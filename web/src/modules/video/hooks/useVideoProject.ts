import { useState, useEffect, useCallback, useMemo } from "react";
import type { VideoSpec, ClipSpec, AudioSpec, OutputSpec, TextSpec } from "../engine/spec";
import { totalDuration } from "../engine/timing";
import { videoApi, type VideoProject } from "../api/client";

function generateId(): string {
  if (typeof crypto !== "undefined" && crypto.randomUUID) {
    return crypto.randomUUID();
  }
  return "c_" + Math.random().toString(36).substring(2, 11);
}

export function createDefaultSpec(): VideoSpec {
  return {
    version: 1,
    output: {
      aspect: "9:16",
      fit: "cover",
      fps: 30,
    },
    clips: [],
  };
}

export function createDefaultClip(imageKey: string): ClipSpec {
  return {
    id: generateId(),
    image_key: imageKey,
    duration: 3,
    crop: { x: 0, y: 0, w: 1, h: 1 },
    rotate: 0,
    filter: "none",
    adjust: {
      brightness: 0,
      contrast: 1,
      saturation: 1,
    },
    texts: [],
  };
}

export interface UseVideoProjectResult {
  project: VideoProject | null;
  spec: VideoSpec;
  title: string;
  revision: number;
  isDirty: boolean;
  isSaving: boolean;
  isLoading: boolean;
  error: string | null;
  activeClipId: string | null;
  activeClip: ClipSpec | null;
  activeClipIndex: number;
  totalTime: number;
  currentTime: number;
  isPlaying: boolean;

  // Actions
  setTitle: (title: string) => void;
  setActiveClipId: (id: string | null) => void;
  setCurrentTime: (time: number) => void;
  setIsPlaying: (playing: boolean | ((prev: boolean) => boolean)) => void;
  addClip: (imageKey: string) => void;
  removeClip: (id: string) => void;
  updateClip: (id: string, patch: Partial<ClipSpec>) => void;
  reorderClips: (fromIndex: number, toIndex: number) => void;
  updateOutput: (patch: Partial<OutputSpec>) => void;
  setAudio: (audio: AudioSpec | undefined) => void;
  addTextToClip: (clipId: string, text?: Partial<TextSpec>) => void;
  updateClipText: (clipId: string, textIndex: number, patch: Partial<TextSpec>) => void;
  removeClipText: (clipId: string, textIndex: number) => void;
  saveProject: () => Promise<VideoProject | null>;
}

