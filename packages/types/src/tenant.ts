export interface Tenant {
  id: string;
  slug: string;
  displayName: string;
  logoUrl: string | null;
  colorPrimary: string;
  colorSecondary: string;
  colorAccent: string;
  phone: string | null;
  email: string | null;
  address: string | null;
  currency: string;
  slotIntervalMinutes: number;
  timezone: string;
  defaultLocale: "en" | "pt";
  enabled: boolean;
  createdAt: string;
  updatedAt: string;
}
