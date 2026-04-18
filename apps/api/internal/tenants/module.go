package tenants

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/zenvikar/api/internal/platform"
	"github.com/zenvikar/api/internal/platform/authn"
	"github.com/zenvikar/api/internal/platform/authz"
	"github.com/zenvikar/api/internal/platform/httpapi"
	"github.com/zenvikar/api/internal/tenant_memberships"
)

// TenantsModule implements the platform.Module interface for the tenants domain.
type TenantsModule struct{}

// New creates a new TenantsModule.
func New() *TenantsModule {
	return &TenantsModule{}
}

// Name returns the module name.
func (m *TenantsModule) Name() string {
	return "tenants"
}

// RegisterRoutes registers tenant-related HTTP routes.
func (m *TenantsModule) RegisterRoutes(router chi.Router, deps platform.Dependencies) {
	repo := NewRepository(deps.DB)
	svc := NewService(repo, deps.Redis)
	membershipSvc := tenant_memberships.NewService(tenant_memberships.NewRepository(deps.DB))
	authzSvc := authz.NewService(authz.NewPlatformAdminChecker(deps.DB), membershipSvc)
	h := newHandler(svc, authzSvc)

	router.Get("/api/v1/tenants/resolve", func(w http.ResponseWriter, r *http.Request) {
		slug := r.URL.Query().Get("slug")
		if slug == "" {
			httpapi.WriteJSON(w, http.StatusBadRequest, map[string]string{
				"error":   "missing_slug",
				"message": "slug query parameter is required",
			})
			return
		}

		tenant, err := svc.ResolveTenantBySlug(r.Context(), slug)
		if err != nil {
			switch err {
			case ErrTenantNotFound:
				httpapi.WriteJSON(w, http.StatusNotFound, map[string]string{
					"error":   "tenant_not_found",
					"message": "No tenant found for this address",
				})
			case ErrTenantDisabled:
				httpapi.WriteJSON(w, http.StatusForbidden, map[string]string{
					"error":   "tenant_disabled",
					"message": "This tenant is currently disabled",
				})
			default:
				httpapi.WriteJSON(w, http.StatusBadRequest, map[string]string{
					"error":   "invalid_request",
					"message": err.Error(),
				})
			}
			return
		}

		httpapi.WriteJSON(w, http.StatusOK, tenant)
	})

	h.register(router, authn.RequireAuth(deps.Config.JWTSecret, deps.Config.JWTTTLMinutes))
}

// Migrate is a no-op — migrations are handled centrally by the migrations package.
func (m *TenantsModule) Migrate(db *sql.DB) error {
	return nil
}
