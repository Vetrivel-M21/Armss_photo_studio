import type { VideoSpec } from "../engine/spec";

export interface VideoProject {
  id: string;
  user_id: string;
  title: string;
  spec: VideoSpec;
  spec_version: number;
  revision: number;
  created_at: string;
  updated_at: string;
}

export interface VideoJob {
  id: string;
  project_id?: string;
  user_id: string;
  spec: VideoSpec;
  resolution: "720p" | "1080p";
  status: "queued" | "processing" | "done" | "failed";
  progress: number;
  output_key?: string;
  error?: string;
  started_at?: string;
  finished_at?: string;
  created_at: string;
  updated_at: string;
}

export interface AudioPreset {
  id: string;
  title: string;
  artist?: string;
  s3_key: string;
  duration: number;
  license: string;
  source_url: string;
  attribution?: string;
  created_at: string;
}

const API_BASE = "/api/v1/video";

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      ...options?.headers,
    },
    credentials: "include",
  });

  if (!res.ok) {
    const errorBody = await res.json().catch(() => ({ error: "Request failed" }));
    throw new Error(errorBody.error || `HTTP ${res.status}`);
  }

  return res.json();
}

export const videoApi = {
  async listProjects(limit = 20, offset = 0): Promise<{ projects: VideoProject[]; limit: number; offset: number }> {
    return request<{ projects: VideoProject[]; limit: number; offset: number }>(`/projects?limit=${limit}&offset=${offset}`);
  },

  async getProject(id: string): Promise<VideoProject> {
    return request<VideoProject>(`/projects/${id}`);
  },

  async createProject(title: string, spec?: VideoSpec): Promise<VideoProject> {
    return request<VideoProject>("/projects", {
      method: "POST",
      body: JSON.stringify({ title, spec }),
    });
  },

  async updateProject(id: string, title: string, spec: VideoSpec, revision: number): Promise<VideoProject> {
    return request<VideoProject>(`/projects/${id}`, {
      method: "PUT",
      body: JSON.stringify({ title, spec, revision }),
    });
  },

  async deleteProject(id: string): Promise<{ status: string }> {
    return request<{ status: string }>(`/projects/${id}`, {
      method: "DELETE",
    });
  },

  async createJob(spec: VideoSpec, resolution: "720p" | "1080p" = "1080p", projectId?: string): Promise<VideoJob> {
    return request<VideoJob>("/jobs", {
      method: "POST",
      body: JSON.stringify({
        project_id: projectId,
        resolution,
        spec,
      }),
    });
  },

  async getJob(id: string): Promise<VideoJob> {
    return request<VideoJob>(`/jobs/${id}`);
  },

  async listAudioPresets(): Promise<{ presets: AudioPreset[] }> {
    return request<{ presets: AudioPreset[] }>("/audio-presets");
  },

  async uploadAudio(file: File): Promise<{ key: string; url: string; filename: string; sizeBytes: number }> {
    const form = new FormData();
    form.append("file", file);
    const res = await fetch(`${API_BASE}/audio-uploads`, {
      method: "POST",
      body: form,
      credentials: "include",
    });
    if (!res.ok) {
      const errorBody = await res.json().catch(() => ({ error: "Audio upload failed" }));
      throw new Error(errorBody.error || `HTTP ${res.status}`);
    }
    return res.json();
  },
};
