package users

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// VerifiedSocialIdentity is the provider-backed identity claims.
type VerifiedSocialIdentity struct {
	Provider       string
	ProviderUserID string
	Email          string
	Name           string
	EmailVerified  bool
}

// SocialVerifier validates social provider tokens with provider APIs.
type SocialVerifier struct {
	httpClient     *http.Client
	googleClientID string
	facebookAppID  string
}

// NewSocialVerifier creates a verifier with provider configuration.
func NewSocialVerifier(googleClientID, facebookAppID string) *SocialVerifier {
	return &SocialVerifier{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		googleClientID: strings.TrimSpace(googleClientID),
		facebookAppID:  strings.TrimSpace(facebookAppID),
	}
}

// Verify validates provider tokens and returns trusted identity claims.
func (v *SocialVerifier) Verify(ctx context.Context, input SocialLoginInput) (*VerifiedSocialIdentity, error) {
	switch strings.ToLower(strings.TrimSpace(input.Provider)) {
	case AuthProviderGoogle:
		return v.verifyGoogle(ctx, input)
	case AuthProviderFacebook:
		return v.verifyFacebook(ctx, input)
	default:
		return nil, ErrUnsupportedProvider
	}
}

func (v *SocialVerifier) verifyGoogle(ctx context.Context, input SocialLoginInput) (*VerifiedSocialIdentity, error) {
	if strings.TrimSpace(input.GoogleIDToken) != "" {
		return v.verifyGoogleIDToken(ctx, input.GoogleIDToken)
	}
	if strings.TrimSpace(input.GoogleAccessToken) != "" {
		return v.verifyGoogleAccessToken(ctx, input.GoogleAccessToken)
	}
	return nil, fmt.Errorf("%w: googleIdToken or googleAccessToken is required", ErrInvalidInput)
}

func (v *SocialVerifier) verifyGoogleIDToken(ctx context.Context, idToken string) (*VerifiedSocialIdentity, error) {
	endpoint := "https://oauth2.googleapis.com/tokeninfo?id_token=" + url.QueryEscape(strings.TrimSpace(idToken))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	res, err := v.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: google token verification failed", ErrSocialValidationFailed)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: google token rejected", ErrSocialValidationFailed)
	}

	var payload struct {
		Sub           string `json:"sub"`
		Email         string `json:"email"`
		Name          string `json:"name"`
		EmailVerified string `json:"email_verified"`
		Audience      string `json:"aud"`
	}
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("%w: google token response decode failed", ErrSocialValidationFailed)
	}

	if strings.TrimSpace(payload.Sub) == "" || strings.TrimSpace(payload.Email) == "" {
		return nil, fmt.Errorf("%w: google token missing identity fields", ErrSocialValidationFailed)
	}
	if v.googleClientID != "" && strings.TrimSpace(payload.Audience) != v.googleClientID {
		return nil, fmt.Errorf("%w: google token audience mismatch", ErrSocialValidationFailed)
	}

	verified := strings.EqualFold(strings.TrimSpace(payload.EmailVerified), "true")

	return &VerifiedSocialIdentity{
		Provider:       AuthProviderGoogle,
		ProviderUserID: strings.TrimSpace(payload.Sub),
		Email:          strings.TrimSpace(strings.ToLower(payload.Email)),
		Name:           strings.TrimSpace(payload.Name),
		EmailVerified:  verified,
	}, nil
}

func (v *SocialVerifier) verifyGoogleAccessToken(ctx context.Context, accessToken string) (*VerifiedSocialIdentity, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://www.googleapis.com/oauth2/v3/userinfo", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(accessToken))

	res, err := v.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: google userinfo request failed", ErrSocialValidationFailed)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: google access token rejected", ErrSocialValidationFailed)
	}

	var payload struct {
		Sub           string `json:"sub"`
		Email         string `json:"email"`
		Name          string `json:"name"`
		EmailVerified bool   `json:"email_verified"`
	}
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("%w: google userinfo decode failed", ErrSocialValidationFailed)
	}

	if strings.TrimSpace(payload.Sub) == "" || strings.TrimSpace(payload.Email) == "" {
		return nil, fmt.Errorf("%w: google userinfo missing identity fields", ErrSocialValidationFailed)
	}

	return &VerifiedSocialIdentity{
		Provider:       AuthProviderGoogle,
		ProviderUserID: strings.TrimSpace(payload.Sub),
		Email:          strings.TrimSpace(strings.ToLower(payload.Email)),
		Name:           strings.TrimSpace(payload.Name),
		EmailVerified:  payload.EmailVerified,
	}, nil
}

func (v *SocialVerifier) verifyFacebook(ctx context.Context, input SocialLoginInput) (*VerifiedSocialIdentity, error) {
	accessToken := strings.TrimSpace(input.FacebookAccessToken)
	if accessToken == "" {
		return nil, fmt.Errorf("%w: facebookAccessToken is required", ErrInvalidInput)
	}

	endpoint := "https://graph.facebook.com/me?fields=id,name,email&access_token=" + url.QueryEscape(accessToken)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	res, err := v.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: facebook profile request failed", ErrSocialValidationFailed)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: facebook access token rejected", ErrSocialValidationFailed)
	}

	var payload struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("%w: facebook profile decode failed", ErrSocialValidationFailed)
	}

	if strings.TrimSpace(payload.ID) == "" || strings.TrimSpace(payload.Email) == "" {
		return nil, fmt.Errorf("%w: facebook account missing email or id", ErrSocialValidationFailed)
	}

	return &VerifiedSocialIdentity{
		Provider:       AuthProviderFacebook,
		ProviderUserID: strings.TrimSpace(payload.ID),
		Email:          strings.TrimSpace(strings.ToLower(payload.Email)),
		Name:           strings.TrimSpace(payload.Name),
		EmailVerified:  true,
	}, nil
}
