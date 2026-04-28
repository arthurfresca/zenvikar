package availability

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ErrNotFound indicates the requested availability resource does not exist in scope.
var ErrNotFound = fmt.Errorf("not found")

type serviceMemberContext struct {
	TenantID        uuid.UUID
	MembershipID    uuid.UUID
	ServiceID       uuid.UUID
	DurationMinutes int
	BufferBefore    int
	BufferAfter     int
	Interval        int
	Timezone        string
}

type existingBookingWindow struct {
	StartTime time.Time
	EndTime   time.Time
	Status    string
}

// AvailableTime is a bookable time window for a service member.
type AvailableTime struct {
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
}

// Repository provides availability persistence.
type Repository struct {
	db *sql.DB
}

// NewRepository creates an availability repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// ListPublicTimes returns bookable times for a service member on a date.
func (r *Repository) ListPublicTimes(ctx context.Context, tenantSlug string, serviceMemberID uuid.UUID, day time.Time) ([]AvailableTime, error) {
	ctxData, err := r.getServiceMemberContextBySlug(ctx, tenantSlug, serviceMemberID)
	if err != nil {
		return nil, err
	}
	return r.listTimesForDay(ctx, ctxData, serviceMemberID, day)
}

func (r *Repository) ListOpeningHours(ctx context.Context, tenantID, serviceMemberID uuid.UUID) ([]OpeningHours, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT oh.id, oh.service_member_id, oh.day_of_week, oh.open_time, oh.close_time, oh.enabled
		FROM opening_hours oh
		JOIN service_members sm ON sm.id = oh.service_member_id
		JOIN services s ON s.id = sm.service_id
		WHERE s.tenant_id = $1 AND oh.service_member_id = $2
		ORDER BY oh.day_of_week ASC
	`, tenantID, serviceMemberID)
	if err != nil {
		return nil, fmt.Errorf("listing opening hours: %w", err)
	}
	defer rows.Close()

	var out []OpeningHours
	for rows.Next() {
		var item OpeningHours
		if err := rows.Scan(&item.ID, &item.ServiceMemberID, &item.DayOfWeek, &item.OpenTime, &item.CloseTime, &item.Enabled); err != nil {
			return nil, fmt.Errorf("scanning opening hours: %w", err)
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

// ServiceMemberBelongsToMembership reports whether the service member belongs to the given tenant membership.
func (r *Repository) ServiceMemberBelongsToMembership(ctx context.Context, tenantID, serviceMemberID, membershipID uuid.UUID) (bool, error) {
	var ok bool
	err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1
			FROM service_members sm
			JOIN services s ON s.id = sm.service_id
			WHERE s.tenant_id = $1 AND sm.id = $2 AND sm.membership_id = $3
		)
	`, tenantID, serviceMemberID, membershipID).Scan(&ok)
	if err != nil {
		return false, fmt.Errorf("checking service member ownership: %w", err)
	}
	return ok, nil
}

func (r *Repository) UpsertOpeningHour(ctx context.Context, tenantID, serviceMemberID uuid.UUID, dayOfWeek int, openTime, closeTime string, enabled bool) (*OpeningHours, error) {
	var ok bool
	err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM service_members sm
			JOIN services s ON s.id = sm.service_id
			WHERE s.tenant_id = $1 AND sm.id = $2
		)
	`, tenantID, serviceMemberID).Scan(&ok)
	if err != nil {
		return nil, fmt.Errorf("validating opening hours scope: %w", err)
	}
	if !ok {
		return nil, ErrNotFound
	}

	var item OpeningHours
	err = r.db.QueryRowContext(ctx, `
		INSERT INTO opening_hours (id, service_member_id, day_of_week, open_time, close_time, enabled)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (service_member_id, day_of_week)
		DO UPDATE SET open_time = EXCLUDED.open_time, close_time = EXCLUDED.close_time, enabled = EXCLUDED.enabled
		RETURNING id, service_member_id, day_of_week, open_time::text, close_time::text, enabled
	`, uuid.New(), serviceMemberID, dayOfWeek, openTime, closeTime, enabled).Scan(
		&item.ID, &item.ServiceMemberID, &item.DayOfWeek, &item.OpenTime, &item.CloseTime, &item.Enabled,
	)
	if err != nil {
		return nil, fmt.Errorf("upserting opening hour: %w", err)
	}
	return &item, nil
}

func (r *Repository) ListBlockedDates(ctx context.Context, tenantID, membershipID uuid.UUID) ([]BlockedDate, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT bd.id, bd.membership_id, bd.date, bd.reason
		FROM blocked_dates bd
		JOIN tenant_memberships tm ON tm.id = bd.membership_id
		WHERE tm.tenant_id = $1 AND bd.membership_id = $2
		ORDER BY bd.date ASC
	`, tenantID, membershipID)
	if err != nil {
		return nil, fmt.Errorf("listing blocked dates: %w", err)
	}
	defer rows.Close()

	var out []BlockedDate
	for rows.Next() {
		var item BlockedDate
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

func (r *Repository) CreateBlockedDate(ctx context.Context, tenantID, membershipID uuid.UUID, date time.Time, reason *string) (*BlockedDate, error) {
	var ok bool
	err := r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM tenant_memberships WHERE tenant_id = $1 AND id = $2)`, tenantID, membershipID).Scan(&ok)
	if err != nil {
		return nil, fmt.Errorf("validating membership scope: %w", err)
	}
	if !ok {
		return nil, ErrNotFound
	}

	var item BlockedDate
	var reasonSQL sql.NullString
	if reason != nil {
		reasonSQL = sql.NullString{String: *reason, Valid: true}
	}
	err = r.db.QueryRowContext(ctx, `
		INSERT INTO blocked_dates (id, membership_id, date, reason)
		VALUES ($1, $2, $3, $4)
		RETURNING id, membership_id, date, reason
	`, uuid.New(), membershipID, date, reasonSQL).Scan(&item.ID, &item.MembershipID, &item.Date, &reasonSQL)
	if err != nil {
		return nil, fmt.Errorf("creating blocked date: %w", err)
	}
	if reasonSQL.Valid {
		item.Reason = &reasonSQL.String
	}
	return &item, nil
}

func (r *Repository) DeleteBlockedDate(ctx context.Context, tenantID, membershipID uuid.UUID, date time.Time) error {
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM blocked_dates bd
		USING tenant_memberships tm
		WHERE bd.membership_id = tm.id AND tm.tenant_id = $1 AND bd.membership_id = $2 AND bd.date = $3
	`, tenantID, membershipID, date)
	if err != nil {
		return fmt.Errorf("deleting blocked date: %w", err)
	}
	rows, err := result.RowsAffected()
	if err == nil && rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) getServiceMemberContextBySlug(ctx context.Context, tenantSlug string, serviceMemberID uuid.UUID) (*serviceMemberContext, error) {
	var item serviceMemberContext
	err := r.db.QueryRowContext(ctx, `
		SELECT s.tenant_id, sm.membership_id, s.id, s.duration_minutes, s.buffer_before_minutes,
		       s.buffer_after_minutes, t.slot_interval_minutes, t.timezone
		FROM service_members sm
		JOIN services s ON s.id = sm.service_id
		JOIN tenants t ON t.id = s.tenant_id
		WHERE t.slug = $1 AND t.enabled = true AND s.enabled = true AND sm.id = $2
	`, tenantSlug, serviceMemberID).Scan(
		&item.TenantID, &item.MembershipID, &item.ServiceID, &item.DurationMinutes, &item.BufferBefore, &item.BufferAfter, &item.Interval, &item.Timezone,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("loading service member context: %w", err)
	}
	return &item, nil
}

