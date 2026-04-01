package db

import (
	"database/sql"
	"fmt"
	"math"
	"strings"
	"time"
)

// CarritoCompra representa un carrito de compra por empresa.
type CarritoCompra struct {
	ID                 int64   `json:"id"`
	EmpresaID          int64   `json:"empresa_id"`
	Codigo             string  `json:"codigo"`
	Nombre             string  `json:"nombre"`
	CanalVenta         string  `json:"canal_venta,omitempty"`
	ClienteID          int64   `json:"cliente_id,omitempty"`
	ClienteNombre      string  `json:"cliente_nombre,omitempty"`
	EstadoCarrito      string  `json:"estado_carrito,omitempty"`
	Moneda             string  `json:"moneda,omitempty"`
	ReferenciaExterna  string  `json:"referencia_externa,omitempty"`
	Subtotal           float64 `json:"subtotal"`
	DescuentoTotal     float64 `json:"descuento_total"`
	ImpuestoTotal      float64 `json:"impuesto_total"`
	Total              float64 `json:"total"`
	ActivadoEn         string  `json:"activado_en,omitempty"`
	PagadoEn           string  `json:"pagado_en,omitempty"`
	DescuentoTipo      string  `json:"descuento_tipo,omitempty"`
	DescuentoCodigo    string  `json:"descuento_codigo,omitempty"`
	DescuentoValor     float64 `json:"descuento_valor"`
	DevolucionTotal    float64 `json:"devolucion_total"`
	TotalPagado        float64 `json:"total_pagado"`
	ItemCount          int64   `json:"item_count"`
	FechaCreacion      string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string  `json:"usuario_creador,omitempty"`
	Estado             string  `json:"estado,omitempty"`
	Observaciones      string  `json:"observaciones,omitempty"`
}

// CarritoCompraItem representa un item dentro de un carrito de compra.
type CarritoCompraItem struct {
	ID                  int64   `json:"id"`
	EmpresaID           int64   `json:"empresa_id"`
	CarritoID           int64   `json:"carrito_id"`
	TipoItem            string  `json:"tipo_item,omitempty"`
	ReferenciaID        int64   `json:"referencia_id,omitempty"`
	CodigoItem          string  `json:"codigo_item,omitempty"`
	Descripcion         string  `json:"descripcion"`
	UnidadMedida        string  `json:"unidad_medida,omitempty"`
	Cantidad            float64 `json:"cantidad"`
	PrecioUnitario      float64 `json:"precio_unitario"`
	DescuentoPorcentaje float64 `json:"descuento_porcentaje"`
	ImpuestoPorcentaje  float64 `json:"impuesto_porcentaje"`
	ImpuestoCodigo      string  `json:"impuesto_codigo,omitempty"`
	BaseGravable        float64 `json:"base_gravable"`
	ValorDescuento      float64 `json:"valor_descuento"`
	ValorImpuesto       float64 `json:"valor_impuesto"`
	SubtotalLinea       float64 `json:"subtotal_linea"`
	TotalLinea          float64 `json:"total_linea"`
	FechaCreacion       string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion  string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador      string  `json:"usuario_creador,omitempty"`
	Estado              string  `json:"estado,omitempty"`
	Observaciones       string  `json:"observaciones,omitempty"`
}

