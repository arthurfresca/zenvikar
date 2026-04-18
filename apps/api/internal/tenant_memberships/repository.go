package tenant_memberships

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ErrNotFound indicates the requested membership was not found.
var ErrNotFound = fmt.Errorf("not found")

// MembershipDetails extends a membership with user identity details.
type MembershipDetails struct {
	TenantMembership
	User MembershipUser `json:"user"`
}

// MembershipUser is the user payload embedded in membership details.
type MembershipUser struct {
	ID               uuid.UUID `json:"id"`
	Email            string    `json:"email"`
	Name             string    `json:"name"`
	Phone            *string   `json:"phone"`
	PreferredContact string    `json:"preferredContact"`
	Locale           string    `json:"locale"`
	EmailVerified    bool      `json:"emailVerified"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

// CreateMembershipInput holds membership creation data.
type CreateMembershipInput struct {
	UserID      uuid.UUID
	Role        string
	PhotoURL    *string
	Description *string
}

// UpdateMembershipInput holds optional membership updates.
type UpdateMembershipInput struct {
	Role        *string
	PhotoURL    **string
	Description **string
}

// Repository provides data access for tenant memberships.
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new tenant membership repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// FindByUserAndTenant retrieves a membership by user ID and tenant ID.
// Returns the membership or an error if not found.
func (r *Repository) FindByUserAndTenant(ctx context.Context, userID, tenantID string) (*TenantMembership, error) {
	var m TenantMembership
	var photoURL, description sql.NullString

	err := r.db.QueryRowContext(ctx,
		`SELECT id, tenant_id, user_id, role, photo_url, description, created_at, updated_at
		 FROM tenant_memberships
		 WHERE user_id = $1 AND tenant_id = $2`, userID, tenantID,
	).Scan(
		&m.ID, &m.TenantID, &m.UserID, &m.Role,
		&photoURL, &description,
		&m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("membership not found for user %s in tenant %s", userID, tenantID)
		}
		return nil, fmt.Errorf("querying membership: %w", err)
	}

	if photoURL.Valid {
		m.PhotoURL = &photoURL.String
	}
	if description.Valid {
		m.Description = &description.String
	}

	return &m, nil
}

// ListByTenant returns memberships with user details for tenant management.
func (r *Repository) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]MembershipDetails, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT tm.id, tm.tenant_id, tm.user_id, tm.role, tm.photo_url, tm.description, tm.created_at, tm.updated_at,
		       u.id, u.email, u.name, u.phone, u.preferred_contact, u.locale, u.email_verified, u.created_at, u.updated_at
		FROM tenant_memberships tm
		JOIN users u ON u.id = tm.user_id
		WHERE tm.tenant_id = $1
		ORDER BY u.name ASC
	`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("listing memberships: %w", err)
	}
	defer rows.Close()

	var out []MembershipDetails
	for rows.Next() {
		var item MembershipDetails
		var membershipPhoto, membershipDescription, userPhone sql.NullString
		if err := rows.Scan(
			&item.ID, &item.TenantID, &item.UserID, &item.Role, &membershipPhoto, &membershipDescription, &item.CreatedAt, &item.UpdatedAt,
			&item.User.ID, &item.User.Email, &item.User.Name, &userPhone, &item.User.PreferredContact, &item.User.Locale, &item.User.EmailVerified, &item.User.CreatedAt, &item.User.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning membership: %w", err)
		}
		if membershipPhoto.Valid {
			item.PhotoURL = &membershipPhoto.String
		}
		if membershipDescription.Valid {
			item.Description = &membershipDescription.String
		}
		if userPhone.Valid {
			item.User.Phone = &userPhone.String
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

// GetByTenant returns a membership with user details in tenant scope.
func (r *Repository) GetByTenant(ctx context.Context, tenantID, membershipID uuid.UUID) (*MembershipDetails, error) {
	var item MembershipDetails
	var membershipPhoto, membershipDescription, userPhone sql.NullString
	err := r.db.QueryRowContext(ctx, `
		SELECT tm.id, tm.tenant_id, tm.user_id, tm.role, tm.photo_url, tm.description, tm.created_at, tm.updated_at,
		       u.id, u.email, u.name, u.phone, u.preferred_contact, u.locale, u.email_verified, u.created_at, u.updated_at
		FROM tenant_memberships tm
		JOIN users u ON u.id = tm.user_id
		WHERE tm.tenant_id = $1 AND tm.id = $2
	`, tenantID, membershipID).Scan(
		&item.ID, &item.TenantID, &item.UserID, &item.Role, &membershipPhoto, &membershipDescription, &item.CreatedAt, &item.UpdatedAt,
		&item.User.ID, &item.User.Email, &item.User.Name, &userPhone, &item.User.PreferredContact, &item.User.Locale, &item.User.EmailVerified, &item.User.CreatedAt, &item.User.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("getting membership: %w", err)
	}
	if membershipPhoto.Valid {
		item.PhotoURL = &membershipPhoto.String
	}
	if membershipDescription.Valid {
		item.Description = &membershipDescription.String
	}
	if userPhone.Valid {
		item.User.Phone = &userPhone.String
	}
	return &item, nil
}

// Create inserts a tenant membership.
func (r *Repository) Create(ctx context.Context, tenantID uuid.UUID, input CreateMembershipInput) (*TenantMembership, error) {
	now := time.Now().UTC()
	var item TenantMembership
	var photo, description sql.NullString
	if input.PhotoURL != nil {
		photo = sql.NullString{String: *input.PhotoURL, Valid: true}
	}
	if input.Description != nil {
		description = sql.NullString{String: *input.Description, Valid: true}
	}
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO tenant_memberships (id, tenant_id, user_id, role, photo_url, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, tenant_id, user_id, role, photo_url, description, created_at, updated_at
	`, uuid.New(), tenantID, input.UserID, input.Role, photo, description, now, now).Scan(
		&item.ID, &item.TenantID, &item.UserID, &item.Role, &photo, &description, &item.CreatedAt, &item.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("creating membership: %w", err)
	}
	if photo.Valid {
		item.PhotoURL = &photo.String
	}
	if description.Valid {
		item.Description = &description.String
	}
	return &item, nil
}

// Update updates a membership in tenant scope.
func (r *Repository) Update(ctx context.Context, tenantID, membershipID uuid.UUID, input UpdateMembershipInput) (*TenantMembership, error) {
	current, err := r.FindByID(ctx, tenantID, membershipID)
	if err != nil {
		return nil, err
	}
	if input.Role != nil {
		current.Role = *input.Role
	}
	if input.PhotoURL != nil {
		current.PhotoURL = *input.PhotoURL
	}
	if input.Description != nil {
		current.Description = *input.Description
	}
	current.UpdatedAt = time.Now().UTC()
	var photo, description sql.NullString
	if current.PhotoURL != nil {
		photo = sql.NullString{String: *current.PhotoURL, Valid: true}
	}
	if current.Description != nil {
		description = sql.NullString{String: *current.Description, Valid: true}
	}
	_, err = r.db.ExecContext(ctx, `
		UPDATE tenant_memberships
		SET role = $3, photo_url = $4, description = $5, updated_at = $6
		WHERE tenant_id = $1 AND id = $2
	`, tenantID, membershipID, current.Role, photo, description, current.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("updating membership: %w", err)
	}
	return current, nil
}

// Delete removes a membership in tenant scope.
func (r *Repository) Delete(ctx context.Context, tenantID, membershipID uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM tenant_memberships WHERE tenant_id = $1 AND id = $2`, tenantID, membershipID)
	if err != nil {
		return fmt.Errorf("deleting membership: %w", err)
	}
	rows, err := result.RowsAffected()
	if err == nil && rows == 0 {
		return ErrNotFound
	}
	return nil
}

// FindByID returns a membership in tenant scope.
func (r *Repository) FindByID(ctx context.Context, tenantID, membershipID uuid.UUID) (*TenantMembership, error) {
	var item TenantMembership
	var photo, description sql.NullString
	err := r.db.QueryRowContext(ctx, `
		SELECT id, tenant_id, user_id, role, photo_url, description, created_at, updated_at
		FROM tenant_memberships WHERE tenant_id = $1 AND id = $2
	`, tenantID, membershipID).Scan(&item.ID, &item.TenantID, &item.UserID, &item.Role, &photo, &description, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("finding membership: %w", err)
	}
	if photo.Valid {
		item.PhotoURL = &photo.String
	}
	if description.Valid {
		item.Description = &description.String
	}
	return &item, nil
}
