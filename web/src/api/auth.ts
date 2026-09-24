import { api } from "./client";
import type { User } from "./types";

export const authApi = {
  login: (email: string, password: string) => api.post<User>("/auth/login", { email, password }),
  logout: () => api.post<{ status: string }>("/auth/logout"),
  me: () => api.get<User>("/auth/me"),
};
