import React, { useState, useEffect, useMemo } from "react";
import type { VideoSpec, ClipSpec, TransitionType, AudioSpec } from "../engine/spec";
import { resolveMediaUrl } from "../utils/mediaUtils";
import {
  Plus,
  Trash2,
  ChevronLeft,
  ChevronRight,
  Music,
  Image as ImageIcon,
  ZoomIn,
  ZoomOut,
  Maximize2,
  Sliders,
  GripVertical,
} from "lucide-react";

interface TimelineProps {
  spec: VideoSpec;
  activeClipId: string | null;
  isAudioSelected?: boolean;
  currentTime: number;
  totalTime: number;
  onSelectClip: (clipId: string) => void;
  onSelectAudio?: () => void;
  onRemoveClip: (clipId: string) => void;
  onReorderClips: (from: number, to: number) => void;
  onUpdateClip: (clipId: string, patch: Partial<ClipSpec>) => void;
  onUpdateAudio?: (patch: Partial<AudioSpec>) => void;
  onAddMediaClick: () => void;
  onAudioClick: () => void;
  onSeek: (time: number) => void;
}

const TRANSITIONS: { type: TransitionType; label: string }[] = [
  { type: "fade", label: "Fade" },
  { type: "dissolve", label: "Dissolve" },
  { type: "slide-left", label: "Slide Left" },
  { type: "slide-right", label: "Slide Right" },
  { type: "zoom-in", label: "Zoom In" },
  { type: "zoom-out", label: "Zoom Out" },
];

