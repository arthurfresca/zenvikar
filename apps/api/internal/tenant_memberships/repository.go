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

	err := r.db.QueryRowContext(ctx,
		`SELECT id, tenant_id, user_id, role, created_at, updated_at
		 FROM tenant_memberships
		 WHERE user_id = $1 AND tenant_id = $2`, userID, tenantID,
	).Scan(
		&m.ID, &m.TenantID, &m.UserID, &m.Role,
		&m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("membership not found for user %s in tenant %s", userID, tenantID)
		}
		return nil, fmt.Errorf("querying membership: %w", err)
	}

	return &m, nil
}
