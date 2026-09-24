import { TextArea } from "../common/TextArea";

interface Props {
  value: string;
  onChange: (value: string) => void;
}

export function OccasionTextArea({ value, onChange }: Props) {
  const quickIdeas = ["Diwali 2026 Celebration", "Grand Opening & Hiring Drive", "Pongal Harvest Festival", "Tech Summit 2026", "Campus Placement Week"];

  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between">
        <label htmlFor="occasion" className="text-xs font-bold uppercase tracking-wider text-slate-300">
          Occasion &amp; Theme Prompt
        </label>
        <span className="text-[11px] text-slate-400">Natural language brief</span>
      </div>
      <TextArea
        id="occasion"
        rows={3}
        placeholder='e.g. "Diwali 2026 celebration with festive diyas and prosperity vibes" or "Campus hiring announcement for tech graduates"'
        value={value}
        onChange={(e) => onChange(e.target.value)}
        className="text-base leading-relaxed bg-[#0A0D16]/90 border-white/15 focus:border-brand-primary focus:ring-brand-primary/30"
      />
      <div className="flex flex-wrap items-center gap-1.5 pt-1">
        <span className="text-[10px] font-semibold uppercase tracking-wider text-slate-500 mr-1">Quick ideas:</span>
        {quickIdeas.map((idea) => (
          <button
            key={idea}
            type="button"
            onClick={() => onChange(idea)}
            className="rounded-full border border-white/10 bg-white/[0.04] px-2.5 py-0.5 text-[11px] text-slate-300 transition-all hover:border-brand-primary/40 hover:bg-brand-primary/10 hover:text-white"
          >
            + {idea}
          </button>
        ))}
      </div>
    </div>
  );
}
