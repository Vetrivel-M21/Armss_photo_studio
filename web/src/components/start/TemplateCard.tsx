import { motion } from "framer-motion";
import type { Template } from "../../api/types";
import { resolveCategoryStyle } from "../../lib/templateCategoryStyle";
import { useMotionSafe } from "../../hooks/useMotionSafe";
import { springPremium, springSnappy } from "../../lib/motion";

interface Props {
  template: Template;
  selected: boolean;
  onSelect: () => void;
  reason?: string;
}

// Deterministic per-card aspect ratio for the masonry gallery — Pinterest's
// look comes from tiles of varying height packing tightly into columns, not
// from uniform same-height rows. Real uploaded photos keep their natural crop
// (aspect-[4/5]); only the no-image gradient fallback tiles vary, since those
// are the ones with no real content dictating a "correct" shape.
const FALLBACK_ASPECTS = ["aspect-[4/5]", "aspect-square", "aspect-[3/4]", "aspect-[5/6]"];

function fallbackAspectFor(id: string): string {
  const sum = Array.from(id).reduce((acc, ch) => acc + ch.charCodeAt(0), 0);
  return FALLBACK_ASPECTS[sum % FALLBACK_ASPECTS.length];
}

export function TemplateCard({ template, selected, onSelect, reason }: Props) {
  const imageSrc = template.thumbUrl || template.imageUrl;
  const motionSafe = useMotionSafe();

  return (
    <motion.button
      type="button"
      onClick={onSelect}
      aria-pressed={selected}
      whileHover={motionSafe ? { y: -5, scale: 1.02, transition: springPremium } : undefined}
      whileTap={motionSafe ? { scale: 0.97, transition: springSnappy } : undefined}
      className={`group relative block w-full overflow-hidden rounded-2xl text-left transition-all duration-300 ${
        selected
          ? "ring-2 ring-brand-primary ring-offset-2 ring-offset-[#090B10] shadow-glow"
          : "border border-white/10 hover:border-white/30 hover:shadow-glow-sm"
      }`}
    >
      {imageSrc ? (
        <div className="relative aspect-[4/5] w-full overflow-hidden bg-[#111420]">
          <img
            src={imageSrc}
            alt={template.title}
            className="h-full w-full object-cover transition-transform duration-500 ease-out group-hover:scale-108 motion-reduce:transition-none motion-reduce:group-hover:scale-100"
            loading="lazy"
          />
          <div className="absolute inset-x-0 bottom-0 bg-gradient-to-t from-[#090B10]/95 via-[#090B10]/60 to-transparent p-3 pt-8">
            {template.occasionTags[0] && (
              <span className="mb-1 inline-block rounded-full bg-white/15 px-2 py-0.5 text-[9px] font-bold uppercase tracking-wider text-white backdrop-blur-md">
                {template.occasionTags[0]}
              </span>
            )}
            <p className="truncate text-xs font-bold text-white tracking-tight">{template.title}</p>
          </div>
        </div>
      ) : (
        (() => {
          const { icon: CategoryIcon, gradientClass } = resolveCategoryStyle(template);
          return (
            <div className={`relative flex w-full items-end ${fallbackAspectFor(template.id)} ${gradientClass} overflow-hidden`}>
              <div className="absolute inset-0 bg-black/25 backdrop-blur-[2px]" />
              <CategoryIcon
                className="absolute left-1/2 top-1/2 h-14 w-14 -translate-x-1/2 -translate-y-1/2 text-white/30 transition-transform duration-500 group-hover:scale-110"
                aria-hidden="true"
              />
              <div className="relative z-10 flex w-full flex-col gap-1 bg-gradient-to-t from-[#090B10]/95 via-[#090B10]/60 to-transparent p-3 pt-8">
                {template.occasionTags[0] && (
                  <span className="inline-flex w-fit items-center gap-1 rounded-full bg-white/20 px-2 py-0.5 text-[9px] font-bold uppercase tracking-wider text-white/90 backdrop-blur-md">
                    {template.occasionTags[0]}
                  </span>
                )}
                <p className="truncate text-xs font-bold text-white tracking-tight">{template.title}</p>
              </div>
            </div>
          );
        })()
      )}
      {selected && (
        <span className="absolute right-2.5 top-2.5 flex items-center gap-1 rounded-full bg-gradient-to-r from-brand-primary to-brand-secondary px-2.5 py-0.5 text-[10px] font-bold text-white shadow-glow-sm">
          ✓ Selected
        </span>
      )}
      {reason && (
        <div className="border-t border-white/10 bg-[#111420]/90 px-3 py-2 backdrop-blur-md">
          <p className="text-[11px] text-slate-300 font-medium">{reason}</p>
        </div>
      )}
    </motion.button>
  );
}
