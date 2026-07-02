"use client";

import { useEffect, useState } from "react";
import { api } from "@/lib/api";
import { useRequireAuth } from "@/lib/auth";
import type { Profile, Link } from "@/lib/types";

export default function ProfilePage() {
  const user = useRequireAuth();
  const [p, setP] = useState<Profile | null>(null);
  const [err, setErr] = useState("");
  const [saved, setSaved] = useState(false);
  const [uploading, setUploading] = useState(false);

  useEffect(() => {
    if (!user) return;
    api.getProfile().then(({ profile }) => setP(profile)).catch((e) => setErr((e as Error).message));
  }, [user]);

  if (!user || !p) return <main className="mx-auto max-w-2xl px-5 py-10 text-[13px] text-muted">Loading…</main>;

  const set = (patch: Partial<Profile>) => setP({ ...p, ...patch });

  const save = async () => {
    setErr(""); setSaved(false);
    try { const { profile } = await api.saveProfile(p); setP(profile); setSaved(true); setTimeout(() => setSaved(false), 1800); }
    catch (e) { setErr((e as Error).message); }
  };

  const upload = async (file: File) => {
    setUploading(true); setErr("");
    try { const { resumeFile } = await api.uploadResume(file); set({ resumeFile }); }
    catch (e) { setErr((e as Error).message); }
    finally { setUploading(false); }
  };

  return (
    <main className="mx-auto max-w-2xl px-5 py-8">
      <header className="flex items-end justify-between border-b border-line pb-4">
        <div>
          <h1 className="font-display text-2xl font-700 tracking-tight text-ink">Your profile</h1>
          <p className="mt-1 text-[13px] text-muted">This is what tailors your jobs and applications. The more you fill in, the better.</p>
        </div>
        <button onClick={save} className="rounded-lg bg-signal px-4 py-2 text-[13px] font-600 text-white hover:opacity-90">
          {saved ? "Saved ✓" : "Save"}
        </button>
      </header>

      {err && <div className="mt-4 rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-[13px] text-red-600">{err}</div>}

      <Section title="About you">
        <Row><Text label="Name" value={p.name} onChange={(v) => set({ name: v })} /><Text label="Headline" value={p.headline} onChange={(v) => set({ headline: v })} placeholder="Frontend & Mobile Engineer" /></Row>
        <Row><Text label="Based in" value={p.basedIn} onChange={(v) => set({ basedIn: v })} placeholder="Abuja, Nigeria" /><Text label="Timezone" value={p.timezone} onChange={(v) => set({ timezone: v })} placeholder="UTC+1" /></Row>
        <LinksEditor links={p.links} onChange={(links) => set({ links })} />
      </Section>

      <Section title="Résumé">
        <p className="mb-2 text-[12px] text-muted">Paste your résumé text — this is what every application is tailored from. Optionally attach the PDF too.</p>
        <textarea className="w-full rounded-lg border border-line bg-paper p-3 text-[13px] leading-relaxed outline-none focus:border-signal" rows={8} value={p.resumeText} onChange={(e) => set({ resumeText: e.target.value })} placeholder="Your résumé, in plain text…" />
        <div className="mt-2 flex items-center gap-3 text-[13px]">
          <label className="cursor-pointer rounded-md border border-line px-3 py-1.5 font-600 text-ink hover:bg-paper">
            {uploading ? "Uploading…" : "Attach PDF"}
            <input type="file" accept=".pdf,.doc,.docx,.txt" className="hidden" onChange={(e) => e.target.files?.[0] && upload(e.target.files[0])} />
          </label>
          {p.resumeFile && <span className="font-mono text-[12px] text-muted">attached ✓</span>}
        </div>
      </Section>

      <Section title="What you want (drives matching)">
        <Chips label="Roles" values={p.roles} onChange={(roles) => set({ roles })} placeholder="frontend, mobile…" />
        <Chips label="Stack" values={p.stack} onChange={(stack) => set({ stack })} placeholder="react native, go…" />
        <Chips label="Seniority" values={p.seniorityWant} onChange={(seniorityWant) => set({ seniorityWant })} placeholder="junior, intern…" />
        <Chips label="Job types" values={p.jobTypes} onChange={(jobTypes) => set({ jobTypes })} placeholder="internship, full-time…" />
        <label className="mt-2 flex items-center gap-2 text-[13px] text-ink">
          <input type="checkbox" checked={p.remoteOnly} onChange={(e) => set({ remoteOnly: e.target.checked })} className="accent-signal" />
          Remote only
        </label>
        <Text label="Work eligibility note" value={p.eligibleNote} onChange={(v) => set({ eligibleNote: v })} placeholder="Can work from Nigeria; open to global / EMEA remote" />
      </Section>

      <Section title="Questions (help tailor your applications)">
        <Q label="What kind of role are you looking for?" value={p.lookingFor} onChange={(v) => set({ lookingFor: v })} />
        <Q label="What are your strongest projects / skills?" value={p.strengths} onChange={(v) => set({ strengths: v })} />
        <Q label="Any companies or industries you're targeting?" value={p.targets} onChange={(v) => set({ targets: v })} />
        <Q label="Constraints (visa, timezone, availability)?" value={p.constraints} onChange={(v) => set({ constraints: v })} />
        <Q label="Anything else for your applications?" value={p.extraContext} onChange={(v) => set({ extraContext: v })} />
      </Section>

      <div className="mt-6 flex justify-end">
        <button onClick={save} className="rounded-lg bg-signal px-5 py-2.5 text-[14px] font-600 text-white hover:opacity-90">
          {saved ? "Saved ✓" : "Save profile"}
        </button>
      </div>
    </main>
  );
}

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <section className="mt-6 rounded-xl border border-line bg-surface p-5">
      <h2 className="mb-3 font-display text-[15px] font-600 text-ink">{title}</h2>
      <div className="space-y-3">{children}</div>
    </section>
  );
}
function Row({ children }: { children: React.ReactNode }) {
  return <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">{children}</div>;
}
function Text({ label, value, onChange, placeholder }: { label: string; value: string; onChange: (v: string) => void; placeholder?: string }) {
  return (
    <label className="block">
      <span className="text-[12px] font-600 uppercase tracking-wide text-muted">{label}</span>
      <input className="mt-1 w-full rounded-lg border border-line px-3 py-2 text-[14px] outline-none focus:border-signal" value={value} onChange={(e) => onChange(e.target.value)} placeholder={placeholder} />
    </label>
  );
}
function Q({ label, value, onChange }: { label: string; value: string; onChange: (v: string) => void }) {
  return (
    <label className="block">
      <span className="text-[13px] text-ink">{label}</span>
      <textarea className="mt-1 w-full rounded-lg border border-line bg-paper p-2.5 text-[13px] outline-none focus:border-signal" rows={2} value={value} onChange={(e) => onChange(e.target.value)} />
    </label>
  );
}
function Chips({ label, values, onChange, placeholder }: { label: string; values: string[]; onChange: (v: string[]) => void; placeholder?: string }) {
  const [draft, setDraft] = useState("");
  const add = () => { const t = draft.trim(); if (t && !values.includes(t)) onChange([...values, t]); setDraft(""); };
  return (
    <div>
      <span className="text-[12px] font-600 uppercase tracking-wide text-muted">{label}</span>
      <div className="mt-1 flex flex-wrap items-center gap-1.5 rounded-lg border border-line p-2">
        {values.map((v) => (
          <span key={v} className="flex items-center gap-1 rounded-md bg-signal-dim px-2 py-0.5 text-[12px] text-signal">
            {v}<button onClick={() => onChange(values.filter((x) => x !== v))} className="text-signal/70 hover:text-signal">×</button>
          </span>
        ))}
        <input className="min-w-[120px] flex-1 text-[13px] outline-none" value={draft} placeholder={placeholder} onChange={(e) => setDraft(e.target.value)} onKeyDown={(e) => { if (e.key === "Enter" || e.key === ",") { e.preventDefault(); add(); } }} onBlur={add} />
      </div>
    </div>
  );
}
function LinksEditor({ links, onChange }: { links: Link[]; onChange: (l: Link[]) => void }) {
  return (
    <div>
      <span className="text-[12px] font-600 uppercase tracking-wide text-muted">Portfolio &amp; links</span>
      <div className="mt-1 space-y-2">
        {links.map((l, i) => (
          <div key={i} className="flex gap-2">
            <input className="w-32 rounded-lg border border-line px-2 py-1.5 text-[13px] outline-none focus:border-signal" placeholder="Label" value={l.label} onChange={(e) => onChange(links.map((x, j) => (j === i ? { ...x, label: e.target.value } : x)))} />
            <input className="flex-1 rounded-lg border border-line px-2 py-1.5 text-[13px] outline-none focus:border-signal" placeholder="https://…" value={l.url} onChange={(e) => onChange(links.map((x, j) => (j === i ? { ...x, url: e.target.value } : x)))} />
            <button onClick={() => onChange(links.filter((_, j) => j !== i))} className="px-2 text-muted hover:text-ink">×</button>
          </div>
        ))}
        <button onClick={() => onChange([...links, { label: "", url: "" }])} className="text-[13px] font-600 text-signal hover:opacity-80">+ Add link</button>
      </div>
    </div>
  );
}
