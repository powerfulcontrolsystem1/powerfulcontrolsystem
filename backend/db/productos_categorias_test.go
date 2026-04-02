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
