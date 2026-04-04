package db

import (
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func openFinanzasTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "finanzas_test.db")
	dbConn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	dbConn.SetMaxOpenConns(1)
	t.Cleanup(func() {
		_ = dbConn.Close()
	})
	return dbConn
}

func TestEmpresaFinanzasConfiguracionUpsertAndGet(t *testing.T) {
	dbConn := openFinanzasTestDB(t)
	if err := EnsureEmpresaFinanzasSchema(dbConn); err != nil {
		t.Fatalf("ensure finanzas schema: %v", err)
	}

	cfgDefault, err := GetEmpresaFinanzasConfiguracion(dbConn, 77)
	if err != nil {
		t.Fatalf("get default config: %v", err)
	}
	if !cfgDefault.HabilitarIngresos || !cfgDefault.HabilitarEgresos {
		t.Fatalf("expected default ingresos/egresos enabled")
	}

	_, err = UpsertEmpresaFinanzasConfiguracion(dbConn, EmpresaFinanzasConfiguracion{
		EmpresaID:                  77,
		HabilitarIngresos:          true,
		HabilitarEgresos:           true,
		Moneda:                     "COP",
		CategoriasIngreso:          "ventas\nservicios",
		CategoriasEgreso:           "compras\nnomina",
		PrefijoIngreso:             "ING",
		PrefijoEgreso:              "EGR",
		FormatoImpresion:           "pos",
		RequiereAprobacion:         true,
		IntegracionContableDestino: "siigo",
		CuentaCajaBancos:           "110505",
		CuentaIngresos:             "413510",
		CuentaIVAGenerado:          "240801",
		CuentaGastos:               "519510",
		CuentaIVADescontable:       "240805",
		CuentasIngresoCategoria:    "ventas=413510",
		CuentasEgresoCategoria:     "compras=613510",
		UsuarioCreador:             "tester",
	})
	if err != nil {
		t.Fatalf("upsert config: %v", err)
	}

	cfg, err := GetEmpresaFinanzasConfiguracion(dbConn, 77)
	if err != nil {
		t.Fatalf("get config: %v", err)
	}
	if cfg.EmpresaID != 77 {
		t.Fatalf("expected empresa_id=77, got %d", cfg.EmpresaID)
	}
	if cfg.FormatoImpresion != "pos" {
		t.Fatalf("expected formato pos, got %s", cfg.FormatoImpresion)
	}
	if !cfg.RequiereAprobacion {
		t.Fatalf("expected requiere_aprobacion=true")
	}
	if cfg.IntegracionContableDestino != "siigo" {
		t.Fatalf("expected integracion siigo, got %s", cfg.IntegracionContableDestino)
	}
	if cfg.CuentaIngresos != "413510" {
		t.Fatalf("expected cuenta ingresos 413510, got %s", cfg.CuentaIngresos)
	}
	if cfg.CuentasIngresoCategoria != "ventas=413510" {
		t.Fatalf("expected cuentas ingreso map ventas=413510, got %s", cfg.CuentasIngresoCategoria)
	}
}

func TestEmpresaFinanzasMovimientosCRUDFlow(t *testing.T) {
	dbConn := openFinanzasTestDB(t)
	if err := EnsureEmpresaFinanzasSchema(dbConn); err != nil {
		t.Fatalf("ensure finanzas schema: %v", err)
	}

	id, err := CreateEmpresaFinanzasMovimiento(dbConn, EmpresaFinanzasMovimiento{
		EmpresaID:       1,
		TipoMovimiento:  "ingreso",
		Concepto:        "Ingreso test",
		Categoria:       "ventas",
		MetodoPago:      "efectivo",
		Moneda:          "COP",
		Monto:           100000,
		Impuesto:        0,
		Total:           100000,
		TipoComprobante: "recibo_interno",
		UsuarioCreador:  "tester",
	})
	if err != nil {
		t.Fatalf("create movimiento: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected id > 0")
	}

	rows, err := ListEmpresaFinanzasMovimientos(dbConn, 1, EmpresaFinanzasMovimientoFilter{Limit: 50})
	if err != nil {
		t.Fatalf("list movimientos: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 movimiento, got %d", len(rows))
	}

	mov := rows[0]
	mov.Concepto = "Ingreso test actualizado"
	mov.Total = 110000
	if err := UpdateEmpresaFinanzasMovimiento(dbConn, mov); err != nil {
		t.Fatalf("update movimiento: %v", err)
	}

	if err := SetEmpresaFinanzasMovimientoEstado(dbConn, 1, mov.ID, "inactivo"); err != nil {
		t.Fatalf("set estado: %v", err)
	}

	activos, err := ListEmpresaFinanzasMovimientos(dbConn, 1, EmpresaFinanzasMovimientoFilter{Limit: 50})
	if err != nil {
		t.Fatalf("list activos: %v", err)
	}
	if len(activos) != 0 {
		t.Fatalf("expected 0 activos after inactivar, got %d", len(activos))
	}

	incluyendoInactivos, err := ListEmpresaFinanzasMovimientos(dbConn, 1, EmpresaFinanzasMovimientoFilter{IncludeInactive: true, Limit: 50})
	if err != nil {
		t.Fatalf("list include inactive: %v", err)
	}
	if len(incluyendoInactivos) != 1 {
		t.Fatalf("expected 1 movimiento including inactive, got %d", len(incluyendoInactivos))
	}

	if err := DeleteEmpresaFinanzasMovimiento(dbConn, 1, mov.ID); err != nil {
		t.Fatalf("delete movimiento: %v", err)
	}

	finalRows, err := ListEmpresaFinanzasMovimientos(dbConn, 1, EmpresaFinanzasMovimientoFilter{IncludeInactive: true, Limit: 50})
	if err != nil {
		t.Fatalf("list final: %v", err)
	}
	if len(finalRows) != 0 {
		t.Fatalf("expected 0 movimientos after delete, got %d", len(finalRows))
	}
}

