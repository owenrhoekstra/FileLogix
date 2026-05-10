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
			w.Header().Set("X-Require-Elevation", "action")
			w.Header().Set("X-Toast", "You lack the privileges to perform this elevated action")
			w.WriteHeader(http.StatusForbidden)
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
			w.Header().Set("X-Require-Elevation", "view")
			w.Header().Set("X-Toast", "You lack the privileges to perform this elevated action")
			w.WriteHeader(http.StatusForbidden)
			return
		}

		_ = TouchElevation(token, ViewElevation)
		next(w, r)
	})
}
