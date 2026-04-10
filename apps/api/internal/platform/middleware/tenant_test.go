package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestExtractTenantSlugFromHost(t *testing.T) {
	const baseDomain = "zenvikar.localhost"

	tests := []struct {
		name    string
		host    string
		want    string
		wantErr error
	}{
		{
			name: "simple subdomain",
			host: "acme.zenvikar.localhost",
			want: "acme",
		},
		{
			name: "subdomain with port",
			host: "acme.zenvikar.localhost:8080",
			want: "acme",
		},
		{
			name: "uppercase normalized to lowercase",
			host: "ACME.Zenvikar.Localhost",
			want: "acme",
		},
		{
			name: "mixed case with port",
			host: "Beta-Salon.ZENVIKAR.LOCALHOST:3000",
			want: "beta-salon",
		},
		{
			name: "base domain only returns ErrNoSubdomain",
			host: "zenvikar.localhost",
			want:    "",
			wantErr: ErrNoSubdomain,
		},
		{
			name:    "base domain with port returns ErrNoSubdomain",
			host:    "zenvikar.localhost:8080",
			want:    "",
			wantErr: ErrNoSubdomain,
		},
		{
			name:    "unrelated host returns ErrInvalidHost",
			host:    "example.com",
			want:    "",
			wantErr: ErrInvalidHost,
		},
		{
			name:    "nested subdomain returns ErrInvalidHost",
			host:    "a.b.zenvikar.localhost",
			want:    "",
			wantErr: ErrInvalidHost,
		},
		{
			name:    "deeply nested subdomain returns ErrInvalidHost",
			host:    "x.y.z.zenvikar.localhost",
			want:    "",
			wantErr: ErrInvalidHost,
		},
		{
			name: "hyphenated slug",
			host: "my-tenant.zenvikar.localhost",
			want: "my-tenant",
		},
		{
			name:    "completely different domain",
			host:    "other.domain.com",
			want:    "",
			wantErr: ErrInvalidHost,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExtractTenantSlugFromHost(tt.host, baseDomain)
			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTenantFromContext_NoContext(t *testing.T) {
	ctx := context.Background()
	tc := TenantFromContext(ctx)
	if tc != nil {
		t.Fatalf("expected nil TenantContext, got %+v", tc)
	}
}

func TestTenantResolutionMiddleware(t *testing.T) {
	tests := []struct {
		name     string
		header   string
		wantSlug string
		wantNil  bool
	}{
		{
			name:    "no header sets no context",
			header:  "",
			wantNil: true,
		},
		{
			name:     "header sets slug in context",
			header:   "acme",
			wantSlug: "acme",
		},
		{
			name:     "header is lowercased and trimmed",
			header:   "  ACME  ",
			wantSlug: "acme",
		},
		{
			name:     "hyphenated slug preserved",
			header:   "my-tenant",
			wantSlug: "my-tenant",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotTC *TenantContext

			handler := TenantResolution(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotTC = TenantFromContext(r.Context())
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.header != "" {
				req.Header.Set("X-Tenant-ID", tt.header)
			}
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if tt.wantNil {
				if gotTC != nil {
					t.Fatalf("expected nil TenantContext, got %+v", gotTC)
				}
				return
			}

			if gotTC == nil {
				t.Fatal("expected TenantContext, got nil")
			}
			if gotTC.Slug != tt.wantSlug {
				t.Errorf("got slug %q, want %q", gotTC.Slug, tt.wantSlug)
			}
		})
	}
}
