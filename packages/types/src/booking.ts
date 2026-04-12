export interface Service {
  id: string;
  tenantId: string;
  name: string;
  description: string | null;
  durationMinutes: number;
  bufferBefore: number;
  bufferAfter: number;
  enabled: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface ServiceMember {
  id: string;
  serviceId: string;
  membershipId: string;
  priceCents: number;
  description: string | null;
}

export interface OpeningHours {
  id: string;
  serviceMemberId: string;
  dayOfWeek: number;
  openTime: string;
  closeTime: string;
  enabled: boolean;
}

export interface BlockedDate {
  id: string;
  membershipId: string;
  date: string;
  reason: string | null;
}

export interface Booking {
  id: string;
  tenantId: string;
  serviceMemberId: string;
  customerId: string;
  priceCents: number;
  startTime: string;
  endTime: string;
  status: "pending" | "confirmed" | "cancelled";
  createdAt: string;
  updatedAt: string;
}
