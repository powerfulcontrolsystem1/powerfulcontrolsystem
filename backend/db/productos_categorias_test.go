package db

import (
	"database/sql"
	"errors"
	"path/filepath"
	"strings"
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

func TestCombosProductoCRUDConReceta(t *testing.T) {
	dbConn := openProductosTestDB(t)
	if err := EnsureEmpresaProductosSchema(dbConn); err != nil {
		t.Fatalf("ensure schema: %v", err)
	}

	bodegaID, err := CreateBodega(dbConn, Bodega{EmpresaID: 82, Codigo: "BOD-CMB-82", Nombre: "Bodega combos"})
	if err != nil {
		t.Fatalf("create bodega: %v", err)
	}

	prodA, err := CreateProducto(dbConn, Producto{EmpresaID: 82, BodegaPrincipalID: bodegaID, Nombre: "Insumo A", UnidadMedida: "unidad", Precio: 100, Costo: 40}, 20, "TEST")
	if err != nil {
		t.Fatalf("create producto A: %v", err)
	}
	prodB, err := CreateProducto(dbConn, Producto{EmpresaID: 82, BodegaPrincipalID: bodegaID, Nombre: "Insumo B", UnidadMedida: "unidad", Precio: 200, Costo: 80}, 15, "TEST")
	if err != nil {
		t.Fatalf("create producto B: %v", err)
	}

	comboID, err := CreateComboProducto(dbConn, ComboProducto{
		EmpresaID:          82,
		Codigo:             "CMB-82-01",
		Nombre:             "Combo 82",
		Precio:             5000,
		ImpuestoPorcentaje: 19,
		Estado:             "activo",
	}, []ComboProductoDetalle{
		{ProductoID: prodA, Cantidad: 2, UnidadMedida: "unidad"},
		{ProductoID: prodB, Cantidad: 1, UnidadMedida: "unidad"},
	})
	if err != nil {
		t.Fatalf("create combo: %v", err)
	}

	rows, err := GetCombosProductosByEmpresa(dbConn, 82, "", "", true, 50, 0)
	if err != nil {
		t.Fatalf("list combos: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 combo, got %d", len(rows))
	}
	if rows[0].IngredientesCount != 2 {
		t.Fatalf("expected ingredientes_count=2, got %d", rows[0].IngredientesCount)
	}

	combo, err := GetComboProductoByID(dbConn, 82, comboID)
	if err != nil {
		t.Fatalf("get combo by id: %v", err)
	}
	if combo.Nombre != "Combo 82" {
		t.Fatalf("expected combo name Combo 82, got %q", combo.Nombre)
	}
	if combo.RecetaVersion != 1 {
		t.Fatalf("expected receta_version=1 after create, got %d", combo.RecetaVersion)
	}
	if combo.CostoTeorico != 160 {
		t.Fatalf("expected costo_teorico=160, got %.2f", combo.CostoTeorico)
	}
	if combo.CostoReal != 160 {
		t.Fatalf("expected costo_real=160, got %.2f", combo.CostoReal)
	}
	if len(combo.Ingredientes) != 2 {
		t.Fatalf("expected 2 ingredientes, got %d", len(combo.Ingredientes))
	}

	if err := UpdateComboProducto(dbConn, ComboProducto{
		ID:                 comboID,
		EmpresaID:          82,
		Codigo:             "CMB-82-01",
		Nombre:             "Combo 82 Ajustado",
		Precio:             6200,
		ImpuestoPorcentaje: 19,
		Estado:             "activo",
	}, []ComboProductoDetalle{
		{ProductoID: prodA, Cantidad: 3, UnidadMedida: "unidad"},
	}); err != nil {
		t.Fatalf("update combo: %v", err)
	}

	comboActualizado, err := GetComboProductoByID(dbConn, 82, comboID)
	if err != nil {
		t.Fatalf("get combo updated: %v", err)
	}
	if comboActualizado.Nombre != "Combo 82 Ajustado" {
		t.Fatalf("expected updated combo name, got %q", comboActualizado.Nombre)
	}
	if comboActualizado.Precio != 6200 {
		t.Fatalf("expected updated combo price 6200, got %.2f", comboActualizado.Precio)
	}
	if comboActualizado.RecetaVersion != 2 {
		t.Fatalf("expected receta_version=2 after recipe update, got %d", comboActualizado.RecetaVersion)
	}
	if comboActualizado.CostoTeorico != 120 {
		t.Fatalf("expected updated costo_teorico=120, got %.2f", comboActualizado.CostoTeorico)
	}
	if comboActualizado.CostoReal != 120 {
		t.Fatalf("expected updated costo_real=120, got %.2f", comboActualizado.CostoReal)
	}
	if len(comboActualizado.Ingredientes) != 1 {
		t.Fatalf("expected 1 ingrediente after update, got %d", len(comboActualizado.Ingredientes))
	}
	if comboActualizado.Ingredientes[0].Cantidad != 3 {
		t.Fatalf("expected ingrediente cantidad 3, got %.2f", comboActualizado.Ingredientes[0].Cantidad)
	}

	var versiones int64
	if err := dbConn.QueryRow(`SELECT COUNT(1) FROM combos_productos_versiones WHERE empresa_id = ? AND combo_id = ?`, 82, comboID).Scan(&versiones); err != nil {
		t.Fatalf("count combo versions: %v", err)
	}
	if versiones < 2 {
		t.Fatalf("expected at least 2 combo recipe versions, got %d", versiones)
	}

	if err := SetComboProductoEstado(dbConn, 82, comboID, "inactivo"); err != nil {
		t.Fatalf("set combo estado: %v", err)
	}

	if err := DeleteComboProducto(dbConn, 82, comboID); err != nil {
		t.Fatalf("delete combo: %v", err)
	}

	if _, err := GetComboProductoByID(dbConn, 82, comboID); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows after combo delete, got %v", err)
	}
}

