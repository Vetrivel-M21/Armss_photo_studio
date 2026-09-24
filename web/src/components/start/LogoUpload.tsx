import { useRef, useState } from "react";
import { Upload, X } from "lucide-react";
import { uploadsApi } from "../../api/uploads";
import { Button } from "../common/Button";
import { ApiError } from "../../api/client";

const MAX_BYTES = 10 * 1024 * 1024;
const ALLOWED_TYPES = ["image/png", "image/jpeg", "image/webp"];

interface Attachment {
  previewUrl: string;
  uploadId?: string;
  status: "uploading" | "done" | "error";
  error?: string;
}

interface Props {
  onChange: (uploadId: string | undefined) => void;
}

// Single-file logo upload, distinct from FileAttach (which feeds style
// analysis, not compositing). This file replaces the global default logo on
// just this poster, so any organization — not only ARMSS — can use the tool.
export function LogoUpload({ onChange }: Props) {
  const [attachment, setAttachment] = useState<Attachment | null>(null);
  const inputRef = useRef<HTMLInputElement>(null);

  async function handleFile(file: File) {
    if (attachment?.previewUrl) URL.revokeObjectURL(attachment.previewUrl);

    if (!ALLOWED_TYPES.includes(file.type)) {
      setAttachment({ previewUrl: "", status: "error", error: "Only PNG, JPEG, or WebP images are accepted." });
      onChange(undefined);
      return;
    }
    if (file.size > MAX_BYTES) {
      setAttachment({ previewUrl: "", status: "error", error: "This file is over 10MB." });
      onChange(undefined);
      return;
    }

    const previewUrl = URL.createObjectURL(file);
    setAttachment({ previewUrl, status: "uploading" });
    onChange(undefined);

    try {
      const uploaded = await uploadsApi.create(file);
      setAttachment({ previewUrl, uploadId: uploaded.id, status: "done" });
      onChange(uploaded.id);
    } catch (err) {
      setAttachment({ previewUrl, status: "error", error: err instanceof ApiError ? err.message : "Upload failed." });
      onChange(undefined);
    }
  }

  function remove() {
    if (attachment?.previewUrl) URL.revokeObjectURL(attachment.previewUrl);
    setAttachment(null);
    onChange(undefined);
  }

  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between">
        <label className="text-sm font-medium text-neutral-700">Your logo (optional)</label>
        {!attachment && (
          <Button type="button" variant="ghost" onClick={() => inputRef.current?.click()} className="px-2 py-1 text-sm">
            <Upload className="h-4 w-4" aria-hidden="true" />
            Add logo
          </Button>
        )}
      </div>
      <p className="text-xs text-neutral-500">Replaces the default logo on this poster only.</p>
      <input
        ref={inputRef}
        type="file"
        accept="image/png,image/jpeg,image/webp"
        className="hidden"
        onChange={(e) => {
          const file = e.target.files?.[0];
          if (file) void handleFile(file);
          e.target.value = "";
        }}
      />
      {attachment && (
        <div className="relative inline-block h-20 w-20 overflow-hidden rounded border border-neutral-200 bg-neutral-50">
          {attachment.previewUrl && <img src={attachment.previewUrl} alt="" className="h-full w-full object-contain" />}
          {attachment.status === "uploading" && (
            <span className="absolute inset-0 flex items-center justify-center bg-white/70 text-[10px] text-neutral-600">
              Uploading...
            </span>
          )}
          {attachment.status === "error" && (
            <span className="absolute inset-0 flex items-center justify-center bg-danger/10 p-1 text-center text-[9px] text-danger">
              {attachment.error}
            </span>
          )}
          <Button
            type="button"
            variant="ghost"
            iconOnly
            onClick={remove}
            aria-label="Remove logo"
            className="absolute right-0.5 top-0.5 h-5 w-5 rounded-full bg-black/60 p-0 text-white hover:bg-black/80 hover:text-white"
          >
            <X className="h-3 w-3" aria-hidden="true" />
          </Button>
        </div>
      )}
    </div>
  );
}
