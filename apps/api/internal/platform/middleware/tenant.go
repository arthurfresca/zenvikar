package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/zenvikar/api/internal/tenants"
)

// ErrNoSubdomain is returned when the host matches the base domain exactly (no subdomain present).
var ErrNoSubdomain = errors.New("no subdomain present")

// ErrInvalidHost is returned when the host does not match the expected base domain
// or contains nested subdomains.
var ErrInvalidHost = errors.New("invalid host")

// tenantContextKey is the context key for storing tenant information.
type tenantContextKey struct{}

// TenantContext holds the resolved tenant information injected into request context.
type TenantContext struct {
	Slug   string
	Tenant *tenants.Tenant
}

// TenantFromContext retrieves the TenantContext from the request context.
// Returns nil if no tenant context is set.
func TenantFromContext(ctx context.Context) *TenantContext {
	tc, _ := ctx.Value(tenantContextKey{}).(*TenantContext)
	return tc
}

// TenantResolution is middleware that extracts the tenant slug from the X-Tenant-ID
// header and injects a TenantContext into the request context.
// The middleware stores the slug; actual tenant resolution is done by the handler.
func TenantResolution(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slug := r.Header.Get("X-Tenant-ID")
		if slug == "" {
			next.ServeHTTP(w, r)
			return
		}

		tc := &TenantContext{
			Slug: strings.ToLower(strings.TrimSpace(slug)),
		}
		ctx := context.WithValue(r.Context(), tenantContextKey{}, tc)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ExtractTenantSlugFromHost extracts the tenant slug from a hostname.
// It strips any port, normalizes to lowercase, verifies the host ends with
// ".{baseDomain}", and rejects nested subdomains.
func ExtractTenantSlugFromHost(host string, baseDomain string) (string, error) {
	// Step 1: Strip port if present
	hostname := host
	if idx := strings.LastIndex(host, ":"); idx != -1 {
		hostname = host[:idx]
	}
	hostname = strings.ToLower(hostname)

	// Step 2: Check if host ends with base domain
	if !strings.HasSuffix(hostname, "."+baseDomain) {
		if hostname == baseDomain {
			return "", ErrNoSubdomain
		}
		return "", ErrInvalidHost
	}

	// Step 3: Extract subdomain
	slug := strings.TrimSuffix(hostname, "."+baseDomain)

	// Step 4: Validate no nested subdomains
	if strings.Contains(slug, ".") {
		return "", ErrInvalidHost
	}

	return slug, nil
}
