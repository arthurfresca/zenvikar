export interface User {
  id: string;
  email: string;
  name: string;
  locale: "en" | "pt";
}

export type TenantRole =
  | "tenant_owner"
  | "tenant_manager"
  | "tenant_staff"
  | "tenant_finance_viewer";

export type PlatformRole =
  | "admin"
  | "support_admin"
  | "finance_admin";

export interface TenantMembership {
  tenantId: string;
  userId: string;
  role: TenantRole;
}
