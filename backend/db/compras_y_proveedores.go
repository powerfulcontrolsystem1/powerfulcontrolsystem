package db

import (
	"database/sql"
	"fmt"
)

// Proveedor representa una entidad externa que suministra productos o servicios.
type EmpresaProveedor struct {
	ID                 int64  `json:"id"`
	EmpresaID          int64  `json:"empresa_id"`
	NIT                string `json:"nit,omitempty"`
	NombreComercial    string `json:"nombre_comercial"`
	RazonSocial        string `json:"razon_social,omitempty"`
	Direccion          string `json:"direccion,omitempty"`
	Telefono           string `json:"telefono,omitempty"`
	Email              string `json:"email,omitempty"`
	CuentaBancaria     string `json:"cuenta_bancaria,omitempty"`
	PlazoDiasPago      int    `json:"plazo_dias_pago"`
	FechaCreacion      string `json:"fecha_creacion,omitempty"`
	FechaActualizacion string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string `json:"usuario_creador,omitempty"`
	Estado             string `json:"estado"`
	Observaciones      string `json:"observaciones,omitempty"`
}

// OrdenCompra representa un documento de requisición enviado a un proveedor.
type OrdenCompra struct {
	ID                 int64   `json:"id"`
	EmpresaID          int64   `json:"empresa_id"`
	ProveedorID        int64   `json:"proveedor_id"`
	BodegaDestinoID    int64   `json:"bodega_destino_id"` // 0 si no aplica
	NumeroOrden        string  `json:"numero_orden"`
	ReferenciaExterna  string  `json:"referencia_externa,omitempty"`
	Moneda             string  `json:"moneda"`
	Total              float64 `json:"total"`
	TotalImpuestos     float64 `json:"total_impuestos"`
	EstadoOrden        string  `json:"estado_orden"` // borrador, emitida, aprobada, recibida, cancelada
	EstadoPago         string  `json:"estado_pago"`  // pendiente, parcial, pagado
	FechaEmision       string  `json:"fecha_emision,omitempty"`
	FechaEsperada      string  `json:"fecha_esperada,omitempty"`
	FechaCreacion      string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string  `json:"usuario_creador,omitempty"`
	Estado             string  `json:"estado"`
	Observaciones      string  `json:"observaciones,omitempty"`
}

// OrdenCompraItem representa una línea dentro de la Orden de Compra.
type OrdenCompraItem struct {
	ID                 int64   `json:"id"`
	EmpresaID          int64   `json:"empresa_id"`
	OrdenCompraID      int64   `json:"orden_compra_id"`
	ProductoID         int64   `json:"producto_id"` // ID en empresa_productos
	Descripcion        string  `json:"descripcion"`
	UnidadMedida       string  `json:"unidad_medida,omitempty"`
	Cantidad           float64 `json:"cantidad"`
	CantidadRecibida   float64 `json:"cantidad_recibida"`
	PrecioUnitario     float64 `json:"precio_unitario"`
	ImpuestoPorcentaje float64 `json:"impuesto_porcentaje"`
	Subtotal           float64 `json:"subtotal"`
	Total              float64 `json:"total"`
	FechaCreacion      string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion string  `json:"fecha_actualizacion,omitempty"`
	Estado             string  `json:"estado"`
	Observaciones      string  `json:"observaciones,omitempty"`
}

// RecepcionInventario documenta el ingreso de mercancia (afecta stock) derivado o no de una OC.
type RecepcionInventario struct {
	ID                 int64   `json:"id"`
	EmpresaID          int64   `json:"empresa_id"`
	OrdenCompraID      int64   `json:"orden_compra_id"` // 0 si es ingreso libre sin orden
	ProveedorID        int64   `json:"proveedor_id"`
	BodegaID           int64   `json:"bodega_id"`
	NumeroFactura      string  `json:"numero_factura,omitempty"`
	TotalRecepcion     float64 `json:"total_recepcion"`
	FechaRecepcion     string  `json:"fecha_recepcion,omitempty"`
	FechaCreacion      string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string  `json:"usuario_creador,omitempty"`
	Estado             string  `json:"estado"`
	Observaciones      string  `json:"observaciones,omitempty"`
}

