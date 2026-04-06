package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	dbpkg "github.com/you/pos-backend/db"
)

func TestEmpresaCategoriasProductosHandlerCRUD(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_categorias_handler.db")
	if err := dbpkg.EnsureEmpresaProductosSchema(dbEmp); err != nil {
		t.Fatalf("ensure productos schema: %v", err)
	}

	h := EmpresaCategoriasProductosHandler(dbEmp)

	createReq := httptest.NewRequest(http.MethodPost, "/api/empresa/categorias_productos", strings.NewReader(`{"empresa_id":7,"nombre":"Bebidas","codigo":"CAT-BEB"}`))
	createReq.Header.Set("Content-Type", "application/json")
	createRR := httptest.NewRecorder()
	h.ServeHTTP(createRR, createReq)
	if createRR.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, createRR.Code, createRR.Body.String())
	}

	var createResp map[string]interface{}
	if err := json.Unmarshal(createRR.Body.Bytes(), &createResp); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	categoriaID := int64(createResp["id"].(float64))
	if categoriaID <= 0 {
		t.Fatalf("invalid categoria id: %v", createResp)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/empresa/categorias_productos?empresa_id=7&include_inactive=1", nil)
	listRR := httptest.NewRecorder()
	h.ServeHTTP(listRR, listReq)
	if listRR.Code != http.StatusOK {
		t.Fatalf("expected list status %d, got %d body=%s", http.StatusOK, listRR.Code, listRR.Body.String())
	}
	var categorias []dbpkg.CategoriaProducto
	if err := json.Unmarshal(listRR.Body.Bytes(), &categorias); err != nil {
		t.Fatalf("decode categorias: %v", err)
	}
	if len(categorias) != 1 {
		t.Fatalf("expected 1 categoria, got %d", len(categorias))
	}

	updateBody := `{"id":` + strconv.FormatInt(categoriaID, 10) + `,"empresa_id":7,"nombre":"Bebidas frias","codigo":"CAT-BEB"}`
	updateReq := httptest.NewRequest(http.MethodPut, "/api/empresa/categorias_productos", strings.NewReader(updateBody))
	updateReq.Header.Set("Content-Type", "application/json")
	updateRR := httptest.NewRecorder()
	h.ServeHTTP(updateRR, updateReq)
	if updateRR.Code != http.StatusNoContent {
		t.Fatalf("expected update status %d, got %d body=%s", http.StatusNoContent, updateRR.Code, updateRR.Body.String())
	}

	toggleReq := httptest.NewRequest(http.MethodPut, "/api/empresa/categorias_productos?empresa_id=7&id="+strconv.FormatInt(categoriaID, 10)+"&action=activar&activo=0", nil)
	toggleRR := httptest.NewRecorder()
	h.ServeHTTP(toggleRR, toggleReq)
	if toggleRR.Code != http.StatusNoContent {
		t.Fatalf("expected toggle status %d, got %d body=%s", http.StatusNoContent, toggleRR.Code, toggleRR.Body.String())
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, "/api/empresa/categorias_productos?empresa_id=7&id="+strconv.FormatInt(categoriaID, 10), nil)
	deleteRR := httptest.NewRecorder()
	h.ServeHTTP(deleteRR, deleteReq)
	if deleteRR.Code != http.StatusNoContent {
		t.Fatalf("expected delete status %d, got %d body=%s", http.StatusNoContent, deleteRR.Code, deleteRR.Body.String())
	}
}

