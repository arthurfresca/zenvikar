package tenants

import (
	"fmt"
	"regexp"
	"strings"
)

// ReservedSlugs is the set of slugs that cannot be used as tenant slugs.
var ReservedSlugs = map[string]struct{}{
	"www":      {},
	"api":      {},
	"admin":    {},
	"manage":   {},
	"app":      {},
	"mail":     {},
	"smtp":     {},
	"ftp":      {},
	"ssh":      {},
	"git":      {},
	"cdn":      {},
	"static":   {},
	"assets":   {},
	"media":    {},
	"blog":     {},
	"docs":     {},
	"help":     {},
	"support":  {},
	"status":   {},
	"billing":  {},
	"zenvikar": {},
}

// slugPattern matches valid tenant slugs: starts and ends with alphanumeric,
// middle characters can include hyphens.
var slugPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*[a-z0-9]$`)

// ValidateTenantSlug validates a tenant slug against all rules.
// Returns nil if valid, or a descriptive error for the first failing rule.
func ValidateTenantSlug(slug string) error {
	n := len(slug)

	if n < 3 || n > 63 {
		return fmt.Errorf("slug must be between 3 and 63 characters, got %d", n)
	}

	if !slugPattern.MatchString(slug) {
		return fmt.Errorf("slug must start and end with a lowercase letter or digit and contain only lowercase letters, digits, and hyphens")
	}

	if strings.Contains(slug, "--") {
		return fmt.Errorf("slug must not contain consecutive hyphens")
	}

	if _, ok := ReservedSlugs[slug]; ok {
		return fmt.Errorf("slug %q is reserved and cannot be used", slug)
	}

	return nil
}
