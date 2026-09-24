// Package ideogram implements imagegen.ImageProvider against Ideogram v2 API.
package ideogram

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"armss-poster-studio/internal/imagegen"
	"armss-poster-studio/internal/retry"
)

var ErrUnavailable = &imagegen.ProviderError{
	Provider:  "ideogram",
	Message:   "Ideogram image generation is disabled (no API key configured)",
	Retryable: false,
}

type Client struct {
	apiKey     string
	baseURL    string
	model      string
	httpClient *http.Client
	timeout    time.Duration
	retryCfg   retry.Config
	costCents  int
}

func NewClient(apiKey, baseURL, model string, timeout time.Duration, costCents int) *Client {
	if baseURL == "" {
		baseURL = "https://api.ideogram.ai"
	}
	if model == "" {
		model = "V_2"
	}
	if costCents <= 0 {
		costCents = 8
	}
	return &Client{
		apiKey:     apiKey,
		baseURL:    strings.TrimRight(baseURL, "/"),
		model:      model,
		httpClient: &http.Client{},
		timeout:    timeout,
		retryCfg:   retry.DefaultConfig(),
		costCents:  costCents,
	}
}

func (c *Client) Name() string                { return "ideogram" }
func (c *Client) SupportsTextRendering() bool { return true }
func (c *Client) Enabled() bool               { return c.apiKey != "" }
func (c *Client) CostCentsPerImage() int      { return c.costCents }

type imageRequestPayload struct {
	Prompt            string `json:"prompt"`
	AspectRatio       string `json:"aspect_ratio,omitempty"`
	Model             string `json:"model,omitempty"`
	MagicPromptOption string `json:"magic_prompt_option,omitempty"`
	NegativePrompt    string `json:"negative_prompt,omitempty"`
	StyleType         string `json:"style_type,omitempty"`
}

type ideogramRequestBody struct {
	ImageRequest imageRequestPayload `json:"image_request"`
}

type ideogramResponseBody struct {
	Data []struct {
		URL         string `json:"url"`
		Prompt      string `json:"prompt"`
		Resolution  string `json:"resolution"`
		IsImageSafe bool   `json:"is_image_safe"`
	} `json:"data"`
	Detail  string `json:"detail,omitempty"`
	Message string `json:"message,omitempty"`
}

func (c *Client) Generate(ctx context.Context, req imagegen.GenerateRequest) ([]imagegen.GeneratedImage, error) {
	if !c.Enabled() {
		return nil, ErrUnavailable
	}

	body := ideogramRequestBody{
		ImageRequest: imageRequestPayload{
			Prompt:            req.Prompt,
			AspectRatio:       mapAspectRatio(req.AspectRatio),
			Model:             c.model,
			MagicPromptOption: "AUTO",
			NegativePrompt:    req.NegativePrompt,
			StyleType:         "DESIGN",
		},
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal ideogram request: %w", err)
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
		var perr *imagegen.ProviderError
		if errors.As(err, &perr) {
			return nil, perr
		}
		return nil, &imagegen.ProviderError{Provider: "ideogram", Message: err.Error(), Retryable: false}
	}
	return images, nil
}

func (c *Client) doGenerate(ctx context.Context, payload []byte) ([]imagegen.GeneratedImage, bool, error) {
	endpoint := c.baseURL + "/generate"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, false, fmt.Errorf("build ideogram request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Api-Key", c.apiKey)
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey) // supported fallback for some gateways

	httpRes, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, true, &imagegen.ProviderError{Provider: "ideogram", Message: "failed to connect to Ideogram API", Retryable: true}
	}
	defer httpRes.Body.Close()

	respBody, err := io.ReadAll(httpRes.Body)
	if err != nil {
		return nil, true, &imagegen.ProviderError{Provider: "ideogram", Message: "could not read Ideogram response", Retryable: true}
	}

	if httpRes.StatusCode == http.StatusTooManyRequests || httpRes.StatusCode >= 500 {
		return nil, true, &imagegen.ProviderError{
			Provider:  "ideogram",
			Message:   fmt.Sprintf("Ideogram returned %d: %s", httpRes.StatusCode, string(respBody)),
			Retryable: true,
		}
	}
	if httpRes.StatusCode >= 400 {
		return nil, false, &imagegen.ProviderError{
			Provider:  "ideogram",
			Message:   fmt.Sprintf("Ideogram rejected request (%d): %s", httpRes.StatusCode, string(respBody)),
			Retryable: false,
		}
	}

	var parsed ideogramResponseBody
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return nil, true, &imagegen.ProviderError{Provider: "ideogram", Message: "Ideogram returned invalid JSON", Retryable: true}
	}

	if len(parsed.Data) == 0 {
		msg := "Ideogram returned no images"
		if parsed.Detail != "" {
			msg = parsed.Detail
		} else if parsed.Message != "" {
			msg = parsed.Message
		}
		return nil, false, &imagegen.ProviderError{Provider: "ideogram", Message: msg, Retryable: false}
	}

	imgData, err := c.fetchImageURL(ctx, parsed.Data[0].URL)
	if err != nil {
		return nil, true, err
	}

	return []imagegen.GeneratedImage{{
		Data:        imgData,
		ContentType: "image/png",
	}}, false, nil
}

func (c *Client) fetchImageURL(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build download request: %w", err)
	}
	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, &imagegen.ProviderError{Provider: "ideogram", Message: "failed to download Ideogram image", Retryable: true}
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, &imagegen.ProviderError{Provider: "ideogram", Message: fmt.Sprintf("download returned %d", res.StatusCode), Retryable: true}
	}
	return io.ReadAll(res.Body)
}

func (c *Client) Edit(ctx context.Context, req imagegen.EditRequest) ([]imagegen.GeneratedImage, error) {
	return nil, &imagegen.ProviderError{Provider: "ideogram", Message: "Ideogram edit is not supported", Retryable: false}
}

func mapAspectRatio(ratio string) string {
	switch ratio {
	case "1:1":
		return "ASPECT_1_1"
	case "9:16":
		return "ASPECT_9_16"
	case "16:9":
		return "ASPECT_16_9"
	case "4:5", "3:4":
		return "ASPECT_3_4"
	case "4:3":
		return "ASPECT_4_3"
	default:
		return "ASPECT_10_16"
	}
}
