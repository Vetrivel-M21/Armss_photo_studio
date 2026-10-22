import React, { useEffect, useRef, useState, useCallback, useMemo } from "react";
import type { VideoSpec, ClipSpec } from "../engine/spec";
import {
  totalDuration,
  clipStartTime,
  textAlpha,
} from "../engine/timing";
import {
  resolveFilter,
  toCSSFilter,
  kenBurnsFrame,
  kenBurnsToPixels,
  containLayout,
} from "../engine/effects";
import { Play, Pause, RotateCcw, Volume2, VolumeX } from "lucide-react";
import { resolveMediaUrl } from "../utils/mediaUtils";

interface CanvasPreviewProps {
  spec: VideoSpec;
  currentTime: number;
  isPlaying: boolean;
  onTimeUpdate: (time: number) => void;
  onTogglePlay: () => void;
  onSelectClip?: (clipId: string) => void;
  activeClipId?: string | null;
}

export const CanvasPreview: React.FC<CanvasPreviewProps> = ({
  spec,
  currentTime,
  isPlaying,
  onTimeUpdate,
  onTogglePlay,
}) => {
  const canvasRef = useRef<HTMLCanvasElement | null>(null);
  const containerRef = useRef<HTMLDivElement | null>(null);
  const imageCacheRef = useRef<Map<string, HTMLImageElement>>(new Map());
  const [imagesLoadedVersion, setImagesLoadedVersion] = useState(0);
  const [isMuted, setIsMuted] = useState(false);
  const audioRef = useRef<HTMLAudioElement | null>(null);
  const currentAudioKeyRef = useRef<string | null>(null);
  const currentTimeRef = useRef(currentTime);

  // Playable duration: max of clips duration or audio duration
  const clipsDuration = useMemo(() => totalDuration(spec.clips), [spec.clips]);
  const duration = useMemo(() => {
    const audioDur = spec.audio?.duration || 0;
    return Math.max(clipsDuration, audioDur, 1);
  }, [clipsDuration, spec.audio?.duration]);

  // Dimension determination
  const { width: outW, height: outH } = useMemo(() => {
    switch (spec.output.aspect) {
      case "16:9":
        return { width: 1920, height: 1080 };
      case "1:1":
        return { width: 1080, height: 1080 };
      case "9:16":
      default:
        return { width: 1080, height: 1920 };
    }
  }, [spec.output.aspect]);

  // Preload images into cache and trigger re-render on load
  useEffect(() => {
    spec.clips.forEach((clip) => {
      const url = resolveMediaUrl(clip.image_key);
      if (url && !imageCacheRef.current.has(url)) {
        const img = new Image();
        img.crossOrigin = "anonymous";
        img.onload = () => {
          setImagesLoadedVersion((v) => v + 1);
        };
        img.onerror = () => {
          if (img.crossOrigin) {
            const fallbackImg = new Image();
            fallbackImg.onload = () => {
              imageCacheRef.current.set(url, fallbackImg);
              setImagesLoadedVersion((v) => v + 1);
            };
            fallbackImg.src = url;
          }
        };
        img.src = url;
        imageCacheRef.current.set(url, img);
      }
    });
  }, [spec.clips]);

  // Audio track element creation and volume management
  useEffect(() => {
    const audioKey = spec.audio?.key;
    if (!audioKey) {
      if (audioRef.current) {
        audioRef.current.pause();
        audioRef.current.src = "";
        audioRef.current = null;
      }
      currentAudioKeyRef.current = null;
      return;
    }

    const audioUrl = resolveMediaUrl(audioKey);
    if (currentAudioKeyRef.current !== audioKey || !audioRef.current) {
      if (audioRef.current) {
        audioRef.current.pause();
      }
      const a = new Audio(audioUrl);
      a.preload = "auto";
      const vol = isMuted ? 0 : Math.min(1, Math.max(0, spec.audio?.volume ?? 1));
      a.volume = vol;
      audioRef.current = a;
      currentAudioKeyRef.current = audioKey;
    } else {
      const vol = isMuted ? 0 : Math.min(1, Math.max(0, spec.audio?.volume ?? 1));
      audioRef.current.volume = vol;
    }
  }, [spec.audio?.key, spec.audio?.volume, isMuted]);

  // Audio play/pause state synchronization (triggered strictly when isPlaying changes)
  useEffect(() => {
    const a = audioRef.current;
    if (!a) return;

    if (isPlaying) {
      const trimStart = spec.audio?.trim_start || 0;
      const targetTime = trimStart + currentTimeRef.current;
      try {
        a.currentTime = Math.max(0, targetTime);
      } catch {}
      const playPromise = a.play();
      if (playPromise !== undefined) {
        playPromise.catch((err) => console.warn("Audio autoplay prevented:", err));
      }
    } else {
      a.pause();
    }
  }, [isPlaying, spec.audio?.trim_start]);

  // Scrub synchronization: when paused and currentTime moves (slider or timeline click)
  useEffect(() => {
    const a = audioRef.current;
    if (!a || isPlaying) return;
    const trimStart = spec.audio?.trim_start || 0;
    try {
      a.currentTime = Math.max(0, trimStart + currentTime);
    } catch {}
  }, [currentTime, isPlaying, spec.audio?.trim_start]);


  // Main Canvas Render function for a given time
  const renderFrameAtTime = useCallback((time: number) => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    const ctx = canvas.getContext("2d");
    if (!ctx) return;

    // Clear background
    ctx.fillStyle = "#000000";
    ctx.fillRect(0, 0, canvas.width, canvas.height);

    if (spec.clips.length === 0) {
      // Empty or audio visualizer state
      ctx.fillStyle = "#0f172a";
      ctx.fillRect(0, 0, canvas.width, canvas.height);

      if (spec.audio?.key) {
        ctx.fillStyle = "#c084fc";
        ctx.font = "bold 40px sans-serif";
        ctx.textAlign = "center";
        ctx.textBaseline = "middle";
        ctx.fillText("🎵 Audio Track Active", canvas.width / 2, canvas.height / 2 - 30);
        ctx.fillStyle = "#e2e8f0";
        ctx.font = "24px sans-serif";
        ctx.fillText(spec.audio.title || "Background Music", canvas.width / 2, canvas.height / 2 + 20);
        ctx.fillStyle = "#94a3b8";
        ctx.font = "18px sans-serif";
        ctx.fillText("Add photo clips below to build your video", canvas.width / 2, canvas.height / 2 + 65);
      } else {
        ctx.fillStyle = "#475569";
        ctx.font = "bold 36px sans-serif";
        ctx.textAlign = "center";
        ctx.textBaseline = "middle";
        ctx.fillText("No clips added to timeline", canvas.width / 2, canvas.height / 2);
        ctx.font = "24px sans-serif";
        ctx.fillText("Click '+ Add Media' to begin", canvas.width / 2, canvas.height / 2 + 50);
      }
      return;
    }

    // Find active clip & transition
    let activeClipIdx = 0;
    let clipStart = 0;
    for (let i = 0; i < spec.clips.length; i++) {
      const c = spec.clips[i];
      const start = clipStartTime(spec.clips, i);
      const end = start + c.duration;
      if (time >= start && time < end) {
        activeClipIdx = i;
        clipStart = start;
        break;
      }
      if (i === spec.clips.length - 1) {
        activeClipIdx = i;
        clipStart = start;
      }
    }

    const currentClip = spec.clips[activeClipIdx];
    const clipLocalTime = Math.min(currentClip.duration, Math.max(0, time - clipStart));

    // Draw single clip helper
    const drawClip = (clip: ClipSpec, localTime: number, alpha: number) => {
      const url = resolveMediaUrl(clip.image_key);
      const img = imageCacheRef.current.get(url);
      if (!img || !img.complete || img.naturalWidth === 0) {
        ctx.fillStyle = "#1e293b";
        ctx.fillRect(0, 0, canvas.width, canvas.height);
        ctx.fillStyle = "#94a3b8";
        ctx.font = "28px sans-serif";
        ctx.textAlign = "center";
        ctx.fillText("Loading image...", canvas.width / 2, canvas.height / 2);
        return;
      }

      ctx.save();
      ctx.globalAlpha = alpha;

      // CSS filter
      const resolved = resolveFilter(clip.filter, clip.adjust);
      ctx.filter = toCSSFilter(resolved);

      const imgW = img.naturalWidth;
      const imgH = img.naturalHeight;

      // Fit mode
      if (spec.output.fit === "contain") {
        // Blurred background
        const layout = containLayout(imgW, imgH, canvas.width, canvas.height);
        ctx.save();
        ctx.filter = `blur(${Math.max(10, Math.round(canvas.width / 30))}px)`;
        ctx.drawImage(img, 0, 0, canvas.width, canvas.height);
        ctx.restore();

        // Foreground sharp image
        ctx.drawImage(
          img,
          0,
          0,
          imgW,
          imgH,
          layout.fgRect.x,
          layout.fgRect.y,
          layout.fgRect.w,
          layout.fgRect.h
        );
      } else {
        // "cover" with optional Ken Burns
        if (clip.ken_burns) {
          const u = clip.duration > 0 ? localTime / clip.duration : 0;
          const frame = kenBurnsFrame(clip.ken_burns, u);
          const rect = kenBurnsToPixels(frame, canvas.width, canvas.height, imgW, imgH);
          ctx.drawImage(img, rect.px, rect.py, rect.pw, rect.ph, 0, 0, canvas.width, canvas.height);
        } else {
          // Normal cover crop
          const targetAspect = canvas.width / canvas.height;
          const imgAspect = imgW / imgH;
          let sx = 0, sy = 0, sw = imgW, sh = imgH;
          if (imgAspect > targetAspect) {
            sw = imgH * targetAspect;
            sx = (imgW - sw) / 2;
          } else {
            sh = imgW / targetAspect;
            sy = (imgH - sh) / 2;
          }
          ctx.drawImage(img, sx, sy, sw, sh, 0, 0, canvas.width, canvas.height);
        }
      }

      ctx.restore();

      // Render Text Overlays
      if (clip.texts && clip.texts.length > 0) {
        clip.texts.forEach((text) => {
          const tAlpha = textAlpha(0, text, localTime, clip.duration);
          if (tAlpha <= 0) return;

          ctx.save();
          ctx.globalAlpha = alpha * tAlpha * (text.opacity ?? 1);

          // Scaled font size relative to canvas dimension
          const scaleFactor = canvas.width / 1080;
          const pxSize = Math.round((text.size || 48) * scaleFactor);
          ctx.font = `bold ${pxSize}px "${text.font || 'Inter'}", sans-serif`;
          ctx.fillStyle = text.color || "#FFFFFF";
          ctx.textBaseline = "top";
          ctx.textAlign = "left";

          let posX = (text.x ?? 0.1) * canvas.width;
          let posY = (text.y ?? 0.8) * canvas.height;

          // Animation slide offset
          if (text.anim === "slide-up") {
            const progress = text.anim_duration > 0 ? Math.min(1, Math.max(0, (localTime - text.start) / text.anim_duration)) : 1;
            const slideOffset = (1 - progress) * 40 * scaleFactor;
            posY += slideOffset;
          }

          // Subtle text shadow for legibility over any photo
          ctx.shadowColor = "rgba(0, 0, 0, 0.8)";
          ctx.shadowBlur = 12 * scaleFactor;
          ctx.shadowOffsetX = 0;
          ctx.shadowOffsetY = 2 * scaleFactor;

          ctx.fillText(text.text, posX, posY);
          ctx.restore();
        });
      }
    };

    // Check transition overlap with next clip
    const nextClipIdx = activeClipIdx + 1;
    const trans = currentClip.transition_out;
    const hasTransition = trans && trans.duration > 0 && nextClipIdx < spec.clips.length;
    const transitionWindowStart = currentClip.duration - (hasTransition ? trans.duration : 0);

    if (hasTransition && clipLocalTime >= transitionWindowStart) {
      const transLocalTime = clipLocalTime - transitionWindowStart;
      const progress = Math.min(1, Math.max(0, transLocalTime / trans.duration));
      const nextClip = spec.clips[nextClipIdx];

      switch (trans.type) {
        case "fade":
        case "dissolve":
          drawClip(currentClip, clipLocalTime, 1 - progress);
          drawClip(nextClip, transLocalTime, progress);
          break;
        case "slide-left": {
          ctx.save();
          drawClip(currentClip, clipLocalTime, 1);
          ctx.restore();
          ctx.save();
          const offsetX = canvas.width * (1 - progress);
          ctx.translate(offsetX, 0);
          drawClip(nextClip, transLocalTime, 1);
          ctx.restore();
          break;
        }
        case "slide-right": {
          ctx.save();
          drawClip(currentClip, clipLocalTime, 1);
          ctx.restore();
          ctx.save();
          const offsetX = -canvas.width * (1 - progress);
          ctx.translate(offsetX, 0);
          drawClip(nextClip, transLocalTime, 1);
          ctx.restore();
          break;
        }
        default:
          drawClip(currentClip, clipLocalTime, 1 - progress);
          drawClip(nextClip, transLocalTime, progress);
          break;
      }
    } else {
      drawClip(currentClip, clipLocalTime, 1);
    }
  }, [spec, imagesLoadedVersion]);

  // Keep ref synchronized and render when currentTime changes (slider drag or timeline click)
  useEffect(() => {
    currentTimeRef.current = currentTime;
    renderFrameAtTime(currentTime);
  }, [currentTime, renderFrameAtTime]);

  // Animation frame loop for continuous 60fps playback without tearing down on state updates
  useEffect(() => {
    if (!isPlaying) return;

    let animId: number;
    let lastStamp = performance.now();

    const loop = (stamp: number) => {
      const delta = (stamp - lastStamp) / 1000;
      lastStamp = stamp;

      let nextTime = currentTimeRef.current + delta;
      const a = audioRef.current;
      if (a && !a.paused && a.duration > 0) {
        const trimStart = spec.audio?.trim_start || 0;
        const audioTime = a.currentTime - trimStart;
        if (Math.abs(audioTime - nextTime) < 0.4 && audioTime >= 0) {
          nextTime = audioTime;
        }
      }

      if (nextTime >= duration) {
        currentTimeRef.current = 0;
        onTimeUpdate(0);
        renderFrameAtTime(0);
        if (a) {
          a.pause();
          a.currentTime = spec.audio?.trim_start || 0;
        }
        onTogglePlay();
      } else {
        currentTimeRef.current = nextTime;
        renderFrameAtTime(nextTime);
        onTimeUpdate(nextTime);
        animId = requestAnimationFrame(loop);
      }
    };

    animId = requestAnimationFrame(loop);
    return () => cancelAnimationFrame(animId);
  }, [isPlaying, duration, onTimeUpdate, onTogglePlay, renderFrameAtTime, spec.audio?.trim_start]);

  // Aspect ratio helper CSS class
  const aspectClass = useMemo(() => {
    switch (spec.output.aspect) {
      case "16:9":
        return "aspect-video max-h-[500px]";
      case "1:1":
        return "aspect-square max-h-[520px]";
      case "9:16":
      default:
        return "aspect-[9/16] max-h-[580px]";
    }
  }, [spec.output.aspect]);

  const formatSeconds = (sec: number) => {
    const m = Math.floor(sec / 60);
    const s = Math.floor(sec % 60);
    const ms = Math.floor((sec % 1) * 10);
    return `${m.toString().padStart(2, "0")}:${s.toString().padStart(2, "0")}.${ms}`;
  };

  return (
    <div className="flex flex-col items-center justify-center w-full bg-slate-950/60 rounded-2xl border border-white/10 p-4 shadow-2xl backdrop-blur-md">
      {/* Canvas container */}
      <div
        ref={containerRef}
        className={`relative w-full ${aspectClass} max-w-full flex items-center justify-center rounded-xl overflow-hidden bg-black/90 shadow-inner group border border-white/5`}
      >
        <canvas
          ref={canvasRef}
          width={outW}
          height={outH}
          className="w-full h-full object-contain cursor-pointer"
          onClick={onTogglePlay}
        />

        {/* Hover play/pause overlay indicator */}
        <div
          className="absolute inset-0 flex items-center justify-center pointer-events-none opacity-0 group-hover:opacity-100 transition-opacity bg-black/20"
        >
          <div className="w-14 h-14 rounded-full bg-black/60 backdrop-blur-md border border-white/20 flex items-center justify-center text-white shadow-xl">
            {isPlaying ? <Pause className="w-6 h-6" /> : <Play className="w-6 h-6 ml-0.5" />}
          </div>
        </div>

        {/* Top Badges */}
        <div className="absolute top-3 left-3 flex items-center gap-2 pointer-events-none">
          <span className="px-2.5 py-1 text-xs font-semibold rounded-md bg-black/70 border border-white/15 text-slate-200 backdrop-blur-md">
            {spec.output.aspect}
          </span>
          <span className="px-2.5 py-1 text-xs font-semibold rounded-md bg-brand-primary/80 text-white backdrop-blur-md">
            {spec.output.fps} FPS
          </span>
        </div>
      </div>

      {/* Control bar */}
      <div className="w-full mt-3 flex items-center justify-between gap-4 px-2">
        <div className="flex items-center gap-2">
          <button
            type="button"
            onClick={onTogglePlay}
            className="p-2.5 rounded-xl bg-white/10 hover:bg-white/20 text-white transition-colors"
            title={isPlaying ? "Pause (Space)" : "Play (Space)"}
          >
            {isPlaying ? <Pause className="w-4 h-4" /> : <Play className="w-4 h-4 ml-0.5" />}
          </button>

          <button
            type="button"
            onClick={() => onTimeUpdate(0)}
            className="p-2.5 rounded-xl bg-white/5 hover:bg-white/15 text-slate-300 hover:text-white transition-colors"
            title="Restart from beginning"
          >
            <RotateCcw className="w-4 h-4" />
          </button>

          {spec.audio?.key && (
            <button
              type="button"
              onClick={() => setIsMuted((m) => !m)}
              className="p-2.5 rounded-xl bg-white/5 hover:bg-white/15 text-slate-300 hover:text-white transition-colors"
              title={isMuted ? "Unmute" : "Mute"}
            >
              {isMuted ? <VolumeX className="w-4 h-4 text-rose-400" /> : <Volume2 className="w-4 h-4" />}
            </button>
          )}

          <div className="text-xs font-mono text-slate-300 ml-1">
            <span className="text-white font-semibold">{formatSeconds(currentTime)}</span>
            <span className="text-slate-500"> / {formatSeconds(duration)}</span>
          </div>
        </div>

        {/* Scrubber timeline mini-bar */}
        <div className="flex-1 max-w-md mx-3">
          <input
            type="range"
            min={0}
            max={Math.max(0.1, duration)}
            step={0.05}
            value={currentTime}
            onChange={(e) => onTimeUpdate(parseFloat(e.target.value))}
            className="w-full accent-brand-primary cursor-pointer h-1.5 bg-white/10 rounded-lg"
          />
        </div>

        <div className="flex items-center gap-2 text-xs text-slate-400">
          <span>{spec.clips.length} clip{spec.clips.length === 1 ? "" : "s"}</span>
        </div>
      </div>
    </div>
  );
};
