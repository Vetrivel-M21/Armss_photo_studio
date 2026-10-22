// Package validate contains server-side validation for video specs and uploaded
// media. All limits are read from config so they can be changed via env vars
// without code changes.
//
// IMPORTANT: The server validates the full spec on every create, update, and
// render request. The client is never trusted.
package validate

import (
	"fmt"
	"math"
	"regexp"
	"strings"

	"armss-poster-studio/internal/video/render"
)

// Limits holds the configurable hard limits for spec validation.
// Populate from video.Config at startup.
type Limits struct {
	MaxClips          int
	MaxTotalDurationS float64
	MinClipDurationS  float64
	MaxClipDurationS  float64
	MaxTextsPerClip   int
	MaxTextLength     int
	MaxProjectsPerUser int
}

// DefaultLimits returns the defaults supporting up to 30-minute videos and 500 clips.
func DefaultLimits() Limits {
	return Limits{
		MaxClips:           500,
		MaxTotalDurationS:  1800,
		MinClipDurationS:   1,
		MaxClipDurationS:   1800,
		MaxTextsPerClip:    5,
		MaxTextLength:      200,
		MaxProjectsPerUser: 100,
	}
}

// ---- Spec validation -------------------------------------------------------

// ValidationError is a list of human-readable validation failures.
type ValidationError struct {
	Fields []string
}

func (e *ValidationError) Error() string {
	return "spec validation failed: " + strings.Join(e.Fields, "; ")
}

func (e *ValidationError) add(msg string, args ...any) {
	e.Fields = append(e.Fields, fmt.Sprintf(msg, args...))
}

func (e *ValidationError) hasErrors() bool { return len(e.Fields) > 0 }

// Spec validates a VideoSpec against the given limits and the rules defined
// in plan §7. Returns nil if valid, *ValidationError otherwise.
//
// userPrefix is the S3 key prefix the requesting user owns, e.g.
// "video/abc-123-user-id/". Every image_key and audio key must start with
// this prefix OR with the global preset prefix "video/presets/".
func Spec(spec render.VideoSpec, limits Limits, userPrefix string) error {
	ve := &ValidationError{}

	// ---- version -----------------------------------------------------------
	if spec.Version != render.SpecVersion {
		ve.add("unsupported spec version %d (current: %d)", spec.Version, render.SpecVersion)
		return ve // can't validate further if version is wrong
	}

	// ---- output ------------------------------------------------------------
	validateOutput(spec.Output, ve)

	// ---- clips -------------------------------------------------------------
	if len(spec.Clips) == 0 {
		ve.add("clips: at least one clip is required")
	}
	if len(spec.Clips) > limits.MaxClips {
		ve.add("clips: max %d clips allowed, got %d", limits.MaxClips, len(spec.Clips))
	}

	seenIDs := make(map[string]bool, len(spec.Clips))
	for i, c := range spec.Clips {
		prefix := fmt.Sprintf("clips[%d]", i)
		validateClip(c, i, spec.Clips, limits, userPrefix, seenIDs, prefix, ve)
	}

	// ---- total duration ----------------------------------------------------
	if len(spec.Clips) > 0 {
		total := render.TotalDuration(spec.Clips)
		if total <= 0 {
			ve.add("total duration must be positive (got %.3fs — transitions may be too long)", total)
		}
		if total > limits.MaxTotalDurationS {
			ve.add("total duration %.1fs exceeds max %.0fs", total, limits.MaxTotalDurationS)
		}
	}

	// ---- audio -------------------------------------------------------------
	if spec.Audio != nil {
		validateAudio(*spec.Audio, userPrefix, ve)
	}

	if ve.hasErrors() {
		return ve
	}
	return nil
}

// ---- internal validators ---------------------------------------------------

var validAspects = map[string]bool{"9:16": true, "1:1": true, "16:9": true}
var validFits = map[string]bool{"cover": true, "contain": true}
var validFPS = map[int]bool{24: true, 30: true}

func validateOutput(o render.OutputSpec, ve *ValidationError) {
	if !validAspects[o.Aspect] {
		ve.add("output.aspect: must be one of 9:16, 1:1, 16:9; got %q", o.Aspect)
	}
	if !validFits[o.Fit] {
		ve.add("output.fit: must be cover or contain; got %q", o.Fit)
	}
	if !validFPS[o.FPS] {
		ve.add("output.fps: must be 24 or 30; got %d", o.FPS)
	}
}

var validFilters = map[string]bool{
	"none": true, "vivid": true, "warm": true,
	"cold": true, "bw": true, "vintage": true,
}
var validTransitions = map[string]bool{
	"fade": true, "slide-left": true, "slide-right": true,
	"zoom-in": true, "zoom-out": true, "dissolve": true,
}
var validEasings = map[string]bool{"linear": true, "ease-in-out": true}
var validAnims = map[string]bool{"none": true, "fade-in": true, "slide-up": true}
var validRotations = map[int]bool{0: true, 90: true, 180: true, 270: true}
var hexColorRE = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

