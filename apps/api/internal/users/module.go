package users

import (
	"database/sql"

	"github.com/go-chi/chi/v5"

	"github.com/zenvikar/api/internal/platform"
	"github.com/zenvikar/api/internal/tenants"
)

// UsersModule implements the platform.Module interface for the users domain.
type UsersModule struct{}

// New creates a new UsersModule.
func New() *UsersModule {
	return &UsersModule{}
}

// Name returns the module name.
func (m *UsersModule) Name() string {
	return "users"
}

// RegisterRoutes registers user-related HTTP routes.
func (m *UsersModule) RegisterRoutes(router chi.Router, deps platform.Dependencies) {
	repo := NewRepository(deps.DB)
	tokenManager := NewTokenManager(deps.Config.JWTSecret, deps.Config.JWTTTLMinutes)
	tenantSvc := tenants.NewService(tenants.NewRepository(deps.DB), deps.Redis)
	svc := NewService(repo, tokenManager, tenantSvc)
	handler := newAuthHandler(svc, deps.Config)

	router.Post("/api/v1/auth/signup", handler.signup)
	router.Post("/api/v1/auth/login", handler.login)
	router.Post("/api/v1/auth/social", handler.socialLogin)
	router.Post("/api/v1/auth/admin/login", handler.adminLogin)
	router.Post("/api/v1/auth/tenant/login", handler.tenantLogin)
	router.Get("/api/v1/auth/me", handler.me)
	router.Get("/api/v1/auth/tenants", handler.tenants)
}

// Migrate is a no-op — migrations are handled centrally by the migrations package.
func (m *UsersModule) Migrate(db *sql.DB) error {
	return nil
}
