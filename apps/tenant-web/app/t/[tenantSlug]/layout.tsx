import { notFound } from "next/navigation";
import type { Tenant } from "@zenvikar/types";

interface Props {
  params: Promise<{ tenantSlug: string }>;
  children: React.ReactNode;
}

async function resolveTenantBySlug(slug: string): Promise<Tenant | null> {
  const res = await fetch(
    `${process.env.API_INTERNAL_URL}/api/v1/tenants/resolve?slug=${slug}`,
    { next: { revalidate: 300 } }
  );
  if (!res.ok) return null;
  return res.json();
}

export default async function TenantLayout({ params, children }: Props) {
  const { tenantSlug } = await params;
  const tenant = await resolveTenantBySlug(tenantSlug);
  if (!tenant) return notFound();

  return (
    <div className="min-h-screen">
      <nav className="border-b px-6 py-3 flex items-center gap-4">
        <span className="font-semibold text-lg">{tenant.displayName}</span>
        <span className="text-sm text-gray-400">/{tenant.slug}</span>
      </nav>
      <main className="p-6">{children}</main>
    </div>
  );
}
