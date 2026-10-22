import React, { useState } from "react";
import { useParams } from "react-router-dom";
import { useVideoProject } from "../hooks/useVideoProject";
import { CanvasPreview } from "../components/CanvasPreview";
import { Timeline } from "../components/Timeline";
import { ClipInspector } from "../components/ClipInspector";
import { AudioInspector } from "../components/AudioInspector";
import { MediaLibraryModal } from "../components/MediaLibraryModal";
import { AudioLibraryModal } from "../components/AudioLibraryModal";
import { ExportModal } from "../components/ExportModal";
import {
  Save,
  Download,
  Film,
  Smartphone,
  Square,
  Monitor,
  Loader2,
  Music,
  Sliders,
} from "lucide-react";

export const VideoEditorPage: React.FC = () => {
  const { id } = useParams<{ id?: string }>();
  const {
    project,
    spec,
    title,
    isDirty,
    isSaving,
    isLoading,
    error,
    activeClipId,
    activeClip,
    activeClipIndex,
    totalTime,
    currentTime,
    isPlaying,
    setTitle,
    setActiveClipId,
    setCurrentTime,
    setIsPlaying,
    addClip,
    removeClip,
    updateClip,
    reorderClips,
    updateOutput,
    setAudio,
    addTextToClip,
    updateClipText,
    removeClipText,
    saveProject,
  } = useVideoProject(id);

  // Modals & Active Inspector Tab
  const [isMediaModalOpen, setIsMediaModalOpen] = useState(false);
  const [isAudioModalOpen, setIsAudioModalOpen] = useState(false);
  const [isExportModalOpen, setIsExportModalOpen] = useState(false);
  const [selectedTab, setSelectedTab] = useState<"clip" | "audio">("clip");

  if (isLoading) {
    return (
      <div className="flex min-h-[70vh] flex-col items-center justify-center gap-3">
        <Loader2 className="w-8 h-8 text-brand-primary animate-spin" />
        <span className="text-sm font-medium text-slate-400">Loading video project...</span>
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-5 w-full pb-10 animate-fade-in">
      {/* Top Header & Toolbar */}
      <header className="flex flex-wrap items-center justify-between gap-4 p-4 rounded-2xl bg-slate-900/80 border border-white/10 backdrop-blur-md shadow-xl">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-brand-primary via-brand-secondary to-purple-600 flex items-center justify-center text-white shadow-glow-sm">
            <Film className="w-5 h-5" />
          </div>

          <div className="flex flex-col">
            <input
              type="text"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder="Untitled Video"
              className="text-base font-bold text-white bg-transparent outline-none hover:bg-white/5 focus:bg-white/10 px-2 py-0.5 rounded-lg transition-colors border border-transparent focus:border-brand-primary/50"
            />
            <div className="flex items-center gap-2 px-2 text-xs text-slate-400">
              <span>{isDirty ? "• Unsaved changes" : "• Saved"}</span>
              {error && <span className="text-rose-400">({error})</span>}
            </div>
          </div>
        </div>

        {/* Global Video Settings */}
        <div className="flex items-center gap-3">
          {/* Aspect Ratio Switcher */}
          <div className="flex items-center bg-black/40 border border-white/10 rounded-xl p-1 gap-1">
            <button
              type="button"
              onClick={() => updateOutput({ aspect: "9:16" })}
              className={`flex items-center gap-1 px-2.5 py-1 rounded-lg text-xs font-semibold transition-all ${
                spec.output.aspect === "9:16"
                  ? "bg-brand-primary text-white shadow-glow-sm"
                  : "text-slate-400 hover:text-white"
              }`}
              title="9:16 Vertical (Reels / Shorts)"
            >
              <Smartphone className="w-3.5 h-3.5" />
              9:16
            </button>
            <button
              type="button"
              onClick={() => updateOutput({ aspect: "1:1" })}
              className={`flex items-center gap-1 px-2.5 py-1 rounded-lg text-xs font-semibold transition-all ${
                spec.output.aspect === "1:1"
                  ? "bg-brand-primary text-white shadow-glow-sm"
                  : "text-slate-400 hover:text-white"
              }`}
              title="1:1 Square (Instagram Post)"
            >
              <Square className="w-3.5 h-3.5" />
              1:1
            </button>
            <button
              type="button"
              onClick={() => updateOutput({ aspect: "16:9" })}
              className={`flex items-center gap-1 px-2.5 py-1 rounded-lg text-xs font-semibold transition-all ${
                spec.output.aspect === "16:9"
                  ? "bg-brand-primary text-white shadow-glow-sm"
                  : "text-slate-400 hover:text-white"
              }`}
              title="16:9 Landscape (YouTube / Standard)"
            >
              <Monitor className="w-3.5 h-3.5" />
              16:9
            </button>
          </div>

          {/* Fit mode */}
          <select
            value={spec.output.fit}
            onChange={(e) => updateOutput({ fit: e.target.value as "cover" | "contain" })}
            className="bg-black/40 border border-white/10 text-xs text-slate-300 font-semibold rounded-xl px-3 py-1.5 outline-none cursor-pointer"
          >
            <option value="cover">Crop to Fill (Cover)</option>
            <option value="contain">Letterbox Blur (Contain)</option>
          </select>

          {/* Save Button */}
          <button
            type="button"
            onClick={saveProject}
            disabled={isSaving}
            className={`flex items-center gap-1.5 px-3.5 py-1.5 rounded-xl text-xs font-semibold border transition-all ${
              isDirty
                ? "bg-white/10 border-brand-primary text-white hover:bg-white/20"
                : "bg-white/5 border-white/10 text-slate-400 hover:text-white"
            }`}
          >
            {isSaving ? (
              <Loader2 className="w-3.5 h-3.5 animate-spin" />
            ) : (
              <Save className="w-3.5 h-3.5" />
            )}
            {isSaving ? "Saving..." : "Save"}
          </button>

          {/* Export Video Button */}
          <button
            type="button"
            onClick={() => setIsExportModalOpen(true)}
            className="flex items-center gap-1.5 px-4 py-1.5 rounded-xl text-xs font-bold bg-gradient-to-r from-brand-primary via-brand-secondary to-purple-600 text-white shadow-glow-sm hover:opacity-95 transition-opacity"
          >
            <Download className="w-3.5 h-3.5" />
            Export MP4
          </button>
        </div>
      </header>

      {/* Main Studio Workspace: Player + Inspector */}
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-5 items-start">
        {/* Left: Real-time Canvas Player */}
        <div className="lg:col-span-7 xl:col-span-8 flex flex-col items-center">
          <CanvasPreview
            spec={spec}
            currentTime={currentTime}
            isPlaying={isPlaying}
            onTimeUpdate={setCurrentTime}
            onTogglePlay={() => setIsPlaying((p) => !p)}
            onSelectClip={(id) => {
              setActiveClipId(id);
              setSelectedTab("clip");
            }}
            activeClipId={activeClipId}
          />
        </div>

        {/* Right: Inspector with Tabs for Clip and Audio */}
        <div className="lg:col-span-5 xl:col-span-4 w-full flex flex-col gap-3">
          <div className="flex items-center p-1 rounded-xl bg-slate-900/60 border border-white/10 gap-1 backdrop-blur-md">
            <button
              type="button"
              onClick={() => setSelectedTab("clip")}
              className={`flex-1 flex items-center justify-center gap-1.5 py-1.5 px-3 rounded-lg text-xs font-semibold transition-all ${
                selectedTab === "clip"
                  ? "bg-brand-primary text-white shadow-glow-sm"
                  : "text-slate-400 hover:text-white"
              }`}
            >
              <Sliders className="w-3.5 h-3.5" />
              Photo Clip Options
            </button>

            <button
              type="button"
              onClick={() => setSelectedTab("audio")}
              className={`flex-1 flex items-center justify-center gap-1.5 py-1.5 px-3 rounded-lg text-xs font-semibold transition-all ${
                selectedTab === "audio"
                  ? "bg-purple-600 text-white shadow-glow-sm"
                  : "text-slate-400 hover:text-white"
              }`}
            >
              <Music className="w-3.5 h-3.5" />
              Audio Track Options
            </button>
          </div>

          {selectedTab === "clip" ? (
            <ClipInspector
              clip={activeClip}
              clipIndex={activeClipIndex}
              onUpdateClip={updateClip}
              onAddText={addTextToClip}
              onUpdateText={updateClipText}
              onRemoveText={removeClipText}
            />
          ) : (
            <AudioInspector
              audio={spec.audio}
              currentTime={currentTime}
              totalVideoTime={totalTime}
              onUpdateAudio={(patch) => {
                if (spec.audio) {
                  setAudio({ ...spec.audio, ...patch });
                }
              }}
              onRemoveAudio={() => setAudio(undefined)}
              onChangeAudioClick={() => setIsAudioModalOpen(true)}
              onFitVideoToAudio={() => {
                if (!spec.audio?.duration || spec.clips.length === 0) return;
                const perClip = Math.max(1, Math.min(1800, spec.audio.duration / spec.clips.length));
                spec.clips.forEach((c) => {
                  updateClip(c.id, { duration: parseFloat(perClip.toFixed(1)) });
                });
              }}
            />
          )}
        </div>
      </div>

      {/* Bottom: Sequencer Timeline */}
      <div className="w-full">
        <Timeline
          spec={spec}
          activeClipId={activeClipId}
          isAudioSelected={selectedTab === "audio"}
          currentTime={currentTime}
          totalTime={totalTime}
          onSelectClip={(id) => {
            setActiveClipId(id);
            setSelectedTab("clip");
          }}
          onSelectAudio={() => setSelectedTab("audio")}
          onRemoveClip={removeClip}
          onReorderClips={reorderClips}
          onUpdateClip={updateClip}
          onUpdateAudio={(patch) => {
            if (spec.audio) {
              setAudio({ ...spec.audio, ...patch });
            }
          }}
          onAddMediaClick={() => setIsMediaModalOpen(true)}
          onAudioClick={() => setIsAudioModalOpen(true)}
          onSeek={setCurrentTime}
        />
      </div>

      {/* Modals */}
      <MediaLibraryModal
        isOpen={isMediaModalOpen}
        onClose={() => setIsMediaModalOpen(false)}
        onSelectImage={addClip}
      />

      <AudioLibraryModal
        isOpen={isAudioModalOpen}
        onClose={() => setIsAudioModalOpen(false)}
        currentAudio={spec.audio}
        onSetAudio={setAudio}
      />

      <ExportModal
        isOpen={isExportModalOpen}
        onClose={() => setIsExportModalOpen(false)}
        spec={spec}
        projectId={project?.id}
        projectTitle={title}
      />
    </div>
  );
};
