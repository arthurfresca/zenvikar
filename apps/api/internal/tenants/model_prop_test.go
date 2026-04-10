package tenants

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"pgregory.net/rapid"
)

// hexColorGen generates valid hex color codes like "#1a2b3c".
func hexColorGen() *rapid.Generator[string] {
	return rapid.Custom(func(t *rapid.T) string {
		r := rapid.IntRange(0, 255).Draw(t, "r")
		g := rapid.IntRange(0, 255).Draw(t, "g")
		b := rapid.IntRange(0, 255).Draw(t, "b")
		return fmt.Sprintf("#%02x%02x%02x", r, g, b)
	})
}

// tenantSlugGen generates valid tenant slugs: 3-63 chars, lowercase alphanumeric + hyphens,
// starts/ends with alphanumeric, no consecutive hyphens.
func tenantSlugGen() *rapid.Generator[string] {
	return rapid.Custom(func(t *rapid.T) string {
		length := rapid.IntRange(3, 63).Draw(t, "slugLen")
		alnum := "abcdefghijklmnopqrstuvwxyz0123456789"
		alnumHyphen := "abcdefghijklmnopqrstuvwxyz0123456789-"

		chars := make([]byte, length)
		chars[0] = alnum[rapid.IntRange(0, len(alnum)-1).Draw(t, "first")]
		chars[length-1] = alnum[rapid.IntRange(0, len(alnum)-1).Draw(t, "last")]

		for i := 1; i < length-1; i++ {
			if chars[i-1] == '-' {
				chars[i] = alnum[rapid.IntRange(0, len(alnum)-1).Draw(t, fmt.Sprintf("mid%d", i))]
			} else {
				chars[i] = alnumHyphen[rapid.IntRange(0, len(alnumHyphen)-1).Draw(t, fmt.Sprintf("mid%d", i))]
			}
		}
		return string(chars)
	})
}

// optionalLogoURLGen generates nil or a non-empty string pointer for LogoURL.
func optionalLogoURLGen() *rapid.Generator[*string] {
	return rapid.Custom(func(t *rapid.T) *string {
		isNil := rapid.Bool().Draw(t, "logoNil")
		if isNil {
			return nil
		}
		s := rapid.StringMatching(`https://[a-z]{3,10}\.[a-z]{2,4}/[a-z0-9]{1,20}\.(png|jpg|svg)`).Draw(t, "logoURL")
		return &s
	})
}

// tenantGen generates valid Tenant objects for property testing.
func tenantGen() *rapid.Generator[Tenant] {
	return rapid.Custom(func(t *rapid.T) Tenant {
		id := uuid.New()
		slug := tenantSlugGen().Draw(t, "slug")
		displayName := rapid.StringMatching(`[A-Za-z][A-Za-z0-9 ]{0,49}`).Draw(t, "displayName")
		logoURL := optionalLogoURLGen().Draw(t, "logoURL")
		colorPrimary := hexColorGen().Draw(t, "colorPrimary")
		colorSecondary := hexColorGen().Draw(t, "colorSecondary")
		colorAccent := hexColorGen().Draw(t, "colorAccent")
		timezone := rapid.SampledFrom([]string{
			"UTC", "America/New_York", "America/Sao_Paulo",
			"Europe/London", "Europe/Lisbon", "Asia/Tokyo",
		}).Draw(t, "timezone")
		locale := rapid.SampledFrom([]string{"en", "pt"}).Draw(t, "locale")
		enabled := rapid.Bool().Draw(t, "enabled")

		// Truncate to seconds to avoid sub-second precision issues with JSON round-trip.
		// JSON marshals time.Time with nanosecond precision in RFC3339Nano, but
		// we want stable round-trips so we truncate to second precision.
		now := time.Now().UTC().Truncate(time.Second)
		createdOffset := rapid.IntRange(0, 86400*365).Draw(t, "createdOffset")
		createdAt := now.Add(-time.Duration(createdOffset) * time.Second)
		updatedOffset := rapid.IntRange(0, createdOffset).Draw(t, "updatedOffset")
		updatedAt := now.Add(-time.Duration(updatedOffset) * time.Second)

		return Tenant{
			ID:             id,
			Slug:           slug,
			DisplayName:    displayName,
			LogoURL:        logoURL,
			ColorPrimary:   colorPrimary,
			ColorSecondary: colorSecondary,
			ColorAccent:    colorAccent,
			Timezone:       timezone,
			DefaultLocale:  locale,
			Enabled:        enabled,
			CreatedAt:      createdAt,
			UpdatedAt:      updatedAt,
		}
	})
}

