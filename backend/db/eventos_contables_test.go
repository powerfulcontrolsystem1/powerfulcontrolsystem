package db

import "testing"

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
