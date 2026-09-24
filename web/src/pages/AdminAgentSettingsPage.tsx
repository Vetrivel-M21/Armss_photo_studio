import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Settings, Image as ImageIcon, Cpu, Type } from "lucide-react";
import { agentSettingsApi, type KeyProvider } from "../api/agentSettings";
import { Button } from "../components/common/Button";
import { Input } from "../components/common/Input";
import { Spinner } from "../components/common/Spinner";
import { ErrorBanner } from "../components/common/ErrorBanner";
import { Card, CardHeader } from "../components/common/Card";
import { Badge, type BadgeTone } from "../components/common/Badge";
import { useToast } from "../components/common/Toast";
import { ApiError } from "../api/client";

const DEFAULT_MODELS: Record<"anthropic" | "gemini", string> = {
  anthropic: "claude-sonnet-5",
  gemini: "gemini-3.6-flash",
};

function PageHeaderBadge() {
  return (
    <div className="flex h-12 w-12 shrink-0 items-center justify-center rounded-2xl bg-gradient-to-br from-brand-primary via-brand-secondary to-brand-accent shadow-glow-sm">
      <Settings className="h-6 w-6 text-white" aria-hidden="true" />
    </div>
  );
}

const KEY_LABELS: Record<KeyProvider, { label: string; hint: string; tag: string }> = {
  anthropic: {
    label: "Anthropic (Claude)",
    hint: "Drives the 7-agent pipeline (Brief, Copy, Concepts, Art Direction) with structured JSON.",
    tag: "Pipeline LLM",
  },
  gemini: {
    label: "Google Gemini",
    hint: "High-speed multimodal agent pipeline provider with expansive context windows.",
    tag: "Pipeline LLM",
  },
  flux: {
    label: "Flux.1 (Together.ai / Fal.ai)",
    hint: "Black Forest Labs FLUX.1 (Schnell/Dev) — High photorealism for Indian jewelry & festive lighting (~$0.003-$0.03/img).",
    tag: "Image Generator",
  },
  ideogram: {
    label: "Ideogram v2",
    hint: "State-of-the-art English typography rendering for festival headlines & graphic posters (~$0.08/img).",
    tag: "Image Generator",
  },
  openai: {
    label: "OpenAI (DALL-E 3)",
    hint: "General-purpose image generation backend via OpenAI image generation endpoints.",
    tag: "Image Generator",
  },
};

const IMAGE_PROVIDERS: Array<{
  id: "flux" | "ideogram" | "openai";
  name: string;
  provider: string;
  cost: string;
  badgeTone: BadgeTone;
  bestFor: string;
  desc: string;
}> = [
  {
    id: "flux",
    name: "FLUX.1 (Schnell / Dev)",
    provider: "Together.ai / Fal.ai",
    cost: "~$0.003 - $0.03 / poster",
    badgeTone: "success",
    bestFor: "Indian Festivals, Gold Jewelry, Realistic Lighting & Cost Efficiency",
    desc: "12B parameter rectified flow transformer. Top photorealism and Indian cultural aesthetics at 1/10th the cost of proprietary models.",
  },
  {
    id: "ideogram",
    name: "Ideogram v2",
    provider: "Ideogram API",
    cost: "~$0.08 / poster",
    badgeTone: "info",
    bestFor: "Big 3D English Titles (DIWALI, PONGAL, SALE) & Graphic Design",
    desc: "Unmatched rendering of English typography, decorative fonts, and clean commercial vector-style compositions.",
  },
  {
    id: "openai",
    name: "OpenAI (DALL-E 3)",
    provider: "OpenAI API",
    cost: "~$0.08 / poster",
    badgeTone: "neutral",
    bestFor: "Conceptual Illustrations & Stylized Artwork",
    desc: "Standard diffusion model with high prompt adherence for surrealist and illustrated poster backgrounds.",
  },
];

