package compositor

import (
	"image"
	"testing"

	"armss-poster-studio/internal/agents"
)

func f(v float64) *float64 { return &v }

func TestZoneToPixelRect(t *testing.T) {
	tests := []struct {
		name       string
		zone       agents.Zone
		imgW, imgH int
		want       image.Rectangle
	}{
		{
			name: "top-left logo zone on a square image",
			zone: agents.Zone{X: 0.03, Y: 0.03, W: 0.22, H: 0.09},
			imgW: 1000, imgH: 1000,
			want: image.Rect(30, 30, 250, 120),
		},
		{
			name: "full-width footer bar on a portrait image",
			zone: agents.Zone{X: 0, Y: 0.9, W: 1.0, H: 0.1},
			imgW: 800, imgH: 1000,
			want: image.Rect(0, 900, 800, 1000),
		},
		{
			name: "zero-size zone collapses to a point rect",
			zone: agents.Zone{X: 0.5, Y: 0.5, W: 0, H: 0},
			imgW: 400, imgH: 400,
			want: image.Rect(200, 200, 200, 200),
		},
		{
			name: "non-square image scales axes independently",
			zone: agents.Zone{X: 0.1, Y: 0.2, W: 0.5, H: 0.5},
			imgW: 2000, imgH: 500,
			want: image.Rect(200, 100, 1200, 350),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ZoneToPixelRect(tt.zone, tt.imgW, tt.imgH)
			if got != tt.want {
				t.Errorf("ZoneToPixelRect(%+v, %d, %d) = %v, want %v", tt.zone, tt.imgW, tt.imgH, got, tt.want)
			}
		})
	}
}

func TestApplyOverride(t *testing.T) {
	base := agents.Zone{ID: "logo", X: 0.1, Y: 0.1, W: 0.2, H: 0.1, Fill: "#000000"}

	tests := []struct {
		name     string
		override ZoneOverride
		want     agents.Zone
	}{
		{
			name:     "no fields set leaves the zone unchanged",
			override: ZoneOverride{},
			want:     base,
		},
		{
			name:     "nudging x only",
			override: ZoneOverride{X: f(0.5)},
			want:     agents.Zone{ID: "logo", X: 0.5, Y: 0.1, W: 0.2, H: 0.1, Fill: "#000000"},
		},
		{
			name:     "nudging all four axes",
			override: ZoneOverride{X: f(0.2), Y: f(0.3), W: f(0.4), H: f(0.15)},
			want:     agents.Zone{ID: "logo", X: 0.2, Y: 0.3, W: 0.4, H: 0.15, Fill: "#000000"},
		},
		{
			name:     "out-of-range values are clamped to [0,1]",
			override: ZoneOverride{X: f(-0.5), W: f(1.5)},
			want:     agents.Zone{ID: "logo", X: 0, Y: 0.1, W: 1, H: 0.1, Fill: "#000000"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ApplyOverride(base, tt.override)
			if got != tt.want {
				t.Errorf("ApplyOverride(%+v, %+v) = %+v, want %+v", base, tt.override, got, tt.want)
			}
		})
	}
}
