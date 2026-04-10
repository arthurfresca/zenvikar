package users

import (
	"database/sql"
	"embed"
	"fmt"

	"github.com/go-chi/chi/v5"

	"github.com/zenvikar/api/internal/platform"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// UsersModule implements the platform.Module interface for the users domain.
type UsersModule struct{}

// New creates a new UsersModule.
func New() *UsersModule {
	return &UsersModule{}
}

// Name returns the module name.
func (m *UsersModule) Name() string {
	return "users"
}

// RegisterRoutes registers user-related HTTP routes.
// Currently a placeholder — routes will be added when auth is implemented.
func (m *UsersModule) RegisterRoutes(router chi.Router, deps platform.Dependencies) {
	// Routes will be registered when user management endpoints are implemented.
}

// Migrate runs the users database migrations.
func (m *UsersModule) Migrate(db *sql.DB) error {
	migrations := []string{
		"migrations/002_create_users.sql",
		"migrations/004_create_platform_admins.sql",
	}

	for _, file := range migrations {
		data, err := migrationsFS.ReadFile(file)
		if err != nil {
			return fmt.Errorf("reading %s: %w", file, err)
		}
		if _, err := db.Exec(string(data)); err != nil {
			return fmt.Errorf("executing %s: %w", file, err)
		}
	}

	return nil
}
