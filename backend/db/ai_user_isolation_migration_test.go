package db

import "testing"

func TestEmpresaAIUserIsolationMigrationIsRegistered(t *testing.T) {
	migrations, err := PlatformMigrations(MigrationTargetEmpresas)
	if err != nil {
		t.Fatal(err)
	}
	for _, migration := range migrations {
		if migration.Version == "20260725-001-ai-user-isolation-v1" {
			if migration.Body != empresaAIUserIsolationSchemaFingerprint || migration.Apply == nil {
				t.Fatal("migracion de aislamiento IA incompleta")
			}
			return
		}
	}
	t.Fatal("migracion de aislamiento IA no registrada")
}
