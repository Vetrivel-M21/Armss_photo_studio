package fixtures

import (
	"context"
	"testing"
)

func TestRunIntake(t *testing.T) {
	client := newFixtureServer(t, loadFixture(t, "intake_response.json"))

	brief, usage, err := client.RunIntake(context.Background(), "Diwali 2026", nil)
	if err != nil {
		t.Fatalf("RunIntake returned error: %v", err)
	}
	if brief.Occasion != "Diwali 2026" {
		t.Errorf("unexpected occasion: %q", brief.Occasion)
	}
	if len(brief.Assumptions) == 0 {
		t.Error("expected at least one assumption to be recorded")
	}
	if len(brief.Tone) != 3 {
		t.Errorf("expected 3 tone words, got %d", len(brief.Tone))
	}
	if usage.TokensIn != 380 || usage.TokensOut != 210 {
		t.Errorf("unexpected usage: %+v", usage)
	}
}
