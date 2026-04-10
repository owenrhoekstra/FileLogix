package authentication

import (
	"encoding/json"
	"log"
	"net/http"

	"file-tracker-backend/sessions"
)

func RegisterChallengeHandler(w http.ResponseWriter, r *http.Request) {
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
		log.Println("getUser error:", err)
		http.Error(w, "user lookup failed", http.StatusInternalServerError)
		return
	}

	// 🔥 CHECK: if user already has a credential, they should login instead
	if userAlreadyHasCredential(u.ID) {
		log.Println("User", req.Email, "already has a passkey, registration blocked")
		http.Error(w, "user already has a passkey, use login instead", http.StatusBadRequest)
		return
	}

	options, sessionData, err := webAuthn.BeginRegistration(u)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Registration challenge created - RPID: %s", options.Response.RelyingParty.ID)

	sessionID := regSessions.set(req.Email, sessionData)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"options":   options,
		"sessionId": sessionID,
	})
}

func RegisterVerifyHandler(w http.ResponseWriter, r *http.Request) {
	// 🔥 READ FROM HEADERS NOW
	email := r.Header.Get("X-Email")
	sessionID := r.Header.Get("X-Session-Id")

	if email == "" || sessionID == "" {
		http.Error(w, "missing email or sessionId", http.StatusBadRequest)
		return
	}

	sessionData, ok := regSessions.get(sessionID)
	if !ok || sessionData == nil {
		http.Error(w, "missing session", http.StatusBadRequest)
		return
	}

	u, err := getUser(email)
	if err != nil {
		http.Error(w, "user lookup failed", http.StatusInternalServerError)
		return
	}

	if userAlreadyHasCredential(u.ID) {
		http.Error(w, "already registered", http.StatusBadRequest)
		return
	}

	credential, err := webAuthn.FinishRegistration(u, *sessionData, r)
	if err != nil {
		log.Println("FinishRegistration error:", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Println("Registration succeeded, credential ID:", credential.ID)

	if err := saveCredential(u.ID, credential); err != nil {
		log.Println("saveCredential error:", err)
		http.Error(w, "failed to save credential", http.StatusInternalServerError)
		return
	}

	// create login session after successful registration
	token, err := sessions.CreateSession(u.ID)
	if err != nil {
		log.Println("session creation error:", err)
		http.Error(w, "failed to create session", http.StatusInternalServerError)
		return
	}

	sessions.SetSessionCookie(w, token)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}

func userAlreadyHasCredential(userID []byte) bool {
	creds, err := getCredentialsByUserID(userID)
	if err != nil {
		return false
	}
	return len(creds) > 0
}