func TestComboProductoValidaCostoTeoricoVsReal(t *testing.T) {
	dbConn := openProductosTestDB(t)
	if err := EnsureEmpresaProductosSchema(dbConn); err != nil {
		t.Fatalf("ensure schema: %v", err)
	}

	bodegaID, err := CreateBodega(dbConn, Bodega{EmpresaID: 91, Codigo: "BOD-CST-91", Nombre: "Bodega costo"})
	if err != nil {
		t.Fatalf("create bodega: %v", err)
	}

	productoID, err := CreateProducto(dbConn, Producto{
		EmpresaID:         91,
		BodegaPrincipalID: bodegaID,
		Nombre:            "Insumo costo mixto",
		UnidadMedida:      "unidad",
		Precio:            120,
		Costo:             10,
		Estado:            "activo",
	}, 5, "TEST")
	if err != nil {
		t.Fatalf("create producto: %v", err)
	}

	if _, err := dbConn.Exec(`DELETE FROM inventario_costos_lotes WHERE empresa_id = ? AND producto_id = ?`, 91, productoID); err != nil {
		t.Fatalf("clean lotes: %v", err)
	}
	if _, err := dbConn.Exec(`INSERT INTO inventario_costos_lotes (
		empresa_id,
		producto_id,
		bodega_id,
		cantidad_disponible,
		costo_unitario,
		referencia,
		usuario_creador,
		estado,
		observaciones
	) VALUES (?, ?, ?, ?, ?, ?, ?, 'activo', ?)`,
		91,
		productoID,
		bodegaID,
		5,
		60,
		"TEST-REAL-COST",
		"TEST",
		"costo real superior para validar variacion",
	); err != nil {
		t.Fatalf("insert lotes high cost: %v", err)
	}

	_, err = CreateComboProducto(dbConn, ComboProducto{
		EmpresaID:          91,
		Codigo:             "CMB-CST-91",
		Nombre:             "Combo validacion costos",
		Precio:             200,
		ImpuestoPorcentaje: 0,
		Estado:             "activo",
	}, []ComboProductoDetalle{{ProductoID: productoID, Cantidad: 1, UnidadMedida: "unidad"}})
	if err == nil {
		t.Fatal("expected costo teorico vs real validation error, got nil")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "variacion") {
		t.Fatalf("expected variacion error, got %v", err)
	}
}