export function useVideoProject(projectId?: string): UseVideoProjectResult {
  const [project, setProject] = useState<VideoProject | null>(null);
  const [spec, setSpec] = useState<VideoSpec>(createDefaultSpec);
  const [title, setTitleState] = useState<string>("Untitled Video");
  const [revision, setRevision] = useState<number>(1);
  const [isDirty, setIsDirty] = useState<boolean>(false);
  const [isSaving, setIsSaving] = useState<boolean>(false);
  const [isLoading, setIsLoading] = useState<boolean>(Boolean(projectId));
  const [error, setError] = useState<string | null>(null);

  const [activeClipId, setActiveClipId] = useState<string | null>(null);
  const [currentTime, setCurrentTime] = useState<number>(0);
  const [isPlaying, setIsPlaying] = useState<boolean>(false);

  // Load project on mount if projectId is passed
  useEffect(() => {
    if (!projectId) {
      setIsLoading(false);
      return;
    }
    let cancelled = false;
    setIsLoading(true);
    setError(null);

    videoApi
      .getProject(projectId)
      .then((data) => {
        if (cancelled) return;
        setProject(data);
        setSpec(data.spec);
        setTitleState(data.title);
        setRevision(data.revision);
        if (data.spec.clips.length > 0) {
          setActiveClipId(data.spec.clips[0].id);
        }
        setIsDirty(false);
      })
      .catch((err) => {
        if (cancelled) return;
        setError(err instanceof Error ? err.message : "Failed to load project");
      })
      .finally(() => {
        if (!cancelled) setIsLoading(false);
      });

    return () => {
      cancelled = true;
    };
  }, [projectId]);

  // Total duration in seconds
  const totalTime = useMemo(() => {
    return totalDuration(spec.clips);
  }, [spec.clips]);

  // Active clip resolution
  const { activeClip, activeClipIndex } = useMemo(() => {
    const idx = spec.clips.findIndex((c) => c.id === activeClipId);
    return {
      activeClip: idx >= 0 ? spec.clips[idx] : null,
      activeClipIndex: idx,
    };
  }, [spec.clips, activeClipId]);

  // Set title
  const setTitle = useCallback((newTitle: string) => {
    setTitleState(newTitle);
    setIsDirty(true);
  }, []);

  // Modify clips
  const addClip = useCallback((imageKey: string) => {
    const newClip = createDefaultClip(imageKey);
    setSpec((prev) => {
      const updatedClips = [...prev.clips, newClip];
      return { ...prev, clips: updatedClips };
    });
    setActiveClipId(newClip.id);
    setIsDirty(true);
  }, []);

  const removeClip = useCallback(
    (id: string) => {
      setSpec((prev) => {
        const nextClips = prev.clips.filter((c) => c.id !== id);
        return { ...prev, clips: nextClips };
      });
      if (activeClipId === id) {
        setSpec((prev) => {
          setActiveClipId(prev.clips[0]?.id || null);
          return prev;
        });
      }
      setIsDirty(true);
    },
    [activeClipId]
  );

  const updateClip = useCallback((id: string, patch: Partial<ClipSpec>) => {
    setSpec((prev) => ({
      ...prev,
      clips: prev.clips.map((clip) => (clip.id === id ? { ...clip, ...patch } : clip)),
    }));
    setIsDirty(true);
  }, []);

  const reorderClips = useCallback((fromIndex: number, toIndex: number) => {
    setSpec((prev) => {
      const nextClips = [...prev.clips];
      const [removed] = nextClips.splice(fromIndex, 1);
      nextClips.splice(toIndex, 0, removed);
      return { ...prev, clips: nextClips };
    });
    setIsDirty(true);
  }, []);

  const updateOutput = useCallback((patch: Partial<OutputSpec>) => {
    setSpec((prev) => ({
      ...prev,
      output: { ...prev.output, ...patch },
    }));
    setIsDirty(true);
  }, []);

  const setAudio = useCallback((audio: AudioSpec | undefined) => {
    setSpec((prev) => ({
      ...prev,
      audio,
    }));
    setIsDirty(true);
  }, []);

  const addTextToClip = useCallback((clipId: string, textPatch?: Partial<TextSpec>) => {
    const defaultText: TextSpec = {
      text: "Your Caption Here",
      font: "Inter",
      size: 48,
      color: "#FFFFFF",
      opacity: 1,
      x: 0.1,
      y: 0.8,
      start: 0,
      end: 0,
      anim: "fade-in",
      anim_duration: 0.5,
      ...textPatch,
    };
    setSpec((prev) => ({
      ...prev,
      clips: prev.clips.map((clip) =>
        clip.id === clipId
          ? { ...clip, texts: [...(clip.texts || []), defaultText] }
          : clip
      ),
    }));
    setIsDirty(true);
  }, []);

  const updateClipText = useCallback((clipId: string, textIndex: number, patch: Partial<TextSpec>) => {
    setSpec((prev) => ({
      ...prev,
      clips: prev.clips.map((clip) => {
        if (clip.id !== clipId) return clip;
        const texts = [...(clip.texts || [])];
        if (texts[textIndex]) {
          texts[textIndex] = { ...texts[textIndex], ...patch };
        }
        return { ...clip, texts };
      }),
    }));
    setIsDirty(true);
  }, []);

  const removeClipText = useCallback((clipId: string, textIndex: number) => {
    setSpec((prev) => ({
      ...prev,
      clips: prev.clips.map((clip) => {
        if (clip.id !== clipId) return clip;
        const texts = [...(clip.texts || [])];
        texts.splice(textIndex, 1);
        return { ...clip, texts };
      }),
    }));
    setIsDirty(true);
  }, []);

  // Save project
  const saveProject = useCallback(async (): Promise<VideoProject | null> => {
    setIsSaving(true);
    setError(null);
    try {
      if (project?.id) {
        const updated = await videoApi.updateProject(project.id, title, spec, revision);
        setProject(updated);
        setRevision(updated.revision);
        setIsDirty(false);
        return updated;
      } else {
        const created = await videoApi.createProject(title, spec);
        setProject(created);
        setRevision(created.revision);
        setIsDirty(false);
        return created;
      }
    } catch (err) {
      const msg = err instanceof Error ? err.message : "Failed to save project";
      setError(msg);
      return null;
    } finally {
      setIsSaving(false);
    }
  }, [project, title, spec, revision]);

  return {
    project,
    spec,
    title,
    revision,
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
  };
}
