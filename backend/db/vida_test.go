package db

import (
	"strings"
	"testing"
)

func TestEmpresaVidaSchemaUsesTenantAndUserIsolation(t *testing.T) {
	joined := strings.ToLower(strings.Join(empresaVidaSchemaStatements(), "\n"))
	for _, required := range []string{
		"empresa_vida_gastos", "empresa_vida_suscripciones",
		"empresa_id", "usuario_id", "numeric(18,2)",
		"ux_empresa_vida_gastos_idempotencia", "ux_empresa_vida_suscripciones_idempotencia",
	} {
		if !strings.Contains(joined, required) {
			t.Fatalf("vida schema is missing %q", required)
		}
	}
	if !strings.Contains(joined, "(empresa_id, usuario_id, client_request_id)") {
		t.Fatal("vida idempotency must be scoped by tenant and authenticated user")
	}
}

func TestEmpresaVidaPriceHistorySchemaUsesTenantUserAndExpense(t *testing.T) {
	joined := strings.ToLower(strings.Join(empresaVidaPriceHistorySchemaStatements(), "\n"))
	for _, required := range []string{
		"empresa_vida_precios", "empresa_id", "usuario_id", "gasto_id",
		"codigo_barras", "precio_unitario", "ia_factura", "on delete cascade",
		"ix_empresa_vida_precios_usuario_codigo",
	} {
		if !strings.Contains(joined, required) {
			t.Fatalf("Vida price history schema is missing %q", required)
		}
	}
	if strings.Contains(joined, "producto_id") {
		t.Fatal("Vida personal price history must not depend on company inventory products")
	}
}

func TestEmpresaVidaSubscriptionProjection(t *testing.T) {
	tests := []struct {
		period string
		cost   float64
		month  float64
		year   float64
	}{
		{period: "mensual", cost: 20, month: 20, year: 240},
		{period: "trimestral", cost: 90, month: 30, year: 360},
		{period: "semestral", cost: 120, month: 20, year: 240},
		{period: "anual", cost: 120, month: 10, year: 120},
	}
	for _, tt := range tests {
		month, year := EmpresaVidaSubscriptionProjection(tt.cost, tt.period, 1)
		if month != tt.month || year != tt.year {
			t.Fatalf("%s projection = %.2f/%.2f, want %.2f/%.2f", tt.period, month, year, tt.month, tt.year)
		}
	}
}

func TestEmpresaVidaReportsAndReminderSchemaStayPersonalAndIdempotent(t *testing.T) {
	joined := strings.ToLower(strings.Join(empresaVidaReportsNotificationsSchemaStatements(), "\n"))
	for _, required := range []string{"empresa_vida_notificacion_configuracion", "empresa_vida_notificaciones", "empresa_id", "usuario_id", "suscripcion_id", "email_activa", "whatsapp_activa", "unique (empresa_id, usuario_id, suscripcion_id, proxima_renovacion, canal)"} {
		if !strings.Contains(joined, required) {
			t.Fatalf("Vida reports/reminders schema is missing %q", required)
		}
	}
	if strings.Contains(joined, "inventario") || strings.Contains(joined, "contabilidad") {
		t.Fatal("Vida reminders must remain independent from business modules")
	}
}

func TestNormalizeVidaReportDatesBoundsPeriod(t *testing.T) {
	if _, _, err := normalizeVidaReportDates("2026-09-05", "2026-09-04"); err == nil {
		t.Fatal("reversed report range was accepted")
	}
	if _, _, err := normalizeVidaReportDates("2025-01-01", "2026-09-05"); err == nil {
		t.Fatal("unbounded report range was accepted")
	}
	if from, to, err := normalizeVidaReportDates("2026-08-01", "2026-09-05"); err != nil || from != "2026-08-01" || to != "2026-09-05" {
		t.Fatalf("valid report range failed: %q %q %v", from, to, err)
	}
}
