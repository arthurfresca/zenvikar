package users

import (
	"time"

	"github.com/google/uuid"
)

// PreferredContact represents the user's preferred contact method.
const (
	ContactEmail    = "email"
	ContactPhone    = "phone"
	ContactWhatsApp = "whatsapp"
)

// User represents a platform user.
type User struct {
	ID               uuid.UUID `json:"id" db:"id"`
	Email            string    `json:"email" db:"email"`
	Name             string    `json:"name" db:"name"`
	PasswordHash     *string   `json:"-" db:"password_hash"`
	Phone            *string   `json:"phone" db:"phone"`
	PreferredContact string    `json:"preferredContact" db:"preferred_contact"`
	Locale           string    `json:"locale" db:"locale"`
	EmailVerified    bool      `json:"emailVerified" db:"email_verified"`
	CreatedAt        time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt        time.Time `json:"updatedAt" db:"updated_at"`
}

// Auth provider constants.
const (
	AuthProviderEmail    = "email"
	AuthProviderGoogle   = "google"
	AuthProviderFacebook = "facebook"
)

// UserAuthProvider links a user to an external authentication provider.
type UserAuthProvider struct {
	ID         uuid.UUID `json:"id" db:"id"`
	UserID     uuid.UUID `json:"userId" db:"user_id"`
	Provider   string    `json:"provider" db:"provider"`
	ProviderID string    `json:"providerId" db:"provider_id"`
	CreatedAt  time.Time `json:"createdAt" db:"created_at"`
}
