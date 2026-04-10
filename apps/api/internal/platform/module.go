package platform

import (
	"database/sql"
	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/trace"

	"github.com/zenvikar/api/internal/platform/config"
)

// Module defines the interface that each domain module must implement.
// Modules register their routes and run their database migrations.
type Module interface {
	Name() string
	RegisterRoutes(router chi.Router, deps Dependencies)
	Migrate(db *sql.DB) error
}

// Dependencies holds shared infrastructure dependencies injected into each module.
type Dependencies struct {
	DB     *sql.DB
	Redis  *redis.Client
	Logger *slog.Logger
	Tracer trace.Tracer
	Config *config.Config
}
