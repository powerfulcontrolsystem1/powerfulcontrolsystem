package db

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestAuthTokenExpirationFailsClosed(t *testing.T) {
	for _, invalid := range []string{"", "sin-fecha", "2026/08/21"} {
		if _, ok := parseAuthTokenExpiration(invalid); ok {
			t.Fatalf("invalid expiration %q was accepted", invalid)
		}
	}
	if parsed, ok := parseAuthTokenExpiration("2026-08-21 12:30:00"); !ok || parsed.IsZero() {
		t.Fatal("valid login token expiration was rejected")
	}
	if parsed, ok := parseAuthTokenExpiration(time.Date(2026, 8, 21, 12, 30, 0, 0, time.UTC).Format(time.RFC3339)); !ok || parsed.IsZero() {
		t.Fatal("valid RFC3339 expiration was rejected")
	}
}

func TestAuthTokensAreConsumedConditionally(t *testing.T) {
	adminSource, err := os.ReadFile("db.go")
	if err != nil {
		t.Fatal(err)
	}
	companySource, err := os.ReadFile("usuarios_empresa.go")
	if err != nil {
		t.Fatal(err)
	}
	admin := string(adminSource)
	company := string(companySource)
	for name, source := range map[string]string{
		"admin confirmation":   admin,
		"company confirmation": company,
	} {
		if !strings.Contains(source, "FOR UPDATE") || !strings.Contains(source, "email_confirm_token = ?") {
			t.Fatalf("%s token is not locked and conditionally consumed", name)
		}
	}
	if !strings.Contains(admin, "SetAdministradorPasswordFromResetToken") || !strings.Contains(admin, "password_reset_token = ?") {
		t.Fatal("administrator password reset token must be consumed by the password update")
	}
	if !strings.Contains(company, "SetEmpresaUsuarioPasswordFromResetToken") || !strings.Contains(company, "password_reset_token = ?") {
		t.Fatal("enterprise-user password reset token must be consumed by the password update")
	}
}
