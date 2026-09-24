package agents

import (
	"context"
	"encoding/json"
	"fmt"
)

const styleSystemPrompt = `You are the Style Analyst for ARMSS Poster Studio, a poster design tool for
ARMSS Employment Solutions. Given one or more reference poster images, extract a structured
description of the house style: color palette, typography feel, layout patterns, imagery
style, finish (photoreal composite / 3D / painting / flat), tone, how the logo is typically
treated, and any furniture (recurring elements) that appears on every poster. Be specific and
concrete — this profile drives every future poster generated in this style. Never mention or
describe the ARMSS logo mark itself in a way that could be used to recreate it; only describe
where and how it is placed (logoTreatment).`

func styleToolSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"summary": map[string]any{"type": "string"},
			"palette": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"hex":  map[string]any{"type": "string"},
						"role": map[string]any{"type": "string"},
					},
					"required": []string{"hex", "role"},
				},
			},
			"typography":     map[string]any{"type": "string"},
			"layoutPatterns": map[string]any{"type": "string"},
			"imagery":        map[string]any{"type": "string"},
			"finish":         map[string]any{"type": "string"},
			"tone":           map[string]any{"type": "string"},
			"logoTreatment":  map[string]any{"type": "string"},
			"fixedElements":  map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			"consistency":    map[string]any{"type": "string"},
		},
		"required": []string{
			"summary", "palette", "typography", "layoutPatterns", "imagery",
			"finish", "tone", "logoTreatment", "fixedElements", "consistency",
		},
	}
}

// AnalyzeStyle is agent 1 (Style Analyst): a vision call that turns reference
// images into a StyleProfile. Cached per template — never re-run for library
// templates once analysed (spec §5 note).
func (c *Client) AnalyzeStyle(ctx context.Context, images []ImageInput) (StyleProfile, Usage, error) {
	if len(images) == 0 {
		return StyleProfile{}, Usage{}, fmt.Errorf("at least one reference image is required")
	}

	result, err := c.CallTool(ctx, ToolCallRequest{
		System:   styleSystemPrompt,
		UserText: "Analyse the attached reference poster image(s) and extract the house style profile.",
		Images:   images,
		Tool: ToolSpec{
			Name:        "emit_style_profile",
			Description: "Emit the structured style profile extracted from the reference image(s).",
			InputSchema: styleToolSchema(),
		},
		MaxTokens: MaxTokensStyle,
	})
	if err != nil {
		return StyleProfile{}, Usage{}, err
	}

	var profile StyleProfile
	if err := json.Unmarshal(result.Raw, &profile); err != nil {
		return StyleProfile{}, Usage{}, fmt.Errorf("unmarshal style profile: %w", err)
	}
	return profile, Usage{TokensIn: result.TokensIn, TokensOut: result.TokensOut}, nil
}
