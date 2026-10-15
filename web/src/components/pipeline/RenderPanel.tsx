import { useState, useEffect, useRef } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { AlertTriangle, ChevronDown, Download, RefreshCw, Wand2, Zap, Move } from "lucide-react";
import type { ArtDirection, RenderView, ZoneOverride } from "../../api/types";
import { Button } from "../common/Button";
import { Toggle } from "../common/Toggle";
import { ErrorBanner } from "../common/ErrorBanner";
import { Card } from "../common/Card";
import { useToast } from "../common/Toast";
import { cn } from "../../lib/cn";
import { projectsApi } from "../../api/projects";
import { ApiError } from "../../api/client";
import { InteractiveLogoOverlay } from "./InteractiveLogoOverlay";

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
  const [logoNudge, setLogoNudge] = useState<{ x?: number; y?: number; w?: number; h?: number }>({});
  const [nudgeOpen, setNudgeOpen] = useState(true);
  const [posterAspectRatio, setPosterAspectRatio] = useState<string>(() => art?.aspectRatio || "3:4");
  const queryClient = useQueryClient();
  const { toast } = useToast();
  const [selectedProvider, setSelectedProvider] = useState<"flux" | "ideogram" | "openai" | undefined>(undefined);
  const canvasContainerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (art?.aspectRatio) {
      setPosterAspectRatio(art.aspectRatio);
    }
  }, [art?.aspectRatio]);

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
    mutationFn: (idempKey: string) => projectsApi.render(projectId, selectedProvider, posterAspectRatio, idempKey),
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
        logoNudge.x !== undefined || logoNudge.y !== undefined || logoNudge.w !== undefined || logoNudge.h !== undefined
          ? { logo: logoNudge }
          : undefined;
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

  const [keepCentered, setKeepCentered] = useState(true);
  const [lockAspect, setLockAspect] = useState(true);

  const currentLogoW = logoNudge.w ?? logoZone?.w ?? 0.22;
  const currentLogoH = logoNudge.h ?? logoZone?.h ?? 0.12;
  const currentLogoX = logoNudge.x ?? logoZone?.x ?? 0.39;
  const currentLogoY = logoNudge.y ?? logoZone?.y ?? 0.04;

  // The horizontal center of the logo is X + W/2.
  // When it equals 0.50, any vertical line dividing the poster into two halves also divides the logo exactly into two equal halves.
  const logoCenterPercent = (currentLogoX + currentLogoW / 2) * 100;
  const isTrulyCentered = Math.abs(logoCenterPercent - 50.0) < 0.2;

  const handleApplyPreset = (type: "top-center" | "top-left" | "top-right" | "bottom-center" | "bottom-left" | "bottom-right") => {
    const w = currentLogoW;
    const h = currentLogoH;
    let x = 0.05;
    let y = 0.04;

    if (type === "top-center") {
      x = parseFloat(((1.0 - w) / 2.0).toFixed(4));
      y = 0.04;
      setKeepCentered(true);
    } else if (type === "bottom-center") {
      x = parseFloat(((1.0 - w) / 2.0).toFixed(4));
      y = 0.84;
      setKeepCentered(true);
    } else {
      setKeepCentered(false);
      if (type === "top-left") {
        x = 0.05;
        y = 0.04;
      } else if (type === "top-right") {
        x = parseFloat((1.0 - w - 0.05).toFixed(4));
        y = 0.04;
      } else if (type === "bottom-left") {
        x = 0.05;
        y = 0.84;
      } else if (type === "bottom-right") {
        x = parseFloat((1.0 - w - 0.05).toFixed(4));
        y = 0.84;
      }
    }

    setLogoNudge({ x, y, w, h });
  };

  const handleSnapCenter = () => {
    const w = currentLogoW;
    const exactX = parseFloat(((1.0 - w) / 2.0).toFixed(4));
    setKeepCentered(true);
    setLogoNudge((n) => ({ ...n, x: exactX, w }));
  };

  // Synchronously adjusts width while maintaining 50% true vertical centerline
  const handleWidthChange = (val: number) => {
    if (isNaN(val) || val <= 0) return;
    const clampedW = parseFloat(Math.min(Math.max(val, 0.05), 0.8).toFixed(4));
    let newH = currentLogoH;
    if (lockAspect && currentLogoW > 0) {
      const ratio = currentLogoH / currentLogoW;
      newH = parseFloat(Math.min(Math.max(clampedW * ratio, 0.03), 0.5).toFixed(4));
    }

    if (keepCentered || isTrulyCentered) {
      const centeredX = parseFloat(((1.0 - clampedW) / 2.0).toFixed(4));
      setLogoNudge((n) => ({
        ...n,
        w: clampedW,
        h: newH,
        x: centeredX,
      }));
    } else {
      setLogoNudge((n) => ({
        ...n,
        w: clampedW,
        h: newH,
      }));
    }
  };

  // Synchronously adjusts height and maintains aspect ratio and center alignment
  const handleHeightChange = (val: number) => {
    if (isNaN(val) || val <= 0) return;
    const clampedH = parseFloat(Math.min(Math.max(val, 0.03), 0.5).toFixed(4));
    let newW = currentLogoW;
    if (lockAspect && currentLogoH > 0) {
      const ratio = currentLogoW / currentLogoH;
      newW = parseFloat(Math.min(Math.max(clampedH * ratio, 0.05), 0.8).toFixed(4));
    }

    if (keepCentered || isTrulyCentered) {
      const centeredX = parseFloat(((1.0 - newW) / 2.0).toFixed(4));
      setLogoNudge((n) => ({
        ...n,
        h: clampedH,
        w: newW,
        x: centeredX,
      }));
    } else {
      setLogoNudge((n) => ({
        ...n,
        h: clampedH,
        w: newW,
      }));
    }
  };

  const handleScaleSlider = (percent: number) => {
    handleWidthChange(percent / 100);
  };

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

      {shown && logoZone && (
        <div className="flex items-center justify-between rounded-lg bg-white/[0.04] px-2.5 py-1 text-[11px] text-slate-300 border border-white/5">
          <span className="flex items-center gap-1.5 font-medium">
            <Move className="h-3.5 w-3.5 text-brand-accent" />
            Drag &amp; resize the logo directly on the poster canvas
          </span>
          <button
            type="button"
            onClick={handleSnapCenter}
            className="text-[10px] text-brand-primary hover:text-brand-accent font-semibold transition-colors"
          >
            Snap Center
          </button>
        </div>
      )}

      {/* Cinematic Viewport Canvas */}
      <div
        className={cn(
          "relative overflow-hidden rounded-2xl border border-white/10 bg-gradient-to-b from-[#181D2D] to-[#0A0D15] p-1.5 shadow-2xl",
          isBusy && "ring-2 ring-brand-highlight/60 shadow-glow-amber",
        )}
      >
        <div
          ref={canvasContainerRef}
          className={cn(
            "relative flex items-center justify-center overflow-hidden rounded-xl bg-[#080A10] transition-all duration-300",
            posterAspectRatio === "1:1" && "aspect-square",
            posterAspectRatio === "4:5" && "aspect-[4/5]",
            posterAspectRatio === "3:4" && "aspect-[3/4]",
            posterAspectRatio === "2:3" && "aspect-[2/3]",
            posterAspectRatio === "9:16" && "aspect-[9/16]",
            posterAspectRatio === "16:9" && "aspect-[16/9]",
            !["1:1", "4:5", "3:4", "2:3", "9:16", "16:9"].includes(posterAspectRatio) && "aspect-[3/4]",
            !shown && "border border-dashed border-white/10",
          )}
        >
          {shown ? (
            <>
              <img
                src={shown.url}
                alt={view === 0 ? "Raw generated poster" : "Composited final poster"}
                className={`h-full w-full object-contain transition-opacity duration-300 ${shown.stale ? "opacity-40" : ""}`}
              />
              {/* Interactive Visual Drag & Drop Logo Overlay */}
              {logoZone && (
                <InteractiveLogoOverlay
                  containerRef={canvasContainerRef}
                  x={currentLogoX}
                  y={currentLogoY}
                  w={currentLogoW}
                  h={currentLogoH}
                  onChange={(coords) => {
                    setKeepCentered(Math.abs(coords.x + coords.w / 2 - 0.5) < 0.02);
                    setLogoNudge(coords);
                  }}
                  onApply={handleComposite}
                  onSnapCenter={handleSnapCenter}
                  isTrulyCentered={isTrulyCentered}
                  isCompositing={isCompositing}
                  disabled={isBusy}
                />
              )}
            </>
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

      {/* Poster Aspect Ratio / Canvas Dimensions */}
      <div className="space-y-1.5">
        <div className="flex items-center justify-between">
          <label className="text-[11px] font-bold uppercase tracking-wider text-slate-400">
            Poster Aspect Ratio / Size
          </label>
          <span className="font-mono text-[10px] font-bold text-brand-accent">
            {posterAspectRatio} Selected
          </span>
        </div>
        <div className="grid grid-cols-3 gap-1.5 rounded-xl border border-white/10 bg-black/40 p-1.5">
          {[
            { ratio: "3:4", label: "3:4", desc: "Portrait Poster" },
            { ratio: "4:5", label: "4:5", desc: "Social Feed" },
            { ratio: "1:1", label: "1:1", desc: "Square" },
            { ratio: "9:16", label: "9:16", desc: "Story / Reels" },
            { ratio: "16:9", label: "16:9", desc: "Landscape Banner" },
            { ratio: "2:3", label: "2:3", desc: "Classic 4x6" },
          ].map((item) => {
            const isSelected = posterAspectRatio === item.ratio;
            return (
              <button
                key={item.ratio}
                type="button"
                disabled={isBusy}
                onClick={() => setPosterAspectRatio(item.ratio)}
                className={cn(
                  "flex flex-col items-center justify-center rounded-lg py-1.5 px-1 text-center transition-all disabled:opacity-50",
                  isSelected
                    ? "bg-brand-primary text-white shadow-glow-sm font-bold border border-brand-primary/60"
                    : "text-slate-400 hover:text-white hover:bg-white/5 font-medium border border-transparent"
                )}
              >
                <div className="flex items-center gap-1 text-xs font-bold leading-tight">
                  {item.label}
                  {isSelected && <span className="h-1.5 w-1.5 rounded-full bg-emerald-400" />}
                </div>
                <div className="text-[9px] opacity-70 leading-tight">{item.desc}</div>
              </button>
            );
          })}
        </div>
      </div>

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
                <span>Logo placement &amp; space allocation</span>
                <ChevronDown className={cn("h-3.5 w-3.5 transition-transform", nudgeOpen && "rotate-180")} aria-hidden="true" />
              </Button>
              <div
                className={cn(
                  "grid transition-[grid-template-rows] duration-200 motion-reduce:transition-none",
                  nudgeOpen ? "grid-rows-[1fr]" : "grid-rows-[0fr]",
                )}
              >
                <div className="overflow-hidden pt-2 space-y-3">
                  {/* Position Presets with mathematical centering */}
                  <div className="space-y-1.5">
                    <div className="flex items-center justify-between">
                      <span className="text-[10px] uppercase font-bold text-slate-400">Placement Presets</span>
                      {isTrulyCentered ? (
                        <span className="flex items-center gap-1 rounded bg-emerald-500/20 px-2 py-0.5 text-[10px] font-bold text-emerald-300">
                          ✓ Bisected at 50% Center Line
                        </span>
                      ) : (
                        <button
                          type="button"
                          onClick={handleSnapCenter}
                          className="rounded bg-brand-primary/20 hover:bg-brand-primary/30 px-2 py-0.5 text-[10px] font-bold text-brand-accent transition-colors"
                        >
                          🎯 Snap to 50% True Center
                        </button>
                      )}
                    </div>
                    <div className="grid grid-cols-3 gap-1.5">
                      {[
                        { id: "top-left" as const, label: "Top Left" },
                        { id: "top-center" as const, label: "Top Center" },
                        { id: "top-right" as const, label: "Top Right" },
                        { id: "bottom-left" as const, label: "Bottom Left" },
                        { id: "bottom-center" as const, label: "Bottom Center" },
                        { id: "bottom-right" as const, label: "Bottom Right" },
                      ].map((preset) => {
                        const isActive =
                          preset.id === "top-center"
                            ? isTrulyCentered && currentLogoY < 0.2
                            : preset.id === "bottom-center"
                            ? isTrulyCentered && currentLogoY > 0.5
                            : preset.id === "top-left"
                            ? currentLogoX < 0.15 && currentLogoY < 0.2
                            : preset.id === "top-right"
                            ? currentLogoX > 0.55 && currentLogoY < 0.2
                            : preset.id === "bottom-left"
                            ? currentLogoX < 0.15 && currentLogoY > 0.5
                            : currentLogoX > 0.55 && currentLogoY > 0.5;

                        return (
                          <button
                            key={preset.id}
                            type="button"
                            disabled={isBusy}
                            onClick={() => handleApplyPreset(preset.id)}
                            className={cn(
                              "rounded-lg border px-2 py-1.5 text-[11px] font-semibold transition-all",
                              isActive
                                ? "border-brand-primary bg-brand-primary/25 text-white shadow-glow-sm"
                                : "border-white/10 bg-[#0C0F1A] text-slate-300 hover:border-brand-primary/50 hover:text-white hover:bg-white/5"
                            )}
                          >
                            {preset.label}
                          </button>
                        );
                      })}
                    </div>
                  </div>

                  {/* Scale & Aspect Ratio Controls */}
                  <div className="space-y-1.5 rounded-xl border border-white/10 bg-[#0C0F1A]/80 p-2.5">
                    <div className="flex items-center justify-between text-xs">
                      <span className="font-semibold text-slate-300">Logo Scale / Size</span>
                      <span className="font-mono text-brand-accent font-bold">{Math.round(currentLogoW * 100)}% Width</span>
                    </div>
                    <input
                      type="range"
                      min={8}
                      max={60}
                      step={1}
                      disabled={isBusy}
                      value={Math.round(currentLogoW * 100)}
                      onChange={(e) => handleScaleSlider(parseFloat(e.target.value))}
                      className="w-full accent-brand-primary cursor-pointer"
                    />
                    <div className="flex items-center justify-between pt-1">
                      <label className="flex items-center gap-1.5 text-[10px] text-slate-300 cursor-pointer select-none">
                        <input
                          type="checkbox"
                          checked={keepCentered}
                          onChange={(e) => {
                            setKeepCentered(e.target.checked);
                            if (e.target.checked) handleSnapCenter();
                          }}
                          className="rounded border-white/20 text-brand-primary focus:ring-0"
                        />
                        <span>Auto-Center 50% Axis</span>
                      </label>
                      <label className="flex items-center gap-1.5 text-[10px] text-slate-300 cursor-pointer select-none">
                        <input
                          type="checkbox"
                          checked={lockAspect}
                          onChange={(e) => setLockAspect(e.target.checked)}
                          className="rounded border-white/20 text-brand-primary focus:ring-0"
                        />
                        <span>Lock Aspect Ratio</span>
                      </label>
                    </div>
                  </div>

                  {/* Manual Coordinate & Size Inputs */}
                  <div className="grid grid-cols-2 gap-2">
                    <label className="text-[11px] text-slate-400">
                      <div className="flex items-center justify-between">
                        <span>Offset X (0-1)</span>
                        <span className="text-[9px] text-slate-500 font-mono">Axis: {logoCenterPercent.toFixed(1)}%</span>
                      </div>
                      <input
                        type="number"
                        step={0.01}
                        min={0}
                        max={1}
                        disabled={isBusy}
                        value={logoNudge.x ?? logoZone.x}
                        onChange={(e) => {
                          setKeepCentered(false);
                          setLogoNudge((n) => ({ ...n, x: parseFloat(e.target.value) }));
                        }}
                        className="mt-1 w-full rounded-lg border border-white/10 bg-[#0C0F1A] px-2 py-1 text-xs text-white disabled:opacity-50"
                      />
                    </label>
                    <label className="text-[11px] text-slate-400">
                      Offset Y (0-1)
                      <input
                        type="number"
                        step={0.01}
                        min={0}
                        max={1}
                        disabled={isBusy}
                        value={logoNudge.y ?? logoZone.y}
                        onChange={(e) => setLogoNudge((n) => ({ ...n, y: parseFloat(e.target.value) }))}
                        className="mt-1 w-full rounded-lg border border-white/10 bg-[#0C0F1A] px-2 py-1 text-xs text-white disabled:opacity-50"
                      />
                    </label>
                    <label className="text-[11px] text-slate-400">
                      <div className="flex items-center justify-between">
                        <span>Width W (0-1)</span>
                        <span className="text-[9px] text-slate-500 font-mono">{Math.round(currentLogoW * 100)}%</span>
                      </div>
                      <input
                        type="number"
                        step={0.01}
                        min={0.05}
                        max={0.8}
                        disabled={isBusy}
                        value={logoNudge.w ?? logoZone.w}
                        onChange={(e) => handleWidthChange(parseFloat(e.target.value))}
                        className="mt-1 w-full rounded-lg border border-white/10 bg-[#0C0F1A] px-2 py-1 text-xs text-white disabled:opacity-50"
                      />
                    </label>
                    <label className="text-[11px] text-slate-400">
                      <div className="flex items-center justify-between">
                        <span>Height H (0-1)</span>
                        <span className="text-[9px] text-slate-500 font-mono">{Math.round(currentLogoH * 100)}%</span>
                      </div>
                      <input
                        type="number"
                        step={0.01}
                        min={0.03}
                        max={0.5}
                        disabled={isBusy}
                        value={logoNudge.h ?? logoZone.h}
                        onChange={(e) => handleHeightChange(parseFloat(e.target.value))}
                        className="mt-1 w-full rounded-lg border border-white/10 bg-[#0C0F1A] px-2 py-1 text-xs text-white disabled:opacity-50"
                      />
                    </label>
                  </div>

                  {/* One-Click Quick Re-composite when nudged */}
                  {Object.keys(logoNudge).length > 0 && (
                    <Button
                      type="button"
                      variant="primary"
                      className="w-full justify-center gap-1.5 py-2 text-xs shadow-glow"
                      disabled={isBusy}
                      loading={isCompositing}
                      onClick={handleComposite}
                    >
                      <Wand2 className="h-3.5 w-3.5" />
                      Apply &amp; Re-composite Logo Now
                    </Button>
                  )}
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
