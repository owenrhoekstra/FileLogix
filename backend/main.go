package main

import (
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strconv"

	"file-tracker-backend/authentication"
	"file-tracker-backend/database"
	"file-tracker-backend/middleware"
)

func main() {
	authentication.InitWebAuthn()
	database.Init()
	database.InitRedis()
	database.RunMigrations(database.DB)

	// 🔓 PUBLIC ROUTES
	http.HandleFunc("/api/auth/check-email", authentication.CheckEmailHandler)
	http.HandleFunc("/api/auth/passkey/register-challenge", authentication.RegisterChallengeHandler)
	http.HandleFunc("/api/auth/passkey/register-verify", authentication.RegisterVerifyHandler)
	http.HandleFunc("/api/auth/passkey/login-challenge", authentication.LoginChallengeHandler)
	http.HandleFunc("/api/auth/passkey/login-verify", authentication.LoginVerifyHandler)
	http.HandleFunc("/api/auth/logout", middleware.LogoutHandler)

	http.HandleFunc("/api/auth/me", middleware.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(middleware.UserIDKey).([]byte)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"userId": hex.EncodeToString(userID),
		})
	}))

	// 🔒 PROTECTED ROUTES (example — add yours here)
	http.HandleFunc("/api/protected/test", middleware.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(middleware.UserIDKey).([]byte)
		w.Write([]byte("You are authenticated. UserID length: " + strconv.Itoa(len(userID))))
	}))

	http.ListenAndServe(":8080", nil)
}
