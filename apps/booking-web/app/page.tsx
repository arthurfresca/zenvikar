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

  return (
    <main className="mx-auto flex min-h-screen w-full max-w-5xl flex-col gap-6 px-6 py-10">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">{tenant.displayName}</h1>
          <p className="text-sm text-gray-600">
            Locale: {tenant.defaultLocale} · Timezone: {tenant.timezone}
          </p>
        </div>
        {session ? <LogoutButton /> : <a href="/login" className="rounded bg-black px-4 py-2 text-sm text-white">Sign In</a>}
      </div>

      <section className="rounded border border-gray-200 p-4">
        <h2 className="text-lg font-semibold">Booking Portal</h2>
        <p className="text-sm text-gray-600">
          {session
            ? `Signed in as ${session.email} (${session.name}). Pick a slot below to book instantly.`
            : "Browse services and availability without signing in. You will only be asked to log in when you book an appointment."}
        </p>
      </section>

      {params.booking === "created" ? (
        <section className="rounded border border-green-200 bg-green-50 p-4 text-sm text-green-800">
          Your booking was created successfully.
        </section>
      ) : null}

      {params.bookingError ? (
        <section className="rounded border border-red-200 bg-red-50 p-4 text-sm text-red-700">
          We could not complete the booking. Please try again.
        </section>
      ) : null}

      <section className="grid gap-6 lg:grid-cols-[1.2fr_0.8fr]">
        <div className="rounded border border-gray-200 p-4">
          <h2 className="text-lg font-semibold">Services</h2>
          {services.length === 0 ? (
            <p className="mt-3 text-sm text-gray-600">This tenant has no public services available yet.</p>
          ) : (
            <div className="mt-4 space-y-4">
              {services.map((service) => (
                <article key={service.id} className="rounded border border-gray-100 p-4">
                  <div className="flex items-start justify-between gap-3">
                    <div>
                      <h3 className="font-semibold text-gray-900">{service.name}</h3>
                      {service.description ? <p className="mt-1 text-sm text-gray-600">{service.description}</p> : null}
                    </div>
                    <span className="rounded bg-gray-100 px-2 py-1 text-xs text-gray-700">
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
                          className={`rounded border p-3 text-sm ${selected ? "border-black bg-black text-white" : "border-gray-200 hover:border-gray-400"}`}
                        >
                          <div className="font-medium">{member.memberName}</div>
                          <div className={`mt-1 ${selected ? "text-gray-200" : "text-gray-600"}`}>{formatMoney(member.priceCents, tenant.currency)}</div>
                          {member.description ? (
                            <div className={`mt-2 text-xs ${selected ? "text-gray-300" : "text-gray-500"}`}>{member.description}</div>
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

        <div className="rounded border border-gray-200 p-4">
          <h2 className="text-lg font-semibold">Availability</h2>
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
                    className="mt-1 w-full rounded border border-gray-300 px-3 py-2"
                  />
                </label>
                <button type="submit" className="w-full rounded border border-gray-300 px-3 py-2 text-sm hover:bg-gray-50">
                  Check availability
                </button>
              </form>

              <div className="mt-4 rounded bg-gray-50 p-3 text-sm text-gray-700">
                <div className="font-medium">{selectedMember.memberName}</div>
                <div>{selectedMember.serviceName}</div>
                <div className="text-gray-500">{formatMoney(selectedMember.priceCents, tenant.currency)}</div>
              </div>

              <div className="mt-4 space-y-3">
                {slots.length === 0 ? (
                  <p className="text-sm text-gray-600">No bookable slots were found for this date.</p>
                ) : (
                  slots.map((slot) => (
                    <form key={slot.startTime} action="/book" method="POST" className="flex items-center justify-between rounded border border-gray-200 p-3">
                      <div>
                        <div className="font-medium">{formatSlotLabel(slot.startTime, slot.endTime, tenant.timezone)}</div>
                        <div className="text-xs text-gray-500">{selectedDate}</div>
                      </div>
                      <input type="hidden" name="tenantSlug" value={tenant.slug} />
                      <input type="hidden" name="serviceMemberId" value={selectedMember.id} />
                      <input type="hidden" name="startTime" value={slot.startTime} />
                      <input type="hidden" name="returnTo" value={returnTo} />
                      <button type="submit" className="rounded bg-black px-3 py-2 text-sm text-white">
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
    </main>
  );
}
