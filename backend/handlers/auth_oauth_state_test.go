package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func oauthCookiesFromRecorder(rec *httptest.ResponseRecorder) []*http.Cookie {
	return rec.Result().Cookies()
}

func cookieHeader(cookies []*http.Cookie) string {
	values := make([]string, 0, len(cookies))
	for _, cookie := range cookies {
		if cookie.MaxAge >= 0 && cookie.Value != "" {
			values = append(values, cookie.Name+"="+cookie.Value)
		}
	}
	return strings.Join(values, "; ")
}

func TestGoogleOAuthStartUsesRandomStateAndPKCE(t *testing.T) {
	handler := HandleGoogleLogin("client-id", "https://example.test/auth/google/callback")
	first := httptest.NewRecorder()
	handler(first, httptest.NewRequest(http.MethodGet, "/auth/google/login", nil))
	second := httptest.NewRecorder()
	handler(second, httptest.NewRequest(http.MethodGet, "/auth/google/login", nil))

	for _, rec := range []*httptest.ResponseRecorder{first, second} {
		if rec.Code != http.StatusFound {
			t.Fatalf("expected OAuth redirect, got %d", rec.Code)
		}
		location, err := url.Parse(rec.Header().Get("Location"))
		if err != nil {
			t.Fatalf("parse redirect: %v", err)
		}
		state := location.Query().Get("state")
		if len(state) < 64 || location.Query().Get("code_challenge_method") != "S256" || location.Query().Get("code_challenge") == "" {
			t.Fatalf("OAuth authorization request missing secure state/PKCE: %s", location.String())
		}
		foundStateCookie := false
		for _, cookie := range oauthCookiesFromRecorder(rec) {
			if cookie.Name == googleOAuthStateAdminCookieName {
				foundStateCookie = true
				if cookie.HttpOnly == false || cookie.Path != googleOAuthCookiePath || cookie.MaxAge != googleOAuthFlowTTL || cookie.Value == state {
					t.Fatalf("state cookie must be protected and store only a hash")
				}
			}
		}
		if !foundStateCookie {
			t.Fatal("state cookie not issued")
		}
	}
	firstURL, _ := url.Parse(first.Header().Get("Location"))
	secondURL, _ := url.Parse(second.Header().Get("Location"))
	if firstURL.Query().Get("state") == secondURL.Query().Get("state") {
		t.Fatal("OAuth state must be unique per login attempt")
	}
}

func TestGoogleOAuthCallbackRejectsMissingInvalidAndReusedState(t *testing.T) {
	start := httptest.NewRecorder()
	HandleGoogleLogin("client-id", "https://example.test/auth/google/callback")(start, httptest.NewRequest(http.MethodGet, "/auth/google/login", nil))
	cookies := oauthCookiesFromRecorder(start)

	for _, rawQuery := range []string{"code=x", "code=x&state=wrong"} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/auth/google/callback?"+rawQuery, nil)
		req.Header.Set("Cookie", cookieHeader(cookies))
		HandleGoogleCallback(nil, nil, "client-id", "secret", "https://example.test/auth/google/callback")(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected invalid OAuth state rejection for %q, got %d", rawQuery, rec.Code)
		}
	}

	stateURL, _ := url.Parse(start.Header().Get("Location"))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/auth/google/callback?code=x&state="+url.QueryEscape(stateURL.Query().Get("state")), nil)
	req.Header.Set("Cookie", cookieHeader(cookies))
	_, _, err := validateAndConsumeGoogleOAuthState(rec, req)
	if err != nil {
		t.Fatalf("expected first state validation to work: %v", err)
	}
	reused := httptest.NewRecorder()
	reusedReq := httptest.NewRequest(http.MethodGet, "/auth/google/callback?code=x&state="+url.QueryEscape(stateURL.Query().Get("state")), nil)
	if _, _, err := validateAndConsumeGoogleOAuthState(reused, reusedReq); err == nil {
		t.Fatal("reused OAuth callback without the consumed cookie must be rejected")
	}
}
