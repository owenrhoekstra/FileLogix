package documentTypes

import (
	"encoding/json"
	"net/http"
	"strings"

	"FileLogix/middleware"
	"FileLogix/utilities/logger"

	"github.com/google/uuid"
)

func AddDocumentType(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := ctx.Value(middleware.RequestIDKey).(uuid.UUID)
	userID := ctx.Value(middleware.UserIDKey).(uuid.UUID)

	var body struct {
		DocumentLabel string `json:"documentLabel"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || strings.TrimSpace(body.DocumentLabel) == "" {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := addDocumentTypeService(ctx, strings.TrimSpace(body.DocumentLabel)); err != nil {
		logger.Errorf(requestID, userID, "AddDocumentType: service error: %v", err)
		http.Error(w, "failed to add document type", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
