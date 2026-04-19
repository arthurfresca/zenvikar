package bookings

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/zenvikar/api/internal/availability"
	servicesdomain "github.com/zenvikar/api/internal/services"
)

// ErrNotFound indicates the requested booking resource does not exist.
var ErrNotFound = fmt.Errorf("not found")

// Repository provides booking persistence and availability data loading.
type Repository struct {
	db *sql.DB
}

// NewRepository creates a bookings repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// GetBlockedDates loads blocked dates for a service member via its membership.
func (r *Repository) GetBlockedDates(ctx context.Context, serviceMemberID uuid.UUID, date time.Time) ([]availability.BlockedDate, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT bd.id, bd.membership_id, bd.date, bd.reason
		FROM blocked_dates bd
		JOIN service_members sm ON sm.membership_id = bd.membership_id
		WHERE sm.id = $1 AND bd.date = $2
	`, serviceMemberID, time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC))
	if err != nil {
		return nil, fmt.Errorf("loading blocked dates: %w", err)
	}
	defer rows.Close()

	var out []availability.BlockedDate
	for rows.Next() {
		var item availability.BlockedDate
		var reason sql.NullString
		if err := rows.Scan(&item.ID, &item.MembershipID, &item.Date, &reason); err != nil {
			return nil, fmt.Errorf("scanning blocked date: %w", err)
		}
		if reason.Valid {
			item.Reason = &reason.String
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

// GetOpeningHours loads opening hours for a service member.
func (r *Repository) GetOpeningHours(ctx context.Context, serviceMemberID uuid.UUID) ([]availability.OpeningHours, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, service_member_id, day_of_week, open_time::text, close_time::text, enabled
		FROM opening_hours
		WHERE service_member_id = $1
		ORDER BY day_of_week ASC
	`, serviceMemberID)
	if err != nil {
		return nil, fmt.Errorf("loading opening hours: %w", err)
	}
	defer rows.Close()

	var out []availability.OpeningHours
	for rows.Next() {
		var item availability.OpeningHours
		if err := rows.Scan(&item.ID, &item.ServiceMemberID, &item.DayOfWeek, &item.OpenTime, &item.CloseTime, &item.Enabled); err != nil {
			return nil, fmt.Errorf("scanning opening hour: %w", err)
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

// GetExistingBookings loads overlapping bookings for a service member.
func (r *Repository) GetExistingBookings(ctx context.Context, serviceMemberID uuid.UUID, rangeStart, rangeEnd time.Time) ([]Booking, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, tenant_id, service_member_id, customer_id, price_cents, start_time, end_time, status, created_at, updated_at
		FROM bookings
		WHERE service_member_id = $1 AND status != 'cancelled' AND start_time < $3 AND end_time > $2
		ORDER BY start_time ASC
	`, serviceMemberID, rangeStart, rangeEnd)
	if err != nil {
		return nil, fmt.Errorf("loading existing bookings: %w", err)
	}
	defer rows.Close()

	var out []Booking
	for rows.Next() {
		var item Booking
		if err := rows.Scan(&item.ID, &item.TenantID, &item.ServiceMemberID, &item.CustomerID, &item.PriceCents, &item.StartTime, &item.EndTime, &item.Status, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning booking: %w", err)
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

// GetServiceByMember loads service settings for a service member.
func (r *Repository) GetServiceByMember(ctx context.Context, serviceMemberID uuid.UUID) (*servicesdomain.Service, error) {
	var item servicesdomain.Service
	var description sql.NullString
	err := r.db.QueryRowContext(ctx, `
		SELECT s.id, s.tenant_id, s.name, s.description, s.duration_minutes, s.buffer_before_minutes,
		       s.buffer_after_minutes, s.enabled, s.created_at, s.updated_at
		FROM services s
		JOIN service_members sm ON sm.service_id = s.id
		WHERE sm.id = $1
	`, serviceMemberID).Scan(&item.ID, &item.TenantID, &item.Name, &description, &item.DurationMinutes, &item.BufferBefore, &item.BufferAfter, &item.Enabled, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("loading service by member: %w", err)
	}
	if description.Valid {
		item.Description = &description.String
	}
	return &item, nil
}

// GetServiceMemberPrice loads the configured price for a service member.
func (r *Repository) GetServiceMemberPrice(ctx context.Context, serviceMemberID uuid.UUID) (int, error) {
	var priceCents int
	err := r.db.QueryRowContext(ctx, `SELECT price_cents FROM service_members WHERE id = $1`, serviceMemberID).Scan(&priceCents)
	if err != nil {
		return 0, fmt.Errorf("loading service member price: %w", err)
	}
	return priceCents, nil
}

// GetSlotIntervalMinutes loads tenant slot interval.
func (r *Repository) GetSlotIntervalMinutes(ctx context.Context, tenantID uuid.UUID) (int, error) {
	var slotInterval int
	err := r.db.QueryRowContext(ctx, `SELECT slot_interval_minutes FROM tenants WHERE id = $1`, tenantID).Scan(&slotInterval)
	if err != nil {
		return 0, fmt.Errorf("loading slot interval: %w", err)
	}
	return slotInterval, nil
}

// ServiceMemberBelongsToTenant reports whether a service member belongs to a tenant.
func (r *Repository) ServiceMemberBelongsToTenant(ctx context.Context, tenantID, serviceMemberID uuid.UUID) (bool, error) {
	var ok bool
	err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM service_members sm
			JOIN services s ON s.id = sm.service_id
			WHERE s.tenant_id = $1 AND sm.id = $2 AND s.enabled = true
		)
	`, tenantID, serviceMemberID).Scan(&ok)
	if err != nil {
		return false, fmt.Errorf("validating service member scope: %w", err)
	}
	return ok, nil
}

// GetByCustomer returns one booking owned by a customer.
func (r *Repository) GetByCustomer(ctx context.Context, bookingID, customerID uuid.UUID) (*Booking, error) {
	return r.getOne(ctx, `
		SELECT id, tenant_id, service_member_id, customer_id, price_cents, start_time, end_time, status, created_at, updated_at
		FROM bookings WHERE id = $1 AND customer_id = $2
	`, bookingID, customerID)
}

// ListByCustomer returns bookings owned by a customer.
func (r *Repository) ListByCustomer(ctx context.Context, customerID uuid.UUID) ([]Booking, error) {
	return r.list(ctx, `
		SELECT id, tenant_id, service_member_id, customer_id, price_cents, start_time, end_time, status, created_at, updated_at
		FROM bookings WHERE customer_id = $1 ORDER BY start_time DESC
	`, customerID)
}

// CancelByCustomer cancels a booking owned by a customer.
func (r *Repository) CancelByCustomer(ctx context.Context, bookingID, customerID uuid.UUID) (*Booking, error) {
	result, err := r.db.ExecContext(ctx, `
		UPDATE bookings SET status = 'cancelled', updated_at = NOW() WHERE id = $1 AND customer_id = $2 AND status != 'cancelled'
	`, bookingID, customerID)
	if err != nil {
		return nil, fmt.Errorf("cancelling booking: %w", err)
	}
	rows, err := result.RowsAffected()
	if err == nil && rows == 0 {
		return nil, ErrNotFound
	}
	return r.GetByCustomer(ctx, bookingID, customerID)
}

// GetByTenant returns one booking in tenant scope.
func (r *Repository) GetByTenant(ctx context.Context, bookingID, tenantID uuid.UUID) (*Booking, error) {
	return r.getOne(ctx, `
		SELECT id, tenant_id, service_member_id, customer_id, price_cents, start_time, end_time, status, created_at, updated_at
		FROM bookings WHERE id = $1 AND tenant_id = $2
	`, bookingID, tenantID)
}

// ListByTenant returns bookings in tenant scope.
func (r *Repository) ListByTenant(ctx context.Context, tenantID uuid.UUID, from, to *time.Time) ([]Booking, error) {
	return r.listByTenant(ctx, tenantID, nil, from, to)
}

// ListByTenantMembership returns bookings in tenant scope limited to one membership's service members.
func (r *Repository) ListByTenantMembership(ctx context.Context, tenantID, membershipID uuid.UUID, from, to *time.Time) ([]Booking, error) {
	return r.listByTenant(ctx, tenantID, &membershipID, from, to)
}

func (r *Repository) listByTenant(ctx context.Context, tenantID uuid.UUID, membershipID *uuid.UUID, from, to *time.Time) ([]Booking, error) {
	query := `
		SELECT id, tenant_id, service_member_id, customer_id, price_cents, start_time, end_time, status, created_at, updated_at
		FROM bookings b`
	args := []any{tenantID}
	if membershipID != nil {
		query += ` JOIN service_members sm ON sm.id = b.service_member_id WHERE b.tenant_id = $1 AND sm.membership_id = $2`
		args = append(args, *membershipID)
	} else {
		query += ` WHERE b.tenant_id = $1`
	}
	if from != nil {
		query += fmt.Sprintf(" AND b.end_time >= $%d", len(args)+1)
		args = append(args, *from)
	}
	if to != nil {
		query += fmt.Sprintf(" AND b.start_time <= $%d", len(args)+1)
		args = append(args, *to)
	}
	query += ` ORDER BY b.start_time DESC`
	return r.list(ctx, query, args...)
}

// UpdateStatusInTenant updates a booking status in tenant scope.
func (r *Repository) UpdateStatusInTenant(ctx context.Context, bookingID, tenantID uuid.UUID, status string) (*Booking, error) {
	result, err := r.db.ExecContext(ctx, `
		UPDATE bookings SET status = $3, updated_at = NOW() WHERE id = $1 AND tenant_id = $2
	`, bookingID, tenantID, status)
	if err != nil {
		return nil, fmt.Errorf("updating booking status: %w", err)
	}
	rows, err := result.RowsAffected()
	if err == nil && rows == 0 {
		return nil, ErrNotFound
	}
	return r.GetByTenant(ctx, bookingID, tenantID)
}

// BookingBelongsToMembership reports whether a booking is assigned to one of the membership's service members.
func (r *Repository) BookingBelongsToMembership(ctx context.Context, bookingID, tenantID, membershipID uuid.UUID) (bool, error) {
	var ok bool
	err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1
			FROM bookings b
			JOIN service_members sm ON sm.id = b.service_member_id
			WHERE b.id = $1 AND b.tenant_id = $2 AND sm.membership_id = $3
		)
	`, bookingID, tenantID, membershipID).Scan(&ok)
	if err != nil {
		return false, fmt.Errorf("checking booking ownership: %w", err)
	}
	return ok, nil
}

func (r *Repository) getOne(ctx context.Context, query string, args ...any) (*Booking, error) {
	var item Booking
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&item.ID, &item.TenantID, &item.ServiceMemberID, &item.CustomerID, &item.PriceCents, &item.StartTime, &item.EndTime, &item.Status, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("loading booking: %w", err)
	}
	return &item, nil
}

func (r *Repository) list(ctx context.Context, query string, args ...any) ([]Booking, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("listing bookings: %w", err)
	}
	defer rows.Close()
	var out []Booking
	for rows.Next() {
		var item Booking
		if err := rows.Scan(&item.ID, &item.TenantID, &item.ServiceMemberID, &item.CustomerID, &item.PriceCents, &item.StartTime, &item.EndTime, &item.Status, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning booking: %w", err)
		}
		out = append(out, item)
	}
	return out, rows.Err()
}
