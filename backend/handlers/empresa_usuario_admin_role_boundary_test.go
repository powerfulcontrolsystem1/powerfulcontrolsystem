package handlers

import (
	"os"
	"strings"
	"testing"
)

func TestEmpresaUsuarioSessionPreservesAdministrativeRoleBoundary(t *testing.T) {
	t.Parallel()

	raw, err := os.ReadFile("usuarios_empresa.go")
	if err != nil {
		t.Fatal(err)
	}
	body := extractEmpresaUsuarioFunctionForTest(t, string(raw), "func createEmpresaUsuarioSession(")
	if !strings.Contains(body, "UpsertAdministradorIdentityForEmpresaUserPreservingRole") {
		t.Fatal("enterprise user session must use the role-preserving identity upsert")
	}
	if strings.Contains(body, "UpsertAdministrador(dbSuper") {
		t.Fatal("enterprise user session must not write its operational role into administradores")
	}
	if !strings.Contains(body, "CreateEmpresaUsuarioSession") {
		t.Fatal("enterprise user session must keep its role in the typed session principal")
	}
}

func extractEmpresaUsuarioFunctionForTest(t *testing.T, source, signature string) string {
	t.Helper()
	start := strings.Index(source, signature)
	if start < 0 {
		t.Fatalf("function %q not found", signature)
	}
	rest := source[start:]
	open := strings.Index(rest, "{")
	if open < 0 {
		t.Fatalf("function %q has no body", signature)
	}
	depth := 0
	for i := open; i < len(rest); i++ {
		switch rest[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return rest[open : i+1]
			}
		}
	}
	t.Fatalf("function %q body is incomplete", signature)
	return ""
}
