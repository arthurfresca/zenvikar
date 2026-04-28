package bookings

import (
	"time"

	"github.com/google/uuid"
)

// Booking status constants.
const (
	StatusPending   = "pending"
	StatusConfirmed = "confirmed"
	StatusCancelled = "cancelled"
)

// Booking represents a booking made by a customer for a specific service member.
type Booking struct {
	ID              uuid.UUID `json:"id" db:"id"`
	TenantID        uuid.UUID `json:"tenantId" db:"tenant_id"`
	ServiceID       uuid.UUID `json:"serviceId,omitempty" db:"-"`
	ServiceMemberID uuid.UUID `json:"serviceMemberId" db:"service_member_id"`
	CustomerID      uuid.UUID `json:"customerId" db:"customer_id"`
	PriceCents      int       `json:"priceCents" db:"price_cents"`
	StartTime       time.Time `json:"startTime" db:"start_time"`
	EndTime         time.Time `json:"endTime" db:"end_time"`
	Status          string    `json:"status" db:"status"`
	CreatedAt       time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt       time.Time `json:"updatedAt" db:"updated_at"`
}

// BookingDetails enriches tenant-side booking responses with joined display data.
type BookingDetails struct {
	Booking
	CustomerName  string `json:"customerName"`
	CustomerEmail string `json:"customerEmail"`
	MemberName    string `json:"memberName"`
	ServiceName   string `json:"serviceName"`
}
