package services

import (
	"database/sql"
	"embed"
	"fmt"

	"github.com/go-chi/chi/v5"

	"github.com/zenvikar/api/internal/platform"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

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
// Currently a placeholder — routes will be added when service management endpoints are implemented.
func (m *ServicesModule) RegisterRoutes(router chi.Router, deps platform.Dependencies) {
	// Routes will be registered when service management endpoints are implemented.
}

// Migrate runs the services database migrations.
func (m *ServicesModule) Migrate(db *sql.DB) error {
	data, err := migrationsFS.ReadFile("migrations/005_create_booking_domain.sql")
	if err != nil {
		return fmt.Errorf("reading services migration: %w", err)
	}

	if _, err := db.Exec(string(data)); err != nil {
		return fmt.Errorf("executing services migration: %w", err)
	}

	return nil
}
