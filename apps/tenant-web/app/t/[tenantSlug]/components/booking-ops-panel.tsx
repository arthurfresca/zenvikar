"use client";

import { useMemo, useState } from "react";
import type { TenantBooking } from "@/lib/tenant-session";
import { getTranslations, type Locale } from "@/lib/i18n";

type Props = {
  bookings: TenantBooking[];
  tenantSlug: string;
  timezone: string;
  currency: string;
  locale: Locale;
  variant?: "compact" | "full";
};

function formatMoney(cents: number, currency: string, intlLocale: string) {
  return new Intl.NumberFormat(intlLocale, { style: "currency", currency }).format(cents / 100);
}

function formatDateTime(value: string, timezone: string, intlLocale: string) {
  return new Intl.DateTimeFormat(intlLocale, {
    weekday: "short", month: "short", day: "numeric",
    hour: "numeric", minute: "2-digit", timeZone: timezone,
  }).format(new Date(value));
}

function StatusBadge({ status, t }: { status: string; t: ReturnType<typeof getTranslations> }) {
  if (status === "confirmed") return <span className="rounded-full bg-emerald-100 px-2.5 py-1 text-[11px] font-semibold text-emerald-700 dark:bg-emerald-400/15 dark:text-emerald-300">{t.status_confirmed}</span>;
  if (status === "cancelled") return <span className="rounded-full bg-red-100 px-2.5 py-1 text-[11px] font-semibold text-red-700 dark:bg-red-400/15 dark:text-red-300">{t.status_cancelled}</span>;
  return <span className="rounded-full bg-amber-100 px-2.5 py-1 text-[11px] font-semibold text-amber-700 dark:bg-amber-400/15 dark:text-amber-300">{t.status_pending}</span>;
}

function StatusForm({ tenantSlug, bookingId, status, label, className }: { tenantSlug: string; bookingId: string; status: string; label: string; className: string }) {
  return (
    <form action={`/t/${tenantSlug}/services`} method="POST">
      <input type="hidden" name="action" value="update-booking-status" />
      <input type="hidden" name="bookingId" value={bookingId} />
      <input type="hidden" name="status" value={status} />
      <button className={`rounded-xl px-4 py-2 text-xs font-semibold transition ${className}`}>{label}</button>
    </form>
  );
}

