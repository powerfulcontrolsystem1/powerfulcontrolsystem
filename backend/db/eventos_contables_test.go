package db

import (
	"encoding/json"
	"testing"
)

func TestEmpresaEventosContablesCreateAndList(t *testing.T) {
	dbConn := openFinanzasTestDB(t)
	if err := EnsureEmpresaEventosContablesSchema(dbConn); err != nil {
		t.Fatalf("ensure eventos contables schema: %v", err)
	}

	id, err := CreateEmpresaEventoContable(dbConn, EmpresaEventoContable{
		EmpresaID:       9,
		Modulo:          "ventas",
		Evento:          "venta_pagada",
		Entidad:         "carrito_compra",
		EntidadID:       101,
		DocumentoTipo:   "carrito",
		DocumentoCodigo: "CAR-101",
		MontoTotal:      45000,
		Moneda:          "COP",
		PayloadJSON:     `{"total_pagado":45000}`,
		Origen:          "test_db",
		UsuarioCreador:  "tester",
	})
	if err != nil {
		t.Fatalf("create evento contable: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected id > 0, got %d", id)
	}

	rows, err := ListEmpresaEventosContables(dbConn, 9, EmpresaEventoContableFilter{Modulo: "ventas", Limit: 10})
	if err != nil {
		t.Fatalf("list eventos contables: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 evento, got %d", len(rows))
	}
	if rows[0].Evento != "venta_pagada" {
		t.Fatalf("expected evento venta_pagada, got %q", rows[0].Evento)
	}
	if rows[0].PeriodoContable == "" {
		t.Fatalf("expected periodo_contable generated")
	}
}

func TestEmpresaEventosContablesRejectInvalidContrato(t *testing.T) {
	dbConn := openFinanzasTestDB(t)
	if err := EnsureEmpresaEventosContablesSchema(dbConn); err != nil {
		t.Fatalf("ensure eventos contables schema: %v", err)
	}

	_, err := CreateEmpresaEventoContable(dbConn, EmpresaEventoContable{
		EmpresaID: 1,
		Modulo:    "ventas",
		Evento:    "evento_no_soportado",
		Entidad:   "carrito_compra",
	})
	if err == nil {
		t.Fatalf("expected error for invalid contract event")
	}
}

func TestProcessEmpresaEventosContablesPendientesGeneraAsientosIdempotentes(t *testing.T) {
	dbConn := openFinanzasTestDB(t)
	if err := EnsureEmpresaFinanzasSchema(dbConn); err != nil {
		t.Fatalf("ensure finanzas schema: %v", err)
	}
	if err := EnsureEmpresaEventosContablesSchema(dbConn); err != nil {
		t.Fatalf("ensure eventos contables schema: %v", err)
	}

	empresaID := int64(21)
	eventoID, err := CreateEmpresaEventoContable(dbConn, EmpresaEventoContable{
		EmpresaID:       empresaID,
		Modulo:          "finanzas",
		Evento:          "movimiento_ingreso_registrado",
		Entidad:         "finanzas_movimiento",
		EntidadID:       9001,
		DocumentoTipo:   "comprobante",
		DocumentoCodigo: "ING-9001",
		PeriodoContable: "2026-04",
		MontoTotal:      120000,
		Moneda:          "COP",
		PayloadJSON:     `{"tipo_movimiento":"ingreso","categoria":"ventas"}`,
		UsuarioCreador:  "tester",
	})
	if err != nil {
		t.Fatalf("create evento contable: %v", err)
	}

	resultado, err := ProcessEmpresaEventosContablesPendientes(dbConn, empresaID, "tester", 20)
	if err != nil {
		t.Fatalf("process eventos pendientes: %v", err)
	}
	if resultado.EventosRevisados != 1 {
		t.Fatalf("expected eventos_revisados=1, got %d", resultado.EventosRevisados)
	}
	if resultado.EventosProcesados != 1 {
		t.Fatalf("expected eventos_procesados=1, got %d", resultado.EventosProcesados)
	}
	if resultado.AsientosCreados != 1 {
		t.Fatalf("expected asientos_creados=1, got %d", resultado.AsientosCreados)
	}

	var asientosCount int64
	if err := dbConn.QueryRow(`SELECT COALESCE(COUNT(1), 0) FROM empresa_asientos_contables WHERE empresa_id = ?`, empresaID).Scan(&asientosCount); err != nil {
		t.Fatalf("count asientos: %v", err)
	}
	if asientosCount != 1 {
		t.Fatalf("expected asientos count=1, got %d", asientosCount)
	}

	if _, err := dbConn.Exec(`UPDATE empresa_eventos_contables SET procesado = 0, fecha_procesado = NULL WHERE id = ?`, eventoID); err != nil {
		t.Fatalf("reopen evento for idempotency check: %v", err)
	}

	resultado2, err := ProcessEmpresaEventosContablesPendientes(dbConn, empresaID, "tester", 20)
	if err != nil {
		t.Fatalf("re-process eventos pendientes: %v", err)
	}
	if resultado2.EventosProcesados != 1 {
		t.Fatalf("expected eventos_procesados=1 in second run, got %d", resultado2.EventosProcesados)
	}
	if resultado2.AsientosExistentes != 1 {
		t.Fatalf("expected asientos_existentes=1 in second run, got %d", resultado2.AsientosExistentes)
	}

	if err := dbConn.QueryRow(`SELECT COALESCE(COUNT(1), 0) FROM empresa_asientos_contables WHERE empresa_id = ?`, empresaID).Scan(&asientosCount); err != nil {
		t.Fatalf("count asientos after second run: %v", err)
	}
	if asientosCount != 1 {
		t.Fatalf("expected asientos count still 1, got %d", asientosCount)
	}

	eventos, err := ListEmpresaEventosContables(dbConn, empresaID, EmpresaEventoContableFilter{Limit: 10})
	if err != nil {
		t.Fatalf("list eventos after processing: %v", err)
	}
	if len(eventos) != 1 || !eventos[0].Procesado {
		t.Fatalf("expected processed event in listing")
	}
	if eventos[0].AsientoContableID <= 0 {
		t.Fatalf("expected asiento_contable_id assigned, got %d", eventos[0].AsientoContableID)
	}

	asientos, err := ListEmpresaAsientosContables(dbConn, empresaID, EmpresaAsientoContableFilter{Limit: 10})
	if err != nil {
		t.Fatalf("list asientos: %v", err)
	}
	if len(asientos) != 1 {
		t.Fatalf("expected 1 asiento, got %d", len(asientos))
	}
	if asientos[0].EventoContableID != eventoID {
		t.Fatalf("expected asiento linked to evento %d, got %d", eventoID, asientos[0].EventoContableID)
	}
	if asientos[0].TotalDebito != 120000 || asientos[0].TotalCredito != 120000 {
		t.Fatalf("expected asiento totals 120000/120000, got %.2f/%.2f", asientos[0].TotalDebito, asientos[0].TotalCredito)
	}
}

