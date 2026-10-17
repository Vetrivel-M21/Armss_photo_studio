// Package agents implements the LLM pipeline: one file per agent, sharing a
// single tool-use-enforced Anthropic client. Agent output shape is enforced
// via tool-use input schemas, never by prompting for JSON and parsing text
// (spec §17 — that breaks silently on truncation).
package agents

import (
	"encoding/json"
	"strings"
)

// PaletteHex is one color in a StyleProfile's palette, with its role in the design.
type PaletteHex struct {
	Hex  string `json:"hex"`
	Role string `json:"role"`
}

type StyleProfile struct {
	Summary        string       `json:"summary"`
	Palette        []PaletteHex `json:"palette"`
	Typography     string       `json:"typography"`
	LayoutPatterns string       `json:"layoutPatterns"`
	Imagery        string       `json:"imagery"`
	Finish         string       `json:"finish"`
	Tone           string       `json:"tone"`
	LogoTreatment  string       `json:"logoTreatment"`
	FixedElements  []string     `json:"fixedElements"`
	Consistency    string       `json:"consistency"`
}

type Brief struct {
	Objective   string   `json:"objective"`
	Audience    string   `json:"audience"`
	KeyMessage  string   `json:"keyMessage"`
	Occasion    string   `json:"occasion"`
	EventDate   string   `json:"eventDate"`
	Tone        []string `json:"tone"`
	Format      string   `json:"format"`
	Dimensions  string   `json:"dimensions"`
	Palette     []string `json:"palette"`
	MustInclude []string `json:"mustInclude"`
	Avoid       []string `json:"avoid"`
	Assumptions []string `json:"assumptions"`
	Questions   []string `json:"questions"`
}

type Concept struct {
	Name           string `json:"name"`
	BigIdea        string `json:"bigIdea"`
	VisualMetaphor string `json:"visualMetaphor"`
	Mood           string `json:"mood"`
	Palette        string `json:"palette"`
}

func (c *Concept) UnmarshalJSON(data []byte) error {
	type alias Concept
	var a alias
	if err := json.Unmarshal(data, &a); err == nil && (a.Name != "" && a.BigIdea != "" && a.VisualMetaphor != "") {
		*c = Concept(a)
		return nil
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	getStr := func(keys ...string) string {
		for _, k := range keys {
			if v, ok := raw[k]; ok {
				if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
					return strings.TrimSpace(s)
				}
			}
		}
		return ""
	}

	c.Name = getStr("name", "title", "concept", "conceptName", "concept_name")
	c.BigIdea = getStr("bigIdea", "big_idea", "idea", "concept_idea", "description")
	c.VisualMetaphor = getStr("visualMetaphor", "visual_metaphor", "metaphor", "visual", "visuals")
	c.Mood = getStr("mood", "atmosphere", "tone")
	c.Palette = getStr("palette", "color_palette", "colors", "colorPalette")

	return nil
}

type Copy struct {
	Eyebrow    string   `json:"eyebrow,omitempty"`
	Headline   string   `json:"headline"`
	Subhead    string   `json:"subhead"`
	CTA        string   `json:"cta"`
	ValueWords []string `json:"valueWords"`
	DateText   string   `json:"dateText,omitempty"`
	Tagline    string   `json:"tagline,omitempty"`
	Alternates []string `json:"alternates"`
}

// Zone IDs the Art Director is asked to emit — the compositor keys off these.
const (
	ZoneLogo    = "logo"
	ZoneTagline = "tagline"
	ZoneFooter  = "footer"
)

// Zone is a normalized (0-1) rectangle the image prompt is told to keep clean,
// which the compositor later fills with a brand asset. This is the mechanism
// that keeps the logo pixel-exact — never let the image model draw it (spec §6/§17).
type Zone struct {
	ID   string  `json:"id"` // "logo" | "tagline" | "footer"
	X    float64 `json:"x"`
	Y    float64 `json:"y"`
	W    float64 `json:"w"`
	H    float64 `json:"h"`
	Fill string  `json:"fill"`
	Note string  `json:"note"`
}

