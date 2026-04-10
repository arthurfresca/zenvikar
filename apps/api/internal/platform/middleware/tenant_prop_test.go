package middleware

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"pgregory.net/rapid"
)

// validSlugGen generates valid tenant slugs: lowercase alphanumeric + hyphens,
// 3-63 chars, starts/ends with alphanumeric, no consecutive hyphens.
func validSlugGen() *rapid.Generator[string] {
	return rapid.Custom(func(t *rapid.T) string {
		// Length between 3 and 63
		length := rapid.IntRange(3, 63).Draw(t, "slugLen")

		chars := make([]byte, length)
		alnum := "abcdefghijklmnopqrstuvwxyz0123456789"
		alnumHyphen := "abcdefghijklmnopqrstuvwxyz0123456789-"

		// First char: alphanumeric only
		chars[0] = alnum[rapid.IntRange(0, len(alnum)-1).Draw(t, "first")]
		// Last char: alphanumeric only
		chars[length-1] = alnum[rapid.IntRange(0, len(alnum)-1).Draw(t, "last")]

		// Middle chars: alphanumeric + hyphen, but no consecutive hyphens
		for i := 1; i < length-1; i++ {
			if chars[i-1] == '-' {
				// Previous was hyphen, this must be alphanumeric
				chars[i] = alnum[rapid.IntRange(0, len(alnum)-1).Draw(t, fmt.Sprintf("mid%d", i))]
			} else {
				chars[i] = alnumHyphen[rapid.IntRange(0, len(alnumHyphen)-1).Draw(t, fmt.Sprintf("mid%d", i))]
			}
		}

		return string(chars)
	})
}

// baseDomainGen generates realistic base domains like "zenvikar.localhost" or "example.com".
func baseDomainGen() *rapid.Generator[string] {
	return rapid.Custom(func(t *rapid.T) string {
		label := rapid.StringMatching(`[a-z]{3,12}`).Draw(t, "domainLabel")
		tld := rapid.SampledFrom([]string{"localhost", "com", "io", "dev", "net"}).Draw(t, "tld")
		return label + "." + tld
	})
}

// portGen generates an optional port suffix (empty string or ":NNNN").
func portGen() *rapid.Generator[string] {
	return rapid.Custom(func(t *rapid.T) string {
		hasPort := rapid.Bool().Draw(t, "hasPort")
		if !hasPort {
			return ""
		}
		port := rapid.IntRange(1, 65535).Draw(t, "port")
		return fmt.Sprintf(":%d", port)
	})
}

// TestPropertySubdomainExtractionRoundTrip verifies Property 2: Subdomain Extraction Round-Trip.
// For any valid tenant slug and any base domain, constructing a host as `{slug}.{baseDomain}`
// (with or without a port suffix) and then calling ExtractTenantSlugFromHost SHALL return
// the original slug.
//
// **Validates: Requirements 3.1, 3.4**
func TestPropertySubdomainExtractionRoundTrip(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		slug := validSlugGen().Draw(t, "slug")
		baseDomain := baseDomainGen().Draw(t, "baseDomain")
		port := portGen().Draw(t, "port")

		host := slug + "." + baseDomain + port

		got, err := ExtractTenantSlugFromHost(host, baseDomain)
		if err != nil {
			t.Fatalf("expected no error for host %q (slug=%q, baseDomain=%q), got: %v",
				host, slug, baseDomain, err)
		}
		if got != slug {
			t.Fatalf("expected slug %q, got %q for host %q", slug, got, host)
		}
	})
}

// TestPropertyInvalidHostRejection verifies Property 3: Invalid Host Rejection.
// For any hostname that does not end with `.{baseDomain}` and is not equal to `baseDomain`,
// the Slug_Extractor SHALL return ErrInvalidHost.
//
// **Validates: Requirement 3.3**
func TestPropertyInvalidHostRejection(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		baseDomain := baseDomainGen().Draw(t, "baseDomain")

		// Generate a hostname that does NOT end with ".{baseDomain}" and is NOT equal to baseDomain.
		// Strategy: generate a random domain label + a different TLD/suffix.
		hostname := rapid.Custom(func(t *rapid.T) string {
			for {
				label := rapid.StringMatching(`[a-z0-9][a-z0-9.-]{1,30}[a-z0-9]`).Draw(t, "hostLabel")
				tld := rapid.SampledFrom([]string{"org", "xyz", "test", "example", "invalid"}).Draw(t, "hostTld")
				h := label + "." + tld

				// Ensure it doesn't accidentally match the base domain pattern
				if !strings.HasSuffix(h, "."+baseDomain) && h != baseDomain {
					// Optionally add a port
					port := portGen().Draw(t, "port")
					return h + port
				}
			}
		}).Draw(t, "hostname")

		_, err := ExtractTenantSlugFromHost(hostname, baseDomain)
		if err == nil {
			t.Fatalf("expected error for host %q with baseDomain %q, got nil", hostname, baseDomain)
		}
		if !errors.Is(err, ErrInvalidHost) {
			t.Fatalf("expected ErrInvalidHost for host %q with baseDomain %q, got: %v",
				hostname, baseDomain, err)
		}
	})
}
