package users

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/zenvikar/api/internal/tenants"
)

var (
	// ErrInvalidCredentials indicates email/password mismatch.
	ErrInvalidCredentials = errors.New("invalid credentials")
	// ErrEmailAlreadyExists indicates duplicated user email.
	ErrEmailAlreadyExists = errors.New("email already exists")
	// ErrAccessDenied indicates user does not match required role/context.
	ErrAccessDenied = errors.New("access denied")
	// ErrUnsupportedProvider indicates unsupported social auth provider.
	ErrUnsupportedProvider = errors.New("unsupported provider")
	// ErrSocialValidationFailed indicates provider token rejected/untrusted.
	ErrSocialValidationFailed = errors.New("social identity validation failed")
	// ErrInvalidInput indicates malformed input payload.
	ErrInvalidInput = errors.New("invalid input")
)

// AuthResult contains token + user profile returned by auth endpoints.
type AuthResult struct {
	Token             string            `json:"token"`
	TokenType         string            `json:"tokenType"`
	ExpiresAt         time.Time         `json:"expiresAt"`
	User              *User             `json:"user"`
	PlatformRole      string            `json:"platformRole,omitempty"`
	TenantRoles       map[string]string `json:"tenantRoles,omitempty"`
	CurrentTenantID   string            `json:"currentTenantId,omitempty"`
	CurrentTenantSlug string            `json:"currentTenantSlug,omitempty"`
}

// EmailSignupInput is used for email/password registration.
type EmailSignupInput struct {
	Email    string
	Name     string
	Password string
	Phone    *string
	Locale   string
}

// EmailLoginInput is used for email/password login.
type EmailLoginInput struct {
	Email    string
	Password string
}

// SocialLoginInput is used after OAuth is completed client-side.
type SocialLoginInput struct {
	Provider            string
	GoogleIDToken       string
	GoogleAccessToken   string
	FacebookAccessToken string
}

// Service provides auth use cases for booking/admin/tenant apps.
type Service struct {
	repo    *Repository
	tokens  *TokenManager
	tenants *tenants.Service
	social  *SocialVerifier
}

// NewService creates a user auth service.
func NewService(repo *Repository, tokens *TokenManager, tenantSvc *tenants.Service, social *SocialVerifier) *Service {
	return &Service{
		repo:    repo,
		tokens:  tokens,
		tenants: tenantSvc,
		social:  social,
	}
}

// SignupEmail creates a user and signs them in for booking-web.
func (s *Service) SignupEmail(ctx context.Context, input EmailSignupInput) (*AuthResult, error) {
	email := normalizeEmail(input.Email)
	name := strings.TrimSpace(input.Name)
	password := strings.TrimSpace(input.Password)
	locale := normalizeLocale(input.Locale)

	if email == "" || name == "" || password == "" {
		return nil, ErrInvalidInput
	}
	if len(password) < 8 {
		return nil, fmt.Errorf("%w: password must have at least 8 characters", ErrInvalidInput)
	}

	if _, err := s.repo.FindUserByEmail(ctx, email); err == nil {
		return nil, ErrEmailAlreadyExists
	} else if !errors.Is(err, errNoRows) {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hashing password: %w", err)
	}
	hashStr := string(hash)

	user, err := s.repo.CreateUser(ctx, CreateUserParams{
		Email:            email,
		Name:             name,
		PasswordHash:     &hashStr,
		Phone:            input.Phone,
		PreferredContact: ContactEmail,
		Locale:           locale,
		EmailVerified:    false,
	})
	if err != nil {
		return nil, err
	}

	if err := s.repo.CreateAuthProvider(ctx, user.ID.String(), AuthProviderEmail, email); err != nil {
		return nil, err
	}

	return s.issueAuthResult(ctx, user, "booking-web", nil)
}

