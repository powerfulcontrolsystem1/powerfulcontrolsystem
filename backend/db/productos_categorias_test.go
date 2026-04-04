package db

import (
	"database/sql"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func openProductosTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "productos_test.db")
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

func TestCategoriasProductosCRUDAndSync(t *testing.T) {
	dbConn := openProductosTestDB(t)
	if err := EnsureEmpresaProductosSchema(dbConn); err != nil {
		t.Fatalf("ensure schema: %v", err)
	}

	categoriaID, err := CreateCategoriaProducto(dbConn, CategoriaProducto{
		EmpresaID: 1,
		Nombre:    "Lacteos",
		Codigo:    "CAT-LAC",
		Orden:     1,
	})
	if err != nil {
		t.Fatalf("create categoria: %v", err)
	}

	productoID, err := CreateProducto(dbConn, Producto{
		EmpresaID:   1,
		Nombre:      "Leche Entera",
		CategoriaID: categoriaID,
		Precio:      7200,
		Costo:       4800,
	}, 0, "TEST")
	if err != nil {
		t.Fatalf("create producto: %v", err)
	}

	p, err := GetProductoByID(dbConn, 1, productoID)
	if err != nil {
		t.Fatalf("get producto by id: %v", err)
	}
	if p.CategoriaID != categoriaID {
		t.Fatalf("expected categoria_id=%d, got %d", categoriaID, p.CategoriaID)
	}
	if p.Categoria != "Lacteos" {
		t.Fatalf("expected categoria Lacteos, got %q", p.Categoria)
	}

	if err := UpdateCategoriaProducto(dbConn, CategoriaProducto{
		ID:        categoriaID,
		EmpresaID: 1,
		Nombre:    "Lacteos Premium",
		Codigo:    "CAT-LAC",
		Orden:     2,
	}); err != nil {
		t.Fatalf("update categoria: %v", err)
	}

	pActualizado, err := GetProductoByID(dbConn, 1, productoID)
	if err != nil {
		t.Fatalf("get producto by id after update categoria: %v", err)
	}
	if pActualizado.Categoria != "Lacteos Premium" {
		t.Fatalf("expected categoria sincronizada Lacteos Premium, got %q", pActualizado.Categoria)
	}

	if err := DeleteCategoriaProducto(dbConn, 1, categoriaID); err == nil {
		t.Fatal("expected error deleting categoria associated to products")
	}
}

func TestEnsureSchemaMigratesLegacyCategorias(t *testing.T) {
	dbConn := openProductosTestDB(t)

	_, err := dbConn.Exec(`CREATE TABLE IF NOT EXISTS productos (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		empresa_id INTEGER NOT NULL,
		bodega_principal_id INTEGER,
		proveedor_principal_id INTEGER,
		sku TEXT,
		codigo_barras TEXT,
		nombre TEXT NOT NULL,
		descripcion TEXT,
		categoria TEXT,
		marca TEXT,
		unidad_medida TEXT DEFAULT 'unidad',
		costo REAL DEFAULT 0,
		precio REAL DEFAULT 0,
		impuesto_porcentaje REAL DEFAULT 0,
		stock_minimo REAL DEFAULT 0,
		stock_maximo REAL DEFAULT 0,
		imagen_url TEXT,
		fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
		fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
		usuario_creador TEXT,
		estado TEXT DEFAULT 'activo',
		observaciones TEXT
	);`)
	if err != nil {
		t.Fatalf("create legacy productos table: %v", err)
	}

	res, err := dbConn.Exec(`INSERT INTO productos (empresa_id, nombre, categoria) VALUES (9, 'Cafe molido', 'Abarrotes')`)
	if err != nil {
		t.Fatalf("insert legacy producto: %v", err)
	}
	productoID, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("last insert id: %v", err)
	}

	if err := EnsureEmpresaProductosSchema(dbConn); err != nil {
		t.Fatalf("ensure schema legacy migration: %v", err)
	}

	categorias, err := GetCategoriasProductoByEmpresa(dbConn, 9, true, "")
	if err != nil {
		t.Fatalf("list categorias migrated: %v", err)
	}
	if len(categorias) != 1 {
		t.Fatalf("expected 1 categoria migrated, got %d", len(categorias))
	}
	if categorias[0].Nombre != "Abarrotes" {
		t.Fatalf("expected migrated categoria Abarrotes, got %q", categorias[0].Nombre)
	}

	p, err := GetProductoByID(dbConn, 9, productoID)
	if err != nil {
		t.Fatalf("get producto migrated: %v", err)
	}
	if p.CategoriaID <= 0 {
		t.Fatalf("expected categoria_id backfilled, got %d", p.CategoriaID)
	}
}

