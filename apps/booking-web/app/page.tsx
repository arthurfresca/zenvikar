import { headers } from "next/headers";
import { cookies } from "next/headers";
import { resolveTenant } from "@/lib/tenant";
import { fetchServerApi } from "@/lib/server-api";
import type { Service, ServiceMember, Tenant } from "@zenvikar/types";
import { LogoutButton } from "./components/logout-button";

const TOKEN_COOKIE = "zenvikar_booking_token";

type BookingSession = {
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
  if (!res.ok) {
    return [];
  }
  const data = (await res.json()) as { services?: PublicService[] };
  return data.services || [];
}

async function loadAvailability(tenantSlug: string, serviceMemberId: string, date: string) {
  const res = await fetchServerApi({
    path: `/api/v1/tenants/${encodeURIComponent(tenantSlug)}/service-members/${serviceMemberId}/availability?date=${encodeURIComponent(date)}`,
    next: { revalidate: 60 },
  });
  if (!res.ok) {
    return [] as Array<{ startTime: string; endTime: string }>;
  }
  const data = (await res.json()) as { slots?: Array<{ startTime: string; endTime: string }> };
  return data.slots || [];
}

function formatMoney(cents: number, currency: string) {
  return new Intl.NumberFormat("en-US", {
    style: "currency",
    currency,
  }).format(cents / 100);
}

function formatSlotLabel(startTime: string, endTime: string, timezone: string) {
  const start = new Date(startTime);
  const end = new Date(endTime);
  const timeFormat = new Intl.DateTimeFormat("en-US", {
    hour: "numeric",
    minute: "2-digit",
    timeZone: timezone,
  });
  return `${timeFormat.format(start)} to ${timeFormat.format(end)}`;
}

function getInitialDate(value?: string) {
  if (value) {
    return value;
  }
  return new Date().toISOString().slice(0, 10);
}

function withAlpha(hex: string, alpha: string) {
  return `${hex}${alpha}`;
}

