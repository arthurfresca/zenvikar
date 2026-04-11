package users

import (
	"database/sql"

	"github.com/go-chi/chi/v5"

	"github.com/zenvikar/api/internal/platform"
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
func (m *UsersModule) RegisterRoutes(router chi.Router, deps platform.Dependencies) {}

// Migrate is a no-op — migrations are handled centrally by the migrations package.
func (m *UsersModule) Migrate(db *sql.DB) error {
	return nil
}
