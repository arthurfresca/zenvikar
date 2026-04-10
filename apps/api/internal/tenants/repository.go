package tenants

import (
	"context"
	"database/sql"
	"fmt"
)

// Repository provides data access for tenants.
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new tenant repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// FindBySlug retrieves a tenant by its slug.
// Returns the tenant or an error if not found or a database error occurs.
func (r *Repository) FindBySlug(ctx context.Context, slug string) (*Tenant, error) {
	var t Tenant
	var logoURL sql.NullString

	err := r.db.QueryRowContext(ctx,
		`SELECT id, slug, display_name, logo_url, color_primary, color_secondary,
		        color_accent, timezone, default_locale, enabled, created_at, updated_at
		 FROM tenants WHERE slug = $1`, slug,
	).Scan(
		&t.ID, &t.Slug, &t.DisplayName, &logoURL,
		&t.ColorPrimary, &t.ColorSecondary, &t.ColorAccent,
		&t.Timezone, &t.DefaultLocale, &t.Enabled,
		&t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("tenant with slug %q not found", slug)
		}
		return nil, fmt.Errorf("querying tenant by slug: %w", err)
	}

	if logoURL.Valid {
		t.LogoURL = &logoURL.String
	}

	return &t, nil
}
