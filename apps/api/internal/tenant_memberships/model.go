package tenant_memberships

import (
	"time"

	"github.com/google/uuid"
)

// Tenant role constants.
const (
	RoleTenantOwner         = "tenant_owner"
	RoleTenantManager       = "tenant_manager"
	RoleTenantStaff         = "tenant_staff"
	RoleTenantFinanceViewer = "tenant_finance_viewer"
)

// TenantMembership represents a user's membership and role within a tenant.
type TenantMembership struct {
	ID        uuid.UUID `json:"id" db:"id"`
	TenantID  uuid.UUID `json:"tenantId" db:"tenant_id"`
	UserID    uuid.UUID `json:"userId" db:"user_id"`
	Role      string    `json:"role" db:"role"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
}
