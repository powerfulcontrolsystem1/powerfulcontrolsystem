package db

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

var (
	// ErrStockInsuficiente se usa cuando una salida/traslado excede la existencia disponible.
	ErrStockInsuficiente = errors.New("stock insuficiente")
)

// Bodega representa una ubicación de inventario dentro de una empresa.
type Bodega struct {
	ID                 int64  `json:"id"`
	EmpresaID          int64  `json:"empresa_id"`
	Codigo             string `json:"codigo,omitempty"`
	Nombre             string `json:"nombre"`
	Ubicacion          string `json:"ubicacion,omitempty"`
	Responsable        string `json:"responsable,omitempty"`
	FechaCreacion      string `json:"fecha_creacion,omitempty"`
	FechaActualizacion string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string `json:"usuario_creador,omitempty"`
	Estado             string `json:"estado,omitempty"`
	Observaciones      string `json:"observaciones,omitempty"`
}

// CategoriaProducto representa una categoría de productos dentro de una empresa.
type CategoriaProducto struct {
	ID                 int64  `json:"id"`
	EmpresaID          int64  `json:"empresa_id"`
	Codigo             string `json:"codigo,omitempty"`
	Nombre             string `json:"nombre"`
	Descripcion        string `json:"descripcion,omitempty"`
	ColorHex           string `json:"color_hex,omitempty"`
	Orden              int64  `json:"orden,omitempty"`
	FechaCreacion      string `json:"fecha_creacion,omitempty"`
	FechaActualizacion string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string `json:"usuario_creador,omitempty"`
	Estado             string `json:"estado,omitempty"`
	Observaciones      string `json:"observaciones,omitempty"`
}

// Producto representa un producto asociado a una empresa.
type Producto struct {
	ID                   int64   `json:"id"`
	EmpresaID            int64   `json:"empresa_id"`
	BodegaPrincipalID    int64   `json:"bodega_principal_id,omitempty"`
	ProveedorPrincipalID int64   `json:"proveedor_principal_id,omitempty"`
	CategoriaID          int64   `json:"categoria_id,omitempty"`
	SKU                  string  `json:"sku,omitempty"`
	CodigoBarras         string  `json:"codigo_barras,omitempty"`
	Nombre               string  `json:"nombre"`
	Descripcion          string  `json:"descripcion,omitempty"`
	Categoria            string  `json:"categoria,omitempty"`
	Marca                string  `json:"marca,omitempty"`
	UnidadMedida         string  `json:"unidad_medida,omitempty"`
	Costo                float64 `json:"costo"`
	Precio               float64 `json:"precio"`
	ImpuestoPorcentaje   float64 `json:"impuesto_porcentaje"`
	StockMinimo          float64 `json:"stock_minimo"`
	StockMaximo          float64 `json:"stock_maximo"`
	StockTotal           float64 `json:"stock_total"`
	ImagenURL            string  `json:"imagen_url,omitempty"`
	FechaCreacion        string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion   string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador       string  `json:"usuario_creador,omitempty"`
	Estado               string  `json:"estado,omitempty"`
	Observaciones        string  `json:"observaciones,omitempty"`
}

// Proveedor representa un proveedor comercial de la empresa.
type Proveedor struct {
	ID                 int64  `json:"id"`
	EmpresaID          int64  `json:"empresa_id"`
	Codigo             string `json:"codigo,omitempty"`
	Nombre             string `json:"nombre"`
	Documento          string `json:"documento,omitempty"`
	Contacto           string `json:"contacto,omitempty"`
	Telefono           string `json:"telefono,omitempty"`
	Email              string `json:"email,omitempty"`
	Direccion          string `json:"direccion,omitempty"`
	FechaCreacion      string `json:"fecha_creacion,omitempty"`
	FechaActualizacion string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string `json:"usuario_creador,omitempty"`
	Estado             string `json:"estado,omitempty"`
	Observaciones      string `json:"observaciones,omitempty"`
}

// Servicio representa un servicio vendible como si fuese un producto.
type Servicio struct {
	ID                 int64   `json:"id"`
	EmpresaID          int64   `json:"empresa_id"`
	Codigo             string  `json:"codigo,omitempty"`
	Nombre             string  `json:"nombre"`
	Descripcion        string  `json:"descripcion,omitempty"`
	Categoria          string  `json:"categoria,omitempty"`
	DuracionMinutos    int     `json:"duracion_minutos,omitempty"`
	CostoReferencial   float64 `json:"costo_referencial"`
	Precio             float64 `json:"precio"`
	ImpuestoPorcentaje float64 `json:"impuesto_porcentaje"`
	ImagenURL          string  `json:"imagen_url,omitempty"`
	FechaCreacion      string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string  `json:"usuario_creador,omitempty"`
	Estado             string  `json:"estado,omitempty"`
	Observaciones      string  `json:"observaciones,omitempty"`
}

// ProductoPrecioHistorial representa cambios de precio/costo/impuesto de productos.
type ProductoPrecioHistorial struct {
	ID                 int64   `json:"id"`
	EmpresaID          int64   `json:"empresa_id"`
	ProductoID         int64   `json:"producto_id"`
	ProductoNombre     string  `json:"producto_nombre,omitempty"`
	CostoAnterior      float64 `json:"costo_anterior"`
	CostoNuevo         float64 `json:"costo_nuevo"`
	PrecioAnterior     float64 `json:"precio_anterior"`
	PrecioNuevo        float64 `json:"precio_nuevo"`
	ImpuestoAnterior   float64 `json:"impuesto_anterior"`
	ImpuestoNuevo      float64 `json:"impuesto_nuevo"`
	Motivo             string  `json:"motivo,omitempty"`
	Referencia         string  `json:"referencia,omitempty"`
	FechaCambio        string  `json:"fecha_cambio,omitempty"`
	FechaCreacion      string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string  `json:"usuario_creador,omitempty"`
	Estado             string  `json:"estado,omitempty"`
	Observaciones      string  `json:"observaciones,omitempty"`
}

// InventarioExistencia representa el stock por producto y bodega.
type InventarioExistencia struct {
	ID                 int64   `json:"id"`
	EmpresaID          int64   `json:"empresa_id"`
	ProductoID         int64   `json:"producto_id"`
	ProductoNombre     string  `json:"producto_nombre,omitempty"`
	BodegaID           int64   `json:"bodega_id"`
	BodegaNombre       string  `json:"bodega_nombre,omitempty"`
	Cantidad           float64 `json:"cantidad"`
	FechaCreacion      string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string  `json:"usuario_creador,omitempty"`
	Estado             string  `json:"estado,omitempty"`
	Observaciones      string  `json:"observaciones,omitempty"`
}

// InventarioMovimiento representa un evento de inventario (entrada/salida/traslado/ajuste).
type InventarioMovimiento struct {
	ID                  int64   `json:"id"`
	EmpresaID           int64   `json:"empresa_id"`
	ProductoID          int64   `json:"producto_id"`
	ProductoNombre      string  `json:"producto_nombre,omitempty"`
	BodegaOrigenID      int64   `json:"bodega_origen_id,omitempty"`
	BodegaOrigenNombre  string  `json:"bodega_origen_nombre,omitempty"`
	BodegaDestinoID     int64   `json:"bodega_destino_id,omitempty"`
	BodegaDestinoNombre string  `json:"bodega_destino_nombre,omitempty"`
	Tipo                string  `json:"tipo"`
	Cantidad            float64 `json:"cantidad"`
	CostoUnitario       float64 `json:"costo_unitario"`
	Referencia          string  `json:"referencia,omitempty"`
	FechaMovimiento     string  `json:"fecha_movimiento,omitempty"`
	FechaCreacion       string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion  string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador      string  `json:"usuario_creador,omitempty"`
	Estado              string  `json:"estado,omitempty"`
	Observaciones       string  `json:"observaciones,omitempty"`
}

