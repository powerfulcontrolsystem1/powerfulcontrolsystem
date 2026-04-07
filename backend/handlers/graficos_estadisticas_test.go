package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	dbpkg "github.com/you/pos-backend/db"
)

// testingDBExec adapta *sql.DB para helpers locales de seed.
type testingDBExec struct {
	execFn func(string, ...interface{}) error
}

func (d *testingDBExec) Exec(query string, args ...interface{}) error {
	if d == nil || d.execFn == nil {
		return nil
	}
	return d.execFn(query, args...)
}

func TestEmpresaGraficosEstadisticasHandlerPanelYAcciones(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_graficos_estadisticas_handler.db")
	ensureEmpresaReportesSchema(t, dbEmp)
	if err := dbpkg.EnsureEmpresaAsistenciaSchema(dbEmp); err != nil {
		t.Fatalf("EnsureEmpresaAsistenciaSchema: %v", err)
	}

	exec := &testingDBExec{execFn: func(query string, args ...interface{}) error {
		_, err := dbEmp.Exec(query, args...)
		return err
	}}
	seedGraficosData(t, exec)

	handler := EmpresaGraficosEstadisticasHandler(dbEmp)

	reqPanel := httptest.NewRequest(http.MethodGet, "/api/empresa/graficos_estadisticas?action=panel&empresa_id=7&desde=2026-04-01&hasta=2026-04-03", nil)
	rrPanel := httptest.NewRecorder()
	handler.ServeHTTP(rrPanel, reqPanel)
	if rrPanel.Code != http.StatusOK {
		t.Fatalf("panel status=%d body=%s", rrPanel.Code, rrPanel.Body.String())
	}

	var panel empresaGraficosPanelResponse
	if err := json.Unmarshal(rrPanel.Body.Bytes(), &panel); err != nil {
		t.Fatalf("unmarshal panel: %v", err)
	}
	if panel.EmpresaID != 7 {
		t.Fatalf("empresa_id esperado=7 obtenido=%d", panel.EmpresaID)
	}
	if len(panel.Series.Ventas) == 0 {
		t.Fatalf("series ventas vacia")
	}
	if len(panel.Series.Finanzas) == 0 {
		t.Fatalf("series finanzas vacia")
	}
	if len(panel.Series.Compras) == 0 {
		t.Fatalf("series compras vacia")
	}
	if len(panel.Series.Asistencia) == 0 {
		t.Fatalf("series asistencia vacia")
	}
	if len(panel.Rankings.TopProductos) == 0 {
		t.Fatalf("ranking top productos vacio")
	}
	if len(panel.Rankings.TopClientes) == 0 {
		t.Fatalf("ranking top clientes vacio")
	}
	if len(panel.Distribuciones.StockEstado) == 0 {
		t.Fatalf("distribucion stock vacia")
	}
	if len(panel.Distribuciones.AsistenciaEstado) == 0 {
		t.Fatalf("distribucion asistencia vacia")
	}

	reqSerie := httptest.NewRequest(http.MethodGet, "/api/empresa/graficos_estadisticas?action=serie&serie=ventas&empresa_id=7", nil)
	rrSerie := httptest.NewRecorder()
	handler.ServeHTTP(rrSerie, reqSerie)
	if rrSerie.Code != http.StatusOK {
		t.Fatalf("serie status=%d body=%s", rrSerie.Code, rrSerie.Body.String())
	}

	var serieResp struct {
		Serie string                      `json:"serie"`
		Data  []empresaGraficoSerieVentas `json:"data"`
	}
	if err := json.Unmarshal(rrSerie.Body.Bytes(), &serieResp); err != nil {
		t.Fatalf("unmarshal serie: %v", err)
	}
	if serieResp.Serie != "ventas" {
		t.Fatalf("serie esperada=ventas obtenida=%s", serieResp.Serie)
	}
	if len(serieResp.Data) == 0 {
		t.Fatalf("data de serie ventas vacia")
	}

	reqCatalog := httptest.NewRequest(http.MethodGet, "/api/empresa/graficos_estadisticas?action=catalogo&empresa_id=7", nil)
	rrCatalog := httptest.NewRecorder()
	handler.ServeHTTP(rrCatalog, reqCatalog)
	if rrCatalog.Code != http.StatusOK {
		t.Fatalf("catalogo status=%d body=%s", rrCatalog.Code, rrCatalog.Body.String())
	}
}