// EnsureEmpresaCarritosSchema crea y migra tablas de carritos de compra en empresas.db.
func EnsureEmpresaCarritosSchema(dbConn *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS carritos_compras (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			codigo TEXT,
			nombre TEXT NOT NULL,
			canal_venta TEXT DEFAULT 'mostrador',
			cliente_id INTEGER,
			estado_carrito TEXT DEFAULT 'abierto',
			moneda TEXT DEFAULT 'COP',
			referencia_externa TEXT,
			subtotal REAL DEFAULT 0,
			descuento_total REAL DEFAULT 0,
			impuesto_total REAL DEFAULT 0,
			total REAL DEFAULT 0,
			activado_en TEXT,
			pagado_en TEXT,
			descuento_tipo TEXT,
			descuento_codigo TEXT,
			descuento_valor REAL DEFAULT 0,
			devolucion_total REAL DEFAULT 0,
			total_pagado REAL DEFAULT 0,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS carrito_compra_items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			carrito_id INTEGER NOT NULL,
			tipo_item TEXT DEFAULT 'producto',
			referencia_id INTEGER,
			codigo_item TEXT,
			descripcion TEXT NOT NULL,
			unidad_medida TEXT DEFAULT 'unidad',
			cantidad REAL NOT NULL DEFAULT 1,
			precio_unitario REAL NOT NULL DEFAULT 0,
			descuento_porcentaje REAL DEFAULT 0,
			impuesto_porcentaje REAL DEFAULT 0,
			impuesto_codigo TEXT DEFAULT 'IVA',
			base_gravable REAL DEFAULT 0,
			valor_descuento REAL DEFAULT 0,
			valor_impuesto REAL DEFAULT 0,
			subtotal_linea REAL DEFAULT 0,
			total_linea REAL DEFAULT 0,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_carritos_empresa_codigo ON carritos_compras(empresa_id, codigo);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_carritos_empresa_nombre ON carritos_compras(empresa_id, nombre);`,
		`CREATE INDEX IF NOT EXISTS ix_carritos_empresa_estado ON carritos_compras(empresa_id, estado, estado_carrito);`,
		`CREATE INDEX IF NOT EXISTS ix_carrito_items_empresa_carrito ON carrito_compra_items(empresa_id, carrito_id);`,
		`CREATE INDEX IF NOT EXISTS ix_carrito_items_empresa_referencia ON carrito_compra_items(empresa_id, referencia_id);`,
	}
	for _, stmt := range stmts {
		if _, err := dbConn.Exec(stmt); err != nil {
			return err
		}
	}

	if err := ensureColumnIfMissing(dbConn, "carritos_compras", "codigo", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "carritos_compras", "canal_venta", "TEXT DEFAULT 'mostrador'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "carritos_compras", "cliente_id", "INTEGER"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "carritos_compras", "estado_carrito", "TEXT DEFAULT 'abierto'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "carritos_compras", "moneda", "TEXT DEFAULT 'COP'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "carritos_compras", "referencia_externa", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "carritos_compras", "subtotal", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "carritos_compras", "descuento_total", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "carritos_compras", "impuesto_total", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "carritos_compras", "total", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "carritos_compras", "activado_en", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "carritos_compras", "pagado_en", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "carritos_compras", "descuento_tipo", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "carritos_compras", "descuento_codigo", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "carritos_compras", "descuento_valor", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "carritos_compras", "devolucion_total", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "carritos_compras", "total_pagado", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "carritos_compras", "fecha_actualizacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "carritos_compras", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "carritos_compras", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "carritos_compras", "observaciones", "TEXT"); err != nil {
		return err
	}

	if err := ensureColumnIfMissing(dbConn, "carrito_compra_items", "tipo_item", "TEXT DEFAULT 'producto'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "carrito_compra_items", "referencia_id", "INTEGER"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "carrito_compra_items", "codigo_item", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "carrito_compra_items", "unidad_medida", "TEXT DEFAULT 'unidad'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "carrito_compra_items", "descuento_porcentaje", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "carrito_compra_items", "impuesto_porcentaje", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "carrito_compra_items", "impuesto_codigo", "TEXT DEFAULT 'IVA'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "carrito_compra_items", "base_gravable", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "carrito_compra_items", "valor_descuento", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "carrito_compra_items", "valor_impuesto", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "carrito_compra_items", "subtotal_linea", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "carrito_compra_items", "total_linea", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "carrito_compra_items", "fecha_actualizacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "carrito_compra_items", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "carrito_compra_items", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "carrito_compra_items", "observaciones", "TEXT"); err != nil {
		return err
	}

	return nil
}

func nextCarritoCodigo() string {
	return fmt.Sprintf("CAR-%d", time.Now().UnixNano())
}

func defaultCanalVenta(v string) string {
	v = strings.TrimSpace(strings.ToLower(v))
	if v == "" {
		return "mostrador"
	}
	return v
}

func defaultMoneda(v string) string {
	v = strings.TrimSpace(strings.ToUpper(v))
	if v == "" {
		return "COP"
	}
	return v
}

func defaultEstadoCarrito(v string) string {
	v = strings.TrimSpace(strings.ToLower(v))
	if v == "" {
		return "abierto"
	}
	return v
}

func defaultTipoItem(v string) string {
	v = strings.TrimSpace(strings.ToLower(v))
	if v == "" {
		return "producto"
	}
	return v
}

func defaultUnidadCarrito(v string) string {
	v = strings.TrimSpace(strings.ToLower(v))
	if v == "" {
		return "unidad"
	}
	return v
}

func defaultImpuestoCodigo(v string) string {
	v = strings.TrimSpace(strings.ToUpper(v))
	if v == "" {
		return "IVA"
	}
	return v
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

func calcItemTotals(item *CarritoCompraItem) {
	if item.Cantidad <= 0 {
		item.Cantidad = 1
	}
	if item.PrecioUnitario < 0 {
		item.PrecioUnitario = 0
	}
	if item.DescuentoPorcentaje < 0 {
		item.DescuentoPorcentaje = 0
	}
	if item.DescuentoPorcentaje > 100 {
		item.DescuentoPorcentaje = 100
	}
	if item.ImpuestoPorcentaje < 0 {
		item.ImpuestoPorcentaje = 0
	}

	base := item.Cantidad * item.PrecioUnitario
	descuento := base * (item.DescuentoPorcentaje / 100)
	baseGravable := base - descuento
	if baseGravable < 0 {
		baseGravable = 0
	}
	impuesto := baseGravable * (item.ImpuestoPorcentaje / 100)
	total := baseGravable + impuesto

	item.BaseGravable = round2(baseGravable)
	item.ValorDescuento = round2(descuento)
	item.ValorImpuesto = round2(impuesto)
	item.SubtotalLinea = round2(baseGravable)
	item.TotalLinea = round2(total)
}

// CreateCarritoCompra crea un carrito por empresa.
func CreateCarritoCompra(dbConn *sql.DB, payload CarritoCompra) (int64, error) {
	if strings.TrimSpace(payload.Codigo) == "" {
		payload.Codigo = nextCarritoCodigo()
	}
	res, err := dbConn.Exec(`INSERT INTO carritos_compras (
		empresa_id,
		codigo,
		nombre,
		canal_venta,
		cliente_id,
		estado_carrito,
		moneda,
		referencia_externa,
		usuario_creador,
		estado,
		observaciones,
		activado_en,
		pagado_en,
		descuento_tipo,
		descuento_codigo,
		descuento_valor,
		devolucion_total,
		total_pagado,
		fecha_creacion,
		fecha_actualizacion,
		subtotal,
		descuento_total,
		impuesto_total,
		total
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 'activo', ?, NULL, NULL, '', '', 0, 0, 0, datetime('now','localtime'), datetime('now','localtime'), 0, 0, 0, 0)`,
		payload.EmpresaID,
		strings.TrimSpace(payload.Codigo),
		strings.TrimSpace(payload.Nombre),
		defaultCanalVenta(payload.CanalVenta),
		nullableInt64(payload.ClienteID),
		defaultEstadoCarrito(payload.EstadoCarrito),
		defaultMoneda(payload.Moneda),
		strings.TrimSpace(payload.ReferenciaExterna),
		strings.TrimSpace(payload.UsuarioCreador),
		strings.TrimSpace(payload.Observaciones),
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// GetCarritosCompraByEmpresa lista carritos por empresa.
func GetCarritosCompraByEmpresa(dbConn *sql.DB, empresaID int64, includeInactive bool, q string) ([]CarritoCompra, error) {
	query := `SELECT
		c.id,
		c.empresa_id,
		COALESCE(c.codigo, ''),
		COALESCE(c.nombre, ''),
		COALESCE(c.canal_venta, 'mostrador'),
		COALESCE(c.cliente_id, 0),
		COALESCE(cl.nombre_razon_social, ''),
		COALESCE(c.estado_carrito, 'abierto'),
		COALESCE(c.moneda, 'COP'),
		COALESCE(c.referencia_externa, ''),
		COALESCE(c.subtotal, 0),
		COALESCE(c.descuento_total, 0),
		COALESCE(c.impuesto_total, 0),
		COALESCE(c.total, 0),
		COALESCE(c.activado_en, ''),
		COALESCE(c.pagado_en, ''),
		COALESCE(c.descuento_tipo, ''),
		COALESCE(c.descuento_codigo, ''),
		COALESCE(c.descuento_valor, 0),
		COALESCE(c.devolucion_total, 0),
		COALESCE(c.total_pagado, 0),
		COALESCE(COUNT(i.id), 0),
		COALESCE(c.fecha_creacion, ''),
		COALESCE(c.fecha_actualizacion, ''),
		COALESCE(c.usuario_creador, ''),
		COALESCE(c.estado, 'activo'),
		COALESCE(c.observaciones, '')
	FROM carritos_compras c
	LEFT JOIN clientes cl ON cl.empresa_id = c.empresa_id AND cl.id = c.cliente_id
	LEFT JOIN carrito_compra_items i ON i.empresa_id = c.empresa_id AND i.carrito_id = c.id AND COALESCE(i.estado, 'activo') = 'activo'
	WHERE c.empresa_id = ?`
	args := []interface{}{empresaID}

	if !includeInactive {
		query += ` AND COALESCE(c.estado, 'activo') = 'activo'`
	}
	q = strings.TrimSpace(q)
	if q != "" {
		pat := "%" + q + "%"
		query += ` AND (
			lower(COALESCE(c.nombre, '')) LIKE lower(?) OR
			lower(COALESCE(c.codigo, '')) LIKE lower(?) OR
			lower(COALESCE(cl.nombre_razon_social, '')) LIKE lower(?)
		)`
		args = append(args, pat, pat, pat)
	}
	query += ` GROUP BY c.id ORDER BY c.id DESC`

	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]CarritoCompra, 0)
	for rows.Next() {
		var item CarritoCompra
		if err := rows.Scan(
			&item.ID,
			&item.EmpresaID,
			&item.Codigo,
			&item.Nombre,
			&item.CanalVenta,
			&item.ClienteID,
			&item.ClienteNombre,
			&item.EstadoCarrito,
			&item.Moneda,
			&item.ReferenciaExterna,
			&item.Subtotal,
			&item.DescuentoTotal,
			&item.ImpuestoTotal,
			&item.Total,
			&item.ActivadoEn,
			&item.PagadoEn,
			&item.DescuentoTipo,
			&item.DescuentoCodigo,
			&item.DescuentoValor,
			&item.DevolucionTotal,
			&item.TotalPagado,
			&item.ItemCount,
			&item.FechaCreacion,
			&item.FechaActualizacion,
			&item.UsuarioCreador,
			&item.Estado,
			&item.Observaciones,
		); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, nil
}

// GetCarritoCompraByID obtiene un carrito puntual por empresa.
func GetCarritoCompraByID(dbConn *sql.DB, empresaID, carritoID int64) (*CarritoCompra, error) {
	query := `SELECT
		id,
		empresa_id,
		COALESCE(codigo, ''),
		COALESCE(nombre, ''),
		COALESCE(canal_venta, 'mostrador'),
		COALESCE(cliente_id, 0),
		COALESCE(estado_carrito, 'abierto'),
		COALESCE(moneda, 'COP'),
		COALESCE(referencia_externa, ''),
		COALESCE(subtotal, 0),
		COALESCE(descuento_total, 0),
		COALESCE(impuesto_total, 0),
		COALESCE(total, 0),
		COALESCE(activado_en, ''),
		COALESCE(pagado_en, ''),
		COALESCE(descuento_tipo, ''),
		COALESCE(descuento_codigo, ''),
		COALESCE(descuento_valor, 0),
		COALESCE(devolucion_total, 0),
		COALESCE(total_pagado, 0),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM carritos_compras
	WHERE empresa_id = ? AND id = ?
	LIMIT 1`
	row := dbConn.QueryRow(query, empresaID, carritoID)

	var item CarritoCompra
	if err := row.Scan(
		&item.ID,
		&item.EmpresaID,
		&item.Codigo,
		&item.Nombre,
		&item.CanalVenta,
		&item.ClienteID,
		&item.EstadoCarrito,
		&item.Moneda,
		&item.ReferenciaExterna,
		&item.Subtotal,
		&item.DescuentoTotal,
		&item.ImpuestoTotal,
		&item.Total,
		&item.ActivadoEn,
		&item.PagadoEn,
		&item.DescuentoTipo,
		&item.DescuentoCodigo,
		&item.DescuentoValor,
		&item.DevolucionTotal,
		&item.TotalPagado,
		&item.FechaCreacion,
		&item.FechaActualizacion,
		&item.UsuarioCreador,
		&item.Estado,
		&item.Observaciones,
	); err != nil {
		return nil, err
	}
	return &item, nil
}

// UpdateCarritoCompra actualiza los campos principales del carrito.
func UpdateCarritoCompra(dbConn *sql.DB, payload CarritoCompra) error {
	_, err := dbConn.Exec(`UPDATE carritos_compras SET
		codigo = ?,
		nombre = ?,
		canal_venta = ?,
		cliente_id = ?,
		estado_carrito = ?,
		moneda = ?,
		referencia_externa = ?,
		observaciones = ?,
		fecha_actualizacion = datetime('now','localtime')
	WHERE id = ? AND empresa_id = ?`,
		strings.TrimSpace(payload.Codigo),
		strings.TrimSpace(payload.Nombre),
		defaultCanalVenta(payload.CanalVenta),
		nullableInt64(payload.ClienteID),
		defaultEstadoCarrito(payload.EstadoCarrito),
		defaultMoneda(payload.Moneda),
		strings.TrimSpace(payload.ReferenciaExterna),
		strings.TrimSpace(payload.Observaciones),
		payload.ID,
		payload.EmpresaID,
	)
	return err
}

// DeleteCarritoCompra elimina el carrito y sus items.
func DeleteCarritoCompra(dbConn *sql.DB, empresaID, carritoID int64) error {
	tx, err := dbConn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM carrito_compra_items WHERE empresa_id = ? AND carrito_id = ?`, empresaID, carritoID); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM carritos_compras WHERE empresa_id = ? AND id = ?`, empresaID, carritoID); err != nil {
		return err
	}

	return tx.Commit()
}

// SetCarritoCompraEstado activa o desactiva el registro del carrito.
func SetCarritoCompraEstado(dbConn *sql.DB, empresaID, carritoID int64, estado string) error {
	_, err := dbConn.Exec(`UPDATE carritos_compras SET estado = ?, fecha_actualizacion = datetime('now','localtime') WHERE empresa_id = ? AND id = ?`, strings.TrimSpace(estado), empresaID, carritoID)
	return err
}

// SetCarritoOperacionEstado cambia estado operativo del carrito (abierto/cerrado).
func SetCarritoOperacionEstado(dbConn *sql.DB, empresaID, carritoID int64, estadoCarrito string) error {
	_, err := dbConn.Exec(`UPDATE carritos_compras SET estado_carrito = ?, fecha_actualizacion = datetime('now','localtime') WHERE empresa_id = ? AND id = ?`, strings.TrimSpace(estadoCarrito), empresaID, carritoID)
	return err
}

// ActivateCarritoStationSession activa un carrito de estación y opcionalmente reinicia sus items.
func ActivateCarritoStationSession(dbConn *sql.DB, empresaID, carritoID int64, resetItems bool) error {
	tx, err := dbConn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if resetItems {
		if _, err := tx.Exec(`DELETE FROM carrito_compra_items WHERE empresa_id = ? AND carrito_id = ?`, empresaID, carritoID); err != nil {
			return err
		}
	}

	if _, err := tx.Exec(`UPDATE carritos_compras SET
		estado = 'activo',
		estado_carrito = 'abierto',
		activado_en = datetime('now','localtime'),
		pagado_en = NULL,
		descuento_tipo = '',
		descuento_codigo = '',
		descuento_valor = 0,
		devolucion_total = 0,
		total_pagado = 0,
		fecha_actualizacion = datetime('now','localtime')
	WHERE empresa_id = ? AND id = ?`, empresaID, carritoID); err != nil {
		return err
	}

	if err := recalculateCarritoTotalsTx(tx, empresaID, carritoID); err != nil {
		return err
	}

	return tx.Commit()
}

// PayCarritoStationSession marca un carrito como pagado/inactivo y guarda resumen de cobro.
func PayCarritoStationSession(dbConn *sql.DB, empresaID, carritoID int64, descuentoTipo, descuentoCodigo string, descuentoValor, devolucionTotal, totalPagado float64) error {
	if descuentoValor < 0 {
		descuentoValor = 0
	}
	if devolucionTotal < 0 {
		devolucionTotal = 0
	}
	if totalPagado < 0 {
		totalPagado = 0
	}

	_, err := dbConn.Exec(`UPDATE carritos_compras SET
		estado = 'inactivo',
		estado_carrito = 'cerrado',
		pagado_en = datetime('now','localtime'),
		descuento_tipo = ?,
		descuento_codigo = ?,
		descuento_valor = ?,
		devolucion_total = ?,
		total_pagado = ?,
		fecha_actualizacion = datetime('now','localtime')
	WHERE empresa_id = ? AND id = ?`,
		strings.TrimSpace(descuentoTipo),
		strings.TrimSpace(descuentoCodigo),
		round2(descuentoValor),
		round2(devolucionTotal),
		round2(totalPagado),
		empresaID,
		carritoID,
	)
	return err
}

// CreateCarritoCompraItem crea un item y recalcula totales del carrito.
func CreateCarritoCompraItem(dbConn *sql.DB, payload CarritoCompraItem) (int64, error) {
	payload.TipoItem = defaultTipoItem(payload.TipoItem)
	payload.UnidadMedida = defaultUnidadCarrito(payload.UnidadMedida)
	payload.ImpuestoCodigo = defaultImpuestoCodigo(payload.ImpuestoCodigo)
	calcItemTotals(&payload)

	tx, err := dbConn.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	if err := validateCarritoEnEmpresaTx(tx, payload.EmpresaID, payload.CarritoID); err != nil {
		return 0, err
	}

	res, err := tx.Exec(`INSERT INTO carrito_compra_items (
		empresa_id,
		carrito_id,
		tipo_item,
		referencia_id,
		codigo_item,
		descripcion,
		unidad_medida,
		cantidad,
		precio_unitario,
		descuento_porcentaje,
		impuesto_porcentaje,
		impuesto_codigo,
		base_gravable,
		valor_descuento,
		valor_impuesto,
		subtotal_linea,
		total_linea,
		usuario_creador,
		estado,
		observaciones,
		fecha_creacion,
		fecha_actualizacion
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'activo', ?, datetime('now','localtime'), datetime('now','localtime'))`,
		payload.EmpresaID,
		payload.CarritoID,
		payload.TipoItem,
		nullableInt64(payload.ReferenciaID),
		strings.TrimSpace(payload.CodigoItem),
		strings.TrimSpace(payload.Descripcion),
		payload.UnidadMedida,
		payload.Cantidad,
		payload.PrecioUnitario,
		payload.DescuentoPorcentaje,
		payload.ImpuestoPorcentaje,
		payload.ImpuestoCodigo,
		payload.BaseGravable,
		payload.ValorDescuento,
		payload.ValorImpuesto,
		payload.SubtotalLinea,
		payload.TotalLinea,
		strings.TrimSpace(payload.UsuarioCreador),
		strings.TrimSpace(payload.Observaciones),
	)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	if err := recalculateCarritoTotalsTx(tx, payload.EmpresaID, payload.CarritoID); err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return id, nil
}

// GetCarritoCompraItems lista items de un carrito.
func GetCarritoCompraItems(dbConn *sql.DB, empresaID, carritoID int64, includeInactive bool) ([]CarritoCompraItem, error) {
	query := `SELECT
		id,
		empresa_id,
		carrito_id,
		COALESCE(tipo_item, 'producto'),
		COALESCE(referencia_id, 0),
		COALESCE(codigo_item, ''),
		COALESCE(descripcion, ''),
		COALESCE(unidad_medida, 'unidad'),
		COALESCE(cantidad, 0),
		COALESCE(precio_unitario, 0),
		COALESCE(descuento_porcentaje, 0),
		COALESCE(impuesto_porcentaje, 0),
		COALESCE(impuesto_codigo, 'IVA'),
		COALESCE(base_gravable, 0),
		COALESCE(valor_descuento, 0),
		COALESCE(valor_impuesto, 0),
		COALESCE(subtotal_linea, 0),
		COALESCE(total_linea, 0),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM carrito_compra_items
	WHERE empresa_id = ? AND carrito_id = ?`
	args := []interface{}{empresaID, carritoID}
	if !includeInactive {
		query += ` AND COALESCE(estado, 'activo') = 'activo'`
	}
	query += ` ORDER BY id DESC`

	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]CarritoCompraItem, 0)
	for rows.Next() {
		var item CarritoCompraItem
		if err := rows.Scan(
			&item.ID,
			&item.EmpresaID,
			&item.CarritoID,
			&item.TipoItem,
			&item.ReferenciaID,
			&item.CodigoItem,
			&item.Descripcion,
			&item.UnidadMedida,
			&item.Cantidad,
			&item.PrecioUnitario,
			&item.DescuentoPorcentaje,
			&item.ImpuestoPorcentaje,
			&item.ImpuestoCodigo,
			&item.BaseGravable,
			&item.ValorDescuento,
			&item.ValorImpuesto,
			&item.SubtotalLinea,
			&item.TotalLinea,
			&item.FechaCreacion,
			&item.FechaActualizacion,
			&item.UsuarioCreador,
			&item.Estado,
			&item.Observaciones,
		); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, nil
}

// UpdateCarritoCompraItem actualiza un item del carrito y recalcula totales.
func UpdateCarritoCompraItem(dbConn *sql.DB, payload CarritoCompraItem) error {
	payload.TipoItem = defaultTipoItem(payload.TipoItem)
	payload.UnidadMedida = defaultUnidadCarrito(payload.UnidadMedida)
	payload.ImpuestoCodigo = defaultImpuestoCodigo(payload.ImpuestoCodigo)
	calcItemTotals(&payload)

	tx, err := dbConn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`UPDATE carrito_compra_items SET
		tipo_item = ?,
		referencia_id = ?,
		codigo_item = ?,
		descripcion = ?,
		unidad_medida = ?,
		cantidad = ?,
		precio_unitario = ?,
		descuento_porcentaje = ?,
		impuesto_porcentaje = ?,
		impuesto_codigo = ?,
		base_gravable = ?,
		valor_descuento = ?,
		valor_impuesto = ?,
		subtotal_linea = ?,
		total_linea = ?,
		observaciones = ?,
		fecha_actualizacion = datetime('now','localtime')
	WHERE id = ? AND empresa_id = ? AND carrito_id = ?`,
		payload.TipoItem,
		nullableInt64(payload.ReferenciaID),
		strings.TrimSpace(payload.CodigoItem),
		strings.TrimSpace(payload.Descripcion),
		payload.UnidadMedida,
		payload.Cantidad,
		payload.PrecioUnitario,
		payload.DescuentoPorcentaje,
		payload.ImpuestoPorcentaje,
		payload.ImpuestoCodigo,
		payload.BaseGravable,
		payload.ValorDescuento,
		payload.ValorImpuesto,
		payload.SubtotalLinea,
		payload.TotalLinea,
		strings.TrimSpace(payload.Observaciones),
		payload.ID,
		payload.EmpresaID,
		payload.CarritoID,
	); err != nil {
		return err
	}

	if err := recalculateCarritoTotalsTx(tx, payload.EmpresaID, payload.CarritoID); err != nil {
		return err
	}
	return tx.Commit()
}

// DeleteCarritoCompraItem elimina item y recalcula totales.
func DeleteCarritoCompraItem(dbConn *sql.DB, empresaID, carritoID, itemID int64) error {
	tx, err := dbConn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM carrito_compra_items WHERE empresa_id = ? AND carrito_id = ? AND id = ?`, empresaID, carritoID, itemID); err != nil {
		return err
	}
	if err := recalculateCarritoTotalsTx(tx, empresaID, carritoID); err != nil {
		return err
	}
	return tx.Commit()
}

