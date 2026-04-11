package availability

import (
	"database/sql"

	"github.com/go-chi/chi/v5"

	"github.com/zenvikar/api/internal/platform"
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
func (m *AvailabilityModule) RegisterRoutes(router chi.Router, deps platform.Dependencies) {}

// Migrate is a no-op — migrations are handled centrally by the migrations package.
func (m *AvailabilityModule) Migrate(db *sql.DB) error {
	return nil
}
