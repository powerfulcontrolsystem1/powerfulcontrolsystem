package db

import (
	"math"
	"testing"
)

func TestEmpresaNominaConfigEmpleadoFestivoCRUD(t *testing.T) {
	dbConn := openCarritoInventarioTestDB(t)
	if err := EnsureEmpresaNominaSchema(dbConn); err != nil {
		t.Fatalf("ensure nomina schema: %v", err)
	}

	cfgID, err := UpsertEmpresaNominaConfiguracion(dbConn, EmpresaNominaConfiguracion{
		EmpresaID:                            7,
		PaisCodigo:                           "CO",
		Moneda:                               "COP",
		HorasOrdinariasSemana:                44,
		HorasOrdinariasDia:                   8,
		DiasNominaMes:                        30,
		RecargoNocturnoPorcentaje:            35,
		HoraExtraDiurnaPorcentaje:            25,
		HoraExtraNocturnaPorcentaje:          75,
		RecargoDominicalDiurnoPorcentaje:     75,
		RecargoDominicalNocturnoPorcentaje:   110,
		HoraExtraDominicalDiurnaPorcentaje:   100,
		HoraExtraDominicalNocturnaPorcentaje: 150,
		DeduccionSaludPorcentaje:             4,
		DeduccionPensionPorcentaje:           4,
		UsuarioCreador:                       "qa@empresa.com",
	})
	if err != nil {
		t.Fatalf("upsert nomina config: %v", err)
	}
	if cfgID <= 0 {
		t.Fatalf("expected cfg id > 0, got %d", cfgID)
	}

	cfg, err := GetEmpresaNominaConfiguracion(dbConn, 7)
	if err != nil {
		t.Fatalf("get nomina config: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected cfg not nil")
	}
	if cfg.DivisorHoraOrdinaria <= 0 {
		t.Fatalf("expected divisor_hora_ordinaria > 0, got %.2f", cfg.DivisorHoraOrdinaria)
	}

	empleadoID, err := CreateEmpresaNominaEmpleado(dbConn, EmpresaNominaEmpleado{
		EmpresaID:                7,
		EmpleadoID:               701,
		EmpleadoCodigo:           "EMP-701",
		EmpleadoNombre:           "Laura Gomez",
		EmpleadoDocumento:        "10990011",
		Cargo:                    "Cajera",
		SalarioBasicoMensual:     1800000,
		AuxilioTransporteMensual: 162000,
		BonificacionFijaMensual:  80000,
		DeduccionFijaMensual:     20000,
		JornadaHorasDia:          8,
		IncluirAuxilioTransporte: true,
		UsuarioCreador:           "qa@empresa.com",
	})
	if err != nil {
		t.Fatalf("create nomina empleado: %v", err)
	}
	if empleadoID <= 0 {
		t.Fatalf("expected empleado id > 0, got %d", empleadoID)
	}

	empleados, err := ListEmpresaNominaEmpleados(dbConn, 7, false, "", 50)
	if err != nil {
		t.Fatalf("list nomina empleados: %v", err)
	}
	if len(empleados) != 1 {
		t.Fatalf("expected 1 empleado, got %d", len(empleados))
	}
	if empleados[0].EmpleadoNombre != "Laura Gomez" {
		t.Fatalf("unexpected empleado nombre: %q", empleados[0].EmpleadoNombre)
	}

	festivoID, err := CreateEmpresaNominaFestivo(dbConn, EmpresaNominaFestivo{
		EmpresaID:      7,
		FechaFestivo:   "2026-04-03",
		Descripcion:    "Festivo local",
		UsuarioCreador: "qa@empresa.com",
	})
	if err != nil {
		t.Fatalf("create nomina festivo: %v", err)
	}
	if festivoID <= 0 {
		t.Fatalf("expected festivo id > 0, got %d", festivoID)
	}

	festivos, err := ListEmpresaNominaFestivos(dbConn, 7, false, "2026-04-01", "2026-04-10", 50)
	if err != nil {
		t.Fatalf("list nomina festivos: %v", err)
	}
	if len(festivos) != 1 {
		t.Fatalf("expected 1 festivo, got %d", len(festivos))
	}
}

