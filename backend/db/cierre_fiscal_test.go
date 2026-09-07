package db

import "testing"

func TestNormalizeCierreFiscalPeriodoRowDefaults(t *testing.T) {
	got := normalizeCierreFiscalPeriodoRow(EmpresaCierreFiscalPeriodo{
		Periodo:       "2026-05-31",
		EstadoPeriodo: "raro",
		TipoCierre:    "",
	})
	if got.Periodo != "2026-05" || got.FechaDesde != "2026-05-01" || got.FechaHasta == "" {
		t.Fatalf("periodo normalizado inesperado: %#v", got)
	}
	if got.EstadoPeriodo != "abierto" || got.TipoCierre != "mensual" {
		t.Fatalf("defaults inesperados: %#v", got)
	}
	if !got.BloqueaVentas || !got.BloqueaCompras || !got.BloqueaCaja || !got.BloqueaInventario || !got.BloqueaContabilidad || !got.BloqueaFacturacion {
		t.Fatalf("bloqueos default incompletos: %#v", got)
	}
}

func TestCierreFiscalPeriodoBloqueaModulo(t *testing.T) {
	p := EmpresaCierreFiscalPeriodo{BloqueaVentas: true, BloqueaCompras: false, BloqueaCaja: true, BloqueaInventario: false, BloqueaContabilidad: true, BloqueaFacturacion: false}
	if !cierreFiscalPeriodoBloqueaModulo(p, "pos") {
		t.Fatalf("pos debe mapear a ventas")
	}
	if cierreFiscalPeriodoBloqueaModulo(p, "compras") {
		t.Fatalf("compras no debe estar bloqueado")
	}
	if !cierreFiscalPeriodoBloqueaModulo(p, "tesoreria_presupuesto") {
		t.Fatalf("tesoreria debe mapear a caja")
	}
	if !cierreFiscalPeriodoBloqueaModulo(p, "contabilidad_colombia") {
		t.Fatalf("contabilidad_colombia debe mapear a contabilidad")
	}
}

func TestNormalizeCierreFiscalPolitica(t *testing.T) {
	got := normalizeCierreFiscalPolitica(EmpresaCierreFiscalPolitica{
		Modulo:                 " Facturacion Electronica ",
		DiasEdicionRetroactiva: -5,
		Estado:                 "x",
	})
	if got.Modulo != "facturacion" || got.DiasEdicionRetroactiva != 0 || got.Estado != "activo" {
		t.Fatalf("politica normalizada inesperada: %#v", got)
	}
	if got.Nombre == "" {
		t.Fatalf("nombre default vacio")
	}
}
