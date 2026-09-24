// Package openai implements imagegen.ImageProvider against OpenAI's
// gpt-image-1, the only vendor implemented in this build (spec §7 asks for
// at least one; the ImageProvider interface stays vendor-agnostic so more
// can be added later without touching the pipeline).
package openai

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"armss-poster-studio/internal/imagegen"
	"armss-poster-studio/internal/retry"
)

type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
	timeout    time.Duration
	retryCfg   retry.Config
	costCents  int
}

func NewClient(apiKey, baseURL string, timeout time.Duration, costCents int) *Client {
	return &Client{
		apiKey:     apiKey,
		baseURL:    baseURL,
		httpClient: &http.Client{},
		timeout:    timeout,
		retryCfg:   retry.DefaultConfig(),
		costCents:  costCents,
	}
}

func (c *Client) Name() string                { return "openai" }
func (c *Client) SupportsTextRendering() bool { return true } // gpt-image-1 renders text directly — Generate receives FullPosterPrompt
func (c *Client) Enabled() bool               { return c.apiKey != "" }
func (c *Client) CostCentsPerImage() int      { return c.costCents }

type generateRequestBody struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Size   string `json:"size"`
	N      int    `json:"n"`
}

type generateResponseBody struct {
	Data []struct {
		B64JSON string `json:"b64_json"`
	} `json:"data"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (c *Client) Generate(ctx context.Context, req imagegen.GenerateRequest) ([]imagegen.GeneratedImage, error) {
	if !c.Enabled() {
		return nil, ErrUnavailable
	}

	prompt := req.Prompt
	if req.NegativePrompt != "" {
		prompt = fmt.Sprintf("%s\n\nAvoid: %s", prompt, req.NegativePrompt)
	}
	n := req.N
	if n <= 0 {
		n = 1
	}

	body := generateRequestBody{Model: "gpt-image-1", Prompt: prompt, Size: mapAspectRatio(req.AspectRatio), N: n}
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal openai request: %w", err)
	}

	var images []imagegen.GeneratedImage
	err = retry.Do(ctx, c.retryCfg, func(attempt int) (bool, error) {
		callCtx, cancel := context.WithTimeout(ctx, c.timeout)
		defer cancel()

		out, retryable, callErr := c.doGenerate(callCtx, payload)
		if callErr != nil {
			return retryable, callErr
		}
		images = out
		return false, nil
	})
	if err != nil {
		if perr, ok := err.(*imagegen.ProviderError); ok {
			return nil, perr
		}
		return nil, providerErr(err.Error(), false)
	}
	return images, nil
}

func (c *Client) doGenerate(ctx context.Context, payload []byte) ([]imagegen.GeneratedImage, bool, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/images/generations", bytes.NewReader(payload))
	if err != nil {
		return nil, false, fmt.Errorf("build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	httpRes, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, true, providerErr("the request to OpenAI failed to connect", true)
	}
	defer httpRes.Body.Close()

	respBody, err := io.ReadAll(httpRes.Body)
	if err != nil {
		return nil, true, providerErr("could not read OpenAI's response", true)
	}

	if httpRes.StatusCode == http.StatusTooManyRequests || httpRes.StatusCode >= 500 {
		return nil, true, providerErr(fmt.Sprintf("OpenAI returned %d", httpRes.StatusCode), true)
	}
	if httpRes.StatusCode >= 400 {
		return nil, false, providerErr(fmt.Sprintf("OpenAI rejected the request (%d)", httpRes.StatusCode), false)
	}

	var parsed generateResponseBody
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return nil, true, providerErr("OpenAI returned a malformed response", true)
	}
	if parsed.Error != nil {
		return nil, true, providerErr(parsed.Error.Message, true)
	}
	if len(parsed.Data) == 0 {
		return nil, true, providerErr("OpenAI returned no image data", true)
	}

	images := make([]imagegen.GeneratedImage, 0, len(parsed.Data))
	for _, d := range parsed.Data {
		data, err := base64.StdEncoding.DecodeString(d.B64JSON)
		if err != nil {
			return nil, true, providerErr("OpenAI returned an unreadable image", true)
		}
		images = append(images, imagegen.GeneratedImage{Data: data, ContentType: "image/png"})
	}
	return images, false, nil
}

// Edit is part of the ImageProvider interface but isn't used by anything in
// this build (no partial-regeneration feature is in scope) — implemented as a
// clear, honest error rather than unused multipart request-building code.
func (c *Client) Edit(ctx context.Context, req imagegen.EditRequest) ([]imagegen.GeneratedImage, error) {
	return nil, providerErr("editing an existing image isn't supported yet", false)
}
