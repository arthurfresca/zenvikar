package bookings

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/zenvikar/api/internal/availability"
	"github.com/zenvikar/api/internal/services"
)

// Service-level errors.
var (
	ErrSlotUnavailable = errors.New("slot_unavailable")
)

// CreateBookingRequest holds the data needed to create a booking.
type CreateBookingRequest struct {
	ServiceMemberID uuid.UUID
	CustomerID      uuid.UUID
	StartTime       time.Time
}

// AvailabilityDataLoader defines the interface for loading availability data.
type AvailabilityDataLoader interface {
	GetBlockedDates(ctx context.Context, membershipID uuid.UUID, date time.Time) ([]availability.BlockedDate, error)
	GetOpeningHours(ctx context.Context, serviceMemberID uuid.UUID) ([]availability.OpeningHours, error)
	GetExistingBookings(ctx context.Context, serviceMemberID uuid.UUID, rangeStart, rangeEnd time.Time) ([]Booking, error)
	GetServiceByMember(ctx context.Context, serviceMemberID uuid.UUID) (*services.Service, error)
	GetServiceMemberPrice(ctx context.Context, serviceMemberID uuid.UUID) (int, error)
	GetSlotIntervalMinutes(ctx context.Context, tenantID uuid.UUID) (int, error)
}

// BookingService handles booking creation with availability checking.
type BookingService struct {
	db     *sql.DB
	loader AvailabilityDataLoader
}

// NewBookingService creates a new BookingService.
func NewBookingService(db *sql.DB, loader AvailabilityDataLoader) *BookingService {
	return &BookingService{db: db, loader: loader}
}

// CreateBooking creates a new booking after checking availability.
func (s *BookingService) CreateBooking(
	ctx context.Context,
	tenantID uuid.UUID,
	req CreateBookingRequest,
) (*Booking, error) {
	svc, err := s.loader.GetServiceByMember(ctx, req.ServiceMemberID)
	if err != nil {
		return nil, fmt.Errorf("loading service: %w", err)
	}

	blockedDates, err := s.loader.GetBlockedDates(ctx, req.ServiceMemberID, req.StartTime)
	if err != nil {
		return nil, fmt.Errorf("loading blocked dates: %w", err)
	}

	openingHours, err := s.loader.GetOpeningHours(ctx, req.ServiceMemberID)
	if err != nil {
		return nil, fmt.Errorf("loading opening hours: %w", err)
	}

	endTime := req.StartTime.Add(time.Duration(svc.DurationMinutes) * time.Minute)
	rangeStart := req.StartTime.Add(-time.Duration(svc.BufferBefore) * time.Minute)
	rangeEnd := endTime.Add(time.Duration(svc.BufferAfter) * time.Minute)

	existingBookings, err := s.loader.GetExistingBookings(ctx, req.ServiceMemberID, rangeStart, rangeEnd)
	if err != nil {
		return nil, fmt.Errorf("loading existing bookings: %w", err)
	}

	slotInterval, err := s.loader.GetSlotIntervalMinutes(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("loading slot interval: %w", err)
	}

	result, err := CheckAvailability(blockedDates, openingHours, existingBookings, *svc, req.StartTime, slotInterval)
	if err != nil {
		return nil, fmt.Errorf("checking availability: %w", err)
	}

	if !result.Available {
		return nil, fmt.Errorf("%w: %s", ErrSlotUnavailable, result.Reason)
	}

	priceCents, err := s.loader.GetServiceMemberPrice(ctx, req.ServiceMemberID)
	if err != nil {
		return nil, fmt.Errorf("loading service member price: %w", err)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

	booking := &Booking{
		ID:              uuid.New(),
		TenantID:        tenantID,
		ServiceMemberID: req.ServiceMemberID,
		CustomerID:      req.CustomerID,
		PriceCents:      priceCents,
		StartTime:       req.StartTime,
		EndTime:         result.EndTime,
		Status:          StatusPending,
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO bookings (id, tenant_id, service_member_id, customer_id, price_cents, start_time, end_time, status, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		booking.ID, booking.TenantID, booking.ServiceMemberID, booking.CustomerID,
		booking.PriceCents, booking.StartTime, booking.EndTime, booking.Status,
		booking.CreatedAt, booking.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("inserting booking: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("committing transaction: %w", err)
	}

	return booking, nil
}
