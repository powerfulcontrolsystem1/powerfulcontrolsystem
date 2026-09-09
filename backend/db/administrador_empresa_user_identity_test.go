package db

import (
	"os"
	"strings"
	"testing"
)

func TestEmpresaUserIdentityUpsertNeverOverwritesGlobalAdminRole(t *testing.T) {
	t.Parallel()

	raw, err := os.ReadFile("db.go")
	if err != nil {
		t.Fatal(err)
	}
	body := extractFunctionBodyForTest(t, string(raw), "func UpsertAdministradorIdentityForEmpresaUserPreservingRole(")
	upper := strings.ToUpper(body)
	conflictAt := strings.Index(upper, "ON CONFLICT")
	if conflictAt < 0 {
		t.Fatal("enterprise user identity must use an atomic conflict-safe upsert")
	}
	conflictClause := upper[conflictAt:]
	if strings.Contains(conflictClause, "ROLE =") || strings.Contains(conflictClause, "PASSWORD_") || strings.Contains(conflictClause, "EMAIL_CONFIRM") {
		t.Fatal("enterprise user identity conflict path must preserve global role and credentials")
	}
	for _, required := range []string{"NAME =", "PHOTO =", "USUARIO_CREADOR =", "FECHA_ACTUALIZACION ="} {
		if !strings.Contains(conflictClause, required) {
			t.Fatalf("enterprise user identity conflict path missing safe field %q", required)
		}
	}
}

func extractFunctionBodyForTest(t *testing.T, source, signature string) string {
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
