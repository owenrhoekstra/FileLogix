package viewRecord

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"strings"

	"FileLogix/internal"
	"FileLogix/middleware"
	"FileLogix/utilities/logger"

	"github.com/google/uuid"
)

func FetchRecordList(w http.ResponseWriter, r *http.Request) {
	requestID := r.Context().Value(middleware.RequestIDKey).(uuid.UUID)
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	q := r.URL.Query()

	perms, err := middleware.GetUserPermissions(r)
	if err != nil {
		logger.Errorf(requestID, userID, "FetchRecordList: failed to get permissions: %v", err)
		perms = map[string]bool{}
	}

	canViewSensitive := perms["can_view_sensitive"]
	canDelete := perms["can_delete"]

	var types []string
	if t := q.Get("types"); t != "" {
		types = strings.Split(t, ",")
	}

	params := fetchParams{
		SortBy:        q.Get("sortBy"),
		Limit:         q.Get("limit"),
		Offset:        q.Get("offset"),
		Name:          q.Get("name"),
		Types:         types,
		DocDateFrom:   q.Get("docDateFrom"),
		DocDateTo:     q.Get("docDateTo"),
		FiledDateFrom: q.Get("filedDateFrom"),
		FiledDateTo:   q.Get("filedDateTo"),
	}

	results, err := fetchRecords(r.Context(), params, canViewSensitive, canDelete)
	if err != nil {
		logger.Errorf(requestID, userID, "FetchRecordList: failed to fetch records: %v", err)
		http.Error(w, "failed to fetch records", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func FetchRecordDetails(w http.ResponseWriter, r *http.Request) {
	requestID := r.Context().Value(middleware.RequestIDKey).(uuid.UUID)
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		logger.Errorf(requestID, userID, "FetchRecordDetails: invalid document id %q: %v", idStr, err)
		http.Error(w, "invalid document id", http.StatusBadRequest)
		return
	}

	perms, err := middleware.GetUserPermissions(r)
	if err != nil {
		logger.Errorf(requestID, userID, "FetchRecordDetails: failed to get permissions: %v", err)
		perms = map[string]bool{}
	}

	canDelete := perms["can_delete"]

	detail, err := buildDocumentDetail(r.Context(), id, canDelete)
	if err != nil {
		logger.Errorf(requestID, userID, "FetchRecordDetails: failed to build detail for doc %s: %v", id, err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if detail == nil {
		logger.Infof(requestID, userID, "FetchRecordDetails: doc %s not found", id)
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(detail)
}

func DeleteRecord(w http.ResponseWriter, r *http.Request) {
	requestID := r.Context().Value(middleware.RequestIDKey).(uuid.UUID)
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	docID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		logger.Errorf(requestID, userID, "DeleteDocument: invalid document id: %v", err)
		http.Error(w, "invalid document id", http.StatusBadRequest)
		return
	}

	if err := softDeleteDocument(requestID, docID, userID); err != nil {
		logger.Errorf(requestID, userID, "DeleteDocument: db error for doc %s: %v", docID, err)
		http.Error(w, "failed to delete document", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

var allowedExts = map[string]bool{
	".webp": true,
	".avif": true,
}

func FetchDocumentImages(w http.ResponseWriter, r *http.Request) {
	requestID := r.Context().Value(middleware.RequestIDKey).(uuid.UUID)
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	rawPath := r.PathValue("path")
	if rawPath == "" {
		logger.Infof(requestID, userID, "FetchDocumentImages: empty path")
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	cleaned := filepath.Clean("/" + rawPath)
	if strings.Contains(cleaned, "..") {
		logger.Errorf(requestID, userID, "FetchDocumentImages: path traversal attempt: %s", rawPath)
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}

	ext := strings.ToLower(filepath.Ext(cleaned))
	if !allowedExts[ext] {
		logger.Errorf(requestID, userID, "FetchDocumentImages: forbidden extension %q for path %s", ext, cleaned)
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	fullPath := filepath.Join(internal.StorageRoot, cleaned)
	http.ServeFile(w, r, fullPath)
}
