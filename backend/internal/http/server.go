package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

// NewServer wraps the router in an http.Server with sane timeouts.
func NewServer(port string, router *chi.Mux) *http.Server {
	return &http.Server{
		Addr:              fmt.Sprintf(":%s", port),
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      0, // SSE streams hold connections open; per-request timeouts are set via context instead
		IdleTimeout:       120 * time.Second,
	}
}

// Shutdown gracefully stops the server, waiting up to the given timeout for
// in-flight requests (e.g. an open SSE stream) to finish.
func Shutdown(srv *http.Server, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return srv.Shutdown(ctx)
}
