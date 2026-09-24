import { AlertCircle } from "lucide-react";
import { Button } from "./Button";

interface Props {
  message: string;
  onRetry?: () => void;
}

// Every failed stage shows this, never a stuck spinner (spec §12/acceptance
// criterion 6): a clear message plus a retry action.
export function ErrorBanner({ message, onRetry }: Props) {
  return (
    <div className="flex items-center justify-between gap-3 rounded border border-danger/30 bg-danger/5 px-4 py-3 text-sm text-danger">
      <span className="flex items-center gap-2">
        <AlertCircle className="h-4 w-4 shrink-0" aria-hidden="true" />
        {message}
      </span>
      {onRetry && (
        <Button variant="danger" onClick={onRetry} className="shrink-0">
          Retry
        </Button>
      )}
    </div>
  );
}
