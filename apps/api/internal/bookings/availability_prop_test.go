package bookings

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"pgregory.net/rapid"

	"github.com/zenvikar/api/internal/availability"
	"github.com/zenvikar/api/internal/services"
)

// serviceGen generates a valid Service with positive duration and non-negative buffers.
func serviceGen() *rapid.Generator[services.Service] {
	return rapid.Custom(func(t *rapid.T) services.Service {
		return services.Service{
			ID:              uuid.New(),
			TenantID:        uuid.New(),
			Name:            "Test Service",
			DurationMinutes: rapid.IntRange(15, 120).Draw(t, "duration"),
			BufferBefore:    rapid.IntRange(0, 30).Draw(t, "bufferBefore"),
			BufferAfter:     rapid.IntRange(0, 30).Draw(t, "bufferAfter"),
			Enabled:         true,
		}
	})
}

// openingHoursForDay generates opening hours for a specific day of the week.
func openingHoursForDay(dayOfWeek int) *rapid.Generator[availability.OpeningHours] {
	return rapid.Custom(func(t *rapid.T) availability.OpeningHours {
		openHour := rapid.IntRange(6, 10).Draw(t, "openHour")
		closeHour := rapid.IntRange(17, 22).Draw(t, "closeHour")
		return availability.OpeningHours{
			ID:        uuid.New(),
			TenantID:  uuid.New(),
			DayOfWeek: dayOfWeek,
			OpenTime:  fmt.Sprintf("%02d:00", openHour),
			CloseTime: fmt.Sprintf("%02d:00", closeHour),
			Enabled:   true,
		}
	})
}

// baseDate returns a Monday (weekday 1) at midnight UTC for predictable testing.
func baseDate() time.Time {
	// 2025-01-06 is a Monday
	return time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC)
}

// TestProperty6_BlockedDateBookingRejection verifies that booking requests
// targeting blocked dates are always rejected with "date_blocked".
//
// **Validates: Requirement 7.1**
func TestProperty6_BlockedDateBookingRejection(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		svc := serviceGen().Draw(t, "service")
		base := baseDate()

		// Generate a blocked date
		dayOffset := rapid.IntRange(0, 30).Draw(t, "dayOffset")
		blockedDay := base.AddDate(0, 0, dayOffset)

		blockedDates := []availability.BlockedDate{
			{
				ID:       uuid.New(),
				TenantID: svc.TenantID,
				Date:     blockedDay,
			},
		}

		// Generate opening hours for the blocked day's weekday so it would
		// otherwise be valid
		dow := int(blockedDay.Weekday())
		hours := []availability.OpeningHours{
			{
				ID:        uuid.New(),
				TenantID:  svc.TenantID,
				DayOfWeek: dow,
				OpenTime:  "06:00",
				CloseTime: "23:00",
				Enabled:   true,
			},
		}

		// Request a booking on the blocked date
		startHour := rapid.IntRange(8, 16).Draw(t, "startHour")
		startTime := time.Date(
			blockedDay.Year(), blockedDay.Month(), blockedDay.Day(),
			startHour, 0, 0, 0, time.UTC,
		)

		result, err := CheckAvailability(blockedDates, hours, nil, svc, startTime)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Available {
			t.Fatalf("booking on blocked date should be rejected, got available=true")
		}
		if result.Reason != "date_blocked" {
			t.Fatalf("expected reason 'date_blocked', got %q", result.Reason)
		}
	})
}

// TestProperty7_OutsideHoursBookingRejection verifies that booking requests
// outside the tenant's opening hours are rejected with "outside_hours".
//
// **Validates: Requirement 7.2**
func TestProperty7_OutsideHoursBookingRejection(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		svc := serviceGen().Draw(t, "service")
		base := baseDate() // Monday

		// Opening hours: 09:00 - 17:00
		hours := []availability.OpeningHours{
			{
				ID:        uuid.New(),
				TenantID:  svc.TenantID,
				DayOfWeek: int(base.Weekday()), // Monday = 1
				OpenTime:  "09:00",
				CloseTime: "17:00",
				Enabled:   true,
			},
		}

		// Strategy: pick a start time that guarantees the booking window
		// falls outside opening hours. Either start before open or end after close.
		outsideCase := rapid.IntRange(0, 1).Draw(t, "outsideCase")

		var startTime time.Time
		if outsideCase == 0 {
			// Before opening: start at 00:00 - 08:59 (ensure end is also before open)
			// Use hours 0-7 so that even with max duration (120min), end <= 09:00
			h := rapid.IntRange(0, 6).Draw(t, "earlyHour")
			startTime = time.Date(base.Year(), base.Month(), base.Day(), h, 0, 0, 0, time.UTC)
		} else {
			// After closing: start late enough that end_time > 17:00
			// With min duration 15min, starting at 16:46+ guarantees end > 17:00
			// Use 17:00+ to be safe
			h := rapid.IntRange(17, 23).Draw(t, "lateHour")
			startTime = time.Date(base.Year(), base.Month(), base.Day(), h, 0, 0, 0, time.UTC)
		}

		result, err := CheckAvailability(nil, hours, nil, svc, startTime)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Available {
			t.Fatalf("booking outside hours should be rejected (case=%d, start=%v, duration=%d)",
				outsideCase, startTime, svc.DurationMinutes)
		}
		if result.Reason != "outside_hours" {
			t.Fatalf("expected reason 'outside_hours', got %q", result.Reason)
		}
	})
}

