package endpointutil

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/zenvikar/api/internal/platform/authn"
	"github.com/zenvikar/api/internal/platform/authz"
	"github.com/zenvikar/api/internal/platform/httpapi"
)

// ParseUUIDParam parses a UUID route parameter and writes a 400 response on failure.
func ParseUUIDParam(w http.ResponseWriter, r *http.Request, name string) (uuid.UUID, bool) {
	value := chi.URLParam(r, name)
	id, err := uuid.Parse(value)
	if err != nil {
		httpapi.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error":   "invalid_request",
			"message": "invalid " + name,
		})
		return uuid.UUID{}, false
	}
	return id, true
}

// CurrentUserID returns the authenticated user ID or writes a 401 response.
func CurrentUserID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	userID, ok := authn.UserIDFromContext(r.Context())
	if !ok {
		httpapi.WriteJSON(w, http.StatusUnauthorized, map[string]string{
			"error":   "unauthorized",
			"message": "invalid or expired token",
		})
		return uuid.UUID{}, false
	}
	return userID, true
}

// CurrentClaims returns parsed auth claims or writes a 401 response.
func CurrentClaims(w http.ResponseWriter, r *http.Request) (*authn.Claims, bool) {
	claims := authn.ClaimsFromContext(r.Context())
	if claims == nil {
		httpapi.WriteJSON(w, http.StatusUnauthorized, map[string]string{
			"error":   "unauthorized",
			"message": "invalid or expired token",
		})
		return nil, false
	}
	return claims, true
}

// RequireCurrentTenantID checks that the current token is scoped to the requested tenant.
func RequireCurrentTenantID(w http.ResponseWriter, r *http.Request, tenantID uuid.UUID) bool {
	claims, ok := CurrentClaims(w, r)
	if !ok {
		return false
	}
	if strings.TrimSpace(claims.CurrentTenantID) == "" {
		return true
	}
	if claims.CurrentTenantID != tenantID.String() {
		httpapi.WriteJSON(w, http.StatusForbidden, map[string]string{
			"error":   "forbidden",
			"message": "session is scoped to a different tenant",
		})
		return false
	}
	return true
}

// RequireCurrentTenantSlug checks that the current token is scoped to the requested tenant slug.
func RequireCurrentTenantSlug(w http.ResponseWriter, r *http.Request, tenantSlug string) bool {
	claims, ok := CurrentClaims(w, r)
	if !ok {
		return false
	}
	if strings.TrimSpace(claims.CurrentTenantSlug) == "" {
		return true
	}
	if !strings.EqualFold(claims.CurrentTenantSlug, strings.TrimSpace(tenantSlug)) {
		httpapi.WriteJSON(w, http.StatusForbidden, map[string]string{
			"error":   "forbidden",
			"message": "session is scoped to a different tenant",
		})
		return false
	}
	return true
}

// RequireTenantPermission checks the tenant permission for the current user.
func RequireTenantPermission(w http.ResponseWriter, r *http.Request, authzSvc *authz.Service, tenantID uuid.UUID, permission string) bool {
	if !RequireCurrentTenantID(w, r, tenantID) {
		return false
	}
	userID, ok := CurrentUserID(w, r)
	if !ok {
		return false
	}
	if err := authzSvc.Authorize(r.Context(), userID, tenantID, permission); err != nil {
		httpapi.WriteJSON(w, http.StatusForbidden, map[string]string{
			"error":   "forbidden",
			"message": err.Error(),
		})
		return false
	}
	return true
}
