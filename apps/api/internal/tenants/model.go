package tenants

import (
	"time"

	"github.com/google/uuid"
)

// Tenant represents a tenant in the multi-tenant platform.
type Tenant struct {
	ID             uuid.UUID `json:"id" db:"id"`
	Slug           string    `json:"slug" db:"slug"`
	DisplayName    string    `json:"displayName" db:"display_name"`
	LogoURL        *string   `json:"logoUrl" db:"logo_url"`
	ColorPrimary   string    `json:"colorPrimary" db:"color_primary"`
	ColorSecondary string    `json:"colorSecondary" db:"color_secondary"`
	ColorAccent    string    `json:"colorAccent" db:"color_accent"`
	Timezone       string    `json:"timezone" db:"timezone"`
	DefaultLocale  string    `json:"defaultLocale" db:"default_locale"`
	Enabled        bool      `json:"enabled" db:"enabled"`
	CreatedAt      time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt      time.Time `json:"updatedAt" db:"updated_at"`
}
