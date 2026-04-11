package bookings

import (
	"errors"
	"fmt"
	"time"

	"github.com/zenvikar/api/internal/availability"
	"github.com/zenvikar/api/internal/services"
)

// Availability check errors.
var (
	ErrDateBlocked       = errors.New("date_blocked")
	ErrOutsideHours      = errors.New("outside_hours")
	ErrSlotTaken         = errors.New("slot_taken")
)

// AvailabilityResult holds the result of an availability check.
type AvailabilityResult struct {
	Available bool
	Reason    string
	EndTime   time.Time
}

// CheckAvailability is a pure function that determines if a time slot is bookable.
// It takes all required data as input so it can be tested without a database.
//
// Parameters:
//   - blockedDates: list of blocked dates for the member
//   - openingHours: list of opening hours for the service_member (all days)
//   - existingBookings: list of existing non-cancelled bookings for the member
//   - service: the service being booked (for duration and buffers)
//   - startTime: the requested booking start time
//   - slotIntervalMinutes: the tenant's slot interval (e.g. 15). Start time must align to this.
func CheckAvailability(
	blockedDates []availability.BlockedDate,
	openingHours []availability.OpeningHours,
	existingBookings []Booking,
	service services.Service,
	startTime time.Time,
	slotIntervalMinutes int,
) (*AvailabilityResult, error) {
	// Step 0: Validate slot alignment
	if slotIntervalMinutes > 0 {
		minuteOfDay := startTime.Hour()*60 + startTime.Minute()
		if minuteOfDay%slotIntervalMinutes != 0 || startTime.Second() != 0 {
			return &AvailabilityResult{Available: false, Reason: "invalid_slot_time"}, nil
		}
	}

	// Step 1: Check if date is blocked
	startDate := time.Date(startTime.Year(), startTime.Month(), startTime.Day(), 0, 0, 0, 0, startTime.Location())
	for _, bd := range blockedDates {
		blockedDay := time.Date(bd.Date.Year(), bd.Date.Month(), bd.Date.Day(), 0, 0, 0, 0, startTime.Location())
		if startDate.Equal(blockedDay) {
			return &AvailabilityResult{Available: false, Reason: "date_blocked"}, nil
		}
	}

	// Step 2: Calculate end time
	endTime := startTime.Add(time.Duration(service.DurationMinutes) * time.Minute)

	// Step 3: Check opening hours for the day
	dayOfWeek := int(startTime.Weekday())
	var dayHours *availability.OpeningHours
	for i := range openingHours {
		if openingHours[i].DayOfWeek == dayOfWeek {
			dayHours = &openingHours[i]
			break
		}
	}

	if dayHours == nil || !dayHours.Enabled {
		return &AvailabilityResult{Available: false, Reason: "outside_hours"}, nil
	}

	// Step 4: Verify time falls within opening hours
	openTime, err := parseTimeOnDate(dayHours.OpenTime, startTime)
	if err != nil {
		return nil, fmt.Errorf("parsing open time: %w", err)
	}
	closeTime, err := parseTimeOnDate(dayHours.CloseTime, startTime)
	if err != nil {
		return nil, fmt.Errorf("parsing close time: %w", err)
	}

	if startTime.Before(openTime) || endTime.After(closeTime) {
		return &AvailabilityResult{Available: false, Reason: "outside_hours"}, nil
	}

	// Step 5: Check for overlapping bookings (including buffers)
	effectiveStart := startTime.Add(-time.Duration(service.BufferBefore) * time.Minute)
	effectiveEnd := endTime.Add(time.Duration(service.BufferAfter) * time.Minute)

	for _, existing := range existingBookings {
		if existing.Status == StatusCancelled {
			continue
		}
		// Two intervals overlap if one starts before the other ends and vice versa
		if effectiveStart.Before(existing.EndTime) && effectiveEnd.After(existing.StartTime) {
			return &AvailabilityResult{Available: false, Reason: "slot_taken"}, nil
		}
	}

	return &AvailabilityResult{Available: true, EndTime: endTime}, nil
}

// parseTimeOnDate parses a time string like "09:00" and returns it as a time.Time
// on the same date as the reference time, in the same location.
func parseTimeOnDate(timeStr string, ref time.Time) (time.Time, error) {
	t, err := time.Parse("15:04", timeStr)
	if err != nil {
		return time.Time{}, err
	}
	return time.Date(
		ref.Year(), ref.Month(), ref.Day(),
		t.Hour(), t.Minute(), 0, 0,
		ref.Location(),
	), nil
}
