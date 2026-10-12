package agents

import (
	"testing"
)

func TestParseConcepts_StandardJSON(t *testing.T) {
	raw := []byte(`{
		"concepts": [
			{"name": "C1", "bigIdea": "Idea 1", "visualMetaphor": "Metaphor 1", "mood": "Mood 1", "palette": "Palette 1"},
			{"name": "C2", "bigIdea": "Idea 2", "visualMetaphor": "Metaphor 2", "mood": "Mood 2", "palette": "Palette 2"},
			{"name": "C3", "bigIdea": "Idea 3", "visualMetaphor": "Metaphor 3", "mood": "Mood 3", "palette": "Palette 3"}
		]
	}`)
	brief := Brief{Occasion: "Diwali", Objective: "Festive poster"}
	res := parseConcepts(raw, brief, nil)
	if len(res) != 3 {
		t.Fatalf("expected 3 concepts, got %d", len(res))
	}
	if res[0].Name != "C1" {
		t.Fatalf("expected name C1, got %s", res[0].Name)
	}
}

func TestParseConcepts_StringifiedJSON(t *testing.T) {
	// This is the exact error case: "cannot unmarshal string into Go struct field conceptsInput.concepts"
	raw := []byte(`{
		"concepts": "[{\"name\": \"StringC1\", \"big_idea\": \"Idea 1\", \"visual_metaphor\": \"Metaphor 1\", \"mood\": \"Mood 1\", \"palette\": \"Palette 1\"}]"
	}`)
	brief := Brief{Occasion: "Festival", Objective: "Celebration"}
	res := parseConcepts(raw, brief, nil)
	if len(res) != 3 {
		t.Fatalf("expected 3 concepts, got %d", len(res))
	}
	if res[0].Name != "StringC1" {
		t.Fatalf("expected name StringC1, got %s", res[0].Name)
	}
	if res[0].BigIdea != "Idea 1" {
		t.Fatalf("expected bigIdea Idea 1, got %s", res[0].BigIdea)
	}
}

func TestParseConcepts_RootArray(t *testing.T) {
	raw := []byte(`[
		{"name": "Root1", "bigIdea": "Idea 1", "visualMetaphor": "Metaphor 1", "mood": "Mood 1", "palette": "Palette 1"}
	]`)
	brief := Brief{Occasion: "Hiring", Objective: "Recruitment"}
	res := parseConcepts(raw, brief, nil)
	if len(res) != 3 {
		t.Fatalf("expected 3 concepts (padded), got %d", len(res))
	}
	if res[0].Name != "Root1" {
		t.Fatalf("expected name Root1, got %s", res[0].Name)
	}
}

func TestParseConcepts_TextFormat(t *testing.T) {
	raw := []byte(`{
		"concepts": "Concept 1: Modern Dawn\nBig Idea: Clean achievement\nVisual Metaphor: Soaring lines\nMood: Inspiring\nPalette: Blue and gold\n\nConcept 2: Golden Festive\nBig Idea: Radiance\nVisual Metaphor: Warm lights\nMood: Joyful\nPalette: Amber\n\nConcept 3: Royal Classic\nBig Idea: Heritage\nVisual Metaphor: Ornate patterns\nMood: Regal\nPalette: Crimson"
	}`)
	brief := Brief{Occasion: "Anniversary", Objective: "Milestone"}
	res := parseConcepts(raw, brief, nil)
	if len(res) != 3 {
		t.Fatalf("expected 3 concepts, got %d", len(res))
	}
	if res[0].Name != "Modern Dawn" {
		t.Fatalf("expected name Modern Dawn, got %s", res[0].Name)
	}
}

func TestParseConcepts_EmptyFallback(t *testing.T) {
	brief := Brief{Occasion: "Tech Summit", Objective: "Showcase innovation", Tone: []string{"Futuristic"}, Palette: []string{"Cyan", "Navy"}}
	res := parseConcepts(nil, brief, nil)
	if len(res) != 3 {
		t.Fatalf("expected 3 concepts, got %d", len(res))
	}
	for _, c := range res {
		if c.Name == "" || c.BigIdea == "" || c.VisualMetaphor == "" || c.Mood == "" || c.Palette == "" {
			t.Fatalf("unexpected empty field in synthesized concept: %+v", c)
		}
	}
}
