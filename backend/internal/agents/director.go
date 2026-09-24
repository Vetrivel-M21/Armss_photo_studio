package agents

import (
	"context"
	"encoding/json"
	"fmt"
)

const directorSystemPrompt = `You are the Creative Director for ARMSS Poster Studio. Given a
creative brief and (optionally) a house style profile, produce exactly three genuinely
different big ideas for the poster — different concepts, not three colourways of the same
idea. Each concept needs its own visual metaphor and mood, while still fitting the brief's
occasion and the house style's tone.`

func directorToolSchema() map[string]any {
	concept := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name":           map[string]any{"type": "string"},
			"bigIdea":        map[string]any{"type": "string"},
			"visualMetaphor": map[string]any{"type": "string"},
			"mood":           map[string]any{"type": "string"},
			"palette":        map[string]any{"type": "string"},
		},
		"required": []string{"name", "bigIdea", "visualMetaphor", "mood", "palette"},
	}
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"concepts": map[string]any{
				"type":     "array",
				"items":    concept,
				"minItems": 3,
				"maxItems": 3,
			},
		},
		"required": []string{"concepts"},
	}
}

type conceptsInput struct {
	Concepts []Concept `json:"concepts"`
}

// RunCreativeDirector is agent 3: Brief + StyleProfile → exactly three Concepts.
func (c *Client) RunCreativeDirector(ctx context.Context, brief Brief, style *StyleProfile) ([]Concept, Usage, error) {
	briefJSON, _ := json.MarshalIndent(brief, "", "  ")
	prompt := fmt.Sprintf("Creative brief:\n%s\n\n%s", string(briefJSON), styleContext(style))

	result, err := c.CallTool(ctx, ToolCallRequest{
		System:   directorSystemPrompt,
		UserText: prompt,
		Tool: ToolSpec{
			Name:        "emit_concepts",
			Description: "Emit exactly three distinct creative concepts for this poster.",
			InputSchema: directorToolSchema(),
		},
		MaxTokens: MaxTokensDirector,
	})
	if err != nil {
		return nil, Usage{}, err
	}

	var parsed conceptsInput
	if err := json.Unmarshal(result.Raw, &parsed); err != nil {
		return nil, Usage{}, fmt.Errorf("unmarshal concepts: %w", err)
	}
	return parsed.Concepts, Usage{TokensIn: result.TokensIn, TokensOut: result.TokensOut}, nil
}
