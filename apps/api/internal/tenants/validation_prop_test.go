package tenants

import (
	"regexp"
	"strings"
	"testing"

	"pgregory.net/rapid"
)

// slugPatternProp is the reference regex for independent validation in tests.
var slugPatternProp = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*[a-z0-9]$`)

// expectedValid independently computes whether a slug should be accepted.
func expectedValid(s string) bool {
	n := len(s)
	if n < 3 || n > 63 {
		return false
	}
	if !slugPatternProp.MatchString(s) {
		return false
	}
	if strings.Contains(s, "--") {
		return false
	}
	if _, ok := ReservedSlugs[s]; ok {
		return false
	}
	return true
}

// TestPropertySlugValidationCorrectness verifies Property 1: Slug Validation Correctness.
// For any string input, ValidateTenantSlug accepts the string if and only if:
// length 3-63, matches ^[a-z0-9][a-z0-9-]*[a-z0-9]$, no consecutive hyphens, not reserved.
//
// **Validates: Requirements 2.1, 2.2, 2.3, 2.4**
func TestPropertySlugValidationCorrectness(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		s := rapid.String().Draw(t, "slug")

		err := ValidateTenantSlug(s)
		want := expectedValid(s)

		if want && err != nil {
			t.Fatalf("expected slug %q to be valid, got error: %v", s, err)
		}
		if !want && err == nil {
			t.Fatalf("expected slug %q to be invalid, got nil error", s)
		}
	})
}