function ApiKeyRow({ provider, configured }: { provider: KeyProvider; configured: boolean }) {
  const [value, setValue] = useState("");
  const queryClient = useQueryClient();
  const { toast } = useToast();
  const { label, hint, tag } = KEY_LABELS[provider];

  const save = useMutation({
    mutationFn: () => agentSettingsApi.setKey(provider, value),
    onSuccess: (updated) => {
      queryClient.setQueryData(["agent-settings"], updated);
      setValue("");
      toast({ title: `${label} key securely saved`, variant: "success" });
    },
  });

  return (
    <div className="space-y-2 rounded-2xl border border-white/5 bg-[#111420]/60 p-4 backdrop-blur-sm transition-all hover:border-white/10">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <div className="flex items-center gap-2">
          <label className="text-sm font-semibold text-white">{label}</label>
          <span className="rounded-full bg-white/5 px-2 py-0.5 text-[10px] font-medium text-slate-400">
            {tag}
          </span>
        </div>
        <Badge tone={configured ? "success" : "danger"}>
          {configured ? "Active / Configured" : "Missing Key"}
        </Badge>
      </div>
      <p className="text-xs text-slate-400">{hint}</p>
      <div className="flex gap-2 pt-1">
        <Input
          type="password"
          autoComplete="off"
          value={value}
          onChange={(e) => setValue(e.target.value)}
          placeholder={configured ? "Paste a new key to rotate" : "Paste API key (sk-... or key-...)"}
          className="flex-1 font-mono text-xs"
        />
        <Button
          type="button"
          variant="secondary"
          disabled={!value.trim() || save.isPending}
          loading={save.isPending}
          loadingText="Saving..."
          onClick={() => save.mutate()}
        >
          Save Key
        </Button>
      </div>
      {save.isError && (
        <ErrorBanner message={save.error instanceof ApiError ? save.error.message : "Could not save this key."} />
      )}
    </div>
  );
}