func TestEmpresaNominaGenerateLiquidacionesFromAsistencia(t *testing.T) {
	dbConn := openCarritoInventarioTestDB(t)
	if err := EnsureEmpresaNominaSchema(dbConn); err != nil {
		t.Fatalf("ensure nomina schema: %v", err)
	}
	if err := EnsureEmpresaAsistenciaSchema(dbConn); err != nil {
		t.Fatalf("ensure asistencia schema: %v", err)
	}

	if _, err := UpsertEmpresaNominaConfiguracion(dbConn, EmpresaNominaConfiguracion{
		EmpresaID:                            21,
		HorasOrdinariasSemana:                44,
		HorasOrdinariasDia:                   8,
		DiasNominaMes:                        30,
		HoraNocturnaDesde:                    "21:00:00",
		HoraNocturnaHasta:                    "06:00:00",
		RecargoNocturnoPorcentaje:            35,
		HoraExtraDiurnaPorcentaje:            25,
		HoraExtraNocturnaPorcentaje:          75,
		RecargoDominicalDiurnoPorcentaje:     75,
		RecargoDominicalNocturnoPorcentaje:   110,
		HoraExtraDominicalDiurnaPorcentaje:   100,
		HoraExtraDominicalNocturnaPorcentaje: 150,
		DeduccionSaludPorcentaje:             4,
		DeduccionPensionPorcentaje:           4,
		UsuarioCreador:                       "qa@empresa.com",
	}); err != nil {
		t.Fatalf("upsert nomina config: %v", err)
	}

	empleadoID, err := CreateEmpresaNominaEmpleado(dbConn, EmpresaNominaEmpleado{
		EmpresaID:                21,
		EmpleadoID:               2101,
		EmpleadoCodigo:           "EMP-2101",
		EmpleadoNombre:           "Carlos Ruiz",
		EmpleadoDocumento:        "800100200",
		Cargo:                    "Supervisor",
		SalarioBasicoMensual:     2200000,
		AuxilioTransporteMensual: 162000,
		JornadaHorasDia:          8,
		IncluirAuxilioTransporte: true,
		UsuarioCreador:           "qa@empresa.com",
	})
	if err != nil {
		t.Fatalf("create empleado nomina: %v", err)
	}

	if _, err := CreateEmpresaNominaFestivo(dbConn, EmpresaNominaFestivo{
		EmpresaID:      21,
		FechaFestivo:   "2026-04-03",
		Descripcion:    "Festivo prueba",
		UsuarioCreador: "qa@empresa.com",
	}); err != nil {
		t.Fatalf("create festivo: %v", err)
	}

	// Jornada con 2 horas extra diurnas.
	if _, err := CreateEmpresaAsistenciaEmpleado(dbConn, EmpresaAsistenciaEmpleado{
		EmpresaID:         21,
		EmpleadoID:        2101,
		EmpleadoCodigo:    "EMP-2101",
		EmpleadoNombre:    "Carlos Ruiz",
		EmpleadoDocumento: "800100200",
		Cargo:             "Supervisor",
		FechaAsistencia:   "2026-04-01",
		HoraEntrada:       "08:00:00",
		HoraSalida:        "18:00:00",
		HorasTrabajadas:   10,
		EstadoAsistencia:  "presente",
		UsuarioCreador:    "qa@empresa.com",
	}); err != nil {
		t.Fatalf("create asistencia extra diurna: %v", err)
	}

	// Jornada nocturna ordinaria con recargo.
	if _, err := CreateEmpresaAsistenciaEmpleado(dbConn, EmpresaAsistenciaEmpleado{
		EmpresaID:         21,
		EmpleadoID:        2101,
		EmpleadoCodigo:    "EMP-2101",
		EmpleadoNombre:    "Carlos Ruiz",
		EmpleadoDocumento: "800100200",
		Cargo:             "Supervisor",
		FechaAsistencia:   "2026-04-02",
		HoraEntrada:       "21:00:00",
		HoraSalida:        "23:00:00",
		HorasTrabajadas:   2,
		EstadoAsistencia:  "presente",
		UsuarioCreador:    "qa@empresa.com",
	}); err != nil {
		t.Fatalf("create asistencia recargo nocturno: %v", err)
	}

	// Jornada en festivo (se liquida como dominical/festivo diurno).
	if _, err := CreateEmpresaAsistenciaEmpleado(dbConn, EmpresaAsistenciaEmpleado{
		EmpresaID:         21,
		EmpleadoID:        2101,
		EmpleadoCodigo:    "EMP-2101",
		EmpleadoNombre:    "Carlos Ruiz",
		EmpleadoDocumento: "800100200",
		Cargo:             "Supervisor",
		FechaAsistencia:   "2026-04-03",
		HoraEntrada:       "08:00:00",
		HoraSalida:        "12:00:00",
		HorasTrabajadas:   4,
		EstadoAsistencia:  "presente",
		UsuarioCreador:    "qa@empresa.com",
	}); err != nil {
		t.Fatalf("create asistencia festivo: %v", err)
	}

	result, err := GenerateEmpresaNominaLiquidaciones(dbConn, EmpresaNominaCalculoRequest{
		EmpresaID:        21,
		PeriodoDesde:     "2026-04-01",
		PeriodoHasta:     "2026-04-10",
		EmpleadoNominaID: empleadoID,
		Overwrite:        true,
		UsuarioCreador:   "qa@empresa.com",
	})
	if err != nil {
		t.Fatalf("generate nomina liquidaciones: %v", err)
	}
	if result == nil {
		t.Fatal("expected result not nil")
	}
	if result.Calculados != 1 {
		t.Fatalf("expected calculados=1, got %d", result.Calculados)
	}
	if len(result.Liquidaciones) != 1 {
		t.Fatalf("expected 1 liquidacion, got %d", len(result.Liquidaciones))
	}

	liq := result.Liquidaciones[0]
	if liq.HorasExtraDiurnas < 1.9 {
		t.Fatalf("expected horas_extra_diurnas >= 1.9, got %.2f", liq.HorasExtraDiurnas)
	}
	if liq.HorasRecargoNocturno < 1.9 {
		t.Fatalf("expected horas_recargo_nocturno >= 1.9, got %.2f", liq.HorasRecargoNocturno)
	}
	if liq.HorasDominicalesDiurnas < 3.9 {
		t.Fatalf("expected horas_dominicales_diurnas >= 3.9, got %.2f", liq.HorasDominicalesDiurnas)
	}
	if liq.DevengadoTotal <= 0 {
		t.Fatalf("expected devengado_total > 0, got %.2f", liq.DevengadoTotal)
	}
	if liq.NetoPagar <= 0 {
		t.Fatalf("expected neto_pagar > 0, got %.2f", liq.NetoPagar)
	}

	rows, err := ListEmpresaNominaLiquidaciones(dbConn, 21, EmpresaNominaLiquidacionFilter{
		PeriodoDesde: "2026-04-01",
		PeriodoHasta: "2026-04-30",
		Limit:        20,
	})
	if err != nil {
		t.Fatalf("list nomina liquidaciones: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 persisted liquidacion, got %d", len(rows))
	}
}

