"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { useAuth } from "@/lib/auth";

export default function LoginPage() {
  const { user, login, register } = useAuth();
  const router = useRouter();
  const [mode, setMode] = useState<"login" | "register">("login");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [invite, setInvite] = useState("");
  const [err, setErr] = useState("");
  const [busy, setBusy] = useState(false);

  useEffect(() => { if (user) router.replace("/"); }, [user, router]);

  const submit = async () => {
    setBusy(true); setErr("");
    try {
      if (mode === "login") await login(email, password);
      else await register(email, password, invite);
      router.replace(mode === "register" ? "/profile" : "/");
    } catch (e) { setErr((e as Error).message); }
    finally { setBusy(false); }
  };

  return (
    <main className="mx-auto max-w-sm px-5 py-16">
      <h1 className="font-display text-2xl font-700 text-ink">
        {mode === "login" ? "Welcome back" : "Create your account"}
      </h1>
      <p className="mt-1 text-[13px] text-muted">
        {mode === "login" ? "Log in to your JobScout." : "Invite-only — ask the host for a code."}
      </p>

      {err && <div className="mt-4 rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-[13px] text-red-600">{err}</div>}

      <div className="mt-6 space-y-3">
        <input className="w-full rounded-lg border border-line px-3 py-2 text-[14px] outline-none focus:border-signal" placeholder="Email" type="email" value={email} onChange={(e) => setEmail(e.target.value)} />
        <input className="w-full rounded-lg border border-line px-3 py-2 text-[14px] outline-none focus:border-signal" placeholder="Password (8+ characters)" type="password" value={password} onChange={(e) => setPassword(e.target.value)} />
        {mode === "register" && (
          <input className="w-full rounded-lg border border-line px-3 py-2 text-[14px] outline-none focus:border-signal" placeholder="Invite code" value={invite} onChange={(e) => setInvite(e.target.value)} />
        )}
        <button onClick={submit} disabled={busy} className="w-full rounded-lg bg-signal px-3 py-2.5 text-[14px] font-600 text-white hover:opacity-90 disabled:opacity-50">
          {busy ? "…" : mode === "login" ? "Log in" : "Create account"}
        </button>
      </div>

      <button onClick={() => { setMode(mode === "login" ? "register" : "login"); setErr(""); }} className="mt-4 text-[13px] text-signal hover:opacity-80">
        {mode === "login" ? "Need an account? Sign up" : "Have an account? Log in"}
      </button>
    </main>
  );
}
