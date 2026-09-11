package db

import "testing"

func TestComisionesProductosMigrationIsRegistered(t *testing.T) {
	migrations, err := PlatformMigrations(MigrationTargetEmpresas)
	if err != nil {
		t.Fatal(err)
	}
	for _, migration := range migrations {
		if migration.Version == "20260910-001-comisiones-productos-v1" {
			if migration.Body != empresaComisionesProductosSchemaFingerprint || migration.Apply == nil {
				t.Fatalf("invalid commission products migration: %#v", migration)
			}
			return
		}
	}
	t.Fatal("commission products migration is not registered")
}
