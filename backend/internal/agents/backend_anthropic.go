package agents

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"armss-poster-studio/internal/retry"
)

const anthropicVersion = "2023-06-01"

// anthropicBackend implements Backend against Claude's Messages API, using
// tool_choice to force the model to call exactly one tool (spec §17: enforce
// output shape with tool-use schemas, never prompt-and-parse).
type anthropicBackend struct {
	apiKey     string
	model      string
	baseURL    string
	httpClient *http.Client
	timeout    time.Duration
	retryCfg   retry.Config
}

func newAnthropicBackend(apiKey, model, baseURL string, timeout time.Duration) *anthropicBackend {
	return &anthropicBackend{
		apiKey:     apiKey,
		model:      model,
		baseURL:    baseURL,
		httpClient: newHTTPClient(),
		timeout:    timeout,
		retryCfg:   sharedRetryConfig(),
	}
}

func (b *anthropicBackend) Enabled() bool     { return b.apiKey != "" }
func (b *anthropicBackend) ModelName() string { return b.model }

type anthropicContentBlock struct {
	Type   string              `json:"type"`
	Text   string              `json:"text,omitempty"`
	Source *anthropicImgSource `json:"source,omitempty"`
	ID     string              `json:"id,omitempty"`
	Name   string              `json:"name,omitempty"`
	Input  json.RawMessage     `json:"input,omitempty"`
}

type anthropicImgSource struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type"`
	Data      string `json:"data"`
}

type anthropicMessage struct {
	Role    string                  `json:"role"`
	Content []anthropicContentBlock `json:"content"`
}

type anthropicToolChoice struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

type anthropicToolDef struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"input_schema"`
}

type anthropicRequest struct {
	Model      string              `json:"model"`
	MaxTokens  int                 `json:"max_tokens"`
	System     string              `json:"system,omitempty"`
	Messages   []anthropicMessage  `json:"messages"`
	Tools      []anthropicToolDef  `json:"tools"`
	ToolChoice anthropicToolChoice `json:"tool_choice"`
}

type anthropicUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

type anthropicResponse struct {
	Content    []anthropicContentBlock `json:"content"`
	StopReason string                  `json:"stop_reason"`
	Usage      anthropicUsage          `json:"usage"`
	Error      *anthropicAPIError      `json:"error,omitempty"`
}

type anthropicAPIError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

func (b *anthropicBackend) CallTool(ctx context.Context, req ToolCallRequest) (ToolCallResult, error) {
	if !b.Enabled() {
		return ToolCallResult{}, ErrAgentsUnavailable
	}

	blocks := []anthropicContentBlock{{Type: "text", Text: req.UserText}}
	for _, img := range req.Images {
		blocks = append(blocks, anthropicContentBlock{
			Type: "image",
			Source: &anthropicImgSource{
				Type:      "base64",
				MediaType: img.MediaType,
				Data:      base64.StdEncoding.EncodeToString(img.Data),
			},
		})
	}

	body := anthropicRequest{
		Model:     b.model,
		MaxTokens: req.MaxTokens,
		System:    req.System,
		Messages:  []anthropicMessage{{Role: "user", Content: blocks}},
		Tools: []anthropicToolDef{{
			Name:        req.Tool.Name,
			Description: req.Tool.Description,
			InputSchema: req.Tool.InputSchema,
		}},
		ToolChoice: anthropicToolChoice{Type: "tool", Name: req.Tool.Name},
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return ToolCallResult{}, fmt.Errorf("marshal request: %w", err)
	}

	var result ToolCallResult
	err = retry.Do(ctx, b.retryCfg, func(attempt int) (bool, error) {
		callCtx, cancel := context.WithTimeout(ctx, b.timeout)
		defer cancel()

		res, retryable, callErr := b.doCall(callCtx, payload)
		if callErr != nil {
			return retryable, callErr
		}
		result = res
		return false, nil
	})
	if err != nil {
		return ToolCallResult{}, err
	}
	return result, nil
}

func (b *anthropicBackend) doCall(ctx context.Context, payload []byte) (ToolCallResult, bool, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, b.baseURL+"/v1/messages", bytes.NewReader(payload))
	if err != nil {
		return ToolCallResult{}, false, fmt.Errorf("build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", b.apiKey)
	httpReq.Header.Set("anthropic-version", anthropicVersion)

	httpRes, err := b.httpClient.Do(httpReq)
	if err != nil {
		return ToolCallResult{}, true, fmt.Errorf("call anthropic: %w", err)
	}
	defer httpRes.Body.Close()

	respBody, err := io.ReadAll(httpRes.Body)
	if err != nil {
		return ToolCallResult{}, true, fmt.Errorf("read response: %w", err)
	}

	if httpRes.StatusCode == http.StatusTooManyRequests || httpRes.StatusCode >= 500 {
		return ToolCallResult{}, true, fmt.Errorf("anthropic returned %d", httpRes.StatusCode)
	}
	if httpRes.StatusCode >= 400 {
		return ToolCallResult{}, false, fmt.Errorf("anthropic returned %d: %s", httpRes.StatusCode, string(respBody))
	}

	var parsed anthropicResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		// Malformed body from a 200 response — retry rather than crash.
		return ToolCallResult{}, true, fmt.Errorf("unmarshal response: %w", err)
	}
	if parsed.Error != nil {
		return ToolCallResult{}, true, fmt.Errorf("anthropic error: %s", parsed.Error.Message)
	}

	for _, block := range parsed.Content {
		if block.Type == "tool_use" && len(block.Input) > 0 {
			return ToolCallResult{
				Raw:        block.Input,
				TokensIn:   parsed.Usage.InputTokens,
				TokensOut:  parsed.Usage.OutputTokens,
				StopReason: parsed.StopReason,
			}, false, nil
		}
	}
	// No tool_use block, or it was empty (truncated) — retryable.
	return ToolCallResult{}, true, fmt.Errorf("no tool_use content in response (stop_reason=%s)", parsed.StopReason)
}
