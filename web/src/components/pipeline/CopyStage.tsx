import { useEffect, useState } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { Type } from "lucide-react";
import type { Copy, SSEStatus } from "../../api/types";
import { StageCard } from "./StageCard";
import { Button } from "../common/Button";
import { Input } from "../common/Input";
import { ErrorBanner } from "../common/ErrorBanner";
import { projectsApi } from "../../api/projects";
import { ApiError } from "../../api/client";

interface Props {
  projectId: string;
  copy?: Copy;
  liveStatus?: SSEStatus;
}

// Editing the headline and re-running invalidates copy and art direction (not
// the brief or concepts) — acceptance criterion 4. Saving here calls PATCH
// /copy, which the backend also uses to regenerate art direction from the
// edited copy when AI is available.
export function CopyStage({ projectId, copy, liveStatus }: Props) {
  const [editing, setEditing] = useState(false);
  const [draft, setDraft] = useState<Copy | undefined>(copy);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const queryClient = useQueryClient();

  useEffect(() => setDraft(copy), [copy]);

  const save = useMutation({
    mutationFn: (value: Copy) => projectsApi.patchCopy(projectId, value),
    onSuccess: (project) => {
      queryClient.setQueryData(["project", projectId], project);
      setEditing(false);
    },
    onSettled: () => setIsSubmitting(false),
  });

  const isBusy = save.isPending || isSubmitting;
  const handleSave = () => {
    if (!draft || isBusy) return;
    setIsSubmitting(true);
    save.mutate(draft);
  };

  return (
    <StageCard
      title="Copy"
      icon={Type}
      liveStatus={liveStatus}
      hasData={!!copy}
      actions={
        copy && (
          <Button variant="secondary" disabled={isBusy} onClick={() => setEditing((v) => !v)}>
            {editing ? "Cancel" : "Edit"}
          </Button>
        )
      }
    >
      {save.isError && (
        <ErrorBanner message={save.error instanceof ApiError ? save.error.message : "Could not save this copy."} />
      )}

      {copy && !editing && (
        <div className="space-y-3 rounded-xl border border-white/5 bg-white/[0.02] p-4">
          <div>
            <span className="text-[10px] font-bold uppercase tracking-wider text-slate-400">Headline</span>
            <p className="mt-0.5 text-lg font-bold text-white tracking-tight">{copy.headline}</p>
          </div>
          <div>
            <span className="text-[10px] font-bold uppercase tracking-wider text-slate-400">Subheadline</span>
            <p className="mt-0.5 text-sm text-slate-300">{copy.subhead}</p>
          </div>
          <div className="flex flex-wrap items-center gap-2 pt-1">
            <span className="rounded-full bg-brand-primary/20 border border-brand-primary/30 px-3 py-1 text-xs font-bold text-brand-accent">
              CTA: {copy.cta}
            </span>
            {copy.valueWords.length > 0 && (
              <span className="text-xs text-slate-400 font-mono">
                {copy.valueWords.join(" • ")}
              </span>
            )}
          </div>
          {copy.alternates.length > 0 && (
            <div className="border-t border-white/10 pt-3 mt-2">
              <p className="text-[10px] font-bold uppercase tracking-wider text-slate-400">Alternate Catchphrases</p>
              <ul className="mt-1 space-y-1 text-xs text-slate-300">
                {copy.alternates.map((a) => (
                  <li key={a} className="flex items-center gap-2">
                    <span className="h-1 w-1 rounded-full bg-brand-accent" />
                    {a}
                  </li>
                ))}
              </ul>
            </div>
          )}
        </div>
      )}

      {editing && draft && (
        <div className="space-y-3">
          <div className="space-y-1">
            <label className="text-xs font-medium text-neutral-500">Headline</label>
            <Input value={draft.headline} onChange={(e) => setDraft({ ...draft, headline: e.target.value })} />
          </div>
          <div className="space-y-1">
            <label className="text-xs font-medium text-neutral-500">Subhead</label>
            <Input value={draft.subhead} onChange={(e) => setDraft({ ...draft, subhead: e.target.value })} />
          </div>
          <div className="space-y-1">
            <label className="text-xs font-medium text-neutral-500">Call to action</label>
            <Input value={draft.cta} onChange={(e) => setDraft({ ...draft, cta: e.target.value })} />
          </div>
          <div className="flex justify-end">
            <Button
              onClick={handleSave}
              disabled={isBusy}
              loading={isBusy}
              loadingText="Saving & re-running art direction..."
            >
              Save & re-run art direction
            </Button>
          </div>
        </div>
      )}
    </StageCard>
  );
}