func (r *Repository) listTimesForDay(ctx context.Context, data *serviceMemberContext, serviceMemberID uuid.UUID, day time.Time) ([]AvailableTime, error) {
	dateOnly := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, time.UTC)
	var blocked bool
	err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS(SELECT 1 FROM blocked_dates WHERE membership_id = $1 AND date = $2)
	`, data.MembershipID, dateOnly).Scan(&blocked)
	if err != nil {
		return nil, fmt.Errorf("checking blocked dates: %w", err)
	}
	if blocked {
		return []AvailableTime{}, nil
	}

	var openTime, closeTime string
	var enabled bool
	err = r.db.QueryRowContext(ctx, `
		SELECT open_time::text, close_time::text, enabled
		FROM opening_hours
		WHERE service_member_id = $1 AND day_of_week = $2
	`, serviceMemberID, int(day.Weekday())).Scan(&openTime, &closeTime, &enabled)
	if err != nil {
		if err == sql.ErrNoRows {
			return []AvailableTime{}, nil
		}
		return nil, fmt.Errorf("loading opening hours: %w", err)
	}
	if !enabled {
		return []AvailableTime{}, nil
	}

	loc, err := time.LoadLocation(data.Timezone)
	if err != nil {
		loc = time.UTC
	}
	localDay := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, loc)
	openAt, err := parseClockOnDate(openTime, localDay)
	if err != nil {
		return nil, err
	}
	closeAt, err := parseClockOnDate(closeTime, localDay)
	if err != nil {
		return nil, err
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT start_time, end_time, status
		FROM bookings
		WHERE service_member_id = $1 AND status != 'cancelled' AND start_time < $3 AND end_time > $2
		ORDER BY start_time ASC
	`, serviceMemberID, openAt.UTC().Add(-time.Duration(data.BufferBefore)*time.Minute), closeAt.UTC().Add(time.Duration(data.BufferAfter)*time.Minute))
	if err != nil {
		return nil, fmt.Errorf("loading existing bookings: %w", err)
	}
	defer rows.Close()

	var existing []existingBookingWindow
	for rows.Next() {
		var item existingBookingWindow
		if err := rows.Scan(&item.StartTime, &item.EndTime, &item.Status); err != nil {
			return nil, fmt.Errorf("scanning booking window: %w", err)
		}
		existing = append(existing, item)
	}

	step := time.Duration(data.Interval) * time.Minute
	serviceDuration := time.Duration(data.DurationMinutes) * time.Minute
	var times []AvailableTime
	for candidate := openAt; !candidate.Add(serviceDuration).After(closeAt); candidate = candidate.Add(step) {
		endAt := candidate.Add(serviceDuration)
		effectiveStart := candidate.Add(-time.Duration(data.BufferBefore) * time.Minute)
		effectiveEnd := endAt.Add(time.Duration(data.BufferAfter) * time.Minute)
		available := true
		for _, item := range existing {
			if effectiveStart.Before(item.EndTime.In(loc)) && effectiveEnd.After(item.StartTime.In(loc)) {
				available = false
				break
			}
		}
		if available {
			times = append(times, AvailableTime{StartTime: candidate.UTC(), EndTime: endAt.UTC()})
		}
	}
	return times, nil
}

func parseClockOnDate(timeOfDay string, day time.Time) (time.Time, error) {
	t, err := time.Parse("15:04:05", timeOfDay)
	if err != nil {
		t, err = time.Parse("15:04", timeOfDay)
		if err != nil {
			return time.Time{}, fmt.Errorf("parsing time: %w", err)
		}
	}
	return time.Date(day.Year(), day.Month(), day.Day(), t.Hour(), t.Minute(), 0, 0, day.Location()), nil
}
