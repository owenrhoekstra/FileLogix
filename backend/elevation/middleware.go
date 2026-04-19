package elevation

import (
	"net/http"

	"FileLogix/middleware"
)

func RequireActionElevation(next http.HandlerFunc) http.HandlerFunc {
	return middleware.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		token, _ := middleware.GetSessionFromRequest(r)

		_, ok := GetElevation(token, ActionElevation)
		if !ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(`{"require_elevation":"action"}`))
			return
		}

		_ = TouchElevation(token, ActionElevation)
		next(w, r)
	})
}

func RequireViewElevation(next http.HandlerFunc) http.HandlerFunc {
	return middleware.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		token, _ := middleware.GetSessionFromRequest(r)

		_, ok := GetElevation(token, ViewElevation)
		if !ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(`{"require_elevation":"view"}`))
			return
		}

		_ = TouchElevation(token, ViewElevation)
		next(w, r)
	})
}
