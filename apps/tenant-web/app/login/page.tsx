"use client";

import { FormEvent, useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { getApiPublicUrl } from "@zenvikar/config";
import { clearTenantToken, currentTenantLoginSlug, persistTenantToken } from "@/lib/auth";

export default function TenantLoginPage() {
  const router = useRouter();
  const [nextPath, setNextPath] = useState("/");
  const [isReauth, setIsReauth] = useState(false);
  const [origin, setOrigin] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    setNextPath(params.get("next") || "/");
    setOrigin(window.location.origin);
    setIsReauth(params.get("reauth") === "1");

    const authToken = params.get("authToken");
    const authExpiresAt = params.get("authExpiresAt");
    const oauthError = params.get("error");
    if (authToken && authExpiresAt) {
      persistTenantToken(authToken, authExpiresAt);
      router.replace(params.get("next") || "/");
      router.refresh();
      return;
    }
    if (oauthError) setError(`Social login failed (${oauthError}).`);
  }, []);

  useEffect(() => {
    if (isReauth) clearTenantToken();
  }, [isReauth]);

  async function onSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setLoading(true);
    setError(null);
    try {
      const res = await fetch(`${getApiPublicUrl()}/api/v1/auth/tenant/login`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          ...(currentTenantLoginSlug(nextPath) ? { "X-Tenant-ID": currentTenantLoginSlug(nextPath) } : {}),
        },
        body: JSON.stringify({ email, password }),
      });
      const data = await readResponseBody(res);
      if (!res.ok) throw new Error(data?.message || "Login failed");
      if (!data.token || !data.expiresAt) throw new Error("Login response was missing auth token");
      persistTenantToken(data.token, data.expiresAt);
      router.push(nextPath || "/");
      router.refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Login failed");
    } finally {
      setLoading(false);
    }
  }

  return (
    <main className="flex min-h-screen bg-white dark:bg-slate-950">
      {/* Left decorative panel — desktop only */}
      <div className="relative hidden overflow-hidden bg-gray-950 lg:flex lg:w-[420px] lg:flex-col lg:justify-between lg:p-10 xl:w-[480px]">
        <div className="absolute inset-0 opacity-30" style={{ background: "radial-gradient(ellipse at 20% 80%, rgba(99,102,241,0.4) 0%, transparent 55%), radial-gradient(ellipse at 80% 20%, rgba(20,184,166,0.25) 0%, transparent 55%)" }} />
        <div className="relative flex items-center gap-3">
          <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-white/10">
            <svg className="h-5 w-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
            </svg>
          </div>
          <span className="text-sm font-semibold text-white">Zenvikar</span>
        </div>
        <div className="relative">
          <blockquote className="text-xl font-medium leading-relaxed text-white">
            "Manage your entire booking operation from one intelligent workspace."
          </blockquote>
          <p className="mt-4 text-sm text-white/50">Tenant workspace · Zenvikar Platform</p>
        </div>
        <p className="relative text-xs text-white/30">© 2025 Zenvikar. All rights reserved.</p>
      </div>

      {/* Right form panel */}
      <div className="flex flex-1 flex-col items-center justify-center px-6 py-12">
        <div className="w-full max-w-sm">
          {/* Mobile brand */}
          <div className="mb-6 flex items-center gap-2 lg:hidden">
            <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-gray-900">
              <svg className="h-4 w-4 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
              </svg>
            </div>
            <span className="text-sm font-semibold text-gray-900 dark:text-white">Zenvikar</span>
          </div>

          <h1 className="text-2xl font-semibold tracking-tight text-gray-900 dark:text-white">Sign in to your workspace</h1>
          <p className="mt-1.5 text-sm text-gray-500 dark:text-slate-400">Access your tenant dashboard and manage services</p>

          <div className="mt-8 rounded-3xl border border-gray-200 bg-white p-8 shadow-sm dark:border-white/10 dark:bg-white/4">
            <form className="space-y-4" onSubmit={onSubmit}>
              <div>
                <label htmlFor="email" className="block text-sm font-medium text-gray-700 dark:text-slate-300">Email address</label>
                <input
                  id="email" type="email" required value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  placeholder="you@example.com"
                  className="mt-1.5 w-full rounded-xl border border-gray-300 bg-white px-4 py-3 text-sm text-gray-900 outline-none transition placeholder:text-gray-400 focus:border-gray-500 focus:ring-2 focus:ring-gray-100 dark:border-white/10 dark:bg-white/5 dark:text-white dark:placeholder:text-slate-600 dark:focus:border-white/25 dark:focus:ring-white/5"
                />
              </div>
              <div>
                <label htmlFor="password" className="block text-sm font-medium text-gray-700 dark:text-slate-300">Password</label>
                <input
                  id="password" type="password" required value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  placeholder="••••••••"
                  className="mt-1.5 w-full rounded-xl border border-gray-300 bg-white px-4 py-3 text-sm text-gray-900 outline-none transition placeholder:text-gray-400 focus:border-gray-500 focus:ring-2 focus:ring-gray-100 dark:border-white/10 dark:bg-white/5 dark:text-white dark:placeholder:text-slate-600 dark:focus:border-white/25 dark:focus:ring-white/5"
                />
              </div>

              {error ? (
                <div className="flex items-start gap-3 rounded-xl border border-red-200 bg-red-50 px-4 py-3 dark:border-red-400/20 dark:bg-red-400/10">
                  <svg className="mt-0.5 h-4 w-4 flex-shrink-0 text-red-500 dark:text-red-400" fill="currentColor" viewBox="0 0 20 20">
                    <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
                  </svg>
                  <p className="text-sm text-red-700 dark:text-red-300">{error}</p>
                </div>
              ) : null}

              <button
                type="submit" disabled={loading}
                className="flex w-full items-center justify-center gap-2 rounded-xl bg-gray-900 px-4 py-3 text-sm font-semibold text-white transition hover:bg-gray-700 active:scale-[0.98] disabled:opacity-60 dark:bg-white dark:text-slate-950 dark:hover:bg-slate-100"
              >
                {loading ? (
                  <>
                    <svg className="h-4 w-4 animate-spin" fill="none" viewBox="0 0 24 24">
                      <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                      <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
                    </svg>
                    Signing in…
                  </>
                ) : "Sign in"}
              </button>
            </form>

            <div className="my-6 flex items-center gap-3">
              <div className="h-px flex-1 bg-gray-200 dark:bg-white/8" />
              <span className="text-xs font-medium text-gray-400 dark:text-slate-600">or continue with</span>
              <div className="h-px flex-1 bg-gray-200 dark:bg-white/8" />
            </div>

            <div className="grid grid-cols-2 gap-3">
              <a
                href={buildOAuthURL("google", origin, nextPath)}
                className="flex items-center justify-center gap-2 rounded-xl border border-gray-300 bg-white px-4 py-3 text-sm font-medium text-gray-700 transition hover:bg-gray-50 active:scale-[0.98] dark:border-white/10 dark:bg-white/5 dark:text-slate-300 dark:hover:bg-white/10 dark:hover:text-white"
              >
                <svg className="h-4 w-4" viewBox="0 0 24 24" fill="none">
                  <path d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z" fill="#4285F4" />
                  <path d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z" fill="#34A853" />
                  <path d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z" fill="#FBBC05" />
                  <path d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z" fill="#EA4335" />
                </svg>
                Google
              </a>
              <a
                href={buildOAuthURL("facebook", origin, nextPath)}
                className="flex items-center justify-center gap-2 rounded-xl border border-gray-300 bg-white px-4 py-3 text-sm font-medium text-gray-700 transition hover:bg-gray-50 active:scale-[0.98] dark:border-white/10 dark:bg-white/5 dark:text-slate-300 dark:hover:bg-white/10 dark:hover:text-white"
              >
                <svg className="h-4 w-4" fill="#1877F2" viewBox="0 0 24 24">
                  <path d="M24 12.073c0-6.627-5.373-12-12-12s-12 5.373-12 12c0 5.99 4.388 10.954 10.125 11.854v-8.385H7.078v-3.47h3.047V9.43c0-3.007 1.792-4.669 4.533-4.669 1.312 0 2.686.235 2.686.235v2.953H15.83c-1.491 0-1.956.925-1.956 1.874v2.25h3.328l-.532 3.47h-2.796v8.385C19.612 23.027 24 18.062 24 12.073z" />
                </svg>
                Facebook
              </a>
            </div>
          </div>
        </div>
      </div>
    </main>
  );
}

function buildOAuthURL(provider: "google" | "facebook", origin: string, nextPath: string) {
  const url = new URL(`/api/v1/auth/oauth/${provider}/start`, getApiPublicUrl());
  url.searchParams.set("audience", "tenant-web");
  const tenantSlug = currentTenantLoginSlug(nextPath);
  if (tenantSlug) url.searchParams.set("tenantSlug", tenantSlug);
  if (origin) {
    const redirect = new URL("/login", origin);
    const normalizedNext = nextPath || "/";
    if (normalizedNext !== "/") redirect.searchParams.set("next", normalizedNext);
    url.searchParams.set("redirect", redirect.toString());
  }
  return url.toString();
}

async function readResponseBody(res: Response): Promise<{ message?: string; token?: string; expiresAt?: string }> {
  const text = await res.text();
  if (!text) return {};
  try { return JSON.parse(text); }
  catch { return { message: `${res.status} ${res.statusText}: ${text}` }; }
}
