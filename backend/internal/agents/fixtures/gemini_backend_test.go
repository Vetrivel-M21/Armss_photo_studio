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

func newGeminiFixtureServer(t *testing.T, body []byte) *agents.Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(body)
	}))
	t.Cleanup(srv.Close)
	return agents.NewGeminiClient("test-key", "gemini-2.5-flash", srv.URL, 5*time.Second)
}

const geminiStyleResponse = `{
  "candidates": [
    {
      "content": {
        "parts": [
          {
            "functionCall": {
              "name": "emit_style_profile",
              "args": {
                "summary": "Warm festival poster style",
                "palette": [{"hex": "#16265A", "role": "primary"}],
                "typography": "Bold sans",
                "layoutPatterns": "Centered hero",
                "imagery": "Photoreal",
                "finish": "photoreal composite",
                "tone": "celebratory",
                "logoTreatment": "Top-left corner",
                "fixedElements": ["footer bar"],
                "consistency": "High"
              }
            }
          }
        ]
      },
      "finishReason": "STOP"
    }
  ],
  "usageMetadata": {"promptTokenCount": 300, "candidatesTokenCount": 150}
}`

func TestGeminiBackend_CallTool_Success(t *testing.T) {
	client := newGeminiFixtureServer(t, []byte(geminiStyleResponse))

	result, err := client.CallTool(context.Background(), testToolCallRequest())
	if err != nil {
		t.Fatalf("CallTool returned error: %v", err)
	}
	if result.TokensIn != 300 || result.TokensOut != 150 {
		t.Fatalf("unexpected usage: %+v", result)
	}
	if len(result.Raw) == 0 {
		t.Fatal("expected non-empty raw function-call args")
	}
}

func TestGeminiBackend_AnalyzeStyle(t *testing.T) {
	client := newGeminiFixtureServer(t, []byte(geminiStyleResponse))

	profile, usage, err := client.AnalyzeStyle(context.Background(), []agents.ImageInput{
		{Data: []byte("fake-image-bytes"), MediaType: "image/png"},
	})
	if err != nil {
		t.Fatalf("AnalyzeStyle returned error: %v", err)
	}
	if profile.Summary == "" {
		t.Error("expected non-empty summary")
	}
	if len(profile.Palette) != 1 || profile.Palette[0].Hex != "#16265A" {
		t.Errorf("unexpected palette: %+v", profile.Palette)
	}
	if usage.TokensIn != 300 || usage.TokensOut != 150 {
		t.Errorf("unexpected usage: %+v", usage)
	}
}

func TestGeminiBackend_DisabledWithoutKey(t *testing.T) {
	client := agents.NewGeminiClient("", "gemini-2.5-flash", "http://unused.invalid", 5*time.Second)

	_, err := client.CallTool(context.Background(), testToolCallRequest())
	if err != agents.ErrAgentsUnavailable {
		t.Fatalf("expected ErrAgentsUnavailable, got %v", err)
	}
}

func TestGeminiBackend_RetriesOnMissingFunctionCall(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&calls, 1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if n < 2 {
			// A text-only reply with no functionCall — treated like a truncated/refused response.
			_, _ = w.Write([]byte(`{"candidates":[{"content":{"parts":[{"text":"sorry"}]},"finishReason":"STOP"}],"usageMetadata":{"promptTokenCount":10,"candidatesTokenCount":5}}`))
			return
		}
		_, _ = w.Write([]byte(geminiStyleResponse))
	}))
	defer srv.Close()

	client := agents.NewGeminiClient("test-key", "gemini-2.5-flash", srv.URL, 5*time.Second)
	_, err := client.CallTool(context.Background(), testToolCallRequest())
	if err != nil {
		t.Fatalf("expected eventual success after retry, got error: %v", err)
	}
	if atomic.LoadInt32(&calls) != 2 {
		t.Fatalf("expected exactly 2 calls, got %d", calls)
	}
}

func TestGeminiBackend_DoesNotRetryOn400(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error": {"code": 400, "message": "bad schema", "status": "INVALID_ARGUMENT"}}`))
	}))
	defer srv.Close()

	client := agents.NewGeminiClient("test-key", "gemini-2.5-flash", srv.URL, 5*time.Second)
	_, err := client.CallTool(context.Background(), testToolCallRequest())
	if err == nil {
		t.Fatal("expected an error for a 400 response")
	}
	if atomic.LoadInt32(&calls) != 1 {
		t.Fatalf("expected exactly 1 call for a non-retryable 400, got %d", calls)
	}
}
