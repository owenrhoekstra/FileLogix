package authentication

import (
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"

	"file-tracker-backend/database"
	"file-tracker-backend/middleware"
)

func LoginChallengeHandler(w http.ResponseWriter, r *http.Request) {
	var req EmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
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
		http.Error(w, "user lookup failed", http.StatusInternalServerError)
		return
	}

	options, sessionData, err := webAuthn.BeginLogin(u)
	if err != nil {
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
	// 🔥 READ FROM HEADERS (same as registration)
	email := r.Header.Get("X-Email")
	sessionID := r.Header.Get("X-Session-Id")

	if email == "" || sessionID == "" {
		http.Error(w, "missing email or sessionId", http.StatusBadRequest)
		return
	}

	sessionData, ok := loginSessions.get(sessionID)
	if !ok || sessionData == nil {
		http.Error(w, "missing session", http.StatusBadRequest)
		return
	}

	u, err := getUser(email)
	if err != nil {
		http.Error(w, "user lookup failed", http.StatusInternalServerError)
		return
	}

	credential, err := webAuthn.FinishLogin(u, *sessionData, r)
	if err != nil {
		log.Println("FinishLogin error:", err)
		http.Error(w, "authentication failed", http.StatusUnauthorized)
		return
	}

	// Update stored credential flags (handles BE/BS flag drift)
	credIDHex := hex.EncodeToString(credential.ID)
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
		credIDHex,
		hex.EncodeToString(u.ID),
	)
	if err != nil {
		log.Println("credential update error:", err)
		// non-fatal, session is still valid
	}

	// 🔥 CREATE REAL SESSION (this is what you were missing)
	token, err := middleware.CreateSession(u.ID)
	if err != nil {
		log.Println("session creation error:", err)
		http.Error(w, "failed to create session", http.StatusInternalServerError)
		return
	}

	middleware.SetSessionCookie(w, token)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}
