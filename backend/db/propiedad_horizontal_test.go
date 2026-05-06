package db

import "testing"

func TestNormalizePropiedadHorizontalUnidad(t *testing.T) {
	got := normalizePropiedadHorizontalUnidad(EmpresaPropiedadHorizontalUnidad{
		Codigo:     " t1-301 ",
		TipoUnidad: "raro",
		Estado:     "x",
		AreaM2:     -10,
		CuotaBase:  -1,
	})
	if got.Codigo != "T1-301" {
		t.Fatalf("codigo normalizado inesperado: %#v", got)
	}
	if got.TipoUnidad != "apartamento" || got.Estado != "ocupada" {
		t.Fatalf("defaults inesperados: %#v", got)
	}
	if got.AreaM2 != 0 || got.CuotaBase != 0 {
		t.Fatalf("valores negativos no saneados: %#v", got)
	}
}

func TestNormalizePropiedadHorizontalCargo(t *testing.T) {
	got := normalizePropiedadHorizontalCargo(EmpresaPropiedadHorizontalCargo{
		Periodo:     "2026-05-15",
		Concepto:    " Cuota ",
		ValorBase:   100000,
		InteresMora: 5000,
		Descuento:   10000,
	})
	if got.Periodo != "2026-05" {
		t.Fatalf("periodo inesperado: %#v", got)
	}
	if got.Total != 95000 || got.SaldoPendiente != 95000 {
		t.Fatalf("totales inesperados: %#v", got)
	}
}
