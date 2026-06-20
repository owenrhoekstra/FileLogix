package middleware

import (
	"context"
	"net/http"

	"FileLogix/utilities/logger"

	"github.com/google/uuid"
)

const RequestIDKey ContextKey = "requestID"

func WithRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := uuid.New()

		ctx := context.WithValue(r.Context(), RequestIDKey, id)

		w.Header().Set("X-Request-ID", id.String())

		logger.Infof(id, uuid.Nil, "%s %s", r.Method, r.URL.Path)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
