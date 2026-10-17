package agents

import (
	"regexp"
	"strings"
)

// AuditVisualConsistency performs a systematic visual consistency review and sanitization
// on the generated ArtDirection against the user's CreativeBrief.
// It resolves:
// 1. Natural objects vs. decorative sculptures/motifs (e.g., natural sunrise vs. unrequested decorative sun motif).
// 2. Optical lighting effects vs. physical objects (e.g., natural sunlight vs. physical sun halo or flower-shaped rays).
// 3. Duplicate major subjects (e.g., multiple suns, duplicate moons).
// 4. Unrequested surreal transformations (e.g., flowers becoming flower-shaped architecture).
// 5. Explicit user constraints (e.g., preserving explicit artistic requests, or enforcing text-free scenes).
func AuditVisualConsistency(art ArtDirection, brief Brief) ArtDirection {
	userIntentText := strings.ToLower(strings.Join([]string{
		brief.Occasion,
		brief.Objective,
		brief.KeyMessage,
		strings.Join(brief.MustInclude, " "),
		strings.Join(brief.Tone, " "),
	}, " "))

	avoidText := strings.ToLower(strings.Join(brief.Avoid, " "))

	explicitSculptedSun := containsAny(userIntentText,
		"sculpted sun", "sun sculpture", "sun motif", "golden sun motif",
		"flower-shaped sun", "sun emblem", "decorative sun", "mythic sun", "sculpture of a sun",
	)

	explicitSculptedMoon := containsAny(userIntentText,
		"sculpted moon", "moon sculpture", "crescent sculpture", "moon motif",
		"decorative moon", "ornamental moon",
	)

	explicitFloralArch := containsAny(userIntentText,
		"flower-shaped building", "floral architecture", "oversized floral sculpture",
		"flower sculpture", "floral building", "flower building",
	)

	explicitFloatingFlames := containsAny(userIntentText,
		"floating flame", "flame sculpture", "floating fire", "sculpted flame",
	)

	wantsNoText := containsAny(userIntentText,
		"no text", "no typography", "text-free", "without text", "clean image without text",
		"no letters", "no words", "no title",
	) || containsAny(avoidText, "text", "typography", "letters", "words", "font")

	// 1. Clean FullPosterPrompt and BackgroundPrompt of unintended duplicate/decorative elements
	art.FullPosterPrompt = sanitizePromptText(art.FullPosterPrompt, explicitSculptedSun, explicitSculptedMoon, explicitFloralArch, explicitFloatingFlames, wantsNoText)
	art.BackgroundPrompt = sanitizePromptText(art.BackgroundPrompt, explicitSculptedSun, explicitSculptedMoon, explicitFloralArch, explicitFloatingFlames, true)

	// If user explicitly requested no text, clear TextToRender
	if wantsNoText {
		art.TextToRender = nil
	}

	// 2. Clean SceneObjects list
	art.SceneObjects = sanitizeSceneObjects(art.SceneObjects, explicitSculptedSun, explicitSculptedMoon, explicitFloralArch, explicitFloatingFlames)

	// 3. Enrich NegativePrompt with context-aware anti-duplication constraints
	art.NegativePrompt = enrichNegativePrompt(art.NegativePrompt, art.FullPosterPrompt+" "+art.BackgroundPrompt, explicitSculptedSun, explicitSculptedMoon, explicitFloralArch, explicitFloatingFlames, wantsNoText)

	return art
}

