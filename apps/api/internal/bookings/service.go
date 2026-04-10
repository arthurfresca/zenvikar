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
	ServiceID  uuid.UUID
	CustomerID uuid.UUID
	StartTime  time.Time
	Timezone   string
}

// AvailabilityDataLoader defines the interface for loading availability data.
// This allows the service to be tested with mock data loaders.
type AvailabilityDataLoader interface {
	GetBlockedDates(ctx context.Context, tenantID uuid.UUID, date time.Time) ([]availability.BlockedDate, error)
	GetOpeningHours(ctx context.Context, tenantID uuid.UUID) ([]availability.OpeningHours, error)
	GetExistingBookings(ctx context.Context, tenantID, serviceID uuid.UUID, rangeStart, rangeEnd time.Time) ([]Booking, error)
	GetService(ctx context.Context, serviceID uuid.UUID) (*services.Service, error)
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
// The booking is created within a transaction with status "pending".
func (s *BookingService) CreateBooking(
	ctx context.Context,
	tenantID uuid.UUID,
	req CreateBookingRequest,
) (*Booking, error) {
	// Load service
	svc, err := s.loader.GetService(ctx, req.ServiceID)
	if err != nil {
		return nil, fmt.Errorf("loading service: %w", err)
	}

	// Load availability data
	blockedDates, err := s.loader.GetBlockedDates(ctx, tenantID, req.StartTime)
	if err != nil {
		return nil, fmt.Errorf("loading blocked dates: %w", err)
	}

	openingHours, err := s.loader.GetOpeningHours(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("loading opening hours: %w", err)
	}

	// Calculate the time range to check for existing bookings (with buffer)
	endTime := req.StartTime.Add(time.Duration(svc.DurationMinutes) * time.Minute)
	rangeStart := req.StartTime.Add(-time.Duration(svc.BufferBefore) * time.Minute)
	rangeEnd := endTime.Add(time.Duration(svc.BufferAfter) * time.Minute)

	existingBookings, err := s.loader.GetExistingBookings(ctx, tenantID, req.ServiceID, rangeStart, rangeEnd)
	if err != nil {
		return nil, fmt.Errorf("loading existing bookings: %w", err)
	}

	// Check availability using the pure function
	result, err := CheckAvailability(blockedDates, openingHours, existingBookings, *svc, req.StartTime)
	if err != nil {
		return nil, fmt.Errorf("checking availability: %w", err)
	}

	if !result.Available {
		return nil, fmt.Errorf("%w: %s", ErrSlotUnavailable, result.Reason)
	}

	// Create booking within a transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

	booking := &Booking{
		ID:         uuid.New(),
		TenantID:   tenantID,
		ServiceID:  req.ServiceID,
		CustomerID: req.CustomerID,
		StartTime:  req.StartTime,
		EndTime:    result.EndTime,
		Status:     StatusPending,
		Timezone:   req.Timezone,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO bookings (id, tenant_id, service_id, customer_id, start_time, end_time, status, timezone, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		booking.ID, booking.TenantID, booking.ServiceID, booking.CustomerID,
		booking.StartTime, booking.EndTime, booking.Status, booking.Timezone,
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
