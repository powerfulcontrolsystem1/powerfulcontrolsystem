package utils

import (
	"net/http"
	"net/http/httptest"
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
}
