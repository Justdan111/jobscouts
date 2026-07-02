"use client";

import Link from "next/link";
import { AuthProvider, useAuth } from "@/lib/auth";

function Nav() {
  const { user, logout } = useAuth();
  return (
    <div className="mx-auto max-w-5xl px-5">
      <nav className="flex items-center justify-between border-b border-line py-4">
        <Link href="/" className="font-display text-lg font-700 tracking-tight text-ink">
          Job<span className="text-signal">Scout</span>
        </Link>
        <div className="flex items-center gap-5 text-[13px] text-muted">
          {user ? (
            <>
              <Link href="/" className="hover:text-ink">Discover</Link>
              <Link href="/applications" className="hover:text-ink">Applications</Link>
              <Link href="/profile" className="hover:text-ink">Profile</Link>
              <button onClick={logout} className="rounded-md border border-line px-2.5 py-1 font-500 text-ink hover:bg-paper">Log out</button>
            </>
          ) : (
            <Link href="/login" className="hover:text-ink">Log in</Link>
          )}
        </div>
      </nav>
    </div>
  );
}

export default function AppShell({ children }: { children: React.ReactNode }) {
  return (
    <AuthProvider>
      <Nav />
      {children}
    </AuthProvider>
  );
}
