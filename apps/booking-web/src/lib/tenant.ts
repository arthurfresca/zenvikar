import type { Tenant } from "@zenvikar/types";

export async function resolveTenant(host: string): Promise<Tenant> {
  const baseDomain =
    process.env.NEXT_PUBLIC_BASE_DOMAIN || "zenvikar.localhost";
  const slug = host.replace(`.${baseDomain}`, "").split(":")[0];
  const res = await fetch(
    `${process.env.API_INTERNAL_URL}/api/v1/tenants/resolve?slug=${slug}`,
    { next: { revalidate: 300 } }
  );
  if (!res.ok) throw new Error(`Tenant not found: ${slug}`);
  return res.json();
}
