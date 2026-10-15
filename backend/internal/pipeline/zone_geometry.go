package pipeline

import (
	"armss-poster-studio/internal/agents"
	"armss-poster-studio/internal/store/mysql"
)

// scaleWidth returns the normalised logo/slogan width for a given scale token.
// These values match the labels shown in the UI (compact=12%, standard=18%, prominent=25%).
func scaleWidth(scale string) float64 {
	switch scale {
	case "compact":
		return 0.12
	case "prominent":
		return 0.25
	default: // "standard" or anything unrecognised
		return 0.18
	}
}

// logoHeight returns the normalised height for the logo zone, proportional to
// width so the aspect ratio of the asset is broadly preserved.
func logoHeight(w float64) float64 {
	// Most ARMSS logo lockups are roughly 4:1 wide, so height ≈ w/4, clamped to
	// at least 6% so it doesn't disappear on tall canvases.
	h := w / 4.0
	if h < 0.06 {
		h = 0.06
	}
	return h
}

// sloganHeight returns the normalised height for the slogan/tagline zone.
// Tagline lockups tend to be wider and thinner than the logo (≈8:1).
func sloganHeight(w float64) float64 {
	h := w / 8.0
	if h < 0.04 {
		h = 0.04
	}
	return h
}

// applyZoneGeometry overwrites the X/Y/W/H of every logo, tagline, and footer
// zone in the Art Director's output using the deterministic coordinates derived
// from the project's stored ZoneHints. This is the canonical source of truth
// for zone placement; the Art Director's own coordinates are used only as a
// fallback when a zone ID is not covered by the hints.
func applyZoneGeometry(zones []agents.Zone, project *mysql.Project) {
	logoW := scaleWidth(project.LogoScale)
	logoH := logoHeight(logoW)

	sloganW := scaleWidth(project.SloganScale)
	sloganH := sloganHeight(sloganW)

	// Resolve logo anchor first so slogan's "below-logo" position can reference it.
	logoX, logoY := logoAnchor(project.LogoPosition, logoW, logoH)
	sloganX, sloganY := sloganAnchor(project.SloganPosition, sloganW, sloganH, logoX, logoY, logoH)

	for i := range zones {
		switch zones[i].ID {
		case agents.ZoneLogo:
			zones[i].X = logoX
			zones[i].Y = logoY
			zones[i].W = logoW
			zones[i].H = logoH

		case agents.ZoneTagline:
			zones[i].X = sloganX
			zones[i].Y = sloganY
			zones[i].W = sloganW
			zones[i].H = sloganH

		case agents.ZoneFooter:
			// Footer is always a full-width strip at the very bottom.
			zones[i].X = 0
			zones[i].Y = 0.91
			zones[i].W = 1.0
			zones[i].H = 0.09
		}
	}
}

// logoAnchor converts a logoPosition string into normalised top-left (X,Y)
// coordinates, clamped so the zone never bleeds off canvas.
func logoAnchor(position string, w, h float64) (x, y float64) {
	const pad = 0.04 // breathing room from the canvas edge
	switch position {
	case "top-left":
		return pad, pad
	case "top-right":
		return 1.0 - w - pad, pad
	case "bottom-left":
		return pad, 1.0 - h - pad
	case "bottom-right":
		return 1.0 - w - pad, 1.0 - h - pad
	default: // "top-center"
		return (1.0 - w) / 2.0, pad
	}
}

// sloganAnchor resolves the tagline zone position relative to the logo zone
// (for "below-logo") or independently for the other positions.
func sloganAnchor(position string, w, h, logoX, logoY, logoH float64) (x, y float64) {
	const pad = 0.015
	switch position {
	case "below-logo":
		// Horizontally centred under the logo zone; gap of pad.
		return logoX, logoY + logoH + pad
	case "above-footer":
		return (1.0 - w) / 2.0, 0.91 - h - pad
	case "top":
		return (1.0 - w) / 2.0, 0.04
	default: // "bottom"
		return (1.0 - w) / 2.0, 0.88 - h
	}
}

// zoneHintPromptText returns a human-readable paragraph that the Art Director
// agent receives as part of its user message. It translates the numeric zone
// geometry into photographic framing language so the image prompt composition
// naturally accommodates both assets without explicit percentage references.
func zoneHintPromptText(project *mysql.Project) string {
	logoW := scaleWidth(project.LogoScale)
	logoH := logoHeight(logoW)
	logoX, logoY := logoAnchor(project.LogoPosition, logoW, logoH)

	sloganW := scaleWidth(project.SloganScale)
	sloganH := sloganHeight(sloganW)
	sloganX, sloganY := sloganAnchor(project.SloganPosition, sloganW, sloganH, logoX, logoY, logoH)

	footerDesc := ""
	if project.FooterVisible {
		footerDesc = " The bottom 9% of the canvas is reserved for a full-width contact footer bar."
	}

	return "Poster canvas proportions: " + project.AspectRatio + ". " +
		"The vector corporate LOGO will be composited at normalised position " +
		formatZone(logoX, logoY, logoW, logoH) + " — " + logoPositionPhrase(project.LogoPosition) + ". " +
		"The SLOGAN lockup will be composited at " +
		formatZone(sloganX, sloganY, sloganW, sloganH) + " — " + sloganPositionPhrase(project.SloganPosition) + ". " +
		"Translate these constraints into natural photographic framing: " +
		"keep those areas atmospherically clear (soft sky, gentle bokeh, or a dark gradient) " +
		"so the overlaid brand assets sit cleanly without fighting the artwork. " +
		"NEVER mention percentages, coordinates, or zone labels in the image prompt." +
		footerDesc
}

func formatZone(x, y, w, h float64) string {
	pct := func(v float64) int { return int(v * 100) }
	return "[x=" + itoa(pct(x)) + "% y=" + itoa(pct(y)) + "% w=" + itoa(pct(w)) + "% h=" + itoa(pct(h)) + "%]"
}

func itoa(n int) string {
	// Avoids importing strconv in a file that otherwise has no stdlib deps.
	if n == 0 {
		return "0"
	}
	b := make([]byte, 0, 4)
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}

func logoPositionPhrase(pos string) string {
	switch pos {
	case "top-left":
		return "upper-left corner — the scene should open with atmospheric sky in the top-left quadrant"
	case "top-right":
		return "upper-right corner — the scene should open with atmospheric sky in the top-right quadrant"
	case "bottom-left":
		return "lower-left corner — the foreground should be clear in the bottom-left"
	case "bottom-right":
		return "lower-right corner — the foreground should be clear in the bottom-right"
	default:
		return "upper-center — the scene should open with atmospheric sky or a soft gradient at the top-center"
	}
}

func sloganPositionPhrase(pos string) string {
	switch pos {
	case "below-logo":
		return "immediately below the logo zone — a thin calm band beneath the top"
	case "above-footer":
		return "just above the footer bar — a slender strip of clear space near the bottom"
	case "top":
		return "near the top of the canvas"
	default:
		return "near the bottom of the canvas"
	}
}
