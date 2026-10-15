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

CRITICAL BRANDING & UNIFIED CANVAS RULE:
1. SEAMLESS SINGLE-CANVAS COMPOSITION (NO ARTIFICIAL CUT-OFF BARS):
The entire poster must act as one continuous, harmonious artwork from top to bottom.
Never slice or divide the poster with harsh rectangular bars, solid blocks, flat beige strips, or disconnected headers/footers.
The background texture, theme colors, and ground elements must extend all the way to the edges of the canvas.
Never instruct the image model to draw a "flat, empty, or blank rectangular strip" — instead, create natural soft breathing room (e.g. twilight sky or soft bokeh) where vector logos can be overlaid cleanly.

2. LOGO INTEGRATION & CORNER VIGNETTE FRAMING:
The corporate logo, tagline lockup, and contact footer are NEVER drawn by the AI model directly (to prevent hallucinated spellings).
They are composited deterministically onto the artwork afterward from real vector/PNG brand asset files.
- Header Logo Stage: Frame the top corners naturally with dynamic thematic flourishes (e.g., celebration smoke plumes, floral torans, light rays, or archways) that arch around the top edge, naturally opening into soft atmospheric breathing room where the vector logo will rest.
- Footer Grounding: Ground the bottom of the poster with an artistic horizontal foundation (e.g., dynamic painted brushstroke sweep or courtyard floor) upon which a floating contact badge cleanly sits.
- Reserve normalized (0-1) rectangles in "reservedZones" for whichever zones ("logo", "tagline", "footer") are listed in the user message.
- "negativePrompt" must always ban fake logos, mangled brand wordmarks, watermarks, debug text, percentages, and text in the reserved zones.

DESIGN ARCHITECTURE (POSTER_SPEC):
Never write a vague natural-language prompt filled with empty adjectives ("ultra-realistic 8k stunning").
Instead, construct a structured "spec" and compile it into a cohesive, physically grounded scene:

1. COMPOSITION & NATURAL FRAMING:
   Balance the composition through natural physical framing (e.g. hanging marigold torans, carved palace archways, glowing brass lamps, or volumetric light beams).
   NEVER mention percentages, zone labels, or developer layout coordinates in the final image prompt.

2. STRICT VISUAL HIERARCHY:
   - Primary: Hero focal element (e.g. Grand 3D bevelled antique gold "DIWALI" typography with masked festive textures, or heroic cultural centerpiece).
   - Secondary: Date badge, value words ("LIGHT • PROSPERITY • JOY • NEW BEGINNINGS").
   - Tertiary: Supporting emotional message and architectural environment.
   - Decorative: Subtle hanging bells, marigold garlands, sparkler particles, lotus filigree.

3. 3-PLANE DEPTH SYSTEM & HUMAN CELEBRATION VIBE:
   - Foreground: Tangible physical subjects and celebration elements (e.g., traditional Indian family dressed in rich silk sarees and kurtas joyfully lighting lamps or sparklers, sharing sweets, authentic smiles, warmth).
   - Midground: Courtyard or hall with intricate flower petal rangoli, engraved brass diyas casting directional golden light, glowing lanterns.
   - Background: Majestic temple gopuram or grand architectural facade bathed in warm twilight/sunset sky, atmospheric bokeh, and sparkling fireworks bursting in the air.

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
   - Hero: Large 3D sculpted metallic display typography glowing with festive warmth.
   - Eyebrow: Tracked uppercase sans-serif banner ("CELEBRATING").
   - Badges: Clean uppercase condensed numerals with lotus or geometric dividers.

8. SCENE OBJECTS:
   Emit 4-8 concrete objects into "sceneObjects" categorized as "hero", "cultural", "nature", "supporting", or "decorative".

