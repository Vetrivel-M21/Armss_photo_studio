// Package render — effects.go
// Effect math that must produce identical results in Go (this file) and
// TypeScript (web/src/modules/video/engine/effects.ts). Both files consume the
// same parameter types from spec.go / spec.ts.
//
// FFmpeg filter equivalents for each preset are documented inline so the
// FFmpeg builder (ffmpeg.go) can reference the same constants without
// reimplementing the math.
package render

import (
	"math"
	"strconv"
)

// ---- Ken Burns -------------------------------------------------------------

// KenBurnsFrame computes the viewport rectangle (in normalized image coords)
// for a clip at normalized time u ∈ [0,1].
//
// The viewport is a rect of size (1/scale × 1/scale) centered at (cx, cy).
// FFmpeg implements this by pre-scaling the image to 2–4× output size and
// then using the zoompan filter. This function returns the parameters that
// feed both the Canvas preview and the zoompan args.
//
// Returns: centerX, centerY, scale (all in [0,1] normalized image space).
func KenBurnsFrame(kb KenBurnsSpec, u float64) (cx, cy, scale float64) {
	cx = KenBurnsValue(kb.From.X, kb.To.X, u, kb.Easing)
	cy = KenBurnsValue(kb.From.Y, kb.To.Y, u, kb.Easing)
	scale = KenBurnsValue(kb.From.Scale, kb.To.Scale, u, kb.Easing)
	return
}

// KenBurnsToPixels converts normalized Ken Burns parameters to pixel-space
// values for a given output width and height. Returns the top-left corner
// (px, py) and the viewport size (pw, ph) in pixels within the source image
// of dimensions (imgW, imgH).
func KenBurnsToPixels(cx, cy, scale float64, outW, outH, imgW, imgH int) (px, py, pw, ph int) {
	// Viewport dimensions in normalized image space
	aspectOut := float64(outW) / float64(outH)
	aspectImg := float64(imgW) / float64(imgH)

	// viewW/viewH in image-normalized coords, at scale=1 the viewport covers
	// the whole image dimension that's the limiting axis.
	var viewW, viewH float64
	if aspectImg > aspectOut {
		// Image is wider: height is the limiting axis
		viewH = 1.0 / scale
		viewW = viewH * aspectOut / aspectImg
	} else {
		viewW = 1.0 / scale
		viewH = viewW * aspectImg / aspectOut
	}

	// Top-left corner in normalized image space, clamped so we don't go OOB
	x := cx - viewW/2
	y := cy - viewH/2
	x = math.Max(0, math.Min(x, 1-viewW))
	y = math.Max(0, math.Min(y, 1-viewH))

	pw = int(math.Round(viewW * float64(imgW)))
	ph = int(math.Round(viewH * float64(imgH)))
	px = int(math.Round(x * float64(imgW)))
	py = int(math.Round(y * float64(imgH)))
	return
}

// ---- Color filters ---------------------------------------------------------

// FilterParams describes the FFmpeg eq/colorbalance parameters for a named
// filter preset. The Canvas preview applies the same numeric parameters via
// CSS filters + a canvas pixel pass for operations CSS can't express.
type FilterParams struct {
	// eq filter params (all are multiplicative on the decoded pixel value)
	Brightness float64 // eq brightness [−1,+1], 0 = no change
	Contrast   float64 // eq contrast [0.5,2.0], 1.0 = no change
	Saturation float64 // eq saturation [0,3.0], 1.0 = no change
	Gamma      float64 // eq gamma [0.1,10.0], 1.0 = no change

	// colorbalance: shadow/mid/highlight shift per channel in [−1,+1], 0 = no change
	RedShadow, GreenShadow, BlueShadow       float64
	RedMid, GreenMid, BlueMid               float64
	RedHigh, GreenHigh, BlueHigh           float64
}

