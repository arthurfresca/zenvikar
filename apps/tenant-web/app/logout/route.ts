import { NextResponse } from "next/server";

const TOKEN_COOKIE = "zenvikar_tenant_token";

export async function GET() {
  const response = new NextResponse(null, {
    status: 307,
    headers: { Location: "/login?reauth=1" },
  });
  response.cookies.set(TOKEN_COOKIE, "", {
    path: "/",
    maxAge: 0,
  });
  return response;
}
