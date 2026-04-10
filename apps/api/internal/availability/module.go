package availability

import (
	"database/sql"
	"embed"
	"fmt"

	"github.com/go-chi/chi/v5"

	"github.com/zenvikar/api/internal/platform"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

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
// Currently a placeholder — routes will be added when availability management endpoints are implemented.
func (m *AvailabilityModule) RegisterRoutes(router chi.Router, deps platform.Dependencies) {
	// Routes will be registered when availability management endpoints are implemented.
}

// Migrate runs the availability database migrations.
func (m *AvailabilityModule) Migrate(db *sql.DB) error {
	data, err := migrationsFS.ReadFile("migrations/006_create_availability.sql")
	if err != nil {
		return fmt.Errorf("reading availability migration: %w", err)
	}

	if _, err := db.Exec(string(data)); err != nil {
		return fmt.Errorf("executing availability migration: %w", err)
	}

	return nil
}
