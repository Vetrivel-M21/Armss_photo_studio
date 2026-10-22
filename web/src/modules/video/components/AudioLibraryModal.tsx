import React, { useState, useRef } from "react";
import type { AudioSpec } from "../engine/spec";
import { videoApi } from "../api/client";
import {
  Music,
  Play,
  Pause,
  X,
  Trash2,
  Check,
  Upload,
  Sparkles,
  Loader2,
  FileAudio,
} from "lucide-react";

interface AudioLibraryModalProps {
  isOpen: boolean;
  onClose: () => void;
  currentAudio?: AudioSpec;
  onSetAudio: (audio: AudioSpec | undefined) => void;
}

const PRESET_TRACKS = [
  {
    key: "video/presets/audio_lofi_chill.mp3",
    title: "Midnight Lofi Chill",
    artist: "ARMSS Studio Audio",
    duration: 64,
    genre: "Lofi / Relaxed",
    url: "https://cdn.freesound.org/previews/612/612642_11861866-lq.mp3",
  },
  {
    key: "video/presets/audio_cinematic_ambient.mp3",
    title: "Cinematic Horizons",
    artist: "ARMSS Studio Audio",
    duration: 82,
    genre: "Ambient / Inspiring",
    url: "https://cdn.freesound.org/previews/648/648455_1386366-lq.mp3",
  },
  {
    key: "video/presets/audio_upbeat_reels.mp3",
    title: "Modern Groove Beat",
    artist: "ARMSS Studio Audio",
    duration: 48,
    genre: "Reels / Upbeat",
    url: "https://cdn.freesound.org/previews/573/573381_11861866-lq.mp3",
  },
  {
    key: "video/presets/audio_acoustic_warmth.mp3",
    title: "Acoustic Warmth",
    artist: "ARMSS Studio Audio",
    duration: 75,
    genre: "Acoustic / Warm",
    url: "https://cdn.freesound.org/previews/518/518305_2402876-lq.mp3",
  },
];

