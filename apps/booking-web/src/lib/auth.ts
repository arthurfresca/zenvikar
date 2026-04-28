export const BOOKING_TOKEN_COOKIE = "zenvikar_booking_token";

const RESERVED_SLUGS = new Set(["www", "api", "admin", "manage", "app", "mail", "smtp", "ftp", "ssh", "git", "cdn", "static", "assets", "media", "blog", "docs", "help", "support", "status", "billing", "zenvikar"]);

export function persistAuthToken(token: string, expiresAtISO?: string) {
  const maxAge = toMaxAge(expiresAtISO);
  const domain = getCookieDomain();
  document.cookie = `${BOOKING_TOKEN_COOKIE}=${encodeURIComponent(token)}; Path=/; Max-Age=${maxAge}; SameSite=Lax${domain ? `; Domain=${domain}` : ""}`;
}

export function clearAuthToken() {
  const domain = getCookieDomain();
  document.cookie = `${BOOKING_TOKEN_COOKIE}=; Path=/; Max-Age=0; SameSite=Lax`;
  if (domain) {
    document.cookie = `${BOOKING_TOKEN_COOKIE}=; Path=/; Max-Age=0; SameSite=Lax; Domain=${domain}`;
  }
}

export function currentBookingTenantSlug() {
  if (typeof window === "undefined") {
    return "";
  }
  const hostname = window.location.hostname;
  const baseDomain = process.env.NEXT_PUBLIC_BASE_DOMAIN || "zenvikar.localhost";
  if (hostname === baseDomain || hostname === "localhost") {
    return "";
  }
  if (hostname.endsWith(`.${baseDomain}`)) {
	    const slug = hostname.slice(0, hostname.length-(`.${baseDomain}`).length);
	    return RESERVED_SLUGS.has(slug) ? "" : slug;
  }
  return "";
}

function getCookieDomain() {
  const hostname = window.location.hostname;
  const configuredBaseDomain = process.env.NEXT_PUBLIC_BASE_DOMAIN || "zenvikar.localhost";
  if (hostname === "localhost") {
    return "";
  }
  if (hostname === configuredBaseDomain || hostname.endsWith(`.${configuredBaseDomain}`)) {
    return configuredBaseDomain;
  }
  return "";
}

function toMaxAge(expiresAtISO?: string): number {
  if (!expiresAtISO) return 60 * 60 * 24;
  const expiresAt = new Date(expiresAtISO).getTime();
  const seconds = Math.floor((expiresAt - Date.now()) / 1000);
  return seconds > 0 ? seconds : 60 * 60;
}
