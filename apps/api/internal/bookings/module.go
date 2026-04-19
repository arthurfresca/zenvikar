package bookings

import (
	"database/sql"

	"github.com/go-chi/chi/v5"

	"github.com/zenvikar/api/internal/platform"
	"github.com/zenvikar/api/internal/platform/authn"
	"github.com/zenvikar/api/internal/platform/authz"
	"github.com/zenvikar/api/internal/tenant_memberships"
	"github.com/zenvikar/api/internal/tenants"
)

// BookingsModule implements the platform.Module interface for the bookings domain.
type BookingsModule struct{}

// New creates a new BookingsModule.
func New() *BookingsModule {
	return &BookingsModule{}
}

// Name returns the module name.
func (m *BookingsModule) Name() string {
	return "bookings"
}

// RegisterRoutes registers booking-related HTTP routes.
func (m *BookingsModule) RegisterRoutes(router chi.Router, deps platform.Dependencies) {
	repo := NewRepository(deps.DB)
	bookingSvc := NewBookingService(deps.DB, repo)
	tenantSvc := tenants.NewService(tenants.NewRepository(deps.DB), deps.Redis)
	membershipSvc := tenant_memberships.NewService(tenant_memberships.NewRepository(deps.DB))
	authzSvc := authz.NewService(authz.NewPlatformAdminChecker(deps.DB), membershipSvc)
	h := newHandler(repo, bookingSvc, tenantSvc, authzSvc, membershipSvc)
	h.register(router, authn.RequireAuth(deps.Config.JWTSecret, deps.Config.JWTTTLMinutes))
}

// Migrate is a no-op — migrations are handled centrally by the migrations package.
func (m *BookingsModule) Migrate(db *sql.DB) error {
	return nil
}
