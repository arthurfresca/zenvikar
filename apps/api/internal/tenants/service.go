package tenants

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// Sentinel errors for tenant resolution.
var (
	ErrTenantNotFound = errors.New("tenant not found")
	ErrTenantDisabled = errors.New("tenant is disabled")
)

const tenantCacheTTL = 5 * time.Minute

// Service provides tenant resolution with Redis caching.
type Service struct {
	repo  *Repository
	redis *redis.Client
}

// NewService creates a new tenant service.
func NewService(repo *Repository, rdb *redis.Client) *Service {
	return &Service{
		repo:  repo,
		redis: rdb,
	}
}

// ResolveTenantBySlug resolves a tenant by slug with Redis caching.
// Algorithm: validate slug → check Redis cache → query DB → check enabled → cache in Redis → return.
func (s *Service) ResolveTenantBySlug(ctx context.Context, slug string) (*Tenant, error) {
	// Step 1: Validate slug format.
	if err := ValidateTenantSlug(slug); err != nil {
		return nil, fmt.Errorf("invalid slug: %w", err)
	}

	// Step 2: Check Redis cache.
	cacheKey := fmt.Sprintf("tenant:slug:%s", slug)
	cached, err := s.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var tenant Tenant
		if err := json.Unmarshal([]byte(cached), &tenant); err == nil {
			return &tenant, nil
		}
	}

	// Step 3: Query database.
	tenant, err := s.repo.FindBySlug(ctx, slug)
	if err != nil {
		return nil, ErrTenantNotFound
	}

	// Step 4: Check enabled status.
	if !tenant.Enabled {
		return nil, ErrTenantDisabled
	}

	// Step 5: Cache result in Redis with 5min TTL.
	data, err := json.Marshal(tenant)
	if err == nil {
		_ = s.redis.Set(ctx, cacheKey, data, tenantCacheTTL).Err()
	}

	return tenant, nil
}

// GetByID returns a tenant by ID.
func (s *Service) GetByID(ctx context.Context, tenantID uuid.UUID) (*Tenant, error) {
	return s.repo.FindByID(ctx, tenantID)
}

// Update updates tenant settings.
func (s *Service) Update(ctx context.Context, tenantID uuid.UUID, input UpdateTenantInput) (*Tenant, error) {
	return s.repo.Update(ctx, tenantID, input)
}
