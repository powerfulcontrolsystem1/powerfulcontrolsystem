package utils

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCSRFMiddlewareRejectsCrossOriginCookieMutation(t *testing.T) {
	h := CSRFMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) }))
	req := httptest.NewRequest(http.MethodPost, "https://service.test/api/empresa/productos", nil)
	req.AddCookie(&http.Cookie{Name: "session_token", Value: "opaque"})
	req.Header.Set("Origin", "https://attacker.test")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("cross-origin cookie mutation accepted: %d", rec.Code)
	}
}

func TestCSRFMiddlewareAllowsSameOriginCookieMutation(t *testing.T) {
	h := CSRFMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) }))
	req := httptest.NewRequest(http.MethodPost, "https://service.test/api/empresa/productos", nil)
	req.AddCookie(&http.Cookie{Name: "session_token", Value: "opaque"})
	req.AddCookie(&http.Cookie{Name: csrfCookieName, Value: "csrf-value"})
	req.Header.Set("Origin", "https://service.test")
	req.Header.Set(csrfHeaderName, "csrf-value")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("same-origin cookie mutation rejected: %d", rec.Code)
	}
}

func TestCSRFMiddlewareRejectsMissingOrIncorrectToken(t *testing.T) {
	for _, token := range []string{"", "different"} {
		h := CSRFMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) }))
		req := httptest.NewRequest(http.MethodPost, "https://service.test/api/empresa/productos", nil)
		req.AddCookie(&http.Cookie{Name: "session_token", Value: "opaque"})
		req.AddCookie(&http.Cookie{Name: csrfCookieName, Value: "csrf-value"})
		req.Header.Set("Origin", "https://service.test")
		if token != "" {
			req.Header.Set(csrfHeaderName, token)
		}
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusForbidden {
			t.Fatalf("token %q accepted: %d", token, rec.Code)
		}
	}
}

func TestCSRFMiddlewareAllowsPublicLoginWithStaleSessionCookie(t *testing.T) {
	h := CSRFMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) }))
	req := httptest.NewRequest(http.MethodPost, "https://service.test/super/api/administradores/login", nil)
	req.AddCookie(&http.Cookie{Name: "session_token", Value: "stale"})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("public login rejected because of stale cookie: %d", rec.Code)
	}
}

func TestCSRFMiddlewareRejectsDifferentPortAndSubdomain(t *testing.T) {
	for _, origin := range []string{"https://service.test:8443", "https://sub.service.test"} {
		h := CSRFMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) }))
		req := httptest.NewRequest(http.MethodPost, "https://service.test/api/empresa/productos", nil)
		req.AddCookie(&http.Cookie{Name: "session_token", Value: "opaque"})
		req.Header.Set("Origin", origin)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusForbidden {
			t.Fatalf("origin %q accepted: %d", origin, rec.Code)
		}
	}
}
