package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

var (
	ErrGoogleNotConfigured    = errors.New("google auth is not configured")
	ErrInvalidGoogleToken     = errors.New("invalid google token")
	ErrGoogleEmailNotVerified = errors.New("google email is not verified")
	ErrGoogleCodeExchange     = errors.New("failed to exchange google auth code")
	ErrGoogleProfileFetch     = errors.New("failed to fetch google profile")
)

const (
	googleTokenInfoURL = "https://oauth2.googleapis.com/tokeninfo"
	googleAuthURL      = "https://accounts.google.com/o/oauth2/v2/auth"
	googleTokenURL     = "https://oauth2.googleapis.com/token"
	googleUserInfoURL  = "https://openidconnect.googleapis.com/v1/userinfo"
)

// GoogleIdentityPayload contains the normalized Google identity data used by auth flows.
type GoogleIdentityPayload struct {
	Subject       string
	Email         string
	Name          string
	PictureURL    string
	EmailVerified bool
}

// GoogleOAuthTokens contains the token response from Google's OAuth token endpoint.
type GoogleOAuthTokens struct {
	AccessToken  string
	RefreshToken string
	IDToken      string
}

// GoogleAuthService handles Google auth verification and OAuth exchange.
type GoogleAuthService struct {
	clientID     string
	clientSecret string
	redirectURL  string
	scopes       []string
	httpClient   *http.Client
}

// NewGoogleAuthService creates a GoogleAuthService from environment configuration.
func NewGoogleAuthService() *GoogleAuthService {
	scopes := strings.Fields(os.Getenv("GOOGLE_OAUTH_SCOPES"))
	if len(scopes) == 0 {
		scopes = []string{"openid", "email", "profile"}
	}

	return &GoogleAuthService{
		clientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		clientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		redirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		scopes:       scopes,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// IsConfigured reports whether the Google auth service has the required configuration.
func (s *GoogleAuthService) IsConfigured() bool {
	return s.clientID != "" && s.clientSecret != "" && s.redirectURL != ""
}

// VerifyIDToken verifies a Google ID token and returns normalized identity data.
func (s *GoogleAuthService) VerifyIDToken(ctx context.Context, idToken string) (*GoogleIdentityPayload, error) {
	if s.clientID == "" {
		return nil, ErrGoogleNotConfigured
	}
	if idToken == "" {
		return nil, ErrInvalidGoogleToken
	}

	endpoint, err := url.Parse(googleTokenInfoURL)
	if err != nil {
		return nil, err
	}

	query := endpoint.Query()
	query.Set("id_token", idToken)
	endpoint.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build google token verification request: %v", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to verify google token: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ErrInvalidGoogleToken
	}

	var body struct {
		Audience      string `json:"aud"`
		ExpiresIn     string `json:"expires_in"`
		Issuer        string `json:"iss"`
		Subject       string `json:"sub"`
		Email         string `json:"email"`
		EmailVerified string `json:"email_verified"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("failed to decode google token verification response: %v", err)
	}

	if body.Audience != s.clientID || body.Subject == "" || body.Email == "" {
		return nil, ErrInvalidGoogleToken
	}

	if body.Issuer != "accounts.google.com" && body.Issuer != "https://accounts.google.com" {
		return nil, ErrInvalidGoogleToken
	}

	if body.EmailVerified != "true" {
		return nil, ErrGoogleEmailNotVerified
	}

	return &GoogleIdentityPayload{
		Subject:       body.Subject,
		Email:         body.Email,
		Name:          body.Name,
		PictureURL:    body.Picture,
		EmailVerified: true,
	}, nil
}

// BuildAuthURL creates a Google OAuth authorization URL.
func (s *GoogleAuthService) BuildAuthURL(state string) (string, error) {
	if !s.IsConfigured() {
		return "", ErrGoogleNotConfigured
	}

	values := url.Values{}
	values.Set("client_id", s.clientID)
	values.Set("redirect_uri", s.redirectURL)
	values.Set("response_type", "code")
	values.Set("scope", strings.Join(s.scopes, " "))
	values.Set("state", state)
	values.Set("access_type", "offline")
	values.Set("include_granted_scopes", "true")
	values.Set("prompt", "consent")

	return fmt.Sprintf("%s?%s", googleAuthURL, values.Encode()), nil
}

// ExchangeCode exchanges an OAuth authorization code for Google tokens.
func (s *GoogleAuthService) ExchangeCode(ctx context.Context, code string) (*GoogleOAuthTokens, error) {
	if !s.IsConfigured() {
		return nil, ErrGoogleNotConfigured
	}
	if code == "" {
		return nil, ErrGoogleCodeExchange
	}

	form := url.Values{}
	form.Set("code", code)
	form.Set("client_id", s.clientID)
	form.Set("client_secret", s.clientSecret)
	form.Set("redirect_uri", s.redirectURL)
	form.Set("grant_type", "authorization_code")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, googleTokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to build google code exchange request: %v", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange google auth code: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ErrGoogleCodeExchange
	}

	var body struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		IDToken      string `json:"id_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("failed to decode google token exchange response: %v", err)
	}

	if body.AccessToken == "" && body.IDToken == "" {
		return nil, ErrGoogleCodeExchange
	}

	return &GoogleOAuthTokens{
		AccessToken:  body.AccessToken,
		RefreshToken: body.RefreshToken,
		IDToken:      body.IDToken,
	}, nil
}

// FetchUserInfo retrieves Google profile data with an access token.
func (s *GoogleAuthService) FetchUserInfo(ctx context.Context, accessToken string) (*GoogleIdentityPayload, error) {
	if accessToken == "" {
		return nil, ErrGoogleProfileFetch
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, googleUserInfoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build google userinfo request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch google userinfo: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ErrGoogleProfileFetch
	}

	var body struct {
		Subject       string `json:"sub"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("failed to decode google userinfo response: %v", err)
	}

	if body.Subject == "" || body.Email == "" {
		return nil, ErrGoogleProfileFetch
	}
	if !body.EmailVerified {
		return nil, ErrGoogleEmailNotVerified
	}

	return &GoogleIdentityPayload{
		Subject:       body.Subject,
		Email:         body.Email,
		Name:          body.Name,
		PictureURL:    body.Picture,
		EmailVerified: body.EmailVerified,
	}, nil
}
