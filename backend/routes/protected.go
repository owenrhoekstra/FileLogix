package routes

import (
	"FileLogix/middleware"
	"FileLogix/ocr"
	"net/http"
)

func ProtectedRoutes() http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/test",
		middleware.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("protected test"))
		}),
	)

	mux.Handle("/api/auth/elevate/verify",
		middleware.RequireRole("superuser", "manager", "user", "contributor")(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { ocr.OcrEndpoint() }),
		),
	)

	return mux
}