func TestEmpresaGraficosEstadisticasHandlerErrores(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_graficos_estadisticas_handler_errores.db")
	ensureEmpresaReportesSchema(t, dbEmp)
	if err := dbpkg.EnsureEmpresaAsistenciaSchema(dbEmp); err != nil {
		t.Fatalf("EnsureEmpresaAsistenciaSchema: %v", err)
	}

	handler := EmpresaGraficosEstadisticasHandler(dbEmp)

	reqMissingEmpresa := httptest.NewRequest(http.MethodGet, "/api/empresa/graficos_estadisticas?action=panel", nil)
	rrMissingEmpresa := httptest.NewRecorder()
	handler.ServeHTTP(rrMissingEmpresa, reqMissingEmpresa)
	if rrMissingEmpresa.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 missing empresa_id, got %d", rrMissingEmpresa.Code)
	}

	reqBadAction := httptest.NewRequest(http.MethodGet, "/api/empresa/graficos_estadisticas?action=foo&empresa_id=7", nil)
	rrBadAction := httptest.NewRecorder()
	handler.ServeHTTP(rrBadAction, reqBadAction)
	if rrBadAction.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 bad action, got %d", rrBadAction.Code)
	}

	reqBadMaxPoints := httptest.NewRequest(http.MethodGet, "/api/empresa/graficos_estadisticas?action=panel&empresa_id=7&max_points=abc", nil)
	rrBadMaxPoints := httptest.NewRecorder()
	handler.ServeHTTP(rrBadMaxPoints, reqBadMaxPoints)
	if rrBadMaxPoints.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 bad max_points, got %d", rrBadMaxPoints.Code)
	}
}

