package fixtures

import (
	"context"
	"strings"
	"testing"

	"armss-poster-studio/internal/agents"
)

func TestRunCopywriter(t *testing.T) {
	client := newFixtureServer(t, loadFixture(t, "copy_response.json"))

	result, usage, err := client.RunCopywriter(context.Background(),
		agents.Brief{Occasion: "Diwali 2026"},
		agents.Concept{Name: "Lamps of Opportunity"})
	if err != nil {
		t.Fatalf("RunCopywriter returned error: %v", err)
	}
	if result.Headline == "" {
		t.Error("expected non-empty headline")
	}
	if strings.Contains(result.Headline, "!") || strings.Contains(result.Subhead, "!") || strings.Contains(result.CTA, "!") {
		t.Error("copy must never contain exclamation marks")
	}
	if len(result.ValueWords) == 0 {
		t.Error("expected at least one value word")
	}
	if usage.TokensIn != 290 || usage.TokensOut != 140 {
		t.Errorf("unexpected usage: %+v", usage)
	}
}
