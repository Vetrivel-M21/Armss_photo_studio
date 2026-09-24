package fixtures

import (
	"context"
	"testing"

	"armss-poster-studio/internal/agents"
)

func TestRunCreativeDirector(t *testing.T) {
	client := newFixtureServer(t, loadFixture(t, "director_response.json"))

	concepts, usage, err := client.RunCreativeDirector(context.Background(), agents.Brief{Occasion: "Diwali 2026"}, nil)
	if err != nil {
		t.Fatalf("RunCreativeDirector returned error: %v", err)
	}
	if len(concepts) != 3 {
		t.Fatalf("expected exactly 3 concepts, got %d", len(concepts))
	}
	names := map[string]bool{}
	for _, c := range concepts {
		if c.Name == "" || c.BigIdea == "" {
			t.Errorf("concept missing name/bigIdea: %+v", c)
		}
		names[c.Name] = true
	}
	if len(names) != 3 {
		t.Error("expected 3 genuinely distinct concept names, found duplicates")
	}
	if usage.TokensIn != 610 || usage.TokensOut != 320 {
		t.Errorf("unexpected usage: %+v", usage)
	}
}