export const Timeline: React.FC<TimelineProps> = ({
  spec,
  activeClipId,
  isAudioSelected,
  currentTime,
  totalTime,
  onSelectClip,
  onSelectAudio,
  onRemoveClip,
  onReorderClips,
  onUpdateClip,
  onUpdateAudio,
  onAddMediaClick,
  onAudioClick,
  onSeek,
}) => {
  // Max time across clips, audio, and minimum 15 seconds
  const audioTrackDuration = spec.audio?.duration || 0;
  const maxTimelineSec = Math.max(totalTime, audioTrackDuration, 15);

  // Zoom levels: fit, 0.25x, 0.5x, 1x, 2x, 4x
  type ZoomLevel = "fit" | "0.25x" | "0.5x" | "1x" | "2x" | "4x";
  const [zoom, setZoom] = useState<ZoomLevel>(() => {
    if (maxTimelineSec > 300) return "fit"; // default to fit for 5-25 min videos
    if (maxTimelineSec > 60) return "0.5x";
    return "1x";
  });

  // Pixels per second for horizontal scaling based on zoom
  const pxPerSec = useMemo(() => {
    switch (zoom) {
      case "fit":
        return Math.max(0.6, Math.min(60, 960 / Math.max(maxTimelineSec, 1)));
      case "0.25x":
        return 4; // 25 mins = 6000px
      case "0.5x":
        return 12; // 25 mins = 18000px
      case "1x":
        return 28;
      case "2x":
        return 50;
      case "4x":
        return 75;
      default:
        return 28;
    }
  }, [zoom, maxTimelineSec]);

  const trackWidth = Math.max(920, Math.ceil(maxTimelineSec * pxPerSec + 240));

  // Playhead position
  const playheadLeft = currentTime * pxPerSec;

  const handleTimelineClick = (e: React.MouseEvent<HTMLDivElement>) => {
    // If clicking on ruler or track background
    const rect = e.currentTarget.getBoundingClientRect();
    const clickX = e.clientX - rect.left;
    const targetSec = Math.max(0, Math.min(maxTimelineSec, clickX / pxPerSec));
    onSeek(targetSec);
  };

  const formatRulerTime = (sec: number) => {
    if (isNaN(sec) || sec < 0) sec = 0;
    const m = Math.floor(sec / 60);
    const s = Math.floor(sec % 60);
    return `${m}:${s.toString().padStart(2, "0")}`;
  };

  const formatDetailedTime = (sec: number) => {
    if (isNaN(sec) || sec < 0) sec = 0;
    const m = Math.floor(sec / 60);
    const s = Math.floor(sec % 60);
    const ms = Math.floor((sec % 1) * 10);
    return `${m.toString().padStart(2, "0")}:${s.toString().padStart(2, "0")}.${ms}`;
  };

  // Adaptive ruler intervals so the DOM never lags on 20-25 minute videos
  const rulerTicks = useMemo(() => {
    const minPixelGap = 65;
    const stepOptions = [1, 2, 5, 10, 15, 30, 60, 120, 300, 600];
    let stepSec = 1;
    for (const step of stepOptions) {
      if (step * pxPerSec >= minPixelGap) {
        stepSec = step;
        break;
      }
      stepSec = step;
    }

    const ticks: { sec: number; isMajor: boolean }[] = [];
    for (let s = 0; s <= maxTimelineSec + stepSec; s += stepSec) {
      // Major ticks on minutes or intervals of 5
      const isMajor = stepSec >= 60 ? s % 60 === 0 : s % (stepSec * 2) === 0 || s === 0;
      ticks.push({ sec: s, isMajor });
    }
    return ticks;
  }, [maxTimelineSec, pxPerSec]);

  // ---- Audio Dragging & Visual Trimming State -------------------------------
  type DragMode = "body" | "left-trim" | "right-trim" | null;
  const [dragMode, setDragMode] = useState<DragMode>(null);
  const [dragInfo, setDragInfo] = useState<{
    startX: number;
    initialTrimStart: number;
    initialTrimEnd: number;
    currentTrimStart: number;
    currentTrimEnd: number;
  } | null>(null);

  const audioEffectiveEnd = useMemo(() => {
    const end = spec.audio?.trim_end || 0;
    return end > 0 && end <= audioTrackDuration ? end : audioTrackDuration;
  }, [spec.audio?.trim_end, audioTrackDuration]);

  const audioEffectiveStart = spec.audio?.trim_start || 0;

  const handleAudioMouseDown = (
    mode: "body" | "left-trim" | "right-trim",
    e: React.MouseEvent
  ) => {
    if (!spec.audio?.key) return;
    e.stopPropagation();
    e.preventDefault();

    onSelectAudio?.();

    const startX = e.clientX;
    const initialTrimStart = spec.audio.trim_start || 0;
    const initialTrimEnd = audioEffectiveEnd;

    setDragMode(mode);
    setDragInfo({
      startX,
      initialTrimStart,
      initialTrimEnd,
      currentTrimStart: initialTrimStart,
      currentTrimEnd: initialTrimEnd,
    });
  };

  useEffect(() => {
    if (!dragMode || !dragInfo) return;

    const handleMouseMove = (e: MouseEvent) => {
      const deltaPx = e.clientX - dragInfo.startX;
      const deltaSec = deltaPx / pxPerSec;
      const totalDur = audioTrackDuration > 0 ? audioTrackDuration : maxTimelineSec;

      if (dragMode === "body") {
        // Shift start time
        const span = Math.max(1, dragInfo.initialTrimEnd - dragInfo.initialTrimStart);
        let nextStart = Math.max(0, dragInfo.initialTrimStart + deltaSec);
        let nextEnd = nextStart + span;

        if (nextEnd > totalDur) {
          nextEnd = totalDur;
          nextStart = Math.max(0, nextEnd - span);
        }

        setDragInfo((prev) =>
          prev
            ? {
                ...prev,
                currentTrimStart: parseFloat(nextStart.toFixed(1)),
                currentTrimEnd: parseFloat(nextEnd.toFixed(1)),
              }
            : null
        );
      } else if (dragMode === "left-trim") {
        // Adjust trim_start
        const maxStart = dragInfo.initialTrimEnd - 0.5;
        const nextStart = Math.max(0, Math.min(maxStart, dragInfo.initialTrimStart + deltaSec));

        setDragInfo((prev) =>
          prev
            ? {
                ...prev,
                currentTrimStart: parseFloat(nextStart.toFixed(1)),
              }
            : null
        );
      } else if (dragMode === "right-trim") {
        // Adjust trim_end
        const minEnd = dragInfo.initialTrimStart + 0.5;
        const nextEnd = Math.max(minEnd, Math.min(totalDur, dragInfo.initialTrimEnd + deltaSec));

        setDragInfo((prev) =>
          prev
            ? {
                ...prev,
                currentTrimEnd: parseFloat(nextEnd.toFixed(1)),
              }
            : null
        );
      }
    };

    const handleMouseUp = () => {
      if (dragInfo && onUpdateAudio) {
        onUpdateAudio({
          trim_start: dragInfo.currentTrimStart,
          trim_end:
            dragInfo.currentTrimEnd >= (audioTrackDuration || maxTimelineSec)
              ? 0
              : dragInfo.currentTrimEnd,
        });
      }
      setDragMode(null);
      setDragInfo(null);
    };

    window.addEventListener("mousemove", handleMouseMove);
    window.addEventListener("mouseup", handleMouseUp);

    return () => {
      window.removeEventListener("mousemove", handleMouseMove);
      window.removeEventListener("mouseup", handleMouseUp);
    };
  }, [dragMode, dragInfo, pxPerSec, audioTrackDuration, maxTimelineSec, onUpdateAudio]);

  const activeTrimStart = dragInfo ? dragInfo.currentTrimStart : audioEffectiveStart;
  const activeTrimEnd = dragInfo ? dragInfo.currentTrimEnd : audioEffectiveEnd;
  const activePlayingAudioDur = Math.max(0.1, activeTrimEnd - activeTrimStart);

  return (
    <div className="w-full bg-slate-900/80 border border-white/10 rounded-2xl p-4 shadow-xl flex flex-col gap-3 backdrop-blur-md select-none">
      {/* Timeline Toolbar with Zoom Controls & Track Information */}
      <div className="flex flex-wrap items-center justify-between gap-3 pb-2 border-b border-white/10">
        <div className="flex items-center gap-3">
          <span className="text-xs font-bold uppercase tracking-wider text-slate-400">
            Timeline Sequencer
          </span>
          <span className="text-xs px-2 py-0.5 rounded bg-white/10 text-slate-300 font-mono">
            {spec.clips.length} clips • {formatRulerTime(totalTime)} ({totalTime.toFixed(1)}s)
          </span>
          {audioTrackDuration > 0 && (
            <span className="text-xs px-2 py-0.5 rounded bg-purple-950/80 text-purple-300 border border-purple-500/30 font-mono hidden sm:inline-block">
              Music: {formatRulerTime(audioTrackDuration)}
            </span>
          )}
        </div>

        {/* Zoom Controls & Media Buttons */}
        <div className="flex items-center gap-2 flex-wrap">
          {/* Zoom controls for long-form 20-25 min videos */}
          <div className="flex items-center bg-black/40 border border-white/10 rounded-xl p-1 gap-1">
            <button
              type="button"
              onClick={() => {
                const levels: ZoomLevel[] = ["fit", "0.25x", "0.5x", "1x", "2x", "4x"];
                const currIdx = levels.indexOf(zoom);
                if (currIdx > 0) setZoom(levels[currIdx - 1]);
              }}
              className="p-1 rounded-lg text-slate-400 hover:text-white hover:bg-white/10 transition-colors"
              title="Zoom Out"
            >
              <ZoomOut className="w-3.5 h-3.5" />
            </button>

            <span className="text-[10px] font-mono text-slate-300 px-1 font-semibold min-w-[34px] text-center">
              {zoom}
            </span>

            <button
              type="button"
              onClick={() => {
                const levels: ZoomLevel[] = ["fit", "0.25x", "0.5x", "1x", "2x", "4x"];
                const currIdx = levels.indexOf(zoom);
                if (currIdx < levels.length - 1) setZoom(levels[currIdx + 1]);
              }}
              className="p-1 rounded-lg text-slate-400 hover:text-white hover:bg-white/10 transition-colors"
              title="Zoom In"
            >
              <ZoomIn className="w-3.5 h-3.5" />
            </button>

            <button
              type="button"
              onClick={() => setZoom("fit")}
              className={`flex items-center gap-1 px-2 py-0.5 rounded-lg text-[10px] font-semibold transition-all ${
                zoom === "fit"
                  ? "bg-brand-primary text-white shadow-glow-sm"
                  : "text-slate-400 hover:text-white hover:bg-white/5"
              }`}
              title="Fit entire video and audio to screen width"
            >
              <Maximize2 className="w-2.5 h-2.5" />
              Fit
            </button>
          </div>

          {/* Audio & Media Buttons */}
          <button
            type="button"
            onClick={onAudioClick}
            className={`flex items-center gap-1.5 px-3 py-1.5 rounded-xl text-xs font-semibold transition-all ${
              spec.audio?.key
                ? "bg-purple-600/30 text-purple-300 border border-purple-500/40 hover:bg-purple-600/50"
                : "bg-white/5 text-slate-300 border border-white/10 hover:bg-white/10"
            }`}
          >
            <Music className="w-3.5 h-3.5" />
            {spec.audio?.key ? "Audio Track Set" : "Add Audio"}
          </button>

          <button
            type="button"
            onClick={onAddMediaClick}
            className="flex items-center gap-1.5 px-3 py-1.5 rounded-xl text-xs font-semibold bg-brand-primary hover:bg-brand-secondary text-white shadow-glow-sm transition-all"
          >
            <Plus className="w-3.5 h-3.5" />
            Add Media
          </button>
        </div>
      </div>

      {/* Scrollable Sequencer Tracks Area */}
      <div className="relative overflow-x-auto pb-2 scrollbar-thin scrollbar-thumb-white/20">
        <div
          className="relative min-h-[200px] pt-7"
          style={{ width: `${trackWidth}px` }}
        >
          {/* Time Ruler (Minute & Second Scale - Adaptive) */}
          <div
            className="absolute top-0 left-0 right-0 h-7 border-b border-white/10 flex items-center cursor-pointer select-none"
            onClick={handleTimelineClick}
          >
            {rulerTicks.map(({ sec, isMajor }) => {
              return (
                <div
                  key={sec}
                  className="absolute font-mono flex flex-col items-center"
                  style={{ left: `${sec * pxPerSec}px` }}
                >
                  <span
                    className={`text-[9px] ${
                      isMajor ? "text-slate-300 font-semibold" : "text-slate-500 text-[8px]"
                    }`}
                  >
                    {formatRulerTime(sec)}
                  </span>
                  <div
                    className={`mt-0.5 ${
                      isMajor ? "h-2.5 w-px bg-white/40" : "h-1 w-px bg-white/15"
                    }`}
                  />
                </div>
              );
            })}
          </div>

          {/* Red Playhead Line */}
          <div
            className="absolute top-0 bottom-0 w-0.5 bg-rose-500 z-30 pointer-events-none"
            style={{ left: `${playheadLeft}px` }}
          >
            <div className="relative">
              <div className="w-3.5 h-3.5 bg-rose-500 rotate-45 -translate-x-[6px] -translate-y-1 shadow-glow-sm flex items-center justify-center">
                <div className="w-1.5 h-1.5 bg-white rounded-full" />
              </div>
              <div className="absolute top-3.5 -translate-x-1/2 px-1.5 py-0.5 rounded bg-rose-600/90 text-white font-mono text-[9px] font-bold shadow-md whitespace-nowrap border border-white/20">
                {formatDetailedTime(currentTime)}
              </div>
            </div>
          </div>

          {/* Clip Track */}
          <div className="flex items-center gap-2 mt-2">
            {spec.clips.map((clip, idx) => {
              const isActive = clip.id === activeClipId;
              const clipWidth = Math.max(90, clip.duration * pxPerSec);
              const imgUrl = resolveMediaUrl(clip.image_key);

              return (
                <React.Fragment key={clip.id}>
                  {/* Clip Card */}
                  <div
                    onClick={() => onSelectClip(clip.id)}
                    style={{ width: `${clipWidth}px` }}
                    className={`group relative h-28 rounded-xl border flex flex-col justify-between overflow-hidden cursor-pointer transition-all duration-200 shrink-0 ${
                      isActive
                        ? "border-brand-primary ring-2 ring-brand-primary/40 bg-white/[0.08] shadow-lg shadow-brand-primary/20"
                        : "border-white/10 bg-slate-800/60 hover:border-white/30 hover:bg-slate-800"
                    }`}
                  >
                    {/* Background thumbnail */}
                    {imgUrl ? (
                      <img
                        src={imgUrl}
                        alt="clip"
                        className="absolute inset-0 w-full h-full object-cover opacity-60 group-hover:opacity-75 transition-opacity pointer-events-none"
                      />
                    ) : (
                      <div className="absolute inset-0 flex items-center justify-center text-slate-600">
                        <ImageIcon className="w-8 h-8 opacity-40" />
                      </div>
                    )}

                    {/* Gradient overlay */}
                    <div className="absolute inset-0 bg-gradient-to-t from-black/80 via-black/20 to-black/40 pointer-events-none" />

                    {/* Top Row: Index & Actions */}
                    <div className="relative z-10 flex items-center justify-between p-1.5">
                      <span className="text-[10px] font-mono font-bold px-1.5 py-0.5 rounded bg-black/60 text-white backdrop-blur-sm border border-white/10">
                        #{idx + 1}
                      </span>

                      <div className="flex items-center gap-0.5 opacity-0 group-hover:opacity-100 transition-opacity bg-black/60 rounded-lg p-0.5 backdrop-blur-sm border border-white/10">
                        {idx > 0 && (
                          <button
                            type="button"
                            onClick={(e) => {
                              e.stopPropagation();
                              onReorderClips(idx, idx - 1);
                            }}
                            className="p-1 hover:text-brand-primary text-slate-300"
                            title="Move clip left"
                          >
                            <ChevronLeft className="w-3 h-3" />
                          </button>
                        )}
                        {idx < spec.clips.length - 1 && (
                          <button
                            type="button"
                            onClick={(e) => {
                              e.stopPropagation();
                              onReorderClips(idx, idx + 1);
                            }}
                            className="p-1 hover:text-brand-primary text-slate-300"
                            title="Move clip right"
                          >
                            <ChevronRight className="w-3 h-3" />
                          </button>
                        )}
                        <button
                          type="button"
                          onClick={(e) => {
                            e.stopPropagation();
                            onRemoveClip(clip.id);
                          }}
                          className="p-1 hover:text-rose-400 text-slate-300"
                          title="Delete clip"
                        >
                          <Trash2 className="w-3 h-3" />
                        </button>
                      </div>
                    </div>

                    {/* Bottom Row: Duration Badge & Filter Info */}
                    <div className="relative z-10 p-1.5 flex items-center justify-between">
                      <span className="text-[10px] font-mono px-1.5 py-0.5 rounded bg-black/60 text-brand-primary font-bold border border-brand-primary/30 backdrop-blur-sm">
                        {clip.duration >= 60
                          ? `${formatRulerTime(clip.duration)} (${clip.duration.toFixed(0)}s)`
                          : `${clip.duration.toFixed(1)}s`}
                      </span>
                      {clip.filter !== "none" && (
                        <span className="text-[9px] px-1 py-0.5 rounded bg-purple-900/60 text-purple-200 uppercase font-bold border border-purple-500/30">
                          {clip.filter}
                        </span>
                      )}
                    </div>
                  </div>

                  {/* Inter-clip Transition Badge */}
                  {idx < spec.clips.length - 1 && (
                    <div className="flex flex-col items-center justify-center shrink-0 z-10 px-0.5">
                      <select
                        value={clip.transition_out?.type || "fade"}
                        onChange={(e) => {
                          const newType = e.target.value as TransitionType;
                          onUpdateClip(clip.id, {
                            transition_out: {
                              type: newType,
                              duration: clip.transition_out?.duration || 0.8,
                            },
                          });
                        }}
                        className="bg-purple-950/90 hover:bg-purple-900 text-purple-200 border border-purple-500/40 rounded-md text-[10px] px-1 py-0.5 outline-none cursor-pointer shadow-sm transition-colors"
                      >
                        {TRANSITIONS.map((t) => (
                          <option key={t.type} value={t.type}>
                            {t.label}
                          </option>
                        ))}
                      </select>
                      {clip.transition_out && (
                        <span className="text-[9px] font-mono text-purple-300/80 mt-0.5">
                          {clip.transition_out.duration}s
                        </span>
                      )}
                    </div>
                  )}
                </React.Fragment>
              );
            })}

            {/* Quick Add Button */}
            <button
              type="button"
              onClick={onAddMediaClick}
              className="h-28 w-24 rounded-xl border border-dashed border-white/20 hover:border-brand-primary/60 bg-white/[0.03] hover:bg-white/[0.06] flex flex-col items-center justify-center gap-1.5 text-slate-400 hover:text-white transition-all shrink-0"
            >
              <Plus className="w-5 h-5 text-brand-primary" />
              <span className="text-[11px] font-medium">Add Clip</span>
            </button>
          </div>

          {/* Audio Sequencer Track Lane with Direct Timeline Dragging */}
          {(() => {
            const audioTotalDur =
              audioTrackDuration > 0 ? audioTrackDuration : Math.max(totalTime, 15);
            const audioTrackWidth = Math.max(
              200,
              Math.min(trackWidth, audioTotalDur * pxPerSec)
            );
            const playheadInAudio = Math.max(
              0,
              Math.min(audioTrackWidth, currentTime * pxPerSec)
            );

            // Left trim offset and trimmed active width in pixels
            const trimStartPx = activeTrimStart * pxPerSec;
            const activeWidthPx = Math.max(40, activePlayingAudioDur * pxPerSec);

            return (
              <div className="mt-4 flex flex-col gap-1.5">
                <div className="flex items-center justify-between pr-4">
                  <div className="flex items-center gap-2">
                    <span className="text-[10px] font-bold uppercase tracking-wider text-purple-400 bg-purple-950/80 px-2.5 py-1 rounded-md border border-purple-500/30">
                      Audio Track
                    </span>
                    {spec.audio?.key && (
                      <span className="text-xs text-white font-semibold truncate max-w-[200px]">
                        {spec.audio.title || spec.audio.key.split("/").pop()}
                      </span>
                    )}
                  </div>
                  {spec.audio?.key && (
                    <div className="flex items-center gap-2.5 text-xs font-mono">
                      <div className="flex items-center gap-1.5 px-2.5 py-1 rounded-lg bg-purple-950/90 border border-purple-500/40 text-purple-200 shadow-sm">
                        <span
                          className={`w-2 h-2 rounded-full ${
                            currentTime > 0 ? "bg-emerald-400 animate-pulse" : "bg-purple-400"
                          }`}
                        />
                        <span className="font-bold text-white">
                          Playing: {formatRulerTime(currentTime)} ({currentTime.toFixed(1)}s)
                        </span>
                        <span className="text-purple-400">/</span>
                        <span>
                          Total: {formatRulerTime(Math.round(audioTotalDur))} (
                          {audioTotalDur.toFixed(1)}s)
                        </span>
                      </div>
                      <span className="text-slate-400 text-[11px]">
                        Vol: {Math.round((spec.audio.volume ?? 1) * 100)}%
                      </span>
                    </div>
                  )}
                </div>

                {spec.audio?.key ? (
                  <div className="relative">
                    {/* Floating Drag HUD Tooltip */}
                    {dragMode && dragInfo && (
                      <div
                        className="absolute -top-9 z-40 px-3 py-1 rounded-xl bg-purple-600 text-white font-mono text-xs font-bold shadow-xl border border-white/20 flex items-center gap-2 animate-fade-in"
                        style={{
                          left: `${Math.max(10, trimStartPx)}px`,
                        }}
                      >
                        <Sliders className="w-3.5 h-3.5" />
                        <span>
                          Start: {formatDetailedTime(activeTrimStart)} | End:{" "}
                          {formatDetailedTime(activeTrimEnd)} | Duration:{" "}
                          {formatDetailedTime(activePlayingAudioDur)}
                        </span>
                      </div>
                    )}

                    {/* Audio track container */}
                    <div
                      onClick={() => onSelectAudio?.()}
                      style={{ width: `${audioTrackWidth}px` }}
                      className={`group relative h-24 rounded-xl bg-gradient-to-r from-purple-950/90 via-purple-900/60 to-purple-950/90 border cursor-pointer overflow-hidden shadow-xl transition-all flex flex-col justify-between p-2 select-none ${
                        isAudioSelected
                          ? "border-purple-400 ring-2 ring-purple-400/60 shadow-purple-500/30"
                          : "border-purple-500/40 hover:border-purple-400"
                      }`}
                    >
                      {/* Active played progress fill on audio track */}
                      <div
                        className="absolute inset-y-0 left-0 bg-purple-600/25 pointer-events-none border-r-2 border-purple-400 z-10 transition-none"
                        style={{ width: `${playheadInAudio}px` }}
                      >
                        <div className="absolute top-1 right-1 px-1.5 py-0.5 rounded bg-purple-500 text-white font-mono text-[9px] font-bold shadow-md whitespace-nowrap">
                          {currentTime.toFixed(1)}s
                        </div>
                      </div>

                      {/* Interactive Trimmable & Draggable Active Audio Block */}
                      <div
                        className={`absolute inset-y-1 rounded-lg border-2 z-20 flex items-center justify-between transition-colors ${
                          dragMode
                            ? "bg-purple-600/30 border-purple-300 ring-2 ring-purple-400/50"
                            : "bg-purple-600/20 hover:bg-purple-600/30 border-purple-400/70 cursor-grab active:cursor-grabbing"
                        }`}
                        style={{
                          left: `${trimStartPx}px`,
                          width: `${activeWidthPx}px`,
                        }}
                        onMouseDown={(e) => handleAudioMouseDown("body", e)}
                        title="Click and drag to slide audio track horizontally"
                      >
                        {/* Left Trim Handle */}
                        <div
                          onMouseDown={(e) => handleAudioMouseDown("left-trim", e)}
                          className="h-full w-4 bg-purple-500/60 hover:bg-purple-400 flex items-center justify-center cursor-ew-resize rounded-l-md transition-colors"
                          title="Drag to crop audio start"
                        >
                          <GripVertical className="w-3 h-3 text-white" />
                        </div>

                        {/* Center Drag Label */}
                        <div className="flex items-center gap-1.5 text-[10px] font-mono text-purple-200 pointer-events-none px-2 truncate">
                          <span className="font-bold text-white">
                            {formatRulerTime(activeTrimStart)}
                          </span>
                          <span>→</span>
                          <span>{formatRulerTime(activeTrimEnd)}</span>
                        </div>

                        {/* Right Trim Handle */}
                        <div
                          onMouseDown={(e) => handleAudioMouseDown("right-trim", e)}
                          className="h-full w-4 bg-purple-500/60 hover:bg-purple-400 flex items-center justify-center cursor-ew-resize rounded-r-md transition-colors"
                          title="Drag to crop audio end"
                        >
                          <GripVertical className="w-3 h-3 text-white" />
                        </div>
                      </div>

                      {/* Waveform graphic pattern */}
                      <div className="absolute inset-x-2 top-3 bottom-7 flex items-center gap-[3px] opacity-35 group-hover:opacity-50 transition-opacity pointer-events-none overflow-hidden">
                        {Array.from({ length: Math.min(180, Math.ceil(audioTrackWidth / 5)) }).map(
                          (_, i) => {
                            const h = Math.round(
                              12 + Math.sin(i * 0.45) * 10 + Math.cos(i * 0.9) * 8
                            );
                            return (
                              <div
                                key={i}
                                className="w-[3px] rounded-full bg-purple-400"
                                style={{ height: `${Math.max(6, Math.min(26, h))}px` }}
                              />
                            );
                          }
                        )}
                      </div>

                      {/* Top info row */}
                      <div className="relative z-10 flex items-center justify-between">
                        <div className="flex items-center gap-2">
                          <div className="w-5 h-5 rounded-md bg-purple-600/40 flex items-center justify-center text-purple-200">
                            <Music className="w-3 h-3" />
                          </div>
                          <span className="text-xs font-bold text-white truncate max-w-[220px]">
                            {spec.audio.title || spec.audio.key.split("/").pop()}
                          </span>
                          <span className="text-[10px] px-2 py-0.5 rounded bg-purple-900/80 border border-purple-400/30 text-purple-200 font-mono">
                            {formatRulerTime(Math.round(audioTotalDur))} (
                            {audioTotalDur.toFixed(1)}s)
                          </span>
                        </div>

                        <div className="flex items-center gap-2">
                          <span className="text-[10px] px-2 py-0.5 rounded bg-black/40 text-slate-300 font-mono">
                            Vol: {Math.round((spec.audio.volume ?? 1) * 100)}%
                          </span>
                          <button
                            type="button"
                            onClick={(e) => {
                              e.stopPropagation();
                              onAudioClick();
                            }}
                            className="text-[10px] px-2.5 py-1 rounded bg-purple-600/50 hover:bg-purple-600 text-white font-medium transition-colors"
                          >
                            Change Track
                          </button>
                        </div>
                      </div>

                      {/* Bottom second & minute time scale ruler */}
                      <div className="relative z-10 flex items-center justify-between text-[10px] font-mono text-purple-200/90 border-t border-purple-500/30 pt-1">
                        <span className="text-[9px] text-purple-300/80">
                          Trim: {formatDetailedTime(activeTrimStart)} →{" "}
                          {formatDetailedTime(activeTrimEnd)} (
                          {activePlayingAudioDur.toFixed(1)}s)
                        </span>
                        <span className="shrink-0 font-bold text-white bg-purple-950/80 px-2 py-0.5 rounded border border-purple-400/30 text-[9px]">
                          End: {formatRulerTime(Math.round(audioTotalDur))}
                        </span>
                      </div>
                    </div>
                  </div>
                ) : (
                  <div
                    onClick={onAudioClick}
                    className="h-14 w-80 rounded-xl border border-dashed border-purple-500/30 hover:border-purple-400/80 bg-purple-950/20 hover:bg-purple-950/40 flex items-center justify-center gap-2 cursor-pointer text-purple-300 hover:text-white transition-all text-xs"
                  >
                    <Music className="w-4 h-4 text-purple-400" />
                    <span className="font-semibold">+ Add Background Music / Audio Track</span>
                  </div>
                )}
              </div>
            );
          })()}
        </div>
      </div>
    </div>
  );
};
