package users

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

var errNoRows = errors.New("not found")

// Repository provides data access for user/auth operations.
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new users repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// CreateUserParams holds the values required to create a user.
type CreateUserParams struct {
	Email            string
	Name             string
	PasswordHash     *string
	Phone            *string
	PreferredContact string
	Locale           string
	EmailVerified    bool
}

// TenantAccess represents a user's membership context with tenant details.
type TenantAccess struct {
	TenantID   string `json:"tenantId"`
	TenantSlug string `json:"tenantSlug"`
	TenantName string `json:"tenantName"`
	Role       string `json:"role"`
}

// FindUserByEmail returns a user by email.
func (r *Repository) FindUserByEmail(ctx context.Context, email string) (*User, error) {
	var u User
	var passwordHash, phone sql.NullString

	err := r.db.QueryRowContext(ctx, `
		SELECT id, email, name, password_hash, phone, preferred_contact, locale, email_verified, created_at, updated_at
		FROM users
		WHERE LOWER(email) = LOWER($1)
	`, strings.TrimSpace(email)).Scan(
		&u.ID,
		&u.Email,
		&u.Name,
		&passwordHash,
		&phone,
		&u.PreferredContact,
		&u.Locale,
		&u.EmailVerified,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errNoRows
		}
		return nil, fmt.Errorf("finding user by email: %w", err)
	}

	if passwordHash.Valid {
		u.PasswordHash = &passwordHash.String
	}
	if phone.Valid {
		u.Phone = &phone.String
	}

	return &u, nil
}

// FindUserByProvider returns a user by external auth provider identity.
func (r *Repository) FindUserByProvider(ctx context.Context, provider, providerID string) (*User, error) {
	var u User
	var passwordHash, phone sql.NullString

	err := r.db.QueryRowContext(ctx, `
		SELECT u.id, u.email, u.name, u.password_hash, u.phone, u.preferred_contact, u.locale, u.email_verified, u.created_at, u.updated_at
		FROM users u
		JOIN user_auth_providers p ON p.user_id = u.id
		WHERE p.provider = $1 AND p.provider_id = $2
	`, provider, providerID).Scan(
		&u.ID,
		&u.Email,
		&u.Name,
		&passwordHash,
		&phone,
		&u.PreferredContact,
		&u.Locale,
		&u.EmailVerified,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errNoRows
		}
		return nil, fmt.Errorf("finding user by provider: %w", err)
	}

	if passwordHash.Valid {
		u.PasswordHash = &passwordHash.String
	}
	if phone.Valid {
		u.Phone = &phone.String
	}

	return &u, nil
}

// CreateUser inserts and returns a new user.
func (r *Repository) CreateUser(ctx context.Context, params CreateUserParams) (*User, error) {
	var u User
	var passwordHash, phone sql.NullString

	if params.PasswordHash != nil {
		passwordHash = sql.NullString{String: *params.PasswordHash, Valid: true}
	}
	if params.Phone != nil {
		phone = sql.NullString{String: *params.Phone, Valid: true}
	}

	err := r.db.QueryRowContext(ctx, `
		INSERT INTO users (email, name, password_hash, phone, preferred_contact, locale, email_verified)
		VALUES (LOWER($1), $2, $3, $4, $5, $6, $7)
		RETURNING id, email, name, password_hash, phone, preferred_contact, locale, email_verified, created_at, updated_at
	`, strings.TrimSpace(params.Email), strings.TrimSpace(params.Name), passwordHash, phone, params.PreferredContact, params.Locale, params.EmailVerified).Scan(
		&u.ID,
		&u.Email,
		&u.Name,
		&passwordHash,
		&phone,
		&u.PreferredContact,
		&u.Locale,
		&u.EmailVerified,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("creating user: %w", err)
	}

	if passwordHash.Valid {
		u.PasswordHash = &passwordHash.String
	}
	if phone.Valid {
		u.Phone = &phone.String
	}

	return &u, nil
}

// CreateAuthProvider links a user to an auth provider.
func (r *Repository) CreateAuthProvider(ctx context.Context, userID, provider, providerID string) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO user_auth_providers (user_id, provider, provider_id)
		VALUES ($1, $2, $3)
		ON CONFLICT (provider, provider_id) DO NOTHING
	`, userID, provider, providerID)
	if err != nil {
		return fmt.Errorf("creating auth provider link: %w", err)
	}
	return nil
}

// FindPlatformRole returns platform role for a user, empty when not admin.
func (r *Repository) FindPlatformRole(ctx context.Context, userID string) (string, error) {
	var role string
	err := r.db.QueryRowContext(ctx, `
		SELECT role
		FROM platform_admins
		WHERE user_id = $1
	`, userID).Scan(&role)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", fmt.Errorf("finding platform role: %w", err)
	}
	return role, nil
}

// ListTenantRoles returns map of tenantID -> role for a user.
func (r *Repository) ListTenantRoles(ctx context.Context, userID string) (map[string]string, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT tenant_id::text, role
		FROM tenant_memberships
		WHERE user_id = $1
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("listing tenant roles: %w", err)
	}
	defer rows.Close()

	out := map[string]string{}
	for rows.Next() {
		var tenantID, role string
		if err := rows.Scan(&tenantID, &role); err != nil {
			return nil, fmt.Errorf("scanning tenant role: %w", err)
		}
		out[tenantID] = role
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating tenant roles: %w", err)
	}
	return out, nil
}

// FindTenantRoleForUser returns user role in a tenant, empty when not a member.
func (r *Repository) FindTenantRoleForUser(ctx context.Context, userID, tenantID string) (string, error) {
	var role string
	err := r.db.QueryRowContext(ctx, `
		SELECT role
		FROM tenant_memberships
		WHERE user_id = $1 AND tenant_id = $2
	`, userID, tenantID).Scan(&role)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", fmt.Errorf("finding tenant role for user: %w", err)
	}
	return role, nil
}

// ListUserTenantAccess returns all tenants a user can access.
func (r *Repository) ListUserTenantAccess(ctx context.Context, userID string) ([]TenantAccess, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT t.id::text, t.slug, t.display_name, tm.role
		FROM tenant_memberships tm
		JOIN tenants t ON t.id = tm.tenant_id
		WHERE tm.user_id = $1
		ORDER BY t.display_name ASC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("listing user tenant access: %w", err)
	}
	defer rows.Close()

	var out []TenantAccess
	for rows.Next() {
		var item TenantAccess
		if err := rows.Scan(&item.TenantID, &item.TenantSlug, &item.TenantName, &item.Role); err != nil {
			return nil, fmt.Errorf("scanning user tenant access: %w", err)
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating user tenant access: %w", err)
	}
	return out, nil
}
