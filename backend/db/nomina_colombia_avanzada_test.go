package db

import (
	"strings"
	"testing"
)

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

func TestNominaColombiaConceptosProfesionalesIncluyePrestaciones(t *testing.T) {
	rows := nominaColombiaConceptosProfesionales(9, "qa")
	seen := map[string]bool{}
	for _, row := range rows {
		seen[row.Codigo] = true
		if row.EmpresaID != 9 || row.UsuarioCreador != "qa" {
			t.Fatalf("concepto sin empresa/usuario esperado: %+v", row)
		}
	}
	for _, code := range []string{"BASICO", "HED", "SALUD", "PENSION", "ARL", "CAJA", "CESANTIAS", "PRIMA", "PROVVAC"} {
		if !seen[code] {
			t.Fatalf("falta concepto profesional %s", code)
		}
	}
}

func TestAplicarNovedadesAprobadasEnLiquidacion(t *testing.T) {
	liq := &EmpresaNominaLiquidacion{
		DevengadoTotal:        1000000,
		IngresoBaseCotizacion: 1000000,
		Bonificacion:          0,
		DeduccionFija:         10000,
		OtrasDeducciones:      0,
		ResumenJSON:           `{"asistencia_registros":4}`,
	}
	cfg := &EmpresaNominaConfiguracion{DeduccionSaludPorcentaje: 4, DeduccionPensionPorcentaje: 4}
	aplicadas, devengado, deduccion := aplicarNovedadesAprobadasEnLiquidacion(liq, cfg, []EmpresaNominaNovedadColombia{
		{Tipo: "devengado", ValorTotal: 100000, AfectaIBC: true, EstadoAprobacion: "aprobado", Estado: "activo"},
		{Tipo: "deduccion", ValorTotal: 50000, EstadoAprobacion: "aprobado", Estado: "activo"},
		{Tipo: "devengado", ValorTotal: 999999, AfectaIBC: true, EstadoAprobacion: "pendiente", Estado: "activo"},
	})
	if aplicadas != 2 || devengado != 100000 || deduccion != 50000 {
		t.Fatalf("novedades aplicadas=%d dev=%v ded=%v", aplicadas, devengado, deduccion)
	}
	if liq.DevengadoTotal != 1100000 || liq.IngresoBaseCotizacion != 1100000 {
		t.Fatalf("totales dev/ibc = %v/%v", liq.DevengadoTotal, liq.IngresoBaseCotizacion)
	}
	if liq.DeduccionSalud != 44000 || liq.DeduccionPension != 44000 {
		t.Fatalf("deducciones salud/pension = %v/%v", liq.DeduccionSalud, liq.DeduccionPension)
	}
	if liq.DeduccionTotal != 148000 || liq.NetoPagar != 952000 {
		t.Fatalf("total ded/neto = %v/%v", liq.DeduccionTotal, liq.NetoPagar)
	}
	if !strings.Contains(liq.ResumenJSON, `"novedades_colombia":2`) {
		t.Fatalf("resumen no incluye novedades: %s", liq.ResumenJSON)
	}
}
