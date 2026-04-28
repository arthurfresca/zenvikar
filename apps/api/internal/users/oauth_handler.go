package users

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	appmiddleware "github.com/zenvikar/api/internal/platform/middleware"
	"github.com/zenvikar/api/internal/tenants"
)

const (
	oauthStateTTL             = 10 * time.Minute
	oauthAudienceBookingWeb   = "booking-web"
	oauthAudienceTenantWeb    = "tenant-web"
	oauthStateErrorKey        = "error"
	oauthStateErrorBadRequest = "invalid_request"
)

type oauthStatePayload struct {
	Provider   string `json:"provider"`
	Redirect   string `json:"redirect"`
	Audience   string `json:"audience"`
	TenantSlug string `json:"tenantSlug,omitempty"`
	IssuedAt   int64  `json:"iat"`
}

func (h *authHandler) googleOAuthStart(w http.ResponseWriter, r *http.Request) {
	h.oauthStart(w, r, AuthProviderGoogle)
}

func (h *authHandler) facebookOAuthStart(w http.ResponseWriter, r *http.Request) {
	h.oauthStart(w, r, AuthProviderFacebook)
}

func (h *authHandler) googleOAuthCallback(w http.ResponseWriter, r *http.Request) {
	h.oauthCallback(w, r, AuthProviderGoogle)
}

func (h *authHandler) facebookOAuthCallback(w http.ResponseWriter, r *http.Request) {
	h.oauthCallback(w, r, AuthProviderFacebook)
}

func (h *authHandler) oauthStart(w http.ResponseWriter, r *http.Request, provider string) {
	redirectTo := strings.TrimSpace(r.URL.Query().Get("redirect"))
	if redirectTo == "" {
		redirectTo = fallbackOAuthRedirectFromReferer(r.Referer())
	}
	if !isAllowedOAuthRedirect(redirectTo, h.cfg.BaseDomain) {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error":   oauthStateErrorBadRequest,
			"message": "redirect must be an allowed zenvikar URL",
		})
		return
	}

	audience := normalizeOAuthAudience(r.URL.Query().Get("audience"))
	if audience == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error":   oauthStateErrorBadRequest,
			"message": "audience is not supported",
		})
		return
	}

	tenantSlug := oauthTenantSlug(redirectTo, h.cfg.BaseDomain)
	if tenantSlug == "" && audience == oauthAudienceTenantWeb {
		tenantSlug = strings.ToLower(strings.TrimSpace(r.URL.Query().Get("tenantSlug")))
	}
	state, err := h.encodeOAuthState(oauthStatePayload{
		Provider:   provider,
		Redirect:   redirectTo,
		Audience:   audience,
		TenantSlug: tenantSlug,
		IssuedAt:   time.Now().UTC().Unix(),
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error":   "internal_error",
			"message": "could not create oauth state",
		})
		return
	}

	var authURL string
	switch provider {
	case AuthProviderGoogle:
		authURL, err = h.buildGoogleAuthorizeURL(state)
	case AuthProviderFacebook:
		authURL, err = h.buildFacebookAuthorizeURL(state)
	default:
		err = ErrUnsupportedProvider
	}
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error":   oauthStateErrorBadRequest,
			"message": err.Error(),
		})
		return
	}

	http.Redirect(w, r, authURL, http.StatusFound)
}

func (h *authHandler) oauthCallback(w http.ResponseWriter, r *http.Request, provider string) {
	state, err := h.decodeOAuthState(r.URL.Query().Get("state"))
	if err != nil || state.Provider != provider || !isAllowedOAuthRedirect(state.Redirect, h.cfg.BaseDomain) {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error":   oauthStateErrorBadRequest,
			"message": "invalid oauth state",
		})
		return
	}

	if providerErr := strings.TrimSpace(r.URL.Query().Get("error")); providerErr != "" {
		http.Redirect(w, r, appendQueryString(state.Redirect, map[string]string{
			oauthStateErrorKey: "social_auth_denied",
		}), http.StatusFound)
		return
	}

	code := strings.TrimSpace(r.URL.Query().Get("code"))
	if code == "" {
		http.Redirect(w, r, appendQueryString(state.Redirect, map[string]string{
			oauthStateErrorKey: "missing_oauth_code",
		}), http.StatusFound)
		return
	}

	var input SocialLoginInput
	switch provider {
	case AuthProviderGoogle:
		tokens, err := h.exchangeGoogleCode(r.Context(), code)
		if err != nil {
			http.Redirect(w, r, appendQueryString(state.Redirect, map[string]string{
				oauthStateErrorKey: "google_exchange_failed",
			}), http.StatusFound)
			return
		}
		input = SocialLoginInput{
			Provider:          AuthProviderGoogle,
			GoogleIDToken:     tokens.IDToken,
			GoogleAccessToken: tokens.AccessToken,
		}
	case AuthProviderFacebook:
		accessToken, err := h.exchangeFacebookCode(r.Context(), code)
		if err != nil {
			http.Redirect(w, r, appendQueryString(state.Redirect, map[string]string{
				oauthStateErrorKey: "facebook_exchange_failed",
			}), http.StatusFound)
			return
		}
		input = SocialLoginInput{
			Provider:            AuthProviderFacebook,
			FacebookAccessToken: accessToken,
		}
	default:
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error":   oauthStateErrorBadRequest,
			"message": "provider is not supported",
		})
		return
	}

	result, err := h.oauthLoginResult(r.Context(), input, state)
	if err != nil {
		http.Redirect(w, r, appendQueryString(state.Redirect, map[string]string{
			oauthStateErrorKey: "social_login_failed",
		}), http.StatusFound)
		return
	}

	http.Redirect(w, r, appendQueryString(state.Redirect, map[string]string{
		"authToken":     result.Token,
		"authExpiresAt": result.ExpiresAt.UTC().Format(time.RFC3339Nano),
	}), http.StatusFound)
}