export function BookingOpsPanel({ bookings, tenantSlug, timezone, currency, locale, variant = "compact" }: Props) {
  const t = getTranslations(locale);
  const [query, setQuery] = useState("");
  const [status, setStatus] = useState("all");
  const [selectedId, setSelectedId] = useState(bookings[0]?.id || "");

  const filterTabs = [
    { value: "all",       label: t.filter_all },
    { value: "pending",   label: t.filter_pending },
    { value: "confirmed", label: t.filter_confirmed },
    { value: "cancelled", label: t.filter_cancelled },
  ];

  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase();
    return bookings.filter((b) => {
      const matchesSearch = !q || [b.customerName, b.customerEmail, b.serviceName, b.memberName].some((v) => v.toLowerCase().includes(q));
      const matchesStatus = status === "all" || b.status === status;
      return matchesSearch && matchesStatus;
    });
  }, [bookings, query, status]);

  const selected = filtered.find((b) => b.id === selectedId) || filtered[0] || null;

  const searchAndFilter = (
    <div className="space-y-2 p-4">
      <div className="relative">
        <svg className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400 dark:text-slate-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
          <path strokeLinecap="round" strokeLinejoin="round" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
        </svg>
        <input
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          placeholder={t.search_placeholder}
          className="w-full rounded-xl border border-gray-200 bg-white py-2.5 pl-10 pr-4 text-sm text-gray-900 outline-none transition placeholder:text-gray-400 focus:border-gray-400 dark:border-white/10 dark:bg-slate-950/60 dark:text-white dark:placeholder:text-slate-600"
        />
      </div>
      <div className="flex gap-1 rounded-xl border border-gray-200 bg-gray-50 p-1 dark:border-white/10 dark:bg-slate-950/40">
        {filterTabs.map((tab) => (
          <button
            key={tab.value}
            type="button"
            onClick={() => setStatus(tab.value)}
            className={`flex-1 rounded-lg px-2 py-1.5 text-xs font-medium capitalize transition ${
              status === tab.value
                ? "bg-white text-gray-900 shadow-sm dark:bg-white/10 dark:text-white"
                : "text-gray-500 hover:text-gray-700 dark:text-slate-500 dark:hover:text-slate-300"
            }`}
          >
            {tab.label}
          </button>
        ))}
      </div>
    </div>
  );

  const bookingList = (
    <div className="flex-1 overflow-y-auto">
      {filtered.length === 0 ? (
        <p className="px-4 py-10 text-center text-sm text-gray-400 dark:text-slate-500">
          {t.no_bookings_filter}
        </p>
      ) : (
        <div className="divide-y divide-gray-100 dark:divide-white/8">
          {filtered.map((booking) => (
            <button
              key={booking.id}
              type="button"
              onClick={() => setSelectedId(booking.id)}
              className={`w-full px-4 py-3.5 text-left transition ${
                selected?.id === booking.id
                  ? "bg-blue-50 dark:bg-cyan-300/8"
                  : "hover:bg-gray-50 dark:hover:bg-white/3"
              }`}
            >
              <div className="flex items-start justify-between gap-2">
                <p className="text-sm font-medium text-gray-900 dark:text-white">{formatDateTime(booking.startTime, timezone, t.intl_locale)}</p>
                <StatusBadge status={booking.status} t={t} />
              </div>
              <p className="mt-1 text-xs text-gray-600 dark:text-slate-400">{booking.serviceName}</p>
              <p className="text-xs text-gray-400 dark:text-slate-500">{booking.customerName}</p>
            </button>
          ))}
        </div>
      )}
    </div>
  );

  const bookingDetail = selected ? (
    <div className="h-full p-6">
      <div className="flex items-start justify-between gap-3">
        <div>
          <p className="text-lg font-semibold text-gray-900 dark:text-white">{selected.customerName}</p>
          <p className="text-sm text-gray-500 dark:text-slate-500">{selected.customerEmail}</p>
        </div>
        <StatusBadge status={selected.status} t={t} />
      </div>

      <dl className="mt-6 space-y-4">
        {([
          { label: t.service_label,    value: selected.serviceName },
          { label: t.specialist_label, value: selected.memberName },
          { label: t.start_label,      value: formatDateTime(selected.startTime, timezone, t.intl_locale) },
          { label: t.end_label,        value: formatDateTime(selected.endTime, timezone, t.intl_locale) },
          { label: t.value_label,      value: formatMoney(selected.priceCents, currency, t.intl_locale) },
        ] as const).map(({ label, value }) => (
          <div key={label} className="flex items-start justify-between gap-4">
            <dt className="text-sm text-gray-400 dark:text-slate-500">{label}</dt>
            <dd className="text-right text-sm font-medium text-gray-900 dark:text-white">{value}</dd>
          </div>
        ))}
      </dl>

      <div className="mt-8 space-y-2">
        {selected.status !== "confirmed" && (
          <StatusForm tenantSlug={tenantSlug} bookingId={selected.id} status="confirmed" label={t.confirm_booking} className="w-full bg-emerald-100 text-emerald-700 hover:bg-emerald-200 dark:bg-emerald-400/15 dark:text-emerald-200 dark:hover:bg-emerald-400/25" />
        )}
        {selected.status !== "pending" && (
          <StatusForm tenantSlug={tenantSlug} bookingId={selected.id} status="pending" label={t.mark_pending} className="w-full bg-amber-100 text-amber-700 hover:bg-amber-200 dark:bg-amber-400/15 dark:text-amber-200 dark:hover:bg-amber-400/25" />
        )}
        {selected.status !== "cancelled" && (
          <StatusForm tenantSlug={tenantSlug} bookingId={selected.id} status="cancelled" label={t.cancel_booking} className="w-full bg-red-100 text-red-700 hover:bg-red-200 dark:bg-red-400/15 dark:text-red-200 dark:hover:bg-red-400/25" />
        )}
      </div>
    </div>
  ) : (
    <div className="flex h-full items-center justify-center p-10 text-center">
      <p className="text-sm text-gray-400 dark:text-slate-500">{t.select_booking_hint}</p>
    </div>
  );

  if (variant === "full") {
    return (
      <div className="grid min-h-[520px] gap-4 lg:grid-cols-[380px_1fr]">
        <div className="flex flex-col overflow-hidden rounded-2xl border border-gray-200 bg-white dark:border-white/10 dark:bg-white/4">
          <div className="border-b border-gray-100 px-4 py-4 dark:border-white/8">
            <p className="text-xs font-semibold uppercase tracking-widest text-gray-400 dark:text-slate-500">{t.bookings_eyebrow}</p>
            <p className="mt-0.5 text-base font-semibold text-gray-900 dark:text-white">
              {filtered.length} {filtered.length === 1 ? t.booking_singular : t.bookings_plural_lc}
            </p>
          </div>
          {searchAndFilter}
          {bookingList}
        </div>
        <div className="overflow-hidden rounded-2xl border border-gray-200 bg-white dark:border-white/10 dark:bg-white/4">
          {bookingDetail}
        </div>
      </div>
    );
  }

  return (
    <section className="rounded-3xl border border-gray-200 bg-white p-5 shadow-sm dark:border-white/8 dark:bg-white/4">
      <p className="text-xs font-semibold uppercase tracking-widest text-gray-400 dark:text-slate-500">{t.bookings_eyebrow}</p>
      <h3 className="mt-1 text-base font-semibold text-gray-900 dark:text-white">{t.bookings_ops_title}</h3>
      <div className="mt-4">
        {searchAndFilter}
      </div>
      <div className="mt-2">
        {bookingList}
      </div>
      {selected && (
        <div className="mt-4 rounded-2xl border border-gray-200 bg-gray-50 dark:border-white/10 dark:bg-slate-950/60">
          {bookingDetail}
        </div>
      )}
    </section>
  );
}
