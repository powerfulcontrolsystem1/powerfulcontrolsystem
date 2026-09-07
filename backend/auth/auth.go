package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const maxGoogleResponseBytes = 1 << 20

// googleHTTPClient is deliberately explicit: identity-provider calls must not
// inherit global client behavior or wait forever on a remote dependency.
var googleHTTPClient = &http.Client{Timeout: 15 * time.Second}

// ExchangeCodeForToken realiza POST a Google token endpoint
func ExchangeCodeForToken(code, clientID, clientSecret, redirectURL string) (*TokenResponse, error) {
	return ExchangeCodeForTokenWithPKCE(context.Background(), code, clientID, clientSecret, redirectURL, "")
}

// ExchangeCodeForTokenWithPKCE exchanges a Google authorization code and uses
// a PKCE verifier when the authorization flow supplied one.
func ExchangeCodeForTokenWithPKCE(ctx context.Context, code, clientID, clientSecret, redirectURL, codeVerifier string) (*TokenResponse, error) {
	data := url.Values{}
	data.Set("code", code)
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("redirect_uri", redirectURL)
	data.Set("grant_type", "authorization_code")
	if strings.TrimSpace(codeVerifier) != "" {
		data.Set("code_verifier", codeVerifier)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://oauth2.googleapis.com/token", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := googleHTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, readErr := io.ReadAll(io.LimitReader(resp.Body, maxGoogleResponseBytes+1))
	if readErr != nil {
		return nil, fmt.Errorf("read token response: %w", readErr)
	}
	if len(b) > maxGoogleResponseBytes {
		return nil, fmt.Errorf("token response too large")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token endpoint returned %d", resp.StatusCode)
	}
	var tr TokenResponse
	if err := json.Unmarshal(b, &tr); err != nil {
		return nil, err
	}
	return &tr, nil
}

// FetchUserInfo solicita el endpoint userinfo
func FetchUserInfo(accessToken string) (*UserInfo, error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://www.googleapis.com/oauth2/v3/userinfo", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := googleHTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, readErr := io.ReadAll(io.LimitReader(resp.Body, maxGoogleResponseBytes+1))
	if readErr != nil {
		return nil, fmt.Errorf("read userinfo response: %w", readErr)
	}
	if len(b) > maxGoogleResponseBytes {
		return nil, fmt.Errorf("userinfo response too large")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("userinfo endpoint %d", resp.StatusCode)
	}
	var u UserInfo
	if err := json.Unmarshal(b, &u); err != nil {
		return nil, err
	}
	return &u, nil
}

// Tipos para respuestas
type TokenResponse struct {
	AccessToken string `json:"access_token"`
}

type UserInfo struct {
	Sub           string `json:"sub"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
}
