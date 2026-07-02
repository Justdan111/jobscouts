"use client";

import { createContext, useContext, useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { api, tokenStore } from "./api";
import type { User } from "./types";

type AuthCtx = {
  user: User | null;
  ready: boolean;
  login: (email: string, password: string) => Promise<void>;
  register: (email: string, password: string, invite: string) => Promise<void>;
  logout: () => void;
};

const Ctx = createContext<AuthCtx>(null as unknown as AuthCtx);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [ready, setReady] = useState(false);

  // On load, if we have a token, assume a session (the API will 401 if invalid).
  useEffect(() => {
    const t = tokenStore.get();
    if (t) setUser({ id: "me", email: "" });
    setReady(true);
  }, []);

  const login = async (email: string, password: string) => {
    const { token, user } = await api.login(email, password);
    tokenStore.set(token);
    setUser(user);
  };
  const register = async (email: string, password: string, invite: string) => {
    const { token, user } = await api.register(email, password, invite);
    tokenStore.set(token);
    setUser(user);
  };
  const logout = () => {
    tokenStore.clear();
    setUser(null);
  };

  return <Ctx.Provider value={{ user, ready, login, register, logout }}>{children}</Ctx.Provider>;
}

export const useAuth = () => useContext(Ctx);

// Redirect to /login if not signed in.
export function useRequireAuth() {
  const { user, ready } = useAuth();
  const router = useRouter();
  useEffect(() => {
    if (ready && !user) router.replace("/login");
  }, [ready, user, router]);
  return user;
}