// EnsureEmpresaProductosSchema crea y migra las tablas del módulo de productos en empresas.db.
func EnsureEmpresaProductosSchema(dbConn *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS bodegas (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			codigo TEXT,
			nombre TEXT NOT NULL,
			ubicacion TEXT,
			responsable TEXT,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS categorias_productos (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			codigo TEXT,
			nombre TEXT NOT NULL,
			descripcion TEXT,
			color_hex TEXT,
			orden INTEGER DEFAULT 0,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS productos (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			bodega_principal_id INTEGER,
			proveedor_principal_id INTEGER,
			categoria_id INTEGER,
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
		);`,
		`CREATE TABLE IF NOT EXISTS proveedores (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			codigo TEXT,
			nombre TEXT NOT NULL,
			documento TEXT,
			contacto TEXT,
			telefono TEXT,
			email TEXT,
			direccion TEXT,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS servicios (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			codigo TEXT,
			nombre TEXT NOT NULL,
			descripcion TEXT,
			categoria TEXT,
			duracion_minutos INTEGER DEFAULT 0,
			costo_referencial REAL DEFAULT 0,
			precio REAL DEFAULT 0,
			impuesto_porcentaje REAL DEFAULT 0,
			imagen_url TEXT,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS producto_precios_historial (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			producto_id INTEGER NOT NULL,
			costo_anterior REAL DEFAULT 0,
			costo_nuevo REAL DEFAULT 0,
			precio_anterior REAL DEFAULT 0,
			precio_nuevo REAL DEFAULT 0,
			impuesto_anterior REAL DEFAULT 0,
			impuesto_nuevo REAL DEFAULT 0,
			motivo TEXT,
			referencia TEXT,
			fecha_cambio TEXT DEFAULT (datetime('now','localtime')),
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS inventario_existencias (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			producto_id INTEGER NOT NULL,
			bodega_id INTEGER NOT NULL,
			cantidad REAL NOT NULL DEFAULT 0,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS inventario_movimientos (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			producto_id INTEGER NOT NULL,
			bodega_origen_id INTEGER,
			bodega_destino_id INTEGER,
			tipo TEXT NOT NULL,
			cantidad REAL NOT NULL,
			costo_unitario REAL DEFAULT 0,
			referencia TEXT,
			fecha_movimiento TEXT DEFAULT (datetime('now','localtime')),
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_bodegas_empresa_codigo ON bodegas(empresa_id, codigo);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_bodegas_empresa_nombre ON bodegas(empresa_id, nombre);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_categorias_productos_empresa_codigo ON categorias_productos(empresa_id, codigo);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_categorias_productos_empresa_nombre ON categorias_productos(empresa_id, nombre);`,
		`CREATE INDEX IF NOT EXISTS ix_categorias_productos_empresa_orden ON categorias_productos(empresa_id, orden, nombre);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_productos_empresa_sku ON productos(empresa_id, sku);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_productos_empresa_barras ON productos(empresa_id, codigo_barras);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_proveedores_empresa_codigo ON proveedores(empresa_id, codigo);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_proveedores_empresa_nombre ON proveedores(empresa_id, nombre);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_servicios_empresa_codigo ON servicios(empresa_id, codigo);`,
		`CREATE INDEX IF NOT EXISTS ix_historial_precios_empresa_producto ON producto_precios_historial(empresa_id, producto_id);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_existencias_empresa_prod_bodega ON inventario_existencias(empresa_id, producto_id, bodega_id);`,
		`CREATE INDEX IF NOT EXISTS ix_existencias_empresa_bodega ON inventario_existencias(empresa_id, bodega_id);`,
		`CREATE INDEX IF NOT EXISTS ix_movimientos_empresa_producto ON inventario_movimientos(empresa_id, producto_id);`,
	}

	for _, stmt := range stmts {
		if _, err := dbConn.Exec(stmt); err != nil {
			return err
		}
	}

	// Migraciones seguras para instalaciones previas.
	if err := ensureColumnIfMissing(dbConn, "bodegas", "codigo", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "bodegas", "ubicacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "bodegas", "responsable", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "bodegas", "fecha_actualizacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "bodegas", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "bodegas", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "bodegas", "observaciones", "TEXT"); err != nil {
		return err
	}

	if err := ensureColumnIfMissing(dbConn, "categorias_productos", "codigo", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "categorias_productos", "nombre", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "categorias_productos", "descripcion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "categorias_productos", "color_hex", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "categorias_productos", "orden", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "categorias_productos", "fecha_actualizacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "categorias_productos", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "categorias_productos", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "categorias_productos", "observaciones", "TEXT"); err != nil {
		return err
	}

	if err := ensureColumnIfMissing(dbConn, "productos", "bodega_principal_id", "INTEGER"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "productos", "proveedor_principal_id", "INTEGER"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "productos", "categoria_id", "INTEGER"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "productos", "codigo_barras", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "productos", "categoria", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "productos", "marca", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "productos", "unidad_medida", "TEXT DEFAULT 'unidad'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "productos", "costo", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "productos", "precio", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "productos", "impuesto_porcentaje", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "productos", "stock_minimo", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "productos", "stock_maximo", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "productos", "imagen_url", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "productos", "fecha_actualizacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "productos", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "productos", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "productos", "observaciones", "TEXT"); err != nil {
		return err
	}
	if _, err := dbConn.Exec(`CREATE INDEX IF NOT EXISTS ix_productos_empresa_categoria_id ON productos(empresa_id, categoria_id);`); err != nil {
		return err
	}

	if err := ensureColumnIfMissing(dbConn, "proveedores", "codigo", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "proveedores", "documento", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "proveedores", "contacto", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "proveedores", "telefono", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "proveedores", "email", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "proveedores", "direccion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "proveedores", "fecha_actualizacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "proveedores", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "proveedores", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "proveedores", "observaciones", "TEXT"); err != nil {
		return err
	}

	if err := ensureColumnIfMissing(dbConn, "servicios", "codigo", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "servicios", "categoria", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "servicios", "duracion_minutos", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "servicios", "costo_referencial", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "servicios", "precio", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "servicios", "impuesto_porcentaje", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "servicios", "imagen_url", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "servicios", "fecha_actualizacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "servicios", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "servicios", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "servicios", "observaciones", "TEXT"); err != nil {
		return err
	}

	if err := ensureColumnIfMissing(dbConn, "producto_precios_historial", "costo_anterior", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "producto_precios_historial", "costo_nuevo", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "producto_precios_historial", "precio_anterior", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "producto_precios_historial", "precio_nuevo", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "producto_precios_historial", "impuesto_anterior", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "producto_precios_historial", "impuesto_nuevo", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "producto_precios_historial", "motivo", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "producto_precios_historial", "referencia", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "producto_precios_historial", "fecha_cambio", "TEXT DEFAULT (datetime('now','localtime'))"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "producto_precios_historial", "fecha_actualizacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "producto_precios_historial", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "producto_precios_historial", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "producto_precios_historial", "observaciones", "TEXT"); err != nil {
		return err
	}

	if err := ensureColumnIfMissing(dbConn, "inventario_existencias", "fecha_actualizacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "inventario_existencias", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "inventario_existencias", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "inventario_existencias", "observaciones", "TEXT"); err != nil {
		return err
	}

	if err := ensureColumnIfMissing(dbConn, "inventario_movimientos", "fecha_movimiento", "TEXT DEFAULT (datetime('now','localtime'))"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "inventario_movimientos", "fecha_actualizacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "inventario_movimientos", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "inventario_movimientos", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "inventario_movimientos", "observaciones", "TEXT"); err != nil {
		return err
	}
	if err := seedCategoriasProductosFromLegacy(dbConn); err != nil {
		return err
	}
	if err := backfillProductoCategoriaIDs(dbConn); err != nil {
		return err
	}

	return nil
}

func ensureColumnIfMissing(dbConn *sql.DB, table, column, columnDef string) error {
	rows, err := dbConn.Query(fmt.Sprintf("PRAGMA table_info(%s);", table))
	if err != nil {
		return err
	}
	defer rows.Close()

	lowerColumn := strings.ToLower(column)
	for rows.Next() {
		var cid int
		var name string
		var ctype string
		var notnull int
		var dflt sql.NullString
		var pk int
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			return err
		}
		if strings.ToLower(name) == lowerColumn {
			return nil
		}
	}
	_, err = dbConn.Exec(fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s", table, column+" "+columnDef))
	return err
}

// CreateBodega inserta una nueva bodega para una empresa.
func CreateBodega(dbConn *sql.DB, b Bodega) (int64, error) {
	res, err := dbConn.Exec(`INSERT INTO bodegas (empresa_id, codigo, nombre, ubicacion, responsable, usuario_creador, estado, observaciones, fecha_creacion, fecha_actualizacion)
		VALUES (?, ?, ?, ?, ?, ?, COALESCE(NULLIF(?, ''), 'activo'), ?, datetime('now','localtime'), datetime('now','localtime'))`,
		b.EmpresaID, strings.TrimSpace(b.Codigo), strings.TrimSpace(b.Nombre), strings.TrimSpace(b.Ubicacion), strings.TrimSpace(b.Responsable), strings.TrimSpace(b.UsuarioCreador), strings.TrimSpace(b.Estado), strings.TrimSpace(b.Observaciones))
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// GetBodegasByEmpresa lista bodegas por empresa.
func GetBodegasByEmpresa(dbConn *sql.DB, empresaID int64, incluirInactivas bool) ([]Bodega, error) {
	query := `SELECT id, empresa_id, codigo, nombre, ubicacion, responsable, fecha_creacion, fecha_actualizacion, usuario_creador, estado, observaciones
		FROM bodegas
		WHERE empresa_id = ?`
	args := []interface{}{empresaID}
	if !incluirInactivas {
		query += ` AND estado = 'activo'`
	}
	query += ` ORDER BY nombre ASC`

	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]Bodega, 0)
	for rows.Next() {
		var b Bodega
		var codigo, ubicacion, responsable, fechaCreacion, fechaAct, usuario, estado, obs sql.NullString
		if err := rows.Scan(&b.ID, &b.EmpresaID, &codigo, &b.Nombre, &ubicacion, &responsable, &fechaCreacion, &fechaAct, &usuario, &estado, &obs); err != nil {
			return nil, err
		}
		if codigo.Valid {
			b.Codigo = codigo.String
		}
		if ubicacion.Valid {
			b.Ubicacion = ubicacion.String
		}
		if responsable.Valid {
			b.Responsable = responsable.String
		}
		if fechaCreacion.Valid {
			b.FechaCreacion = fechaCreacion.String
		}
		if fechaAct.Valid {
			b.FechaActualizacion = fechaAct.String
		}
		if usuario.Valid {
			b.UsuarioCreador = usuario.String
		}
		if estado.Valid {
			b.Estado = estado.String
		}
		if obs.Valid {
			b.Observaciones = obs.String
		}
		out = append(out, b)
	}
	return out, nil
}

// UpdateBodega actualiza los datos editables de una bodega.
func UpdateBodega(dbConn *sql.DB, b Bodega) error {
	_, err := dbConn.Exec(`UPDATE bodegas
		SET codigo = ?, nombre = ?, ubicacion = ?, responsable = ?, observaciones = ?, fecha_actualizacion = datetime('now','localtime')
		WHERE id = ? AND empresa_id = ?`,
		strings.TrimSpace(b.Codigo), strings.TrimSpace(b.Nombre), strings.TrimSpace(b.Ubicacion), strings.TrimSpace(b.Responsable), strings.TrimSpace(b.Observaciones), b.ID, b.EmpresaID)
	return err
}

// DeleteBodega elimina una bodega de la empresa.
func DeleteBodega(dbConn *sql.DB, empresaID, bodegaID int64) error {
	_, err := dbConn.Exec("DELETE FROM bodegas WHERE id = ? AND empresa_id = ?", bodegaID, empresaID)
	return err
}

// SetBodegaEstado activa/desactiva una bodega.
func SetBodegaEstado(dbConn *sql.DB, empresaID, bodegaID int64, estado string) error {
	_, err := dbConn.Exec("UPDATE bodegas SET estado = ?, fecha_actualizacion = datetime('now','localtime') WHERE id = ? AND empresa_id = ?", strings.TrimSpace(estado), bodegaID, empresaID)
	return err
}

// CreateCategoriaProducto inserta una categoría de producto para una empresa.
func CreateCategoriaProducto(dbConn *sql.DB, c CategoriaProducto) (int64, error) {
	res, err := dbConn.Exec(`INSERT INTO categorias_productos (
		empresa_id, codigo, nombre, descripcion, color_hex, orden,
		usuario_creador, estado, observaciones, fecha_creacion, fecha_actualizacion
	) VALUES (?, NULLIF(?, ''), ?, ?, ?, ?, ?, COALESCE(NULLIF(?, ''), 'activo'), ?, datetime('now','localtime'), datetime('now','localtime'))`,
		c.EmpresaID, strings.TrimSpace(c.Codigo), strings.TrimSpace(c.Nombre), strings.TrimSpace(c.Descripcion), strings.TrimSpace(c.ColorHex), c.Orden,
		strings.TrimSpace(c.UsuarioCreador), strings.TrimSpace(c.Estado), strings.TrimSpace(c.Observaciones))
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// GetCategoriasProductoByEmpresa lista categorías de producto por empresa.
func GetCategoriasProductoByEmpresa(dbConn *sql.DB, empresaID int64, incluirInactivas bool, filtro string) ([]CategoriaProducto, error) {
	query := `SELECT id, empresa_id, codigo, nombre, descripcion, color_hex, orden,
		fecha_creacion, fecha_actualizacion, usuario_creador, estado, observaciones
	FROM categorias_productos
	WHERE empresa_id = ?`
	args := []interface{}{empresaID}
	if !incluirInactivas {
		query += ` AND estado = 'activo'`
	}
	if strings.TrimSpace(filtro) != "" {
		like := "%" + strings.TrimSpace(filtro) + "%"
		query += ` AND (nombre LIKE ? OR codigo LIKE ? OR descripcion LIKE ?)`
		args = append(args, like, like, like)
	}
	query += ` ORDER BY orden ASC, nombre ASC`

	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]CategoriaProducto, 0)
	for rows.Next() {
		var c CategoriaProducto
		var codigo, descripcion, colorHex, fechaCreacion, fechaActualizacion, usuario, estado, observaciones sql.NullString
		if err := rows.Scan(&c.ID, &c.EmpresaID, &codigo, &c.Nombre, &descripcion, &colorHex, &c.Orden, &fechaCreacion, &fechaActualizacion, &usuario, &estado, &observaciones); err != nil {
			return nil, err
		}
		if codigo.Valid {
			c.Codigo = codigo.String
		}
		if descripcion.Valid {
			c.Descripcion = descripcion.String
		}
		if colorHex.Valid {
			c.ColorHex = colorHex.String
		}
		if fechaCreacion.Valid {
			c.FechaCreacion = fechaCreacion.String
		}
		if fechaActualizacion.Valid {
			c.FechaActualizacion = fechaActualizacion.String
		}
		if usuario.Valid {
			c.UsuarioCreador = usuario.String
		}
		if estado.Valid {
			c.Estado = estado.String
		}
		if observaciones.Valid {
			c.Observaciones = observaciones.String
		}
		out = append(out, c)
	}
	return out, nil
}

// UpdateCategoriaProducto actualiza una categoría y sincroniza el nombre en productos asociados.
func UpdateCategoriaProducto(dbConn *sql.DB, c CategoriaProducto) error {
	tx, err := dbConn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`UPDATE categorias_productos
		SET codigo = NULLIF(?, ''), nombre = ?, descripcion = ?, color_hex = ?, orden = ?, observaciones = ?, fecha_actualizacion = datetime('now','localtime')
		WHERE id = ? AND empresa_id = ?`,
		strings.TrimSpace(c.Codigo), strings.TrimSpace(c.Nombre), strings.TrimSpace(c.Descripcion), strings.TrimSpace(c.ColorHex), c.Orden, strings.TrimSpace(c.Observaciones), c.ID, c.EmpresaID); err != nil {
		return err
	}

	if _, err := tx.Exec(`UPDATE productos
		SET categoria = ?, fecha_actualizacion = datetime('now','localtime')
		WHERE empresa_id = ? AND categoria_id = ?`, strings.TrimSpace(c.Nombre), c.EmpresaID, c.ID); err != nil {
		return err
	}

	return tx.Commit()
}

// DeleteCategoriaProducto elimina una categoría si no está asociada a productos.
func DeleteCategoriaProducto(dbConn *sql.DB, empresaID, categoriaID int64) error {
	var totalProductos int64
	if err := dbConn.QueryRow(`SELECT COUNT(1) FROM productos WHERE empresa_id = ? AND categoria_id = ?`, empresaID, categoriaID).Scan(&totalProductos); err != nil {
		return err
	}
	if totalProductos > 0 {
		return fmt.Errorf("no se puede eliminar la categoría porque está asociada a productos")
	}
	_, err := dbConn.Exec(`DELETE FROM categorias_productos WHERE id = ? AND empresa_id = ?`, categoriaID, empresaID)
	return err
}

// SetCategoriaProductoEstado activa/desactiva una categoría de producto.
func SetCategoriaProductoEstado(dbConn *sql.DB, empresaID, categoriaID int64, estado string) error {
	_, err := dbConn.Exec(`UPDATE categorias_productos SET estado = ?, fecha_actualizacion = datetime('now','localtime') WHERE id = ? AND empresa_id = ?`, strings.TrimSpace(estado), categoriaID, empresaID)
	return err
}

// CreateProducto crea un producto y opcionalmente su stock inicial.
func CreateProducto(dbConn *sql.DB, p Producto, stockInicial float64, referenciaInicial string) (int64, error) {
	tx, err := dbConn.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	if p.BodegaPrincipalID > 0 {
		if err := validateBodegaEmpresaTx(tx, p.EmpresaID, p.BodegaPrincipalID); err != nil {
			return 0, err
		}
	}
	if p.ProveedorPrincipalID > 0 {
		if err := validateProveedorEmpresaTx(tx, p.EmpresaID, p.ProveedorPrincipalID); err != nil {
			return 0, err
		}
	}
	if p.CategoriaID > 0 {
		categoriaNombre, err := resolveCategoriaProductoTx(tx, p.EmpresaID, p.CategoriaID)
		if err != nil {
			return 0, err
		}
		p.Categoria = categoriaNombre
	}

	res, err := tx.Exec(`INSERT INTO productos (
		empresa_id, bodega_principal_id, proveedor_principal_id, categoria_id, sku, codigo_barras, nombre, descripcion, categoria, marca, unidad_medida,
		costo, precio, impuesto_porcentaje, stock_minimo, stock_maximo, imagen_url,
		usuario_creador, estado, observaciones, fecha_creacion, fecha_actualizacion
	) VALUES (?, ?, ?, ?, NULLIF(?, ''), NULLIF(?, ''), ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, COALESCE(NULLIF(?, ''), 'activo'), ?, datetime('now','localtime'), datetime('now','localtime'))`,
		p.EmpresaID, nullableInt64(p.BodegaPrincipalID), nullableInt64(p.ProveedorPrincipalID), nullableInt64(p.CategoriaID), strings.TrimSpace(p.SKU), strings.TrimSpace(p.CodigoBarras), strings.TrimSpace(p.Nombre), strings.TrimSpace(p.Descripcion), strings.TrimSpace(p.Categoria), strings.TrimSpace(p.Marca), defaultUnidad(p.UnidadMedida),
		p.Costo, p.Precio, p.ImpuestoPorcentaje, p.StockMinimo, p.StockMaximo, strings.TrimSpace(p.ImagenURL),
		strings.TrimSpace(p.UsuarioCreador), strings.TrimSpace(p.Estado), strings.TrimSpace(p.Observaciones))
	if err != nil {
		return 0, err
	}
	productoID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	if stockInicial > 0 && p.BodegaPrincipalID > 0 {
		if err := upsertExistenciaTx(tx, p.EmpresaID, productoID, p.BodegaPrincipalID, stockInicial, p.UsuarioCreador, "entrada inicial"); err != nil {
			return 0, err
		}
		if err := insertMovimientoTx(tx, InventarioMovimiento{
			EmpresaID:       p.EmpresaID,
			ProductoID:      productoID,
			BodegaDestinoID: p.BodegaPrincipalID,
			Tipo:            "entrada",
			Cantidad:        stockInicial,
			CostoUnitario:   p.Costo,
			Referencia:      strings.TrimSpace(referenciaInicial),
			UsuarioCreador:  strings.TrimSpace(p.UsuarioCreador),
			Estado:          "activo",
			Observaciones:   "stock inicial al crear producto",
		}); err != nil {
			return 0, err
		}
	}

	if err := insertProductoPrecioHistorialTx(tx, ProductoPrecioHistorial{
		EmpresaID:        p.EmpresaID,
		ProductoID:       productoID,
		CostoAnterior:    0,
		CostoNuevo:       p.Costo,
		PrecioAnterior:   0,
		PrecioNuevo:      p.Precio,
		ImpuestoAnterior: 0,
		ImpuestoNuevo:    p.ImpuestoPorcentaje,
		Motivo:           "creacion_producto",
		Referencia:       strings.TrimSpace(referenciaInicial),
		UsuarioCreador:   strings.TrimSpace(p.UsuarioCreador),
		Estado:           "activo",
		Observaciones:    "historial inicial de precio",
	}); err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return productoID, nil
}

// GetProductosByEmpresa lista productos con filtros y stock total.
func GetProductosByEmpresa(dbConn *sql.DB, empresaID int64, filtro, estado string, bodegaID, categoriaID int64, limit, offset int) ([]Producto, error) {
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}

	query := `SELECT
		p.id, p.empresa_id, p.bodega_principal_id, p.proveedor_principal_id, p.categoria_id, p.sku, p.codigo_barras, p.nombre, p.descripcion, COALESCE(NULLIF(cp.nombre, ''), p.categoria), p.marca, p.unidad_medida,
		p.costo, p.precio, p.impuesto_porcentaje, p.stock_minimo, p.stock_maximo, p.imagen_url,
		p.fecha_creacion, p.fecha_actualizacion, p.usuario_creador, p.estado, p.observaciones,
		COALESCE(SUM(e.cantidad), 0) AS stock_total
	FROM productos p
	LEFT JOIN categorias_productos cp ON cp.empresa_id = p.empresa_id AND cp.id = p.categoria_id
	LEFT JOIN inventario_existencias e ON e.empresa_id = p.empresa_id AND e.producto_id = p.id
	WHERE p.empresa_id = ?`

	args := []interface{}{empresaID}
	if strings.TrimSpace(estado) != "" {
		query += " AND p.estado = ?"
		args = append(args, strings.TrimSpace(estado))
	}
	if strings.TrimSpace(filtro) != "" {
		like := "%" + strings.TrimSpace(filtro) + "%"
		query += " AND (p.nombre LIKE ? OR p.sku LIKE ? OR p.codigo_barras LIKE ? OR p.marca LIKE ? OR p.categoria LIKE ? OR cp.nombre LIKE ?)"
		args = append(args, like, like, like, like, like, like)
	}
	if bodegaID > 0 {
		query += ` AND EXISTS (
			SELECT 1 FROM inventario_existencias ex
			WHERE ex.empresa_id = p.empresa_id AND ex.producto_id = p.id AND ex.bodega_id = ?
		)`
		args = append(args, bodegaID)
	}
	if categoriaID > 0 {
		query += ` AND p.categoria_id = ?`
		args = append(args, categoriaID)
	}

	query += ` GROUP BY p.id ORDER BY p.id DESC LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]Producto, 0)
	for rows.Next() {
		var p Producto
		var bodegaPrincipalID sql.NullInt64
		var proveedorPrincipalID sql.NullInt64
		var categoriaIDVal sql.NullInt64
		var sku, codigoBarras, desc, categoria, marca, unidad, imagenURL, fechaCre, fechaAct, usuario, estadoVal, obs sql.NullString
		if err := rows.Scan(
			&p.ID, &p.EmpresaID, &bodegaPrincipalID, &proveedorPrincipalID, &categoriaIDVal, &sku, &codigoBarras, &p.Nombre, &desc, &categoria, &marca, &unidad,
			&p.Costo, &p.Precio, &p.ImpuestoPorcentaje, &p.StockMinimo, &p.StockMaximo, &imagenURL,
			&fechaCre, &fechaAct, &usuario, &estadoVal, &obs,
			&p.StockTotal,
		); err != nil {
			return nil, err
		}
		if bodegaPrincipalID.Valid {
			p.BodegaPrincipalID = bodegaPrincipalID.Int64
		}
		if proveedorPrincipalID.Valid {
			p.ProveedorPrincipalID = proveedorPrincipalID.Int64
		}
		if categoriaIDVal.Valid {
			p.CategoriaID = categoriaIDVal.Int64
		}
		if sku.Valid {
			p.SKU = sku.String
		}
		if codigoBarras.Valid {
			p.CodigoBarras = codigoBarras.String
		}
		if desc.Valid {
			p.Descripcion = desc.String
		}
		if categoria.Valid {
			p.Categoria = categoria.String
		}
		if marca.Valid {
			p.Marca = marca.String
		}
		if unidad.Valid {
			p.UnidadMedida = unidad.String
		}
		if imagenURL.Valid {
			p.ImagenURL = imagenURL.String
		}
		if fechaCre.Valid {
			p.FechaCreacion = fechaCre.String
		}
		if fechaAct.Valid {
			p.FechaActualizacion = fechaAct.String
		}
		if usuario.Valid {
			p.UsuarioCreador = usuario.String
		}
		if estadoVal.Valid {
			p.Estado = estadoVal.String
		}
		if obs.Valid {
			p.Observaciones = obs.String
		}
		out = append(out, p)
	}

	return out, nil
}

// GetProductoByID devuelve un producto específico por empresa.
func GetProductoByID(dbConn *sql.DB, empresaID, productoID int64) (*Producto, error) {
	query := `SELECT p.id, p.empresa_id, p.bodega_principal_id, p.proveedor_principal_id, p.categoria_id, p.sku, p.codigo_barras, p.nombre, p.descripcion, COALESCE(NULLIF(cp.nombre, ''), p.categoria), p.marca, p.unidad_medida,
		p.costo, p.precio, p.impuesto_porcentaje, p.stock_minimo, p.stock_maximo, p.imagen_url,
		p.fecha_creacion, p.fecha_actualizacion, p.usuario_creador, p.estado, p.observaciones
	FROM productos p
	LEFT JOIN categorias_productos cp ON cp.empresa_id = p.empresa_id AND cp.id = p.categoria_id
	WHERE p.empresa_id = ? AND p.id = ? LIMIT 1`

	row := dbConn.QueryRow(query, empresaID, productoID)
	var p Producto
	var bodegaPrincipalID sql.NullInt64
	var proveedorPrincipalID sql.NullInt64
	var categoriaID sql.NullInt64
	var sku, codigoBarras, desc, categoria, marca, unidad, imagenURL, fechaCre, fechaAct, usuario, estadoVal, obs sql.NullString
	if err := row.Scan(
		&p.ID, &p.EmpresaID, &bodegaPrincipalID, &proveedorPrincipalID, &categoriaID, &sku, &codigoBarras, &p.Nombre, &desc, &categoria, &marca, &unidad,
		&p.Costo, &p.Precio, &p.ImpuestoPorcentaje, &p.StockMinimo, &p.StockMaximo, &imagenURL,
		&fechaCre, &fechaAct, &usuario, &estadoVal, &obs,
	); err != nil {
		return nil, err
	}
	if bodegaPrincipalID.Valid {
		p.BodegaPrincipalID = bodegaPrincipalID.Int64
	}
	if proveedorPrincipalID.Valid {
		p.ProveedorPrincipalID = proveedorPrincipalID.Int64
	}
	if categoriaID.Valid {
		p.CategoriaID = categoriaID.Int64
	}
	if sku.Valid {
		p.SKU = sku.String
	}
	if codigoBarras.Valid {
		p.CodigoBarras = codigoBarras.String
	}
	if desc.Valid {
		p.Descripcion = desc.String
	}
	if categoria.Valid {
		p.Categoria = categoria.String
	}
	if marca.Valid {
		p.Marca = marca.String
	}
	if unidad.Valid {
		p.UnidadMedida = unidad.String
	}
	if imagenURL.Valid {
		p.ImagenURL = imagenURL.String
	}
	if fechaCre.Valid {
		p.FechaCreacion = fechaCre.String
	}
	if fechaAct.Valid {
		p.FechaActualizacion = fechaAct.String
	}
	if usuario.Valid {
		p.UsuarioCreador = usuario.String
	}
	if estadoVal.Valid {
		p.Estado = estadoVal.String
	}
	if obs.Valid {
		p.Observaciones = obs.String
	}
	return &p, nil
}

// UpdateProducto actualiza un producto de la empresa.
func UpdateProducto(dbConn *sql.DB, p Producto, motivoCambio, referenciaCambio string) error {
	tx, err := dbConn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if p.BodegaPrincipalID > 0 {
		if err := validateBodegaEmpresaTx(tx, p.EmpresaID, p.BodegaPrincipalID); err != nil {
			return err
		}
	}
	if p.ProveedorPrincipalID > 0 {
		if err := validateProveedorEmpresaTx(tx, p.EmpresaID, p.ProveedorPrincipalID); err != nil {
			return err
		}
	}
	if p.CategoriaID > 0 {
		categoriaNombre, err := resolveCategoriaProductoTx(tx, p.EmpresaID, p.CategoriaID)
		if err != nil {
			return err
		}
		p.Categoria = categoriaNombre
	}

	var costoAnterior, precioAnterior, impuestoAnterior float64
	if err := tx.QueryRow(`SELECT costo, precio, impuesto_porcentaje FROM productos WHERE id = ? AND empresa_id = ? LIMIT 1`, p.ID, p.EmpresaID).Scan(&costoAnterior, &precioAnterior, &impuestoAnterior); err != nil {
		return err
	}

	if _, err := tx.Exec(`UPDATE productos
		SET bodega_principal_id = ?, proveedor_principal_id = ?, categoria_id = ?, sku = NULLIF(?, ''), codigo_barras = NULLIF(?, ''), nombre = ?, descripcion = ?, categoria = ?, marca = ?, unidad_medida = ?,
			costo = ?, precio = ?, impuesto_porcentaje = ?, stock_minimo = ?, stock_maximo = ?, imagen_url = ?, observaciones = ?, fecha_actualizacion = datetime('now','localtime')
		WHERE id = ? AND empresa_id = ?`,
		nullableInt64(p.BodegaPrincipalID), nullableInt64(p.ProveedorPrincipalID), nullableInt64(p.CategoriaID), strings.TrimSpace(p.SKU), strings.TrimSpace(p.CodigoBarras), strings.TrimSpace(p.Nombre), strings.TrimSpace(p.Descripcion), strings.TrimSpace(p.Categoria), strings.TrimSpace(p.Marca), defaultUnidad(p.UnidadMedida),
		p.Costo, p.Precio, p.ImpuestoPorcentaje, p.StockMinimo, p.StockMaximo, strings.TrimSpace(p.ImagenURL), strings.TrimSpace(p.Observaciones), p.ID, p.EmpresaID); err != nil {
		return err
	}

	if costoAnterior != p.Costo || precioAnterior != p.Precio || impuestoAnterior != p.ImpuestoPorcentaje {
		if strings.TrimSpace(motivoCambio) == "" {
			motivoCambio = "actualizacion_precio"
		}
		if err := insertProductoPrecioHistorialTx(tx, ProductoPrecioHistorial{
			EmpresaID:        p.EmpresaID,
			ProductoID:       p.ID,
			CostoAnterior:    costoAnterior,
			CostoNuevo:       p.Costo,
			PrecioAnterior:   precioAnterior,
			PrecioNuevo:      p.Precio,
			ImpuestoAnterior: impuestoAnterior,
			ImpuestoNuevo:    p.ImpuestoPorcentaje,
			Motivo:           strings.TrimSpace(motivoCambio),
			Referencia:       strings.TrimSpace(referenciaCambio),
			UsuarioCreador:   strings.TrimSpace(p.UsuarioCreador),
			Estado:           "activo",
			Observaciones:    "historial automático por cambio de precio",
		}); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// DeleteProducto elimina un producto y su data de inventario asociada.
func DeleteProducto(dbConn *sql.DB, empresaID, productoID int64) error {
	tx, err := dbConn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("DELETE FROM inventario_movimientos WHERE empresa_id = ? AND producto_id = ?", empresaID, productoID); err != nil {
		return err
	}
	if _, err := tx.Exec("DELETE FROM inventario_existencias WHERE empresa_id = ? AND producto_id = ?", empresaID, productoID); err != nil {
		return err
	}
	if _, err := tx.Exec("DELETE FROM productos WHERE empresa_id = ? AND id = ?", empresaID, productoID); err != nil {
		return err
	}
	return tx.Commit()
}

// SetProductoEstado activa/desactiva un producto.
func SetProductoEstado(dbConn *sql.DB, empresaID, productoID int64, estado string) error {
	_, err := dbConn.Exec("UPDATE productos SET estado = ?, fecha_actualizacion = datetime('now','localtime') WHERE id = ? AND empresa_id = ?", strings.TrimSpace(estado), productoID, empresaID)
	return err
}

// UpdateProductoImagen actualiza la URL de imagen/logo de un producto.
func UpdateProductoImagen(dbConn *sql.DB, empresaID, productoID int64, imagenURL string) error {
	_, err := dbConn.Exec("UPDATE productos SET imagen_url = ?, fecha_actualizacion = datetime('now','localtime') WHERE id = ? AND empresa_id = ?", strings.TrimSpace(imagenURL), productoID, empresaID)
	return err
}

// GetExistenciasByEmpresa devuelve stock por bodega y producto.
func GetExistenciasByEmpresa(dbConn *sql.DB, empresaID, productoID, bodegaID int64, limit int, offset int) ([]InventarioExistencia, error) {
	if limit <= 0 || limit > 1000 {
		limit = 300
	}
	if offset < 0 {
		offset = 0
	}

	query := `SELECT e.id, e.empresa_id, e.producto_id, p.nombre, e.bodega_id, b.nombre, e.cantidad,
		e.fecha_creacion, e.fecha_actualizacion, e.usuario_creador, e.estado, e.observaciones
	FROM inventario_existencias e
	JOIN productos p ON p.id = e.producto_id AND p.empresa_id = e.empresa_id
	JOIN bodegas b ON b.id = e.bodega_id AND b.empresa_id = e.empresa_id
	WHERE e.empresa_id = ?`
	args := []interface{}{empresaID}
	if productoID > 0 {
		query += " AND e.producto_id = ?"
		args = append(args, productoID)
	}
	if bodegaID > 0 {
		query += " AND e.bodega_id = ?"
		args = append(args, bodegaID)
	}
	query += " ORDER BY p.nombre ASC, b.nombre ASC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]InventarioExistencia, 0)
	for rows.Next() {
		var e InventarioExistencia
		var fechaCre, fechaAct, usuario, estado, obs sql.NullString
		if err := rows.Scan(&e.ID, &e.EmpresaID, &e.ProductoID, &e.ProductoNombre, &e.BodegaID, &e.BodegaNombre, &e.Cantidad, &fechaCre, &fechaAct, &usuario, &estado, &obs); err != nil {
			return nil, err
		}
		if fechaCre.Valid {
			e.FechaCreacion = fechaCre.String
		}
		if fechaAct.Valid {
			e.FechaActualizacion = fechaAct.String
		}
		if usuario.Valid {
			e.UsuarioCreador = usuario.String
		}
		if estado.Valid {
			e.Estado = estado.String
		}
		if obs.Valid {
			e.Observaciones = obs.String
		}
		out = append(out, e)
	}
	return out, nil
}

// TransferirProductoEntreBodegas mueve unidades entre bodegas de una empresa y registra movimiento.
func TransferirProductoEntreBodegas(dbConn *sql.DB, empresaID, productoID, bodegaOrigenID, bodegaDestinoID int64, cantidad float64, referencia, usuario, observaciones string) error {
	if cantidad <= 0 {
		return fmt.Errorf("cantidad debe ser mayor a 0")
	}
	if bodegaOrigenID == bodegaDestinoID {
		return fmt.Errorf("la bodega origen y destino no pueden ser iguales")
	}

	tx, err := dbConn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := validateBodegaEmpresaTx(tx, empresaID, bodegaOrigenID); err != nil {
		return err
	}
	if err := validateBodegaEmpresaTx(tx, empresaID, bodegaDestinoID); err != nil {
		return err
	}

	var costoUnitario float64
	if err := tx.QueryRow("SELECT costo FROM productos WHERE empresa_id = ? AND id = ?", empresaID, productoID).Scan(&costoUnitario); err != nil {
		return err
	}

	var stockOrigen float64
	err = tx.QueryRow("SELECT cantidad FROM inventario_existencias WHERE empresa_id = ? AND producto_id = ? AND bodega_id = ?", empresaID, productoID, bodegaOrigenID).Scan(&stockOrigen)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrStockInsuficiente
		}
		return err
	}
	if stockOrigen < cantidad {
		return ErrStockInsuficiente
	}

	if _, err := tx.Exec(`UPDATE inventario_existencias
		SET cantidad = cantidad - ?, fecha_actualizacion = datetime('now','localtime')
		WHERE empresa_id = ? AND producto_id = ? AND bodega_id = ?`, cantidad, empresaID, productoID, bodegaOrigenID); err != nil {
		return err
	}

	if err := upsertExistenciaTx(tx, empresaID, productoID, bodegaDestinoID, cantidad, usuario, "traslado entre bodegas"); err != nil {
		return err
	}

	if err := insertMovimientoTx(tx, InventarioMovimiento{
		EmpresaID:       empresaID,
		ProductoID:      productoID,
		BodegaOrigenID:  bodegaOrigenID,
		BodegaDestinoID: bodegaDestinoID,
		Tipo:            "traslado",
		Cantidad:        cantidad,
		CostoUnitario:   costoUnitario,
		Referencia:      strings.TrimSpace(referencia),
		UsuarioCreador:  strings.TrimSpace(usuario),
		Estado:          "activo",
		Observaciones:   strings.TrimSpace(observaciones),
	}); err != nil {
		return err
	}

	return tx.Commit()
}

// GetMovimientosByEmpresa devuelve historial de movimientos de inventario.
func GetMovimientosByEmpresa(dbConn *sql.DB, empresaID, productoID int64, limit int, offset int) ([]InventarioMovimiento, error) {
	if limit <= 0 || limit > 1000 {
		limit = 300
	}
	if offset < 0 {
		offset = 0
	}

	query := `SELECT m.id, m.empresa_id, m.producto_id, p.nombre,
		m.bodega_origen_id, bo.nombre,
		m.bodega_destino_id, bd.nombre,
		m.tipo, m.cantidad, m.costo_unitario, m.referencia, m.fecha_movimiento,
		m.fecha_creacion, m.fecha_actualizacion, m.usuario_creador, m.estado, m.observaciones
	FROM inventario_movimientos m
	JOIN productos p ON p.id = m.producto_id AND p.empresa_id = m.empresa_id
	LEFT JOIN bodegas bo ON bo.id = m.bodega_origen_id AND bo.empresa_id = m.empresa_id
	LEFT JOIN bodegas bd ON bd.id = m.bodega_destino_id AND bd.empresa_id = m.empresa_id
	WHERE m.empresa_id = ?`
	args := []interface{}{empresaID}
	if productoID > 0 {
		query += " AND m.producto_id = ?"
		args = append(args, productoID)
	}
	query += " ORDER BY m.id DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]InventarioMovimiento, 0)
	for rows.Next() {
		var m InventarioMovimiento
		var bodegaOrigenID, bodegaDestinoID sql.NullInt64
		var bodegaOrigenNombre, bodegaDestinoNombre sql.NullString
		var referencia, fechaMov, fechaCre, fechaAct, usuario, estado, obs sql.NullString
		if err := rows.Scan(
			&m.ID, &m.EmpresaID, &m.ProductoID, &m.ProductoNombre,
			&bodegaOrigenID, &bodegaOrigenNombre,
			&bodegaDestinoID, &bodegaDestinoNombre,
			&m.Tipo, &m.Cantidad, &m.CostoUnitario, &referencia, &fechaMov,
			&fechaCre, &fechaAct, &usuario, &estado, &obs,
		); err != nil {
			return nil, err
		}
		if bodegaOrigenID.Valid {
			m.BodegaOrigenID = bodegaOrigenID.Int64
		}
		if bodegaDestinoID.Valid {
			m.BodegaDestinoID = bodegaDestinoID.Int64
		}
		if bodegaOrigenNombre.Valid {
			m.BodegaOrigenNombre = bodegaOrigenNombre.String
		}
		if bodegaDestinoNombre.Valid {
			m.BodegaDestinoNombre = bodegaDestinoNombre.String
		}
		if referencia.Valid {
			m.Referencia = referencia.String
		}
		if fechaMov.Valid {
			m.FechaMovimiento = fechaMov.String
		}
		if fechaCre.Valid {
			m.FechaCreacion = fechaCre.String
		}
		if fechaAct.Valid {
			m.FechaActualizacion = fechaAct.String
		}
		if usuario.Valid {
			m.UsuarioCreador = usuario.String
		}
		if estado.Valid {
			m.Estado = estado.String
		}
		if obs.Valid {
			m.Observaciones = obs.String
		}
		out = append(out, m)
	}

	return out, nil
}

// CreateProveedor inserta un proveedor para la empresa.
func CreateProveedor(dbConn *sql.DB, p Proveedor) (int64, error) {
	res, err := dbConn.Exec(`INSERT INTO proveedores (
		empresa_id, codigo, nombre, documento, contacto, telefono, email, direccion,
		usuario_creador, estado, observaciones, fecha_creacion, fecha_actualizacion
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, COALESCE(NULLIF(?, ''), 'activo'), ?, datetime('now','localtime'), datetime('now','localtime'))`,
		p.EmpresaID, strings.TrimSpace(p.Codigo), strings.TrimSpace(p.Nombre), strings.TrimSpace(p.Documento), strings.TrimSpace(p.Contacto), strings.TrimSpace(p.Telefono), strings.TrimSpace(p.Email), strings.TrimSpace(p.Direccion), strings.TrimSpace(p.UsuarioCreador), strings.TrimSpace(p.Estado), strings.TrimSpace(p.Observaciones))
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// GetProveedoresByEmpresa lista proveedores por empresa.
func GetProveedoresByEmpresa(dbConn *sql.DB, empresaID int64, incluirInactivos bool) ([]Proveedor, error) {
	query := `SELECT id, empresa_id, codigo, nombre, documento, contacto, telefono, email, direccion,
		fecha_creacion, fecha_actualizacion, usuario_creador, estado, observaciones
	FROM proveedores WHERE empresa_id = ?`
	args := []interface{}{empresaID}
	if !incluirInactivos {
		query += ` AND estado = 'activo'`
	}
	query += ` ORDER BY nombre ASC`

	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]Proveedor, 0)
	for rows.Next() {
		var p Proveedor
		var codigo, documento, contacto, telefono, email, direccion, fechaCre, fechaAct, usuario, estado, obs sql.NullString
		if err := rows.Scan(&p.ID, &p.EmpresaID, &codigo, &p.Nombre, &documento, &contacto, &telefono, &email, &direccion, &fechaCre, &fechaAct, &usuario, &estado, &obs); err != nil {
			return nil, err
		}
		if codigo.Valid {
			p.Codigo = codigo.String
		}
		if documento.Valid {
			p.Documento = documento.String
		}
		if contacto.Valid {
			p.Contacto = contacto.String
		}
		if telefono.Valid {
			p.Telefono = telefono.String
		}
		if email.Valid {
			p.Email = email.String
		}
		if direccion.Valid {
			p.Direccion = direccion.String
		}
		if fechaCre.Valid {
			p.FechaCreacion = fechaCre.String
		}
		if fechaAct.Valid {
			p.FechaActualizacion = fechaAct.String
		}
		if usuario.Valid {
			p.UsuarioCreador = usuario.String
		}
		if estado.Valid {
			p.Estado = estado.String
		}
		if obs.Valid {
			p.Observaciones = obs.String
		}
		out = append(out, p)
	}
	return out, nil
}

// UpdateProveedor actualiza proveedor.
func UpdateProveedor(dbConn *sql.DB, p Proveedor) error {
	_, err := dbConn.Exec(`UPDATE proveedores
		SET codigo = ?, nombre = ?, documento = ?, contacto = ?, telefono = ?, email = ?, direccion = ?, observaciones = ?, fecha_actualizacion = datetime('now','localtime')
		WHERE id = ? AND empresa_id = ?`,
		strings.TrimSpace(p.Codigo), strings.TrimSpace(p.Nombre), strings.TrimSpace(p.Documento), strings.TrimSpace(p.Contacto), strings.TrimSpace(p.Telefono), strings.TrimSpace(p.Email), strings.TrimSpace(p.Direccion), strings.TrimSpace(p.Observaciones), p.ID, p.EmpresaID)
	return err
}

// DeleteProveedor elimina proveedor.
func DeleteProveedor(dbConn *sql.DB, empresaID, proveedorID int64) error {
	_, err := dbConn.Exec(`DELETE FROM proveedores WHERE id = ? AND empresa_id = ?`, proveedorID, empresaID)
	return err
}

// SetProveedorEstado activa/desactiva proveedor.
func SetProveedorEstado(dbConn *sql.DB, empresaID, proveedorID int64, estado string) error {
	_, err := dbConn.Exec(`UPDATE proveedores SET estado = ?, fecha_actualizacion = datetime('now','localtime') WHERE id = ? AND empresa_id = ?`, strings.TrimSpace(estado), proveedorID, empresaID)
	return err
}

// CreateServicio crea un servicio comercial por empresa.
func CreateServicio(dbConn *sql.DB, s Servicio) (int64, error) {
	res, err := dbConn.Exec(`INSERT INTO servicios (
		empresa_id, codigo, nombre, descripcion, categoria, duracion_minutos, costo_referencial, precio, impuesto_porcentaje, imagen_url,
		usuario_creador, estado, observaciones, fecha_creacion, fecha_actualizacion
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, COALESCE(NULLIF(?, ''), 'activo'), ?, datetime('now','localtime'), datetime('now','localtime'))`,
		s.EmpresaID, strings.TrimSpace(s.Codigo), strings.TrimSpace(s.Nombre), strings.TrimSpace(s.Descripcion), strings.TrimSpace(s.Categoria), s.DuracionMinutos, s.CostoReferencial, s.Precio, s.ImpuestoPorcentaje, strings.TrimSpace(s.ImagenURL), strings.TrimSpace(s.UsuarioCreador), strings.TrimSpace(s.Estado), strings.TrimSpace(s.Observaciones))
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// GetServiciosByEmpresa lista servicios por empresa.
func GetServiciosByEmpresa(dbConn *sql.DB, empresaID int64, filtro, estado string, limit, offset int) ([]Servicio, error) {
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}

	query := `SELECT id, empresa_id, codigo, nombre, descripcion, categoria, duracion_minutos, costo_referencial, precio, impuesto_porcentaje, imagen_url,
		fecha_creacion, fecha_actualizacion, usuario_creador, estado, observaciones
	FROM servicios WHERE empresa_id = ?`
	args := []interface{}{empresaID}
	if strings.TrimSpace(estado) != "" {
		query += ` AND estado = ?`
		args = append(args, strings.TrimSpace(estado))
	}
	if strings.TrimSpace(filtro) != "" {
		like := "%" + strings.TrimSpace(filtro) + "%"
		query += ` AND (nombre LIKE ? OR codigo LIKE ? OR categoria LIKE ?)`
		args = append(args, like, like, like)
	}
	query += ` ORDER BY id DESC LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]Servicio, 0)
	for rows.Next() {
		var s Servicio
		var codigo, desc, categoria, imagen, fechaCre, fechaAct, usuario, estadoVal, obs sql.NullString
		if err := rows.Scan(&s.ID, &s.EmpresaID, &codigo, &s.Nombre, &desc, &categoria, &s.DuracionMinutos, &s.CostoReferencial, &s.Precio, &s.ImpuestoPorcentaje, &imagen, &fechaCre, &fechaAct, &usuario, &estadoVal, &obs); err != nil {
			return nil, err
		}
		if codigo.Valid {
			s.Codigo = codigo.String
		}
		if desc.Valid {
			s.Descripcion = desc.String
		}
		if categoria.Valid {
			s.Categoria = categoria.String
		}
		if imagen.Valid {
			s.ImagenURL = imagen.String
		}
		if fechaCre.Valid {
			s.FechaCreacion = fechaCre.String
		}
		if fechaAct.Valid {
			s.FechaActualizacion = fechaAct.String
		}
		if usuario.Valid {
			s.UsuarioCreador = usuario.String
		}
		if estadoVal.Valid {
			s.Estado = estadoVal.String
		}
		if obs.Valid {
			s.Observaciones = obs.String
		}
		out = append(out, s)
	}
	return out, nil
}

// UpdateServicio actualiza servicio.
func UpdateServicio(dbConn *sql.DB, s Servicio) error {
	_, err := dbConn.Exec(`UPDATE servicios
		SET codigo = ?, nombre = ?, descripcion = ?, categoria = ?, duracion_minutos = ?, costo_referencial = ?, precio = ?, impuesto_porcentaje = ?, imagen_url = ?, observaciones = ?, fecha_actualizacion = datetime('now','localtime')
		WHERE id = ? AND empresa_id = ?`,
		strings.TrimSpace(s.Codigo), strings.TrimSpace(s.Nombre), strings.TrimSpace(s.Descripcion), strings.TrimSpace(s.Categoria), s.DuracionMinutos, s.CostoReferencial, s.Precio, s.ImpuestoPorcentaje, strings.TrimSpace(s.ImagenURL), strings.TrimSpace(s.Observaciones), s.ID, s.EmpresaID)
	return err
}

// DeleteServicio elimina servicio.
func DeleteServicio(dbConn *sql.DB, empresaID, servicioID int64) error {
	_, err := dbConn.Exec(`DELETE FROM servicios WHERE id = ? AND empresa_id = ?`, servicioID, empresaID)
	return err
}

// SetServicioEstado activa/desactiva servicio.
func SetServicioEstado(dbConn *sql.DB, empresaID, servicioID int64, estado string) error {
	_, err := dbConn.Exec(`UPDATE servicios SET estado = ?, fecha_actualizacion = datetime('now','localtime') WHERE id = ? AND empresa_id = ?`, strings.TrimSpace(estado), servicioID, empresaID)
	return err
}

// GetProductoPrecioHistorialByEmpresa lista historial de precios de productos.
func GetProductoPrecioHistorialByEmpresa(dbConn *sql.DB, empresaID, productoID int64, limit, offset int) ([]ProductoPrecioHistorial, error) {
	if limit <= 0 || limit > 1000 {
		limit = 300
	}
	if offset < 0 {
		offset = 0
	}

	query := `SELECT h.id, h.empresa_id, h.producto_id, p.nombre,
		h.costo_anterior, h.costo_nuevo, h.precio_anterior, h.precio_nuevo, h.impuesto_anterior, h.impuesto_nuevo,
		h.motivo, h.referencia, h.fecha_cambio, h.fecha_creacion, h.fecha_actualizacion, h.usuario_creador, h.estado, h.observaciones
	FROM producto_precios_historial h
	JOIN productos p ON p.id = h.producto_id AND p.empresa_id = h.empresa_id
	WHERE h.empresa_id = ?`
	args := []interface{}{empresaID}
	if productoID > 0 {
		query += ` AND h.producto_id = ?`
		args = append(args, productoID)
	}
	query += ` ORDER BY h.id DESC LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]ProductoPrecioHistorial, 0)
	for rows.Next() {
		var h ProductoPrecioHistorial
		var motivo, referencia, fechaCambio, fechaCre, fechaAct, usuario, estadoVal, obs sql.NullString
		if err := rows.Scan(&h.ID, &h.EmpresaID, &h.ProductoID, &h.ProductoNombre, &h.CostoAnterior, &h.CostoNuevo, &h.PrecioAnterior, &h.PrecioNuevo, &h.ImpuestoAnterior, &h.ImpuestoNuevo, &motivo, &referencia, &fechaCambio, &fechaCre, &fechaAct, &usuario, &estadoVal, &obs); err != nil {
			return nil, err
		}
		if motivo.Valid {
			h.Motivo = motivo.String
		}
		if referencia.Valid {
			h.Referencia = referencia.String
		}
		if fechaCambio.Valid {
			h.FechaCambio = fechaCambio.String
		}
		if fechaCre.Valid {
			h.FechaCreacion = fechaCre.String
		}
		if fechaAct.Valid {
			h.FechaActualizacion = fechaAct.String
		}
		if usuario.Valid {
			h.UsuarioCreador = usuario.String
		}
		if estadoVal.Valid {
			h.Estado = estadoVal.String
		}
		if obs.Valid {
			h.Observaciones = obs.String
		}
		out = append(out, h)
	}
	return out, nil
}

// RegistrarMovimientoInventario registra entradas, salidas, devoluciones y pérdidas con impacto en stock.
func RegistrarMovimientoInventario(dbConn *sql.DB, empresaID, productoID, bodegaID int64, tipo string, cantidad float64, referencia, usuario, observaciones string) error {
	tipo = strings.ToLower(strings.TrimSpace(tipo))
	if cantidad <= 0 {
		return fmt.Errorf("cantidad debe ser mayor a 0")
	}
	if tipo == "" {
		return fmt.Errorf("tipo de movimiento requerido")
	}

	incoming := tipo == "entrada" || tipo == "devolucion" || tipo == "ajuste_entrada" || tipo == "ajuste_positivo"
	outgoing := tipo == "salida" || tipo == "perdida" || tipo == "ajuste_salida" || tipo == "ajuste_negativo"
	if !incoming && !outgoing {
		return fmt.Errorf("tipo de movimiento no soportado: %s", tipo)
	}

	tx, err := dbConn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := validateBodegaEmpresaTx(tx, empresaID, bodegaID); err != nil {
		return err
	}

	var costoUnitario float64
	if err := tx.QueryRow(`SELECT costo FROM productos WHERE empresa_id = ? AND id = ? LIMIT 1`, empresaID, productoID).Scan(&costoUnitario); err != nil {
		return err
	}

	if incoming {
		if err := upsertExistenciaTx(tx, empresaID, productoID, bodegaID, cantidad, usuario, observaciones); err != nil {
			return err
		}
		if err := insertMovimientoTx(tx, InventarioMovimiento{
			EmpresaID:       empresaID,
			ProductoID:      productoID,
			BodegaDestinoID: bodegaID,
			Tipo:            tipo,
			Cantidad:        cantidad,
			CostoUnitario:   costoUnitario,
			Referencia:      strings.TrimSpace(referencia),
			UsuarioCreador:  strings.TrimSpace(usuario),
			Estado:          "activo",
			Observaciones:   strings.TrimSpace(observaciones),
		}); err != nil {
			return err
		}
		return tx.Commit()
	}

	var stockActual float64
	err = tx.QueryRow(`SELECT cantidad FROM inventario_existencias WHERE empresa_id = ? AND producto_id = ? AND bodega_id = ?`, empresaID, productoID, bodegaID).Scan(&stockActual)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrStockInsuficiente
		}
		return err
	}
	if stockActual < cantidad {
		return ErrStockInsuficiente
	}

	if _, err := tx.Exec(`UPDATE inventario_existencias SET cantidad = cantidad - ?, fecha_actualizacion = datetime('now','localtime') WHERE empresa_id = ? AND producto_id = ? AND bodega_id = ?`, cantidad, empresaID, productoID, bodegaID); err != nil {
		return err
	}
	if err := insertMovimientoTx(tx, InventarioMovimiento{
		EmpresaID:      empresaID,
		ProductoID:     productoID,
		BodegaOrigenID: bodegaID,
		Tipo:           tipo,
		Cantidad:       cantidad,
		CostoUnitario:  costoUnitario,
		Referencia:     strings.TrimSpace(referencia),
		UsuarioCreador: strings.TrimSpace(usuario),
		Estado:         "activo",
		Observaciones:  strings.TrimSpace(observaciones),
	}); err != nil {
		return err
	}

	return tx.Commit()
}

