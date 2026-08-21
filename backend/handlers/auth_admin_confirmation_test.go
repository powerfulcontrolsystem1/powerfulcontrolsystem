package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestConfirmarAdminHandlerRedirectsMissingTokenToStyledLogin(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/auth/confirmar_admin", nil)
	rr := httptest.NewRecorder()

	ConfirmarAdminHandler(nil).ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusSeeOther)
	}
	if location := rr.Header().Get("Location"); location != "/login.html?confirmacion=error" {
		t.Fatalf("Location = %q, want styled login confirmation error", location)
	}
}

func TestConfirmarAdminHandlerRejectsNonGET(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/auth/confirmar_admin", nil)
	rr := httptest.NewRecorder()

	ConfirmarAdminHandler(nil).ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusMethodNotAllowed)
	}
}
