import { api, newIdempotencyKey } from "./client";
import type { ArtDirection, Copy, ProjectDetail, ProjectSummary, ZoneOverride } from "./types";

export const projectsApi = {
  create: (body: {
    text: string;
    templateId?: string;
    uploadIds?: string[];
    logoUploadId?: string;
    footerText?: string;
  }) => api.post<ProjectDetail>("/projects", body),
  list: () => api.get<ProjectSummary[]>("/projects"),
  get: (id: string) => api.get<ProjectDetail>(`/projects/${id}`),
  downloadUrl: (id: string, kind: "raw" | "composited" = "composited") => `/api/v1/projects/${id}/download?kind=${kind}`,

  rerunBrief: (id: string, extraDetail: string) =>
    api.post<ProjectDetail>(`/projects/${id}/brief/rerun`, { extraDetail }),
  selectConcept: (id: string, index: number) =>
    api.post<ProjectDetail>(`/projects/${id}/concepts/${index}/select`),
  patchCopy: (id: string, copy: Copy) => api.patch<ProjectDetail>(`/projects/${id}/copy`, copy),
  patchArtDirection: (id: string, art: ArtDirection) =>
    api.patch<ProjectDetail>(`/projects/${id}/art-direction`, art),
  rerunArtDirection: (id: string, excludedObjects: string[]) =>
    api.post<ProjectDetail>(`/projects/${id}/art-direction/rerun`, { excludedObjects }),

  // Render and composite cost real money per click, so each call uses an
  // idempotency key — a retried request replays the stored result instead of
  // generating (or compositing) a second time (spec §13).
  render: (id: string, provider?: string, aspectRatio?: string, idempotencyKey?: string) =>
    api.post<ProjectDetail>(
      `/projects/${id}/render`,
      provider || aspectRatio ? { ...(provider ? { provider } : {}), ...(aspectRatio ? { aspectRatio } : {}) } : {},
      {
        "Idempotency-Key": idempotencyKey || newIdempotencyKey(),
      }
    ),
  composite: (id: string, zoneOverrides?: Record<string, ZoneOverride>, idempotencyKey?: string) =>
    api.post<ProjectDetail>(`/projects/${id}/composite`, zoneOverrides ? { zoneOverrides } : {}, {
      "Idempotency-Key": idempotencyKey || newIdempotencyKey(),
    }),
};
