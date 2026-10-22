import React from "react";
import type { AudioSpec } from "../engine/spec";
import {
  Music,
  Volume2,
  VolumeX,
  Clock,
  Trash2,
  Upload,
  Scissors,
  Wand2,
} from "lucide-react";

interface AudioInspectorProps {
  audio: AudioSpec | undefined;
  currentTime?: number;
  totalVideoTime: number;
  onUpdateAudio: (patch: Partial<AudioSpec>) => void;
  onRemoveAudio: () => void;
  onChangeAudioClick: () => void;
  onFitVideoToAudio?: () => void;
}

function formatTimeMMSS(sec: number): string {
  if (isNaN(sec) || sec < 0) sec = 0;
  const m = Math.floor(sec / 60);
  const s = Math.floor(sec % 60);
  const ms = Math.floor((sec % 1) * 10);
  return `${m.toString().padStart(2, "0")}:${s.toString().padStart(2, "0")}.${ms}`;
}

export const AudioInspector: React.FC<AudioInspectorProps> = ({
  audio,
  currentTime,
  totalVideoTime,
  onUpdateAudio,
  onRemoveAudio,
  onChangeAudioClick,
  onFitVideoToAudio,
}) => {
  if (!audio?.key) {
    return (
      <div className="w-full bg-slate-900/80 border border-white/10 rounded-2xl p-6 shadow-xl flex flex-col items-center justify-center text-center gap-3 backdrop-blur-md min-h-[380px]">
        <div className="w-12 h-12 rounded-2xl bg-purple-500/10 border border-purple-500/20 flex items-center justify-center text-purple-400">
          <Music className="w-6 h-6" />
        </div>
        <h4 className="text-sm font-bold text-white">No Audio Track Selected</h4>
        <p className="text-xs text-slate-400 max-w-xs">
          Add background music or an audio track from your computer or royalty-free library to edit volume, fades, and timing.
        </p>
        <button
          type="button"
          onClick={onChangeAudioClick}
          className="mt-2 flex items-center gap-2 px-4 py-2 rounded-xl text-xs font-semibold bg-purple-600 hover:bg-purple-500 text-white transition-all shadow-glow-sm"
        >
          <Upload className="w-3.5 h-3.5" />
          Choose Audio Track
        </button>
      </div>
    );
  }

  const trackDuration = audio.duration && audio.duration > 0 ? audio.duration : Math.max(totalVideoTime, 30);
  const trimStart = audio.trim_start || 0;
  const trimEnd = audio.trim_end || 0;
  const effectiveEnd = trimEnd > trimStart ? trimEnd : trackDuration;
  const effectiveDuration = Math.max(0, effectiveEnd - trimStart);

  return (
    <div className="w-full bg-slate-900/80 border border-white/10 rounded-2xl p-4 shadow-xl flex flex-col gap-4 backdrop-blur-md animate-fade-in">
      {/* Header */}
      <div className="flex items-center justify-between border-b border-white/10 pb-3">
        <div className="flex items-center gap-2.5">
          <div className="w-8 h-8 rounded-xl bg-purple-500/20 border border-purple-500/30 flex items-center justify-center text-purple-300">
            <Music className="w-4 h-4" />
          </div>
          <div className="flex flex-col">
            <h3 className="text-sm font-bold text-white truncate max-w-[190px]">
              {audio.title || audio.key.split("/").pop()}
            </h3>
            <span className="text-[11px] font-mono text-purple-300">
              Total Duration: {formatTimeMMSS(trackDuration)}
            </span>
          </div>
        </div>

        <div className="flex items-center gap-1.5">
          <button
            type="button"
            onClick={onChangeAudioClick}
            className="text-[11px] px-2.5 py-1 rounded-lg bg-white/5 hover:bg-white/10 text-slate-200 border border-white/10 transition-colors"
            title="Replace audio file"
          >
            Replace
          </button>
          <button
            type="button"
            onClick={onRemoveAudio}
            className="p-1.5 rounded-lg bg-rose-500/20 hover:bg-rose-500 text-rose-300 hover:text-white transition-colors"
            title="Remove audio track"
          >
            <Trash2 className="w-3.5 h-3.5" />
          </button>
        </div>
      </div>

      {/* Timing and Trimming Section */}
      <div className="flex flex-col gap-3 p-3 rounded-xl bg-white/[0.02] border border-white/10">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-1.5 text-xs font-semibold text-white">
            <Clock className="w-3.5 h-3.5 text-purple-400" />
            <span>Audio Timing & Trimming</span>
          </div>
          <span className="text-[11px] font-mono text-purple-300 font-semibold bg-purple-950/80 px-2 py-0.5 rounded border border-purple-500/30">
            Playing: {effectiveDuration.toFixed(1)}s ({formatTimeMMSS(effectiveDuration)})
          </span>
        </div>

        {/* Live Playhead Position */}
        {currentTime !== undefined && (
          <div className="flex items-center justify-between p-2 rounded-lg bg-purple-950/70 border border-purple-500/30 text-xs">
            <span className="text-slate-300 flex items-center gap-1.5">
              <span className={`w-2 h-2 rounded-full ${currentTime > 0 ? "bg-emerald-400 animate-pulse" : "bg-purple-400"}`} />
              Now Playing Second:
            </span>
            <span className="font-mono text-emerald-300 font-bold text-xs bg-black/40 px-2 py-0.5 rounded border border-emerald-500/30">
              {currentTime.toFixed(1)}s ({formatTimeMMSS(currentTime)})
            </span>
          </div>
        )}

        {/* Trim Start Slider & Numeric Input */}
        <div className="flex flex-col gap-1 text-xs">
          <div className="flex items-center justify-between text-slate-300">
            <span className="flex items-center gap-1 font-medium">
              <Scissors className="w-3 h-3 text-purple-400" />
              Trim Start (Skip Intro)
            </span>
            <div className="flex items-center gap-1">
              <input
                type="number"
                min={0}
                max={Math.max(0, trackDuration - 1)}
                step={0.5}
                value={trimStart}
                onChange={(e) => {
                  const val = parseFloat(e.target.value);
                  if (!isNaN(val) && val >= 0 && val < trackDuration) {
                    onUpdateAudio({ trim_start: val });
                  }
                }}
                className="w-14 px-1 py-0.5 rounded bg-white/10 text-white font-mono text-[11px] text-right border border-white/20 focus:border-purple-400 outline-none"
              />
              <span className="font-mono text-purple-300 text-[10px]">s ({formatTimeMMSS(trimStart)})</span>
            </div>
          </div>
          <input
            type="range"
            min={0}
            max={Math.max(0, trackDuration - 1)}
            step={0.5}
            value={trimStart}
            onChange={(e) => {
              const val = parseFloat(e.target.value);
              onUpdateAudio({ trim_start: val });
            }}
            className="accent-purple-500 w-full cursor-pointer h-1.5 bg-white/10 rounded-lg"
          />
        </div>

        {/* Trim End Slider & Numeric Input */}
        <div className="flex flex-col gap-1 text-xs">
          <div className="flex items-center justify-between text-slate-300">
            <span className="flex items-center gap-1 font-medium">
              <Scissors className="w-3 h-3 text-purple-400 rotate-180" />
              Trim End (Cut Outro)
            </span>
            <div className="flex items-center gap-1">
              <input
                type="number"
                min={0}
                max={trackDuration}
                step={0.5}
                value={trimEnd || trackDuration}
                onChange={(e) => {
                  const val = parseFloat(e.target.value);
                  if (!isNaN(val) && val >= 0 && val <= trackDuration) {
                    onUpdateAudio({ trim_end: val >= trackDuration ? 0 : val });
                  }
                }}
                className="w-14 px-1 py-0.5 rounded bg-white/10 text-white font-mono text-[11px] text-right border border-white/20 focus:border-purple-400 outline-none"
              />
              <span className="font-mono text-purple-300 text-[10px]">
                {trimEnd > 0 ? `s (${formatTimeMMSS(trimEnd)})` : `Full (${trackDuration.toFixed(0)}s)`}
              </span>
            </div>
          </div>
          <input
            type="range"
            min={0}
            max={trackDuration}
            step={0.5}
            value={trimEnd || trackDuration}
            onChange={(e) => {
              const val = parseFloat(e.target.value);
              onUpdateAudio({ trim_end: val >= trackDuration ? 0 : val });
            }}
            className="accent-purple-500 w-full cursor-pointer h-1.5 bg-white/10 rounded-lg"
          />
        </div>

        {/* Smart Video Sync Helpers */}
        <div className="flex items-center gap-2 pt-1 border-t border-white/5">
          {onFitVideoToAudio && (
            <button
              type="button"
              onClick={onFitVideoToAudio}
              className="flex-1 flex items-center justify-center gap-1.5 py-1.5 px-2 rounded-lg bg-purple-900/30 hover:bg-purple-900/50 border border-purple-500/30 text-[11px] font-semibold text-purple-200 transition-colors"
              title="Automatically distribute photo durations to match audio length"
            >
              <Wand2 className="w-3 h-3 text-purple-400" />
              Fit Video to Music
            </button>
          )}

          <button
            type="button"
            onClick={() => {
              onUpdateAudio({ trim_end: totalVideoTime });
            }}
            className="flex-1 flex items-center justify-center gap-1.5 py-1.5 px-2 rounded-lg bg-white/5 hover:bg-white/10 border border-white/10 text-[11px] font-semibold text-slate-300 transition-colors"
            title="Cut audio so it stops exactly when the last photo ends"
          >
            <Scissors className="w-3 h-3 text-slate-400" />
            Trim to Video End
          </button>
        </div>
      </div>

      {/* Volume and Dynamics */}
      <div className="flex flex-col gap-3 p-3 rounded-xl bg-white/[0.02] border border-white/10">
        <div className="flex items-center justify-between text-xs">
          <div className="flex items-center gap-1.5 font-semibold text-white">
            <Volume2 className="w-3.5 h-3.5 text-purple-400" />
            <span>Volume Level</span>
          </div>
          <div className="flex items-center gap-2">
            <button
              type="button"
              onClick={() => onUpdateAudio({ volume: (audio.volume ?? 1) > 0 ? 0 : 1 })}
              className="p-1 rounded bg-white/5 hover:bg-white/10 text-slate-300 transition-colors"
              title={(audio.volume ?? 1) > 0 ? "Mute" : "Unmute"}
            >
              {(audio.volume ?? 1) > 0 ? <Volume2 className="w-3 h-3" /> : <VolumeX className="w-3 h-3 text-rose-400" />}
            </button>
            <span className="font-mono text-purple-300 font-bold text-xs">
              {Math.round((audio.volume ?? 1) * 100)}%
            </span>
          </div>
        </div>

        <input
          type="range"
          min={0}
          max={1.5}
          step={0.05}
          value={audio.volume ?? 1}
          onChange={(e) => onUpdateAudio({ volume: parseFloat(e.target.value) })}
          className="accent-purple-500 w-full cursor-pointer h-1.5 bg-white/10 rounded-lg"
        />
      </div>

      {/* Fades */}
      <div className="grid grid-cols-2 gap-3 p-3 rounded-xl bg-white/[0.02] border border-white/10 text-xs">
        <div className="flex flex-col gap-1.5">
          <div className="flex items-center justify-between text-slate-300">
            <span>Fade In</span>
            <span className="font-mono text-purple-300 font-bold">
              {(audio.fade_in ?? 1).toFixed(1)}s
            </span>
          </div>
          <input
            type="range"
            min={0}
            max={4}
            step={0.5}
            value={audio.fade_in ?? 1}
            onChange={(e) => onUpdateAudio({ fade_in: parseFloat(e.target.value) })}
            className="accent-purple-500 w-full cursor-pointer h-1.5 bg-white/10 rounded-lg"
          />
        </div>

        <div className="flex flex-col gap-1.5">
          <div className="flex items-center justify-between text-slate-300">
            <span>Fade Out</span>
            <span className="font-mono text-purple-300 font-bold">
              {(audio.fade_out ?? 1).toFixed(1)}s
            </span>
          </div>
          <input
            type="range"
            min={0}
            max={4}
            step={0.5}
            value={audio.fade_out ?? 1}
            onChange={(e) => onUpdateAudio({ fade_out: parseFloat(e.target.value) })}
            className="accent-purple-500 w-full cursor-pointer h-1.5 bg-white/10 rounded-lg"
          />
        </div>
      </div>
    </div>
  );
};
