package compositor

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"

	"armss-poster-studio/internal/agents"
)

const testFontPath = "../../assets/fonts/Go-Regular.ttf"

func encodePNG(t *testing.T, img image.Image) []byte {
	t.Helper()
	buf := new(bytes.Buffer)
	if err := png.Encode(buf, img); err != nil {
		t.Fatalf("encode test png: %v", err)
	}
	return buf.Bytes()
}

func solidRGBA(w, h int, c color.Color) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, c)
		}
	}
	return img
}

func decodePNG(t *testing.T, data []byte) image.Image {
	t.Helper()
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("decode composited png: %v", err)
	}
	return img
}

func TestComposite_DrawsFooterBarInBrandNavy(t *testing.T) {
	c, err := New(Config{NavyHex: "#16265A", FontPath: testFontPath})
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	base := encodePNG(t, solidRGBA(1000, 1000, color.White))
	zones := []agents.Zone{
		{ID: agents.ZoneFooter, X: 0, Y: 0.9, W: 1.0, H: 0.1, Fill: "#16265A"},
	}

	out, err := c.Composite(base, zones, nil, BrandAssetSet{}, "ARMSS Employment Solutions")
	if err != nil {
		t.Fatalf("Composite() error: %v", err)
	}

	img := decodePNG(t, out)
	// The center of the floating pill at (500, 950) should be drawn with brand navy.
	r, g, b, _ := img.At(500, 950).RGBA()
	gotR, gotG, gotB := r>>8, g>>8, b>>8
	// Pill interior should have brand navy tone
	if gotR > 0x40 || gotB < 0x30 {
		t.Errorf("center of floating badge = rgb(%d,%d,%d), expected navy tone", gotR, gotG, gotB)
	}

	// Outside the pill at the edge of canvas (980, 950), artwork is preserved (white), not severed by a bar!
	rEdge, gEdge, bEdge, _ := img.At(980, 950).RGBA()
	if rEdge>>8 != 0xff || gEdge>>8 != 0xff || bEdge>>8 != 0xff {
		t.Errorf("artwork edge at (980,950) was cut off by a solid bar: rgb(%d,%d,%d)", rEdge>>8, gEdge>>8, bEdge>>8)
	}

	// Outside the footer zone completely (500, 100) must be untouched.
	r2, g2, b2, _ := img.At(500, 100).RGBA()
	if r2>>8 != 0xff || g2>>8 != 0xff || b2>>8 != 0xff {
		t.Errorf("area outside footer zone was modified: rgb(%d,%d,%d)", r2>>8, g2>>8, b2>>8)
	}
}

// TestComposite_EmptyFooterTextSkipsTheBar covers the per-project generalization:
// posters made for organizations other than ARMSS shouldn't get ARMSS's footer,
// so an empty footer text must skip the bar entirely, not just leave it textless.
func TestComposite_EmptyFooterTextSkipsTheBar(t *testing.T) {
	c, err := New(Config{NavyHex: "#16265A", FontPath: testFontPath})
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	base := encodePNG(t, solidRGBA(1000, 1000, color.White))
	zones := []agents.Zone{
		{ID: agents.ZoneFooter, X: 0, Y: 0.9, W: 1.0, H: 0.1, Fill: "#16265A"},
	}

	out, err := c.Composite(base, zones, nil, BrandAssetSet{}, "")
	if err != nil {
		t.Fatalf("Composite() error: %v", err)
	}

	img := decodePNG(t, out)
	r, g, b, _ := img.At(950, 950).RGBA()
	if r>>8 != 0xff || g>>8 != 0xff || b>>8 != 0xff {
		t.Errorf("footer bar was drawn despite empty footer text: rgb(%d,%d,%d), want white (untouched)", r>>8, g>>8, b>>8)
	}
}

func TestComposite_DrawsLogoInsideItsZone(t *testing.T) {
	c, err := New(Config{NavyHex: "#16265A", FontPath: testFontPath})
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	base := encodePNG(t, solidRGBA(1000, 1000, color.White))
	logo := encodePNG(t, solidRGBA(200, 200, color.RGBA{R: 255, A: 255})) // opaque red square
	zones := []agents.Zone{
		{ID: agents.ZoneLogo, X: 0.03, Y: 0.03, W: 0.2, H: 0.2, Fill: "#FFFFFF"},
	}

	out, err := c.Composite(base, zones, nil, BrandAssetSet{Logo: logo}, "")
	if err != nil {
		t.Fatalf("Composite() error: %v", err)
	}

	img := decodePNG(t, out)
	// Center of the logo zone: (0.03+0.1)*1000 = 130, same for y.
	r, g, b, _ := img.At(130, 130).RGBA()
	if r>>8 != 0xff || g>>8 != 0 || b>>8 != 0 {
		t.Errorf("logo zone center = rgb(%d,%d,%d), want opaque red", r>>8, g>>8, b>>8)
	}

	// Far outside the logo zone must remain the white background.
	r2, g2, b2, _ := img.At(900, 900).RGBA()
	if r2>>8 != 0xff || g2>>8 != 0xff || b2>>8 != 0xff {
		t.Errorf("area outside the logo zone was modified: rgb(%d,%d,%d)", r2>>8, g2>>8, b2>>8)
	}
}

func TestComposite_ZoneOverrideMovesTheLogo(t *testing.T) {
	c, err := New(Config{NavyHex: "#16265A", FontPath: testFontPath})
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	base := encodePNG(t, solidRGBA(1000, 1000, color.White))
	logo := encodePNG(t, solidRGBA(100, 100, color.RGBA{G: 255, A: 255})) // opaque green square
	zones := []agents.Zone{
		{ID: agents.ZoneLogo, X: 0.03, Y: 0.03, W: 0.1, H: 0.1, Fill: "#FFFFFF"},
	}
	overrides := Overrides{agents.ZoneLogo: ZoneOverride{X: f(0.7), Y: f(0.7)}}

	out, err := c.Composite(base, zones, overrides, BrandAssetSet{Logo: logo}, "")
	if err != nil {
		t.Fatalf("Composite() error: %v", err)
	}
	img := decodePNG(t, out)

	// Original position should now be untouched background.
	r, g, b, _ := img.At(80, 80).RGBA()
	if r>>8 != 0xff || g>>8 != 0xff || b>>8 != 0xff {
		t.Errorf("original logo position should be unmodified after override, got rgb(%d,%d,%d)", r>>8, g>>8, b>>8)
	}

	// Nudged position (0.7*1000=700, plus half the 100px zone) should show the logo.
	r2, g2, b2, _ := img.At(750, 750).RGBA()
	if g2>>8 != 0xff || r2>>8 != 0 || b2>>8 != 0 {
		t.Errorf("nudged logo position = rgb(%d,%d,%d), want opaque green", r2>>8, g2>>8, b2>>8)
	}
}
