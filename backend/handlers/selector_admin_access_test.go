package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestWithAdminAuditoriaRejectsUnauthenticatedRequest(t *testing.T) {
	called := false
	h := WithAdminAuditoria(nil, "selector", func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	})
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/super/api/empresas", nil))
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
	if called {
		t.Fatal("wrapped selector handler ran without an authenticated administrator")
	}
}

func TestWithAdminAuditoriaRejectsEmpresaUserPrincipal(t *testing.T) {
	called := false
	h := WithAdminAuditoria(nil, "selector", func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	})
	req := httptest.NewRequest(http.MethodGet, "/super/api/empresas", nil)
	req = req.WithContext(context.WithValue(req.Context(), "sessionPrincipalType", "empresa_usuario"))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusForbidden)
	}
	if called {
		t.Fatal("enterprise user principal reached account selector endpoint")
	}
}

func TestSelectorRoutesUseAccountScopedAuditWrappers(t *testing.T) {
	body, err := os.ReadFile("../main.go")
	if err != nil {
		t.Fatal(err)
	}
	source := string(body)
	required := []string{
		`"/super/api/tipos_empresas", handlers.WithAdminReadSuperWriteAuditoria`,
		`"/super/api/empresas", handlers.WithAdminAuditoria`,
		`"/super/api/empresas/compartidos", handlers.WithAdminAuditoria`,
		`"/super/api/empresas/compartidos/aceptar", handlers.WithAdminAuditoria`,
		`"/super/api/licencias", handlers.WithAdminAuditoria`,
	}
	for _, contract := range required {
		if !strings.Contains(source, contract) {
			t.Fatalf("selector route missing account-scoped wrapper: %s", contract)
		}
	}
}