function getContrastColor(hex: string) {
  const normalized = hex.replace("#", "");
  const value = normalized.length === 3
    ? normalized.split("").map((char) => char + char).join("")
    : normalized;
  const r = parseInt(value.slice(0, 2), 16);
  const g = parseInt(value.slice(2, 4), 16);
  const b = parseInt(value.slice(4, 6), 16);
  const brightness = (r * 299 + g * 587 + b * 114) / 1000;
  return brightness > 155 ? "#111827" : "#ffffff";
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
  const session = token ? ((await loadSession(token)) as BookingSession | null) : null;

  let tenant: Tenant;
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

  const services = await loadPublicServices(tenant.slug);
  const members = services.flatMap((service) =>
    service.members.map((member) => ({
      ...member,
      serviceName: service.name,
      durationMinutes: service.durationMinutes,
    }))
  );
  const selectedDate = getInitialDate(params.date);
  const selectedMember =
    members.find((member) => member.id === params.member) || members[0] || null;
  const slots = selectedMember
    ? await loadAvailability(tenant.slug, selectedMember.id, selectedDate)
    : [];
  const returnTo = `/${selectedMember ? `?member=${encodeURIComponent(selectedMember.id)}&date=${encodeURIComponent(selectedDate)}` : ""}`;
  const primaryTextColor = getContrastColor(tenant.colorPrimary);
  const accentTextColor = getContrastColor(tenant.colorAccent);

  return (
    <main
      className="min-h-screen"
      style={{
        background: `linear-gradient(180deg, ${withAlpha(tenant.colorSecondary, "14")} 0%, #ffffff 32%)`,
      }}
    >
      <div className="mx-auto flex w-full max-w-5xl flex-col gap-6 px-6 py-10">
      <section
        className="overflow-hidden rounded-3xl border shadow-sm"
        style={{ borderColor: withAlpha(tenant.colorPrimary, "22") }}
      >
        <div
          className="flex flex-col gap-6 p-6 sm:flex-row sm:items-start sm:justify-between"
          style={{
            background: `linear-gradient(135deg, ${tenant.colorPrimary} 0%, ${tenant.colorSecondary} 100%)`,
            color: primaryTextColor,
          }}
        >
          <div className="flex items-start gap-4">
            {tenant.logoUrl ? (
              <img
                src={tenant.logoUrl}
                alt={`${tenant.displayName} logo`}
                className="h-16 w-16 rounded-2xl border bg-white/90 object-cover p-2"
              />
            ) : null}
            <div>
              <p className="text-xs uppercase tracking-[0.24em] opacity-75">{tenant.slug}</p>
              <h1 className="mt-2 text-3xl font-semibold">{tenant.displayName}</h1>
              <p className="mt-2 max-w-2xl text-sm opacity-85">
                Browse real-time availability, choose your preferred specialist, and confirm an appointment when you are ready.
              </p>
              <div className="mt-4 flex flex-wrap gap-2 text-xs opacity-80">
                <span className="rounded-full border px-3 py-1">Locale: {tenant.defaultLocale}</span>
                <span className="rounded-full border px-3 py-1">Timezone: {tenant.timezone}</span>
                <span className="rounded-full border px-3 py-1">Every {tenant.slotIntervalMinutes} min</span>
              </div>
            </div>
          </div>
          <div className="flex flex-col items-start gap-3 sm:items-end">
            {session ? (
              <LogoutButton />
            ) : (
              <a
                href="/login"
                className="rounded-full px-4 py-2 text-sm font-medium"
                style={{ backgroundColor: tenant.colorAccent, color: accentTextColor }}
              >
                Sign In To Book
              </a>
            )}
            <div className="text-sm opacity-85 sm:text-right">
              {tenant.phone ? <div>{tenant.phone}</div> : null}
              {tenant.email ? <div>{tenant.email}</div> : null}
              {tenant.address ? <div>{tenant.address}</div> : null}
            </div>
          </div>
        </div>
      </section>

      <div className="flex items-center justify-between gap-4 rounded-2xl border bg-white p-4 shadow-sm" style={{ borderColor: withAlpha(tenant.colorPrimary, "18") }}>
        <div>
          <h2 className="text-lg font-semibold" style={{ color: tenant.colorPrimary }}>Booking Portal</h2>
          <p className="text-sm text-gray-600">
            {session
              ? `Signed in as ${session.email} (${session.name}). Pick a slot below to book instantly.`
              : "Browse services and availability without signing in. You will only be asked to log in when you book an appointment."}
          </p>
        </div>
        <div className="hidden rounded-full px-3 py-1 text-xs font-medium sm:block" style={{ backgroundColor: withAlpha(tenant.colorAccent, "24"), color: tenant.colorPrimary }}>
          {session ? "Authenticated" : "Guest browsing"}
        </div>
      </div>

      {params.booking === "created" ? (
        <section className="rounded-2xl border p-4 text-sm" style={{ borderColor: withAlpha(tenant.colorAccent, "55"), backgroundColor: withAlpha(tenant.colorAccent, "18"), color: tenant.colorPrimary }}>
          Your booking was created successfully.
        </section>
      ) : null}

      {params.bookingError ? (
        <section className="rounded border border-red-200 bg-red-50 p-4 text-sm text-red-700">
          We could not complete the booking. Please try again.
        </section>
      ) : null}

      <section className="grid gap-6 lg:grid-cols-[1.2fr_0.8fr]">
        <div className="rounded-2xl border bg-white p-4 shadow-sm" style={{ borderColor: withAlpha(tenant.colorPrimary, "18") }}>
          <h2 className="text-lg font-semibold" style={{ color: tenant.colorPrimary }}>Services</h2>
          {services.length === 0 ? (
            <p className="mt-3 text-sm text-gray-600">This tenant has no public services available yet.</p>
          ) : (
            <div className="mt-4 space-y-4">
              {services.map((service) => (
                <article key={service.id} className="rounded-2xl border p-4" style={{ borderColor: withAlpha(tenant.colorSecondary, "26"), backgroundColor: withAlpha(tenant.colorSecondary, "08") }}>
                  <div className="flex items-start justify-between gap-3">
                    <div>
                      <h3 className="font-semibold" style={{ color: tenant.colorPrimary }}>{service.name}</h3>
                      {service.description ? <p className="mt-1 text-sm text-gray-600">{service.description}</p> : null}
                    </div>
                    <span className="rounded-full px-2 py-1 text-xs" style={{ backgroundColor: withAlpha(tenant.colorAccent, "20"), color: tenant.colorPrimary }}>
                      {service.durationMinutes} min
                    </span>
                  </div>
                  <div className="mt-4 grid gap-3 sm:grid-cols-2">
                    {service.members.map((member) => {
                      const selected = selectedMember?.id === member.id;
                      return (
                        <a
                          key={member.id}
                          href={`/?member=${encodeURIComponent(member.id)}&date=${encodeURIComponent(selectedDate)}`}
                          className="rounded-2xl border p-3 text-sm transition-colors"
                          style={selected
                            ? {
                                borderColor: tenant.colorPrimary,
                                backgroundColor: tenant.colorPrimary,
                                color: primaryTextColor,
                              }
                            : {
                                borderColor: withAlpha(tenant.colorPrimary, "22"),
                                backgroundColor: "#ffffff",
                                color: tenant.colorPrimary,
                              }}
                        >
                          <div className="font-medium">{member.memberName}</div>
                          <div className="mt-1" style={{ color: selected ? withAlpha(primaryTextColor, "CC") : tenant.colorSecondary }}>{formatMoney(member.priceCents, tenant.currency)}</div>
                          {member.description ? (
                            <div className="mt-2 text-xs" style={{ color: selected ? withAlpha(primaryTextColor, "B3") : "#6b7280" }}>{member.description}</div>
                          ) : null}
                        </a>
                      );
                    })}
                  </div>
                </article>
              ))}
            </div>
          )}
        </div>

        <div className="rounded-2xl border bg-white p-4 shadow-sm" style={{ borderColor: withAlpha(tenant.colorPrimary, "18") }}>
          <h2 className="text-lg font-semibold" style={{ color: tenant.colorPrimary }}>Availability</h2>
          {selectedMember ? (
            <>
              <form className="mt-4 space-y-3" method="GET">
                <input type="hidden" name="member" value={selectedMember.id} />
                <label className="block text-sm text-gray-700">
                  Date
                  <input
                    type="date"
                    name="date"
                    defaultValue={selectedDate}
                    className="mt-1 w-full rounded border px-3 py-2"
                    style={{ borderColor: withAlpha(tenant.colorPrimary, "35") }}
                  />
                </label>
                <button type="submit" className="w-full rounded px-3 py-2 text-sm" style={{ border: `1px solid ${withAlpha(tenant.colorSecondary, "44")}`, backgroundColor: withAlpha(tenant.colorSecondary, "10"), color: tenant.colorPrimary }}>
                  Check availability
                </button>
              </form>

              <div className="mt-4 rounded-2xl p-3 text-sm" style={{ backgroundColor: withAlpha(tenant.colorAccent, "16"), color: tenant.colorPrimary }}>
                <div className="font-medium">{selectedMember.memberName}</div>
                <div>{selectedMember.serviceName}</div>
                <div style={{ color: tenant.colorSecondary }}>{formatMoney(selectedMember.priceCents, tenant.currency)}</div>
              </div>

              <div className="mt-4 space-y-3">
                {slots.length === 0 ? (
                  <p className="text-sm text-gray-600">No bookable slots were found for this date.</p>
                ) : (
                  slots.map((slot) => (
                    <form key={slot.startTime} action="/book" method="POST" className="flex items-center justify-between rounded-2xl border p-3" style={{ borderColor: withAlpha(tenant.colorPrimary, "18") }}>
                      <div>
                        <div className="font-medium" style={{ color: tenant.colorPrimary }}>{formatSlotLabel(slot.startTime, slot.endTime, tenant.timezone)}</div>
                        <div className="text-xs text-gray-500">{selectedDate}</div>
                      </div>
                      <input type="hidden" name="tenantSlug" value={tenant.slug} />
                      <input type="hidden" name="serviceMemberId" value={selectedMember.id} />
                      <input type="hidden" name="startTime" value={slot.startTime} />
                      <input type="hidden" name="returnTo" value={returnTo} />
                      <button type="submit" className="rounded px-3 py-2 text-sm font-medium" style={{ backgroundColor: tenant.colorAccent, color: accentTextColor }}>
                        {session ? "Book now" : "Sign in to book"}
                      </button>
                    </form>
                  ))
                )}
              </div>
            </>
          ) : (
            <p className="mt-3 text-sm text-gray-600">Choose a service provider to view available appointment times.</p>
          )}
        </div>
      </section>
      </div>
    </main>
  );
}
