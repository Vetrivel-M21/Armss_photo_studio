import { useState } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { Check, Lightbulb } from "lucide-react";
import type { Concept, SSEStatus } from "../../api/types";
import { StageCard } from "./StageCard";
import { Button } from "../common/Button";
import { Card } from "../common/Card";
import { ErrorBanner } from "../common/ErrorBanner";
import { cn } from "../../lib/cn";
import { projectsApi } from "../../api/projects";
import { ApiError } from "../../api/client";

interface Props {
  projectId: string;
  concepts?: Concept[];
  selectedIndex?: number;
  liveStatus?: SSEStatus;
}

export function ConceptsStage({ projectId, concepts, selectedIndex, liveStatus }: Props) {
  const queryClient = useQueryClient();
  const [selectedPendingIndex, setSelectedPendingIndex] = useState<number | null>(null);

  const select = useMutation({
    mutationFn: (index: number) => projectsApi.selectConcept(projectId, index),
    onSuccess: (project) => queryClient.setQueryData(["project", projectId], project),
    onSettled: () => setSelectedPendingIndex(null),
  });

  const isBusy = select.isPending || selectedPendingIndex !== null;

  const handleSelect = (index: number) => {
    if (isBusy) return;
    setSelectedPendingIndex(index);
    select.mutate(index);
  };

  return (
    <StageCard
      title="Creative direction"
      icon={Lightbulb}
      liveStatus={liveStatus}
      hasData={!!concepts}
      defaultExpanded={selectedIndex === undefined ? true : undefined}
    >
      {select.isError && (
        <ErrorBanner message={select.error instanceof ApiError ? select.error.message : "Could not select this direction."} />
      )}
      {concepts && (
        <div className="grid gap-3 sm:grid-cols-3">
          {concepts.map((c, i) => {
            const isSelected = selectedIndex === i;
            const isCurrentlyPending = selectedPendingIndex === i;
            return (
              <Card
                key={c.name + i}
                padding="md"
                className={cn(
                  "flex flex-col justify-between gap-3.5 transition-all duration-300",
                  isSelected
                    ? "border-brand-primary bg-brand-primary/10 shadow-glow-sm"
                    : "border-white/10 bg-[#0C0F1A]/90 hover:border-white/20",
                )}
              >
                <div>
                  <div className="flex items-center justify-between gap-1.5">
                    <p className="text-sm font-bold text-white tracking-tight">{c.name}</p>
                    {isSelected && (
                      <span className="flex h-5 w-5 items-center justify-center rounded-full bg-brand-primary text-white shadow-glow-sm">
                        <Check className="h-3 w-3" aria-hidden="true" />
                      </span>
                    )}
                  </div>
                  <p className="mt-2 text-xs leading-relaxed text-slate-300">{c.bigIdea}</p>
                  <div className="mt-3 rounded-lg bg-white/[0.03] p-2 border border-white/5 space-y-1.5">
                    <p className="text-[10px] font-bold uppercase tracking-wider text-slate-400">Visual Metaphor</p>
                    <p className="text-xs text-slate-300">{c.visualMetaphor}</p>
                    <p className="text-[10px] font-bold uppercase tracking-wider text-slate-400 pt-1">Mood &amp; Palette</p>
                    <p className="text-xs text-slate-300">
                      {c.mood} · {c.palette}
                    </p>
                  </div>
                </div>
                <Button
                  variant={isSelected ? "secondary" : "primary"}
                  disabled={isBusy}
                  loading={isCurrentlyPending}
                  loadingText="Selecting..."
                  onClick={() => handleSelect(i)}
                  className="w-full text-xs font-semibold py-2"
                >
                  {isSelected ? "Active Direction" : "Select Concept"}
                </Button>
              </Card>
            );
          })}
        </div>
      )}
    </StageCard>
  );
}
