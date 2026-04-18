package services

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/zenvikar/api/internal/platform/authz"
	"github.com/zenvikar/api/internal/platform/endpointutil"
	"github.com/zenvikar/api/internal/platform/httpapi"
	"github.com/zenvikar/api/internal/tenants"
)

type handler struct {
	repo      *Repository
	tenantSvc *tenants.Service
	authzSvc  *authz.Service
}

func newHandler(repo *Repository, tenantSvc *tenants.Service, authzSvc *authz.Service) *handler {
	return &handler{repo: repo, tenantSvc: tenantSvc, authzSvc: authzSvc}
}

func (h *handler) register(router chi.Router, requireAuth func(http.Handler) http.Handler) {
	router.Get("/api/v1/tenants/{tenantSlug}/services", h.listPublic)

	router.Route("/api/v1/tenant/tenants/{tenantId}/services", func(r chi.Router) {
		r.Use(requireAuth)
		r.Get("/", h.listTenant)
		r.Post("/", h.create)
		r.Get("/{serviceId}", h.get)
		r.Patch("/{serviceId}", h.update)
		r.Delete("/{serviceId}", h.remove)
		r.Get("/{serviceId}/members", h.listMembers)
		r.Post("/{serviceId}/members", h.addMember)
		r.Delete("/{serviceId}/members/{serviceMemberId}", h.removeMember)
	})
}

func (h *handler) listPublic(w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimSpace(chi.URLParam(r, "tenantSlug"))
	if slug == "" {
		httpapi.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_request", "message": "missing tenantSlug"})
		return
	}
	if _, err := h.tenantSvc.ResolveTenantBySlug(r.Context(), slug); err != nil {
		httpapi.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "tenant_not_found", "message": "No tenant found for this address"})
		return
	}
	items, err := h.repo.ListPublicByTenantSlug(r.Context(), slug)
	if err != nil {
		httpapi.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal_error", "message": "failed to load services"})
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, map[string]any{"services": items})
}

func (h *handler) listTenant(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := endpointutil.ParseUUIDParam(w, r, "tenantId")
	if !ok || !endpointutil.RequireTenantPermission(w, r, h.authzSvc, tenantID, "services:read") {
		return
	}
	items, err := h.repo.ListByTenant(r.Context(), tenantID)
	if err != nil {
		httpapi.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal_error", "message": "failed to load services"})
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, map[string]any{"services": items})
}

type serviceRequest struct {
	Name            string  `json:"name"`
	Description     *string `json:"description"`
	DurationMinutes int     `json:"durationMinutes"`
	BufferBefore    int     `json:"bufferBefore"`
	BufferAfter     int     `json:"bufferAfter"`
	Enabled         *bool   `json:"enabled"`
}

func (h *handler) create(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := endpointutil.ParseUUIDParam(w, r, "tenantId")
	if !ok || !endpointutil.RequireTenantPermission(w, r, h.authzSvc, tenantID, "services:create") {
		return
	}
	var req serviceRequest
	if !httpapi.DecodeJSON(w, r, &req) {
		return
	}
	if strings.TrimSpace(req.Name) == "" || req.DurationMinutes <= 0 || req.BufferBefore < 0 || req.BufferAfter < 0 {
		httpapi.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_request", "message": "invalid service payload"})
		return
	}
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	svc, err := h.repo.Create(r.Context(), tenantID, CreateServiceInput{
		Name:            strings.TrimSpace(req.Name),
		Description:     req.Description,
		DurationMinutes: req.DurationMinutes,
		BufferBefore:    req.BufferBefore,
		BufferAfter:     req.BufferAfter,
		Enabled:         enabled,
	})
	if err != nil {
		httpapi.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal_error", "message": "failed to create service"})
		return
	}
	httpapi.WriteJSON(w, http.StatusCreated, svc)
}

func (h *handler) get(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := endpointutil.ParseUUIDParam(w, r, "tenantId")
	if !ok || !endpointutil.RequireTenantPermission(w, r, h.authzSvc, tenantID, "services:read") {
		return
	}
	serviceID, ok := endpointutil.ParseUUIDParam(w, r, "serviceId")
	if !ok {
		return
	}
	svc, err := h.repo.GetByTenant(r.Context(), tenantID, serviceID)
	if err != nil {
		status := http.StatusInternalServerError
		message := "failed to load service"
		if err == ErrNotFound {
			status = http.StatusNotFound
			message = "service not found"
		}
		httpapi.WriteJSON(w, status, map[string]string{"error": "not_found", "message": message})
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, svc)
}