func TestEmpresaNominaCalculoPorPaisYEmpresa(t *testing.T) {
	dbConn := openCarritoInventarioTestDB(t)
	if err := EnsureEmpresaNominaSchema(dbConn); err != nil {
		t.Fatalf("ensure nomina schema: %v", err)
	}
	if err := EnsureEmpresaAsistenciaSchema(dbConn); err != nil {
		t.Fatalf("ensure asistencia schema: %v", err)
	}

	if _, err := UpsertEmpresaNominaConfiguracion(dbConn, EmpresaNominaConfiguracion{
		EmpresaID:                  91,
		PaisCodigo:                 "CO",
		Moneda:                     "COP",
		HorasOrdinariasSemana:      44,
		HorasOrdinariasDia:         8,
		DiasNominaMes:              30,
		DivisorHoraOrdinaria:       220,
		DeduccionSaludPorcentaje:   4,
		DeduccionPensionPorcentaje: 4,
	}); err != nil {
		t.Fatalf("upsert config CO: %v", err)
	}

	if _, err := UpsertEmpresaNominaConfiguracion(dbConn, EmpresaNominaConfiguracion{
		EmpresaID:                  92,
		PaisCodigo:                 "MX",
		Moneda:                     "MXN",
		HorasOrdinariasSemana:      48,
		HorasOrdinariasDia:         8,
		DiasNominaMes:              30,
		DivisorHoraOrdinaria:       240,
		DeduccionSaludPorcentaje:   3,
		DeduccionPensionPorcentaje: 2,
	}); err != nil {
		t.Fatalf("upsert config MX: %v", err)
	}

	empCO, err := CreateEmpresaNominaEmpleado(dbConn, EmpresaNominaEmpleado{
		EmpresaID:            91,
		EmpleadoID:           9101,
		EmpleadoCodigo:       "EMP-CO-1",
		EmpleadoNombre:       "Empleado CO",
		EmpleadoDocumento:    "DOC-CO-1",
		SalarioBasicoMensual: 2400000,
		JornadaHorasDia:      8,
	})
	if err != nil {
		t.Fatalf("create empleado CO: %v", err)
	}
	empMX, err := CreateEmpresaNominaEmpleado(dbConn, EmpresaNominaEmpleado{
		EmpresaID:            92,
		EmpleadoID:           9201,
		EmpleadoCodigo:       "EMP-MX-1",
		EmpleadoNombre:       "Empleado MX",
		EmpleadoDocumento:    "DOC-MX-1",
		SalarioBasicoMensual: 2400000,
		JornadaHorasDia:      8,
	})
	if err != nil {
		t.Fatalf("create empleado MX: %v", err)
	}

	if _, err := CreateEmpresaAsistenciaEmpleado(dbConn, EmpresaAsistenciaEmpleado{
		EmpresaID:         91,
		EmpleadoID:        9101,
		EmpleadoCodigo:    "EMP-CO-1",
		EmpleadoNombre:    "Empleado CO",
		EmpleadoDocumento: "DOC-CO-1",
		FechaAsistencia:   "2026-04-01",
		HoraEntrada:       "08:00:00",
		HoraSalida:        "16:00:00",
		HorasTrabajadas:   8,
		EstadoAsistencia:  "presente",
	}); err != nil {
		t.Fatalf("create asistencia CO: %v", err)
	}
	if _, err := CreateEmpresaAsistenciaEmpleado(dbConn, EmpresaAsistenciaEmpleado{
		EmpresaID:         92,
		EmpleadoID:        9201,
		EmpleadoCodigo:    "EMP-MX-1",
		EmpleadoNombre:    "Empleado MX",
		EmpleadoDocumento: "DOC-MX-1",
		FechaAsistencia:   "2026-04-01",
		HoraEntrada:       "08:00:00",
		HoraSalida:        "16:00:00",
		HorasTrabajadas:   8,
		EstadoAsistencia:  "presente",
	}); err != nil {
		t.Fatalf("create asistencia MX: %v", err)
	}

	resCO, err := GenerateEmpresaNominaLiquidaciones(dbConn, EmpresaNominaCalculoRequest{
		EmpresaID:        91,
		PeriodoDesde:     "2026-04-01",
		PeriodoHasta:     "2026-04-01",
		EmpleadoNominaID: empCO,
		Overwrite:        true,
	})
	if err != nil {
		t.Fatalf("generate CO: %v", err)
	}
	resMX, err := GenerateEmpresaNominaLiquidaciones(dbConn, EmpresaNominaCalculoRequest{
		EmpresaID:        92,
		PeriodoDesde:     "2026-04-01",
		PeriodoHasta:     "2026-04-01",
		EmpleadoNominaID: empMX,
		Overwrite:        true,
	})
	if err != nil {
		t.Fatalf("generate MX: %v", err)
	}

	if len(resCO.Liquidaciones) != 1 || len(resMX.Liquidaciones) != 1 {
		t.Fatalf("expected one liquidacion per company: co=%d mx=%d", len(resCO.Liquidaciones), len(resMX.Liquidaciones))
	}
	valorHoraCO := resCO.Liquidaciones[0].ValorHoraOrdinaria
	valorHoraMX := resMX.Liquidaciones[0].ValorHoraOrdinaria
	if !(valorHoraCO > valorHoraMX) {
		t.Fatalf("expected valor hora CO > MX due divisor 220 vs 240, got co=%.2f mx=%.2f", valorHoraCO, valorHoraMX)
	}

	if _, err := UpsertEmpresaNominaConfiguracion(dbConn, EmpresaNominaConfiguracion{
		EmpresaID:                  92,
		PaisCodigo:                 "MX",
		Moneda:                     "MXN",
		HorasOrdinariasSemana:      48,
		HorasOrdinariasDia:         8,
		DiasNominaMes:              30,
		DivisorHoraOrdinaria:       200,
		DeduccionSaludPorcentaje:   3,
		DeduccionPensionPorcentaje: 2,
	}); err != nil {
		t.Fatalf("upsert config MX override: %v", err)
	}

	resMXOverride, err := GenerateEmpresaNominaLiquidaciones(dbConn, EmpresaNominaCalculoRequest{
		EmpresaID:        92,
		PeriodoDesde:     "2026-04-01",
		PeriodoHasta:     "2026-04-01",
		EmpleadoNominaID: empMX,
		Overwrite:        true,
	})
	if err != nil {
		t.Fatalf("generate MX override: %v", err)
	}
	if len(resMXOverride.Liquidaciones) != 1 {
		t.Fatalf("expected one liquidacion for MX override, got %d", len(resMXOverride.Liquidaciones))
	}
	if !(resMXOverride.Liquidaciones[0].ValorHoraOrdinaria > valorHoraMX) {
		t.Fatalf("expected company override divisor to increase valor hora, before=%.2f after=%.2f", valorHoraMX, resMXOverride.Liquidaciones[0].ValorHoraOrdinaria)
	}
}

