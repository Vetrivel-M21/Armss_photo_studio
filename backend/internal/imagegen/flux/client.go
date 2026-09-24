// Package flux implements imagegen.ImageProvider against Flux.1 (Black Forest Labs)
// via Together.ai, Fal.ai, Replicate, or direct BFL API endpoints.
package flux

import (
	"bytes"
	"context"
	"encoding/base64"
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
	Provider:  "flux",
	Message:   "Flux image generation is disabled (no API key configured)",
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
		baseURL = "https://api.together.xyz"
	}
	if model == "" {
		model = "black-forest-labs/FLUX.1-schnell"
	}
	if costCents <= 0 {
		costCents = 3 // ~3 cents
	}
	if timeout <= 0 {
		timeout = 120 * time.Second
	}
	return &Client{
		apiKey:     apiKey,
		baseURL:    strings.TrimRight(baseURL, "/"),
		model:      model,
		httpClient: &http.Client{Timeout: timeout},
		timeout:    timeout,
		retryCfg:   retry.DefaultConfig(),
		costCents:  costCents,
	}
}

func (c *Client) Name() string                { return "flux" }
func (c *Client) SupportsTextRendering() bool { return true }
func (c *Client) Enabled() bool               { return c.apiKey != "" }
func (c *Client) CostCentsPerImage() int      { return c.costCents }

type togetherRequestBody struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	N      int    `json:"n"`
	Steps  int    `json:"steps,omitempty"`
}

