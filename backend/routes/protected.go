package routes

import (
	"FileLogix/middleware"
	"net/http"
)

func ProtectedRoutes() http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/test",
		middleware.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("protected test"))
		}),
	)

	return mux
}
