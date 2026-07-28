package db

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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

func TestSchemaStatementLoopsUseRuntimeDDLGuard(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".go" || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		raw, err := os.ReadFile(entry.Name())
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(raw), "dbConn.Exec(stmt)") {
			t.Fatalf("%s bypasses runtime DDL guard for a schema statement loop", entry.Name())
		}
	}
}