func TestProcessEmpresaEventosContablesPendientesConPoliticaRespetaMaxReintentos(t *testing.T) {
	dbConn := openFinanzasTestDB(t)
	if err := EnsureEmpresaFinanzasSchema(dbConn); err != nil {
		t.Fatalf("ensure finanzas schema: %v", err)
	}
	if err := EnsureEmpresaEventosContablesSchema(dbConn); err != nil {
		t.Fatalf("ensure eventos contables schema: %v", err)
	}

	empresaID := int64(33)
	idProcesable, err := CreateEmpresaEventoContable(dbConn, EmpresaEventoContable{
		EmpresaID:       empresaID,
		Modulo:          "finanzas",
		Evento:          "movimiento_ingreso_registrado",
		Entidad:         "finanzas_movimiento",
		EntidadID:       3301,
		DocumentoTipo:   "comprobante",
		DocumentoCodigo: "ING-3301",
		PeriodoContable: "2026-04",
		MontoTotal:      90000,
		Moneda:          "COP",
		PayloadJSON:     `{"tipo_movimiento":"ingreso","categoria":"ventas"}`,
		UsuarioCreador:  "tester",
	})
	if err != nil {
		t.Fatalf("create procesable evento: %v", err)
	}

	idBloqueado, err := CreateEmpresaEventoContable(dbConn, EmpresaEventoContable{
		EmpresaID:       empresaID,
		Modulo:          "finanzas",
		Evento:          "movimiento_egreso_registrado",
		Entidad:         "finanzas_movimiento",
		EntidadID:       3302,
		DocumentoTipo:   "comprobante",
		DocumentoCodigo: "EGR-3302",
		PeriodoContable: "2026-04",
		MontoTotal:      45000,
		Moneda:          "COP",
		PayloadJSON:     `{"tipo_movimiento":"egreso","categoria":"compras"}`,
		UsuarioCreador:  "tester",
	})
	if err != nil {
		t.Fatalf("create bloqueado evento: %v", err)
	}

	if _, err := dbConn.Exec(`UPDATE empresa_eventos_contables SET intentos_procesamiento = 5 WHERE id = ?`, idBloqueado); err != nil {
		t.Fatalf("set intentos_procesamiento for blocked event: %v", err)
	}

	resultado, err := ProcessEmpresaEventosContablesPendientesConPolitica(dbConn, empresaID, "tester", 20, 5)
	if err != nil {
		t.Fatalf("process eventos with policy: %v", err)
	}
	if resultado.EventosRevisados != 1 {
		t.Fatalf("expected eventos_revisados=1 with max_reintentos=5, got %d", resultado.EventosRevisados)
	}
	if resultado.EventosProcesados != 1 {
		t.Fatalf("expected eventos_procesados=1, got %d", resultado.EventosProcesados)
	}

	var procesadoFlag int64
	if err := dbConn.QueryRow(`SELECT COALESCE(procesado, 0) FROM empresa_eventos_contables WHERE id = ?`, idProcesable).Scan(&procesadoFlag); err != nil {
		t.Fatalf("read procesado for procesable event: %v", err)
	}
	if procesadoFlag != 1 {
		t.Fatalf("expected procesable event as procesado=1, got %d", procesadoFlag)
	}

	if err := dbConn.QueryRow(`SELECT COALESCE(procesado, 0) FROM empresa_eventos_contables WHERE id = ?`, idBloqueado).Scan(&procesadoFlag); err != nil {
		t.Fatalf("read procesado for blocked event: %v", err)
	}
	if procesadoFlag != 0 {
		t.Fatalf("expected blocked event to remain procesado=0, got %d", procesadoFlag)
	}
}

