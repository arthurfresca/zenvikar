import Link from "next/link";
import { redirect } from "next/navigation";
import { LogoutButton } from "@/components/logout-button";
import { getTenantToken, loadTenantAccess, loadTenantSession } from "@/lib/tenant-session";

export default async function TenantHomePage() {
  const token = await getTenantToken();
  if (!token) redirect("/login");

  const session = await loadTenantSession(token);
  if (!session) redirect("/login?reauth=1");
  if (session.currentTenantSlug) redirect(`/t/${session.currentTenantSlug}`);

  const tenants = await loadTenantAccess(token);
  if (tenants.length === 0) redirect("/login?reauth=1");
  if (tenants.length === 1) redirect(`/t/${tenants[0].tenantSlug}`);

  return (
    <main className="min-h-screen bg-gray-50 px-6 py-12 dark:bg-slate-950">
      <div className="mx-auto max-w-4xl">
        <div className="mb-10 flex items-start justify-between gap-4">
          <div>
            <div className="flex items-center gap-2.5">
              <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-gray-900 dark:bg-indigo-500/20">
                <svg className="h-4 w-4 text-white dark:text-indigo-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                  <path strokeLinecap="round" strokeLinejoin="round" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
                </svg>
              </div>
              <span className="text-sm font-semibold text-gray-500 dark:text-slate-400">Zenvikar</span>
            </div>
            <h1 className="mt-4 text-3xl font-bold tracking-tight text-gray-900 dark:text-white">Choose a workspace</h1>
            <p className="mt-2 text-sm text-gray-500 dark:text-slate-400">
              Signed in as <span className="font-medium text-gray-700 dark:text-slate-300">{session.email}</span>.
              Select the workspace you want to manage.
            </p>
          </div>
          <LogoutButton />
        </div>

        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {tenants.map((tenant) => (
            <Link
              key={tenant.tenantId}
              href={`/t/${tenant.tenantSlug}`}
              className="group rounded-3xl border border-gray-200 bg-white p-6 shadow-sm transition hover:-translate-y-0.5 hover:border-gray-300 hover:shadow-md dark:border-white/10 dark:bg-white/5 dark:hover:border-white/20 dark:hover:bg-white/8"
            >
              <div className="flex items-start justify-between gap-3">
                <div className="min-w-0">
                  <p className="truncate text-base font-semibold text-gray-900 dark:text-white">{tenant.tenantName}</p>
                  <p className="mt-1 truncate text-sm text-gray-400 dark:text-slate-500">/{tenant.tenantSlug}</p>
                </div>
                <span className="flex-shrink-0 rounded-full border border-gray-200 bg-gray-50 px-2.5 py-1 text-xs font-medium text-gray-600 dark:border-white/10 dark:bg-slate-800 dark:text-slate-300">
                  {tenant.role.replace("tenant_", "")}
                </span>
              </div>
              <div className="mt-8 flex items-center justify-between text-sm">
                <span className="text-gray-400 transition group-hover:text-gray-600 dark:text-slate-500 dark:group-hover:text-slate-400">Open workspace</span>
                <svg className="h-4 w-4 text-gray-400 transition group-hover:translate-x-1 group-hover:text-gray-600 dark:text-slate-600 dark:group-hover:text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                  <path strokeLinecap="round" strokeLinejoin="round" d="M9 5l7 7-7 7" />
                </svg>
              </div>
            </Link>
          ))}
        </div>
      </div>
    </main>
  );
}