func TestEmpresaFinanzasMovimientoMontoInvalido(t *testing.T) {
	dbConn := openFinanzasTestDB(t)
	if err := EnsureEmpresaFinanzasSchema(dbConn); err != nil {
		t.Fatalf("ensure finanzas schema: %v", err)
	}

	_, err := CreateEmpresaFinanzasMovimiento(dbConn, EmpresaFinanzasMovimiento{
		EmpresaID:      1,
		TipoMovimiento: "egreso",
		Concepto:       "Pago invalido",
		Monto:          0,
	})
	if err == nil {
		t.Fatalf("expected error for monto <= 0")
	}
}

func TestGetEmpresaReportesTableroResumen(t *testing.T) {
	dbConn := openFinanzasTestDB(t)
	if err := EnsureEmpresaCarritosSchema(dbConn); err != nil {
		t.Fatalf("ensure carritos schema: %v", err)
	}
	if err := EnsureEmpresaClientesSchema(dbConn); err != nil {
		t.Fatalf("ensure clientes schema: %v", err)
	}
	if err := EnsureEmpresaProductosSchema(dbConn); err != nil {
		t.Fatalf("ensure productos schema: %v", err)
	}
	if err := EnsureEmpresaFinanzasSchema(dbConn); err != nil {
		t.Fatalf("ensure finanzas schema: %v", err)
	}
	if err := EnsureEmpresaEventosContablesSchema(dbConn); err != nil {
		t.Fatalf("ensure eventos contables schema: %v", err)
	}
	if err := EnsureEmpresaDocumentosTransaccionalesSchema(dbConn); err != nil {
		t.Fatalf("ensure documentos transaccionales schema: %v", err)
	}

	empresaID := int64(88)
	todayDate := time.Now().Format("2006-01-02")
	todayStamp := time.Now().Format("2006-01-02 15:04:05")

	if _, err := dbConn.Exec(`INSERT INTO carritos_compras (
		empresa_id, codigo, nombre, estado_carrito, total, total_pagado, pagado_en, estado
	) VALUES (?, 'C-001', 'Carrito KPI', 'cerrado', 150000, 150000, ?, 'activo')`, empresaID, todayStamp); err != nil {
		t.Fatalf("insert carrito: %v", err)
	}

	if _, err := dbConn.Exec(`INSERT INTO clientes (
		empresa_id, tipo_documento, numero_documento, nombre_razon_social, estado
	) VALUES (?, 'CC', '123', 'Cliente KPI', 'activo')`, empresaID); err != nil {
		t.Fatalf("insert cliente: %v", err)
	}

	if _, err := dbConn.Exec(`INSERT INTO productos (
		empresa_id, nombre, stock_minimo, estado
	) VALUES (?, 'Producto KPI', 5, 'activo')`, empresaID); err != nil {
		t.Fatalf("insert producto: %v", err)
	}

	var productoID int64
	if err := dbConn.QueryRow(`SELECT id FROM productos WHERE empresa_id = ? LIMIT 1`, empresaID).Scan(&productoID); err != nil {
		t.Fatalf("select producto id: %v", err)
	}

	if _, err := dbConn.Exec(`INSERT INTO inventario_existencias (
		empresa_id, producto_id, bodega_id, cantidad, estado
	) VALUES (?, ?, 1, 3, 'activo')`, empresaID, productoID); err != nil {
		t.Fatalf("insert inventario existencia: %v", err)
	}

	if _, err := dbConn.Exec(`INSERT INTO inventario_movimientos (
		empresa_id, producto_id, tipo, cantidad, costo_unitario, referencia, fecha_movimiento, estado
	) VALUES (?, ?, 'entrada', 4, 5000, 'COMP-001', ?, 'activo')`, empresaID, productoID, todayStamp); err != nil {
		t.Fatalf("insert inventario movimiento compra: %v", err)
	}

	if _, err := CreateEmpresaFinanzasMovimiento(dbConn, EmpresaFinanzasMovimiento{
		EmpresaID:       empresaID,
		TipoMovimiento:  "ingreso",
		Concepto:        "Ingreso KPI",
		Categoria:       "ventas",
		MetodoPago:      "efectivo",
		Moneda:          "COP",
		Monto:           100000,
		Total:           100000,
		FechaMovimiento: todayStamp,
		UsuarioCreador:  "tester",
	}); err != nil {
		t.Fatalf("create ingreso: %v", err)
	}

	if _, err := CreateEmpresaFinanzasMovimiento(dbConn, EmpresaFinanzasMovimiento{
		EmpresaID:       empresaID,
		TipoMovimiento:  "egreso",
		Concepto:        "Egreso KPI",
		Categoria:       "compras",
		MetodoPago:      "transferencia",
		Moneda:          "COP",
		Monto:           50000,
		Total:           50000,
		FechaMovimiento: todayStamp,
		UsuarioCreador:  "tester",
	}); err != nil {
		t.Fatalf("create egreso: %v", err)
	}

	if _, err := UpsertEmpresaFinanzasPeriodo(dbConn, EmpresaFinanzasPeriodo{
		EmpresaID:      empresaID,
		Periodo:        time.Now().Format("2006-01"),
		Estado:         "abierto",
		FechaInicio:    todayDate,
		FechaFin:       todayDate,
		UsuarioCreador: "tester",
	}); err != nil {
		t.Fatalf("upsert periodo abierto: %v", err)
	}

	if _, err := UpsertEmpresaFinanzasPeriodo(dbConn, EmpresaFinanzasPeriodo{
		EmpresaID:      empresaID,
		Periodo:        time.Now().AddDate(0, -1, 0).Format("2006-01"),
		Estado:         "cerrado",
		FechaInicio:    todayDate,
		FechaFin:       todayDate,
		UsuarioCreador: "tester",
		CerradoPor:     "tester",
	}); err != nil {
		t.Fatalf("upsert periodo cerrado: %v", err)
	}

	eventoPendienteID, err := CreateEmpresaEventoContable(dbConn, EmpresaEventoContable{
		EmpresaID:       empresaID,
		Modulo:          "ventas",
		Evento:          "venta_pagada",
		Entidad:         "carrito_compra",
		EntidadID:       1,
		DocumentoTipo:   "carrito",
		DocumentoCodigo: "C-001",
		PeriodoContable: time.Now().Format("2006-01"),
		MontoTotal:      150000,
		Moneda:          "COP",
		UsuarioCreador:  "tester",
	})
	if err != nil {
		t.Fatalf("create evento pendiente: %v", err)
	}

	eventoProcesadoID, err := CreateEmpresaEventoContable(dbConn, EmpresaEventoContable{
		EmpresaID:       empresaID,
		Modulo:          "finanzas",
		Evento:          "movimiento_ingreso_registrado",
		Entidad:         "finanzas_movimiento",
		EntidadID:       2,
		DocumentoTipo:   "comprobante",
		DocumentoCodigo: "ING-001",
		PeriodoContable: time.Now().Format("2006-01"),
		MontoTotal:      100000,
		Moneda:          "COP",
		UsuarioCreador:  "tester",
	})
	if err != nil {
		t.Fatalf("create evento procesado: %v", err)
	}

	if _, err := dbConn.Exec(`UPDATE empresa_eventos_contables SET procesado = 1, fecha_procesado = ? WHERE id = ?`, todayStamp, eventoProcesadoID); err != nil {
		t.Fatalf("mark evento procesado: %v", err)
	}
	if eventoPendienteID <= 0 {
		t.Fatalf("expected evento pendiente id > 0")
	}

	if _, err := UpsertEmpresaDocumentoFacturacion(dbConn, EmpresaDocumentoFacturacion{
		EmpresaID:       empresaID,
		TipoDocumento:   "factura_electronica",
		DocumentoCodigo: "FAC-001",
		EstadoDocumento: "emitida",
		EventoUltimo:    "factura_emitida",
		PeriodoContable: time.Now().Format("2006-01"),
		MontoTotal:      150000,
		Moneda:          "COP",
		UsuarioCreador:  "tester",
	}); err != nil {
		t.Fatalf("upsert documento facturacion: %v", err)
	}

	if _, err := UpsertEmpresaDocumentoCompra(dbConn, EmpresaDocumentoCompra{
		EmpresaID:       empresaID,
		TipoDocumento:   "orden_compra",
		DocumentoCodigo: "OC-001",
		EstadoDocumento: "emitida",
		EventoUltimo:    "orden_compra_emitida",
		PeriodoContable: time.Now().Format("2006-01"),
		MontoTotal:      20000,
		Moneda:          "COP",
		UsuarioCreador:  "tester",
	}); err != nil {
		t.Fatalf("upsert documento compra: %v", err)
	}

	resumen, err := GetEmpresaReportesTableroResumen(dbConn, empresaID, todayDate, todayDate)
	if err != nil {
		t.Fatalf("get tablero resumen: %v", err)
	}

	if resumen.Operativo.VentasCerradas != 1 {
		t.Fatalf("expected ventas_cerradas=1, got %d", resumen.Operativo.VentasCerradas)
	}
	if resumen.Operativo.IngresosVentas != 150000 {
		t.Fatalf("expected ingresos_ventas=150000, got %.2f", resumen.Operativo.IngresosVentas)
	}
	if resumen.Operativo.ClientesActivos != 1 {
		t.Fatalf("expected clientes_activos=1, got %d", resumen.Operativo.ClientesActivos)
	}
	if resumen.Operativo.ProductosActivos != 1 {
		t.Fatalf("expected productos_activos=1, got %d", resumen.Operativo.ProductosActivos)
	}
	if resumen.Operativo.ProductosBajoMinimo != 1 {
		t.Fatalf("expected productos_bajo_minimo=1, got %d", resumen.Operativo.ProductosBajoMinimo)
	}
	if resumen.Operativo.ComprasMovimientos != 1 {
		t.Fatalf("expected compras_movimientos=1, got %d", resumen.Operativo.ComprasMovimientos)
	}
	if resumen.Operativo.ComprasCosto != 20000 {
		t.Fatalf("expected compras_costo=20000, got %.2f", resumen.Operativo.ComprasCosto)
	}

	if resumen.Financiero.MovimientosIngresos != 1 {
		t.Fatalf("expected movimientos_ingresos=1, got %d", resumen.Financiero.MovimientosIngresos)
	}
	if resumen.Financiero.MovimientosEgresos != 1 {
		t.Fatalf("expected movimientos_egresos=1, got %d", resumen.Financiero.MovimientosEgresos)
	}
	if resumen.Financiero.Ingresos != 100000 {
		t.Fatalf("expected ingresos=100000, got %.2f", resumen.Financiero.Ingresos)
	}
	if resumen.Financiero.Egresos != 50000 {
		t.Fatalf("expected egresos=50000, got %.2f", resumen.Financiero.Egresos)
	}
	if resumen.Financiero.Balance != 50000 {
		t.Fatalf("expected balance=50000, got %.2f", resumen.Financiero.Balance)
	}
	if resumen.Financiero.PeriodosAbiertos != 1 {
		t.Fatalf("expected periodos_abiertos=1, got %d", resumen.Financiero.PeriodosAbiertos)
	}
	if resumen.Financiero.PeriodosCerrados != 1 {
		t.Fatalf("expected periodos_cerrados=1, got %d", resumen.Financiero.PeriodosCerrados)
	}

	if resumen.Contable.EventosPendientes != 1 {
		t.Fatalf("expected eventos_pendientes=1, got %d", resumen.Contable.EventosPendientes)
	}
	if resumen.Contable.EventosProcesados != 1 {
		t.Fatalf("expected eventos_procesados=1, got %d", resumen.Contable.EventosProcesados)
	}
	if resumen.Contable.EventosTotal != 2 {
		t.Fatalf("expected eventos_total=2, got %d", resumen.Contable.EventosTotal)
	}
	if resumen.Contable.DocumentosFacturacionActivos != 1 {
		t.Fatalf("expected documentos_facturacion_activos=1, got %d", resumen.Contable.DocumentosFacturacionActivos)
	}
	if resumen.Contable.DocumentosComprasActivos != 1 {
		t.Fatalf("expected documentos_compras_activos=1, got %d", resumen.Contable.DocumentosComprasActivos)
	}
}

