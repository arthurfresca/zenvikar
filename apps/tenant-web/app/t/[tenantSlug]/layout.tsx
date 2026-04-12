import { notFound, redirect } from "next/navigation";
import { cookies } from "next/headers";
import type { Tenant } from "@zenvikar/types";
import { LogoutButton } from "@/components/logout-button";

interface Props {
  params: Promise<{ tenantSlug: string }>;
  children: React.ReactNode;
}

const TOKEN_COOKIE = "zenvikar_tenant_token";

async function resolveTenantBySlug(slug: string): Promise<Tenant | null> {
  const apiURL = process.env.API_INTERNAL_URL || "http://api:8080";
  const res = await fetch(
    `${apiURL}/api/v1/tenants/resolve?slug=${slug}`,
    { next: { revalidate: 300 } }
  );
  if (!res.ok) return null;
  return res.json();
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

export default async function TenantLayout({ params, children }: Props) {
  const { tenantSlug } = await params;
  const cookieStore = await cookies();
  const token = cookieStore.get(TOKEN_COOKIE)?.value;

  if (!token) {
    redirect(`/login?next=/t/${encodeURIComponent(tenantSlug)}`);
  }

  const tenant = await resolveTenantBySlug(tenantSlug);
  if (!tenant) return notFound();

  const session = await loadSession(token);
  if (!session?.tenantRoles || !session.tenantRoles[tenant.id]) {
    redirect(`/login?reauth=1&next=/t/${encodeURIComponent(tenantSlug)}`);
  }

  return (
    <div className="min-h-screen">
      <nav className="border-b px-6 py-3 flex items-center gap-4">
        <span className="font-semibold text-lg">{tenant.displayName}</span>
        <span className="text-sm text-gray-400">/{tenant.slug}</span>
        <span className="ml-auto text-sm text-gray-500">{session.email}</span>
        <LogoutButton />
      </nav>
      <main className="p-6">{children}</main>
    </div>
  );
}
