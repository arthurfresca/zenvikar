package notifications

import (
	"database/sql"

	"github.com/go-chi/chi/v5"

	"github.com/zenvikar/api/internal/platform"
)

// NotificationsModule implements the platform.Module interface for the notifications domain.
type NotificationsModule struct{}

// New creates a new NotificationsModule.
func New() *NotificationsModule {
	return &NotificationsModule{}
}

// Name returns the module name.
func (m *NotificationsModule) Name() string {
	return "notifications"
}

// RegisterRoutes registers notification-related HTTP routes.
func (m *NotificationsModule) RegisterRoutes(router chi.Router, deps platform.Dependencies) {
	// Placeholder — routes will be added when notification features are implemented.
}

// Migrate runs the notifications database migrations.
func (m *NotificationsModule) Migrate(db *sql.DB) error {
	return nil
}
