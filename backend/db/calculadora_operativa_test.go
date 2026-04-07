package db

import "testing"

func TestEmpresaCalculadoraConfiguracionYHistorialFlow(t *testing.T) {
	dbConn := openCarritoInventarioTestDB(t)
	if err := EnsureEmpresaCalculadoraSchema(dbConn); err != nil {
		t.Fatalf("ensure calculadora schema: %v", err)
	}

	cfg, err := GetEmpresaCalculadoraConfiguracion(dbConn, 31)
	if err != nil {
		t.Fatalf("get default config: %v", err)
	}
	if cfg == nil || !cfg.IntegrarCarritos || !cfg.IntegrarCotizaciones {
		t.Fatalf("expected default integrations enabled, got %+v", cfg)
	}

	if _, err := UpsertEmpresaCalculadoraConfiguracion(dbConn, EmpresaCalculadoraConfiguracion{
		EmpresaID:            31,
		IntegrarCarritos:     false,
		IntegrarCotizaciones: true,
		UsuarioCreador:       "qa_calculadora",
		Estado:               "activo",
		Observaciones:        "prueba config",
	}); err != nil {
		t.Fatalf("upsert config: %v", err)
	}

	updatedCfg, err := GetEmpresaCalculadoraConfiguracion(dbConn, 31)
	if err != nil {
		t.Fatalf("get updated config: %v", err)
	}
	if updatedCfg.IntegrarCarritos {
		t.Fatalf("expected integrar_carritos=false, got %+v", updatedCfg)
	}

	idA, err := CreateEmpresaCalculadoraOperacion(dbConn, EmpresaCalculadoraOperacion{
		EmpresaID:       31,
		Expresion:       "10+20",
		Resultado:       "30",
		Etiquetas:       []string{"cierre", "turno"},
		FechaOperacion:  "2026-04-07 09:00:00",
		UsuarioCreador:  "ana@empresa.com",
		ClienteID:       100,
		DocumentoTipo:   "carrito",
		DocumentoCodigo: "CAR-001",
		Estado:          "activo",
	})
	if err != nil {
		t.Fatalf("create op A: %v", err)
	}
	if idA <= 0 {
		t.Fatalf("expected op A id > 0, got %d", idA)
	}

	idB, err := CreateEmpresaCalculadoraOperacion(dbConn, EmpresaCalculadoraOperacion{
		EmpresaID:      31,
		Expresion:      "50*2",
		Resultado:      "100",
		Etiquetas:      []string{"apertura"},
		FechaOperacion: "2026-04-08 11:00:00",
		UsuarioCreador: "luis@empresa.com",
		ClienteID:      101,
		Estado:         "activo",
	})
	if err != nil {
		t.Fatalf("create op B: %v", err)
	}
	if idB <= 0 {
		t.Fatalf("expected op B id > 0, got %d", idB)
	}

	rowsUsuario, totalUsuario, err := ListEmpresaCalculadoraOperaciones(dbConn, 31, EmpresaCalculadoraOperacionFilter{
		UsuarioCreador: "ana@empresa.com",
		Desde:          "2026-04-01",
		Hasta:          "2026-04-30",
		Limit:          20,
		Offset:         0,
	})
	if err != nil {
		t.Fatalf("list by usuario: %v", err)
	}
	if totalUsuario != 1 || len(rowsUsuario) != 1 {
		t.Fatalf("expected one row for ana, total=%d len=%d", totalUsuario, len(rowsUsuario))
	}
	if rowsUsuario[0].ID != idA {
		t.Fatalf("expected row id=%d, got=%d", idA, rowsUsuario[0].ID)
	}

	desactivados, err := SetEmpresaCalculadoraOperacionesEstado(dbConn, 31, EmpresaCalculadoraOperacionFilter{
		UsuarioCreador: "ana@empresa.com",
		Desde:          "2026-04-01",
		Hasta:          "2026-04-30",
	}, "inactivo")
	if err != nil {
		t.Fatalf("set estado filtered: %v", err)
	}
	if desactivados != 1 {
		t.Fatalf("expected 1 deactivated row, got %d", desactivados)
	}

	rowsActivas, totalActivas, err := ListEmpresaCalculadoraOperaciones(dbConn, 31, EmpresaCalculadoraOperacionFilter{
		IncludeInactive: false,
		Limit:           20,
		Offset:          0,
	})
	if err != nil {
		t.Fatalf("list active rows: %v", err)
	}
	if totalActivas != 1 || len(rowsActivas) != 1 {
		t.Fatalf("expected one active row after clear, total=%d len=%d", totalActivas, len(rowsActivas))
	}
	if rowsActivas[0].ID != idB {
		t.Fatalf("expected active row id=%d, got=%d", idB, rowsActivas[0].ID)
	}
}
