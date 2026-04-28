"use client";

import { useMemo, useState } from "react";
import type { OpeningHours } from "@zenvikar/types";
import type { BlockedDateDetails, ServiceMemberDetails } from "@/lib/tenant-session";
import { getTranslations, type Locale } from "@/lib/i18n";

type DayState = {
  dayOfWeek: number;
  enabled: boolean;
  openMinutes: number;
  closeMinutes: number;
};

type Props = {
  tenantSlug: string;
  member: ServiceMemberDetails;
  openingHours: OpeningHours[];
  blockedDates: BlockedDateDetails[];
  copySources: Array<{ id: string; name: string }>;
  locale: Locale;
};

function parseMinutes(value: string) {
  const [hours, minutes] = value.slice(0, 5).split(":").map(Number);
  return hours * 60 + minutes;
}

function formatMinutes(value: number) {
  return `${Math.floor(value / 60).toString().padStart(2, "0")}:${Math.floor(value % 60).toString().padStart(2, "0")}`;
}

export function InteractiveScheduleEditor({ tenantSlug, member, openingHours, blockedDates, copySources, locale }: Props) {
  const t = getTranslations(locale);

  const initialDays = useMemo<DayState[]>(() => {
    const byDay = new Map(openingHours.map((item) => [item.dayOfWeek, item]));
    return t.days_short.map((_, dayOfWeek) => {
      const item = byDay.get(dayOfWeek);
      return {
        dayOfWeek,
        enabled: item?.enabled ?? true,
        openMinutes: parseMinutes(item?.openTime || "09:00"),
        closeMinutes: parseMinutes(item?.closeTime || "18:00"),
      };
    });
  }, [openingHours, t]);

  const [days, setDays] = useState(initialDays);

  function updateDay(dayOfWeek: number, patch: Partial<DayState>) {
    setDays((current) =>
      current.map((item) => {
        if (item.dayOfWeek !== dayOfWeek) return item;
        const next = { ...item, ...patch };
        if (next.closeMinutes <= next.openMinutes) {
          next.closeMinutes = Math.min(next.openMinutes + 60, 23 * 60 + 45);
        }
        return next;
      })
    );
  }

  return (
    <div className="overflow-hidden rounded-2xl border border-gray-200 bg-gray-50 dark:border-white/8 dark:bg-slate-950/55">
      {/* Header */}
      <div className="flex items-center justify-between gap-3 border-b border-gray-200 px-5 py-4 dark:border-white/8">
        <div className="flex items-center gap-3">
          <div className="flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-full bg-gray-200 text-xs font-bold text-gray-600 dark:bg-white/8 dark:text-slate-300">
            {member.memberName.split(" ").map((n: string) => n[0]).slice(0, 2).join("").toUpperCase()}
          </div>
          <div>
            <p className="text-sm font-semibold text-gray-900 dark:text-white">{member.memberName}</p>
            <p className="text-xs text-gray-500 dark:text-slate-500">{member.memberEmail}</p>
          </div>
        </div>
        <span className="rounded-full border border-gray-200 px-2.5 py-1 text-[11px] font-medium text-gray-500 dark:border-white/10 dark:text-slate-400">
          {t.calendar_badge}
        </span>
      </div>

      <div className="grid gap-6 p-5 lg:grid-cols-2">
        {/* Opening hours */}
        <div className="space-y-4">
          <p className="text-sm font-semibold text-gray-900 dark:text-white">{t.weekly_hours}</p>

          {/* Quick templates */}
          <div className="space-y-2">
            <form action={`/t/${tenantSlug}/services`} method="POST" className="flex items-center gap-2 rounded-xl border border-gray-200 bg-white p-3 dark:border-white/8 dark:bg-slate-900/60">
              <input type="hidden" name="action" value="apply-schedule-template" />
              <input type="hidden" name="serviceMemberId" value={member.id} />
              <select
                name="template"
                defaultValue="weekdays"
                className="flex-1 rounded-lg border border-gray-300 bg-white px-3 py-2 text-sm text-gray-700 outline-none dark:border-white/10 dark:bg-slate-950 dark:text-white"
              >
                <option value="weekdays">{t.template_weekdays}</option>
                <option value="extended">{t.template_extended}</option>
                <option value="weekend">{t.template_weekend}</option>
              </select>
              <button className="flex-shrink-0 rounded-lg bg-gray-900 px-3 py-2 text-xs font-semibold text-white transition hover:bg-gray-700 dark:bg-white dark:text-slate-950 dark:hover:bg-slate-100">
                {t.apply_template}
              </button>
            </form>

            {copySources.length > 0 ? (
              <form action={`/t/${tenantSlug}/services`} method="POST" className="flex items-center gap-2 rounded-xl border border-gray-200 bg-white p-3 dark:border-white/8 dark:bg-slate-900/60">
                <input type="hidden" name="action" value="copy-schedule-from-member" />
                <input type="hidden" name="targetServiceMemberId" value={member.id} />
                <select
                  name="sourceServiceMemberId"
                  defaultValue=""
                  className="flex-1 rounded-lg border border-gray-300 bg-white px-3 py-2 text-sm text-gray-700 outline-none dark:border-white/10 dark:bg-slate-950 dark:text-white"
                >
                  <option value="" disabled>{t.copy_from_specialist}</option>
                  {copySources.map((source) => (
                    <option key={source.id} value={source.id}>{source.name}</option>
                  ))}
                </select>
                <button className="flex-shrink-0 rounded-lg border border-gray-200 bg-gray-50 px-3 py-2 text-xs font-medium text-gray-700 transition hover:bg-gray-100 dark:border-white/10 dark:bg-white/5 dark:text-white dark:hover:bg-white/10">
                  {t.copy}
                </button>
              </form>
            ) : null}
          </div>

          {/* Day rows */}
          <div className="space-y-2">
            {days.map((item) => {
              const startPct = (item.openMinutes / (24 * 60)) * 100;
              const widthPct = Math.max(((item.closeMinutes - item.openMinutes) / (24 * 60)) * 100, 3);
              return (
                <form
                  key={item.dayOfWeek}
                  action={`/t/${tenantSlug}/services`}
                  method="POST"
                  className={`overflow-hidden rounded-2xl border transition ${
                    item.enabled
                      ? "border-gray-200 bg-white dark:border-white/10 dark:bg-slate-900/60"
                      : "border-gray-100 bg-gray-50 opacity-60 dark:border-white/5 dark:bg-slate-950/30"
                  }`}
                >
                  <input type="hidden" name="action" value="upsert-opening-hour" />
                  <input type="hidden" name="serviceMemberId" value={member.id} />
                  <input type="hidden" name="dayOfWeek" value={item.dayOfWeek} />
                  <input type="hidden" name="openTime" value={formatMinutes(item.openMinutes)} />
                  <input type="hidden" name="closeTime" value={formatMinutes(item.closeMinutes)} />
                  <input type="hidden" name="enabled" value={item.enabled ? "on" : ""} />

                  <div className="flex items-center justify-between gap-3 px-4 py-3">
                    <div className="flex items-center gap-3">
                      {/* Toggle */}
                      <label className="flex cursor-pointer items-center">
                        <input
                          type="checkbox"
                          checked={item.enabled}
                          onChange={(e) => updateDay(item.dayOfWeek, { enabled: e.target.checked })}
                          className="sr-only"
                        />
                        <div className={`flex h-5 w-9 items-center rounded-full border transition ${item.enabled ? "border-indigo-300 bg-indigo-100 dark:border-cyan-400/50 dark:bg-cyan-400/20" : "border-gray-300 bg-gray-100 dark:border-white/15 dark:bg-white/5"}`}>
                          <div className={`mx-0.5 h-3.5 w-3.5 rounded-full transition-all ${item.enabled ? "translate-x-4 bg-indigo-600 dark:bg-cyan-400" : "translate-x-0 bg-gray-400 dark:bg-slate-600"}`} />
                        </div>
                      </label>
                      <span className="w-9 text-sm font-medium text-gray-700 dark:text-slate-300">{t.days_short[item.dayOfWeek]}</span>
                    </div>
                    {item.enabled ? (
                      <span className="font-mono text-xs text-gray-500 dark:text-slate-400">
                        {formatMinutes(item.openMinutes)}–{formatMinutes(item.closeMinutes)}
                      </span>
                    ) : (
                      <span className="text-xs text-gray-400 dark:text-slate-600">{t.closed}</span>
                    )}
                  </div>

                  {item.enabled ? (
                    <div className="space-y-3 border-t border-gray-100 px-4 pb-3 pt-3 dark:border-white/8">
                      {/* Visual bar */}
                      <div className="h-2 overflow-hidden rounded-full bg-gray-100 dark:bg-white/5">
                        <div
                          className="h-full rounded-full bg-indigo-400 dark:bg-cyan-400/70"
                          style={{ marginLeft: `${startPct}%`, width: `${widthPct}%` }}
                        />
                      </div>
                      <div className="grid gap-3 sm:grid-cols-2">
                        <label className="block">
                          <span className="text-[11px] font-medium text-gray-400 dark:text-slate-500">{t.open_label} · {formatMinutes(item.openMinutes)}</span>
                          <input
                            type="range" min={0} max={23 * 60} step={15}
                            value={item.openMinutes}
                            onChange={(e) => updateDay(item.dayOfWeek, { openMinutes: Number(e.target.value) })}
                            className="mt-1.5 w-full accent-indigo-600 dark:accent-cyan-400"
                          />
                        </label>
                        <label className="block">
                          <span className="text-[11px] font-medium text-gray-400 dark:text-slate-500">{t.close_label} · {formatMinutes(item.closeMinutes)}</span>
                          <input
                            type="range" min={60} max={24 * 60} step={15}
                            value={item.closeMinutes}
                            onChange={(e) => updateDay(item.dayOfWeek, { closeMinutes: Number(e.target.value) })}
                            className="mt-1.5 w-full accent-indigo-600 dark:accent-cyan-400"
                          />
                        </label>
                      </div>
                      <div className="flex justify-end">
                        <button className="rounded-lg border border-gray-200 bg-gray-50 px-3 py-1.5 text-xs font-medium text-gray-700 transition hover:bg-gray-100 dark:border-white/10 dark:bg-white/5 dark:text-white dark:hover:bg-white/10">
                          {t.save_prefix} {t.days_short[item.dayOfWeek]}
                        </button>
                      </div>
                    </div>
                  ) : null}
                </form>
              );
            })}
          </div>
        </div>

        {/* Blocked dates */}
        <div className="space-y-4">
          <p className="text-sm font-semibold text-gray-900 dark:text-white">{t.blocked_dates}</p>

          <form action={`/t/${tenantSlug}/services`} method="POST" className="space-y-3 rounded-2xl border border-gray-200 bg-white p-4 dark:border-white/8 dark:bg-slate-900/60">
            <input type="hidden" name="action" value="create-blocked-range" />
            <input type="hidden" name="membershipId" value={member.membershipId} />
            <div className="grid gap-2 sm:grid-cols-2">
              <div>
                <label className="mb-1.5 block text-[11px] font-medium text-gray-500 dark:text-slate-500">{t.date_from}</label>
                <input
                  type="date" name="startDate" required
                  className="w-full rounded-xl border border-gray-300 bg-white px-3 py-2.5 text-sm text-gray-900 outline-none transition focus:border-gray-500 dark:border-white/10 dark:bg-slate-950 dark:text-white dark:[color-scheme:dark]"
                />
              </div>
              <div>
                <label className="mb-1.5 block text-[11px] font-medium text-gray-500 dark:text-slate-500">{t.date_to}</label>
                <input
                  type="date" name="endDate" required
                  className="w-full rounded-xl border border-gray-300 bg-white px-3 py-2.5 text-sm text-gray-900 outline-none transition focus:border-gray-500 dark:border-white/10 dark:bg-slate-950 dark:text-white dark:[color-scheme:dark]"
                />
              </div>
            </div>
            <input
              name="reason"
              placeholder={t.reason_placeholder}
              className="w-full rounded-xl border border-gray-300 bg-white px-3 py-2.5 text-sm text-gray-900 outline-none transition placeholder:text-gray-400 focus:border-gray-500 dark:border-white/10 dark:bg-slate-950 dark:text-white dark:placeholder:text-slate-600"
            />
            <button className="w-full rounded-xl bg-gray-900 px-4 py-2.5 text-sm font-semibold text-white transition hover:bg-gray-700 active:scale-[0.98] dark:bg-white dark:text-slate-950 dark:hover:bg-slate-100">
              {t.block_date_range}
            </button>
          </form>

          <div className="space-y-2">
            {blockedDates.length === 0 ? (
              <div className="rounded-2xl border border-dashed border-gray-200 p-4 text-center text-sm text-gray-400 dark:border-white/10 dark:text-slate-600">
                {t.no_blocked_dates}
              </div>
            ) : (
              blockedDates.slice(0, 8).map((item) => (
                <form
                  key={`${item.membershipId}-${item.date}`}
                  action={`/t/${tenantSlug}/services`}
                  method="POST"
                  className="flex items-center justify-between gap-3 rounded-2xl border border-gray-100 bg-white px-4 py-3 dark:border-white/8 dark:bg-slate-900/60"
                >
                  <input type="hidden" name="action" value="delete-blocked-date" />
                  <input type="hidden" name="membershipId" value={member.membershipId} />
                  <input type="hidden" name="date" value={item.date.slice(0, 10)} />
                  <div className="flex items-center gap-3">
                    <div className="flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-lg bg-red-50 dark:bg-red-400/10">
                      <svg className="h-3.5 w-3.5 text-red-500 dark:text-red-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                        <path strokeLinecap="round" strokeLinejoin="round" d="M18.364 18.364A9 9 0 005.636 5.636m12.728 12.728A9 9 0 015.636 5.636m12.728 12.728L5.636 5.636" />
                      </svg>
                    </div>
                    <div>
                      <p className="text-sm font-medium text-gray-900 dark:text-white">{item.date.slice(0, 10)}</p>
                      {item.reason ? <p className="text-xs text-gray-500 dark:text-slate-500">{item.reason}</p> : null}
                    </div>
                  </div>
                  <button className="flex-shrink-0 rounded-lg border border-gray-200 px-2.5 py-1.5 text-xs text-gray-500 transition hover:bg-gray-50 dark:border-white/10 dark:text-slate-400 dark:hover:bg-white/5 dark:hover:text-white">
                    {t.remove}
                  </button>
                </form>
              ))
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