// RegistrarCambioProducto registra cambio de un producto por otro afectando existencias y movimientos.
func RegistrarCambioProducto(dbConn *sql.DB, empresaID, productoOrigenID, productoDestinoID, bodegaID int64, cantidad float64, referencia, usuario, observaciones string) error {
	if cantidad <= 0 {
		return fmt.Errorf("cantidad debe ser mayor a 0")
	}
	if productoOrigenID <= 0 || productoDestinoID <= 0 {
		return fmt.Errorf("producto_origen_id y producto_destino_id son requeridos")
	}

	tx, err := dbConn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := validateBodegaEmpresaTx(tx, empresaID, bodegaID); err != nil {
		return err
	}

	var costoOrigen float64
	if err := tx.QueryRow(`SELECT costo FROM productos WHERE empresa_id = ? AND id = ? LIMIT 1`, empresaID, productoOrigenID).Scan(&costoOrigen); err != nil {
		return err
	}
	var costoDestino float64
	if err := tx.QueryRow(`SELECT costo FROM productos WHERE empresa_id = ? AND id = ? LIMIT 1`, empresaID, productoDestinoID).Scan(&costoDestino); err != nil {
		return err
	}

	var stockOrigen float64
	err = tx.QueryRow(`SELECT cantidad FROM inventario_existencias WHERE empresa_id = ? AND producto_id = ? AND bodega_id = ?`, empresaID, productoOrigenID, bodegaID).Scan(&stockOrigen)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrStockInsuficiente
		}
		return err
	}
	if stockOrigen < cantidad {
		return ErrStockInsuficiente
	}

	if _, err := tx.Exec(`UPDATE inventario_existencias SET cantidad = cantidad - ?, fecha_actualizacion = datetime('now','localtime') WHERE empresa_id = ? AND producto_id = ? AND bodega_id = ?`, cantidad, empresaID, productoOrigenID, bodegaID); err != nil {
		return err
	}
	if err := upsertExistenciaTx(tx, empresaID, productoDestinoID, bodegaID, cantidad, usuario, "cambio de producto"); err != nil {
		return err
	}

	if err := insertMovimientoTx(tx, InventarioMovimiento{
		EmpresaID:      empresaID,
		ProductoID:     productoOrigenID,
		BodegaOrigenID: bodegaID,
		Tipo:           "cambio_salida",
		Cantidad:       cantidad,
		CostoUnitario:  costoOrigen,
		Referencia:     strings.TrimSpace(referencia),
		UsuarioCreador: strings.TrimSpace(usuario),
		Estado:         "activo",
		Observaciones:  strings.TrimSpace(observaciones),
	}); err != nil {
		return err
	}
	if err := insertMovimientoTx(tx, InventarioMovimiento{
		EmpresaID:       empresaID,
		ProductoID:      productoDestinoID,
		BodegaDestinoID: bodegaID,
		Tipo:            "cambio_entrada",
		Cantidad:        cantidad,
		CostoUnitario:   costoDestino,
		Referencia:      strings.TrimSpace(referencia),
		UsuarioCreador:  strings.TrimSpace(usuario),
		Estado:          "activo",
		Observaciones:   strings.TrimSpace(observaciones),
	}); err != nil {
		return err
	}

	return tx.Commit()
}

