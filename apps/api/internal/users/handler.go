package users

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/zenvikar/api/internal/platform/config"
	appmiddleware "github.com/zenvikar/api/internal/platform/middleware"
	"github.com/zenvikar/api/internal/tenants"
)

type authHandler struct {
	svc *Service
	cfg *config.Config
}

func newAuthHandler(svc *Service, cfg *config.Config) *authHandler {
	return &authHandler{
		svc: svc,
		cfg: cfg,
	}
}

type signupRequest struct {
	Email    string  `json:"email"`
	Name     string  `json:"name"`
	Password string  `json:"password"`
	Phone    *string `json:"phone"`
	Locale   string  `json:"locale"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type socialLoginRequest struct {
	Provider            string `json:"provider"`
	GoogleIDToken       string `json:"googleIdToken"`
	GoogleAccessToken   string `json:"googleAccessToken"`
	FacebookAccessToken string `json:"facebookAccessToken"`
}

func (h *authHandler) signup(w http.ResponseWriter, r *http.Request) {
	var req signupRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	result, err := h.svc.SignupEmail(r.Context(), EmailSignupInput{
		Email:    req.Email,
		Name:     req.Name,
		Password: req.Password,
		Phone:    req.Phone,
		Locale:   req.Locale,
	})
	if err != nil {
		writeAuthError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, result)
}

func (h *authHandler) login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	result, err := h.svc.LoginEmail(r.Context(), EmailLoginInput{
		Email:    req.Email,
		Password: req.Password,
	}, "booking-web")
	if err != nil {
		writeAuthError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (h *authHandler) socialLogin(w http.ResponseWriter, r *http.Request) {
	var req socialLoginRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	result, err := h.svc.LoginSocial(r.Context(), SocialLoginInput{
		Provider:            req.Provider,
		GoogleIDToken:       req.GoogleIDToken,
		GoogleAccessToken:   req.GoogleAccessToken,
		FacebookAccessToken: req.FacebookAccessToken,
	}, "booking-web")
	if err != nil {
		writeAuthError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (h *authHandler) adminLogin(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	result, err := h.svc.LoginAdmin(r.Context(), EmailLoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		writeAuthError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (h *authHandler) tenantLogin(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	slug := resolveTenantSlug(r, h.cfg.BaseDomain)
	var (
		result *AuthResult
		err    error
	)
	if slug == "" {
		result, err = h.svc.LoginTenantPortal(r.Context(), EmailLoginInput{
			Email:    req.Email,
			Password: req.Password,
		})
	} else {
		result, err = h.svc.LoginTenant(r.Context(), EmailLoginInput{
			Email:    req.Email,
			Password: req.Password,
		}, slug)
	}
	if err != nil {
		writeAuthError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (h *authHandler) me(w http.ResponseWriter, r *http.Request) {
	token := bearerToken(r.Header.Get("Authorization"))
	if token == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{
			"error":   "unauthorized",
			"message": "missing bearer token",
		})
		return
	}

	claims, err := h.svc.ParseToken(token)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{
			"error":   "unauthorized",
			"message": "invalid or expired token",
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"userId":       claims.Subject,
		"email":        claims.Email,
		"name":         claims.Name,
		"audience":     claims.Audience,
		"platformRole": claims.PlatformRole,
		"tenantRoles":  claims.TenantRoles,
		"issuedAt":     claims.IssuedAt,
		"expiresAt":    claims.ExpiresAt,
	})
}

func (h *authHandler) tenants(w http.ResponseWriter, r *http.Request) {
	token := bearerToken(r.Header.Get("Authorization"))
	if token == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{
			"error":   "unauthorized",
			"message": "missing bearer token",
		})
		return
	}

	claims, err := h.svc.ParseToken(token)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{
			"error":   "unauthorized",
			"message": "invalid or expired token",
		})
		return
	}

	tenants, err := h.svc.ListUserTenantAccess(r.Context(), claims.Subject)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error":   "internal_error",
			"message": "failed to load tenant access",
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"tenants": tenants,
	})
}

func resolveTenantSlug(r *http.Request, baseDomain string) string {
	if slug := strings.ToLower(strings.TrimSpace(r.Header.Get("X-Tenant-ID"))); slug != "" {
		return slug
	}

	if slug, err := appmiddleware.ExtractTenantSlugFromHost(r.Host, baseDomain); err == nil {
		normalized := strings.ToLower(strings.TrimSpace(slug))
		// Ignore platform/app subdomains (api/manage/admin/www/etc) as tenant context.
		if _, reserved := tenants.ReservedSlugs[normalized]; reserved {
			return ""
		}
		return normalized
	}

	return ""
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error":   "invalid_request",
			"message": "invalid JSON payload",
		})
		return false
	}
	return true
}

func writeAuthError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrInvalidCredentials):
		writeJSON(w, http.StatusUnauthorized, map[string]string{
			"error":   "invalid_credentials",
			"message": "invalid email or password",
		})
	case errors.Is(err, ErrEmailAlreadyExists):
		writeJSON(w, http.StatusConflict, map[string]string{
			"error":   "email_already_exists",
			"message": "an account with this email already exists",
		})
	case errors.Is(err, ErrAccessDenied):
		writeJSON(w, http.StatusForbidden, map[string]string{
			"error":   "forbidden",
			"message": "you do not have access to this area",
		})
	case errors.Is(err, ErrUnsupportedProvider):
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error":   "unsupported_provider",
			"message": "provider must be google or facebook",
		})
	case errors.Is(err, ErrSocialValidationFailed):
		writeJSON(w, http.StatusUnauthorized, map[string]string{
			"error":   "social_validation_failed",
			"message": "could not validate social account with provider",
		})
	case errors.Is(err, ErrInvalidInput):
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error":   "invalid_request",
			"message": err.Error(),
		})
	default:
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error":   "internal_error",
			"message": "unexpected error",
		})
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func bearerToken(authHeader string) string {
	parts := strings.SplitN(strings.TrimSpace(authHeader), " ", 2)
	if len(parts) != 2 {
		return ""
	}
	if !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}
