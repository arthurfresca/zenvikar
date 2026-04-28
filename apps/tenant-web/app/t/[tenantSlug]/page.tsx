import type { OpeningHours, Service } from "@zenvikar/types";
import Link from "next/link";
import {
  loadBlockedDates,
  loadOpeningHours,
  loadServiceMembers,
  loadTenantBookings,
  loadTenantMemberships,
  loadTenantServices,
  requireTenantWorkspace,
  type BlockedDateDetails,
  type MembershipDetails,
  type ServiceMemberDetails,
  type TenantBooking,
} from "@/lib/tenant-session";
import { BookingOpsPanel } from "./components/booking-ops-panel";
import { InteractiveScheduleEditor } from "./components/interactive-schedule-editor";
import { ColorField } from "./components/color-field";
import { CurrencyInput } from "./components/currency-input";
import { getTranslations, type Translations } from "@/lib/i18n";

type Props = {
  params: Promise<{ tenantSlug: string }>;
  searchParams: Promise<{
    tab?: string;
    serviceAction?: string;
    serviceError?: string;
    customer?: string;
    from?: string;
    to?: string;
    newService?: string;
  }>;
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

function summarizeBookings(bookings: TenantBooking[]) {
  return bookings.reduce(
    (acc, b) => {
      acc.total += 1;
      if (b.status === "confirmed") acc.confirmed += 1;
      if (b.status === "pending") acc.pending += 1;
      if (b.status === "cancelled") acc.cancelled += 1;
      return acc;
    },
    { total: 0, confirmed: 0, pending: 0, cancelled: 0 }
  );
}

function memberMap(memberships: MembershipDetails[]) {
  return new Map(memberships.map((m) => [m.id, m]));
}

function customerSummary(bookings: TenantBooking[]) {
  const grouped = new Map<string, { name: string; email: string; visits: number; spent: number; lastVisit: string }>();
  for (const b of bookings) {
    const cur = grouped.get(b.customerId);
    if (cur) {
      cur.visits += 1;
      cur.spent += b.priceCents;
      if (b.startTime > cur.lastVisit) cur.lastVisit = b.startTime;
    } else {
      grouped.set(b.customerId, { name: b.customerName, email: b.customerEmail, visits: 1, spent: b.priceCents, lastVisit: b.startTime });
    }
  }
  return [...grouped.values()].sort((a, b) => b.visits - a.visits).slice(0, 10);
}

const TIMEZONES = [
  "UTC",
  "America/New_York","America/Chicago","America/Denver","America/Los_Angeles",
  "America/Toronto","America/Vancouver","America/Sao_Paulo","America/Argentina/Buenos_Aires",
  "America/Mexico_City","America/Bogota","America/Lima",
  "Europe/London","Europe/Lisbon","Europe/Paris","Europe/Berlin","Europe/Madrid",
  "Europe/Rome","Europe/Amsterdam","Europe/Zurich","Europe/Moscow",
  "Asia/Dubai","Asia/Kolkata","Asia/Bangkok","Asia/Singapore",
  "Asia/Shanghai","Asia/Hong_Kong","Asia/Tokyo","Asia/Seoul",
  "Australia/Sydney","Australia/Melbourne","Pacific/Auckland","Pacific/Honolulu",
];

export default async function TenantDashboardPage({ params, searchParams }: Props) {
  const { tenantSlug } = await params;
  const query = await searchParams;
  const workspace = await requireTenantWorkspace(tenantSlug);
  const t = getTranslations(workspace.tenant.defaultLocale);
  const tab = query.tab || "overview";

  const [services, memberships, bookings] = await Promise.all([
    loadTenantServices(workspace.token, workspace.tenant.id),
    loadTenantMemberships(workspace.token, workspace.tenant.id),
    loadTenantBookings(workspace.token, workspace.tenant.id, query.from, query.to),
  ]);

  const serviceMembersEntries = await Promise.all(
    services.map(async (service) => [
      service.id,
      await loadServiceMembers(workspace.token, workspace.tenant.id, service.id),
    ] as const)
  );

  const serviceMembers = new Map<string, ServiceMemberDetails[]>(serviceMembersEntries);
  const assignedMembers = serviceMembersEntries.flatMap(([, members]) => members);

  const openingHoursEntries = await Promise.all(
    assignedMembers.map(async (member) => [
      member.id,
      await loadOpeningHours(workspace.token, workspace.tenant.id, member.id),
    ] as const)
  );
  const blockedDatesEntries = await Promise.all(
    [...new Set(assignedMembers.map((m) => m.membershipId))].map(async (membershipId) => [
      membershipId,
      await loadBlockedDates(workspace.token, workspace.tenant.id, membershipId),
    ] as const)
  );

  const openingHoursByMember = new Map<string, OpeningHours[]>(openingHoursEntries);
  const blockedDatesByMembership = new Map<string, BlockedDateDetails[]>(blockedDatesEntries);
  const membershipById = memberMap(memberships);
  const bookingSummary = summarizeBookings(bookings);
  const customers = customerSummary(bookings);
  const selectedCustomer = customers.find((c) => c.email === query.customer) || null;
  const customerBookings = selectedCustomer
    ? bookings.filter((b) => b.customerEmail === selectedCustomer.email).sort((a, b) => +new Date(b.startTime) - +new Date(a.startTime)).slice(0, 8)
    : [];
  const enabledServices = services.filter((s) => s.enabled).length;
  const assignedCount = new Set(assignedMembers.map((m) => m.membershipId)).size;

  const defaultFrom = query.from || new Date(Date.now() - 86400000).toISOString().slice(0, 10);
  const defaultTo = query.to || new Date(Date.now() + 86400000 * 14).toISOString().slice(0, 10);

  const pendingBookings = bookings.filter((b) => b.status === "pending");
  const upcomingBookings = bookings
    .filter((b) => b.status === "confirmed" && new Date(b.startTime) >= new Date())
    .sort((a, b) => new Date(a.startTime).getTime() - new Date(b.startTime).getTime())
    .slice(0, 6);

  const slug = workspace.tenant.slug;

  return (
    <div className="space-y-5">
      {query.serviceAction && (
        <div className="flex items-center gap-3 rounded-xl border border-emerald-200 bg-emerald-50 px-4 py-3.5 text-sm text-emerald-700 dark:border-emerald-400/20 dark:bg-emerald-400/10 dark:text-emerald-200">
          <svg className="h-4 w-4 flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}><path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7" /></svg>
          {t.action_completed}: {query.serviceAction.replaceAll("_", " ")}
        </div>
      )}
      {query.serviceError && (
        <div className="flex items-center gap-3 rounded-xl border border-red-200 bg-red-50 px-4 py-3.5 text-sm text-red-700 dark:border-red-400/20 dark:bg-red-400/10 dark:text-red-200">
          <svg className="h-4 w-4 flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}><path strokeLinecap="round" strokeLinejoin="round" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
          {t.could_not_complete}: {query.serviceError.replaceAll("_", " ")}
        </div>
      )}

      {/* ── OVERVIEW ── */}
      {tab === "overview" && (
        <div className="space-y-6">
          <div className="flex flex-wrap items-start justify-between gap-4">
            <div>
              <h1 className="text-2xl font-bold text-gray-900 dark:text-white">{workspace.tenant.displayName}</h1>
              <p className="mt-1 flex items-center gap-1.5 text-sm">
                {workspace.tenant.enabled ? (
                  <><span className="h-1.5 w-1.5 rounded-full bg-emerald-500" /><span className="text-emerald-600 dark:text-emerald-400">{t.booking_page_live}</span></>
                ) : (
                  <><span className="h-1.5 w-1.5 rounded-full bg-gray-300 dark:bg-slate-600" /><span className="text-gray-400 dark:text-slate-500">{t.hidden_badge}</span></>
                )}
              </p>
            </div>
            <DateRangeFilter defaultFrom={defaultFrom} defaultTo={defaultTo} tab="overview" applyLabel={t.apply} />
          </div>

          <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
            <MetricCard icon="services" label={t.metric_live_services} value={enabledServices.toString()} detail={`${services.length - enabledServices} ${t.hidden_lower}`} />
            <MetricCard icon="team" label={t.metric_team_members} value={memberships.length.toString()} detail={`${assignedCount} ${t.assigned_count}`} />
            <MetricCard icon="confirmed" label={t.metric_confirmed} value={bookingSummary.confirmed.toString()} detail={`${bookingSummary.total} ${t.total_bookings}`} />
            <MetricCard icon="pending" label={t.metric_pending} value={bookingSummary.pending.toString()} detail={`${bookingSummary.cancelled} ${t.cancelled_count}`} />
          </div>

          <div className="grid gap-6 xl:grid-cols-[1fr_300px]">
            <div className="space-y-4">
              {pendingBookings.length > 0 && (
                <div className="overflow-hidden rounded-2xl border border-amber-200 dark:border-amber-400/20">
                  <div className="flex items-center gap-3 border-b border-amber-200 bg-amber-50 px-5 py-3.5 dark:border-amber-400/20 dark:bg-amber-400/8">
                    <span className="flex h-5 w-5 items-center justify-center rounded-full bg-amber-500 text-[11px] font-bold text-white">
                      {pendingBookings.length}
                    </span>
                    <p className="text-sm font-semibold text-amber-800 dark:text-amber-300">{t.needs_attention}</p>
                    <Link href={`?tab=bookings`} className="ml-auto text-xs text-amber-600 hover:underline dark:text-amber-400">
                      {t.view_all_link}
                    </Link>
                  </div>
                  <div className="divide-y divide-amber-100 bg-white dark:divide-amber-400/10 dark:bg-amber-400/3">
                    {pendingBookings.slice(0, 5).map((booking) => (
                      <div key={booking.id} className="flex items-center gap-4 px-5 py-3.5">
                        <div className="min-w-0 flex-1">
                          <p className="text-sm font-medium text-gray-900 dark:text-white">{booking.customerName}</p>
                          <p className="mt-0.5 text-xs text-gray-500 dark:text-slate-400">
                            {booking.serviceName} · {formatDateTime(booking.startTime, workspace.tenant.timezone, t.intl_locale)}
                          </p>
                        </div>
                        <div className="flex flex-shrink-0 gap-2">
                          <form action={`/t/${slug}/services`} method="POST">
                            <input type="hidden" name="action" value="update-booking-status" />
                            <input type="hidden" name="bookingId" value={booking.id} />
                            <input type="hidden" name="status" value="confirmed" />
                            <button className="rounded-lg bg-emerald-100 px-3 py-1.5 text-xs font-semibold text-emerald-700 transition hover:bg-emerald-200 dark:bg-emerald-400/15 dark:text-emerald-300 dark:hover:bg-emerald-400/25">
                              {t.confirm_booking}
                            </button>
                          </form>
                          <form action={`/t/${slug}/services`} method="POST">
                            <input type="hidden" name="action" value="update-booking-status" />
                            <input type="hidden" name="bookingId" value={booking.id} />
                            <input type="hidden" name="status" value="cancelled" />
                            <button className="rounded-lg bg-red-50 px-3 py-1.5 text-xs font-semibold text-red-600 transition hover:bg-red-100 dark:bg-red-400/10 dark:text-red-300 dark:hover:bg-red-400/20">
                              {t.cancel_booking}
                            </button>
                          </form>
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              )}

              <div className="overflow-hidden rounded-2xl border border-gray-200 bg-white dark:border-white/10 dark:bg-white/4">
                <div className="flex items-center justify-between border-b border-gray-100 px-5 py-3.5 dark:border-white/8">
                  <h3 className="text-sm font-semibold text-gray-900 dark:text-white">{t.upcoming}</h3>
                  <Link href={`?tab=bookings`} className="text-xs text-gray-400 transition hover:text-gray-700 dark:text-slate-500 dark:hover:text-slate-300">
                    {t.all_bookings_link}
                  </Link>
                </div>
                {upcomingBookings.length === 0 ? (
                  <p className="px-5 py-10 text-center text-sm text-gray-400 dark:text-slate-500">
                    {t.no_upcoming}
                  </p>
                ) : (
                  <div className="divide-y divide-gray-100 dark:divide-white/8">
                    {upcomingBookings.map((booking) => (
                      <div key={booking.id} className="flex items-center gap-4 px-5 py-3.5">
                        <div className="flex h-10 w-10 flex-shrink-0 flex-col items-center justify-center rounded-xl bg-gray-100 text-center dark:bg-white/8">
                          <span className="text-[10px] font-medium uppercase text-gray-500 dark:text-slate-400">
                            {new Intl.DateTimeFormat(t.intl_locale, { month: "short", timeZone: workspace.tenant.timezone }).format(new Date(booking.startTime))}
                          </span>
                          <span className="text-sm font-bold leading-none text-gray-900 dark:text-white">
                            {new Intl.DateTimeFormat(t.intl_locale, { day: "numeric", timeZone: workspace.tenant.timezone }).format(new Date(booking.startTime))}
                          </span>
                        </div>
                        <div className="min-w-0 flex-1">
                          <p className="truncate text-sm font-medium text-gray-900 dark:text-white">{booking.customerName}</p>
                          <p className="mt-0.5 truncate text-xs text-gray-500 dark:text-slate-400">
                            {booking.serviceName} · {new Intl.DateTimeFormat(t.intl_locale, { hour: "numeric", minute: "2-digit", timeZone: workspace.tenant.timezone }).format(new Date(booking.startTime))}
                          </p>
                        </div>
                        <span className="flex-shrink-0 text-xs text-gray-400 dark:text-slate-500">
                          {booking.memberName.split(" ")[0]}
                        </span>
                      </div>
                    ))}
                  </div>
                )}
              </div>
            </div>

            <div className="space-y-4">
              <div className="rounded-2xl border border-gray-200 bg-white p-5 dark:border-white/10 dark:bg-white/4">
                <p className="text-xs font-semibold uppercase tracking-widest text-gray-400 dark:text-slate-500">{t.workspace}</p>
                <dl className="mt-4 space-y-3 text-sm">
                  {([
                    { label: t.timezone_label, value: workspace.tenant.timezone },
                    { label: t.currency_label, value: workspace.tenant.currency },
                    { label: t.locale_label, value: workspace.tenant.defaultLocale },
                  ] as const).map(({ label, value }) => (
                    <div key={label} className="flex items-center justify-between gap-4">
                      <dt className="text-gray-500 dark:text-slate-500">{label}</dt>
                      <dd className="font-medium text-gray-900 dark:text-white">{value}</dd>
                    </div>
                  ))}
                </dl>
              </div>

              <div className="grid grid-cols-2 gap-2">
                {([
                  { tab: "bookings",  label: t.nav_bookings,  badge: bookingSummary.pending > 0 ? bookingSummary.pending : null, iconPath: <path strokeLinecap="round" strokeLinejoin="round" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" /> },
                  { tab: "services",  label: t.nav_services,  badge: null, iconPath: <path strokeLinecap="round" strokeLinejoin="round" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" /> },
                  { tab: "team",      label: t.nav_team,      badge: null, iconPath: <path strokeLinecap="round" strokeLinejoin="round" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z" /> },
                  { tab: "settings",  label: t.nav_settings,  badge: null, iconPath: <path strokeLinecap="round" strokeLinejoin="round" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z M15 12a3 3 0 11-6 0 3 3 0 016 0z" /> },
                ] as const).map((item) => (
                  <Link
                    key={item.tab}
                    href={`?tab=${item.tab}`}
                    className="flex flex-col gap-3 rounded-xl border border-gray-200 bg-white p-4 transition hover:border-gray-300 hover:bg-gray-50 dark:border-white/10 dark:bg-white/4 dark:hover:border-white/20 dark:hover:bg-white/8"
                  >
                    <svg className="h-5 w-5 text-gray-400 dark:text-slate-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.75}>{item.iconPath}</svg>
                    <div className="flex items-center justify-between gap-1">
                      <span className="text-xs font-medium text-gray-700 dark:text-slate-300">{item.label}</span>
                      {item.badge !== null && (
                        <span className="rounded-full bg-amber-100 px-1.5 py-0.5 text-[10px] font-bold text-amber-700 dark:bg-amber-400/20 dark:text-amber-300">
                          {item.badge}
                        </span>
                      )}
                    </div>
                  </Link>
                ))}
              </div>
            </div>
          </div>
        </div>
      )}

      {/* ── BOOKINGS ── */}
      {tab === "bookings" && (
        <div className="space-y-5">
          <div className="flex flex-wrap items-center justify-between gap-4">
            <h1 className="text-2xl font-bold text-gray-900 dark:text-white">{t.nav_bookings}</h1>
            <DateRangeFilter defaultFrom={defaultFrom} defaultTo={defaultTo} tab="bookings" applyLabel={t.apply} />
          </div>
          <BookingOpsPanel
            bookings={bookings}
            tenantSlug={slug}
            timezone={workspace.tenant.timezone}
            currency={workspace.tenant.currency}
            locale={workspace.tenant.defaultLocale}
            variant="full"
          />
        </div>
      )}

      {/* ── SERVICES ── */}
      {tab === "services" && (
        <div className="space-y-5">
          <div className="flex flex-wrap items-center justify-between gap-4">
            <div>
              <h1 className="text-2xl font-bold text-gray-900 dark:text-white">{t.nav_services}</h1>
              <p className="mt-0.5 text-sm text-gray-500 dark:text-slate-400">
                {enabledServices} {t.live_lower} · {services.length - enabledServices} {t.hidden_lower}
              </p>
            </div>
            <Link
              href={`?tab=services&newService=1`}
              className="rounded-xl bg-gray-900 px-4 py-2.5 text-sm font-semibold text-white transition hover:bg-gray-700 dark:bg-white dark:text-slate-950 dark:hover:bg-slate-100"
            >
              + {t.create_service}
            </Link>
          </div>

          {query.newService === "1" && (
            <div className="overflow-hidden rounded-2xl border border-gray-200 bg-white shadow-sm dark:border-white/10 dark:bg-white/4">
              <div className="flex items-center justify-between border-b border-gray-100 px-6 py-4 dark:border-white/8">
                <h3 className="text-sm font-semibold text-gray-900 dark:text-white">{t.add_service_title}</h3>
                <Link href={`?tab=services`} className="text-gray-400 hover:text-gray-600 dark:text-slate-500 dark:hover:text-slate-300">
                  <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}><path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" /></svg>
                </Link>
              </div>
              <div className="p-6">
                <form action={`/t/${slug}/services`} method="POST" className="grid gap-4 md:grid-cols-2">
                  <input type="hidden" name="action" value="create-service" />
                  <FormGroup label={t.service_name} className="md:col-span-2">
                    <FormInput name="name" required placeholder={t.description_placeholder} />
                  </FormGroup>
                  <FormGroup label={t.description} className="md:col-span-2">
                    <textarea name="description" rows={2} className="w-full resize-none rounded-xl border border-gray-300 bg-white px-4 py-3 text-sm text-gray-900 outline-none transition placeholder:text-gray-400 focus:border-gray-500 focus:ring-2 focus:ring-gray-100 dark:border-white/10 dark:bg-slate-950 dark:text-white dark:placeholder:text-slate-600" placeholder={t.description_placeholder} />
                  </FormGroup>
                  <FormGroup label={t.duration_min}>
                    <FormInput name="durationMinutes" type="number" min={5} step={5} defaultValue={45} required />
                  </FormGroup>
                  <FormGroup label={t.buffer_before}>
                    <FormInput name="bufferBefore" type="number" min={0} step={5} defaultValue={0} />
                  </FormGroup>
                  <FormGroup label={t.buffer_after}>
                    <FormInput name="bufferAfter" type="number" min={0} step={5} defaultValue={0} />
                  </FormGroup>
                  <FormGroup label={t.visibility} className="flex items-end">
                    <label className="flex cursor-pointer items-center gap-3 rounded-xl border border-gray-200 bg-gray-50 px-4 py-3 text-sm text-gray-700 hover:bg-gray-100 dark:border-white/10 dark:bg-slate-950/50 dark:text-slate-300">
                      <input type="checkbox" name="enabled" defaultChecked className="h-4 w-4 rounded" />
                      {t.publish_immediately}
                    </label>
                  </FormGroup>
                  <div className="flex gap-3 md:col-span-2">
                    <button className="rounded-xl bg-gray-900 px-5 py-2.5 text-sm font-semibold text-white transition hover:bg-gray-700 active:scale-[0.98] dark:bg-white dark:text-slate-950 dark:hover:bg-slate-100">
                      {t.create_service}
                    </button>
                    <Link href={`?tab=services`} className="rounded-xl border border-gray-200 px-5 py-2.5 text-sm text-gray-600 transition hover:bg-gray-50 dark:border-white/10 dark:text-slate-400 dark:hover:bg-white/5">
                      {t.cancel}
                    </Link>
                  </div>
                </form>
              </div>
            </div>
          )}

          {services.length === 0 ? (
            <div className="flex flex-col items-center justify-center rounded-3xl border border-dashed border-gray-200 bg-white py-20 dark:border-white/10 dark:bg-white/3">
              <div className="mb-4 flex h-14 w-14 items-center justify-center rounded-2xl bg-gray-100 dark:bg-white/8">
                <svg className="h-7 w-7 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
                  <path strokeLinecap="round" strokeLinejoin="round" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
                </svg>
              </div>
              <p className="text-sm font-medium text-gray-600 dark:text-slate-400">{t.no_services_title}</p>
              <p className="mt-1 text-sm text-gray-400 dark:text-slate-500">{t.no_services_body}</p>
              <Link href={`?tab=services&newService=1`} className="mt-5 rounded-xl bg-gray-900 px-5 py-2.5 text-sm font-semibold text-white hover:bg-gray-700 dark:bg-white dark:text-slate-950">
                + {t.create_service}
              </Link>
            </div>
          ) : (
            <div className="space-y-4">
              {services.map((service) => {
                const assigned = serviceMembers.get(service.id) || [];
                const assignable = memberships.filter((m) => !assigned.some((a) => a.membershipId === m.id));
                return (
                  <ServiceCard
                    key={service.id}
                    service={service}
                    members={assigned}
                    memberships={assignable}
                    membershipById={membershipById}
                    openingHoursByMember={openingHoursByMember}
                    blockedDatesByMembership={blockedDatesByMembership}
                    copySources={assigned.map((m) => ({ id: m.id, name: m.memberName }))}
                    tenantSlug={slug}
                    currency={workspace.tenant.currency}
                    locale={workspace.tenant.defaultLocale}
                    t={t}
                  />
                );
              })}
            </div>
          )}
        </div>
      )}

      {/* ── TEAM ── */}
      {tab === "team" && (
        <div className="space-y-5">
          <div>
            <h1 className="text-2xl font-bold text-gray-900 dark:text-white">{t.nav_team}</h1>
            <p className="mt-0.5 text-sm text-gray-500 dark:text-slate-400">
              {memberships.length} {memberships.length === 1 ? t.member_singular : t.members_plural} · {assignedCount} {t.assigned_count}
            </p>
          </div>
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {memberships.map((membership) => {
              const memberServiceCount = serviceMembersEntries.filter(([, members]) =>
                members.some((m) => m.membershipId === membership.id)
              ).length;
              const role = membership.role;
              const roleBadge = role.includes("owner")
                ? "bg-purple-100 text-purple-700 dark:bg-purple-400/15 dark:text-purple-300"
                : role.includes("manager")
                  ? "bg-blue-100 text-blue-700 dark:bg-blue-400/15 dark:text-blue-300"
                  : "bg-teal-100 text-teal-700 dark:bg-teal-400/15 dark:text-teal-300";
              return (
                <div key={membership.id} className="rounded-2xl border border-gray-200 bg-white p-5 dark:border-white/10 dark:bg-white/4">
                  <div className="flex items-start gap-4">
                    <div className="flex h-11 w-11 flex-shrink-0 items-center justify-center rounded-full bg-gray-100 text-sm font-bold text-gray-600 dark:bg-white/10 dark:text-slate-300">
                      {membership.user.name.split(" ").map((n: string) => n[0]).slice(0, 2).join("").toUpperCase()}
                    </div>
                    <div className="min-w-0 flex-1">
                      <p className="truncate font-semibold text-gray-900 dark:text-white">{membership.user.name}</p>
                      <p className="truncate text-sm text-gray-500 dark:text-slate-400">{membership.user.email}</p>
                      <div className="mt-2.5 flex flex-wrap items-center gap-2">
                        <span className={`rounded-full px-2 py-0.5 text-[11px] font-medium ${roleBadge}`}>
                          {role.replace("tenant_", "")}
                        </span>
                        {memberServiceCount > 0 && (
                          <span className="text-xs text-gray-400 dark:text-slate-500">
                            {memberServiceCount} {memberServiceCount === 1 ? t.service_singular_lc : t.services_plural_lc}
                          </span>
                        )}
                      </div>
                    </div>
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      )}

      {/* ── CUSTOMERS ── */}
      {tab === "customers" && (
        <div className="space-y-5">
          <div className="flex flex-wrap items-center justify-between gap-4">
            <div>
              <h1 className="text-2xl font-bold text-gray-900 dark:text-white">{t.nav_customers}</h1>
              <p className="mt-0.5 text-sm text-gray-500 dark:text-slate-400">{t.top_guests_subtitle}</p>
            </div>
            <DateRangeFilter defaultFrom={defaultFrom} defaultTo={defaultTo} tab="customers" applyLabel={t.apply} />
          </div>
          <div className="grid gap-5 xl:grid-cols-[1fr_360px]">
            <div className="overflow-hidden rounded-2xl border border-gray-200 bg-white dark:border-white/10 dark:bg-white/4">
              {customers.length === 0 ? (
                <p className="px-6 py-14 text-center text-sm text-gray-400 dark:text-slate-500">{t.no_customers_yet}</p>
              ) : (
                <>
                  <div className="grid grid-cols-[1fr_64px_96px] border-b border-gray-100 px-5 py-3 text-xs font-medium uppercase tracking-wider text-gray-400 dark:border-white/8 dark:text-slate-500">
                    <span>Customer</span>
                    <span className="text-right">Visits</span>
                    <span className="text-right">Lifetime</span>
                  </div>
                  {customers.map((customer) => {
                    const selected = selectedCustomer?.email === customer.email;
                    const params = new URLSearchParams({ tab: "customers", customer: customer.email, ...(query.from ? { from: query.from } : {}), ...(query.to ? { to: query.to } : {}) });
                    return (
                      <a
                        key={customer.email}
                        href={`?${params.toString()}`}
                        className={`grid grid-cols-[1fr_64px_96px] items-center border-b border-gray-100 px-5 py-3.5 transition last:border-0 dark:border-white/8 ${selected ? "bg-blue-50 dark:bg-cyan-400/5" : "hover:bg-gray-50 dark:hover:bg-white/3"}`}
                      >
                        <div className="flex items-center gap-3 min-w-0">
                          <div className="flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-full bg-gray-200 text-xs font-bold text-gray-600 dark:bg-white/8 dark:text-slate-400">
                            {customer.name.split(" ").map((n: string) => n[0]).slice(0, 2).join("").toUpperCase()}
                          </div>
                          <div className="min-w-0">
                            <p className="truncate text-sm font-medium text-gray-900 dark:text-white">{customer.name}</p>
                            <p className="truncate text-xs text-gray-500 dark:text-slate-400">{customer.email}</p>
                          </div>
                        </div>
                        <span className="text-right text-sm font-medium text-gray-900 dark:text-white">{customer.visits}</span>
                        <span className="text-right text-sm text-gray-500 dark:text-slate-400">{formatMoney(customer.spent, workspace.tenant.currency, t.intl_locale)}</span>
                      </a>
                    );
                  })}
                </>
              )}
            </div>

            {selectedCustomer ? (
              <div className="rounded-2xl border border-gray-200 bg-white p-5 dark:border-white/10 dark:bg-white/4">
                <div className="flex items-center gap-3">
                  <div className="flex h-11 w-11 flex-shrink-0 items-center justify-center rounded-full bg-blue-100 text-sm font-bold text-blue-600 dark:bg-cyan-300/15 dark:text-cyan-300">
                    {selectedCustomer.name.split(" ").map((n: string) => n[0]).slice(0, 2).join("").toUpperCase()}
                  </div>
                  <div>
                    <p className="font-semibold text-gray-900 dark:text-white">{selectedCustomer.name}</p>
                    <p className="text-sm text-gray-500 dark:text-slate-400">{selectedCustomer.email}</p>
                  </div>
                </div>
                <div className="mt-4 grid grid-cols-2 gap-3">
                  <div className="rounded-xl border border-gray-100 bg-gray-50 p-4 text-center dark:border-white/8 dark:bg-white/4">
                    <p className="text-2xl font-bold text-gray-900 dark:text-white">{selectedCustomer.visits}</p>
                    <p className="mt-0.5 text-xs text-gray-400 dark:text-slate-500">{t.visits}</p>
                  </div>
                  <div className="rounded-xl border border-gray-100 bg-gray-50 p-4 text-center dark:border-white/8 dark:bg-white/4">
                    <p className="text-base font-bold text-gray-900 dark:text-white">{formatMoney(selectedCustomer.spent, workspace.tenant.currency, t.intl_locale)}</p>
                    <p className="mt-0.5 text-xs text-gray-400 dark:text-slate-500">{t.lifetime}</p>
                  </div>
                </div>
                <h4 className="mt-5 mb-3 text-sm font-semibold text-gray-900 dark:text-white">{t.booking_history}</h4>
                <div className="space-y-2">
                  {customerBookings.map((booking) => (
                    <div key={booking.id} className="rounded-xl border border-gray-100 bg-gray-50 px-4 py-3 dark:border-white/8 dark:bg-white/3">
                      <div className="flex items-center justify-between gap-2">
                        <p className="text-sm font-medium text-gray-900 dark:text-white">{booking.serviceName}</p>
                        <StatusBadge status={booking.status} t={t} />
                      </div>
                      <p className="mt-0.5 text-xs text-gray-500 dark:text-slate-400">
                        {formatDateTime(booking.startTime, workspace.tenant.timezone, t.intl_locale)}
                      </p>
                    </div>
                  ))}
                </div>
              </div>
            ) : (
              <div className="flex items-center justify-center rounded-2xl border border-dashed border-gray-200 bg-gray-50 p-10 text-center dark:border-white/10 dark:bg-white/2">
                <p className="text-sm text-gray-400 dark:text-slate-500">{t.select_customer_hint}</p>
              </div>
            )}
          </div>
        </div>
      )}

      {/* ── SETTINGS ── */}
      {tab === "settings" && (
        <div className="space-y-6">
          <div>
            <h1 className="text-2xl font-bold text-gray-900 dark:text-white">{t.nav_settings}</h1>
            <p className="mt-0.5 text-sm text-gray-400 dark:text-slate-500">/{workspace.tenant.slug}</p>
          </div>

          <form action={`/t/${slug}/services`} method="POST" className="space-y-4">
            <input type="hidden" name="action" value="update-tenant" />

            <SettingsSection title={t.settings_brand} description={t.settings_brand_desc}>
              <div className="grid gap-4 md:grid-cols-3">
                <FormGroup label={t.display_name} className="md:col-span-2">
                  <FormInput name="displayName" required defaultValue={workspace.tenant.displayName} />
                </FormGroup>
                <FormGroup label={t.logo_url}>
                  <FormInput name="logoUrl" defaultValue={workspace.tenant.logoUrl || ""} placeholder="https://…" />
                </FormGroup>
                <ColorFormGroup label={t.primary_color} hint={t.primary_color_hint}>
                  <ColorField name="colorPrimary" defaultValue={workspace.tenant.colorPrimary} />
                </ColorFormGroup>
                <ColorFormGroup label={t.secondary_color} hint={t.secondary_color_hint}>
                  <ColorField name="colorSecondary" defaultValue={workspace.tenant.colorSecondary} />
                </ColorFormGroup>
                <ColorFormGroup label={t.accent_color} hint={t.accent_color_hint}>
                  <ColorField name="colorAccent" defaultValue={workspace.tenant.colorAccent} />
                </ColorFormGroup>
              </div>
            </SettingsSection>

            <SettingsSection title={t.settings_contact} description={t.settings_contact_desc}>
              <div className="grid gap-4 md:grid-cols-3">
                <FormGroup label={t.contact_phone}>
                  <FormInput name="phone" defaultValue={workspace.tenant.phone || ""} placeholder="+1 555 0100" />
                </FormGroup>
                <FormGroup label={t.contact_email}>
                  <FormInput name="email" defaultValue={workspace.tenant.email || ""} placeholder="hello@example.com" />
                </FormGroup>
                <FormGroup label={t.address}>
                  <FormInput name="address" defaultValue={workspace.tenant.address || ""} placeholder="123 Main St, City" />
                </FormGroup>
              </div>
            </SettingsSection>

            <SettingsSection title={t.settings_preferences} description={t.settings_preferences_desc}>
              <div className="grid gap-4 md:grid-cols-3">
                <FormGroup label={t.currency}>
                  <select name="currency" defaultValue={workspace.tenant.currency} className="w-full rounded-xl border border-gray-300 bg-white px-4 py-2.5 text-sm text-gray-900 outline-none transition focus:border-gray-500 focus:ring-2 focus:ring-gray-100 dark:border-white/10 dark:bg-slate-950 dark:text-white">
                    <option value="USD">USD — US Dollar</option>
                    <option value="CAD">CAD — Canadian Dollar</option>
                    <option value="BRL">BRL — Brazilian Real</option>
                  </select>
                </FormGroup>
                <FormGroup label={t.timezone}>
                  <div className="relative">
                    <FormInput name="timezone" defaultValue={workspace.tenant.timezone} placeholder="Search timezone…" list="timezone-list" />
                    <datalist id="timezone-list">
                      {TIMEZONES.map((tz) => <option key={tz} value={tz} />)}
                    </datalist>
                  </div>
                </FormGroup>
                <FormGroup label={t.locale}>
                  <select name="defaultLocale" defaultValue={workspace.tenant.defaultLocale} className="w-full rounded-xl border border-gray-300 bg-white px-4 py-2.5 text-sm text-gray-900 outline-none transition focus:border-gray-500 focus:ring-2 focus:ring-gray-100 dark:border-white/10 dark:bg-slate-950 dark:text-white">
                    <option value="en">English (en)</option>
                    <option value="pt">Português (pt)</option>
                  </select>
                </FormGroup>
              </div>
            </SettingsSection>

            <SettingsSection title={t.booking_page} description={t.settings_booking_page_desc}>
              <label className="inline-flex cursor-pointer items-start gap-4 rounded-xl border border-gray-200 bg-gray-50 px-5 py-4 text-sm transition hover:bg-gray-100 dark:border-white/10 dark:bg-slate-950/50 dark:hover:bg-white/5">
                <input type="checkbox" name="enabled" defaultChecked={workspace.tenant.enabled} className="mt-0.5 h-4 w-4 rounded" />
                <div>
                  <p className="font-medium text-gray-800 dark:text-slate-200">{t.enabled_for_customers}</p>
                  <p className="mt-0.5 text-xs text-gray-400 dark:text-slate-500">{t.settings_booking_page_hint}</p>
                </div>
              </label>
            </SettingsSection>

            <div className="pt-2">
              <button className="rounded-xl bg-gray-900 px-6 py-2.5 text-sm font-semibold text-white transition hover:bg-gray-700 active:scale-[0.98] dark:bg-white dark:text-slate-950 dark:hover:bg-slate-100">
                {t.save_settings}
              </button>
            </div>
          </form>
        </div>
      )}
    </div>
  );
}

// ── Shared components ────────────────────────────────────────────────────────

function DateRangeFilter({ defaultFrom, defaultTo, tab, applyLabel }: { defaultFrom: string; defaultTo: string; tab: string; applyLabel: string }) {
  return (
    <form method="GET" className="flex items-center gap-2 rounded-xl border border-gray-200 bg-white px-3 py-2 shadow-sm dark:border-white/10 dark:bg-white/5">
      <input type="hidden" name="tab" value={tab} />
      <svg className="h-3.5 w-3.5 flex-shrink-0 text-gray-400 dark:text-slate-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
        <path strokeLinecap="round" strokeLinejoin="round" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
      </svg>
      <input type="date" name="from" defaultValue={defaultFrom} className="w-32 bg-transparent text-sm text-gray-700 outline-none dark:text-slate-300 dark:[color-scheme:dark]" />
      <span className="text-xs text-gray-300 dark:text-slate-700">–</span>
      <input type="date" name="to" defaultValue={defaultTo} className="w-32 bg-transparent text-sm text-gray-700 outline-none dark:text-slate-300 dark:[color-scheme:dark]" />
      <button type="submit" className="rounded-lg bg-gray-900 px-3 py-1.5 text-xs font-semibold text-white transition hover:bg-gray-700 dark:bg-white dark:text-slate-950 dark:hover:bg-slate-100">
        {applyLabel}
      </button>
    </form>
  );
}

function SettingsSection({ title, description, children }: { title: string; description: string; children: React.ReactNode }) {
  return (
    <div className="overflow-hidden rounded-2xl border border-gray-200 bg-white dark:border-white/10 dark:bg-white/4">
      <div className="border-b border-gray-100 px-6 py-4 dark:border-white/8">
        <p className="font-medium text-gray-900 dark:text-white">{title}</p>
        <p className="mt-0.5 text-xs text-gray-400 dark:text-slate-500">{description}</p>
      </div>
      <div className="px-6 py-5">{children}</div>
    </div>
  );
}

function MetricCard({ icon, label, value, detail }: { icon: string; label: string; value: string; detail: string }) {
  const configs: Record<string, { bg: string; darkBg: string; iconPath: React.ReactNode; color: string }> = {
    services: {
      bg: "bg-indigo-50", darkBg: "dark:bg-indigo-500/10",
      color: "text-indigo-600 dark:text-indigo-400",
      iconPath: <path strokeLinecap="round" strokeLinejoin="round" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />,
    },
    team: {
      bg: "bg-teal-50", darkBg: "dark:bg-teal-500/10",
      color: "text-teal-600 dark:text-teal-400",
      iconPath: <path strokeLinecap="round" strokeLinejoin="round" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z" />,
    },
    confirmed: {
      bg: "bg-emerald-50", darkBg: "dark:bg-emerald-500/10",
      color: "text-emerald-600 dark:text-emerald-400",
      iconPath: <path strokeLinecap="round" strokeLinejoin="round" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />,
    },
    pending: {
      bg: "bg-amber-50", darkBg: "dark:bg-amber-500/10",
      color: "text-amber-600 dark:text-amber-400",
      iconPath: <path strokeLinecap="round" strokeLinejoin="round" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />,
    },
  };
  const cfg = configs[icon];
  return (
    <div className="rounded-2xl border border-gray-200 bg-white p-5 dark:border-white/8 dark:bg-white/4">
      <div className={`flex h-9 w-9 items-center justify-center rounded-xl ${cfg.bg} ${cfg.darkBg}`}>
        <svg className={`h-[18px] w-[18px] ${cfg.color}`} fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
          {cfg.iconPath}
        </svg>
      </div>
      <p className="mt-4 text-3xl font-bold tracking-tight text-gray-900 dark:text-white">{value}</p>
      <p className="mt-1 text-xs font-semibold uppercase tracking-widest text-gray-400 dark:text-slate-500">{label}</p>
      <p className="mt-0.5 text-xs text-gray-400 dark:text-slate-600">{detail}</p>
    </div>
  );
}

function StatusBadge({ status, t }: { status: string; t: Translations }) {
  if (status === "confirmed") return <span className="rounded-full bg-emerald-100 px-2.5 py-0.5 text-[11px] font-semibold text-emerald-700 dark:bg-emerald-400/15 dark:text-emerald-300">{t.status_confirmed}</span>;
  if (status === "cancelled") return <span className="rounded-full bg-red-100 px-2.5 py-0.5 text-[11px] font-semibold text-red-700 dark:bg-red-400/15 dark:text-red-300">{t.status_cancelled}</span>;
  return <span className="rounded-full bg-amber-100 px-2.5 py-0.5 text-[11px] font-semibold text-amber-700 dark:bg-amber-400/15 dark:text-amber-300">{t.status_pending}</span>;
}

function FormGroup({ label, className, children }: { label: string; className?: string; children: React.ReactNode }) {
  return (
    <div className={className}>
      <label className="block text-sm font-medium text-gray-600 dark:text-slate-400">{label}</label>
      <div className="mt-1.5">{children}</div>
    </div>
  );
}

function ColorFormGroup({ label, hint, children }: { label: string; hint: string; children: React.ReactNode }) {
  return (
    <div>
      <div className="flex items-center gap-1.5">
        <span className="text-sm font-medium text-gray-600 dark:text-slate-400">{label}</span>
        <div className="group relative inline-flex">
          <svg className="h-3.5 w-3.5 cursor-help text-gray-400 dark:text-slate-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
            <path strokeLinecap="round" strokeLinejoin="round" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <div className="pointer-events-none absolute bottom-full left-1/2 z-10 mb-2 w-52 -translate-x-1/2 rounded-xl bg-gray-900 px-3 py-2 text-xs leading-relaxed text-white opacity-0 transition-opacity group-hover:opacity-100 dark:bg-zinc-700">
            {hint}
            <div className="absolute left-1/2 top-full -translate-x-1/2 border-4 border-transparent border-t-gray-900 dark:border-t-zinc-700" />
          </div>
        </div>
      </div>
      <div className="mt-1.5">{children}</div>
    </div>
  );
}

function FormInput({ className, ...props }: React.InputHTMLAttributes<HTMLInputElement>) {
  return (
    <input
      {...props}
      className={`w-full rounded-xl border border-gray-300 bg-white px-4 py-2.5 text-sm text-gray-900 outline-none transition placeholder:text-gray-400 focus:border-gray-500 focus:ring-2 focus:ring-gray-100 dark:border-white/10 dark:bg-slate-950 dark:text-white dark:placeholder:text-slate-600 dark:focus:border-white/25 dark:focus:ring-white/5 ${className ?? ""}`}
    />
  );
}

function ServiceCard({
  service, members, memberships, membershipById,
  openingHoursByMember, blockedDatesByMembership,
  copySources, tenantSlug, currency, locale, t,
}: {
  service: Service;
  members: ServiceMemberDetails[];
  memberships: MembershipDetails[];
  membershipById: Map<string, MembershipDetails>;
  openingHoursByMember: Map<string, OpeningHours[]>;
  blockedDatesByMembership: Map<string, BlockedDateDetails[]>;
  copySources: Array<{ id: string; name: string }>;
  tenantSlug: string;
  currency: string;
  locale: string;
  t: Translations;
}) {
  return (
    <article className="overflow-hidden rounded-2xl border border-gray-200 bg-white dark:border-white/8 dark:bg-white/4">
      <div className="flex flex-wrap items-start justify-between gap-4 border-b border-gray-100 px-6 py-4 dark:border-white/8">
        <div className="flex items-start gap-4">
          <div className="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-xl bg-gray-100 text-sm font-bold text-gray-600 dark:bg-white/8 dark:text-slate-300">
            {service.name.slice(0, 2).toUpperCase()}
          </div>
          <div>
            <div className="flex items-center gap-2.5">
              <h4 className="text-base font-semibold text-gray-900 dark:text-white">{service.name}</h4>
              <StatusBadge status={service.enabled ? "confirmed" : "cancelled"} t={t} />
            </div>
            {service.description && (
              <p className="mt-0.5 max-w-xl text-sm text-gray-500 dark:text-slate-400">{service.description}</p>
            )}
            <div className="mt-1.5 flex flex-wrap items-center gap-3 text-xs text-gray-400 dark:text-slate-500">
              <span>{service.durationMinutes} min</span>
              {service.bufferBefore > 0 && <span>+{service.bufferBefore} min before</span>}
              {service.bufferAfter > 0 && <span>+{service.bufferAfter} min after</span>}
              <span>{members.length} {members.length === 1 ? t.member_singular : t.members_plural}</span>
            </div>
          </div>
        </div>
        <form action={`/t/${tenantSlug}/services`} method="POST">
          <input type="hidden" name="action" value="toggle-service" />
          <input type="hidden" name="serviceId" value={service.id} />
          <input type="hidden" name="enabled" value={service.enabled ? "false" : "true"} />
          <button className="rounded-xl border border-gray-200 bg-gray-50 px-4 py-2 text-sm text-gray-700 transition hover:bg-gray-100 dark:border-white/10 dark:bg-slate-950/60 dark:text-white dark:hover:bg-white/8">
            {service.enabled ? t.hide_service : t.publish_service}
          </button>
        </form>
      </div>

      <div className="space-y-5 p-6">
        <div className="grid gap-5 lg:grid-cols-2">
          <div>
            <p className="mb-3 text-sm font-semibold text-gray-900 dark:text-white">
              {t.assigned_specialists}
              <span className="ml-2 text-xs font-normal text-gray-400 dark:text-slate-500">{members.length} {t.active_lower}</span>
            </p>
            <div className="space-y-2">
              {members.length === 0 ? (
                <p className="rounded-xl border border-dashed border-gray-200 p-4 text-sm text-gray-400 dark:border-white/10 dark:text-slate-500">{t.no_specialists_yet}</p>
              ) : (
                members.map((member) => {
                  const membership = membershipById.get(member.membershipId);
                  return (
                    <div key={member.id} className="flex items-center gap-3 rounded-xl border border-gray-100 bg-gray-50 px-4 py-3 dark:border-white/8 dark:bg-slate-950/40">
                      <div className="flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-full bg-gray-200 text-xs font-bold text-gray-600 dark:bg-white/8 dark:text-slate-300">
                        {member.memberName.split(" ").map((n: string) => n[0]).slice(0, 2).join("").toUpperCase()}
                      </div>
                      <div className="min-w-0 flex-1">
                        <p className="truncate text-sm font-medium text-gray-900 dark:text-white">{member.memberName}</p>
                        <p className="text-xs text-gray-500 dark:text-slate-500">
                          {new Intl.NumberFormat(locale === "pt" ? "pt-BR" : "en-US", { style: "currency", currency }).format(member.priceCents / 100)}
                          {membership?.role ? ` · ${membership.role.replace("tenant_", "")}` : ""}
                        </p>
                      </div>
                      <form action={`/t/${tenantSlug}/services`} method="POST">
                        <input type="hidden" name="action" value="remove-member" />
                        <input type="hidden" name="serviceId" value={service.id} />
                        <input type="hidden" name="serviceMemberId" value={member.id} />
                        <button className="rounded-lg border border-gray-200 px-2.5 py-1.5 text-xs text-gray-500 transition hover:bg-gray-100 dark:border-white/10 dark:text-slate-400 dark:hover:bg-white/5">
                          {t.remove}
                        </button>
                      </form>
                    </div>
                  );
                })
              )}
            </div>
          </div>

          <div className="rounded-xl border border-gray-100 bg-gray-50 p-4 dark:border-white/8 dark:bg-slate-950/40">
            <p className="mb-3 text-sm font-semibold text-gray-900 dark:text-white">{t.add_specialist}</p>
            <form action={`/t/${tenantSlug}/services`} method="POST" className="space-y-3">
              <input type="hidden" name="action" value="assign-member" />
              <input type="hidden" name="serviceId" value={service.id} />
              <FormGroup label={t.team_member}>
                <select name="membershipId" required disabled={memberships.length === 0} defaultValue="" className="w-full rounded-xl border border-gray-300 bg-white px-4 py-2.5 text-sm text-gray-900 outline-none transition focus:border-gray-500 disabled:cursor-not-allowed disabled:opacity-50 dark:border-white/10 dark:bg-slate-950 dark:text-white">
                  <option value="" disabled>{memberships.length === 0 ? t.all_members_assigned : t.choose_team_member}</option>
                  {memberships.map((m) => <option key={m.id} value={m.id}>{m.user.name} · {m.role.replace("tenant_", "")}</option>)}
                </select>
              </FormGroup>
              <FormGroup label={t.price_label}>
                <CurrencyInput defaultValueCents={4500} currency={currency} locale={locale} />
              </FormGroup>
              <FormGroup label={t.internal_note}>
                <FormInput name="description" placeholder={t.internal_note_placeholder} />
              </FormGroup>
              <button disabled={memberships.length === 0} className="w-full rounded-xl bg-gray-900 px-4 py-2.5 text-sm font-semibold text-white transition hover:bg-gray-700 active:scale-[0.98] disabled:cursor-not-allowed disabled:opacity-40 dark:bg-white dark:text-slate-950 dark:hover:bg-slate-100">
                {t.add_specialist}
              </button>
            </form>
          </div>
        </div>

        <details className="group rounded-xl border border-gray-200 bg-gray-50 dark:border-white/8 dark:bg-slate-950/40">
          <summary className="flex cursor-pointer list-none items-center justify-between px-5 py-3.5 text-sm font-semibold text-gray-900 dark:text-white">
            {t.edit_service_details}
            <svg className="h-4 w-4 text-gray-400 transition group-open:rotate-180 dark:text-slate-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M19 9l-7 7-7-7" />
            </svg>
          </summary>
          <div className="border-t border-gray-200 p-5 dark:border-white/8">
            <form action={`/t/${tenantSlug}/services`} method="POST" className="grid gap-4 md:grid-cols-2">
              <input type="hidden" name="action" value="update-service" />
              <input type="hidden" name="serviceId" value={service.id} />
              <FormGroup label={t.service_name} className="md:col-span-2"><FormInput name="name" required defaultValue={service.name} /></FormGroup>
              <FormGroup label={t.description} className="md:col-span-2">
                <textarea name="description" rows={2} defaultValue={service.description || ""} className="w-full resize-none rounded-xl border border-gray-300 bg-white px-4 py-3 text-sm text-gray-900 outline-none transition focus:border-gray-500 dark:border-white/10 dark:bg-slate-950 dark:text-white" />
              </FormGroup>
              <FormGroup label={t.duration_min}><FormInput name="durationMinutes" type="number" min={5} step={5} defaultValue={service.durationMinutes} /></FormGroup>
              <FormGroup label={t.buffer_before}><FormInput name="bufferBefore" type="number" min={0} step={5} defaultValue={service.bufferBefore} /></FormGroup>
              <FormGroup label={t.buffer_after}><FormInput name="bufferAfter" type="number" min={0} step={5} defaultValue={service.bufferAfter} /></FormGroup>
              <FormGroup label={t.visibility} className="flex items-end">
                <label className="flex cursor-pointer items-center gap-3 rounded-xl border border-gray-200 bg-white px-4 py-3 text-sm text-gray-700 dark:border-white/10 dark:bg-slate-950/50 dark:text-slate-300">
                  <input type="checkbox" name="enabled" defaultChecked={service.enabled} className="h-4 w-4 rounded" />
                  {t.visible_on_booking_page}
                </label>
              </FormGroup>
              <div className="flex flex-wrap gap-3 md:col-span-2">
                <button className="rounded-xl bg-gray-900 px-5 py-2.5 text-sm font-semibold text-white transition hover:bg-gray-700 dark:bg-white dark:text-slate-950 dark:hover:bg-slate-100">{t.save_changes}</button>
              </div>
            </form>
            <div className="mt-4 flex flex-wrap gap-3 border-t border-gray-200 pt-4 dark:border-white/8">
              <form action={`/t/${tenantSlug}/services`} method="POST">
                <input type="hidden" name="action" value="duplicate-service" />
                <input type="hidden" name="serviceId" value={service.id} />
                <button className="rounded-xl border border-gray-200 bg-white px-4 py-2 text-sm text-gray-700 transition hover:bg-gray-50 dark:border-white/10 dark:bg-white/5 dark:text-slate-300">{t.duplicate}</button>
              </form>
              <form action={`/t/${tenantSlug}/services`} method="POST">
                <input type="hidden" name="action" value="delete-service" />
                <input type="hidden" name="serviceId" value={service.id} />
                <button className="rounded-xl border border-red-200 bg-red-50 px-4 py-2 text-sm font-medium text-red-600 transition hover:bg-red-100 dark:border-red-400/20 dark:bg-red-400/8 dark:text-red-300">{t.delete_service}</button>
              </form>
            </div>
          </div>
        </details>

        {members.length > 0 && (
          <div>
            <p className="mb-3 text-sm font-semibold text-gray-900 dark:text-white">{t.calendar_controls}</p>
            <div className="space-y-4">
              {members.map((member) => (
                <InteractiveScheduleEditor
                  key={member.id}
                  tenantSlug={tenantSlug}
                  member={member}
                  openingHours={openingHoursByMember.get(member.id) || []}
                  blockedDates={blockedDatesByMembership.get(member.membershipId) || []}
                  copySources={copySources.filter((c) => c.id !== member.id)}
                  locale={locale as "en" | "pt"}
                />
              ))}
            </div>
          </div>
        )}
      </div>
    </article>
  );
}
