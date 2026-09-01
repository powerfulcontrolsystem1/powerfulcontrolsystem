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
	t.Setenv("NEXTCLOUD_BASE_URL", "https://nextcloud.example.test")
	t.Setenv("PCS_CSP_CONNECT_ORIGINS", "https://api.example.test")
	t.Setenv("PCS_CSP_IMG_ORIGINS", "https://images.example.test")
	t.Setenv("PCS_CSP_SCRIPT_ORIGINS", "https://scripts.example.test")
	t.Setenv("PCS_CSP_STYLE_ORIGINS", "https://styles.example.test")
	t.Setenv("PCS_CSP_FONT_ORIGINS", "https://fonts.example.test")
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
	for _, expected := range []string{"form-action 'self'", "font-src 'self' data:", "https://onlyoffice.example.test", "https://nextcloud.example.test", "https://mail.powerfulcontrolsystem.com", "https://www.google.com", "https://www.gstatic.com", "https://api.example.test", "https://images.example.test", "https://scripts.example.test", "https://styles.example.test", "https://fonts.example.test", "https://lh3.googleusercontent.com"} {
		if !strings.Contains(policy, expected) {
			t.Fatalf("CSP missing explicit source %q: %q", expected, policy)
		}
	}
	reportOnly := rec.Header().Get("Content-Security-Policy-Report-Only")
	if reportOnly != policy {
		t.Fatalf("CSP report-only must match the enforced compatibility policy unless strict reporting is enabled: %q != %q", reportOnly, policy)
	}
}

func TestSecurityHeadersCanEnableStrictReportOnlyCSPWithoutBlockingCompatibility(t *testing.T) {
	t.Setenv("PCS_CSP_REPORT_ONLY_STRICT", "true")
	h := SecurityHeadersMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) }))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/administrar_empresa/administrar_productos.html", nil))

	enforced := rec.Header().Get("Content-Security-Policy")
	reportOnly := rec.Header().Get("Content-Security-Policy-Report-Only")
	if !strings.Contains(enforced, "'unsafe-inline'") {
		t.Fatalf("compatibility CSP unexpectedly removed inline support: %q", enforced)
	}
	if strings.Contains(reportOnly, "'unsafe-inline'") {
		t.Fatalf("strict report-only CSP must omit unsafe-inline: %q", reportOnly)
	}
	for _, expected := range []string{"default-src 'self'", "form-action 'self'", "script-src 'self'", "style-src 'self'"} {
		if !strings.Contains(reportOnly, expected) {
			t.Fatalf("strict report-only CSP missing %q: %q", expected, reportOnly)
		}
	}
}

func TestSecurityHeadersAllowGoogleRecaptchaWithoutWildcardOrigins(t *testing.T) {
	h := SecurityHeadersMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) }))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/login.html", nil))

	for _, header := range []string{"Content-Security-Policy", "Content-Security-Policy-Report-Only"} {
		policy := rec.Header().Get(header)
		for _, directive := range []string{"script-src", "connect-src", "frame-src"} {
			start := strings.Index(policy, directive+" ")
			if start < 0 {
				t.Fatalf("%s missing %s: %q", header, directive, policy)
			}
			section := policy[start:]
			if end := strings.Index(section, ";"); end >= 0 {
				section = section[:end]
			}
			for _, origin := range []string{"https://www.google.com", "https://www.gstatic.com"} {
				if !strings.Contains(section, origin) {
					t.Fatalf("%s %s missing reCAPTCHA origin %s: %q", header, directive, origin, section)
				}
			}
		}
		if strings.Contains(policy, "https://*.google.com") {
			t.Fatalf("%s must not use a broad Google wildcard: %q", header, policy)
		}
	}
}

func TestSecurityHeadersUseCompactDenyAllPolicyForGoogleOAuth(t *testing.T) {
	h := SecurityHeadersMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusFound) }))
	for _, path := range []string{"/auth/google/login", "/auth/google/usuario/login", "/auth/google/callback"} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		for _, header := range []string{"Content-Security-Policy", "Content-Security-Policy-Report-Only"} {
			if got := rec.Header().Get(header); got != googleOAuthResponseCSP {
				t.Fatalf("%s %s = %q, want compact OAuth policy %q", path, header, got, googleOAuthResponseCSP)
			}
		}
		if strings.Contains(rec.Header().Get("Content-Security-Policy"), "unsafe-inline") {
			t.Fatalf("%s OAuth CSP must deny inline content", path)
		}
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
