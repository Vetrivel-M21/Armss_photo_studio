package agents

import (
	"context"
	"encoding/json"
	"fmt"
)

const copySystemPrompt = `You are the Copywriter for ARMSS Poster Studio. Given a creative brief
and a chosen concept, write the poster's typographic copy hierarchy.
Keep copy crisp, active, and dignified. Never use exclamation marks.
Structure the text into distinct hierarchy roles:
- "eyebrow": Short introductory uppercase banner (e.g. "CELEBRATING" or "ANNUAL GALA")
- "headline": Primary focal word or main event name (e.g. "DIWALI")
- "subhead": Supporting poetic or inspiring message (e.g. "Lighting Every Home. Brightening Every Future.")
- "cta": Festive greeting or call to action (e.g. "Wishing You Prosperity")
- "valueWords": 3-5 single uppercase words capturing the emotional and cultural pillars (e.g. ["LIGHT", "PROSPERITY", "JOY", "NEW BEGINNINGS"])
- "dateText": Exact formatted event date from the brief or occasion (e.g. "8th November 2026")
- "tagline": Corporate slogan or campaign tagline (e.g. "Think Different Think Money")
- "alternates": 2-3 alternate headline phrasing options.`

func copyToolSchema() map[string]any {
	strArr := map[string]any{"type": "array", "items": map[string]any{"type": "string"}}
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"eyebrow":    map[string]any{"type": "string"},
			"headline":   map[string]any{"type": "string"},
			"subhead":    map[string]any{"type": "string"},
			"cta":        map[string]any{"type": "string"},
			"valueWords": strArr,
			"dateText":   map[string]any{"type": "string"},
			"tagline":    map[string]any{"type": "string"},
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
