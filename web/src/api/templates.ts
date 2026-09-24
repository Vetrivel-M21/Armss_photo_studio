import { api } from "./client";
import type { Template } from "./types";

export const templatesApi = {
  list: (occasion?: string) => api.get<Template[]>(`/templates${occasion ? `?occasion=${encodeURIComponent(occasion)}` : ""}`),
  get: (id: string) => api.get<Template>(`/templates/${id}`),
  create: (form: FormData) => api.postForm<Template>("/templates", form),
  patch: (
    id: string,
    body: { styleProfile?: unknown; occasionTags?: string[]; season?: string; isActive?: boolean },
  ) => api.patch<Template>(`/templates/${id}`, body),
  importStarterLibrary: () => api.post<{ imported: number; skipped: number }>("/templates/import-starter-library"),
};
