package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

func TestDecorateEmpresasByEffectiveAccessIncludesPrincipalOwnedForDelegatedAdmin(t *testing.T) {
	t.Parallel()

	empresas := []dbpkg.Empresa{
		{ID: 10, EmpresaID: 10, Nombre: "Principal A", UsuarioCreador: "principal@example.com"},
		{ID: 11, EmpresaID: 11, Nombre: "Otra", UsuarioCreador: "otra@example.com"},
	}

	got, err := decorateEmpresasByEffectiveAccess(nil, "delegado@example.com", "principal@example.com", empresas)
	if err != nil {
		t.Fatalf("decorateEmpresasByEffectiveAccess returned error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected only principal-owned empresa, got %d", len(got))
	}
	if got[0].EmpresaID != 10 {
		t.Fatalf("expected empresa 10, got %d", got[0].EmpresaID)
	}
	if got[0].AccessSource != "delegated" {
		t.Fatalf("expected delegated access source, got %q", got[0].AccessSource)
	}
}

func TestFilterAdministradoresForPrincipalScopeExcludesPrincipal(t *testing.T) {
	t.Parallel()

	admins := []dbpkg.Admin{
		{Email: "principal@example.com", UsuarioCreador: ""},
		{Email: "delegado@example.com", UsuarioCreador: "principal@example.com", EmailConfirmado: 0},
	}

	got, err := filterAdministradoresForPrincipalScope(nil, "principal@example.com", admins)
	if err != nil {
		t.Fatalf("filterAdministradoresForPrincipalScope returned error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected only delegated admin, got %d", len(got))
	}
	if got[0].Email != "delegado@example.com" {
		t.Fatalf("expected delegated admin, got %q", got[0].Email)
	}
	if got[0].InvitationStatus != "pendiente" {
		t.Fatalf("expected pending invitation status, got %q", got[0].InvitationStatus)
	}
}

func TestAdministradoresEffectivePrincipalScopeForScopedSuperRequest(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/super/api/administradores?scope=principal", nil)
	admin := &dbpkg.Admin{Email: "principal@example.com", Role: "super_administrador"}

	got := administradoresEffectivePrincipalScope(req, admin, "")
	if got != "principal@example.com" {
		t.Fatalf("expected scoped principal email, got %q", got)
	}

	globalReq := httptest.NewRequest(http.MethodGet, "/super/api/administradores", nil)
	if got := administradoresEffectivePrincipalScope(globalReq, admin, ""); got != "" {
		t.Fatalf("expected global super scope to remain empty, got %q", got)
	}
}

func TestValidatePendingAdminInvitationToken(t *testing.T) {
	expira := time.Now().Add(time.Hour).Format("2006-01-02 15:04:05")
	admin := &dbpkg.Admin{EmailConfirmToken: "token-ok", EmailConfirmExpira: expira}

	if status, msg := validatePendingAdminInvitationToken(admin, "token-ok", time.Now()); status != http.StatusOK || msg != "" {
		t.Fatalf("expected valid invitation token, got status=%d msg=%q", status, msg)
	}
	if status, _ := validatePendingAdminInvitationToken(admin, "token-malo", time.Now()); status != http.StatusForbidden {
		t.Fatalf("expected forbidden for wrong token, got %d", status)
	}
	if status, _ := validatePendingAdminInvitationToken(admin, "token-ok", time.Now().Add(2*time.Hour)); status != http.StatusGone {
		t.Fatalf("expected gone for expired token, got %d", status)
	}
}
