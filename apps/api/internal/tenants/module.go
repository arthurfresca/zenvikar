package tenants

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/zenvikar/api/internal/platform"
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
	svc := NewService(NewRepository(deps.DB), deps.Redis)

	router.Get("/api/v1/tenants/resolve", func(w http.ResponseWriter, r *http.Request) {
		slug := r.URL.Query().Get("slug")
		if slug == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error":   "missing_slug",
				"message": "slug query parameter is required",
			})
			return
		}

		tenant, err := svc.ResolveTenantBySlug(r.Context(), slug)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			switch err {
			case ErrTenantNotFound:
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]string{
					"error":   "tenant_not_found",
					"message": "No tenant found for this address",
				})
			case ErrTenantDisabled:
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(map[string]string{
					"error":   "tenant_disabled",
					"message": "This tenant is currently disabled",
				})
			default:
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{
					"error":   "invalid_request",
					"message": err.Error(),
				})
			}
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tenant)
	})
}

// Migrate is a no-op — migrations are handled centrally by the migrations package.
func (m *TenantsModule) Migrate(db *sql.DB) error {
	return nil
}
