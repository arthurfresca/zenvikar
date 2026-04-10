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
