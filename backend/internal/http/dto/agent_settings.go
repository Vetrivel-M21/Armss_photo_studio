package dto

type AgentSettingsResponse struct {
	Provider            string `json:"provider"`
	Model               string `json:"model"`
	AnthropicConfigured bool   `json:"anthropicConfigured"`
	GeminiConfigured    bool   `json:"geminiConfigured"`
	OpenAIConfigured    bool   `json:"openaiConfigured"`
	FluxConfigured      bool   `json:"fluxConfigured"`
	IdeogramConfigured  bool   `json:"ideogramConfigured"`
	ImageProvider       string `json:"imageProvider"`
}

type AgentSettingsRequest struct {
	Provider      string `json:"provider,omitempty"`
	Model         string `json:"model,omitempty"`
	ImageProvider string `json:"imageProvider,omitempty"`
}

// SetAPIKeyRequest is intentionally write-only: the key is never echoed back
// in any response, only whether a key is now configured (see
// AgentSettingsResponse.*Configured).
type SetAPIKeyRequest struct {
	Provider string `json:"provider"` // "anthropic" | "gemini" | "openai" | "flux" | "ideogram"
	APIKey   string `json:"apiKey"`
}