// SceneObject is one concrete object/prop the Art Director wants rendered
// into the scene, grouped into a fixed taxonomy so the UI can show a
// consistent categorized breakdown regardless of occasion (a corporate
// poster may simply have zero "nature" objects — categories are not all
// required to be populated).
type SceneObject struct {
	Name     string `json:"name"`
	Category string `json:"category"` // "hero" | "cultural" | "nature" | "supporting" | "decorative"
}

const (
	ObjectCategoryHero       = "hero"
	ObjectCategoryCultural   = "cultural"
	ObjectCategoryNature     = "nature"
	ObjectCategorySupporting = "supporting"
	ObjectCategoryDecorative = "decorative"
)

// SectionAllocation defines vertical layout zones with adaptive percentages.
type SectionAllocation struct {
	Section    string `json:"section"`    // "header" | "hero" | "middle" | "foreground" | "footer"
	Percentage int    `json:"percentage"` // e.g. 12, 28, 22, 26, 12
	Role       string `json:"role"`
}

// CompositionPlan details layout structure, framing, and layout strategy.
type CompositionPlan struct {
	Strategy    string              `json:"strategy"` // "hero_focused" | "editorial_framed" | "layered_cinematic" | "centered_radial"
	Description string              `json:"description"`
	Sections    []SectionAllocation `json:"sections"`
	Framing     string              `json:"framing"` // e.g. "Ornate brass bells and marigold toran framing top arch"
}

// VisualHierarchy details the prioritized design elements.
type VisualHierarchy struct {
	PrimaryElement     string   `json:"primaryElement"`    // e.g. "DIWALI hero 3D metallic display typography"
	SecondaryElement   string   `json:"secondaryElement"`  // e.g. "8th November 2026 date badge with value pills"
	TertiaryElement    string   `json:"tertiaryElement"`   // e.g. "Temple corridor illuminated at sunset"
	DecorativeElements []string `json:"decorativeElements"`// e.g. ["Hanging brass bells", "Marigold garland toran", "Sparkler flares"]
}

// DepthPlanes defines foreground, midground, and background layers.
type DepthPlanes struct {
	Foreground string `json:"foreground"` // e.g. "Indian family in silk attire lighting brass lamps on marble floor"
	Midground  string `json:"midground"`  // e.g. "Intricate golden floor rangoli with hero engraved diya and illuminated temple corridor"
	Background string `json:"background"` // e.g. "Majestic South Indian temple gopuram against twilight sky with floating ember particles"
	FocusDepth string `json:"focusDepth"` // e.g. "Deep cinematic depth-of-field, sharp subjects with soft atmospheric perspective"
}

// RelationalIntegration models physical interactions between objects.
type RelationalIntegration struct {
	Interactions []string `json:"interactions"` // e.g. ["Diya flame casts warm golden light onto rangoli", "Reflections shimmer on polished marble"]
	Shadows      string   `json:"shadows"`      // e.g. "Directional soft shadows falling outward from the central diya"
	LightSpill   string   `json:"lightSpill"`   // e.g. "Warm amber light spill on family hands and silk fabric folds"
}

// MaterialRealism models tangible, realistic surfaces instead of generic terms.
type MaterialRealism struct {
	Metals   string `json:"metals"`   // e.g. "Hand-carved antique brass with soft specular highlights, polished warm reflections"
	Fabrics  string `json:"fabrics"`  // e.g. "Rich Kanchipuram silk with woven gold zari threads and natural drape"
	Surfaces string `json:"surfaces"` // e.g. "Polished dark granite/marble reflecting diya flames and flower petals"
	Organic  string `json:"organic"`  // e.g. "Fresh orange and yellow marigold flowers with visible moisture and petal texture"
}

