"use client";

import { useEffect, useMemo, useState } from "react";
import { api } from "@/lib/api";
import { useRequireAuth } from "@/lib/auth";
import type { Company } from "@/lib/types";

export default function CompaniesPage() {
  const user = useRequireAuth();
  const [companies, setCompanies] = useState<Company[]>([]);
  const [loading, setLoading] = useState(true);
  const [err, setErr] = useState("");
  const [hiringOnly, setHiringOnly] = useState(false);
  const [remoteOnly, setRemoteOnly] = useState(false);
  const [batch, setBatch] = useState("all");
  const [q, setQ] = useState("");

  useEffect(() => {
    if (!user) return;
    setLoading(true);
    setErr("");
    api
      .listCompanies(false)
      .then((r) => setCompanies(r.companies))
      .catch((e) => setErr((e as Error).message))
      .finally(() => setLoading(false));
  }, [user]);

  const batches = useMemo(() => {
    const s = new Set<string>();
    for (const c of companies) if (c.batch) s.add(c.batch);
    return Array.from(s).sort();
  }, [companies]);

  const filtered = useMemo(() => {
    const needle = q.trim().toLowerCase();
    return companies.filter((c) => {
      if (hiringOnly && !c.isHiring) return false;
      if (remoteOnly && !c.remote) return false;
      if (batch !== "all" && c.batch !== batch) return false;
      if (needle) {
        const hay = [c.name, c.oneLiner, c.longDesc, ...(c.industries || [])]
          .join(" ")
          .toLowerCase();
        if (!hay.includes(needle)) return false;
      }
      return true;
    });
  }, [companies, hiringOnly, remoteOnly, batch, q]);

  return (
    <main className="mx-auto max-w-5xl px-5 py-8">
      <header className="border-b border-line pb-5">
        <h1 className="font-display text-2xl font-700 tracking-tight text-ink">Newly-funded startups</h1>
        <p className="mt-1 text-[13px] text-muted">
          Recent YC batches — hiring or not. Great for cold-emailing founders directly.
        </p>
      </header>

      <div className="mt-5 flex flex-wrap items-center gap-3">
        <input
          value={q}
          onChange={(e) => setQ(e.target.value)}
          placeholder="Search name, industry, description…"
          className="min-w-[220px] flex-1 rounded-lg border border-line bg-surface px-3 py-2 text-[13px] outline-none focus:border-signal"
        />
        <select
          value={batch}
          onChange={(e) => setBatch(e.target.value)}
          className="rounded-lg border border-line bg-surface px-2.5 py-2 text-[13px]"
        >
          <option value="all">All batches</option>
          {batches.map((b) => (
            <option key={b} value={b}>{b}</option>
          ))}
        </select>
        <label className="flex items-center gap-1.5 text-[13px] text-ink">
          <input type="checkbox" checked={hiringOnly} onChange={(e) => setHiringOnly(e.target.checked)} />
          Hiring only
        </label>
        <label className="flex items-center gap-1.5 text-[13px] text-ink">
          <input type="checkbox" checked={remoteOnly} onChange={(e) => setRemoteOnly(e.target.checked)} />
          Remote-friendly
        </label>
      </div>

      {err && (
        <div className="mt-4 rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-[13px] text-red-600">{err}</div>
      )}

      {loading ? (
        <p className="mt-8 text-center text-[13px] text-muted">Loading…</p>
      ) : filtered.length === 0 ? (
        <div className="mt-6 rounded-xl border border-dashed border-line bg-surface px-4 py-10 text-center text-[13px] text-muted">
          No companies match those filters.
        </div>
      ) : (
        <>
          <p className="mt-4 text-[12px] text-muted">
            {filtered.length} of {companies.length} companies
          </p>
          <ul className="mt-3 grid grid-cols-1 gap-3 md:grid-cols-2">
            {filtered.map((c) => (
              <li key={c.slug} className="rounded-lg border border-line bg-surface p-4">
                <div className="flex items-start justify-between gap-3">
                  <div className="min-w-0">
                    <h3 className="font-display text-[15px] font-700 text-ink">{c.name}</h3>
                    <p className="mt-0.5 text-[12px] text-muted">
                      {c.batch}
                      {c.location ? ` · ${c.location}` : ""}
                    </p>
                  </div>
                  {c.isHiring && (
                    <span className="shrink-0 rounded-full bg-signal/10 px-2 py-0.5 text-[11px] font-600 text-signal">
                      Hiring
                    </span>
                  )}
                </div>

                {c.oneLiner && (
                  <p className="mt-2 line-clamp-3 text-[13px] text-ink/80">{c.oneLiner}</p>
                )}

                {c.industries && c.industries.length > 0 && (
                  <div className="mt-2 flex flex-wrap gap-1">
                    {c.industries.slice(0, 4).map((tag) => (
                      <span key={tag} className="rounded-full bg-line px-1.5 py-0.5 text-[10px] text-muted">
                        {tag}
                      </span>
                    ))}
                  </div>
                )}

                <div className="mt-3 flex items-center gap-3 text-[12px]">
                  {c.website && (
                    <a
                      href={c.website}
                      target="_blank"
                      rel="noreferrer"
                      className="font-600 text-signal hover:opacity-80"
                    >
                      Website ↗
                    </a>
                  )}
                  {c.ycUrl && (
                    <a
                      href={c.ycUrl}
                      target="_blank"
                      rel="noreferrer"
                      className="text-muted hover:text-ink"
                    >
                      YC profile ↗
                    </a>
                  )}
                </div>
              </li>
            ))}
          </ul>
        </>
      )}
    </main>
  );
}
