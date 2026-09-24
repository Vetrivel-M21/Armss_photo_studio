package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// artSystemPrompt encodes the single most important correctness rule in the whole
// system (spec §6/§17): the image model must NEVER be asked to draw the ARMSS logo,
// tagline lockup, or footer bar. A prior attempt at this produced "OARMSS". Instead,
// the Art Director reserves normalized rectangles for those elements and the backend
// composites the real brand assets in afterward.
const artSystemPrompt = `You are the Art Director for ARMSS Poster Studio. Given a creative
brief, the chosen concept, the copy, and (optionally) a house style profile, produce a full
art direction spec for the image-generation step.

CRITICAL RULE: the ARMSS logo, tagline lockup, and contact footer are never drawn by the
image model. They are composited onto the artwork afterward from real brand asset files.
Do not name, describe, or reference the ARMSS logo mark anywhere in your prompts — describing
it (even to ask for it) causes the image model to attempt a fake logo and fail. Instead:

- The user message tells you exactly which zones ("logo", "tagline", "footer") will
  actually have something composited into them for this poster — reserve space ONLY for
  those. If none are listed, emit an empty "reservedZones" array and use the full canvas:
  let the hero visual, headline, and decorative elements naturally fill the space that
  would otherwise sit empty. Never reserve a zone for something that won't be
  composited — an unused reserved zone becomes a visibly empty gap in the final poster.
- For each zone you DO reserve: a normalized (0-1) rectangle the artwork must leave
  visually clean and uncluttered at those coordinates, with a "fill" hex the background
  should be in that area and a short "note" describing what will be placed there later
  (for prompt context only, e.g. "space reserved for brand lockup").
- "fullPosterPrompt" must explicitly instruct the image model to keep any reserved zones
  clean, uncluttered, and free of any text or graphic marks.
- "negativePrompt" must always include: logo, brand marks, wordmark, watermark, text in the
  top band — in addition to any other negative terms appropriate to the concept.
- "backgroundPrompt" is the same artwork direction but without any headline/copy text baked
  in, for providers that can't render text reliably (the backend then typesets the headline
  itself in that case).
- "textToRender" lists the exact copy strings (headline, subhead, CTA) the artwork should
  render, for providers that do support text rendering.

VISUAL RICHNESS: A poster with an empty or vague background reads as unfinished. Beyond
the reserved brand zones, "fullPosterPrompt" and "backgroundPrompt" must describe 3-6
concrete, specific objects or props that belong to this occasion — not generic phrases
like "festive decorations" or "decorative elements." Ground these in what you were given:
translate the concept's "visualMetaphor" into literal objects, pull directly from the
style profile's "fixedElements" and "imagery" fields when a style profile is supplied, and
infer occasion-appropriate props from the brief's "occasion" and "keyMessage" when no style
profile is supplied. For example, for a Diwali poster: "terracotta diyas with lit flames,
marigold garlands, a brass rangoli pattern in the foreground" rather than "festive Diwali
decorations." Name real, renderable objects, materials, and their placement — never a
generic mood phrase as a substitute for actual content.

Also emit every one of those objects again as "sceneObjects" — a flat list of {name,
category} pairs describing exactly the same objects named in fullPosterPrompt, so the UI
can show the user a categorized breakdown of what's about to be generated. Assign each
object to exactly one of five fixed categories: "hero" (the single largest/most central
visual subject — usually 1 object), "cultural" (occasion-specific cultural/traditional
objects), "nature" (natural elements — sky, plants, landscape), "supporting" (secondary
objects that reinforce the scene without being decorative filler), "decorative" (patterns,
motifs, ornamental flourishes, particle/light effects). Not every category needs an entry —
a corporate poster may have zero "nature" objects — but do not invent objects solely to
fill a category; only list what you actually described in the prompts. If the user message
includes an "Excluded objects" list, do not include any of those objects anywhere — in
fullPosterPrompt, backgroundPrompt, or sceneObjects — and, if needed, replace them with a
different concrete object of the same category so the scene stays visually rich rather than
simply emptier.

When a reference style profile is supplied, it governs composition, layout, visual
hierarchy, and lighting direction — how the poster looks. Reflect its "summary" field
concretely in "layout", "focalPoint", and "fullPosterPrompt" — don't just acknowledge it in
the abstract. It never governs what the poster says: the user's brief, chosen concept, and
copy always take priority over the style profile's decorative details. Adapt the style
profile's colors to any brand palette the brief specifies. Never invent event details,
dates, prices, or claims that weren't provided in the brief or copy.`

