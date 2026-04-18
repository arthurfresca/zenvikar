package tenant_memberships

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/zenvikar/api/internal/platform/authn"
	"github.com/zenvikar/api/internal/platform/httpapi"
)

type handler struct {
	svc *Service
	db  *sql.DB
}

type createMembershipRequest struct {
	UserID      string  `json:"userId"`
	Role        string  `json:"role"`
	PhotoURL    *string `json:"photoUrl"`
	Description *string `json:"description"`
}

type updateMembershipRequest struct {
	Role        *string `json:"role"`
	PhotoURL    *string `json:"photoUrl"`
	Description *string `json:"description"`
}

func newHandler(svc *Service, db *sql.DB) *handler {
	return &handler{svc: svc, db: db}
}

func (h *handler) register(router chi.Router, requireAuth func(http.Handler) http.Handler) {
	router.Route("/api/v1/tenant/tenants/{tenantId}/memberships", func(r chi.Router) {
		r.Use(requireAuth)
		r.Get("/", h.list)
		r.Post("/", h.create)
		r.Get("/{membershipId}", h.get)
		r.Patch("/{membershipId}", h.update)
		r.Delete("/{membershipId}", h.remove)
	})
}

func (h *handler) list(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := parseUUIDParam(w, r, "tenantId")
	if !ok || !h.requirePermission(w, r, tenantID, "staff:read") {
		return
	}
	items, err := h.svc.ListByTenant(r.Context(), tenantID)
	if err != nil {
		httpapi.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal_error", "message": "failed to load memberships"})
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, map[string]any{"memberships": items})
}

func (h *handler) create(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := parseUUIDParam(w, r, "tenantId")
	if !ok || !h.requirePermission(w, r, tenantID, "staff:create") {
		return
	}
	var req createMembershipRequest
	if !httpapi.DecodeJSON(w, r, &req) {
		return
	}
	userID, err := uuid.Parse(req.UserID)
	if err != nil || !ValidRole(req.Role) {
		httpapi.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_request", "message": "invalid membership payload"})
		return
	}
	item, err := h.svc.Create(r.Context(), tenantID, CreateMembershipInput{UserID: userID, Role: req.Role, PhotoURL: req.PhotoURL, Description: req.Description})
	if err != nil {
		httpapi.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal_error", "message": "failed to create membership"})
		return
	}
	httpapi.WriteJSON(w, http.StatusCreated, item)
}

func (h *handler) get(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := parseUUIDParam(w, r, "tenantId")
	if !ok || !h.requirePermission(w, r, tenantID, "staff:read") {
		return
	}
	membershipID, ok := parseUUIDParam(w, r, "membershipId")
	if !ok {
		return
	}
	item, err := h.svc.GetByTenant(r.Context(), tenantID, membershipID)
	if err != nil {
		status := http.StatusInternalServerError
		message := "failed to load membership"
		if err == ErrNotFound {
			status = http.StatusNotFound
			message = "membership not found"
		}
		httpapi.WriteJSON(w, status, map[string]string{"error": "not_found", "message": message})
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, item)
}

func (h *handler) update(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := parseUUIDParam(w, r, "tenantId")
	if !ok || !h.requirePermission(w, r, tenantID, "staff:update") {
		return
	}
	membershipID, ok := parseUUIDParam(w, r, "membershipId")
	if !ok {
		return
	}
	var req updateMembershipRequest
	if !httpapi.DecodeJSON(w, r, &req) {
		return
	}
	if req.Role != nil && !ValidRole(*req.Role) {
		httpapi.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_request", "message": "invalid role"})
		return
	}
	var photo, description **string
	if req.PhotoURL != nil {
		photo = &req.PhotoURL
	}
	if req.Description != nil {
		description = &req.Description
	}
	item, err := h.svc.Update(r.Context(), tenantID, membershipID, UpdateMembershipInput{Role: req.Role, PhotoURL: photo, Description: description})
	if err != nil {
		status := http.StatusInternalServerError
		message := "failed to update membership"
		if err == ErrNotFound {
			status = http.StatusNotFound
			message = "membership not found"
		}
		httpapi.WriteJSON(w, status, map[string]string{"error": "not_found", "message": message})
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, item)
}

func (h *handler) remove(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := parseUUIDParam(w, r, "tenantId")
	if !ok || !h.requirePermission(w, r, tenantID, "staff:delete") {
		return
	}
	membershipID, ok := parseUUIDParam(w, r, "membershipId")
	if !ok {
		return
	}
	if err := h.svc.Delete(r.Context(), tenantID, membershipID); err != nil {
		status := http.StatusInternalServerError
		message := "failed to delete membership"
		if err == ErrNotFound {
			status = http.StatusNotFound
			message = "membership not found"
		}
		httpapi.WriteJSON(w, status, map[string]string{"error": "not_found", "message": message})
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (h *handler) requirePermission(w http.ResponseWriter, r *http.Request, tenantID uuid.UUID, permission string) bool {
	userID, ok := authn.UserIDFromContext(r.Context())
	if !ok {
		httpapi.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized", "message": "invalid or expired token"})
		return false
	}
	admin, err := h.isPlatformAdmin(r.Context(), userID)
	if err == nil && admin {
		return true
	}
	membership, err := h.svc.CheckMembership(r.Context(), userID, tenantID)
	if err != nil || !roleHasPermission(membership.Role, permission) {
		httpapi.WriteJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden", "message": "user does not have the required permission"})
		return false
	}
	return true
}

func (h *handler) isPlatformAdmin(ctx context.Context, userID uuid.UUID) (bool, error) {
	var exists bool
	err := h.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM platform_admins WHERE user_id = $1)`, userID).Scan(&exists)
	return exists, err
}

func roleHasPermission(role, permission string) bool {
	switch role {
	case RoleTenantOwner:
		return true
	case RoleTenantManager:
		return permission == "staff:read" || permission == "staff:create" || permission == "staff:update" || permission == "staff:delete"
	default:
		return false
	}
}

func parseUUIDParam(w http.ResponseWriter, r *http.Request, name string) (uuid.UUID, bool) {
	value := chi.URLParam(r, name)
	id, err := uuid.Parse(value)
	if err != nil {
		httpapi.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_request", "message": "invalid " + name})
		return uuid.UUID{}, false
	}
	return id, true
}
