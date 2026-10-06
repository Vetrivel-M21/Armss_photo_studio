import { useEffect, useState } from "react";
import { useMutation, useQuery } from "@tanstack/react-query";
import { useNavigate } from "react-router-dom";
import { motion } from "framer-motion";
import { ChevronDown, Sparkles, Wand2 } from "lucide-react";
import { TemplateGallery } from "../components/start/TemplateGallery";
import { OccasionTextArea } from "../components/start/OccasionTextArea";
import { FileAttach } from "../components/start/FileAttach";
import { LogoUpload } from "../components/start/LogoUpload";
import { RecommendationList } from "../components/start/RecommendationList";
import { Button } from "../components/common/Button";
import { Input } from "../components/common/Input";
import { ErrorBanner } from "../components/common/ErrorBanner";
import { SparkleField } from "../components/common/SparkleField";
import { cn } from "../lib/cn";
import { useMotionSafe } from "../hooks/useMotionSafe";
import { recommendationsApi } from "../api/recommendations";
import { projectsApi } from "../api/projects";
import { ApiError } from "../api/client";

const RECOMMEND_MIN_CHARS = 8;
const DEBOUNCE_MS = 400;

// The Start screen offers all three input affordances together (spec §3/§12):
// a template gallery, a large text area, and a file attach control. The user
// can combine them — pick a template *and* attach an extra reference. Any
// occasion works — a company announcement, a school celebration, or ARMSS's
// own festival posters — and the logo/footer below are what make a poster
// belong to whichever organization is making it, not always ARMSS.
export function StartPage() {
  const [text, setText] = useState("");
  const [debouncedText, setDebouncedText] = useState("");
  const [templateId, setTemplateId] = useState<string | undefined>();
  const [uploadIds, setUploadIds] = useState<string[]>([]);
  const [logoUploadId, setLogoUploadId] = useState<string | undefined>();
  const [footerText, setFooterText] = useState("");
  const [showBranding, setShowBranding] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const navigate = useNavigate();

  useEffect(() => {
    const timer = setTimeout(() => setDebouncedText(text), DEBOUNCE_MS);
    return () => clearTimeout(timer);
  }, [text]);

  const recommendEnabled = !templateId && debouncedText.trim().length >= RECOMMEND_MIN_CHARS;
  const { data: recommendations } = useQuery({
    queryKey: ["recommendations", debouncedText],
    queryFn: () => recommendationsApi.recommend(debouncedText),
    enabled: recommendEnabled,
  });

  const create = useMutation({
    mutationFn: (args?: { text: string; templateId?: string; uploadIds?: string[]; logoUploadId?: string; footerText?: string }) =>
      projectsApi.create(
        args ?? {
          text,
          templateId,
          uploadIds: uploadIds.length ? uploadIds : undefined,
          logoUploadId,
          footerText: footerText.trim() ? footerText.trim() : undefined,
        }
      ),
    onSuccess: (project) => navigate(`/projects/${project.id}`),
    onSettled: () => setIsSubmitting(false),
  });

  const isBusy = create.isPending || isSubmitting;
  const canGenerate = text.trim().length > 0 && !isBusy;
  const motionSafe = useMotionSafe();

  const [showAdvanced, setShowAdvanced] = useState(false);
  const [artStyle, setArtStyle] = useState("auto");
  const [layoutStrategy, setLayoutStrategy] = useState("auto");
  const [colorMood, setColorMood] = useState("auto");
  const [referenceQuality, setReferenceQuality] = useState(true);

  const handleCreate = () => {
    if (!canGenerate) return;
    setIsSubmitting(true);

    const promptDirectives = [
      text.trim(),
      artStyle !== "auto" ? `Preferred Art Direction: ${artStyle}` : "",
      layoutStrategy !== "auto" ? `Composition Strategy: ${layoutStrategy}` : "",
      colorMood !== "auto" ? `Color Palette: ${colorMood}` : "",
      referenceQuality ? `Mode: Premium Reference Quality (ad-grade composition, 3-plane depth, material realism, motivated lighting, non-hallucinated branding)` : "",
    ].filter(Boolean).join("\n\n");

    create.mutate({
      text: promptDirectives,
      templateId,
      uploadIds: uploadIds.length ? uploadIds : undefined,
      logoUploadId,
      footerText: footerText.trim() ? footerText.trim() : undefined,
    } as any);
  };

  return (
    <div className="space-y-10">
      {/* Cinematic Streaming Hero Banner */}
      <div className="relative overflow-hidden rounded-3xl border border-white/10 bg-gradient-to-b from-[#161C2C]/90 via-[#101422]/90 to-[#090C14]/95 p-6 shadow-2xl backdrop-blur-2xl sm:p-8">
        <motion.div
          className="pointer-events-none absolute -right-16 -top-16 h-72 w-72 rounded-full bg-brand-primary/20 blur-[100px]"
          aria-hidden="true"
          animate={motionSafe ? { x: [0, 20, -10, 0], y: [0, -15, 10, 0], scale: [1, 1.1, 0.95, 1] } : undefined}
          transition={{ duration: 14, repeat: Infinity, ease: "easeInOut" }}
        />
        <motion.div
          className="pointer-events-none absolute -bottom-20 -left-10 h-64 w-64 rounded-full bg-purple-600/20 blur-[100px]"
          aria-hidden="true"
          animate={motionSafe ? { x: [0, -18, 12, 0], y: [0, 12, -14, 0], scale: [1, 0.92, 1.08, 1] } : undefined}
          transition={{ duration: 17, repeat: Infinity, ease: "easeInOut" }}
        />
        <SparkleField />

        <div className="relative flex flex-wrap items-center justify-between gap-4">
          <div className="flex items-center gap-3.5">
            <div className="flex h-12 w-12 shrink-0 items-center justify-center rounded-2xl bg-gradient-to-br from-brand-primary via-brand-secondary to-purple-600 shadow-glow">
              <Sparkles className="h-6 w-6 text-white animate-pulse" aria-hidden="true" />
            </div>
            <div>
              <div className="flex items-center gap-2">
                <span className="rounded-full bg-brand-primary/20 border border-brand-primary/30 px-2.5 py-0.5 text-[10px] font-extrabold uppercase tracking-wider text-brand-accent">
                  AI Creative Suite
                </span>
                <span className="text-xs text-slate-400">v2.0 Production</span>
              </div>
              <h1 className="mt-1 text-2xl font-extrabold tracking-tight text-white sm:text-3xl">
                Create On-Brand Posters <span className="bg-gradient-to-r from-brand-primary via-rose-400 to-amber-300 bg-clip-text text-transparent">In Minutes</span>
              </h1>
            </div>
          </div>
          {templateId && (
            <div className="flex items-center gap-2 rounded-full border border-brand-primary/40 bg-brand-primary/10 px-3.5 py-1.5 backdrop-blur-md">
              <span className="h-2 w-2 rounded-full bg-brand-primary animate-ping" />
              <span className="text-xs font-semibold text-white">Template Selected</span>
              <button
                type="button"
                onClick={() => setTemplateId(undefined)}
                className="ml-1 text-xs text-slate-400 hover:text-white"
              >
                ✕
              </button>
            </div>
          )}
        </div>

        {/* Input Console */}
        <div className="relative mt-6 space-y-5 rounded-2xl border border-white/10 bg-[#0C101A]/90 p-5 shadow-glass backdrop-blur-xl sm:mt-7 sm:p-6">
          <OccasionTextArea value={text} onChange={setText} />

          {recommendEnabled && recommendations && recommendations.length > 0 && (
            <RecommendationList recommendations={recommendations} selectedId={templateId} onSelect={setTemplateId} />
          )}

          <div className="border-t border-white/10 pt-4">
            <Button
              type="button"
              variant="ghost"
              onClick={() => setShowBranding((v) => !v)}
              className="w-full justify-between px-0 text-slate-300 hover:bg-transparent hover:text-white font-medium"
              aria-expanded={showBranding}
            >
              <span className="flex items-center gap-2 text-xs font-bold uppercase tracking-wider text-slate-300">
                Custom Branding &amp; Reference Assets
              </span>
              <ChevronDown className={cn("h-4 w-4 transition-transform text-slate-400", showBranding && "rotate-180")} aria-hidden="true" />
            </Button>
            {showBranding && (
              <motion.div
                initial={motionSafe ? { opacity: 0, y: -10 } : undefined}
                animate={motionSafe ? { opacity: 1, y: 0 } : undefined}
                className="mt-4 space-y-4 rounded-xl border border-white/10 bg-white/[0.02] p-4"
              >
                <FileAttach onChange={setUploadIds} />
                <LogoUpload onChange={setLogoUploadId} />
                <div className="space-y-1.5">
                  <label htmlFor="footerText" className="text-xs font-medium text-slate-300">
                    Contact / footer bar text (optional)
                  </label>
                  <Input
                    id="footerText"
                    placeholder="e.g. ARMSS Employment Solutions | +91 98765 43210"
                    value={footerText}
                    onChange={(e) => setFooterText(e.target.value)}
                  />
                  <p className="text-[11px] text-slate-500">Leave blank for no footer bar on this poster.</p>
                </div>
              </motion.div>
            )}
          </div>

          {/* Advanced Design Controls (Art Direction, Depth, Palette) */}
          <div className="border-t border-white/10 pt-4">
            <Button
              type="button"
              variant="ghost"
              onClick={() => setShowAdvanced((v) => !v)}
              className="w-full justify-between px-0 text-slate-300 hover:bg-transparent hover:text-white font-medium"
              aria-expanded={showAdvanced}
            >
              <span className="flex items-center gap-2 text-xs font-bold uppercase tracking-wider text-slate-300">
                <Sparkles className="h-3.5 w-3.5 text-brand-primary" />
                Advanced Design Controls (Art Direction &amp; Depth)
              </span>
              <ChevronDown className={cn("h-4 w-4 transition-transform text-slate-400", showAdvanced && "rotate-180")} aria-hidden="true" />
            </Button>
            {showAdvanced && (
              <motion.div
                initial={motionSafe ? { opacity: 0, y: -10 } : undefined}
                animate={motionSafe ? { opacity: 1, y: 0 } : undefined}
                className="mt-4 space-y-4 rounded-xl border border-white/10 bg-white/[0.02] p-4"
              >
                <div className="grid gap-3 sm:grid-cols-2">
                  <div className="space-y-1">
                    <label className="text-[11px] font-semibold uppercase tracking-wider text-slate-400">
                      Art Direction Style
                    </label>
                    <select
                      value={artStyle}
                      onChange={(e) => setArtStyle(e.target.value)}
                      className="w-full rounded-xl border border-white/10 bg-[#0C101A] px-3 py-2 text-xs text-slate-200 outline-none focus:border-brand-primary"
                    >
                      <option value="auto">Auto / Context-Adaptive</option>
                      <option value="Premium Corporate Festival (Editorial Advertising)">Premium Corporate Festival (Editorial Advertising)</option>
                      <option value="Luxury Indian Heritage (Palace Architecture, Antique Gold)">Luxury Indian Heritage (Palace Architecture, Antique Gold)</option>
                      <option value="Cinematic Festival (Sunset Hour, Golden Illumination)">Cinematic Festival (Sunset Hour, Golden Illumination)</option>
                      <option value="Modern Corporate Grid (Clean, Contemporary)">Modern Corporate Grid (Clean, Contemporary)</option>
                      <option value="Minimal Luxury (Deep Matte, Refined Gold Accent)">Minimal Luxury (Deep Matte, Refined Gold Accent)</option>
                    </select>
                  </div>

                  <div className="space-y-1">
                    <label className="text-[11px] font-semibold uppercase tracking-wider text-slate-400">
                      Composition Layout
                    </label>
                    <select
                      value={layoutStrategy}
                      onChange={(e) => setLayoutStrategy(e.target.value)}
                      className="w-full rounded-xl border border-white/10 bg-[#0C101A] px-3 py-2 text-xs text-slate-200 outline-none focus:border-brand-primary"
                    >
                      <option value="auto">Auto Proportional Grid</option>
                      <option value="Hero-Focused (Large 3D Typography, Central Diya/Prop)">Hero-Focused (Large 3D Typography, Central Diya/Prop)</option>
                      <option value="Layered Cinematic (Foreground Family, Midground Temple, Twilight Sky)">Layered Cinematic (Foreground Family, Midground Temple, Twilight Sky)</option>
                      <option value="Editorial Arch Framing (Ornate Bells &amp; Floral Toran)">Editorial Arch Framing (Ornate Bells &amp; Floral Toran)</option>
                      <option value="Grand Symmetrical Courtyard">Grand Symmetrical Courtyard</option>
                    </select>
                  </div>

                  <div className="space-y-1">
                    <label className="text-[11px] font-semibold uppercase tracking-wider text-slate-400">
                      Color Palette &amp; Mood
                    </label>
                    <select
                      value={colorMood}
                      onChange={(e) => setColorMood(e.target.value)}
                      className="w-full rounded-xl border border-white/10 bg-[#0C101A] px-3 py-2 text-xs text-slate-200 outline-none focus:border-brand-primary"
                    >
                      <option value="auto">Auto Balanced Harmonious</option>
                      <option value="Royal Navy (45%), Sunset Crimson (25%), Antique Gold (20%), Marigold (10%)">Royal Navy, Sunset Crimson &amp; Antique Gold</option>
                      <option value="Deep Maroon &amp; Saffron Glow">Deep Maroon &amp; Saffron Glow</option>
                      <option value="Imperial Emerald &amp; Warm Brass">Imperial Emerald &amp; Warm Brass</option>
                      <option value="Midnight Twilight &amp; Amber Flame">Midnight Twilight &amp; Amber Flame</option>
                    </select>
                  </div>

                  <div className="space-y-1">
                    <label className="text-[11px] font-semibold uppercase tracking-wider text-slate-400">
                      Generation Pipeline
                    </label>
                    <label className="flex items-center gap-2.5 rounded-xl border border-white/10 bg-[#0C101A] px-3 py-2 text-xs text-slate-200 cursor-pointer">
                      <input
                        type="checkbox"
                        checked={referenceQuality}
                        onChange={(e) => setReferenceQuality(e.target.checked)}
                        className="h-4 w-4 rounded accent-brand-primary"
                      />
                      <span className="font-semibold text-brand-accent">
                        Premium Reference Quality Mode
                      </span>
                    </label>
                  </div>
                </div>
                <p className="text-[11px] text-slate-400 leading-relaxed">
                  Enforces strict 3-plane depth (Foreground, Midground, Background), authentic material surface physics (brass, silk, marble), motivated flame reflections, and high-end typography hierarchy.
                </p>
              </motion.div>
            )}
          </div>

          {create.isError && (
            <ErrorBanner
              message={create.error instanceof ApiError ? create.error.message : "Could not start this project. Try again."}
              onRetry={() => create.mutate()}
            />
          )}

          <Button
            onClick={handleCreate}
            disabled={!canGenerate}
            loading={isBusy}
            loadingText="Orchestrating AI Pipeline..."
            className="w-full justify-center gap-2.5 py-3.5 text-base font-bold shadow-glow hover:shadow-glow-violet"
          >
            <Wand2 className="h-5 w-5" aria-hidden="true" />
            Generate AI Poster
          </Button>
        </div>
      </div>

      <section className="space-y-4">
        <TemplateGallery selectedId={templateId} onSelect={setTemplateId} />
      </section>
    </div>
  );
}
