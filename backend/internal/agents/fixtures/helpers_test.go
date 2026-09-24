// Package fixtures holds recorded-fixture tests for every agent, run against an
// httptest fake Anthropic server so no live API calls happen in CI (spec §13).
package fixtures

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"armss-poster-studio/internal/agents"
)

// loadFixture reads a recorded Anthropic /v1/messages response body from testdata.
func loadFixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	return data
}

// newFixtureServer starts an httptest server that always serves body verbatim as
// the /v1/messages response, and a Client pointed at it.
func newFixtureServer(t *testing.T, body []byte) *agents.Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(body)
	}))
	t.Cleanup(srv.Close)
	return agents.NewClient("test-key", "claude-sonnet-4-6", srv.URL, 5*time.Second)
}
