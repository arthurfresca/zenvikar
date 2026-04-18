package availability

import (
	"time"

	"github.com/google/uuid"
)

// OpeningHours represents when a specific member offers a specific service on a day of the week.
// For example: "Bob does Haircut on Monday 09:00-18:00".
type OpeningHours struct {
	ID              uuid.UUID `json:"id" db:"id"`
	TenantID        uuid.UUID `json:"tenantId,omitempty" db:"-"`
	ServiceMemberID uuid.UUID `json:"serviceMemberId" db:"service_member_id"`
	DayOfWeek       int       `json:"dayOfWeek" db:"day_of_week"` // 0=Sunday, 6=Saturday
	OpenTime        string    `json:"openTime" db:"open_time"`    // "09:00"
	CloseTime       string    `json:"closeTime" db:"close_time"`  // "18:00"
	Enabled         bool      `json:"enabled" db:"enabled"`
}

// BlockedDate represents a date on which a member is not available.
// Stays at the membership level since a day off blocks all services.
type BlockedDate struct {
	ID           uuid.UUID `json:"id" db:"id"`
	TenantID     uuid.UUID `json:"tenantId,omitempty" db:"-"`
	MembershipID uuid.UUID `json:"membershipId" db:"membership_id"`
	Date         time.Time `json:"date" db:"date"`
	Reason       *string   `json:"reason" db:"reason"`
}
