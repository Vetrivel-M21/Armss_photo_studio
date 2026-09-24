package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"armss-poster-studio/internal/observability"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

// Flush forwards to the underlying ResponseWriter's http.Flusher. Without
// this, wrapping ResponseWriter in statusRecorder silently breaks streaming
// handlers (SSE) — the embedded interface only promotes Header/Write/
// WriteHeader, not Flush, so a wrapped writer fails a http.Flusher type
// assertion even though the real writer underneath supports it.
func (r *statusRecorder) Flush() {
	if f, ok := r.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// Logging emits one structured JSON log line per request with method, path,
// status, duration, and request id.
func Logging(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rec, r)
			observability.FromContext(r.Context(), logger).Info("http_request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", rec.status,
				"duration_ms", time.Since(start).Milliseconds(),
			)
		})
	}
}
