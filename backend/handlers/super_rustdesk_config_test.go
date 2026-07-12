package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRustDeskConfigHandlerRequiresSuperAdmin(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/super/api/rustdesk/config", nil)
	res := httptest.NewRecorder()

	RustDeskConfigHandler(nil).ServeHTTP(res, req)

	if res.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusUnauthorized)
	}
}
