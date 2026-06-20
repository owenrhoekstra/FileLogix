package authentication

import (
	"encoding/json"
	"net/http"

	"FileLogix/database"
	"FileLogix/middleware"
	"FileLogix/utilities/logger"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
)

func RegisterChallengeHandler(w http.ResponseWriter, r *http.Request) {
	requestID := r.Context().Value(middleware.RequestIDKey).(uuid.UUID)

	var req EmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Infof(requestID, uuid.Nil, "RegisterChallengeHandler: failed to decode request: %v", err)
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if req.Email == "" {
		http.Error(w, "missing email", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if !isAllowed(req.Email) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"options":   nil,
			"sessionId": "",
		})
		return
	}

	u, err := getUser(req.Email)
	if err != nil {
		logger.Errorf(requestID, uuid.Nil, "RegisterChallengeHandler: getUser failed: %v", err)
		http.Error(w, "user lookup failed", http.StatusInternalServerError)
		return
	}

	options, sessionData, err := webAuthn.BeginRegistration(
		u,
		webauthn.WithExclusions(credentialsToDescriptors(u.WebAuthnCredentials())),
	)
	if err != nil {
		logger.Errorf(requestID, u.ID, "RegisterChallengeHandler: BeginRegistration failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sessionID := regSessions.set(req.Email, sessionData)

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"options":   options,
		"sessionId": sessionID,
	})
}

func RegisterVerifyHandler(w http.ResponseWriter, r *http.Request) {
	requestID := r.Context().Value(middleware.RequestIDKey).(uuid.UUID)

	email := r.Header.Get("X-Email")
	sessionID := r.Header.Get("X-Session-Id")

	if email == "" || sessionID == "" {
		http.Error(w, "missing email or sessionId", http.StatusBadRequest)
		return
	}

	sessionData, ok := regSessions.get(sessionID)
	if !ok || sessionData == nil {
		logger.Infof(requestID, uuid.Nil, "RegisterVerifyHandler: session not found for sessionID %q", sessionID)
		http.Error(w, "missing session", http.StatusBadRequest)
		return
	}

	u, err := getUser(email)
	if err != nil {
		logger.Errorf(requestID, uuid.Nil, "RegisterVerifyHandler: getUser failed: %v", err)
		http.Error(w, "user lookup failed", http.StatusInternalServerError)
		return
	}

	credential, err := webAuthn.FinishRegistration(u, *sessionData, r)
	if err != nil {
		logger.Infof(requestID, u.ID, "RegisterVerifyHandler: FinishRegistration failed: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := saveCredential(u.ID, credential); err != nil {
		logger.Errorf(requestID, u.ID, "RegisterVerifyHandler: saveCredential failed: %v", err)
		http.Error(w, "failed to save credential", http.StatusInternalServerError)
		return
	}

	roleName, permissions, err := database.GetUserRole(u.ID)
	if err != nil {
		logger.Errorf(requestID, u.ID, "RegisterVerifyHandler: GetUserRole failed: %v", err)
		http.Error(w, "failed to load user role", http.StatusInternalServerError)
		return
	}

	token, err := middleware.CreateSession(u.ID, roleName, permissions)
	if err != nil {
		logger.Errorf(requestID, u.ID, "RegisterVerifyHandler: CreateSession failed: %v", err)
		http.Error(w, "failed to create session", http.StatusInternalServerError)
		return
	}

	middleware.SetSessionCookie(w, token)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func userAlreadyHasCredential(userID uuid.UUID) bool {
	creds, err := getCredentialsByUserID(userID)
	if err != nil {
		return false
	}
	return len(creds) > 0
}
