package recovery

import (
	"FileLogix/utilities/logger"
	"net/http"
	"runtime/debug"

	"github.com/google/uuid"
)

func PanicPrevent(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger.Errorf(uuid.Nil, uuid.Nil, "Panic recovered: %v\n%s", err, debug.Stack())
				w.WriteHeader(http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
