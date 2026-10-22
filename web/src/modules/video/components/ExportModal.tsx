import React, { useState, useEffect } from "react";
import type { VideoSpec } from "../engine/spec";
import { videoApi, type VideoJob } from "../api/client";
import {
  Download,
  Film,
  Loader2,
  CheckCircle2,
  AlertCircle,
  X,
} from "lucide-react";

interface ExportModalProps {
  isOpen: boolean;
  onClose: () => void;
  spec: VideoSpec;
  projectId?: string;
  projectTitle: string;
}

export const ExportModal: React.FC<ExportModalProps> = ({
  isOpen,
  onClose,
  spec,
  projectId,
  projectTitle,
}) => {
  const [resolution, setResolution] = useState<"720p" | "1080p">("1080p");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [job, setJob] = useState<VideoJob | null>(null);
  const [error, setError] = useState<string | null>(null);

  // Poll for job updates if active
  useEffect(() => {
    if (!job || job.status === "done" || job.status === "failed") return;

    const interval = setInterval(async () => {
      try {
        const updated = await videoApi.getJob(job.id);
        setJob(updated);
        if (updated.status === "done" || updated.status === "failed") {
          clearInterval(interval);
        }
      } catch (err) {
        // Continue polling on transient network hiccup
      }
    }, 1500);

    return () => clearInterval(interval);
  }, [job]);

  if (!isOpen) return null;

  const handleStartExport = async () => {
    setIsSubmitting(true);
    setError(null);

    try {
      const createdJob = await videoApi.createJob(spec, resolution, projectId);
      setJob(createdJob);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to initiate video export");
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleReset = () => {
    setJob(null);
    setError(null);
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 backdrop-blur-sm animate-fade-in">
      <div className="relative w-full max-w-lg bg-slate-900 border border-white/10 rounded-2xl shadow-2xl overflow-hidden flex flex-col">
        {/* Modal Header */}
        <div className="flex items-center justify-between px-6 py-4 border-b border-white/10">
          <div className="flex items-center gap-2">
            <Film className="w-5 h-5 text-brand-primary" />
            <h2 className="text-base font-bold text-white">Export Finished Video</h2>
          </div>
          <button
            type="button"
            onClick={onClose}
            className="p-1 rounded-lg text-slate-400 hover:text-white hover:bg-white/10 transition-colors"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Modal Body */}
        <div className="p-6 flex flex-col gap-5">
          {!job ? (
            <>
              <div className="flex flex-col gap-1">
                <span className="text-sm font-semibold text-white">{projectTitle}</span>
                <span className="text-xs text-slate-400">
                  {spec.clips.length} clips • {spec.output.aspect} aspect ratio • {spec.output.fps} FPS
                </span>
              </div>

              {/* Resolution options */}
              <div className="flex flex-col gap-2">
                <span className="text-xs font-semibold uppercase tracking-wider text-slate-400">
                  Select Export Quality
                </span>

                <div className="grid grid-cols-2 gap-3">
                  <div
                    onClick={() => setResolution("1080p")}
                    className={`p-4 rounded-xl border cursor-pointer transition-all ${
                      resolution === "1080p"
                        ? "border-brand-primary bg-brand-primary/10 shadow-glow-sm"
                        : "border-white/10 bg-white/[0.02] hover:bg-white/[0.05]"
                    }`}
                  >
                    <span className="block text-sm font-bold text-white">Full HD 1080p</span>
                    <span className="block text-xs text-slate-400 mt-0.5">
                      Recommended for Reels, TikTok & YouTube Shorts
                    </span>
                  </div>

                  <div
                    onClick={() => setResolution("720p")}
                    className={`p-4 rounded-xl border cursor-pointer transition-all ${
                      resolution === "720p"
                        ? "border-brand-primary bg-brand-primary/10 shadow-glow-sm"
                        : "border-white/10 bg-white/[0.02] hover:bg-white/[0.05]"
                    }`}
                  >
                    <span className="block text-sm font-bold text-white">HD 720p</span>
                    <span className="block text-xs text-slate-400 mt-0.5">
                      Fast render, smaller file size
                    </span>
                  </div>
                </div>
              </div>

              {error && (
                <div className="p-3 rounded-xl bg-rose-500/10 border border-rose-500/20 text-xs text-rose-400 flex items-center gap-2">
                  <AlertCircle className="w-4 h-4 shrink-0" />
                  <span>{error}</span>
                </div>
              )}

              {/* Export Button */}
              <button
                type="button"
                disabled={isSubmitting || spec.clips.length === 0}
                onClick={handleStartExport}
                className="w-full py-3 rounded-xl bg-gradient-to-r from-brand-primary via-brand-secondary to-purple-600 text-white font-bold text-sm shadow-glow-sm hover:opacity-95 transition-opacity disabled:opacity-50 flex items-center justify-center gap-2"
              >
                {isSubmitting ? (
                  <>
                    <Loader2 className="w-4 h-4 animate-spin" />
                    Initializing Worker...
                  </>
                ) : (
                  <>
                    <Download className="w-4 h-4" />
                    Render MP4 Video
                  </>
                )}
              </button>
            </>
          ) : (
            /* Active / Completed Job View */
            <div className="flex flex-col items-center justify-center py-4 gap-4 text-center">
              {job.status === "queued" && (
                <>
                  <div className="w-14 h-14 rounded-2xl bg-amber-500/20 border border-amber-500/30 flex items-center justify-center text-amber-400">
                    <Loader2 className="w-7 h-7 animate-spin" />
                  </div>
                  <div>
                    <h3 className="text-base font-bold text-white">Queued in Render Worker</h3>
                    <p className="text-xs text-slate-400 mt-1">Waiting for next available FFmpeg slot...</p>
                  </div>
                </>
              )}

              {job.status === "processing" && (
                <>
                  <div className="w-14 h-14 rounded-2xl bg-brand-primary/20 border border-brand-primary/30 flex items-center justify-center text-brand-primary">
                    <Loader2 className="w-7 h-7 animate-spin" />
                  </div>
                  <div className="w-full">
                    <h3 className="text-base font-bold text-white">Rendering MP4 with FFmpeg</h3>
                    <p className="text-xs text-slate-400 mt-1">Compositing Ken Burns, filters, text & audio...</p>
                    {/* Progress bar */}
                    <div className="w-full bg-white/10 rounded-full h-2 mt-4 overflow-hidden">
                      <div
                        className="bg-brand-primary h-full transition-all duration-300 rounded-full shadow-glow-sm"
                        style={{ width: `${Math.max(5, job.progress)}%` }}
                      />
                    </div>
                    <span className="text-xs font-mono text-brand-primary font-bold mt-1.5 block">
                      {job.progress}%
                    </span>
                  </div>
                </>
              )}

              {job.status === "done" && (
                <>
                  <div className="w-14 h-14 rounded-2xl bg-emerald-500/20 border border-emerald-500/30 flex items-center justify-center text-emerald-400 shadow-glow-sm">
                    <CheckCircle2 className="w-8 h-8" />
                  </div>
                  <div>
                    <h3 className="text-base font-bold text-white">Export Finished Successfully!</h3>
                    <p className="text-xs text-slate-400 mt-1">Your video is ready to download or share.</p>
                  </div>

                  <div className="w-full flex items-center gap-2 mt-2">
                    <a
                      href={`/files/${job.output_key || ""}`}
                      download={`${projectTitle.replace(/\s+/g, "_")}.mp4`}
                      className="flex-1 py-3 rounded-xl bg-emerald-600 hover:bg-emerald-500 text-white font-bold text-sm shadow-glow-sm flex items-center justify-center gap-2 transition-colors"
                    >
                      <Download className="w-4 h-4" />
                      Download MP4 Video
                    </a>

                    <button
                      type="button"
                      onClick={handleReset}
                      className="px-4 py-3 rounded-xl bg-white/10 hover:bg-white/20 text-slate-200 text-sm font-semibold transition-colors"
                    >
                      New Export
                    </button>
                  </div>
                </>
              )}

              {job.status === "failed" && (
                <>
                  <div className="w-14 h-14 rounded-2xl bg-rose-500/20 border border-rose-500/30 flex items-center justify-center text-rose-400">
                    <AlertCircle className="w-8 h-8" />
                  </div>
                  <div>
                    <h3 className="text-base font-bold text-white">Rendering Encountered an Error</h3>
                    <p className="text-xs text-rose-400 mt-1">{job.error || "Unknown render error"}</p>
                  </div>
                  <button
                    type="button"
                    onClick={handleReset}
                    className="px-6 py-2.5 rounded-xl bg-white/10 hover:bg-white/20 text-white text-xs font-semibold transition-colors"
                  >
                    Try Again
                  </button>
                </>
              )}
            </div>
          )}
        </div>
      </div>
    </div>
  );
};
