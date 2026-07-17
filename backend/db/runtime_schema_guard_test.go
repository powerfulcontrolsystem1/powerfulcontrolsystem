package db

import "testing"

func TestRuntimeDDLGuardBlocksProductionAPI(t *testing.T) {
	t.Setenv("PCS_ENV", "production")
	t.Setenv("PCS_RUNTIME_ROLE", "api")
	if !runtimeDDLBlocked("CREATE TABLE should_not_run (id BIGINT)") {
		t.Fatal("production API must block DDL")
	}
	if runtimeDDLBlocked("UPDATE empresas SET estado = 'activo'") {
		t.Fatal("business DML must remain available")
	}
}

func TestRuntimeDDLGuardAllowsMigrator(t *testing.T) {
	t.Setenv("PCS_ENV", "production")
	t.Setenv("PCS_RUNTIME_ROLE", "migrate")
	if runtimeDDLBlocked("ALTER TABLE empresas ADD COLUMN ejemplo TEXT") {
		t.Fatal("migration role must be allowed to execute DDL")
	}
}
