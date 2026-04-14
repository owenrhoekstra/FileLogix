package authentication

import (
	"encoding/json"
	"log"
	"net/http"

	"FileLogix/middleware"

	"github.com/go-webauthn/webauthn/webauthn"
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

	w.Header().Set("Content-Type", "application/json")

	if !isAllowed(req.Email) {
		// Return fake 200 to prevent email enumeration
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"options":   nil,
			"sessionId": "",
		})
		return
	}

	u, err := getUser(req.Email)
	if err != nil {
		log.Println("getUser error:", err)
		http.Error(w, "user lookup failed", http.StatusInternalServerError)
		return
	}

	options, sessionData, err := webAuthn.BeginRegistration(
		u,
		webauthn.WithExclusions(credentialsToDescriptors(u.WebAuthnCredentials())),
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Registration challenge created - RPID: %s", options.Response.RelyingParty.ID)

	sessionID := regSessions.set(req.Email, sessionData)

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"options":   options,
		"sessionId": sessionID,
	})
}

func RegisterVerifyHandler(w http.ResponseWriter, r *http.Request) {
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

func userAlreadyHasCredential(userID []byte) bool {
	creds, err := getCredentialsByUserID(userID)
	if err != nil {
		return false
	}
	return len(creds) > 0
}
