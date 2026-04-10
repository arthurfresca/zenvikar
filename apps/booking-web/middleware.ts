import { NextRequest, NextResponse } from "next/server";

export function middleware(request: NextRequest) {
  const host =
    request.headers.get("x-forwarded-host") ||
    request.headers.get("host") ||
    "";
  const response = NextResponse.next();
  response.headers.set("x-tenant-host", host);
  return response;
}