// sanitizePromptText strips unintended decorative motifs or sculptures when only natural elements were requested.
func sanitizePromptText(prompt string, explicitSun, explicitMoon, explicitFloralArch, explicitFlames, noText bool) string {
	if prompt == "" {
		return ""
	}

	cleaned := prompt

	// 1. Sun disambiguation: If prompt describes a natural sunrise or sun in the sky, remove unrequested decorative sun motifs/sculptures
	if !explicitSun {
		sunMotifRegexes := []*regexp.Regexp{
			// "radiant mythic sun motif with sculpted, flower-shaped rays behind the Pongal pot"
			regexp.MustCompile(`(?i)(?:,\s*|\s+)?(?:with\s+a\s+|a\s+|featuring\s+a\s+)?(?:radiant\s+|mythic\s+|decorative\s+|sculpted\s+|ornamental\s+|golden\s+)*sun\s+motif(?:\s+with\s+sculpted[^\.,;]*)?(?:\s+behind[^\.,;]*)?`),
			// "sculpted, flower-shaped rays behind the pot"
			regexp.MustCompile(`(?i)(?:,\s*|\s+)?(?:with\s+)?sculpted,?\s*flower-shaped\s+rays(?:\s+behind[^\.,;]*)?`),
			// "flower-shaped sun" or "sculpted golden sun behind..."
			regexp.MustCompile(`(?i)(?:,\s*|\s+)?(?:a\s+)?(?:sculpted|flower-shaped)\s+(?:golden\s+)?sun(?:\s+(?:sculpture|halo|motif|disc))?(?:\s+behind[^\.,;]*)?`),
			// "physical sun halo" or "decorative sun halo"
			regexp.MustCompile(`(?i)(?:,\s*|\s+)?(?:a\s+)?(?:physical|sculpted|decorative|ornamental)\s+sun\s+halo(?:\s+behind[^\.,;]*)?`),
			// "second sun" or "duplicate sun"
			regexp.MustCompile(`(?i)(?:,\s*|\s+)?(?:a\s+)?(?:second|duplicate|additional)\s+sun[^\.,;]*`),
		}
		for _, re := range sunMotifRegexes {
			cleaned = re.ReplaceAllString(cleaned, "")
		}
	}

	// 2. Moon disambiguation
	if !explicitMoon {
		moonRegexes := []*regexp.Regexp{
			regexp.MustCompile(`(?i)(?:,\s*|\s+)?(?:with\s+a\s+|a\s+)?(?:sculpted|decorative|ornamental|metallic|acrylic)\s+(?:crescent\s+)?moon\s*(?:sculpture|prop|motif|structure)?(?:\s+in\s+the\s+foreground[^\.,;]*)?`),
			regexp.MustCompile(`(?i)(?:,\s*|\s+)?(?:a\s+)?(?:second|duplicate|additional)\s+moon[^\.,;]*`),
		}
		for _, re := range moonRegexes {
			cleaned = re.ReplaceAllString(cleaned, "")
		}
	}

	// 3. Floating flame disambiguation
	if !explicitFlames {
		flameRegexes := []*regexp.Regexp{
			regexp.MustCompile(`(?i)(?:,\s*|\s+)?(?:with\s+)?(?:floating|disembodied|levitating)\s+flame\s*(?:sculptures?|decorations?|shapes?)[^\.,;]*`),
			regexp.MustCompile(`(?i)(?:,\s*|\s+)?flame-shaped\s+decorations?[^\.,;]*`),
		}
		for _, re := range flameRegexes {
			cleaned = re.ReplaceAllString(cleaned, "")
		}
	}

	// 4. Floral architecture disambiguation
	if !explicitFloralArch {
		floralRegexes := []*regexp.Regexp{
			regexp.MustCompile(`(?i)(?:,\s*|\s+)?(?:with\s+)?flower-shaped\s+(?:architecture|buildings?|structures?|pavilions?)[^\.,;]*`),
			regexp.MustCompile(`(?i)(?:,\s*|\s+)?oversized\s+floral\s+(?:sculptures?|monuments?|architecture)[^\.,;]*`),
		}
		for _, re := range floralRegexes {
			cleaned = re.ReplaceAllString(cleaned, "")
		}
	}

	// 5. Text / typography suppression if text-free requested
	if noText {
		textRegexes := []*regexp.Regexp{
			regexp.MustCompile(`(?i)(?:,\s*|\s+)?(?:with\s+)?(?:sculpted\s+|bold\s+|embossed\s+|3d\s+)?(?:metallic\s+|gold\s+)?(?:display\s+)?typography\s*(?:stating|reading|featuring)?[^,\.;]*`),
			regexp.MustCompile(`(?i)(?:,\s*|\s+)?render\s+the\s+headline[^,\.;]*`),
			regexp.MustCompile(`(?i)(?:,\s*|\s+)?(?:headline|title|lettering)\s+["'][^"']+["'][^,\.;]*`),
			regexp.MustCompile(`(?i)(?:,\s*|\s+)?["'][A-Z0-9\s]{3,}["']\s+lettering[^,\.;]*`),
		}
		for _, re := range textRegexes {
			cleaned = re.ReplaceAllString(cleaned, "")
		}
	}

	// Clean up duplicate commas, spaces, and punctuation artifacts from removals
	cleaned = regexp.MustCompile(`\s*,\s*,+`).ReplaceAllString(cleaned, ",")
	cleaned = regexp.MustCompile(`\s+`).ReplaceAllString(cleaned, " ")
	cleaned = regexp.MustCompile(`\s*,\s*\.`).ReplaceAllString(cleaned, ".")
	cleaned = regexp.MustCompile(`\s*\.\s*\.`).ReplaceAllString(cleaned, ".")
	cleaned = strings.TrimSpace(cleaned)

	return cleaned
}

