import Link from "next/link";
import { Suspense } from "react";
import { LogoutButton } from "@/components/logout-button";
import { ThemeToggle } from "@/components/theme-toggle";
import { requireTenantWorkspace, loadTenantBookings } from "@/lib/tenant-session";
import { getTranslations } from "@/lib/i18n";
import { SidebarNav } from "./components/sidebar-nav";
import { CollapsibleSidebarShell } from "./components/collapsible-sidebar-shell";

type Props = {
  params: Promise<{ tenantSlug: string }>;
  children: React.ReactNode;
};

export default async function TenantLayout({ params, children }: Props) {
  const { tenantSlug } = await params;
  const workspace = await requireTenantWorkspace(tenantSlug);
  const { tenant, session, token } = workspace;
  const t = getTranslations(tenant.defaultLocale);
  const role = session.tenantRoles[tenant.id] || "tenant_user";

  const allBookings = await loadTenantBookings(token, tenant.id, undefined, undefined);
  const pendingCount = allBookings.filter((b) => b.status === "pending").length;

  const initials = (name: string) =>
    name.split(" ").map((n) => n[0]).slice(0, 2).join("").toUpperCase();

  const roleBadgeStyle =
    role.includes("owner")
      ? "bg-purple-100 text-purple-700 dark:bg-purple-400/15 dark:text-purple-300"
      : role.includes("manager")
        ? "bg-blue-100 text-blue-700 dark:bg-blue-400/15 dark:text-blue-300"
        : "bg-teal-100 text-teal-700 dark:bg-teal-400/15 dark:text-teal-300";

  const sidebar = (
    <>
      {/* Workspace brand */}
      <div className="flex items-center border-b border-gray-100 px-3 py-[18px] dark:border-zinc-800/60 group-data-[collapsed=true]:justify-center group-data-[collapsed=true]:px-0">
        <div className="relative flex-shrink-0">
          {tenant.logoUrl ? (
            <img
              src={tenant.logoUrl}
              alt={tenant.displayName}
              className="h-9 w-9 rounded-xl border border-gray-200 bg-white object-cover shadow-sm dark:border-zinc-700"
            />
          ) : (
            <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-gray-100 text-sm font-bold text-gray-600 dark:bg-zinc-800 dark:text-zinc-300">
              {tenant.displayName.slice(0, 2).toUpperCase()}
            </div>
          )}
          {/* Live dot on the logo when collapsed */}
          {tenant.enabled ? (
            <span className="absolute -right-0.5 -top-0.5 hidden h-2.5 w-2.5 rounded-full border-2 border-white bg-emerald-500 dark:border-zinc-950 group-data-[collapsed=true]:block" />
          ) : null}
        </div>

        {/* Name + slug — hidden when collapsed */}
        <div className="ml-3 min-w-0 flex-1 group-data-[collapsed=true]:hidden">
          <p className="truncate text-sm font-semibold text-gray-900 dark:text-white">{tenant.displayName}</p>
          <p className="truncate text-xs text-gray-400 dark:text-zinc-500">/{tenant.slug}</p>
        </div>

        {/* Live dot when expanded */}
        {tenant.enabled ? (
          <span className="ml-2 flex h-2 w-2 flex-shrink-0 rounded-full bg-emerald-500 group-data-[collapsed=true]:hidden" title={t.live_badge} />
        ) : (
          <span className="ml-2 flex h-2 w-2 flex-shrink-0 rounded-full bg-gray-300 dark:bg-zinc-600 group-data-[collapsed=true]:hidden" title={t.hidden_badge} />
        )}
      </div>

      {/* Navigation */}
      <div className="flex-1 overflow-y-auto px-3 py-4 group-data-[collapsed=true]:px-2">
        <Suspense fallback={null}>
          <SidebarNav
            tenantSlug={tenant.slug}
            locale={tenant.defaultLocale}
            pendingCount={pendingCount}
          />
        </Suspense>

        {/* Switch workspace */}
        <div className="mt-6 border-t border-gray-100 pt-4 dark:border-zinc-800/60">
          <div className="group/switchws relative">
            <Link
              href="/"
              className="flex items-center rounded-xl py-2.5 text-sm text-gray-500 transition-colors hover:bg-gray-50 hover:text-gray-700 dark:text-zinc-500 dark:hover:bg-white/5 dark:hover:text-zinc-300 group-data-[collapsed=true]:justify-center group-data-[collapsed=true]:px-2 group-data-[collapsed=false]:gap-3 group-data-[collapsed=false]:px-3"
            >
              <svg className="h-[18px] w-[18px] flex-shrink-0 opacity-60" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
              </svg>
              <span className="group-data-[collapsed=true]:hidden">{t.nav_switch_workspace}</span>
            </Link>
            {/* Tooltip when collapsed */}
            <div className="pointer-events-none absolute left-full top-1/2 z-50 ml-3 -translate-y-1/2 rounded-xl bg-gray-900 px-3 py-2 text-xs font-medium text-white opacity-0 shadow-lg transition-opacity group-hover/switchws:group-data-[collapsed=true]:opacity-100 dark:bg-zinc-700 group-data-[collapsed=false]:hidden">
              {t.nav_switch_workspace}
            </div>
          </div>
        </div>
      </div>

      {/* User footer */}
      <div className="border-t border-gray-100 px-3 py-3 dark:border-zinc-800/60 group-data-[collapsed=true]:px-2">
        <div className="flex items-center gap-3 rounded-xl px-2 py-2 group-data-[collapsed=true]:justify-center group-data-[collapsed=true]:gap-0 group-data-[collapsed=true]:px-0">
          <div className="flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-full bg-gray-200 text-xs font-bold text-gray-600 dark:bg-zinc-700 dark:text-zinc-300">
            {initials(session.name)}
          </div>
          {/* Name + role — hidden when collapsed */}
          <div className="min-w-0 flex-1 group-data-[collapsed=true]:hidden">
            <p className="truncate text-xs font-medium text-gray-900 dark:text-zinc-100">{session.name}</p>
            <span className={`inline-block rounded-full px-1.5 py-0.5 text-[10px] font-medium ${roleBadgeStyle}`}>
              {role.replace("tenant_", "")}
            </span>
          </div>
          {/* Actions — hidden when collapsed */}
          <div className="flex items-center gap-1 group-data-[collapsed=true]:hidden">
            <ThemeToggle />
            <LogoutButton label="" />
          </div>
        </div>
      </div>
    </>
  );

  return (
    <CollapsibleSidebarShell sidebar={sidebar}>
      {children}
    </CollapsibleSidebarShell>
  );
}
