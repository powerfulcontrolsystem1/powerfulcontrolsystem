package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSuperVPS2HandlerRequiresSuperAdmin(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/super/api/vps2?action=status", nil)
	res := httptest.NewRecorder()

	SuperVPS2Handler(nil).ServeHTTP(res, req)

	if res.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusUnauthorized)
	}
}
