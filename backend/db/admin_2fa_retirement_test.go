package db

import (
	"os"
	"strings"
	"testing"
)

func TestSuperCatalogIncludesAdmin2FARetirement(t *testing.T) {
	t.Parallel()
	migrations, err := PlatformMigrations(MigrationTargetSuper)
	if err != nil {
		t.Fatal(err)
	}
	for _, migration := range migrations {
		if migration.Version != "20260907-001-retire-admin-2fa-v1" {
			continue
		}
		if migration.Apply == nil || migration.Body != admin2FARetirementFingerprint {
			t.Fatal("administrator 2FA retirement migration must be executable and immutable")
		}
		return
	}
	t.Fatal("administrator 2FA retirement migration is missing from super catalog")
}

func TestAdmin2FARetirementClearsAllUsableMaterial(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile("admin_2fa_retirement.go")
	if err != nil {
		t.Fatal(err)
	}
	source := string(raw)
	for _, marker := range []string{
		"totp_enabled = 0",
		"totp_secret = ''",
		"totp_last_counter = -1",
		"DELETE FROM administrador_totp_recovery_codes",
		"security.admin_2fa.enabled",
		"UPDATE sesiones",
		"activo = 0",
	} {
		if !strings.Contains(source, marker) {
			t.Fatalf("retirement migration is missing %q", marker)
		}
	}
}
