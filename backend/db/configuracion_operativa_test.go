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
