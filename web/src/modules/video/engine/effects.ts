/**
 * effects.ts — TypeScript mirror of backend/internal/video/render/effects.go
 *
 * SYNC RULE: FilterPresets values here must exactly match those in effects.go.
 * The Canvas renderer uses these to apply CSS filters and a pixel-level
 * colorbalance pass so preview visually matches the FFmpeg export.
 */

import type { AdjustSpec, FilterName, KenBurnsEasing, KenBurnsSpec } from "./spec";

// ---- Filter presets --------------------------------------------------------

export interface FilterParams {
  // eq-equivalent params
  brightness: number; // [-1, +1], 0 = no change
  contrast:   number; // [0.5, 2.0], 1.0 = no change
  saturation: number; // [0, 3.0], 1.0 = no change
  gamma:      number; // [0.1, 10.0], 1.0 = no change

  // colorbalance: shadow/mid/highlight shift per channel [-1, +1]
  redShadow:   number; greenShadow:   number; blueShadow:   number;
  redMid:      number; greenMid:      number; blueMid:      number;
  redHigh:     number; greenHigh:     number; blueHigh:     number;
}

const IDENTITY: FilterParams = {
  brightness: 0, contrast: 1, saturation: 1, gamma: 1,
  redShadow: 0, greenShadow: 0, blueShadow: 0,
  redMid: 0,    greenMid: 0,   blueMid: 0,
  redHigh: 0,   greenHigh: 0,  blueHigh: 0,
};

/**
 * FilterPresets — must stay in sync with effects.go → FilterPresets.
 * Values are the source of truth for the Canvas renderer.
 */
export const FilterPresets: Record<FilterName, FilterParams> = {
  none:    { ...IDENTITY },
  vivid:   { ...IDENTITY, contrast: 1.15, saturation: 1.4, gamma: 0.95 },
  warm: {
    ...IDENTITY, saturation: 1.1,
    redHigh: 0.08, greenHigh: 0.02, blueHigh: -0.06,
    redMid: 0.04, blueMid: -0.04,
  },
  cold: {
    ...IDENTITY,
    blueHigh: 0.10, greenHigh: 0.03, redHigh: -0.05,
    blueMid: 0.05, redMid: -0.03,
  },
  bw: {
    ...IDENTITY, saturation: 0, contrast: 1.05,
  },
  vintage: {
    ...IDENTITY,
    brightness: 0.04, contrast: 0.90, saturation: 0.80, gamma: 1.08,
    redShadow: 0.06, greenShadow: -0.02, blueShadow: -0.04,
    redHigh: 0.06, blueHigh: -0.08,
  },
};

/**
 * resolveFilter merges the named preset with the per-clip AdjustSpec.
 * Mirrors Go's ResolveFilter exactly.
 */
export function resolveFilter(filter: FilterName, adj: AdjustSpec): FilterParams {
  const base = { ...FilterPresets[filter] ?? IDENTITY };

  base.brightness = clamp(base.brightness + adj.brightness, -1, 1);
  base.contrast   = clamp(base.contrast * adj.contrast,   0.5, 2.0);
  if (adj.saturation !== 1) {
    base.saturation = clamp(base.saturation * adj.saturation, 0, 3.0);
  }

  return base;
}

/**
 * toCSSFilter converts FilterParams to a CSS filter string suitable for a
 * <canvas> drawImage or an overlay <img>. Used for the live preview.
 *
 * NOTE: colorbalance shifts (red/green/blue mid/shadow/highlight) cannot be
 * expressed in CSS alone. They are applied via a pixel-level pass in
 * renderFrame.ts using applyColorBalance().
 */
