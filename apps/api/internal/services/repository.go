package services

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ErrNotFound indicates the requested service resource does not exist.
var ErrNotFound = fmt.Errorf("not found")

// ServiceMemberDetails contains service-member assignment data with member identity.
type ServiceMemberDetails struct {
	ServiceMember
	MemberName  string    `json:"memberName"`
	MemberEmail string    `json:"memberEmail"`
	TenantID    uuid.UUID `json:"tenantId"`
}

// PublicService contains a public-facing service with bookable members.
type PublicService struct {
	Service
	Members []ServiceMemberDetails `json:"members"`
}

// CreateServiceInput holds service creation data.
type CreateServiceInput struct {
	Name            string
	Description     *string
	DurationMinutes int
	BufferBefore    int
	BufferAfter     int
	Enabled         bool
}

// UpdateServiceInput holds optional service updates.
type UpdateServiceInput struct {
	Name            *string
	Description     **string
	DurationMinutes *int
	BufferBefore    *int
	BufferAfter     *int
	Enabled         *bool
}

// AddServiceMemberInput holds service-member assignment data.
type AddServiceMemberInput struct {
	MembershipID uuid.UUID
	PriceCents   int
	Description  *string
}

// Repository provides service and service-member persistence.
type Repository struct {
	db *sql.DB
}

// NewRepository creates a services repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// ListPublicByTenantSlug returns enabled services and members for a tenant slug.
func (r *Repository) ListPublicByTenantSlug(ctx context.Context, slug string) ([]PublicService, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT s.id, s.tenant_id, s.name, s.description, s.duration_minutes,
		       s.buffer_before_minutes, s.buffer_after_minutes, s.enabled,
		       s.created_at, s.updated_at,
		       sm.id, sm.membership_id, sm.price_cents, sm.description,
		       u.name, u.email
		FROM tenants t
		JOIN services s ON s.tenant_id = t.id
		JOIN service_members sm ON sm.service_id = s.id
		JOIN tenant_memberships tm ON tm.id = sm.membership_id
		JOIN users u ON u.id = tm.user_id
		WHERE t.slug = $1 AND t.enabled = true AND s.enabled = true
		ORDER BY s.name ASC, u.name ASC
	`, slug)
	if err != nil {
		return nil, fmt.Errorf("listing public services: %w", err)
	}
	defer rows.Close()

	servicesByID := map[uuid.UUID]*PublicService{}
	order := make([]uuid.UUID, 0)
	for rows.Next() {
		var svc PublicService
		var member ServiceMemberDetails
		var svcDescription, memberDescription sql.NullString
		if err := rows.Scan(
			&svc.ID, &svc.TenantID, &svc.Name, &svcDescription, &svc.DurationMinutes,
			&svc.BufferBefore, &svc.BufferAfter, &svc.Enabled, &svc.CreatedAt, &svc.UpdatedAt,
			&member.ID, &member.MembershipID, &member.PriceCents, &memberDescription,
			&member.MemberName, &member.MemberEmail,
		); err != nil {
			return nil, fmt.Errorf("scanning public service: %w", err)
		}
		member.ServiceID = svc.ID
		member.TenantID = svc.TenantID
		if svcDescription.Valid {
			svc.Description = &svcDescription.String
		}
		if memberDescription.Valid {
			member.Description = &memberDescription.String
		}
		if _, ok := servicesByID[svc.ID]; !ok {
			copy := svc
			copy.Members = []ServiceMemberDetails{}
			servicesByID[svc.ID] = &copy
			order = append(order, svc.ID)
		}
		servicesByID[svc.ID].Members = append(servicesByID[svc.ID].Members, member)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating public services: %w", err)
	}

	out := make([]PublicService, 0, len(order))
	for _, id := range order {
		out = append(out, *servicesByID[id])
	}
	return out, nil
}

// ListByTenant returns services for tenant management.
func (r *Repository) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]Service, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, tenant_id, name, description, duration_minutes, buffer_before_minutes,
		       buffer_after_minutes, enabled, created_at, updated_at
		FROM services
		WHERE tenant_id = $1
		ORDER BY name ASC
	`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("listing tenant services: %w", err)
	}
	defer rows.Close()

	var out []Service
	for rows.Next() {
		var svc Service
		var description sql.NullString
		if err := rows.Scan(&svc.ID, &svc.TenantID, &svc.Name, &description, &svc.DurationMinutes, &svc.BufferBefore, &svc.BufferAfter, &svc.Enabled, &svc.CreatedAt, &svc.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning tenant service: %w", err)
		}
		if description.Valid {
			svc.Description = &description.String
		}
		out = append(out, svc)
	}
	return out, rows.Err()
}

