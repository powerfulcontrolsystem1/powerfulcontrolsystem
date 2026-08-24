package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestEmpresaUsuarioCredentialResolversRequireTenantScope(t *testing.T) {
	if _, err := resolveUniqueEmpresaUsuarioByEmail(nil, "qa@example.test", 0); !errors.Is(err, errEmpresaUsuarioScopeRequired) {
		t.Fatalf("unique lookup without empresa_id must fail with scope error, got %v", err)
	}
	if _, _, err := resolveEmpresaUsuarioForPasswordLogin(nil, "qa@example.test", "secret", 0); !errors.Is(err, errEmpresaUsuarioScopeRequired) {
		t.Fatalf("password lookup without empresa_id must fail with scope error, got %v", err)
	}
	if _, err := resolveEmpresaUsuarioForPasswordReset(nil, "qa@example.test", "token", 0); !errors.Is(err, errEmpresaUsuarioScopeRequired) {
		t.Fatalf("reset lookup without empresa_id must fail with scope error, got %v", err)
	}
}

func TestRequireEmpresaUsuarioScopeRejectsMissingAndMismatchedTenant(t *testing.T) {
	missing := httptest.NewRequest(http.MethodPost, "https://service.test/api/empresa/usuarios/login", nil)
	if _, err := requireEmpresaUsuarioScope(missing, 0); !errors.Is(err, errEmpresaUsuarioScopeRequired) {
		t.Fatalf("missing empresa_id must fail closed, got %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "https://service.test/api/empresa/usuarios/login?empresa_id=12", nil)
	req = requestWithTenantContext(req, TenantContext{EmpresaID: 12, Module: "publico"})
	if got, err := requireEmpresaUsuarioScope(req, 12); err != nil || got != 12 {
		t.Fatalf("matching validated tenant must be accepted, got id=%d err=%v", got, err)
	}
	if _, err := requireEmpresaUsuarioScope(req, 13); err == nil {
		t.Fatal("mismatched body tenant must be rejected")
	}
}

func TestWithEmpresaPublicScopeRejectsRequestWithoutEmpresaID(t *testing.T) {
	called := false
	h := WithEmpresaPublicScope(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "https://service.test/api/empresa/usuarios/login", strings.NewReader(`{"email":"qa@example.test"}`)))
	if called {
		t.Fatal("public company-user handler must not run without empresa_id")
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("missing empresa_id status=%d want=%d", rec.Code, http.StatusBadRequest)
	}
}
