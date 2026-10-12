package compositor

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	"image/png"
	"math"
	"os"

	"github.com/disintegration/imaging"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"

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

// Composite draws footer, then tagline, then logo (footer first, so any accidental
// overlap is harmless) onto baseImage's reserved zones.
// An empty footerText skips the footer entirely rather than drawing a default.
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

// drawFooter places a floating capsule pill badge with the phone icon, contact numbers,
// and decorative side pinstripes centered over the bottom artwork. The background artwork
// continues naturally to the bottom edge without being severed by a solid rectangle.
func (c *Compositor) drawFooter(canvas *image.NRGBA, z agents.Zone, icon []byte, contactString string) error {
	rect := ZoneToPixelRect(z, canvas.Bounds().Dx(), canvas.Bounds().Dy())
	canvasW := canvas.Bounds().Dx()
	centerY := rect.Min.Y + rect.Dy()/2

	zoneH := rect.Dy()
	if zoneH <= 0 {
		return nil
	}
	pillH := int(float64(zoneH) * 0.72)
	if pillH < 32 {
		pillH = 32
	}
	if pillH > zoneH {
		pillH = zoneH
	}

	fontSize := float64(pillH) * 0.44
	textWidth := measureTextWidth(c.font, contactString, fontSize)

	iconDiameter := int(float64(pillH) * 0.75)
	circleR := iconDiameter / 2
	gap := int(float64(pillH) * 0.25)
	paddingH := int(float64(pillH) * 0.45)

	contentWidth := iconDiameter + gap + textWidth
	pillW := contentWidth + 2*paddingH
	if pillW > canvasW-40 {
		pillW = canvasW - 40
	}

	pillCX := canvasW / 2
	pillCY := centerY

	var navyNRGBA color.NRGBA
	if r, g, b, a := c.navy.RGBA(); a > 0 {
		navyNRGBA = color.NRGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8), A: 255}
	} else {
		navyNRGBA = color.NRGBA{R: 0x16, G: 0x26, B: 0x5A, A: 255}
	}

	// 1. Soft ambient shadow under the floating pill
	shadowColor := color.NRGBA{R: 0, G: 0, B: 0, A: 70}
	drawRoundedPill(canvas, pillCX, pillCY+3, pillW+4, pillH+4, shadowColor)

	// 2. Floating pill container in brand navy
	drawRoundedPill(canvas, pillCX, pillCY, pillW, pillH, navyNRGBA)

	// 3. White circular badge on the left inside the pill
	iconCX := pillCX - contentWidth/2 + circleR
	iconCY := pillCY
	whiteColor := color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	drawCircle(canvas, iconCX, iconCY, circleR, whiteColor)

	// 4. Phone icon inside the white circle
	if len(icon) > 0 {
		if iconImg, _, err := image.Decode(bytes.NewReader(icon)); err == nil {
			iconSize := iconDiameter - 6
			if iconSize > 0 {
				resizedIcon := imaging.Fit(iconImg, iconSize, iconSize, imaging.Lanczos)
				iconRect := image.Rect(0, 0, resizedIcon.Bounds().Dx(), resizedIcon.Bounds().Dy())
				iconPos := image.Pt(iconCX-resizedIcon.Bounds().Dx()/2, iconCY-resizedIcon.Bounds().Dy()/2)
				draw.Draw(canvas, iconRect.Add(iconPos), resizedIcon, image.Point{}, draw.Over)
			}
		}
	} else {
		drawPhoneGlyph(canvas, iconCX, iconCY, int(float64(circleR)*0.65), navyNRGBA)
	}

	// 5. Contact text in bold white
	textX := iconCX + circleR + gap
	if err := drawText(canvas, c.font, contactString, textX, pillCY, fontSize, whiteColor); err != nil {
		return fmt.Errorf("draw contact string: %w", err)
	}

	// 6. Decorative horizontal pinstripes flanking the pill (= [ pill ] =)
	lineLen := int(float64(canvasW) * 0.08)
	if lineLen > 90 {
		lineLen = 90
	}
	if lineLen < 30 {
		lineLen = 30
	}
	lineGap := int(float64(pillH) * 0.28)
	accentColor1 := color.NRGBA{R: 245, G: 166, B: 35, A: 220} // Saffron/Gold
	accentColor2 := color.NRGBA{R: 255, G: 255, B: 255, A: 200} // White

	leftEnd := pillCX - pillW/2 - lineGap
	leftStart := leftEnd - lineLen
	rightStart := pillCX + pillW/2 + lineGap
	rightEnd := rightStart + lineLen

	if leftStart > 10 {
		drawDecorativeLine(canvas, leftStart, leftEnd, pillCY-3, 2, accentColor1)
		drawDecorativeLine(canvas, leftStart, leftEnd, pillCY+3, 2, accentColor2)
	}
	if rightEnd < canvasW-10 {
		drawDecorativeLine(canvas, rightStart, rightEnd, pillCY-3, 2, accentColor1)
		drawDecorativeLine(canvas, rightStart, rightEnd, pillCY+3, 2, accentColor2)
	}

	return nil
}

// drawAsset scales assetPNG to fit inside the zone preserving aspect ratio,
// centers it, and adds a soft natural ambient shadow so it rests seamlessly on artwork.
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
			return color.NRGBA{R: 0, G: 0, B: 0, A: c.A / 6}
		}), 4)
		shadowPos := pos.Add(image.Pt(0, 2))
		draw.Draw(canvas, shadow.Bounds().Add(shadowPos), shadow, image.Point{}, draw.Over)
	}

	draw.Draw(canvas, fitted.Bounds().Add(pos), fitted, image.Point{}, draw.Over)
	return nil
}