// EnsureEmpresasComprasSchema crea o actualiza las tablas del módulo de Compras / Proveedores.
func EnsureEmpresasComprasSchema(dbConn *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_proveedores (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			nit TEXT,
			nombre_comercial TEXT NOT NULL,
			razon_social TEXT,
			direccion TEXT,
			telefono TEXT,
			email TEXT,
			cuenta_bancaria TEXT,
			plazo_dias_pago INTEGER DEFAULT 0,
			fecha_creacion TEXT DEFAULT CURRENT_TIMESTAMP,
			fecha_actualizacion TEXT DEFAULT CURRENT_TIMESTAMP,
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS empresa_ordenes_compra (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			proveedor_id INTEGER NOT NULL,
			bodega_destino_id INTEGER DEFAULT 0,
			numero_orden TEXT NOT NULL,
			referencia_externa TEXT,
			moneda TEXT DEFAULT 'COP',
			total REAL DEFAULT 0,
			total_impuestos REAL DEFAULT 0,
			estado_orden TEXT DEFAULT 'borrador',
			estado_pago TEXT DEFAULT 'pendiente',
			fecha_emision TEXT,
			fecha_esperada TEXT,
			fecha_creacion TEXT DEFAULT CURRENT_TIMESTAMP,
			fecha_actualizacion TEXT DEFAULT CURRENT_TIMESTAMP,
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS empresa_ordenes_compra_items (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			orden_compra_id INTEGER NOT NULL,
			producto_id INTEGER NOT NULL,
			descripcion TEXT NOT NULL,
			unidad_medida TEXT,
			cantidad REAL DEFAULT 0,
			cantidad_recibida REAL DEFAULT 0,
			precio_unitario REAL DEFAULT 0,
			impuesto_porcentaje REAL DEFAULT 0,
			subtotal REAL DEFAULT 0,
			total REAL DEFAULT 0,
			fecha_creacion TEXT DEFAULT CURRENT_TIMESTAMP,
			fecha_actualizacion TEXT DEFAULT CURRENT_TIMESTAMP,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS empresa_compras_recepciones (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			orden_compra_id INTEGER DEFAULT 0,
			proveedor_id INTEGER DEFAULT 0,
			bodega_id INTEGER DEFAULT 0,
			numero_factura TEXT,
			total_recepcion REAL DEFAULT 0,
			fecha_recepcion TEXT DEFAULT CURRENT_TIMESTAMP,
			fecha_creacion TEXT DEFAULT CURRENT_TIMESTAMP,
			fecha_actualizacion TEXT DEFAULT CURRENT_TIMESTAMP,
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		)`,
	}

	for _, stmt := range stmts {
		if _, err := dbConn.Exec(stmt); err != nil {
			return fmt.Errorf("error asegurando schema de compras (CREATE): %w", err)
		}
	}

	// Evolución - usando el método del repo ensureColumnIfMissing de db.go / utils locales
	// (como usualmente se hace en cada módulo para agregar campos futuros).
	columns := []struct{ table, col, def string }{
		{"empresa_proveedores", "cuenta_bancaria", "TEXT"},
		{"empresa_proveedores", "plazo_dias_pago", "INTEGER DEFAULT 0"},
		{"empresa_ordenes_compra", "bodega_destino_id", "INTEGER DEFAULT 0"},
		{"empresa_ordenes_compra_items", "impuesto_porcentaje", "REAL DEFAULT 0"},
		{"empresa_compras_recepciones", "proveedor_id", "INTEGER DEFAULT 0"},
	}

	for _, c := range columns {
		if err := ensureColumnIfMissing(dbConn, c.table, c.col, c.def); err != nil {
			return fmt.Errorf("error en %s columna %s: %w", c.table, c.col, err)
		}
	}

	return nil
}

func CreateEmpresaProveedor(dbConn *sql.DB, p EmpresaProveedor) (int64, error) {
	stmt := `INSERT INTO empresa_proveedores (empresa_id, nit, nombre_comercial, razon_social, direccion, telefono, email, cuenta_bancaria, plazo_dias_pago, usuario_creador) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	res, err := dbConn.Exec(stmt, p.EmpresaID, p.NIT, p.NombreComercial, p.RazonSocial, p.Direccion, p.Telefono, p.Email, p.CuentaBancaria, p.PlazoDiasPago, p.UsuarioCreador)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func UpdateEmpresaProveedor(dbConn *sql.DB, p EmpresaProveedor) error {
	stmt := `UPDATE empresa_proveedores SET nit=?, nombre_comercial=?, razon_social=?, direccion=?, telefono=?, email=?, cuenta_bancaria=?, plazo_dias_pago=?, fecha_actualizacion=pcs_ts('now', 'localtime'), observaciones=? WHERE id=? AND empresa_id=?`
	_, err := dbConn.Exec(stmt, p.NIT, p.NombreComercial, p.RazonSocial, p.Direccion, p.Telefono, p.Email, p.CuentaBancaria, p.PlazoDiasPago, p.Observaciones, p.ID, p.EmpresaID)
	return err
}

func SetEstadoEmpresaProveedor(dbConn *sql.DB, id, empresaID int64, estado string) error {
	stmt := `UPDATE empresa_proveedores SET estado=?, fecha_actualizacion=pcs_ts('now', 'localtime') WHERE id=? AND empresa_id=?`
	_, err := dbConn.Exec(stmt, estado, id, empresaID)
	return err
}

func GetEmpresaProveedores(dbConn *sql.DB, empresaID int64, paramEstado string) ([]EmpresaProveedor, error) {
	var query string
	var rows *sql.Rows
	var err error
	if paramEstado != "" {
		query = `SELECT id, empresa_id, COALESCE(nit,''), nombre_comercial, COALESCE(razon_social,''), COALESCE(direccion,''), COALESCE(telefono,''), COALESCE(email,''), COALESCE(cuenta_bancaria,''), plazo_dias_pago, fecha_creacion, COALESCE(fecha_actualizacion,''), COALESCE(usuario_creador,''), estado, COALESCE(observaciones,'') FROM empresa_proveedores WHERE empresa_id=? AND estado=? ORDER BY nombre_comercial ASC`
		rows, err = dbConn.Query(query, empresaID, paramEstado)
	} else {
		query = `SELECT id, empresa_id, COALESCE(nit,''), nombre_comercial, COALESCE(razon_social,''), COALESCE(direccion,''), COALESCE(telefono,''), COALESCE(email,''), COALESCE(cuenta_bancaria,''), plazo_dias_pago, fecha_creacion, COALESCE(fecha_actualizacion,''), COALESCE(usuario_creador,''), estado, COALESCE(observaciones,'') FROM empresa_proveedores WHERE empresa_id=? ORDER BY nombre_comercial ASC`
		rows, err = dbConn.Query(query, empresaID)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lista = make([]EmpresaProveedor, 0)
	for rows.Next() {
		var p EmpresaProveedor
		err := rows.Scan(&p.ID, &p.EmpresaID, &p.NIT, &p.NombreComercial, &p.RazonSocial, &p.Direccion, &p.Telefono, &p.Email, &p.CuentaBancaria, &p.PlazoDiasPago, &p.FechaCreacion, &p.FechaActualizacion, &p.UsuarioCreador, &p.Estado, &p.Observaciones)
		if err != nil {
			return nil, err
		}
		lista = append(lista, p)
	}
	return lista, nil
}
