interface Props {
  params: Promise<{ tenantSlug: string }>;
}

async function resolveTenantName(slug: string): Promise<string> {
  const apiURL = process.env.API_INTERNAL_URL || "http://api:8080";
  const res = await fetch(`${apiURL}/api/v1/tenants/resolve?slug=${slug}`, {
    next: { revalidate: 300 },
  });
  if (!res.ok) return slug;
  const tenant = await res.json();
  return tenant.displayName || slug;
}

export default async function TenantDashboardPage({ params }: Props) {
  const { tenantSlug } = await params;
  const tenantName = await resolveTenantName(tenantSlug);

  return (
    <div className="space-y-4">
      <h1 className="text-2xl font-bold">Current Tenant</h1>
      <p className="text-gray-600">
        <strong>{tenantName}</strong> ({tenantSlug})
      </p>
    </div>
  );
}
