import { api } from "./client";
import type { UploadResponse } from "./types";

export const uploadsApi = {
  create: (file: File) => {
    const form = new FormData();
    form.append("file", file);
    return api.postForm<UploadResponse>("/uploads", form);
  },
};
