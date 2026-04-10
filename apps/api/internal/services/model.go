package services

import (
	"time"

	"github.com/google/uuid"
)

// Service represents a bookable service offered by a tenant.
type Service struct {
	ID              uuid.UUID `json:"id" db:"id"`
	TenantID        uuid.UUID `json:"tenantId" db:"tenant_id"`
	Name            string    `json:"name" db:"name"`
	DurationMinutes int       `json:"durationMinutes" db:"duration_minutes"`
	BufferBefore    int       `json:"bufferBefore" db:"buffer_before_minutes"`
	BufferAfter     int       `json:"bufferAfter" db:"buffer_after_minutes"`
	Enabled         bool      `json:"enabled" db:"enabled"`
	CreatedAt       time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt       time.Time `json:"updatedAt" db:"updated_at"`
}
