import { useQuery } from "@tanstack/react-query";
import { useParams } from "react-router-dom";
import { projectsApi } from "../api/projects";
import { useSSE } from "../hooks/useSSE";
import { StageList } from "../components/pipeline/StageList";
import { StageProgress } from "../components/pipeline/StageProgress";
import { RenderPanel } from "../components/pipeline/RenderPanel";
import { Spinner } from "../components/common/Spinner";
import { ErrorBanner } from "../components/common/ErrorBanner";
import { Badge } from "../components/common/Badge";
import { ApiError } from "../api/client";
import type { ProjectState } from "../api/types";

const stateLabels: Record<ProjectState, string> = {
  draft: "Draft",
  briefing: "Briefing",
  concepts_ready: "Concepts ready",
  developing: "Developing",
  render_ready: "Render ready",
  composited: "Composited",
  approved: "Approved",
};

// The working view (spec §12): a vertical run of stages on the left, a sticky
// render panel on the right. SSE drives live per-stage status; the canonical
// project state refetches whenever a stage finishes or errors.
export function PipelinePage() {
  const { id } = useParams<{ id: string }>();
  const { statuses } = useSSE(id);

  const { data: project, isLoading, isError, error, refetch } = useQuery({
    queryKey: ["project", id],
    queryFn: () => projectsApi.get(id!),
    enabled: !!id,
  });

  if (isLoading) {
    return (
      <div className="flex justify-center py-16">
        <Spinner label="Loading project..." />
      </div>
    );
  }
  if (isError || !project) {
    return (
      <ErrorBanner
        message={error instanceof ApiError ? error.message : "Could not load this project."}
        onRetry={() => refetch()}
      />
    );
  }

  return (
    <div className="space-y-6">
      {/* Studio Header Bar */}
      <div className="flex flex-wrap items-center justify-between gap-4 rounded-2xl border border-white/10 bg-gradient-to-r from-[#141824]/95 via-[#161B28]/90 to-[#1D162B]/95 p-5 shadow-glass backdrop-blur-xl">
        <div>
          <div className="flex items-center gap-2 mb-1">
            <span className="text-[10px] font-bold uppercase tracking-wider text-slate-400">Poster Job</span>
            <span className="text-slate-600">•</span>
            <span className="font-mono text-[10px] text-slate-400">ID: {project.id.slice(0, 8)}</span>
          </div>
          <h1 className="text-xl font-extrabold tracking-tight text-white sm:text-2xl">
            {project.title || project.inputText}
          </h1>
        </div>
        <div className="flex items-center gap-3">
          <Badge tone="brand" size="md" className="shadow-glow-sm">
            <span className="h-1.5 w-1.5 rounded-full bg-white animate-ping mr-1" />
            {stateLabels[project.state]}
          </Badge>
        </div>
      </div>

      {/* 2-Column Responsive Studio Workspace */}
      <div className="grid gap-6 lg:grid-cols-[1fr_380px] xl:grid-cols-[1fr_420px]">
        <div className="space-y-4">
          <StageProgress project={project} liveStatuses={statuses} />
          <StageList project={project} liveStatuses={statuses} />
        </div>
        <div>
          <RenderPanel projectId={project.id} art={project.stages.art_direction?.payload} renders={project.renders} />
        </div>
      </div>
    </div>
  );
}
