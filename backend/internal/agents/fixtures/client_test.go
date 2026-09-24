package fixtures

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"armss-poster-studio/internal/agents"
)

func testToolCallRequest() agents.ToolCallRequest {
	return agents.ToolCallRequest{
		System:   "test system prompt",
		UserText: "test user text",
		Tool: agents.ToolSpec{
			Name:        "emit_style_profile",
			Description: "test tool",
			InputSchema: map[string]any{"type": "object", "properties": map[string]any{}},
		},
		MaxTokens: 100,
	}
}

func TestCallTool_Success(t *testing.T) {
	client := newFixtureServer(t, loadFixture(t, "style_response.json"))

	result, err := client.CallTool(context.Background(), testToolCallRequest())
	if err != nil {
		t.Fatalf("CallTool returned error: %v", err)
	}
	if result.TokensIn != 512 || result.TokensOut != 240 {
		t.Fatalf("unexpected usage: %+v", result)
	}
	if len(result.Raw) == 0 {
		t.Fatal("expected non-empty raw tool input")
	}
}

func TestCallTool_DisabledWithoutKey(t *testing.T) {
	client := agents.NewClient("", "claude-sonnet-4-6", "http://unused.invalid", 5*time.Second)

	_, err := client.CallTool(context.Background(), testToolCallRequest())
	if err != agents.ErrAgentsUnavailable {
		t.Fatalf("expected ErrAgentsUnavailable, got %v", err)
	}
}

func TestCallTool_RetriesOnMalformedThenSucceeds(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&calls, 1)
		w.Header().Set("Content-Type", "application/json")
		if n == 1 {
			// Simulate a truncated/malformed response — must be retried, not crash.
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"content": [{"type": "tool_use", "input": `)) // invalid JSON
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(loadFixture(t, "style_response.json"))
	}))
	defer srv.Close()

	client := agents.NewClient("test-key", "claude-sonnet-4-6", srv.URL, 5*time.Second)
	result, err := client.CallTool(context.Background(), testToolCallRequest())
	if err != nil {
		t.Fatalf("expected eventual success after retry, got error: %v", err)
	}
	if atomic.LoadInt32(&calls) < 2 {
		t.Fatalf("expected at least 2 calls (initial + retry), got %d", calls)
	}
	if len(result.Raw) == 0 {
		t.Fatal("expected non-empty raw tool input after retry")
	}
}

func TestCallTool_RetriesOnEmptyToolUse(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&calls, 1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if n < 3 {
			// Valid JSON but no tool_use block — e.g. the model stopped early.
			_, _ = w.Write([]byte(`{"content": [{"type": "text", "text": "sorry"}], "stop_reason": "max_tokens", "usage": {"input_tokens": 10, "output_tokens": 5}}`))
			return
		}
		_, _ = w.Write(loadFixture(t, "style_response.json"))
	}))
	defer srv.Close()

	client := agents.NewClient("test-key", "claude-sonnet-4-6", srv.URL, 5*time.Second)
	_, err := client.CallTool(context.Background(), testToolCallRequest())
	if err != nil {
		t.Fatalf("expected eventual success after retries, got error: %v", err)
	}
	if atomic.LoadInt32(&calls) != 3 {
		t.Fatalf("expected exactly 3 calls, got %d", calls)
	}
}

func TestCallTool_DoesNotRetryOn400(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error": {"type": "invalid_request_error", "message": "bad schema"}}`))
	}))
	defer srv.Close()

	client := agents.NewClient("test-key", "claude-sonnet-4-6", srv.URL, 5*time.Second)
	_, err := client.CallTool(context.Background(), testToolCallRequest())
	if err == nil {
		t.Fatal("expected an error for a 400 response")
	}
	if atomic.LoadInt32(&calls) != 1 {
		t.Fatalf("expected exactly 1 call for a non-retryable 400, got %d", calls)
	}
}

func TestCallTool_TimesOut(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(loadFixture(t, "style_response.json"))
	}))
	defer srv.Close()

	client := agents.NewClient("test-key", "claude-sonnet-4-6", srv.URL, 20*time.Millisecond)
	_, err := client.CallTool(context.Background(), testToolCallRequest())
	if err == nil {
		t.Fatal("expected a timeout error")
	}
}
