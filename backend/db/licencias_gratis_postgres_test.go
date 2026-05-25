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

func TestActivateLicenciaGratisStoresOnlyAssignedActivation(t *testing.T) {
	body, err := os.ReadFile("licencias_gratis.go")
	if err != nil {
		t.Fatalf("read licencias_gratis.go: %v", err)
	}
	src := string(body)
	insertCount := strings.Count(src, "INSERT INTO licencias_activaciones_gratis (licencia_id, empresa_id")
	if insertCount != 1 {
		t.Fatalf("la activacion gratis debe registrar una sola fila para no abortar la transaccion en PostgreSQL; inserts=%d", insertCount)
	}
	if !strings.Contains(src, "assignedLicenciaID = licenciaID") {
		t.Fatal("la activacion gratis debe usar la licencia asignada real y conservar fallback a la licencia solicitada")
	}
}
