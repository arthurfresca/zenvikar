export type PreferredContact = "email" | "phone" | "whatsapp";

export interface User {
  id: string;
  email: string;
  name: string;
  phone: string | null;
  preferredContact: PreferredContact;
  locale: "en" | "pt";
}

export type AuthProvider = "email" | "google" | "facebook";

export interface UserAuthProvider {
  id: string;
  userId: string;
  provider: AuthProvider;
  providerId: string;
  createdAt: string;
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
  photoUrl: string | null;
  description: string | null;
}
