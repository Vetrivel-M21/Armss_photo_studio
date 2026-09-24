import { useState } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { AlertTriangle, ChevronDown, Download, RefreshCw, Wand2, Zap } from "lucide-react";
import type { ArtDirection, RenderView, ZoneOverride } from "../../api/types";
import { Button } from "../common/Button";
import { Toggle } from "../common/Toggle";
import { ErrorBanner } from "../common/ErrorBanner";
import { Card } from "../common/Card";
import { useToast } from "../common/Toast";
import { cn } from "../../lib/cn";
import { projectsApi } from "../../api/projects";
import { ApiError } from "../../api/client";

interface Props {
  projectId: string;
  art?: ArtDirection;
  renders: RenderView[];
}

// The sticky right panel (spec §12): the render, then the composited final,
// with a raw/composited toggle. Image generation and compositing are the
// only steps that cost money or overwrite brand assets, so both are strictly
// protected with synchronous one-click locks, stable idempotency keys, and
// immediate button disabling to prevent duplicate API billing.
export function RenderPanel({ projectId, art, renders }: Props) {
  const [view, setView] = useState<0 | 1>(1); // 0 = raw, 1 = composited
  const [logoNudge, setLogoNudge] = useState<{ x?: number; y?: number }>({});
  const [nudgeOpen, setNudgeOpen] = useState(false);
  const queryClient = useQueryClient();
  const { toast } = useToast();
  const [selectedProvider, setSelectedProvider] = useState<"flux" | "ideogram" | "openai" | undefined>(undefined);

  // Scoped idempotency keys and synchronous hardware locks to eliminate duplicate requests
  const [activeRenderKey, setActiveRenderKey] = useState<string>(() => crypto.randomUUID());
  const [activeCompositeKey, setActiveCompositeKey] = useState<string>(() => crypto.randomUUID());
  const [isRenderLocked, setIsRenderLocked] = useState(false);
  const [isCompositeLocked, setIsCompositeLocked] = useState(false);

  const rawRenders = (renders ?? []).filter((r) => r.kind === "raw");
  const compositedRenders = (renders ?? []).filter((r) => r.kind === "composited");
  const latestRaw = rawRenders[0];
  const latestComposited = compositedRenders[0];
  const shown = view === 0 ? latestRaw : latestComposited;

  const render = useMutation({
    mutationFn: (idempKey: string) => projectsApi.render(projectId, selectedProvider, idempKey),
    onSuccess: (project) => {
      queryClient.setQueryData(["project", projectId], project);
      setView(0);
      toast({ title: "Artwork synthesized successfully", variant: "success" });
    },
    onSettled: () => {
      setIsRenderLocked(false);
      setActiveRenderKey(crypto.randomUUID());
    },
  });

  const composite = useMutation({
    mutationFn: (idempKey: string) => {
      const overrides: Record<string, ZoneOverride> | undefined =
        logoNudge.x !== undefined || logoNudge.y !== undefined ? { logo: logoNudge } : undefined;
      return projectsApi.composite(projectId, overrides, idempKey);
    },
    onSuccess: (project) => {
      queryClient.setQueryData(["project", projectId], project);
      setView(1);
      toast({ title: "Poster composited with brand vector assets", variant: "success" });
    },
    onSettled: () => {
      setIsCompositeLocked(false);
      setActiveCompositeKey(crypto.randomUUID());
    },
  });

  const isRendering = render.isPending || isRenderLocked;
  const isCompositing = composite.isPending || isCompositeLocked;
  const isBusy = isRendering || isCompositing;

  const handleGenerate = () => {
    if (!art || isBusy) return;
    setIsRenderLocked(true);
    render.mutate(activeRenderKey);
  };

  const handleComposite = () => {
    if (isBusy) return;
    setIsCompositeLocked(true);
    composite.mutate(activeCompositeKey);
  };

  const logoZone = art?.reservedZones.find((z) => z.id === "logo");
  const effectiveProvider = selectedProvider ?? "flux";

  return (
    <Card padding="md" className="lg:sticky lg:top-6 space-y-5 border-white/15 bg-[#111422]/90 backdrop-blur-2xl shadow-glass">
      <div className="flex items-center justify-between border-b border-white/10 pb-3">
        <div>
          <h3 className="text-sm font-bold tracking-tight text-white">Poster Viewport</h3>
          <p className="text-[11px] text-slate-400">Live Artwork &amp; Compositing</p>
        </div>
        {(latestRaw || latestComposited) && (
          <Toggle options={["Raw", "Composited"]} value={view} onChange={setView} />
        )}
      </div>

      {/* Cinematic Viewport Canvas */}
      <div
        className={cn(
          "relative overflow-hidden rounded-2xl border border-white/10 bg-gradient-to-b from-[#181D2D] to-[#0A0D15] p-1.5 shadow-2xl",
          isBusy && "ring-2 ring-brand-highlight/60 shadow-glow-amber",
        )}
      >
        <div
          className={cn(
            "relative flex aspect-[4/5] items-center justify-center overflow-hidden rounded-xl bg-[#080A10]",
            !shown && "border border-dashed border-white/10",
          )}
        >
          {shown ? (
            <img
              src={shown.url}
              alt={view === 0 ? "Raw generated poster" : "Composited final poster"}
              className={`h-full w-full object-contain transition-opacity duration-300 ${shown.stale ? "opacity-40" : ""}`}
            />
          ) : (
            <div className="p-6 text-center">
              <div className="mx-auto mb-2 flex h-10 w-10 items-center justify-center rounded-xl bg-white/[0.04] text-slate-500">
                <Wand2 className="h-5 w-5" />
              </div>
              <p className="text-xs font-medium text-slate-400">
                {art ? "Ready for generation" : "Select creative concept to unlock"}
              </p>
              <p className="mt-1 text-[11px] text-slate-500">
                AI generates base art; Go compositor applies brand assets
              </p>
            </div>
          )}

          {/* Live Scanner beam while generating */}
          {isRendering && (
            <div className="pointer-events-none absolute inset-0 overflow-hidden">
              <div className="h-full w-full animate-shimmer bg-gradient-to-b from-transparent via-brand-primary/25 to-transparent" />
            </div>
          )}
        </div>
        {isBusy && (
          <div className="pointer-events-none absolute inset-1 overflow-hidden rounded-xl" aria-hidden="true">
            <div className="absolute inset-y-0 left-0 w-1/2 animate-shimmer bg-gradient-to-r from-transparent via-brand-highlight/25 to-transparent" />
          </div>
        )}
      </div>

      {shown?.stale && (
        <p className="flex items-center gap-1.5 rounded-xl bg-rose-500/10 border border-rose-500/20 p-2 text-xs text-rose-300">
          <AlertTriangle className="h-4 w-4 shrink-0 text-rose-400" aria-hidden="true" />
          Art direction updated — regenerate to update artwork.
        </p>
      )}

      {render.isError && (
        <ErrorBanner
          message={render.error instanceof ApiError ? render.error.message : "Image generation failed."}
          onRetry={handleGenerate}
        />
      )}

      {/* Neural Image Engine Selector */}
      <div className="space-y-1.5">
        <div className="flex items-center justify-between">
          <label className="text-[11px] font-bold uppercase tracking-wider text-slate-400">
            Render Engine
          </label>
          <span className="flex items-center gap-1 text-[10px] font-semibold text-emerald-400">
            {effectiveProvider === "flux" ? (
              <>
                <Zap className="h-3 w-3 text-amber-400" />
                ~$0.003/img · 90% savings
              </>
            ) : (
              "~$0.08/img"
            )}
          </span>
        </div>
        <div className="grid grid-cols-3 gap-1.5 rounded-xl border border-white/10 bg-black/40 p-1">
          {[
            { id: "flux", label: "Flux.1", sub: "Jewelry / Glow", highlight: true },
            { id: "ideogram", label: "Ideogram", sub: "3D Typography", highlight: false },
            { id: "openai", label: "DALL-E 3", sub: "General", highlight: false },
          ].map((item) => {
            const isChosen = effectiveProvider === item.id;
            return (
              <button
                key={item.id}
                type="button"
                disabled={isBusy}
                onClick={() => setSelectedProvider(item.id as "flux" | "ideogram" | "openai")}
                className={cn(
                  "rounded-lg py-1.5 px-1 text-center transition-all disabled:opacity-50 disabled:cursor-not-allowed",
                  isChosen
                    ? "bg-brand-primary text-white shadow-glow-sm font-bold"
                    : "text-slate-400 hover:text-white hover:bg-white/5 font-medium"
                )}
              >
                <div className="flex items-center justify-center gap-1 text-xs leading-tight">
                  {item.label}
                  {item.highlight && <span className="h-1.5 w-1.5 rounded-full bg-emerald-400" />}
                </div>
                <div className="text-[9px] opacity-70 leading-tight">{item.sub}</div>
              </button>
            );
          })}
        </div>
      </div>

      <Button
        className="w-full justify-center gap-2 py-3 shadow-glow"
        disabled={!art || isBusy}
        loading={isRendering}
        loadingText="Synthesizing Artwork..."
        onClick={handleGenerate}
      >
        <RefreshCw className="h-4 w-4" aria-hidden="true" />
        {latestRaw ? "Regenerate Artwork" : "Generate Artwork"}
      </Button>

      {latestRaw && (
        <div className="space-y-3 border-t border-white/10 pt-4">
          {logoZone && (
            <div className="text-xs text-slate-400">
              <Button
                type="button"
                variant="ghost"
                disabled={isBusy}
                onClick={() => setNudgeOpen((v) => !v)}
                aria-expanded={nudgeOpen}
                className="w-full justify-between px-0 py-1 text-xs font-semibold text-slate-300 hover:bg-transparent hover:text-white"
              >
                <span>Fine-tune logo coordinates</span>
                <ChevronDown className={cn("h-3.5 w-3.5 transition-transform", nudgeOpen && "rotate-180")} aria-hidden="true" />
              </Button>
              <div
                className={cn(
                  "grid transition-[grid-template-rows] duration-200 motion-reduce:transition-none",
                  nudgeOpen ? "grid-rows-[1fr]" : "grid-rows-[0fr]",
                )}
              >
                <div className="overflow-hidden pt-2">
                  <div className="flex gap-2">
                    <label className="flex-1 text-[11px] text-slate-400">
                      Offset X (0-1)
                      <input
                        type="number"
                        step={0.01}
                        min={0}
                        max={1}
                        disabled={isBusy}
                        defaultValue={logoZone.x}
                        onChange={(e) => setLogoNudge((n) => ({ ...n, x: parseFloat(e.target.value) }))}
                        className="mt-1 w-full rounded-lg border border-white/10 bg-[#0C0F1A] px-2.5 py-1.5 text-xs text-white disabled:opacity-50"
                      />
                    </label>
                    <label className="flex-1 text-[11px] text-slate-400">
                      Offset Y (0-1)
                      <input
                        type="number"
                        step={0.01}
                        min={0}
                        max={1}
                        disabled={isBusy}
                        defaultValue={logoZone.y}
                        onChange={(e) => setLogoNudge((n) => ({ ...n, y: parseFloat(e.target.value) }))}
                        className="mt-1 w-full rounded-lg border border-white/10 bg-[#0C0F1A] px-2.5 py-1.5 text-xs text-white disabled:opacity-50"
                      />
                    </label>
                  </div>
                </div>
              </div>
            </div>
          )}

          {composite.isError && (
            <ErrorBanner
              message={composite.error instanceof ApiError ? composite.error.message : "Compositing failed."}
              onRetry={handleComposite}
            />
          )}

          <Button
            variant="secondary"
            className="w-full justify-center gap-2"
            disabled={isBusy}
            loading={isCompositing}
            loadingText="Compositing brand assets..."
            onClick={handleComposite}
          >
            <Wand2 className="h-4 w-4 text-brand-primary" aria-hidden="true" />
            {latestComposited ? "Re-composite Vector Assets" : "Composite Vector Brand Assets"}
          </Button>
        </div>
      )}

      {latestComposited && (
        <a
          href={projectsApi.downloadUrl(projectId, "composited")}
          className="flex w-full items-center justify-center gap-2 rounded-xl bg-gradient-to-r from-emerald-500 to-teal-600 px-4 py-3 text-center text-sm font-bold text-white shadow-[0_0_20px_-3px_rgba(16,185,129,0.4)] transition-all hover:brightness-110 active:scale-95"
        >
          <Download className="h-4 w-4" aria-hidden="true" />
          Download Production Poster (PNG)
        </a>
      )}
    </Card>
  );
}
