export interface Service {
  id: string;
  tenantId: string;
  name: string;
  durationMinutes: number;
  bufferBefore: number;
  bufferAfter: number;
  enabled: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface OpeningHours {
  id: string;
  tenantId: string;
  dayOfWeek: number;
  openTime: string;
  closeTime: string;
  enabled: boolean;
}

export interface BlockedDate {
  id: string;
  tenantId: string;
  date: string;
  reason: string | null;
}

export interface Booking {
  id: string;
  tenantId: string;
  serviceId: string;
  customerId: string;
  startTime: string;
  endTime: string;
  status: "pending" | "confirmed" | "cancelled";
  timezone: string;
  createdAt: string;
  updatedAt: string;
}
