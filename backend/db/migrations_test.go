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

func TestEmpresaCatalogIncludesNextcloudSchemaMigration(t *testing.T) {
	t.Parallel()
	migrations, err := PlatformMigrations(MigrationTargetEmpresas)
	if err != nil {
		t.Fatal(err)
	}
	for _, migration := range migrations {
		if migration.Version == "20260716-002-nextcloud-accounts-v1" {
			if migration.Body != empresaNextcloudSchemaFingerprint || migration.Apply == nil {
				t.Fatal("nextcloud migration must be immutable and executable")
			}
			return
		}
	}
	t.Fatal("nextcloud migration is missing from empresas catalog")
}

func TestSuperCatalogIncludesSystemMetricsMigration(t *testing.T) {
	t.Parallel()
	migrations, err := PlatformMigrations(MigrationTargetSuper)
	if err != nil {
		t.Fatal(err)
	}
	for _, migration := range migrations {
		if migration.Version == "20260716-004-system-metrics-v1" {
			if migration.Body != metricsSchemaFingerprint || migration.Apply == nil {
				t.Fatal("metrics migration must be immutable and executable")
			}
			return
		}
	}
	t.Fatal("system metrics migration is missing from super catalog")
}

func TestPlatformCatalogsFreezeLegacySchemaManifest(t *testing.T) {
	t.Parallel()
	for _, target := range []string{MigrationTargetEmpresas, MigrationTargetSuper} {
		migrations, err := PlatformMigrations(target)
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for _, migration := range migrations {
			if migration.Version != "20260717-001-legacy-schema-manifest-v1" {
				continue
			}
			if migration.Apply != nil || migration.Body == "" {
				t.Fatalf("legacy manifest migration for %s must be immutable metadata only", target)
			}
			found = true
		}
		if !found {
			t.Fatalf("legacy manifest migration is missing for %s", target)
		}
	}
}

func TestSuperCatalogIncludesSessionTokenMigration(t *testing.T) {
	t.Parallel()
	migrations, err := PlatformMigrations(MigrationTargetSuper)
	if err != nil {
		t.Fatal(err)
	}
	for _, migration := range migrations {
		if migration.Version == "20260722-001-session-token-hashes-v1" {
			if migration.Apply == nil || migration.Body == "" {
				t.Fatal("session token migration must be executable and checksummed")
			}
			return
		}
	}
	t.Fatal("session token migration is missing from super catalog")
}

func TestLegacySchemaManifestV1KeepsReleasedChecksums(t *testing.T) {
	t.Parallel()
	expected := map[string]string{
		MigrationTargetEmpresas: "0ca48d62a466acdcf0484dbc41c788a6fd3767056c1d96759b0d47ba4ad52603",
		MigrationTargetSuper:    "39556fb4724b5f82a4eb5aa2450b101d87a4137b31b2b92df70feb372f11ae14",
	}
	for target, want := range expected {
		migrations, err := PlatformMigrations(target)
		if err != nil {
			t.Fatalf("PlatformMigrations(%s): %v", target, err)
		}
		found := false
		for _, migration := range migrations {
			if migration.Version == "20260717-001-legacy-schema-manifest-v1" {
				found = true
				if got := MigrationChecksum(platformMigrationScope, migration); got != want {
					t.Fatalf("legacy manifest checksum for %s = %s, want %s", target, got, want)
				}
				break
			}
		}
		if !found {
			t.Fatalf("legacy manifest migration is missing for %s", target)
		}
	}
}
