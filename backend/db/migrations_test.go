package db

import "testing"

func TestMigrationChecksumIncludesImmutableBody(t *testing.T) {
	t.Parallel()
	base := Migration{Version: "20260716-001", Description: "test", Body: "CREATE TABLE example"}
	changed := base
	changed.Body = "CREATE TABLE changed_example"
	if MigrationChecksum(platformMigrationScope, base) == MigrationChecksum(platformMigrationScope, changed) {
		t.Fatal("migration checksum must change when the migration body changes")
	}
}

func TestValidateMigrationCatalogRejectsInvalidOrderingAndDuplicates(t *testing.T) {
	t.Parallel()
	if err := ValidateMigrationCatalog([]Migration{
		{Version: "20260716-002", Description: "second", Body: "second"},
		{Version: "20260716-001", Description: "first", Body: "first"},
	}); err == nil {
		t.Fatal("expected ordering validation error")
	}
	if err := ValidateMigrationCatalog([]Migration{
		{Version: "20260716-001", Description: "first", Body: "first"},
		{Version: "20260716-001", Description: "duplicate", Body: "duplicate"},
	}); err == nil {
		t.Fatal("expected duplicate validation error")
	}
}

func TestPlatformMigrationCatalogsAreOrderedAndChecksummed(t *testing.T) {
	t.Parallel()
	for _, target := range []string{MigrationTargetEmpresas, MigrationTargetSuper} {
		migrations, err := PlatformMigrations(target)
		if err != nil {
			t.Fatalf("PlatformMigrations(%s): %v", target, err)
		}
		if err := ValidateMigrationCatalog(migrations); err != nil {
			t.Fatalf("ValidateMigrationCatalog(%s): %v", target, err)
		}
		for _, migration := range migrations {
			if MigrationChecksum(platformMigrationScope, migration) == "" {
				t.Fatalf("empty checksum for %s/%s", target, migration.Version)
			}
		}
	}
}
