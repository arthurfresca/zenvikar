package availability

import (
	"database/sql"

	"github.com/go-chi/chi/v5"

	"github.com/zenvikar/api/internal/platform"
	"github.com/zenvikar/api/internal/platform/authn"
	"github.com/zenvikar/api/internal/platform/authz"
	"github.com/zenvikar/api/internal/tenant_memberships"
)

// AvailabilityModule implements the platform.Module interface for the availability domain.
type AvailabilityModule struct{}

// New creates a new AvailabilityModule.
func New() *AvailabilityModule {
	return &AvailabilityModule{}
}

// Name returns the module name.
func (m *AvailabilityModule) Name() string {
	return "availability"
}

// RegisterRoutes registers availability-related HTTP routes.
func (m *AvailabilityModule) RegisterRoutes(router chi.Router, deps platform.Dependencies) {
	repo := NewRepository(deps.DB)
	membershipSvc := tenant_memberships.NewService(tenant_memberships.NewRepository(deps.DB))
	authzSvc := authz.NewService(authz.NewPlatformAdminChecker(deps.DB), membershipSvc)
	h := newHandler(repo, authzSvc)
	h.register(router, authn.RequireAuth(deps.Config.JWTSecret, deps.Config.JWTTTLMinutes))
}

// Migrate is a no-op — migrations are handled centrally by the migrations package.
func (m *AvailabilityModule) Migrate(db *sql.DB) error {
	return nil
}
