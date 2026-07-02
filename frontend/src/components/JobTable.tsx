"use client";

import { Fragment, useState } from "react";
import type { Job, TriageStatus, Eligibility } from "@/lib/types";

const ELIG: Record<Eligibility, { label: string; cls: string }> = {
  global: { label: "Global", cls: "bg-signal-dim text-signal" },
  "us-only": { label: "US-only", cls: "bg-red-50 text-red-600" },
  unknown: { label: "Check", cls: "bg-gray-100 text-muted" },
};

function timeAgo(iso: string): string {
  const d = (Date.now() - new Date(iso).getTime()) / 86400000;
  if (isNaN(d)) return "—";
  if (d < 1) return "today";
  if (d < 2) return "1d";
  if (d < 30) return `${Math.floor(d)}d`;
  return `${Math.floor(d / 30)}mo`;
}

function Badges({ job }: { job: Job }) {
  return (
    <div className="mt-1 flex flex-wrap gap-1">
      {job.yc && (
        <span className="rounded bg-heat/15 px-1.5 py-0.5 font-mono text-[9.5px] font-600 uppercase text-heat">
          YC {job.ycBatch ? abbrevBatch(job.ycBatch) : ""}
        </span>
      )}
      {!job.yc && job.funded && (
        <span className="rounded bg-signal-dim px-1.5 py-0.5 font-mono text-[9.5px] font-600 uppercase text-signal">VC-funded</span>
      )}
      {job.internship && (
        <span className="rounded bg-gray-100 px-1.5 py-0.5 font-mono text-[9.5px] font-600 uppercase text-muted">Intern</span>
      )}
    </div>
  );
}

function abbrevBatch(b: string): string {
  // "Summer 2025" -> "S25"
  const m = b.match(/(Winter|Spring|Summer|Fall|Autumn)\s+(\d{4})/i);
  if (!m) return b;
  return m[1][0].toUpperCase() + m[2].slice(2);
}

export default function JobTable({
  jobs, onStatus, onApply,
}: {
  jobs: Job[];
  onStatus: (id: string, status: TriageStatus) => void;
  onApply: (job: Job) => void;
}) {
  const [open, setOpen] = useState<string | null>(null);

  return (
    <div className="overflow-x-auto rounded-xl border border-line bg-surface">
      <table className="w-full min-w-[720px] border-collapse text-[13px]">
        <thead>
          <tr className="border-b border-line text-left font-mono text-[10px] uppercase tracking-wider text-muted">
            <th className="px-3 py-2.5 font-500">Company</th>
            <th className="px-3 py-2.5 font-500">Role</th>
            <th className="px-3 py-2.5 font-500">Location</th>
            <th className="px-3 py-2.5 font-500">Eligible</th>
            <th className="px-3 py-2.5 font-500">Score</th>
            <th className="px-3 py-2.5 font-500">Posted</th>
            <th className="px-3 py-2.5 font-500"></th>
          </tr>
        </thead>
        <tbody>
          {jobs.map((j) => {
            const elig = ELIG[j.eligibility];
            const isOpen = open === j.id;
            return (
              <Fragment key={j.id}>
                <tr className="cursor-pointer border-b border-line align-top hover:bg-paper" onClick={() => setOpen(isOpen ? null : j.id)}>
                  <td className="px-3 py-3">
                    <div className="font-600 text-ink">{j.company}</div>
                    <Badges job={j} />
                  </td>
                  <td className="px-3 py-3">
                    <a href={j.url} target="_blank" rel="noreferrer" onClick={(e) => e.stopPropagation()} className="font-500 text-ink hover:text-signal">
                      {j.title}
                    </a>
                    <div className="mt-0.5 font-mono text-[10px] uppercase text-muted">{j.source}</div>
                  </td>
                  <td className="px-3 py-3 text-muted">{j.location}</td>
                  <td className="px-3 py-3">
                    <span className={`rounded-full px-2 py-0.5 text-[11px] font-600 ${elig.cls}`}>{elig.label}</span>
                  </td>
                  <td className="px-3 py-3">
                    <div className="flex items-center gap-1.5">
                      <span className="font-mono font-600 text-ink">{j.score}</span>
                      <span className="h-1 w-10 overflow-hidden rounded-full bg-line">
                        <span className="heatbar block h-full" style={{ width: `${j.score}%` }} />
                      </span>
                    </div>
                  </td>
                  <td className="px-3 py-3 font-mono text-[12px] text-muted">{timeAgo(j.postedAt)}</td>
                  <td className="px-3 py-3 text-right">
                    <button onClick={(e) => { e.stopPropagation(); onApply(j); }} className="whitespace-nowrap rounded-md bg-signal px-2.5 py-1 text-[12px] font-600 text-white hover:opacity-90">
                      ⚡ Apply
                    </button>
                  </td>
                </tr>
                {isOpen && (
                  <tr className="border-b border-line bg-paper/60">
                    <td colSpan={7} className="px-3 py-3">
                      <p className="text-[13px] leading-relaxed text-ink">{j.reason}</p>
                      {j.description && (
                        <p className="mt-2 max-w-2xl text-[12.5px] leading-relaxed text-muted">
                          {j.description.slice(0, 460)}{j.description.length > 460 ? "…" : ""}
                        </p>
                      )}
                      {j.tags && j.tags.length > 0 && (
                        <div className="mt-2 flex flex-wrap gap-1.5">
                          {j.tags.slice(0, 8).map((t) => (
                            <span key={t} className="rounded border border-line px-1.5 py-0.5 font-mono text-[10.5px] text-muted">{t}</span>
                          ))}
                        </div>
                      )}
                      <div className="mt-3 flex flex-wrap gap-1.5">
                        <button onClick={() => onApply(j)} className="rounded-md bg-signal px-3 py-1.5 text-[12px] font-600 text-white hover:opacity-90">
                          ⚡ Apply now — tailor my résumé &amp; cover
                        </button>
                        <a href={j.url} target="_blank" rel="noreferrer" className="rounded-md border border-line px-3 py-1.5 text-[12px] font-600 text-ink hover:bg-surface">Open posting ↗</a>
                        {j.status !== "saved" && <button onClick={() => onStatus(j.id, "saved")} className="rounded-md border border-line px-2.5 py-1.5 text-[12px] font-500 hover:bg-surface">Save</button>}
                        {j.status !== "dismissed" && <button onClick={() => onStatus(j.id, "dismissed")} className="rounded-md px-2.5 py-1.5 text-[12px] font-500 text-muted hover:text-ink">Dismiss</button>}
                      </div>
                    </td>
                  </tr>
                )}
              </Fragment>
            );
          })}
        </tbody>
      </table>
    </div>
  );
}
