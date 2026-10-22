// Package render defines the VideoSpec — the single source of truth that both
// the Canvas preview engine (TypeScript) and the FFmpeg worker (Go) read. Any
// change to this struct requires a matching change in
// web/src/modules/video/engine/spec.ts AND a spec_version bump.
package render

// SpecVersion is incremented whenever the shape of VideoSpec changes in a
// backward-incompatible way. A migration function must be provided for every
// version bump so that saved projects can be upgraded on load.
const SpecVersion = 1

// ---- top-level spec --------------------------------------------------------

// VideoSpec is the complete description of one video project. It is stored as
// JSON in video_projects.spec and snapshot-copied into video_jobs.spec when an
// export is submitted.
type VideoSpec struct {
	Version int        `json:"version"`
	Output  OutputSpec `json:"output"`
	Clips   []ClipSpec `json:"clips"`
	Audio   *AudioSpec `json:"audio,omitempty"`
}

// ---- output ----------------------------------------------------------------

// OutputSpec controls the final video dimensions and frame rate.
type OutputSpec struct {
	// Aspect is one of "9:16" | "1:1" | "16:9". Default: "9:16".
	Aspect string `json:"aspect"`

	// Fit controls how each image fills the frame.
	// "cover"   — scale-and-crop to fill (default).
	// "contain" — letterbox with blurred background fill (same image, heavily
	//             blurred and scaled to fill, then sharp image overlaid).
	Fit string `json:"fit"`

	// FPS is the output frame rate. Must be 24 or 30. Default: 30.
	FPS int `json:"fps"`
}

// ---- clip ------------------------------------------------------------------

// ClipSpec is one image slide. Order in the Clips slice determines play order.
type ClipSpec struct {
	// ID is a stable client-generated identifier (UUID v4).
	ID string `json:"id"`

	// ImageKey is the S3 key of the source image, always under
	// video/{user_id}/images/. Validated server-side against the requesting
	// user's prefix before any render.
	ImageKey string `json:"image_key"`

	// Duration is the clip's on-screen time in seconds, including the
	// outgoing transition overlap. Range: 1–10. Default: 3.
	Duration float64 `json:"duration"`

	// Crop defines a sub-rectangle of the image to use, in normalized coords.
	// Zero value {0,0,1,1} means "use the full image".
	Crop CropSpec `json:"crop"`

	// Rotate is clockwise rotation in degrees: 0 | 90 | 180 | 270.
	Rotate int `json:"rotate"`

	// KenBurns defines an animated zoom-pan over the clip's duration.
	// Omit (nil) for a static frame.
	KenBurns *KenBurnsSpec `json:"ken_burns,omitempty"`

	// Filter is one of the named color presets. "none" means no filter.
	// Valid values: "none" | "vivid" | "warm" | "cold" | "bw" | "vintage"
	Filter string `json:"filter"`

	// Adjust contains per-channel scalar adjustments applied on top of Filter.
	Adjust AdjustSpec `json:"adjust"`

	// Texts is the list of text overlays for this clip. Max 5.
	Texts []TextSpec `json:"texts,omitempty"`

	// TransitionOut is the transition from this clip to the next.
	// Nil on the last clip.
	TransitionOut *TransitionSpec `json:"transition_out,omitempty"`
}

// CropSpec is a normalized sub-rectangle of the source image.
// All values in [0, 1]. X+W ≤ 1 and Y+H ≤ 1 are enforced by the validator.
type CropSpec struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	W float64 `json:"w"`
	H float64 `json:"h"`
}

// KenBurnsSpec animates the viewport from one position/scale to another over
// the clip duration. Coordinates are the center of the viewport in normalized
// image space [0,1]. Scale 1.0 = fit-to-frame; values > 1 zoom in.
type KenBurnsSpec struct {
	From   KenBurnsKeyframe `json:"from"`
	To     KenBurnsKeyframe `json:"to"`
	Easing string           `json:"easing"` // "linear" | "ease-in-out"
}

