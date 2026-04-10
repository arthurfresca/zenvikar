package authz

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/google/uuid"

	"github.com/zenvikar/api/internal/tenant_memberships"
)

// Sentinel errors for authorization.
var (
	ErrNotAMember              = errors.New("user is not a member of this tenant")
	ErrInsufficientPermissions = errors.New("user does not have the required permission")
)

// RolePermissions defines the static permission matrix for tenant roles.
// tenant_owner gets all permissions via the wildcard "*".
var RolePermissions = map[string][]string{
	tenant_memberships.RoleTenantOwner:         {"*"},
	tenant_memberships.RoleTenantManager:       {"bookings:*", "services:*", "staff:*", "availability:*", "branding:read"},
	tenant_memberships.RoleTenantStaff:         {"bookings:read", "bookings:create", "bookings:update", "services:read", "availability:read"},
	tenant_memberships.RoleTenantFinanceViewer: {"bookings:read", "billing:read", "reports:read"},
}

// PlatformAdminChecker checks if a user is a platform admin.
type PlatformAdminChecker struct {
	db *sql.DB
}

// NewPlatformAdminChecker creates a new platform admin checker.
func NewPlatformAdminChecker(db *sql.DB) *PlatformAdminChecker {
	return &PlatformAdminChecker{db: db}
}

// IsPlatformAdmin checks if the given user is a platform admin.
func (c *PlatformAdminChecker) IsPlatformAdmin(ctx context.Context, userID uuid.UUID) (bool, error) {
	var exists bool
	err := c.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM platform_admins WHERE user_id = $1)`,
		userID.String(),
	).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// Service provides RBAC authorization.
type Service struct {
	adminChecker *PlatformAdminChecker
	memberships  *tenant_memberships.Service
}

// NewService creates a new authorization service.
func NewService(adminChecker *PlatformAdminChecker, memberships *tenant_memberships.Service) *Service {
	return &Service{
		adminChecker: adminChecker,
		memberships:  memberships,
	}
}

// Authorize checks if a user has the required permission for a tenant action.
// Platform admins bypass tenant RBAC entirely.
func (s *Service) Authorize(ctx context.Context, userID, tenantID uuid.UUID, permission string) error {
	// Step 1: Check platform admin (bypasses tenant RBAC).
	isAdmin, err := s.adminChecker.IsPlatformAdmin(ctx, userID)
	if err == nil && isAdmin {
		return nil
	}

	// Step 2: Check tenant membership.
	membership, err := s.memberships.CheckMembership(ctx, userID, tenantID)
	if err != nil {
		return ErrNotAMember
	}

	// Step 3: Check role has permission.
	if !RoleHasPermission(membership.Role, permission) {
		return ErrInsufficientPermissions
	}

	return nil
}

// RoleHasPermission checks whether a role grants the specified permission
// using the static permission matrix.
func RoleHasPermission(role, permission string) bool {
	perms, ok := RolePermissions[role]
	if !ok {
		return false
	}

	for _, p := range perms {
		if p == "*" {
			return true
		}
		if p == permission {
			return true
		}
		// Wildcard match: "bookings:*" matches "bookings:read", "bookings:create", etc.
		if strings.HasSuffix(p, ":*") {
			prefix := strings.TrimSuffix(p, "*")
			if strings.HasPrefix(permission, prefix) {
				return true
			}
		}
	}

	return false
}