// SetCarritoCompraItemEstado activa/desactiva item y recalcula totales.
func SetCarritoCompraItemEstado(dbConn *sql.DB, empresaID, carritoID, itemID int64, estado string) error {
	tx, err := dbConn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`UPDATE carrito_compra_items SET estado = ?, fecha_actualizacion = datetime('now','localtime') WHERE empresa_id = ? AND carrito_id = ? AND id = ?`, strings.TrimSpace(estado), empresaID, carritoID, itemID); err != nil {
		return err
	}
	if err := recalculateCarritoTotalsTx(tx, empresaID, carritoID); err != nil {
		return err
	}
	return tx.Commit()
}

// RecalculateCarritoCompraTotals recalcula totales del carrito basado en items activos.
func RecalculateCarritoCompraTotals(dbConn *sql.DB, empresaID, carritoID int64) error {
	tx, err := dbConn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := recalculateCarritoTotalsTx(tx, empresaID, carritoID); err != nil {
		return err
	}
	return tx.Commit()
}

func recalculateCarritoTotalsTx(tx *sql.Tx, empresaID, carritoID int64) error {
	var subtotal float64
	var descuento float64
	var impuesto float64
	var total float64
	if err := tx.QueryRow(`SELECT
		COALESCE(SUM(subtotal_linea), 0),
		COALESCE(SUM(valor_descuento), 0),
		COALESCE(SUM(valor_impuesto), 0),
		COALESCE(SUM(total_linea), 0)
	FROM carrito_compra_items
	WHERE empresa_id = ? AND carrito_id = ? AND COALESCE(estado, 'activo') = 'activo'`, empresaID, carritoID).Scan(&subtotal, &descuento, &impuesto, &total); err != nil {
		return err
	}

	_, err := tx.Exec(`UPDATE carritos_compras SET
		subtotal = ?,
		descuento_total = ?,
		impuesto_total = ?,
		total = ?,
		fecha_actualizacion = datetime('now','localtime')
	WHERE empresa_id = ? AND id = ?`, round2(subtotal), round2(descuento), round2(impuesto), round2(total), empresaID, carritoID)
	return err
}

func validateCarritoEnEmpresaTx(tx *sql.Tx, empresaID, carritoID int64) error {
	var count int64
	if err := tx.QueryRow(`SELECT COUNT(1) FROM carritos_compras WHERE empresa_id = ? AND id = ?`, empresaID, carritoID).Scan(&count); err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("carrito no existe para la empresa")
	}
	return nil
}