func measureTextWidth(f *truetype.Font, text string, size float64) int {
	opts := truetype.Options{Size: size, DPI: 72}
	face := truetype.NewFace(f, &opts)
	defer face.Close()
	return font.MeasureString(face, text).Ceil()
}

func drawRoundedPill(dst *image.NRGBA, cx, cy, w, h int, fillColor color.NRGBA) {
	if w <= 0 || h <= 0 {
		return
	}
	r := float64(h) / 2.0
	straightHalfW := float64(w)/2.0 - r
	if straightHalfW < 0 {
		straightHalfW = 0
	}

	minX := int(float64(cx) - float64(w)/2.0 - 2)
	maxX := int(float64(cx) + float64(w)/2.0 + 2)
	minY := int(float64(cy) - r - 2)
	maxY := int(float64(cy) + r + 2)

	b := dst.Bounds()
	if minX < b.Min.X { minX = b.Min.X }
	if maxX > b.Max.X { maxX = b.Max.X }
	if minY < b.Min.Y { minY = b.Min.Y }
	if maxY > b.Max.Y { maxY = b.Max.Y }

	for y := minY; y < maxY; y++ {
		dy := float64(y) + 0.5 - float64(cy)
		for x := minX; x < maxX; x++ {
			px := float64(x) + 0.5 - float64(cx)
			clampedX := px
			if clampedX < -straightHalfW {
				clampedX = -straightHalfW
			} else if clampedX > straightHalfW {
				clampedX = straightHalfW
			}
			dist := math.Hypot(px-clampedX, dy)
			if dist <= r {
				alphaFactor := 1.0
				if dist > r-1.0 {
					alphaFactor = r - dist
				}
				blendPixel(dst, x, y, fillColor, alphaFactor)
			}
		}
	}
}

func drawCircle(dst *image.NRGBA, cx, cy, r int, fillColor color.NRGBA) {
	rF := float64(r)
	minX := cx - r - 2
	maxX := cx + r + 2
	minY := cy - r - 2
	maxY := cy + r + 2

	b := dst.Bounds()
	if minX < b.Min.X { minX = b.Min.X }
	if maxX > b.Max.X { maxX = b.Max.X }
	if minY < b.Min.Y { minY = b.Min.Y }
	if maxY > b.Max.Y { maxY = b.Max.Y }

	for y := minY; y < maxY; y++ {
		dy := float64(y) + 0.5 - float64(cy)
		for x := minX; x < maxX; x++ {
			dx := float64(x) + 0.5 - float64(cx)
			dist := math.Hypot(dx, dy)
			if dist <= rF {
				alphaFactor := 1.0
				if dist > rF-1.0 {
					alphaFactor = rF - dist
				}
				blendPixel(dst, x, y, fillColor, alphaFactor)
			}
		}
	}
}

func drawDecorativeLine(dst *image.NRGBA, x0, x1, y, thickness int, col color.NRGBA) {
	if x1 <= x0 {
		return
	}
	b := dst.Bounds()
	halfT := thickness / 2
	for lineY := y - halfT; lineY <= y+halfT; lineY++ {
		if lineY < b.Min.Y || lineY >= b.Max.Y {
			continue
		}
		for x := x0; x <= x1; x++ {
			if x >= b.Min.X && x < b.Max.X {
				blendPixel(dst, x, lineY, col, 1.0)
			}
		}
	}
}

func drawPhoneGlyph(dst *image.NRGBA, cx, cy, r int, col color.NRGBA) {
	for dy := -r; dy <= r; dy++ {
		for dx := -r; dx <= r; dx++ {
			fx := float64(dx) / float64(r)
			fy := float64(dy) / float64(r)
			rx := 0.7071 * (fx - fy)
			ry := 0.7071 * (fx + fy)
			inHandle := math.Abs(rx) < 0.22 && math.Abs(ry) < 0.65
			inEarpiece := math.Hypot(rx-0.15, ry+0.55) < 0.35
			inMouthpiece := math.Hypot(rx-0.15, ry-0.55) < 0.35
			if (inHandle || inEarpiece || inMouthpiece) && math.Hypot(fx, fy) <= 0.95 {
				blendPixel(dst, cx+dx, cy+dy, col, 1.0)
			}
		}
	}
}

func blendPixel(dst *image.NRGBA, x, y int, src color.NRGBA, factor float64) {
	if factor <= 0 {
		return
	}
	if factor > 1.0 {
		factor = 1.0
	}
	srcA := float64(src.A) * factor / 255.0
	if srcA <= 0 {
		return
	}
	idx := dst.PixOffset(x, y)
	dstR := float64(dst.Pix[idx+0])
	dstG := float64(dst.Pix[idx+1])
	dstB := float64(dst.Pix[idx+2])
	dstA := float64(dst.Pix[idx+3]) / 255.0

	outA := srcA + dstA*(1.0-srcA)
	if outA > 0 {
		outR := (float64(src.R)*srcA + dstR*dstA*(1.0-srcA)) / outA
		outG := (float64(src.G)*srcA + dstG*dstA*(1.0-srcA)) / outA
		outB := (float64(src.B)*srcA + dstB*dstA*(1.0-srcA)) / outA

		dst.Pix[idx+0] = uint8(math.Round(outR))
		dst.Pix[idx+1] = uint8(math.Round(outG))
		dst.Pix[idx+2] = uint8(math.Round(outB))
		dst.Pix[idx+3] = uint8(math.Round(outA * 255.0))
	}
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
