package users

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	errInvalidToken = errors.New("invalid token")
	errTokenExpired = errors.New("token expired")
)

type tokenClaims struct {
	Issuer       string            `json:"iss"`
	Subject      string            `json:"sub"`
	Audience     string            `json:"aud"`
	IssuedAt     int64             `json:"iat"`
	ExpiresAt    int64             `json:"exp"`
	Email        string            `json:"email"`
	Name         string            `json:"name"`
	PlatformRole string            `json:"platformRole,omitempty"`
	TenantRoles  map[string]string `json:"tenantRoles,omitempty"`
}

// TokenClaims exposes parsed auth claims outside the users package.
type TokenClaims = tokenClaims

// TokenManager creates and validates JWT-like HS256 tokens.
type TokenManager struct {
	secret []byte
	ttl    time.Duration
}

// NewTokenManager initializes a token manager.
func NewTokenManager(secret string, ttlMinutes int) *TokenManager {
	if ttlMinutes <= 0 {
		ttlMinutes = 120
	}
	return &TokenManager{
		secret: []byte(secret),
		ttl:    time.Duration(ttlMinutes) * time.Minute,
	}
}

// IssueToken creates a signed auth token and returns token + expiry.
func (tm *TokenManager) IssueToken(user *User, audience, platformRole string, tenantRoles map[string]string) (string, time.Time, error) {
	now := time.Now().UTC()
	exp := now.Add(tm.ttl)

	claims := tokenClaims{
		Issuer:       "zenvikar-api",
		Subject:      user.ID.String(),
		Audience:     audience,
		IssuedAt:     now.Unix(),
		ExpiresAt:    exp.Unix(),
		Email:        user.Email,
		Name:         user.Name,
		PlatformRole: platformRole,
		TenantRoles:  tenantRoles,
	}

	token, err := tm.sign(claims)
	if err != nil {
		return "", time.Time{}, err
	}

	return token, exp, nil
}

// ParseToken validates and decodes a token.
func (tm *TokenManager) ParseToken(token string) (*tokenClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errInvalidToken
	}

	signingInput := parts[0] + "." + parts[1]
	sig, err := tm.signRaw(signingInput)
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

	var claims tokenClaims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return nil, errInvalidToken
	}

	if time.Now().UTC().Unix() >= claims.ExpiresAt {
		return nil, errTokenExpired
	}

	return &claims, nil
}

func (tm *TokenManager) sign(claims tokenClaims) (string, error) {
	header := map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	}

	hb, err := json.Marshal(header)
	if err != nil {
		return "", fmt.Errorf("encoding token header: %w", err)
	}
	cb, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("encoding token claims: %w", err)
	}

	headerSeg := base64.RawURLEncoding.EncodeToString(hb)
	claimsSeg := base64.RawURLEncoding.EncodeToString(cb)
	signingInput := headerSeg + "." + claimsSeg

	sig, err := tm.signRaw(signingInput)
	if err != nil {
		return "", err
	}

	return signingInput + "." + sig, nil
}

func (tm *TokenManager) signRaw(input string) (string, error) {
	mac := hmac.New(sha256.New, tm.secret)
	if _, err := mac.Write([]byte(input)); err != nil {
		return "", fmt.Errorf("signing token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil)), nil
}