func seedCategoriasProductosFromLegacy(dbConn *sql.DB) error {
	rows, err := dbConn.Query(`SELECT DISTINCT empresa_id, TRIM(categoria)
		FROM productos
		WHERE TRIM(COALESCE(categoria, '')) <> ''`)
	if err != nil {
		return err
	}
	type categoriaSeed struct {
		empresaID int64
		nombre    string
	}
	seeds := make([]categoriaSeed, 0)

	for rows.Next() {
		var empresaID int64
		var nombre string
		if err := rows.Scan(&empresaID, &nombre); err != nil {
			rows.Close()
			return err
		}
		nombre = strings.TrimSpace(nombre)
		if empresaID <= 0 || nombre == "" {
			continue
		}
		seeds = append(seeds, categoriaSeed{empresaID: empresaID, nombre: nombre})
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	if err := rows.Close(); err != nil {
		return err
	}

	for _, seed := range seeds {
		if _, err := dbConn.Exec(`INSERT OR IGNORE INTO categorias_productos (
			empresa_id, nombre, descripcion, usuario_creador, estado, observaciones, fecha_creacion, fecha_actualizacion
		) VALUES (?, ?, ?, 'migracion', 'activo', ?, datetime('now','localtime'), datetime('now','localtime'))`,
			seed.empresaID, seed.nombre, "categoria migrada desde productos.categoria", "registro automático por migración"); err != nil {
			return err
		}
	}
	return nil
}

func backfillProductoCategoriaIDs(dbConn *sql.DB) error {
	if _, err := dbConn.Exec(`UPDATE productos
		SET categoria_id = (
			SELECT cp.id
			FROM categorias_productos cp
			WHERE cp.empresa_id = productos.empresa_id
				AND LOWER(TRIM(cp.nombre)) = LOWER(TRIM(productos.categoria))
			LIMIT 1
		)
		WHERE (categoria_id IS NULL OR categoria_id <= 0)
			AND TRIM(COALESCE(categoria, '')) <> ''`); err != nil {
		return err
	}

	if _, err := dbConn.Exec(`UPDATE productos
		SET categoria = (
			SELECT cp.nombre
			FROM categorias_productos cp
			WHERE cp.empresa_id = productos.empresa_id
				AND cp.id = productos.categoria_id
			LIMIT 1
		)
		WHERE categoria_id IS NOT NULL AND categoria_id > 0`); err != nil {
		return err
	}
	return nil
}

func resolveCategoriaProductoTx(tx *sql.Tx, empresaID, categoriaID int64) (string, error) {
	var nombre string
	if err := tx.QueryRow(`SELECT nombre FROM categorias_productos WHERE empresa_id = ? AND id = ? LIMIT 1`, empresaID, categoriaID).Scan(&nombre); err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("categoria %d no pertenece a la empresa %d", categoriaID, empresaID)
		}
		return "", err
	}
	nombre = strings.TrimSpace(nombre)
	if nombre == "" {
		return "", fmt.Errorf("la categoría seleccionada no tiene nombre")
	}
	return nombre, nil
}

