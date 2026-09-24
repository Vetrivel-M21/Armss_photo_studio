import { useState } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { FileText } from "lucide-react";
import type { Brief, SSEStatus } from "../../api/types";
import { StageCard } from "./StageCard";
import { Button } from "../common/Button";
import { TextArea } from "../common/TextArea";
import { ErrorBanner } from "../common/ErrorBanner";
import { projectsApi } from "../../api/projects";
import { ApiError } from "../../api/client";

interface Props {
  projectId: string;
  brief?: Brief;
  liveStatus?: SSEStatus;
}

function Field({ label, value }: { label: string; value?: string | string[] }) {
  if (!value || (Array.isArray(value) && value.length === 0)) return null;
  return (
    <div className="rounded-xl border border-white/5 bg-white/[0.02] p-2.5">
      <dt className="text-[10px] font-bold uppercase tracking-wider text-slate-400">{label}</dt>
      <dd className="mt-0.5 text-sm font-medium text-slate-100">{Array.isArray(value) ? value.join(", ") : value}</dd>
    </div>
  );
}

export function BriefStage({ projectId, brief, liveStatus }: Props) {
  const [editing, setEditing] = useState(false);
  const [extraDetail, setExtraDetail] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const queryClient = useQueryClient();

  const rerun = useMutation({
    mutationFn: () => projectsApi.rerunBrief(projectId, extraDetail),
    onSuccess: (project) => {
      queryClient.setQueryData(["project", projectId], project);
      setEditing(false);
      setExtraDetail("");
    },
    onSettled: () => setIsSubmitting(false),
  });

  const isBusy = rerun.isPending || isSubmitting;
  const handleRerun = () => {
    if (!extraDetail.trim() || isBusy) return;
    setIsSubmitting(true);
    rerun.mutate();
  };

  return (
    <StageCard
      title="Brief"
      icon={FileText}
      liveStatus={liveStatus}
      hasData={!!brief}
      actions={
        <Button variant="secondary" disabled={isBusy} onClick={() => setEditing((v) => !v)}>
          {editing ? "Cancel" : "Add detail & re-run"}
        </Button>
      }
    >
      {brief && (
        <dl className="grid gap-3 sm:grid-cols-2">
          <Field label="Objective" value={brief.objective} />
          <Field label="Audience" value={brief.audience} />
          <Field label="Key message" value={brief.keyMessage} />
          <Field label="Occasion" value={brief.occasion} />
          <Field label="Event date" value={brief.eventDate} />
          <Field label="Tone" value={brief.tone} />
          <Field label="Format" value={brief.format} />
          <Field label="Dimensions" value={brief.dimensions} />
          <Field label="Palette" value={brief.palette} />
          <Field label="Must include" value={brief.mustInclude} />
          <Field label="Avoid" value={brief.avoid} />
          <Field label="Assumptions made" value={brief.assumptions} />
          <Field label="Open questions" value={brief.questions} />
        </dl>
      )}

      {editing && (
        <div className="mt-4 space-y-2 border-t border-neutral-100 pt-4">
          <TextArea
            rows={2}
            placeholder="Add or correct a detail — this re-runs the brief and every concept built from it."
            value={extraDetail}
            disabled={isBusy}
            onChange={(e) => setExtraDetail(e.target.value)}
          />
          {rerun.isError && (
            <ErrorBanner message={rerun.error instanceof ApiError ? rerun.error.message : "Could not re-run the brief."} />
          )}
          <div className="flex justify-end">
            <Button
              onClick={handleRerun}
              disabled={!extraDetail.trim() || isBusy}
              loading={isBusy}
              loadingText="Re-running brief..."
            >
              Re-run brief
            </Button>
          </div>
        </div>
      )}
    </StageCard>
  );
}
