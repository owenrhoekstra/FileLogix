package middleware

import (
	"net/http"
)

func RequirePermission(permission string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return RequireAuth(func(w http.ResponseWriter, r *http.Request) {
			permissions, ok := r.Context().Value(PermissionsKey).(map[string]bool)
			if !ok {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}

			if !permissions[permission] {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}

			next(w, r)
		})
	}
}
