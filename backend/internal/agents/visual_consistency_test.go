package agents

import (
	"strings"
	"testing"
)

// Test A — Pongal:
// Input requests a natural sunrise and a traditional Pongal pot.
// Expected: one natural sun in the sky, no decorative sun sculpture.
func TestVisualConsistency_PongalNaturalSun(t *testing.T) {
	brief := Brief{
		Occasion:    "Pongal 2026",
		Objective:   "Celebrate Pongal harvest festival in a realistic South Indian courtyard",
		KeyMessage:  "May the harvest bring prosperity and joy",
		MustInclude: []string{"natural golden-orange sunrise", "one realistic sun in distant sky", "traditional terracotta Pongal pot with rice and milk", "sugarcane", "kolam", "marigold flowers", "jasmine", "lit diyas"},
	}

	// Draft with duplicate/conflicting elements (as Claude previously generated)
	draft := ArtDirection{
		FullPosterPrompt: "A wide cinematic shot of a South Indian courtyard under a natural golden-orange sunrise with one realistic sun in the distant sky. In the center, a traditional terracotta Pongal pot overflowing with boiled milk, with a radiant mythic sun motif with sculpted, flower-shaped rays behind the Pongal pot. Traditional kolam and lit brass diyas rest on the floor.",
		BackgroundPrompt: "A wide cinematic shot of a South Indian courtyard under a natural golden-orange sunrise with one realistic sun in the distant sky. In the center, a traditional terracotta Pongal pot, with a radiant mythic sun motif behind the Pongal pot.",
		NegativePrompt:   "no debug text, no watermarks, no fake logos",
		SceneObjects: []SceneObject{
			{Name: "Natural Sunrise", Category: ObjectCategoryNature},
			{Name: "Terracotta Pongal Pot", Category: ObjectCategoryHero},
			{Name: "Radiant Mythic Sun Motif", Category: ObjectCategoryDecorative},
			{Name: "Lit Diyas", Category: ObjectCategoryCultural},
		},
	}

	audited := AuditVisualConsistency(draft, brief)

	// Verify unrequested decorative sun motif is removed
	if strings.Contains(strings.ToLower(audited.FullPosterPrompt), "mythic sun motif") ||
		strings.Contains(strings.ToLower(audited.FullPosterPrompt), "sculpted, flower-shaped rays") {
		t.Errorf("expected decorative sun motif to be removed from fullPosterPrompt, got: %s", audited.FullPosterPrompt)
	}

	if strings.Contains(strings.ToLower(audited.BackgroundPrompt), "mythic sun motif") {
		t.Errorf("expected decorative sun motif to be removed from backgroundPrompt, got: %s", audited.BackgroundPrompt)
	}

	// Verify natural sun and Pongal pot are preserved
	if !strings.Contains(strings.ToLower(audited.FullPosterPrompt), "natural golden-orange sunrise") ||
		!strings.Contains(strings.ToLower(audited.FullPosterPrompt), "pongal pot") {
		t.Errorf("expected natural sunrise and pongal pot to be preserved, got: %s", audited.FullPosterPrompt)
	}

	// Verify negative prompt includes anti-duplication terms
	for _, term := range []string{"no duplicate sun", "no second sun", "no decorative sun motif", "no sculpted sun"} {
		if !strings.Contains(strings.ToLower(audited.NegativePrompt), term) {
			t.Errorf("expected negativePrompt to contain %q, got: %s", term, audited.NegativePrompt)
		}
	}

	// Verify SceneObjects does not contain the decorative sun motif
	for _, obj := range audited.SceneObjects {
		if strings.Contains(strings.ToLower(obj.Name), "sun motif") || strings.Contains(strings.ToLower(obj.Name), "sculpted sun") {
			t.Errorf("expected sceneObjects to omit decorative sun motif, but found: %+v", obj)
		}
	}
}

