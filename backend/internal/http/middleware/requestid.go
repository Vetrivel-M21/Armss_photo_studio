package middleware

import (
	"net/http"

	"github.com/google/uuid"

	"armss-poster-studio/internal/observability"
)

// RequestID stamps every request with a unique id, stored on the context and
// echoed back as X-Request-Id, for correlating logs with client-visible errors.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-Id")
		if id == "" {
			id = uuid.NewString()
		}
		w.Header().Set("X-Request-Id", id)
		ctx := observability.WithRequestID(r.Context(), id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
