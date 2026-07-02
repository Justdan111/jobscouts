"use client";

import { useEffect, useState } from "react";
import { api } from "@/lib/api";
import type { Job, Application } from "@/lib/types";

export default function ApplyDrawer({ job, onClose }: { job: Job; onClose: () => void }) {
  const [app, setApp] = useState<Application | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [copied, setCopied] = useState<string>("");

  const generate = async () => {
    setLoading(true);
    setError("");
    try {
      const { application } = await api.draft(job.id);
      setApp(application);
    } catch (e) {
      setError((e as Error).message);
    } finally {
      setLoading(false);
    }
  };

  // Auto-generate on open.
  useEffect(() => {
    generate();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const save = async (patch: Partial<Application>) => {
    if (!app) return;
    const next = { ...app, ...patch };
    setApp(next);
    await api.updateApplication(job.id, {
      stage: next.stage,
      notes: next.notes,
      resumeHighlights: next.resumeHighlights,
      coverEmail: next.coverEmail,
      pitch: next.pitch,
    });
  };

  const copy = async (text: string, which: string) => {
    await navigator.clipboard.writeText(text);
    setCopied(which);
    setTimeout(() => setCopied(""), 1500);
  };

  return (
    <div className="fixed inset-0 z-50 flex justify-end bg-black/30" onClick={onClose}>
      <div
        className="h-full w-full max-w-xl overflow-y-auto bg-surface p-6 shadow-xl"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-start justify-between">
          <div>
            <h2 className="font-display text-lg font-700 text-ink">{job.title}</h2>
            <p className="text-[13px] text-muted">{job.company} · {job.location}</p>
          </div>
          <button onClick={onClose} className="text-muted hover:text-ink">✕</button>
        </div>

        <a href={job.url} target="_blank" rel="noreferrer" className="mt-2 inline-block text-[12px] font-600 text-signal">
          Open the posting ↗
        </a>

        {loading && (
          <div className="mt-8 rounded-lg border border-dashed border-line px-4 py-10 text-center text-[13px] text-muted">
            Tailoring your résumé &amp; cover letter to this role…
          </div>
        )}

        {error && (
          <div className="mt-6 rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-[13px] text-red-600">
            {error}
            <div className="mt-1 text-[12px] text-muted">
              Application drafting needs <span className="font-mono">ANTHROPIC_API_KEY</span> set on the backend.
            </div>
          </div>
        )}

        {app && !loading && (
          <div className="mt-6 space-y-5">
            <Field
              label="Tailored résumé highlights"
              value={app.resumeHighlights}
              rows={8}
              onChange={(v) => setApp({ ...app, resumeHighlights: v })}
              onBlur={() => save({ resumeHighlights: app.resumeHighlights })}
              onCopy={() => copy(app.resumeHighlights, "resume")}
              copied={copied === "resume"}
            />
            <Field
              label="Cover email"
              value={app.coverEmail}
              rows={12}
              onChange={(v) => setApp({ ...app, coverEmail: v })}
              onBlur={() => save({ coverEmail: app.coverEmail })}
              onCopy={() => copy(app.coverEmail, "email")}
              copied={copied === "email"}
            />
            <Field
              label="Short pitch (DM / application box)"
              value={app.pitch}
              rows={4}
              onChange={(v) => setApp({ ...app, pitch: v })}
              onBlur={() => save({ pitch: app.pitch })}
              onCopy={() => copy(app.pitch, "pitch")}
              copied={copied === "pitch"}
            />

            <div>
              <label className="text-[12px] font-600 uppercase tracking-wide text-muted">Notes</label>
              <textarea
                className="mt-1 w-full rounded-lg border border-line bg-paper p-3 text-[13px] outline-none focus:border-signal"
                rows={2}
                value={app.notes}
                onChange={(e) => setApp({ ...app, notes: e.target.value })}
                onBlur={() => save({ notes: app.notes })}
                placeholder="Recruiter name, referral, follow-up date…"
              />
            </div>

            <div className="flex items-center justify-between border-t border-line pt-4">
              <div className="flex items-center gap-2">
                <span className="text-[12px] text-muted">Stage:</span>
                <select
                  value={app.stage}
                  onChange={(e) => save({ stage: e.target.value as Application["stage"] })}
                  className="rounded-md border border-line bg-surface px-2 py-1 text-[13px]"
                >
                  <option value="drafted">Drafted</option>
                  <option value="applied">Applied</option>
                  <option value="interviewing">Interviewing</option>
                  <option value="offer">Offer</option>
                  <option value="rejected">Rejected</option>
                </select>
              </div>
              <button onClick={generate} className="rounded-md border border-line px-3 py-1.5 text-[13px] font-600 text-ink hover:bg-paper">
                ↻ Regenerate
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}

function Field({
  label, value, rows, onChange, onBlur, onCopy, copied,
}: {
  label: string; value: string; rows: number;
  onChange: (v: string) => void; onBlur: () => void; onCopy: () => void; copied: boolean;
}) {
  return (
    <div>
      <div className="flex items-center justify-between">
        <label className="text-[12px] font-600 uppercase tracking-wide text-muted">{label}</label>
        <button onClick={onCopy} className="text-[12px] font-600 text-signal hover:opacity-80">
          {copied ? "Copied ✓" : "Copy"}
        </button>
      </div>
      <textarea
        className="mt-1 w-full rounded-lg border border-line bg-paper p-3 text-[13px] leading-relaxed outline-none focus:border-signal"
        rows={rows}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        onBlur={onBlur}
      />
    </div>
  );
}
