package db

import "testing"

func TestEmpresaAIUsoDiarioUniqueMigrationIsRegistered(t *testing.T) {
	migrations, err := PlatformMigrations(MigrationTargetEmpresas)
	if err != nil {
		t.Fatalf("PlatformMigrations(empresas): %v", err)
	}
	for _, migration := range migrations {
		if migration.Version != "20260731-001-ai-usage-unique-v1" {
			continue
		}
		if migration.Body != empresaAIUsoDiarioUniqueSchemaFingerprint || migration.Apply == nil {
			t.Fatal("AI daily usage uniqueness migration must be executable and checksummed")
		}
		return
	}
	t.Fatal("AI daily usage uniqueness migration is missing from enterprise catalog")
}
