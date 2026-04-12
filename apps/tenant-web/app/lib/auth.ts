export const TENANT_TOKEN_COOKIE = "zenvikar_tenant_token";

export function persistTenantToken(token: string, expiresAtISO?: string) {
  const maxAge = toMaxAge(expiresAtISO);
  document.cookie = `${TENANT_TOKEN_COOKIE}=${encodeURIComponent(token)}; Path=/; Max-Age=${maxAge}; SameSite=Lax`;
}

export function clearTenantToken() {
  document.cookie = `${TENANT_TOKEN_COOKIE}=; Path=/; Max-Age=0; SameSite=Lax`;
}

function toMaxAge(expiresAtISO?: string): number {
  if (!expiresAtISO) return 60 * 60 * 24;
  const expiresAt = new Date(expiresAtISO).getTime();
  const seconds = Math.floor((expiresAt - Date.now()) / 1000);
  return seconds > 0 ? seconds : 60 * 60;
}
