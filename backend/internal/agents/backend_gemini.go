package agents

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"armss-poster-studio/internal/retry"
)

// geminiBackend implements Backend against Google Gemini's generateContent
// API, using function calling in forced ("ANY") mode as Gemini's equivalent
// of Anthropic's tool_choice — same "never prompt-and-parse" guarantee
// (spec §17), different vendor underneath. Selected via AGENT_PROVIDER=gemini.
type geminiBackend struct {
	apiKey     string
	model      string
	baseURL    string
	httpClient *http.Client
	timeout    time.Duration
	retryCfg   retry.Config
}

func newGeminiBackend(apiKey, model, baseURL string, timeout time.Duration) *geminiBackend {
	return &geminiBackend{
		apiKey:     apiKey,
		model:      model,
		baseURL:    baseURL,
		httpClient: newHTTPClient(),
		timeout:    timeout,
		retryCfg:   sharedRetryConfig(),
	}
}

func (b *geminiBackend) Enabled() bool     { return b.apiKey != "" }
func (b *geminiBackend) ModelName() string { return b.model }

type geminiPart struct {
	Text         string              `json:"text,omitempty"`
	InlineData   *geminiInlineData   `json:"inline_data,omitempty"`
	FunctionCall *geminiFunctionCall `json:"functionCall,omitempty"`
}

type geminiInlineData struct {
	MimeType string `json:"mime_type"`
	Data     string `json:"data"`
}

type geminiFunctionCall struct {
	Name string          `json:"name"`
	Args json.RawMessage `json:"args"`
}

type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}

type geminiFunctionDeclaration struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

type geminiTool struct {
	FunctionDeclarations []geminiFunctionDeclaration `json:"function_declarations"`
}

type geminiToolConfig struct {
	FunctionCallingConfig geminiFunctionCallingConfig `json:"function_calling_config"`
}

type geminiFunctionCallingConfig struct {
	Mode                 string   `json:"mode"`
	AllowedFunctionNames []string `json:"allowed_function_names"`
}

type geminiGenerationConfig struct {
	MaxOutputTokens int `json:"maxOutputTokens"`
}

type geminiRequest struct {
	SystemInstruction *geminiContent         `json:"system_instruction,omitempty"`
	Contents          []geminiContent        `json:"contents"`
	Tools             []geminiTool           `json:"tools"`
	ToolConfig        geminiToolConfig       `json:"tool_config"`
	GenerationConfig  geminiGenerationConfig `json:"generationConfig"`
}

type geminiCandidate struct {
	Content      geminiContent `json:"content"`
	FinishReason string        `json:"finishReason"`
}

type geminiUsageMetadata struct {
	PromptTokenCount     int `json:"promptTokenCount"`
	CandidatesTokenCount int `json:"candidatesTokenCount"`
}

type geminiResponse struct {
	Candidates    []geminiCandidate    `json:"candidates"`
	UsageMetadata *geminiUsageMetadata `json:"usageMetadata"`
	Error         *geminiAPIError      `json:"error,omitempty"`
}

type geminiAPIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

