/**
 * timing.ts — TypeScript mirror of backend/internal/video/render/timing.go
 *
 * SYNC RULE: Every formula here must produce the same result as the Go
 * counterpart, to within 1 millisecond. Both files are tested against
 * the shared golden vectors in
 * backend/internal/video/render/testdata/timing_vectors.json.
 *
 * Import only from spec.ts — no browser APIs in this file.
 */

import type { ClipSpec, KenBurnsEasing, TextSpec } from "./spec";

// ---- Total duration --------------------------------------------------------

/**
 * Returns the wall-clock length of the rendered video in seconds.
 *
 *   total = Σ clip.duration − Σ transition.duration
 *
 * The last clip has no outgoing transition. A transition is "owned" by the
 * clip it exits.
 */
export function totalDuration(clips: ClipSpec[]): number {
  let total = 0;
  for (const c of clips) {
    total += c.duration;
    if (c.transition_out) {
      total -= c.transition_out.duration;
    }
  }
  return roundMs(total);
}

// ---- Clip start times ------------------------------------------------------

/**
 * Returns the absolute start time (within the video) of clips[i] in seconds.
 *
 * clips[0] always starts at t=0.
 * clips[i] starts after the non-overlapping portion of clips[i-1].
 */
export function clipStartTime(clips: ClipSpec[], i: number): number {
  let t = 0;
  for (let j = 0; j < i; j++) {
    t += clips[j].duration;
    if (clips[j].transition_out) {
      t -= clips[j].transition_out!.duration;
    }
  }
  return roundMs(t);
}

// ---- Transition start time -------------------------------------------------

/**
 * Returns the absolute start time of the outgoing transition of clips[i].
 * Throws if clips[i] has no transition_out.
 */
export function transitionStartTime(clips: ClipSpec[], i: number): number {
  const tr = clips[i].transition_out;
  if (!tr) throw new Error(`clips[${i}] has no transition_out`);
  return roundMs(clipStartTime(clips, i) + clips[i].duration - tr.duration);
}

/**
 * xfadeOffset — the `offset` parameter for FFmpeg's xfade filter between
 * clips[i] and clips[i+1]. Identical to transitionStartTime; provided under
 * this name to make the FFmpeg builder code self-documenting.
 */
export const xfadeOffset = transitionStartTime;

// ---- Transition max allowed ------------------------------------------------

/**
 * Returns the maximum allowed transition duration (in seconds) between
 * clips[i] and clips[i+1], based on the 0.5 × min(adjacent durations) rule.
 * Also returns whether `transitionDur` satisfies the rule.
 */
export function validateTransitionDuration(
  clips: ClipSpec[],
  i: number,
  transitionDur: number
): { ok: boolean; maxAllowed: number } {
  const minAdj = Math.min(clips[i].duration, clips[i + 1].duration);
  const maxAllowed = roundMs(0.5 * minAdj);
  return { ok: transitionDur <= maxAllowed, maxAllowed };
}

// ---- Clip local time -------------------------------------------------------

/**
 * Converts an absolute video time t into the normalized position u ∈ [0, 1]
 * within clips[i]. Returns -1 if t is outside the clip's active window.
 */
export function clipLocalTime(clips: ClipSpec[], i: number, t: number): number {
  const start = clipStartTime(clips, i);
  const end = start + clips[i].duration;
  if (t < start || t >= end) return -1;
  return clamp01((t - start) / clips[i].duration);
}

// ---- Ken Burns easing ------------------------------------------------------

/**
 * Interpolates a scalar Ken Burns parameter (scale, x, or y) at normalized
 * time u ∈ [0, 1] using the given easing function.
 *
 * Must match Go's KenBurnsValue exactly.
 */
export function kenBurnsValue(
  from: number,
  to: number,
  u: number,
  easing: KenBurnsEasing
): number {
  const e = applyEasing(clamp01(u), easing);
  return from + (to - from) * e;
}

// ---- Text alpha & slide offset ---------------------------------------------

/**
 * Returns the opacity of a text overlay at absolute video time t.
 * 0 if outside [start, end]; 1 after animation completes; [0,1] during entry.
 *
 * clipEnd is the absolute end time of the clip (used when text.end === 0).
 */
export function textAlpha(
  clipStart: number,
  text: TextSpec,
  t: number,
  clipEnd: number
): number {
  const absStart = clipStart + text.start;
  const absEnd = text.end === 0 ? clipEnd : clipStart + text.end;
  if (t < absStart || t >= absEnd) return 0;
  const animEnd = absStart + text.anim_duration;
  if (t >= animEnd) return 1;
  return clamp01((t - absStart) / text.anim_duration);
}

/**
 * Returns the normalized vertical slide offset for a "slide-up" animation.
 * 1 = fully off-screen below (at start), 0 = at final position (after anim).
 * Multiply by a pixel height to get the Y offset in pixels.
 */
export function textSlideOffsetNorm(
  clipStart: number,
  text: TextSpec,
  t: number
): number {
  const absStart = clipStart + text.start;
  const animEnd = absStart + text.anim_duration;
  if (t <= absStart) return 1;
  if (t >= animEnd) return 0;
  return 1 - clamp01((t - absStart) / text.anim_duration);
}

// ---- Helpers ---------------------------------------------------------------

function applyEasing(u: number, easing: KenBurnsEasing): number {
  if (easing === "ease-in-out") {
    // Cubic ease-in-out: 3u² − 2u³  (matches Go implementation)
    return u * u * (3 - 2 * u);
  }
  return u; // linear
}

function clamp01(v: number): number {
  return Math.max(0, Math.min(1, v));
}

/** Round to millisecond precision (matches Go's math.Round(v*1000)/1000). */
function roundMs(v: number): number {
  return Math.round(v * 1000) / 1000;
}
