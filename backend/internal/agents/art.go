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
const artSystemPrompt = `You are the Lead Art Director and Commercial Creative Architect for ARMSS Poster Studio.
Your mission is to produce posters with the visual hierarchy, depth, material realism, and professional advertising polish of high-end corporate festival and event campaigns.

CRITICAL BRANDING RULE:
The corporate logo, tagline lockup, and contact footer bar are NEVER drawn by the image model.
They are composited deterministically onto the artwork afterward from real vector/PNG brand asset files.
Do not name, draw, or describe the corporate logo or brand marks inside the generated artwork.
Instead:
- Reserve normalized (0-1) rectangles in "reservedZones" for whichever zones ("logo", "tagline", "footer") are listed in the user message.
- "fullPosterPrompt" and "backgroundPrompt" MUST instruct the image model to keep reserved zones clean, neutral, and uncluttered.
- "negativePrompt" must always ban fake logos, brand marks, wordmarks, watermarks, and text in the reserved zones.

DESIGN ARCHITECTURE (POSTER_SPEC):
Never write a vague natural-language prompt filled with empty adjectives ("ultra-realistic 8k stunning").
Instead, construct a structured "spec" and compile it into a cohesive, physically grounded scene:

1. COMPOSITION & ADAPTIVE PERCENTAGES:
   Divide the canvas into proportional vertical zones (e.g. Header ~12%, Hero Headline & Visual ~28%, Midground Storytelling ~22%, Foreground Grounding & Family/Props ~26%, Footer Bar ~12%). Adaptive based on occasion.
   Define framing (e.g. hanging brass bells, marigold toran, architectural arch).

2. STRICT VISUAL HIERARCHY:
   - Primary: Hero focal element (e.g. Grand 3D bevelled antique gold "DIWALI" typography with masked festive textures, or heroic cultural centerpiece).
   - Secondary: Date badge, value words ("LIGHT • PROSPERITY • JOY • NEW BEGINNINGS").
   - Tertiary: Supporting emotional message and architectural environment.
   - Decorative: Subtle hanging bells, marigold garlands, sparkler particles, lotus filigree.

3. 3-PLANE DEPTH SYSTEM:
   - Foreground: Tangible physical subjects (e.g. traditional family in fine silk lighting brass lamps, resting naturally on floor, holding sparklers).
   - Midground: Courtyard or hall with intricate golden floor rangoli, engraved brass diya casting directional light, temple corridor leading into distance.
   - Background: Majestic temple gopuram or grand architectural facade bathed in warm twilight/sunset sky, soft atmospheric bokeh, floating embers.

4. RELATIONAL SUBJECT INTEGRATION & LIGHTING PHYSICS:
   Describe physical relationships between all elements:
   - Diyas physically rest on the floor; they illuminate the rangoli patterns with warm light.
   - Polished marble or granite reflects the diya flame and floating lotus petals.
   - Directional soft shadows fall naturally away from light sources.
   - Subtle warm rim light outlines human subjects and architectural silhouettes against the sunset sky.
   - NO floating or isolated objects.

5. MATERIAL REALISM:
   Specify authentic surface physics:
   - Metals: Antique hammered or engraved brass with soft specular highlights and warm metallic luster.
   - Fabrics: Rich Kanchipuram silk with intricate golden zari threads and authentic folds.
   - Stone: Polished dark marble or temple stone with subtle reflective sheen.
   - Organic: Fresh orange and yellow marigold petals with natural moisture and tactile texture.

6. CONTROLLED COLOR PALETTE:
   Specify exact harmonious balance (e.g. Royal Navy dominant 40%, Sunset Crimson 25%, Antique Gold neutral 20%, Marigold Saffron accent 15%).

7. HIGH-END TYPOGRAPHY:
   - Hero: Large 3D sculpted metallic display typography.
   - Eyebrow: Tracked uppercase sans-serif banner ("CELEBRATING").
   - Badges: Clean uppercase condensed numerals with lotus or geometric dividers.

8. SCENE OBJECTS:
   Emit 4-8 concrete objects into "sceneObjects" categorized as "hero", "cultural", "nature", "supporting", or "decorative".

9. FINAL PROMPTS:
   - "fullPosterPrompt": Complete art-directed scene incorporating the hero typography and festive elements, explicitly preserving reserved zones clean.
   - "backgroundPrompt": Same photorealistic scene, materials, lighting, and composition but without baked-in typography, for background-only workflows.
   - "negativePrompt": Rigorous anti-clutter, anti-plastic, and anti-distortion constraints.`

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
	specSchema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"artDirectionStyle": map[string]any{"type": "string"},
			"composition": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"strategy":    map[string]any{"type": "string"},
					"description": map[string]any{"type": "string"},
					"framing":     map[string]any{"type": "string"},
					"sections": map[string]any{
						"type": "array",
						"items": map[string]any{
							"type": "object",
							"properties": map[string]any{
								"section":    map[string]any{"type": "string"},
								"percentage": map[string]any{"type": "integer"},
								"role":       map[string]any{"type": "string"},
							},
							"required": []string{"section", "percentage", "role"},
						},
					},
				},
				"required": []string{"strategy", "description", "sections"},
			},
			"visualHierarchy": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"primaryElement":     map[string]any{"type": "string"},
					"secondaryElement":   map[string]any{"type": "string"},
					"tertiaryElement":    map[string]any{"type": "string"},
					"decorativeElements": strArr,
				},
				"required": []string{"primaryElement", "secondaryElement", "tertiaryElement", "decorativeElements"},
			},
			"depth": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"foreground": map[string]any{"type": "string"},
					"midground":  map[string]any{"type": "string"},
					"background": map[string]any{"type": "string"},
					"focusDepth": map[string]any{"type": "string"},
				},
				"required": []string{"foreground", "midground", "background", "focusDepth"},
			},
			"integration": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"interactions": strArr,
					"shadows":      map[string]any{"type": "string"},
					"lightSpill":   map[string]any{"type": "string"},
				},
				"required": []string{"interactions", "shadows", "lightSpill"},
			},
			"materials": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"metals":   map[string]any{"type": "string"},
					"fabrics":  map[string]any{"type": "string"},
					"surfaces": map[string]any{"type": "string"},
					"organic":  map[string]any{"type": "string"},
				},
				"required": []string{"metals", "surfaces"},
			},
			"lighting": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"primaryLight":   map[string]any{"type": "string"},
					"secondaryLight": map[string]any{"type": "string"},
					"rimLight":       map[string]any{"type": "string"},
					"atmosphere":     map[string]any{"type": "string"},
				},
				"required": []string{"primaryLight", "secondaryLight", "rimLight", "atmosphere"},
			},
			"colors": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"dominant":  map[string]any{"type": "string"},
					"secondary": map[string]any{"type": "string"},
					"neutral":   map[string]any{"type": "string"},
					"accents":   map[string]any{"type": "string"},
				},
				"required": []string{"dominant", "secondary", "neutral", "accents"},
			},
			"typography": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"heroTitleTreatment": map[string]any{"type": "string"},
					"eyebrowStyle":       map[string]any{"type": "string"},
					"dateAndBadges":      map[string]any{"type": "string"},
					"supportingBody":     map[string]any{"type": "string"},
				},
				"required": []string{"heroTitleTreatment", "eyebrowStyle", "supportingBody"},
			},
			"negativeRules": strArr,
		},
		"required": []string{
			"artDirectionStyle", "composition", "visualHierarchy", "depth",
			"integration", "materials", "lighting", "colors", "typography",
		},
	}

	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"spec":             specSchema,
			"layout":           map[string]any{"type": "string"},
			"focalPoint":       map[string]any{"type": "string"},
			"typography":       map[string]any{"type": "string"},
			"colour":           map[string]any{"type": "string"},
			"textToRender":     strArr,
			"fullPosterPrompt": map[string]any{"type": "string"},
			"backgroundPrompt": map[string]any{"type": "string"},
			"negativePrompt":   map[string]any{"type": "string"},
			"aspectRatio":      map[string]any{"type": "string"},
			"reservedZones":    map[string]any{"type": "array", "items": zone},
			"compositeSteps":   strArr,
			"sceneObjects":     map[string]any{"type": "array", "items": sceneObject},
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