func TestGetEmpresaConciliacionContablePorPeriodo(t *testing.T) {
	dbConn := openFinanzasTestDB(t)
	if err := EnsureEmpresaFinanzasSchema(dbConn); err != nil {
		t.Fatalf("ensure finanzas schema: %v", err)
	}
	if err := EnsureEmpresaEventosContablesSchema(dbConn); err != nil {
		t.Fatalf("ensure eventos contables schema: %v", err)
	}

	empresaID := int64(44)
	if _, err := CreateEmpresaEventoContable(dbConn, EmpresaEventoContable{
		EmpresaID:       empresaID,
		Modulo:          "finanzas",
		Evento:          "movimiento_ingreso_registrado",
		Entidad:         "finanzas_movimiento",
		EntidadID:       4401,
		DocumentoTipo:   "comprobante",
		DocumentoCodigo: "ING-4401",
		PeriodoContable: "2026-04",
		MontoTotal:      100000,
		Moneda:          "COP",
		PayloadJSON:     `{"tipo_movimiento":"ingreso","categoria":"ventas"}`,
		UsuarioCreador:  "tester",
	}); err != nil {
		t.Fatalf("create primer evento: %v", err)
	}
	pendienteID, err := CreateEmpresaEventoContable(dbConn, EmpresaEventoContable{
		EmpresaID:       empresaID,
		Modulo:          "finanzas",
		Evento:          "movimiento_egreso_registrado",
		Entidad:         "finanzas_movimiento",
		EntidadID:       4402,
		DocumentoTipo:   "comprobante",
		DocumentoCodigo: "EGR-4402",
		PeriodoContable: "2026-04",
		MontoTotal:      35000,
		Moneda:          "COP",
		PayloadJSON:     `{"tipo_movimiento":"egreso","categoria":"compras"}`,
		UsuarioCreador:  "tester",
	})
	if err != nil {
		t.Fatalf("create segundo evento: %v", err)
	}

	resultado, err := ProcessEmpresaEventosContablesPendientes(dbConn, empresaID, "tester", 1)
	if err != nil {
		t.Fatalf("process eventos pendientes: %v", err)
	}
	if resultado.EventosProcesados != 1 {
		t.Fatalf("expected eventos_procesados=1, got %d", resultado.EventosProcesados)
	}

	if _, err := dbConn.Exec(`UPDATE empresa_eventos_contables SET error_procesamiento = 'fallo temporal', intentos_procesamiento = 2 WHERE id = ?`, pendienteID); err != nil {
		t.Fatalf("mark pendiente con error: %v", err)
	}

	resumen, err := GetEmpresaConciliacionContablePorPeriodo(dbConn, empresaID, EmpresaConciliacionContableFilter{PeriodoContable: "2026-04", Limit: 12})
	if err != nil {
		t.Fatalf("get conciliacion por periodo: %v", err)
	}
	if resumen.TotalPeriodos != 1 {
		t.Fatalf("expected total_periodos=1, got %d", resumen.TotalPeriodos)
	}
	if len(resumen.Filas) != 1 {
		t.Fatalf("expected 1 fila de conciliacion, got %d", len(resumen.Filas))
	}

	fila := resumen.Filas[0]
	if fila.PeriodoContable != "2026-04" {
		t.Fatalf("expected periodo_contable=2026-04, got %q", fila.PeriodoContable)
	}
	if fila.EventosTotal != 2 {
		t.Fatalf("expected eventos_total=2, got %d", fila.EventosTotal)
	}
	if fila.EventosProcesados != 1 {
		t.Fatalf("expected eventos_procesados=1, got %d", fila.EventosProcesados)
	}
	if fila.EventosPendientes != 1 {
		t.Fatalf("expected eventos_pendientes=1, got %d", fila.EventosPendientes)
	}
	if fila.EventosConError != 1 {
		t.Fatalf("expected eventos_con_error=1, got %d", fila.EventosConError)
	}
	if fila.AsientosTotal != 1 {
		t.Fatalf("expected asientos_total=1, got %d", fila.AsientosTotal)
	}
	if fila.EstadoConciliacion != "con_pendientes" {
		t.Fatalf("expected estado_conciliacion=con_pendientes, got %q", fila.EstadoConciliacion)
	}
}

