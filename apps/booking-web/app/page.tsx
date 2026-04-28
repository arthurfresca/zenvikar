import { headers } from "next/headers";
import { cookies } from "next/headers";
import { resolveTenant } from "@/lib/tenant";
import { fetchServerApi } from "@/lib/server-api";
import { getTranslations } from "@/lib/i18n";
import type { Booking, Service, ServiceMember, Tenant } from "@zenvikar/types";
import { LogoutButton } from "./components/logout-button";

const TOKEN_COOKIE = "zenvikar_booking_token";

type BookingSession = {
  userId: string;
  email: string;
  name: string;
};

type PublicServiceMember = ServiceMember & {
  memberName: string;
  memberEmail: string;
  tenantId: string;
};

type PublicService = Service & {
  members: PublicServiceMember[];
};

type PageSearchParams = Promise<{
  member?: string;
  date?: string;
  booking?: string;
  bookingError?: string;
}>;

async function loadSession(token: string) {
  const res = await fetchServerApi({
    path: "/api/v1/auth/me",
    headers: { Authorization: `Bearer ${token}` },
    cache: "no-store",
  });
  if (!res.ok) return null;
  return res.json();
}

async function loadPublicServices(tenantSlug: string): Promise<PublicService[]> {
  const res = await fetchServerApi({
    path: `/api/v1/tenants/${encodeURIComponent(tenantSlug)}/services`,
    next: { revalidate: 300 },
  });
  if (!res.ok) return [];
  const data = (await res.json()) as { services?: PublicService[] };
  return data.services || [];
}

async function loadMyBookings(token: string): Promise<Booking[]> {
  const res = await fetchServerApi({
    path: "/api/v1/me/bookings",
    headers: { Authorization: `Bearer ${token}` },
    cache: "no-store",
  });
  if (!res.ok) return [];
  const data = await res.json();
  return Array.isArray(data?.bookings) ? data.bookings : [];
}

async function loadAvailability(tenantSlug: string, serviceMemberId: string, date: string) {
  const res = await fetchServerApi({
    path: `/api/v1/tenants/${encodeURIComponent(tenantSlug)}/service-members/${serviceMemberId}/availability?date=${encodeURIComponent(date)}`,
    next: { revalidate: 60 },
  });
  if (!res.ok) return [] as Array<{ startTime: string; endTime: string }>;
  const data = (await res.json()) as { times?: Array<{ startTime: string; endTime: string }> };
  return data.times || [];
}

function formatMoney(cents: number, currency: string) {
  return new Intl.NumberFormat("en-US", { style: "currency", currency }).format(cents / 100);
}

function formatTime(value: string, timezone: string) {
  return new Intl.DateTimeFormat("en-US", {
    hour: "numeric",
    minute: "2-digit",
    timeZone: timezone,
  }).format(new Date(value));
}

function getInitialDate(value?: string) {
  if (value) return value;
  return new Date().toISOString().slice(0, 10);
}

function formatDateTime(value: string, timezone: string) {
  return new Intl.DateTimeFormat("en-US", {
    weekday: "short",
    month: "short",
    day: "numeric",
    hour: "numeric",
    minute: "2-digit",
    timeZone: timezone,
  }).format(new Date(value));
}

function withAlpha(hex: string, alpha: string) {
  return `${hex}${alpha}`;
}

function getContrastColor(hex: string) {
  const normalized = hex.replace("#", "");
  const value =
    normalized.length === 3
      ? normalized.split("").map((char) => char + char).join("")
      : normalized;
  const r = parseInt(value.slice(0, 2), 16);
  const g = parseInt(value.slice(2, 4), 16);
  const b = parseInt(value.slice(4, 6), 16);
  const brightness = (r * 299 + g * 587 + b * 114) / 1000;
  return brightness > 155 ? "#111827" : "#ffffff";
}

function getInitials(name: string) {
  return name
    .split(" ")
    .map((n) => n[0])
    .slice(0, 2)
    .join("")
    .toUpperCase();
}

function statusStyle(status: string) {
  if (status === "confirmed") return { bg: "#d1fae5", text: "#065f46" };
  if (status === "cancelled") return { bg: "#fee2e2", text: "#991b1b" };
  return { bg: "#fef3c7", text: "#92400e" };
}

