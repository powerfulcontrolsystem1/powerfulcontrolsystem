package db

import (
	"os"
	"strings"
	"testing"
)

func TestLegacyCoreSchemaDefinesMigrationRoots(t *testing.T) {
	raw, err := os.ReadFile("legacy_core_schema.go")
	if err != nil {
		t.Fatal(err)
	}
	source := string(raw)
	for _, required := range []string{
		"func BootstrapLegacyCoreEmpresasSchema(",
		"CREATE TABLE IF NOT EXISTS empresas",
		"func BootstrapLegacyCoreSuperSchema(",
		"CREATE TABLE IF NOT EXISTS administradores",
		"CREATE TABLE IF NOT EXISTS sesiones",
		"CREATE TABLE IF NOT EXISTS configuraciones",
		"CREATE TABLE IF NOT EXISTS tipos_de_empresas",
	} {
		if !strings.Contains(source, required) {
			t.Fatalf("legacy empty-database bootstrap is missing %q", required)
		}
	}
}

func TestMainRunsCoreRootsBeforeLegacyExtensions(t *testing.T) {
	raw, err := os.ReadFile("../main.go")
	if err != nil {
		t.Fatal(err)
	}
	source := string(raw)
	core := strings.Index(source, "BootstrapLegacyCoreSuperSchema")
	auth := strings.Index(source, "EnsureAdministradoresAuthSchema")
	if core < 0 || auth < 0 || core >= auth {
		t.Fatal("super core schema must be created before legacy auth extensions")
	}
}
