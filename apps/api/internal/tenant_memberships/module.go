package tenant_memberships

import (
	"database/sql"
	"embed"
	"fmt"

	"github.com/go-chi/chi/v5"

	"github.com/zenvikar/api/internal/platform"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// TenantMembershipsModule implements the platform.Module interface for tenant memberships.
type TenantMembershipsModule struct{}

// New creates a new TenantMembershipsModule.
func New() *TenantMembershipsModule {
	return &TenantMembershipsModule{}
}

// Name returns the module name.
func (m *TenantMembershipsModule) Name() string {
	return "tenant_memberships"
}

// RegisterRoutes registers tenant membership HTTP routes.
// Currently a placeholder — routes will be added when membership management endpoints are implemented.
func (m *TenantMembershipsModule) RegisterRoutes(router chi.Router, deps platform.Dependencies) {
	// Routes will be registered when membership management endpoints are implemented.
}

// Migrate runs the tenant memberships database migrations.
func (m *TenantMembershipsModule) Migrate(db *sql.DB) error {
	data, err := migrationsFS.ReadFile("migrations/003_create_tenant_memberships.sql")
	if err != nil {
		return fmt.Errorf("reading tenant_memberships migration: %w", err)
	}

	if _, err := db.Exec(string(data)); err != nil {
		return fmt.Errorf("executing tenant_memberships migration: %w", err)
	}

	return nil
}