func validateClip(
	c render.ClipSpec,
	i int,
	all []render.ClipSpec,
	lim Limits,
	userPrefix string,
	seenIDs map[string]bool,
	prefix string,
	ve *ValidationError,
) {
	// ID
	if c.ID == "" {
		ve.add("%s.id: must not be empty", prefix)
	} else if seenIDs[c.ID] {
		ve.add("%s.id: duplicate clip id %q", prefix, c.ID)
	} else {
		seenIDs[c.ID] = true
	}

	// image_key ownership
	if !isOwnedKey(c.ImageKey, userPrefix) {
		ve.add("%s.image_key: key %q does not belong to the requesting user", prefix, c.ImageKey)
	}

	// duration
	if c.Duration < lim.MinClipDurationS || c.Duration > lim.MaxClipDurationS {
		ve.add("%s.duration: must be in [%.0f, %.0f], got %.2f",
			prefix, lim.MinClipDurationS, lim.MaxClipDurationS, c.Duration)
	}

	// crop
	if err := validateCrop(c.Crop); err != nil {
		ve.add("%s.crop: %s", prefix, err)
	}

	// rotate
	if !validRotations[c.Rotate] {
		ve.add("%s.rotate: must be 0, 90, 180, or 270; got %d", prefix, c.Rotate)
	}

	// ken_burns
	if c.KenBurns != nil {
		validateKenBurns(*c.KenBurns, prefix, ve)
	}

	// filter
	if !validFilters[c.Filter] {
		ve.add("%s.filter: unknown filter %q", prefix, c.Filter)
	}

	// adjust
	validateAdjust(c.Adjust, prefix, ve)

	// texts
	if len(c.Texts) > lim.MaxTextsPerClip {
		ve.add("%s.texts: max %d overlays, got %d", prefix, lim.MaxTextsPerClip, len(c.Texts))
	}
	for j, t := range c.Texts {
		validateText(t, c.Duration, lim, fmt.Sprintf("%s.texts[%d]", prefix, j), ve)
	}

	// transition_out
	if c.TransitionOut != nil {
		if i == len(all)-1 {
			ve.add("%s.transition_out: last clip must not have a transition", prefix)
		} else {
			validateTransition(*c.TransitionOut, all, i, prefix, ve)
		}
	}
}

func validateCrop(cr render.CropSpec) error {
	if cr.X < 0 || cr.Y < 0 || cr.W <= 0 || cr.H <= 0 {
		return fmt.Errorf("x,y must be ≥ 0; w,h must be > 0")
	}
	if cr.X+cr.W > 1.0+1e-9 {
		return fmt.Errorf("x+w must be ≤ 1 (got %.4f)", cr.X+cr.W)
	}
	if cr.Y+cr.H > 1.0+1e-9 {
		return fmt.Errorf("y+h must be ≤ 1 (got %.4f)", cr.Y+cr.H)
	}
	return nil
}

func validateKenBurns(kb render.KenBurnsSpec, prefix string, ve *ValidationError) {
	for _, frame := range []struct {
		name string
		kf   render.KenBurnsKeyframe
	}{{"from", kb.From}, {"to", kb.To}} {
		if frame.kf.Scale < 1.0 || frame.kf.Scale > 2.0 {
			ve.add("%s.ken_burns.%s.scale: must be in [1.0, 2.0], got %.3f", prefix, frame.name, frame.kf.Scale)
		}
		if frame.kf.X < 0 || frame.kf.X > 1 {
			ve.add("%s.ken_burns.%s.x: must be in [0,1], got %.3f", prefix, frame.name, frame.kf.X)
		}
		if frame.kf.Y < 0 || frame.kf.Y > 1 {
			ve.add("%s.ken_burns.%s.y: must be in [0,1], got %.3f", prefix, frame.name, frame.kf.Y)
		}
	}
	if !validEasings[kb.Easing] {
		ve.add("%s.ken_burns.easing: must be linear or ease-in-out; got %q", prefix, kb.Easing)
	}
}

func validateAdjust(adj render.AdjustSpec, prefix string, ve *ValidationError) {
	if adj.Brightness < -1 || adj.Brightness > 1 {
		ve.add("%s.adjust.brightness: must be in [-1,+1], got %.3f", prefix, adj.Brightness)
	}
	if adj.Contrast < 0.5 || adj.Contrast > 2.0 {
		ve.add("%s.adjust.contrast: must be in [0.5, 2.0], got %.3f", prefix, adj.Contrast)
	}
	if adj.Saturation < 0.5 || adj.Saturation > 2.0 {
		ve.add("%s.adjust.saturation: must be in [0.5, 2.0], got %.3f", prefix, adj.Saturation)
	}
}

