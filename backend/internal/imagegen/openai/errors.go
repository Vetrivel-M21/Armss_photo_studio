package openai

import (
	"errors"

	"armss-poster-studio/internal/imagegen"
)

// ErrUnavailable is returned when no OpenAI key is configured. Handlers must
// map this to a 503 naming the provider, never a crash (mirrors
// agents.ErrAgentsUnavailable's contract).
var ErrUnavailable = errors.New("openai api key not configured")

func providerErr(message string, retryable bool) *imagegen.ProviderError {
	return &imagegen.ProviderError{Provider: "openai", Message: message, Retryable: retryable}
}
