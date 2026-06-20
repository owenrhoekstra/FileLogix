package middleware

import (
	"context"
	"net/http"

	"FileLogix/utilities/logger"

	"github.com/google/uuid"
)

type ContextKey string

const (
	UserIDKey      ContextKey = "userID"
	RoleKey        ContextKey = "role"
	PermissionsKey ContextKey = "permissions"
)

func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Context().Value(RequestIDKey).(uuid.UUID)

		token, err := GetSessionFromRequest(r)
		if err != nil {
			logger.Infof(requestID, uuid.Nil, "RequireAuth: missing or invalid session cookie: %v", err)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		session, err := GetSession(token)
		if err != nil {
			switch err {
			case ErrSessionExpired:
				logger.Infof(requestID, uuid.Nil, "RequireAuth: session expired")
				http.Error(w, "session expired", http.StatusUnauthorized)
			case ErrSessionIdle:
				logger.Infof(requestID, uuid.Nil, "RequireAuth: session idle timeout")
				http.Error(w, "session idle timeout", http.StatusUnauthorized)
			default:
				logger.Errorf(requestID, uuid.Nil, "RequireAuth: session lookup failed: %v", err)
				http.Error(w, "invalid session", http.StatusUnauthorized)
			}
			return
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
			requestID := r.Context().Value(RequestIDKey).(uuid.UUID)
			role := r.Context().Value(RoleKey).(string)

			for _, allowed := range roles {
				if role == allowed {
					next(w, r)
					return
				}
			}

			logger.Infof(requestID, uuid.Nil, "RequireRole: access denied for role %q", role)
			http.Error(w, "forbidden", http.StatusForbidden)
		})
	}
}
