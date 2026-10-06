import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { useNavigate } from "react-router-dom";
import { ArrowRight, Calendar, FolderKanban, Plus, Search, Sparkles } from "lucide-react";
import { projectsApi } from "../api/projects";
import type { ProjectState } from "../api/types";
import { Card } from "../components/common/Card";
import { Badge, type BadgeTone } from "../components/common/Badge";
import { Button } from "../components/common/Button";
import { Input } from "../components/common/Input";
import { Spinner } from "../components/common/Spinner";
import { ErrorBanner } from "../components/common/ErrorBanner";

const STATE_CONFIG: Record<ProjectState, { label: string; tone: BadgeTone }> = {
  draft: { label: "Draft", tone: "neutral" },
  briefing: { label: "Briefing", tone: "info" },
  concepts_ready: { label: "Concepts Ready", tone: "brand" },
  developing: { label: "In Development", tone: "warning" },
  render_ready: { label: "Artwork Rendered", tone: "brand" },
  composited: { label: "Composited & Ready", tone: "success" },
  approved: { label: "Approved", tone: "success" },
};

export function ProjectsPage() {
  const navigate = useNavigate();
  const [search, setSearch] = useState("");

  const { data: projects, isLoading, isError, error, refetch } = useQuery({
    queryKey: ["projects"],
    queryFn: projectsApi.list,
  });

  const filtered = (projects ?? []).filter((p) => {
    if (!search.trim()) return true;
    const term = search.toLowerCase();
    const title = (p.title || p.inputText).toLowerCase();
    const state = (STATE_CONFIG[p.state]?.label || p.state).toLowerCase();
    return title.includes(term) || state.includes(term);
  });

  return (
    <div className="space-y-8">
      {/* Header bar */}
      <div className="flex flex-wrap items-center justify-between gap-4 rounded-3xl border border-white/10 bg-gradient-to-r from-[#141824]/95 via-[#161B28]/90 to-[#1D162B]/95 p-6 shadow-glass backdrop-blur-xl">
        <div className="flex items-center gap-3.5">
          <div className="flex h-12 w-12 shrink-0 items-center justify-center rounded-2xl bg-gradient-to-br from-brand-primary via-brand-secondary to-purple-600 shadow-glow">
            <FolderKanban className="h-6 w-6 text-white" aria-hidden="true" />
          </div>
          <div>
            <h1 className="text-xl font-extrabold tracking-tight text-white sm:text-2xl">
              My Posters
            </h1>
            <p className="mt-0.5 text-xs text-slate-400">
              Browse and resume your generated poster projects
            </p>
          </div>
        </div>

        <div className="flex items-center gap-3">
          <Button
            onClick={() => navigate("/")}
            className="gap-2 shadow-glow font-bold text-xs"
          >
            <Plus className="h-4 w-4" aria-hidden="true" />
            Create New Poster
          </Button>
        </div>
      </div>

      {/* Search and counter */}
      <div className="flex flex-wrap items-center justify-between gap-4">
        <div className="relative w-full max-w-sm">
          <Search className="pointer-events-none absolute left-3.5 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-400" />
          <Input
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Search by occasion or status..."
            className="pl-9.5 text-xs"
          />
        </div>

        {projects && (
          <span className="text-xs font-medium text-slate-400">
            Showing {filtered.length} of {projects.length} posters
          </span>
        )}
      </div>

      {isLoading && (
        <div className="flex justify-center py-20">
          <Spinner label="Loading posters..." />
        </div>
      )}

      {isError && (
        <ErrorBanner
          message={error instanceof Error ? error.message : "Failed to load projects."}
          onRetry={() => refetch()}
        />
      )}

      {projects && projects.length === 0 && (
        <div className="flex flex-col items-center justify-center rounded-3xl border border-dashed border-white/10 bg-[#111420]/50 p-12 text-center backdrop-blur-md">
          <div className="flex h-14 w-14 items-center justify-center rounded-2xl bg-white/[0.04] text-slate-400 mb-4">
            <Sparkles className="h-7 w-7 text-brand-primary" />
          </div>
          <h3 className="text-base font-bold text-white">No posters created yet</h3>
          <p className="mt-1 max-w-sm text-xs text-slate-400">
            Pick a template or enter an occasion description to generate your first on-brand poster.
          </p>
          <Button
            onClick={() => navigate("/")}
            className="mt-6 gap-2 shadow-glow text-xs font-bold"
          >
            <Plus className="h-4 w-4" />
            Start First Poster
          </Button>
        </div>
      )}

      {projects && projects.length > 0 && filtered.length === 0 && (
        <div className="rounded-2xl border border-dashed border-white/10 bg-[#111420]/40 p-8 text-center">
          <p className="text-xs text-slate-400">
            No posters found matching &quot;{search}&quot;.{" "}
            <button
              type="button"
              onClick={() => setSearch("")}
              className="font-semibold text-brand-accent hover:underline"
            >
              Clear search
            </button>
          </p>
        </div>
      )}

      {filtered.length > 0 && (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {filtered.map((p) => {
            const config = STATE_CONFIG[p.state] || { label: p.state, tone: "neutral" as BadgeTone };
            const formattedDate = new Date(p.createdAt).toLocaleDateString(undefined, {
              month: "short",
              day: "numeric",
              year: "numeric",
            });

            return (
              <Card
                key={p.id}
                interactive
                padding="md"
                onClick={() => navigate(`/projects/${p.id}`)}
                className="group flex flex-col justify-between gap-4 cursor-pointer hover:border-brand-primary/40"
              >
                <div>
                  <div className="flex items-center justify-between gap-2 mb-2.5">
                    <span className="font-mono text-[10px] text-slate-400">
                      ID: {p.id.slice(0, 8)}
                    </span>
                    <Badge tone={config.tone} size="sm">
                      {config.label}
                    </Badge>
                  </div>

                  <h3 className="text-sm font-bold text-white line-clamp-2 group-hover:text-brand-accent transition-colors">
                    {p.title || p.inputText}
                  </h3>

                  <p className="mt-1.5 text-xs text-slate-400 line-clamp-2">
                    {p.inputText}
                  </p>
                </div>

                <div className="flex items-center justify-between border-t border-white/5 pt-3 text-[11px] text-slate-400">
                  <span className="flex items-center gap-1.5">
                    <Calendar className="h-3.5 w-3.5 text-slate-400" />
                    {formattedDate}
                  </span>
                  <span className="flex items-center gap-1 font-semibold text-brand-accent group-hover:translate-x-0.5 transition-transform">
                    Open Studio
                    <ArrowRight className="h-3.5 w-3.5" />
                  </span>
                </div>
              </Card>
            );
          })}
        </div>
      )}
    </div>
  );
}
