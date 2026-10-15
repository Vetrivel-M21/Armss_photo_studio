import React, { useState } from "react";
import { Move, Target, Check, RefreshCw } from "lucide-react";
import { cn } from "../../lib/cn";

interface Props {
  containerRef: React.RefObject<HTMLDivElement | null>;
  x: number;
  y: number;
  w: number;
  h: number;
  onChange: (coords: { x: number; y: number; w: number; h: number }) => void;
  onApply: () => void;
  onSnapCenter: () => void;
  isTrulyCentered: boolean;
  isCompositing?: boolean;
  disabled?: boolean;
  logoUrl?: string;
}

export function InteractiveLogoOverlay({
  containerRef,
  x,
  y,
  w,
  h,
  onChange,
  onApply,
  onSnapCenter,
  isTrulyCentered,
  isCompositing = false,
  disabled = false,
  logoUrl,
}: Props) {
  const [isDragging, setIsDragging] = useState(false);
  const [isResizing, setIsResizing] = useState(false);
  const [dragStart, setDragStart] = useState<{ pointerX: number; pointerY: number; initX: number; initY: number } | null>(null);
  const [resizeStart, setResizeStart] = useState<{ pointerX: number; pointerY: number; initW: number; initH: number; initX: number } | null>(null);

  // Dragging handler
  const handlePointerDownDrag = (e: React.PointerEvent<HTMLDivElement>) => {
    if (disabled || isResizing) return;
    e.stopPropagation();
    e.preventDefault();
    (e.currentTarget as HTMLElement).setPointerCapture(e.pointerId);
    setIsDragging(true);
    setDragStart({
      pointerX: e.clientX,
      pointerY: e.clientY,
      initX: x,
      initY: y,
    });
  };

  const handlePointerMoveDrag = (e: React.PointerEvent<HTMLDivElement>) => {
    if (!isDragging || !dragStart || !containerRef.current) return;
    const rect = containerRef.current.getBoundingClientRect();
    if (rect.width <= 0 || rect.height <= 0) return;

    const dx = (e.clientX - dragStart.pointerX) / rect.width;
    const dy = (e.clientY - dragStart.pointerY) / rect.height;

    let newX = Math.min(Math.max(dragStart.initX + dx, 0), 1 - w);
    const newY = Math.min(Math.max(dragStart.initY + dy, 0), 1 - h);

    // Magnetic snap to horizontal 50% center
    const centerAxis = newX + w / 2;
    if (Math.abs(centerAxis - 0.5) < 0.035) {
      newX = (1.0 - w) / 2.0;
    }

    onChange({
      x: parseFloat(newX.toFixed(4)),
      y: parseFloat(newY.toFixed(4)),
      w,
      h,
    });
  };

  const handlePointerUpDrag = (e: React.PointerEvent<HTMLDivElement>) => {
    if (isDragging) {
      try {
        (e.currentTarget as HTMLElement).releasePointerCapture(e.pointerId);
      } catch {
        // Ignore if already released
      }
      setIsDragging(false);
      setDragStart(null);
    }
  };

  // Resizing handler (from bottom-right handle)
  const handlePointerDownResize = (e: React.PointerEvent<HTMLDivElement>) => {
    if (disabled) return;
    e.stopPropagation();
    e.preventDefault();
    (e.currentTarget as HTMLElement).setPointerCapture(e.pointerId);
    setIsResizing(true);
    setResizeStart({
      pointerX: e.clientX,
      pointerY: e.clientY,
      initW: w,
      initH: h,
      initX: x,
    });
  };

  const handlePointerMoveResize = (e: React.PointerEvent<HTMLDivElement>) => {
    if (!isResizing || !resizeStart || !containerRef.current) return;
    const rect = containerRef.current.getBoundingClientRect();
    if (rect.width <= 0 || rect.height <= 0) return;

    const dx = (e.clientX - resizeStart.pointerX) / rect.width;
    const aspectRatio = resizeStart.initH / resizeStart.initW;

    const newW = Math.min(Math.max(resizeStart.initW + dx, 0.08), 0.7);
    const newH = Math.min(Math.max(newW * aspectRatio, 0.04), 0.5);

    // If currently centered, scale outward symmetrically from center
    let newX = x;
    if (isTrulyCentered) {
      newX = (1.0 - newW) / 2.0;
    } else {
      newX = Math.min(resizeStart.initX, 1 - newW);
    }

    onChange({
      x: parseFloat(newX.toFixed(4)),
      y: parseFloat(y.toFixed(4)),
      w: parseFloat(newW.toFixed(4)),
      h: parseFloat(newH.toFixed(4)),
    });
  };

  const handlePointerUpResize = (e: React.PointerEvent<HTMLDivElement>) => {
    if (isResizing) {
      try {
        (e.currentTarget as HTMLElement).releasePointerCapture(e.pointerId);
      } catch {
        // Ignore if already released
      }
      setIsResizing(false);
      setResizeStart(null);
    }
  };

  const percentLeft = Math.max(0, Math.min(100, x * 100));
  const percentTop = Math.max(0, Math.min(100, y * 100));
  const percentW = Math.max(5, Math.min(100, w * 100));
  const percentH = Math.max(3, Math.min(100, h * 100));

  return (
    <>
      {/* Center alignment guide line when snapped */}
      {(isTrulyCentered || isDragging) && (
        <div
          className={cn(
            "pointer-events-none absolute inset-y-0 left-1/2 w-px -translate-x-1/2 z-10 transition-opacity duration-200",
            isTrulyCentered
              ? "bg-emerald-400/80 shadow-[0_0_8px_rgba(52,211,153,0.8)] opacity-100"
              : "border-l border-dashed border-white/20 opacity-40"
          )}
        />
      )}

      {/* Interactive Draggable Box */}
      <div
        style={{
          left: `${percentLeft}%`,
          top: `${percentTop}%`,
          width: `${percentW}%`,
          height: `${percentH}%`,
        }}
        onPointerDown={handlePointerDownDrag}
        onPointerMove={handlePointerMoveDrag}
        onPointerUp={handlePointerUpDrag}
        onPointerCancel={handlePointerUpDrag}
        className={cn(
          "absolute z-20 touch-none select-none rounded-lg transition-shadow",
          "border-2 border-dashed cursor-grab active:cursor-grabbing",
          isDragging || isResizing
            ? "border-amber-400 bg-amber-400/10 shadow-[0_0_20px_rgba(251,191,36,0.35)] ring-2 ring-amber-400/50"
            : isTrulyCentered
            ? "border-emerald-400 bg-emerald-400/10 hover:border-emerald-300 ring-1 ring-emerald-400/30"
            : "border-brand-primary bg-brand-primary/10 hover:border-brand-highlight ring-1 ring-brand-primary/30",
          disabled && "pointer-events-none opacity-50"
        )}
      >
        {/* Preview image inside or placeholder */}
        <div className="flex h-full w-full items-center justify-center p-1 pointer-events-none">
          {logoUrl ? (
            <img src={logoUrl} alt="Logo" className="h-full w-full object-contain filter drop-shadow" />
          ) : (
            <div className="flex flex-col items-center justify-center text-center">
              <Move className="h-4 w-4 text-white/80 animate-pulse" />
              <span className="text-[9px] font-bold tracking-wider uppercase text-white/90 drop-shadow">
                Logo
              </span>
            </div>
          )}
        </div>

        {/* Position coordinates pill badge */}
        <div className="pointer-events-none absolute -top-6 left-1/2 -translate-x-1/2 whitespace-nowrap rounded-full bg-slate-900/90 px-2 py-0.5 text-[9px] font-mono font-bold text-white shadow border border-white/10 backdrop-blur-sm">
          {isTrulyCentered ? "🎯 50% Center" : `X: ${(x * 100).toFixed(0)}%`} • Y: {(y * 100).toFixed(0)}% • W: {(w * 100).toFixed(0)}%
        </div>

        {/* Resize Handle - Bottom Right Corner */}
        <div
          onPointerDown={handlePointerDownResize}
          onPointerMove={handlePointerMoveResize}
          onPointerUp={handlePointerUpResize}
          onPointerCancel={handlePointerUpResize}
          className="absolute -bottom-2 -right-2 flex h-5 w-5 cursor-nwse-resize items-center justify-center rounded-full border-2 border-white bg-brand-primary text-white shadow hover:scale-125 transition-transform"
          title="Drag to resize logo"
        >
          <div className="h-1.5 w-1.5 rounded-full bg-white" />
        </div>

        {/* Floating Quick Action Bar */}
        <div
          className="absolute -bottom-9 left-1/2 -translate-x-1/2 flex items-center gap-1.5 pointer-events-auto"
          onPointerDown={(e) => e.stopPropagation()}
        >
          {!isTrulyCentered && (
            <button
              type="button"
              onClick={onSnapCenter}
              disabled={disabled || isCompositing}
              className="flex items-center gap-1 rounded-full bg-slate-900/90 px-2 py-0.5 text-[10px] font-semibold text-brand-accent shadow border border-white/15 hover:bg-slate-800 transition-colors"
              title="Snap to 50% horizontal center"
            >
              <Target className="h-3 w-3" />
              Center
            </button>
          )}

          <button
            type="button"
            onClick={onApply}
            disabled={disabled || isCompositing}
            className="flex items-center gap-1 rounded-full bg-gradient-to-r from-emerald-500 to-teal-600 px-2.5 py-0.5 text-[10px] font-bold text-white shadow-lg hover:brightness-110 active:scale-95 transition-all"
            title="Bake logo into poster at this position"
          >
            {isCompositing ? (
              <RefreshCw className="h-3 w-3 animate-spin" />
            ) : (
              <Check className="h-3 w-3" />
            )}
            Apply
          </button>
        </div>
      </div>
    </>
  );
}
