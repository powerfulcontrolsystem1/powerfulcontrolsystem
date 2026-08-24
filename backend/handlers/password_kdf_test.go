package handlers

import (
	"strings"
	"testing"
)

func TestEmpresaUsuarioPasswordKDFAndLegacyUpgrade(t *testing.T) {
	const password = "Prueba-Login-2026!"
	const salt = "sal-segura-de-prueba"
	current := currentEmpresaUsuarioPasswordHash(password, salt)
	if !strings.HasPrefix(current, empresaUsuarioPasswordKDFName+"$") {
		t.Fatalf("new password verifier is not versioned: %q", current)
	}
	valid, upgrade := verifyEmpresaUsuarioPasswordHash(password, salt, current)
	if !valid || upgrade {
		t.Fatalf("current verifier result valid=%t upgrade=%t", valid, upgrade)
	}
	legacy := legacyEmpresaUsuarioPasswordHash(password, salt)
	valid, upgrade = verifyEmpresaUsuarioPasswordHash(password, salt, legacy)
	if !valid || !upgrade {
		t.Fatalf("legacy verifier must authenticate once and request upgrade: valid=%t upgrade=%t", valid, upgrade)
	}
	if valid, _ := verifyEmpresaUsuarioPasswordHash("incorrecta", salt, current); valid {
		t.Fatal("incorrect password matched the current verifier")
	}
}
