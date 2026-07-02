"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { api } from "@/lib/api";
import type { Job, TriageStatus } from "@/lib/types";
import { isFunded } from "@/lib/funded";
import { useRequireAuth } from "@/lib/auth";
import JobCard from "./JobCard";
import JobTable from "./JobTable";
import ApplyDrawer from "./ApplyDrawer";

type StatusFilter = "all" | TriageStatus;
type View = "cards" | "table";
type FundTag = "newlyFunded" | "yc" | "vc";

export default function Dashboard() {
  const user = useRequireAuth();
  const [jobs, setJobs] = useState<Job[]>([]);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [err, setErr] = useState("");
  const [status, setStatus] = useState<StatusFilter>("new");
  const [hideUsOnly, setHideUsOnly] = useState(true);
  const [internOnly, setInternOnly] = useState(false);
  const [fund, setFund] = useState<Set<FundTag>>(new Set());
  const [minScore, setMinScore] = useState(0);
  const [view, setView] = useState<View>("table");
  const [applying, setApplying] = useState<Job | null>(null);

  const load = useCallback(async () => {
    setLoading(true);
    setErr("");
    try {
      const { jobs } = await api.listJobs();
      setJobs(jobs);
    } catch (e) {
      setErr((e as Error).message);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { if (user) load(); }, [user, load]);

  const setJobStatus = async (id: string, next: TriageStatus) => {
    setJobs((prev) => prev.map((j) => (j.id === id ? { ...j, status: next } : j)));
    await api.setJobStatus(id, next);
  };

  const refresh = async () => {
    setRefreshing(true);
    try { await api.refresh(); await load(); }
    catch (e) { setErr((e as Error).message); }
    finally { setRefreshing(false); }
  };

  const toggleFund = (t: FundTag) =>
    setFund((prev) => {
      const n = new Set(prev);
      n.has(t) ? n.delete(t) : n.add(t);
      return n;
    });

  const matchesFund = (j: Job): boolean => {
    if (fund.size === 0) return true; // OR within the funding group
    if (fund.has("newlyFunded") && j.newlyFunded) return true;
    if (fund.has("yc") && j.yc) return true;
    if (fund.has("vc") && j.funded && !j.yc) return true;
    return false;
  };

  const visible = useMemo(
    () =>
      jobs
        .filter((j) => (status === "all" ? true : j.status === status))
        .filter((j) => (hideUsOnly ? j.eligibility !== "us-only" : true))
        .filter((j) => (internOnly ? j.internship : true))
        .filter(matchesFund)
        .filter((j) => j.score >= minScore),
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [jobs, status, hideUsOnly, internOnly, fund, minScore]
  );

  const counts = useMemo(() => {
    const c = { new: 0, newlyFunded: 0, intern: 0 };
    for (const j of jobs) {
      if (j.status === "new") c.new++;
      if (j.newlyFunded) c.newlyFunded++;
      if (j.internship) c.intern++;
    }
    return c;
  }, [jobs]);

  const chip = (active: boolean) =>
    `rounded-full px-3 py-1 text-[12px] font-600 ${active ? "bg-signal text-white" : "border border-line text-muted hover:text-ink"}`;

  return (
    <main className={`mx-auto px-5 py-8 ${view === "table" ? "max-w-5xl" : "max-w-3xl"}`}>
      <header className="flex items-end justify-between border-b border-line pb-5">
        <div>
          <h1 className="font-display text-2xl font-700 tracking-tight text-ink">Discover</h1>
          <p className="mt-1 text-[13px] text-muted">Newly funded startups &amp; remote roles, scored for you.</p>
        </div>
        <button onClick={refresh} disabled={refreshing} className="rounded-lg bg-ink px-3.5 py-2 text-[13px] font-600 text-white hover:opacity-90 disabled:opacity-50">
          {refreshing ? "Scanning…" : "Run scan"}
        </button>
      </header>

      {err && (
        <div className="mt-4 rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-[13px] text-red-600">
          {err} — is the Go backend running on <span className="font-mono">{process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080"}</span>?
        </div>
      )}

      <section className="mt-5 grid grid-cols-3 gap-3">
        <Stat label="New" value={counts.new} />
        <Stat label="Newly funded" value={counts.newlyFunded} accent />
        <Stat label="Internships" value={counts.intern} />
      </section>

      {/* status */}
      <section className="mt-5 flex flex-wrap items-center gap-2">
        {(["new", "saved", "all"] as StatusFilter[]).map((s) => (
          <button key={s} onClick={() => setStatus(s)} className={chip(status === s)}>{s}</button>
        ))}
        <div className="ml-auto flex overflow-hidden rounded-md border border-line text-[12px] font-600">
          <button onClick={() => setView("table")} className={`px-2.5 py-1 ${view === "table" ? "bg-ink text-white" : "text-muted hover:text-ink"}`}>Table</button>
          <button onClick={() => setView("cards")} className={`px-2.5 py-1 ${view === "cards" ? "bg-ink text-white" : "text-muted hover:text-ink"}`}>Cards</button>
        </div>
      </section>

      {/* funding + type filters */}
      <section className="mt-2 flex flex-wrap items-center gap-2">
        <span className="font-mono text-[10px] uppercase tracking-wider text-muted">Funding</span>
        <button onClick={() => toggleFund("newlyFunded")} className={chip(fund.has("newlyFunded"))}>Newly funded</button>
        <button onClick={() => toggleFund("yc")} className={chip(fund.has("yc"))}>YC</button>
        <button onClick={() => toggleFund("vc")} className={chip(fund.has("vc"))}>VC-funded</button>
        <span className="mx-1 h-4 w-px bg-line" />
        <button onClick={() => setInternOnly(!internOnly)} className={chip(internOnly)}>Internships</button>
        <label className="ml-auto flex items-center gap-1.5 text-[12px] text-muted">
          <input type="checkbox" checked={hideUsOnly} onChange={(e) => setHideUsOnly(e.target.checked)} className="accent-signal" />
          Hide US-only
        </label>
        <label className="flex items-center gap-1.5 text-[12px] text-muted">
          min
          <input type="range" min={0} max={100} step={10} value={minScore} onChange={(e) => setMinScore(Number(e.target.value))} className="accent-signal" />
          <span className="w-6 font-mono">{minScore}</span>
        </label>
      </section>

      <section className="mt-4">
        {loading ? (
          <Empty>Loading the radar…</Empty>
        ) : visible.length === 0 ? (
          <Empty>No jobs match these filters. Loosen them or run a fresh scan.</Empty>
        ) : view === "table" ? (
          <JobTable jobs={visible} onStatus={setJobStatus} onApply={setApplying} />
        ) : (
          <div className="space-y-3">
            {visible.map((j) => <JobCard key={j.id} job={j} onStatus={setJobStatus} onApply={setApplying} />)}
          </div>
        )}
      </section>

      {applying && <ApplyDrawer job={applying} onClose={() => setApplying(null)} />}
    </main>
  );
}

function Stat({ label, value, accent }: { label: string; value: number; accent?: boolean }) {
  return (
    <div className="rounded-lg border border-line bg-surface px-3 py-2.5">
      <div className={`font-display text-xl font-700 ${accent ? "text-signal" : "text-ink"}`}>{value}</div>
      <div className="text-[11px] uppercase tracking-wide text-muted">{label}</div>
    </div>
  );
}

function Empty({ children }: { children: React.ReactNode }) {
  return <div className="rounded-xl border border-dashed border-line bg-surface px-4 py-10 text-center text-[13px] text-muted">{children}</div>;
}
