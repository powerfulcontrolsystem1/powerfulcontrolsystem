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
