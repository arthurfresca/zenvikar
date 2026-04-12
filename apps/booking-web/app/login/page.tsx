"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { FormEvent, useEffect, useState } from "react";
import { getApiPublicUrl } from "@zenvikar/config";
import { clearAuthToken, persistAuthToken } from "@/lib/auth";

export default function BookingLoginPage() {
  const router = useRouter();
  const [nextPath, setNextPath] = useState("/");
  const [isReauth, setIsReauth] = useState(false);

  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [socialEmail, setSocialEmail] = useState("");
  const [socialName, setSocialName] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    setNextPath(params.get("next") || "/");
    const reauth = params.get("reauth") === "1";
    setIsReauth(reauth);
  }, []);

  useEffect(() => {
    if (isReauth) {
      clearAuthToken();
    }
  }, [isReauth]);

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
      const data = await res.json();
      if (!res.ok) throw new Error(data?.message || "Login failed");

      persistAuthToken(data.token, data.expiresAt);
      router.push(nextPath);
      router.refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Login failed");
    } finally {
      setLoading(false);
    }
  }

  async function socialLogin(provider: "google" | "facebook") {
    setLoading(true);
    setError(null);

    try {
      const res = await fetch(`${getApiPublicUrl()}/api/v1/auth/social`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          provider,
          providerUserId: `${provider}:${socialEmail.toLowerCase()}`,
          email: socialEmail,
          name: socialName,
        }),
      });
      const data = await res.json();
      if (!res.ok) throw new Error(data?.message || "Social login failed");

      persistAuthToken(data.token, data.expiresAt);
      router.push(nextPath);
      router.refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Social login failed");
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
        <input
          type="email"
          required
          value={socialEmail}
          onChange={(e) => setSocialEmail(e.target.value)}
          placeholder="Social Email"
          className="w-full rounded border border-gray-300 px-3 py-2"
        />
        <input
          type="text"
          required
          value={socialName}
          onChange={(e) => setSocialName(e.target.value)}
          placeholder="Full Name"
          className="w-full rounded border border-gray-300 px-3 py-2"
        />
        <button
          type="button"
          disabled={loading || !socialEmail || !socialName}
          onClick={() => socialLogin("google")}
          className="w-full rounded border border-gray-300 px-3 py-2 disabled:opacity-60"
        >
          Continue with Google
        </button>
        <button
          type="button"
          disabled={loading || !socialEmail || !socialName}
          onClick={() => socialLogin("facebook")}
          className="w-full rounded border border-gray-300 px-3 py-2 disabled:opacity-60"
        >
          Continue with Facebook
        </button>
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
