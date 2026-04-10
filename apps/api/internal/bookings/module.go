package bookings

import (
	"database/sql"
	"embed"
	"fmt"

	"github.com/go-chi/chi/v5"

	"github.com/zenvikar/api/internal/platform"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

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
// Currently a placeholder — routes will be added when booking endpoints are implemented.
func (m *BookingsModule) RegisterRoutes(router chi.Router, deps platform.Dependencies) {
	// Routes will be registered when booking endpoints are implemented.
}

// Migrate runs the bookings database migrations.
func (m *BookingsModule) Migrate(db *sql.DB) error {
	data, err := migrationsFS.ReadFile("migrations/007_create_bookings.sql")
	if err != nil {
		return fmt.Errorf("reading bookings migration: %w", err)
	}

	if _, err := db.Exec(string(data)); err != nil {
		return fmt.Errorf("executing bookings migration: %w", err)
	}

	return nil
}
