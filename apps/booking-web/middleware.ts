import { NextRequest, NextResponse } from "next/server";

const TOKEN_COOKIE = "zenvikar_booking_token";
const AUTH_PAGES = ["/login", "/signup"];

export function middleware(request: NextRequest) {
  const { pathname, search } = request.nextUrl;
  const host =
    request.headers.get("x-forwarded-host") ||
    request.headers.get("host") ||
    "";

  const isAuthPage = AUTH_PAGES.includes(pathname);
  const token = request.cookies.get(TOKEN_COOKIE)?.value;

  const isReauth = request.nextUrl.searchParams.get("reauth") === "1";

  if (token && isAuthPage && !isReauth) {
    return NextResponse.redirect(new URL("/", request.url));
  }

  const requestHeaders = new Headers(request.headers);
  requestHeaders.set("x-tenant-host", host);

  return NextResponse.next({
    request: { headers: requestHeaders },
  });
}

export const config = {
  matcher: ["/((?!_next/static|_next/image|favicon.ico).*)"],
};