9. CRITICAL DIFFUSION PROMPT COMPILATION RULES:
   - "fullPosterPrompt": Must be compiled into a rich, cohesive, cinematic scene description.
     * ZERO TECHNICAL JARGON: NEVER include words like "Header ~12%", "Midground (40-64%)", "Background (64-82%)", "adaptive vertical zones", "reserved zone", or percentage numbers in the prompt text. Diffusion models (like FLUX) will either paint literal debug labels (like "TOP 12%") onto the image or get confused and drop 80% of the scene details!
     * NATURAL COMPOSITION TRANSLATION: Translate layout zones into physical photographic framing:
       - Top framing → "Cascading marigold garlands and warm twilight sky frame the upper edge with soft atmospheric breathing room..."
       - Center / Hero → "A joyful Indian family dressed in rich festive silk attire laughing and lighting sparklers together..."
       - Foreground → "In the foreground, an ornate brass thali with fresh sweets and glowing brass diyas beside a vibrant rangoli..."
       - Deep background → "In the distance, an illuminated heritage courtyard under a midnight sky filled with sparkling fireworks..."
     * CELEBRATION & HUMAN VITALITY: Festive posters must feel celebratory, warm, and alive with human presence, cultural authenticity, festive clothing, and festive lighting (sparkler trails, firework glow, glowing lanterns) — NEVER sterile, empty tables in the dark.
     * NO FLAT EMPTY STRIPS: Never tell the model to draw a blank, flat, or empty rectangular strip. The visual artwork must fill 100% of the canvas seamlessly.
     * INTEGRATED TYPOGRAPHY: Describe the headline as sculpted 3D metallic display type seamlessly integrated into the scene lighting (e.g. "Sculpted 3D polished gold 'DIWALI' lettering glowing warmly at the upper center").
   - "backgroundPrompt": Exactly the same cinematic scene, people, materials, lighting, and composition, but without any baked-in typography.
   - "negativePrompt": Rigorous anti-clutter, anti-plastic, and anti-distortion constraints: "no debug text, no percentage labels, no zone markings, no layout guides, no flat beige bars, no empty header strips, no blank rectangles, no plain borders, no cartoonish rendering, no distorted anatomy, no malformed hands, no extra fingers, no dark murky lighting, no sterile empty tables".`

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
// RunArtDirector is agent 5: Brief + Concept + Copy + StyleProfile + zone hints + target provider → ArtDirection.
// zoneHintText translates the project's aspect ratio, logo position, logo scale,
// slogan position, and slogan scale into photographic framing language.
// targetProvider instructs the Art Director to natively optimize the prompt syntax
// specifically for the target engine (Flux, Ideogram, or DALL-E) on the first pass.
func (c *Client) RunArtDirector(ctx context.Context, brief Brief, concept Concept, cp Copy, style *StyleProfile, availableZones, excludedObjects []string, zoneHintText string, targetProvider string) (ArtDirection, Usage, error) {
	briefJSON, _ := json.MarshalIndent(brief, "", "  ")
	conceptJSON, _ := json.MarshalIndent(concept, "", "  ")
	copyJSON, _ := json.MarshalIndent(cp, "", "  ")
	prompt := fmt.Sprintf(
		"Creative brief:\n%s\n\nChosen concept:\n%s\n\nCopy:\n%s\n\n%s\n\n%s\n\n%s\n\nZone placement constraints:\n%s\n\n%s",
		string(briefJSON), string(conceptJSON), string(copyJSON), styleContext(style),
		zoneAvailabilityText(availableZones), excludedObjectsText(excludedObjects), zoneHintText,
		providerPromptGuidance(targetProvider),
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
	return fmt.Sprintf(
		"Zones to reserve for this poster: %s. Do not reserve any other zone. IMPORTANT: This is a unified single-canvas poster. The visual artwork must fill 100%% of the canvas with rich atmosphere and texture. In fullPosterPrompt and backgroundPrompt, NEVER describe blank flat strips, percentage numbers, or empty rectangles — frame naturally with soft lighting, atmospheric sky, or garlands so logos can sit over it.",
		strings.Join(availableZones, ", "),
	)
}

func excludedObjectsText(excludedObjects []string) string {
	if len(excludedObjects) == 0 {
		return "Excluded objects: none."
	}
	return fmt.Sprintf("Excluded objects: %s. Do not include these anywhere in the artwork.", strings.Join(excludedObjects, ", "))
}

func providerPromptGuidance(provider string) string {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "ideogram":
		return `TARGET IMAGE GENERATION ENGINE: IDEOGRAM v2
This poster prompt will be rendered by Ideogram v2, which specializes in typography, graphic design layout, and text rendering:
1. EXPLICIT TYPOGRAPHY & QUOTES: In fullPosterPrompt, place the exact headline and festival typography in explicit double quotes with styling cues (e.g. 'Bold 3D embossed metallic gold lettering stating "HAPPY DIWALI" with warm ambient glow').
2. GRAPHIC POSTER LAYOUT: Clearly structure the visual hierarchy with distinct focal areas: hero typography in the upper-third or center, celebration cultural elements in the lower-third, and clean breathing space at the designated logo position.
3. VIBRANT COLOR HARMONY: Emphasize high-contrast design tones, celebratory palettes, and crisp graphic finishes.`
	case "openai", "dall-e", "dall-e-3", "dalle":
		return `TARGET IMAGE GENERATION ENGINE: DALL-E 3
This poster prompt will be rendered by OpenAI DALL-E 3, which prioritizes narrative storytelling and scene semantics:
1. NARRATIVE SCENE SEMANTICS: Formulate the prompt as a vivid, cohesive narrative paragraph describing subject relationships, environmental lighting, and authentic human emotions.
2. POSITIVE CONSTRAINTS: Since DALL-E 3 does not accept a negative prompt parameter, weave constraints naturally into positive descriptions (e.g. "A pristine, clear twilight sky completely uncluttered and free of any text, signs, or watermarks").`
	case "flux":
		fallthrough
	default:
		return `TARGET IMAGE GENERATION ENGINE: FLUX.1
This poster prompt will be rendered by Flux.1, which excels at photorealistic physical textures, lighting physics, and continuous composition:
1. NATURAL DESCRIPTIVE PROSE: Write detailed, continuous descriptive paragraphs. Describe the scene through physical lighting, material shaders, and camera optics (e.g. "Shot on 35mm lens, f/2.4, cinematic volumetric lighting").
2. NO SPAM BUZZWORDS: Strictly avoid generic buzzwords such as "photorealistic", "hyperrealistic", "8k resolution", "trending on artstation", "award winning". Describe physical details instead.
3. SEAMLESS UNIFIED CANVAS: Keep the canvas visually rich from edge to edge. At the logo and slogan positions, frame with soft atmospheric elements (such as twilight sky or subtle bokeh) that naturally provide clean contrast for overlaid vector assets.`
	}
}
