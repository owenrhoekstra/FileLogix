package createRecord

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"FileLogix/database"
	"FileLogix/middleware"
	"FileLogix/utilities/logger"

	"github.com/google/uuid"
)

const (
	maxFormSize = 500 << 20 // 500MB
	maxFileSize = 25 << 20  // 25MB
)

func Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID, reqOk := ctx.Value(middleware.RequestIDKey).(uuid.UUID)
	userID, userOk := ctx.Value(middleware.UserIDKey).(uuid.UUID)
	if !reqOk || !userOk {
		logger.Errorf(uuid.Nil, uuid.Nil, "Create: missing requestID or userID in context")
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	if err := r.ParseMultipartForm(maxFormSize); err != nil {
		logger.Errorf(requestID, userID, "Create: ParseMultipartForm error: %v", err)
		http.Error(w, "invalid form data", http.StatusBadRequest)
		return
	}

	// ---- FIELD EXTRACTION ----
	documentName := strings.TrimSpace(r.FormValue("documentName"))
	documentDate := strings.TrimSpace(r.FormValue("documentDate"))
	documentSensitivityRaw := strings.TrimSpace(r.FormValue("documentSensitivity"))
	documentTypes := r.Form["documentType"]
	files := r.MultipartForm.File["photos"]

	// ---- VALIDATION ----
	var validationErrors []string

	if documentName == "" || len(documentName) > 50 {
		validationErrors = append(validationErrors, "documentName must be between 1 and 50 characters")
	}

	if documentDate == "" {
		validationErrors = append(validationErrors, "documentDate is required")
	} else if _, err := time.Parse("2006-01-02", documentDate); err != nil {
		validationErrors = append(validationErrors, "documentDate must be a valid date (YYYY-MM-DD)")
	}

	sensitivity := false
	switch documentSensitivityRaw {
	case "true":
		sensitivity = true
	case "false":
		sensitivity = false
	default:
		validationErrors = append(validationErrors, "documentSensitivity must be true or false")
	}

	if len(documentTypes) == 0 || len(documentTypes) > 3 {
		validationErrors = append(validationErrors, "documentType must have between 1 and 3 values")
	}

	if len(files) == 0 {
		validationErrors = append(validationErrors, "at least one file is required")
	}

	for _, fh := range files {
		if fh.Size > maxFileSize {
			validationErrors = append(validationErrors, fh.Filename+": exceeds 25MB limit")
			continue
		}
		if fh.Header.Get("Content-Type") != "image/webp" {
			validationErrors = append(validationErrors, fh.Filename+": only image/webp is accepted")
		}
	}

	if len(validationErrors) > 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"errors": validationErrors,
		})
		return
	}

	// ---- BUILD INPUT ----
	input := CreateRecordInput{
		Name:       documentName,
		DateOfDoc:  documentDate,
		Sensitive:  sensitivity,
		Types:      documentTypes,
		Files:      files,
		UploadedBy: userID,
	}

	// ---- SERVICE ----
	documentID, err := CreateRecord(ctx, input)
	if err != nil {
		logger.Errorf(requestID, userID, "Create: CreateRecord error: %v", err)
		http.Error(w, "failed to create record", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"id": documentID.String(),
	})
}

func PrintLabel(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID, reqOk := ctx.Value(middleware.RequestIDKey).(uuid.UUID)
	userID, userOk := ctx.Value(middleware.UserIDKey).(uuid.UUID)
	if !reqOk || !userOk {
		logger.Errorf(uuid.Nil, uuid.Nil, "PrintLabel: missing requestID or userID in context")
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		logger.Errorf(requestID, userID, "PrintLabel: missing id path value")
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	pdfBytes, err := GenerateLabel(id)
	if err != nil {
		logger.Errorf(requestID, userID, "PrintLabel: GenerateLabel failed for id %s: %v", id, err)
		http.Error(w, "failed to generate label", http.StatusInternalServerError)
		return
	}

	w.Header().Set("X-Frame-Options", "SAMEORIGIN")
	w.Header().Set("Content-Security-Policy",
		"default-src 'self'; "+
			"script-src 'self'; "+
			"style-src 'self' 'unsafe-inline'; "+
			"img-src 'self' data:; "+
			"connect-src 'self'; "+
			"object-src 'self'; "+
			"base-uri 'self'; "+
			"frame-ancestors 'self';")
	w.Header().Set("Content-Encoding", "identity")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "inline")

	w.Write(pdfBytes)
}

func TypeOptions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := ctx.Value(middleware.RequestIDKey).(uuid.UUID)
	userID := ctx.Value(middleware.UserIDKey).(uuid.UUID)

	var value json.RawMessage
	err := database.DB.QueryRowContext(ctx, `
		SELECT value FROM settings WHERE key = 'document_types'
	`).Scan(&value)
	if err != nil {
		logger.Errorf(requestID, userID, "TypeOptions: scan error: %v", err)
		http.Error(w, "failed to load document types", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"documentTypes":`))
	w.Write(value)
	w.Write([]byte(`}`))
}
