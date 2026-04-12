import { NextRequest, NextResponse } from "next/server";

const TOKEN_COOKIE = "zenvikar_tenant_token";

export function middleware(request: NextRequest) {
  const { pathname, search } = request.nextUrl;
  const token = request.cookies.get(TOKEN_COOKIE)?.value;
  const isLogin = pathname === "/login";
  const isReauth = request.nextUrl.searchParams.get("reauth") === "1";

  if (!token && !isLogin) {
    const loginURL = new URL("/login", request.url);
    loginURL.searchParams.set("next", `${pathname}${search}`);

    return NextResponse.redirect(loginURL);
  }

  if (token && isLogin && !isReauth) {
    const nextPath = request.nextUrl.searchParams.get("next");
    return NextResponse.redirect(new URL(nextPath || "/", request.url));
  }

  return NextResponse.next();
}

export const config = {
  matcher: ["/((?!_next/static|_next/image|favicon.ico).*)"],
};
