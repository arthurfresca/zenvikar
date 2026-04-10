package users

import (
	"time"

	"github.com/google/uuid"
)

// Platform admin role constants.
const (
	RoleAdmin        = "admin"
	RoleSupportAdmin = "support_admin"
	RoleFinanceAdmin = "finance_admin"
)

// PlatformAdmin represents a user with platform-level administrative privileges.
type PlatformAdmin struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"userId" db:"user_id"`
	Role      string    `json:"role" db:"role"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
}
