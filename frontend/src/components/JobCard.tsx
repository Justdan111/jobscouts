"use client";

import type { Job, TriageStatus, Eligibility } from "@/lib/types";

const ELIG: Record<Eligibility, { label: string; cls: string }> = {
  global: { label: "Global", cls: "bg-signal-dim text-signal" },
  "us-only": { label: "US-only", cls: "bg-red-50 text-red-600" },
  unknown: { label: "Check post", cls: "bg-gray-100 text-muted" },
};

function abbrevBatch(b: string): string {
  const m = b.match(/(Winter|Spring|Summer|Fall|Autumn)\s+(\d{4})/i);
  return m ? m[1][0].toUpperCase() + m[2].slice(2) : b;
}

export default function JobCard({
  job, onStatus, onApply,
}: {
  job: Job;
  onStatus: (id: string, status: TriageStatus) => void;
  onApply: (job: Job) => void;
}) {
  const elig = ELIG[job.eligibility];
  return (
    <article className="rounded-xl border border-line bg-surface p-4 transition-shadow hover:shadow-[0_1px_12px_rgba(17,124,111,0.08)]">
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0">
          <a href={job.url} target="_blank" rel="noreferrer" className="font-display text-[15px] font-600 leading-snug text-ink hover:text-signal">
            {job.title}
          </a>
          <p className="mt-0.5 truncate text-[13px] text-muted">{job.company} · {job.location}</p>
        </div>
        <span className="shrink-0 font-mono text-[13px] font-600 text-ink">{job.score}</span>
      </div>

      <div className="mt-2 flex flex-wrap gap-1">
        {job.yc && <span className="rounded bg-heat/15 px-1.5 py-0.5 font-mono text-[9.5px] font-600 uppercase text-heat">YC {abbrevBatch(job.ycBatch)}</span>}
        {!job.yc && job.funded && <span className="rounded bg-signal-dim px-1.5 py-0.5 font-mono text-[9.5px] font-600 uppercase text-signal">VC-funded</span>}
        {job.internship && <span className="rounded bg-gray-100 px-1.5 py-0.5 font-mono text-[9.5px] font-600 uppercase text-muted">Intern</span>}
      </div>

      <div className="mt-3 flex items-center gap-3">
        <div className="h-1 flex-1 overflow-hidden rounded-full bg-line">
          <div className="heatbar h-full" style={{ width: `${job.score}%` }} />
        </div>
        <span className={`rounded-full px-2 py-0.5 text-[11px] font-600 ${elig.cls}`}>{elig.label}</span>
      </div>

      <p className="mt-2 text-[12px] leading-relaxed text-muted">{job.reason}</p>

      <div className="mt-3 flex flex-wrap items-center gap-1.5 text-[12px]">
        <button onClick={() => onApply(job)} className="rounded-md bg-signal px-2.5 py-1 font-600 text-white hover:opacity-90">
          ⚡ Apply now
        </button>
        {job.status !== "saved" && (
          <button onClick={() => onStatus(job.id, "saved")} className="rounded-md border border-line px-2.5 py-1 font-500 text-ink hover:bg-paper">Save</button>
        )}
        {job.status !== "dismissed" && (
          <button onClick={() => onStatus(job.id, "dismissed")} className="rounded-md px-2.5 py-1 font-500 text-muted hover:text-ink">Dismiss</button>
        )}
        {job.status !== "new" && (
          <span className="ml-auto font-mono text-[11px] uppercase tracking-wide text-muted">{job.status}</span>
        )}
      </div>
    </article>
  );
}