func insertProductoPrecioHistorialTx(tx *sql.Tx, h ProductoPrecioHistorial) error {
	_, err := tx.Exec(`INSERT INTO producto_precios_historial (
		empresa_id, producto_id, costo_anterior, costo_nuevo, precio_anterior, precio_nuevo, impuesto_anterior, impuesto_nuevo,
		motivo, referencia, fecha_cambio, fecha_creacion, fecha_actualizacion, usuario_creador, estado, observaciones
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now','localtime'), datetime('now','localtime'), datetime('now','localtime'), ?, COALESCE(NULLIF(?, ''), 'activo'), ?)`,
		h.EmpresaID, h.ProductoID, h.CostoAnterior, h.CostoNuevo, h.PrecioAnterior, h.PrecioNuevo, h.ImpuestoAnterior, h.ImpuestoNuevo,
		strings.TrimSpace(h.Motivo), strings.TrimSpace(h.Referencia), strings.TrimSpace(h.UsuarioCreador), strings.TrimSpace(h.Estado), strings.TrimSpace(h.Observaciones))
	return err
}

func validateBodegaEmpresaTx(tx *sql.Tx, empresaID, bodegaID int64) error {
	var exists int
	if err := tx.QueryRow("SELECT 1 FROM bodegas WHERE empresa_id = ? AND id = ? LIMIT 1", empresaID, bodegaID).Scan(&exists); err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("bodega %d no pertenece a la empresa %d", bodegaID, empresaID)
		}
		return err
	}
	return nil
}

