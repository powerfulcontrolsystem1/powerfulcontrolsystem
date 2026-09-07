package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPlatformComposeSeparatesMigrationAndRuntimeDatabaseUsers(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "deploy", "docker-compose.platform.yml"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(raw)
	for _, required := range []string{
		"PCS_RUNTIME_DB_USER: ${PCS_RUNTIME_DB_USER:?Defina PCS_RUNTIME_DB_USER}",
		"PCS_RUNTIME_DB_PASSWORD: ${PCS_RUNTIME_DB_PASSWORD:?Defina PCS_RUNTIME_DB_PASSWORD}",
		"postgres://${PCS_RUNTIME_DB_USER:?Defina PCS_RUNTIME_DB_USER}:${PCS_RUNTIME_DB_PASSWORD:?Defina PCS_RUNTIME_DB_PASSWORD}@postgres:5432/pcs_empresas",
		"postgres://${PCS_RUNTIME_DB_USER:?Defina PCS_RUNTIME_DB_USER}:${PCS_RUNTIME_DB_PASSWORD:?Defina PCS_RUNTIME_DB_PASSWORD}@postgres:5432/pcs_superadministrador",
	} {
		if !strings.Contains(content, required) {
			t.Fatalf("runtime database separation is missing %q", required)
		}
	}
	if got := strings.Count(content, "postgres://${POSTGRES_USER:-pcs}:${POSTGRES_PASSWORD}@postgres:5432/"); got != 2 {
		t.Fatalf("only migrate may use owner DSNs; found %d owner references, want 2", got)
	}
}
