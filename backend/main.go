package main

import (
	"FileLogix/authentication"
	"FileLogix/daemons"
	"FileLogix/database"
	"FileLogix/elevation"
	"FileLogix/middleware"
	"FileLogix/routes"
	"FileLogix/utilities/recovery"

	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/didip/tollbooth/v7"
	"github.com/google/uuid"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

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
			userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

			var metadataComplete bool
			var email string
			err := database.DB.QueryRow(`
            SELECT metadata_complete, email FROM users WHERE id = $1
        `, userID).Scan(&metadataComplete, &email)
			if err != nil {
				http.Error(w, "failed to load user", http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"userId":           userID.String(),
				"metadataComplete": metadataComplete,
				"email":            email,
			})
		}),
	)

	mux.Handle("/api/user/",
		http.StripPrefix("/api/user", routes.UserRoutes()),
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
			middleware.RequireAuth(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					http.StripPrefix("/api/protected", routes.ProtectedRoutes()).ServeHTTP(w, r)
				}),
			),
		),
	)

	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
	})

	handler := middleware.CORS(mux)
	handler = middleware.SecurityHeaders(handler)
	handler = middleware.WithRequestID(handler)
	handler = http.TimeoutHandler(handler, 30*time.Second, `{"error":"request timeout"}`)
	handler = recovery.PanicPrevent(handler)

	daemons.StartAll(ctx)

	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatal(err)
	}
}
