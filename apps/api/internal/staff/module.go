package staff

import (
	"database/sql"

	"github.com/go-chi/chi/v5"

	"github.com/zenvikar/api/internal/platform"
)

// StaffModule implements the platform.Module interface for the staff domain.
type StaffModule struct{}

// New creates a new StaffModule.
func New() *StaffModule {
	return &StaffModule{}
}

// Name returns the module name.
func (m *StaffModule) Name() string {
	return "staff"
}

// RegisterRoutes registers staff-related HTTP routes.
func (m *StaffModule) RegisterRoutes(router chi.Router, deps platform.Dependencies) {
	// Placeholder — routes will be added when staff management features are implemented.
}

// Migrate runs the staff database migrations.
func (m *StaffModule) Migrate(db *sql.DB) error {
	return nil
}
