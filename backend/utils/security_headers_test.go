package utils

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestForwardedHeadersIgnoredOutsideTrustedProxy(t *testing.T) {
	t.Setenv("PCS_TRUSTED_PROXY_CIDRS", "127.0.0.1/32")
	req := httptest.NewRequest(http.MethodGet, "https://service.test/", nil)
	req.RemoteAddr = "198.51.100.24:1234"
	req.Header.Set("X-Forwarded-For", "203.0.113.10")
	req.Header.Set("X-Forwarded-Host", "attacker.test")
	if got := requestClientIP(req); got != "198.51.100.24" {
		t.Fatalf("untrusted forwarded IP accepted: %q", got)
	}
	if got := resolveRequestHost(req); got != "service.test" {
		t.Fatalf("untrusted forwarded host accepted: %q", got)
	}
}

func TestSecurityHeadersAndNoStoreOnLogin(t *testing.T) {
	t.Setenv("ONLYOFFICE_DOCUMENT_SERVER_URL", "https://onlyoffice.example.test")
	t.Setenv("PCS_CSP_CONNECT_ORIGINS", "https://api.example.test")
	t.Setenv("PCS_CSP_IMG_ORIGINS", "https://images.example.test")
	h := SecurityHeadersMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) }))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/login.html", nil))
	for header, expected := range map[string]string{
		"X-Content-Type-Options": "nosniff",
		"Referrer-Policy":        "strict-origin-when-cross-origin",
		"Cache-Control":          "no-store",
	} {
		if got := rec.Header().Get(header); got != expected {
			t.Fatalf("%s = %q, want %q", header, got, expected)
		}
	}
	if rec.Header().Get("Content-Security-Policy") == "" {
		t.Fatal("CSP header missing")
	}
	policy := rec.Header().Get("Content-Security-Policy")
	for _, forbidden := range []string{"img-src 'self' data: https:", "connect-src 'self' https: ", "connect-src 'self' wss: ", "https://*.google.com"} {
		if strings.Contains(policy, forbidden) {
			t.Fatalf("CSP keeps broad source %q: %q", forbidden, policy)
		}
	}
	for _, expected := range []string{"form-action 'self'", "https://onlyoffice.example.test", "https://api.example.test", "https://images.example.test"} {
		if !strings.Contains(policy, expected) {
			t.Fatalf("CSP missing explicit source %q: %q", expected, policy)
		}
	}
	reportOnly := rec.Header().Get("Content-Security-Policy-Report-Only")
	if reportOnly != policy {
		t.Fatalf("CSP report-only must be generated from the same policy: %q != %q", reportOnly, policy)
	}
}

func TestCSPOriginRejectsWildcardPathAndCredentials(t *testing.T) {
	t.Parallel()
	for _, raw := range []string{"https://*.example.test", "https://example.test/path", "https://user:pass@example.test", "javascript:alert(1)"} {
		if got := cspOrigin(raw); got != "" {
			t.Fatalf("unsafe CSP source accepted %q as %q", raw, got)
		}
	}
}

func TestSecurityContentSecurityPolicyUpgradesRequestsOnlyInProduction(t *testing.T) {
	t.Setenv("PCS_ENV", "production")
	if policy := securityContentSecurityPolicy(); !strings.Contains(policy, "upgrade-insecure-requests") {
		t.Fatalf("production CSP must upgrade insecure requests: %q", policy)
	}
	t.Setenv("PCS_ENV", "development")
	if policy := securityContentSecurityPolicy(); strings.Contains(policy, "upgrade-insecure-requests") {
		t.Fatalf("development CSP must not force upgrade-insecure-requests: %q", policy)
	}
}