// GetByTenant returns one service in tenant scope.
func (r *Repository) GetByTenant(ctx context.Context, tenantID, serviceID uuid.UUID) (*Service, error) {
	var svc Service
	var description sql.NullString
	err := r.db.QueryRowContext(ctx, `
		SELECT id, tenant_id, name, description, duration_minutes, buffer_before_minutes,
		       buffer_after_minutes, enabled, created_at, updated_at
		FROM services WHERE tenant_id = $1 AND id = $2
	`, tenantID, serviceID).Scan(&svc.ID, &svc.TenantID, &svc.Name, &description, &svc.DurationMinutes, &svc.BufferBefore, &svc.BufferAfter, &svc.Enabled, &svc.CreatedAt, &svc.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("getting service: %w", err)
	}
	if description.Valid {
		svc.Description = &description.String
	}
	return &svc, nil
}

// Create inserts a service for a tenant.
func (r *Repository) Create(ctx context.Context, tenantID uuid.UUID, input CreateServiceInput) (*Service, error) {
	now := time.Now().UTC()
	serviceID := uuid.New()
	var svc Service
	var description sql.NullString
	if input.Description != nil {
		description = sql.NullString{String: *input.Description, Valid: true}
	}
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO services (id, tenant_id, name, description, duration_minutes, buffer_before_minutes, buffer_after_minutes, enabled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, tenant_id, name, description, duration_minutes, buffer_before_minutes, buffer_after_minutes, enabled, created_at, updated_at
	`, serviceID, tenantID, input.Name, description, input.DurationMinutes, input.BufferBefore, input.BufferAfter, input.Enabled, now, now).Scan(
		&svc.ID, &svc.TenantID, &svc.Name, &description, &svc.DurationMinutes, &svc.BufferBefore, &svc.BufferAfter, &svc.Enabled, &svc.CreatedAt, &svc.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("creating service: %w", err)
	}
	if description.Valid {
		svc.Description = &description.String
	}
	return &svc, nil
}

// Update updates a service in tenant scope.
func (r *Repository) Update(ctx context.Context, tenantID, serviceID uuid.UUID, input UpdateServiceInput) (*Service, error) {
	current, err := r.GetByTenant(ctx, tenantID, serviceID)
	if err != nil {
		return nil, err
	}
	if input.Name != nil {
		current.Name = *input.Name
	}
	if input.Description != nil {
		current.Description = *input.Description
	}
	if input.DurationMinutes != nil {
		current.DurationMinutes = *input.DurationMinutes
	}
	if input.BufferBefore != nil {
		current.BufferBefore = *input.BufferBefore
	}
	if input.BufferAfter != nil {
		current.BufferAfter = *input.BufferAfter
	}
	if input.Enabled != nil {
		current.Enabled = *input.Enabled
	}
	current.UpdatedAt = time.Now().UTC()

	var description sql.NullString
	if current.Description != nil {
		description = sql.NullString{String: *current.Description, Valid: true}
	}
	_, err = r.db.ExecContext(ctx, `
		UPDATE services
		SET name = $3, description = $4, duration_minutes = $5, buffer_before_minutes = $6,
		    buffer_after_minutes = $7, enabled = $8, updated_at = $9
		WHERE tenant_id = $1 AND id = $2
	`, tenantID, serviceID, current.Name, description, current.DurationMinutes, current.BufferBefore, current.BufferAfter, current.Enabled, current.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("updating service: %w", err)
	}
	return current, nil
}

// Delete removes a service in tenant scope.
func (r *Repository) Delete(ctx context.Context, tenantID, serviceID uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM services WHERE tenant_id = $1 AND id = $2`, tenantID, serviceID)
	if err != nil {
		return fmt.Errorf("deleting service: %w", err)
	}
	rows, err := result.RowsAffected()
	if err == nil && rows == 0 {
		return ErrNotFound
	}
	return nil
}

