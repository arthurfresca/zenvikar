import type { Metadata } from "next";
import { headers } from "next/headers";
import { resolveTenant } from "@/lib/tenant";
import "./globals.css";

export const metadata: Metadata = {
  title: "Book an Appointment — Zenvikar",
  description: "Book an appointment with your favorite service provider",
};

export default async function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const headersList = await headers();
  const host = headersList.get("x-tenant-host") || headersList.get("host") || "";
  let locale = "en";
  try {
    const tenant = await resolveTenant(host);
    locale = tenant.defaultLocale;
  } catch {
    // fallback to "en" if tenant cannot be resolved
  }

  return (
    <html lang={locale}>
      <body>{children}</body>
    </html>
  );
}
