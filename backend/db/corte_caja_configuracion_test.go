package db

import "testing"

func TestEmpresaCorteCajaReportesDesdeConfiguracionMapsChecksToSections(t *testing.T) {
	cfg := DefaultEmpresaCorteCajaConfiguracion(7)
	cfg.MostrarMovimientos = false
	cfg.MostrarIngresos = false
	cfg.MostrarEgresos = false
	cfg.MostrarItems = false
	cfg.MostrarSensoresPuertas = false

	got := EmpresaCorteCajaReportesDesdeConfiguracion(&cfg)
	want := map[string]bool{
		"resumen":     true,
		"ventas":      true,
		"anulaciones": true,
		"auditoria":   true,
	}
	if len(got) != len(want) {
		t.Fatalf("reportes inesperados: got %#v want keys %#v", got, want)
	}
	for _, item := range got {
		if !want[item] {
			t.Fatalf("reporte no esperado: %s en %#v", item, got)
		}
	}
}

func TestDefaultEmpresaCorteCajaConfiguracionEnablesTurnoReportFields(t *testing.T) {
	cfg := DefaultEmpresaCorteCajaConfiguracion(7)
	checks := map[string]bool{
		"encabezado":              cfg.MostrarEncabezado,
		"datos_empresa":           cfg.MostrarEmpresaDatos,
		"fecha_hora":              cfg.MostrarFechaHora,
		"usuario":                 cfg.MostrarUsuarioReporte,
		"consecutivo":             cfg.MostrarConsecutivo,
		"cantidad_ventas":         cfg.MostrarCantidadVentas,
		"total_descuentos":        cfg.MostrarTotalDescuentos,
		"detalle_fecha_entrada":   cfg.MostrarDetalleEntrada,
		"detalle_fecha_salida":    cfg.MostrarDetalleSalida,
		"detalle_numero_venta":    cfg.MostrarDetalleNumero,
		"detalle_estacion":        cfg.MostrarDetalleEstacion,
		"detalle_cajero":          cfg.MostrarDetalleCajero,
		"detalle_medio_pago":      cfg.MostrarDetalleMetodo,
		"detalle_total":           cfg.MostrarDetalleTotal,
		"resumen_total_productos": cfg.MostrarTotalProductos,
		"resumen_total_servicios": cfg.MostrarTotalServicios,
	}
	for name, enabled := range checks {
		if !enabled {
			t.Fatalf("campo de reporte de turno apagado por defecto: %s", name)
		}
	}
}

func TestEmpresaCorteCajaReportesDesdeConfiguracionFallsBackToResumen(t *testing.T) {
	cfg := EmpresaCorteCajaConfiguracion{EmpresaID: 9}
	got := EmpresaCorteCajaReportesDesdeConfiguracion(&cfg)
	if len(got) != 1 || got[0] != "resumen" {
		t.Fatalf("debe volver a resumen cuando todo esta apagado: %#v", got)
	}
}

func TestNormalizeEmpresaCorteCajaConfiguracionFormat(t *testing.T) {
	cfg := normalizeEmpresaCorteCajaConfiguracion(EmpresaCorteCajaConfiguracion{
		EmpresaID:        5,
		FormatoImpresion: "ticket",
		Estado:           "otro",
		Observaciones:    "  ok  ",
	})
	if cfg.FormatoImpresion != "pos" {
		t.Fatalf("formato inesperado: %s", cfg.FormatoImpresion)
	}
	if cfg.Estado != "activo" {
		t.Fatalf("estado inesperado: %s", cfg.Estado)
	}
	if cfg.Observaciones != "ok" {
		t.Fatalf("observaciones no normalizadas: %q", cfg.Observaciones)
	}
}
