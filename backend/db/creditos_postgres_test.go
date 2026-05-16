package db

import (
	"os"
	"strings"
	"testing"
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
	if strings.Contains(body, "datetime(") {
		t.Fatalf("CreateEmpresaCredito no debe usar datetime() en runtime PostgreSQL: %s", body)
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
	if strings.Contains(body, "datetime(") {
		t.Fatalf("creditoGenerateCuotasTxWithStart no debe usar datetime() en runtime PostgreSQL: %s", body)
	}
	if !strings.Contains(body, "execTxSQLCompat") {
		t.Fatalf("creditoGenerateCuotasTxWithStart debe rebindear INSERT transaccional con execTxSQLCompat: %s", body)
	}
	if !strings.Contains(body, "sqlNowExpr()") {
		t.Fatalf("creditoGenerateCuotasTxWithStart debe usar sqlNowExpr() para fechas runtime: %s", body)
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