func TestEmpresaGraficosEstadisticasHandlerFiltrosComparativoYCache(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_graficos_estadisticas_handler_filtros.db")
	ensureEmpresaReportesSchema(t, dbEmp)
	if err := dbpkg.EnsureEmpresaAsistenciaSchema(dbEmp); err != nil {
		t.Fatalf("EnsureEmpresaAsistenciaSchema: %v", err)
	}

	exec := &testingDBExec{execFn: func(query string, args ...interface{}) error {
		_, err := dbEmp.Exec(query, args...)
		return err
	}}
	seedGraficosData(t, exec)

	mustExec := func(query string, args ...interface{}) {
		t.Helper()
		if err := exec.Exec(query, args...); err != nil {
			t.Fatalf("seed extra query failed: %v | query=%s", err, query)
		}
	}

	mustExec(`INSERT INTO carritos_compras (
		empresa_id, codigo, nombre, cliente_id, estado_carrito, total, total_pagado, pagado_en,
		estado, fecha_creacion, fecha_actualizacion
	) VALUES (
		7, 'CRT-7002', 'Venta demo 7002', 1, 'cerrado', 220000, 220000, '2026-04-02 14:20:00',
		'activo', datetime('now','localtime'), datetime('now','localtime')
	);`)

	mustExec(`INSERT INTO carrito_compra_items (
		empresa_id, carrito_id, tipo_item, referencia_id, codigo_item, descripcion,
		cantidad, precio_unitario, subtotal_linea, total_linea,
		estado, fecha_creacion, fecha_actualizacion
	) VALUES (
		7, 2, 'producto', 11, 'SKU-MOUSE', 'Mouse inalambrico',
		2, 110000, 220000, 220000,
		'activo', datetime('now','localtime'), datetime('now','localtime')
	);`)

	mustExec(`INSERT INTO carritos_compras (
		empresa_id, codigo, nombre, cliente_id, estado_carrito, total, total_pagado, pagado_en,
		estado, fecha_creacion, fecha_actualizacion
	) VALUES (
		7, 'CRT-7003', 'Venta demo 7003', 1, 'cerrado', 150000, 150000, '2026-03-30 10:05:00',
		'activo', datetime('now','localtime'), datetime('now','localtime')
	);`)

	mustExec(`INSERT INTO carrito_compra_items (
		empresa_id, carrito_id, tipo_item, referencia_id, codigo_item, descripcion,
		cantidad, precio_unitario, subtotal_linea, total_linea,
		estado, fecha_creacion, fecha_actualizacion
	) VALUES (
		7, 3, 'producto', 12, 'SKU-AUDIFONOS', 'Audifonos bluetooth',
		1, 150000, 150000, 150000,
		'activo', datetime('now','localtime'), datetime('now','localtime')
	);`)

	mustExec(`INSERT INTO empresa_ventas_estacion_metricas (
		empresa_id, carrito_id, estacion_id, estacion_codigo, estacion_nombre, evento_operacion,
		fecha_evento, estado, fecha_creacion, fecha_actualizacion
	) VALUES
		(7, 1, 1, 'EST-1', 'Mesa 1', 'venta_pagada', '2026-04-01 11:20:00', 'activo', datetime('now','localtime'), datetime('now','localtime')),
		(7, 2, 2, 'EST-2', 'Mesa 2', 'venta_pagada', '2026-04-02 14:20:00', 'activo', datetime('now','localtime'), datetime('now','localtime')),
		(7, 3, 2, 'EST-2', 'Mesa 2', 'venta_pagada', '2026-03-30 10:05:00', 'activo', datetime('now','localtime'), datetime('now','localtime'));
	`)

	mustExec(`INSERT INTO empresa_cierres_caja (
		empresa_id, sucursal_id, caja_codigo, turno, fecha_operacion, estado_cierre,
		estado, fecha_creacion, fecha_actualizacion
	) VALUES (
		7, 9, 'CAJA-09', 'manana', '2026-04-02', 'cerrado',
		'activo', datetime('now','localtime'), datetime('now','localtime')
	);`)

	handler := EmpresaGraficosEstadisticasHandler(dbEmp)

	url := "/api/empresa/graficos_estadisticas?action=panel&empresa_id=7&desde=2026-04-01&hasta=2026-04-03&estacion_id=2&segmento=frecuente&comparar=true"
	req := httptest.NewRequest(http.MethodGet, url, nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("panel filtros status=%d body=%s", rr.Code, rr.Body.String())
	}

	var panel empresaGraficosPanelResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &panel); err != nil {
		t.Fatalf("unmarshal panel filtros: %v", err)
	}
	if panel.Cache.Hit {
		t.Fatalf("primer request no debe ser cache hit")
	}
	if panel.Filtros.EstacionID != 2 {
		t.Fatalf("estacion_id esperado=2 obtenido=%d", panel.Filtros.EstacionID)
	}
	if panel.Filtros.Segmento != "frecuente" {
		t.Fatalf("segmento esperado=frecuente obtenido=%s", panel.Filtros.Segmento)
	}
	if panel.Comparativo == nil {
		t.Fatalf("comparativo no debe ser nil")
	}
	if len(panel.Series.Ventas) != 1 {
		t.Fatalf("ventas filtradas esperadas=1 obtenido=%d", len(panel.Series.Ventas))
	}
	if panel.Tablero.Operativo.VentasCerradas != 1 {
		t.Fatalf("ventas cerradas esperadas=1 obtenido=%d", panel.Tablero.Operativo.VentasCerradas)
	}
	metricaVentas, ok := panel.Comparativo.Metricas["ventas_cerradas"]
	if !ok {
		t.Fatalf("comparativo sin metrica ventas_cerradas")
	}
	if metricaVentas.Anterior != 1 {
		t.Fatalf("comparativo ventas anterior esperado=1 obtenido=%.2f", metricaVentas.Anterior)
	}

	reqCached := httptest.NewRequest(http.MethodGet, url, nil)
	rrCached := httptest.NewRecorder()
	handler.ServeHTTP(rrCached, reqCached)
	if rrCached.Code != http.StatusOK {
		t.Fatalf("panel cache status=%d body=%s", rrCached.Code, rrCached.Body.String())
	}
	var panelCached empresaGraficosPanelResponse
	if err := json.Unmarshal(rrCached.Body.Bytes(), &panelCached); err != nil {
		t.Fatalf("unmarshal panel cache: %v", err)
	}
	if !panelCached.Cache.Hit {
		t.Fatalf("segundo request debe ser cache hit")
	}

	reqSucursal := httptest.NewRequest(http.MethodGet, "/api/empresa/graficos_estadisticas?action=panel&empresa_id=7&desde=2026-04-01&hasta=2026-04-03&sucursal_id=9", nil)
	rrSucursal := httptest.NewRecorder()
	handler.ServeHTTP(rrSucursal, reqSucursal)
	if rrSucursal.Code != http.StatusOK {
		t.Fatalf("panel sucursal status=%d body=%s", rrSucursal.Code, rrSucursal.Body.String())
	}
	var panelSucursal empresaGraficosPanelResponse
	if err := json.Unmarshal(rrSucursal.Body.Bytes(), &panelSucursal); err != nil {
		t.Fatalf("unmarshal panel sucursal: %v", err)
	}
	if panelSucursal.Filtros.SucursalID != 9 {
		t.Fatalf("sucursal_id esperado=9 obtenido=%d", panelSucursal.Filtros.SucursalID)
	}
	if panelSucursal.Tablero.Operativo.VentasCerradas != 1 {
		t.Fatalf("ventas cerradas por sucursal esperadas=1 obtenido=%d", panelSucursal.Tablero.Operativo.VentasCerradas)
	}
}

