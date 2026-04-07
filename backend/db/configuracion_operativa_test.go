package db

import "testing"

func TestEmpresaConfiguracionOperativaDefaultYResolucionPorRol(t *testing.T) {
	dbConn := openCarritoInventarioTestDB(t)
	if err := EnsureEmpresaConfiguracionOperativaSchema(dbConn); err != nil {
		t.Fatalf("ensure configuracion operativa schema: %v", err)
	}

	cfg, err := GetEmpresaConfiguracionOperativa(dbConn, 1)
	if err != nil {
		t.Fatalf("get default config operativa: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected cfg not nil")
	}
	if !cfg.MetodoPagoEfectivo || !cfg.MetodoPagoTarjetaCredito || !cfg.MetodoPagoTarjetaDebito || !cfg.MetodoPagoTransferenciaBancaria || !cfg.MetodoPagoMixto || !cfg.MetodoPagoCodigoDescuento {
		t.Fatalf("expected all payment methods enabled by default, got %+v", cfg)
	}
	if !cfg.HabilitarPropinas || !cfg.HabilitarComisiones {
		t.Fatalf("expected propinas/comisiones enabled by default, got %+v", cfg)
	}

	if _, err := UpsertEmpresaConfiguracionOperativa(dbConn, EmpresaConfiguracionOperativa{
		EmpresaID:                       1,
		MetodoPagoEfectivo:              true,
		MetodoPagoTarjetaCredito:        true,
		MetodoPagoTarjetaDebito:         true,
		MetodoPagoTransferenciaBancaria: false,
		MetodoPagoMixto:                 true,
		MetodoPagoCodigoDescuento:       false,
		HabilitarPropinas:               true,
		HabilitarComisiones:             true,
		UsuarioCreador:                  "qa@empresa.com",
		Estado:                          "activo",
	}); err != nil {
		t.Fatalf("upsert base config operativa: %v", err)
	}

	if _, err := UpsertEmpresaConfiguracionOperativaRol(dbConn, EmpresaConfiguracionOperativaRol{
		EmpresaID:                       1,
		Rol:                             "cajero",
		MetodoPagoEfectivo:              true,
		MetodoPagoTarjetaCredito:        false,
		MetodoPagoTarjetaDebito:         false,
		MetodoPagoTransferenciaBancaria: false,
		MetodoPagoMixto:                 false,
		MetodoPagoCodigoDescuento:       false,
		HabilitarPropinas:               false,
		HabilitarComisiones:             false,
		UsuarioCreador:                  "qa@empresa.com",
		Estado:                          "activo",
	}); err != nil {
		t.Fatalf("upsert role config operativa: %v", err)
	}

	cfg, err = GetEmpresaConfiguracionOperativa(dbConn, 1)
	if err != nil {
		t.Fatalf("get config after upserts: %v", err)
	}
	if len(cfg.Roles) != 1 {
		t.Fatalf("expected 1 role row, got %d", len(cfg.Roles))
	}

	resolvedCajero := ResolveEmpresaConfiguracionOperativaParaRol(cfg, "cajero")
	if !resolvedCajero.MetodoPagoEfectivo {
		t.Fatal("expected efectivo enabled for cajero")
	}
	if resolvedCajero.MetodoPagoTarjetaCredito || resolvedCajero.MetodoPagoTarjetaDebito || resolvedCajero.MetodoPagoTransferenciaBancaria || resolvedCajero.MetodoPagoMixto || resolvedCajero.MetodoPagoCodigoDescuento {
		t.Fatalf("expected card/transfer/mixto/codigo disabled for cajero, got %+v", resolvedCajero)
	}
	if resolvedCajero.HabilitarPropinas || resolvedCajero.HabilitarComisiones {
		t.Fatalf("expected propinas/comisiones disabled for cajero, got %+v", resolvedCajero)
	}

	resolvedAdmin := ResolveEmpresaConfiguracionOperativaParaRol(cfg, "admin_empresa")
	if !resolvedAdmin.MetodoPagoTarjetaCredito || !resolvedAdmin.MetodoPagoTarjetaDebito {
		t.Fatalf("expected tarjetas enabled for admin_empresa based on company config, got %+v", resolvedAdmin)
	}
	if resolvedAdmin.MetodoPagoCodigoDescuento {
		t.Fatalf("expected codigo_descuento disabled at company level for admin_empresa, got %+v", resolvedAdmin)
	}
}

