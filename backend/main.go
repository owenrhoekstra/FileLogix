package main

import (
	"FileLogix/authentication"
	"FileLogix/database"
	"FileLogix/elevation"
	"FileLogix/middleware"
	"FileLogix/routes"
	"encoding/hex"
	"encoding/json"
	"net/http"

	"github.com/didip/tollbooth/v7"
)

func main() {
	authentication.InitWebAuthn()
	database.Init()
	database.InitRedis()
	database.RunMigrations(database.DB)

	// Wire WebAuthn instance into elevation package
	elevation.WebAuthn = authentication.GetWebAuthn()

	emailLimiter := tollbooth.NewLimiter(1, nil)
	authLimiter := tollbooth.NewLimiter(3, nil)
	emailLimiter.SetIPLookups([]string{"CF-Connecting-IP", "X-Forwarded-For", "RemoteAddr"})
	authLimiter.SetIPLookups([]string{"CF-Connecting-IP", "X-Forwarded-For", "RemoteAddr"})

	mux := http.NewServeMux()

	// 🔓 PUBLIC ROUTES
	mux.Handle("/api/auth/check-email",
		middleware.RateLimit(emailLimiter)(
			http.HandlerFunc(authentication.CheckEmailHandler),
		),
	)

	mux.Handle("/api/auth/passkey/register-challenge",
		middleware.RateLimit(authLimiter)(
			http.HandlerFunc(authentication.RegisterChallengeHandler),
		),
	)

	mux.Handle("/api/auth/passkey/register-verify",
		middleware.RateLimit(authLimiter)(
			http.HandlerFunc(authentication.RegisterVerifyHandler),
		),
	)

	mux.Handle("/api/auth/passkey/login-challenge",
		middleware.RateLimit(authLimiter)(
			http.HandlerFunc(authentication.LoginChallengeHandler),
		),
	)

	mux.Handle("/api/auth/passkey/login-verify",
		middleware.RateLimit(authLimiter)(
			http.HandlerFunc(authentication.LoginVerifyHandler),
		),
	)

	mux.Handle("/api/auth/logout",
		middleware.RateLimit(authLimiter)(
			http.HandlerFunc(middleware.LogoutHandler),
		),
	)

	// 🔒 AUTHENTICATED ROUTES
	mux.Handle("/api/auth/me",
		middleware.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
			userID := r.Context().Value(middleware.UserIDKey).([]byte)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{
				"userId": hex.EncodeToString(userID),
			})
		}),
	)

	// 🔒 ELEVATION ROUTES
	mux.Handle("/api/auth/elevate/challenge",
		middleware.RateLimit(authLimiter)(
			middleware.RequireAuth(
				http.HandlerFunc(elevation.ChallengeHandler),
			),
		),
	)

	mux.Handle("/api/auth/elevate/verify",
		middleware.RateLimit(authLimiter)(
			middleware.RequireAuth(
				http.HandlerFunc(elevation.VerifyHandler),
			),
		),
	)

	// 🔒 PROTECTED ROUTES
	mux.Handle("/api/protected/",
		middleware.RateLimit(authLimiter)(
			http.StripPrefix("/api/protected", routes.ProtectedRoutes()),
		),
	)

	handler := middleware.CORS(mux)
	handler = middleware.SecurityHeaders(handler)

	http.ListenAndServe(":8080", handler)
}