func TestEmpresaProductosHandlerFiltraPorCategoriaID(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_productos_categoria_filter.db")
	if err := dbpkg.EnsureEmpresaProductosSchema(dbEmp); err != nil {
		t.Fatalf("ensure productos schema: %v", err)
	}

	catA, err := dbpkg.CreateCategoriaProducto(dbEmp, dbpkg.CategoriaProducto{EmpresaID: 11, Nombre: "Tecnologia"})
	if err != nil {
		t.Fatalf("create categoria A: %v", err)
	}
	catB, err := dbpkg.CreateCategoriaProducto(dbEmp, dbpkg.CategoriaProducto{EmpresaID: 11, Nombre: "Hogar"})
	if err != nil {
		t.Fatalf("create categoria B: %v", err)
	}

	if _, err := dbpkg.CreateProducto(dbEmp, dbpkg.Producto{EmpresaID: 11, Nombre: "Laptop", CategoriaID: catA, Precio: 1000, Costo: 800}, 0, "TEST"); err != nil {
		t.Fatalf("create producto categoria A: %v", err)
	}
	if _, err := dbpkg.CreateProducto(dbEmp, dbpkg.Producto{EmpresaID: 11, Nombre: "Silla", CategoriaID: catB, Precio: 200, Costo: 120}, 0, "TEST"); err != nil {
		t.Fatalf("create producto categoria B: %v", err)
	}

	h := EmpresaProductosHandler(dbEmp)
	req := httptest.NewRequest(http.MethodGet, "/api/empresa/productos?empresa_id=11&categoria_id="+strconv.FormatInt(catA, 10), nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var productos []dbpkg.Producto
	if err := json.Unmarshal(rr.Body.Bytes(), &productos); err != nil {
		t.Fatalf("decode productos: %v", err)
	}
	if len(productos) != 1 {
		t.Fatalf("expected 1 producto filtered, got %d", len(productos))
	}
	if productos[0].CategoriaID != catA {
		t.Fatalf("expected categoria_id=%d, got %d", catA, productos[0].CategoriaID)
	}
}

func TestEmpresaInventarioAlertasHandlerDevuelveQuiebrePorBodega(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_inventario_alertas_handler.db")
	if err := dbpkg.EnsureEmpresaProductosSchema(dbEmp); err != nil {
		t.Fatalf("ensure productos schema: %v", err)
	}

	bodegaPrincipalID, err := dbpkg.CreateBodega(dbEmp, dbpkg.Bodega{EmpresaID: 21, Codigo: "BOD-PRI-21", Nombre: "Bodega Principal"})
	if err != nil {
		t.Fatalf("create bodega principal: %v", err)
	}
	bodegaSecundariaID, err := dbpkg.CreateBodega(dbEmp, dbpkg.Bodega{EmpresaID: 21, Codigo: "BOD-SEC-21", Nombre: "Bodega Secundaria"})
	if err != nil {
		t.Fatalf("create bodega secundaria: %v", err)
	}

	productoBajoID, err := dbpkg.CreateProducto(dbEmp, dbpkg.Producto{EmpresaID: 21, Nombre: "Arroz", StockMinimo: 10, StockMaximo: 40}, 0, "TEST")
	if err != nil {
		t.Fatalf("create producto bajo minimo: %v", err)
	}
	productoSinStockID, err := dbpkg.CreateProducto(dbEmp, dbpkg.Producto{EmpresaID: 21, Nombre: "Azucar", StockMinimo: 0, StockMaximo: 30}, 0, "TEST")
	if err != nil {
		t.Fatalf("create producto sin stock: %v", err)
	}
	productoOKID, err := dbpkg.CreateProducto(dbEmp, dbpkg.Producto{EmpresaID: 21, Nombre: "Cafe", StockMinimo: 5, StockMaximo: 30}, 0, "TEST")
	if err != nil {
		t.Fatalf("create producto ok: %v", err)
	}

	if _, err := dbEmp.Exec(`INSERT INTO inventario_existencias (empresa_id, producto_id, bodega_id, cantidad, estado) VALUES (?, ?, ?, 3, 'activo')`, 21, productoBajoID, bodegaPrincipalID); err != nil {
		t.Fatalf("insert existencia bajo minimo: %v", err)
	}
	if _, err := dbEmp.Exec(`INSERT INTO inventario_existencias (empresa_id, producto_id, bodega_id, cantidad, estado) VALUES (?, ?, ?, 0, 'activo')`, 21, productoSinStockID, bodegaSecundariaID); err != nil {
		t.Fatalf("insert existencia sin stock: %v", err)
	}
	if _, err := dbEmp.Exec(`INSERT INTO inventario_existencias (empresa_id, producto_id, bodega_id, cantidad, estado) VALUES (?, ?, ?, 18, 'activo')`, 21, productoOKID, bodegaPrincipalID); err != nil {
		t.Fatalf("insert existencia ok: %v", err)
	}

	h := EmpresaInventarioAlertasHandler(dbEmp)
	req := httptest.NewRequest(http.MethodGet, "/api/empresa/inventario/alertas?empresa_id=21&limit=50", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var alertas []dbpkg.InventarioAlertaQuiebre
	if err := json.Unmarshal(rr.Body.Bytes(), &alertas); err != nil {
		t.Fatalf("decode alertas: %v", err)
	}
	if len(alertas) != 2 {
		t.Fatalf("expected 2 alertas, got %d", len(alertas))
	}

	hasSinStock := false
	hasBajoMinimo := false
	for _, alerta := range alertas {
		if alerta.EstadoStock == "sin_stock" {
			hasSinStock = true
		}
		if alerta.EstadoStock == "bajo_minimo" {
			hasBajoMinimo = true
		}
	}
	if !hasSinStock || !hasBajoMinimo {
		t.Fatalf("expected alertas con estado sin_stock y bajo_minimo, got %+v", alertas)
	}

	filteredReq := httptest.NewRequest(http.MethodGet, "/api/empresa/inventario/alertas?empresa_id=21&bodega_id="+strconv.FormatInt(bodegaPrincipalID, 10), nil)
	filteredRR := httptest.NewRecorder()
	h.ServeHTTP(filteredRR, filteredReq)
	if filteredRR.Code != http.StatusOK {
		t.Fatalf("expected filtered status %d, got %d body=%s", http.StatusOK, filteredRR.Code, filteredRR.Body.String())
	}

	var filtered []dbpkg.InventarioAlertaQuiebre
	if err := json.Unmarshal(filteredRR.Body.Bytes(), &filtered); err != nil {
		t.Fatalf("decode filtered alertas: %v", err)
	}
	if len(filtered) != 1 {
		t.Fatalf("expected 1 alerta filtrada por bodega principal, got %d", len(filtered))
	}
	if filtered[0].ProductoID != productoBajoID {
		t.Fatalf("expected alerta de producto bajo minimo (%d), got %d", productoBajoID, filtered[0].ProductoID)
	}
}

func TestEmpresaInventarioMovimientosHandlerFiltraPorBodegaTipoYRango(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_inventario_movimientos_handler.db")
	if err := dbpkg.EnsureEmpresaProductosSchema(dbEmp); err != nil {
		t.Fatalf("ensure productos schema: %v", err)
	}

	bodegaOrigenID, err := dbpkg.CreateBodega(dbEmp, dbpkg.Bodega{EmpresaID: 31, Codigo: "BOD-ORI-31", Nombre: "Origen"})
	if err != nil {
		t.Fatalf("create bodega origen: %v", err)
	}
	bodegaDestinoID, err := dbpkg.CreateBodega(dbEmp, dbpkg.Bodega{EmpresaID: 31, Codigo: "BOD-DES-31", Nombre: "Destino"})
	if err != nil {
		t.Fatalf("create bodega destino: %v", err)
	}

	productoID, err := dbpkg.CreateProducto(dbEmp, dbpkg.Producto{EmpresaID: 31, Nombre: "Producto Kardex", StockMinimo: 1, StockMaximo: 20}, 0, "TEST")
	if err != nil {
		t.Fatalf("create producto: %v", err)
	}

	if _, err := dbEmp.Exec(`INSERT INTO inventario_movimientos (
		empresa_id, producto_id, bodega_origen_id, bodega_destino_id, tipo, cantidad, costo_unitario, referencia, fecha_movimiento, estado
	) VALUES (?, ?, NULL, ?, 'entrada', 8, 1000, 'REF-ENTRADA', '2026-04-01 08:00:00', 'activo')`, 31, productoID, bodegaOrigenID); err != nil {
		t.Fatalf("insert movimiento entrada: %v", err)
	}
	if _, err := dbEmp.Exec(`INSERT INTO inventario_movimientos (
		empresa_id, producto_id, bodega_origen_id, bodega_destino_id, tipo, cantidad, costo_unitario, referencia, fecha_movimiento, estado
	) VALUES (?, ?, ?, ?, 'traslado', 3, 1000, 'REF-TRASLADO', '2026-04-03 10:30:00', 'activo')`, 31, productoID, bodegaOrigenID, bodegaDestinoID); err != nil {
		t.Fatalf("insert movimiento traslado: %v", err)
	}
	if _, err := dbEmp.Exec(`INSERT INTO inventario_movimientos (
		empresa_id, producto_id, bodega_origen_id, bodega_destino_id, tipo, cantidad, costo_unitario, referencia, fecha_movimiento, estado
	) VALUES (?, ?, ?, NULL, 'salida', 1, 1000, 'REF-SALIDA', '2026-04-10 18:00:00', 'activo')`, 31, productoID, bodegaDestinoID); err != nil {
		t.Fatalf("insert movimiento salida: %v", err)
	}

	h := EmpresaInventarioMovimientosHandler(dbEmp)
	req := httptest.NewRequest(http.MethodGet, "/api/empresa/inventario/movimientos?empresa_id=31&bodega_id="+strconv.FormatInt(bodegaOrigenID, 10)+"&tipo=traslado&desde=2026-04-01&hasta=2026-04-05", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var rows []dbpkg.InventarioMovimiento
	if err := json.Unmarshal(rr.Body.Bytes(), &rows); err != nil {
		t.Fatalf("decode movimientos filtrados: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 movimiento filtrado, got %d", len(rows))
	}
	if rows[0].Tipo != "traslado" {
		t.Fatalf("expected tipo traslado, got %q", rows[0].Tipo)
	}
	if rows[0].Referencia != "REF-TRASLADO" {
		t.Fatalf("expected referencia REF-TRASLADO, got %q", rows[0].Referencia)
	}

	invalidDateReq := httptest.NewRequest(http.MethodGet, "/api/empresa/inventario/movimientos?empresa_id=31&desde=2026-99-01", nil)
	invalidDateRR := httptest.NewRecorder()
	h.ServeHTTP(invalidDateRR, invalidDateReq)
	if invalidDateRR.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid date status %d, got %d body=%s", http.StatusBadRequest, invalidDateRR.Code, invalidDateRR.Body.String())
	}
}

func TestEmpresaInventarioResumenHandlerDevuelveKPIsPorRango(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_inventario_resumen_handler.db")
	if err := dbpkg.EnsureEmpresaProductosSchema(dbEmp); err != nil {
		t.Fatalf("ensure productos schema: %v", err)
	}

	bodegaA, err := dbpkg.CreateBodega(dbEmp, dbpkg.Bodega{EmpresaID: 41, Codigo: "BOD-A-41", Nombre: "Bodega A"})
	if err != nil {
		t.Fatalf("create bodega A: %v", err)
	}
	bodegaB, err := dbpkg.CreateBodega(dbEmp, dbpkg.Bodega{EmpresaID: 41, Codigo: "BOD-B-41", Nombre: "Bodega B"})
	if err != nil {
		t.Fatalf("create bodega B: %v", err)
	}

	prodSinStock, err := dbpkg.CreateProducto(dbEmp, dbpkg.Producto{EmpresaID: 41, Nombre: "Harina", StockMinimo: 8, StockMaximo: 30}, 0, "TEST")
	if err != nil {
		t.Fatalf("create producto sin stock: %v", err)
	}
	prodBajoMin, err := dbpkg.CreateProducto(dbEmp, dbpkg.Producto{EmpresaID: 41, Nombre: "Sal", StockMinimo: 6, StockMaximo: 25}, 0, "TEST")
	if err != nil {
		t.Fatalf("create producto bajo minimo: %v", err)
	}

	if _, err := dbEmp.Exec(`INSERT INTO inventario_existencias (empresa_id, producto_id, bodega_id, cantidad, estado) VALUES (?, ?, ?, 0, 'activo')`, 41, prodSinStock, bodegaA); err != nil {
		t.Fatalf("insert existencia sin stock: %v", err)
	}
	if _, err := dbEmp.Exec(`INSERT INTO inventario_existencias (empresa_id, producto_id, bodega_id, cantidad, estado) VALUES (?, ?, ?, 3, 'activo')`, 41, prodBajoMin, bodegaB); err != nil {
		t.Fatalf("insert existencia bajo minimo: %v", err)
	}

	if _, err := dbEmp.Exec(`INSERT INTO inventario_movimientos (
		empresa_id, producto_id, bodega_origen_id, bodega_destino_id, tipo, cantidad, costo_unitario, referencia, fecha_movimiento, estado
	) VALUES (?, ?, NULL, ?, 'entrada', 4, 100, 'R-ENT', '2026-04-01 09:00:00', 'activo')`, 41, prodBajoMin, bodegaB); err != nil {
		t.Fatalf("insert movimiento entrada: %v", err)
	}
	if _, err := dbEmp.Exec(`INSERT INTO inventario_movimientos (
		empresa_id, producto_id, bodega_origen_id, bodega_destino_id, tipo, cantidad, costo_unitario, referencia, fecha_movimiento, estado
	) VALUES (?, ?, ?, ?, 'traslado', 2, 100, 'R-TRS', '2026-04-02 11:30:00', 'activo')`, 41, prodBajoMin, bodegaA, bodegaB); err != nil {
		t.Fatalf("insert movimiento traslado: %v", err)
	}
	if _, err := dbEmp.Exec(`INSERT INTO inventario_movimientos (
		empresa_id, producto_id, bodega_origen_id, bodega_destino_id, tipo, cantidad, costo_unitario, referencia, fecha_movimiento, estado
	) VALUES (?, ?, ?, NULL, 'ajuste_negativo', 1, 100, 'R-AJN', '2026-04-03 18:15:00', 'activo')`, 41, prodBajoMin, bodegaB); err != nil {
		t.Fatalf("insert movimiento ajuste negativo: %v", err)
	}
	if _, err := dbEmp.Exec(`INSERT INTO inventario_movimientos (
		empresa_id, producto_id, bodega_origen_id, bodega_destino_id, tipo, cantidad, costo_unitario, referencia, fecha_movimiento, estado
	) VALUES (?, ?, ?, NULL, 'salida', 1, 100, 'R-SAL-OLD', '2026-03-20 08:00:00', 'activo')`, 41, prodBajoMin, bodegaB); err != nil {
		t.Fatalf("insert movimiento salida fuera de rango: %v", err)
	}

	h := EmpresaInventarioResumenHandler(dbEmp)
	req := httptest.NewRequest(http.MethodGet, "/api/empresa/inventario/resumen?empresa_id=41&desde=2026-04-01&hasta=2026-04-05", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var resumen dbpkg.InventarioResumen
	if err := json.Unmarshal(rr.Body.Bytes(), &resumen); err != nil {
		t.Fatalf("decode resumen: %v", err)
	}
	if resumen.AlertasTotal != 2 || resumen.AlertasSinStock != 1 || resumen.AlertasBajoMinimo != 1 {
		t.Fatalf("unexpected alertas resumen: %+v", resumen)
	}
	if resumen.DeficitTotal != 11 {
		t.Fatalf("expected deficit_total=11, got %.2f", resumen.DeficitTotal)
	}
	if resumen.MovimientosTotal != 3 {
		t.Fatalf("expected movimientos_total=3 in range, got %d", resumen.MovimientosTotal)
	}
	if resumen.MovimientosEntrada != 1 || resumen.MovimientosSalida != 1 || resumen.MovimientosTraslado != 1 || resumen.MovimientosAjuste != 1 {
		t.Fatalf("unexpected movimientos resumen: %+v", resumen)
	}

	invalidReq := httptest.NewRequest(http.MethodGet, "/api/empresa/inventario/resumen?empresa_id=41&desde=2026-99-01", nil)
	invalidRR := httptest.NewRecorder()
	h.ServeHTTP(invalidRR, invalidReq)
	if invalidRR.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid date status %d, got %d body=%s", http.StatusBadRequest, invalidRR.Code, invalidRR.Body.String())
	}
}

func TestEmpresaInventarioTendenciaHandlerDevuelveSeriePorRango(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_inventario_tendencia_handler.db")
	if err := dbpkg.EnsureEmpresaProductosSchema(dbEmp); err != nil {
		t.Fatalf("ensure productos schema: %v", err)
	}

	bodegaA, err := dbpkg.CreateBodega(dbEmp, dbpkg.Bodega{EmpresaID: 51, Codigo: "BOD-A-51", Nombre: "A"})
	if err != nil {
		t.Fatalf("create bodega A: %v", err)
	}
	bodegaB, err := dbpkg.CreateBodega(dbEmp, dbpkg.Bodega{EmpresaID: 51, Codigo: "BOD-B-51", Nombre: "B"})
	if err != nil {
		t.Fatalf("create bodega B: %v", err)
	}

	productoID, err := dbpkg.CreateProducto(dbEmp, dbpkg.Producto{EmpresaID: 51, Nombre: "Producto tendencia", StockMinimo: 1, StockMaximo: 20}, 0, "TEST")
	if err != nil {
		t.Fatalf("create producto: %v", err)
	}

	if _, err := dbEmp.Exec(`INSERT INTO inventario_movimientos (
		empresa_id, producto_id, bodega_origen_id, bodega_destino_id, tipo, cantidad, costo_unitario, referencia, fecha_movimiento, estado
	) VALUES (?, ?, NULL, ?, 'entrada', 4, 100, 'TEND-1', '2026-04-01 09:00:00', 'activo')`, 51, productoID, bodegaA); err != nil {
		t.Fatalf("insert tendencia 1: %v", err)
	}
	if _, err := dbEmp.Exec(`INSERT INTO inventario_movimientos (
		empresa_id, producto_id, bodega_origen_id, bodega_destino_id, tipo, cantidad, costo_unitario, referencia, fecha_movimiento, estado
	) VALUES (?, ?, ?, NULL, 'salida', 1, 100, 'TEND-2', '2026-04-02 12:00:00', 'activo')`, 51, productoID, bodegaA); err != nil {
		t.Fatalf("insert tendencia 2: %v", err)
	}
	if _, err := dbEmp.Exec(`INSERT INTO inventario_movimientos (
		empresa_id, producto_id, bodega_origen_id, bodega_destino_id, tipo, cantidad, costo_unitario, referencia, fecha_movimiento, estado
	) VALUES (?, ?, ?, ?, 'traslado', 2, 100, 'TEND-3', '2026-04-03 15:00:00', 'activo')`, 51, productoID, bodegaA, bodegaB); err != nil {
		t.Fatalf("insert tendencia 3: %v", err)
	}

	h := EmpresaInventarioTendenciaHandler(dbEmp)
	req := httptest.NewRequest(http.MethodGet, "/api/empresa/inventario/tendencia?empresa_id=51&bodega_id="+strconv.FormatInt(bodegaA, 10)+"&desde=2026-04-01&hasta=2026-04-03", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var serie []dbpkg.InventarioTendenciaDia
	if err := json.Unmarshal(rr.Body.Bytes(), &serie); err != nil {
		t.Fatalf("decode tendencia: %v", err)
	}
	if len(serie) != 3 {
		t.Fatalf("expected 3 dias en tendencia, got %d", len(serie))
	}
	if serie[0].Fecha != "2026-04-01" || serie[0].Entradas != 4 {
		t.Fatalf("unexpected dia 1: %+v", serie[0])
	}
	if serie[1].Fecha != "2026-04-02" || serie[1].Salidas != 1 {
		t.Fatalf("unexpected dia 2: %+v", serie[1])
	}

	invalidReq := httptest.NewRequest(http.MethodGet, "/api/empresa/inventario/tendencia?empresa_id=51&desde=2026-99-01", nil)
	invalidRR := httptest.NewRecorder()
	h.ServeHTTP(invalidRR, invalidReq)
	if invalidRR.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid date status %d, got %d body=%s", http.StatusBadRequest, invalidRR.Code, invalidRR.Body.String())
	}
}

func TestEmpresaInventarioBalanceBodegasHandlerDevuelveResumenPorBodega(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_inventario_balance_bodegas_handler.db")
	if err := dbpkg.EnsureEmpresaProductosSchema(dbEmp); err != nil {
		t.Fatalf("ensure productos schema: %v", err)
	}

	bodegaA, err := dbpkg.CreateBodega(dbEmp, dbpkg.Bodega{EmpresaID: 61, Codigo: "BOD-A-61", Nombre: "A"})
	if err != nil {
		t.Fatalf("create bodega A: %v", err)
	}
	bodegaB, err := dbpkg.CreateBodega(dbEmp, dbpkg.Bodega{EmpresaID: 61, Codigo: "BOD-B-61", Nombre: "B"})
	if err != nil {
		t.Fatalf("create bodega B: %v", err)
	}

	productoID, err := dbpkg.CreateProducto(dbEmp, dbpkg.Producto{EmpresaID: 61, Nombre: "Producto balance", StockMinimo: 1, StockMaximo: 20}, 0, "TEST")
	if err != nil {
		t.Fatalf("create producto: %v", err)
	}

	if _, err := dbEmp.Exec(`INSERT INTO inventario_movimientos (
		empresa_id, producto_id, bodega_origen_id, bodega_destino_id, tipo, cantidad, costo_unitario, referencia, fecha_movimiento, estado
	) VALUES (?, ?, NULL, ?, 'entrada', 4, 100, 'BAL-H-1', '2026-04-01 09:00:00', 'activo')`, 61, productoID, bodegaA); err != nil {
		t.Fatalf("insert movimiento 1: %v", err)
	}
	if _, err := dbEmp.Exec(`INSERT INTO inventario_movimientos (
		empresa_id, producto_id, bodega_origen_id, bodega_destino_id, tipo, cantidad, costo_unitario, referencia, fecha_movimiento, estado
	) VALUES (?, ?, ?, ?, 'traslado', 1, 100, 'BAL-H-2', '2026-04-02 13:10:00', 'activo')`, 61, productoID, bodegaA, bodegaB); err != nil {
		t.Fatalf("insert movimiento 2: %v", err)
	}

	h := EmpresaInventarioBalanceBodegasHandler(dbEmp)
	req := httptest.NewRequest(http.MethodGet, "/api/empresa/inventario/balance_bodegas?empresa_id=61&bodega_id="+strconv.FormatInt(bodegaA, 10)+"&desde=2026-04-01&hasta=2026-04-03", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var rows []dbpkg.InventarioBalanceBodega
	if err := json.Unmarshal(rr.Body.Bytes(), &rows); err != nil {
		t.Fatalf("decode balance bodegas: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row de balance por filtro bodega, got %d", len(rows))
	}
	if rows[0].BodegaID != bodegaA || rows[0].Entradas != 4 || rows[0].TrasladosSalida != 1 || rows[0].Neto != 3 {
		t.Fatalf("unexpected balance row: %+v", rows[0])
	}

	invalidReq := httptest.NewRequest(http.MethodGet, "/api/empresa/inventario/balance_bodegas?empresa_id=61&hasta=2026-99-01", nil)
	invalidRR := httptest.NewRecorder()
	h.ServeHTTP(invalidRR, invalidReq)
	if invalidRR.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid date status %d, got %d body=%s", http.StatusBadRequest, invalidRR.Code, invalidRR.Body.String())
	}
}

func TestEmpresaInventarioProyeccionQuiebreHandlerDevuelveRiesgo(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_inventario_proyeccion_quiebre_handler.db")
	if err := dbpkg.EnsureEmpresaProductosSchema(dbEmp); err != nil {
		t.Fatalf("ensure productos schema: %v", err)
	}

	bodegaA, err := dbpkg.CreateBodega(dbEmp, dbpkg.Bodega{EmpresaID: 71, Codigo: "BOD-A-71", Nombre: "A"})
	if err != nil {
		t.Fatalf("create bodega A: %v", err)
	}

	productoID, err := dbpkg.CreateProducto(dbEmp, dbpkg.Producto{EmpresaID: 71, Nombre: "Producto proyectado", StockMinimo: 4, StockMaximo: 18}, 0, "TEST")
	if err != nil {
		t.Fatalf("create producto: %v", err)
	}

	if _, err := dbEmp.Exec(`INSERT INTO inventario_existencias (empresa_id, producto_id, bodega_id, cantidad, estado) VALUES (71, ?, ?, 3, 'activo')`, productoID, bodegaA); err != nil {
		t.Fatalf("insert existencia: %v", err)
	}
	if _, err := dbEmp.Exec(`INSERT INTO inventario_movimientos (
		empresa_id, producto_id, bodega_origen_id, bodega_destino_id, tipo, cantidad, costo_unitario, referencia, fecha_movimiento, estado
	) VALUES (71, ?, ?, NULL, 'salida', 6, 100, 'PROY-H-1', datetime('now','-1 day'), 'activo')`, productoID, bodegaA); err != nil {
		t.Fatalf("insert movimiento salida: %v", err)
	}

	h := EmpresaInventarioProyeccionQuiebreHandler(dbEmp)
	req := httptest.NewRequest(http.MethodGet, "/api/empresa/inventario/proyeccion_quiebre?empresa_id=71&dias_ventana=7&limit=20", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var rows []dbpkg.InventarioProyeccionQuiebre
	if err := json.Unmarshal(rr.Body.Bytes(), &rows); err != nil {
		t.Fatalf("decode proyeccion quiebre: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row in proyeccion quiebre, got %d", len(rows))
	}
	if rows[0].ProductoID != productoID {
		t.Fatalf("expected producto_id=%d, got %d", productoID, rows[0].ProductoID)
	}
	if rows[0].EstadoProyeccion == "estable" {
		t.Fatalf("expected non-estable estado, got %+v", rows[0])
	}

	invalidReq := httptest.NewRequest(http.MethodGet, "/api/empresa/inventario/proyeccion_quiebre?empresa_id=71&dias_ventana=abc", nil)
	invalidRR := httptest.NewRecorder()
	h.ServeHTTP(invalidRR, invalidReq)
	if invalidRR.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid dias_ventana status %d, got %d body=%s", http.StatusBadRequest, invalidRR.Code, invalidRR.Body.String())
	}
}

func TestEmpresaInventarioPlanReposicionHandlerDevuelveCostoEstimado(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_inventario_plan_reposicion_handler.db")
	if err := dbpkg.EnsureEmpresaProductosSchema(dbEmp); err != nil {
		t.Fatalf("ensure productos schema: %v", err)
	}

	bodegaA, err := dbpkg.CreateBodega(dbEmp, dbpkg.Bodega{EmpresaID: 72, Codigo: "BOD-A-72", Nombre: "A"})
	if err != nil {
		t.Fatalf("create bodega A: %v", err)
	}

	proveedorID, err := dbpkg.CreateProveedor(dbEmp, dbpkg.Proveedor{EmpresaID: 72, Nombre: "Proveedor Test"})
	if err != nil {
		t.Fatalf("create proveedor: %v", err)
	}

	productoID, err := dbpkg.CreateProducto(dbEmp, dbpkg.Producto{EmpresaID: 72, Nombre: "Producto plan", ProveedorPrincipalID: proveedorID, Costo: 9.5, StockMinimo: 3, StockMaximo: 14}, 0, "TEST")
	if err != nil {
		t.Fatalf("create producto: %v", err)
	}

	if _, err := dbEmp.Exec(`INSERT INTO inventario_existencias (empresa_id, producto_id, bodega_id, cantidad, estado) VALUES (72, ?, ?, 2, 'activo')`, productoID, bodegaA); err != nil {
		t.Fatalf("insert existencia: %v", err)
	}
	if _, err := dbEmp.Exec(`INSERT INTO inventario_movimientos (
		empresa_id, producto_id, bodega_origen_id, bodega_destino_id, tipo, cantidad, costo_unitario, referencia, fecha_movimiento, estado
	) VALUES (72, ?, ?, NULL, 'salida', 6, 9.5, 'PLAN-H-1', datetime('now','-1 day'), 'activo')`, productoID, bodegaA); err != nil {
		t.Fatalf("insert movimiento salida: %v", err)
	}

	h := EmpresaInventarioPlanReposicionHandler(dbEmp)
	req := httptest.NewRequest(http.MethodGet, "/api/empresa/inventario/plan_reposicion?empresa_id=72&dias_ventana=7&solo_riesgo=1", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var rows []dbpkg.InventarioPlanReposicionItem
	if err := json.Unmarshal(rr.Body.Bytes(), &rows); err != nil {
		t.Fatalf("decode plan reposicion: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row in plan reposicion, got %d", len(rows))
	}
	if rows[0].ProductoID != productoID {
		t.Fatalf("expected producto_id=%d, got %d", productoID, rows[0].ProductoID)
	}
	if rows[0].CostoEstimado <= 0 {
		t.Fatalf("expected costo_estimado > 0, got %+v", rows[0])
	}

	invalidReq := httptest.NewRequest(http.MethodGet, "/api/empresa/inventario/plan_reposicion?empresa_id=72&solo_riesgo=quizas", nil)
	invalidRR := httptest.NewRecorder()
	h.ServeHTTP(invalidRR, invalidReq)
	if invalidRR.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid solo_riesgo status %d, got %d body=%s", http.StatusBadRequest, invalidRR.Code, invalidRR.Body.String())
	}
}

func TestEmpresaInventarioPlanReposicionResumenHandlerAgrupaProveedor(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_inventario_plan_reposicion_resumen_handler.db")
	if err := dbpkg.EnsureEmpresaProductosSchema(dbEmp); err != nil {
		t.Fatalf("ensure productos schema: %v", err)
	}

	bodegaA, err := dbpkg.CreateBodega(dbEmp, dbpkg.Bodega{EmpresaID: 73, Codigo: "BOD-A-73", Nombre: "A"})
	if err != nil {
		t.Fatalf("create bodega A: %v", err)
	}

	proveedorID, err := dbpkg.CreateProveedor(dbEmp, dbpkg.Proveedor{EmpresaID: 73, Nombre: "Proveedor Resumen"})
	if err != nil {
		t.Fatalf("create proveedor: %v", err)
	}

	productoID, err := dbpkg.CreateProducto(dbEmp, dbpkg.Producto{EmpresaID: 73, Nombre: "Producto resumen", ProveedorPrincipalID: proveedorID, Costo: 10, StockMinimo: 2, StockMaximo: 12}, 0, "TEST")
	if err != nil {
		t.Fatalf("create producto: %v", err)
	}

	if _, err := dbEmp.Exec(`INSERT INTO inventario_existencias (empresa_id, producto_id, bodega_id, cantidad, estado) VALUES (73, ?, ?, 1, 'activo')`, productoID, bodegaA); err != nil {
		t.Fatalf("insert existencia: %v", err)
	}
	if _, err := dbEmp.Exec(`INSERT INTO inventario_movimientos (
		empresa_id, producto_id, bodega_origen_id, bodega_destino_id, tipo, cantidad, costo_unitario, referencia, fecha_movimiento, estado
	) VALUES (73, ?, ?, NULL, 'salida', 4, 10, 'RES-H-1', datetime('now','-1 day'), 'activo')`, productoID, bodegaA); err != nil {
		t.Fatalf("insert movimiento salida: %v", err)
	}

	h := EmpresaInventarioPlanReposicionResumenHandler(dbEmp)
	req := httptest.NewRequest(http.MethodGet, "/api/empresa/inventario/plan_reposicion_resumen?empresa_id=73&dias_ventana=7&solo_riesgo=1", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var rows []dbpkg.InventarioPlanReposicionProveedorResumen
	if err := json.Unmarshal(rr.Body.Bytes(), &rows); err != nil {
		t.Fatalf("decode plan reposicion resumen: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row in plan reposicion resumen, got %d", len(rows))
	}
	if rows[0].ProveedorID != proveedorID {
		t.Fatalf("expected proveedor_id=%d, got %d", proveedorID, rows[0].ProveedorID)
	}
	if rows[0].CostoTotal <= 0 {
		t.Fatalf("expected costo_total > 0, got %+v", rows[0])
	}

	invalidReq := httptest.NewRequest(http.MethodGet, "/api/empresa/inventario/plan_reposicion_resumen?empresa_id=73&solo_riesgo=quizas", nil)
	invalidRR := httptest.NewRecorder()
	h.ServeHTTP(invalidRR, invalidReq)
	if invalidRR.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid solo_riesgo status %d, got %d body=%s", http.StatusBadRequest, invalidRR.Code, invalidRR.Body.String())
	}
}

func TestEmpresaInventarioPlanReposicionBorradorHandlerConstruyeDocumento(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_inventario_plan_reposicion_borrador_handler.db")
	if err := dbpkg.EnsureEmpresaProductosSchema(dbEmp); err != nil {
		t.Fatalf("ensure productos schema: %v", err)
	}

	bodegaA, err := dbpkg.CreateBodega(dbEmp, dbpkg.Bodega{EmpresaID: 74, Codigo: "BOD-A-74", Nombre: "A"})
	if err != nil {
		t.Fatalf("create bodega A: %v", err)
	}

	proveedorID, err := dbpkg.CreateProveedor(dbEmp, dbpkg.Proveedor{EmpresaID: 74, Nombre: "Proveedor Borrador"})
	if err != nil {
		t.Fatalf("create proveedor: %v", err)
	}

	productoID, err := dbpkg.CreateProducto(dbEmp, dbpkg.Producto{EmpresaID: 74, Nombre: "Producto borrador", ProveedorPrincipalID: proveedorID, Costo: 11, StockMinimo: 3, StockMaximo: 14}, 0, "TEST")
	if err != nil {
		t.Fatalf("create producto: %v", err)
	}

	if _, err := dbEmp.Exec(`INSERT INTO inventario_existencias (empresa_id, producto_id, bodega_id, cantidad, estado) VALUES (74, ?, ?, 1, 'activo')`, productoID, bodegaA); err != nil {
		t.Fatalf("insert existencia: %v", err)
	}
	if _, err := dbEmp.Exec(`INSERT INTO inventario_movimientos (
		empresa_id, producto_id, bodega_origen_id, bodega_destino_id, tipo, cantidad, costo_unitario, referencia, fecha_movimiento, estado
	) VALUES (74, ?, ?, NULL, 'salida', 5, 11, 'BORR-H-1', datetime('now','-1 day'), 'activo')`, productoID, bodegaA); err != nil {
		t.Fatalf("insert movimiento salida: %v", err)
	}

	h := EmpresaInventarioPlanReposicionBorradorHandler(dbEmp)
	req := httptest.NewRequest(http.MethodGet, "/api/empresa/inventario/plan_reposicion_borrador?empresa_id=74&proveedor_id="+strconv.FormatInt(proveedorID, 10)+"&dias_ventana=7&solo_riesgo=1", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var row dbpkg.InventarioPlanReposicionBorradorCompra
	if err := json.Unmarshal(rr.Body.Bytes(), &row); err != nil {
		t.Fatalf("decode plan reposicion borrador: %v", err)
	}
	if row.ProveedorID != proveedorID {
		t.Fatalf("expected proveedor_id=%d, got %d", proveedorID, row.ProveedorID)
	}
	if row.TotalItems != 1 {
		t.Fatalf("expected total_items=1, got %+v", row)
	}
	if row.CostoTotal <= 0 {
		t.Fatalf("expected costo_total > 0, got %+v", row)
	}

	invalidReq := httptest.NewRequest(http.MethodGet, "/api/empresa/inventario/plan_reposicion_borrador?empresa_id=74", nil)
	invalidRR := httptest.NewRecorder()
	h.ServeHTTP(invalidRR, invalidReq)
	if invalidRR.Code != http.StatusBadRequest {
		t.Fatalf("expected missing proveedor_id status %d, got %d body=%s", http.StatusBadRequest, invalidRR.Code, invalidRR.Body.String())
	}

	invalidSoloReq := httptest.NewRequest(http.MethodGet, "/api/empresa/inventario/plan_reposicion_borrador?empresa_id=74&proveedor_id="+strconv.FormatInt(proveedorID, 10)+"&solo_riesgo=quizas", nil)
	invalidSoloRR := httptest.NewRecorder()
	h.ServeHTTP(invalidSoloRR, invalidSoloReq)
	if invalidSoloRR.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid solo_riesgo status %d, got %d body=%s", http.StatusBadRequest, invalidSoloRR.Code, invalidSoloRR.Body.String())
	}
}

func TestEmpresaComprasPlanReposicionEmitirOrdenHandlerEmiteDocumento(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_compras_plan_reposicion_emitir_handler.db")
	if err := dbpkg.EnsureEmpresaProductosSchema(dbEmp); err != nil {
		t.Fatalf("ensure productos schema: %v", err)
	}
	if err := dbpkg.EnsureEmpresaDocumentosTransaccionalesSchema(dbEmp); err != nil {
		t.Fatalf("ensure documentos transaccionales schema: %v", err)
	}
	if err := dbpkg.EnsureEmpresaEventosContablesSchema(dbEmp); err != nil {
		t.Fatalf("ensure eventos contables schema: %v", err)
	}

	bodegaA, err := dbpkg.CreateBodega(dbEmp, dbpkg.Bodega{EmpresaID: 75, Codigo: "BOD-A-75", Nombre: "A"})
	if err != nil {
		t.Fatalf("create bodega A: %v", err)
	}

	proveedorID, err := dbpkg.CreateProveedor(dbEmp, dbpkg.Proveedor{EmpresaID: 75, Nombre: "Proveedor F11"})
	if err != nil {
		t.Fatalf("create proveedor: %v", err)
	}

	productoID, err := dbpkg.CreateProducto(dbEmp, dbpkg.Producto{EmpresaID: 75, Nombre: "Producto F11", ProveedorPrincipalID: proveedorID, Costo: 14, StockMinimo: 3, StockMaximo: 16}, 0, "TEST")
	if err != nil {
		t.Fatalf("create producto: %v", err)
	}

	if _, err := dbEmp.Exec(`INSERT INTO inventario_existencias (empresa_id, producto_id, bodega_id, cantidad, estado) VALUES (75, ?, ?, 1, 'activo')`, productoID, bodegaA); err != nil {
		t.Fatalf("insert existencia: %v", err)
	}
	if _, err := dbEmp.Exec(`INSERT INTO inventario_movimientos (
		empresa_id, producto_id, bodega_origen_id, bodega_destino_id, tipo, cantidad, costo_unitario, referencia, fecha_movimiento, estado
	) VALUES (75, ?, ?, NULL, 'salida', 5, 14, 'EMI-H-1', datetime('now','-1 day'), 'activo')`, productoID, bodegaA); err != nil {
		t.Fatalf("insert salida: %v", err)
	}

	h := EmpresaComprasPlanReposicionEmitirOrdenHandler(dbEmp)
	req := httptest.NewRequest(http.MethodPost, "/api/empresa/compras/plan_reposicion/emitir_orden", strings.NewReader(`{"empresa_id":75,"proveedor_id":`+strconv.FormatInt(proveedorID, 10)+`,"bodega_id":`+strconv.FormatInt(bodegaA, 10)+`,"dias_ventana":7,"solo_riesgo":true,"periodo_contable":"2026-04","moneda":"COP"}`))
	req = req.WithContext(context.WithValue(req.Context(), "adminEmail", "compras-f11@test.com"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var resp struct {
		OK        bool                                       `json:"ok"`
		Resultado dbpkg.InventarioPlanReposicionOrdenEmitida `json:"resultado"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode emitir orden response: %v", err)
	}
	if !resp.OK {
		t.Fatalf("expected ok=true, got %+v", resp)
	}
	if resp.Resultado.EstadoNuevo != "emitida" {
		t.Fatalf("expected estado_nuevo emitida, got %+v", resp.Resultado)
	}
	if resp.Resultado.EntidadID <= 0 {
		t.Fatalf("expected entidad_id > 0, got %+v", resp.Resultado)
	}

	doc, err := dbpkg.GetEmpresaDocumentoCompraByCodigo(dbEmp, 75, "orden_compra", resp.Resultado.DocumentoCodigo)
	if err != nil {
		t.Fatalf("get emitted doc by codigo: %v", err)
	}
	if doc.EstadoDocumento != "emitida" {
		t.Fatalf("expected estado emitida, got %+v", doc)
	}

	invalidReq := httptest.NewRequest(http.MethodPost, "/api/empresa/compras/plan_reposicion/emitir_orden", strings.NewReader(`{"empresa_id":75}`))
	invalidReq.Header.Set("Content-Type", "application/json")
	invalidRR := httptest.NewRecorder()
	h.ServeHTTP(invalidRR, invalidReq)
	if invalidRR.Code != http.StatusBadRequest {
		t.Fatalf("expected missing proveedor_id status %d, got %d body=%s", http.StatusBadRequest, invalidRR.Code, invalidRR.Body.String())
	}
}

func TestEmpresaComprasPlanReposicionActualizarEstadoHandlerGestionaCiclo(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_compras_plan_reposicion_actualizar_estado_handler.db")
	if err := dbpkg.EnsureEmpresaProductosSchema(dbEmp); err != nil {
		t.Fatalf("ensure productos schema: %v", err)
	}
	if err := dbpkg.EnsureEmpresaDocumentosTransaccionalesSchema(dbEmp); err != nil {
		t.Fatalf("ensure documentos transaccionales schema: %v", err)
	}
	if err := dbpkg.EnsureEmpresaEventosContablesSchema(dbEmp); err != nil {
		t.Fatalf("ensure eventos contables schema: %v", err)
	}

	bodegaA, err := dbpkg.CreateBodega(dbEmp, dbpkg.Bodega{EmpresaID: 76, Codigo: "BOD-A-76", Nombre: "A"})
	if err != nil {
		t.Fatalf("create bodega A: %v", err)
	}

	proveedorID, err := dbpkg.CreateProveedor(dbEmp, dbpkg.Proveedor{EmpresaID: 76, Nombre: "Proveedor F12"})
	if err != nil {
		t.Fatalf("create proveedor: %v", err)
	}

	productoID, err := dbpkg.CreateProducto(dbEmp, dbpkg.Producto{EmpresaID: 76, Nombre: "Producto F12", ProveedorPrincipalID: proveedorID, Costo: 18, StockMinimo: 4, StockMaximo: 18}, 0, "TEST")
	if err != nil {
		t.Fatalf("create producto: %v", err)
	}

	if _, err := dbEmp.Exec(`INSERT INTO inventario_existencias (empresa_id, producto_id, bodega_id, cantidad, estado) VALUES (76, ?, ?, 1, 'activo')`, productoID, bodegaA); err != nil {
		t.Fatalf("insert existencia: %v", err)
	}
	if _, err := dbEmp.Exec(`INSERT INTO inventario_movimientos (
		empresa_id, producto_id, bodega_origen_id, bodega_destino_id, tipo, cantidad, costo_unitario, referencia, fecha_movimiento, estado
	) VALUES (76, ?, ?, NULL, 'salida', 7, 18, 'F12-H-1', datetime('now','-1 day'), 'activo')`, productoID, bodegaA); err != nil {
		t.Fatalf("insert salida: %v", err)
	}

	hEmitir := EmpresaComprasPlanReposicionEmitirOrdenHandler(dbEmp)
	reqEmitir := httptest.NewRequest(http.MethodPost, "/api/empresa/compras/plan_reposicion/emitir_orden", strings.NewReader(`{"empresa_id":76,"proveedor_id":`+strconv.FormatInt(proveedorID, 10)+`,"bodega_id":`+strconv.FormatInt(bodegaA, 10)+`,"dias_ventana":7,"solo_riesgo":true,"periodo_contable":"2026-04","moneda":"COP"}`))
	reqEmitir = reqEmitir.WithContext(context.WithValue(reqEmitir.Context(), "adminEmail", "compras-f12@test.com"))
	reqEmitir.Header.Set("Content-Type", "application/json")
	rrEmitir := httptest.NewRecorder()
	hEmitir.ServeHTTP(rrEmitir, reqEmitir)
	if rrEmitir.Code != http.StatusOK {
		t.Fatalf("expected emitir status %d, got %d body=%s", http.StatusOK, rrEmitir.Code, rrEmitir.Body.String())
	}

	var emitirResp struct {
		OK        bool                                       `json:"ok"`
		Resultado dbpkg.InventarioPlanReposicionOrdenEmitida `json:"resultado"`
	}
	if err := json.Unmarshal(rrEmitir.Body.Bytes(), &emitirResp); err != nil {
		t.Fatalf("decode emitir response: %v", err)
	}
	if !emitirResp.OK || strings.TrimSpace(emitirResp.Resultado.DocumentoCodigo) == "" {
		t.Fatalf("respuesta emitir invalida: %+v", emitirResp)
	}

	h := EmpresaComprasPlanReposicionActualizarEstadoHandler(dbEmp)
	reqRecepcionar := httptest.NewRequest(http.MethodPost, "/api/empresa/compras/plan_reposicion/actualizar_estado", strings.NewReader(`{"empresa_id":76,"proveedor_id":`+strconv.FormatInt(proveedorID, 10)+`,"documento_codigo":"`+emitirResp.Resultado.DocumentoCodigo+`","accion":"recepcionar_compra","periodo_contable":"2026-04"}`))
	reqRecepcionar = reqRecepcionar.WithContext(context.WithValue(reqRecepcionar.Context(), "adminEmail", "compras-f12@test.com"))
	reqRecepcionar.Header.Set("Content-Type", "application/json")
	rrRecepcionar := httptest.NewRecorder()
	h.ServeHTTP(rrRecepcionar, reqRecepcionar)
	if rrRecepcionar.Code != http.StatusOK {
		t.Fatalf("expected recepcionar status %d, got %d body=%s", http.StatusOK, rrRecepcionar.Code, rrRecepcionar.Body.String())
	}

	var recepResp struct {
		OK        bool                                                 `json:"ok"`
		Resultado dbpkg.InventarioPlanReposicionOrdenEstadoActualizado `json:"resultado"`
	}
	if err := json.Unmarshal(rrRecepcionar.Body.Bytes(), &recepResp); err != nil {
		t.Fatalf("decode recepcionar response: %v", err)
	}
	if !recepResp.OK || recepResp.Resultado.EstadoNuevo != "recepcionada" {
		t.Fatalf("respuesta recepcionar invalida: %+v", recepResp)
	}

	reqContabilizar := httptest.NewRequest(http.MethodPost, "/api/empresa/compras/plan_reposicion/actualizar_estado", strings.NewReader(`{"empresa_id":76,"proveedor_id":`+strconv.FormatInt(proveedorID, 10)+`,"documento_codigo":"`+emitirResp.Resultado.DocumentoCodigo+`","accion":"contabilizar_compra","periodo_contable":"2026-04"}`))
	reqContabilizar = reqContabilizar.WithContext(context.WithValue(reqContabilizar.Context(), "adminEmail", "compras-f12@test.com"))
	reqContabilizar.Header.Set("Content-Type", "application/json")
	rrContabilizar := httptest.NewRecorder()
	h.ServeHTTP(rrContabilizar, reqContabilizar)
	if rrContabilizar.Code != http.StatusOK {
		t.Fatalf("expected contabilizar status %d, got %d body=%s", http.StatusOK, rrContabilizar.Code, rrContabilizar.Body.String())
	}

	var contResp struct {
		OK        bool                                                 `json:"ok"`
		Resultado dbpkg.InventarioPlanReposicionOrdenEstadoActualizado `json:"resultado"`
	}
	if err := json.Unmarshal(rrContabilizar.Body.Bytes(), &contResp); err != nil {
		t.Fatalf("decode contabilizar response: %v", err)
	}
	if !contResp.OK || contResp.Resultado.EstadoNuevo != "contabilizada" {
		t.Fatalf("respuesta contabilizar invalida: %+v", contResp)
	}

	reqConflict := httptest.NewRequest(http.MethodPost, "/api/empresa/compras/plan_reposicion/actualizar_estado", strings.NewReader(`{"empresa_id":76,"proveedor_id":`+strconv.FormatInt(proveedorID, 10)+`,"documento_codigo":"`+emitirResp.Resultado.DocumentoCodigo+`","accion":"recepcionar_compra"}`))
	reqConflict.Header.Set("Content-Type", "application/json")
	rrConflict := httptest.NewRecorder()
	h.ServeHTTP(rrConflict, reqConflict)
	if rrConflict.Code != http.StatusConflict {
		t.Fatalf("expected conflict status %d, got %d body=%s", http.StatusConflict, rrConflict.Code, rrConflict.Body.String())
	}
}

func TestEmpresaInventarioConfiguracionYConteoCiclicoHandler(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_inventario_config_conteo_handler.db")
	if err := dbpkg.EnsureEmpresaProductosSchema(dbEmp); err != nil {
		t.Fatalf("ensure productos schema: %v", err)
	}

	bodegaID, err := dbpkg.CreateBodega(dbEmp, dbpkg.Bodega{EmpresaID: 120, Codigo: "BOD-120", Nombre: "Principal"})
	if err != nil {
		t.Fatalf("create bodega: %v", err)
	}
	productoID, err := dbpkg.CreateProducto(dbEmp, dbpkg.Producto{EmpresaID: 120, BodegaPrincipalID: bodegaID, Nombre: "Producto conteo", Costo: 12, StockMinimo: 1, StockMaximo: 20}, 0, "TEST")
	if err != nil {
		t.Fatalf("create producto: %v", err)
	}
	if err := dbpkg.RegistrarMovimientoInventario(dbEmp, 120, productoID, bodegaID, "entrada", 10, "H-CNT-ENT", "tester", "stock inicial"); err != nil {
		t.Fatalf("entrada inicial: %v", err)
	}

	hConfig := EmpresaInventarioConfiguracionHandler(dbEmp)
	putReq := httptest.NewRequest(http.MethodPut, "/api/empresa/inventario/configuracion", strings.NewReader(`{"empresa_id":120,"politica_costo":"peps","observaciones":"politica fase 11"}`))
	putReq = putReq.WithContext(context.WithValue(putReq.Context(), "adminEmail", "inventario@test.com"))
	putReq.Header.Set("Content-Type", "application/json")
	putRR := httptest.NewRecorder()
	hConfig.ServeHTTP(putRR, putReq)
	if putRR.Code != http.StatusOK {
		t.Fatalf("expected config put status %d, got %d body=%s", http.StatusOK, putRR.Code, putRR.Body.String())
	}

	var cfg dbpkg.EmpresaInventarioConfiguracion
	if err := json.Unmarshal(putRR.Body.Bytes(), &cfg); err != nil {
		t.Fatalf("decode config put response: %v", err)
	}
	if cfg.PoliticaCosto != "peps" {
		t.Fatalf("expected politica peps, got %+v", cfg)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/api/empresa/inventario/configuracion?empresa_id=120", nil)
	getRR := httptest.NewRecorder()
	hConfig.ServeHTTP(getRR, getReq)
	if getRR.Code != http.StatusOK {
		t.Fatalf("expected config get status %d, got %d body=%s", http.StatusOK, getRR.Code, getRR.Body.String())
	}

	hConteo := EmpresaInventarioConteoCiclicoHandler(dbEmp)
	postReq := httptest.NewRequest(http.MethodPost, "/api/empresa/inventario/conteo_ciclico", strings.NewReader(`{"empresa_id":120,"producto_id":`+strconv.FormatInt(productoID, 10)+`,"bodega_id":`+strconv.FormatInt(bodegaID, 10)+`,"cantidad_contada":7,"referencia":"H-CNT-001","observaciones":"conteo semanal"}`))
	postReq = postReq.WithContext(context.WithValue(postReq.Context(), "adminEmail", "auditor@test.com"))
	postReq.Header.Set("Content-Type", "application/json")
	postRR := httptest.NewRecorder()
	hConteo.ServeHTTP(postRR, postReq)
	if postRR.Code != http.StatusOK {
		t.Fatalf("expected conteo post status %d, got %d body=%s", http.StatusOK, postRR.Code, postRR.Body.String())
	}

	var conteoResp struct {
		ID        int64                         `json:"id"`
		Resultado dbpkg.InventarioConteoCiclico `json:"resultado"`
	}
	if err := json.Unmarshal(postRR.Body.Bytes(), &conteoResp); err != nil {
		t.Fatalf("decode conteo post response: %v", err)
	}
	if conteoResp.ID <= 0 || conteoResp.Resultado.TipoAjuste != "ajuste_negativo" {
		t.Fatalf("unexpected conteo response: %+v", conteoResp)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/empresa/inventario/conteo_ciclico?empresa_id=120&producto_id="+strconv.FormatInt(productoID, 10), nil)
	listRR := httptest.NewRecorder()
	hConteo.ServeHTTP(listRR, listReq)
	if listRR.Code != http.StatusOK {
		t.Fatalf("expected conteo list status %d, got %d body=%s", http.StatusOK, listRR.Code, listRR.Body.String())
	}

	var conteos []dbpkg.InventarioConteoCiclico
	if err := json.Unmarshal(listRR.Body.Bytes(), &conteos); err != nil {
		t.Fatalf("decode conteo list response: %v", err)
	}
	if len(conteos) != 1 {
		t.Fatalf("expected 1 conteo row, got %d", len(conteos))
	}
}

func TestEmpresaInventarioAlertasHandlerProactivasIncluyeSobrestock(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_inventario_alertas_proactivas_handler.db")
	if err := dbpkg.EnsureEmpresaProductosSchema(dbEmp); err != nil {
		t.Fatalf("ensure productos schema: %v", err)
	}

	bodegaID, err := dbpkg.CreateBodega(dbEmp, dbpkg.Bodega{EmpresaID: 121, Codigo: "BOD-121", Nombre: "Bodega"})
	if err != nil {
		t.Fatalf("create bodega: %v", err)
	}
	productoID, err := dbpkg.CreateProducto(dbEmp, dbpkg.Producto{EmpresaID: 121, Nombre: "Producto sobrestock", StockMinimo: 2, StockMaximo: 5}, 0, "TEST")
	if err != nil {
		t.Fatalf("create producto: %v", err)
	}
	if _, err := dbEmp.Exec(`INSERT INTO inventario_existencias (empresa_id, producto_id, bodega_id, cantidad, estado) VALUES (121, ?, ?, 8, 'activo')`, productoID, bodegaID); err != nil {
		t.Fatalf("insert existencia sobrestock: %v", err)
	}

	h := EmpresaInventarioAlertasHandler(dbEmp)
	req := httptest.NewRequest(http.MethodGet, "/api/empresa/inventario/alertas?empresa_id=121&action=proactivas", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected proactivas status %d, got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var rows []dbpkg.InventarioAlertaOperativa
	if err := json.Unmarshal(rr.Body.Bytes(), &rows); err != nil {
		t.Fatalf("decode alertas proactivas: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 alerta proactiva, got %d", len(rows))
	}
	if rows[0].EstadoStock != "sobrestock" {
		t.Fatalf("expected estado_stock sobrestock, got %+v", rows[0])
	}
}