// Test B — Diwali:
// Input requests a natural moonlit sky with diyas.
// Expected: a natural moon and ordinary diyas, without unintended moon sculptures or extra floating flames.
func TestVisualConsistency_DiwaliMoonAndDiyas(t *testing.T) {
	brief := Brief{
		Occasion:    "Diwali",
		Objective:   "Diwali celebration under a natural moonlit sky",
		MustInclude: []string{"natural moonlit night sky", "traditional oil diyas glowing on floor", "rangoli"},
	}

	draft := ArtDirection{
		FullPosterPrompt: "A serene temple courtyard under a natural moonlit sky with a silver moon. In the foreground, traditional brass diyas glowing warmly on a floor rangoli, with a sculpted crescent moon sculpture in the foreground and floating flame sculptures in the air.",
		BackgroundPrompt: "A serene temple courtyard under a natural moonlit sky. Traditional brass diyas on floor rangoli, with a sculpted crescent moon sculpture and floating flame decorations.",
		NegativePrompt:   "no debug text, no watermark",
		SceneObjects: []SceneObject{
			{Name: "Natural Moonlit Sky", Category: ObjectCategoryNature},
			{Name: "Brass Diyas", Category: ObjectCategoryCultural},
			{Name: "Sculpted Crescent Moon", Category: ObjectCategoryDecorative},
			{Name: "Floating Flame Sculpture", Category: ObjectCategoryDecorative},
		},
	}

	audited := AuditVisualConsistency(draft, brief)

	// Verify unintended moon sculpture and floating flames are removed
	if strings.Contains(strings.ToLower(audited.FullPosterPrompt), "sculpted crescent moon") ||
		strings.Contains(strings.ToLower(audited.FullPosterPrompt), "floating flame sculptures") {
		t.Errorf("expected moon sculpture and floating flames removed, got: %s", audited.FullPosterPrompt)
	}

	// Verify natural moon and diyas are kept
	if !strings.Contains(strings.ToLower(audited.FullPosterPrompt), "natural moonlit sky") ||
		!strings.Contains(strings.ToLower(audited.FullPosterPrompt), "brass diyas") {
		t.Errorf("expected natural moon and diyas preserved, got: %s", audited.FullPosterPrompt)
	}

	// Verify negative prompt constraints
	for _, term := range []string{"no duplicate moon", "no floating flame sculptures"} {
		if !strings.Contains(strings.ToLower(audited.NegativePrompt), term) {
			t.Errorf("expected negativePrompt to contain %q, got: %s", term, audited.NegativePrompt)
		}
	}

	// Verify SceneObjects
	for _, obj := range audited.SceneObjects {
		if strings.Contains(strings.ToLower(obj.Name), "moon sculpture") || strings.Contains(strings.ToLower(obj.Name), "floating flame") {
			t.Errorf("expected sceneObjects to omit moon/flame sculptures, got: %+v", obj)
		}
	}
}

// Test C — Floral campaign:
// Input requests flowers arranged around a product.
// Expected: natural flowers, without flower-shaped buildings or oversized floral sculptures.
func TestVisualConsistency_FloralCampaign(t *testing.T) {
	brief := Brief{
		Occasion:    "Spring Launch",
		Objective:   "Luxury skincare product campaign",
		MustInclude: []string{"fresh jasmine and marigold flowers arranged around skincare bottle on marble table"},
	}

	draft := ArtDirection{
		FullPosterPrompt: "A sleek minimalist studio shot of a glass skincare bottle on a marble table with fresh jasmine and marigold flowers arranged around the product, set inside a grand flower-shaped building with oversized floral architecture.",
		BackgroundPrompt: "A glass skincare bottle on a marble table with fresh jasmine and marigold flowers, set inside a flower-shaped building.",
		NegativePrompt:   "no logos, no watermarks",
		SceneObjects: []SceneObject{
			{Name: "Skincare Bottle", Category: ObjectCategoryHero},
			{Name: "Fresh Jasmine and Marigolds", Category: ObjectCategoryNature},
			{Name: "Flower-Shaped Building", Category: ObjectCategorySupporting},
		},
	}

	audited := AuditVisualConsistency(draft, brief)

	if strings.Contains(strings.ToLower(audited.FullPosterPrompt), "flower-shaped building") ||
		strings.Contains(strings.ToLower(audited.FullPosterPrompt), "oversized floral architecture") {
		t.Errorf("expected floral architecture removed, got: %s", audited.FullPosterPrompt)
	}

	if !strings.Contains(strings.ToLower(audited.FullPosterPrompt), "fresh jasmine and marigold flowers") {
		t.Errorf("expected natural flowers preserved, got: %s", audited.FullPosterPrompt)
	}

	if !strings.Contains(strings.ToLower(audited.NegativePrompt), "no flower-shaped architecture") {
		t.Errorf("expected negativePrompt to contain 'no flower-shaped architecture', got: %s", audited.NegativePrompt)
	}
}