func TestCreateAndUpdateProductoValidanStockMinMax(t *testing.T) {
	dbConn := openProductosTestDB(t)
	if err := EnsureEmpresaProductosSchema(dbConn); err != nil {
		t.Fatalf("ensure schema: %v", err)
	}

	if _, err := CreateProducto(dbConn, Producto{
		EmpresaID:    15,
		Nombre:       "Producto invalido",
		StockMinimo:  20,
		StockMaximo:  10,
		Precio:       100,
		Costo:        70,
		UnidadMedida: "unidad",
	}, 0, "TEST"); err == nil {
		t.Fatal("expected error creating producto when stock_minimo > stock_maximo")
	}

	productoID, err := CreateProducto(dbConn, Producto{
		EmpresaID:    15,
		Nombre:       "Producto valido",
		StockMinimo:  2,
		StockMaximo:  10,
		Precio:       100,
		Costo:        70,
		UnidadMedida: "unidad",
	}, 0, "TEST")
	if err != nil {
		t.Fatalf("create producto valido: %v", err)
	}

	p, err := GetProductoByID(dbConn, 15, productoID)
	if err != nil {
		t.Fatalf("get producto valido: %v", err)
	}
	p.StockMinimo = 12
	p.StockMaximo = 6

	if err := UpdateProducto(dbConn, *p, "", ""); err == nil {
		t.Fatal("expected error updating producto when stock_minimo > stock_maximo")
	}
}

