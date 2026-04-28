import { NextRequest, NextResponse } from "next/server";
import { fetchServerApi } from "@/lib/server-api";
import { requireTenantWorkspace } from "@/lib/tenant-session";

function eachDate(startDate: string, endDate: string) {
  const dates: string[] = [];
  const cursor = new Date(`${startDate}T00:00:00Z`);
  const end = new Date(`${endDate}T00:00:00Z`);
  while (cursor <= end) {
    dates.push(cursor.toISOString().slice(0, 10));
    cursor.setUTCDate(cursor.getUTCDate() + 1);
  }
  return dates;
}

function redirectBack(request: NextRequest, tenantSlug: string, search: URLSearchParams) {
  const url = new URL(`/t/${encodeURIComponent(tenantSlug)}`, request.url);
  url.search = search.toString();
  return NextResponse.redirect(url);
}

export async function POST(
  request: NextRequest,
  { params }: { params: Promise<{ tenantSlug: string }> }
) {
  const { tenantSlug } = await params;
  const workspace = await requireTenantWorkspace(tenantSlug);
  const formData = await request.formData();
  const action = String(formData.get("action") || "").trim();
  const search = new URLSearchParams();

  try {
    if (action === "create-service") {
      const payload = {
        name: String(formData.get("name") || "").trim(),
        description: String(formData.get("description") || "").trim() || null,
        durationMinutes: Number(formData.get("durationMinutes") || 0),
        bufferBefore: Number(formData.get("bufferBefore") || 0),
        bufferAfter: Number(formData.get("bufferAfter") || 0),
        enabled: formData.get("enabled") === "on",
      };
      const res = await fetchServerApi({
        path: `/api/v1/tenant/tenants/${workspace.tenant.id}/services`,
        method: "POST",
        headers: {
          Authorization: `Bearer ${workspace.token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify(payload),
      });
      search.set(res.ok ? "serviceAction" : "serviceError", res.ok ? "created" : "create_failed");
      return redirectBack(request, tenantSlug, search);
    }

    if (action === "update-tenant") {
      const payload = {
        displayName: String(formData.get("displayName") || "").trim(),
        logoUrl: String(formData.get("logoUrl") || "").trim() || null,
        colorPrimary: String(formData.get("colorPrimary") || "").trim(),
        colorSecondary: String(formData.get("colorSecondary") || "").trim(),
        colorAccent: String(formData.get("colorAccent") || "").trim(),
        phone: String(formData.get("phone") || "").trim() || null,
        email: String(formData.get("email") || "").trim() || null,
        address: String(formData.get("address") || "").trim() || null,
        currency: String(formData.get("currency") || "").trim(),
        timezone: String(formData.get("timezone") || "").trim(),
        defaultLocale: String(formData.get("defaultLocale") || "en").trim(),
        enabled: formData.get("enabled") === "on",
      };
      const res = await fetchServerApi({
        path: `/api/v1/tenant/tenants/${workspace.tenant.id}`,
        method: "PATCH",
        headers: {
          Authorization: `Bearer ${workspace.token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify(payload),
      });
      search.set(res.ok ? "serviceAction" : "serviceError", res.ok ? "tenant_saved" : "tenant_save_failed");
      return redirectBack(request, tenantSlug, search);
    }

    if (action === "toggle-service") {
      const serviceId = String(formData.get("serviceId") || "").trim();
      const enabled = formData.get("enabled") === "true";
      const res = await fetchServerApi({
        path: `/api/v1/tenant/tenants/${workspace.tenant.id}/services/${serviceId}`,
        method: "PATCH",
        headers: {
          Authorization: `Bearer ${workspace.token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ enabled }),
      });
      search.set(res.ok ? "serviceAction" : "serviceError", res.ok ? "updated" : "update_failed");
      return redirectBack(request, tenantSlug, search);
    }

    if (action === "update-service") {
      const serviceId = String(formData.get("serviceId") || "").trim();
      const payload = {
        name: String(formData.get("name") || "").trim(),
        description: String(formData.get("description") || "").trim() || null,
        durationMinutes: Number(formData.get("durationMinutes") || 0),
        bufferBefore: Number(formData.get("bufferBefore") || 0),
        bufferAfter: Number(formData.get("bufferAfter") || 0),
        enabled: formData.get("enabled") === "on",
      };
      const res = await fetchServerApi({
        path: `/api/v1/tenant/tenants/${workspace.tenant.id}/services/${serviceId}`,
        method: "PATCH",
        headers: {
          Authorization: `Bearer ${workspace.token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify(payload),
      });
      search.set(res.ok ? "serviceAction" : "serviceError", res.ok ? "service_saved" : "service_save_failed");
      return redirectBack(request, tenantSlug, search);
    }

    if (action === "delete-service") {
      const serviceId = String(formData.get("serviceId") || "").trim();
      const res = await fetchServerApi({
        path: `/api/v1/tenant/tenants/${workspace.tenant.id}/services/${serviceId}`,
        method: "DELETE",
        headers: { Authorization: `Bearer ${workspace.token}` },
      });
      search.set(res.ok ? "serviceAction" : "serviceError", res.ok ? "service_deleted" : "service_delete_failed");
      return redirectBack(request, tenantSlug, search);
    }

    if (action === "duplicate-service") {
      const serviceId = String(formData.get("serviceId") || "").trim();
      const currentRes = await fetchServerApi({
        path: `/api/v1/tenant/tenants/${workspace.tenant.id}/services/${serviceId}`,
        headers: { Authorization: `Bearer ${workspace.token}` },
        cache: "no-store",
      });
      if (!currentRes.ok) {
        search.set("serviceError", "service_duplicate_failed");
        return redirectBack(request, tenantSlug, search);
      }
      const current = await currentRes.json();
      const createRes = await fetchServerApi({
        path: `/api/v1/tenant/tenants/${workspace.tenant.id}/services`,
        method: "POST",
        headers: {
          Authorization: `Bearer ${workspace.token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          name: `${current.name} Copy`,
          description: current.description,
          durationMinutes: current.durationMinutes,
          bufferBefore: current.bufferBefore,
          bufferAfter: current.bufferAfter,
          enabled: false,
        }),
      });
      search.set(createRes.ok ? "serviceAction" : "serviceError", createRes.ok ? "service_duplicated" : "service_duplicate_failed");
      return redirectBack(request, tenantSlug, search);
    }

    if (action === "assign-member") {
      const serviceId = String(formData.get("serviceId") || "").trim();
      const membershipId = String(formData.get("membershipId") || "").trim();
      const priceCents = Number(formData.get("priceCents") || 0);
      const description = String(formData.get("description") || "").trim() || null;
      const res = await fetchServerApi({
        path: `/api/v1/tenant/tenants/${workspace.tenant.id}/services/${serviceId}/members`,
        method: "POST",
        headers: {
          Authorization: `Bearer ${workspace.token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ membershipId, priceCents, description }),
      });
      search.set(res.ok ? "serviceAction" : "serviceError", res.ok ? "member_added" : "member_add_failed");
      return redirectBack(request, tenantSlug, search);
    }

    if (action === "remove-member") {
      const serviceId = String(formData.get("serviceId") || "").trim();
      const serviceMemberId = String(formData.get("serviceMemberId") || "").trim();
      const res = await fetchServerApi({
        path: `/api/v1/tenant/tenants/${workspace.tenant.id}/services/${serviceId}/members/${serviceMemberId}`,
        method: "DELETE",
        headers: { Authorization: `Bearer ${workspace.token}` },
      });
      search.set(res.ok ? "serviceAction" : "serviceError", res.ok ? "member_removed" : "member_remove_failed");
      return redirectBack(request, tenantSlug, search);
    }

    if (action === "upsert-opening-hour") {
      const serviceMemberId = String(formData.get("serviceMemberId") || "").trim();
      const dayOfWeek = String(formData.get("dayOfWeek") || "").trim();
      const payload = {
        openTime: String(formData.get("openTime") || "").trim(),
        closeTime: String(formData.get("closeTime") || "").trim(),
        enabled: formData.get("enabled") === "on",
      };
      const res = await fetchServerApi({
        path: `/api/v1/tenant/tenants/${workspace.tenant.id}/service-members/${serviceMemberId}/opening-hours/${dayOfWeek}`,
        method: "PUT",
        headers: {
          Authorization: `Bearer ${workspace.token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify(payload),
      });
      search.set(res.ok ? "serviceAction" : "serviceError", res.ok ? "calendar_saved" : "calendar_save_failed");
      return redirectBack(request, tenantSlug, search);
    }

    if (action === "apply-schedule-template") {
      const serviceMemberId = String(formData.get("serviceMemberId") || "").trim();
      const template = String(formData.get("template") || "").trim();
      const templates: Record<string, Array<{ dayOfWeek: number; openTime: string; closeTime: string; enabled: boolean }>> = {
        weekdays: [1, 2, 3, 4, 5].map((day) => ({ dayOfWeek: day, openTime: "09:00", closeTime: "18:00", enabled: true })),
        extended: [1, 2, 3, 4, 5].map((day) => ({ dayOfWeek: day, openTime: "08:00", closeTime: "20:00", enabled: true })),
        weekend: [6, 0].map((day) => ({ dayOfWeek: day, openTime: "10:00", closeTime: "16:00", enabled: true })),
      };
      const templateItems = templates[template];
      if (!templateItems) {
        search.set("serviceError", "calendar_template_failed");
        return redirectBack(request, tenantSlug, search);
      }
      const results = await Promise.all(
        templateItems.map((item) =>
          fetchServerApi({
            path: `/api/v1/tenant/tenants/${workspace.tenant.id}/service-members/${serviceMemberId}/opening-hours/${item.dayOfWeek}`,
            method: "PUT",
            headers: {
              Authorization: `Bearer ${workspace.token}`,
              "Content-Type": "application/json",
            },
            body: JSON.stringify(item),
          })
        )
      );
      search.set(results.every((res) => res.ok) ? "serviceAction" : "serviceError", results.every((res) => res.ok) ? "calendar_template_applied" : "calendar_template_failed");
      return redirectBack(request, tenantSlug, search);
    }

    if (action === "create-blocked-date") {
      const membershipId = String(formData.get("membershipId") || "").trim();
      const payload = {
        date: String(formData.get("date") || "").trim(),
        reason: String(formData.get("reason") || "").trim() || null,
      };
      const res = await fetchServerApi({
        path: `/api/v1/tenant/tenants/${workspace.tenant.id}/memberships/${membershipId}/blocked-dates`,
        method: "POST",
        headers: {
          Authorization: `Bearer ${workspace.token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify(payload),
      });
      search.set(res.ok ? "serviceAction" : "serviceError", res.ok ? "blocked_date_added" : "blocked_date_failed");
      return redirectBack(request, tenantSlug, search);
    }

    if (action === "create-blocked-range") {
      const membershipId = String(formData.get("membershipId") || "").trim();
      const startDate = String(formData.get("startDate") || "").trim();
      const endDate = String(formData.get("endDate") || "").trim();
      const reason = String(formData.get("reason") || "").trim() || null;
      const dates = eachDate(startDate, endDate || startDate);
      const results = await Promise.all(
        dates.map((date) =>
          fetchServerApi({
            path: `/api/v1/tenant/tenants/${workspace.tenant.id}/memberships/${membershipId}/blocked-dates`,
            method: "POST",
            headers: {
              Authorization: `Bearer ${workspace.token}`,
              "Content-Type": "application/json",
            },
            body: JSON.stringify({ date, reason }),
          })
        )
      );
      search.set(results.every((res) => res.ok) ? "serviceAction" : "serviceError", results.every((res) => res.ok) ? "blocked_range_added" : "blocked_range_failed");
      return redirectBack(request, tenantSlug, search);
    }

    if (action === "delete-blocked-date") {
      const membershipId = String(formData.get("membershipId") || "").trim();
      const date = String(formData.get("date") || "").trim();
      const res = await fetchServerApi({
        path: `/api/v1/tenant/tenants/${workspace.tenant.id}/memberships/${membershipId}/blocked-dates/${date}`,
        method: "DELETE",
        headers: { Authorization: `Bearer ${workspace.token}` },
      });
      search.set(res.ok ? "serviceAction" : "serviceError", res.ok ? "blocked_date_removed" : "blocked_date_remove_failed");
      return redirectBack(request, tenantSlug, search);
    }

    if (action === "update-booking-status") {
      const bookingId = String(formData.get("bookingId") || "").trim();
      const status = String(formData.get("status") || "").trim();
      const res = await fetchServerApi({
        path: `/api/v1/tenant/tenants/${workspace.tenant.id}/bookings/${bookingId}`,
        method: "PATCH",
        headers: {
          Authorization: `Bearer ${workspace.token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ status }),
      });
      search.set(res.ok ? "serviceAction" : "serviceError", res.ok ? "booking_updated" : "booking_update_failed");
      return redirectBack(request, tenantSlug, search);
    }

    if (action === "copy-schedule-from-member") {
      const sourceServiceMemberId = String(formData.get("sourceServiceMemberId") || "").trim();
      const targetServiceMemberId = String(formData.get("targetServiceMemberId") || "").trim();
      const sourceRes = await fetchServerApi({
        path: `/api/v1/tenant/tenants/${workspace.tenant.id}/service-members/${sourceServiceMemberId}/opening-hours`,
        headers: { Authorization: `Bearer ${workspace.token}` },
        cache: "no-store",
      });
      if (!sourceRes.ok) {
        search.set("serviceError", "schedule_copy_failed");
        return redirectBack(request, tenantSlug, search);
      }
      const sourceData = await sourceRes.json();
      const openingHours = Array.isArray(sourceData?.openingHours) ? sourceData.openingHours : [];
      const results = await Promise.all(
        openingHours.map((item: { dayOfWeek: number; openTime: string; closeTime: string; enabled: boolean }) =>
          fetchServerApi({
            path: `/api/v1/tenant/tenants/${workspace.tenant.id}/service-members/${targetServiceMemberId}/opening-hours/${item.dayOfWeek}`,
            method: "PUT",
            headers: {
              Authorization: `Bearer ${workspace.token}`,
              "Content-Type": "application/json",
            },
            body: JSON.stringify({
              openTime: item.openTime.slice(0, 5),
              closeTime: item.closeTime.slice(0, 5),
              enabled: item.enabled,
            }),
          })
        )
      );
      search.set(results.every((res) => res.ok) ? "serviceAction" : "serviceError", results.every((res) => res.ok) ? "schedule_copied" : "schedule_copy_failed");
      return redirectBack(request, tenantSlug, search);
    }
  } catch {
    search.set("serviceError", "unexpected_error");
    return redirectBack(request, tenantSlug, search);
  }

  search.set("serviceError", "invalid_action");
  return redirectBack(request, tenantSlug, search);
}
