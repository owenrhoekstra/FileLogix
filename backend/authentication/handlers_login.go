package authentication

import (
	"encoding/json"
	"net/http"

	"FileLogix/database"
	"FileLogix/middleware"
	"FileLogix/utilities/logger"

	"github.com/google/uuid"
)

func LoginChallengeHandler(w http.ResponseWriter, r *http.Request) {
	requestID := r.Context().Value(middleware.RequestIDKey).(uuid.UUID)

	var req EmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Infof(requestID, uuid.Nil, "LoginChallengeHandler: failed to decode request: %v", err)
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if req.Email == "" {
		http.Error(w, "missing email", http.StatusBadRequest)
		return
	}

	if !isAllowed(req.Email) {
		http.Error(w, "email not allowed", http.StatusForbidden)
		return
	}

	u, err := getUser(req.Email)
	if err != nil {
		logger.Errorf(requestID, uuid.Nil, "LoginChallengeHandler: getUser failed: %v", err)
		http.Error(w, "user lookup failed", http.StatusInternalServerError)
		return
	}

	options, sessionData, err := webAuthn.BeginLogin(u)
	if err != nil {
		logger.Errorf(requestID, u.ID, "LoginChallengeHandler: BeginLogin failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sessionID := loginSessions.set(req.Email, sessionData)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"options":   options,
		"sessionId": sessionID,
	})
}

func LoginVerifyHandler(w http.ResponseWriter, r *http.Request) {
	requestID := r.Context().Value(middleware.RequestIDKey).(uuid.UUID)

	email := r.Header.Get("X-Email")
	sessionID := r.Header.Get("X-Session-Id")

	if email == "" || sessionID == "" {
		http.Error(w, "missing email or sessionId", http.StatusBadRequest)
		return
	}

	sessionData, ok := loginSessions.get(sessionID)
	if !ok || sessionData == nil {
		logger.Infof(requestID, uuid.Nil, "LoginVerifyHandler: session not found for sessionID %q", sessionID)
		http.Error(w, "missing session", http.StatusBadRequest)
		return
	}

	u, err := getUser(email)
	if err != nil {
		logger.Errorf(requestID, uuid.Nil, "LoginVerifyHandler: getUser failed: %v", err)
		http.Error(w, "user lookup failed", http.StatusInternalServerError)
		return
	}

	credential, err := webAuthn.FinishLogin(u, *sessionData, r)
	if err != nil {
		logger.Infof(requestID, u.ID, "LoginVerifyHandler: FinishLogin failed: %v", err)
		http.Error(w, "authentication failed", http.StatusUnauthorized)
		return
	}

	_, err = database.DB.Exec(`
        UPDATE credentials
        SET backup_eligible = $1,
            backup_state    = $2,
            sign_count      = $3
        WHERE credential_id = $4
          AND user_id       = $5
    `,
		credential.Flags.BackupEligible,
		credential.Flags.BackupState,
		credential.Authenticator.SignCount,
		credential.ID,
		u.ID,
	)
	if err != nil {
		logger.Errorf(requestID, u.ID, "LoginVerifyHandler: credential update failed: %v", err)
	}

	roleName, permissions, err := database.GetUserRole(u.ID)
	if err != nil {
		logger.Errorf(requestID, u.ID, "LoginVerifyHandler: GetUserRole failed: %v", err)
		http.Error(w, "failed to load user role", http.StatusInternalServerError)
		return
	}

	token, err := middleware.CreateSession(u.ID, roleName, permissions)
	if err != nil {
		logger.Errorf(requestID, u.ID, "LoginVerifyHandler: CreateSession failed: %v", err)
		http.Error(w, "failed to create session", http.StatusInternalServerError)
		return
	}

	middleware.SetSessionCookie(w, token)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
