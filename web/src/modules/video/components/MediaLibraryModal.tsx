import React, { useState, useRef } from "react";
import { Upload, X, Image as ImageIcon, Loader2, Sparkles } from "lucide-react";
import { uploadsApi } from "../../../api/uploads";

interface MediaLibraryModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSelectImage: (imageKey: string) => void;
}

// Preset starter photos available for immediate test / demo creation
const PRESET_PHOTOS = [
  {
    key: "video/presets/studio_sample_1.jpg",
    name: "Urban Fashion",
    url: "https://images.unsplash.com/photo-1515886657613-9f3515b0c78f?auto=format&fit=crop&w=800&q=80",
  },
  {
    key: "video/presets/studio_sample_2.jpg",
    name: "Portrait Glow",
    url: "https://images.unsplash.com/photo-1534528741775-53994a69daeb?auto=format&fit=crop&w=800&q=80",
  },
  {
    key: "video/presets/studio_sample_3.jpg",
    name: "Editorial Model",
    url: "https://images.unsplash.com/photo-1529139574466-a303027c1d8b?auto=format&fit=crop&w=800&q=80",
  },
  {
    key: "video/presets/studio_sample_4.jpg",
    name: "Golden Hour",
    url: "https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?auto=format&fit=crop&w=800&q=80",
  },
  {
    key: "video/presets/studio_sample_5.jpg",
    name: "Modern Street",
    url: "https://images.unsplash.com/photo-1524504388940-b1c1722653e1?auto=format&fit=crop&w=800&q=80",
  },
  {
    key: "video/presets/studio_sample_6.jpg",
    name: "Moody Aesthetic",
    url: "https://images.unsplash.com/photo-1492562080023-ab3db95bfbce?auto=format&fit=crop&w=800&q=80",
  },
];

export const MediaLibraryModal: React.FC<MediaLibraryModalProps> = ({
  isOpen,
  onClose,
  onSelectImage,
}) => {
  const [activeTab, setActiveTab] = useState<"upload" | "presets">("upload");
  const [isUploading, setIsUploading] = useState(false);
  const [uploadError, setUploadError] = useState<string | null>(null);
  const fileInputRef = useRef<HTMLInputElement | null>(null);

  if (!isOpen) return null;

  const handleFileChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    setIsUploading(true);
    setUploadError(null);

    try {
      const res = await uploadsApi.create(file);
      // If objectKey returned, use it; otherwise use uploads/{res.id}
      const key = res.objectKey || `uploads/${res.id}`;
      onSelectImage(key);
      onClose();
    } catch (err) {
      setUploadError(err instanceof Error ? err.message : "Failed to upload image");
    } finally {
      setIsUploading(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 backdrop-blur-sm animate-fade-in">
      <div className="relative w-full max-w-2xl bg-slate-900 border border-white/10 rounded-2xl shadow-2xl overflow-hidden flex flex-col max-h-[85vh]">
        {/* Modal Header */}
        <div className="flex items-center justify-between px-6 py-4 border-b border-white/10">
          <div className="flex items-center gap-2">
            <ImageIcon className="w-5 h-5 text-brand-primary" />
            <h2 className="text-base font-bold text-white">Add Photo to Video</h2>
          </div>
          <button
            type="button"
            onClick={onClose}
            className="p-1 rounded-lg text-slate-400 hover:text-white hover:bg-white/10 transition-colors"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Tab Navigation */}
        <div className="flex px-6 pt-3 border-b border-white/10 gap-4">
          <button
            type="button"
            onClick={() => setActiveTab("upload")}
            className={`pb-2.5 text-xs font-semibold flex items-center gap-1.5 border-b-2 transition-all ${
              activeTab === "upload"
                ? "border-brand-primary text-white"
                : "border-transparent text-slate-400 hover:text-white"
            }`}
          >
            <Upload className="w-3.5 h-3.5" />
            Upload New Photo
          </button>
          <button
            type="button"
            onClick={() => setActiveTab("presets")}
            className={`pb-2.5 text-xs font-semibold flex items-center gap-1.5 border-b-2 transition-all ${
              activeTab === "presets"
                ? "border-brand-primary text-white"
                : "border-transparent text-slate-400 hover:text-white"
            }`}
          >
            <Sparkles className="w-3.5 h-3.5" />
            Sample Studio Library
          </button>
        </div>

        {/* Modal Content */}
        <div className="p-6 overflow-y-auto flex-1">
          {activeTab === "upload" && (
            <div className="flex flex-col items-center justify-center gap-4">
              <input
                ref={fileInputRef}
                type="file"
                accept="image/png,image/jpeg,image/webp"
                className="hidden"
                onChange={handleFileChange}
              />

              <div
                onClick={() => !isUploading && fileInputRef.current?.click()}
                className={`w-full h-56 border-2 border-dashed rounded-2xl flex flex-col items-center justify-center p-6 cursor-pointer transition-all ${
                  isUploading
                    ? "border-brand-primary/50 bg-brand-primary/5 cursor-wait"
                    : "border-white/20 hover:border-brand-primary/80 bg-white/[0.02] hover:bg-white/[0.05]"
                }`}
              >
                {isUploading ? (
                  <div className="flex flex-col items-center gap-3">
                    <Loader2 className="w-8 h-8 text-brand-primary animate-spin" />
                    <span className="text-sm font-medium text-slate-200">
                      Processing and optimizing photo...
                    </span>
                    <span className="text-xs text-slate-500">Auto-downscaling to 1500px</span>
                  </div>
                ) : (
                  <div className="flex flex-col items-center gap-2 text-center">
                    <div className="w-12 h-12 rounded-xl bg-brand-primary/20 flex items-center justify-center text-brand-primary mb-1">
                      <Upload className="w-6 h-6" />
                    </div>
                    <span className="text-sm font-semibold text-white">
                      Click to browse or drag & drop photo
                    </span>
                    <span className="text-xs text-slate-400">
                      Supports PNG, JPEG, WebP (up to 10MB)
                    </span>
                  </div>
                )}
              </div>

              {uploadError && (
                <div className="w-full p-3 rounded-xl bg-rose-500/10 border border-rose-500/20 text-xs text-rose-400">
                  {uploadError}
                </div>
              )}
            </div>
          )}

          {activeTab === "presets" && (
            <div className="flex flex-col gap-3">
              <p className="text-xs text-slate-400">
                Click any high-resolution portrait sample below to add it directly to your video timeline:
              </p>
              <div className="grid grid-cols-3 gap-3">
                {PRESET_PHOTOS.map((p) => (
                  <div
                    key={p.key}
                    onClick={() => {
                      onSelectImage(p.url);
                      onClose();
                    }}
                    className="group relative aspect-[3/4] rounded-xl overflow-hidden cursor-pointer border border-white/10 hover:border-brand-primary hover:shadow-lg hover:shadow-brand-primary/20 transition-all"
                  >
                    <img
                      src={p.url}
                      alt={p.name}
                      className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300"
                    />
                    <div className="absolute inset-0 bg-gradient-to-t from-black/80 via-transparent to-transparent opacity-80 group-hover:opacity-100 transition-opacity" />
                    <div className="absolute bottom-2 left-2 right-2 flex items-center justify-between text-white">
                      <span className="text-xs font-semibold truncate">{p.name}</span>
                      <span className="text-[10px] px-1.5 py-0.5 rounded bg-brand-primary/80 opacity-0 group-hover:opacity-100 transition-opacity">
                        Add
                      </span>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};