func (h *handler) update(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := endpointutil.ParseUUIDParam(w, r, "tenantId")
	if !ok || !endpointutil.RequireTenantPermission(w, r, h.authzSvc, tenantID, "services:update") {
		return
	}
	serviceID, ok := endpointutil.ParseUUIDParam(w, r, "serviceId")
	if !ok {
		return
	}
	var req struct {
		Name            *string `json:"name"`
		Description     *string `json:"description"`
		DurationMinutes *int    `json:"durationMinutes"`
		BufferBefore    *int    `json:"bufferBefore"`
		BufferAfter     *int    `json:"bufferAfter"`
		Enabled         *bool   `json:"enabled"`
	}
	if !httpapi.DecodeJSON(w, r, &req) {
		return
	}
	var description **string
	if req.Description != nil {
		description = &req.Description
	}
	svc, err := h.repo.Update(r.Context(), tenantID, serviceID, UpdateServiceInput{
		Name:            req.Name,
		Description:     description,
		DurationMinutes: req.DurationMinutes,
		BufferBefore:    req.BufferBefore,
		BufferAfter:     req.BufferAfter,
		Enabled:         req.Enabled,
	})
	if err != nil {
		status := http.StatusInternalServerError
		message := "failed to update service"
		if err == ErrNotFound {
			status = http.StatusNotFound
			message = "service not found"
		}
		httpapi.WriteJSON(w, status, map[string]string{"error": "not_found", "message": message})
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, svc)
}

func (h *handler) remove(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := endpointutil.ParseUUIDParam(w, r, "tenantId")
	if !ok || !endpointutil.RequireTenantPermission(w, r, h.authzSvc, tenantID, "services:delete") {
		return
	}
	serviceID, ok := endpointutil.ParseUUIDParam(w, r, "serviceId")
	if !ok {
		return
	}
	if err := h.repo.Delete(r.Context(), tenantID, serviceID); err != nil {
		status := http.StatusInternalServerError
		message := "failed to delete service"
		if err == ErrNotFound {
			status = http.StatusNotFound
			message = "service not found"
		}
		httpapi.WriteJSON(w, status, map[string]string{"error": "not_found", "message": message})
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (h *handler) listMembers(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := endpointutil.ParseUUIDParam(w, r, "tenantId")
	if !ok || !endpointutil.RequireTenantPermission(w, r, h.authzSvc, tenantID, "services:read") {
		return
	}
	serviceID, ok := endpointutil.ParseUUIDParam(w, r, "serviceId")
	if !ok {
		return
	}
	items, err := h.repo.ListMembers(r.Context(), tenantID, serviceID)
	if err != nil {
		httpapi.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal_error", "message": "failed to load service members"})
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, map[string]any{"members": items})
}

func (h *handler) addMember(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := endpointutil.ParseUUIDParam(w, r, "tenantId")
	if !ok || !endpointutil.RequireTenantPermission(w, r, h.authzSvc, tenantID, "services:update") {
		return
	}
	serviceID, ok := endpointutil.ParseUUIDParam(w, r, "serviceId")
	if !ok {
		return
	}
	var req struct {
		MembershipID string  `json:"membershipId"`
		PriceCents   int     `json:"priceCents"`
		Description  *string `json:"description"`
	}
	if !httpapi.DecodeJSON(w, r, &req) {
		return
	}
	membershipID, err := uuid.Parse(req.MembershipID)
	if err != nil || req.PriceCents < 0 {
		httpapi.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_request", "message": "invalid service member payload"})
		return
	}
	item, err := h.repo.AddMember(r.Context(), tenantID, serviceID, AddServiceMemberInput{MembershipID: membershipID, PriceCents: req.PriceCents, Description: req.Description})
	if err != nil {
		status := http.StatusInternalServerError
		message := "failed to add service member"
		if err == ErrNotFound {
			status = http.StatusNotFound
			message = "service or membership not found"
		}
		httpapi.WriteJSON(w, status, map[string]string{"error": "not_found", "message": message})
		return
	}
	httpapi.WriteJSON(w, http.StatusCreated, item)
}

func (h *handler) removeMember(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := endpointutil.ParseUUIDParam(w, r, "tenantId")
	if !ok || !endpointutil.RequireTenantPermission(w, r, h.authzSvc, tenantID, "services:update") {
		return
	}
	serviceID, ok := endpointutil.ParseUUIDParam(w, r, "serviceId")
	if !ok {
		return
	}
	serviceMemberID, ok := endpointutil.ParseUUIDParam(w, r, "serviceMemberId")
	if !ok {
		return
	}
	if err := h.repo.RemoveMember(r.Context(), tenantID, serviceID, serviceMemberID); err != nil {
		status := http.StatusInternalServerError
		message := "failed to remove service member"
		if err == ErrNotFound {
			status = http.StatusNotFound
			message = "service member not found"
		}
		httpapi.WriteJSON(w, status, map[string]string{"error": "not_found", "message": message})
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
