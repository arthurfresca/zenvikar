package bookings

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/zenvikar/api/internal/platform/authz"
	"github.com/zenvikar/api/internal/platform/endpointutil"
	"github.com/zenvikar/api/internal/platform/httpapi"
	"github.com/zenvikar/api/internal/tenant_memberships"
	"github.com/zenvikar/api/internal/tenants"
)

type handler struct {
	repo       *Repository
	bookingSvc *BookingService
	tenantSvc  *tenants.Service
	authzSvc   *authz.Service
	memberSvc  *tenant_memberships.Service
}

type createBookingRequest struct {
	ServiceMemberID string `json:"serviceMemberId"`
	StartTime       string `json:"startTime"`
}

type updateTenantBookingRequest struct {
	Status string `json:"status"`
}

func newHandler(repo *Repository, bookingSvc *BookingService, tenantSvc *tenants.Service, authzSvc *authz.Service, memberSvc *tenant_memberships.Service) *handler {
	return &handler{repo: repo, bookingSvc: bookingSvc, tenantSvc: tenantSvc, authzSvc: authzSvc, memberSvc: memberSvc}
}

func (h *handler) register(router chi.Router, requireAuth func(http.Handler) http.Handler) {
	router.With(requireAuth).Post("/api/v1/tenants/{tenantSlug}/bookings", h.create)
	router.With(requireAuth).Get("/api/v1/me/bookings", h.listMine)
	router.With(requireAuth).Get("/api/v1/me/bookings/{bookingId}", h.getMine)
	router.With(requireAuth).Post("/api/v1/me/bookings/{bookingId}/cancel", h.cancelMine)

	router.Route("/api/v1/tenant/tenants/{tenantId}/bookings", func(r chi.Router) {
		r.Use(requireAuth)
		r.Get("/", h.listTenant)
		r.Get("/{bookingId}", h.getTenant)
		r.Patch("/{bookingId}", h.updateTenant)
	})
}

func (h *handler) create(w http.ResponseWriter, r *http.Request) {
	userID, ok := endpointutil.CurrentUserID(w, r)
	if !ok {
		return
	}
	tenantSlug := strings.TrimSpace(chi.URLParam(r, "tenantSlug"))
	if !endpointutil.RequireCurrentTenantSlug(w, r, tenantSlug) {
		return
	}
	tenant, err := h.tenantSvc.ResolveTenantBySlug(r.Context(), tenantSlug)
	if err != nil {
		httpapi.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "tenant_not_found", "message": "No tenant found for this address"})
		return
	}
	if !endpointutil.RequireCurrentTenantID(w, r, tenant.ID) {
		return
	}
	var req createBookingRequest
	if !httpapi.DecodeJSON(w, r, &req) {
		return
	}
	serviceMemberID, err := uuid.Parse(req.ServiceMemberID)
	if err != nil {
		httpapi.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_request", "message": "invalid serviceMemberId"})
		return
	}
	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		httpapi.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_request", "message": "startTime must be a valid RFC3339 timestamp"})
		return
	}
	inScope, err := h.repo.ServiceMemberBelongsToTenant(r.Context(), tenant.ID, serviceMemberID)
	if err != nil || !inScope {
		httpapi.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "not_found", "message": "service member not found"})
		return
	}
	booking, err := h.bookingSvc.CreateBooking(r.Context(), tenant.ID, CreateBookingRequest{ServiceMemberID: serviceMemberID, CustomerID: userID, StartTime: startTime})
	if err != nil {
		status := http.StatusInternalServerError
		message := "failed to create booking"
		if errors.Is(err, ErrTimeUnavailable) {
			status = http.StatusConflict
			message = "The requested time is no longer available"
		}
		httpapi.WriteJSON(w, status, map[string]string{"error": "time_unavailable", "message": message})
		return
	}
	httpapi.WriteJSON(w, http.StatusCreated, booking)
}

func (h *handler) listMine(w http.ResponseWriter, r *http.Request) {
	userID, ok := endpointutil.CurrentUserID(w, r)
	if !ok {
		return
	}
	items, err := h.repo.ListByCustomer(r.Context(), userID)
	if err != nil {
		httpapi.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal_error", "message": "failed to load bookings"})
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, map[string]any{"bookings": items})
}

func (h *handler) getMine(w http.ResponseWriter, r *http.Request) {
	userID, ok := endpointutil.CurrentUserID(w, r)
	if !ok {
		return
	}
	bookingID, ok := endpointutil.ParseUUIDParam(w, r, "bookingId")
	if !ok {
		return
	}
	item, err := h.repo.GetByCustomer(r.Context(), bookingID, userID)
	if err != nil {
		httpapi.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "not_found", "message": "booking not found"})
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, item)
}

