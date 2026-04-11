package bookings

import (
	"database/sql"

	"github.com/go-chi/chi/v5"

	"github.com/zenvikar/api/internal/platform"
)

// BookingsModule implements the platform.Module interface for the bookings domain.
type BookingsModule struct{}

// New creates a new BookingsModule.
func New() *BookingsModule {
	return &BookingsModule{}
}

// Name returns the module name.
func (m *BookingsModule) Name() string {
	return "bookings"
}

// RegisterRoutes registers booking-related HTTP routes.
func (m *BookingsModule) RegisterRoutes(router chi.Router, deps platform.Dependencies) {}

// Migrate is a no-op — migrations are handled centrally by the migrations package.
func (m *BookingsModule) Migrate(db *sql.DB) error {
	return nil
}
