package authn

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

type claimsContextKey struct{}

// Claims represents the authenticated JWT payload used by the API.
type Claims struct {
	Issuer            string            `json:"iss"`
	Subject           string            `json:"sub"`
	Audience          string            `json:"aud"`
	IssuedAt          int64             `json:"iat"`
	ExpiresAt         int64             `json:"exp"`
	Email             string            `json:"email"`
	Name              string            `json:"name"`
	PlatformRole      string            `json:"platformRole,omitempty"`
	TenantRoles       map[string]string `json:"tenantRoles,omitempty"`
	CurrentTenantID   string            `json:"currentTenantId,omitempty"`
	CurrentTenantSlug string            `json:"currentTenantSlug,omitempty"`
}

// ClaimsFromContext returns authenticated user claims when present.
func ClaimsFromContext(ctx context.Context) *Claims {
	claims, _ := ctx.Value(claimsContextKey{}).(*Claims)
	return claims
}

// UserIDFromContext returns the authenticated user ID when present and valid.
func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	claims := ClaimsFromContext(ctx)
	if claims == nil {
		return uuid.UUID{}, false
	}
	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.UUID{}, false
	}
	return userID, true
}

// RequireAuth parses a bearer token and stores claims in request context.
func RequireAuth(secret string, ttlMinutes int) func(http.Handler) http.Handler {
	_ = ttlMinutes
	secretBytes := []byte(secret)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := bearerToken(r.Header.Get("Authorization"))
			if token == "" {
				writeAuthError(w, "missing bearer token")
				return
			}
			claims, err := parseToken(token, secretBytes)
			if err != nil {
				writeAuthError(w, "invalid or expired token")
				return
			}
			ctx := context.WithValue(r.Context(), claimsContextKey{}, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func parseToken(token string, secret []byte) (*Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errInvalidToken
	}
	signingInput := parts[0] + "." + parts[1]
	sig, err := signRaw(signingInput, secret)
	if err != nil {
		return nil, err
	}
	if !hmac.Equal([]byte(sig), []byte(parts[2])) {
		return nil, errInvalidToken
	}
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, errInvalidToken
	}
	var claims Claims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return nil, errInvalidToken
	}
	if time.Now().UTC().Unix() >= claims.ExpiresAt {
		return nil, errInvalidToken
	}
	return &claims, nil
}

func signRaw(input string, secret []byte) (string, error) {
	mac := hmac.New(sha256.New, secret)
	if _, err := mac.Write([]byte(input)); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil)), nil
}

var errInvalidToken = http.ErrNoCookie

func bearerToken(authHeader string) string {
	parts := strings.SplitN(strings.TrimSpace(authHeader), " ", 2)
	if len(parts) != 2 {
		return ""
	}
	if !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func writeAuthError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error":   "unauthorized",
		"message": message,
	})
}
