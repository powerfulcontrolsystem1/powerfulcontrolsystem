package db

import (
	"strings"
	"testing"
)

func TestOfflineCashReportMigrationIsRegistered(t *testing.T) {
	migrations, err := PlatformMigrations(MigrationTargetEmpresas)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, migration := range migrations {
		if migration.Version != "20260906-004-offline-cash-report-v1" {
			continue
		}
		found = true
		if migration.Body != empresaOfflineCashReportSchemaFingerprint || migration.Apply == nil {
			t.Fatalf("invalid offline cash report migration: %#v", migration)
		}
	}
	if !found {
		t.Fatal("offline cash report migration is not registered")
	}
}

func TestOfflineCashReportMigrationKeepsTenantCashOperationIdentity(t *testing.T) {
	raw := empresaOfflineCashReportSchemaFingerprint
	for _, required := range []string{"offline", "cash", "operation"} {
		if !strings.Contains(raw, required) {
			t.Fatalf("migration fingerprint is missing %q", required)
		}
	}
}
