/**
 * timing.test.ts — TypeScript golden tests for timing.ts
 *
 * Loads the same timing_vectors.json that Go's timing_test.go reads.
 * Both suites MUST pass for the preview/export parity guarantee to hold.
 *
 * Run with: npx vitest run src/modules/video/engine/timing.test.ts
 * (Vitest is already a Vite project; add it to devDependencies if not present:
 *   npm install -D vitest)
 */

import { describe, it, expect } from "vitest";
import {
  totalDuration,
  clipStartTime,
  transitionStartTime,
  validateTransitionDuration,
  kenBurnsValue,
  textAlpha,
} from "./timing";
import type { ClipSpec, TextSpec } from "./spec";

// ---- Load vectors ----------------------------------------------------------

import fs from "node:fs";
import path from "node:path";

// Load golden vectors from backend testdata
const vectorPath = path.resolve(__dirname, "../../../../../backend/internal/video/render/testdata/timing_vectors.json");
const vectors = JSON.parse(fs.readFileSync(vectorPath, "utf-8"));

type WireTransition = { type: string; duration: number } | null;
type WireClip = { duration: number; transition_out: WireTransition };

function toClips(wires: WireClip[]): ClipSpec[] {
  return wires.map((w) => ({
    id: "x",
    image_key: "video/presets/x.jpg",
    duration: w.duration,
    crop: { x: 0, y: 0, w: 1, h: 1 },
    rotate: 0 as const,
    filter: "none" as const,
    adjust: { brightness: 0, contrast: 1, saturation: 1 },
    transition_out: w.transition_out
      ? { type: w.transition_out.type as ClipSpec["transition_out"] extends { type: infer T } ? T : never, duration: w.transition_out.duration }
      : undefined,
  }));
}

const EPS = 1e-6;

function approxEq(a: number, b: number): boolean {
  return Math.abs(a - b) < EPS;
}

// ---- total_duration vectors ------------------------------------------------

describe("totalDuration — golden vectors", () => {
  for (const v of (vectors as any).total_duration) {
    it(v.id, () => {
      let clips: ClipSpec[];
      if (v.clips_repeat) {
        const r = v.clips_repeat;
        clips = Array.from({ length: r.count }, () => ({
          id: "x", image_key: "video/presets/x.jpg",
          duration: r.duration,
          crop: { x: 0, y: 0, w: 1, h: 1 }, rotate: 0 as const,
          filter: "none" as const,
          adjust: { brightness: 0, contrast: 1, saturation: 1 },
          transition_out: r.transition_out
            ? { type: r.transition_out.type, duration: r.transition_out.duration }
            : undefined,
        }));
      } else {
        clips = toClips(v.clips);
      }
      const got = totalDuration(clips);
      expect(approxEq(got, v.expected)).toBe(true);
    });
  }
});

// ---- clip_start_times vectors ----------------------------------------------

describe("clipStartTime — golden vectors", () => {
  for (const v of (vectors as any).clip_start_times) {
    it(v.id, () => {
      const clips = toClips(v.clips);
      v.expected_starts.forEach((want: number, i: number) => {
        const got = clipStartTime(clips, i);
        expect(approxEq(got, want)).toBe(true);
      });
    });
  }
});

// ---- xfade_offsets vectors -------------------------------------------------

describe("transitionStartTime (xfadeOffset) — golden vectors", () => {
  for (const v of (vectors as any).xfade_offsets) {
    it(v.id, () => {
      const clips = toClips(v.clips);
      let offIdx = 0;
      clips.forEach((c, i) => {
        if (!c.transition_out) return;
        const want = v.expected_offsets[offIdx++];
        const got = transitionStartTime(clips, i);
        expect(approxEq(got, want)).toBe(true);
      });
    });
  }
});

// ---- transition_max_allowed vectors ----------------------------------------

describe("validateTransitionDuration — golden vectors", () => {
  for (const v of (vectors as any).transition_max_allowed) {
    it(v.id, () => {
      const clips = toClips(v.clips);
      const { maxAllowed } = validateTransitionDuration(clips, v.between_index, 0);
      expect(approxEq(maxAllowed, v.expected_max)).toBe(true);
    });
  }
});

// ---- ken_burns_easing vectors ----------------------------------------------

describe("kenBurnsValue — golden vectors", () => {
  for (const v of (vectors as any).ken_burns_easing) {
    it(v.id, () => {
      const got = kenBurnsValue(v.from, v.to, v.u, v.easing);
      expect(approxEq(got, v.expected)).toBe(true);
    });
  }
});

// ---- text_alpha vectors ----------------------------------------------------

describe("textAlpha — golden vectors", () => {
  for (const v of (vectors as any).text_alpha) {
    it(v.id, () => {
      const clipEnd = v.clip_start + 10; // large enough for all test cases
      const spec: TextSpec = {
        text: "test", font: "Inter", size: 32, color: "#ffffff",
        opacity: 1, x: 0.5, y: 0.5,
        start: v.text.start,
        end:   v.text.end,
        anim:  v.text.anim,
        anim_duration: v.text.anim_duration,
      };
      const got = textAlpha(v.clip_start, spec, v.t, clipEnd);
      expect(approxEq(got, v.expected_alpha)).toBe(true);
    });
  }
});