// TestPropertyTenantJSONSerializationRoundTrip verifies Property 4: Tenant JSON Serialization Round-Trip.
// For any valid Tenant object, serializing it to JSON and deserializing it back SHALL produce
// an object equal to the original. This ensures cache consistency between Redis and database results.
//
// **Validates: Requirements 4.6**
func TestPropertyTenantJSONSerializationRoundTrip(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		original := tenantGen().Draw(t, "tenant")

		// Serialize to JSON
		data, err := json.Marshal(original)
		if err != nil {
			t.Fatalf("failed to marshal tenant: %v", err)
		}

		// Deserialize back
		var restored Tenant
		if err := json.Unmarshal(data, &restored); err != nil {
			t.Fatalf("failed to unmarshal tenant: %v", err)
		}

		// Compare all fields
		if original.ID != restored.ID {
			t.Fatalf("ID mismatch: %v != %v", original.ID, restored.ID)
		}
		if original.Slug != restored.Slug {
			t.Fatalf("Slug mismatch: %q != %q", original.Slug, restored.Slug)
		}
		if original.DisplayName != restored.DisplayName {
			t.Fatalf("DisplayName mismatch: %q != %q", original.DisplayName, restored.DisplayName)
		}
		// LogoURL: compare pointer values
		if (original.LogoURL == nil) != (restored.LogoURL == nil) {
			t.Fatalf("LogoURL nil mismatch: original=%v, restored=%v", original.LogoURL, restored.LogoURL)
		}
		if original.LogoURL != nil && *original.LogoURL != *restored.LogoURL {
			t.Fatalf("LogoURL mismatch: %q != %q", *original.LogoURL, *restored.LogoURL)
		}
		if original.ColorPrimary != restored.ColorPrimary {
			t.Fatalf("ColorPrimary mismatch: %q != %q", original.ColorPrimary, restored.ColorPrimary)
		}
		if original.ColorSecondary != restored.ColorSecondary {
			t.Fatalf("ColorSecondary mismatch: %q != %q", original.ColorSecondary, restored.ColorSecondary)
		}
		if original.ColorAccent != restored.ColorAccent {
			t.Fatalf("ColorAccent mismatch: %q != %q", original.ColorAccent, restored.ColorAccent)
		}
		if original.Timezone != restored.Timezone {
			t.Fatalf("Timezone mismatch: %q != %q", original.Timezone, restored.Timezone)
		}
		if original.DefaultLocale != restored.DefaultLocale {
			t.Fatalf("DefaultLocale mismatch: %q != %q", original.DefaultLocale, restored.DefaultLocale)
		}
		if original.Enabled != restored.Enabled {
			t.Fatalf("Enabled mismatch: %v != %v", original.Enabled, restored.Enabled)
		}
		// time.Time comparison using Equal() to handle timezone/monotonic clock differences
		if !original.CreatedAt.Equal(restored.CreatedAt) {
			t.Fatalf("CreatedAt mismatch: %v != %v", original.CreatedAt, restored.CreatedAt)
		}
		if !original.UpdatedAt.Equal(restored.UpdatedAt) {
			t.Fatalf("UpdatedAt mismatch: %v != %v", original.UpdatedAt, restored.UpdatedAt)
		}
	})
}
