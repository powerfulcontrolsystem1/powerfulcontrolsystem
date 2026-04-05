package db

import (
	"math"
	"testing"
)

func TestEmpresaComisionesServicioConfigUpsertAndGet(t *testing.T) {
	dbConn := openCarritoInventarioTestDB(t)
	if err := EnsureEmpresaComisionesServicioSchema(dbConn); err != nil {
		t.Fatalf("ensure comisiones schema: %v", err)
	}

	id, err := UpsertEmpresaComisionesServicioConfiguracion(dbConn, EmpresaComisionesServicioConfiguracion{
		EmpresaID:              1,
		HabilitarComisiones:    true,
		PorcentajeComision:     18,
		FiltroServicio:         "lavado",
		AplicarAutomaticamente: true,
		UsuarioCreador:         "contabilidad@empresa.com",
		Estado:                 "activo",
	})
	if err != nil {
		t.Fatalf("upsert comisiones config: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected id > 0, got %d", id)
	}

	cfg, err := GetEmpresaComisionesServicioConfiguracion(dbConn, 1)
	if err != nil {
		t.Fatalf("get comisiones config: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected config not nil")
	}
	if !cfg.HabilitarComisiones {
		t.Fatal("expected habilitar_comisiones=true")
	}
	if math.Abs(cfg.PorcentajeComision-18) > 0.001 {
		t.Fatalf("expected porcentaje_comision 18, got %.4f", cfg.PorcentajeComision)
	}
	if cfg.FiltroServicio != "lavado" {
		t.Fatalf("expected filtro_servicio=lavado, got %q", cfg.FiltroServicio)
	}
}

func TestEmpresaComisionesServicioRegistroDesdeCarritoYReporte(t *testing.T) {
	dbConn := openCarritoInventarioTestDB(t)
	if err := EnsureEmpresaCarritosSchema(dbConn); err != nil {
		t.Fatalf("ensure carritos schema: %v", err)
	}
	if err := EnsureEmpresaProductosSchema(dbConn); err != nil {
		t.Fatalf("ensure productos schema: %v", err)
	}
	if err := EnsureEmpresaComisionesServicioSchema(dbConn); err != nil {
		t.Fatalf("ensure comisiones schema: %v", err)
	}

	if _, err := UpsertEmpresaComisionesServicioConfiguracion(dbConn, EmpresaComisionesServicioConfiguracion{
		EmpresaID:              1,
		HabilitarComisiones:    true,
		PorcentajeComision:     20,
		FiltroServicio:         "lavado",
		AplicarAutomaticamente: true,
		UsuarioCreador:         "qa@empresa.com",
		Estado:                 "activo",
	}); err != nil {
		t.Fatalf("upsert comisiones config: %v", err)
	}

	servicioID, err := CreateServicio(dbConn, Servicio{
		EmpresaID:          1,
		Codigo:             "LAV-001",
		Nombre:             "Lavado premium",
		Categoria:          "lavado",
		Precio:             10000,
		ImpuestoPorcentaje: 0,
		UsuarioCreador:     "qa@empresa.com",
		Estado:             "activo",
	})
	if err != nil {
		t.Fatalf("create servicio: %v", err)
	}

	carritoID, err := CreateCarritoCompra(dbConn, CarritoCompra{
		EmpresaID:     1,
		Codigo:        "EST-1-001",
		Nombre:        "Caja Lavadero 1",
		CanalVenta:    "mostrador",
		Moneda:        "COP",
		Estado:        "activo",
		EstadoCarrito: "abierto",
	})
	if err != nil {
		t.Fatalf("create carrito: %v", err)
	}

	if _, err := CreateCarritoCompraItem(dbConn, CarritoCompraItem{
		EmpresaID:           1,
		CarritoID:           carritoID,
		TipoItem:            "servicio",
		ReferenciaID:        servicioID,
		CodigoItem:          "LAV-001",
		Descripcion:         "Lavado premium de auto",
		UnidadMedida:        "servicio",
		Cantidad:            1,
		PrecioUnitario:      10000,
		DescuentoPorcentaje: 0,
		ImpuestoPorcentaje:  0,
		ImpuestoCodigo:      "IVA",
		UsuarioCreador:      "qa@empresa.com",
		Estado:              "activo",
	}); err != nil {
		t.Fatalf("create servicio item: %v", err)
	}

	result, err := RegisterEmpresaComisionesServicioDesdeCarrito(dbConn, 1, carritoID, "lavador1@empresa.com", "cajero1@empresa.com")
	if err != nil {
		t.Fatalf("register comisiones desde carrito: %v", err)
	}
	if result == nil {
		t.Fatal("expected result not nil")
	}
	if !result.Aplicada {
		t.Fatalf("expected aplicada=true, got %+v", result)
	}
	if result.MovimientosRegistrados != 1 {
		t.Fatalf("expected 1 movimiento, got %d", result.MovimientosRegistrados)
	}
	if math.Abs(result.BaseServicios-10000) > 0.001 {
		t.Fatalf("expected base_servicios 10000, got %.4f", result.BaseServicios)
	}
	if math.Abs(result.MontoComision-2000) > 0.001 {
		t.Fatalf("expected monto_comision 2000, got %.4f", result.MontoComision)
	}

	report, err := GetEmpresaComisionesServicioReporte(dbConn, 1, EmpresaComisionServicioMovimientoFilter{Limit: 50})
	if err != nil {
		t.Fatalf("get reporte comisiones: %v", err)
	}
	if report == nil {
		t.Fatal("expected report not nil")
	}
	if report.Resumen.CantidadMovimientos != 1 {
		t.Fatalf("expected cantidad_movimientos 1, got %d", report.Resumen.CantidadMovimientos)
	}
	if len(report.Lavadores) != 1 {
		t.Fatalf("expected 1 row in lavadores, got %d", len(report.Lavadores))
	}
	if report.Lavadores[0].UsuarioLavador != "lavador1@empresa.com" {
		t.Fatalf("expected lavador lavador1@empresa.com, got %q", report.Lavadores[0].UsuarioLavador)
	}
	if math.Abs(report.Lavadores[0].TotalComision-2000) > 0.001 {
		t.Fatalf("expected total comision 2000, got %.4f", report.Lavadores[0].TotalComision)
	}
}
