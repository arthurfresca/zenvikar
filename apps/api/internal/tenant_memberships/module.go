package tenant_memberships

import (
	"database/sql"

	"github.com/go-chi/chi/v5"

	"github.com/zenvikar/api/internal/platform"
	"github.com/zenvikar/api/internal/platform/authn"
)

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
func (m *TenantMembershipsModule) RegisterRoutes(router chi.Router, deps platform.Dependencies) {
	repo := NewRepository(deps.DB)
	svc := NewService(repo)
	h := newHandler(svc, deps.DB)
	h.register(router, authn.RequireAuth(deps.Config.JWTSecret, deps.Config.JWTTTLMinutes))
}

// Migrate is a no-op — migrations are handled centrally by the migrations package.
func (m *TenantMembershipsModule) Migrate(db *sql.DB) error {
	return nil
}
