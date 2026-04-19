package middleware

import (
	"context"
	"net/http"
)

type ContextKey string

const (
	UserIDKey      ContextKey = "userID"
	RoleKey        ContextKey = "role"
	PermissionsKey ContextKey = "permissions"
)

func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, err := GetSessionFromRequest(r)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		session, err := GetSession(token)
		if err != nil {
			switch err {
			case ErrSessionExpired:
				http.Error(w, "session expired", http.StatusUnauthorized)
				return
			case ErrSessionIdle:
				http.Error(w, "session idle timeout", http.StatusUnauthorized)
				return
			default:
				http.Error(w, "invalid session", http.StatusUnauthorized)
				return
			}
		}

		_ = TouchSession(session)

		ctx := r.Context()
		ctx = context.WithValue(ctx, UserIDKey, session.UserID)
		ctx = context.WithValue(ctx, RoleKey, session.RoleName)
		ctx = context.WithValue(ctx, PermissionsKey, session.Permissions)

		next(w, r.WithContext(ctx))
	}
}

func RequireRole(roles ...string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return RequireAuth(func(w http.ResponseWriter, r *http.Request) {
			role := r.Context().Value(RoleKey).(string)

			for _, allowed := range roles {
				if role == allowed {
					next(w, r)
					return
				}
			}

			http.Error(w, "forbidden", http.StatusForbidden)
		})
	}
}
