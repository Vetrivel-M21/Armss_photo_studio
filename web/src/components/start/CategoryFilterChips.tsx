import type { CategoryStyle } from "../../lib/templateCategoryStyle";
import { cn } from "../../lib/cn";

interface Props {
  categories: CategoryStyle[];
  active: string | null;
  onChange: (category: string | null) => void;
}

export function CategoryFilterChips({ categories, active, onChange }: Props) {
  return (
    <div className="-mx-1 flex gap-2 overflow-x-auto px-1 pb-2 scrollbar-none">
      <button
        type="button"
        onClick={() => onChange(null)}
        className={cn(
          "shrink-0 rounded-full px-4 py-1.5 text-xs font-medium transition-all duration-200 motion-reduce:transition-none",
          active === null
            ? "bg-gradient-to-r from-brand-primary via-brand-secondary to-purple-600 text-white font-semibold shadow-glow-sm border border-white/25"
            : "border border-white/10 bg-[#111420]/80 backdrop-blur-md text-slate-300 hover:bg-white/[0.08] hover:text-white hover:border-white/20",
        )}
      >
        All Categories
      </button>
      {categories.map((c) => {
        const isActive = active === c.category;
        return (
          <button
            key={c.category}
            type="button"
            onClick={() => onChange(c.category)}
            className={cn(
              "flex shrink-0 items-center gap-1.5 rounded-full px-3.5 py-1.5 text-xs font-medium transition-all duration-200 motion-reduce:transition-none",
              isActive
                ? "bg-gradient-to-r from-brand-primary via-brand-secondary to-purple-600 text-white font-semibold shadow-glow-sm border border-white/25"
                : "border border-white/10 bg-[#111420]/80 backdrop-blur-md text-slate-300 hover:bg-white/[0.08] hover:text-white hover:border-white/20",
            )}
          >
            <c.icon className="h-3.5 w-3.5" aria-hidden="true" />
            {c.category}
          </button>
        );
      })}
    </div>
  );
}
