"use client";

import Link from "next/link";
import { useSearchParams } from "next/navigation";
import { getTranslations, type Locale } from "@/lib/i18n";
import { useSidebarContext } from "./sidebar-context";

const NAV_ICONS = {
  overview:  <path strokeLinecap="round" strokeLinejoin="round" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6" />,
  bookings:  <path strokeLinecap="round" strokeLinejoin="round" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />,
  services:  <path strokeLinecap="round" strokeLinejoin="round" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />,
  team:      <path strokeLinecap="round" strokeLinejoin="round" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z" />,
  customers: <path strokeLinecap="round" strokeLinejoin="round" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />,
  settings:  <path strokeLinecap="round" strokeLinejoin="round" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z M15 12a3 3 0 11-6 0 3 3 0 016 0z" />,
} as const;

type Props = {
  tenantSlug: string;
  locale: Locale;
  pendingCount: number;
};

export function SidebarNav({ tenantSlug, locale, pendingCount }: Props) {
  const t = getTranslations(locale);
  const { collapsed } = useSidebarContext();
  const searchParams = useSearchParams();
  const active = searchParams.get("tab") || "overview";

  const primaryNav = [
    { id: "overview",  label: t.nav_overview },
    { id: "bookings",  label: t.nav_bookings },
    { id: "services",  label: t.nav_services },
    { id: "team",      label: t.nav_team },
    { id: "customers", label: t.nav_customers },
  ] as const;

  const itemClass = (isActive: boolean) =>
    [
      "group flex w-full items-center rounded-xl py-2.5 text-sm transition-colors",
      collapsed ? "justify-center px-2" : "gap-3 px-3",
      isActive
        ? "bg-gray-100 font-medium text-gray-900 dark:bg-white/8 dark:text-white"
        : "font-normal text-gray-600 hover:bg-gray-50 hover:text-gray-900 dark:text-zinc-400 dark:hover:bg-white/5 dark:hover:text-zinc-100",
    ].join(" ");

  return (
    <nav className="flex flex-col gap-0.5">
      {primaryNav.map((item) => {
        const isActive = item.id === active;
        const hasBadge = item.id === "bookings" && pendingCount > 0;
        return (
          <div key={item.id} className="group/navitem relative">
            <Link href={`/t/${tenantSlug}?tab=${item.id}`} className={itemClass(isActive)}>
              {/* Icon */}
              <svg
                className={`h-[18px] w-[18px] flex-shrink-0 transition-opacity ${isActive ? "opacity-100" : "opacity-60 group-hover:opacity-80"}`}
                fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={isActive ? 2.5 : 2}
              >
                {NAV_ICONS[item.id]}
              </svg>

              {/* Label (hidden when collapsed) */}
              {!collapsed && <span className="flex-1">{item.label}</span>}

              {/* Badge */}
              {hasBadge && !collapsed && (
                <span className="flex h-5 min-w-[20px] items-center justify-center rounded-full bg-amber-100 px-1.5 text-[10px] font-bold text-amber-700 dark:bg-amber-400/20 dark:text-amber-300">
                  {pendingCount}
                </span>
              )}

              {/* Dot badge when collapsed */}
              {hasBadge && collapsed && (
                <span className="absolute right-1.5 top-1.5 h-2 w-2 rounded-full bg-amber-500 ring-2 ring-white dark:ring-zinc-950" />
              )}
            </Link>

            {/* Tooltip — only visible when collapsed + hovered */}
            {collapsed && (
              <div className="pointer-events-none absolute left-full top-1/2 z-50 ml-3 -translate-y-1/2 flex items-center gap-2 rounded-xl bg-gray-900 px-3 py-2 text-xs font-medium text-white opacity-0 shadow-lg transition-opacity group-hover/navitem:opacity-100 dark:bg-zinc-700">
                {item.label}
                {hasBadge && (
                  <span className="rounded-full bg-amber-500 px-1.5 py-0.5 text-[10px] font-bold text-white">
                    {pendingCount}
                  </span>
                )}
              </div>
            )}
          </div>
        );
      })}

      <div className="my-1.5 h-px bg-gray-100 dark:bg-white/8" />

      {/* Settings */}
      <div className="group/navitem relative">
        <Link
          href={`/t/${tenantSlug}?tab=settings`}
          className={itemClass(active === "settings")}
        >
          <svg
            className={`h-[18px] w-[18px] flex-shrink-0 transition-opacity ${active === "settings" ? "opacity-100" : "opacity-60 group-hover:opacity-80"}`}
            fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={active === "settings" ? 2.5 : 2}
          >
            {NAV_ICONS.settings}
          </svg>
          {!collapsed && <span>{t.nav_settings}</span>}
        </Link>
        {collapsed && (
          <div className="pointer-events-none absolute left-full top-1/2 z-50 ml-3 -translate-y-1/2 rounded-xl bg-gray-900 px-3 py-2 text-xs font-medium text-white opacity-0 shadow-lg transition-opacity group-hover/navitem:opacity-100 dark:bg-zinc-700">
            {t.nav_settings}
          </div>
        )}
      </div>
    </nav>
  );
}
