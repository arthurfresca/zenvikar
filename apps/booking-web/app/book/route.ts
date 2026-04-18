import { NextRequest, NextResponse } from "next/server";
import { BOOKING_TOKEN_COOKIE } from "@/lib/auth";
import { fetchServerApi } from "@/lib/server-api";

function appendResult(url: URL, key: string, value: string) {
  url.searchParams.delete("booking");
  url.searchParams.delete("bookingError");
  url.searchParams.set(key, value);
}

export async function POST(request: NextRequest) {
  const formData = await request.formData();
  const tenantSlug = String(formData.get("tenantSlug") || "").trim();
  const serviceMemberId = String(formData.get("serviceMemberId") || "").trim();
  const startTime = String(formData.get("startTime") || "").trim();
  const returnToRaw = String(formData.get("returnTo") || "/").trim() || "/";
  const returnTo = new URL(returnToRaw, request.url);

  if (!tenantSlug || !serviceMemberId || !startTime) {
    appendResult(returnTo, "bookingError", "missing_fields");
    return NextResponse.redirect(returnTo);
  }

  const token = request.cookies.get(BOOKING_TOKEN_COOKIE)?.value;
  if (!token) {
    const loginURL = new URL("/login", request.url);
    loginURL.searchParams.set("next", `${returnTo.pathname}${returnTo.search}`);
    return NextResponse.redirect(loginURL);
  }

  const response = await fetchServerApi({
    path: `/api/v1/tenants/${encodeURIComponent(tenantSlug)}/bookings`,
    method: "POST",
    headers: {
      Authorization: `Bearer ${token}`,
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ serviceMemberId, startTime }),
    cache: "no-store",
  });

  if (!response.ok) {
    const errorText = await response.text();
    appendResult(returnTo, "bookingError", errorText ? "booking_failed" : "booking_failed");
    return NextResponse.redirect(returnTo);
  }

  appendResult(returnTo, "booking", "created");
  return NextResponse.redirect(returnTo);
}
