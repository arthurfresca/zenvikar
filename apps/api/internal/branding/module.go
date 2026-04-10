package branding

import (
	"database/sql"

	"github.com/go-chi/chi/v5"

	"github.com/zenvikar/api/internal/platform"
)

// BrandingModule implements the platform.Module interface for the branding domain.
type BrandingModule struct{}

// New creates a new BrandingModule.
func New() *BrandingModule {
	return &BrandingModule{}
}

// Name returns the module name.
func (m *BrandingModule) Name() string {
	return "branding"
}

// RegisterRoutes registers branding-related HTTP routes.
func (m *BrandingModule) RegisterRoutes(router chi.Router, deps platform.Dependencies) {
	// Placeholder — routes will be added when branding features are implemented.
}

// Migrate runs the branding database migrations.
func (m *BrandingModule) Migrate(db *sql.DB) error {
	return nil
}
