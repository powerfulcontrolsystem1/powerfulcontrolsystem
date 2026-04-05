package db

import (
	"database/sql"
	"fmt"
	"math"
	"sort"
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
	EstadoVenta        string  `json:"estado_venta,omitempty"`
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
	MetodoPago         string  `json:"metodo_pago,omitempty"`
	ReferenciaPago     string  `json:"referencia_pago,omitempty"`
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
			metodo_pago TEXT DEFAULT 'efectivo',
			referencia_pago TEXT,
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
	if err := ensureColumnIfMissing(dbConn, "carritos_compras", "metodo_pago", "TEXT DEFAULT 'efectivo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "carritos_compras", "referencia_pago", "TEXT"); err != nil {
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

// NormalizeMetodoPagoCarrito normaliza metodos de pago aceptados en el flujo de carrito.
func NormalizeMetodoPagoCarrito(v string) string {
	normalized := strings.TrimSpace(strings.ToLower(v))
	normalized = strings.ReplaceAll(normalized, "-", "_")
	normalized = strings.ReplaceAll(normalized, " ", "_")
	switch normalized {
	case "", "efectivo", "cash":
		return "efectivo"
	case "tarjeta_credito", "credito":
		return "tarjeta_credito"
	case "tarjeta_debito", "debito", "debito_tarjeta":
		return "tarjeta_debito"
	case "transferencia", "transferencia_bancaria", "bank_transfer":
		return "transferencia_bancaria"
	case "codigo_descuento", "descuento", "codigo":
		return "codigo_descuento"
	case "mixto", "mixed", "pago_mixto":
		return "mixto"
	default:
		return ""
	}
}

// IsMetodoPagoCarritoValido valida si el metodo de pago pertenece al catalogo permitido.
func IsMetodoPagoCarritoValido(v string) bool {
	return NormalizeMetodoPagoCarrito(v) != ""
}

func resolveCarritoEstadoVenta(estadoCarrito, estadoRegistro, pagadoEn string) string {
	estadoOp := strings.TrimSpace(strings.ToLower(estadoCarrito))
	if estadoOp == "" {
		estadoOp = "abierto"
	}
	estadoReg := strings.TrimSpace(strings.ToLower(estadoRegistro))
	if estadoReg == "" {
		estadoReg = "activo"
	}
	if estadoOp == "cerrado" && strings.TrimSpace(pagadoEn) != "" {
		return "venta_pagada"
	}
	if estadoOp == "cerrado" {
		return "venta_cerrada"
	}
	if estadoReg == "inactivo" {
		return "venta_suspendida"
	}
	return "venta_abierta"
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
	metodoPago := NormalizeMetodoPagoCarrito(payload.MetodoPago)
	if metodoPago == "" {
		metodoPago = "efectivo"
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
		metodo_pago,
		referencia_pago,
		fecha_creacion,
		fecha_actualizacion,
		subtotal,
		descuento_total,
		impuesto_total,
		total
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 'activo', ?, NULL, NULL, '', '', 0, 0, 0, ?, ?, datetime('now','localtime'), datetime('now','localtime'), 0, 0, 0, 0)`,
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
		metodoPago,
		strings.TrimSpace(payload.ReferenciaPago),
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
		COALESCE(c.metodo_pago, 'efectivo'),
		COALESCE(c.referencia_pago, ''),
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
			&item.MetodoPago,
			&item.ReferenciaPago,
			&item.ItemCount,
			&item.FechaCreacion,
			&item.FechaActualizacion,
			&item.UsuarioCreador,
			&item.Estado,
			&item.Observaciones,
		); err != nil {
			return nil, err
		}
		item.EstadoVenta = resolveCarritoEstadoVenta(item.EstadoCarrito, item.Estado, item.PagadoEn)
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
		COALESCE(metodo_pago, 'efectivo'),
		COALESCE(referencia_pago, ''),
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
		&item.MetodoPago,
		&item.ReferenciaPago,
		&item.FechaCreacion,
		&item.FechaActualizacion,
		&item.UsuarioCreador,
		&item.Estado,
		&item.Observaciones,
	); err != nil {
		return nil, err
	}
	item.EstadoVenta = resolveCarritoEstadoVenta(item.EstadoCarrito, item.Estado, item.PagadoEn)
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

	estadoCarrito, err := getCarritoEstadoTx(tx, empresaID, carritoID)
	if err != nil {
		return err
	}
	if !isCarritoCerrado(estadoCarrito) {
		if err := restoreCarritoItemsStockTx(tx, empresaID, carritoID, "eliminacion_carrito"); err != nil {
			return err
		}
	}

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

	estadoPrevio, err := getCarritoEstadoTx(tx, empresaID, carritoID)
	if err != nil {
		return err
	}

	if resetItems {
		if !isCarritoCerrado(estadoPrevio) {
			if err := restoreCarritoItemsStockTx(tx, empresaID, carritoID, "reset_estacion"); err != nil {
				return err
			}
		}
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
		metodo_pago = 'efectivo',
		referencia_pago = '',
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
func PayCarritoStationSession(dbConn *sql.DB, empresaID, carritoID int64, metodoPago, referenciaPago, descuentoTipo, descuentoCodigo string, descuentoValor, devolucionTotal, totalPagado float64, codigoDescuentoID int64) error {
	metodoPago = NormalizeMetodoPagoCarrito(metodoPago)
	if metodoPago == "" {
		return fmt.Errorf("metodo_pago invalido")
	}
	if descuentoValor < 0 {
		descuentoValor = 0
	}
	if devolucionTotal < 0 {
		devolucionTotal = 0
	}
	if totalPagado < 0 {
		totalPagado = 0
	}

	tx, err := dbConn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if codigoDescuentoID > 0 {
		if err := markCodigoDescuentoUsoTx(tx, empresaID, codigoDescuentoID); err != nil {
			return err
		}
	}

	_, err = tx.Exec(`UPDATE carritos_compras SET
		estado = 'inactivo',
		estado_carrito = 'cerrado',
		pagado_en = datetime('now','localtime'),
		metodo_pago = ?,
		referencia_pago = ?,
		descuento_tipo = ?,
		descuento_codigo = ?,
		descuento_valor = ?,
		devolucion_total = ?,
		total_pagado = ?,
		fecha_actualizacion = datetime('now','localtime')
	WHERE empresa_id = ? AND id = ?`,
		metodoPago,
		strings.TrimSpace(referenciaPago),
		strings.TrimSpace(descuentoTipo),
		strings.TrimSpace(descuentoCodigo),
		round2(descuentoValor),
		round2(devolucionTotal),
		round2(totalPagado),
		empresaID,
		carritoID,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// CreateCarritoCompraItem crea un item y recalcula totales del carrito.
func CreateCarritoCompraItem(dbConn *sql.DB, payload CarritoCompraItem) (int64, error) {
	payload.TipoItem = defaultTipoItem(payload.TipoItem)
	payload.UnidadMedida = defaultUnidadCarrito(payload.UnidadMedida)
	payload.ImpuestoCodigo = defaultImpuestoCodigo(payload.ImpuestoCodigo)
	payload.Estado = strings.TrimSpace(payload.Estado)
	if payload.Estado == "" {
		payload.Estado = "activo"
	}
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

	if isItemActivo(payload.Estado) {
		referencia := fmt.Sprintf("carrito:%d:item:%d", payload.CarritoID, id)
		if err := adjustCarritoItemStockTx(
			tx,
			payload.EmpresaID,
			payload.CarritoID,
			payload.TipoItem,
			payload.ReferenciaID,
			payload.Cantidad,
			true,
			referencia,
			payload.UsuarioCreador,
			"reserva por adicion al carrito",
		); err != nil {
			return 0, err
		}
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

	prev, err := getCarritoItemSnapshotTx(tx, payload.EmpresaID, payload.CarritoID, payload.ID)
	if err != nil {
		return err
	}

	if isItemActivo(prev.Estado) {
		if isTrackableProduct(prev.TipoItem, prev.ReferenciaID) && isTrackableProduct(payload.TipoItem, payload.ReferenciaID) && prev.ReferenciaID == payload.ReferenciaID {
			delta := payload.Cantidad - prev.Cantidad
			if delta > 0 {
				referencia := fmt.Sprintf("carrito:%d:item:%d", payload.CarritoID, payload.ID)
				if err := adjustCarritoItemStockTx(
					tx,
					payload.EmpresaID,
					payload.CarritoID,
					payload.TipoItem,
					payload.ReferenciaID,
					delta,
					true,
					referencia,
					payload.UsuarioCreador,
					"reserva adicional por actualizacion de item",
				); err != nil {
					return err
				}
			}
			if delta < 0 {
				referencia := fmt.Sprintf("carrito:%d:item:%d", payload.CarritoID, payload.ID)
				if err := adjustCarritoItemStockTx(
					tx,
					payload.EmpresaID,
					payload.CarritoID,
					payload.TipoItem,
					payload.ReferenciaID,
					-delta,
					false,
					referencia,
					payload.UsuarioCreador,
					"liberacion por disminucion de item",
				); err != nil {
					return err
				}
			}
		} else {
			if isTrackableProduct(prev.TipoItem, prev.ReferenciaID) {
				referencia := fmt.Sprintf("carrito:%d:item:%d", payload.CarritoID, payload.ID)
				if err := adjustCarritoItemStockTx(
					tx,
					payload.EmpresaID,
					payload.CarritoID,
					prev.TipoItem,
					prev.ReferenciaID,
					prev.Cantidad,
					false,
					referencia,
					payload.UsuarioCreador,
					"liberacion por cambio de referencia de item",
				); err != nil {
					return err
				}
			}
			if isTrackableProduct(payload.TipoItem, payload.ReferenciaID) {
				referencia := fmt.Sprintf("carrito:%d:item:%d", payload.CarritoID, payload.ID)
				if err := adjustCarritoItemStockTx(
					tx,
					payload.EmpresaID,
					payload.CarritoID,
					payload.TipoItem,
					payload.ReferenciaID,
					payload.Cantidad,
					true,
					referencia,
					payload.UsuarioCreador,
					"reserva por cambio de referencia de item",
				); err != nil {
					return err
				}
			}
		}
	}

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

	estadoCarrito, err := getCarritoEstadoTx(tx, empresaID, carritoID)
	if err != nil {
		return err
	}
	item, err := getCarritoItemSnapshotTx(tx, empresaID, carritoID, itemID)
	if err != nil {
		return err
	}
	if isItemActivo(item.Estado) && !isCarritoCerrado(estadoCarrito) {
		referencia := fmt.Sprintf("carrito:%d:item:%d", carritoID, itemID)
		if err := adjustCarritoItemStockTx(
			tx,
			empresaID,
			carritoID,
			item.TipoItem,
			item.ReferenciaID,
			item.Cantidad,
			false,
			referencia,
			item.UsuarioCreador,
			"liberacion por eliminacion de item",
		); err != nil {
			return err
		}
	}

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

	nuevoEstado := strings.TrimSpace(estado)
	if nuevoEstado == "" {
		nuevoEstado = "activo"
	}

	estadoCarrito, err := getCarritoEstadoTx(tx, empresaID, carritoID)
	if err != nil {
		return err
	}

	item, err := getCarritoItemSnapshotTx(tx, empresaID, carritoID, itemID)
	if err != nil {
		return err
	}

	estadoActual := strings.TrimSpace(item.Estado)
	if estadoActual == "" {
		estadoActual = "activo"
	}

	if !isCarritoCerrado(estadoCarrito) && isTrackableProduct(item.TipoItem, item.ReferenciaID) {
		if isItemActivo(estadoActual) && !isItemActivo(nuevoEstado) {
			referencia := fmt.Sprintf("carrito:%d:item:%d", carritoID, itemID)
			if err := adjustCarritoItemStockTx(
				tx,
				empresaID,
				carritoID,
				item.TipoItem,
				item.ReferenciaID,
				item.Cantidad,
				false,
				referencia,
				item.UsuarioCreador,
				"liberacion por desactivacion de item",
			); err != nil {
				return err
			}
		}
		if !isItemActivo(estadoActual) && isItemActivo(nuevoEstado) {
			referencia := fmt.Sprintf("carrito:%d:item:%d", carritoID, itemID)
			if err := adjustCarritoItemStockTx(
				tx,
				empresaID,
				carritoID,
				item.TipoItem,
				item.ReferenciaID,
				item.Cantidad,
				true,
				referencia,
				item.UsuarioCreador,
				"reserva por activacion de item",
			); err != nil {
				return err
			}
		}
	}

	if _, err := tx.Exec(`UPDATE carrito_compra_items SET estado = ?, fecha_actualizacion = datetime('now','localtime') WHERE empresa_id = ? AND carrito_id = ? AND id = ?`, nuevoEstado, empresaID, carritoID, itemID); err != nil {
		return err
	}
	if err := recalculateCarritoTotalsTx(tx, empresaID, carritoID); err != nil {
		return err
	}
	return tx.Commit()
}

// RecalculateCarritoCompraTotals recalcula totales del carrito basado en items activos.
func RecalculateCarritoCompraTotals(dbConn *sql.DB, empresaID, carritoID int64) error {
	_, err := RefreshCarritoTotalConTarifaPorDia(dbConn, empresaID, carritoID, time.Now())
	return err
}

type carritoItemSnapshot struct {
	ID             int64
	TipoItem       string
	ReferenciaID   int64
	Cantidad       float64
	Descripcion    string
	UsuarioCreador string
	Estado         string
}

func getCarritoEstadoTx(tx *sql.Tx, empresaID, carritoID int64) (string, error) {
	var estadoCarrito string
	err := tx.QueryRow(`SELECT COALESCE(estado_carrito, 'abierto') FROM carritos_compras WHERE empresa_id = ? AND id = ? LIMIT 1`, empresaID, carritoID).Scan(&estadoCarrito)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(strings.ToLower(estadoCarrito)), nil
}

func getCarritoItemSnapshotTx(tx *sql.Tx, empresaID, carritoID, itemID int64) (*carritoItemSnapshot, error) {
	row := tx.QueryRow(`SELECT
		id,
		COALESCE(tipo_item, 'producto'),
		COALESCE(referencia_id, 0),
		COALESCE(cantidad, 0),
		COALESCE(descripcion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo')
	FROM carrito_compra_items
	WHERE empresa_id = ? AND carrito_id = ? AND id = ?
	LIMIT 1`, empresaID, carritoID, itemID)

	item := &carritoItemSnapshot{}
	if err := row.Scan(&item.ID, &item.TipoItem, &item.ReferenciaID, &item.Cantidad, &item.Descripcion, &item.UsuarioCreador, &item.Estado); err != nil {
		return nil, err
	}
	item.TipoItem = defaultTipoItem(item.TipoItem)
	item.Estado = strings.TrimSpace(item.Estado)
	if item.Estado == "" {
		item.Estado = "activo"
	}
	return item, nil
}

func isCarritoCerrado(estadoCarrito string) bool {
	return strings.EqualFold(strings.TrimSpace(estadoCarrito), "cerrado")
}

func isItemActivo(estado string) bool {
	trim := strings.TrimSpace(strings.ToLower(estado))
	if trim == "" {
		return true
	}
	return trim == "activo"
}

func isTrackableProduct(tipoItem string, referenciaID int64) bool {
	if referenciaID <= 0 {
		return false
	}
	itemType := strings.TrimSpace(strings.ToLower(tipoItem))
	return itemType == "producto" || itemType == "combo"
}

type carritoStockComponent struct {
	ProductoID int64
	Cantidad   float64
}

type carritoStockContext struct {
	ProductoID    int64
	Cantidad      float64
	BodegaID      int64
	CostoUnitario float64
}

func resolveCarritoStockComponentsTx(tx *sql.Tx, empresaID int64, tipoItem string, referenciaID int64, cantidad float64, requireActiveCombo bool) ([]carritoStockComponent, error) {
	tipo := strings.TrimSpace(strings.ToLower(tipoItem))
	if referenciaID <= 0 || cantidad <= 0 {
		return nil, nil
	}

	if tipo == "producto" {
		return []carritoStockComponent{{ProductoID: referenciaID, Cantidad: cantidad}}, nil
	}
	if tipo != "combo" {
		return nil, nil
	}

	comboQuery := `SELECT COUNT(1) FROM combos_productos WHERE empresa_id = ? AND id = ?`
	comboArgs := []interface{}{empresaID, referenciaID}
	if requireActiveCombo {
		comboQuery += ` AND COALESCE(estado, 'activo') = 'activo'`
	}
	var comboCount int64
	if err := tx.QueryRow(comboQuery, comboArgs...).Scan(&comboCount); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "no such table") {
			return nil, fmt.Errorf("modulo de combos no disponible en la base de datos")
		}
		return nil, err
	}
	if comboCount == 0 {
		if requireActiveCombo {
			return nil, fmt.Errorf("combo no encontrado o inactivo")
		}
		return nil, fmt.Errorf("combo no encontrado")
	}

	rows, err := tx.Query(`SELECT
		COALESCE(producto_id, 0),
		COALESCE(cantidad, 0)
	FROM combos_productos_detalle
	WHERE empresa_id = ? AND combo_id = ? AND COALESCE(estado, 'activo') = 'activo'`, empresaID, referenciaID)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "no such table") {
			return nil, fmt.Errorf("detalle de combos no disponible en la base de datos")
		}
		return nil, err
	}
	defer rows.Close()

	merged := make(map[int64]float64)
	for rows.Next() {
		var productoID int64
		var cantidadPorCombo float64
		if err := rows.Scan(&productoID, &cantidadPorCombo); err != nil {
			return nil, err
		}
		if productoID <= 0 || cantidadPorCombo <= 0 {
			continue
		}
		merged[productoID] = round2(merged[productoID] + (cantidadPorCombo * cantidad))
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(merged) == 0 {
		return nil, fmt.Errorf("el combo no tiene ingredientes activos")
	}

	ids := make([]int64, 0, len(merged))
	for id := range merged {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })

	components := make([]carritoStockComponent, 0, len(ids))
	for _, id := range ids {
		qty := round2(merged[id])
		if qty <= 0 {
			continue
		}
		components = append(components, carritoStockComponent{ProductoID: id, Cantidad: qty})
	}
	if len(components) == 0 {
		return nil, fmt.Errorf("el combo no tiene ingredientes validos")
	}
	return components, nil
}

func resolveProductoStockContextTx(tx *sql.Tx, empresaID, productoID int64) (int64, float64, error) {
	var bodegaPrincipal sql.NullInt64
	var costo float64
	err := tx.QueryRow(`SELECT bodega_principal_id, COALESCE(costo, 0) FROM productos WHERE empresa_id = ? AND id = ? LIMIT 1`, empresaID, productoID).Scan(&bodegaPrincipal, &costo)
	if err != nil {
		return 0, 0, err
	}

	bodegaID := int64(0)
	if bodegaPrincipal.Valid {
		bodegaID = bodegaPrincipal.Int64
	}
	if bodegaID > 0 {
		return bodegaID, costo, nil
	}

	err = tx.QueryRow(`SELECT bodega_id FROM inventario_existencias WHERE empresa_id = ? AND producto_id = ? ORDER BY cantidad DESC, bodega_id ASC LIMIT 1`, empresaID, productoID).Scan(&bodegaID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, 0, fmt.Errorf("producto %d sin bodega de inventario", productoID)
		}
		return 0, 0, err
	}
	return bodegaID, costo, nil
}

func adjustCarritoItemStockTx(tx *sql.Tx, empresaID, carritoID int64, tipoItem string, referenciaID int64, cantidad float64, reservar bool, referencia, usuario, observaciones string) error {
	if cantidad <= 0 {
		return nil
	}
	if !isTrackableProduct(tipoItem, referenciaID) {
		return nil
	}

	components, err := resolveCarritoStockComponentsTx(tx, empresaID, tipoItem, referenciaID, cantidad, reservar)
	if err != nil {
		return err
	}
	if len(components) == 0 {
		return nil
	}

	if strings.TrimSpace(usuario) == "" {
		usuario = seedCarritoSystemUser
	}

	contexts := make([]carritoStockContext, 0, len(components))
	for _, component := range components {
		bodegaID, costoUnitario, err := resolveProductoStockContextTx(tx, empresaID, component.ProductoID)
		if err != nil {
			return err
		}

		ctx := carritoStockContext{
			ProductoID:    component.ProductoID,
			Cantidad:      component.Cantidad,
			BodegaID:      bodegaID,
			CostoUnitario: costoUnitario,
		}

		if reservar {
			var stockActual float64
			err := tx.QueryRow(`SELECT cantidad FROM inventario_existencias WHERE empresa_id = ? AND producto_id = ? AND bodega_id = ?`, empresaID, component.ProductoID, bodegaID).Scan(&stockActual)
			if err != nil {
				if err == sql.ErrNoRows {
					return ErrStockInsuficiente
				}
				return err
			}
			if stockActual < component.Cantidad {
				return ErrStockInsuficiente
			}
		}

		contexts = append(contexts, ctx)
	}

	normalizedTipo := strings.TrimSpace(strings.ToLower(tipoItem))
	baseReferencia := strings.TrimSpace(referencia)
	if baseReferencia == "" {
		baseReferencia = fmt.Sprintf("carrito:%d:item", carritoID)
	}

	for _, ctx := range contexts {
		movRef := baseReferencia
		if normalizedTipo == "combo" {
			movRef = fmt.Sprintf("%s:combo:%d:producto:%d", baseReferencia, referenciaID, ctx.ProductoID)
		}

		if reservar {
			if _, err := tx.Exec(`UPDATE inventario_existencias SET cantidad = cantidad - ?, fecha_actualizacion = datetime('now','localtime') WHERE empresa_id = ? AND producto_id = ? AND bodega_id = ?`, ctx.Cantidad, empresaID, ctx.ProductoID, ctx.BodegaID); err != nil {
				return err
			}
			if err := insertMovimientoTx(tx, InventarioMovimiento{
				EmpresaID:      empresaID,
				ProductoID:     ctx.ProductoID,
				BodegaOrigenID: ctx.BodegaID,
				Tipo:           "salida",
				Cantidad:       ctx.Cantidad,
				CostoUnitario:  ctx.CostoUnitario,
				Referencia:     strings.TrimSpace(movRef),
				UsuarioCreador: strings.TrimSpace(usuario),
				Estado:         "activo",
				Observaciones:  strings.TrimSpace(observaciones),
			}); err != nil {
				return err
			}
			continue
		}

		if err := upsertExistenciaTx(tx, empresaID, ctx.ProductoID, ctx.BodegaID, ctx.Cantidad, usuario, observaciones); err != nil {
			return err
		}
		if err := insertMovimientoTx(tx, InventarioMovimiento{
			EmpresaID:       empresaID,
			ProductoID:      ctx.ProductoID,
			BodegaDestinoID: ctx.BodegaID,
			Tipo:            "devolucion",
			Cantidad:        ctx.Cantidad,
			CostoUnitario:   ctx.CostoUnitario,
			Referencia:      strings.TrimSpace(movRef),
			UsuarioCreador:  strings.TrimSpace(usuario),
			Estado:          "activo",
			Observaciones:   strings.TrimSpace(observaciones),
		}); err != nil {
			return err
		}
	}

	return nil
}

func restoreCarritoItemsStockTx(tx *sql.Tx, empresaID, carritoID int64, motivo string) error {
	rows, err := tx.Query(`SELECT
		id,
		COALESCE(tipo_item, 'producto'),
		COALESCE(referencia_id, 0),
		COALESCE(cantidad, 0),
		COALESCE(usuario_creador, '')
	FROM carrito_compra_items
	WHERE empresa_id = ? AND carrito_id = ? AND COALESCE(estado, 'activo') = 'activo'`, empresaID, carritoID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var itemID int64
		var tipoItem string
		var referenciaID int64
		var cantidad float64
		var usuario string
		if err := rows.Scan(&itemID, &tipoItem, &referenciaID, &cantidad, &usuario); err != nil {
			return err
		}
		referencia := fmt.Sprintf("carrito:%d:item:%d", carritoID, itemID)
		if err := adjustCarritoItemStockTx(
			tx,
			empresaID,
			carritoID,
			tipoItem,
			referenciaID,
			cantidad,
			false,
			referencia,
			usuario,
			"liberacion de stock por "+motivo,
		); err != nil {
			return err
		}
	}
	return rows.Err()
}

const seedCarritoSystemUser = "sistema_carrito"

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
