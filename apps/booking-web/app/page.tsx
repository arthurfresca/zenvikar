import { headers } from "next/headers";
import { resolveTenant } from "@/lib/tenant";

export default async function BookingPage() {
  const headersList = await headers();
  const host = headersList.get("x-tenant-host") || headersList.get("host") || "";

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
    <main className="flex min-h-screen flex-col items-center justify-center p-8">
      <h1 className="text-3xl font-bold mb-2">{tenant.displayName}</h1>
      <p className="text-lg text-gray-600 mb-4">
        Book an Appointment
      </p>
      <p className="text-sm text-gray-400">
        Locale: {tenant.defaultLocale} · Timezone: {tenant.timezone}
      </p>
    </main>
  );
}