export default async function BookingPage({
  searchParams,
}: {
  searchParams: PageSearchParams;
}) {
  const headersList = await headers();
  const cookieStore = await cookies();
  const params = await searchParams;
  const host = headersList.get("x-tenant-host") || headersList.get("host") || "";
  const token = cookieStore.get(TOKEN_COOKIE)?.value;
  let session = token
    ? ((await loadSession(token)) as (BookingSession & { currentTenantSlug?: string }) | null)
    : null;

  let tenant: Tenant;
  try {
    tenant = await resolveTenant(host);
  } catch {
    const t = getTranslations("en");
    return (
      <main className="flex min-h-screen flex-col items-center justify-center bg-gray-50 p-8">
        <div className="rounded-3xl border border-gray-200 bg-white p-12 text-center shadow-sm">
          <div className="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-2xl bg-red-100">
            <svg className="h-6 w-6 text-red-600" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
            </svg>
          </div>
          <h1 className="text-xl font-semibold text-gray-900">{t.workspace_not_found}</h1>
          <p className="mt-2 text-sm text-gray-500">{t.workspace_not_found_desc} {host}</p>
        </div>
      </main>
    );
  }

  const t = getTranslations(tenant.defaultLocale);

  const sessionScopedElsewhere = Boolean(
    session?.currentTenantSlug && session.currentTenantSlug !== tenant.slug
  );
  if (sessionScopedElsewhere) session = null;

  const services = await loadPublicServices(tenant.slug);
  const myBookings = session && token ? await loadMyBookings(token) : [];
  const members = services.flatMap((service) =>
    service.members.map((member) => ({
      ...member,
      serviceName: service.name,
      durationMinutes: service.durationMinutes,
    }))
  );
  const selectedDate = getInitialDate(params.date);
  const selectedMember = members.find((member) => member.id === params.member) || members[0] || null;
  const times = selectedMember
    ? await loadAvailability(tenant.slug, selectedMember.id, selectedDate)
    : [];
  const returnTo = `/${selectedMember ? `?member=${encodeURIComponent(selectedMember.id)}&date=${encodeURIComponent(selectedDate)}` : ""}`;
  const loginHref = `/login?${new URLSearchParams({
    reauth: sessionScopedElsewhere ? "1" : "0",
    next: returnTo || "/",
    locale: tenant.defaultLocale,
  }).toString()}`;
  const primaryTextColor = getContrastColor(tenant.colorPrimary);
  const accentTextColor = getContrastColor(tenant.colorAccent);

  return (
    <main className="min-h-screen bg-gray-50">
      {/* Hero */}
      <div
        className="relative overflow-hidden"
        style={{
          background: `linear-gradient(135deg, ${tenant.colorPrimary} 0%, ${tenant.colorSecondary} 100%)`,
        }}
      >
        <div
          className="absolute inset-0 opacity-20"
          style={{
            backgroundImage: `radial-gradient(circle at 20% 80%, ${tenant.colorAccent} 0%, transparent 50%), radial-gradient(circle at 80% 20%, #ffffff 0%, transparent 50%)`,
          }}
        />
        <div className="relative mx-auto flex max-w-6xl items-start justify-between gap-6 px-6 py-10 sm:items-center">
          <div className="flex items-start gap-4 sm:items-center">
            {tenant.logoUrl ? (
              <img
                src={tenant.logoUrl}
                alt={`${tenant.displayName} logo`}
                className="h-14 w-14 flex-shrink-0 rounded-2xl border-2 bg-white/90 object-cover p-1.5 shadow-md"
                style={{ borderColor: withAlpha(primaryTextColor, "30") }}
              />
            ) : (
              <div
                className="flex h-14 w-14 flex-shrink-0 items-center justify-center rounded-2xl text-xl font-bold shadow-md"
                style={{ backgroundColor: withAlpha(primaryTextColor, "18"), color: primaryTextColor }}
              >
                {getInitials(tenant.displayName)}
              </div>
            )}
            <div style={{ color: primaryTextColor }}>
              <p className="text-xs font-semibold uppercase tracking-widest opacity-70">{tenant.slug}</p>
              <h1 className="mt-0.5 text-2xl font-bold sm:text-3xl">{tenant.displayName}</h1>
              <div className="mt-2 flex flex-wrap items-center gap-3 text-xs opacity-75">
                {tenant.phone ? (
                  <span className="flex items-center gap-1">
                    <svg className="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                      <path strokeLinecap="round" strokeLinejoin="round" d="M3 5a2 2 0 012-2h3.28a1 1 0 01.948.684l1.498 4.493a1 1 0 01-.502 1.21l-2.257 1.13a11.042 11.042 0 005.516 5.516l1.13-2.257a1 1 0 011.21-.502l4.493 1.498a1 1 0 01.684.949V19a2 2 0 01-2 2h-1C9.716 21 3 14.284 3 6V5z" />
                    </svg>
                    {tenant.phone}
                  </span>
                ) : null}
                {tenant.email ? (
                  <span className="flex items-center gap-1">
                    <svg className="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                      <path strokeLinecap="round" strokeLinejoin="round" d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
                    </svg>
                    {tenant.email}
                  </span>
                ) : null}
                {tenant.address ? (
                  <span className="flex items-center gap-1">
                    <svg className="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                      <path strokeLinecap="round" strokeLinejoin="round" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" />
                      <path strokeLinecap="round" strokeLinejoin="round" d="M15 11a3 3 0 11-6 0 3 3 0 016 0z" />
                    </svg>
                    {tenant.address}
                  </span>
                ) : null}
              </div>
            </div>
          </div>
          <div className="flex flex-shrink-0 items-center gap-3">
            {session ? (
              <div className="flex items-center gap-3">
                <div
                  className="hidden rounded-full px-3 py-1.5 text-xs font-medium sm:block"
                  style={{ backgroundColor: withAlpha(primaryTextColor, "18"), color: primaryTextColor }}
                >
                  {session.name}
                </div>
                <LogoutButton />
              </div>
            ) : (
              <a
                href={loginHref}
                className="flex items-center gap-2 rounded-full px-5 py-2.5 text-sm font-semibold shadow-md transition hover:shadow-lg active:scale-[0.98]"
                style={{ backgroundColor: tenant.colorAccent, color: accentTextColor }}
              >
                <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                  <path strokeLinecap="round" strokeLinejoin="round" d="M11 16l-4-4m0 0l4-4m-4 4h14m-5 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h7a3 3 0 013 3v1" />
                </svg>
                {t.sign_in_to_book}
              </a>
            )}
          </div>
        </div>
      </div>

      <div className="mx-auto max-w-6xl space-y-8 px-6 py-8">
        {/* Notifications */}
        {params.booking === "created" ? (
          <div
            className="flex items-center gap-3 rounded-2xl border p-4 text-sm font-medium"
            style={{
              borderColor: withAlpha(tenant.colorAccent, "40"),
              backgroundColor: withAlpha(tenant.colorAccent, "12"),
              color: tenant.colorPrimary,
            }}
          >
            <svg className="h-5 w-5 flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            {t.booking_success}
          </div>
        ) : null}
        {params.bookingError ? (
          <div className="flex items-center gap-3 rounded-2xl border border-red-200 bg-red-50 p-4 text-sm font-medium text-red-700">
            <svg className="h-5 w-5 flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            {t.booking_error}
          </div>
        ) : null}

        {/* My upcoming visits */}
        {session && myBookings.length > 0 ? (
          <section>
            <div className="mb-4 flex items-center justify-between gap-4">
              <h2 className="text-lg font-semibold text-gray-900">{t.upcoming_visits}</h2>
              <span
                className="rounded-full px-3 py-1 text-xs font-medium"
                style={{ backgroundColor: withAlpha(tenant.colorAccent, "18"), color: tenant.colorPrimary }}
              >
                {myBookings.length} {t.total}
              </span>
            </div>
            <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-3">
              {myBookings.slice(0, 6).map((booking) => {
                const style = statusStyle(booking.status);
                const statusLabel =
                  booking.status === "confirmed" ? t.status_confirmed :
                  booking.status === "cancelled" ? t.status_cancelled :
                  t.status_pending;
                return (
                  <article
                    key={booking.id}
                    className="flex flex-col gap-3 rounded-2xl border border-gray-200 bg-white p-5 shadow-sm"
                  >
                    <div className="flex items-start justify-between gap-3">
                      <p className="text-sm font-semibold text-gray-900">
                        {formatDateTime(booking.startTime, tenant.timezone)}
                      </p>
                      <span
                        className="flex-shrink-0 rounded-full px-2.5 py-1 text-xs font-semibold"
                        style={{ backgroundColor: style.bg, color: style.text }}
                      >
                        {statusLabel}
                      </span>
                    </div>
                    <div className="flex items-center justify-between gap-3 border-t border-gray-100 pt-3">
                      <p className="text-sm font-semibold" style={{ color: tenant.colorPrimary }}>
                        {formatMoney(booking.priceCents, tenant.currency)}
                      </p>
                      <p className="text-xs text-gray-500">{t.ends} {formatDateTime(booking.endTime, tenant.timezone)}</p>
                    </div>
                  </article>
                );
              })}
            </div>
          </section>
        ) : null}

        {/* Main booking section */}
        <div className="grid gap-6 lg:grid-cols-[1fr_380px]">
          {/* Services */}
          <section>
            <h2 className="mb-4 text-lg font-semibold text-gray-900">
              {selectedMember ? t.change_specialist : t.choose_specialist}
            </h2>
            {services.length === 0 ? (
              <div className="rounded-3xl border border-dashed border-gray-300 bg-white p-10 text-center">
                <div className="mx-auto mb-3 flex h-10 w-10 items-center justify-center rounded-xl bg-gray-100">
                  <svg className="h-5 w-5 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                    <path strokeLinecap="round" strokeLinejoin="round" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
                  </svg>
                </div>
                <p className="text-sm text-gray-500">{t.no_services}</p>
              </div>
            ) : (
              <div className="space-y-4">
                {services.map((service) => (
                  <article
                    key={service.id}
                    className="overflow-hidden rounded-3xl border border-gray-200 bg-white shadow-sm"
                  >
                    <div className="border-b border-gray-100 px-6 py-5">
                      <div className="flex items-start justify-between gap-4">
                        <div>
                          <h3 className="font-semibold text-gray-900">{service.name}</h3>
                          {service.description ? (
                            <p className="mt-1 text-sm text-gray-500">{service.description}</p>
                          ) : null}
                        </div>
                        <span
                          className="flex-shrink-0 rounded-full px-3 py-1 text-xs font-semibold"
                          style={{
                            backgroundColor: withAlpha(tenant.colorAccent, "18"),
                            color: tenant.colorPrimary,
                          }}
                        >
                          {service.durationMinutes} {t.duration_min}
                        </span>
                      </div>
                    </div>
                    <div className="grid gap-3 p-4 sm:grid-cols-2">
                      {service.members.map((member) => {
                        const selected = selectedMember?.id === member.id;
                        return (
                          <a
                            key={member.id}
                            href={`/?member=${encodeURIComponent(member.id)}&date=${encodeURIComponent(selectedDate)}`}
                            className="group flex items-start gap-3 rounded-2xl border p-4 transition"
                            style={
                              selected
                                ? {
                                    borderColor: tenant.colorPrimary,
                                    backgroundColor: tenant.colorPrimary,
                                    color: primaryTextColor,
                                  }
                                : {
                                    borderColor: "#e5e7eb",
                                    backgroundColor: "#f9fafb",
                                    color: "#374151",
                                  }
                            }
                          >
                            <div
                              className="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-xl text-sm font-bold"
                              style={
                                selected
                                  ? { backgroundColor: withAlpha(primaryTextColor, "22"), color: primaryTextColor }
                                  : { backgroundColor: withAlpha(tenant.colorPrimary, "14"), color: tenant.colorPrimary }
                              }
                            >
                              {getInitials(member.memberName)}
                            </div>
                            <div className="min-w-0 flex-1">
                              <p className="truncate font-medium">{member.memberName}</p>
                              <p
                                className="mt-0.5 text-sm font-semibold"
                                style={{ color: selected ? withAlpha(primaryTextColor, "CC") : tenant.colorSecondary }}
                              >
                                {formatMoney(member.priceCents, tenant.currency)}
                              </p>
                              {member.description ? (
                                <p
                                  className="mt-1 text-xs"
                                  style={{ color: selected ? withAlpha(primaryTextColor, "99") : "#9ca3af" }}
                                >
                                  {member.description}
                                </p>
                              ) : null}
                            </div>
                            {selected ? (
                              <svg className="mt-0.5 h-4 w-4 flex-shrink-0" style={{ color: withAlpha(primaryTextColor, "BB") }} fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2.5}>
                                <path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7" />
                              </svg>
                            ) : null}
                          </a>
                        );
                      })}
                    </div>
                  </article>
                ))}
              </div>
            )}
          </section>

          {/* Availability */}
          <aside>
            <div className="sticky top-6 rounded-3xl border border-gray-200 bg-white p-6 shadow-sm">
              {!selectedMember ? (
                <div className="flex flex-col items-center py-6 text-center">
                  <div
                    className="mb-4 flex h-12 w-12 items-center justify-center rounded-2xl"
                    style={{ backgroundColor: withAlpha(tenant.colorAccent, "15") }}
                  >
                    <svg className="h-6 w-6" style={{ color: tenant.colorPrimary }} fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                      <path strokeLinecap="round" strokeLinejoin="round" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
                    </svg>
                  </div>
                  <p className="font-medium text-gray-900">{t.select_specialist}</p>
                  <p className="mt-1 text-sm text-gray-500">{t.choose_provider}</p>
                </div>
              ) : (
                <>
                  {/* Selected specialist info */}
                  <div
                    className="mb-5 flex items-center gap-3 rounded-2xl p-4"
                    style={{ backgroundColor: withAlpha(tenant.colorAccent, "12") }}
                  >
                    <div
                      className="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-xl text-sm font-bold"
                      style={{ backgroundColor: withAlpha(tenant.colorPrimary, "14"), color: tenant.colorPrimary }}
                    >
                      {getInitials(selectedMember.memberName)}
                    </div>
                    <div>
                      <p className="font-semibold" style={{ color: tenant.colorPrimary }}>{selectedMember.memberName}</p>
                      <p className="text-sm" style={{ color: tenant.colorSecondary }}>
                        {selectedMember.serviceName} · {formatMoney(selectedMember.priceCents, tenant.currency)}
                      </p>
                    </div>
                  </div>

                  {/* Date picker */}
                  <form method="GET" className="mb-5">
                    <input type="hidden" name="member" value={selectedMember.id} />
                    <label className="block text-sm font-medium text-gray-700">
                      {t.pick_date}
                    </label>
                    <div className="mt-2 flex gap-2">
                      <input
                        type="date"
                        name="date"
                        defaultValue={selectedDate}
                        className="flex-1 rounded-xl border border-gray-300 px-4 py-2.5 text-sm text-gray-900 outline-none transition focus:border-gray-500 focus:ring-4 focus:ring-gray-100"
                        style={{ colorScheme: "light" }}
                      />
                      <button
                        type="submit"
                        className="flex-shrink-0 rounded-xl px-4 py-2.5 text-sm font-semibold transition hover:opacity-90 active:scale-[0.97]"
                        style={{ backgroundColor: tenant.colorPrimary, color: primaryTextColor }}
                      >
                        {t.check}
                      </button>
                    </div>
                  </form>

                  {/* Available times */}
                  <div>
                    <p className="mb-3 text-sm font-medium text-gray-700">
                      {t.available_times}
                      {times.length > 0 ? (
                        <span className="ml-2 rounded-full bg-gray-100 px-2 py-0.5 text-xs font-medium text-gray-500">
                          {times.length}
                        </span>
                      ) : null}
                    </p>

                    {times.length === 0 ? (
                      <div className="rounded-2xl border border-dashed border-gray-200 bg-gray-50 px-4 py-6 text-center">
                        <p className="text-sm text-gray-500">{t.no_times}</p>
                        <p className="mt-1 text-xs text-gray-400">{t.try_different_day}</p>
                      </div>
                    ) : (
                      <div className="grid grid-cols-2 gap-2">
                        {times.map((time) => (
                          <form key={time.startTime} action="/book" method="POST">
                            <input type="hidden" name="tenantSlug" value={tenant.slug} />
                            <input type="hidden" name="serviceMemberId" value={selectedMember.id} />
                            <input type="hidden" name="startTime" value={time.startTime} />
                            <input type="hidden" name="returnTo" value={returnTo} />
                            {session ? (
                              <button
                                type="submit"
                                className="w-full rounded-xl border px-3 py-2.5 text-center text-sm font-medium transition hover:opacity-90 active:scale-[0.97]"
                                style={{
                                  borderColor: withAlpha(tenant.colorAccent, "55"),
                                  backgroundColor: withAlpha(tenant.colorAccent, "12"),
                                  color: tenant.colorPrimary,
                                }}
                              >
                                {formatTime(time.startTime, tenant.timezone)}
                              </button>
                            ) : (
                              <a
                                href={loginHref}
                                className="block rounded-xl border px-3 py-2.5 text-center text-sm font-medium transition hover:opacity-90"
                                style={{
                                  borderColor: withAlpha(tenant.colorAccent, "55"),
                                  backgroundColor: withAlpha(tenant.colorAccent, "12"),
                                  color: tenant.colorPrimary,
                                }}
                              >
                                {formatTime(time.startTime, tenant.timezone)}
                              </a>
                            )}
                          </form>
                        ))}
                      </div>
                    )}

                    {!session && times.length > 0 ? (
                      <p className="mt-4 text-center text-xs text-gray-500">
                        <a href={loginHref} className="font-semibold underline underline-offset-2" style={{ color: tenant.colorPrimary }}>
                          {t.sign_in}
                        </a>{" "}
                        {t.sign_in_to_confirm}
                      </p>
                    ) : null}
                  </div>
                </>
              )}
            </div>
          </aside>
        </div>
      </div>
    </main>
  );
}
