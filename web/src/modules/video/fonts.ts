/**
 * fonts.ts — TypeScript mirror of backend/internal/video/render/fonts.go
 *
 * SYNC RULE: Any change here must be reflected in fonts.go and vice versa.
 * The font list drives:
 *   - The dropdown in TextOverlayEditor
 *   - The @font-face declarations in the video module CSS
 *   - The FFmpeg drawtext fontfile= parameter (backend)
 *   - The server-side font validator (backend)
 */

import type { FontName } from "./engine/spec";

/** Definitive list of OFL-licensed fonts available for text overlays. */
export const ALLOWED_FONTS: FontName[] = [
  "Inter",
  "Poppins",
  "Montserrat",
  "Playfair Display",
  "Bebas Neue",
  "Lobster",
];

export const ALLOWED_FONT_SET = new Set<string>(ALLOWED_FONTS);

/**
 * Maps logical font name → .ttf filename as it appears in VIDEO_FONTS_DIR
 * (backend) and in the web/public/fonts/ directory (frontend).
 *
 * Must stay in sync with FontFileName in fonts.go.
 */
export const FONT_FILE_NAME: Record<FontName, string> = {
  "Inter":            "Inter-Regular.ttf",
  "Poppins":          "Poppins-Regular.ttf",
  "Montserrat":       "Montserrat-Regular.ttf",
  "Playfair Display": "PlayfairDisplay-Regular.ttf",
  "Bebas Neue":       "BebasNeue-Regular.ttf",
  "Lobster":          "Lobster-Regular.ttf",
};

/**
 * CSS @font-face blocks to inject into the video module stylesheet.
 * Each font is loaded from /fonts/<filename> (served from web/public/fonts/).
 */
export const FONT_FACE_CSS = ALLOWED_FONTS.map(
  (name) =>
    `@font-face {
  font-family: '${name}';
  src: url('/fonts/${FONT_FILE_NAME[name]}') format('truetype');
  font-display: swap;
}`
).join("\n");
