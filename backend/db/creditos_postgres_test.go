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

func TestCreditoCarteraResumenCoalescesEmptyAggregateCounts(t *testing.T) {
	raw, err := os.ReadFile("creditos.go")
	if err != nil {
		t.Fatalf("read creditos.go: %v", err)
	}
	body := extractCreditoFunctionForTest(t, string(raw), "func GetEmpresaCreditosCarteraResumen(", "func GetEmpresaCreditosMoraDashboard(")
	if got := strings.Count(body, "COALESCE(SUM(CASE"); got != 4 {
		t.Fatalf("el resumen debe convertir a cero sus cuatro conteos SUM(CASE) cuando no hay creditos, got=%d: %s", got, body)
	}
	if !strings.Contains(body, "func GetEmpresaCreditosCarteraResumenContext(ctx context.Context") || !strings.Contains(body, "queryRowSQLCompatContext(ctx") {
		t.Fatalf("el resumen debe conservar el contexto del handler hasta PostgreSQL: %s", body)
	}
}

func TestCreditoDisponibilidadClientePreservesRequestContext(t *testing.T) {
	raw, err := os.ReadFile("creditos.go")
	if err != nil {
		t.Fatalf("read creditos.go: %v", err)
	}
	source := string(raw)
	body := extractCreditoFunctionForTest(t, source, "func GetEmpresaCreditoClienteDisponibilidad(", "func GetEmpresaCreditoByVentaOrigenID(")
	if !strings.Contains(source, "func GetEmpresaCreditoClienteLimiteContext(ctx context.Context") {
		t.Fatalf("el limite de credito debe ofrecer variante cancelable: %s", source)
	}
	for _, required := range []string{
		"func GetEmpresaCreditoClienteDisponibilidadContext(ctx context.Context",
		"GetEmpresaCreditoClienteLimiteContext(ctx",
		"queryRowSQLCompatContext(ctx",
	} {
		if !strings.Contains(body, required) {
			t.Fatalf("la disponibilidad de credito debe preservar contexto en %q: %s", required, body)
		}
	}
}

func TestCreditoListadosPreserveRequestContext(t *testing.T) {
	raw, err := os.ReadFile("creditos.go")
	if err != nil {
		t.Fatalf("read creditos.go: %v", err)
	}
	source := string(raw)
	for _, required := range []string{
		"func ListEmpresaCreditoCuotasContext(ctx context.Context",
		"func ListEmpresaCreditoMovimientosContext(ctx context.Context",
		"querySQLCompatContext(ctx, dbConn, query, args...)",
	} {
		if !strings.Contains(source, required) {
			t.Fatalf("los listados de credito deben conservar contexto en %q", required)
		}
	}
}

func TestCreditoDetallePreservesRequestContextDuringHydration(t *testing.T) {
	raw, err := os.ReadFile("creditos.go")
	if err != nil {
		t.Fatalf("read creditos.go: %v", err)
	}
	source := string(raw)
	for _, required := range []string{
		"func creditoHydrateCuotaStatusContext(ctx context.Context",
		"func GetEmpresaCreditoByIDContext(ctx context.Context",
		"queryRowSQLCompatContext(ctx, dbConn",
		"creditoHydrateCuotaStatusContext(ctx, dbConn, empresaID, credito)",
	} {
		if !strings.Contains(source, required) {
			t.Fatalf("el detalle de credito debe conservar contexto en %q", required)
		}
	}
}

func TestCreditoListadoPreservesRequestContextDuringHydration(t *testing.T) {
	raw, err := os.ReadFile("creditos.go")
	if err != nil {
		t.Fatalf("read creditos.go: %v", err)
	}
	source := string(raw)
	for _, required := range []string{
		"func ListEmpresaCreditosContext(ctx context.Context",
		"queryRowSQLCompatContext(ctx, dbConn, countQuery",
		"querySQLCompatContext(ctx, dbConn, query, args...)",
		"creditoHydrateCuotaStatusRowsContext(ctx, dbConn, empresaID, out)",
	} {
		if !strings.Contains(source, required) {
			t.Fatalf("el listado de credito debe conservar contexto en %q", required)
		}
	}
}

func TestCreditoMoraDashboardPreservesRequestContext(t *testing.T) {
	raw, err := os.ReadFile("creditos.go")
	if err != nil {
		t.Fatalf("read creditos.go: %v", err)
	}
	source := string(raw)
	for _, required := range []string{
		"func listEmpresaCreditosByWhereContext(ctx context.Context",
		"func GetEmpresaCreditosMoraDashboardContext(ctx context.Context",
		"listEmpresaCreditosByWhereContext(ctx,",
		"queryRowSQLCompatContext(ctx, dbConn, countQuery",
	} {
		if !strings.Contains(source, required) {
			t.Fatalf("el tablero de mora debe conservar contexto en %q", required)
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
