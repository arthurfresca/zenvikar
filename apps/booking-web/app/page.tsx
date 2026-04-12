import { headers } from "next/headers";
import { cookies } from "next/headers";
import { redirect } from "next/navigation";
import { resolveTenant } from "@/lib/tenant";
import { LogoutButton } from "./components/logout-button";

const TOKEN_COOKIE = "zenvikar_booking_token";

async function loadSession(token: string) {
  const apiURL = process.env.API_INTERNAL_URL || "http://api:8080";
  const res = await fetch(`${apiURL}/api/v1/auth/me`, {
    headers: { Authorization: `Bearer ${token}` },
    cache: "no-store",
  });
  if (!res.ok) return null;
  return res.json();
}

export default async function BookingPage() {
  const headersList = await headers();
  const cookieStore = await cookies();
  const host = headersList.get("x-tenant-host") || headersList.get("host") || "";
  const token = cookieStore.get(TOKEN_COOKIE)?.value;

  if (!token) {
    redirect("/login");
  }

  const session = await loadSession(token);
  if (!session) {
    redirect("/login?reauth=1");
  }

  let tenant;
  try {
    tenant = await resolveTenant(host);
  } catch {
    return (
      <main className="flex min-h-screen flex-col items-center justify-center p-8">
        <h1 className="text-2xl font-bold text-red-600">Tenant not found</h1>
        <p className="text-gray-500 mt-2">
          Could not resolve tenant from host: {host}
        </p>
      </main>
    );
  }

  return (
    <main className="mx-auto flex min-h-screen w-full max-w-3xl flex-col gap-6 px-6 py-10">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">{tenant.displayName}</h1>
          <p className="text-sm text-gray-600">
            Locale: {tenant.defaultLocale} · Timezone: {tenant.timezone}
          </p>
        </div>
        <LogoutButton />
      </div>

      <section className="rounded border border-gray-200 p-4">
        <h2 className="text-lg font-semibold">Welcome</h2>
        <p className="text-sm text-gray-600">
          Signed in as {session.email} ({session.name})
        </p>
      </section>
    </main>
  );
}
