package db

import (
	"os"
	"strings"
	"testing"
)

func TestEmpresaConfiguracionAvanzadaUpsertUsesPostgresRuntimeSQL(t *testing.T) {
	raw, err := os.ReadFile("empresa_configuracion_avanzada.go")
	if err != nil {
		t.Fatalf("read empresa_configuracion_avanzada.go: %v", err)
	}
	src := string(raw)
	body := extractEmpresaConfigAvanzadaFunctionForTest(t, src, "func UpsertEmpresaConfiguracionAvanzada(", "")

	if strings.Contains(body, "pcs_ts(") {
		t.Fatalf("UpsertEmpresaConfiguracionAvanzada no debe usar pcs_ts() en runtime PostgreSQL: %s", body)
	}
	if !strings.Contains(body, "ExecCompat") {
		t.Fatalf("UpsertEmpresaConfiguracionAvanzada debe usar ExecCompat para rebind PostgreSQL: %s", body)
	}
	if !strings.Contains(body, "sqlNowExpr()") {
		t.Fatalf("UpsertEmpresaConfiguracionAvanzada debe usar sqlNowExpr() para fechas runtime: %s", body)
	}
	if !strings.Contains(body, "moneda_codigo,\n\t\tsistema_numerico,\n\t\tusar_decimales,\n\t\tcantidad_decimales") {
		t.Fatalf("UpsertEmpresaConfiguracionAvanzada debe insertar moneda, sistema numerico y decimales: %s", body)
	}
	if !strings.Contains(body, "?, ?, ?, ?, `+nowExpr+`, `+nowExpr+`, ?, ?, ?") {
		t.Fatalf("UpsertEmpresaConfiguracionAvanzada debe tener placeholders para usar_decimales y cantidad_decimales antes de fechas: %s", body)
	}
}

func TestEmpresaConfiguracionAvanzadaSchemaNormalizesLegacyBooleanFlags(t *testing.T) {
	raw, err := os.ReadFile("empresa_configuracion_avanzada.go")
	if err != nil {
		t.Fatalf("read empresa_configuracion_avanzada.go: %v", err)
	}
	src := string(raw)
	schemaBody := extractEmpresaConfigAvanzadaFunctionForTest(t, src, "func EnsureEmpresaConfiguracionAvanzadaSchema(", "func ensureEmpresaConfiguracionAvanzadaFlagColumns(")
	normalizerBody := extractEmpresaConfigAvanzadaFunctionForTest(t, src, "func ensureEmpresaConfiguracionAvanzadaFlagColumns(", "func defaultConfigAvanzada(")

	if !strings.Contains(schemaBody, "ensureEmpresaConfiguracionAvanzadaFlagColumns") {
		t.Fatalf("EnsureEmpresaConfiguracionAvanzadaSchema debe normalizar flags legacy antes de operar: %s", schemaBody)
	}
	for _, required := range []string{
		"information_schema.columns",
		"data_type",
		"boolean",
		"ALTER COLUMN %s TYPE INTEGER USING CASE",
		"enviar_email_venta",
		"mostrar_logo_empresa",
		"mostrar_logo_factura",
		"usar_decimales",
	} {
		if !strings.Contains(normalizerBody, required) {
			t.Fatalf("normalizador de flags debe cubrir %s: %s", required, normalizerBody)
		}
	}
}

func extractEmpresaConfigAvanzadaFunctionForTest(t *testing.T, src, startMarker, endMarker string) string {
	t.Helper()

	start := strings.Index(src, startMarker)
	if start < 0 {
		t.Fatalf("no se encontro %s", startMarker)
	}
	if strings.TrimSpace(endMarker) == "" {
		return src[start:]
	}
	end := strings.Index(src[start:], endMarker)
	if end < 0 {
		t.Fatalf("no se encontro limite %s para %s", endMarker, startMarker)
	}
	return src[start : start+end]
}
