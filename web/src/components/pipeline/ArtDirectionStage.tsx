import { useEffect, useMemo, useState } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { Palette, Sparkles, Wand2 } from "lucide-react";
import type { ArtDirection, SceneObjectCategory, SSEStatus } from "../../api/types";
import { StageCard } from "./StageCard";
import { Button } from "../common/Button";
import { TextArea } from "../common/TextArea";
import { ErrorBanner } from "../common/ErrorBanner";
import { PromptCopyButton } from "./PromptCopyButton";
import { useToast } from "../common/Toast";
import { projectsApi } from "../../api/projects";
import { ApiError } from "../../api/client";

interface Props {
  projectId: string;
  art?: ArtDirection;
  liveStatus?: SSEStatus;
}

const CATEGORY_META: { key: SceneObjectCategory; label: string }[] = [
  { key: "hero", label: "Hero Objects" },
  { key: "cultural", label: "Cultural Objects" },
  { key: "nature", label: "Nature Objects" },
  { key: "supporting", label: "Supporting Objects" },
  { key: "decorative", label: "Decorative" },
];

export function ArtDirectionStage({ projectId, art, liveStatus }: Props) {
  const [editing, setEditing] = useState(false);
  const [draft, setDraft] = useState<ArtDirection | undefined>(art);
  const [excluded, setExcluded] = useState<Set<string>>(new Set());
  const [isSaving, setIsSaving] = useState(false);
  const [isRegenerating, setIsRegenerating] = useState(false);
  const queryClient = useQueryClient();
  const { toast } = useToast();

  useEffect(() => setDraft(art), [art]);
  // A fresh art direction (initial run, or after a scene-objects regenerate)
  // always starts with every object checked — reset local exclusions rather
  // than carrying stale ones across a new sceneObjects list.
  useEffect(() => setExcluded(new Set()), [art?.sceneObjects]);

  const byCategory = useMemo(() => {
    const map = new Map<SceneObjectCategory, string[]>();
    for (const obj of art?.sceneObjects ?? []) {
      const list = map.get(obj.category) ?? [];
      list.push(obj.name);
      map.set(obj.category, list);
    }
    return map;
  }, [art?.sceneObjects]);

  const save = useMutation({
    mutationFn: (value: ArtDirection) => projectsApi.patchArtDirection(projectId, value),
    onSuccess: (project) => {
      queryClient.setQueryData(["project", projectId], project);
      setEditing(false);
    },
    onSettled: () => setIsSaving(false),
  });

  const regenerate = useMutation({
    mutationFn: () => projectsApi.rerunArtDirection(projectId, Array.from(excluded)),
    onSuccess: (project) => {
      queryClient.setQueryData(["project", projectId], project);
      toast({ title: "Art direction regenerated", variant: "success" });
    },
    onSettled: () => setIsRegenerating(false),
  });

  const handleSave = () => {
    if (!draft || isSaving || save.isPending) return;
    setIsSaving(true);
    save.mutate(draft);
  };

  const handleRegenerate = () => {
    if (excluded.size === 0 || isRegenerating || regenerate.isPending) return;
    setIsRegenerating(true);
    regenerate.mutate();
  };

  function toggleObject(name: string) {
    setExcluded((prev) => {
      const next = new Set(prev);
      if (next.has(name)) next.delete(name);
      else next.add(name);
      return next;
    });
  }

  return (
    <StageCard
      title="Art direction"
      icon={Palette}
      liveStatus={liveStatus}
      hasData={!!art}
      actions={
        art && (
          <Button variant="secondary" onClick={() => setEditing((v) => !v)}>
            {editing ? "Cancel" : "Edit prompts"}
          </Button>
        )
      }
    >
      {save.isError && (
        <ErrorBanner message={save.error instanceof ApiError ? save.error.message : "Could not save this art direction."} />
      )}

      {art && !editing && (
        <div className="space-y-4">
          <dl className="grid gap-3 text-sm sm:grid-cols-2">
            <div className="rounded-xl border border-white/5 bg-white/[0.02] p-2.5">
              <dt className="text-[10px] font-bold uppercase tracking-wider text-slate-400">Composition Layout</dt>
              <dd className="mt-0.5 font-medium text-slate-100">{art.layout}</dd>
            </div>
            <div className="rounded-xl border border-white/5 bg-white/[0.02] p-2.5">
              <dt className="text-[10px] font-bold uppercase tracking-wider text-slate-400">Aspect Ratio</dt>
              <dd className="mt-0.5 font-medium text-slate-100">{art.aspectRatio}</dd>
            </div>
          </dl>

          {byCategory.size > 0 && (
            <div className="space-y-3 rounded-xl border border-white/10 bg-white/[0.02] p-3.5">
              <p className="text-[10px] font-bold uppercase tracking-wider text-slate-400">
                Scene Objects — Toggle Elements to Include/Exclude
              </p>
              <div className="grid gap-3 sm:grid-cols-2">
                {CATEGORY_META.filter((c) => byCategory.has(c.key)).map((c) => (
                  <div key={c.key} className="space-y-1.5">
                    <p className="text-xs font-bold text-brand-accent">{c.label}</p>
                    <ul className="space-y-1">
                      {byCategory.get(c.key)!.map((name) => (
                        <li key={name}>
                          <label className="flex items-center gap-2 text-xs text-slate-300 hover:text-white cursor-pointer">
                            <input
                              type="checkbox"
                              className="h-3.5 w-3.5 rounded accent-brand-primary"
                              checked={!excluded.has(name)}
                              onChange={() => toggleObject(name)}
                            />
                            {name}
                          </label>
                        </li>
                      ))}
                    </ul>
                  </div>
                ))}
              </div>

              {regenerate.isError && (
                <ErrorBanner
                  message={regenerate.error instanceof ApiError ? regenerate.error.message : "Could not regenerate art direction."}
                />
              )}
              <div className="flex justify-end">
                <Button
                  variant="secondary"
                  disabled={excluded.size === 0 || isRegenerating || regenerate.isPending}
                  loading={isRegenerating || regenerate.isPending}
                  loadingText="Regenerating..."
                  onClick={handleRegenerate}
                  className="text-xs py-1.5"
                >
                  <Wand2 className="h-3.5 w-3.5" aria-hidden="true" />
                  Apply changes & regenerate
                </Button>
              </div>
            </div>
          )}

          {/* Structured Poster Architecture & Spec Inspector */}
          {art.spec && (
            <div className="space-y-4 rounded-xl border border-brand-primary/30 bg-gradient-to-b from-[#141A29] to-[#0D111A] p-4 shadow-glass">
              <div className="flex items-center justify-between border-b border-white/10 pb-2.5">
                <div className="flex items-center gap-2">
                  <div className="flex h-6 w-6 items-center justify-center rounded-lg bg-brand-primary/20 text-brand-accent">
                    <Sparkles className="h-3.5 w-3.5" />
                  </div>
                  <div>
                    <h4 className="text-xs font-bold text-white tracking-wide">Poster Specification Architecture</h4>
                    <p className="text-[10px] text-slate-400">Commercial Art Direction &amp; Depth Framework</p>
                  </div>
                </div>
                <span className="rounded-full bg-brand-primary/20 px-2 py-0.5 text-[10px] font-bold text-brand-accent">
                  {art.spec.artDirectionStyle}
                </span>
              </div>

              {/* Composition Section Percentages Bar */}
              {art.spec.composition.sections && art.spec.composition.sections.length > 0 && (
                <div className="space-y-1.5">
                  <div className="flex items-center justify-between text-[10px] font-bold uppercase tracking-wider text-slate-400">
                    <span>Layout Proportions ({art.spec.composition.strategy})</span>
                    <span>100% Vertical Frame</span>
                  </div>
                  <div className="flex h-5 w-full overflow-hidden rounded-lg bg-black/40 p-0.5 border border-white/10 text-[9px] font-bold text-white">
                    {art.spec.composition.sections.map((sec, idx) => {
                      const colors = [
                        "bg-blue-600/70",
                        "bg-amber-600/70",
                        "bg-purple-600/70",
                        "bg-rose-600/70",
                        "bg-emerald-600/70",
                      ];
                      return (
                        <div
                          key={sec.section}
                          style={{ width: `${sec.percentage}%` }}
                          title={`${sec.section.toUpperCase()}: ${sec.role} (${sec.percentage}%)`}
                          className={`flex items-center justify-center truncate px-1 transition-all hover:brightness-125 ${colors[idx % colors.length]}`}
                        >
                          {sec.section} {sec.percentage}%
                        </div>
                      );
                    })}
                  </div>
                  {art.spec.composition.framing && (
                    <p className="text-[11px] text-slate-400 italic">
                      Framing: {art.spec.composition.framing}
                    </p>
                  )}
                </div>
              )}

              {/* 3-Plane Depth & Visual Hierarchy */}
              <div className="grid gap-3 sm:grid-cols-2">
                <div className="space-y-1.5 rounded-lg border border-white/5 bg-white/[0.02] p-2.5">
                  <span className="text-[10px] font-bold uppercase tracking-wider text-amber-400">
                    3-Plane Cinematic Depth
                  </span>
                  <div className="space-y-1 text-xs">
                    <p className="text-slate-300">
                      <strong className="text-white">FG:</strong> {art.spec.depth.foreground}
                    </p>
                    <p className="text-slate-300">
                      <strong className="text-white">MG:</strong> {art.spec.depth.midground}
                    </p>
                    <p className="text-slate-300">
                      <strong className="text-white">BG:</strong> {art.spec.depth.background}
                    </p>
                  </div>
                </div>

                <div className="space-y-1.5 rounded-lg border border-white/5 bg-white/[0.02] p-2.5">
                  <span className="text-[10px] font-bold uppercase tracking-wider text-purple-400">
                    Visual Hierarchy
                  </span>
                  <div className="space-y-1 text-xs">
                    <p className="text-slate-300">
                      <strong className="text-white">Primary:</strong> {art.spec.visualHierarchy.primaryElement}
                    </p>
                    <p className="text-slate-300">
                      <strong className="text-white">Secondary:</strong> {art.spec.visualHierarchy.secondaryElement}
                    </p>
                    <p className="text-slate-300">
                      <strong className="text-white">Tertiary:</strong> {art.spec.visualHierarchy.tertiaryElement}
                    </p>
                  </div>
                </div>
              </div>

              {/* Material Physics & Motivated Lighting */}
              <div className="grid gap-3 sm:grid-cols-2">
                <div className="space-y-1 text-xs rounded-lg border border-white/5 bg-white/[0.02] p-2.5">
                  <span className="text-[10px] font-bold uppercase tracking-wider text-emerald-400">
                    Material Surface Realism
                  </span>
                  <p className="text-slate-300">
                    <strong className="text-white">Metals:</strong> {art.spec.materials.metals}
                  </p>
                  {art.spec.materials.surfaces && (
                    <p className="text-slate-300">
                      <strong className="text-white">Surfaces:</strong> {art.spec.materials.surfaces}
                    </p>
                  )}
                </div>

                <div className="space-y-1 text-xs rounded-lg border border-white/5 bg-white/[0.02] p-2.5">
                  <span className="text-[10px] font-bold uppercase tracking-wider text-rose-400">
                    Motivated Illumination
                  </span>
                  <p className="text-slate-300">
                    <strong className="text-white">Primary:</strong> {art.spec.lighting.primaryLight}
                  </p>
                  <p className="text-slate-300">
                    <strong className="text-white">Rim/Ambient:</strong> {art.spec.lighting.rimLight}
                  </p>
                </div>
              </div>
            </div>
          )}

          <div>
            <div className="mb-1.5 flex items-center justify-between">
              <p className="text-[10px] font-bold uppercase tracking-wider text-slate-400">Full Poster Synthesis Prompt</p>
              <PromptCopyButton text={art.fullPosterPrompt} />
            </div>
            <p className="rounded-xl border border-white/10 bg-[#090C14] p-3 font-mono text-xs text-slate-300 leading-relaxed selection:bg-brand-primary">
              {art.fullPosterPrompt}
            </p>
          </div>

          <div>
            <div className="mb-1.5 flex items-center justify-between">
              <p className="text-[10px] font-bold uppercase tracking-wider text-slate-400">Background-Only Artwork Prompt</p>
              <PromptCopyButton text={art.backgroundPrompt} />
            </div>
            <p className="rounded-xl border border-white/10 bg-[#090C14] p-3 font-mono text-xs text-slate-300 leading-relaxed selection:bg-brand-primary">
              {art.backgroundPrompt}
            </p>
          </div>

          <div className="rounded-xl border border-white/5 bg-white/[0.02] p-3">
            <p className="text-[10px] font-bold uppercase tracking-wider text-slate-400">Reserved Brand Zones</p>
            <p className="mt-1 text-xs text-slate-400">
              Zones <span className="font-semibold text-white">{art.reservedZones.map((z) => z.id).join(", ")}</span> are explicitly kept clean in the AI prompt for pixel-exact vector logo &amp; footer overlay.
            </p>
          </div>
        </div>
      )}

      {editing && draft && (
        <div className="space-y-3">
          <div className="space-y-1">
            <label className="text-xs font-medium text-neutral-500">Full poster prompt</label>
            <TextArea
              rows={4}
              value={draft.fullPosterPrompt}
              onChange={(e) => setDraft({ ...draft, fullPosterPrompt: e.target.value })}
            />
          </div>
          <div className="space-y-1">
            <label className="text-xs font-medium text-neutral-500">Background-only prompt</label>
            <TextArea
              rows={4}
              value={draft.backgroundPrompt}
              onChange={(e) => setDraft({ ...draft, backgroundPrompt: e.target.value })}
            />
          </div>
          <div className="space-y-1">
            <label className="text-xs font-medium text-neutral-500">Negative prompt</label>
            <TextArea
              rows={2}
              value={draft.negativePrompt}
              onChange={(e) => setDraft({ ...draft, negativePrompt: e.target.value })}
            />
          </div>
          <div className="flex justify-end">
            <Button
              onClick={handleSave}
              disabled={isSaving || save.isPending}
              loading={isSaving || save.isPending}
              loadingText="Saving..."
            >
              Save
            </Button>
          </div>
        </div>
      )}
    </StageCard>
  );
}
