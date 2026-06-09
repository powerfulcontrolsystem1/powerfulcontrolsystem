package db

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestCreateEmpresaCreditoUsesPostgresCompatibleWrites(t *testing.T) {
	raw, err := os.ReadFile("creditos.go")
	if err != nil {
		t.Fatalf("read creditos.go: %v", err)
	}
	src := string(raw)
	body := extractCreditoFunctionForTest(t, src, "func CreateEmpresaCredito(", "func creditoGenerateCuotasTx(")

	if strings.Contains(body, "tx.Exec(") {
		t.Fatalf("CreateEmpresaCredito debe usar helpers SQL compatibles con PostgreSQL, no tx.Exec directo: %s", body)
	}
	if strings.Contains(body, "pcs_ts(") {
		t.Fatalf("CreateEmpresaCredito no debe usar pcs_ts() en runtime PostgreSQL: %s", body)
	}
	if !strings.Contains(body, "insertTxSQLCompat") || !strings.Contains(body, "execTxSQLCompat") {
		t.Fatalf("CreateEmpresaCredito debe rebindear INSERT/UPDATE transaccionales con helpers SQL compat: %s", body)
	}
	if !strings.Contains(body, "sqlNowExpr()") {
		t.Fatalf("CreateEmpresaCredito debe usar sqlNowExpr() para fechas runtime: %s", body)
	}
}

func TestCreditoGenerateCuotasTxUsesPostgresCompatibleWrites(t *testing.T) {
	raw, err := os.ReadFile("creditos.go")
	if err != nil {
		t.Fatalf("read creditos.go: %v", err)
	}
	src := string(raw)
	body := extractCreditoFunctionForTest(t, src, "func creditoGenerateCuotasTxWithStart(", "func scanEmpresaCredito(")

	if strings.Contains(body, "tx.Exec(") {
		t.Fatalf("creditoGenerateCuotasTxWithStart debe usar helpers SQL compatibles con PostgreSQL, no tx.Exec directo: %s", body)
	}
	if strings.Contains(body, "pcs_ts(") {
		t.Fatalf("creditoGenerateCuotasTxWithStart no debe usar pcs_ts() en runtime PostgreSQL: %s", body)
	}
	if !strings.Contains(body, "execTxSQLCompat") {
		t.Fatalf("creditoGenerateCuotasTxWithStart debe rebindear INSERT transaccional con execTxSQLCompat: %s", body)
	}
	if !strings.Contains(body, "sqlNowExpr()") {
		t.Fatalf("creditoGenerateCuotasTxWithStart debe usar sqlNowExpr() para fechas runtime: %s", body)
	}
}

func TestCreditoDashboardsUsePostgresSafeDateComparisons(t *testing.T) {
	raw, err := os.ReadFile("creditos.go")
	if err != nil {
		t.Fatalf("read creditos.go: %v", err)
	}
	src := string(raw)
	cases := []struct {
		name  string
		start string
		end   string
	}{
		{"hydrate cuotas", "func creditoHydrateCuotaStatus(", "func creditoHydrateCuotaStatusRows("},
		{"filtros creditos", "func creditoBuildWhere(", "// ListEmpresaCreditos lista creditos"},
		{"resumen cartera", "func GetEmpresaCreditosCarteraResumen(", "func GetEmpresaCreditosMoraDashboard("},
		{"dashboard mora", "func GetEmpresaCreditosMoraDashboard(", "func scanEmpresaCreditoWorkflow("},
	}
	for _, tc := range cases {
		body := extractCreditoFunctionForTest(t, src, tc.start, tc.end)
		for _, forbidden := range []string{"pcs_ts(", "date('now'", `date("now"`, "pcs_julian_day("} {
			if strings.Contains(body, forbidden) {
				t.Fatalf("%s no debe usar %s en consultas runtime PostgreSQL: %s", tc.name, forbidden, body)
			}
		}
		if !strings.Contains(body, "time.Now().In(time.Local).Format(\"2006-01-02\")") {
			t.Fatalf("%s debe calcular la fecha actual desde Go y pasarla como parametro SQL: %s", tc.name, body)
		}
	}
}

func TestCreditoDailyScheduleSupportsLongContractsAndSkipsSundays(t *testing.T) {
	if got := creditoMaxCuotas("diaria"); got < 730 {
		t.Fatalf("los creditos diarios deben permitir contratos de al menos dos años, got=%d", got)
	}
	inicio := time.Date(2026, time.May, 23, 0, 0, 0, 0, time.Local)
	primera := creditoNextFechaCuota(inicio, "diaria", 1, true)
	if primera.Weekday() == time.Sunday {
		t.Fatalf("la primera cuota no debe caer domingo cuando se omiten domingos: %s", primera.Format("2006-01-02"))
	}
	if primera.Format("2006-01-02") != "2026-05-25" {
		t.Fatalf("se esperaba saltar domingo 2026-05-24 y vencer 2026-05-25, got=%s", primera.Format("2006-01-02"))
	}
}

func extractCreditoFunctionForTest(t *testing.T, src, startMarker, endMarker string) string {
	t.Helper()

	start := strings.Index(src, startMarker)
	if start < 0 {
		t.Fatalf("no se encontro %s", startMarker)
	}
	end := strings.Index(src[start:], endMarker)
	if end < 0 {
		t.Fatalf("no se encontro limite %s para %s", endMarker, startMarker)
	}
	return src[start : start+end]
}
