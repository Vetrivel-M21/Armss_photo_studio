import { useRef, useState } from "react";
import { Plus, X } from "lucide-react";
import { uploadsApi } from "../../api/uploads";
import { Button } from "../common/Button";
import { ApiError } from "../../api/client";

const MAX_FILES = 4;
const MAX_BYTES = 10 * 1024 * 1024;
const ALLOWED_TYPES = ["image/png", "image/jpeg", "image/webp"];

interface Attachment {
  key: string;
  file: File;
  previewUrl: string;
  uploadId?: string;
  status: "uploading" | "done" | "error";
  error?: string;
}

interface Props {
  onChange: (uploadIds: string[]) => void;
}

// Flow C (spec §3): attach 1-4 of your own reference images, analysed on the
// fly exactly like a library template.
export function FileAttach({ onChange }: Props) {
  const [attachments, setAttachments] = useState<Attachment[]>([]);
  const inputRef = useRef<HTMLInputElement>(null);

  function emitChange(list: Attachment[]) {
    onChange(list.filter((a) => a.status === "done" && a.uploadId).map((a) => a.uploadId!));
  }

  async function handleFiles(files: FileList) {
    const room = MAX_FILES - attachments.length;
    const picked = Array.from(files).slice(0, room);

    for (const file of picked) {
      if (!ALLOWED_TYPES.includes(file.type)) {
        setAttachments((prev) => {
          const next = [...prev, { key: crypto.randomUUID(), file, previewUrl: "", status: "error" as const, error: "Only PNG, JPEG, or WebP images are accepted." }];
          emitChange(next);
          return next;
        });
        continue;
      }
      if (file.size > MAX_BYTES) {
        setAttachments((prev) => {
          const next = [...prev, { key: crypto.randomUUID(), file, previewUrl: "", status: "error" as const, error: "This file is over 10MB." }];
          emitChange(next);
          return next;
        });
        continue;
      }

      const key = crypto.randomUUID();
      const previewUrl = URL.createObjectURL(file);
      setAttachments((prev) => {
        const next = [...prev, { key, file, previewUrl, status: "uploading" as const }];
        emitChange(next);
        return next;
      });

      try {
        const uploaded = await uploadsApi.create(file);
        setAttachments((prev) => {
          const next = prev.map((a) => (a.key === key ? { ...a, uploadId: uploaded.id, status: "done" as const } : a));
          emitChange(next);
          return next;
        });
      } catch (err) {
        setAttachments((prev) => {
          const next = prev.map((a) =>
            a.key === key ? { ...a, status: "error" as const, error: err instanceof ApiError ? err.message : "Upload failed." } : a,
          );
          emitChange(next);
          return next;
        });
      }
    }
  }

  function remove(key: string) {
    setAttachments((prev) => {
      const next = prev.filter((a) => a.key !== key);
      emitChange(next);
      return next;
    });
  }

  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between">
        <label className="text-sm font-medium text-neutral-700">Attach your own reference (optional)</label>
        {attachments.length < MAX_FILES && (
          <Button type="button" variant="ghost" onClick={() => inputRef.current?.click()} className="px-2 py-1 text-sm">
            <Plus className="h-4 w-4" aria-hidden="true" />
            Add image
          </Button>
        )}
      </div>
      <input
        ref={inputRef}
        type="file"
        accept="image/png,image/jpeg,image/webp"
        multiple
        className="hidden"
        onChange={(e) => {
          if (e.target.files) void handleFiles(e.target.files);
          e.target.value = "";
        }}
      />
      {attachments.length > 0 && (
        <ul className="flex flex-wrap gap-2">
          {attachments.map((a) => (
            <li key={a.key} className="relative h-20 w-20 overflow-hidden rounded border border-neutral-200">
              {a.previewUrl && <img src={a.previewUrl} alt="" className="h-full w-full object-cover" />}
              {a.status === "uploading" && (
                <span className="absolute inset-0 flex items-center justify-center bg-white/70 text-[10px] text-neutral-600">
                  Uploading...
                </span>
              )}
              {a.status === "error" && (
                <span className="absolute inset-0 flex items-center justify-center bg-danger/10 p-1 text-center text-[9px] text-danger">
                  {a.error}
                </span>
              )}
              <Button
                type="button"
                variant="ghost"
                iconOnly
                onClick={() => remove(a.key)}
                aria-label="Remove"
                className="absolute right-0.5 top-0.5 h-5 w-5 rounded-full bg-black/60 p-0 text-white hover:bg-black/80 hover:text-white"
              >
                <X className="h-3 w-3" aria-hidden="true" />
              </Button>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
