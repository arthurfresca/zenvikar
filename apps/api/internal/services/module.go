package services

import (
	"database/sql"

	"github.com/go-chi/chi/v5"

	"github.com/zenvikar/api/internal/platform"
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
func (m *ServicesModule) RegisterRoutes(router chi.Router, deps platform.Dependencies) {}

// Migrate is a no-op — migrations are handled centrally by the migrations package.
func (m *ServicesModule) Migrate(db *sql.DB) error {
	return nil
}
