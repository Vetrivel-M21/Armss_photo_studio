package agents

import (
	"context"
	"encoding/json"
	"fmt"
)

const copySystemPrompt = `You are the Copywriter for ARMSS Poster Studio. Given a creative brief
and a chosen concept, write the poster's copy: headline, subhead, and call to action. Keep it
short and in active voice. Never use exclamation marks. Also provide a few value words (single
words or short phrases capturing the emotional core, e.g. LIGHT / PROSPERITY / JOY) and a
couple of alternate headlines.`

func copyToolSchema() map[string]any {
	strArr := map[string]any{"type": "array", "items": map[string]any{"type": "string"}}
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"headline":   map[string]any{"type": "string"},
			"subhead":    map[string]any{"type": "string"},
			"cta":        map[string]any{"type": "string"},
			"valueWords": strArr,
			"alternates": strArr,
		},
		"required": []string{"headline", "subhead", "cta", "valueWords", "alternates"},
	}
}

// RunCopywriter is agent 4: Brief + chosen Concept → Copy.
func (c *Client) RunCopywriter(ctx context.Context, brief Brief, concept Concept) (Copy, Usage, error) {
	briefJSON, _ := json.MarshalIndent(brief, "", "  ")
	conceptJSON, _ := json.MarshalIndent(concept, "", "  ")
	prompt := fmt.Sprintf("Creative brief:\n%s\n\nChosen concept:\n%s", string(briefJSON), string(conceptJSON))

	result, err := c.CallTool(ctx, ToolCallRequest{
		System:   copySystemPrompt,
		UserText: prompt,
		Tool: ToolSpec{
			Name:        "emit_copy",
			Description: "Emit the poster's copy for the chosen concept.",
			InputSchema: copyToolSchema(),
		},
		MaxTokens: MaxTokensCopy,
	})
	if err != nil {
		return Copy{}, Usage{}, err
	}

	var out Copy
	if err := json.Unmarshal(result.Raw, &out); err != nil {
		return Copy{}, Usage{}, fmt.Errorf("unmarshal copy: %w", err)
	}
	return out, Usage{TokensIn: result.TokensIn, TokensOut: result.TokensOut}, nil
}