// KenBurnsKeyframe is one endpoint of a Ken Burns animation.
type KenBurnsKeyframe struct {
	// Scale is the zoom factor relative to the clip's fit-to-frame baseline.
	// Range: 1.0–2.0.
	Scale float64 `json:"scale"`
	// X and Y are the viewport center in normalized image coordinates [0,1].
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// AdjustSpec holds per-channel adjustments layered on top of the named filter.
// Brightness: additive offset in [-1, +1] (0 = no change).
// Contrast and Saturation: multiplicative factor in [0.5, 2.0] (1.0 = no change).
type AdjustSpec struct {
	Brightness float64 `json:"brightness"` // [-1, +1], default 0
	Contrast   float64 `json:"contrast"`   // [0.5, 2.0], default 1
	Saturation float64 `json:"saturation"` // [0.5, 2.0], default 1
}

// TextSpec is one text overlay drawn on top of a clip for a time window.
type TextSpec struct {
	// Text is the display string. Max 200 chars. Never passed raw to FFmpeg;
	// written to a temp file and referenced via textfile=.
	Text string `json:"text"`

	// Font must be one of the fixed OFL font names defined in
	// web/src/modules/video/fonts.ts and backend/internal/video/render/fonts.go.
	// Valid: "Inter" | "Poppins" | "Montserrat" | "Playfair Display" |
	//        "Bebas Neue" | "Lobster"
	Font string `json:"font"`

	// Size is the font size in pixels at the output resolution. Range: 16–200.
	Size int `json:"size"`

	// Color is a CSS hex color string, e.g. "#ffffff". Alpha not supported here;
	// use Opacity instead.
	Color string `json:"color"`

	// Opacity is the text opacity in [0, 1]. Default 1.
	Opacity float64 `json:"opacity"`

	// X and Y are the text anchor position in normalized output coordinates
	// [0,1]. (0.5, 0.5) is center. Anchor is top-left of the bounding box.
	X float64 `json:"x"`
	Y float64 `json:"y"`

	// Start and End are times within the clip in seconds. Start ≥ 0,
	// End ≤ clip.Duration. End 0 is treated as clip.Duration.
	Start float64 `json:"start"`
	End   float64 `json:"end"`

	// Anim is the entry animation. "none" | "fade-in" | "slide-up".
	// The animation runs from Start to Start+AnimDuration.
	Anim string `json:"anim"`

	// AnimDuration is how long the entry animation takes in seconds. Range: 0.1–1.0.
	AnimDuration float64 `json:"anim_duration"`
}

// TransitionSpec describes the transition from one clip into the next.
type TransitionSpec struct {
	// Type is one of: "fade" | "slide-left" | "slide-right" | "zoom-in" |
	//                 "zoom-out" | "dissolve"
	Type string `json:"type"`

	// Duration in seconds. Range: 0.3–1.5.
	// MUST satisfy: duration ≤ 0.5 × min(this clip duration, next clip duration).
	// Enforced by validator and by the UI slider constraint.
	Duration float64 `json:"duration"`
}

// ---- audio -----------------------------------------------------------------

// AudioSpec describes the single global audio track.
type AudioSpec struct {
	// Key is the S3 key of the audio file, either under
	// video/{user_id}/audio/ (user upload) or video/presets/ (global preset).
	// Validated server-side.
	Key string `json:"key"`

	// TrimStart is the offset into the audio file to begin playback, in seconds.
	// Default 0.
	TrimStart float64 `json:"trim_start"`

	// TrimEnd is the offset at which to stop reading the audio file, in seconds.
	// 0 means "read to the end of the file". Must be > TrimStart if nonzero.
	TrimEnd float64 `json:"trim_end"`

	// Volume is a global gain multiplier in [0, 2.0]. Default 1.0.
	Volume float64 `json:"volume"`

	// FadeIn duration in seconds applied from the start of playback. 0 = no fade.
	FadeIn float64 `json:"fade_in"`

	// FadeOut duration in seconds applied before the end of the video. 0 = no fade.
	// Applied relative to video end, not audio file end.
	FadeOut float64 `json:"fade_out"`

	// Title is the friendly display title of the audio track.
	Title string `json:"title,omitempty"`

	// Duration is the full duration in seconds of the source audio file.
	Duration float64 `json:"duration,omitempty"`
}

// ---- defaults --------------------------------------------------------------

// DefaultSpec returns a VideoSpec populated with safe defaults for a new
// project. The caller should append Clips as the user adds images.
func DefaultSpec() VideoSpec {
	return VideoSpec{
		Version: SpecVersion,
		Output: OutputSpec{
			Aspect: "9:16",
			Fit:    "cover",
			FPS:    30,
		},
		Clips: []ClipSpec{},
	}
}

// DefaultClip returns a ClipSpec with safe defaults for a newly added image.
func DefaultClip(imageKey string) ClipSpec {
	return ClipSpec{
		ImageKey: imageKey,
		Duration: 3.0,
		Crop:     CropSpec{X: 0, Y: 0, W: 1, H: 1},
		Rotate:   0,
		Filter:   "none",
		Adjust: AdjustSpec{
			Brightness: 0,
			Contrast:   1.0,
			Saturation: 1.0,
		},
	}
}
