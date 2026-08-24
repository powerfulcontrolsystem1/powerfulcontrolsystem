package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func responseHeaderBytes(header http.Header) int {
	total := 0
	for name, values := range header {
		for _, value := range values {
			total += len(name) + 2 + len(value) + 2
		}
	}
	return total
}

func TestGoogleAdminLoginHeadersFitReverseProxyBudget(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "https://powerfulcontrolsystem.com/auth/google/login", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	rr := httptest.NewRecorder()

	HandleGoogleLogin("client-id", "https://powerfulcontrolsystem.com/auth/google/callback").ServeHTTP(rr, req)

	if rr.Code != http.StatusFound {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusFound)
	}
	if location := rr.Header().Get("Location"); !strings.HasPrefix(location, "https://accounts.google.com/") {
		t.Fatalf("unexpected OAuth redirect: %q", location)
	}
	if got := len(rr.Header().Values("Set-Cookie")); got > 5 {
		t.Fatalf("Set-Cookie count = %d, want <= 5 to avoid proxy 502", got)
	}
	if got := responseHeaderBytes(rr.Header()); got >= 4096 {
		t.Fatalf("OAuth response headers = %d bytes, want < 4096", got)
	}
}

func TestGoogleEmpresaLoginHeadersFitReverseProxyBudget(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "https://powerfulcontrolsystem.com/auth/google/usuario/login?empresa_id=12&token_invitacion=token-prueba", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	rr := httptest.NewRecorder()

	HandleGoogleUsuarioLogin("client-id", "https://powerfulcontrolsystem.com/auth/google/callback").ServeHTTP(rr, req)

	if rr.Code != http.StatusFound {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusFound)
	}
	if got := responseHeaderBytes(rr.Header()); got >= 4096 {
		t.Fatalf("enterprise OAuth response headers = %d bytes, want < 4096", got)
	}
}
