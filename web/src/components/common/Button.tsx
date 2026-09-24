import type { ButtonHTMLAttributes, MouseEvent } from "react";
import { useRef } from "react";
import { motion } from "framer-motion";
import { cn } from "../../lib/cn";
import { useMotionSafe } from "../../hooks/useMotionSafe";
import { springSnappy } from "../../lib/motion";

type Variant = "primary" | "secondary" | "danger" | "ghost";

// framer-motion's motion.button redefines onDrag*/onAnimation* with its own
// gesture/animation-lifecycle signatures, which conflict with the plain DOM
// event types in ButtonHTMLAttributes — omit them since nothing here uses them.
type NativeButtonProps = Omit<
  ButtonHTMLAttributes<HTMLButtonElement>,
  "onDrag" | "onDragStart" | "onDragEnd" | "onAnimationStart" | "onAnimationEnd" | "onAnimationIteration"
>;

export interface ButtonProps extends NativeButtonProps {
  variant?: Variant;
  // Switches padding to a square (icon-only). Callers MUST pass `aria-label`
  // when this is true — not enforced at the type level.
  iconOnly?: boolean;
  // Loading and single-click protection: locks the button synchronously on click,
  // preventing double-firing and duplicate expensive AI calls.
  loading?: boolean;
  loadingText?: string;
  // Debounce time in ms to block rapid multi-clicks (default 400ms)
  clickDebounceMs?: number;
}

const variantClasses: Record<Variant, string> = {
  primary:
    "relative bg-gradient-to-r from-brand-primary via-brand-secondary to-purple-600 text-white font-semibold shadow-glow hover:brightness-110 hover:shadow-glow-violet disabled:opacity-40 disabled:shadow-none disabled:brightness-100",
  secondary:
    "bg-white/[0.06] text-slate-100 border border-white/15 hover:bg-white/[0.12] hover:border-white/30 disabled:opacity-40 font-medium",
  danger:
    "bg-red-500/20 text-red-300 border border-red-500/30 hover:bg-red-500/30 hover:border-red-500/50 disabled:opacity-40 font-medium",
  ghost:
    "bg-transparent text-slate-300 hover:bg-white/[0.08] hover:text-white disabled:opacity-40 font-medium",
};

export function Button({
  variant = "primary",
  iconOnly = false,
  loading = false,
  loadingText,
  clickDebounceMs = 400,
  disabled = false,
  className = "",
  onClick,
  children,
  ...props
}: ButtonProps) {
  const motionSafe = useMotionSafe();
  const lastClickRef = useRef<number>(0);

  const handleClick = (e: MouseEvent<HTMLButtonElement>) => {
    if (loading || disabled) {
      e.preventDefault();
      e.stopPropagation();
      return;
    }
    const now = Date.now();
    if (clickDebounceMs > 0 && now - lastClickRef.current < clickDebounceMs) {
      e.preventDefault();
      e.stopPropagation();
      return;
    }
    lastClickRef.current = now;
    onClick?.(e);
  };

  const isDisabled = disabled || loading;

  return (
    <motion.button
      whileHover={motionSafe && !isDisabled ? { scale: 1.02 } : undefined}
      whileTap={motionSafe && !isDisabled ? { scale: 0.96 } : undefined}
      transition={springSnappy}
      disabled={isDisabled}
      aria-busy={loading ? "true" : undefined}
      onClick={handleClick}
      className={cn(
        "inline-flex items-center justify-center gap-2 rounded-xl text-sm font-medium transition-all duration-200 disabled:cursor-not-allowed motion-reduce:transition-none active:outline-none",
        isDisabled && "pointer-events-none select-none",
        iconOnly ? "p-2.5" : "px-4 py-2.5",
        variantClasses[variant],
        className,
      )}
      {...props}
    >
      {loading ? (
        <>
          <span
            className="inline-block h-4 w-4 shrink-0 animate-spin rounded-full border-2 border-current border-t-transparent"
            aria-hidden="true"
          />
          {loadingText ? <span>{loadingText}</span> : children}
        </>
      ) : (
        children
      )}
    </motion.button>
  );
}
