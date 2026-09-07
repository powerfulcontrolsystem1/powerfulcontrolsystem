package handlers

import (
	"encoding/json"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	dbpkg "github.com/you/pos-backend/db"
)

func TestCorteCajaSoloUsuarioCajaActualRecognizesReporteMiTurno(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/empresa/corte_caja?empresa_id=1&action=reporte_mi_turno", nil)
	if !corteCajaSoloUsuarioCajaActual(req) {
		t.Fatalf("expected reporte_mi_turno to use current-user cash close report")
	}

	plainReq := httptest.NewRequest("GET", "/api/empresa/corte_caja?empresa_id=1", nil)
	if corteCajaSoloUsuarioCajaActual(plainReq) {
		t.Fatalf("plain corte_caja request should not be forced to current-user cash report")
	}
}

func TestApplyCorteCajaPersistedCashSummaryIncludesMixedCash(t *testing.T) {
	resumen := corteCajaResumen{
		AperturaEfectivo:     0,
		EfectivoVentas:       0,
		IngresosEfectivo:     0,
		EgresosEfectivo:      0,
		OtrosMediosVentas:    100,
		EfectivoEsperadoCaja: 0,
	}
	cierre := dbpkg.EmpresaCierreCaja{
		AperturaMonto:    0,
		IngresosEfectivo: 50,
		EgresosEfectivo:  0,
		RetirosEfectivo:  0,
	}

	applyCorteCajaPersistedCashSummary(&resumen, &cierre)

	if resumen.EfectivoVentas != 50 {
		t.Fatalf("expected mixed cash portion 50, got %.2f", resumen.EfectivoVentas)
	}
	if resumen.EfectivoEsperadoCaja != 50 {
		t.Fatalf("expected cash drawer 50, got %.2f", resumen.EfectivoEsperadoCaja)
	}
	if resumen.OtrosMediosVentas != 50 {
		t.Fatalf("expected remaining non-cash mixed portion 50, got %.2f", resumen.OtrosMediosVentas)
	}
}

func TestBuildCorteCajaDetallesVentasKeepsReusableCartSalesIndependent(t *testing.T) {
	detailOne, _ := json.Marshal([]corteCajaVentaItemSnapshot{{Tipo: "producto", Referencia: "QA-1", Producto: "Producto uno", Cantidad: 1, Total: 119}})
	detailTwo, _ := json.Marshal([]corteCajaVentaItemSnapshot{{Tipo: "servicio", Referencia: "QA-2", Producto: "Servicio dos", Cantidad: 2, Total: 200}})
	ventas := []corteCajaVenta{
		{Codigo: "VENTA-DIRECTA-12-CRT-117-PG-1", FechaPago: "2026-08-26 00:21:51", DetalleItems: string(detailOne)},
		{Codigo: "VENTA-DIRECTA-12-CRT-117-PG-2", FechaPago: "2026-08-26 00:25:09", DetalleItems: string(detailTwo)},
	}

	items, productos := buildCorteCajaDetallesVentas(ventas, 50)
	if len(items) != 2 {
		t.Fatalf("items por tipo=%d, want 2", len(items))
	}
	if len(productos) != 2 {
		t.Fatalf("productos vendidos=%d, want 2", len(productos))
	}
	if productos[0].VentaCodigo != ventas[1].Codigo || productos[1].VentaCodigo != ventas[0].Codigo {
		t.Fatalf("products must preserve each immutable sale code: %#v", productos)
	}
	if productos[0].Producto != "Servicio dos" || productos[1].Producto != "Producto uno" {
		t.Fatalf("products must be ordered from newest sale: %#v", productos)
	}
}

func TestCorteCajaRangeCoversFullTurnRejectsPartialCalendarRange(t *testing.T) {
	cierre := &dbpkg.EmpresaCierreCaja{
		FechaApertura: "2026-06-17 09:00:00-05:00",
		FechaCierre:   "",
	}
	if corteCajaRangeCoversFullTurn("2026-08-26 00:00:00", "2026-08-26 23:59:59", cierre) {
		t.Fatalf("calendar-day report must not inherit cash accumulated since an older open shift")
	}
	if !corteCajaRangeCoversFullTurn("2026-06-17 09:00:00", "2026-08-26 23:59:59", cierre) {
		t.Fatalf("report starting at the shift opening must cover the full open turn")
	}
}

func TestCorteCajaVentasReadsImmutableTenantScopedHistory(t *testing.T) {
	raw, err := os.ReadFile("corte_caja.go")
	if err != nil {
		t.Fatalf("read corte_caja.go: %v", err)
	}
	source := string(raw)
	start := strings.Index(source, "func listCorteCajaVentas(")
	end := strings.Index(source, "func listCorteCajaVentasAnuladas(")
	if start < 0 || end <= start {
		t.Fatalf("could not isolate listCorteCajaVentas")
	}
	body := source[start:end]
	for _, required := range []string{
		"FROM empresa_ventas_estacion_metricas m",
		"WHERE m.empresa_id = ?",
		"m.cierre_caja_id",
		"m.caja_codigo",
		"m.detalle_items_json",
	} {
		if !strings.Contains(body, required) {
			t.Errorf("immutable cash report query is missing %q", required)
		}
	}
	if strings.Contains(body, "FROM carritos_compras c") {
		t.Fatalf("cash report must not depend on the mutable reusable cart")
	}
}

func TestCorteCajaOfflineReportUsesImmutableOperationAndCashScope(t *testing.T) {
	raw, err := os.ReadFile("corte_caja.go")
	if err != nil {
		t.Fatalf("read corte_caja.go: %v", err)
	}
	source := string(raw)
	start := strings.Index(source, "func listCorteCajaVentasOffline(")
	end := strings.Index(source, "func listCorteCajaVentasAnuladas(")
	if start < 0 || end <= start {
		t.Fatal("could not isolate offline cash report query")
	}
	body := source[start:end]
	for _, required := range []string{
		"FROM empresa_ventas_offline_sync o",
		"WHERE o.empresa_id = ?",
		"o.operacion_codigo",
		"o.cierre_caja_id",
		"o.caja_codigo",
		"o.usuario_cajero",
	} {
		if !strings.Contains(body, required) {
			t.Errorf("offline cash report query is missing %q", required)
		}
	}
	if strings.Contains(body, "JOIN empresa_ventas_estacion_metricas") {
		t.Fatal("offline report must not infer sale identity from a reusable carrito_id")
	}
}

func TestNormalizeCorteCajaReportesAcceptsOfflineSelection(t *testing.T) {
	got := normalizeCorteCajaReportes([]string{"ventas", "ventas_offline"})
	if strings.Join(got, ",") != "ventas,offline" {
		t.Fatalf("unexpected report selection: %#v", got)
	}
}
