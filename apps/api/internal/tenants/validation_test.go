package tenants

import (
	"testing"
)

func TestValidateTenantSlug_Valid(t *testing.T) {
	valid := []string{
		"abc",                    // minimum length
		"a1b",                    // mixed alphanumeric
		"my-tenant",              // typical slug
		"tenant-123",             // trailing digits
		"a-b-c",                  // multiple hyphens (non-consecutive)
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", // 63 chars
	}
	for _, slug := range valid {
		if err := ValidateTenantSlug(slug); err != nil {
			t.Errorf("ValidateTenantSlug(%q) = %v, want nil", slug, err)
		}
	}
}

func TestValidateTenantSlug_TooShort(t *testing.T) {
	for _, slug := range []string{"", "a", "ab"} {
		if err := ValidateTenantSlug(slug); err == nil {
			t.Errorf("ValidateTenantSlug(%q) = nil, want error for too short", slug)
		}
	}
}

func TestValidateTenantSlug_TooLong(t *testing.T) {
	slug := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" // 64 chars
	if err := ValidateTenantSlug(slug); err == nil {
		t.Errorf("ValidateTenantSlug(%q) = nil, want error for too long", slug)
	}
}

func TestValidateTenantSlug_InvalidPattern(t *testing.T) {
	invalid := []string{
		"-abc",     // starts with hyphen
		"abc-",     // ends with hyphen
		"ABC",      // uppercase
		"my tenant", // space
		"my_tenant", // underscore
		"my.tenant", // dot
	}
	for _, slug := range invalid {
		if err := ValidateTenantSlug(slug); err == nil {
			t.Errorf("ValidateTenantSlug(%q) = nil, want error for invalid pattern", slug)
		}
	}
}

func TestValidateTenantSlug_ConsecutiveHyphens(t *testing.T) {
	if err := ValidateTenantSlug("my--tenant"); err == nil {
		t.Error("ValidateTenantSlug(\"my--tenant\") = nil, want error for consecutive hyphens")
	}
}

func TestValidateTenantSlug_Reserved(t *testing.T) {
	for slug := range ReservedSlugs {
		if len(slug) >= 3 {
			if err := ValidateTenantSlug(slug); err == nil {
				t.Errorf("ValidateTenantSlug(%q) = nil, want error for reserved slug", slug)
			}
		}
	}
}
