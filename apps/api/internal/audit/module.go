package audit

import (
	"database/sql"

	"github.com/go-chi/chi/v5"

	"github.com/zenvikar/api/internal/platform"
)

// AuditModule implements the platform.Module interface for the audit domain.
type AuditModule struct{}

// New creates a new AuditModule.
func New() *AuditModule {
	return &AuditModule{}
}

// Name returns the module name.
func (m *AuditModule) Name() string {
	return "audit"
}

// RegisterRoutes registers audit-related HTTP routes.
func (m *AuditModule) RegisterRoutes(router chi.Router, deps platform.Dependencies) {
	// Placeholder — routes will be added when audit features are implemented.
}

// Migrate runs the audit database migrations.
func (m *AuditModule) Migrate(db *sql.DB) error {
	return nil
}