func TestGetInventarioResumenByEmpresaCalculaIndicadores(t *testing.T) {
	dbConn := openProductosTestDB(t)
	if err := EnsureEmpresaProductosSchema(dbConn); err != nil {
		t.Fatalf("ensure schema: %v", err)
	}

	bodegaA, err := CreateBodega(dbConn, Bodega{EmpresaID: 61, Codigo: "BOD-A-61", Nombre: "A"})
	if err != nil {
		t.Fatalf("create bodega A: %v", err)
	}
	bodegaB, err := CreateBodega(dbConn, Bodega{EmpresaID: 61, Codigo: "BOD-B-61", Nombre: "B"})
	if err != nil {
		t.Fatalf("create bodega B: %v", err)
	}

	prod1, err := CreateProducto(dbConn, Producto{EmpresaID: 61, Nombre: "Producto 1", StockMinimo: 10, StockMaximo: 40}, 0, "TEST")
	if err != nil {
		t.Fatalf("create producto 1: %v", err)
	}
	prod2, err := CreateProducto(dbConn, Producto{EmpresaID: 61, Nombre: "Producto 2", StockMinimo: 5, StockMaximo: 20}, 0, "TEST")
	if err != nil {
		t.Fatalf("create producto 2: %v", err)
	}

	if _, err := dbConn.Exec(`INSERT INTO inventario_existencias (empresa_id, producto_id, bodega_id, cantidad, estado) VALUES (61, ?, ?, 0, 'activo')`, prod1, bodegaA); err != nil {
		t.Fatalf("insert existencia 1: %v", err)
	}
	if _, err := dbConn.Exec(`INSERT INTO inventario_existencias (empresa_id, producto_id, bodega_id, cantidad, estado) VALUES (61, ?, ?, 2, 'activo')`, prod2, bodegaB); err != nil {
		t.Fatalf("insert existencia 2: %v", err)
	}

	if _, err := dbConn.Exec(`INSERT INTO inventario_movimientos (
		empresa_id, producto_id, bodega_origen_id, bodega_destino_id, tipo, cantidad, costo_unitario, referencia, fecha_movimiento, estado
	) VALUES (61, ?, NULL, ?, 'entrada', 2, 10, 'E-1', '2026-04-02 10:00:00', 'activo')`, prod2, bodegaB); err != nil {
		t.Fatalf("insert movimiento entrada: %v", err)
	}
	if _, err := dbConn.Exec(`INSERT INTO inventario_movimientos (
		empresa_id, producto_id, bodega_origen_id, bodega_destino_id, tipo, cantidad, costo_unitario, referencia, fecha_movimiento, estado
	) VALUES (61, ?, ?, NULL, 'salida', 1, 10, 'S-1', '2026-04-03 11:00:00', 'activo')`, prod2, bodegaB); err != nil {
		t.Fatalf("insert movimiento salida: %v", err)
	}
	if _, err := dbConn.Exec(`INSERT INTO inventario_movimientos (
		empresa_id, producto_id, bodega_origen_id, bodega_destino_id, tipo, cantidad, costo_unitario, referencia, fecha_movimiento, estado
	) VALUES (61, ?, ?, ?, 'traslado', 1, 10, 'T-1', '2026-04-04 12:00:00', 'activo')`, prod2, bodegaA, bodegaB); err != nil {
		t.Fatalf("insert movimiento traslado: %v", err)
	}

	resumen, err := GetInventarioResumenByEmpresa(dbConn, 61, "2026-04-01", "2026-04-05")
	if err != nil {
		t.Fatalf("get inventario resumen: %v", err)
	}

	if resumen.TotalExistencias != 2 {
		t.Fatalf("expected total_existencias=2, got %.2f", resumen.TotalExistencias)
	}
	if resumen.ProductosConExistencia != 2 {
		t.Fatalf("expected productos_con_existencia=2, got %d", resumen.ProductosConExistencia)
	}
	if resumen.BodegasConStock != 2 {
		t.Fatalf("expected bodegas_con_stock=2, got %d", resumen.BodegasConStock)
	}
	if resumen.AlertasTotal != 2 || resumen.AlertasSinStock != 1 || resumen.AlertasBajoMinimo != 1 {
		t.Fatalf("unexpected alertas resumen: %+v", resumen)
	}
	if resumen.DeficitTotal != 13 {
		t.Fatalf("expected deficit_total=13, got %.2f", resumen.DeficitTotal)
	}
	if resumen.MovimientosTotal != 3 {
		t.Fatalf("expected movimientos_total=3, got %d", resumen.MovimientosTotal)
	}
	if resumen.MovimientosEntrada != 1 || resumen.MovimientosSalida != 1 || resumen.MovimientosTraslado != 1 {
		t.Fatalf("unexpected movimientos breakdown: %+v", resumen)
	}
}

