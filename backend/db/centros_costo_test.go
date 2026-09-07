package db

import "testing"

func TestNormalizeCentroCostoCodigo(t *testing.T) {
	cases := map[string]string{
		" motel calipso ": "MOTEL-CALIPSO",
		"OBRAS_demo":      "OBRAS-DEMO",
		" ventas/online":  "VENTAS-ONLINE",
	}
	for input, want := range cases {
		if got := normalizeCentroCostoCodigo(input); got != want {
			t.Fatalf("normalizeCentroCostoCodigo(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestBuildEmpresaCentrosCostoDashboardFromRows(t *testing.T) {
	centros := []EmpresaCentroCosto{
		{EmpresaID: 7, Codigo: "MOTEL-CALIPSO", Nombre: "Motel Calipso", Tipo: "sucursal", MetaMargenPct: 35, Estado: "activo"},
		{EmpresaID: 7, Codigo: "ADMIN", Nombre: "Administracion", Tipo: "area", Estado: "activo"},
	}
	reglas := []EmpresaCentroCostoRegla{{EmpresaID: 7, CentroCostoCodigo: "MOTEL-CALIPSO", Nombre: "Compras", Activa: true, Estado: "activo"}}
	pres := []EmpresaCentroCostoPresupuesto{
		{EmpresaID: 7, CentroCostoCodigo: "MOTEL-CALIPSO", Periodo: "2026-05", IngresosPresupuesto: 1000000, EgresosPresupuesto: 400000, MetaMargenPct: 35, Estado: "aprobado"},
	}
	movs := []EmpresaCentroCostoMovimiento{
		{CentroCostoCodigo: "motel calipso", OrigenModulo: "ventas", Periodo: "2026-05", Tipo: "ingreso", Ingresos: 900000, Concepto: "Ventas"},
		{CentroCostoCodigo: "MOTEL-CALIPSO", OrigenModulo: "compras", Periodo: "2026-05", Tipo: "egreso", Egresos: 350000, Concepto: "Compras"},
		{CentroCostoCodigo: "nuevo proyecto", OrigenModulo: "aiu", Periodo: "2026-05", Tipo: "ingreso", Ingresos: 200000, Concepto: "AIU"},
	}

	got := buildEmpresaCentrosCostoDashboardFromRows(7, "2026-05", centros, reglas, pres, movs)
	if got.CentrosActivos != 2 || got.ReglasActivas != 1 || got.MovimientosTotal != 3 {
		t.Fatalf("resumen inesperado: %#v", got)
	}
	if got.IngresosTotal != 1100000 || got.EgresosTotal != 350000 || got.MargenTotal != 750000 {
		t.Fatalf("totales inesperados: ingresos %.2f egresos %.2f margen %.2f", got.IngresosTotal, got.EgresosTotal, got.MargenTotal)
	}
	if len(got.Rentabilidad) != 3 {
		t.Fatalf("rentabilidad esperada para 3 centros, got %d", len(got.Rentabilidad))
	}
	var foundInferred bool
	for _, row := range got.Rentabilidad {
		if row.CentroCostoCodigo == "NUEVO-PROYECTO" && row.InferidoDeMovimientos {
			foundInferred = true
		}
	}
	if !foundInferred {
		t.Fatalf("no se incluyo centro inferido desde movimientos: %#v", got.Rentabilidad)
	}
}

func TestNormalizeEmpresaCentroCostoPresupuestoDefaults(t *testing.T) {
	got := normalizeEmpresaCentroCostoPresupuesto(EmpresaCentroCostoPresupuesto{
		CentroCostoCodigo:   " operaciones ",
		Periodo:             "2026-05-31",
		Escenario:           "",
		IngresosPresupuesto: -10,
		EgresosPresupuesto:  500,
		MetaMargenPct:       140,
		Estado:              "raro",
	})
	if got.CentroCostoCodigo != "OPERACIONES" || got.Periodo != "2026-05" || got.Escenario != "base" {
		t.Fatalf("campos normalizados inesperados: %#v", got)
	}
	if got.IngresosPresupuesto != 0 || got.EgresosPresupuesto != 500 || got.MetaMargenPct != 100 || got.Estado != "aprobado" {
		t.Fatalf("valores/defaults inesperados: %#v", got)
	}
}
