import { useState } from "react";
import { Check, Copy } from "lucide-react";

// Prompt text must be copyable in one click (spec §12) — users paste it into
// other tools.
export function PromptCopyButton({ text }: { text: string }) {
  const [copied, setCopied] = useState(false);

  async function copy() {
    await navigator.clipboard.writeText(text);
    setCopied(true);
    setTimeout(() => setCopied(false), 1500);
  }

  return (
    <button
      type="button"
      onClick={copy}
      aria-label={copied ? "Copied" : "Copy"}
      className="shrink-0 rounded-lg border border-neutral-300 bg-white p-1.5 text-neutral-600 hover:border-brand-primary hover:text-brand-primary"
    >
      {copied ? <Check className="h-3.5 w-3.5" aria-hidden="true" /> : <Copy className="h-3.5 w-3.5" aria-hidden="true" />}
    </button>
  );
}
