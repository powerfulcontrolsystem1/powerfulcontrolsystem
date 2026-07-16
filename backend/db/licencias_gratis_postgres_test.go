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
	src := strings.ReplaceAll(string(body), "\r\n", "\n")
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
	src := strings.ReplaceAll(string(body), "\r\n", "\n")
	insertCount := strings.Count(src, "INSERT INTO licencias_activaciones_gratis (licencia_id, empresa_id")
	if insertCount != 1 {
		t.Fatalf("la activacion gratis debe registrar una sola fila para no abortar la transaccion en PostgreSQL; inserts=%d", insertCount)
	}
	if !strings.Contains(src, "assignedLicenciaID = licenciaID") {
		t.Fatal("la activacion gratis debe usar la licencia asignada real y conservar fallback a la licencia solicitada")
	}
}

func TestLicenciaGratisHistoryBlocksExpiredOrInactiveTrial(t *testing.T) {
	body, err := os.ReadFile("licencias_gratis.go")
	if err != nil {
		t.Fatalf("read licencias_gratis.go: %v", err)
	}
	src := strings.ReplaceAll(string(body), "\r\n", "\n")

	historyFn := extractFunctionForTest(t, src, "func HasAnyLicenciaGratisActivationForEmpresa")
	if strings.Contains(historyFn, "COALESCE(estado, 'activo') = 'activo'") {
		t.Fatal("el historial de prueba gratis no debe depender del estado activo; una prueba vencida o historica tambien bloquea")
	}
	if strings.Contains(historyFn, "COALESCE(activo, 1) = 1") {
		t.Fatal("el fallback por licencias antiguas no debe depender de activo=1; una prueba vencida o inactiva tambien bloquea")
	}
	for _, expected := range []string{"duracion_dias, 0) = 15", "LIKE '%prueba%'", "LIKE '%gratis%'", "LIKE '%trial%'"} {
		if !strings.Contains(historyFn, expected) {
			t.Fatalf("el fallback historico debe reconocer licencias de prueba antiguas por %q", expected)
		}
	}

	activateFn := extractFunctionForTest(t, src, "func ActivateLicenciaGratisForEmpresa")
	zeroTrialBranchMarker := "} else {\n\t\tif err := queryRowTxSQLCompat(tx, `SELECT COUNT(1)\n\t\t\tFROM licencias_activaciones_gratis"
	zeroTrialBranchStart := strings.Index(activateFn, zeroTrialBranchMarker)
	if zeroTrialBranchStart < 0 {
		t.Fatal("no se encontro la rama de activacion de prueba gratis sin descuento pago")
	}
	zeroTrialBranch := activateFn[zeroTrialBranchStart:]
	if strings.Contains(zeroTrialBranch, "COALESCE(estado, 'activo') = 'activo'") {
		t.Fatal("la activacion gratis debe consultar historial completo antes de insertar una segunda prueba")
	}
	if strings.Contains(zeroTrialBranch, "COALESCE(activo, 1) = 1") {
		t.Fatal("la activacion gratis debe bloquear pruebas antiguas aunque la licencia asignada ya este inactiva")
	}
}

func extractFunctionForTest(t *testing.T, src, marker string) string {
	t.Helper()
	start := strings.Index(src, marker)
	if start < 0 {
		t.Fatalf("no se encontro %s", marker)
	}
	bodyStart := strings.Index(src[start:], "{")
	if bodyStart < 0 {
		t.Fatalf("no se encontro cuerpo de %s", marker)
	}
	pos := start + bodyStart
	depth := 0
	for i := pos; i < len(src); i++ {
		switch src[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return src[start : i+1]
			}
		}
	}
	t.Fatalf("cuerpo incompleto para %s", marker)
	return ""
}
