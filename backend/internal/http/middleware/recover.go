package middleware

import (
	"log/slog"
	"net/http"

	"armss-poster-studio/internal/http/respond"
	"armss-poster-studio/internal/observability"
)

// Recover converts a panic in any handler into a 500 response instead of
// crashing the server, and logs the panic with its request id.
func Recover(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					observability.FromContext(r.Context(), logger).Error("panic_recovered", "panic", rec)
					respond.Error(w, http.StatusInternalServerError, "Something went wrong. Please try again.")
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
