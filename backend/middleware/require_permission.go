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
				w.Header().Set("X-Toast", "You lack the necessary permissions to complete the requested action")
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}

			next(w, r)
		})
	}
}

func GetUserPermissions(r *http.Request) (map[string]bool, error) {

	cookie, err := GetSessionFromRequest(r)
	if err != nil {
		return nil, err
	}
	session, err := GetSession(cookie)
	if err != nil {
		return nil, err
	}
	return session.Permissions, nil

}
