package auth

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(req *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func withMockHTTPClient(t *testing.T, fn roundTripFunc) {
	t.Helper()
	original := http.DefaultClient
	http.DefaultClient = &http.Client{Transport: fn}
	t.Cleanup(func() {
		http.DefaultClient = original
	})
}

func TestExchangeCodeForTokenSuccess(t *testing.T) {
	withMockHTTPClient(t, func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", req.Method)
		}
		if req.URL.String() != "https://oauth2.googleapis.com/token" {
			t.Fatalf("unexpected url: %s", req.URL.String())
		}
		if got := req.Header.Get("Content-Type"); got != "application/x-www-form-urlencoded" {
			t.Fatalf("unexpected content-type: %s", got)
		}
		body, err := io.ReadAll(req.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		raw := string(body)
		for _, token := range []string{"code=code123", "client_id=client", "client_secret=secret", "grant_type=authorization_code"} {
			if !strings.Contains(raw, token) {
				t.Fatalf("expected body to contain %q, got %q", token, raw)
			}
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(`{"access_token":"token-abc"}`)),
		}, nil
	})

	resp, err := ExchangeCodeForToken("code123", "client", "secret", "http://localhost/callback")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if resp == nil || resp.AccessToken != "token-abc" {
		t.Fatalf("unexpected token response: %+v", resp)
	}
}

func TestExchangeCodeForTokenNon200(t *testing.T) {
	withMockHTTPClient(t, func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusBadRequest,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader("invalid code")),
		}, nil
	})

	_, err := ExchangeCodeForToken("bad", "client", "secret", "http://localhost/callback")
	if err == nil {
		t.Fatal("expected error for non-200 response")
	}
	if !strings.Contains(err.Error(), "returned 400") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFetchUserInfoSuccess(t *testing.T) {
	withMockHTTPClient(t, func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", req.Method)
		}
		if req.URL.String() != "https://www.googleapis.com/oauth2/v3/userinfo" {
			t.Fatalf("unexpected url: %s", req.URL.String())
		}
		if got := req.Header.Get("Authorization"); got != "Bearer token-xyz" {
			t.Fatalf("unexpected auth header: %s", got)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body: io.NopCloser(strings.NewReader(`{
				"sub":"u1",
				"name":"Jane Doe",
				"email":"jane@example.com",
				"email_verified":true
			}`)),
		}, nil
	})

	user, err := FetchUserInfo("token-xyz")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if user == nil || user.Email != "jane@example.com" || !user.EmailVerified {
		t.Fatalf("unexpected user info: %+v", user)
	}
}

func TestFetchUserInfoNon200(t *testing.T) {
	withMockHTTPClient(t, func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusUnauthorized,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader("unauthorized")),
		}, nil
	})

	_, err := FetchUserInfo("bad-token")
	if err == nil {
		t.Fatal("expected error for non-200 response")
	}
	if !strings.Contains(err.Error(), "userinfo endpoint 401") {
		t.Fatalf("unexpected error: %v", err)
	}
}
