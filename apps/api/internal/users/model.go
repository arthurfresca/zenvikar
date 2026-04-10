package users

import (
	"time"

	"github.com/google/uuid"
)

// User represents a platform user.
type User struct {
	ID            uuid.UUID `json:"id" db:"id"`
	Email         string    `json:"email" db:"email"`
	Name          string    `json:"name" db:"name"`
	PasswordHash  string    `json:"-" db:"password_hash"`
	Locale        string    `json:"locale" db:"locale"`
	EmailVerified bool      `json:"emailVerified" db:"email_verified"`
	CreatedAt     time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt     time.Time `json:"updatedAt" db:"updated_at"`
}
