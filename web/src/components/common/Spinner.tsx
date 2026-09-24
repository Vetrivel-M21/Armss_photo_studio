export function Spinner({ label }: { label?: string }) {
  return (
    <span className="inline-flex items-center gap-2 text-sm text-neutral-500">
      <span
        className="h-4 w-4 animate-spin rounded-full border-2 border-brand-primary/30 border-t-brand-primary motion-reduce:animate-none"
        role="status"
        aria-label={label ?? "Loading"}
      />
      {label}
    </span>
  );
}
