package authentication

import (
	"encoding/json"
	"net/http"

	"FileLogix/middleware"
	"FileLogix/utilities/logger"

	"github.com/google/uuid"
)

func CheckEmailHandler(w http.ResponseWriter, r *http.Request) {
	requestID := r.Context().Value(middleware.RequestIDKey).(uuid.UUID)

	var req EmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Infof(requestID, uuid.Nil, "CheckEmailHandler: failed to decode request: %v", err)
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if !isAllowed(req.Email) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"hasPasskey": false})
		return
	}

	u, err := getUser(req.Email)
	if err != nil {
		logger.Errorf(requestID, uuid.Nil, "CheckEmailHandler: getUser failed: %v", err)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"hasPasskey": false})
		return
	}

	hasPasskey := false
	if u != nil {
		creds, err := getCredentialsByUserID(u.ID)
		if err != nil {
			logger.Errorf(requestID, u.ID, "CheckEmailHandler: getCredentialsByUserID failed: %v", err)
		} else {
			hasPasskey = len(creds) > 0
		}
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{"hasPasskey": hasPasskey})
}
