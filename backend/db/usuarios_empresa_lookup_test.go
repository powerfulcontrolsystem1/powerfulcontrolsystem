package db

import (
	"os"
	"strings"
	"testing"
)

func TestEmpresaUsuarioEmailLookupPrefiereCuentaConfirmadaActiva(t *testing.T) {
	src, err := os.ReadFile("usuarios_empresa.go")
	if err != nil {
		t.Fatalf("no se pudo leer usuarios_empresa.go: %v", err)
	}
	body := string(src)
	required := []string{
		"CASE WHEN COALESCE(estado, 'activo') = 'activo' THEN 0 ELSE 1 END",
		"CASE WHEN COALESCE(email_confirmado, 0) = 1 THEN 0 ELSE 1 END",
		"CASE WHEN COALESCE(password_set, 0) = 1 AND COALESCE(password_hash, '') <> '' THEN 0 ELSE 1 END",
		"id DESC",
	}
	for _, fragment := range required {
		if !strings.Contains(body, fragment) {
			t.Fatalf("lookup de usuario por email debe priorizar cuentas validas; falta fragmento: %s", fragment)
		}
	}
}