func TestGetInventarioTendenciaByEmpresaDevuelveSerieDiaria(t *testing.T) {
	dbConn := openProductosTestDB(t)
	if err := EnsureEmpresaProductosSchema(dbConn); err != nil {
		t.Fatalf("ensure schema: %v", err)
	}

	bodegaA, err := CreateBodega(dbConn, Bodega{EmpresaID: 71, Codigo: "BOD-A-71", Nombre: "A"})
	if err != nil {
		t.Fatalf("create bodega A: %v", err)
	}
	bodegaB, err := CreateBodega(dbConn, Bodega{EmpresaID: 71, Codigo: "BOD-B-71", Nombre: "B"})
	if err != nil {
		t.Fatalf("create bodega B: %v", err)
	}

	prod, err := CreateProducto(dbConn, Producto{EmpresaID: 71, Nombre: "Producto tendencia", StockMinimo: 1, StockMaximo: 20}, 0, "TEST")
	if err != nil {
		t.Fatalf("create producto: %v", err)
	}

	if _, err := dbConn.Exec(`INSERT INTO inventario_movimientos (
		empresa_id, producto_id, bodega_origen_id, bodega_destino_id, tipo, cantidad, costo_unitario, referencia, fecha_movimiento, estado
	) VALUES (71, ?, NULL, ?, 'entrada', 5, 10, 'TEN-1', '2026-04-01 08:00:00', 'activo')`, prod, bodegaA); err != nil {
		t.Fatalf("insert movimiento 1: %v", err)
	}
	if _, err := dbConn.Exec(`INSERT INTO inventario_movimientos (
		empresa_id, producto_id, bodega_origen_id, bodega_destino_id, tipo, cantidad, costo_unitario, referencia, fecha_movimiento, estado
	) VALUES (71, ?, ?, NULL, 'salida', 2, 10, 'TEN-2', '2026-04-02 09:00:00', 'activo')`, prod, bodegaA); err != nil {
		t.Fatalf("insert movimiento 2: %v", err)
	}
	if _, err := dbConn.Exec(`INSERT INTO inventario_movimientos (
		empresa_id, producto_id, bodega_origen_id, bodega_destino_id, tipo, cantidad, costo_unitario, referencia, fecha_movimiento, estado
	) VALUES (71, ?, NULL, ?, 'ajuste_positivo', 1, 10, 'TEN-3', '2026-04-02 14:30:00', 'activo')`, prod, bodegaA); err != nil {
		t.Fatalf("insert movimiento 3: %v", err)
	}
	if _, err := dbConn.Exec(`INSERT INTO inventario_movimientos (
		empresa_id, producto_id, bodega_origen_id, bodega_destino_id, tipo, cantidad, costo_unitario, referencia, fecha_movimiento, estado
	) VALUES (71, ?, ?, ?, 'traslado', 3, 10, 'TEN-4', '2026-04-03 11:45:00', 'activo')`, prod, bodegaA, bodegaB); err != nil {
		t.Fatalf("insert movimiento 4: %v", err)
	}
	if _, err := dbConn.Exec(`INSERT INTO inventario_movimientos (
		empresa_id, producto_id, bodega_origen_id, bodega_destino_id, tipo, cantidad, costo_unitario, referencia, fecha_movimiento, estado
	) VALUES (71, ?, ?, NULL, 'salida', 7, 10, 'TEN-5', '2026-04-04 10:00:00', 'activo')`, prod, bodegaB); err != nil {
		t.Fatalf("insert movimiento fuera de bodega: %v", err)
	}

	serie, err := GetInventarioTendenciaByEmpresa(dbConn, 71, bodegaA, "2026-04-01", "2026-04-03", 0)
	if err != nil {
		t.Fatalf("get inventario tendencia: %v", err)
	}
	if len(serie) != 3 {
		t.Fatalf("expected 3 dias en tendencia, got %d", len(serie))
	}

	if serie[0].Fecha != "2026-04-01" || serie[0].Entradas != 5 || serie[0].Salidas != 0 || serie[0].Neto != 5 {
		t.Fatalf("unexpected dia 1: %+v", serie[0])
	}
	if serie[1].Fecha != "2026-04-02" || serie[1].Entradas != 1 || serie[1].Salidas != 2 || serie[1].Neto != -1 || serie[1].Eventos != 2 {
		t.Fatalf("unexpected dia 2: %+v", serie[1])
	}
	if serie[2].Fecha != "2026-04-03" || serie[2].Traslados != 3 || serie[2].Neto != 0 || serie[2].Eventos != 1 {
		t.Fatalf("unexpected dia 3: %+v", serie[2])
	}
}

