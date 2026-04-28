package tenants

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/zenvikar/api/internal/platform/authz"
	"github.com/zenvikar/api/internal/platform/endpointutil"
	"github.com/zenvikar/api/internal/platform/httpapi"
)

type handler struct {
	svc      *Service
	authzSvc *authz.Service
}

type updateTenantRequest struct {
	DisplayName    *string `json:"displayName"`
	LogoURL        *string `json:"logoUrl"`
	ColorPrimary   *string `json:"colorPrimary"`
	ColorSecondary *string `json:"colorSecondary"`
	ColorAccent    *string `json:"colorAccent"`
	Phone          *string `json:"phone"`
	Email          *string `json:"email"`
	Address        *string `json:"address"`
	Currency       *string `json:"currency"`
	Timezone       *string `json:"timezone"`
	DefaultLocale  *string `json:"defaultLocale"`
	Enabled        *bool   `json:"enabled"`
}

func newHandler(svc *Service, authzSvc *authz.Service) *handler {
	return &handler{svc: svc, authzSvc: authzSvc}
}

func (h *handler) register(router chi.Router, requireAuth func(http.Handler) http.Handler) {
	router.With(requireAuth).Get("/api/v1/tenant/tenants/{tenantId}", h.get)
	router.With(requireAuth).Patch("/api/v1/tenant/tenants/{tenantId}", h.update)
}

func (h *handler) get(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := endpointutil.ParseUUIDParam(w, r, "tenantId")
	if !ok || !endpointutil.RequireTenantPermission(w, r, h.authzSvc, tenantID, "branding:read") {
		return
	}
	tenant, err := h.svc.GetByID(r.Context(), tenantID)
	if err != nil {
		status := http.StatusInternalServerError
		message := "failed to load tenant"
		if err == ErrNotFound {
			status = http.StatusNotFound
			message = "tenant not found"
		}
		httpapi.WriteJSON(w, status, map[string]string{"error": "not_found", "message": message})
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, tenant)
}

func (h *handler) update(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := endpointutil.ParseUUIDParam(w, r, "tenantId")
	if !ok || !endpointutil.RequireTenantPermission(w, r, h.authzSvc, tenantID, "services:update") {
		return
	}
	var req updateTenantRequest
	if !httpapi.DecodeJSON(w, r, &req) {
		return
	}
	var logoURL, phone, email, address **string
	if req.LogoURL != nil {
		logoURL = &req.LogoURL
	}
	if req.Phone != nil {
		phone = &req.Phone
	}
	if req.Email != nil {
		email = &req.Email
	}
	if req.Address != nil {
		address = &req.Address
	}
	tenant, err := h.svc.Update(r.Context(), tenantID, UpdateTenantInput{
		DisplayName:    req.DisplayName,
		LogoURL:        logoURL,
		ColorPrimary:   req.ColorPrimary,
		ColorSecondary: req.ColorSecondary,
		ColorAccent:    req.ColorAccent,
		Phone:          phone,
		Email:          email,
		Address:        address,
		Currency:       req.Currency,
		Timezone:       req.Timezone,
		DefaultLocale:  req.DefaultLocale,
		Enabled:        req.Enabled,
	})
	if err != nil {
		status := http.StatusInternalServerError
		message := "failed to update tenant"
		if err == ErrNotFound {
			status = http.StatusNotFound
			message = "tenant not found"
		}
		httpapi.WriteJSON(w, status, map[string]string{"error": "not_found", "message": message})
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, tenant)
}
