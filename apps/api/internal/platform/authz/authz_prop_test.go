package authz

import (
	"testing"

	"github.com/zenvikar/api/internal/tenant_memberships"
	"pgregory.net/rapid"
)

// **Validates: Requirements 5.5, 5.6, 5.7**

// allTenantRoles is the complete set of tenant roles.
var allTenantRoles = []string{
	tenant_memberships.RoleTenantOwner,
	tenant_memberships.RoleTenantManager,
	tenant_memberships.RoleTenantStaff,
	tenant_memberships.RoleTenantFinanceViewer,
}

// allPermissions is the complete set of permissions used in the system.
var allPermissions = []string{
	"bookings:read",
	"bookings:create",
	"bookings:update",
	"bookings:delete",
	"services:read",
	"services:create",
	"services:update",
	"services:delete",
	"staff:read",
	"staff:create",
	"staff:update",
	"staff:delete",
	"availability:read",
	"availability:create",
	"availability:update",
	"availability:delete",
	"branding:read",
	"branding:update",
	"billing:read",
	"billing:update",
	"reports:read",
	"reports:update",
}

// expectedPermissions defines the ground-truth permission matrix for verification.
// This is intentionally separate from RolePermissions to serve as an independent oracle.
var expectedPermissions = map[string]map[string]bool{
	tenant_memberships.RoleTenantOwner: func() map[string]bool {
		// tenant_owner has ALL permissions
		m := make(map[string]bool)
		for _, p := range allPermissions {
			m[p] = true
		}
		return m
	}(),
	tenant_memberships.RoleTenantManager: {
		"bookings:read":       true,
		"bookings:create":     true,
		"bookings:update":     true,
		"bookings:delete":     true,
		"services:read":       true,
		"services:create":     true,
		"services:update":     true,
		"services:delete":     true,
		"staff:read":          true,
		"staff:create":        true,
		"staff:update":        true,
		"staff:delete":        true,
		"availability:read":   true,
		"availability:create": true,
		"availability:update": true,
		"availability:delete": true,
		"branding:read":       true,
	},
	tenant_memberships.RoleTenantStaff: {
		"bookings:read":    true,
		"bookings:create":  true,
		"bookings:update":  true,
		"services:read":    true,
		"availability:read": true,
	},
	tenant_memberships.RoleTenantFinanceViewer: {
		"bookings:read": true,
		"billing:read":  true,
		"reports:read":  true,
	},
}

// TestProperty5_RBACPermissionMatrixCorrectness verifies that for any (role, permission)
// combination, RoleHasPermission matches the expected static permission matrix.
// Platform admins always authorized; tenant_owner has all permissions.
func TestProperty5_RBACPermissionMatrixCorrectness(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		role := rapid.SampledFrom(allTenantRoles).Draw(t, "role")
		permission := rapid.SampledFrom(allPermissions).Draw(t, "permission")

		got := RoleHasPermission(role, permission)
		want := expectedPermissions[role][permission]

		if got != want {
			t.Fatalf("RoleHasPermission(%q, %q) = %v, want %v", role, permission, got, want)
		}
	})
}

// TestProperty5_TenantOwnerHasAllPermissions verifies that tenant_owner
// is granted every permission in the system.
func TestProperty5_TenantOwnerHasAllPermissions(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		permission := rapid.SampledFrom(allPermissions).Draw(t, "permission")

		if !RoleHasPermission(tenant_memberships.RoleTenantOwner, permission) {
			t.Fatalf("tenant_owner should have permission %q but was denied", permission)
		}
	})
}

// TestProperty5_UnknownRoleHasNoPermissions verifies that an unknown role
// is denied all permissions.
func TestProperty5_UnknownRoleHasNoPermissions(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		role := rapid.StringMatching(`[a-z_]{5,20}`).Draw(t, "unknownRole")
		// Skip if it happens to be a real role
		for _, r := range allTenantRoles {
			if role == r {
				return
			}
		}

		permission := rapid.SampledFrom(allPermissions).Draw(t, "permission")

		if RoleHasPermission(role, permission) {
			t.Fatalf("unknown role %q should not have permission %q", role, permission)
		}
	})
}