// ListMembers returns service-member assignments in tenant scope.
func (r *Repository) ListMembers(ctx context.Context, tenantID, serviceID uuid.UUID) ([]ServiceMemberDetails, error) {
	return r.listMembers(ctx, tenantID, serviceID, nil)
}

// ListMembersByMembership returns service-member assignments for one membership in tenant scope.
func (r *Repository) ListMembersByMembership(ctx context.Context, tenantID, serviceID, membershipID uuid.UUID) ([]ServiceMemberDetails, error) {
	return r.listMembers(ctx, tenantID, serviceID, &membershipID)
}

func (r *Repository) listMembers(ctx context.Context, tenantID, serviceID uuid.UUID, membershipID *uuid.UUID) ([]ServiceMemberDetails, error) {
	query := `
		SELECT sm.id, sm.service_id, sm.membership_id, sm.price_cents, sm.description,
		       u.name, u.email
		FROM service_members sm
		JOIN tenant_memberships tm ON tm.id = sm.membership_id
		JOIN users u ON u.id = tm.user_id
		JOIN services s ON s.id = sm.service_id
		WHERE s.tenant_id = $1 AND sm.service_id = $2`
	args := []any{tenantID, serviceID}
	if membershipID != nil {
		query += ` AND sm.membership_id = $3`
		args = append(args, *membershipID)
	}
	query += ` ORDER BY u.name ASC`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("listing service members: %w", err)
	}
	defer rows.Close()

	var out []ServiceMemberDetails
	for rows.Next() {
		var item ServiceMemberDetails
		var description sql.NullString
		if err := rows.Scan(&item.ID, &item.ServiceID, &item.MembershipID, &item.PriceCents, &description, &item.MemberName, &item.MemberEmail); err != nil {
			return nil, fmt.Errorf("scanning service member: %w", err)
		}
		item.TenantID = tenantID
		if description.Valid {
			item.Description = &description.String
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

// AddMember adds a membership to a service in the same tenant.
func (r *Repository) AddMember(ctx context.Context, tenantID, serviceID uuid.UUID, input AddServiceMemberInput) (*ServiceMember, error) {
	var ok bool
	err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1
			FROM services s
			JOIN tenant_memberships tm ON tm.tenant_id = s.tenant_id
			WHERE s.id = $1 AND s.tenant_id = $2 AND tm.id = $3
		)
	`, serviceID, tenantID, input.MembershipID).Scan(&ok)
	if err != nil {
		return nil, fmt.Errorf("validating service member scope: %w", err)
	}
	if !ok {
		return nil, ErrNotFound
	}

	var item ServiceMember
	var description sql.NullString
	if input.Description != nil {
		description = sql.NullString{String: *input.Description, Valid: true}
	}
	err = r.db.QueryRowContext(ctx, `
		INSERT INTO service_members (id, service_id, membership_id, price_cents, description)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, service_id, membership_id, price_cents, description
	`, uuid.New(), serviceID, input.MembershipID, input.PriceCents, description).Scan(
		&item.ID, &item.ServiceID, &item.MembershipID, &item.PriceCents, &description,
	)
	if err != nil {
		return nil, fmt.Errorf("adding service member: %w", err)
	}
	if description.Valid {
		item.Description = &description.String
	}
	return &item, nil
}

// RemoveMember removes a service-member assignment in tenant scope.
func (r *Repository) RemoveMember(ctx context.Context, tenantID, serviceID, serviceMemberID uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM service_members sm
		USING services s
		WHERE sm.service_id = s.id AND s.tenant_id = $1 AND sm.service_id = $2 AND sm.id = $3
	`, tenantID, serviceID, serviceMemberID)
	if err != nil {
		return fmt.Errorf("removing service member: %w", err)
	}
	rows, err := result.RowsAffected()
	if err == nil && rows == 0 {
		return ErrNotFound
	}
	return nil
}
