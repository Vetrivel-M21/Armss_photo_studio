import type { ReactNode } from "react";
import { cn } from "../../lib/cn";

export type BadgeTone = "neutral" | "info" | "success" | "warning" | "danger" | "brand";
type BadgeSize = "sm" | "md";

interface Props {
  children: ReactNode;
  tone?: BadgeTone;
  size?: BadgeSize;
  icon?: ReactNode;
  className?: string;
}

const toneClasses: Record<BadgeTone, string> = {
  neutral: "bg-white/[0.06] text-slate-400 border border-white/10",
  info: "bg-sky-500/15 text-sky-300 border border-sky-500/30",
  success: "bg-emerald-500/15 text-emerald-300 border border-emerald-500/30 shadow-[0_0_12px_-3px_rgba(16,185,129,0.3)]",
  warning: "bg-amber-500/15 text-amber-300 border border-amber-500/30 shadow-[0_0_12px_-3px_rgba(245,158,11,0.3)]",
  danger: "bg-rose-500/15 text-rose-300 border border-rose-500/30 shadow-[0_0_12px_-3px_rgba(244,63,94,0.3)]",
  brand: "bg-gradient-to-r from-brand-primary to-brand-secondary text-white border border-white/20 shadow-glow-sm",
};

const sizeClasses: Record<BadgeSize, string> = {
  sm: "px-2.5 py-0.5 text-[11px]",
  md: "px-3 py-1 text-xs",
};

export function Badge({ children, tone = "neutral", size = "sm", icon, className = "" }: Props) {
  return (
    <span
      className={cn(
        "inline-flex items-center gap-1.5 rounded-full font-semibold backdrop-blur-md transition-colors",
        toneClasses[tone],
        sizeClasses[size],
        className,
      )}
    >
      {icon}
      {children}
    </span>
  );
}
