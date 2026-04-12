export const BOOKING_TOKEN_COOKIE = "zenvikar_booking_token";

export function persistAuthToken(token: string, expiresAtISO?: string) {
  const maxAge = toMaxAge(expiresAtISO);
  document.cookie = `${BOOKING_TOKEN_COOKIE}=${encodeURIComponent(token)}; Path=/; Max-Age=${maxAge}; SameSite=Lax`;
}

export function clearAuthToken() {
  document.cookie = `${BOOKING_TOKEN_COOKIE}=; Path=/; Max-Age=0; SameSite=Lax`;
}

function toMaxAge(expiresAtISO?: string): number {
  if (!expiresAtISO) return 60 * 60 * 24;
  const expiresAt = new Date(expiresAtISO).getTime();
  const seconds = Math.floor((expiresAt - Date.now()) / 1000);
  return seconds > 0 ? seconds : 60 * 60;
}
