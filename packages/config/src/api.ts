/**
 * API client base configuration and environment helpers.
 */

/** Returns the base URL for the internal API (server-side calls). */
export function getApiInternalUrl(): string {
  return process.env.API_INTERNAL_URL || "http://localhost:8080";
}

/** Returns the base URL for the public API (client-side calls). */
export function getApiPublicUrl(): string {
  return process.env.NEXT_PUBLIC_API_URL || "http://localhost:8081";
}

/** Returns the platform base domain used for tenant subdomain resolution. */
export function getBaseDomain(): string {
  return process.env.NEXT_PUBLIC_BASE_DOMAIN || "zenvikar.localhost";
}
