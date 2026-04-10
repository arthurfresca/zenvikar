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

// Booking represents a booking made by a customer for a service within a tenant.
type Booking struct {
	ID         uuid.UUID `json:"id" db:"id"`
	TenantID   uuid.UUID `json:"tenantId" db:"tenant_id"`
	ServiceID  uuid.UUID `json:"serviceId" db:"service_id"`
	CustomerID uuid.UUID `json:"customerId" db:"customer_id"`
	StartTime  time.Time `json:"startTime" db:"start_time"`
	EndTime    time.Time `json:"endTime" db:"end_time"`
	Status     string    `json:"status" db:"status"`
	Timezone   string    `json:"timezone" db:"timezone"`
	CreatedAt  time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt  time.Time `json:"updatedAt" db:"updated_at"`
}
