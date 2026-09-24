package fixtures

import (
	"context"
	"strings"
	"testing"

	"armss-poster-studio/internal/agents"
)

func TestRunArtDirector(t *testing.T) {
	client := newFixtureServer(t, loadFixture(t, "art_response.json"))

	art, usage, err := client.RunArtDirector(context.Background(),
		agents.Brief{Occasion: "Diwali 2026"},
		agents.Concept{Name: "Lamps of Opportunity"},
		agents.Copy{Headline: "Light the Way Forward"},
		nil,
		[]string{agents.ZoneLogo, agents.ZoneTagline, agents.ZoneFooter},
		nil)
	if err != nil {
		t.Fatalf("RunArtDirector returned error: %v", err)
	}

	// The single most important correctness constraint in the spec: the negative
	// prompt must always exclude the logo/brand marks so the image model never
	// attempts to draw them (spec §6).
	for _, term := range []string{"logo", "brand marks", "wordmark", "watermark"} {
		if !strings.Contains(art.NegativePrompt, term) {
			t.Errorf("negativePrompt missing required term %q: %q", term, art.NegativePrompt)
		}
	}

	// And the three reserved zones (logo, tagline, footer) must be present so the
	// compositor has somewhere to place the real brand assets afterward.
	ids := map[string]bool{}
	for _, z := range art.ReservedZones {
		ids[z.ID] = true
	}
	for _, want := range []string{"logo", "tagline", "footer"} {
		if !ids[want] {
			t.Errorf("expected a reserved zone with id %q, got zones %+v", want, art.ReservedZones)
		}
	}

	if art.FullPosterPrompt == "" || art.BackgroundPrompt == "" {
		t.Error("expected both fullPosterPrompt and backgroundPrompt to be non-empty")
	}
	if usage.TokensIn != 720 || usage.TokensOut != 410 {
		t.Errorf("unexpected usage: %+v", usage)
	}
}
