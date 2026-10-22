import React, { useState } from "react";
import type { ClipSpec, FilterName, TextSpec, KenBurnsEasing } from "../engine/spec";
import { ALLOWED_FONTS } from "../fonts";
import {
  Clock,
  Sparkles,
  Sliders,
  Type,
  ZoomIn,
  Trash2,
  Plus,
} from "lucide-react";

interface ClipInspectorProps {
  clip: ClipSpec | null;
  clipIndex: number;
  onUpdateClip: (id: string, patch: Partial<ClipSpec>) => void;
  onAddText: (clipId: string) => void;
  onUpdateText: (clipId: string, index: number, patch: Partial<TextSpec>) => void;
  onRemoveText: (clipId: string, index: number) => void;
}

const FILTERS: { name: FilterName; label: string }[] = [
  { name: "none", label: "Normal" },
  { name: "vivid", label: "Vivid" },
  { name: "warm", label: "Warm" },
  { name: "cold", label: "Cold" },
  { name: "bw", label: "B&W" },
  { name: "vintage", label: "Vintage" },
];

export const ClipInspector: React.FC<ClipInspectorProps> = ({
  clip,
  clipIndex,
  onUpdateClip,
  onAddText,
  onUpdateText,
  onRemoveText,
}) => {
  const [activeTab, setActiveTab] = useState<"adjust" | "effects" | "text">("adjust");

  if (!clip) {
    return (
      <div className="w-full h-full flex flex-col items-center justify-center p-6 text-center text-slate-500 bg-slate-900/60 rounded-2xl border border-white/10 backdrop-blur-md">
        <Sliders className="w-10 h-10 mb-2 stroke-1 text-slate-600" />
        <p className="text-sm font-medium">Select a clip on the timeline to edit its properties</p>
      </div>
    );
  }

  // Ken Burns presets helper
  const applyKenBurnsPreset = (preset: "zoom-in" | "zoom-out" | "pan-left" | "pan-right") => {
    switch (preset) {
      case "zoom-in":
        onUpdateClip(clip.id, {
          ken_burns: {
            from: { x: 0.5, y: 0.5, scale: 1.0 },
            to: { x: 0.5, y: 0.5, scale: 1.25 },
            easing: "ease-in-out",
          },
        });
        break;
      case "zoom-out":
        onUpdateClip(clip.id, {
          ken_burns: {
            from: { x: 0.5, y: 0.5, scale: 1.25 },
            to: { x: 0.5, y: 0.5, scale: 1.0 },
            easing: "ease-in-out",
          },
        });
        break;
      case "pan-left":
        onUpdateClip(clip.id, {
          ken_burns: {
            from: { x: 0.65, y: 0.5, scale: 1.15 },
            to: { x: 0.35, y: 0.5, scale: 1.15 },
            easing: "ease-in-out",
          },
        });
        break;
      case "pan-right":
        onUpdateClip(clip.id, {
          ken_burns: {
            from: { x: 0.35, y: 0.5, scale: 1.15 },
            to: { x: 0.65, y: 0.5, scale: 1.15 },
            easing: "ease-in-out",
          },
        });
        break;
    }
  };

  return (
    <div className="w-full bg-slate-900/80 border border-white/10 rounded-2xl p-4 shadow-xl flex flex-col gap-4 backdrop-blur-md">
      {/* Header */}
      <div className="flex items-center justify-between border-b border-white/10 pb-3">
        <div>
          <h3 className="text-sm font-bold text-white flex items-center gap-2">
            <span>Clip #{clipIndex + 1} Settings</span>
          </h3>
          <span className="text-[11px] text-slate-400">Duration & Visual Adjustments</span>
        </div>

        {/* Tab Switcher */}
        <div className="flex items-center bg-white/5 p-1 rounded-xl border border-white/10 text-xs">
          <button
            type="button"
            onClick={() => setActiveTab("adjust")}
            className={`px-3 py-1 rounded-lg font-medium transition-all ${
              activeTab === "adjust"
                ? "bg-brand-primary text-white shadow-glow-sm"
                : "text-slate-400 hover:text-white"
            }`}
          >
            Adjust
          </button>
          <button
            type="button"
            onClick={() => setActiveTab("effects")}
            className={`px-3 py-1 rounded-lg font-medium transition-all ${
              activeTab === "effects"
                ? "bg-brand-primary text-white shadow-glow-sm"
                : "text-slate-400 hover:text-white"
            }`}
          >
            Motion
          </button>
          <button
            type="button"
            onClick={() => setActiveTab("text")}
            className={`px-3 py-1 rounded-lg font-medium transition-all ${
              activeTab === "text"
                ? "bg-brand-primary text-white shadow-glow-sm"
                : "text-slate-400 hover:text-white"
            }`}
          >
            Text ({clip.texts?.length || 0})
          </button>
        </div>
      </div>

      {/* Tab 1: Adjust (Duration, Filters, Light) */}
      {activeTab === "adjust" && (
        <div className="flex flex-col gap-4">
          {/* Duration */}
          <div className="flex flex-col gap-2">
            <div className="flex justify-between items-center text-xs">
              <span className="text-slate-300 font-medium flex items-center gap-1.5">
                <Clock className="w-3.5 h-3.5 text-brand-primary" />
                Clip Duration
              </span>
              <div className="flex items-center gap-1">
                <input
                  type="number"
                  min={1}
                  max={1800}
                  step={0.5}
                  value={clip.duration}
                  onChange={(e) => {
                    const val = parseFloat(e.target.value);
                    if (!isNaN(val) && val >= 1 && val <= 1800) {
                      onUpdateClip(clip.id, { duration: val });
                    }
                  }}
                  className="w-16 px-1.5 py-0.5 rounded bg-white/10 text-white font-mono font-bold text-xs text-right border border-white/20 focus:border-brand-primary outline-none"
                />
                <span className="text-xs text-slate-400 font-mono">s</span>
              </div>
            </div>

            <input
              type="range"
              min={1}
              max={Math.max(120, clip.duration)}
              step={0.5}
              value={clip.duration}
              onChange={(e) => onUpdateClip(clip.id, { duration: parseFloat(e.target.value) })}
              className="accent-brand-primary w-full cursor-pointer h-1.5 bg-white/10 rounded-lg"
            />

            {/* Quick Duration Presets for slideshows & long-form videos */}
            <div className="flex items-center gap-1.5 flex-wrap">
              {[
                { label: "3s", val: 3 },
                { label: "5s", val: 5 },
                { label: "10s", val: 10 },
                { label: "30s", val: 30 },
                { label: "1m", val: 60 },
                { label: "5m", val: 300 },
              ].map((p) => (
                <button
                  key={p.label}
                  type="button"
                  onClick={() => onUpdateClip(clip.id, { duration: p.val })}
                  className={`px-2 py-0.5 rounded-md text-[10px] font-mono transition-colors ${
                    clip.duration === p.val
                      ? "bg-brand-primary text-white font-bold shadow-sm"
                      : "bg-white/5 hover:bg-white/10 text-slate-300 border border-white/10"
                  }`}
                >
                  {p.label}
                </button>
              ))}
            </div>
          </div>

          {/* Color Filter Presets */}
          <div className="flex flex-col gap-2">
            <span className="text-xs text-slate-300 font-medium flex items-center gap-1.5">
              <Sparkles className="w-3.5 h-3.5 text-purple-400" />
              Color Filter
            </span>
            <div className="grid grid-cols-3 gap-2">
              {FILTERS.map((f) => (
                <button
                  key={f.name}
                  type="button"
                  onClick={() => onUpdateClip(clip.id, { filter: f.name })}
                  className={`py-2 px-2 rounded-xl text-xs font-semibold border transition-all ${
                    clip.filter === f.name
                      ? "bg-gradient-to-r from-brand-primary to-purple-600 text-white border-transparent shadow-glow-sm"
                      : "bg-white/5 border-white/10 text-slate-300 hover:bg-white/10 hover:text-white"
                  }`}
                >
                  {f.label}
                </button>
              ))}
            </div>
          </div>

          {/* Brightness, Contrast, Saturation */}
          <div className="space-y-3 pt-2 border-t border-white/5">
            {/* Brightness */}
            <div className="flex flex-col gap-1">
              <div className="flex justify-between text-xs">
                <span className="text-slate-400">Brightness</span>
                <span className="font-mono text-slate-200">
                  {clip.adjust.brightness > 0 ? `+${clip.adjust.brightness}` : clip.adjust.brightness}
                </span>
              </div>
              <input
                type="range"
                min={-0.5}
                max={0.5}
                step={0.05}
                value={clip.adjust.brightness}
                onChange={(e) =>
                  onUpdateClip(clip.id, {
                    adjust: { ...clip.adjust, brightness: parseFloat(e.target.value) },
                  })
                }
                className="accent-brand-primary w-full cursor-pointer h-1.5 bg-white/10 rounded-lg"
              />
            </div>

            {/* Contrast */}
            <div className="flex flex-col gap-1">
              <div className="flex justify-between text-xs">
                <span className="text-slate-400">Contrast</span>
                <span className="font-mono text-slate-200">{clip.adjust.contrast.toFixed(2)}x</span>
              </div>
              <input
                type="range"
                min={0.5}
                max={2.0}
                step={0.05}
                value={clip.adjust.contrast}
                onChange={(e) =>
                  onUpdateClip(clip.id, {
                    adjust: { ...clip.adjust, contrast: parseFloat(e.target.value) },
                  })
                }
                className="accent-brand-primary w-full cursor-pointer h-1.5 bg-white/10 rounded-lg"
              />
            </div>

            {/* Saturation */}
            <div className="flex flex-col gap-1">
              <div className="flex justify-between text-xs">
                <span className="text-slate-400">Saturation</span>
                <span className="font-mono text-slate-200">{clip.adjust.saturation.toFixed(2)}x</span>
              </div>
              <input
                type="range"
                min={0.0}
                max={2.0}
                step={0.05}
                value={clip.adjust.saturation}
                onChange={(e) =>
                  onUpdateClip(clip.id, {
                    adjust: { ...clip.adjust, saturation: parseFloat(e.target.value) },
                  })
                }
                className="accent-brand-primary w-full cursor-pointer h-1.5 bg-white/10 rounded-lg"
              />
            </div>
          </div>
        </div>
      )}

      {/* Tab 2: Motion (Ken Burns) */}
      {activeTab === "effects" && (
        <div className="flex flex-col gap-4">
          <div className="flex items-center justify-between">
            <span className="text-xs text-slate-300 font-medium flex items-center gap-1.5">
              <ZoomIn className="w-3.5 h-3.5 text-amber-400" />
              Ken Burns Zoom-Pan
            </span>
            <button
              type="button"
              onClick={() => {
                if (clip.ken_burns) {
                  onUpdateClip(clip.id, { ken_burns: undefined });
                } else {
                  applyKenBurnsPreset("zoom-in");
                }
              }}
              className={`text-xs px-2.5 py-1 rounded-lg font-semibold transition-all ${
                clip.ken_burns
                  ? "bg-amber-500 text-slate-950"
                  : "bg-white/10 text-slate-300 hover:bg-white/20"
              }`}
            >
              {clip.ken_burns ? "Enabled" : "Disabled"}
            </button>
          </div>

          {clip.ken_burns && (
            <div className="flex flex-col gap-3 p-3 rounded-xl bg-white/[0.04] border border-white/10">
              <span className="text-xs text-slate-400">Presets</span>
              <div className="grid grid-cols-2 gap-2">
                <button
                  type="button"
                  onClick={() => applyKenBurnsPreset("zoom-in")}
                  className="p-2 rounded-lg bg-white/5 hover:bg-white/10 text-xs text-slate-200 font-medium border border-white/10"
                >
                  🔍 Slow Zoom In
                </button>
                <button
                  type="button"
                  onClick={() => applyKenBurnsPreset("zoom-out")}
                  className="p-2 rounded-lg bg-white/5 hover:bg-white/10 text-xs text-slate-200 font-medium border border-white/10"
                >
                  🔎 Slow Zoom Out
                </button>
                <button
                  type="button"
                  onClick={() => applyKenBurnsPreset("pan-left")}
                  className="p-2 rounded-lg bg-white/5 hover:bg-white/10 text-xs text-slate-200 font-medium border border-white/10"
                >
                  ⬅ Pan Left
                </button>
                <button
                  type="button"
                  onClick={() => applyKenBurnsPreset("pan-right")}
                  className="p-2 rounded-lg bg-white/5 hover:bg-white/10 text-xs text-slate-200 font-medium border border-white/10"
                >
                  ➡ Pan Right
                </button>
              </div>

              <div className="flex items-center justify-between text-xs pt-2">
                <span className="text-slate-400">Easing</span>
                <select
                  value={clip.ken_burns.easing}
                  onChange={(e) => {
                    if (!clip.ken_burns) return;
                    onUpdateClip(clip.id, {
                      ken_burns: {
                        ...clip.ken_burns,
                        easing: e.target.value as KenBurnsEasing,
                      },
                    });
                  }}
                  className="bg-slate-800 border border-white/10 rounded-lg px-2 py-1 text-white outline-none cursor-pointer"
                >
                  <option value="ease-in-out">Smooth (Ease-In-Out)</option>
                  <option value="linear">Linear</option>
                </select>
              </div>
            </div>
          )}
        </div>
      )}

      {/* Tab 3: Text Overlays */}
      {activeTab === "text" && (
        <div className="flex flex-col gap-3">
          <div className="flex items-center justify-between">
            <span className="text-xs text-slate-300 font-medium flex items-center gap-1.5">
              <Type className="w-3.5 h-3.5 text-sky-400" />
              Captions & Overlays
            </span>
            <button
              type="button"
              onClick={() => onAddText(clip.id)}
              disabled={(clip.texts?.length || 0) >= 5}
              className="flex items-center gap-1 text-xs px-2.5 py-1 rounded-lg font-semibold bg-brand-primary text-white hover:bg-brand-secondary disabled:opacity-50 shadow-glow-sm"
            >
              <Plus className="w-3 h-3" />
              Add Text
            </button>
          </div>

          <div className="flex flex-col gap-2 max-h-[360px] overflow-y-auto pr-1">
            {(!clip.texts || clip.texts.length === 0) && (
              <div className="p-4 text-center text-xs text-slate-500 rounded-xl bg-white/[0.02] border border-dashed border-white/10">
                No captions added yet. Click &quot;Add Text&quot; to overlay titles or captions on this clip.
              </div>
            )}

            {clip.texts?.map((t, tIdx) => (
              <div
                key={tIdx}
                className="p-3 rounded-xl bg-white/[0.04] border border-white/10 flex flex-col gap-2.5"
              >
                <div className="flex items-center justify-between gap-2">
                  <input
                    type="text"
                    value={t.text}
                    onChange={(e) => onUpdateText(clip.id, tIdx, { text: e.target.value })}
                    className="flex-1 bg-black/40 border border-white/10 rounded-lg px-2.5 py-1 text-xs text-white outline-none focus:border-brand-primary"
                    placeholder="Enter caption..."
                  />
                  <button
                    type="button"
                    onClick={() => onRemoveText(clip.id, tIdx)}
                    className="p-1 rounded bg-rose-600/30 text-rose-300 hover:bg-rose-600 hover:text-white transition-colors"
                  >
                    <Trash2 className="w-3.5 h-3.5" />
                  </button>
                </div>

                <div className="grid grid-cols-2 gap-2 text-xs">
                  {/* Font picker */}
                  <div className="flex flex-col gap-1">
                    <span className="text-slate-400 text-[10px]">Font</span>
                    <select
                      value={t.font}
                      onChange={(e) => onUpdateText(clip.id, tIdx, { font: e.target.value as any })}
                      className="bg-slate-800 border border-white/10 rounded-lg px-2 py-1 text-slate-200 outline-none"
                    >
                      {ALLOWED_FONTS.map((fontName) => (
                        <option key={fontName} value={fontName}>
                          {fontName}
                        </option>
                      ))}
                    </select>
                  </div>

                  {/* Animation */}
                  <div className="flex flex-col gap-1">
                    <span className="text-slate-400 text-[10px]">Animation</span>
                    <select
                      value={t.anim}
                      onChange={(e) => onUpdateText(clip.id, tIdx, { anim: e.target.value as any })}
                      className="bg-slate-800 border border-white/10 rounded-lg px-2 py-1 text-slate-200 outline-none"
                    >
                      <option value="none">Static</option>
                      <option value="fade-in">Fade In</option>
                      <option value="slide-up">Slide Up</option>
                    </select>
                  </div>
                </div>

                {/* Color & Size */}
                <div className="flex items-center justify-between gap-3 text-xs pt-1">
                  <div className="flex items-center gap-2">
                    <span className="text-slate-400 text-[10px]">Color</span>
                    <input
                      type="color"
                      value={t.color}
                      onChange={(e) => onUpdateText(clip.id, tIdx, { color: e.target.value })}
                      className="w-6 h-6 rounded border-0 bg-transparent cursor-pointer"
                    />
                  </div>

                  <div className="flex items-center gap-2 flex-1 max-w-[140px]">
                    <span className="text-slate-400 text-[10px]">Size</span>
                    <input
                      type="range"
                      min={24}
                      max={96}
                      step={4}
                      value={t.size}
                      onChange={(e) => onUpdateText(clip.id, tIdx, { size: parseInt(e.target.value, 10) })}
                      className="accent-brand-primary w-full cursor-pointer h-1 bg-white/10 rounded"
                    />
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
};
