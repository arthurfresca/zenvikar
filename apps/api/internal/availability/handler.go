package availability

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/zenvikar/api/internal/platform/authz"
	"github.com/zenvikar/api/internal/platform/endpointutil"
	"github.com/zenvikar/api/internal/platform/httpapi"
)

type handler struct {
	repo     *Repository
	authzSvc *authz.Service
}

type upsertOpeningHourRequest struct {
	OpenTime  string `json:"openTime"`
	CloseTime string `json:"closeTime"`
	Enabled   bool   `json:"enabled"`
}

type createBlockedDateRequest struct {
	Date   string  `json:"date"`
	Reason *string `json:"reason"`
}

func newHandler(repo *Repository, authzSvc *authz.Service) *handler {
	return &handler{repo: repo, authzSvc: authzSvc}
}

func (h *handler) register(router chi.Router, requireAuth func(http.Handler) http.Handler) {
	router.Get("/api/v1/tenants/{tenantSlug}/service-members/{serviceMemberId}/availability", h.listPublic)
	router.With(requireAuth).Get("/api/v1/tenant/tenants/{tenantId}/service-members/{serviceMemberId}/opening-hours", h.listOpeningHours)
	router.With(requireAuth).Put("/api/v1/tenant/tenants/{tenantId}/service-members/{serviceMemberId}/opening-hours/{dayOfWeek}", h.upsertOpeningHour)
	router.With(requireAuth).Get("/api/v1/tenant/tenants/{tenantId}/memberships/{membershipId}/blocked-dates", h.listBlockedDates)
	router.With(requireAuth).Post("/api/v1/tenant/tenants/{tenantId}/memberships/{membershipId}/blocked-dates", h.createBlockedDate)
	router.With(requireAuth).Delete("/api/v1/tenant/tenants/{tenantId}/memberships/{membershipId}/blocked-dates/{date}", h.deleteBlockedDate)
}

func (h *handler) listPublic(w http.ResponseWriter, r *http.Request) {
	serviceMemberID, ok := endpointutil.ParseUUIDParam(w, r, "serviceMemberId")
	if !ok {
		return
	}
	day, err := time.Parse("2006-01-02", r.URL.Query().Get("date"))
	if err != nil {
		httpapi.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_request", "message": "date query parameter is required in YYYY-MM-DD format"})
		return
	}
	slots, err := h.repo.ListPublicSlots(r.Context(), chi.URLParam(r, "tenantSlug"), serviceMemberID, day)
	if err != nil {
		status := http.StatusInternalServerError
		message := "failed to load availability"
		if err == ErrNotFound {
			status = http.StatusNotFound
			message = "service member not found"
		}
		httpapi.WriteJSON(w, status, map[string]string{"error": "not_found", "message": message})
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, map[string]any{"date": day.Format("2006-01-02"), "slots": slots})
}

func (h *handler) listOpeningHours(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := endpointutil.ParseUUIDParam(w, r, "tenantId")
	if !ok || !endpointutil.RequireTenantPermission(w, r, h.authzSvc, tenantID, "availability:read") {
		return
	}
	serviceMemberID, ok := endpointutil.ParseUUIDParam(w, r, "serviceMemberId")
	if !ok {
		return
	}
	items, err := h.repo.ListOpeningHours(r.Context(), tenantID, serviceMemberID)
	if err != nil {
		httpapi.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal_error", "message": "failed to load opening hours"})
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, map[string]any{"openingHours": items})
}

func (h *handler) upsertOpeningHour(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := endpointutil.ParseUUIDParam(w, r, "tenantId")
	if !ok || !endpointutil.RequireTenantPermission(w, r, h.authzSvc, tenantID, "availability:update") {
		return
	}
	serviceMemberID, ok := endpointutil.ParseUUIDParam(w, r, "serviceMemberId")
	if !ok {
		return
	}
	dayOfWeek, err := strconv.Atoi(chi.URLParam(r, "dayOfWeek"))
	if err != nil || dayOfWeek < 0 || dayOfWeek > 6 {
		httpapi.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_request", "message": "dayOfWeek must be between 0 and 6"})
		return
	}
	var req upsertOpeningHourRequest
	if !httpapi.DecodeJSON(w, r, &req) {
		return
	}
	if req.OpenTime == "" || req.CloseTime == "" || req.OpenTime >= req.CloseTime {
		httpapi.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_request", "message": "invalid opening hours payload"})
		return
	}
	item, err := h.repo.UpsertOpeningHour(r.Context(), tenantID, serviceMemberID, dayOfWeek, req.OpenTime, req.CloseTime, req.Enabled)
	if err != nil {
		status := http.StatusInternalServerError
		message := "failed to save opening hours"
		if err == ErrNotFound {
			status = http.StatusNotFound
			message = "service member not found"
		}
		httpapi.WriteJSON(w, status, map[string]string{"error": "not_found", "message": message})
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, item)
}

func (h *handler) listBlockedDates(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := endpointutil.ParseUUIDParam(w, r, "tenantId")
	if !ok || !endpointutil.RequireTenantPermission(w, r, h.authzSvc, tenantID, "availability:read") {
		return
	}
	membershipID, ok := endpointutil.ParseUUIDParam(w, r, "membershipId")
	if !ok {
		return
	}
	items, err := h.repo.ListBlockedDates(r.Context(), tenantID, membershipID)
	if err != nil {
		httpapi.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal_error", "message": "failed to load blocked dates"})
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, map[string]any{"blockedDates": items})
}

func (h *handler) createBlockedDate(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := endpointutil.ParseUUIDParam(w, r, "tenantId")
	if !ok || !endpointutil.RequireTenantPermission(w, r, h.authzSvc, tenantID, "availability:update") {
		return
	}
	membershipID, ok := endpointutil.ParseUUIDParam(w, r, "membershipId")
	if !ok {
		return
	}
	var req createBlockedDateRequest
	if !httpapi.DecodeJSON(w, r, &req) {
		return
	}
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		httpapi.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_request", "message": "date must be in YYYY-MM-DD format"})
		return
	}
	item, err := h.repo.CreateBlockedDate(r.Context(), tenantID, membershipID, date, req.Reason)
	if err != nil {
		status := http.StatusInternalServerError
		message := "failed to create blocked date"
		if err == ErrNotFound {
			status = http.StatusNotFound
			message = "membership not found"
		}
		httpapi.WriteJSON(w, status, map[string]string{"error": "not_found", "message": message})
		return
	}
	httpapi.WriteJSON(w, http.StatusCreated, item)
}

func (h *handler) deleteBlockedDate(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := endpointutil.ParseUUIDParam(w, r, "tenantId")
	if !ok || !endpointutil.RequireTenantPermission(w, r, h.authzSvc, tenantID, "availability:update") {
		return
	}
	membershipID, ok := endpointutil.ParseUUIDParam(w, r, "membershipId")
	if !ok {
		return
	}
	date, err := time.Parse("2006-01-02", chi.URLParam(r, "date"))
	if err != nil {
		httpapi.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_request", "message": "date must be in YYYY-MM-DD format"})
		return
	}
	if err := h.repo.DeleteBlockedDate(r.Context(), tenantID, membershipID, date); err != nil {
		status := http.StatusInternalServerError
		message := "failed to delete blocked date"
		if err == ErrNotFound {
			status = http.StatusNotFound
			message = "blocked date not found"
		}
		httpapi.WriteJSON(w, status, map[string]string{"error": "not_found", "message": message})
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