func validateProveedorEmpresaTx(tx *sql.Tx, empresaID, proveedorID int64) error {
	var exists int
	if err := tx.QueryRow("SELECT 1 FROM proveedores WHERE empresa_id = ? AND id = ? LIMIT 1", empresaID, proveedorID).Scan(&exists); err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("proveedor %d no pertenece a la empresa %d", proveedorID, empresaID)
		}
		return err
	}
	return nil
}

func upsertExistenciaTx(tx *sql.Tx, empresaID, productoID, bodegaID int64, delta float64, usuario, observaciones string) error {
	res, err := tx.Exec(`UPDATE inventario_existencias
		SET cantidad = cantidad + ?, fecha_actualizacion = datetime('now','localtime')
		WHERE empresa_id = ? AND producto_id = ? AND bodega_id = ?`,
		delta, empresaID, productoID, bodegaID)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected > 0 {
		return nil
	}
	_, err = tx.Exec(`INSERT INTO inventario_existencias (
		empresa_id, producto_id, bodega_id, cantidad, usuario_creador, estado, observaciones, fecha_creacion, fecha_actualizacion
	) VALUES (?, ?, ?, ?, ?, 'activo', ?, datetime('now','localtime'), datetime('now','localtime'))`,
		empresaID, productoID, bodegaID, delta, strings.TrimSpace(usuario), strings.TrimSpace(observaciones))
	return err
}

func insertMovimientoTx(tx *sql.Tx, m InventarioMovimiento) error {
	_, err := tx.Exec(`INSERT INTO inventario_movimientos (
		empresa_id, producto_id, bodega_origen_id, bodega_destino_id, tipo, cantidad, costo_unitario, referencia, fecha_movimiento,
		usuario_creador, estado, observaciones, fecha_creacion, fecha_actualizacion
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, datetime('now','localtime'), ?, COALESCE(NULLIF(?, ''), 'activo'), ?, datetime('now','localtime'), datetime('now','localtime'))`,
		m.EmpresaID, m.ProductoID, nullableInt64(m.BodegaOrigenID), nullableInt64(m.BodegaDestinoID), strings.TrimSpace(m.Tipo), m.Cantidad, m.CostoUnitario, strings.TrimSpace(m.Referencia), strings.TrimSpace(m.UsuarioCreador), strings.TrimSpace(m.Estado), strings.TrimSpace(m.Observaciones))
	return err
}

func nullableInt64(v int64) interface{} {
	if v <= 0 {
		return nil
	}
	return v
}

func defaultUnidad(u string) string {
	u = strings.TrimSpace(u)
	if u == "" {
		return "unidad"
	}
	return u
}