func TestEmpresaNominaDesprendibleYConciliacionAsistencia(t *testing.T) {
	dbConn := openCarritoInventarioTestDB(t)
	if err := EnsureEmpresaNominaSchema(dbConn); err != nil {
		t.Fatalf("ensure nomina schema: %v", err)
	}
	if err := EnsureEmpresaAsistenciaSchema(dbConn); err != nil {
		t.Fatalf("ensure asistencia schema: %v", err)
	}

	if _, err := UpsertEmpresaNominaConfiguracion(dbConn, EmpresaNominaConfiguracion{
		EmpresaID:                            95,
		PaisCodigo:                           "CO",
		Moneda:                               "COP",
		HorasOrdinariasSemana:                44,
		HorasOrdinariasDia:                   8,
		DiasNominaMes:                        30,
		DivisorHoraOrdinaria:                 220,
		RecargoNocturnoPorcentaje:            35,
		HoraExtraDiurnaPorcentaje:            25,
		HoraExtraNocturnaPorcentaje:          75,
		RecargoDominicalDiurnoPorcentaje:     75,
		RecargoDominicalNocturnoPorcentaje:   110,
		HoraExtraDominicalDiurnaPorcentaje:   100,
		HoraExtraDominicalNocturnaPorcentaje: 150,
		DeduccionSaludPorcentaje:             4,
		DeduccionPensionPorcentaje:           4,
	}); err != nil {
		t.Fatalf("upsert config: %v", err)
	}

	empID, err := CreateEmpresaNominaEmpleado(dbConn, EmpresaNominaEmpleado{
		EmpresaID:                95,
		EmpleadoID:               9501,
		EmpleadoCodigo:           "EMP-9501",
		EmpleadoNombre:           "Empleado Conciliacion",
		EmpleadoDocumento:        "DOC-9501",
		Cargo:                    "Operador",
		TipoContrato:             "indefinido",
		SalarioBasicoMensual:     2100000,
		AuxilioTransporteMensual: 162000,
		JornadaHorasDia:          8,
		IncluirAuxilioTransporte: true,
	})
	if err != nil {
		t.Fatalf("create empleado: %v", err)
	}

	if _, err := CreateEmpresaAsistenciaEmpleado(dbConn, EmpresaAsistenciaEmpleado{
		EmpresaID:         95,
		EmpleadoID:        9501,
		EmpleadoCodigo:    "EMP-9501",
		EmpleadoNombre:    "Empleado Conciliacion",
		EmpleadoDocumento: "DOC-9501",
		FechaAsistencia:   "2026-04-01",
		HoraEntrada:       "08:00:00",
		HoraSalida:        "16:00:00",
		HorasTrabajadas:   8,
		EstadoAsistencia:  "presente",
	}); err != nil {
		t.Fatalf("create asistencia 1: %v", err)
	}
	if _, err := CreateEmpresaAsistenciaEmpleado(dbConn, EmpresaAsistenciaEmpleado{
		EmpresaID:         95,
		EmpleadoID:        9501,
		EmpleadoCodigo:    "EMP-9501",
		EmpleadoNombre:    "Empleado Conciliacion",
		EmpleadoDocumento: "DOC-9501",
		FechaAsistencia:   "2026-04-02",
		HoraEntrada:       "08:00:00",
		HoraSalida:        "16:00:00",
		HorasTrabajadas:   8,
		EstadoAsistencia:  "presente",
	}); err != nil {
		t.Fatalf("create asistencia 2: %v", err)
	}

	if _, err := GenerateEmpresaNominaLiquidaciones(dbConn, EmpresaNominaCalculoRequest{
		EmpresaID:        95,
		PeriodoDesde:     "2026-04-01",
		PeriodoHasta:     "2026-04-10",
		EmpleadoNominaID: empID,
		Overwrite:        true,
	}); err != nil {
		t.Fatalf("generate liquidacion: %v", err)
	}

	docInicial, err := GetEmpresaNominaDesprendible(dbConn, 95, empID, "2026-04-01", "2026-04-10")
	if err != nil {
		t.Fatalf("get desprendible inicial: %v", err)
	}
	if docInicial.NetoPagar <= 0 {
		t.Fatalf("expected neto pagar > 0, got %.2f", docInicial.NetoPagar)
	}
	if docInicial.HorasAsistencia <= 0 {
		t.Fatalf("expected horas asistencia > 0, got %.2f", docInicial.HorasAsistencia)
	}

	if _, err := CreateEmpresaAsistenciaEmpleado(dbConn, EmpresaAsistenciaEmpleado{
		EmpresaID:         95,
		EmpleadoID:        9501,
		EmpleadoCodigo:    "EMP-9501",
		EmpleadoNombre:    "Empleado Conciliacion",
		EmpleadoDocumento: "DOC-9501",
		FechaAsistencia:   "2026-04-03",
		HoraEntrada:       "08:00:00",
		HoraSalida:        "16:00:00",
		HorasTrabajadas:   8,
		EstadoAsistencia:  "presente",
	}); err != nil {
		t.Fatalf("create asistencia 3: %v", err)
	}

	conciliarDryRun, err := ConciliarEmpresaNominaAsistencia(dbConn, EmpresaNominaConciliacionRequest{
		EmpresaID:        95,
		PeriodoDesde:     "2026-04-01",
		PeriodoHasta:     "2026-04-10",
		EmpleadoNominaID: empID,
		AutoRecalcular:   false,
	})
	if err != nil {
		t.Fatalf("conciliar dry-run: %v", err)
	}
	if conciliarDryRun.TotalInconsistencias <= 0 {
		t.Fatalf("expected inconsistencias > 0 on dry-run, got %d", conciliarDryRun.TotalInconsistencias)
	}

	conciliarFix, err := ConciliarEmpresaNominaAsistencia(dbConn, EmpresaNominaConciliacionRequest{
		EmpresaID:        95,
		PeriodoDesde:     "2026-04-01",
		PeriodoHasta:     "2026-04-10",
		EmpleadoNominaID: empID,
		AutoRecalcular:   true,
		UsuarioCreador:   "qa@empresa.com",
	})
	if err != nil {
		t.Fatalf("conciliar con recalculo: %v", err)
	}
	if conciliarFix.TotalRecalculados <= 0 {
		t.Fatalf("expected recalculados > 0, got %d", conciliarFix.TotalRecalculados)
	}

	docFinal, err := GetEmpresaNominaDesprendible(dbConn, 95, empID, "2026-04-01", "2026-04-10")
	if err != nil {
		t.Fatalf("get desprendible final: %v", err)
	}
	if !(docFinal.HorasAsistencia > docInicial.HorasAsistencia) {
		t.Fatalf("expected horas asistencia to increase after conciliar/recalcular, before=%.2f after=%.2f", docInicial.HorasAsistencia, docFinal.HorasAsistencia)
	}
}

