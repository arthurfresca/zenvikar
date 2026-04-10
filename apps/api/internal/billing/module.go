package billing

import (
	"database/sql"

	"github.com/go-chi/chi/v5"

	"github.com/zenvikar/api/internal/platform"
)

// BillingModule implements the platform.Module interface for the billing domain.
type BillingModule struct{}

// New creates a new BillingModule.
func New() *BillingModule {
	return &BillingModule{}
}

// Name returns the module name.
func (m *BillingModule) Name() string {
	return "billing"
}

// RegisterRoutes registers billing-related HTTP routes.
func (m *BillingModule) RegisterRoutes(router chi.Router, deps platform.Dependencies) {
	// Placeholder — routes will be added when billing features are implemented.
}

// Migrate runs the billing database migrations.
func (m *BillingModule) Migrate(db *sql.DB) error {
	return nil
}
