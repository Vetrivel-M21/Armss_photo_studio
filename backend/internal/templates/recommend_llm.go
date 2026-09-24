package templates

import (
	"context"
	"encoding/json"
	"fmt"

	"armss-poster-studio/internal/agents"
)

const rankSystemPrompt = `You rank a small poster template library against a user's free-text
description of what they need. Pick the templates that best fit their occasion, mood, and
format, most relevant first. Give a one-line reason for each pick.`

func rankToolSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"rankings": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"templateId": map[string]any{"type": "string"},
						"reason":     map[string]any{"type": "string"},
					},
					"required": []string{"templateId", "reason"},
				},
			},
		},
		"required": []string{"rankings"},
	}
}

type rankingsInput struct {
	Rankings []struct {
		TemplateID string `json:"templateId"`
		Reason     string `json:"reason"`
	} `json:"rankings"`
}

// RankWithLLM is spec §8's "step 2": only called when the cheap tag/keyword
// overlap in Rank() returns fewer than `need` results. Ranks title/summary
// against the free text query with a single tool-use call.
func RankWithLLM(ctx context.Context, client *agents.Client, query string, candidates []Candidate, need int) ([]Recommendation, error) {
	if len(candidates) == 0 || need <= 0 {
		return nil, nil
	}

	type candidateView struct {
		ID    string `json:"id"`
		Title string `json:"title"`
	}
	views := make([]candidateView, 0, len(candidates))
	for _, c := range candidates {
		views = append(views, candidateView{ID: c.TemplateID, Title: c.Title})
	}
	candidatesJSON, err := json.Marshal(views)
	if err != nil {
		return nil, fmt.Errorf("marshal candidates: %w", err)
	}

	result, err := client.CallTool(ctx, agents.ToolCallRequest{
		System:   rankSystemPrompt,
		UserText: fmt.Sprintf("User's description:\n%s\n\nCandidate templates:\n%s\n\nReturn the top %d.", query, string(candidatesJSON), need),
		Tool: agents.ToolSpec{
			Name:        "emit_rankings",
			Description: "Emit the ranked template ids with a one-line reason each.",
			InputSchema: rankToolSchema(),
		},
		MaxTokens: 1500,
	})
	if err != nil {
		return nil, fmt.Errorf("rank templates with llm: %w", err)
	}

	var parsed rankingsInput
	if err := json.Unmarshal(result.Raw, &parsed); err != nil {
		return nil, fmt.Errorf("unmarshal rankings: %w", err)
	}

	out := make([]Recommendation, 0, len(parsed.Rankings))
	for i, r := range parsed.Rankings {
		out = append(out, Recommendation{TemplateID: r.TemplateID, Reason: r.Reason, Score: len(parsed.Rankings) - i})
	}
	return out, nil
}