func validateText(t render.TextSpec, clipDuration float64, lim Limits, prefix string, ve *ValidationError) {
	if len([]rune(t.Text)) > lim.MaxTextLength {
		ve.add("%s.text: max %d chars, got %d", prefix, lim.MaxTextLength, len([]rune(t.Text)))
	}
	if _, ok := render.AllowedFontSet[t.Font]; !ok {
		ve.add("%s.font: unknown font %q", prefix, t.Font)
	}
	if t.Size < 16 || t.Size > 200 {
		ve.add("%s.size: must be in [16, 200], got %d", prefix, t.Size)
	}
	if !hexColorRE.MatchString(t.Color) {
		ve.add("%s.color: must be a 6-digit hex like #ffffff, got %q", prefix, t.Color)
	}
	if t.Opacity < 0 || t.Opacity > 1 {
		ve.add("%s.opacity: must be in [0,1], got %.3f", prefix, t.Opacity)
	}
	if t.X < 0 || t.X > 1 {
		ve.add("%s.x: must be in [0,1], got %.3f", prefix, t.X)
	}
	if t.Y < 0 || t.Y > 1 {
		ve.add("%s.y: must be in [0,1], got %.3f", prefix, t.Y)
	}
	if t.Start < 0 || t.Start >= clipDuration {
		ve.add("%s.start: must be in [0, clip.duration), got %.3f", prefix, t.Start)
	}
	if t.End != 0 && t.End <= t.Start {
		ve.add("%s.end: must be > start when nonzero; start=%.3f end=%.3f", prefix, t.Start, t.End)
	}
	if t.End != 0 && t.End > clipDuration {
		ve.add("%s.end: %.3f exceeds clip duration %.3f", prefix, t.End, clipDuration)
	}
	if !validAnims[t.Anim] {
		ve.add("%s.anim: unknown animation %q", prefix, t.Anim)
	}
	if t.AnimDuration < 0.1 || t.AnimDuration > 1.0 {
		ve.add("%s.anim_duration: must be in [0.1, 1.0], got %.3f", prefix, t.AnimDuration)
	}
}

func validateTransition(tr render.TransitionSpec, clips []render.ClipSpec, i int, prefix string, ve *ValidationError) {
	if !validTransitions[tr.Type] {
		ve.add("%s.transition_out.type: unknown transition %q", prefix, tr.Type)
	}
	if tr.Duration < 0.3 || tr.Duration > 1.5 {
		ve.add("%s.transition_out.duration: must be in [0.3, 1.5], got %.3f", prefix, tr.Duration)
	}
	// 0.5 × min(adjacent clip durations) rule
	ok, maxAllowed := render.ValidateTransitionDuration(clips, i, tr.Duration)
	if !ok {
		ve.add(
			"%s.transition_out.duration: %.3fs exceeds allowed max %.3fs (0.5 × min adjacent clip durations)",
			prefix, tr.Duration, maxAllowed,
		)
	}
}

func validateAudio(a render.AudioSpec, userPrefix string, ve *ValidationError) {
	if !isOwnedKey(a.Key, userPrefix) {
		ve.add("audio.key: %q does not belong to the requesting user", a.Key)
	}
	if a.TrimStart < 0 {
		ve.add("audio.trim_start: must be ≥ 0, got %.3f", a.TrimStart)
	}
	if a.TrimEnd != 0 && a.TrimEnd <= a.TrimStart {
		ve.add("audio.trim_end: must be > trim_start when nonzero")
	}
	if a.Volume < 0 || a.Volume > 2.0 {
		ve.add("audio.volume: must be in [0, 2.0], got %.3f", a.Volume)
	}
	if a.FadeIn < 0 {
		ve.add("audio.fade_in: must be ≥ 0, got %.3f", a.FadeIn)
	}
	if a.FadeOut < 0 {
		ve.add("audio.fade_out: must be ≥ 0, got %.3f", a.FadeOut)
	}
}

// isOwnedKey returns true if the S3 key belongs to the user (starts with
// userPrefix) OR is a global preset (starts with "video/presets/").
func isOwnedKey(key, userPrefix string) bool {
	if key == "" {
		return false
	}
	if strings.HasPrefix(key, userPrefix) || strings.HasPrefix(key, "video/presets/") || strings.HasPrefix(key, "audio/presets/") {
		return true
	}
	uid := strings.TrimPrefix(userPrefix, "video/")
	uid = strings.Trim(uid, "/")
	if uid != "" {
		if strings.HasPrefix(key, "uploads/"+uid+"/") || strings.HasPrefix(key, "renders/"+uid+"/") {
			return true
		}
	}
	return false
}

// ---- Resolution validation -------------------------------------------------

var validResolutions = map[string]bool{"720p": true, "1080p": true}

// Resolution validates a resolution preset string.
func Resolution(res string) error {
	if !validResolutions[res] {
		return fmt.Errorf("resolution must be 720p or 1080p, got %q", res)
	}
	return nil
}

// ---- Timing sanity ---------------------------------------------------------

// TotalDurationPositive checks that the spec's total duration is > 0.
// (Also checked inside Spec, exposed separately for unit tests.)
func TotalDurationPositive(clips []render.ClipSpec) error {
	d := render.TotalDuration(clips)
	if d <= 0 || math.IsNaN(d) || math.IsInf(d, 0) {
		return fmt.Errorf("total duration must be > 0, got %.4f (transitions may be too long)", d)
	}
	return nil
}
