package db

import "testing"

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
