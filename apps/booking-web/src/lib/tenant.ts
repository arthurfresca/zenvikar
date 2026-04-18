import type { Tenant } from "@zenvikar/types";
import { fetchServerApi } from "@/lib/server-api";

export async function resolveTenant(host: string): Promise<Tenant> {
  const baseDomain =
    process.env.NEXT_PUBLIC_BASE_DOMAIN || "zenvikar.localhost";
  const slug = host.replace(`.${baseDomain}`, "").split(":")[0];
  const res = await fetchServerApi({
    path: `/api/v1/tenants/resolve?slug=${slug}`,
    cache: "no-store",
  });
  if (!res.ok) throw new Error(`Tenant not found: ${slug}`);
  return res.json();
}