func TestGetInventarioBalanceBodegasByEmpresaConsolidaMovimientos(t *testing.T) {
	dbConn := openProductosTestDB(t)
	if err := EnsureEmpresaProductosSchema(dbConn); err != nil {
		t.Fatalf("ensure schema: %v", err)
	}

	bodegaA, err := CreateBodega(dbConn, Bodega{EmpresaID: 81, Codigo: "BOD-A-81", Nombre: "A"})
	if err != nil {
		t.Fatalf("create bodega A: %v", err)
	}
	bodegaB, err := CreateBodega(dbConn, Bodega{EmpresaID: 81, Codigo: "BOD-B-81", Nombre: "B"})
	if err != nil {
		t.Fatalf("create bodega B: %v", err)
	}

	prod, err := CreateProducto(dbConn, Producto{EmpresaID: 81, Nombre: "Producto balance", StockMinimo: 1, StockMaximo: 30}, 0, "TEST")
	if err != nil {
		t.Fatalf("create producto: %v", err)
	}

	if _, err := dbConn.Exec(`INSERT INTO inventario_movimientos (
		empresa_id, producto_id, bodega_origen_id, bodega_destino_id, tipo, cantidad, costo_unitario, referencia, fecha_movimiento, estado
	) VALUES (81, ?, NULL, ?, 'entrada', 5, 10, 'BAL-1', '2026-04-01 08:00:00', 'activo')`, prod, bodegaA); err != nil {
		t.Fatalf("insert movimiento 1: %v", err)
	}
	if _, err := dbConn.Exec(`INSERT INTO inventario_movimientos (
		empresa_id, producto_id, bodega_origen_id, bodega_destino_id, tipo, cantidad, costo_unitario, referencia, fecha_movimiento, estado
	) VALUES (81, ?, ?, NULL, 'salida', 2, 10, 'BAL-2', '2026-04-02 09:00:00', 'activo')`, prod, bodegaA); err != nil {
		t.Fatalf("insert movimiento 2: %v", err)
	}
	if _, err := dbConn.Exec(`INSERT INTO inventario_movimientos (
		empresa_id, producto_id, bodega_origen_id, bodega_destino_id, tipo, cantidad, costo_unitario, referencia, fecha_movimiento, estado
	) VALUES (81, ?, ?, ?, 'traslado', 1, 10, 'BAL-3', '2026-04-03 10:00:00', 'activo')`, prod, bodegaA, bodegaB); err != nil {
		t.Fatalf("insert movimiento 3: %v", err)
	}
	if _, err := dbConn.Exec(`INSERT INTO inventario_movimientos (
		empresa_id, producto_id, bodega_origen_id, bodega_destino_id, tipo, cantidad, costo_unitario, referencia, fecha_movimiento, estado
	) VALUES (81, ?, NULL, ?, 'ajuste_positivo', 2, 10, 'BAL-4', '2026-04-03 13:20:00', 'activo')`, prod, bodegaB); err != nil {
		t.Fatalf("insert movimiento 4: %v", err)
	}
	if _, err := dbConn.Exec(`INSERT INTO inventario_movimientos (
		empresa_id, producto_id, bodega_origen_id, bodega_destino_id, tipo, cantidad, costo_unitario, referencia, fecha_movimiento, estado
	) VALUES (81, ?, ?, NULL, 'salida', 9, 10, 'BAL-OLD', '2026-03-10 10:00:00', 'activo')`, prod, bodegaA); err != nil {
		t.Fatalf("insert movimiento fuera de rango: %v", err)
	}

	balance, err := GetInventarioBalanceBodegasByEmpresa(dbConn, 81, 0, "2026-04-01", "2026-04-03", 0)
	if err != nil {
		t.Fatalf("get inventario balance bodegas: %v", err)
	}
	if len(balance) != 2 {
		t.Fatalf("expected 2 bodegas en balance, got %d", len(balance))
	}

	var rowA, rowB *InventarioBalanceBodega
	for i := range balance {
		if balance[i].BodegaID == bodegaA {
			rowA = &balance[i]
		}
		if balance[i].BodegaID == bodegaB {
			rowB = &balance[i]
		}
	}
	if rowA == nil || rowB == nil {
		t.Fatalf("expected rows for bodegas A and B, got %+v", balance)
	}

	if rowA.Entradas != 5 || rowA.Salidas != 2 || rowA.TrasladosSalida != 1 || rowA.Neto != 2 {
		t.Fatalf("unexpected balance bodega A: %+v", rowA)
	}
	if rowB.Entradas != 2 || rowB.TrasladosEntrada != 1 || rowB.Neto != 3 {
		t.Fatalf("unexpected balance bodega B: %+v", rowB)
	}
}