// LightingSetup models directional, motivated illumination.
type LightingSetup struct {
	PrimaryLight   string `json:"primaryLight"`   // e.g. "Warm organic flame light from oil diyas"
	SecondaryLight string `json:"secondaryLight"` // e.g. "Golden hour twilight ambient illumination"
	RimLight       string `json:"rimLight"`       // e.g. "Subtle warm rim light outlining family silhouettes and temple architecture"
	Atmosphere     string `json:"atmosphere"`     // e.g. "Subtle floating golden sparkles and holy incense haze"
}

// ColorDistribution specifies the harmonious palette balance.
type ColorDistribution struct {
	Dominant  string `json:"dominant"`  // e.g. "Deep Royal Navy (40%)"
	Secondary string `json:"secondary"` // e.g. "Warm Sunset Crimson (25%)"
	Neutral   string `json:"neutral"`   // e.g. "Lustrous Antique Gold (20%)"
	Accents   string `json:"accents"`   // e.g. "Vibrant Marigold Saffron (15%)"
}

// TypographyPlan defines typographic treatment for each hierarchy level.
type TypographyPlan struct {
	HeroTitleTreatment string `json:"heroTitleTreatment"` // e.g. "Grand 3D bevelled antique gold metallic lettering with masked festival imagery"
	EyebrowStyle       string `json:"eyebrowStyle"`       // e.g. "Clean tracked uppercase sans-serif"
	DateAndBadges      string `json:"dateAndBadges"`      // e.g. "Bold condensed numerals flanked by traditional lotus filigree dividers"
	SupportingBody     string `json:"supportingBody"`     // e.g. "Warm italicized editorial serif typography"
}

// PosterSpec is the complete creative specification constructed before prompt synthesis.
type PosterSpec struct {
	ArtDirectionStyle string                `json:"artDirectionStyle"`
	Composition       CompositionPlan       `json:"composition"`
	VisualHierarchy   VisualHierarchy       `json:"visualHierarchy"`
	Depth             DepthPlanes           `json:"depth"`
	Integration       RelationalIntegration `json:"integration"`
	Materials         MaterialRealism       `json:"materials"`
	Lighting          LightingSetup         `json:"lighting"`
	Colors            ColorDistribution     `json:"colors"`
	Typography        TypographyPlan        `json:"typography"`
	NegativeRules     []string              `json:"negativeRules"`
}

type ArtDirection struct {
	Spec             *PosterSpec   `json:"spec,omitempty"`
	Layout           string        `json:"layout"`
	FocalPoint       string        `json:"focalPoint"`
	Typography       string        `json:"typography"`
	Colour           string        `json:"colour"`
	TextToRender     []string      `json:"textToRender"`
	FullPosterPrompt string        `json:"fullPosterPrompt"`
	BackgroundPrompt string        `json:"backgroundPrompt"`
	NegativePrompt   string        `json:"negativePrompt"`
	AspectRatio      string        `json:"aspectRatio"`
	ReservedZones    []Zone        `json:"reservedZones"`
	CompositeSteps   []string      `json:"compositeSteps"`
	SceneObjects     []SceneObject `json:"sceneObjects"`
}

// Review is agent 7's output. Defined for schema completeness even though the
// Reviewer agent itself is out of scope for this build (spec §15 steps 1-8).
type Review struct {
	Score      int      `json:"score"`
	Verdict    string   `json:"verdict"`
	HouseMatch string   `json:"houseMatch"`
	Working    []string `json:"working"`
	Risks      []string `json:"risks"`
	Fixes      []string `json:"fixes"`
}

// Usage is the token spend of a single agent call, for persisting into stage_outputs.
type Usage struct {
	TokensIn  int
	TokensOut int
}

// Token budgets per spec §5. A 1000-token cap truncates output silently — this
// is the single most likely source of bugs, so these are generous on purpose.
const (
	MaxTokensStyle    = 4000
	MaxTokensIntake   = 4000
	MaxTokensDirector = 4000
	MaxTokensCopy     = 4000
	MaxTokensArt      = 6000
)
