package fixtures

import (
	"context"
	"testing"

	"armss-poster-studio/internal/agents"
)

func TestAnalyzeStyle(t *testing.T) {
	client := newFixtureServer(t, loadFixture(t, "style_response.json"))

	profile, usage, err := client.AnalyzeStyle(context.Background(), []agents.ImageInput{
		{Data: []byte("fake-image-bytes"), MediaType: "image/png"},
	})
	if err != nil {
		t.Fatalf("AnalyzeStyle returned error: %v", err)
	}
	if profile.Summary == "" {
		t.Error("expected non-empty summary")
	}
	if len(profile.Palette) != 3 {
		t.Errorf("expected 3 palette entries, got %d", len(profile.Palette))
	}
	if profile.Palette[0].Hex != "#16265A" || profile.Palette[0].Role != "primary" {
		t.Errorf("unexpected first palette entry: %+v", profile.Palette[0])
	}
	if len(profile.FixedElements) != 2 {
		t.Errorf("expected 2 fixed elements, got %d", len(profile.FixedElements))
	}
	if usage.TokensIn != 512 || usage.TokensOut != 240 {
		t.Errorf("unexpected usage: %+v", usage)
	}
}

func TestAnalyzeStyle_RequiresAtLeastOneImage(t *testing.T) {
	client := newFixtureServer(t, loadFixture(t, "style_response.json"))

	_, _, err := client.AnalyzeStyle(context.Background(), nil)
	if err == nil {
		t.Fatal("expected an error when no images are provided")
	}
}
