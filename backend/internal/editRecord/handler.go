package editRecord

import (
	"encoding/json"
	"net/http"
	"time"

	"FileLogix/middleware"
	"FileLogix/utilities/logger"

	"github.com/google/uuid"
)

func HandleRecordEdit(w http.ResponseWriter, r *http.Request) {
	requestID := r.Context().Value(middleware.RequestIDKey).(uuid.UUID)
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		logger.Errorf(requestID, userID, "HandleRecordEdit: invalid document id %q: %v", idStr, err)
		http.Error(w, "invalid document id", http.StatusBadRequest)
		return
	}

	var payload editPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		logger.Errorf(requestID, userID, "HandleRecordEdit: failed to decode body: %v", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	dateOfDoc, err := time.Parse("2006-01-02", payload.DateOfDoc)
	if err != nil {
		logger.Errorf(requestID, userID, "HandleRecordEdit: invalid dateOfDoc %q: %v", payload.DateOfDoc, err)
		http.Error(w, "invalid dateOfDoc format", http.StatusBadRequest)
		return
	}

	if err := editRecord(r.Context(), id, payload.Name, payload.Sensitive, payload.Types, dateOfDoc); err != nil {
		logger.Errorf(requestID, userID, "HandleRecordEdit: failed to edit record %s: %v", id, err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func HandleRecordRestore(w http.ResponseWriter, r *http.Request) {
	requestID := r.Context().Value(middleware.RequestIDKey).(uuid.UUID)
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		logger.Errorf(requestID, userID, "HandleRecordRestore: invalid document id %q: %v", idStr, err)
		http.Error(w, "invalid document id", http.StatusBadRequest)
		return
	}

	if err := restoreRecord(r.Context(), id); err != nil {
		logger.Errorf(requestID, userID, "HandleRecordRestore: failed to restore record %s: %v", id, err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
