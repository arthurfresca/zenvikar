interface Props {
  params: Promise<{ tenantSlug: string }>;
}

export default async function TenantDashboardPage({ params }: Props) {
  const { tenantSlug } = await params;
  return (
    <div>
      <h1 className="text-2xl font-bold mb-4">Dashboard</h1>
      <p className="text-gray-600">
        Managing tenant: <strong>{tenantSlug}</strong>
      </p>
    </div>
  );
}