type togetherResponseBody struct {
	Data []struct {
		B64JSON string `json:"b64_json"`
		URL     string `json:"url"`
	} `json:"data"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type falRequestBody struct {
	Prompt          string `json:"prompt"`
	ImageSize       string `json:"image_size,omitempty"`
	NumImages       int    `json:"num_images,omitempty"`
	EnableSafetyChk bool   `json:"enable_safety_checker,omitempty"`
}

type falResponseBody struct {
	Images []struct {
		URL         string `json:"url"`
		ContentType string `json:"content_type"`
	} `json:"images"`
	Detail string `json:"detail,omitempty"`
}

type replicateRequestBody struct {
	Input map[string]any `json:"input"`
}

type replicateResponseBody struct {
	ID     string   `json:"id"`
	Status string   `json:"status"`
	Output []string `json:"output"`
	Error  string   `json:"error,omitempty"`
	URLs   struct {
		Get string `json:"get"`
	} `json:"urls"`
}

type bflTaskResponse struct {
	ID     string `json:"id"`
	Status string `json:"status,omitempty"`
}

type bflResultResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Result struct {
		Sample string `json:"sample"`
	} `json:"result"`
}

func (c *Client) Generate(ctx context.Context, req imagegen.GenerateRequest) ([]imagegen.GeneratedImage, error) {
	if !c.Enabled() {
		return nil, ErrUnavailable
	}

	prompt := req.Prompt
	if req.NegativePrompt != "" {
		prompt = fmt.Sprintf("%s\n\nNegative prompt: %s", prompt, req.NegativePrompt)
	}

	var images []imagegen.GeneratedImage
	err := retry.Do(ctx, c.retryCfg, func(attempt int) (bool, error) {
		callCtx, cancel := context.WithTimeout(ctx, c.timeout)
		defer cancel()

		var out []imagegen.GeneratedImage
		var retryable bool
		var callErr error

		if strings.Contains(c.baseURL, "fal.run") || strings.Contains(c.baseURL, "fal.ai") {
			out, retryable, callErr = c.doFalGenerate(callCtx, prompt, req.AspectRatio)
		} else if strings.Contains(c.baseURL, "replicate.com") || strings.HasPrefix(c.apiKey, "r8_") {
			out, retryable, callErr = c.doReplicateGenerate(callCtx, prompt, req.AspectRatio)
		} else if strings.Contains(c.baseURL, "bfl.ml") {
			out, retryable, callErr = c.doBFLGenerate(callCtx, prompt, req.AspectRatio)
		} else {
			out, retryable, callErr = c.doTogetherGenerate(callCtx, prompt, req.AspectRatio)
		}

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
		return nil, &imagegen.ProviderError{Provider: "flux", Message: err.Error(), Retryable: false}
	}
	return images, nil
}

func (c *Client) doTogetherGenerate(ctx context.Context, prompt, ratio string) ([]imagegen.GeneratedImage, bool, error) {
	w, h := mapAspectRatioToDimensions(ratio)
	body := togetherRequestBody{
		Model:  c.model,
		Prompt: prompt,
		Width:  w,
		Height: h,
		N:      1,
		Steps:  4,
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, false, fmt.Errorf("marshal flux request: %w", err)
	}

	endpoint := c.baseURL + "/v1/images/generations"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, false, fmt.Errorf("build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	httpRes, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, true, &imagegen.ProviderError{Provider: "flux", Message: "failed to connect to Together Flux endpoint", Retryable: true}
	}
	defer httpRes.Body.Close()

	respBody, err := io.ReadAll(httpRes.Body)
	if err != nil {
		return nil, true, &imagegen.ProviderError{Provider: "flux", Message: "could not read Flux response", Retryable: true}
	}

	if httpRes.StatusCode == http.StatusTooManyRequests || httpRes.StatusCode >= 500 {
		return nil, true, &imagegen.ProviderError{Provider: "flux", Message: fmt.Sprintf("Flux provider returned %d: %s", httpRes.StatusCode, string(respBody)), Retryable: true}
	}
	if httpRes.StatusCode >= 400 {
		return nil, false, &imagegen.ProviderError{Provider: "flux", Message: fmt.Sprintf("Flux provider rejected request (%d): %s", httpRes.StatusCode, string(respBody)), Retryable: false}
	}

	var parsed togetherResponseBody
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return nil, true, &imagegen.ProviderError{Provider: "flux", Message: "Flux returned invalid JSON response", Retryable: true}
	}
	if parsed.Error != nil && parsed.Error.Message != "" {
		return nil, false, &imagegen.ProviderError{Provider: "flux", Message: parsed.Error.Message, Retryable: false}
	}
	if len(parsed.Data) == 0 {
		return nil, true, &imagegen.ProviderError{Provider: "flux", Message: "Flux returned no image data", Retryable: true}
	}

	item := parsed.Data[0]
	if item.B64JSON != "" {
		data, err := base64.StdEncoding.DecodeString(item.B64JSON)
		if err != nil {
			return nil, false, &imagegen.ProviderError{Provider: "flux", Message: "Failed to decode base64 image", Retryable: false}
		}
		return []imagegen.GeneratedImage{{Data: data, ContentType: "image/png", Width: w, Height: h}}, false, nil
	}

	if item.URL != "" {
		data, err := c.fetchImageURL(ctx, item.URL)
		if err != nil {
			return nil, true, err
		}
		return []imagegen.GeneratedImage{{Data: data, ContentType: "image/png", Width: w, Height: h}}, false, nil
	}

	return nil, false, &imagegen.ProviderError{Provider: "flux", Message: "Flux response contained neither b64_json nor url", Retryable: false}
}

func (c *Client) doFalGenerate(ctx context.Context, prompt, ratio string) ([]imagegen.GeneratedImage, bool, error) {
	falSize := "portrait_4_3"
	if ratio == "1:1" {
		falSize = "square_hd"
	} else if ratio == "16:9" {
		falSize = "landscape_16_9"
	} else if ratio == "4:5" {
		falSize = "portrait_4_3"
	}

	body := falRequestBody{
		Prompt:          prompt,
		ImageSize:       falSize,
		NumImages:       1,
		EnableSafetyChk: true,
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, false, fmt.Errorf("marshal fal request: %w", err)
	}

	endpoint := c.baseURL
	if !strings.HasSuffix(endpoint, "/schnell") && !strings.HasSuffix(endpoint, "/dev") {
		endpoint = endpoint + "/schnell"
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, false, fmt.Errorf("build fal request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Key "+c.apiKey)

	httpRes, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, true, &imagegen.ProviderError{Provider: "flux", Message: "failed to connect to Fal Flux endpoint", Retryable: true}
	}
	defer httpRes.Body.Close()

	respBody, err := io.ReadAll(httpRes.Body)
	if err != nil {
		return nil, true, &imagegen.ProviderError{Provider: "flux", Message: "could not read Fal response", Retryable: true}
	}

	if httpRes.StatusCode >= 400 {
		return nil, httpRes.StatusCode >= 500, &imagegen.ProviderError{
			Provider:  "flux",
			Message:   fmt.Sprintf("Fal Flux error (%d): %s", httpRes.StatusCode, string(respBody)),
			Retryable: httpRes.StatusCode >= 500,
		}
	}

	var parsed falResponseBody
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return nil, true, &imagegen.ProviderError{Provider: "flux", Message: "Fal returned invalid JSON", Retryable: true}
	}
	if len(parsed.Images) == 0 {
		return nil, false, &imagegen.ProviderError{Provider: "flux", Message: "Fal returned no images", Retryable: false}
	}

	imgData, err := c.fetchImageURL(ctx, parsed.Images[0].URL)
	if err != nil {
		return nil, true, err
	}
	return []imagegen.GeneratedImage{{Data: imgData, ContentType: "image/png"}}, false, nil
}

func (c *Client) doReplicateGenerate(ctx context.Context, prompt, ratio string) ([]imagegen.GeneratedImage, bool, error) {
	aspectRatio := ratio
	if aspectRatio == "" {
		aspectRatio = "4:5"
	}
	input := map[string]any{
		"prompt":        prompt,
		"aspect_ratio":  aspectRatio,
		"output_format": "png",
	}

	body := replicateRequestBody{Input: input}
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, false, fmt.Errorf("marshal replicate request: %w", err)
	}

	endpoint := "https://api.replicate.com/v1/models/black-forest-labs/flux-schnell/predictions"
	if strings.Contains(c.baseURL, "replicate.com") {
		endpoint = c.baseURL
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, false, fmt.Errorf("build replicate request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("Prefer", "wait=60")

	httpRes, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, true, &imagegen.ProviderError{Provider: "flux", Message: "failed to connect to Replicate", Retryable: true}
	}
	defer httpRes.Body.Close()

	respBody, err := io.ReadAll(httpRes.Body)
	if err != nil {
		return nil, true, &imagegen.ProviderError{Provider: "flux", Message: "could not read Replicate response", Retryable: true}
	}

	if httpRes.StatusCode >= 400 {
		return nil, httpRes.StatusCode >= 500, &imagegen.ProviderError{
			Provider:  "flux",
			Message:   fmt.Sprintf("Replicate error (%d): %s", httpRes.StatusCode, string(respBody)),
			Retryable: httpRes.StatusCode >= 500,
		}
	}

	var parsed replicateResponseBody
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return nil, true, &imagegen.ProviderError{Provider: "flux", Message: "Replicate returned invalid JSON", Retryable: true}
	}

	// If finished synchronously with output
	if parsed.Status == "succeeded" && len(parsed.Output) > 0 {
		imgData, err := c.fetchImageURL(ctx, parsed.Output[0])
		if err != nil {
			return nil, true, err
		}
		return []imagegen.GeneratedImage{{Data: imgData, ContentType: "image/png"}}, false, nil
	}

	// Poll if needed
	if parsed.URLs.Get != "" {
		return c.pollReplicate(ctx, parsed.URLs.Get)
	}

	return nil, false, &imagegen.ProviderError{Provider: "flux", Message: "Replicate returned no image output", Retryable: false}
}

func (c *Client) pollReplicate(ctx context.Context, pollURL string) ([]imagegen.GeneratedImage, bool, error) {
	ticker := time.NewTicker(1500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, false, ctx.Err()
		case <-ticker.C:
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, pollURL, nil)
			if err != nil {
				return nil, false, err
			}
			req.Header.Set("Authorization", "Bearer "+c.apiKey)
			res, err := c.httpClient.Do(req)
			if err != nil {
				continue
			}
			body, _ := io.ReadAll(res.Body)
			res.Body.Close()

			var parsed replicateResponseBody
			if err := json.Unmarshal(body, &parsed); err != nil {
				continue
			}
			if parsed.Status == "succeeded" && len(parsed.Output) > 0 {
				imgData, err := c.fetchImageURL(ctx, parsed.Output[0])
				if err != nil {
					return nil, true, err
				}
				return []imagegen.GeneratedImage{{Data: imgData, ContentType: "image/png"}}, false, nil
			}
			if parsed.Status == "failed" || parsed.Status == "canceled" {
				return nil, false, &imagegen.ProviderError{Provider: "flux", Message: "Replicate generation " + parsed.Status, Retryable: false}
			}
		}
	}
}

func (c *Client) doBFLGenerate(ctx context.Context, prompt, ratio string) ([]imagegen.GeneratedImage, bool, error) {
	aspectRatio := ratio
	if aspectRatio == "" {
		aspectRatio = "4:5"
	}
	body := map[string]any{
		"prompt":       prompt,
		"aspect_ratio": aspectRatio,
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, false, err
	}

	endpoint := c.baseURL
	if !strings.Contains(endpoint, "/flux-") {
		endpoint = endpoint + "/v1/flux-pro-1.1"
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, false, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-key", c.apiKey)

	httpRes, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, true, &imagegen.ProviderError{Provider: "flux", Message: "failed to connect to BFL API", Retryable: true}
	}
	defer httpRes.Body.Close()

	respBody, err := io.ReadAll(httpRes.Body)
	if err != nil {
		return nil, true, err
	}

	if httpRes.StatusCode >= 400 {
		return nil, httpRes.StatusCode >= 500, &imagegen.ProviderError{
			Provider:  "flux",
			Message:   fmt.Sprintf("BFL error (%d): %s", httpRes.StatusCode, string(respBody)),
			Retryable: httpRes.StatusCode >= 500,
		}
	}

	var task bflTaskResponse
	if err := json.Unmarshal(respBody, &task); err != nil {
		return nil, true, err
	}
	if task.ID == "" {
		return nil, false, &imagegen.ProviderError{Provider: "flux", Message: "BFL returned empty task id", Retryable: false}
	}

	return c.pollBFL(ctx, task.ID)
}

func (c *Client) pollBFL(ctx context.Context, taskID string) ([]imagegen.GeneratedImage, bool, error) {
	pollURL := "https://api.bfl.ml/v1/get_result?id=" + taskID
	ticker := time.NewTicker(1500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, false, ctx.Err()
		case <-ticker.C:
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, pollURL, nil)
			if err != nil {
				return nil, false, err
			}
			req.Header.Set("x-key", c.apiKey)
			res, err := c.httpClient.Do(req)
			if err != nil {
				continue
			}
			body, _ := io.ReadAll(res.Body)
			res.Body.Close()

			var parsed bflResultResponse
			if err := json.Unmarshal(body, &parsed); err != nil {
				continue
			}
			if parsed.Status == "Ready" && parsed.Result.Sample != "" {
				imgData, err := c.fetchImageURL(ctx, parsed.Result.Sample)
				if err != nil {
					return nil, true, err
				}
				return []imagegen.GeneratedImage{{Data: imgData, ContentType: "image/png"}}, false, nil
			}
			if parsed.Status == "Error" || parsed.Status == "Task not found" {
				return nil, false, &imagegen.ProviderError{Provider: "flux", Message: "BFL generation error", Retryable: false}
			}
		}
	}
}

func (c *Client) fetchImageURL(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build image download request: %w", err)
	}
	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, &imagegen.ProviderError{Provider: "flux", Message: "failed to download generated image", Retryable: true}
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, &imagegen.ProviderError{Provider: "flux", Message: fmt.Sprintf("image download returned %d", res.StatusCode), Retryable: true}
	}
	return io.ReadAll(res.Body)
}

func (c *Client) Edit(ctx context.Context, req imagegen.EditRequest) ([]imagegen.GeneratedImage, error) {
	return nil, &imagegen.ProviderError{Provider: "flux", Message: "Flux partial edit is not supported", Retryable: false}
}

func mapAspectRatioToDimensions(ratio string) (int, int) {
	switch ratio {
	case "1:1":
		return 1024, 1024
	case "4:5":
		return 896, 1152
	case "3:4":
		return 768, 1024
	case "9:16":
		return 768, 1344
	case "16:9":
		return 1344, 768
	default:
		return 896, 1152
	}
}
