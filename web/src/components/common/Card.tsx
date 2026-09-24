import type { HTMLAttributes, ReactNode } from "react";
import { cn } from "../../lib/cn";

type Padding = "none" | "sm" | "md" | "lg";

interface Props extends HTMLAttributes<HTMLDivElement> {
  padding?: Padding;
  interactive?: boolean;
}

const paddingClasses: Record<Padding, string> = {
  none: "",
  sm: "p-3",
  md: "p-5",
  lg: "p-8",
};

// Shared card shell — replaces the identical inline
// `rounded-lg border border-neutral-200 bg-white ...` string repeated across
// StartPage/StageCard/admin pages.
export function Card({ children, className = "", padding = "md", interactive = false, ...props }: Props) {
  return (
    <div
      className={cn(
        "rounded-2xl border border-white/10 bg-[#111420]/80 backdrop-blur-xl shadow-glass text-slate-100",
        interactive && "transition-all duration-300 hover:border-white/20 hover:bg-[#161B2B]/90 hover:shadow-glow-sm motion-reduce:transition-none",
        paddingClasses[padding],
        className,
      )}
      {...props}
    >
      {children}
    </div>
  );
}

interface CardHeaderProps {
  title: ReactNode;
  description?: ReactNode;
  actions?: ReactNode;
}

export function CardHeader({ title, description, actions }: CardHeaderProps) {
  return (
    <div className="flex items-start justify-between gap-4">
      <div>
        <h2 className="text-sm font-bold tracking-tight text-white">{title}</h2>
        {description && <p className="mt-1 text-xs text-slate-400">{description}</p>}
      </div>
      {actions && <div className="flex shrink-0 items-center gap-3">{actions}</div>}
    </div>
  );
}
