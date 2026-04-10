export interface TenantColors {
  primary: string;
  secondary: string;
  accent: string;
}

export interface Tenant {
  id: string;
  slug: string;
  displayName: string;
  logoUrl: string | null;
  colors: TenantColors;
  timezone: string;
  defaultLocale: "en" | "pt";
  enabled: boolean;
  createdAt: string;
  updatedAt: string;
}
