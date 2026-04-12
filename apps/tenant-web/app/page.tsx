import Link from "next/link";
import { cookies } from "next/headers";
import { redirect } from "next/navigation";
import { LogoutButton } from "@/components/logout-button";

const TOKEN_COOKIE = "zenvikar_tenant_token";

interface TenantAccess {
  tenantId: string;
  tenantSlug: string;
  tenantName: string;
  role: string;
}

async function loadSession(token: string) {
  const apiURL = process.env.API_INTERNAL_URL || "http://api:8080";
  const res = await fetch(`${apiURL}/api/v1/auth/me`, {
    headers: { Authorization: `Bearer ${token}` },
    cache: "no-store",
  });
  if (!res.ok) return null;
  return res.json();
}

async function loadTenantAccess(token: string): Promise<TenantAccess[] | null> {
  const apiURL = process.env.API_INTERNAL_URL || "http://api:8080";
  const res = await fetch(`${apiURL}/api/v1/auth/tenants`, {
    headers: { Authorization: `Bearer ${token}` },
    cache: "no-store",
  });
  if (!res.ok) return null;
  const data = await res.json();
  return Array.isArray(data?.tenants) ? data.tenants : [];
}

export default async function TenantHomePage() {
  const cookieStore = await cookies();
  const token = cookieStore.get(TOKEN_COOKIE)?.value;

  if (!token) {
    redirect("/login");
  }

  const session = await loadSession(token);
  if (!session) {
    redirect("/login?reauth=1");
  }

  const tenants = await loadTenantAccess(token);
  if (!tenants || tenants.length === 0) {
    redirect("/login?reauth=1");
  }

  if (tenants.length === 1) {
    redirect(`/t/${tenants[0].tenantSlug}`);
  }

  return (
    <main className="mx-auto flex min-h-screen w-full max-w-3xl flex-col gap-6 px-6 py-10">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Choose Tenant</h1>
          <p className="text-sm text-gray-600">Signed in as {session.email}</p>
        </div>
        <LogoutButton />
      </div>

      <div className="space-y-3">
        {tenants.map((tenant) => (
          <Link
            key={tenant.tenantId}
            href={`/t/${tenant.tenantSlug}`}
            className="block rounded border border-gray-200 px-4 py-3 hover:bg-gray-50"
          >
            <p className="font-medium">{tenant.tenantName}</p>
            <p className="text-sm text-gray-600">
              {tenant.tenantSlug} · {tenant.role}
            </p>
          </Link>
        ))}
      </div>
    </main>
  );
}
