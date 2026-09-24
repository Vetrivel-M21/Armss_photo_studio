package compositor

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	"image/png"
	"os"

	"github.com/disintegration/imaging"
	"github.com/golang/freetype/truetype"

	"armss-poster-studio/internal/agents"
)

type Config struct {
	NavyHex       string
	ContactString string
	FontPath      string
}

// Compositor overlays brand assets into an image's reserved zones. It parses
// the bundled font once at construction, not per request.
type Compositor struct {
	cfg  Config
	font *truetype.Font
	navy color.Color
}

func New(cfg Config) (*Compositor, error) {
	fontBytes, err := os.ReadFile(cfg.FontPath)
	if err != nil {
		return nil, fmt.Errorf("read bundled font %s: %w", cfg.FontPath, err)
	}
	f, err := truetype.Parse(fontBytes)
	if err != nil {
		return nil, fmt.Errorf("parse bundled font: %w", err)
	}
	navy, err := parseHexColor(cfg.NavyHex)
	if err != nil {
		return nil, fmt.Errorf("parse navy color: %w", err)
	}
	return &Compositor{cfg: cfg, font: f, navy: navy}, nil
}

// Overrides is keyed by zone id ("logo" | "tagline" | "footer").
type Overrides map[string]ZoneOverride

// Composite draws footer, then tagline, then logo (footer full-bleed first,
// so any accidental overlap is harmless) onto baseImage's reserved zones.
// footerText is per-project (spec: an ARMSS-fixed contact string doesn't make
// sense once posters are made for other organizations) — an empty footerText
// skips the footer bar entirely rather than drawing a default.
func (c *Compositor) Composite(baseImage []byte, zones []agents.Zone, overrides Overrides, assets BrandAssetSet, footerText string) ([]byte, error) {
	base, _, err := image.Decode(bytes.NewReader(baseImage))
	if err != nil {
		return nil, fmt.Errorf("decode base image: %w", err)
	}
	canvas := imaging.Clone(base)

	byID := make(map[string]agents.Zone, len(zones))
	for _, z := range zones {
		byID[z.ID] = z
	}

	if z, ok := byID[agents.ZoneFooter]; ok && footerText != "" {
		z = ApplyOverride(z, overrides[agents.ZoneFooter])
		if err := c.drawFooter(canvas, z, assets.FooterIcon, footerText); err != nil {
			return nil, fmt.Errorf("draw footer: %w", err)
		}
	}
	if z, ok := byID[agents.ZoneTagline]; ok && len(assets.Tagline) > 0 {
		z = ApplyOverride(z, overrides[agents.ZoneTagline])
		if err := c.drawAsset(canvas, z, assets.Tagline, false); err != nil {
			return nil, fmt.Errorf("draw tagline: %w", err)
		}
	}
	if z, ok := byID[agents.ZoneLogo]; ok && len(assets.Logo) > 0 {
		z = ApplyOverride(z, overrides[agents.ZoneLogo])
		if err := c.drawAsset(canvas, z, assets.Logo, true); err != nil {
			return nil, fmt.Errorf("draw logo: %w", err)
		}
	}

	out := new(bytes.Buffer)
	if err := png.Encode(out, canvas); err != nil {
		return nil, fmt.Errorf("encode composited image: %w", err)
	}
	return out.Bytes(), nil
}

// drawFooter fills the zone with a solid navy rect, then places the phone
// icon (if uploaded) and the given contact string using the bundled font
// (spec §6). Caller guarantees contactString is non-empty.
func (c *Compositor) drawFooter(canvas *image.NRGBA, z agents.Zone, icon []byte, contactString string) error {
	rect := ZoneToPixelRect(z, canvas.Bounds().Dx(), canvas.Bounds().Dy())
	draw.Draw(canvas, rect, image.NewUniform(c.navy), image.Point{}, draw.Src)

	barHeight := rect.Dy()
	padding := barHeight / 6
	textX := rect.Min.X + padding
	centerY := rect.Min.Y + barHeight/2

	if len(icon) > 0 {
		iconImg, _, err := image.Decode(bytes.NewReader(icon))
		if err == nil {
			iconSize := barHeight - 2*padding
			resizedIcon := imaging.Fit(iconImg, iconSize, iconSize, imaging.Lanczos)
			iconRect := image.Rect(0, 0, resizedIcon.Bounds().Dx(), resizedIcon.Bounds().Dy())
			iconPos := image.Pt(rect.Min.X+padding, centerY-iconRect.Dy()/2)
			draw.Draw(canvas, iconRect.Add(iconPos), resizedIcon, image.Point{}, draw.Over)
			textX = iconPos.X + iconRect.Dx() + padding
		}
	}

	fontSize := float64(barHeight) * 0.32
	if err := drawText(canvas, c.font, contactString, textX, centerY, fontSize, color.White); err != nil {
		return fmt.Errorf("draw contact string: %w", err)
	}
	return nil
}

// drawAsset scales assetPNG to fit inside the zone preserving aspect ratio,
// centers it, and — for the logo only — adds a soft drop shadow so it reads
// against busy artwork (spec §6).
func (c *Compositor) drawAsset(canvas *image.NRGBA, z agents.Zone, assetPNG []byte, withShadow bool) error {
	asset, _, err := image.Decode(bytes.NewReader(assetPNG))
	if err != nil {
		return fmt.Errorf("decode asset: %w", err)
	}

	rect := ZoneToPixelRect(z, canvas.Bounds().Dx(), canvas.Bounds().Dy())
	fitted := imaging.Fit(asset, rect.Dx(), rect.Dy(), imaging.Lanczos)
	pos := image.Pt(
		rect.Min.X+(rect.Dx()-fitted.Bounds().Dx())/2,
		rect.Min.Y+(rect.Dy()-fitted.Bounds().Dy())/2,
	)

	if withShadow {
		shadow := imaging.Blur(imaging.AdjustFunc(fitted, func(c color.NRGBA) color.NRGBA {
			return color.NRGBA{R: 0, G: 0, B: 0, A: c.A / 2}
		}), 6)
		shadowPos := pos.Add(image.Pt(3, 4))
		draw.Draw(canvas, shadow.Bounds().Add(shadowPos), shadow, image.Point{}, draw.Over)
	}

	draw.Draw(canvas, fitted.Bounds().Add(pos), fitted, image.Point{}, draw.Over)
	return nil
}

func parseHexColor(hex string) (color.Color, error) {
	hex = trimHashPrefix(hex)
	if len(hex) != 6 {
		return nil, fmt.Errorf("expected a 6-digit hex color, got %q", hex)
	}
	var r, g, b uint8
	if _, err := fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b); err != nil {
		return nil, fmt.Errorf("invalid hex color %q: %w", hex, err)
	}
	return color.RGBA{R: r, G: g, B: b, A: 255}, nil
}

func trimHashPrefix(s string) string {
	if len(s) > 0 && s[0] == '#' {
		return s[1:]
	}
	return s
}
