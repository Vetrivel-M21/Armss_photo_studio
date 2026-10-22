// Package render — timing.go
// Timing formulas that must produce identical results in Go (this file) and
// TypeScript (web/src/modules/video/engine/timing.ts). Both files are tested
// against the shared golden vectors in
// backend/internal/video/render/testdata/timing_vectors.json.
package render

import "math"

// ---- public API ------------------------------------------------------------

// TotalDuration returns the wall-clock length of the rendered video in seconds.
//
//	total = Σ clip.Duration − Σ transition.Duration
//
// The last clip has no outgoing transition and is excluded from the transition
// sum. A transition is "owned" by the clip it exits, so clips[i].TransitionOut
// shortens clips[i] and the gap before clips[i+1].
func TotalDuration(clips []ClipSpec) float64 {
	total := 0.0
	for _, c := range clips {
		total += c.Duration
		if c.TransitionOut != nil {
			total -= c.TransitionOut.Duration
		}
	}
	return math.Round(total*1000) / 1000 // round to millisecond
}

// ClipStartTime returns the absolute start time (within the video) of clips[i].
//
// clips[0] always starts at t=0.
// clips[i] starts after the non-overlapping portion of clips[i-1]:
//
//	start[i] = start[i-1] + clips[i-1].Duration − clips[i-1].TransitionOut.Duration
func ClipStartTime(clips []ClipSpec, i int) float64 {
	t := 0.0
	for j := 0; j < i; j++ {
		t += clips[j].Duration
		if clips[j].TransitionOut != nil {
			t -= clips[j].TransitionOut.Duration
		}
	}
	return math.Round(t*1000) / 1000
}

// TransitionStartTime returns the absolute start time of the transition at the
// end of clips[i]. Panics if clips[i].TransitionOut is nil.
//
//	transition_start[i] = clip_start[i] + clips[i].Duration − transition.Duration
func TransitionStartTime(clips []ClipSpec, i int) float64 {
	if clips[i].TransitionOut == nil {
		panic("TransitionStartTime called on clip with no outgoing transition")
	}
	t := ClipStartTime(clips, i) + clips[i].Duration - clips[i].TransitionOut.Duration
	return math.Round(t*1000) / 1000
}

// XFadeOffset returns the `offset` parameter for FFmpeg's xfade filter between
// clips[i] and clips[i+1]. This is the number of seconds into the first stream
// (clips[0..i] concatenated) at which the transition begins.
//
// It is identical to TransitionStartTime and is provided under this name to
// make the FFmpeg builder code self-documenting.
func XFadeOffset(clips []ClipSpec, i int) float64 {
	return TransitionStartTime(clips, i)
}

// ValidateTransitionDuration reports whether a transition duration satisfies
// the "≤ 0.5 × min(adjacent clip durations)" rule.
// Returns the maximum allowed value for convenience (used by the UI slider).
func ValidateTransitionDuration(clips []ClipSpec, i int, transitionDur float64) (ok bool, maxAllowed float64) {
	nextDur := clips[i+1].Duration
	thisDur := clips[i].Duration
	minAdj := math.Min(thisDur, nextDur)
	max := math.Round((0.5*minAdj)*1000) / 1000
	return transitionDur <= max, max
}

// ClipLocalTime converts an absolute video time t into a local time within
// clips[i], expressed as a fraction u ∈ [0,1] of the clip's duration.
// Returns -1 if t is outside the clip's active window (before start or after
// start+duration). This is used by the preview renderer to decide which clip
// to draw and at what animation position.
func ClipLocalTime(clips []ClipSpec, i int, t float64) float64 {
	start := ClipStartTime(clips, i)
	end := start + clips[i].Duration
	if t < start || t >= end {
		return -1
	}
	u := (t - start) / clips[i].Duration
	return clamp01(u)
}

// KenBurnsValue interpolates a single scalar Ken Burns parameter (scale, x, or
// y) at normalized time u ∈ [0,1], using the given easing function.
func KenBurnsValue(from, to float64, u float64, easing string) float64 {
	e := applyEasing(u, easing)
	return from + (to-from)*e
}

// TextAlpha returns the opacity of a text overlay at absolute video time t for
// the given clip start. Returns 0 if t is outside [start, end]; returns 1 when
// the entry animation is complete; returns a value in [0,1] during entry.
//
// Both "fade-in" and "slide-up" use the same alpha ramp (the position offset
// for slide-up is handled separately in the renderer).
func TextAlpha(clipStart float64, text TextSpec, t float64) float64 {
	absStart := clipStart + text.Start
	absEnd := clipStart + text.End
	if text.End == 0 { // sentinel: show until clip end
		// caller must pass the true clip end if needed; we don't have it here
		absEnd = math.MaxFloat64
	}
	if t < absStart || t >= absEnd {
		return 0
	}
	animEnd := absStart + text.AnimDuration
	if t >= animEnd {
		return 1
	}
	u := (t - absStart) / text.AnimDuration
	return clamp01(u)
}

// TextSlideOffsetNorm returns the normalized vertical slide offset (0 = at
// final position, 1 = fully offscreen below) for a "slide-up" animation.
// Use the result as a Y offset: finalY + offset * somePixelHeight.
func TextSlideOffsetNorm(clipStart float64, text TextSpec, t float64) float64 {
	absStart := clipStart + text.Start
	animEnd := absStart + text.AnimDuration
	if t <= absStart {
		return 1
	}
	if t >= animEnd {
		return 0
	}
	u := (t - absStart) / text.AnimDuration
	return 1 - clamp01(u)
}

// ---- helpers ---------------------------------------------------------------

func applyEasing(u float64, easing string) float64 {
	u = clamp01(u)
	switch easing {
	case "ease-in-out":
		// Cubic ease-in-out: 3u²-2u³
		return u * u * (3 - 2*u)
	default: // "linear"
		return u
	}
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