func TestEmpresaCierresCajaFlow(t *testing.T) {
	dbConn := openFinanzasTestDB(t)
	if err := EnsureEmpresaFinanzasSchema(dbConn); err != nil {
		t.Fatalf("ensure finanzas schema: %v", err)
	}

	empresaID := int64(64)
	id, err := CreateEmpresaCierreCaja(dbConn, EmpresaCierreCaja{
		EmpresaID:        empresaID,
		SucursalID:       2,
		CajaCodigo:       "cj_02",
		Turno:            "noche",
		AperturaMonto:    100000,
		IngresosEfectivo: 50000,
		EgresosEfectivo:  10000,
		RetirosEfectivo:  10000,
		Moneda:           "cop",
		UsuarioCreador:   "tester",
	})
	if err != nil {
		t.Fatalf("create cierre caja: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected cierre caja id > 0")
	}

	rows, err := ListEmpresaCierresCaja(dbConn, empresaID, EmpresaCierreCajaFilter{Limit: 10})
	if err != nil {
		t.Fatalf("list cierres caja: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 cierre caja, got %d", len(rows))
	}
	if rows[0].CajaCodigo != "CJ_02" {
		t.Fatalf("expected caja codigo normalized CJ_02, got %s", rows[0].CajaCodigo)
	}
	if rows[0].CajaTeorica != 130000 {
		t.Fatalf("expected caja teorica 130000, got %.2f", rows[0].CajaTeorica)
	}
	if rows[0].EstadoCierre != "abierto" {
		t.Fatalf("expected estado_cierre abierto, got %s", rows[0].EstadoCierre)
	}

	cajaFisica := 125000.0
	if err := SetEmpresaCierreCajaEstado(dbConn, empresaID, id, "cerrado", &cajaFisica, "tester", "arqueo turno noche"); err != nil {
		t.Fatalf("cerrar caja: %v", err)
	}

	cerrados, err := ListEmpresaCierresCaja(dbConn, empresaID, EmpresaCierreCajaFilter{EstadoCierre: "cerrado", Limit: 10})
	if err != nil {
		t.Fatalf("list cierres cerrados: %v", err)
	}
	if len(cerrados) != 1 {
		t.Fatalf("expected 1 cierre cerrado, got %d", len(cerrados))
	}
	if cerrados[0].DiferenciaCaja != 5000 {
		t.Fatalf("expected diferencia caja 5000, got %.2f", cerrados[0].DiferenciaCaja)
	}
	if !cerrados[0].TieneIncidencia {
		t.Fatalf("expected incidencia=true when diferencia != 0")
	}

	if err := SetEmpresaCierreCajaEstado(dbConn, empresaID, id, "aprobado", nil, "supervisor", "cierre aprobado"); err != nil {
		t.Fatalf("aprobar cierre caja: %v", err)
	}

	err = UpdateEmpresaCierreCaja(dbConn, EmpresaCierreCaja{
		ID:               id,
		EmpresaID:        empresaID,
		SucursalID:       2,
		CajaCodigo:       "CJ-02",
		Turno:            "noche",
		FechaOperacion:   time.Now().Format("2006-01-02"),
		EstadoCierre:     "aprobado",
		AperturaMonto:    100000,
		IngresosEfectivo: 50000,
		EgresosEfectivo:  10000,
		RetirosEfectivo:  10000,
		CajaFisica:       125000,
		Moneda:           "COP",
		UsuarioCreador:   "tester",
	})
	if !errors.Is(err, ErrCierreCajaAprobadoBloqueado) {
		t.Fatalf("expected ErrCierreCajaAprobadoBloqueado, got %v", err)
	}

	if err := SetEmpresaCierreCajaRegistroEstado(dbConn, empresaID, id, "inactivo"); err != nil {
		t.Fatalf("set cierre caja estado registro inactivo: %v", err)
	}

	activos, err := ListEmpresaCierresCaja(dbConn, empresaID, EmpresaCierreCajaFilter{Limit: 10})
	if err != nil {
		t.Fatalf("list cierres activos: %v", err)
	}
	if len(activos) != 0 {
		t.Fatalf("expected 0 cierres activos after inactivar, got %d", len(activos))
	}
}
