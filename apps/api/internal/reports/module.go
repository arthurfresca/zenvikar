package reports

import (
	"database/sql"

	"github.com/go-chi/chi/v5"

	"github.com/zenvikar/api/internal/platform"
)

// ReportsModule implements the platform.Module interface for the reports domain.
type ReportsModule struct{}

// New creates a new ReportsModule.
func New() *ReportsModule {
	return &ReportsModule{}
}

// Name returns the module name.
func (m *ReportsModule) Name() string {
	return "reports"
}

// RegisterRoutes registers reports-related HTTP routes.
func (m *ReportsModule) RegisterRoutes(router chi.Router, deps platform.Dependencies) {
	// Placeholder — routes will be added when reporting features are implemented.
}

// Migrate runs the reports database migrations.
func (m *ReportsModule) Migrate(db *sql.DB) error {
	return nil
}