// FilterPresets maps filter name → exact FFmpeg-equivalent parameters.
// "none" returns zero-value FilterParams (identity transform).
// These values are the source of truth; the TS side mirrors them in effects.ts.
var FilterPresets = map[string]FilterParams{
	"none": {
		Contrast: 1, Saturation: 1, Gamma: 1,
	},
	"vivid": {
		Contrast:   1.15,
		Saturation: 1.4,
		Gamma:      0.95,
	},
	"warm": {
		Contrast:   1.0,
		Saturation: 1.1,
		Gamma:      1.0,
		RedHigh:    0.08,
		GreenHigh:  0.02,
		BlueHigh:   -0.06,
		RedMid:     0.04,
		BlueMid:    -0.04,
	},
	"cold": {
		Contrast:   1.0,
		Saturation: 1.0,
		Gamma:      1.0,
		BlueHigh:   0.10,
		GreenHigh:  0.03,
		RedHigh:    -0.05,
		BlueMid:    0.05,
		RedMid:     -0.03,
	},
	"bw": {
		// hue=s=0 in FFmpeg — we express it as saturation=0
		Saturation: 0,
		Contrast:   1.05,
		Gamma:      1.0,
	},
	"vintage": {
		// Approximates a curves lift in shadows + warm cast + faded blacks
		Brightness: 0.04,
		Contrast:   0.90,
		Saturation: 0.80,
		Gamma:      1.08,
		RedShadow:  0.06,
		GreenShadow: -0.02,
		BlueShadow:  -0.04,
		RedHigh:    0.06,
		BlueHigh:   -0.08,
	},
}

// ResolveFilter returns the combined FilterParams for a clip, merging the named
// preset with the per-clip AdjustSpec. AdjustSpec values are applied on top:
//   - Brightness adds to the preset's Brightness
//   - Contrast and Saturation multiply the preset's values
func ResolveFilter(filter string, adj AdjustSpec) FilterParams {
	p, ok := FilterPresets[filter]
	if !ok {
		p = FilterPresets["none"]
	}
	// Ensure identity defaults if the preset left them zero
	if p.Contrast == 0 {
		p.Contrast = 1
	}
	if p.Saturation == 0 && filter != "bw" {
		p.Saturation = 1
	}
	if p.Gamma == 0 {
		p.Gamma = 1
	}

	p.Brightness += adj.Brightness
	p.Contrast *= adj.Contrast
	if adj.Saturation != 1 {
		p.Saturation *= adj.Saturation
	}

	// Clamp to FFmpeg-safe ranges
	p.Brightness = math.Max(-1, math.Min(1, p.Brightness))
	p.Contrast = math.Max(0.5, math.Min(2.0, p.Contrast))
	p.Saturation = math.Max(0, math.Min(3.0, p.Saturation))
	p.Gamma = math.Max(0.1, math.Min(10.0, p.Gamma))
	return p
}

// ---- Output resolution -----------------------------------------------------

// Resolution holds the pixel dimensions of the output video.
type Resolution struct {
	W, H int
}

// AspectResolution maps an aspect ratio name and a preset name to pixel
// dimensions. Only 720p and 1080p are supported in v1.
var AspectResolutions = map[string]map[string]Resolution{
	"9:16": {
		"720p":  {720, 1280},
		"1080p": {1080, 1920},
	},
	"1:1": {
		"720p":  {720, 720},
		"1080p": {1080, 1080},
	},
	"16:9": {
		"720p":  {1280, 720},
		"1080p": {1920, 1080},
	},
}

// OutputResolution returns the pixel Width and Height for a given aspect and
// preset. Returns an error string if the combination is not found.
func OutputResolution(aspect, preset string) (Resolution, bool) {
	byAspect, ok := AspectResolutions[aspect]
	if !ok {
		return Resolution{}, false
	}
	res, ok := byAspect[preset]
	return res, ok
}

// ---- Contain-with-blur layout ----------------------------------------------

// ContainLayout describes the two-layer FFmpeg filter graph for "contain" fit
// mode: a heavily blurred background fill + a sharp centered overlay.
//
// Background: scale the source image to fill the output (cover), then apply a
// heavy boxblur (radius ≈ outW/8) and a slight darkening overlay.
// Foreground: scale the source image to fit within the output (letterbox),
// centered on the blurred background.
//
// Returns blur radius and the foreground scale filter string.
func ContainLayout(outW, outH int) (blurRadius int, fgScaleFilter string) {
	blurRadius = outW / 8
	fgScaleFilter = "scale=iw*min(" +
		itoa(outW) + "/iw\\," +
		itoa(outH) + "/ih):ih*min(" +
		itoa(outW) + "/iw\\," +
		itoa(outH) + "/ih),pad=" +
		itoa(outW) + ":" +
		itoa(outH) + ":(ow-iw)/2:(oh-ih)/2"
	return
}

func itoa(n int) string {
	return strconv.Itoa(n)
}
