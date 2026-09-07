package utils

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSessionCookieMaxAgeUsesConfiguredDuration(t *testing.T) {
	t.Setenv("SESSION_TIMEOUT", "2h")
	if got := SessionCookieMaxAge(); got != 7200 {
		t.Fatalf("expected 7200 seconds, got %d", got)
	}
}

func TestRequestBodyLimitMiddlewareRejectsOversizedBody(t *testing.T) {
	handler := RequestBodyLimitMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "too large", http.StatusRequestEntityTooLarge)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}), 4)
	req := httptest.NewRequest(http.MethodPost, "/api/test", strings.NewReader("12345"))
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected request entity too large, got %d", res.Code)
	}
}