func TestProveedorCRUDIncluyeCatalogoPreciosYCondiciones(t *testing.T) {
	dbConn := openProductosTestDB(t)
	if err := EnsureEmpresaProductosSchema(dbConn); err != nil {
		t.Fatalf("ensure schema: %v", err)
	}

	proveedorID, err := CreateProveedor(dbConn, Proveedor{
		EmpresaID:             62,
		Nombre:                "Proveedor Comercial",
		Contacto:              "Camila R",
		Telefono:              "3000001111",
		Email:                 "compras@proveedor-comercial.com",
		CatalogoReferencia:    "CAT-2026-Q2",
		PrecioBaseReferencial: 12500.75,
		DescuentoPorcentaje:   7.5,
		PlazoPagoDias:         30,
		CondicionEntrega:      "Entrega 24h en bodega principal",
		Observaciones:         "Precio sujeto a volumen",
	})
	if err != nil {
		t.Fatalf("create proveedor comercial: %v", err)
	}

	rows, err := GetProveedoresByEmpresa(dbConn, 62, true)
	if err != nil {
		t.Fatalf("list proveedores: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 proveedor, got %d", len(rows))
	}
	if rows[0].CatalogoReferencia != "CAT-2026-Q2" {
		t.Fatalf("expected catalogo referencia CAT-2026-Q2, got %q", rows[0].CatalogoReferencia)
	}
	if rows[0].PrecioBaseReferencial != 12500.75 {
		t.Fatalf("expected precio_base_referencial 12500.75, got %.2f", rows[0].PrecioBaseReferencial)
	}
	if rows[0].DescuentoPorcentaje != 7.5 {
		t.Fatalf("expected descuento_porcentaje 7.5, got %.2f", rows[0].DescuentoPorcentaje)
	}
	if rows[0].PlazoPagoDias != 30 {
		t.Fatalf("expected plazo_pago_dias 30, got %d", rows[0].PlazoPagoDias)
	}

	if err := UpdateProveedor(dbConn, Proveedor{
		ID:                    proveedorID,
		EmpresaID:             62,
		Nombre:                "Proveedor Comercial",
		PrecioBaseReferencial: 13200,
		DescuentoPorcentaje:   10,
		PlazoPagoDias:         45,
		CondicionEntrega:      "Entrega 48h",
		CatalogoReferencia:    "CAT-2026-Q3",
		Observaciones:         "Ajuste trimestral",
	}); err != nil {
		t.Fatalf("update proveedor comercial: %v", err)
	}

	rows, err = GetProveedoresByEmpresa(dbConn, 62, true)
	if err != nil {
		t.Fatalf("list proveedores after update: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 proveedor after update, got %d", len(rows))
	}
	if rows[0].PrecioBaseReferencial != 13200 {
		t.Fatalf("expected precio_base_referencial 13200, got %.2f", rows[0].PrecioBaseReferencial)
	}
	if rows[0].DescuentoPorcentaje != 10 {
		t.Fatalf("expected descuento_porcentaje 10, got %.2f", rows[0].DescuentoPorcentaje)
	}
	if rows[0].PlazoPagoDias != 45 {
		t.Fatalf("expected plazo_pago_dias 45, got %d", rows[0].PlazoPagoDias)
	}
	if rows[0].CatalogoReferencia != "CAT-2026-Q3" {
		t.Fatalf("expected catalogo referencia CAT-2026-Q3, got %q", rows[0].CatalogoReferencia)
	}

	if _, err := CreateProveedor(dbConn, Proveedor{EmpresaID: 62, Nombre: "Proveedor invalido", PrecioBaseReferencial: -10}); err == nil {
		t.Fatalf("expected error for precio_base_referencial negativo")
	}
	if err := UpdateProveedor(dbConn, Proveedor{ID: proveedorID, EmpresaID: 62, Nombre: "Proveedor Comercial", DescuentoPorcentaje: 150}); err == nil {
		t.Fatalf("expected error for descuento_porcentaje fuera de rango")
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

func TestGetInventarioProyeccionQuiebreByEmpresaPriorizaRiesgo(t *testing.T) {
	dbConn := openProductosTestDB(t)
	if err := EnsureEmpresaProductosSchema(dbConn); err != nil {
		t.Fatalf("ensure schema: %v", err)
	}

	bodegaA, err := CreateBodega(dbConn, Bodega{EmpresaID: 91, Codigo: "BOD-A-91", Nombre: "A"})
	if err != nil {
		t.Fatalf("create bodega A: %v", err)
	}

	prodRiesgo, err := CreateProducto(dbConn, Producto{EmpresaID: 91, Nombre: "Producto Riesgo", StockMinimo: 5, StockMaximo: 20}, 0, "TEST")
	if err != nil {
		t.Fatalf("create producto riesgo: %v", err)
	}
	prodEstable, err := CreateProducto(dbConn, Producto{EmpresaID: 91, Nombre: "Producto Estable", StockMinimo: 3, StockMaximo: 40}, 0, "TEST")
	if err != nil {
		t.Fatalf("create producto estable: %v", err)
	}

	if _, err := dbConn.Exec(`INSERT INTO inventario_existencias (empresa_id, producto_id, bodega_id, cantidad, estado) VALUES (91, ?, ?, 2, 'activo')`, prodRiesgo, bodegaA); err != nil {
		t.Fatalf("insert existencia riesgo: %v", err)
	}
	if _, err := dbConn.Exec(`INSERT INTO inventario_existencias (empresa_id, producto_id, bodega_id, cantidad, estado) VALUES (91, ?, ?, 30, 'activo')`, prodEstable, bodegaA); err != nil {
		t.Fatalf("insert existencia estable: %v", err)
	}

	if _, err := dbConn.Exec(`INSERT INTO inventario_movimientos (
		empresa_id, producto_id, bodega_origen_id, bodega_destino_id, tipo, cantidad, costo_unitario, referencia, fecha_movimiento, estado
	) VALUES (91, ?, ?, NULL, 'salida', 10, 10, 'PROY-R-1', datetime('now','-1 day'), 'activo')`, prodRiesgo, bodegaA); err != nil {
		t.Fatalf("insert salida riesgo: %v", err)
	}
	if _, err := dbConn.Exec(`INSERT INTO inventario_movimientos (
		empresa_id, producto_id, bodega_origen_id, bodega_destino_id, tipo, cantidad, costo_unitario, referencia, fecha_movimiento, estado
	) VALUES (91, ?, ?, NULL, 'salida', 5, 10, 'PROY-E-1', datetime('now','-1 day'), 'activo')`, prodEstable, bodegaA); err != nil {
		t.Fatalf("insert salida estable: %v", err)
	}

	rows, err := GetInventarioProyeccionQuiebreByEmpresa(dbConn, 91, bodegaA, 5, 20, 0)
	if err != nil {
		t.Fatalf("get inventario proyeccion quiebre: %v", err)
	}
	if len(rows) < 2 {
		t.Fatalf("expected at least 2 rows in proyeccion, got %d", len(rows))
	}

	if rows[0].ProductoID != prodRiesgo {
		t.Fatalf("expected first row as riesgo product %d, got %d", prodRiesgo, rows[0].ProductoID)
	}
	if rows[0].EstadoProyeccion != "quiebre_inminente" {
		t.Fatalf("expected estado quiebre_inminente, got %q", rows[0].EstadoProyeccion)
	}
	if rows[0].SugeridoReposicion <= 0 {
		t.Fatalf("expected sugerido_reposicion > 0, got %.2f", rows[0].SugeridoReposicion)
	}
}

func TestGetInventarioPlanReposicionByEmpresaConsolidaProveedorYCosto(t *testing.T) {
	dbConn := openProductosTestDB(t)
	if err := EnsureEmpresaProductosSchema(dbConn); err != nil {
		t.Fatalf("ensure schema: %v", err)
	}

	bodegaA, err := CreateBodega(dbConn, Bodega{EmpresaID: 95, Codigo: "BOD-A-95", Nombre: "A"})
	if err != nil {
		t.Fatalf("create bodega A: %v", err)
	}

	proveedorID, err := CreateProveedor(dbConn, Proveedor{EmpresaID: 95, Nombre: "Proveedor Central"})
	if err != nil {
		t.Fatalf("create proveedor: %v", err)
	}

	prodRiesgo, err := CreateProducto(dbConn, Producto{EmpresaID: 95, Nombre: "Producto Plan", ProveedorPrincipalID: proveedorID, Costo: 12.5, StockMinimo: 4, StockMaximo: 18}, 0, "TEST")
	if err != nil {
		t.Fatalf("create producto plan: %v", err)
	}
	prodEstable, err := CreateProducto(dbConn, Producto{EmpresaID: 95, Nombre: "Producto Estable", ProveedorPrincipalID: proveedorID, Costo: 10, StockMinimo: 2, StockMaximo: 40}, 0, "TEST")
	if err != nil {
		t.Fatalf("create producto estable: %v", err)
	}

	if _, err := dbConn.Exec(`INSERT INTO inventario_existencias (empresa_id, producto_id, bodega_id, cantidad, estado) VALUES (95, ?, ?, 3, 'activo')`, prodRiesgo, bodegaA); err != nil {
		t.Fatalf("insert existencia plan: %v", err)
	}
	if _, err := dbConn.Exec(`INSERT INTO inventario_existencias (empresa_id, producto_id, bodega_id, cantidad, estado) VALUES (95, ?, ?, 25, 'activo')`, prodEstable, bodegaA); err != nil {
		t.Fatalf("insert existencia estable: %v", err)
	}

	if _, err := dbConn.Exec(`INSERT INTO inventario_movimientos (
		empresa_id, producto_id, bodega_origen_id, bodega_destino_id, tipo, cantidad, costo_unitario, referencia, fecha_movimiento, estado
	) VALUES (95, ?, ?, NULL, 'salida', 8, 12.5, 'PLAN-R-1', datetime('now','-1 day'), 'activo')`, prodRiesgo, bodegaA); err != nil {
		t.Fatalf("insert salida riesgo: %v", err)
	}

	rows, err := GetInventarioPlanReposicionByEmpresa(dbConn, 95, bodegaA, 7, true, 20, 0)
	if err != nil {
		t.Fatalf("get inventario plan reposicion: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row in plan reposicion, got %d", len(rows))
	}

	if rows[0].ProductoID != prodRiesgo {
		t.Fatalf("expected producto riesgo %d, got %d", prodRiesgo, rows[0].ProductoID)
	}
	if rows[0].ProveedorID != proveedorID {
		t.Fatalf("expected proveedor_id %d, got %d", proveedorID, rows[0].ProveedorID)
	}
	if rows[0].CostoUnitarioRef != 12.5 {
		t.Fatalf("expected costo_unitario_ref 12.5, got %.2f", rows[0].CostoUnitarioRef)
	}
	if rows[0].CostoEstimado <= 0 {
		t.Fatalf("expected costo_estimado > 0, got %.2f", rows[0].CostoEstimado)
	}
}

func TestGetInventarioPlanReposicionResumenByEmpresaAgrupaProveedor(t *testing.T) {
	dbConn := openProductosTestDB(t)
	if err := EnsureEmpresaProductosSchema(dbConn); err != nil {
		t.Fatalf("ensure schema: %v", err)
	}

	bodegaA, err := CreateBodega(dbConn, Bodega{EmpresaID: 96, Codigo: "BOD-A-96", Nombre: "A"})
	if err != nil {
		t.Fatalf("create bodega A: %v", err)
	}

	proveedorID, err := CreateProveedor(dbConn, Proveedor{EmpresaID: 96, Nombre: "Proveedor Resumen"})
	if err != nil {
		t.Fatalf("create proveedor: %v", err)
	}

	p1, err := CreateProducto(dbConn, Producto{EmpresaID: 96, Nombre: "Producto 1", ProveedorPrincipalID: proveedorID, Costo: 11, StockMinimo: 3, StockMaximo: 15}, 0, "TEST")
	if err != nil {
		t.Fatalf("create producto 1: %v", err)
	}
	p2, err := CreateProducto(dbConn, Producto{EmpresaID: 96, Nombre: "Producto 2", ProveedorPrincipalID: proveedorID, Costo: 8, StockMinimo: 2, StockMaximo: 12}, 0, "TEST")
	if err != nil {
		t.Fatalf("create producto 2: %v", err)
	}

	if _, err := dbConn.Exec(`INSERT INTO inventario_existencias (empresa_id, producto_id, bodega_id, cantidad, estado) VALUES (96, ?, ?, 2, 'activo')`, p1, bodegaA); err != nil {
		t.Fatalf("insert existencia p1: %v", err)
	}
	if _, err := dbConn.Exec(`INSERT INTO inventario_existencias (empresa_id, producto_id, bodega_id, cantidad, estado) VALUES (96, ?, ?, 1.5, 'activo')`, p2, bodegaA); err != nil {
		t.Fatalf("insert existencia p2: %v", err)
	}

	if _, err := dbConn.Exec(`INSERT INTO inventario_movimientos (
		empresa_id, producto_id, bodega_origen_id, bodega_destino_id, tipo, cantidad, costo_unitario, referencia, fecha_movimiento, estado
	) VALUES (96, ?, ?, NULL, 'salida', 6, 11, 'RES-1', datetime('now','-1 day'), 'activo')`, p1, bodegaA); err != nil {
		t.Fatalf("insert salida p1: %v", err)
	}
	if _, err := dbConn.Exec(`INSERT INTO inventario_movimientos (
		empresa_id, producto_id, bodega_origen_id, bodega_destino_id, tipo, cantidad, costo_unitario, referencia, fecha_movimiento, estado
	) VALUES (96, ?, ?, NULL, 'salida', 4, 8, 'RES-2', datetime('now','-1 day'), 'activo')`, p2, bodegaA); err != nil {
		t.Fatalf("insert salida p2: %v", err)
	}

	rows, err := GetInventarioPlanReposicionResumenByEmpresa(dbConn, 96, bodegaA, 7, true, 20, 0)
	if err != nil {
		t.Fatalf("get inventario plan reposicion resumen: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 provider row in resumen, got %d", len(rows))
	}
	if rows[0].ProveedorID != proveedorID {
		t.Fatalf("expected proveedor_id=%d, got %d", proveedorID, rows[0].ProveedorID)
	}
	if rows[0].Items != 2 || rows[0].ProductosUnicos != 2 {
		t.Fatalf("expected 2 items/productos, got %+v", rows[0])
	}
	if rows[0].CostoTotal <= 0 {
		t.Fatalf("expected costo_total > 0, got %+v", rows[0])
	}
}

func TestGetInventarioPlanReposicionBorradorByEmpresaProveedor(t *testing.T) {
	dbConn := openProductosTestDB(t)
	if err := EnsureEmpresaProductosSchema(dbConn); err != nil {
		t.Fatalf("ensure schema: %v", err)
	}

	bodegaA, err := CreateBodega(dbConn, Bodega{EmpresaID: 97, Codigo: "BOD-A-97", Nombre: "A"})
	if err != nil {
		t.Fatalf("create bodega A: %v", err)
	}

	proveedorID, err := CreateProveedor(dbConn, Proveedor{EmpresaID: 97, Nombre: "Proveedor Borrador"})
	if err != nil {
		t.Fatalf("create proveedor: %v", err)
	}

	p1, err := CreateProducto(dbConn, Producto{EmpresaID: 97, Nombre: "Producto B1", ProveedorPrincipalID: proveedorID, Costo: 13, StockMinimo: 4, StockMaximo: 18}, 0, "TEST")
	if err != nil {
		t.Fatalf("create producto 1: %v", err)
	}
	p2, err := CreateProducto(dbConn, Producto{EmpresaID: 97, Nombre: "Producto B2", ProveedorPrincipalID: proveedorID, Costo: 9, StockMinimo: 2, StockMaximo: 12}, 0, "TEST")
	if err != nil {
		t.Fatalf("create producto 2: %v", err)
	}

	if _, err := dbConn.Exec(`INSERT INTO inventario_existencias (empresa_id, producto_id, bodega_id, cantidad, estado) VALUES (97, ?, ?, 1, 'activo')`, p1, bodegaA); err != nil {
		t.Fatalf("insert existencia p1: %v", err)
	}
	if _, err := dbConn.Exec(`INSERT INTO inventario_existencias (empresa_id, producto_id, bodega_id, cantidad, estado) VALUES (97, ?, ?, 0.5, 'activo')`, p2, bodegaA); err != nil {
		t.Fatalf("insert existencia p2: %v", err)
	}

	if _, err := dbConn.Exec(`INSERT INTO inventario_movimientos (
		empresa_id, producto_id, bodega_origen_id, bodega_destino_id, tipo, cantidad, costo_unitario, referencia, fecha_movimiento, estado
	) VALUES (97, ?, ?, NULL, 'salida', 7, 13, 'BORR-1', datetime('now','-1 day'), 'activo')`, p1, bodegaA); err != nil {
		t.Fatalf("insert salida p1: %v", err)
	}
	if _, err := dbConn.Exec(`INSERT INTO inventario_movimientos (
		empresa_id, producto_id, bodega_origen_id, bodega_destino_id, tipo, cantidad, costo_unitario, referencia, fecha_movimiento, estado
	) VALUES (97, ?, ?, NULL, 'salida', 4, 9, 'BORR-2', datetime('now','-1 day'), 'activo')`, p2, bodegaA); err != nil {
		t.Fatalf("insert salida p2: %v", err)
	}

	borrador, err := GetInventarioPlanReposicionBorradorByEmpresa(dbConn, 97, proveedorID, bodegaA, 7, true)
	if err != nil {
		t.Fatalf("get inventario plan reposicion borrador: %v", err)
	}

	if borrador.ProveedorID != proveedorID {
		t.Fatalf("expected proveedor_id=%d, got %d", proveedorID, borrador.ProveedorID)
	}
	if borrador.TotalItems != 2 {
		t.Fatalf("expected total_items=2, got %+v", borrador)
	}
	if borrador.ProductosUnicos != 2 {
		t.Fatalf("expected productos_unicos=2, got %+v", borrador)
	}
	if len(borrador.Items) != 2 {
		t.Fatalf("expected 2 detail rows, got %d", len(borrador.Items))
	}
	if borrador.CostoTotal <= 0 {
		t.Fatalf("expected costo_total > 0, got %+v", borrador)
	}
	if len(borrador.CodigoBorrador) < 8 || borrador.CodigoBorrador[:8] != "BORR-OC-" {
		t.Fatalf("expected codigo borrador with BORR-OC- prefix, got %q", borrador.CodigoBorrador)
	}

	if _, err := GetInventarioPlanReposicionBorradorByEmpresa(dbConn, 97, 0, bodegaA, 7, true); err == nil {
		t.Fatalf("expected error for proveedor_id invalido")
	}
}

func TestEmitirOrdenCompraDesdePlanReposicionBorradorPersistDoc(t *testing.T) {
	dbConn := openProductosTestDB(t)
	if err := EnsureEmpresaProductosSchema(dbConn); err != nil {
		t.Fatalf("ensure productos schema: %v", err)
	}
	if err := EnsureEmpresaDocumentosTransaccionalesSchema(dbConn); err != nil {
		t.Fatalf("ensure documentos transaccionales schema: %v", err)
	}

	bodegaA, err := CreateBodega(dbConn, Bodega{EmpresaID: 98, Codigo: "BOD-A-98", Nombre: "A"})
	if err != nil {
		t.Fatalf("create bodega A: %v", err)
	}

	proveedorID, err := CreateProveedor(dbConn, Proveedor{EmpresaID: 98, Nombre: "Proveedor Emision"})
	if err != nil {
		t.Fatalf("create proveedor: %v", err)
	}

	productoID, err := CreateProducto(dbConn, Producto{EmpresaID: 98, Nombre: "Producto Emitir", ProveedorPrincipalID: proveedorID, Costo: 15, StockMinimo: 4, StockMaximo: 18}, 0, "TEST")
	if err != nil {
		t.Fatalf("create producto: %v", err)
	}

	if _, err := dbConn.Exec(`INSERT INTO inventario_existencias (empresa_id, producto_id, bodega_id, cantidad, estado) VALUES (98, ?, ?, 1, 'activo')`, productoID, bodegaA); err != nil {
		t.Fatalf("insert existencia: %v", err)
	}
	if _, err := dbConn.Exec(`INSERT INTO inventario_movimientos (
		empresa_id, producto_id, bodega_origen_id, bodega_destino_id, tipo, cantidad, costo_unitario, referencia, fecha_movimiento, estado
	) VALUES (98, ?, ?, NULL, 'salida', 6, 15, 'EMI-1', datetime('now','-1 day'), 'activo')`, productoID, bodegaA); err != nil {
		t.Fatalf("insert salida: %v", err)
	}

	resultado, err := EmitirOrdenCompraDesdePlanReposicionBorrador(dbConn, 98, proveedorID, bodegaA, 7, true, "", "2026-04", "COP", "tester@local", "emision desde fase 11")
	if err != nil {
		t.Fatalf("emitir orden desde borrador: %v", err)
	}

	if resultado.EstadoNuevo != "emitida" {
		t.Fatalf("expected estado_nuevo emitida, got %+v", resultado)
	}
	if resultado.Evento != "orden_compra_emitida" {
		t.Fatalf("expected evento orden_compra_emitida, got %+v", resultado)
	}
	if len(resultado.DocumentoCodigo) < 3 || resultado.DocumentoCodigo[:3] != "OC-" {
		t.Fatalf("expected documento_codigo with OC- prefix, got %+v", resultado)
	}
	if resultado.EntidadID <= 0 {
		t.Fatalf("expected entidad_id > 0, got %+v", resultado)
	}

	doc, err := GetEmpresaDocumentoCompraByCodigo(dbConn, 98, "orden_compra", resultado.DocumentoCodigo)
	if err != nil {
		t.Fatalf("get compra documento emitido: %v", err)
	}
	if doc.EstadoDocumento != "emitida" {
		t.Fatalf("expected documento emitida, got %+v", doc)
	}
	if doc.MontoTotal <= 0 {
		t.Fatalf("expected monto_total > 0, got %+v", doc)
	}

	if _, err := EmitirOrdenCompraDesdePlanReposicionBorrador(dbConn, 98, 999999, bodegaA, 7, true, "", "", "", "", ""); err == nil {
		t.Fatalf("expected error when provider has no suggested items")
	}
}

func TestActualizarEstadoOrdenCompraDesdeReposicionCiclo(t *testing.T) {
	dbConn := openProductosTestDB(t)
	if err := EnsureEmpresaProductosSchema(dbConn); err != nil {
		t.Fatalf("ensure productos schema: %v", err)
	}
	if err := EnsureEmpresaDocumentosTransaccionalesSchema(dbConn); err != nil {
		t.Fatalf("ensure documentos transaccionales schema: %v", err)
	}

	bodegaA, err := CreateBodega(dbConn, Bodega{EmpresaID: 99, Codigo: "BOD-A-99", Nombre: "A"})
	if err != nil {
		t.Fatalf("create bodega A: %v", err)
	}

	proveedorID, err := CreateProveedor(dbConn, Proveedor{EmpresaID: 99, Nombre: "Proveedor F12"})
	if err != nil {
		t.Fatalf("create proveedor: %v", err)
	}

	productoID, err := CreateProducto(dbConn, Producto{EmpresaID: 99, Nombre: "Producto F12", ProveedorPrincipalID: proveedorID, Costo: 10, StockMinimo: 4, StockMaximo: 18}, 0, "TEST")
	if err != nil {
		t.Fatalf("create producto: %v", err)
	}

	if _, err := dbConn.Exec(`INSERT INTO inventario_existencias (empresa_id, producto_id, bodega_id, cantidad, estado) VALUES (99, ?, ?, 1, 'activo')`, productoID, bodegaA); err != nil {
		t.Fatalf("insert existencia: %v", err)
	}
	if _, err := dbConn.Exec(`INSERT INTO inventario_movimientos (
		empresa_id, producto_id, bodega_origen_id, bodega_destino_id, tipo, cantidad, costo_unitario, referencia, fecha_movimiento, estado
	) VALUES (99, ?, ?, NULL, 'salida', 6, 10, 'F12-1', datetime('now','-1 day'), 'activo')`, productoID, bodegaA); err != nil {
		t.Fatalf("insert salida: %v", err)
	}

	emitida, err := EmitirOrdenCompraDesdePlanReposicionBorrador(dbConn, 99, proveedorID, bodegaA, 7, true, "", "2026-04", "COP", "tester@local", "emision fase 11")
	if err != nil {
		t.Fatalf("emitir orden desde borrador: %v", err)
	}

	recepcion, err := ActualizarEstadoOrdenCompraDesdeReposicion(dbConn, 99, proveedorID, emitida.DocumentoCodigo, "recepcionar_compra", "", "2026-04", "recepcion fase12", "tester")
	if err != nil {
		t.Fatalf("recepcionar compra: %v", err)
	}
	if recepcion.EstadoNuevo != "recepcionada" {
		t.Fatalf("estado nuevo recepcion inesperado: %q", recepcion.EstadoNuevo)
	}
	if recepcion.Evento != "compra_recepcionada" {
		t.Fatalf("evento recepcion inesperado: %q", recepcion.Evento)
	}

	contabilizada, err := ActualizarEstadoOrdenCompraDesdeReposicion(dbConn, 99, proveedorID, emitida.DocumentoCodigo, "contabilizar_compra", "", "2026-04", "contabilizacion fase12", "tester")
	if err != nil {
		t.Fatalf("contabilizar compra: %v", err)
	}
	if contabilizada.EstadoNuevo != "contabilizada" {
		t.Fatalf("estado nuevo contabilizacion inesperado: %q", contabilizada.EstadoNuevo)
	}
	if contabilizada.Evento != "compra_contabilizada" {
		t.Fatalf("evento contabilizacion inesperado: %q", contabilizada.Evento)
	}

	doc, err := GetEmpresaDocumentoCompraByCodigo(dbConn, 99, "orden_compra", emitida.DocumentoCodigo)
	if err != nil {
		t.Fatalf("consultar documento final: %v", err)
	}
	if doc.EstadoDocumento != "contabilizada" {
		t.Fatalf("estado final inesperado: %q", doc.EstadoDocumento)
	}

	_, err = ActualizarEstadoOrdenCompraDesdeReposicion(dbConn, 99, proveedorID, emitida.DocumentoCodigo, "contabilizar_compra", "", "2026-04", "contabilizacion duplicada", "tester")
	if err == nil {
		t.Fatalf("se esperaba error por transicion invalida")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "transicion invalida") {
		t.Fatalf("error inesperado en transicion invalida: %v", err)
	}
}

func TestInventarioPoliticaCostoPromedioYPEPS(t *testing.T) {
	dbConn := openProductosTestDB(t)
	if err := EnsureEmpresaProductosSchema(dbConn); err != nil {
		t.Fatalf("ensure schema: %v", err)
	}

	bodegaID, err := CreateBodega(dbConn, Bodega{EmpresaID: 110, Codigo: "BOD-110", Nombre: "Principal"})
	if err != nil {
		t.Fatalf("create bodega: %v", err)
	}

	productoID, err := CreateProducto(dbConn, Producto{EmpresaID: 110, BodegaPrincipalID: bodegaID, Nombre: "Producto costos", Costo: 10, StockMinimo: 1, StockMaximo: 30}, 0, "TEST")
	if err != nil {
		t.Fatalf("create producto: %v", err)
	}

	if _, err := UpsertEmpresaInventarioConfiguracion(dbConn, EmpresaInventarioConfiguracion{EmpresaID: 110, PoliticaCosto: "peps", UsuarioCreador: "tester"}); err != nil {
		t.Fatalf("set config peps: %v", err)
	}

	if err := RegistrarMovimientoInventario(dbConn, 110, productoID, bodegaID, "entrada", 10, "COST-ENT-1", "tester", "entrada lote 1"); err != nil {
		t.Fatalf("entrada lote 1: %v", err)
	}
	if _, err := dbConn.Exec(`UPDATE productos SET costo = 20, fecha_actualizacion = datetime('now','localtime') WHERE empresa_id = 110 AND id = ?`, productoID); err != nil {
		t.Fatalf("update costo producto a 20: %v", err)
	}
	if err := RegistrarMovimientoInventario(dbConn, 110, productoID, bodegaID, "entrada", 10, "COST-ENT-2", "tester", "entrada lote 2"); err != nil {
		t.Fatalf("entrada lote 2: %v", err)
	}
	if err := RegistrarMovimientoInventario(dbConn, 110, productoID, bodegaID, "salida", 12, "COST-SAL-PEPS", "tester", "salida peps"); err != nil {
		t.Fatalf("salida peps: %v", err)
	}

	var costoSalidaPEPS float64
	if err := dbConn.QueryRow(`SELECT costo_unitario FROM inventario_movimientos WHERE empresa_id = 110 AND referencia = 'COST-SAL-PEPS' LIMIT 1`).Scan(&costoSalidaPEPS); err != nil {
		t.Fatalf("query costo salida peps: %v", err)
	}
	if costoSalidaPEPS < 11.66 || costoSalidaPEPS > 11.67 {
		t.Fatalf("expected costo peps ~11.67, got %.4f", costoSalidaPEPS)
	}

	if _, err := dbConn.Exec(`UPDATE productos SET costo = 30, fecha_actualizacion = datetime('now','localtime') WHERE empresa_id = 110 AND id = ?`, productoID); err != nil {
		t.Fatalf("update costo producto a 30: %v", err)
	}
	if err := RegistrarMovimientoInventario(dbConn, 110, productoID, bodegaID, "entrada", 10, "COST-ENT-3", "tester", "entrada lote 3"); err != nil {
		t.Fatalf("entrada lote 3: %v", err)
	}

	if _, err := UpsertEmpresaInventarioConfiguracion(dbConn, EmpresaInventarioConfiguracion{EmpresaID: 110, PoliticaCosto: "promedio", UsuarioCreador: "tester"}); err != nil {
		t.Fatalf("set config promedio: %v", err)
	}
	if err := RegistrarMovimientoInventario(dbConn, 110, productoID, bodegaID, "salida", 4, "COST-SAL-PROM", "tester", "salida promedio"); err != nil {
		t.Fatalf("salida promedio: %v", err)
	}

	var costoSalidaProm float64
	if err := dbConn.QueryRow(`SELECT costo_unitario FROM inventario_movimientos WHERE empresa_id = 110 AND referencia = 'COST-SAL-PROM' LIMIT 1`).Scan(&costoSalidaProm); err != nil {
		t.Fatalf("query costo salida promedio: %v", err)
	}
	if costoSalidaProm < 25.55 || costoSalidaProm > 25.57 {
		t.Fatalf("expected costo promedio ~25.56, got %.4f", costoSalidaProm)
	}
}

func TestRegistrarConteoCiclicoInventarioAjustaYAudita(t *testing.T) {
	dbConn := openProductosTestDB(t)
	if err := EnsureEmpresaProductosSchema(dbConn); err != nil {
		t.Fatalf("ensure schema: %v", err)
	}

	bodegaID, err := CreateBodega(dbConn, Bodega{EmpresaID: 111, Codigo: "BOD-111", Nombre: "Conteo"})
	if err != nil {
		t.Fatalf("create bodega: %v", err)
	}

	productoID, err := CreateProducto(dbConn, Producto{EmpresaID: 111, BodegaPrincipalID: bodegaID, Nombre: "Producto conteo", Costo: 9, StockMinimo: 1, StockMaximo: 20}, 0, "TEST")
	if err != nil {
		t.Fatalf("create producto: %v", err)
	}

	if err := RegistrarMovimientoInventario(dbConn, 111, productoID, bodegaID, "entrada", 14, "CNT-ENT", "tester", "stock inicial conteo"); err != nil {
		t.Fatalf("entrada inicial: %v", err)
	}

	conteo, err := RegistrarConteoCiclicoInventario(dbConn, InventarioConteoCiclico{
		EmpresaID:       111,
		ProductoID:      productoID,
		BodegaID:        bodegaID,
		CantidadContada: 11,
		Referencia:      "CNT-CIC-001",
		UsuarioRevisor:  "auditor@local",
		Observaciones:   "conteo semanal bodega principal",
	})
	if err != nil {
		t.Fatalf("registrar conteo ciclico: %v", err)
	}

	if conteo.TipoAjuste != "ajuste_negativo" {
		t.Fatalf("expected tipo_ajuste ajuste_negativo, got %q", conteo.TipoAjuste)
	}
	if conteo.EstadoConteo != "ajustado" {
		t.Fatalf("expected estado_conteo ajustado, got %q", conteo.EstadoConteo)
	}
	if conteo.MovimientoID <= 0 {
		t.Fatalf("expected movimiento_id > 0, got %+v", conteo)
	}

	var stockFinal float64
	if err := dbConn.QueryRow(`SELECT COALESCE(cantidad, 0) FROM inventario_existencias WHERE empresa_id = 111 AND producto_id = ? AND bodega_id = ? LIMIT 1`, productoID, bodegaID).Scan(&stockFinal); err != nil {
		t.Fatalf("query stock final: %v", err)
	}
	if stockFinal < 10.99 || stockFinal > 11.01 {
		t.Fatalf("expected stock final 11, got %.4f", stockFinal)
	}

	rows, err := GetInventarioConteosCiclicosByEmpresa(dbConn, 111, productoID, bodegaID, "", "", "", 20, 0)
	if err != nil {
		t.Fatalf("list conteos ciclicos: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 conteo registrado, got %d", len(rows))
	}
}

func TestGetAlertasOperativasByEmpresaIncluyeSobrestock(t *testing.T) {
	dbConn := openProductosTestDB(t)
	if err := EnsureEmpresaProductosSchema(dbConn); err != nil {
		t.Fatalf("ensure schema: %v", err)
	}

	bodegaID, err := CreateBodega(dbConn, Bodega{EmpresaID: 112, Codigo: "BOD-112", Nombre: "Alertas"})
	if err != nil {
		t.Fatalf("create bodega: %v", err)
	}

	prodSobrestock, err := CreateProducto(dbConn, Producto{EmpresaID: 112, Nombre: "Producto sobrestock", StockMinimo: 2, StockMaximo: 5}, 0, "TEST")
	if err != nil {
		t.Fatalf("create producto sobrestock: %v", err)
	}
	if _, err := dbConn.Exec(`INSERT INTO inventario_existencias (empresa_id, producto_id, bodega_id, cantidad, estado) VALUES (112, ?, ?, 9, 'activo')`, prodSobrestock, bodegaID); err != nil {
		t.Fatalf("insert sobrestock: %v", err)
	}

	rows, err := GetAlertasOperativasByEmpresa(dbConn, 112, 0, 0, 20, 0)
	if err != nil {
		t.Fatalf("get alertas operativas: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 alerta operativa, got %d", len(rows))
	}
	if rows[0].EstadoStock != "sobrestock" {
		t.Fatalf("expected estado sobrestock, got %+v", rows[0])
	}
	if rows[0].Exceso < 3.99 || rows[0].Exceso > 4.01 {
		t.Fatalf("expected exceso 4, got %.4f", rows[0].Exceso)
	}
}
