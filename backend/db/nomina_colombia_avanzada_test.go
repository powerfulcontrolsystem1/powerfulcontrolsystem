package db

import "testing"

func TestNormalizeNominaColombiaConcepto(t *testing.T) {
	item := normalizeNominaConceptoColombia(EmpresaNominaConceptoColombia{
		Codigo:         " bono-extra ",
		Nombre:         "  Bono productividad  ",
		Tipo:           "APORTE",
		Porcentaje:     120,
		ValorFijo:      -50,
		CuentaContable: " 510548 ",
	})

	if item.Codigo != "BONO-EXTRA" {
		t.Fatalf("codigo normalizado = %q", item.Codigo)
	}
	if item.Nombre != "Bono productividad" {
		t.Fatalf("nombre normalizado = %q", item.Nombre)
	}
	if item.Tipo != "aporte" {
		t.Fatalf("tipo normalizado = %q", item.Tipo)
	}
	if item.Porcentaje != 120 {
		t.Fatalf("porcentaje normalizado = %v", item.Porcentaje)
	}
	if item.ValorFijo != 0 {
		t.Fatalf("valor fijo normalizado = %v", item.ValorFijo)
	}
	if item.CuentaContable != "510548" {
		t.Fatalf("cuenta contable normalizada = %q", item.CuentaContable)
	}
}

func TestNormalizeNominaColombiaNovedad(t *testing.T) {
	item := normalizeNominaNovedadColombia(EmpresaNominaNovedadColombia{
		PeriodoDesde:     "2026-05-01",
		Tipo:             "DEDUCCION",
		CodigoConcepto:   " salud ",
		Descripcion:      "  Ajuste salud  ",
		Cantidad:         2,
		ValorUnitario:    15000,
		EstadoAprobacion: "APROBADO",
	})

	if item.Tipo != "deduccion" || item.CodigoConcepto != "SALUD" {
		t.Fatalf("tipo/codigo normalizados = %q/%q", item.Tipo, item.CodigoConcepto)
	}
	if item.Descripcion != "Ajuste salud" {
		t.Fatalf("descripcion normalizada = %q", item.Descripcion)
	}
	if item.ValorTotal != 30000 {
		t.Fatalf("valor total calculado = %v", item.ValorTotal)
	}
	if item.FechaNovedad != "2026-05-01" {
		t.Fatalf("fecha novedad por defecto = %q", item.FechaNovedad)
	}
	if item.EstadoAprobacion != "aprobado" {
		t.Fatalf("estado aprobacion normalizado = %q", item.EstadoAprobacion)
	}
}

func TestBuildNominaPILARowColombia(t *testing.T) {
	cfg := &EmpresaNominaConfiguracion{
		DeduccionSaludPorcentaje:         4,
		DeduccionPensionPorcentaje:       4,
		AporteSaludEmpleadorPorcentaje:   8.5,
		AportePensionEmpleadorPorcentaje: 12,
		AporteARLPorcentaje:              0.522,
		AporteCajaCompensacionPorcentaje: 4,
		AporteICBFPorcentaje:             3,
		AporteSENAPorcentaje:             2,
	}
	row := buildNominaPILARowColombia(7, "2026-05", EmpresaNominaLiquidacion{
		EmpleadoNominaID:      11,
		EmpleadoNombre:        "Empleado QA",
		EmpleadoDocumento:     "123",
		IngresoBaseCotizacion: 2500000,
	}, cfg, "qa")

	if row.SaludEmpleado != 100000 || row.PensionEmpleado != 100000 {
		t.Fatalf("deducciones empleado inesperadas: salud=%v pension=%v", row.SaludEmpleado, row.PensionEmpleado)
	}
	if row.TotalAportes != 950550 {
		t.Fatalf("total aportes PILA = %v", row.TotalAportes)
	}
	if row.Periodo != "2026-05" || row.Estado != "generado" {
		t.Fatalf("periodo/estado = %q/%q", row.Periodo, row.Estado)
	}
}
