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
	req.Header.Set("Origin", "https://service.test")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("same-origin cookie mutation rejected: %d", rec.Code)
	}
}
