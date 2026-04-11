import { NextRequest, NextResponse } from "next/server";

export function middleware(request: NextRequest) {
  const host =
    request.headers.get("x-forwarded-host") ||
    request.headers.get("host") ||
    "";

  // Pass the host to server components via request headers
  const requestHeaders = new Headers(request.headers);
  requestHeaders.set("x-tenant-host", host);

  return NextResponse.next({
    request: { headers: requestHeaders },
  });
}