// LoginEmail authenticates user credentials.
func (s *Service) LoginEmail(ctx context.Context, input EmailLoginInput, audience string) (*AuthResult, error) {
	email := normalizeEmail(input.Email)
	if email == "" || strings.TrimSpace(input.Password) == "" {
		return nil, ErrInvalidInput
	}

	user, err := s.repo.FindUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, errNoRows) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}
	if user.PasswordHash == nil || *user.PasswordHash == "" {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(input.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	return s.issueAuthResult(ctx, user, audience, nil)
}

// LoginBooking authenticates a booking-web user and scopes the session to a tenant.
func (s *Service) LoginBooking(ctx context.Context, input EmailLoginInput, tenantSlug string) (*AuthResult, error) {
	tenant, err := s.resolveTenantContext(ctx, tenantSlug)
	if err != nil {
		return nil, ErrAccessDenied
	}
	result, err := s.LoginEmail(ctx, input, "booking-web")
	if err != nil {
		return nil, err
	}
	return s.issueAuthResult(ctx, result.User, "booking-web", tenant)
}

// LoginSocial signs in or creates user with external provider identity.
// The frontend is expected to finish provider OAuth and send provider identity.
func (s *Service) LoginSocial(ctx context.Context, input SocialLoginInput, audience string) (*AuthResult, error) {
	identity, err := s.social.Verify(ctx, input)
	if err != nil {
		if errors.Is(err, ErrUnsupportedProvider) || errors.Is(err, ErrInvalidInput) {
			return nil, err
		}
		return nil, ErrSocialValidationFailed
	}

	user, err := s.repo.FindUserByProvider(ctx, identity.Provider, identity.ProviderUserID)
	if err == nil {
		return s.issueAuthResult(ctx, user, audience, nil)
	}
	if err != nil && !errors.Is(err, errNoRows) {
		return nil, err
	}

	user, err = s.repo.FindUserByEmail(ctx, identity.Email)
	if err != nil && !errors.Is(err, errNoRows) {
		return nil, err
	}

	if errors.Is(err, errNoRows) {
		user, err = s.repo.CreateUser(ctx, CreateUserParams{
			Email:            identity.Email,
			Name:             fallbackName(identity.Name, identity.Email),
			PasswordHash:     nil,
			Phone:            nil,
			PreferredContact: ContactEmail,
			Locale:           "en",
			EmailVerified:    identity.EmailVerified,
		})
		if err != nil {
			return nil, err
		}
	}

	if err := s.repo.CreateAuthProvider(ctx, user.ID.String(), identity.Provider, identity.ProviderUserID); err != nil {
		return nil, err
	}

	return s.issueAuthResult(ctx, user, audience, nil)
}

// LoginSocialBooking signs in a booking user and scopes the session to a tenant.
func (s *Service) LoginSocialBooking(ctx context.Context, input SocialLoginInput, tenantSlug string) (*AuthResult, error) {
	tenant, err := s.resolveTenantContext(ctx, tenantSlug)
	if err != nil {
		return nil, ErrAccessDenied
	}
	result, err := s.LoginSocial(ctx, input, "booking-web")
	if err != nil {
		return nil, err
	}
	return s.issueAuthResult(ctx, result.User, "booking-web", tenant)
}

// LoginSocialTenant signs in a tenant-web user, enforcing tenant membership and session scope.
func (s *Service) LoginSocialTenant(ctx context.Context, input SocialLoginInput, tenantSlug string) (*AuthResult, error) {
	tenant, err := s.resolveTenantContext(ctx, tenantSlug)
	if err != nil {
		return nil, ErrAccessDenied
	}
	result, err := s.LoginSocial(ctx, input, "tenant-web")
	if err != nil {
		return nil, err
	}
	role, err := s.repo.FindTenantRoleForUser(ctx, result.User.ID.String(), tenant.ID.String())
	if err != nil {
		return nil, err
	}
	if role == "" {
		return nil, ErrAccessDenied
	}
	result.TenantRoles = map[string]string{tenant.ID.String(): role}
	token, exp, err := s.tokens.IssueToken(result.User, "tenant-web", result.PlatformRole, result.TenantRoles, tenant.ID.String(), tenant.Slug)
	if err != nil {
		return nil, err
	}
	result.Token = token
	result.ExpiresAt = exp
	result.CurrentTenantID = tenant.ID.String()
	result.CurrentTenantSlug = tenant.Slug
	return result, nil
}

// LoginAdmin authenticates and enforces platform_admin membership.
func (s *Service) LoginAdmin(ctx context.Context, input EmailLoginInput) (*AuthResult, error) {
	result, err := s.LoginEmail(ctx, input, "admin-web")
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(result.PlatformRole) == "" {
		return nil, ErrAccessDenied
	}
	return result, nil
}

// LoginTenant authenticates and enforces membership in resolved tenant slug.
func (s *Service) LoginTenant(ctx context.Context, input EmailLoginInput, tenantSlug string) (*AuthResult, error) {
	slug := strings.ToLower(strings.TrimSpace(tenantSlug))
	if slug == "" {
		return nil, ErrInvalidInput
	}

	tenant, err := s.tenants.ResolveTenantBySlug(ctx, slug)
	if err != nil {
		return nil, ErrAccessDenied
	}

	result, err := s.LoginEmail(ctx, input, "tenant-web")
	if err != nil {
		return nil, err
	}

	role, err := s.repo.FindTenantRoleForUser(ctx, result.User.ID.String(), tenant.ID.String())
	if err != nil {
		return nil, err
	}
	if role == "" {
		return nil, ErrAccessDenied
	}

	// Limit tenant-web token to current tenant context.
	result.TenantRoles = map[string]string{
		tenant.ID.String(): role,
	}

	token, exp, err := s.tokens.IssueToken(result.User, "tenant-web", result.PlatformRole, result.TenantRoles, tenant.ID.String(), tenant.Slug)
	if err != nil {
		return nil, err
	}
	result.Token = token
	result.ExpiresAt = exp
	result.CurrentTenantID = tenant.ID.String()
	result.CurrentTenantSlug = tenant.Slug

	return result, nil
}

// LoginTenantPortal authenticates a user for tenant-web and ensures they belong
// to at least one tenant membership.
func (s *Service) LoginTenantPortal(ctx context.Context, input EmailLoginInput) (*AuthResult, error) {
	result, err := s.LoginEmail(ctx, input, "tenant-web")
	if err != nil {
		return nil, err
	}
	if len(result.TenantRoles) == 0 {
		return nil, ErrAccessDenied
	}
	return result, nil
}

// ParseToken validates a bearer token and returns decoded claims.
func (s *Service) ParseToken(token string) (*tokenClaims, error) {
	return s.tokens.ParseToken(token)
}

// ListUserTenantAccess returns tenants a user can access.
func (s *Service) ListUserTenantAccess(ctx context.Context, userID string) ([]TenantAccess, error) {
	return s.repo.ListUserTenantAccess(ctx, userID)
}

func (s *Service) issueAuthResult(ctx context.Context, user *User, audience string, currentTenant *tenants.Tenant) (*AuthResult, error) {
	platformRole, err := s.repo.FindPlatformRole(ctx, user.ID.String())
	if err != nil {
		return nil, err
	}

	tenantRoles, err := s.repo.ListTenantRoles(ctx, user.ID.String())
	if err != nil {
		return nil, err
	}

	currentTenantID := ""
	currentTenantSlug := ""
	if currentTenant != nil {
		currentTenantID = currentTenant.ID.String()
		currentTenantSlug = currentTenant.Slug
	}
	token, exp, err := s.tokens.IssueToken(user, audience, platformRole, tenantRoles, currentTenantID, currentTenantSlug)
	if err != nil {
		return nil, err
	}

	return &AuthResult{
		Token:             token,
		TokenType:         "Bearer",
		ExpiresAt:         exp,
		User:              user,
		PlatformRole:      platformRole,
		TenantRoles:       tenantRoles,
		CurrentTenantID:   currentTenantID,
		CurrentTenantSlug: currentTenantSlug,
	}, nil
}

func (s *Service) resolveTenantContext(ctx context.Context, tenantSlug string) (*tenants.Tenant, error) {
	slug := strings.ToLower(strings.TrimSpace(tenantSlug))
	if slug == "" {
		return nil, ErrInvalidInput
	}
	return s.tenants.ResolveTenantBySlug(ctx, slug)
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func normalizeLocale(locale string) string {
	normalized := strings.ToLower(strings.TrimSpace(locale))
	if normalized != "pt" {
		return "en"
	}
	return normalized
}

func fallbackName(name, email string) string {
	trimmed := strings.TrimSpace(name)
	if trimmed != "" {
		return trimmed
	}
	localPart := strings.Split(strings.TrimSpace(email), "@")[0]
	if localPart == "" {
		return "User"
	}
	return localPart
}
