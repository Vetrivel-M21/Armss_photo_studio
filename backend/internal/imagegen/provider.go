// Package imagegen puts the image-generation step behind a vendor-agnostic
// interface (spec §7): the pipeline's thinking (Anthropic) and the pixels
// (an image provider) come from different vendors.
package imagegen

import "context"

type GenerateRequest struct {
	Prompt         string
	NegativePrompt string
	AspectRatio    string
	N              int // always 1 in this build — never auto-generate more than one image per user action (spec §13)
}

type EditRequest struct {
	Prompt    string
	BaseImage []byte
	Mask      []byte
}

type GeneratedImage struct {
	Data        []byte
	ContentType string
	Width       int
	Height      int
}

type ImageProvider interface {
	Name() string
	SupportsTextRendering() bool
	Enabled() bool
	Generate(ctx context.Context, req GenerateRequest) ([]GeneratedImage, error)
	Edit(ctx context.Context, req EditRequest) ([]GeneratedImage, error) // optional; a provider without edit support returns ErrEditUnsupported
}

// ProviderError names the failing provider so the user-facing message never
// has to guess (spec §7: "a clear user-facing message naming the provider
// that failed. Never silently fall back to a different provider").
type ProviderError struct {
	Provider  string
	Message   string
	Retryable bool
}

func (e *ProviderError) Error() string {
	return e.Provider + ": " + e.Message
}