func (b *geminiBackend) CallTool(ctx context.Context, req ToolCallRequest) (ToolCallResult, error) {
	if !b.Enabled() {
		return ToolCallResult{}, ErrAgentsUnavailable
	}

	parts := []geminiPart{{Text: req.UserText}}
	for _, img := range req.Images {
		parts = append(parts, geminiPart{
			InlineData: &geminiInlineData{MimeType: img.MediaType, Data: base64.StdEncoding.EncodeToString(img.Data)},
		})
	}

	body := geminiRequest{
		SystemInstruction: &geminiContent{Parts: []geminiPart{{Text: req.System}}},
		Contents:          []geminiContent{{Role: "user", Parts: parts}},
		Tools: []geminiTool{{FunctionDeclarations: []geminiFunctionDeclaration{{
			Name:        req.Tool.Name,
			Description: req.Tool.Description,
			Parameters:  convertSchemaForGemini(req.Tool.InputSchema),
		}}}},
		ToolConfig: geminiToolConfig{FunctionCallingConfig: geminiFunctionCallingConfig{
			Mode:                 "ANY",
			AllowedFunctionNames: []string{req.Tool.Name},
		}},
		GenerationConfig: geminiGenerationConfig{MaxOutputTokens: req.MaxTokens},
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

func (b *geminiBackend) doCall(ctx context.Context, payload []byte) (ToolCallResult, bool, error) {
	endpoint := fmt.Sprintf("%s/v1beta/models/%s:generateContent?key=%s", b.baseURL, b.model, url.QueryEscape(b.apiKey))
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return ToolCallResult{}, false, fmt.Errorf("build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	httpRes, err := b.httpClient.Do(httpReq)
	if err != nil {
		return ToolCallResult{}, true, fmt.Errorf("call gemini: %w", err)
	}
	defer httpRes.Body.Close()

	respBody, err := io.ReadAll(httpRes.Body)
	if err != nil {
		return ToolCallResult{}, true, fmt.Errorf("read response: %w", err)
	}

	if httpRes.StatusCode == http.StatusTooManyRequests || httpRes.StatusCode >= 500 {
		return ToolCallResult{}, true, fmt.Errorf("gemini returned %d", httpRes.StatusCode)
	}
	if httpRes.StatusCode >= 400 {
		return ToolCallResult{}, false, fmt.Errorf("gemini returned %d: %s", httpRes.StatusCode, string(respBody))
	}

	var parsed geminiResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		// Malformed body from a 200 response — retry rather than crash.
		return ToolCallResult{}, true, fmt.Errorf("unmarshal response: %w", err)
	}
	if parsed.Error != nil {
		return ToolCallResult{}, true, fmt.Errorf("gemini error: %s", parsed.Error.Message)
	}

	var tokensIn, tokensOut int
	if parsed.UsageMetadata != nil {
		tokensIn = parsed.UsageMetadata.PromptTokenCount
		tokensOut = parsed.UsageMetadata.CandidatesTokenCount
	}

	for _, cand := range parsed.Candidates {
		for _, part := range cand.Content.Parts {
			if part.FunctionCall != nil && len(part.FunctionCall.Args) > 0 {
				return ToolCallResult{
					Raw:        part.FunctionCall.Args,
					TokensIn:   tokensIn,
					TokensOut:  tokensOut,
					StopReason: cand.FinishReason,
				}, false, nil
			}
		}
	}
	// No functionCall part, or it was empty (truncated) — retryable.
	return ToolCallResult{}, true, fmt.Errorf("no functionCall content in gemini response")
}

// convertSchemaForGemini strips JSON-Schema keywords Gemini's Schema type
// doesn't recognise (minItems/maxItems) and uppercases "type" values to match
// Gemini's Schema.Type enum (STRING/NUMBER/INTEGER/BOOLEAN/ARRAY/OBJECT),
// recursing into properties/items. Our tool schemas are otherwise plain
// enough (type/properties/items/required/description) to translate directly.
func convertSchemaForGemini(schema map[string]any) map[string]any {
	if schema == nil {
		return nil
	}
	out := map[string]any{}
	if t, ok := schema["type"].(string); ok {
		out["type"] = strings.ToUpper(t)
	}
	if desc, ok := schema["description"]; ok {
		out["description"] = desc
	}
	if required, ok := schema["required"]; ok {
		out["required"] = required
	}
	if props, ok := schema["properties"].(map[string]any); ok {
		converted := make(map[string]any, len(props))
		for k, v := range props {
			if sub, ok := v.(map[string]any); ok {
				converted[k] = convertSchemaForGemini(sub)
			}
		}
		out["properties"] = converted
	}
	if items, ok := schema["items"].(map[string]any); ok {
		out["items"] = convertSchemaForGemini(items)
	}
	return out
}
