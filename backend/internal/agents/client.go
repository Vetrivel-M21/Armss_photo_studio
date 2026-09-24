package agents

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"armss-poster-studio/internal/retry"
)

// ErrAgentsUnavailable is returned by every agent call when the configured
// backend has no API key. Handlers must map this to a 503 naming what's
// missing, never let it propagate as a crash or a stuck request (spec §17,
// acceptance criterion 6).
var ErrAgentsUnavailable = errors.New("agent provider api key not configured")

// ImageInput is a vision-call attachment (used by the Style Analyst).
type ImageInput struct {
	Data      []byte
	MediaType string // e.g. "image/png"
}

// ToolSpec describes the single tool the model is forced to call, so its
// input arrives already shaped to our schema instead of free text. Every
// agent (style.go, intake.go, ...) builds one of these — this shape is
// backend-agnostic, which is what lets the agent files stay unchanged
// regardless of which Backend actually answers the call.
type ToolSpec struct {
	Name        string
	Description string
	InputSchema map[string]any
}

type ToolCallRequest struct {
	System    string
	UserText  string
	Images    []ImageInput
	Tool      ToolSpec
	MaxTokens int
}

type ToolCallResult struct {
	Raw        json.RawMessage
	TokensIn   int
	TokensOut  int
	StopReason string
}

// Backend is one LLM vendor's forced-tool-call implementation. Anthropic
// Claude (backend_anthropic.go) and Google Gemini (backend_gemini.go) both
// implement it; Client is a thin, vendor-agnostic wrapper so every agent
// (style/intake/director/copy/art) is written once against ToolCallRequest/
// ToolCallResult and never touches vendor wire formats directly.
type Backend interface {
	Enabled() bool
	ModelName() string
	CallTool(ctx context.Context, req ToolCallRequest) (ToolCallResult, error)
}

// Client is the single entrypoint every agent calls. The backend is guarded
// by a mutex and swappable at runtime (SwapBackend) — an admin changing the
// active provider/model takes effect immediately for every in-flight and
// future call, without restarting the server or re-wiring the many places
// that already hold a *Client (orchestrator, template analysis,
// recommendation ranking all keep the same pointer forever).
type Client struct {
	mu       sync.RWMutex
	backend  Backend
	provider string
}

// NewClient builds a Client against Anthropic Claude — the spec's default
// pipeline provider, and the one every recorded fixture test is written
// against. Kept as the zero-config constructor for backward compatibility.
func NewClient(apiKey, model, baseURL string, timeout time.Duration) *Client {
	return &Client{backend: newAnthropicBackend(apiKey, model, baseURL, timeout), provider: "anthropic"}
}

// NewGeminiClient builds a Client against Google Gemini's function-calling
// API instead of Anthropic's tool-use API — same ToolCallRequest/Result
// contract, different vendor underneath.
func NewGeminiClient(apiKey, model, baseURL string, timeout time.Duration) *Client {
	return &Client{backend: newGeminiBackend(apiKey, model, baseURL, timeout), provider: "gemini"}
}

// NewFromProvider selects the backend by name ("anthropic" or "gemini"), so
// the whole pipeline's LLM vendor is one config value (AGENT_PROVIDER).
func NewFromProvider(provider string, cfg BackendConfig) (*Client, error) {
	backend, resolvedProvider, err := BuildBackend(provider, cfg)
	if err != nil {
		return nil, err
	}
	return &Client{backend: backend, provider: resolvedProvider}, nil
}

// BuildBackend constructs a Backend for the given provider name without
// wrapping it in a Client — used both by NewFromProvider and by the admin
// agent-settings switch (which swaps a backend into an existing Client).
func BuildBackend(provider string, cfg BackendConfig) (backend Backend, resolvedProvider string, err error) {
	switch provider {
	case "", "anthropic":
		return newAnthropicBackend(cfg.AnthropicAPIKey, cfg.AnthropicModel, cfg.AnthropicBaseURL, cfg.Timeout), "anthropic", nil
	case "gemini":
		return newGeminiBackend(cfg.GeminiAPIKey, cfg.GeminiModel, cfg.GeminiBaseURL, cfg.Timeout), "gemini", nil
	default:
		return nil, "", fmt.Errorf("unknown agent provider %q (expected \"anthropic\" or \"gemini\")", provider)
	}
}

// BackendConfig carries both vendors' settings so NewFromProvider/BuildBackend
// can switch between them without the caller needing to know which fields matter.
type BackendConfig struct {
	AnthropicAPIKey  string
	AnthropicModel   string
	AnthropicBaseURL string
	GeminiAPIKey     string
	GeminiModel      string
	GeminiBaseURL    string
	Timeout          time.Duration
}

// WithModel returns a copy of cfg with the given provider's model field
// overridden — used to build a one-off BackendConfig for a specific
// provider+model pair (the admin settings switch) without touching the
// other provider's configured model.
func (cfg BackendConfig) WithModel(provider, model string) BackendConfig {
	out := cfg
	switch provider {
	case "anthropic":
		out.AnthropicModel = model
	case "gemini":
		out.GeminiModel = model
	}
	return out
}

// WithKey returns a copy of cfg with the given provider's API key overridden
// — used when an admin pastes a replacement key from the settings UI.
func (cfg BackendConfig) WithKey(provider, apiKey string) BackendConfig {
	out := cfg
	switch provider {
	case "anthropic":
		out.AnthropicAPIKey = apiKey
	case "gemini":
		out.GeminiAPIKey = apiKey
	}
	return out
}

func (c *Client) Enabled() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.backend.Enabled()
}

// ModelName returns the configured model, for recording alongside stage_outputs.
func (c *Client) ModelName() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.backend.ModelName()
}

// Provider returns the current backend's vendor name ("anthropic" | "gemini").
func (c *Client) Provider() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.provider
}

// SwapBackend atomically replaces the active backend — every subsequent
// CallTool (from any goroutine) uses it immediately. In-flight calls that
// already captured the old backend finish against it, uninterrupted.
func (c *Client) SwapBackend(provider string, backend Backend) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.backend = backend
	c.provider = provider
}

func (c *Client) CallTool(ctx context.Context, req ToolCallRequest) (ToolCallResult, error) {
	c.mu.RLock()
	backend := c.backend
	c.mu.RUnlock()
	return backend.CallTool(ctx, req)
}

// sharedHTTPClient is reused by both backends — no per-call allocation.
func newHTTPClient() *http.Client { return &http.Client{} }

// sharedRetryConfig is the retry/backoff policy both backends use: 3
// attempts, exponential 500ms→4s with jitter, retrying network errors,
// 429/5xx, and malformed/empty tool-call responses (spec §17).
func sharedRetryConfig() retry.Config { return retry.DefaultConfig() }