export function toCSSFilter(p: FilterParams): string {
  const parts: string[] = [];

  if (p.brightness !== 0) {
    // CSS brightness() is multiplicative, but we treat ours as additive offset.
    // Approximate: offset [-1,+1] → brightness multiplier [0, 2].
    parts.push(`brightness(${(1 + p.brightness).toFixed(3)})`);
  }
  if (p.contrast !== 1) {
    parts.push(`contrast(${p.contrast.toFixed(3)})`);
  }
  if (p.saturation !== 1) {
    parts.push(`saturate(${p.saturation.toFixed(3)})`);
  }
  // gamma: no direct CSS equivalent — handled via gamma LUT in renderFrame.ts

  return parts.join(" ") || "none";
}

// ---- Ken Burns -------------------------------------------------------------

export interface KenBurnsFrame {
  cx: number;    // viewport center X in normalized image space [0,1]
  cy: number;    // viewport center Y in normalized image space [0,1]
  scale: number; // zoom factor
}

/**
 * kenBurnsFrame returns the viewport state for a clip at normalized time
 * u ∈ [0, 1]. Mirrors Go's KenBurnsFrame.
 */
export function kenBurnsFrame(kb: KenBurnsSpec, u: number): KenBurnsFrame {
  const e = applyEasing(clamp01(u), kb.easing);
  return {
    cx:    kb.from.x     + (kb.to.x     - kb.from.x)     * e,
    cy:    kb.from.y     + (kb.to.y     - kb.from.y)     * e,
    scale: kb.from.scale + (kb.to.scale - kb.from.scale) * e,
  };
}

/**
 * kenBurnsToPixels converts normalized Ken Burns viewport parameters to pixel
 * coordinates within a source image of (imgW × imgH), for output of (outW × outH).
 * Returns the crop rect (px, py, pw, ph) to pass to Canvas drawImage.
 */
export function kenBurnsToPixels(
  frame: KenBurnsFrame,
  outW: number, outH: number,
  imgW: number, imgH: number
): { px: number; py: number; pw: number; ph: number } {
  const aspectOut = outW / outH;
  const aspectImg = imgW / imgH;

  let viewW: number, viewH: number;
  if (aspectImg > aspectOut) {
    viewH = 1 / frame.scale;
    viewW = viewH * aspectOut / aspectImg;
  } else {
    viewW = 1 / frame.scale;
    viewH = viewW * aspectImg / aspectOut;
  }

  let x = frame.cx - viewW / 2;
  let y = frame.cy - viewH / 2;
  x = Math.max(0, Math.min(x, 1 - viewW));
  y = Math.max(0, Math.min(y, 1 - viewH));

  return {
    px: Math.round(x * imgW),
    py: Math.round(y * imgH),
    pw: Math.round(viewW * imgW),
    ph: Math.round(viewH * imgH),
  };
}

// ---- Contain-with-blur layout ---------------------------------------------

/**
 * ContainLayout describes parameters for the two-layer Canvas composite used
 * in "contain" fit mode: a blurred background + sharp centered foreground.
 *
 * blurRadius: px, apply to the background canvas layer (CSS filter or
 *   ctx.filter = "blur(Xpx)").
 * fgRect: the destination rect for the sharp image, centered in the output.
 */
export interface ContainLayout {
  blurRadius: number;
  fgRect: { x: number; y: number; w: number; h: number };
}

export function containLayout(
  imgW: number, imgH: number,
  outW: number, outH: number
): ContainLayout {
  const scale = Math.min(outW / imgW, outH / imgH);
  const fw = Math.round(imgW * scale);
  const fh = Math.round(imgH * scale);
  return {
    blurRadius: Math.round(outW / 8),
    fgRect: {
      x: Math.round((outW - fw) / 2),
      y: Math.round((outH - fh) / 2),
      w: fw,
      h: fh,
    },
  };
}

// ---- Helpers ---------------------------------------------------------------

function applyEasing(u: number, easing: KenBurnsEasing): number {
  if (easing === "ease-in-out") return u * u * (3 - 2 * u);
  return u; // linear
}

function clamp01(v: number): number { return Math.max(0, Math.min(1, v)); }
function clamp(v: number, lo: number, hi: number): number {
  return Math.max(lo, Math.min(hi, v));
}
