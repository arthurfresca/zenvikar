import { NextRequest, NextResponse } from "next/server";

const TOKEN_COOKIE = "zenvikar_booking_token";
const PUBLIC_PATHS = ["/login", "/signup"];

export function middleware(request: NextRequest) {
  const { pathname, search } = request.nextUrl;
  const host =
    request.headers.get("x-forwarded-host") ||
    request.headers.get("host") ||
    "";

  const isPublicPath = PUBLIC_PATHS.includes(pathname);
  const token = request.cookies.get(TOKEN_COOKIE)?.value;

  if (!token && !isPublicPath) {
    const loginURL = new URL("/login", request.url);
    loginURL.searchParams.set("next", `${pathname}${search}`);
    return NextResponse.redirect(loginURL);
  }

  const isReauth = request.nextUrl.searchParams.get("reauth") === "1";

  if (token && isPublicPath && !isReauth) {
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