// Test D — Explicit artistic request:
// Input explicitly requests a sculpted golden sun behind a Pongal pot.
// Expected: retain the sculpted sun because the user specifically requested it.
func TestVisualConsistency_ExplicitSculptedSunRetained(t *testing.T) {
	brief := Brief{
		Occasion:    "Pongal Art Exhibition",
		Objective:   "Artistic celebration poster",
		MustInclude: []string{"a sculpted golden sun motif behind the Pongal pot", "terracotta pot"},
	}

	draft := ArtDirection{
		FullPosterPrompt: "A stylized heritage studio scene featuring a sculpted golden sun motif behind the traditional terracotta Pongal pot, illuminated by warm studio spotlighting.",
		BackgroundPrompt: "A sculpted golden sun motif behind the traditional terracotta Pongal pot.",
		NegativePrompt:   "no debug text, no watermarks",
		SceneObjects: []SceneObject{
			{Name: "Sculpted Golden Sun Motif", Category: ObjectCategoryHero},
			{Name: "Terracotta Pongal Pot", Category: ObjectCategoryCultural},
		},
	}

	audited := AuditVisualConsistency(draft, brief)

	// Since the user EXPLICITLY requested it, the sculpted golden sun motif must be retained!
	if !strings.Contains(strings.ToLower(audited.FullPosterPrompt), "sculpted golden sun motif") {
		t.Errorf("expected explicitly requested sculpted sun to be retained, got: %s", audited.FullPosterPrompt)
	}

	// Negative prompt should NOT ban sculpted sun when explicitly requested
	if strings.Contains(strings.ToLower(audited.NegativePrompt), "no sculpted sun") {
		t.Errorf("negativePrompt should not ban sculpted sun when user explicitly requested it, got: %s", audited.NegativePrompt)
	}
}

// Test E — Text-free image:
// Input explicitly requests no text or typography.
// Expected: preserve that constraint in the final prompt and negative prompt.
func TestVisualConsistency_TextFreeImage(t *testing.T) {
	brief := Brief{
		Occasion:  "Heritage Photography",
		Objective: "Pure photographic visual for digital wallpaper, absolutely text-free image without text or typography",
		Avoid:     []string{"text", "typography", "letters", "words"},
	}

	draft := ArtDirection{
		FullPosterPrompt: "A breathtaking golden hour landscape of South Indian hills and palm groves with dramatic volumetric light. Sculpted 3D metallic display typography stating 'TRANQUILITY' glowing at upper center.",
		BackgroundPrompt: "A breathtaking golden hour landscape of South Indian hills and palm groves with dramatic volumetric light.",
		NegativePrompt:   "no debug text, no watermarks",
		TextToRender:     []string{"TRANQUILITY"},
	}

	audited := AuditVisualConsistency(draft, brief)

	// Verify typography rendering instructions are stripped from FullPosterPrompt
	if strings.Contains(strings.ToLower(audited.FullPosterPrompt), "typography stating") ||
		strings.Contains(strings.ToLower(audited.FullPosterPrompt), "tranquility") {
		t.Errorf("expected typography instructions stripped when text-free is requested, got: %s", audited.FullPosterPrompt)
	}

	// Verify TextToRender is cleared
	if len(audited.TextToRender) > 0 {
		t.Errorf("expected TextToRender to be empty, got: %+v", audited.TextToRender)
	}

	// Verify NegativePrompt bans text
	for _, term := range []string{"no text", "no typography", "no letters", "no words"} {
		if !strings.Contains(strings.ToLower(audited.NegativePrompt), term) {
			t.Errorf("expected negativePrompt to contain %q, got: %s", term, audited.NegativePrompt)
		}
	}
}