func TestEmpresaNominaLiquidacionIntegraComisionesServicio(t *testing.T) {
	dbConn := openCarritoInventarioTestDB(t)
	if err := EnsureEmpresaNominaSchema(dbConn); err != nil {
		t.Fatalf("ensure nomina schema: %v", err)
	}
	if err := EnsureEmpresaAsistenciaSchema(dbConn); err != nil {
		t.Fatalf("ensure asistencia schema: %v", err)
	}
	if err := EnsureEmpresaComisionesServicioSchema(dbConn); err != nil {
		t.Fatalf("ensure comisiones schema: %v", err)
	}

	if _, err := UpsertEmpresaNominaConfiguracion(dbConn, EmpresaNominaConfiguracion{
		EmpresaID:                  120,
		PaisCodigo:                 "CO",
		Moneda:                     "COP",
		HorasOrdinariasSemana:      44,
		HorasOrdinariasDia:         8,
		DiasNominaMes:              30,
		DivisorHoraOrdinaria:       220,
		DeduccionSaludPorcentaje:   4,
		DeduccionPensionPorcentaje: 4,
	}); err != nil {
		t.Fatalf("upsert nomina config: %v", err)
	}

	empID, err := CreateEmpresaNominaEmpleado(dbConn, EmpresaNominaEmpleado{
		EmpresaID:            120,
		EmpleadoID:           12001,
		EmpleadoCodigo:       "emp-com-01",
		EmpleadoNombre:       "Lavador Nomina",
		EmpleadoDocumento:    "DOC-COM-01",
		Cargo:                "Lavador",
		SalarioBasicoMensual: 2200000,
		JornadaHorasDia:      8,
	})
	if err != nil {
		t.Fatalf("create empleado nomina: %v", err)
	}

	if _, err := CreateEmpresaAsistenciaEmpleado(dbConn, EmpresaAsistenciaEmpleado{
		EmpresaID:         120,
		EmpleadoID:        12001,
		EmpleadoCodigo:    "emp-com-01",
		EmpleadoNombre:    "Lavador Nomina",
		EmpleadoDocumento: "DOC-COM-01",
		FechaAsistencia:   "2026-04-08",
		HoraEntrada:       "08:00:00",
		HoraSalida:        "16:00:00",
		HorasTrabajadas:   8,
		EstadoAsistencia:  "presente",
	}); err != nil {
		t.Fatalf("create asistencia: %v", err)
	}

	if _, err := CreateEmpresaComisionServicioMovimiento(dbConn, EmpresaComisionServicioMovimiento{
		EmpresaID:          120,
		UsuarioOrigen:      "cajero@empresa.com",
		UsuarioLavador:     "emp-com-01",
		BaseServicio:       10000,
		PorcentajeComision: 10,
		MontoComision:      1000,
		UsuarioCreador:     "cajero@empresa.com",
		Estado:             "activo",
		FechaMovimiento:    "2026-04-08 10:00:00",
	}); err != nil {
		t.Fatalf("create comision 1: %v", err)
	}
	if _, err := CreateEmpresaComisionServicioMovimiento(dbConn, EmpresaComisionServicioMovimiento{
		EmpresaID:          120,
		UsuarioOrigen:      "cajero@empresa.com",
		UsuarioLavador:     "emp-com-01",
		BaseServicio:       5000,
		PorcentajeComision: 10,
		MontoComision:      500,
		UsuarioCreador:     "cajero@empresa.com",
		Estado:             "activo",
		FechaMovimiento:    "2026-04-08 11:00:00",
	}); err != nil {
		t.Fatalf("create comision 2: %v", err)
	}

	result, err := GenerateEmpresaNominaLiquidaciones(dbConn, EmpresaNominaCalculoRequest{
		EmpresaID:        120,
		PeriodoDesde:     "2026-04-01",
		PeriodoHasta:     "2026-04-30",
		EmpleadoNominaID: empID,
		Overwrite:        true,
		UsuarioCreador:   "nomina@empresa.com",
	})
	if err != nil {
		t.Fatalf("generate liquidaciones: %v", err)
	}
	if len(result.Liquidaciones) != 1 {
		t.Fatalf("expected 1 liquidacion, got %d", len(result.Liquidaciones))
	}

	liq := result.Liquidaciones[0]
	if math.Abs(liq.ComisionesServicioTotal-1500) > 0.001 {
		t.Fatalf("expected comisiones_servicio_total 1500, got %.4f", liq.ComisionesServicioTotal)
	}
	if liq.ComisionesServicioMovimientos != 2 {
		t.Fatalf("expected comisiones_servicio_movimientos 2, got %d", liq.ComisionesServicioMovimientos)
	}

	vinculados, err := ListEmpresaComisionServicioMovimientos(dbConn, 120, EmpresaComisionServicioMovimientoFilter{
		LiquidacionNominaID: liq.ID,
		Limit:               20,
	})
	if err != nil {
		t.Fatalf("list comisiones vinculadas: %v", err)
	}
	if len(vinculados) != 2 {
		t.Fatalf("expected 2 comisiones vinculadas, got %d", len(vinculados))
	}

	doc, err := GetEmpresaNominaDesprendible(dbConn, 120, empID, "2026-04-01", "2026-04-30")
	if err != nil {
		t.Fatalf("get desprendible: %v", err)
	}
	if math.Abs(doc.ComisionesServicioTotal-1500) > 0.001 {
		t.Fatalf("expected desprendible comisiones_servicio_total 1500, got %.4f", doc.ComisionesServicioTotal)
	}
}
