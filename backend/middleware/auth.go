package middleware

import (
	"context"
	"net/http"
)

type ContextKey string

const UserIDKey ContextKey = "userID"

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

		// extend idle timer on every valid request
		_ = TouchSession(session)

		ctx := r.Context()
		ctx = context.WithValue(ctx, UserIDKey, session.UserID)

		next(w, r.WithContext(ctx))
	}
}
