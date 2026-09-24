package agents

import (
	"context"
	"encoding/json"
	"fmt"
)

const intakeSystemPrompt = `You are the Intake agent for ARMSS Poster Studio. Turn a short, often
incomplete user description of an occasion into a structured creative brief. Fill gaps with
sensible assumptions rather than blocking — never leave a field empty because information is
missing. Record every assumption you made in "assumptions" and any open question worth asking
the user later in "questions", but always produce a complete, usable brief regardless.`

func intakeToolSchema() map[string]any {
	strArr := map[string]any{"type": "array", "items": map[string]any{"type": "string"}}
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"objective":   map[string]any{"type": "string"},
			"audience":    map[string]any{"type": "string"},
			"keyMessage":  map[string]any{"type": "string"},
			"occasion":    map[string]any{"type": "string"},
			"eventDate":   map[string]any{"type": "string"},
			"tone":        strArr,
			"format":      map[string]any{"type": "string"},
			"dimensions":  map[string]any{"type": "string"},
			"palette":     strArr,
			"mustInclude": strArr,
			"avoid":       strArr,
			"assumptions": strArr,
			"questions":   strArr,
		},
		"required": []string{
			"objective", "audience", "keyMessage", "occasion", "eventDate", "tone",
			"format", "dimensions", "palette", "mustInclude", "avoid", "assumptions", "questions",
		},
	}
}

// RunIntake is agent 2 (Intake): user free text + an optional StyleProfile → a Brief.
func (c *Client) RunIntake(ctx context.Context, userText string, style *StyleProfile) (Brief, Usage, error) {
	prompt := fmt.Sprintf("User's occasion description:\n%s\n\n%s", userText, styleContext(style))

	result, err := c.CallTool(ctx, ToolCallRequest{
		System:   intakeSystemPrompt,
		UserText: prompt,
		Tool: ToolSpec{
			Name:        "emit_brief",
			Description: "Emit the structured creative brief derived from the user's description.",
			InputSchema: intakeToolSchema(),
		},
		MaxTokens: MaxTokensIntake,
	})
	if err != nil {
		return Brief{}, Usage{}, err
	}

	var brief Brief
	if err := json.Unmarshal(result.Raw, &brief); err != nil {
		return Brief{}, Usage{}, fmt.Errorf("unmarshal brief: %w", err)
	}
	return brief, Usage{TokensIn: result.TokensIn, TokensOut: result.TokensOut}, nil
}

// styleContext renders an optional StyleProfile as prompt text shared by every
// downstream agent (Intake, Creative Director, Art Director all take it).
func styleContext(style *StyleProfile) string {
	if style == nil {
		return "No style reference is selected for this project."
	}
	b, _ := json.MarshalIndent(style, "", "  ")
	return fmt.Sprintf("Reference house style profile (match this):\n%s", string(b))
}
