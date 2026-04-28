import { cookies } from "next/headers";
import { redirect } from "next/navigation";
import type { OpeningHours, Service, Tenant, TenantMembership } from "@zenvikar/types";
import { fetchServerApi } from "@/lib/server-api";

const TOKEN_COOKIE = "zenvikar_tenant_token";

export type TenantSession = {
  userId: string;
  email: string;
  name: string;
  audience: string;
  tenantRoles: Record<string, string>;
  currentTenantId?: string;
  currentTenantSlug?: string;
};

export type TenantAccess = {
  tenantId: string;
  tenantSlug: string;
  tenantName: string;
  role: string;
};

export type MembershipDetails = TenantMembership & {
  id: string;
  description: string | null;
  photoUrl: string | null;
  createdAt: string;
  updatedAt: string;
  user: {
    id: string;
    email: string;
    name: string;
    phone: string | null;
    preferredContact: string;
    locale: string;
    emailVerified: boolean;
    createdAt: string;
    updatedAt: string;
  };
};

export type ServiceMemberDetails = {
  id: string;
  serviceId: string;
  membershipId: string;
  priceCents: number;
  description: string | null;
  memberName: string;
  memberEmail: string;
  tenantId: string;
};

export type BlockedDateDetails = {
  id: string;
  membershipId: string;
  date: string;
  reason: string | null;
};

export type TenantBooking = {
  id: string;
  tenantId: string;
  serviceMemberId: string;
  customerId: string;
  priceCents: number;
  startTime: string;
  endTime: string;
  status: "pending" | "confirmed" | "cancelled";
  createdAt: string;
  updatedAt: string;
  customerName: string;
  customerEmail: string;
  memberName: string;
  serviceName: string;
};

export async function getTenantToken() {
  const cookieStore = await cookies();
  return cookieStore.get(TOKEN_COOKIE)?.value || null;
}

export async function loadTenantSession(token: string): Promise<TenantSession | null> {
  const res = await fetchServerApi({
    path: "/api/v1/auth/me",
    headers: { Authorization: `Bearer ${token}` },
    cache: "no-store",
  });
  if (!res.ok) return null;
  return res.json();
}

export async function loadTenantAccess(token: string): Promise<TenantAccess[]> {
  const res = await fetchServerApi({
    path: "/api/v1/auth/tenants",
    headers: { Authorization: `Bearer ${token}` },
    cache: "no-store",
  });
  if (!res.ok) return [];
  const data = await res.json();
  return Array.isArray(data?.tenants) ? data.tenants : [];
}

export async function resolveTenantBySlug(slug: string): Promise<Tenant | null> {
  const res = await fetchServerApi({
    path: `/api/v1/tenants/resolve?slug=${encodeURIComponent(slug)}`,
    cache: "no-store",
  });
  if (!res.ok) return null;
  return res.json();
}

export async function requireTenantWorkspace(tenantSlug: string) {
  const token = await getTenantToken();
  if (!token) {
    redirect(`/login?next=/t/${encodeURIComponent(tenantSlug)}`);
  }

  const [tenant, session] = await Promise.all([
    resolveTenantBySlug(tenantSlug),
    loadTenantSession(token),
  ]);

  if (!tenant || !session || !session.tenantRoles?.[tenant.id]) {
    redirect(`/login?reauth=1&next=/t/${encodeURIComponent(tenantSlug)}`);
  }

  if (session.currentTenantId && session.currentTenantId !== tenant.id) {
    redirect(`/login?reauth=1&next=/t/${encodeURIComponent(tenantSlug)}`);
  }

  return { token, session, tenant };
}

export async function loadTenantServices(token: string, tenantId: string): Promise<Service[]> {
  const res = await fetchServerApi({
    path: `/api/v1/tenant/tenants/${tenantId}/services`,
    headers: { Authorization: `Bearer ${token}` },
    cache: "no-store",
  });
  if (!res.ok) return [];
  const data = await res.json();
  return Array.isArray(data?.services) ? data.services : [];
}

export async function loadServiceMembers(
  token: string,
  tenantId: string,
  serviceId: string
): Promise<ServiceMemberDetails[]> {
  const res = await fetchServerApi({
    path: `/api/v1/tenant/tenants/${tenantId}/services/${serviceId}/members`,
    headers: { Authorization: `Bearer ${token}` },
    cache: "no-store",
  });
  if (!res.ok) return [];
  const data = await res.json();
  return Array.isArray(data?.members) ? data.members : [];
}

export async function loadTenantMemberships(
  token: string,
  tenantId: string
): Promise<MembershipDetails[]> {
  const res = await fetchServerApi({
    path: `/api/v1/tenant/tenants/${tenantId}/memberships`,
    headers: { Authorization: `Bearer ${token}` },
    cache: "no-store",
  });
  if (!res.ok) return [];
  const data = await res.json();
  return Array.isArray(data?.memberships) ? data.memberships : [];
}

export async function loadTenantBookings(
  token: string,
  tenantId: string,
  fromDate?: string,
  toDate?: string
): Promise<TenantBooking[]> {
  const now = new Date();
  const rangeStart = fromDate
    ? new Date(fromDate).toISOString()
    : new Date(now.getTime() - 1000 * 60 * 60 * 24).toISOString();
  const rangeEnd = toDate
    ? new Date(new Date(toDate).getTime() + 1000 * 60 * 60 * 24).toISOString()
    : new Date(now.getTime() + 1000 * 60 * 60 * 24 * 14).toISOString();
  const res = await fetchServerApi({
    path: `/api/v1/tenant/tenants/${tenantId}/bookings?from=${encodeURIComponent(rangeStart)}&to=${encodeURIComponent(rangeEnd)}`,
    headers: { Authorization: `Bearer ${token}` },
    cache: "no-store",
  });
  if (!res.ok) return [];
  const data = await res.json();
  return Array.isArray(data?.bookings) ? data.bookings : [];
}

export async function loadOpeningHours(
  token: string,
  tenantId: string,
  serviceMemberId: string
): Promise<OpeningHours[]> {
  const res = await fetchServerApi({
    path: `/api/v1/tenant/tenants/${tenantId}/service-members/${serviceMemberId}/opening-hours`,
    headers: { Authorization: `Bearer ${token}` },
    cache: "no-store",
  });
  if (!res.ok) return [];
  const data = await res.json();
  return Array.isArray(data?.openingHours) ? data.openingHours : [];
}

export async function loadBlockedDates(
  token: string,
  tenantId: string,
  membershipId: string
): Promise<BlockedDateDetails[]> {
  const res = await fetchServerApi({
    path: `/api/v1/tenant/tenants/${tenantId}/memberships/${membershipId}/blocked-dates`,
    headers: { Authorization: `Bearer ${token}` },
    cache: "no-store",
  });
  if (!res.ok) return [];
  const data = await res.json();
  return Array.isArray(data?.blockedDates) ? data.blockedDates : [];
}
