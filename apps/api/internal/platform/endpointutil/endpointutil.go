package endpointutil

import (
	"net/http"

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

// RequireTenantPermission checks the tenant permission for the current user.
func RequireTenantPermission(w http.ResponseWriter, r *http.Request, authzSvc *authz.Service, tenantID uuid.UUID, permission string) bool {
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
