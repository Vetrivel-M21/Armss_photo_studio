package compositor

import (
	"image"
	"image/color"
	"image/draw"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
)

// drawText renders text left-aligned, vertically centered at y, in the given
// color and point size, using the bundled brand font (spec §6: "Bundle the
// brand font in the repo; do not rely on system fonts").
func drawText(dst draw.Image, f *truetype.Font, text string, x, yCenter int, size float64, c color.Color) error {
	ctx := freetype.NewContext()
	ctx.SetDPI(72)
	ctx.SetFont(f)
	ctx.SetFontSize(size)
	ctx.SetClip(dst.Bounds())
	ctx.SetDst(dst)
	ctx.SetSrc(image.NewUniform(c))
	ctx.SetHinting(font.HintingFull)

	// Baseline sits roughly a third of the em size below vertical center.
	baseline := yCenter + int(size/3)
	pt := freetype.Pt(x, baseline)
	_, err := ctx.DrawString(text, pt)
	return err
}