// Test F — Existing template:
// Run an existing poster template and confirm its composition, branding, and intended visual elements remain intact.
func TestVisualConsistency_ExistingTemplatePreserved(t *testing.T) {
	brief := Brief{
		Occasion:    "Diwali Corporate Campaign",
		Objective:   "Annual festive greetings for clients and employees",
		MustInclude: []string{"golden diyas", "elegant silk backdrop", "celebration sparkles"},
	}

	original := ArtDirection{
		FullPosterPrompt: "An elegant, prestigious corporate Diwali poster: a graceful arrangement of polished brass diyas casting warm golden light over dark royal navy silk, illuminated by warm ambient candlelight and festive bokeh.",
		BackgroundPrompt: "An elegant arrangement of polished brass diyas over dark royal navy silk with festive bokeh.",
		NegativePrompt:   "no fake logos, no brand marks, no wordmark, no watermark, no debug text",
		AspectRatio:      "4:5",
		ReservedZones: []Zone{
			{ID: "logo", X: 0.05, Y: 0.05, W: 0.2, H: 0.1, Fill: "#16265A", Note: "top left logo"},
			{ID: "tagline", X: 0.05, Y: 0.16, W: 0.3, H: 0.05, Fill: "#16265A", Note: "tagline zone"},
			{ID: "footer", X: 0.0, Y: 0.9, W: 1.0, H: 0.1, Fill: "#16265A", Note: "footer bar"},
		},
		SceneObjects: []SceneObject{
			{Name: "Polished Brass Diyas", Category: ObjectCategoryCultural},
			{Name: "Royal Navy Silk Drapery", Category: ObjectCategoryHero},
			{Name: "Festive Bokeh Sparkles", Category: ObjectCategoryDecorative},
		},
	}

	audited := AuditVisualConsistency(original, brief)

	// Verify full prompt and background prompt retain all their core narrative
	if !strings.Contains(audited.FullPosterPrompt, "polished brass diyas") ||
		!strings.Contains(audited.FullPosterPrompt, "royal navy silk") {
		t.Errorf("expected template prompt content preserved, got: %s", audited.FullPosterPrompt)
	}

	// Verify reserved branding zones remain completely intact
	if len(audited.ReservedZones) != 3 {
		t.Fatalf("expected 3 reserved zones, got %d", len(audited.ReservedZones))
	}
	ids := map[string]bool{}
	for _, z := range audited.ReservedZones {
		ids[z.ID] = true
	}
	for _, expectedID := range []string{"logo", "tagline", "footer"} {
		if !ids[expectedID] {
			t.Errorf("expected zone %q to be preserved", expectedID)
		}
	}

	// Verify scene objects remain intact
	if len(audited.SceneObjects) != 3 {
		t.Errorf("expected 3 scene objects, got %d", len(audited.SceneObjects))
	}

	// Verify essential branding exclusions in negative prompt remain intact
	for _, term := range []string{"no fake logos", "brand marks", "wordmark", "watermark"} {
		if !strings.Contains(audited.NegativePrompt, term) {
			t.Errorf("expected negativePrompt to preserve %q, got: %s", term, audited.NegativePrompt)
		}
	}
}
