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
