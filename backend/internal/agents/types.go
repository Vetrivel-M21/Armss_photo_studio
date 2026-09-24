// Package agents implements the LLM pipeline: one file per agent, sharing a
// single tool-use-enforced Anthropic client. Agent output shape is enforced
// via tool-use input schemas, never by prompting for JSON and parsing text
// (spec §17 — that breaks silently on truncation).
package agents

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

type Copy struct {
	Headline   string   `json:"headline"`
	Subhead    string   `json:"subhead"`
	CTA        string   `json:"cta"`
	ValueWords []string `json:"valueWords"`
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

type ArtDirection struct {
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
	MaxTokensCopy     = 3000
	MaxTokensArt      = 4000
)
