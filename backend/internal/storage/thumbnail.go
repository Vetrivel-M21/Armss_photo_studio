package storage

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	"net/http"

	"github.com/disintegration/imaging"
	_ "golang.org/x/image/webp"
)

// AllowedContentTypes are the only image types accepted for uploads (spec §13):
// validated by content sniffing, not file extension.
var AllowedContentTypes = map[string]bool{
	"image/png":  true,
	"image/jpeg": true,
	"image/webp": true,
}

// SniffContentType detects the real content type of image data, ignoring any
// client-supplied filename/extension.
func SniffContentType(data []byte) string {
	return http.DetectContentType(data)
}

// DownscaleLongEdge resizes an image so its longest edge is at most maxPixels,
// preserving aspect ratio. Images already smaller are returned unchanged.
// Full-resolution print files waste vision-agent tokens for no accuracy gain (spec §13/§17).
func DownscaleLongEdge(data []byte, maxPixels int) ([]byte, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("decode image: %w", err)
	}
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w <= maxPixels && h <= maxPixels {
		return data, nil
	}

	var resized image.Image
	if w >= h {
		resized = imaging.Resize(img, maxPixels, 0, imaging.Lanczos)
	} else {
		resized = imaging.Resize(img, 0, maxPixels, imaging.Lanczos)
	}

	out := new(bytes.Buffer)
	if err := png.Encode(out, resized); err != nil {
		return nil, fmt.Errorf("encode resized image: %w", err)
	}
	return out.Bytes(), nil
}
