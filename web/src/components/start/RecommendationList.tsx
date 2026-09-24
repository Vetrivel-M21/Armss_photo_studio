import type { RecommendationView } from "../../api/types";

interface Props {
  recommendations: RecommendationView[];
  selectedId?: string;
  onSelect: (id: string) => void;
}

// Flow B (spec §3): shown inline under the text area once the user has typed
// enough, without blocking them from continuing without picking one.
export function RecommendationList({ recommendations, selectedId, onSelect }: Props) {
  if (recommendations.length === 0) return null;

  return (
    <div className="space-y-2.5 rounded-2xl border border-brand-highlight/30 bg-brand-highlight/5 p-4 backdrop-blur-md">
      <div className="flex items-center gap-2">
        <span className="h-2 w-2 rounded-full bg-brand-highlight animate-ping" />
        <p className="text-xs font-bold uppercase tracking-wider text-brand-highlight">AI Recommended Templates</p>
      </div>
      <div className="grid gap-2.5 sm:grid-cols-3">
        {recommendations.map((r) => (
          <button
            key={r.template.id}
            type="button"
            onClick={() => onSelect(r.template.id)}
            className={`flex items-center gap-3 rounded-xl border p-2.5 text-left transition-all duration-200 motion-reduce:transition-none ${
              selectedId === r.template.id
                ? "border-brand-primary bg-brand-primary/15 shadow-glow-sm"
                : "border-white/10 bg-[#0E121E]/80 hover:border-white/20 hover:bg-[#141928]"
            }`}
          >
            <img src={r.template.thumbUrl || r.template.imageUrl} alt="" className="h-14 w-11 shrink-0 rounded-lg object-cover" />
            <span className="min-w-0">
              <span className="block truncate text-xs font-bold text-white">{r.template.title}</span>
              <span className="mt-0.5 block text-[11px] text-slate-400 line-clamp-2">{r.reason}</span>
            </span>
          </button>
        ))}
      </div>
    </div>
  );
}
