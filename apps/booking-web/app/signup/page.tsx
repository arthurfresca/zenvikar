"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { FormEvent, useState } from "react";
import { getApiPublicUrl } from "@zenvikar/config";
import { persistAuthToken } from "@/lib/auth";

export default function BookingSignupPage() {
  const router = useRouter();

  const [email, setEmail] = useState("");
  const [name, setName] = useState("");
  const [password, setPassword] = useState("");
  const [locale, setLocale] = useState<"en" | "pt">("en");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function onSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setLoading(true);
    setError(null);

    try {
      const res = await fetch(`${getApiPublicUrl()}/api/v1/auth/signup`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email, name, password, locale }),
      });
      const data = await readResponseBody(res);
      if (!res.ok) throw new Error(data?.message || "Signup failed");
      if (!data.token || !data.expiresAt) throw new Error("Signup response was missing auth token");

      persistAuthToken(data.token, data.expiresAt);
      router.push("/");
      router.refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Signup failed");
    } finally {
      setLoading(false);
    }
  }

  return (
    <main className="mx-auto flex min-h-screen w-full max-w-md flex-col justify-center px-6 py-10">
      <h1 className="mb-2 text-2xl font-semibold">Create Account</h1>
      <p className="mb-6 text-sm text-gray-600">Sign up to book appointments.</p>

      <form className="space-y-3" onSubmit={onSubmit}>
        <input
          type="text"
          required
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="Full Name"
          className="w-full rounded border border-gray-300 px-3 py-2"
        />
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
          minLength={8}
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          placeholder="Password (min 8 chars)"
          className="w-full rounded border border-gray-300 px-3 py-2"
        />
        <select
          value={locale}
          onChange={(e) => setLocale(e.target.value as "en" | "pt")}
          className="w-full rounded border border-gray-300 px-3 py-2"
        >
          <option value="en">English</option>
          <option value="pt">Portuguese</option>
        </select>
        <button
          type="submit"
          disabled={loading}
          className="w-full rounded bg-black px-3 py-2 text-white disabled:opacity-60"
        >
          {loading ? "Creating..." : "Create Account"}
        </button>
      </form>

      {error ? <p className="mt-4 text-sm text-red-600">{error}</p> : null}

      <p className="mt-6 text-sm text-gray-600">
        Already have an account?{" "}
        <Link href="/login" className="text-black underline">
          Sign in
        </Link>
      </p>
    </main>
  );
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