func artToolSchema() map[string]any {
	strArr := map[string]any{"type": "array", "items": map[string]any{"type": "string"}}
	zone := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"id":   map[string]any{"type": "string"},
			"x":    map[string]any{"type": "number"},
			"y":    map[string]any{"type": "number"},
			"w":    map[string]any{"type": "number"},
			"h":    map[string]any{"type": "number"},
			"fill": map[string]any{"type": "string"},
			"note": map[string]any{"type": "string"},
		},
		"required": []string{"id", "x", "y", "w", "h", "fill", "note"},
	}
	sceneObject := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{"type": "string"},
			"category": map[string]any{
				"type": "string",
				"enum": []string{
					ObjectCategoryHero, ObjectCategoryCultural, ObjectCategoryNature,
					ObjectCategorySupporting, ObjectCategoryDecorative,
				},
			},
		},
		"required": []string{"name", "category"},
	}
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"layout":           map[string]any{"type": "string"},
			"focalPoint":       map[string]any{"type": "string"},
			"typography":       map[string]any{"type": "string"},
			"colour":           map[string]any{"type": "string"},
			"textToRender":     strArr,
			"fullPosterPrompt": map[string]any{"type": "string"},
			"backgroundPrompt": map[string]any{"type": "string"},
			"negativePrompt":   map[string]any{"type": "string"},
			"aspectRatio":      map[string]any{"type": "string"},
			// No minItems: reservedZones is driven by which brand elements are
			// actually available for this poster (see RunArtDirector's
			// availableZones param) — it can legitimately be empty.
			"reservedZones":  map[string]any{"type": "array", "items": zone},
			"compositeSteps": strArr,
			"sceneObjects":   map[string]any{"type": "array", "items": sceneObject},
		},
		"required": []string{
			"layout", "focalPoint", "typography", "colour", "textToRender",
			"fullPosterPrompt", "backgroundPrompt", "negativePrompt", "aspectRatio",
			"reservedZones", "compositeSteps", "sceneObjects",
		},
	}
}

// RunArtDirector is agent 5: Brief + Concept + Copy + StyleProfile → ArtDirection.
// availableZones lists which of "logo"/"tagline"/"footer" will actually be
// composited in afterward (the caller resolves this from what brand assets
// and footer text are actually available for this project) — only those get
// reserved space in the artwork, so no poster ends up with a blank gap for a
// brand element it was never going to have. excludedObjects (may be nil) names
// scene objects the user has explicitly removed from a previous run's
// sceneObjects list — the Art Director keeps them out of the artwork entirely
// rather than just noting them as unwanted.
func (c *Client) RunArtDirector(ctx context.Context, brief Brief, concept Concept, cp Copy, style *StyleProfile, availableZones, excludedObjects []string) (ArtDirection, Usage, error) {
	briefJSON, _ := json.MarshalIndent(brief, "", "  ")
	conceptJSON, _ := json.MarshalIndent(concept, "", "  ")
	copyJSON, _ := json.MarshalIndent(cp, "", "  ")
	prompt := fmt.Sprintf(
		"Creative brief:\n%s\n\nChosen concept:\n%s\n\nCopy:\n%s\n\n%s\n\n%s\n\n%s",
		string(briefJSON), string(conceptJSON), string(copyJSON), styleContext(style),
		zoneAvailabilityText(availableZones), excludedObjectsText(excludedObjects),
	)

	result, err := c.CallTool(ctx, ToolCallRequest{
		System:   artSystemPrompt,
		UserText: prompt,
		Tool: ToolSpec{
			Name:        "emit_art_direction",
			Description: "Emit the full art direction spec, including reserved zones for brand assets.",
			InputSchema: artToolSchema(),
		},
		MaxTokens: MaxTokensArt,
	})
	if err != nil {
		return ArtDirection{}, Usage{}, err
	}

	var out ArtDirection
	if err := json.Unmarshal(result.Raw, &out); err != nil {
		return ArtDirection{}, Usage{}, fmt.Errorf("unmarshal art direction: %w", err)
	}
	return out, Usage{TokensIn: result.TokensIn, TokensOut: result.TokensOut}, nil
}

func zoneAvailabilityText(availableZones []string) string {
	if len(availableZones) == 0 {
		return "Zones to reserve for this poster: none. Use the full canvas — do not leave any empty gap."
	}
	return fmt.Sprintf("Zones to reserve for this poster: %s. Do not reserve any other zone.", strings.Join(availableZones, ", "))
}

func excludedObjectsText(excludedObjects []string) string {
	if len(excludedObjects) == 0 {
		return "Excluded objects: none."
	}
	return fmt.Sprintf("Excluded objects: %s. Do not include these anywhere in the artwork.", strings.Join(excludedObjects, ", "))
}