func (h *authHandler) oauthLoginResult(ctx context.Context, input SocialLoginInput, state *oauthStatePayload) (*AuthResult, error) {
	if strings.TrimSpace(state.TenantSlug) == "" {
		return h.svc.LoginSocial(ctx, input, state.Audience)
	}
	if state.Audience == oauthAudienceTenantWeb {
		return h.svc.LoginSocialTenant(ctx, input, state.TenantSlug)
	}
	return h.svc.LoginSocialBooking(ctx, input, state.TenantSlug)
}

func oauthTenantSlug(redirectTo, baseDomain string) string {
	parsed, err := url.Parse(strings.TrimSpace(redirectTo))
	if err != nil {
		return ""
	}
	host := parsed.Hostname()
	if host == "" {
		return ""
	}
	if slug, err := appmiddleware.ExtractTenantSlugFromHost(host, baseDomain); err == nil {
		normalized := strings.ToLower(strings.TrimSpace(slug))
		if _, reserved := tenants.ReservedSlugs[normalized]; !reserved {
			return normalized
		}
	}
	return ""
}

type googleTokenExchangeResponse struct {
	AccessToken string `json:"access_token"`
	IDToken     string `json:"id_token"`
}

func (h *authHandler) exchangeGoogleCode(ctx context.Context, code string) (*googleTokenExchangeResponse, error) {
	if strings.TrimSpace(h.cfg.GoogleClientID) == "" || strings.TrimSpace(h.cfg.GoogleClientSecret) == "" {
		return nil, fmt.Errorf("%w: google oauth is not configured", ErrInvalidInput)
	}

	form := url.Values{}
	form.Set("code", code)
	form.Set("client_id", h.cfg.GoogleClientID)
	form.Set("client_secret", h.cfg.GoogleClientSecret)
	form.Set("redirect_uri", h.googleCallbackURL())
	form.Set("grant_type", "authorization_code")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://oauth2.googleapis.com/token", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := oauthHTTPClient().Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: google oauth exchange rejected", ErrSocialValidationFailed)
	}

	var payload googleTokenExchangeResponse
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		return nil, err
	}
	if strings.TrimSpace(payload.IDToken) == "" && strings.TrimSpace(payload.AccessToken) == "" {
		return nil, fmt.Errorf("%w: google oauth missing tokens", ErrSocialValidationFailed)
	}

	return &payload, nil
}

func (h *authHandler) exchangeFacebookCode(ctx context.Context, code string) (string, error) {
	if strings.TrimSpace(h.cfg.FacebookAppID) == "" || strings.TrimSpace(h.cfg.FacebookAppSecret) == "" {
		return "", fmt.Errorf("%w: facebook oauth is not configured", ErrInvalidInput)
	}

	endpoint := "https://graph.facebook.com/v22.0/oauth/access_token"
	values := url.Values{}
	values.Set("client_id", h.cfg.FacebookAppID)
	values.Set("client_secret", h.cfg.FacebookAppSecret)
	values.Set("redirect_uri", h.facebookCallbackURL())
	values.Set("code", code)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"?"+values.Encode(), nil)
	if err != nil {
		return "", err
	}

	res, err := oauthHTTPClient().Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%w: facebook oauth exchange rejected", ErrSocialValidationFailed)
	}

	var payload struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		return "", err
	}
	if strings.TrimSpace(payload.AccessToken) == "" {
		return "", fmt.Errorf("%w: facebook oauth missing access token", ErrSocialValidationFailed)
	}

	return payload.AccessToken, nil
}

func (h *authHandler) buildGoogleAuthorizeURL(state string) (string, error) {
	if strings.TrimSpace(h.cfg.GoogleClientID) == "" {
		return "", fmt.Errorf("%w: google oauth is not configured", ErrInvalidInput)
	}
	values := url.Values{}
	values.Set("client_id", h.cfg.GoogleClientID)
	values.Set("redirect_uri", h.googleCallbackURL())
	values.Set("response_type", "code")
	values.Set("scope", "openid email profile")
	values.Set("state", state)
	values.Set("access_type", "online")
	values.Set("prompt", "select_account")

	return "https://accounts.google.com/o/oauth2/v2/auth?" + values.Encode(), nil
}

