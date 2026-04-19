package services

import (
	"database/sql"

	"github.com/go-chi/chi/v5"

	"github.com/zenvikar/api/internal/platform"
	"github.com/zenvikar/api/internal/platform/authn"
	"github.com/zenvikar/api/internal/platform/authz"
	"github.com/zenvikar/api/internal/tenant_memberships"
	"github.com/zenvikar/api/internal/tenants"
)

// ServicesModule implements the platform.Module interface for the services domain.
type ServicesModule struct{}

// New creates a new ServicesModule.
func New() *ServicesModule {
	return &ServicesModule{}
}

// Name returns the module name.
func (m *ServicesModule) Name() string {
	return "services"
}

// RegisterRoutes registers service-related HTTP routes.
func (m *ServicesModule) RegisterRoutes(router chi.Router, deps platform.Dependencies) {
	repo := NewRepository(deps.DB)
	tenantSvc := tenants.NewService(tenants.NewRepository(deps.DB), deps.Redis)
	membershipSvc := tenant_memberships.NewService(tenant_memberships.NewRepository(deps.DB))
	authzSvc := authz.NewService(authz.NewPlatformAdminChecker(deps.DB), membershipSvc)
	h := newHandler(repo, tenantSvc, authzSvc, membershipSvc)
	h.register(router, authn.RequireAuth(deps.Config.JWTSecret, deps.Config.JWTTTLMinutes))
}

// Migrate is a no-op — migrations are handled centrally by the migrations package.
func (m *ServicesModule) Migrate(db *sql.DB) error {
	return nil
}
