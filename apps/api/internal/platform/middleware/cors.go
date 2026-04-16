package middleware

import (
	"net/http"
	"strings"
)

// CORS returns middleware that sets CORS headers based on the provided allowed origins.
// If allowedOrigins contains "*", all origins are allowed.
// Preflight OPTIONS requests are handled and return immediately.
func CORS(allowedOrigins []string) func(next http.Handler) http.Handler {
	allowAll := false
	originSet := make(map[string]struct{}, len(allowedOrigins))
	for _, o := range allowedOrigins {
		if o == "*" {
			allowAll = true
		}
		originSet[strings.ToLower(o)] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			if allowAll {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			} else if isAllowedOrigin(origin, originSet) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Vary", "Origin")
			}

			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Tenant-ID")
			w.Header().Set("Access-Control-Max-Age", "86400")

			// Handle preflight
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func isAllowedOrigin(origin string, allowed map[string]struct{}) bool {
	normalized := strings.ToLower(strings.TrimSpace(origin))
	if normalized == "" {
		return false
	}
	if _, ok := allowed[normalized]; ok {
		return true
	}

	for pattern := range allowed {
		if strings.Contains(pattern, "://*.") && matchesWildcardOrigin(normalized, pattern) {
			return true
		}
	}

	return false
}

func matchesWildcardOrigin(origin, pattern string) bool {
	originScheme, originHost, ok := strings.Cut(origin, "://")
	if !ok {
		return false
	}
	patternScheme, patternHost, ok := strings.Cut(pattern, "://*.")
	if !ok || originScheme != patternScheme {
		return false
	}

	return strings.HasSuffix(originHost, "."+patternHost) && originHost != patternHost
}
