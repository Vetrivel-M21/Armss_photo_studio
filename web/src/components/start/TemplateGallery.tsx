import { useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { motion } from "framer-motion";
import { LayoutGrid } from "lucide-react";
import { templatesApi } from "../../api/templates";
import { TemplateCard } from "./TemplateCard";
import { CategoryFilterChips } from "./CategoryFilterChips";
import { Spinner } from "../common/Spinner";
import { useMotionSafe } from "../../hooks/useMotionSafe";
import { galleryItem, galleryStagger } from "../../lib/motion";
import { resolveCategoryStyle, type CategoryStyle } from "../../lib/templateCategoryStyle";

interface Props {
  selectedId?: string;
  onSelect: (id: string | undefined) => void;
}

// The primary path (Flow A, spec §3): a gallery of pre-uploaded ARMSS
// designs the user can pick directly, no typing required beyond the occasion.
export function TemplateGallery({ selectedId, onSelect }: Props) {
  const motionSafe = useMotionSafe();
  const [activeCategory, setActiveCategory] = useState<string | null>(null);
  const { data: templates, isLoading, isError } = useQuery({
    queryKey: ["templates"],
    queryFn: () => templatesApi.list(),
  });

  const categories = useMemo(() => {
    if (!templates) return [];
    const map = new Map<string, CategoryStyle>();
    for (const t of templates) {
      const style = resolveCategoryStyle(t);
      if (!map.has(style.category)) map.set(style.category, style);
    }
    return Array.from(map.values()).sort((a, b) => a.category.localeCompare(b.category));
  }, [templates]);

  const filtered = useMemo(() => {
    if (!templates) return [];
    return activeCategory ? templates.filter((t) => resolveCategoryStyle(t).category === activeCategory) : templates;
  }, [templates, activeCategory]);

  if (isLoading) return <Spinner label="Loading templates..." />;
  if (isError) return <p className="text-sm text-neutral-500">Could not load the template library right now.</p>;
  if (!templates || templates.length === 0) {
    return <p className="text-sm text-neutral-500">No templates yet — an admin can add some, or describe what you need below.</p>;
  }

  return (
    <div className="space-y-5">
      <div className="flex flex-wrap items-center justify-between gap-3 border-b border-white/10 pb-4">
        <div className="flex items-center gap-2.5">
          <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-brand-primary/20 text-brand-primary border border-brand-primary/30">
            <LayoutGrid className="h-4 w-4" aria-hidden="true" />
          </div>
          <div>
            <h2 className="text-lg font-bold tracking-tight text-white">Poster Template Gallery</h2>
            <p className="text-xs text-slate-400">Curated house designs &amp; aesthetic composition foundations</p>
          </div>
        </div>
        <span className="rounded-full bg-white/[0.06] border border-white/10 px-3 py-1 text-xs font-semibold text-slate-300">
          {activeCategory ? `${filtered.length} of ${templates.length}` : templates.length} templates
        </span>
      </div>

      <CategoryFilterChips categories={categories} active={activeCategory} onChange={setActiveCategory} />

      {filtered.length === 0 ? (
        <div className="rounded-2xl border border-dashed border-white/10 bg-[#111420]/60 p-8 text-center backdrop-blur-md">
          <p className="text-sm text-slate-400">
            No templates found in &quot;{activeCategory}&quot;.{" "}
            <button
              type="button"
              onClick={() => setActiveCategory(null)}
              className="font-semibold text-brand-accent hover:underline underline-offset-2"
            >
              Reset filter
            </button>
          </p>
        </div>
      ) : (
        <motion.div
          key={activeCategory ?? "all"}
          className="columns-2 gap-4 sm:columns-3 md:columns-4 lg:columns-5"
          variants={galleryStagger}
          initial={motionSafe ? "hidden" : "visible"}
          whileInView="visible"
          viewport={{ once: true, amount: 0.2 }}
        >
          {filtered.map((t) => (
            <motion.div key={t.id} variants={galleryItem} className="mb-4 break-inside-avoid">
              <TemplateCard
                template={t}
                selected={selectedId === t.id}
                onSelect={() => onSelect(selectedId === t.id ? undefined : t.id)}
              />
            </motion.div>
          ))}
        </motion.div>
      )}
    </div>
  );
}