// TestProperty8_BookingOverlapRejectionWithBuffers verifies that booking requests
// overlapping existing bookings (including buffer times) are rejected with "slot_taken".
//
// **Validates: Requirements 7.3, 7.5**
func TestProperty8_BookingOverlapRejectionWithBuffers(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		svc := serviceGen().Draw(t, "service")
		base := baseDate() // Monday

		// Wide opening hours so we don't get outside_hours rejections
		hours := []availability.OpeningHours{
			{
				ID:        uuid.New(),
				TenantID:  svc.TenantID,
				DayOfWeek: int(base.Weekday()),
				OpenTime:  "00:00",
				CloseTime: "23:59",
				Enabled:   true,
			},
		}

		// Create an existing booking at 10:00
		existingStart := time.Date(base.Year(), base.Month(), base.Day(), 10, 0, 0, 0, time.UTC)
		existingEnd := existingStart.Add(time.Duration(svc.DurationMinutes) * time.Minute)

		existingBookings := []Booking{
			{
				ID:        uuid.New(),
				TenantID:  svc.TenantID,
				ServiceID: svc.ID,
				StartTime: existingStart,
				EndTime:   existingEnd,
				Status:    StatusConfirmed,
			},
		}

		// The effective window of the existing booking is:
		//   [existingStart, existingEnd]
		// The effective window of the new booking (with buffers) is:
		//   [newStart - bufferBefore, newEnd + bufferAfter]
		// For overlap, we need:
		//   newStart - bufferBefore < existingEnd AND newEnd + bufferAfter > existingStart

		// Generate a new start time that overlaps the existing booking's window
		// considering buffers. Pick a time within the effective range.
		effectiveExistingStart := existingStart
		effectiveExistingEnd := existingEnd

		// The new booking's effective window [newStart-bufBefore, newStart+duration+bufAfter]
		// must overlap [existingStart, existingEnd].
		// So: newStart - bufBefore < existingEnd AND newStart + duration + bufAfter > existingStart
		// => newStart < existingEnd + bufBefore AND newStart > existingStart - duration - bufAfter

		totalNewWindow := time.Duration(svc.DurationMinutes+svc.BufferAfter) * time.Minute
		earliestOverlap := effectiveExistingStart.Add(-totalNewWindow).Add(time.Minute) // +1min to ensure overlap
		latestOverlap := effectiveExistingEnd.Add(time.Duration(svc.BufferBefore) * time.Minute).Add(-time.Minute)

		if !earliestOverlap.Before(latestOverlap) {
			// Edge case: window too small, just use existing start
			earliestOverlap = existingStart
			latestOverlap = existingStart.Add(time.Minute)
		}

		// Pick a minute offset within the overlap range
		rangeMinutes := int(latestOverlap.Sub(earliestOverlap).Minutes())
		if rangeMinutes <= 0 {
			rangeMinutes = 1
		}
		offsetMin := rapid.IntRange(0, rangeMinutes).Draw(t, "overlapOffset")
		newStart := earliestOverlap.Add(time.Duration(offsetMin) * time.Minute)

		result, err := CheckAvailability(nil, hours, existingBookings, svc, newStart)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Available {
			newEnd := newStart.Add(time.Duration(svc.DurationMinutes) * time.Minute)
			effStart := newStart.Add(-time.Duration(svc.BufferBefore) * time.Minute)
			effEnd := newEnd.Add(time.Duration(svc.BufferAfter) * time.Minute)
			t.Fatalf("overlapping booking should be rejected:\n"+
				"  existing: [%v, %v]\n"+
				"  new effective: [%v, %v]\n"+
				"  buffers: before=%d, after=%d",
				existingStart, existingEnd, effStart, effEnd,
				svc.BufferBefore, svc.BufferAfter)
		}
		if result.Reason != "slot_taken" {
			t.Fatalf("expected reason 'slot_taken', got %q", result.Reason)
		}
	})
}

// TestProperty9_BookingEndTimeCalculation verifies that for valid booking requests,
// end_time equals start_time + service.DurationMinutes.
//
// **Validates: Requirement 7.4**
func TestProperty9_BookingEndTimeCalculation(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		svc := serviceGen().Draw(t, "service")
		base := baseDate() // Monday

		// Wide opening hours
		hours := []availability.OpeningHours{
			{
				ID:        uuid.New(),
				TenantID:  svc.TenantID,
				DayOfWeek: int(base.Weekday()),
				OpenTime:  "00:00",
				CloseTime: "23:59",
				Enabled:   true,
			},
		}

		// Pick a start time early enough that the booking fits within the day
		maxStartHour := 23 - (svc.DurationMinutes / 60) - 1
		if maxStartHour < 0 {
			maxStartHour = 0
		}
		if maxStartHour > 22 {
			maxStartHour = 22
		}
		startHour := rapid.IntRange(0, maxStartHour).Draw(t, "startHour")
		startMin := rapid.IntRange(0, 59).Draw(t, "startMin")

		// Ensure end time doesn't exceed 23:59
		startTime := time.Date(base.Year(), base.Month(), base.Day(), startHour, startMin, 0, 0, time.UTC)
		expectedEnd := startTime.Add(time.Duration(svc.DurationMinutes) * time.Minute)

		// Skip if end time would be past 23:59
		closeTime := time.Date(base.Year(), base.Month(), base.Day(), 23, 59, 0, 0, time.UTC)
		if expectedEnd.After(closeTime) {
			return
		}

		result, err := CheckAvailability(nil, hours, nil, svc, startTime)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.Available {
			t.Fatalf("expected available=true for valid slot, got reason=%q", result.Reason)
		}
		if !result.EndTime.Equal(expectedEnd) {
			t.Fatalf("end_time mismatch: got %v, want %v (start=%v, duration=%d)",
				result.EndTime, expectedEnd, startTime, svc.DurationMinutes)
		}
	})
}
