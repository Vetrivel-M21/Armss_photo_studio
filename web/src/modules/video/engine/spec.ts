/**
 * spec.ts — TypeScript mirror of backend/internal/video/render/spec.go
 *
 * SYNC RULE: Any change to the Go struct must be reflected here AND the
 * SpecVersion constant must be bumped in both files.
 *
 * All coordinates are normalized [0, 1] (resolution-independent).
 */

export const SPEC_VERSION = 1;

// ---- Output ----------------------------------------------------------------

export type AspectRatio = "9:16" | "1:1" | "16:9";
export type FitMode = "cover" | "contain";
export type FPS = 24 | 30;

export interface OutputSpec {
  aspect: AspectRatio;
  /** "cover" = scale-and-crop. "contain" = letterbox with blurred background. */
  fit: FitMode;
  fps: FPS;
}

// ---- Clip ------------------------------------------------------------------

export interface CropSpec {
  x: number; // [0, 1]
  y: number; // [0, 1]
  w: number; // (0, 1], x + w ≤ 1
  h: number; // (0, 1], y + h ≤ 1
}

export interface KenBurnsKeyframe {
  /** Zoom factor relative to fit-to-frame baseline. Range [1.0, 2.0]. */
  scale: number;
  /** Viewport center X in normalized image space [0, 1]. */
  x: number;
  /** Viewport center Y in normalized image space [0, 1]. */
  y: number;
}

export type KenBurnsEasing = "linear" | "ease-in-out";

export interface KenBurnsSpec {
  from: KenBurnsKeyframe;
  to: KenBurnsKeyframe;
  easing: KenBurnsEasing;
}

export type FilterName = "none" | "vivid" | "warm" | "cold" | "bw" | "vintage";

export interface AdjustSpec {
  /** Additive brightness offset [-1, +1]. 0 = no change. */
  brightness: number;
  /** Multiplicative contrast [0.5, 2.0]. 1.0 = no change. */
  contrast: number;
  /** Multiplicative saturation [0.5, 2.0]. 1.0 = no change. */
  saturation: number;
}

export type TextAnim = "none" | "fade-in" | "slide-up";
export type FontName =
  | "Inter"
  | "Poppins"
  | "Montserrat"
  | "Playfair Display"
  | "Bebas Neue"
  | "Lobster";

export interface TextSpec {
  /** Display string. Max 200 chars. */
  text: string;
  font: FontName;
  /** Font size in pixels at output resolution. Range [16, 200]. */
  size: number;
  /** CSS hex color e.g. "#ffffff". */
  color: string;
  /** Opacity [0, 1]. Default 1. */
  opacity: number;
  /** Text anchor X in normalized output coords [0, 1]. Anchor = top-left. */
  x: number;
  /** Text anchor Y in normalized output coords [0, 1]. */
  y: number;
  /** Start time within the clip in seconds. */
  start: number;
  /** End time within the clip in seconds. 0 = show until clip ends. */
  end: number;
  /** Entry animation. */
  anim: TextAnim;
  /** Duration of the entry animation in seconds [0.1, 1.0]. */
  anim_duration: number;
}

export type TransitionType =
  | "fade"
  | "slide-left"
  | "slide-right"
  | "zoom-in"
  | "zoom-out"
  | "dissolve";

export interface TransitionSpec {
  type: TransitionType;
  /** Duration in seconds [0.3, 1.5]. Must be ≤ 0.5 × min(adjacent clips). */
  duration: number;
}

export interface ClipSpec {
  /** Stable client-generated UUID. */
  id: string;
  /**
   * S3 key of the source image. Must be under the user's video/ prefix
   * or under video/presets/. Validated server-side.
   */
  image_key: string;
  /** Clip on-screen duration in seconds [1, 10]. Default 3. */
  duration: number;
  /** Normalized sub-rectangle. {x:0,y:0,w:1,h:1} = full image. */
  crop: CropSpec;
  /** Clockwise rotation: 0 | 90 | 180 | 270. */
  rotate: 0 | 90 | 180 | 270;
  /** Animated zoom-pan. Omit for a static frame. */
  ken_burns?: KenBurnsSpec;
  filter: FilterName;
  adjust: AdjustSpec;
  /** Text overlays. Max 5. */
  texts?: TextSpec[];
  /** Transition from this clip to the next. Null on the last clip. */
  transition_out?: TransitionSpec;
}

// ---- Audio -----------------------------------------------------------------

export interface AudioSpec {
  /** S3 key under user's video/ prefix or video/presets/. */
  key: string;
  /** Offset into audio file to begin playback [s]. Default 0. */
  trim_start: number;
  /** Offset to stop reading [s]. 0 = read to EOF. */
  trim_end: number;
  /** Global gain multiplier [0, 2.0]. Default 1. */
  volume: number;
  /** Fade-in duration [s] from playback start. 0 = none. */
  fade_in: number;
  /** Fade-out duration [s] before video end. 0 = none. */
  fade_out: number;
  /** Friendly track title / filename for display in sequencer. */
  title?: string;
  /** Source audio duration in seconds. */
  duration?: number;
}

// ---- Top-level spec --------------------------------------------------------

export interface VideoSpec {
  version: typeof SPEC_VERSION;
  output: OutputSpec;
  clips: ClipSpec[];
  audio?: AudioSpec;
}

// ---- Defaults --------------------------------------------------------------

export function defaultOutputSpec(): OutputSpec {
  return { aspect: "9:16", fit: "cover", fps: 30 };
}

export function defaultAdjust(): AdjustSpec {
  return { brightness: 0, contrast: 1, saturation: 1 };
}

export function defaultClip(imageKey: string, id: string): ClipSpec {
  return {
    id,
    image_key: imageKey,
    duration: 3,
    crop: { x: 0, y: 0, w: 1, h: 1 },
    rotate: 0,
    filter: "none",
    adjust: defaultAdjust(),
    texts: [],
  };
}

export function defaultSpec(): VideoSpec {
  return {
    version: SPEC_VERSION,
    output: defaultOutputSpec(),
    clips: [],
  };
}

// ---- Output resolution table -----------------------------------------------

export type ResolutionPreset = "720p" | "1080p";

export interface Resolution {
  w: number;
  h: number;
}

/** Same table as effects.go → AspectResolutions. Must stay in sync. */
export const ASPECT_RESOLUTIONS: Record<AspectRatio, Record<ResolutionPreset, Resolution>> = {
  "9:16": { "720p": { w: 720, h: 1280 }, "1080p": { w: 1080, h: 1920 } },
  "1:1":  { "720p": { w: 720, h: 720 },  "1080p": { w: 1080, h: 1080 } },
  "16:9": { "720p": { w: 1280, h: 720 }, "1080p": { w: 1920, h: 1080 } },
};