export const AudioLibraryModal: React.FC<AudioLibraryModalProps> = ({
  isOpen,
  onClose,
  currentAudio,
  onSetAudio,
}) => {
  const [activeTab, setActiveTab] = useState<"upload" | "presets">("upload");
  const [playingKey, setPlayingKey] = useState<string | null>(null);
  const [volume, setVolume] = useState<number>(currentAudio?.volume ?? 1);
  const [fadeIn, setFadeIn] = useState<number>(currentAudio?.fade_in ?? 1);
  const [fadeOut, setFadeOut] = useState<number>(currentAudio?.fade_out ?? 1);

  // Uploading state
  const [isUploading, setIsUploading] = useState(false);
  const [uploadError, setUploadError] = useState<string | null>(null);
  const fileInputRef = useRef<HTMLInputElement | null>(null);

  const previewAudioRef = useRef<HTMLAudioElement | null>(null);

  if (!isOpen) return null;

  const togglePreview = (url: string, key: string) => {
    if (playingKey === key) {
      previewAudioRef.current?.pause();
      setPlayingKey(null);
    } else {
      if (previewAudioRef.current) {
        previewAudioRef.current.pause();
      }
      const audio = new Audio(url);
      audio.volume = volume;
      audio.onended = () => setPlayingKey(null);
      previewAudioRef.current = audio;
      audio.play().catch(() => {});
      setPlayingKey(key);
    }
  };

  const handleSelectTrack = (trackKey: string, title?: string, duration?: number) => {
    if (previewAudioRef.current) {
      previewAudioRef.current.pause();
    }
    onSetAudio({
      key: trackKey,
      title: title || (trackKey.split("/").pop() || "Audio Track"),
      duration: duration || 0,
      trim_start: 0,
      trim_end: 0,
      volume,
      fade_in: fadeIn,
      fade_out: fadeOut,
    });
    onClose();
  };

  const handleRemoveAudio = () => {
    if (previewAudioRef.current) {
      previewAudioRef.current.pause();
    }
    onSetAudio(undefined);
    onClose();
  };

  const handleFileUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    setIsUploading(true);
    setUploadError(null);

    try {
      // Detect audio duration locally via audio element
      let detectedDuration = 0;
      const objectUrl = URL.createObjectURL(file);
      try {
        const tempAudio = new Audio(objectUrl);
        await new Promise<void>((resolve) => {
          tempAudio.onloadedmetadata = () => {
            if (Number.isFinite(tempAudio.duration)) {
              detectedDuration = tempAudio.duration;
            }
            resolve();
          };
          tempAudio.onerror = () => resolve();
          setTimeout(resolve, 2000);
        });
      } catch {
        // Fallback: duration remains 0 if browser can't parse metadata
      } finally {
        URL.revokeObjectURL(objectUrl);
      }

      const res = await videoApi.uploadAudio(file);
      handleSelectTrack(res.key, res.filename || file.name, detectedDuration);
    } catch (err) {
      setUploadError(err instanceof Error ? err.message : "Failed to upload audio file");
    } finally {
      setIsUploading(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 backdrop-blur-sm animate-fade-in">
      <div className="relative w-full max-w-xl bg-slate-900 border border-white/10 rounded-2xl shadow-2xl overflow-hidden flex flex-col max-h-[85vh]">
        {/* Modal Header */}
        <div className="flex items-center justify-between px-6 py-4 border-b border-white/10">
          <div className="flex items-center gap-2">
            <Music className="w-5 h-5 text-purple-400" />
            <h2 className="text-base font-bold text-white">Add Background Audio</h2>
          </div>
          <button
            type="button"
            onClick={() => {
              if (previewAudioRef.current) previewAudioRef.current.pause();
              onClose();
            }}
            className="p-1 rounded-lg text-slate-400 hover:text-white hover:bg-white/10 transition-colors"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Tab Navigation */}
        <div className="flex px-6 pt-3 border-b border-white/10 gap-4">
          <button
            type="button"
            onClick={() => setActiveTab("upload")}
            className={`pb-2.5 text-xs font-semibold flex items-center gap-1.5 border-b-2 transition-all ${
              activeTab === "upload"
                ? "border-purple-500 text-white"
                : "border-transparent text-slate-400 hover:text-white"
            }`}
          >
            <Upload className="w-3.5 h-3.5" />
            Upload from Computer
          </button>
          <button
            type="button"
            onClick={() => setActiveTab("presets")}
            className={`pb-2.5 text-xs font-semibold flex items-center gap-1.5 border-b-2 transition-all ${
              activeTab === "presets"
                ? "border-purple-500 text-white"
                : "border-transparent text-slate-400 hover:text-white"
            }`}
          >
            <Sparkles className="w-3.5 h-3.5" />
            Royalty-Free Presets
          </button>
        </div>

        {/* Modal Content */}
        <div className="p-6 overflow-y-auto flex-1 flex flex-col gap-4">
          {/* Current track banner */}
          {currentAudio?.key && (
            <div className="p-3 rounded-xl bg-purple-950/40 border border-purple-500/30 flex items-center justify-between">
              <div className="flex items-center gap-2 text-xs text-purple-200">
                <Music className="w-4 h-4 text-purple-400 shrink-0" />
                <div className="flex flex-col">
                  <span>
                    Active Track: <strong className="font-semibold text-white">{currentAudio.title || currentAudio.key.split("/").pop()}</strong>
                  </span>
                  {currentAudio.duration ? (
                    <span className="text-[10px] text-purple-300/80 font-mono">
                      Duration: {currentAudio.duration.toFixed(1)}s
                    </span>
                  ) : null}
                </div>
              </div>
              <button
                type="button"
                onClick={handleRemoveAudio}
                className="flex items-center gap-1 text-xs px-2.5 py-1 rounded-lg bg-rose-600/30 text-rose-300 hover:bg-rose-600 hover:text-white transition-colors"
              >
                <Trash2 className="w-3 h-3" />
                Remove
              </button>
            </div>
          )}

          {/* Volume and Fade settings */}
          <div className="grid grid-cols-3 gap-3 p-3.5 rounded-xl bg-white/[0.03] border border-white/10 text-xs">
            <div className="flex flex-col gap-1">
              <span className="text-slate-400">Volume</span>
              <input
                type="range"
                min={0}
                max={1.5}
                step={0.05}
                value={volume}
                onChange={(e) => {
                  const val = parseFloat(e.target.value);
                  setVolume(val);
                  if (previewAudioRef.current) previewAudioRef.current.volume = Math.min(1, val);
                }}
                className="accent-purple-500 w-full cursor-pointer h-1 bg-white/10 rounded"
              />
              <span className="font-mono text-slate-300 text-[10px]">{Math.round(volume * 100)}%</span>
            </div>

            <div className="flex flex-col gap-1">
              <span className="text-slate-400">Fade In</span>
              <input
                type="range"
                min={0}
                max={3}
                step={0.5}
                value={fadeIn}
                onChange={(e) => setFadeIn(parseFloat(e.target.value))}
                className="accent-purple-500 w-full cursor-pointer h-1 bg-white/10 rounded"
              />
              <span className="font-mono text-slate-300 text-[10px]">{fadeIn}s</span>
            </div>

            <div className="flex flex-col gap-1">
              <span className="text-slate-400">Fade Out</span>
              <input
                type="range"
                min={0}
                max={3}
                step={0.5}
                value={fadeOut}
                onChange={(e) => setFadeOut(parseFloat(e.target.value))}
                className="accent-purple-500 w-full cursor-pointer h-1 bg-white/10 rounded"
              />
              <span className="font-mono text-slate-300 text-[10px]">{fadeOut}s</span>
            </div>
          </div>

          {/* Tab 1: Upload from Computer */}
          {activeTab === "upload" && (
            <div className="flex flex-col items-center justify-center gap-4">
              <input
                ref={fileInputRef}
                type="file"
                accept=".mp3,.wav,.aac,.m4a,.ogg,audio/*"
                className="hidden"
                onChange={handleFileUpload}
              />

              <div
                onClick={() => !isUploading && fileInputRef.current?.click()}
                className={`w-full h-52 border-2 border-dashed rounded-2xl flex flex-col items-center justify-center p-6 cursor-pointer transition-all ${
                  isUploading
                    ? "border-purple-500/50 bg-purple-500/5 cursor-wait"
                    : "border-white/20 hover:border-purple-500/80 bg-white/[0.02] hover:bg-white/[0.05]"
                }`}
              >
                {isUploading ? (
                  <div className="flex flex-col items-center gap-3">
                    <Loader2 className="w-8 h-8 text-purple-400 animate-spin" />
                    <span className="text-sm font-medium text-slate-200">
                      Uploading audio file...
                    </span>
                    <span className="text-xs text-slate-500">Preparing for video synchronization</span>
                  </div>
                ) : (
                  <div className="flex flex-col items-center gap-2 text-center">
                    <div className="w-12 h-12 rounded-xl bg-purple-500/20 flex items-center justify-center text-purple-400 mb-1">
                      <FileAudio className="w-6 h-6" />
                    </div>
                    <span className="text-sm font-semibold text-white">
                      Click to choose an audio file from your computer
                    </span>
                    <span className="text-xs text-slate-400">
                      Supports MP3, WAV, AAC, M4A, OGG (up to 25MB)
                    </span>
                  </div>
                )}
              </div>

              {uploadError && (
                <div className="w-full p-3 rounded-xl bg-rose-500/10 border border-rose-500/20 text-xs text-rose-400">
                  {uploadError}
                </div>
              )}
            </div>
          )}

          {/* Tab 2: Royalty-Free Presets */}
          {activeTab === "presets" && (
            <div className="flex flex-col gap-2">
              <span className="text-xs font-semibold uppercase tracking-wider text-slate-400">
                Preset Music Catalog
              </span>

              <div className="space-y-2">
                {PRESET_TRACKS.map((t) => {
                  const isSelected = currentAudio?.key === t.key;
                  const isPlaying = playingKey === t.key;

                  return (
                    <div
                      key={t.key}
                      className={`p-3 rounded-xl border flex items-center justify-between gap-3 transition-all ${
                        isSelected
                          ? "border-purple-500 bg-purple-950/30"
                          : "border-white/10 bg-white/[0.02] hover:bg-white/[0.05]"
                      }`}
                    >
                      <div className="flex items-center gap-3">
                        <button
                          type="button"
                          onClick={() => togglePreview(t.url, t.key)}
                          className="w-9 h-9 rounded-xl bg-purple-600/30 hover:bg-purple-600 flex items-center justify-center text-purple-200 hover:text-white transition-all shadow-sm"
                          title={isPlaying ? "Pause preview" : "Listen preview"}
                        >
                          {isPlaying ? <Pause className="w-4 h-4" /> : <Play className="w-4 h-4 ml-0.5" />}
                        </button>

                        <div className="flex flex-col">
                          <span className="text-xs font-bold text-white">{t.title}</span>
                          <span className="text-[10px] text-slate-400">
                            {t.genre} • {t.duration}s
                          </span>
                        </div>
                      </div>

                      <button
                        type="button"
                        onClick={() => handleSelectTrack(t.key, t.title, t.duration)}
                        className={`text-xs px-3 py-1.5 rounded-lg font-semibold flex items-center gap-1 transition-all ${
                          isSelected
                            ? "bg-purple-600 text-white"
                            : "bg-white/10 hover:bg-white/20 text-slate-200 hover:text-white"
                        }`}
                      >
                        {isSelected ? <Check className="w-3.5 h-3.5" /> : null}
                        {isSelected ? "Selected" : "Use Track"}
                      </button>
                    </div>
                  );
                })}
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};
