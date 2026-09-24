import { useState, type ReactNode } from "react";
import { AlertCircle, CheckCircle2, ChevronDown, Circle, type LucideIcon } from "lucide-react";
import type { SSEStatus } from "../../api/types";
import { Card } from "../common/Card";
import { Badge, type BadgeTone } from "../common/Badge";
import { Button } from "../common/Button";
import { cn } from "../../lib/cn";

interface Props {
  title: string;
  icon?: LucideIcon;
  liveStatus?: SSEStatus;
  hasData: boolean;
  actions?: ReactNode;
  children?: ReactNode;
  // Overrides the default "collapse if already done on mount" behavior —
  // e.g. Concepts must stay open until the user actually picks one, even
  // though the concepts themselves already "have data".
  defaultExpanded?: boolean;
}

const statusTone: Record<string, BadgeTone> = {
  running: "warning",
  error: "danger",
  done: "success",
  pending: "neutral",
};

// Every stage has an edit control and a re-run control (spec §12); re-running
// visibly invalidates the stages below it — reflected here by hasData simply
// being false once the server no longer has a current row for this stage.
export function StageCard({ title, icon: Icon, liveStatus, hasData, actions, children, defaultExpanded }: Props) {
  const status = liveStatus === "running" ? "running" : liveStatus === "error" ? "error" : hasData ? "done" : "pending";
  const label = { running: "Active Agent", error: "Failed", done: "Completed", pending: "Waiting" }[status];
  const statusIcon =
    status === "running" ? (
      <span className="h-2 w-2 rounded-full bg-brand-highlight animate-ping" />
    ) : status === "done" ? (
      <CheckCircle2 className="h-3.5 w-3.5 text-emerald-400" aria-hidden="true" />
    ) : status === "error" ? (
      <AlertCircle className="h-3.5 w-3.5 text-rose-400" aria-hidden="true" />
    ) : (
      <Circle className="h-3 w-3 text-slate-500" aria-hidden="true" />
    );

  const [expanded, setExpanded] = useState(() => defaultExpanded ?? !hasData);

  return (
    <Card
      className={cn(
        "relative transition-all duration-300",
        status === "running"
          ? "border-brand-highlight/50 bg-[#141926] shadow-glow-amber"
          : status === "done"
          ? "border-white/10 bg-[#111420]/90"
          : "border-white/5 bg-[#0D101A]/60 opacity-85",
      )}
    >
      {status === "running" && (
        <div className="pointer-events-none absolute inset-0 overflow-hidden rounded-2xl" aria-hidden="true">
          <div className="absolute inset-y-0 left-0 w-1/2 animate-shimmer bg-gradient-to-r from-transparent via-brand-highlight/15 to-transparent" />
        </div>
      )}
      <header className="relative mb-3 flex items-center justify-between gap-3">
        <div className="flex items-center gap-3">
          {Icon && (
            <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-xl border border-white/10 bg-white/[0.05] shadow-inner">
              <Icon className="h-4 w-4 text-brand-accent" aria-hidden="true" />
            </div>
          )}
          <div className="flex items-center gap-2.5">
            <h3 className="text-sm font-bold tracking-tight text-white">{title}</h3>
            <Badge tone={statusTone[status]} icon={statusIcon}>
              {label}
            </Badge>
          </div>
        </div>
        <div className="flex items-center gap-2">
          {actions}
          {hasData && (
            <Button
              type="button"
              variant="ghost"
              iconOnly
              aria-label="Toggle details"
              onClick={() => setExpanded((v) => !v)}
              className="text-slate-400 hover:text-white"
            >
              <ChevronDown className={cn("h-4 w-4 transition-transform", expanded && "rotate-180")} aria-hidden="true" />
            </Button>
          )}
        </div>
      </header>
      <div
        className={cn(
          "grid transition-[grid-template-rows] duration-200 motion-reduce:transition-none",
          expanded ? "grid-rows-[1fr]" : "grid-rows-[0fr]",
        )}
      >
        <div className="overflow-hidden pt-1">{children}</div>
      </div>
    </Card>
  );
}
