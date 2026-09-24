interface Props {
  options: [string, string];
  value: 0 | 1;
  onChange: (value: 0 | 1) => void;
}

export function Toggle({ options, value, onChange }: Props) {
  return (
    <div className="inline-flex rounded-xl border border-white/10 bg-black/40 p-1 text-xs backdrop-blur-md shadow-inner">
      {options.map((label, i) => (
        <button
          key={label}
          type="button"
          aria-pressed={value === i}
          onClick={() => onChange(i as 0 | 1)}
          className={`rounded-lg px-3.5 py-1.5 font-medium transition-all duration-200 motion-reduce:transition-none ${
            value === i
              ? "bg-gradient-to-r from-brand-primary to-brand-secondary text-white font-semibold shadow-glow-sm"
              : "text-slate-400 hover:text-white hover:bg-white/[0.04]"
          }`}
        >
          {label}
        </button>
      ))}
    </div>
  );
}
