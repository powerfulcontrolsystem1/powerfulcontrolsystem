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

	result, err := RegisterEmpresaComisionesServicioDesdeCarrito(dbConn, 1, carritoID, "lavador1@empresa.com", "cajero1@empresa.com", "cajero")
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

func TestEmpresaComisionesServicioEscalaConTopePorRol(t *testing.T) {
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
		PorcentajeComision:     10,
		FiltroServicio:         "lavado",
		AplicarAutomaticamente: true,
		UsuarioCreador:         "qa@empresa.com",
		Estado:                 "activo",
	}); err != nil {
		t.Fatalf("upsert comisiones config: %v", err)
	}

	if _, err := CreateEmpresaComisionServicioEscala(dbConn, EmpresaComisionServicioEscala{
		EmpresaID:          1,
		RolOperacion:       "cajero",
		ServicioFiltro:     "premium",
		PorcentajeComision: 30,
		TopeComision:       2500,
		Prioridad:          1,
		UsuarioCreador:     "qa@empresa.com",
		Estado:             "activo",
	}); err != nil {
		t.Fatalf("create escala comision: %v", err)
	}

	servicioID, err := CreateServicio(dbConn, Servicio{
		EmpresaID:          1,
		Codigo:             "LAV-PRM",
		Nombre:             "Lavado premium extremo",
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
		Codigo:        "EST-1-TOPE",
		Nombre:        "Caja Tope",
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
		CodigoItem:          "LAV-PRM",
		Descripcion:         "Lavado premium extremo",
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

	result, err := RegisterEmpresaComisionesServicioDesdeCarrito(dbConn, 1, carritoID, "emp-001", "cajero1", "cajero")
	if err != nil {
		t.Fatalf("register comisiones desde carrito: %v", err)
	}
	if result == nil || !result.Aplicada {
		t.Fatalf("expected aplicada=true, got %+v", result)
	}
	if math.Abs(result.MontoComision-2500) > 0.001 {
		t.Fatalf("expected comision con tope 2500, got %.4f", result.MontoComision)
	}
	if math.Abs(result.TotalTopesAplicados-500) > 0.001 {
		t.Fatalf("expected topes aplicados 500, got %.4f", result.TotalTopesAplicados)
	}

	movs, err := ListEmpresaComisionServicioMovimientos(dbConn, 1, EmpresaComisionServicioMovimientoFilter{Limit: 10})
	if err != nil {
		t.Fatalf("list movimientos: %v", err)
	}
	if len(movs) != 1 {
		t.Fatalf("expected 1 movimiento, got %d", len(movs))
	}
	if movs[0].EscalaID <= 0 {
		t.Fatalf("expected escala_id > 0, got %d", movs[0].EscalaID)
	}
	if math.Abs(movs[0].MontoComisionBruto-3000) > 0.001 {
		t.Fatalf("expected monto_comision_bruto 3000, got %.4f", movs[0].MontoComisionBruto)
	}
}

func TestEmpresaComisionesServicioAjusteManualConAprobacion(t *testing.T) {
	dbConn := openCarritoInventarioTestDB(t)
	if err := EnsureEmpresaComisionesServicioSchema(dbConn); err != nil {
		t.Fatalf("ensure comisiones schema: %v", err)
	}

	id, err := CreateEmpresaComisionServicioAjusteManual(dbConn, EmpresaComisionServicioMovimiento{
		EmpresaID:      1,
		UsuarioOrigen:  "supervisor@empresa.com",
		UsuarioLavador: "emp-001",
		MontoComision:  800,
		UsuarioCreador: "supervisor@empresa.com",
		Observaciones:  "Ajuste por servicio fuera de caja",
	})
	if err != nil {
		t.Fatalf("create ajuste manual: %v", err)
	}

	pendientes, err := ListEmpresaComisionServicioMovimientos(dbConn, 1, EmpresaComisionServicioMovimientoFilter{
		SoloPendientes:  true,
		IncludeInactive: true,
		Limit:           20,
	})
	if err != nil {
		t.Fatalf("list pendientes: %v", err)
	}
	if len(pendientes) != 1 {
		t.Fatalf("expected 1 pendiente, got %d", len(pendientes))
	}
	if pendientes[0].AjusteEstado != EmpresaComisionServicioAjustePendiente {
		t.Fatalf("expected ajuste pendiente, got %s", pendientes[0].AjusteEstado)
	}

	resuelto, err := ResolverEmpresaComisionServicioAjusteManual(dbConn, 1, id, true, "gerencia@empresa.com", "Aprobado por gerencia")
	if err != nil {
		t.Fatalf("resolver ajuste manual: %v", err)
	}
	if resuelto.AjusteEstado != EmpresaComisionServicioAjusteAprobado {
		t.Fatalf("expected ajuste aprobado, got %s", resuelto.AjusteEstado)
	}
	if resuelto.Estado != "activo" {
		t.Fatalf("expected estado activo, got %s", resuelto.Estado)
	}
	if resuelto.AprobadoPor != "gerencia@empresa.com" {
		t.Fatalf("expected aprobado_por gerencia@empresa.com, got %s", resuelto.AprobadoPor)
	}
}

func TestEmpresaComisionesServicioVinculoLiquidacion(t *testing.T) {
	dbConn := openCarritoInventarioTestDB(t)
	if err := EnsureEmpresaComisionesServicioSchema(dbConn); err != nil {
		t.Fatalf("ensure comisiones schema: %v", err)
	}

	mov1, err := CreateEmpresaComisionServicioMovimiento(dbConn, EmpresaComisionServicioMovimiento{
		EmpresaID:          1,
		UsuarioOrigen:      "cajero@empresa.com",
		UsuarioLavador:     "emp-001",
		BaseServicio:       10000,
		PorcentajeComision: 10,
		MontoComision:      1000,
		UsuarioCreador:     "cajero@empresa.com",
		Estado:             "activo",
	})
	if err != nil {
		t.Fatalf("create movimiento 1: %v", err)
	}
	mov2, err := CreateEmpresaComisionServicioMovimiento(dbConn, EmpresaComisionServicioMovimiento{
		EmpresaID:          1,
		UsuarioOrigen:      "cajero@empresa.com",
		UsuarioLavador:     "emp-001",
		BaseServicio:       5000,
		PorcentajeComision: 10,
		MontoComision:      500,
		UsuarioCreador:     "cajero@empresa.com",
		Estado:             "activo",
	})
	if err != nil {
		t.Fatalf("create movimiento 2: %v", err)
	}

	if _, err := dbConn.Exec(`UPDATE empresa_comisiones_servicio_movimientos
		SET fecha_movimiento = '2026-04-08 10:00:00'
		WHERE id IN (?, ?)`, mov1, mov2); err != nil {
		t.Fatalf("update fecha movimientos: %v", err)
	}

	resumen, err := GetEmpresaComisionServicioLiquidacionResumen(dbConn, 1, []string{"emp-001"}, "2026-04-01", "2026-04-30")
	if err != nil {
		t.Fatalf("get resumen liquidacion: %v", err)
	}
	if resumen.CantidadMovimientos != 2 {
		t.Fatalf("expected 2 movimientos, got %d", resumen.CantidadMovimientos)
	}
	if math.Abs(resumen.TotalComisiones-1500) > 0.001 {
		t.Fatalf("expected total comisiones 1500, got %.4f", resumen.TotalComisiones)
	}

	if err := VincularEmpresaComisionesServicioALiquidacion(dbConn, 1, 77, "2026-04-01", "2026-04-30", "nomina@empresa.com", resumen.MovimientoIDs); err != nil {
		t.Fatalf("vincular comisiones liquidacion: %v", err)
	}

	vinculados, err := ListEmpresaComisionServicioMovimientos(dbConn, 1, EmpresaComisionServicioMovimientoFilter{
		LiquidacionNominaID: 77,
		Limit:               20,
	})
	if err != nil {
		t.Fatalf("list vinculados: %v", err)
	}
	if len(vinculados) != 2 {
		t.Fatalf("expected 2 vinculados, got %d", len(vinculados))
	}

	if err := LimpiarVinculoEmpresaComisionesServicioLiquidacion(dbConn, 1, 77); err != nil {
		t.Fatalf("limpiar vinculo liquidacion: %v", err)
	}

	noLiquidados, err := ListEmpresaComisionServicioMovimientos(dbConn, 1, EmpresaComisionServicioMovimientoFilter{
		NoLiquidado: true,
		Limit:       20,
	})
	if err != nil {
		t.Fatalf("list no liquidados: %v", err)
	}
	if len(noLiquidados) != 2 {
		t.Fatalf("expected 2 no liquidados tras limpieza, got %d", len(noLiquidados))
	}
}
