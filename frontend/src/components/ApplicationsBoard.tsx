"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { api } from "@/lib/api";
import { useRequireAuth } from "@/lib/auth";
import type { Application, Job, AppStage } from "@/lib/types";

const STAGES: { key: AppStage; label: string }[] = [
  { key: "drafted", label: "Drafted" },
  { key: "applied", label: "Applied" },
  { key: "interviewing", label: "Interviewing" },
  { key: "offer", label: "Offer" },
  { key: "rejected", label: "Rejected" },
];

export default function ApplicationsBoard() {
  const user = useRequireAuth();
  const [apps, setApps] = useState<Application[]>([]);
  const [jobs, setJobs] = useState<Record<string, Job>>({});
  const [loading, setLoading] = useState(true);
  const [err, setErr] = useState("");

  const load = useCallback(async () => {
    setLoading(true);
    setErr("");
    try {
      const [{ applications }, { jobs }] = await Promise.all([api.listApplications(), api.listJobs()]);
      setApps(applications);
      setJobs(Object.fromEntries(jobs.map((j) => [j.id, j])));
    } catch (e) {
      setErr((e as Error).message);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { if (user) load(); }, [user, load]);

  const move = async (a: Application, stage: AppStage) => {
    setApps((prev) => prev.map((x) => (x.postingId === a.postingId ? { ...x, stage } : x)));
    await api.updateApplication(a.postingId, { stage });
  };

  const byStage = useMemo(() => {
    const m: Record<string, Application[]> = {};
    for (const s of STAGES) m[s.key] = [];
    for (const a of apps) (m[a.stage] ||= []).push(a);
    return m;
  }, [apps]);

  return (
    <main className="mx-auto max-w-3xl px-5 py-8">
      <header className="border-b border-line pb-5">
        <h1 className="font-display text-2xl font-700 tracking-tight text-ink">Applications</h1>
        <p className="mt-1 text-[13px] text-muted">Everything you've drafted, and where it stands.</p>
      </header>

      {err && (
        <div className="mt-4 rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-[13px] text-red-600">{err}</div>
      )}

      {loading ? (
        <p className="mt-8 text-center text-[13px] text-muted">Loading…</p>
      ) : apps.length === 0 ? (
        <div className="mt-6 rounded-xl border border-dashed border-line bg-surface px-4 py-10 text-center text-[13px] text-muted">
          No applications yet. Go to <span className="font-600 text-signal">Discover</span> and draft one.
        </div>
      ) : (
        <div className="mt-6 space-y-7">
          {STAGES.map((s) => (
            <section key={s.key}>
              <h2 className="mb-2 flex items-center gap-2 font-mono text-[11px] uppercase tracking-wider text-muted">
                {s.label}
                <span className="rounded-full bg-line px-1.5 text-ink">{byStage[s.key]?.length || 0}</span>
              </h2>
              <div className="space-y-2">
                {(byStage[s.key] || []).map((a) => {
                  const job = jobs[a.postingId];
                  return (
                    <div key={a.postingId} className="rounded-lg border border-line bg-surface p-3">
                      <div className="flex items-center justify-between gap-3">
                        <div className="min-w-0">
                          <a href={job?.url} target="_blank" rel="noreferrer" className="font-display text-[14px] font-600 text-ink hover:text-signal">
                            {job?.title || a.postingId}
                          </a>
                          <p className="truncate text-[12px] text-muted">{job?.company}</p>
                        </div>
                        <select
                          value={a.stage}
                          onChange={(e) => move(a, e.target.value as AppStage)}
                          className="shrink-0 rounded-md border border-line bg-surface px-2 py-1 text-[12px]"
                        >
                          {STAGES.map((x) => <option key={x.key} value={x.key}>{x.label}</option>)}
                        </select>
                      </div>
                      {a.notes && <p className="mt-1.5 text-[12px] text-muted">{a.notes}</p>}
                    </div>
                  );
                })}
              </div>
            </section>
          ))}
        </div>
      )}
    </main>
  );
}