// sanitizeSceneObjects removes redundant or conflicting decorative objects from sceneObjects.
func sanitizeSceneObjects(objects []SceneObject, explicitSun, explicitMoon, explicitFloralArch, explicitFlames bool) []SceneObject {
	if len(objects) == 0 {
		return objects
	}

	var filtered []SceneObject
	for _, obj := range objects {
		nameLower := strings.ToLower(obj.Name)

		if !explicitSun && (strings.Contains(nameLower, "sun motif") || strings.Contains(nameLower, "sculpted sun") || strings.Contains(nameLower, "sun sculpture") || strings.Contains(nameLower, "flower-shaped sun") || strings.Contains(nameLower, "sun halo")) {
			continue
		}

		if !explicitMoon && (strings.Contains(nameLower, "moon sculpture") || strings.Contains(nameLower, "crescent sculpture") || strings.Contains(nameLower, "moon motif")) {
			continue
		}

		if !explicitFloralArch && (strings.Contains(nameLower, "flower-shaped building") || strings.Contains(nameLower, "floral architecture") || strings.Contains(nameLower, "floral sculpture")) {
			continue
		}

		if !explicitFlames && (strings.Contains(nameLower, "floating flame") || strings.Contains(nameLower, "flame sculpture")) {
			continue
		}

		filtered = append(filtered, obj)
	}

	return filtered
}

// enrichNegativePrompt ensures contextual anti-duplication terms are present in the negative prompt.
func enrichNegativePrompt(negPrompt, promptText string, explicitSun, explicitMoon, explicitFloralArch, explicitFlames, noText bool) string {
	combined := strings.ToLower(promptText)
	additions := []string{}

	hasSunContext := strings.Contains(combined, "sun") || strings.Contains(combined, "sunrise") || strings.Contains(combined, "dawn")
	if hasSunContext && !explicitSun {
		if !strings.Contains(strings.ToLower(negPrompt), "duplicate sun") {
			additions = append(additions, "no duplicate sun", "no second sun", "no decorative sun motif", "no sculpted sun", "no sun halo", "no flower-shaped sun", "no artificial sun sculpture")
		}
	}

	hasMoonContext := strings.Contains(combined, "moon") || strings.Contains(combined, "lunar") || strings.Contains(combined, "night sky")
	if hasMoonContext && !explicitMoon {
		if !strings.Contains(strings.ToLower(negPrompt), "duplicate moon") {
			additions = append(additions, "no duplicate moon", "no second moon", "no ornamental moon sculpture", "no crescent prop")
		}
	}

	hasFloralContext := strings.Contains(combined, "flower") || strings.Contains(combined, "marigold") || strings.Contains(combined, "floral")
	if hasFloralContext && !explicitFloralArch {
		if !strings.Contains(strings.ToLower(negPrompt), "flower-shaped architecture") {
			additions = append(additions, "no flower-shaped architecture", "no surreal floral buildings", "no oversized floral sculptures")
		}
	}

	hasFlameContext := strings.Contains(combined, "diya") || strings.Contains(combined, "lamp") || strings.Contains(combined, "flame") || strings.Contains(combined, "fire")
	if hasFlameContext && !explicitFlames {
		if !strings.Contains(strings.ToLower(negPrompt), "floating flame") {
			additions = append(additions, "no floating flame sculptures", "no disembodied flames")
		}
	}

	if noText {
		if !strings.Contains(strings.ToLower(negPrompt), "no text") {
			additions = append(additions, "no text", "no typography", "no letters", "no words", "no script", "no numbers", "no watermark")
		}
	}

	if len(additions) == 0 {
		return negPrompt
	}

	if strings.TrimSpace(negPrompt) == "" {
		return strings.Join(additions, ", ")
	}

	return strings.TrimSpace(negPrompt) + ", " + strings.Join(additions, ", ")
}

func containsAny(s string, terms ...string) bool {
	for _, term := range terms {
		if strings.Contains(s, term) {
			return true
		}
	}
	return false
}
