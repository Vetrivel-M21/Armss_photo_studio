import { api } from "./client";
import type { RecommendationView } from "./types";

export const recommendationsApi = {
  recommend: (text: string) => api.post<RecommendationView[]>("/recommendations", { text }),
};
