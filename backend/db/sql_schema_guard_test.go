package db

import "testing"

func TestSchemaBootstrapDisabledOnlyForProductionNonMigrator(t *testing.T) {
	t.Setenv("PCS_ENV", "production")
	t.Setenv("PCS_RUNTIME_ROLE", "api")
	t.Setenv("PCS_RUNTIME_SCHEMA_BOOTSTRAP", "0")
	if !SchemaBootstrapDisabled() {
		t.Fatal("production API must not execute legacy schema bootstrap")
	}
	if err := EnsureEmpresaCarritosSchema(nil); err != nil {
		t.Fatalf("legacy schema guard must avoid DDL path: %v", err)
	}
	t.Setenv("PCS_RUNTIME_ROLE", "migrate")
	if SchemaBootstrapDisabled() {
		t.Fatal("migration role must retain schema bootstrap")
	}
}
