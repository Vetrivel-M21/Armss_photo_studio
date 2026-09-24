import { FormEvent, useRef, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { AlertTriangle, CheckCircle2, LayoutTemplate } from "lucide-react";
import { templatesApi } from "../api/templates";
import { TemplateTable } from "../components/admin/TemplateTable";
import { TemplateEditDrawer } from "../components/admin/TemplateEditDrawer";
import { Button } from "../components/common/Button";
import { Input } from "../components/common/Input";
import { TextArea } from "../components/common/TextArea";
import { Spinner } from "../components/common/Spinner";
import { ErrorBanner } from "../components/common/ErrorBanner";
import { Card, CardHeader } from "../components/common/Card";
import { Toggle } from "../components/common/Toggle";
import type { Template } from "../api/types";
import { ApiError } from "../api/client";

type AddMode = "image" | "prompt";

// Admin upload flow (spec §8): attach image → run Style Analyst → admin
// reviews and edits the extracted profile → save. A template can also be
// created from a composition prompt alone, with no reference image — either
// one at a time here, or all 40 starter templates in one click below.
export function AdminTemplatesPage() {
  const [editing, setEditing] = useState<Template | null>(null);
  const [mode, setMode] = useState<AddMode>("image");
  const [title, setTitle] = useState("");
  const [occasionTags, setOccasionTags] = useState("");
  const [season, setSeason] = useState("");
  const [compositionPrompt, setCompositionPrompt] = useState("");
  const fileRef = useRef<HTMLInputElement>(null);
  const queryClient = useQueryClient();

  const { data: templates, isLoading } = useQuery({
    queryKey: ["admin-templates"],
    queryFn: () => templatesApi.list(),
  });

  function resetForm() {
    setTitle("");
    setOccasionTags("");
    setSeason("");
    setCompositionPrompt("");
    if (fileRef.current) fileRef.current.value = "";
  }

  const upload = useMutation({
    mutationFn: (form: FormData) => templatesApi.create(form),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin-templates"] });
      resetForm();
    },
  });

  const importLibrary = useMutation({
    mutationFn: () => templatesApi.importStarterLibrary(),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["admin-templates"] }),
  });

  function onSubmit(e: FormEvent) {
    e.preventDefault();
    if (!title) return;
    const form = new FormData();
    form.append("title", title);
    form.append("occasionTags", occasionTags);
    form.append("season", season);
    if (mode === "image") {
      const file = fileRef.current?.files?.[0];
      if (!file) return;
      form.append("file", file);
    } else {
      if (!compositionPrompt.trim()) return;
      form.append("compositionPrompt", compositionPrompt);
    }
    upload.mutate(form);
  }

  const totalCount = templates?.length ?? 0;
  const activeCount = templates?.filter((t) => t.isActive).length ?? 0;
  const needsReviewCount = templates?.filter((t) => t.styleProfileStatus !== "ready").length ?? 0;

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-3">
        <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-gradient-to-br from-brand-primary to-brand-secondary shadow-sm">
          <LayoutTemplate className="h-5 w-5 text-white" aria-hidden="true" />
        </div>
        <div>
          <h1 className="font-serif text-xl font-semibold text-brand-primary">Template library</h1>
          <p className="mt-1 text-sm text-neutral-500">Upload reference designs and correct their extracted style profile.</p>
        </div>
      </div>

      {!isLoading && (
        <div className="grid grid-cols-3 gap-3 sm:gap-4">
          <div className="flex items-center gap-3 rounded-2xl border border-white/10 bg-[#0E121E]/90 p-4 shadow-glass backdrop-blur-xl">
            <span className="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-brand-primary/15 border border-brand-primary/30 text-brand-primary">
              <LayoutTemplate className="h-5 w-5" aria-hidden="true" />
            </span>
            <div className="min-w-0">
              <p className="text-xl font-bold leading-tight text-white">{totalCount}</p>
              <p className="truncate text-xs text-slate-400 font-medium">Total Templates</p>
            </div>
          </div>
          <div className="flex items-center gap-3 rounded-2xl border border-white/10 bg-[#0E121E]/90 p-4 shadow-glass backdrop-blur-xl">
            <span className="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-emerald-500/15 border border-emerald-500/30 text-emerald-400">
              <CheckCircle2 className="h-5 w-5" aria-hidden="true" />
            </span>
            <div className="min-w-0">
              <p className="text-xl font-bold leading-tight text-white">{activeCount}</p>
              <p className="truncate text-xs text-slate-400 font-medium">Active Library</p>
            </div>
          </div>
          <div className="flex items-center gap-3 rounded-2xl border border-white/10 bg-[#0E121E]/90 p-4 shadow-glass backdrop-blur-xl">
            <span className="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-amber-500/15 border border-amber-500/30 text-amber-400">
              <AlertTriangle className="h-5 w-5" aria-hidden="true" />
            </span>
            <div className="min-w-0">
              <p className="text-xl font-bold leading-tight text-white">{needsReviewCount}</p>
              <p className="truncate text-xs text-slate-400 font-medium">Needs Review</p>
            </div>
          </div>
        </div>
      )}

      <Card>
        <CardHeader
          title="Starter template library"
          description="40 ready-made composition prompts — no image required."
          actions={
            <>
              {importLibrary.isSuccess && (
                <span className="text-xs text-neutral-500">
                  Imported {importLibrary.data.imported}, skipped {importLibrary.data.skipped} already present.
                </span>
              )}
              <Button onClick={() => importLibrary.mutate()} disabled={importLibrary.isPending}>
                {importLibrary.isPending ? "Importing..." : "Import starter template library"}
              </Button>
            </>
          }
        />
      </Card>
      {importLibrary.isError && (
        <ErrorBanner message={importLibrary.error instanceof ApiError ? importLibrary.error.message : "Could not import the starter library."} />
      )}

      <form onSubmit={onSubmit}>
        <Card className="space-y-3">
          <div className="flex items-center justify-between">
            <h2 className="text-sm font-semibold text-neutral-700">Add a template</h2>
            <Toggle
              options={["Upload image", "Write a prompt"]}
              value={mode === "image" ? 0 : 1}
              onChange={(v) => setMode(v === 0 ? "image" : "prompt")}
            />
          </div>
          {upload.isError && (
            <ErrorBanner message={upload.error instanceof ApiError ? upload.error.message : "Could not save this template."} />
          )}
          <div className="grid gap-3 sm:grid-cols-3">
            <Input placeholder="Title" value={title} onChange={(e) => setTitle(e.target.value)} required />
            <Input placeholder="Occasion tags (comma separated)" value={occasionTags} onChange={(e) => setOccasionTags(e.target.value)} />
            <Input placeholder="Season" value={season} onChange={(e) => setSeason(e.target.value)} />
          </div>
          {mode === "image" ? (
            <input ref={fileRef} type="file" accept="image/png,image/jpeg,image/webp" required className="text-sm" />
          ) : (
            <TextArea
              rows={4}
              placeholder="Describe the composition, layout, lighting, and decorative language — not event-specific content."
              value={compositionPrompt}
              onChange={(e) => setCompositionPrompt(e.target.value)}
            />
          )}
          <div className="flex justify-end">
            <Button type="submit" disabled={upload.isPending}>
              {upload.isPending ? "Saving..." : mode === "image" ? "Upload & analyse" : "Save template"}
            </Button>
          </div>
        </Card>
      </form>

      {isLoading ? (
        <Spinner label="Loading templates..." />
      ) : (
        <TemplateTable templates={templates ?? []} onEdit={setEditing} />
      )}

      {editing && <TemplateEditDrawer template={editing} onClose={() => setEditing(null)} />}
    </div>
  );
}
