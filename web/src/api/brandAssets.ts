import { api } from "./client";
import type { BrandAsset } from "./types";

export const brandAssetsApi = {
  list: () => api.get<BrandAsset[]>("/brand-assets"),
  put: (kind: "logo" | "tagline" | "footer", file: File) => {
    const form = new FormData();
    form.append("kind", kind);
    form.append("file", file);
    return api.putForm<BrandAsset>("/brand-assets", form);
  },
};
