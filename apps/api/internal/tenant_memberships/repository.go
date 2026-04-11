package tenant_memberships

import (
	"context"
	"database/sql"
	"fmt"
)

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
