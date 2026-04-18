package tenant_memberships

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

// ErrNotAMember is returned when a user has no membership in the specified tenant.
var ErrNotAMember = errors.New("user is not a member of this tenant")

// Service provides tenant membership operations.
type Service struct {
	repo *Repository
}

// NewService creates a new membership service.
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// CheckMembership verifies that a user belongs to a tenant and returns their membership.
// Returns ErrNotAMember if no membership exists.
func (s *Service) CheckMembership(ctx context.Context, userID, tenantID uuid.UUID) (*TenantMembership, error) {
	membership, err := s.repo.FindByUserAndTenant(ctx, userID.String(), tenantID.String())
	if err != nil {
		return nil, ErrNotAMember
	}
	return membership, nil
}

// ListByTenant returns all memberships for a tenant with user details.
func (s *Service) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]MembershipDetails, error) {
	return s.repo.ListByTenant(ctx, tenantID)
}

// GetByTenant returns a membership in tenant scope.
func (s *Service) GetByTenant(ctx context.Context, tenantID, membershipID uuid.UUID) (*MembershipDetails, error) {
	return s.repo.GetByTenant(ctx, tenantID, membershipID)
}

// Create creates a membership in tenant scope.
func (s *Service) Create(ctx context.Context, tenantID uuid.UUID, input CreateMembershipInput) (*TenantMembership, error) {
	return s.repo.Create(ctx, tenantID, input)
}

// Update updates a membership in tenant scope.
func (s *Service) Update(ctx context.Context, tenantID, membershipID uuid.UUID, input UpdateMembershipInput) (*TenantMembership, error) {
	return s.repo.Update(ctx, tenantID, membershipID, input)
}

// Delete deletes a membership in tenant scope.
func (s *Service) Delete(ctx context.Context, tenantID, membershipID uuid.UUID) error {
	return s.repo.Delete(ctx, tenantID, membershipID)
}

// ValidRole reports whether the given role is supported.
func ValidRole(role string) bool {
	switch role {
	case RoleTenantOwner, RoleTenantManager, RoleTenantStaff, RoleTenantFinanceViewer:
		return true
	default:
		return false
	}
}