func TestProcessEmpresaEventosContablesPendientesCreditoAbonoGeneraLineasCartera(t *testing.T) {
	dbConn := openFinanzasTestDB(t)
	if err := EnsureEmpresaFinanzasSchema(dbConn); err != nil {
		t.Fatalf("ensure finanzas schema: %v", err)
	}
	if err := EnsureEmpresaEventosContablesSchema(dbConn); err != nil {
		t.Fatalf("ensure eventos contables schema: %v", err)
	}

	empresaID := int64(55)
	monto := 100000.0
	eventoID, err := CreateEmpresaEventoContable(dbConn, EmpresaEventoContable{
		EmpresaID:       empresaID,
		Modulo:          "creditos",
		Evento:          "credito_abono_registrado",
		Entidad:         "credito_movimiento",
		EntidadID:       5501,
		DocumentoTipo:   "credito",
		DocumentoCodigo: "CR-5501",
		PeriodoContable: "2026-04",
		MontoTotal:      monto,
		Moneda:          "COP",
		PayloadJSON:     `{"capital_aplicado":80000,"interes_aplicado":15000,"mora_aplicada":5000,"canal_pago":"bancos"}`,
		UsuarioCreador:  "tester",
	})
	if err != nil {
		t.Fatalf("create credito evento contable: %v", err)
	}

	resultado, err := ProcessEmpresaEventosContablesPendientes(dbConn, empresaID, "tester", 20)
	if err != nil {
		t.Fatalf("process eventos pendientes: %v", err)
	}
	if resultado.EventosProcesados != 1 || resultado.AsientosCreados != 1 {
		t.Fatalf("expected 1 evento procesado y 1 asiento creado, got %+v", resultado)
	}

	asientos, err := ListEmpresaAsientosContables(dbConn, empresaID, EmpresaAsientoContableFilter{Modulo: "creditos", Evento: "credito_abono_registrado", Limit: 10})
	if err != nil {
		t.Fatalf("list asientos creditos: %v", err)
	}
	if len(asientos) != 1 {
		t.Fatalf("expected 1 asiento de creditos, got %d", len(asientos))
	}
	asiento := asientos[0]
	if asiento.EventoContableID != eventoID {
		t.Fatalf("expected asiento linked to evento=%d, got %d", eventoID, asiento.EventoContableID)
	}
	if asiento.TotalDebito != monto || asiento.TotalCredito != monto {
		t.Fatalf("expected totals %.2f/%.2f, got %.2f/%.2f", monto, monto, asiento.TotalDebito, asiento.TotalCredito)
	}

	var lineas []EmpresaAsientoContableLinea
	if err := json.Unmarshal([]byte(asiento.LineasJSON), &lineas); err != nil {
		t.Fatalf("decode lineas_json: %v", err)
	}
	if len(lineas) < 3 {
		t.Fatalf("expected >=3 lineas para capital/interes/mora, got %d", len(lineas))
	}

	tieneCaja := false
	tieneCartera := false
	tieneInteres := false
	tieneMora := false
	for _, ln := range lineas {
		if ln.Cuenta == "110505" && ln.Debito > 0 {
			tieneCaja = true
		}
		if ln.Cuenta == "130505" && ln.Credito > 0 {
			tieneCartera = true
		}
		if ln.Descripcion == "Ingresos por interes de credito" && ln.Credito > 0 {
			tieneInteres = true
		}
		if ln.Descripcion == "Ingresos por interes de mora" && ln.Credito > 0 {
			tieneMora = true
		}
	}
	if !tieneCaja || !tieneCartera || !tieneInteres || !tieneMora {
		t.Fatalf("expected lineas caja/cartera/interes/mora, got %+v", lineas)
	}
}
