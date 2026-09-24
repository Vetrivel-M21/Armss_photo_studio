import { useRef, useState } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { Upload, type LucideIcon } from "lucide-react";
import { brandAssetsApi } from "../../api/brandAssets";
import type { BrandAsset } from "../../api/types";
import { Button } from "../common/Button";
import { Card } from "../common/Card";
import { ErrorBanner } from "../common/ErrorBanner";
import { useToast } from "../common/Toast";
import { ApiError } from "../../api/client";

interface Props {
  kind: "logo" | "tagline" | "footer";
  label: string;
  hint: string;
  icon: LucideIcon;
  current?: BrandAsset;
}

// Replacing a brand asset always adds a new active version (spec §10) and
// invalidates the compositor's cache immediately server-side.
export function BrandAssetUploader({ kind, label, hint, icon: Icon, current }: Props) {
  const fileRef = useRef<HTMLInputElement>(null);
  const [fileName, setFileName] = useState<string | null>(null);
  const queryClient = useQueryClient();
  const { toast } = useToast();

  const put = useMutation({
    mutationFn: (file: File) => brandAssetsApi.put(kind, file),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["brand-assets"] });
      if (fileRef.current) fileRef.current.value = "";
      setFileName(null);
      toast({ title: `${label} updated`, variant: "success" });
    },
  });

  return (
    <Card>
      <div className="flex items-start justify-between gap-4">
        <div className="flex items-start gap-2.5">
          <span className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-brand-primary/10 text-brand-primary">
            <Icon className="h-4 w-4" aria-hidden="true" />
          </span>
          <div>
            <h3 className="text-sm font-semibold text-neutral-800">{label}</h3>
            <p className="text-xs text-neutral-500">{hint}</p>
            {current && <p className="mt-1 text-xs text-neutral-400">Current version: v{current.version}</p>}
          </div>
        </div>
        {current && (
          <img src={current.imageUrl} alt={`Current ${label}`} className="h-16 w-16 rounded border border-neutral-100 object-contain bg-neutral-50" />
        )}
      </div>

      {put.isError && (
        <ErrorBanner message={put.error instanceof ApiError ? put.error.message : "Could not upload this asset."} />
      )}

      <div className="mt-3 space-y-2">
        <input
          ref={fileRef}
          type="file"
          accept="image/png"
          className="hidden"
          onChange={(e) => setFileName(e.target.files?.[0]?.name ?? null)}
        />
        <button
          type="button"
          onClick={() => fileRef.current?.click()}
          className="flex w-full flex-col items-center justify-center gap-1.5 rounded-xl border-2 border-dashed border-neutral-300 p-4 text-center hover:border-brand-primary"
        >
          <Upload className="h-5 w-5 text-neutral-400" aria-hidden="true" />
          <span className="text-xs text-neutral-500">{fileName ?? "Click to choose a PNG file"}</span>
        </button>
        <Button
          variant="secondary"
          className="w-full"
          disabled={put.isPending || !fileName}
          onClick={() => {
            const file = fileRef.current?.files?.[0];
            if (file) put.mutate(file);
          }}
        >
          {put.isPending ? "Uploading..." : "Replace"}
        </Button>
      </div>
    </Card>
  );
}
