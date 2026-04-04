package handlers

import (
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
