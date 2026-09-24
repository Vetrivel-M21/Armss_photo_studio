import { api } from "./client";

export interface AgentSettings {
  provider: "anthropic" | "gemini";
  model: string;
  anthropicConfigured: boolean;
  geminiConfigured: boolean;
  openaiConfigured: boolean;
  fluxConfigured: boolean;
  ideogramConfigured: boolean;
  imageProvider: "flux" | "ideogram" | "openai";
}

export type KeyProvider = "anthropic" | "gemini" | "openai" | "flux" | "ideogram";

export const agentSettingsApi = {
  get: () => api.get<AgentSettings>("/agent-settings"),
  put: (provider?: "anthropic" | "gemini", model?: string, imageProvider?: "flux" | "ideogram" | "openai") =>
    api.put<AgentSettings>("/agent-settings", { provider, model, imageProvider }),
  setImageProvider: (imageProvider: "flux" | "ideogram" | "openai") =>
    api.put<AgentSettings>("/agent-settings", { imageProvider }),
  setKey: (provider: KeyProvider, apiKey: string) =>
    api.put<AgentSettings>("/agent-settings/keys", { provider, apiKey }),
};
