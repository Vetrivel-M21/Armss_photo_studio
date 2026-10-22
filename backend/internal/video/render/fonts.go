// Package render — fonts.go
// Fixed OFL-licensed font set shared by the frontend and the FFmpeg worker.
// Any change here MUST be mirrored in web/src/modules/video/fonts.ts.
package render

// AllowedFonts is the definitive list of font names that may appear in
// TextSpec.Font. Both the validator and the FFmpeg builder reject any name not
// in this list.
//
// These fonts are OFL-licensed Google Fonts. The .ttf files must be present at
// VIDEO_FONTS_DIR at runtime (bundled in the Docker image).
//
// Frontend: loaded via @font-face with the same file names (see fonts.ts).
// Backend: passed to FFmpeg drawtext via fontfile=<path>/<name>.ttf.
var AllowedFonts = []string{
	"Inter",
	"Poppins",
	"Montserrat",
	"Playfair Display",
	"Bebas Neue",
	"Lobster",
}

// AllowedFontSet is a fast-lookup version of AllowedFonts.
var AllowedFontSet = func() map[string]struct{} {
	m := make(map[string]struct{}, len(AllowedFonts))
	for _, f := range AllowedFonts {
		m[f] = struct{}{}
	}
	return m
}()

// FontFileName maps a logical font name to the expected .ttf filename under
// VIDEO_FONTS_DIR. Filename is the hyphenated Regular weight file distributed
// by Google Fonts.
var FontFileName = map[string]string{
	"Inter":            "Inter-Regular.ttf",
	"Poppins":          "Poppins-Regular.ttf",
	"Montserrat":       "Montserrat-Regular.ttf",
	"Playfair Display": "PlayfairDisplay-Regular.ttf",
	"Bebas Neue":       "BebasNeue-Regular.ttf",
	"Lobster":          "Lobster-Regular.ttf",
}