func (h *authHandler) buildFacebookAuthorizeURL(state string) (string, error) {
	if strings.TrimSpace(h.cfg.FacebookAppID) == "" {
		return "", fmt.Errorf("%w: facebook oauth is not configured", ErrInvalidInput)
	}
	values := url.Values{}
	values.Set("client_id", h.cfg.FacebookAppID)
	values.Set("redirect_uri", h.facebookCallbackURL())
	values.Set("state", state)
	values.Set("scope", "email,public_profile")
	values.Set("response_type", "code")

	return "https://www.facebook.com/v22.0/dialog/oauth?" + values.Encode(), nil
}

func (h *authHandler) googleCallbackURL() string {
	return strings.TrimRight(strings.TrimSpace(h.cfg.APIPublicURL), "/") + "/api/v1/auth/oauth/google/callback"
}

func (h *authHandler) facebookCallbackURL() string {
	return strings.TrimRight(strings.TrimSpace(h.cfg.APIPublicURL), "/") + "/api/v1/auth/oauth/facebook/callback"
}

func (h *authHandler) encodeOAuthState(payload oauthStatePayload) (string, error) {
	serialized, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	mac := hmac.New(sha256.New, []byte(h.cfg.JWTSecret))
	mac.Write(serialized)
	signature := mac.Sum(nil)

	state := base64.RawURLEncoding.EncodeToString(serialized) + "." + base64.RawURLEncoding.EncodeToString(signature)
	return state, nil
}

func (h *authHandler) decodeOAuthState(raw string) (*oauthStatePayload, error) {
	parts := strings.Split(strings.TrimSpace(raw), ".")
	if len(parts) != 2 {
		return nil, fmt.Errorf("%w: malformed state", ErrInvalidInput)
	}

	payloadRaw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("%w: malformed payload", ErrInvalidInput)
	}
	signatureRaw, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("%w: malformed signature", ErrInvalidInput)
	}

	mac := hmac.New(sha256.New, []byte(h.cfg.JWTSecret))
	mac.Write(payloadRaw)
	expected := mac.Sum(nil)
	if subtle.ConstantTimeCompare(signatureRaw, expected) != 1 {
		return nil, fmt.Errorf("%w: invalid signature", ErrInvalidInput)
	}

	var payload oauthStatePayload
	if err := json.Unmarshal(payloadRaw, &payload); err != nil {
		return nil, fmt.Errorf("%w: invalid payload", ErrInvalidInput)
	}
	if payload.IssuedAt <= 0 {
		return nil, fmt.Errorf("%w: missing issuedAt", ErrInvalidInput)
	}
	if time.Since(time.Unix(payload.IssuedAt, 0)) > oauthStateTTL {
		return nil, fmt.Errorf("%w: expired state", ErrInvalidInput)
	}

	return &payload, nil
}

func normalizeOAuthAudience(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", oauthAudienceBookingWeb:
		return oauthAudienceBookingWeb
	case oauthAudienceTenantWeb:
		return oauthAudienceTenantWeb
	default:
		return ""
	}
}

func isAllowedOAuthRedirect(rawURL, baseDomain string) bool {
	if strings.TrimSpace(rawURL) == "" {
		return false
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	if !strings.EqualFold(parsed.Scheme, "http") && !strings.EqualFold(parsed.Scheme, "https") {
		return false
	}

	host := strings.ToLower(strings.TrimSpace(parsed.Hostname()))
	if host == "" {
		return false
	}

	base := strings.ToLower(strings.TrimSpace(baseDomain))
	return host == base || strings.HasSuffix(host, "."+base)
}

func appendQueryString(rawURL string, params map[string]string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	query := parsed.Query()
	for key, value := range params {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			query.Set(key, trimmed)
		}
	}
	parsed.RawQuery = query.Encode()

	return parsed.String()
}

func oauthHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 10 * time.Second,
	}
}

func fallbackOAuthRedirectFromReferer(rawReferer string) string {
	ref := strings.TrimSpace(rawReferer)
	if ref == "" {
		return ""
	}

	parsed, err := url.Parse(ref)
	if err != nil {
		return ""
	}
	if !strings.EqualFold(parsed.Scheme, "http") && !strings.EqualFold(parsed.Scheme, "https") {
		return ""
	}

	// Normalize to the login page while preserving requested "next" if present.
	nextPath := strings.TrimSpace(parsed.Query().Get("next"))
	parsed.Path = "/login"
	parsed.RawQuery = ""
	if nextPath != "" {
		query := url.Values{}
		query.Set("next", nextPath)
		parsed.RawQuery = query.Encode()
	}

	return parsed.String()
}
