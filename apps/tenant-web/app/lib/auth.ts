export const TENANT_TOKEN_COOKIE = "zenvikar_tenant_token";

const RESERVED_SLUGS = new Set(["www", "api", "admin", "manage", "app", "mail", "smtp", "ftp", "ssh", "git", "cdn", "static", "assets", "media", "blog", "docs", "help", "support", "status", "billing", "zenvikar"]);

export function persistTenantToken(token: string, expiresAtISO?: string) {
  const maxAge = toMaxAge(expiresAtISO);
  const domain = getCookieDomain();
  document.cookie = `${TENANT_TOKEN_COOKIE}=${encodeURIComponent(token)}; Path=/; Max-Age=${maxAge}; SameSite=Lax${domain ? `; Domain=${domain}` : ""}`;
}

export function clearTenantToken() {
  const domain = getCookieDomain();
  document.cookie = `${TENANT_TOKEN_COOKIE}=; Path=/; Max-Age=0; SameSite=Lax`;
  if (domain) {
    document.cookie = `${TENANT_TOKEN_COOKIE}=; Path=/; Max-Age=0; SameSite=Lax; Domain=${domain}`;
  }
}

export function currentTenantLoginSlug(nextPath?: string) {
  if (typeof window === "undefined") {
    return extractTenantSlugFromPath(nextPath || "");
  }
  const fromPath = extractTenantSlugFromPath(nextPath || window.location.pathname);
  if (fromPath) {
    return fromPath;
  }
  const hostname = window.location.hostname;
  const baseDomain = process.env.NEXT_PUBLIC_BASE_DOMAIN || "zenvikar.localhost";
  if (hostname.endsWith(`.${baseDomain}`) && hostname !== baseDomain) {
	    const slug = hostname.slice(0, hostname.length - (`.${baseDomain}`).length);
	    return RESERVED_SLUGS.has(slug) ? "" : slug;
  }
  return "";
}

function extractTenantSlugFromPath(pathname: string) {
  const match = pathname.match(/^\/t\/([^/?#]+)/);
  return match?.[1] || "";
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