func TestEmpresaConfiguracionOperativaPoliticaContextoYRollback(t *testing.T) {
	dbConn := openCarritoInventarioTestDB(t)
	if err := EnsureEmpresaConfiguracionOperativaSchema(dbConn); err != nil {
		t.Fatalf("ensure configuracion operativa schema: %v", err)
	}

	if _, err := UpsertEmpresaConfiguracionOperativa(dbConn, EmpresaConfiguracionOperativa{
		EmpresaID:                       9,
		MetodoPagoEfectivo:              true,
		MetodoPagoTarjetaCredito:        true,
		MetodoPagoTarjetaDebito:         true,
		MetodoPagoTransferenciaBancaria: true,
		MetodoPagoMixto:                 true,
		MetodoPagoCodigoDescuento:       false,
		HabilitarPropinas:               true,
		HabilitarComisiones:             true,
		UsuarioCreador:                  "qa@empresa.com",
		Estado:                          "activo",
	}); err != nil {
		t.Fatalf("upsert base config: %v", err)
	}

	if _, err := UpsertEmpresaConfiguracionOperativaRol(dbConn, EmpresaConfiguracionOperativaRol{
		EmpresaID:                       9,
		Rol:                             "cajero",
		MetodoPagoEfectivo:              true,
		MetodoPagoTarjetaCredito:        false,
		MetodoPagoTarjetaDebito:         false,
		MetodoPagoTransferenciaBancaria: false,
		MetodoPagoMixto:                 false,
		MetodoPagoCodigoDescuento:       false,
		HabilitarPropinas:               false,
		HabilitarComisiones:             false,
		UsuarioCreador:                  "qa@empresa.com",
		Estado:                          "activo",
	}); err != nil {
		t.Fatalf("upsert role config: %v", err)
	}

	if _, err := UpsertEmpresaConfiguracionOperativaPolitica(dbConn, EmpresaConfiguracionOperativaPolitica{
		EmpresaID:                       9,
		CanalVenta:                      "app",
		SucursalID:                      7,
		Turno:                           "noche",
		Prioridad:                       10,
		MetodoPagoEfectivo:              false,
		MetodoPagoTarjetaCredito:        true,
		MetodoPagoTarjetaDebito:         true,
		MetodoPagoTransferenciaBancaria: false,
		MetodoPagoMixto:                 false,
		MetodoPagoCodigoDescuento:       false,
		HabilitarPropinas:               true,
		HabilitarComisiones:             false,
		UsuarioCreador:                  "qa@empresa.com",
		Estado:                          "activo",
	}); err != nil {
		t.Fatalf("upsert politica contextual: %v", err)
	}

	cfg, err := GetEmpresaConfiguracionOperativa(dbConn, 9)
	if err != nil {
		t.Fatalf("get config after policy: %v", err)
	}
	resolvedPolicy := ResolveEmpresaConfiguracionOperativaConContexto(cfg, EmpresaConfiguracionOperativaContexto{
		Rol:        "cajero",
		CanalVenta: "app",
		SucursalID: 7,
		Turno:      "noche",
	})
	if !resolvedPolicy.PoliticaAplicada || resolvedPolicy.Fuente != "politica" {
		t.Fatalf("expected policy applied, got %+v", resolvedPolicy)
	}
	if resolvedPolicy.MetodoPagoEfectivo || !resolvedPolicy.MetodoPagoTarjetaCredito || !resolvedPolicy.MetodoPagoTarjetaDebito {
		t.Fatalf("expected contextual policy overrides, got %+v", resolvedPolicy)
	}

	resolvedRole := ResolveEmpresaConfiguracionOperativaConContexto(cfg, EmpresaConfiguracionOperativaContexto{
		Rol:        "cajero",
		CanalVenta: "mostrador",
		SucursalID: 3,
		Turno:      "dia",
	})
	if resolvedRole.Fuente != "rol" {
		t.Fatalf("expected role source without matching policy, got %+v", resolvedRole)
	}
	if resolvedRole.MetodoPagoTarjetaCredito {
		t.Fatalf("expected tarjeta credito disabled at role level, got %+v", resolvedRole)
	}

	historialID, err := CreateEmpresaConfiguracionOperativaHistorialSnapshot(dbConn, EmpresaConfiguracionOperativaHistorialSnapshot{
		EmpresaID:      9,
		Evento:         "publicar",
		UsuarioCreador: "qa@empresa.com",
		Estado:         "activo",
		Observaciones:  "snapshot antes de cambios",
	})
	if err != nil {
		t.Fatalf("create historial snapshot: %v", err)
	}
	if historialID <= 0 {
		t.Fatalf("expected historial id > 0, got %d", historialID)
	}

	if _, err := UpsertEmpresaConfiguracionOperativa(dbConn, EmpresaConfiguracionOperativa{
		EmpresaID:                       9,
		MetodoPagoEfectivo:              false,
		MetodoPagoTarjetaCredito:        false,
		MetodoPagoTarjetaDebito:         false,
		MetodoPagoTransferenciaBancaria: false,
		MetodoPagoMixto:                 false,
		MetodoPagoCodigoDescuento:       false,
		HabilitarPropinas:               false,
		HabilitarComisiones:             false,
		UsuarioCreador:                  "qa@empresa.com",
		Estado:                          "activo",
	}); err != nil {
		t.Fatalf("upsert base modified: %v", err)
	}

	rollbackID, err := ApplyEmpresaConfiguracionOperativaRollback(dbConn, 9, historialID, "qa@empresa.com", "rollback prueba")
	if err != nil {
		t.Fatalf("apply rollback: %v", err)
	}
	if rollbackID <= 0 {
		t.Fatalf("expected rollback id > 0, got %d", rollbackID)
	}

	restored, err := GetEmpresaConfiguracionOperativa(dbConn, 9)
	if err != nil {
		t.Fatalf("get restored config: %v", err)
	}
	if !restored.MetodoPagoEfectivo || !restored.MetodoPagoTarjetaCredito {
		t.Fatalf("expected base config restored with enabled methods, got %+v", restored)
	}
	if len(restored.Politicas) == 0 {
		t.Fatalf("expected policies restored after rollback, got %+v", restored)
	}

	historialRows, err := ListEmpresaConfiguracionOperativaHistorialSnapshots(dbConn, 9, 10)
	if err != nil {
		t.Fatalf("list historial after rollback: %v", err)
	}
	if len(historialRows) < 2 {
		t.Fatalf("expected at least 2 historial rows, got %d", len(historialRows))
	}
	if historialRows[0].Evento != "rollback" {
		t.Fatalf("expected latest historial event rollback, got %+v", historialRows[0])
	}
}
