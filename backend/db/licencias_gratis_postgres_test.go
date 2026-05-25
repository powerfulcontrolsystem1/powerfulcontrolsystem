package db

import (
	"os"
	"strings"
	"testing"
)

func TestLicenciasGratisSchemaUsesPostgresGeneratedID(t *testing.T) {
	body, err := os.ReadFile("licencias_gratis.go")
	if err != nil {
		t.Fatalf("read licencias_gratis.go: %v", err)
	}
	src := string(body)
	if !strings.Contains(src, "id BIGSERIAL PRIMARY KEY") {
		t.Fatal("licencias_activaciones_gratis debe crear id con BIGSERIAL PRIMARY KEY en PostgreSQL")
	}
	if !strings.Contains(src, "EnsurePostgresPrimaryKeySequences(dbConn)") {
		t.Fatal("licencias_activaciones_gratis debe reparar secuencias en tablas existentes sin default")
	}
}
