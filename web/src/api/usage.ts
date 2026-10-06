import { api } from "./client";

export interface UserLimit {
  maxApiCallsPerHour: number;
  maxRendersPerHour: number;
  monthlySpendCapCents: number;
  isUnlimited: boolean;
  isCustom: boolean;
}

export interface UserUsage {
  userId: string;
  email: string;
  role: string;
  createdAt: string;
  totalProjects: number;
  totalTokensIn: number;
  totalTokensOut: number;
  totalRenders: number;
  totalCostCents: number;
  hourApiCalls: number;
  hourRenders: number;
  limits: UserLimit;
}

export interface UsageOverview {
  totalProjects: number;
  totalTokensIn: number;
  totalTokensOut: number;
  totalRenders: number;
  totalCostCents: number;
  globalLimit: UserLimit;
  users: UserUsage[];
}

export interface UpdateLimitsPayload {
  maxApiCallsPerHour: number;
  maxRendersPerHour: number;
  monthlySpendCapCents: number;
  isUnlimited?: boolean;
}

export const usageApi = {
  getOverview: () => api.get<UsageOverview>("/admin/usage"),
  updateGlobalLimits: (payload: UpdateLimitsPayload) => api.put<{ message: string }>("/admin/limits", payload),
  updateUserLimits: (userId: string, payload: UpdateLimitsPayload) =>
    api.put<{ message: string }>(`/admin/users/${userId}/limits`, payload),
};
