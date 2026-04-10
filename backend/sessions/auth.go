package sessions

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
			http.Error(w, "invalid session", http.StatusUnauthorized)
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, UserIDKey, session.UserID)

		next(w, r.WithContext(ctx))
	}
}
