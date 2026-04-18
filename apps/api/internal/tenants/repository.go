package tenants

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ErrNotFound indicates the tenant was not found.
var ErrNotFound = fmt.Errorf("not found")

// UpdateTenantInput holds optional tenant updates.
type UpdateTenantInput struct {
	DisplayName         *string
	LogoURL             **string
	ColorPrimary        *string
	ColorSecondary      *string
	ColorAccent         *string
	Phone               **string
	Email               **string
	Address             **string
	Currency            *string
	SlotIntervalMinutes *int
	Timezone            *string
	DefaultLocale       *string
	Enabled             *bool
}

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
	var logoURL, phone, email, address sql.NullString

	err := r.db.QueryRowContext(ctx,
		`SELECT id, slug, display_name, logo_url, color_primary, color_secondary,
		        color_accent, phone, email, address, currency, slot_interval_minutes,
		        timezone, default_locale, enabled, created_at, updated_at
		 FROM tenants WHERE slug = $1`, slug,
	).Scan(
		&t.ID, &t.Slug, &t.DisplayName, &logoURL,
		&t.ColorPrimary, &t.ColorSecondary, &t.ColorAccent,
		&phone, &email, &address, &t.Currency, &t.SlotIntervalMinutes,
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
	if phone.Valid {
		t.Phone = &phone.String
	}
	if email.Valid {
		t.Email = &email.String
	}
	if address.Valid {
		t.Address = &address.String
	}

	return &t, nil
}

// FindByID retrieves a tenant by ID.
func (r *Repository) FindByID(ctx context.Context, tenantID uuid.UUID) (*Tenant, error) {
	var t Tenant
	var logoURL, phone, email, address sql.NullString
	err := r.db.QueryRowContext(ctx, `
		SELECT id, slug, display_name, logo_url, color_primary, color_secondary,
		       color_accent, phone, email, address, currency, slot_interval_minutes,
		       timezone, default_locale, enabled, created_at, updated_at
		FROM tenants WHERE id = $1
	`, tenantID).Scan(
		&t.ID, &t.Slug, &t.DisplayName, &logoURL, &t.ColorPrimary, &t.ColorSecondary,
		&t.ColorAccent, &phone, &email, &address, &t.Currency, &t.SlotIntervalMinutes,
		&t.Timezone, &t.DefaultLocale, &t.Enabled, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("querying tenant by id: %w", err)
	}
	if logoURL.Valid {
		t.LogoURL = &logoURL.String
	}
	if phone.Valid {
		t.Phone = &phone.String
	}
	if email.Valid {
		t.Email = &email.String
	}
	if address.Valid {
		t.Address = &address.String
	}
	return &t, nil
}

// Update updates tenant settings in place.
func (r *Repository) Update(ctx context.Context, tenantID uuid.UUID, input UpdateTenantInput) (*Tenant, error) {
	current, err := r.FindByID(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	if input.DisplayName != nil {
		current.DisplayName = *input.DisplayName
	}
	if input.LogoURL != nil {
		current.LogoURL = *input.LogoURL
	}
	if input.ColorPrimary != nil {
		current.ColorPrimary = *input.ColorPrimary
	}
	if input.ColorSecondary != nil {
		current.ColorSecondary = *input.ColorSecondary
	}
	if input.ColorAccent != nil {
		current.ColorAccent = *input.ColorAccent
	}
	if input.Phone != nil {
		current.Phone = *input.Phone
	}
	if input.Email != nil {
		current.Email = *input.Email
	}
	if input.Address != nil {
		current.Address = *input.Address
	}
	if input.Currency != nil {
		current.Currency = *input.Currency
	}
	if input.SlotIntervalMinutes != nil {
		current.SlotIntervalMinutes = *input.SlotIntervalMinutes
	}
	if input.Timezone != nil {
		current.Timezone = *input.Timezone
	}
	if input.DefaultLocale != nil {
		current.DefaultLocale = *input.DefaultLocale
	}
	if input.Enabled != nil {
		current.Enabled = *input.Enabled
	}
	current.UpdatedAt = time.Now().UTC()

	var logoURL, phone, email, address sql.NullString
	if current.LogoURL != nil {
		logoURL = sql.NullString{String: *current.LogoURL, Valid: true}
	}
	if current.Phone != nil {
		phone = sql.NullString{String: *current.Phone, Valid: true}
	}
	if current.Email != nil {
		email = sql.NullString{String: *current.Email, Valid: true}
	}
	if current.Address != nil {
		address = sql.NullString{String: *current.Address, Valid: true}
	}
	_, err = r.db.ExecContext(ctx, `
		UPDATE tenants
		SET display_name = $2, logo_url = $3, color_primary = $4, color_secondary = $5,
		    color_accent = $6, phone = $7, email = $8, address = $9, currency = $10,
		    slot_interval_minutes = $11, timezone = $12, default_locale = $13, enabled = $14, updated_at = $15
		WHERE id = $1
	`, tenantID, current.DisplayName, logoURL, current.ColorPrimary, current.ColorSecondary, current.ColorAccent, phone, email, address, current.Currency, current.SlotIntervalMinutes, current.Timezone, current.DefaultLocale, current.Enabled, current.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("updating tenant: %w", err)
	}
	return current, nil
}
