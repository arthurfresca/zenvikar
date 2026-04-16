"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { FormEvent, useEffect, useMemo, useState } from "react";
import { getApiPublicUrl } from "@zenvikar/config";
import { clearAuthToken, persistAuthToken } from "@/lib/auth";

export default function BookingLoginPage() {
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
    setIsReauth(params.get("reauth") === "1");
    setOrigin(window.location.origin);

    const authToken = params.get("authToken");
    const authExpiresAt = params.get("authExpiresAt");
    const oauthError = params.get("error");

    if (authToken && authExpiresAt) {
      persistAuthToken(authToken, authExpiresAt);
      router.replace(params.get("next") || "/");
      router.refresh();
      return;
    }

    if (oauthError) {
      setError(`Social login failed (${oauthError}).`);
    }
  }, []);

  useEffect(() => {
    if (isReauth) {
      clearAuthToken();
    }
  }, [isReauth]);

  const googleOAuthURL = useMemo(
    () => buildOAuthURL("google", origin, nextPath),
    [origin, nextPath]
  );
  const facebookOAuthURL = useMemo(
    () => buildOAuthURL("facebook", origin, nextPath),
    [origin, nextPath]
  );

  async function onLoginSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setLoading(true);
    setError(null);

    try {
      const res = await fetch(`${getApiPublicUrl()}/api/v1/auth/login`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email, password }),
      });
      const data = await readResponseBody(res);
      if (!res.ok) throw new Error(data?.message || "Login failed");
      if (!data.token || !data.expiresAt) throw new Error("Login response was missing auth token");

      persistAuthToken(data.token, data.expiresAt);
      router.push(nextPath);
      router.refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Login failed");
    } finally {
      setLoading(false);
    }
  }

  return (
    <main className="mx-auto flex min-h-screen w-full max-w-md flex-col justify-center px-6 py-10">
      <h1 className="mb-2 text-2xl font-semibold">Sign In</h1>
      <p className="mb-6 text-sm text-gray-600">Access your booking account.</p>

      <form className="space-y-3" onSubmit={onLoginSubmit}>
        <input
          type="email"
          required
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          placeholder="Email"
          className="w-full rounded border border-gray-300 px-3 py-2"
        />
        <input
          type="password"
          required
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          placeholder="Password"
          className="w-full rounded border border-gray-300 px-3 py-2"
        />
        <button
          type="submit"
          disabled={loading}
          className="w-full rounded bg-black px-3 py-2 text-white disabled:opacity-60"
        >
          {loading ? "Signing In..." : "Sign In"}
        </button>
      </form>

      <div className="my-6 h-px bg-gray-200" />

      <div className="space-y-3">
        <a
          href={googleOAuthURL}
          className="block w-full rounded border border-gray-300 px-3 py-2 text-center"
        >
          Continue with Google
        </a>
        <a
          href={facebookOAuthURL}
          className="block w-full rounded border border-gray-300 px-3 py-2 text-center"
        >
          Continue with Facebook
        </a>
      </div>

      {error ? <p className="mt-4 text-sm text-red-600">{error}</p> : null}

      <p className="mt-6 text-sm text-gray-600">
        New here?{" "}
        <Link href="/signup" className="text-black underline">
          Create account
        </Link>
      </p>
    </main>
  );
}

function buildOAuthURL(provider: "google" | "facebook", origin: string, nextPath: string) {
  const url = new URL(`/api/v1/auth/oauth/${provider}/start`, getApiPublicUrl());
  if (origin) {
    const redirect = new URL("/login", origin);
    const normalizedNext = nextPath || "/";
    if (normalizedNext !== "/") {
      redirect.searchParams.set("next", normalizedNext);
    }
    url.searchParams.set("redirect", redirect.toString());
  }
  return url.toString();
}

async function readResponseBody(res: Response): Promise<{ message?: string; token?: string; expiresAt?: string }> {
  const text = await res.text();
  if (!text) {
    return {};
  }

  try {
    return JSON.parse(text);
  } catch {
    return { message: `${res.status} ${res.statusText}: ${text}` };
  }
}