export function AdminAgentSettingsPage() {
  const [provider, setProvider] = useState<"anthropic" | "gemini" | null>(null);
  const [model, setModel] = useState<string | null>(null);
  const queryClient = useQueryClient();
  const { toast } = useToast();

  const { data: settings, isLoading } = useQuery({
    queryKey: ["agent-settings"],
    queryFn: agentSettingsApi.get,
  });

  const activeProvider = provider ?? settings?.provider ?? "anthropic";
  const activeModel = model ?? settings?.model ?? "";
  const activeImageProvider = settings?.imageProvider ?? "flux";

  const saveLLM = useMutation({
    mutationFn: () => agentSettingsApi.put(activeProvider, activeModel),
    onSuccess: (updated) => {
      queryClient.setQueryData(["agent-settings"], updated);
      setProvider(null);
      setModel(null);
      toast({ title: `Switched LLM to ${updated.provider}`, variant: "success" });
    },
  });

  const switchImageProvider = useMutation({
    mutationFn: (newImgProv: "flux" | "ideogram" | "openai") =>
      agentSettingsApi.setImageProvider(newImgProv),
    onSuccess: (updated) => {
      queryClient.setQueryData(["agent-settings"], updated);
      toast({ title: `Default image engine switched to ${updated.imageProvider.toUpperCase()}`, variant: "success" });
    },
  });

  function selectProvider(p: "anthropic" | "gemini") {
    setProvider(p);
    setModel(DEFAULT_MODELS[p]);
  }

  if (isLoading || !settings) {
    return (
      <div className="space-y-6">
        <div className="flex items-center gap-3">
          <PageHeaderBadge />
          <h1 className="font-heading text-xl font-bold text-white">AI Engine Orchestration</h1>
        </div>
        <Spinner label="Loading current AI configurations..." />
      </div>
    );
  }

  return (
    <div className="space-y-8 pb-12">
      {/* Header */}
      <div className="flex items-center gap-4">
        <PageHeaderBadge />
        <div>
          <h1 className="font-heading text-2xl font-extrabold text-white tracking-tight">AI & Image Engine Settings</h1>
          <p className="mt-1 text-sm text-slate-400">
            Control the 7-agent LLM reasoning pipeline, default diffusion renderer (Flux.1 / Ideogram / OpenAI), and rotate keys on the fly.
          </p>
        </div>
      </div>

      {/* Image Generation Engine Selector */}
      <Card className="space-y-5 border-brand-primary/20 bg-gradient-to-b from-[#141826] to-[#0d101a]">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2.5">
            <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-brand-primary/20 text-brand-primary">
              <ImageIcon className="h-4 w-4" />
            </div>
            <div>
              <h2 className="text-base font-bold text-white">Default Image Generation Engine</h2>
              <p className="text-xs text-slate-400">Select which neural renderer generates poster artwork in Stage 6 (Render)</p>
            </div>
          </div>
          <Badge tone="brand">
            Active: {activeImageProvider.toUpperCase()}
          </Badge>
        </div>

        <div className="grid gap-4 sm:grid-cols-3">
          {IMAGE_PROVIDERS.map((ip) => {
            const isSelected = activeImageProvider === ip.id;
            const isConfigured =
              ip.id === "flux"
                ? settings.fluxConfigured
                : ip.id === "ideogram"
                ? settings.ideogramConfigured
                : settings.openaiConfigured;

            return (
              <div
                key={ip.id}
                onClick={() => {
                  if (isSelected || switchImageProvider.isPending) return;
                  switchImageProvider.mutate(ip.id);
                }}
                className={`relative flex flex-col justify-between rounded-2xl border p-4 transition-all duration-300 ${
                  isSelected
                    ? "border-brand-primary bg-brand-primary/10 shadow-glow-sm ring-1 ring-brand-primary/40 cursor-default"
                    : switchImageProvider.isPending
                    ? "border-white/10 bg-white/[0.01] opacity-60 cursor-not-allowed"
                    : "border-white/10 bg-white/[0.02] hover:border-white/20 hover:bg-white/[0.04] cursor-pointer"
                }`}
              >
                <div>
                  <div className="flex items-center justify-between gap-2">
                    <span className="text-sm font-extrabold text-white">{ip.name}</span>
                    <Badge tone={isConfigured ? "success" : "danger"}>
                      {isConfigured ? "Key Set" : "No Key"}
                    </Badge>
                  </div>
                  <div className="mt-1 flex items-center gap-2">
                    <span className="text-[11px] font-mono font-medium text-brand-accent">{ip.cost}</span>
                  </div>
                  <p className="mt-2.5 text-xs text-slate-300 leading-relaxed">{ip.desc}</p>
                  <div className="mt-3 rounded-lg bg-black/30 p-2 text-[11px] text-slate-400 border border-white/5">
                    <span className="font-semibold text-slate-200">Recommended for:</span> {ip.bestFor}
                  </div>
                </div>

                <div className="mt-4 pt-3 border-t border-white/10 flex items-center justify-between">
                  <span className="text-[11px] text-slate-400">{ip.provider}</span>
                  <button
                    type="button"
                    disabled={isSelected || switchImageProvider.isPending}
                    className={`rounded-lg px-3 py-1 text-xs font-semibold transition-all ${
                      isSelected
                        ? "bg-brand-primary text-white cursor-default"
                        : "bg-white/10 text-white hover:bg-white/20 disabled:opacity-40"
                    }`}
                  >
                    {isSelected ? "Selected" : switchImageProvider.isPending ? "Switching..." : "Set as Default"}
                  </button>
                </div>
              </div>
            );
          })}
        </div>
      </Card>

      {/* Tamil Font & Compositor Architecture Info */}
      <Card className="border-amber-500/20 bg-amber-500/[0.03]">
        <div className="flex items-start gap-3">
          <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-xl bg-amber-500/20 text-amber-400">
            <Type className="h-5 w-5" />
          </div>
          <div className="space-y-1">
            <h3 className="text-sm font-bold text-white flex items-center gap-2">
              Bilingual Tamil & Deterministic Brand Compositor
              <Badge tone="warning">NotoSansTamil-Bold.ttf Wired</Badge>
            </h3>
            <p className="text-xs text-slate-400 leading-relaxed">
              Diffusion models (Flux, Ideogram, OpenAI) generate the visual artwork, gold jewelry, and backgrounds.
              Tamil typography, exact ARMSS logos, and footer contact details are rendered deterministically onto reserved pixel zones by the Go backend compositor using the bundled TrueType Tamil font — guaranteeing zero spelling deformities or hallucinatory phone numbers.
            </p>
          </div>
        </div>
      </Card>

      {/* LLM Agent Pipeline Settings */}
      <Card className="space-y-5">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2.5">
            <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-brand-secondary/20 text-brand-secondary">
              <Cpu className="h-4 w-4" />
            </div>
            <div>
              <h2 className="text-base font-bold text-white">LLM Agent Pipeline Provider</h2>
              <p className="text-xs text-slate-400">Controls Stages 1–5: Brief, Creative Concepts, Ad Copy, and Art Direction prompts.</p>
            </div>
          </div>
          <Badge tone="brand">
            Active: {settings.provider.toUpperCase()} ({settings.model})
          </Badge>
        </div>

        <div className="grid gap-3 sm:grid-cols-2" role="radiogroup" aria-label="AI provider">
          {(["anthropic", "gemini"] as const).map((p) => {
            const configured = p === "anthropic" ? settings.anthropicConfigured : settings.geminiConfigured;
            const checked = activeProvider === p;
            return (
              <button
                key={p}
                type="button"
                role="radio"
                aria-checked={checked}
                onClick={() => selectProvider(p)}
                className={`rounded-2xl border p-4 text-left transition-all duration-200 ${
                  checked
                    ? "border-brand-primary bg-brand-primary/10 shadow-glow-sm"
                    : "border-white/10 bg-white/[0.02] hover:border-white/20"
                }`}
              >
                <div className="flex items-center justify-between">
                  <span className="text-sm font-bold capitalize text-white">{p}</span>
                  <Badge tone={configured ? "success" : "danger"}>
                    {configured ? "Configured" : "Missing Key"}
                  </Badge>
                </div>
                <p className="mt-1 text-xs text-slate-400">
                  {p === "anthropic" ? "Claude 3.5 Sonnet Tool-Use Engine." : "Google Gemini Function-Calling Pipeline."}
                </p>
              </button>
            );
          })}
        </div>

        <div className="max-w-md space-y-1.5 pt-2">
          <label className="text-xs font-bold uppercase tracking-wider text-slate-400">Model Identifier</label>
          <Input
            value={activeModel}
            onChange={(e) => setModel(e.target.value)}
            placeholder="e.g. claude-sonnet-5 or gemini-3.6-flash"
          />
        </div>

        {saveLLM.isError && (
          <ErrorBanner message={saveLLM.error instanceof ApiError ? saveLLM.error.message : "Could not save setting."} />
        )}

        <div className="flex justify-end pt-2">
          <Button
            onClick={() => saveLLM.mutate()}
            disabled={saveLLM.isPending || !activeModel.trim()}
            loading={saveLLM.isPending}
            loadingText="Switching..."
          >
            Update Pipeline Model
          </Button>
        </div>
      </Card>

      {/* API Key Vault */}
      <Card className="space-y-4">
        <CardHeader
          title="API Key Vault & Multi-Provider Rotation"
          description="Update credentials for any provider below. Keys are securely stored and hot-swapped in memory immediately without restarting the server."
        />
        <div className="space-y-3">
          <ApiKeyRow provider="flux" configured={settings.fluxConfigured} />
          <ApiKeyRow provider="ideogram" configured={settings.ideogramConfigured} />
          <ApiKeyRow provider="openai" configured={settings.openaiConfigured} />
          <ApiKeyRow provider="anthropic" configured={settings.anthropicConfigured} />
          <ApiKeyRow provider="gemini" configured={settings.geminiConfigured} />
        </div>
      </Card>
    </div>
  );
}
