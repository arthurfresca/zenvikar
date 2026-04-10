package availability

import (
	"time"

	"github.com/google/uuid"
)

// OpeningHours represents the opening hours for a tenant on a specific day of the week.
type OpeningHours struct {
	ID        uuid.UUID `json:"id" db:"id"`
	TenantID  uuid.UUID `json:"tenantId" db:"tenant_id"`
	DayOfWeek int       `json:"dayOfWeek" db:"day_of_week"` // 0=Sunday, 6=Saturday
	OpenTime  string    `json:"openTime" db:"open_time"`    // "09:00"
	CloseTime string    `json:"closeTime" db:"close_time"`  // "18:00"
	Enabled   bool      `json:"enabled" db:"enabled"`
}

// BlockedDate represents a date on which a tenant does not accept bookings.
type BlockedDate struct {
	ID       uuid.UUID `json:"id" db:"id"`
	TenantID uuid.UUID `json:"tenantId" db:"tenant_id"`
	Date     time.Time `json:"date" db:"date"`
	Reason   *string   `json:"reason" db:"reason"`
}
