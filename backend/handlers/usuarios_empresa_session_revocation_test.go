package handlers

import (
	"os"
	"strings"
	"testing"
)

// Este contrato evita que una refactorizacion del endpoint de activacion deje
// cuentas inactivas con sesiones anteriores todavia utilizables. La prueba E2E
// de staging cubre ademas el efecto real contra PostgreSQL.
func TestEmpresaUsuarioDeactivationRevokesExistingSessions(t *testing.T) {
	source, err := os.ReadFile("usuarios_empresa.go")
	if err != nil {
		t.Fatalf("read usuarios_empresa.go: %v", err)
	}
	text := string(source)
	revoke := strings.Index(text, "RevokeSessionsByAdminEmail(dbSuper, item.Email)")
	update := strings.Index(text, "SetEmpresaUsuarioEstado(dbEmp, empresaID, id, estado)")
	if revoke < 0 {
		t.Fatal("deactivation must revoke existing sessions")
	}
	if update < 0 || revoke > update {
		t.Fatal("sessions must be revoked before persisting the inactive state")
	}
}
