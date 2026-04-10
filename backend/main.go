package main

import (
	"net/http"
	"strconv"

	"file-tracker-backend/authentication"
	"file-tracker-backend/database"
	"file-tracker-backend/sessions"
)

func main() {
	authentication.InitWebAuthn()
	database.Init()

	// 🔓 PUBLIC ROUTES
	http.HandleFunc("/api/auth/check-email", authentication.CheckEmailHandler)
	http.HandleFunc("/api/auth/passkey/register-challenge", authentication.RegisterChallengeHandler)
	http.HandleFunc("/api/auth/passkey/register-verify", authentication.RegisterVerifyHandler)
	http.HandleFunc("/api/auth/passkey/login-challenge", authentication.LoginChallengeHandler)
	http.HandleFunc("/api/auth/passkey/login-verify", authentication.LoginVerifyHandler)

	// 🔒 PROTECTED ROUTES (example — add yours here)
	http.HandleFunc("/api/protected/test", sessions.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(sessions.UserIDKey).([]byte)
		w.Write([]byte("You are authenticated. UserID length: " + strconv.Itoa(len(userID))))
	}))

	http.ListenAndServe(":8000", nil)
}
