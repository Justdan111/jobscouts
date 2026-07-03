import type { Application, Company, Job, Profile, TriageStatus, AppStage, User } from "./types";

const BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";
const TOKEN_KEY = "jobscout_token";

export const tokenStore = {
  get: () => (typeof window === "undefined" ? null : localStorage.getItem(TOKEN_KEY)),
  set: (t: string) => localStorage.setItem(TOKEN_KEY, t),
  clear: () => localStorage.removeItem(TOKEN_KEY),
};

async function req<T>(path: string, init: RequestInit = {}, auth = true): Promise<T> {
  const headers: Record<string, string> = { ...(init.headers as Record<string, string>) };
  if (!(init.body instanceof FormData)) headers["Content-Type"] = "Content-Type" in headers ? headers["Content-Type"] : "application/json";
  if (auth) {
    const t = tokenStore.get();
    if (t) headers["Authorization"] = `Bearer ${t}`;
  }
  const res = await fetch(`${BASE}${path}`, { ...init, headers, cache: "no-store" });
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.error || `Request failed (${res.status})`);
  }
  return res.json() as Promise<T>;
}

export const api = {
  register: (email: string, password: string, invite: string) =>
    req<{ token: string; user: User }>("/api/auth/register", { method: "POST", body: JSON.stringify({ email, password, invite }) }, false),
  login: (email: string, password: string) =>
    req<{ token: string; user: User }>("/api/auth/login", { method: "POST", body: JSON.stringify({ email, password }) }, false),

  getProfile: () => req<{ profile: Profile }>("/api/profile"),
  saveProfile: (p: Profile) => req<{ profile: Profile }>("/api/profile", { method: "PUT", body: JSON.stringify(p) }),
  uploadResume: (file: File) => {
    const fd = new FormData();
    fd.append("file", file);
    return req<{ resumeFile: string }>("/api/profile/resume", { method: "POST", body: fd });
  },

  listJobs: () => req<{ jobs: Job[] }>("/api/jobs"),
  setJobStatus: (id: string, status: TriageStatus) =>
    req<{ job: Job }>(`/api/jobs/${id}`, { method: "PATCH", body: JSON.stringify({ status }) }),
  refresh: () => req<{ fetched: number; added: number; scored: number }>("/api/refresh", { method: "POST" }),
  draft: (id: string) => req<{ application: Application }>(`/api/jobs/${id}/draft`, { method: "POST" }),

  listCompanies: (hiringOnly = false) =>
    req<{ companies: Company[] }>(`/api/companies${hiringOnly ? "?hiringOnly=1" : ""}`),

  listApplications: () => req<{ applications: Application[] }>("/api/applications"),
  updateApplication: (id: string, payload: Partial<{ stage: AppStage; notes: string; resumeHighlights: string; coverEmail: string; pitch: string }>) =>
    req<{ application: Application }>(`/api/applications/${id}`, { method: "PUT", body: JSON.stringify(payload) }),
};
