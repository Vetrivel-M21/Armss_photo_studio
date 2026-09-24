// Dialog-level motion intentionally stays on tailwindcss-animate + Radix
// data-state — see web/src/lib/motion.ts for the Framer Motion boundary used
// everywhere else (route transitions, card hover/tap, gallery reveal).
import type { ReactNode } from "react";
import * as Dialog from "@radix-ui/react-dialog";
import { X } from "lucide-react";
import { cn } from "../../lib/cn";

type Side = "center" | "right" | "left";
type Size = "sm" | "md" | "lg";

interface Props {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  title: ReactNode;
  description?: ReactNode;
  children: ReactNode;
  footer?: ReactNode;
  side?: Side;
  size?: Size;
  // An admin could lose in-progress edits to an accidental backdrop click —
  // pass true to require the explicit close control instead.
  disableOutsideClose?: boolean;
}

const sizeClasses: Record<Size, string> = {
  sm: "max-w-sm",
  md: "max-w-md",
  lg: "max-w-lg",
};

const sideContentClasses: Record<Side, string> = {
  right:
    "fixed inset-y-0 right-0 h-full w-full max-w-lg bg-[#0E121E]/95 backdrop-blur-2xl border-l border-white/10 text-slate-100 shadow-2xl data-[state=open]:animate-in data-[state=open]:slide-in-from-right data-[state=closed]:animate-out data-[state=closed]:slide-out-to-right z-50 overflow-y-auto",
  left:
    "fixed inset-y-0 left-0 h-full w-full max-w-lg bg-[#0E121E]/95 backdrop-blur-2xl border-r border-white/10 text-slate-100 shadow-2xl data-[state=open]:animate-in data-[state=open]:slide-in-from-left data-[state=closed]:animate-out data-[state=closed]:slide-out-to-left z-50 overflow-y-auto",
  center:
    "fixed left-1/2 top-1/2 w-[calc(100%-2rem)] -translate-x-1/2 -translate-y-1/2 rounded-2xl bg-[#0E121E]/95 backdrop-blur-2xl border border-white/10 text-slate-100 p-6 shadow-2xl data-[state=open]:animate-in data-[state=open]:fade-in data-[state=open]:zoom-in-95 data-[state=closed]:animate-out data-[state=closed]:fade-out data-[state=closed]:zoom-out-95 z-50 max-h-[85vh] overflow-y-auto",
};

export function Modal({
  open,
  onOpenChange,
  title,
  description,
  children,
  footer,
  side = "center",
  size = "md",
  disableOutsideClose = false,
}: Props) {
  const contentClass =
    side === "center" ? cn(sideContentClasses.center, sizeClasses[size]) : sideContentClasses[side];

  return (
    <Dialog.Root open={open} onOpenChange={onOpenChange}>
      <Dialog.Portal>
        <Dialog.Overlay className="fixed inset-0 bg-black/75 backdrop-blur-md data-[state=open]:animate-in data-[state=open]:fade-in data-[state=closed]:animate-out data-[state=closed]:fade-out z-40" />
        <Dialog.Content
          className={contentClass}
          onPointerDownOutside={(e) => {
            if (disableOutsideClose) e.preventDefault();
          }}
          onInteractOutside={(e) => {
            if (disableOutsideClose) e.preventDefault();
          }}
        >
          <div className={cn("flex items-center justify-between gap-4", side === "center" ? "mb-4" : "border-b border-white/10 p-6 pb-4")}>
            <div>
              <Dialog.Title className="text-base font-bold text-white tracking-tight">{title}</Dialog.Title>
              {description && (
                <Dialog.Description className="mt-1 text-sm text-slate-400">{description}</Dialog.Description>
              )}
            </div>
            <Dialog.Close asChild>
              <button
                type="button"
                aria-label="Close"
                className="rounded-xl p-2 text-slate-400 hover:bg-white/10 hover:text-white transition-colors"
              >
                <X className="h-4 w-4" />
              </button>
            </Dialog.Close>
          </div>
          <div className={side === "center" ? undefined : "p-6 pt-4"}>{children}</div>
          {footer && (
            <div className={cn("flex justify-end gap-2", side === "center" ? "mt-4" : "border-t border-white/10 p-6 pt-4")}>
              {footer}
            </div>
          )}
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}
