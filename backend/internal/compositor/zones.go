// Package compositor overlays the real ARMSS brand assets (logo, tagline,
// footer) into the reserved zones the Art Director agent left clean in the
// generated artwork. This is the mechanism that keeps the logo pixel-exact —
// spec §6/§17's central rule: the image model never draws it.
package compositor

import (
	"image"

	"armss-poster-studio/internal/agents"
)

// ZoneToPixelRect converts a normalized (0-1) Zone into a pixel rectangle for
// an image of the given dimensions. Pure and table-tested — the geometry the
// rest of the compositor builds on.
func ZoneToPixelRect(z agents.Zone, imgW, imgH int) image.Rectangle {
	x0 := int(z.X * float64(imgW))
	y0 := int(z.Y * float64(imgH))
	x1 := x0 + int(z.W*float64(imgW))
	y1 := y0 + int(z.H*float64(imgH))
	return image.Rect(x0, y0, x1, y1)
}

// ZoneOverride is a partial nudge to one zone's normalized rect — nil fields
// are left unchanged. Used when the artwork crowds a reserved zone (spec §6:
// "the user must be able to... nudge the logo zone if the artwork crowds it").
type ZoneOverride struct {
	X, Y, W, H *float64
}

// ApplyOverride merges a partial override into a zone, clamping every
// resulting coordinate to [0,1] so a bad nudge can't push the asset off-canvas.
func ApplyOverride(z agents.Zone, o ZoneOverride) agents.Zone {
	if o.X != nil {
		z.X = clamp01(*o.X)
	}
	if o.Y != nil {
		z.Y = clamp01(*o.Y)
	}
	if o.W != nil {
		z.W = clamp01(*o.W)
	}
	if o.H != nil {
		z.H = clamp01(*o.H)
	}
	return z
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
