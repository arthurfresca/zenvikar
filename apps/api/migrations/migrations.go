// Package migrations provides centralized database migration management.
// All SQL migration files are embedded and executed in numeric order.
package migrations

import (
	"database/sql"
	"embed"
	"fmt"
	"log/slog"
	"sort"
)

//go:embed *.sql
var sqlFiles embed.FS

// RunAll executes all SQL migration files in numeric order.
// Files are sorted by name (001_, 002_, etc.) to ensure correct ordering.
func RunAll(db *sql.DB, logger *slog.Logger) error {
	entries, err := sqlFiles.ReadDir(".")
	if err != nil {
		return fmt.Errorf("reading migration files: %w", err)
	}

	// Collect and sort SQL files
	var files []string
	for _, e := range entries {
		if !e.IsDir() {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)

	for _, file := range files {
		data, err := sqlFiles.ReadFile(file)
		if err != nil {
			return fmt.Errorf("reading %s: %w", file, err)
		}

		if _, err := db.Exec(string(data)); err != nil {
			return fmt.Errorf("executing %s: %w", file, err)
		}

		logger.Info("migration applied", "file", file)
	}

	return nil
}
