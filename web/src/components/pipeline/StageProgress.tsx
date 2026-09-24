import type { ProjectDetail } from "../../api/types";
import type { StageStatusMap } from "../../hooks/useSSE";

interface Props {
  project: ProjectDetail;
  liveStatuses: StageStatusMap;
}

const STEPS: { key: string; label: string; hasData: (p: ProjectDetail) => boolean }[] = [
  { key: "brief", label: "Brief", hasData: (p) => !!p.stages.brief },
  { key: "concepts", label: "Concepts", hasData: (p) => !!p.stages.concepts },
  { key: "copy", label: "Copy", hasData: (p) => !!p.stages.copy },
  { key: "art_direction", label: "Art direction", hasData: (p) => !!p.stages.art_direction },
];

const dotClasses: Record<string, string> = {
  running: "bg-brand-highlight shadow-[0_0_10px_rgba(251,191,36,0.8)] animate-pulse",
  error: "bg-red-500 shadow-[0_0_10px_rgba(239,68,68,0.8)]",
  done: "bg-gradient-to-r from-brand-primary to-purple-500 shadow-[0_0_10px_rgba(255,42,95,0.6)]",
  pending: "bg-slate-700",
};

export function StageProgress({ project, liveStatuses }: Props) {
  return (
    <ol className="mb-6 flex flex-wrap items-center gap-x-3 gap-y-2.5 rounded-2xl border border-white/10 bg-[#111420]/80 px-4 py-3 text-xs backdrop-blur-md">
      {STEPS.map((step, i) => {
        const live = liveStatuses[step.key];
        const status = live === "running" ? "running" : live === "error" ? "error" : step.hasData(project) ? "done" : "pending";
        const prevDone = i > 0 && STEPS[i - 1].hasData(project) && liveStatuses[STEPS[i - 1].key] !== "error";
        return (
          <li key={step.key} className="flex items-center gap-2.5">
            {i > 0 && (
              <span
                className={`h-0.5 w-5 rounded-full transition-all duration-300 ${
                  prevDone
                    ? "bg-gradient-to-r from-brand-primary via-brand-secondary to-purple-500 shadow-[0_0_8px_rgba(255,42,95,0.4)]"
                    : "bg-slate-800"
                }`}
                aria-hidden="true"
              />
            )}
            <div className="relative flex items-center justify-center">
              {status === "running" && (
                <span className="absolute h-4 w-4 rounded-full bg-brand-highlight/40 animate-ping" />
              )}
              <span className={`relative h-2.5 w-2.5 rounded-full ${dotClasses[status]}`} aria-hidden="true" />
            </div>
            <span
              className={`font-semibold tracking-tight transition-colors ${
                status === "pending"
                  ? "text-slate-500"
                  : status === "running"
                  ? "text-brand-highlight"
                  : "text-slate-200"
              }`}
            >
              {step.label}
            </span>
          </li>
        );
      })}
    </ol>
  );
}