func seedGraficosData(t *testing.T, dbExec *testingDBExec) {
	t.Helper()

	mustExec := func(query string, args ...interface{}) {
		t.Helper()
		if err := dbExec.Exec(query, args...); err != nil {
			t.Fatalf("seed query failed: %v | query=%s", err, query)
		}
	}

	mustExec(`INSERT INTO clientes (
		empresa_id, tipo_documento, numero_documento, nombre_razon_social,
		estado, fecha_creacion, fecha_actualizacion
	) VALUES (
		7, 'CC', '900001007', 'Cliente Demo',
		'activo', datetime('now','localtime'), datetime('now','localtime')
	);`)

	mustExec(`INSERT INTO carritos_compras (
		empresa_id, codigo, nombre, cliente_id, estado_carrito, total, total_pagado, pagado_en,
		estado, fecha_creacion, fecha_actualizacion
	) VALUES (
		7, 'CRT-7001', 'Venta demo 7001', 1, 'cerrado', 180000, 180000, '2026-04-01 11:20:00',
		'activo', datetime('now','localtime'), datetime('now','localtime')
	);`)

	mustExec(`INSERT INTO carrito_compra_items (
		empresa_id, carrito_id, tipo_item, referencia_id, codigo_item, descripcion,
		cantidad, precio_unitario, subtotal_linea, total_linea,
		estado, fecha_creacion, fecha_actualizacion
	) VALUES (
		7, 1, 'producto', 10, 'SKU-TECLADO', 'Teclado mecanico',
		2, 90000, 180000, 180000,
		'activo', datetime('now','localtime'), datetime('now','localtime')
	);`)

	mustExec(`INSERT INTO bodegas (
		empresa_id, codigo, nombre, estado, fecha_creacion, fecha_actualizacion
	) VALUES (
		7, 'BOD-01', 'Bodega Principal', 'activo', datetime('now','localtime'), datetime('now','localtime')
	);`)

	mustExec(`INSERT INTO productos (
		empresa_id, bodega_principal_id, sku, nombre, stock_minimo, stock_maximo,
		estado, fecha_creacion, fecha_actualizacion
	) VALUES (
		7, 1, 'SKU-TECLADO', 'Teclado mecanico', 5, 25,
		'activo', datetime('now','localtime'), datetime('now','localtime')
	);`)

	mustExec(`INSERT INTO inventario_existencias (
		empresa_id, producto_id, bodega_id, cantidad, estado, fecha_creacion, fecha_actualizacion
	) VALUES (
		7, 1, 1, 4, 'activo', datetime('now','localtime'), datetime('now','localtime')
	);`)

	mustExec(`INSERT INTO inventario_movimientos (
		empresa_id, producto_id, bodega_destino_id, tipo, cantidad, costo_unitario,
		referencia, fecha_movimiento, estado, fecha_creacion, fecha_actualizacion
	) VALUES (
		7, 1, 1, 'compra', 10, 70000,
		'OC-7001', '2026-04-01 09:10:00', 'activo', datetime('now','localtime'), datetime('now','localtime')
	);`)

	mustExec(`INSERT INTO empresa_finanzas_movimientos (
		empresa_id, tipo_movimiento, codigo, fecha_movimiento, categoria, concepto,
		monto, total, total_neto, estado, fecha_creacion, fecha_actualizacion
	) VALUES (
		7, 'ingreso', 'ING-7001', '2026-04-01 12:00:00', 'ventas', 'venta mostrador',
		180000, 180000, 180000, 'activo', datetime('now','localtime'), datetime('now','localtime')
	);`)

	mustExec(`INSERT INTO empresa_finanzas_movimientos (
		empresa_id, tipo_movimiento, codigo, fecha_movimiento, categoria, concepto,
		monto, total, total_neto, estado, fecha_creacion, fecha_actualizacion
	) VALUES (
		7, 'egreso', 'EGR-7001', '2026-04-01 18:00:00', 'compras', 'compra mercancia',
		70000, 70000, 70000, 'activo', datetime('now','localtime'), datetime('now','localtime')
	);`)

	mustExec(`INSERT INTO empresa_asistencia_empleados (
		empresa_id, empleado_id, empleado_codigo, empleado_nombre, fecha_asistencia,
		hora_entrada, hora_salida, minutos_tarde, horas_trabajadas, estado_asistencia,
		estado, fecha_creacion, fecha_actualizacion
	) VALUES (
		7, 101, 'EMP-101', 'Laura Perez', '2026-04-01',
		'08:10:00', '17:05:00', 10, 8.92, 'presente',
		'activo', datetime('now','localtime'), datetime('now','localtime')
	);`)
}