func (h *handler) cancelMine(w http.ResponseWriter, r *http.Request) {
	userID, ok := endpointutil.CurrentUserID(w, r)
	if !ok {
		return
	}
	bookingID, ok := endpointutil.ParseUUIDParam(w, r, "bookingId")
	if !ok {
		return
	}
	item, err := h.repo.CancelByCustomer(r.Context(), bookingID, userID)
	if err != nil {
		httpapi.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "not_found", "message": "booking not found"})
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, item)
}

func (h *handler) listTenant(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := endpointutil.ParseUUIDParam(w, r, "tenantId")
	if !ok || !endpointutil.RequireTenantPermission(w, r, h.authzSvc, tenantID, "bookings:read") {
		return
	}
	var from, to *time.Time
	if raw := strings.TrimSpace(r.URL.Query().Get("from")); raw != "" {
		value, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			httpapi.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_request", "message": "from must be RFC3339"})
			return
		}
		from = &value
	}
	if raw := strings.TrimSpace(r.URL.Query().Get("to")); raw != "" {
		value, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			httpapi.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_request", "message": "to must be RFC3339"})
			return
		}
		to = &value
	}
	userID, ok := endpointutil.CurrentUserID(w, r)
	if !ok {
		return
	}
	membership, err := h.memberSvc.CheckMembership(r.Context(), userID, tenantID)
	if err != nil {
		httpapi.WriteJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden", "message": "user is not a member of this tenant"})
		return
	}
	var items []BookingDetails
	if membership.Role == tenant_memberships.RoleTenantStaff {
		items, err = h.repo.ListByTenantMembership(r.Context(), tenantID, membership.ID, from, to)
	} else {
		items, err = h.repo.ListByTenant(r.Context(), tenantID, from, to)
	}
	if err != nil {
		httpapi.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal_error", "message": "failed to load tenant bookings"})
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, map[string]any{"bookings": items})
}

func (h *handler) getTenant(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := endpointutil.ParseUUIDParam(w, r, "tenantId")
	if !ok || !endpointutil.RequireTenantPermission(w, r, h.authzSvc, tenantID, "bookings:read") {
		return
	}
	bookingID, ok := endpointutil.ParseUUIDParam(w, r, "bookingId")
	if !ok {
		return
	}
	if !h.ensureStaffOwnBooking(w, r, tenantID, bookingID) {
		return
	}
	item, err := h.repo.GetByTenant(r.Context(), bookingID, tenantID)
	if err != nil {
		httpapi.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "not_found", "message": "booking not found"})
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, item)
}

func (h *handler) updateTenant(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := endpointutil.ParseUUIDParam(w, r, "tenantId")
	if !ok || !endpointutil.RequireTenantPermission(w, r, h.authzSvc, tenantID, "bookings:update") {
		return
	}
	bookingID, ok := endpointutil.ParseUUIDParam(w, r, "bookingId")
	if !ok {
		return
	}
	if !h.ensureStaffOwnBooking(w, r, tenantID, bookingID) {
		return
	}
	var req updateTenantBookingRequest
	if !httpapi.DecodeJSON(w, r, &req) {
		return
	}
	if req.Status != StatusPending && req.Status != StatusConfirmed && req.Status != StatusCancelled {
		httpapi.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_request", "message": "status must be pending, confirmed, or cancelled"})
		return
	}
	item, err := h.repo.UpdateStatusInTenant(r.Context(), bookingID, tenantID, req.Status)
	if err != nil {
		httpapi.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "not_found", "message": "booking not found"})
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, item)
}

func (h *handler) ensureStaffOwnBooking(w http.ResponseWriter, r *http.Request, tenantID, bookingID uuid.UUID) bool {
	userID, ok := endpointutil.CurrentUserID(w, r)
	if !ok {
		return false
	}
	membership, err := h.memberSvc.CheckMembership(r.Context(), userID, tenantID)
	if err != nil {
		httpapi.WriteJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden", "message": "user is not a member of this tenant"})
		return false
	}
	if membership.Role != tenant_memberships.RoleTenantStaff {
		return true
	}
	ok, err = h.repo.BookingBelongsToMembership(r.Context(), bookingID, tenantID, membership.ID)
	if err != nil {
		httpapi.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal_error", "message": "failed to validate booking scope"})
		return false
	}
	if !ok {
		httpapi.WriteJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden", "message": "staff can only access their own calendar"})
		return false
	}
	return true
}
