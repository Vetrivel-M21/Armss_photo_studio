import type { TextareaHTMLAttributes } from "react";
import { cn } from "../../lib/cn";

export function TextArea({ className = "", ...props }: TextareaHTMLAttributes<HTMLTextAreaElement>) {
  return (
    <textarea
      className={cn(
        "w-full rounded-xl border border-white/10 bg-[#0E121D]/90 px-3.5 py-2.5 text-sm text-slate-100 placeholder:text-slate-500 transition-all duration-200 focus:border-brand-primary focus:bg-[#141929] focus:outline-none focus:ring-2 focus:ring-brand-primary/30",
        className,
      )}
      {...props}
    />
  );
}
