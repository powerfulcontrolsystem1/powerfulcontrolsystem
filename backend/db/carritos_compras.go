package db

import (
	"database/sql"
	"fmt"
	"math"
	"sort"
	"strconv"
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

// CarritoStationMetricInput representa una medicion operativa por estacion.
type CarritoStationMetricInput struct {
	EmpresaID           int64
	CarritoID           int64
	EstacionID          int64
	EstacionCodigo      string
	EstacionNombre      string
	EventoOperacion     string
	MetodoPago          string
	Moneda              string
	MontoTotal          float64
	MontoPagado         float64
	MontoAnulado        float64
	DevolucionTotal     float64
	DuracionSegundos    int64
	ActivadoEn          string
	PagadoEn            string
	ReferenciaOperacion string
	FechaEvento         string
	UsuarioCreador      string
	Observaciones       string
}

// CarritoStationMetricSummary consolida rendimiento de ventas simples por estacion.
type CarritoStationMetricSummary struct {
	EstacionID             int64   `json:"estacion_id"`
	EstacionCodigo         string  `json:"estacion_codigo"`
	EstacionNombre         string  `json:"estacion_nombre"`
	VentasPagadas          int64   `json:"ventas_pagadas"`
	Correcciones           int64   `json:"correcciones"`
	MontoVendido           float64 `json:"monto_vendido"`
	MontoPagado            float64 `json:"monto_pagado"`
	MontoAnulado           float64 `json:"monto_anulado"`
	DevolucionTotal        float64 `json:"devolucion_total"`
	TiempoPromedioSegundos float64 `json:"tiempo_promedio_segundos"`
	TiempoMinSegundos      int64   `json:"tiempo_min_segundos"`
	TiempoMaxSegundos      int64   `json:"tiempo_max_segundos"`
	UltimaOperacion        string  `json:"ultima_operacion"`
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
		`CREATE TABLE IF NOT EXISTS empresa_ventas_estacion_metricas (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			carrito_id INTEGER NOT NULL,
			estacion_id INTEGER DEFAULT 0,
			estacion_codigo TEXT,
			estacion_nombre TEXT,
			evento_operacion TEXT NOT NULL DEFAULT 'venta_pagada',
			metodo_pago TEXT DEFAULT 'efectivo',
			moneda TEXT DEFAULT 'COP',
			monto_total REAL DEFAULT 0,
			monto_pagado REAL DEFAULT 0,
			monto_anulado REAL DEFAULT 0,
			devolucion_total REAL DEFAULT 0,
			duracion_segundos INTEGER DEFAULT 0,
			activado_en TEXT,
			pagado_en TEXT,
			referencia_operacion TEXT,
			fecha_evento TEXT DEFAULT (datetime('now','localtime')),
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
	}
	for _, stmt := range stmts {
		if _, err := dbConn.Exec(stmt); err != nil {
			return err
		}
	}

	indexStmts := []struct {
		name  string
		query string
	}{
		{name: "ux_carritos_empresa_codigo", query: `CREATE UNIQUE INDEX IF NOT EXISTS ux_carritos_empresa_codigo ON carritos_compras(empresa_id, codigo);`},
		{name: "ux_carritos_empresa_nombre", query: `CREATE UNIQUE INDEX IF NOT EXISTS ux_carritos_empresa_nombre ON carritos_compras(empresa_id, nombre);`},
		{name: "ix_carritos_empresa_estado", query: `CREATE INDEX IF NOT EXISTS ix_carritos_empresa_estado ON carritos_compras(empresa_id, estado, estado_carrito);`},
		{name: "ix_carrito_items_empresa_carrito", query: `CREATE INDEX IF NOT EXISTS ix_carrito_items_empresa_carrito ON carrito_compra_items(empresa_id, carrito_id);`},
		{name: "ix_carrito_items_empresa_referencia", query: `CREATE INDEX IF NOT EXISTS ix_carrito_items_empresa_referencia ON carrito_compra_items(empresa_id, referencia_id);`},
		{name: "ix_ventas_estacion_metricas_empresa_estacion_fecha", query: `CREATE INDEX IF NOT EXISTS ix_ventas_estacion_metricas_empresa_estacion_fecha ON empresa_ventas_estacion_metricas(empresa_id, estacion_id, fecha_evento DESC);`},
		{name: "ix_ventas_estacion_metricas_empresa_evento", query: `CREATE INDEX IF NOT EXISTS ix_ventas_estacion_metricas_empresa_evento ON empresa_ventas_estacion_metricas(empresa_id, evento_operacion, fecha_evento DESC);`},
		{name: "ix_ventas_estacion_metricas_carrito", query: `CREATE INDEX IF NOT EXISTS ix_ventas_estacion_metricas_carrito ON empresa_ventas_estacion_metricas(empresa_id, carrito_id, fecha_evento DESC);`},
	}
	for _, idx := range indexStmts {
		if err := ensureIndexIfMissing(dbConn, idx.name, idx.query); err != nil {
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

	if err := ensureColumnIfMissing(dbConn, "empresa_ventas_estacion_metricas", "empresa_id", "INTEGER NOT NULL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_ventas_estacion_metricas", "carrito_id", "INTEGER NOT NULL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_ventas_estacion_metricas", "estacion_id", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_ventas_estacion_metricas", "estacion_codigo", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_ventas_estacion_metricas", "estacion_nombre", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_ventas_estacion_metricas", "evento_operacion", "TEXT DEFAULT 'venta_pagada'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_ventas_estacion_metricas", "metodo_pago", "TEXT DEFAULT 'efectivo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_ventas_estacion_metricas", "moneda", "TEXT DEFAULT 'COP'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_ventas_estacion_metricas", "monto_total", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_ventas_estacion_metricas", "monto_pagado", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_ventas_estacion_metricas", "monto_anulado", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_ventas_estacion_metricas", "devolucion_total", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_ventas_estacion_metricas", "duracion_segundos", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_ventas_estacion_metricas", "activado_en", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_ventas_estacion_metricas", "pagado_en", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_ventas_estacion_metricas", "referencia_operacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_ventas_estacion_metricas", "fecha_evento", "TEXT DEFAULT (datetime('now','localtime'))"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_ventas_estacion_metricas", "fecha_creacion", "TEXT DEFAULT (datetime('now','localtime'))"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_ventas_estacion_metricas", "fecha_actualizacion", "TEXT DEFAULT (datetime('now','localtime'))"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_ventas_estacion_metricas", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_ventas_estacion_metricas", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_ventas_estacion_metricas", "observaciones", "TEXT"); err != nil {
		return err
	}

	if !isPostgresDialect() {
		if _, err := dbConn.Exec(`PRAGMA busy_timeout = 5000`); err != nil {
			return err
		}
	}

	return nil
}

func ensureIndexIfMissing(dbConn *sql.DB, indexName, createStmt string) error {
	if dbConn == nil {
		return fmt.Errorf("db connection is nil")
	}
	if shouldUsePostgresCompat(dbConn) {
		_, err := execSQLCompat(dbConn, createStmt)
		return err
	}

	var existingName string
	err := queryRowSQLCompat(dbConn, `SELECT name FROM sqlite_master WHERE type = 'index' AND name = ? LIMIT 1`, indexName).Scan(&existingName)
	if err == nil {
		return nil
	}
	if err != sql.ErrNoRows {
		return err
	}

	stmt := strings.Replace(createStmt, " IF NOT EXISTS", "", 1)
	_, err = execSQLCompat(dbConn, stmt)
	return err
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

func defaultMonedaEmpresa(dbConn *sql.DB, empresaID int64, payloadMoneda string) string {
	if strings.TrimSpace(payloadMoneda) != "" {
		return defaultMoneda(payloadMoneda)
	}
	if empresaID <= 0 {
		return defaultMoneda("")
	}
	cfg, err := GetEmpresaConfiguracionAvanzada(dbConn, empresaID)
	if err != nil || cfg == nil {
		return defaultMoneda("")
	}
	return defaultMoneda(cfg.MonedaCodigo)
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

func normalizeCarritoStationMetricEvent(v string) string {
	event := strings.TrimSpace(strings.ToLower(v))
	switch event {
	case "", "venta_pagada":
		return "venta_pagada"
	case "cierre_parcial_anulado", "sesion_recuperada":
		return event
	default:
		return "operacion"
	}
}

func parseCarritoStationID(referenciaExterna, codigo string, empresaID int64) int64 {
	ref := strings.ToUpper(strings.TrimSpace(referenciaExterna))
	if strings.HasPrefix(ref, "ESTACION_") {
		n, err := strconv.ParseInt(strings.TrimPrefix(ref, "ESTACION_"), 10, 64)
		if err == nil && n > 0 {
			return n
		}
	}
	prefix := strings.ToUpper(fmt.Sprintf("EST-%d-", empresaID))
	code := strings.ToUpper(strings.TrimSpace(codigo))
	if strings.HasPrefix(code, prefix) {
		n, err := strconv.ParseInt(strings.TrimPrefix(code, prefix), 10, 64)
		if err == nil && n > 0 {
			return n
		}
	}
	return 0
}

// ResolveCarritoStationIdentity obtiene metadatos de estacion desde un carrito de ventas.
func ResolveCarritoStationIdentity(carrito *CarritoCompra) (int64, string, string) {
	if carrito == nil {
		return 0, "", ""
	}
	estacionID := parseCarritoStationID(carrito.ReferenciaExterna, carrito.Codigo, carrito.EmpresaID)
	estacionCodigo := strings.TrimSpace(carrito.Codigo)
	if estacionCodigo == "" && estacionID > 0 {
		estacionCodigo = fmt.Sprintf("EST-%d-%d", carrito.EmpresaID, estacionID)
	}
	estacionNombre := strings.TrimSpace(carrito.Nombre)
	if estacionNombre == "" && estacionID > 0 {
		estacionNombre = fmt.Sprintf("Estacion %d", estacionID)
	}
	if estacionNombre == "" {
		estacionNombre = "Estacion"
	}
	return estacionID, estacionCodigo, estacionNombre
}

func parseCarritoDateTime(raw string) (time.Time, bool) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return time.Time{}, false
	}
	layouts := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04",
		"2006-01-02T15:04",
		"2006-01-02",
	}
	for _, layout := range layouts {
		ts, err := time.ParseInLocation(layout, value, time.Local)
		if err == nil {
			return ts, true
		}
	}
	return time.Time{}, false
}

// ResolveCarritoAttentionDurationSeconds calcula tiempo de atencion entre activacion y pago.
func ResolveCarritoAttentionDurationSeconds(activadoEn, pagadoEn string) int64 {
	activadoTS, okActivado := parseCarritoDateTime(activadoEn)
	pagadoTS, okPagado := parseCarritoDateTime(pagadoEn)
	if !okActivado || !okPagado {
		return 0
	}
	delta := pagadoTS.Sub(activadoTS)
	if delta <= 0 {
		return 0
	}
	return int64(delta.Seconds())
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

const (
	carritoTxRetryMaxAttempts = 5
	carritoTxRetryBaseDelay   = 20 * time.Millisecond
)

func isSQLiteBusyOrLockedError(err error) bool {
	if err == nil {
		return false
	}
	lower := strings.ToLower(strings.TrimSpace(err.Error()))
	return strings.Contains(lower, "database is locked") || strings.Contains(lower, "database is busy")
}

func withCarritoTxRetry(dbConn *sql.DB, run func(tx *sql.Tx) error) error {
	var lastRetryErr error
	for attempt := 0; attempt < carritoTxRetryMaxAttempts; attempt++ {
		tx, err := dbConn.Begin()
		if err != nil {
			if isSQLiteBusyOrLockedError(err) {
				lastRetryErr = err
				time.Sleep(time.Duration(attempt+1) * carritoTxRetryBaseDelay)
				continue
			}
			return err
		}

		err = run(tx)
		if err != nil {
			_ = tx.Rollback()
			if isSQLiteBusyOrLockedError(err) && attempt+1 < carritoTxRetryMaxAttempts {
				lastRetryErr = err
				time.Sleep(time.Duration(attempt+1) * carritoTxRetryBaseDelay)
				continue
			}
			return err
		}

		if err := tx.Commit(); err != nil {
			_ = tx.Rollback()
			if isSQLiteBusyOrLockedError(err) && attempt+1 < carritoTxRetryMaxAttempts {
				lastRetryErr = err
				time.Sleep(time.Duration(attempt+1) * carritoTxRetryBaseDelay)
				continue
			}
			return err
		}

		return nil
	}

	if lastRetryErr != nil {
		return lastRetryErr
	}
	return fmt.Errorf("no se pudo completar transaccion de carrito")
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
	moneda := defaultMonedaEmpresa(dbConn, payload.EmpresaID, payload.Moneda)
	return insertSQLCompat(dbConn, `INSERT INTO carritos_compras (
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
		moneda,
		strings.TrimSpace(payload.ReferenciaExterna),
		strings.TrimSpace(payload.UsuarioCreador),
		strings.TrimSpace(payload.Observaciones),
		metodoPago,
		strings.TrimSpace(payload.ReferenciaPago),
	)
}

// GetCarritosCompraByEmpresa lista carritos por empresa.
func GetCarritosCompraByEmpresa(dbConn *sql.DB, empresaID int64, includeInactive bool, q string) ([]CarritoCompra, error) {
	query, args := buildCarritosCompraByEmpresaQuery(empresaID, includeInactive, q, true)
	rows, err := dbConn.Query(query, args...)
	if err != nil {
		if !shouldRetryCarritosCompraWithoutClientes(err) {
			return nil, err
		}
		query, args = buildCarritosCompraByEmpresaQuery(empresaID, includeInactive, q, false)
		rows, err = dbConn.Query(query, args...)
		if err != nil {
			return nil, err
		}
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
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func buildCarritosCompraByEmpresaQuery(empresaID int64, includeInactive bool, q string, includeClientes bool) (string, []interface{}) {
	clienteNombreExpr := `''`
	joinClientes := ""
	joinItemCounts := `
	LEFT JOIN (
		SELECT empresa_id, carrito_id, COUNT(id) AS item_count
		FROM carrito_compra_items
		WHERE COALESCE(estado, 'activo') = 'activo'
		GROUP BY empresa_id, carrito_id
	) ic ON ic.empresa_id = c.empresa_id AND ic.carrito_id = c.id`
	searchClienteExpr := ""
	if includeClientes {
		clienteNombreExpr = `COALESCE(cl.nombre_razon_social, '')`
		joinClientes = `
	LEFT JOIN clientes cl ON cl.empresa_id = c.empresa_id AND cl.id = c.cliente_id`
		searchClienteExpr = ` OR
			lower(COALESCE(cl.nombre_razon_social, '')) LIKE lower(?)`
	}

	query := `SELECT
		c.id,
		c.empresa_id,
		COALESCE(c.codigo, ''),
		COALESCE(c.nombre, ''),
		COALESCE(c.canal_venta, 'mostrador'),
		COALESCE(c.cliente_id, 0),
		` + clienteNombreExpr + `,
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
		COALESCE(ic.item_count, 0),
		COALESCE(c.fecha_creacion, ''),
		COALESCE(c.fecha_actualizacion, ''),
		COALESCE(c.usuario_creador, ''),
		COALESCE(c.estado, 'activo'),
		COALESCE(c.observaciones, '')
	FROM carritos_compras c` + joinClientes + joinItemCounts + `
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
			lower(COALESCE(c.codigo, '')) LIKE lower(?)` + searchClienteExpr + `
		)`
		args = append(args, pat, pat)
		if includeClientes {
			args = append(args, pat)
		}
	}
	query += ` ORDER BY c.id DESC`
	return query, args
}

func shouldRetryCarritosCompraWithoutClientes(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(strings.TrimSpace(err.Error()))
	if msg == "" {
		return false
	}
	return (strings.Contains(msg, "clientes") || strings.Contains(msg, "nombre_razon_social")) &&
		(strings.Contains(msg, "no such table") || strings.Contains(msg, "does not exist") || strings.Contains(msg, "no such column") || strings.Contains(msg, "unknown column"))
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
	if err := revertCodigoDescuentoUsoPorCarritoTx(tx, empresaID, carritoID, "anulada", "carrito eliminado", "sistema"); err != nil {
		return err
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
	tx, err := dbConn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	estadoObjetivo := strings.TrimSpace(estadoCarrito)
	if strings.EqualFold(estadoObjetivo, "abierto") {
		if err := revertCodigoDescuentoUsoPorCarritoTx(tx, empresaID, carritoID, "revertida", "carrito reabierto", "sistema"); err != nil {
			return err
		}
	}

	if _, err := tx.Exec(`UPDATE carritos_compras SET estado_carrito = ?, fecha_actualizacion = datetime('now','localtime') WHERE empresa_id = ? AND id = ?`, estadoObjetivo, empresaID, carritoID); err != nil {
		return err
	}

	return tx.Commit()
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
	if isCarritoCerrado(estadoPrevio) {
		if err := revertCodigoDescuentoUsoPorCarritoTx(tx, empresaID, carritoID, "revertida", "reactivacion de sesion de estacion", "sistema"); err != nil {
			return err
		}
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

// RecoverInterruptedCarritoSession recupera una sesion interrumpida sin perder items reservados.
func RecoverInterruptedCarritoSession(dbConn *sql.DB, empresaID, carritoID int64) error {
	return withCarritoTxRetry(dbConn, func(tx *sql.Tx) error {
		var estado string
		var estadoCarrito string
		var pagadoEn string
		err := tx.QueryRow(`SELECT
			COALESCE(estado, 'activo'),
			COALESCE(estado_carrito, 'abierto'),
			COALESCE(pagado_en, '')
		FROM carritos_compras
		WHERE empresa_id = ? AND id = ?
		LIMIT 1`, empresaID, carritoID).Scan(&estado, &estadoCarrito, &pagadoEn)
		if err != nil {
			return err
		}

		if strings.TrimSpace(pagadoEn) != "" {
			return fmt.Errorf("la venta ya fue pagada y no puede recuperarse como interrumpida")
		}

		estado = strings.TrimSpace(strings.ToLower(estado))
		estadoCarrito = strings.TrimSpace(strings.ToLower(estadoCarrito))
		if estado == "activo" && estadoCarrito == "abierto" {
			if _, err := tx.Exec(`UPDATE carritos_compras SET fecha_actualizacion = datetime('now','localtime') WHERE empresa_id = ? AND id = ?`, empresaID, carritoID); err != nil {
				return err
			}
			return nil
		}

		if _, err := tx.Exec(`UPDATE carritos_compras SET
			estado = 'activo',
			estado_carrito = 'abierto',
			activado_en = datetime('now','localtime'),
			fecha_actualizacion = datetime('now','localtime')
		WHERE empresa_id = ? AND id = ?`, empresaID, carritoID); err != nil {
			return err
		}

		return recalculateCarritoTotalsTx(tx, empresaID, carritoID)
	})
}

// CancelCarritoPartialClosure registra una anulacion parcial de cierre sobre una venta ya pagada.
func CancelCarritoPartialClosure(dbConn *sql.DB, empresaID, carritoID int64, montoAnulado float64) (float64, float64, error) {
	montoAnulado = round2(montoAnulado)
	if montoAnulado <= 0 {
		return 0, 0, fmt.Errorf("monto_anulado debe ser mayor a cero")
	}

	totalPagadoNuevo := 0.0
	devolucionTotalNueva := 0.0

	err := withCarritoTxRetry(dbConn, func(tx *sql.Tx) error {
		var pagadoEn string
		var totalPagadoActual float64
		var devolucionActual float64
		err := tx.QueryRow(`SELECT
			COALESCE(pagado_en, ''),
			COALESCE(total_pagado, 0),
			COALESCE(devolucion_total, 0)
		FROM carritos_compras
		WHERE empresa_id = ? AND id = ?
		LIMIT 1`, empresaID, carritoID).Scan(&pagadoEn, &totalPagadoActual, &devolucionActual)
		if err != nil {
			return err
		}

		if strings.TrimSpace(pagadoEn) == "" {
			return fmt.Errorf("solo se puede anular parcialmente una venta pagada")
		}

		totalPagadoActual = round2(totalPagadoActual)
		if montoAnulado >= totalPagadoActual {
			return fmt.Errorf("monto_anulado debe ser menor al total_pagado actual para mantener anulacion parcial")
		}

		totalPagadoNuevo = round2(totalPagadoActual - montoAnulado)
		devolucionTotalNueva = round2(devolucionActual + montoAnulado)

		_, err = tx.Exec(`UPDATE carritos_compras SET
			total_pagado = ?,
			devolucion_total = ?,
			fecha_actualizacion = datetime('now','localtime')
		WHERE empresa_id = ? AND id = ?`, totalPagadoNuevo, devolucionTotalNueva, empresaID, carritoID)
		return err
	})
	if err != nil {
		return 0, 0, err
	}

	return totalPagadoNuevo, devolucionTotalNueva, nil
}

// PayCarritoStationSession marca un carrito como pagado/inactivo y guarda resumen de cobro.
func PayCarritoStationSession(dbConn *sql.DB, empresaID, carritoID int64, metodoPago, referenciaPago, descuentoTipo, descuentoCodigo string, descuentoValor, devolucionTotal, totalPagado float64, codigoDescuentoID int64, usuarioCreador string) error {
	metodoPago = NormalizeMetodoPagoCarrito(metodoPago)
	if metodoPago == "" {
		return fmt.Errorf("metodo_pago invalido")
	}
	usuarioCreador = strings.TrimSpace(usuarioCreador)
	if usuarioCreador == "" {
		usuarioCreador = "sistema"
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
		if err := markCodigoDescuentoUsoTx(tx, empresaID, codigoDescuentoID, carritoID, descuentoValor, usuarioCreador, strings.TrimSpace(referenciaPago)); err != nil {
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

// RecordCarritoStationMetric persiste una metrica operativa de ventas simples por estacion.
func RecordCarritoStationMetric(dbConn *sql.DB, input CarritoStationMetricInput) (int64, error) {
	if dbConn == nil {
		return 0, fmt.Errorf("conexion de base de datos no disponible")
	}
	if input.EmpresaID <= 0 {
		return 0, fmt.Errorf("empresa_id invalido")
	}
	if input.CarritoID <= 0 {
		return 0, fmt.Errorf("carrito_id invalido")
	}

	evento := normalizeCarritoStationMetricEvent(input.EventoOperacion)
	metodoPago := NormalizeMetodoPagoCarrito(input.MetodoPago)
	if metodoPago == "" {
		metodoPago = "efectivo"
	}
	moneda := strings.TrimSpace(strings.ToUpper(input.Moneda))
	if moneda == "" {
		moneda = "COP"
	}
	estacionCodigo := strings.TrimSpace(input.EstacionCodigo)
	if estacionCodigo == "" && input.EstacionID > 0 {
		estacionCodigo = fmt.Sprintf("EST-%d-%d", input.EmpresaID, input.EstacionID)
	}
	estacionNombre := strings.TrimSpace(input.EstacionNombre)
	if estacionNombre == "" && input.EstacionID > 0 {
		estacionNombre = fmt.Sprintf("Estacion %d", input.EstacionID)
	}
	duracionSegundos := input.DuracionSegundos
	if duracionSegundos <= 0 {
		duracionSegundos = ResolveCarritoAttentionDurationSeconds(input.ActivadoEn, input.PagadoEn)
	}
	if duracionSegundos < 0 {
		duracionSegundos = 0
	}
	fechaEvento := strings.TrimSpace(input.FechaEvento)
	if fechaEvento == "" {
		fechaEvento = strings.TrimSpace(input.PagadoEn)
	}
	if fechaEvento == "" {
		fechaEvento = strings.TrimSpace(input.ActivadoEn)
	}

	id, err := insertSQLCompat(dbConn, `INSERT INTO empresa_ventas_estacion_metricas (
		empresa_id,
		carrito_id,
		estacion_id,
		estacion_codigo,
		estacion_nombre,
		evento_operacion,
		metodo_pago,
		moneda,
		monto_total,
		monto_pagado,
		monto_anulado,
		devolucion_total,
		duracion_segundos,
		activado_en,
		pagado_en,
		referencia_operacion,
		fecha_evento,
		fecha_creacion,
		fecha_actualizacion,
		usuario_creador,
		estado,
		observaciones
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, COALESCE(NULLIF(?, ''), datetime('now','localtime')), datetime('now','localtime'), datetime('now','localtime'), ?, 'activo', ?)`,
		input.EmpresaID,
		input.CarritoID,
		input.EstacionID,
		estacionCodigo,
		estacionNombre,
		evento,
		metodoPago,
		moneda,
		round2(input.MontoTotal),
		round2(input.MontoPagado),
		round2(input.MontoAnulado),
		round2(input.DevolucionTotal),
		duracionSegundos,
		strings.TrimSpace(input.ActivadoEn),
		strings.TrimSpace(input.PagadoEn),
		strings.TrimSpace(input.ReferenciaOperacion),
		fechaEvento,
		strings.TrimSpace(input.UsuarioCreador),
		strings.TrimSpace(input.Observaciones),
	)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// ListCarritoStationMetricSummary resume ventas, correcciones y tiempos por estacion.
func ListCarritoStationMetricSummary(dbConn *sql.DB, empresaID, estacionID int64, days, limit int) ([]CarritoStationMetricSummary, error) {
	if empresaID <= 0 {
		return nil, fmt.Errorf("empresa_id invalido")
	}
	if days <= 0 {
		days = 7
	}
	if days > 365 {
		days = 365
	}
	if limit <= 0 {
		limit = 10
	}
	if limit > 200 {
		limit = 200
	}

	query := `SELECT
		COALESCE(estacion_id, 0) AS estacion_id,
		COALESCE(MAX(NULLIF(estacion_codigo, '')), '') AS estacion_codigo,
		COALESCE(MAX(NULLIF(estacion_nombre, '')), '') AS estacion_nombre,
		SUM(CASE WHEN evento_operacion = 'venta_pagada' THEN 1 ELSE 0 END) AS ventas_pagadas,
		SUM(CASE WHEN evento_operacion = 'cierre_parcial_anulado' THEN 1 ELSE 0 END) AS correcciones,
		COALESCE(SUM(CASE WHEN evento_operacion = 'venta_pagada' THEN COALESCE(monto_total, 0) ELSE 0 END), 0) AS monto_vendido,
		COALESCE(SUM(CASE WHEN evento_operacion = 'venta_pagada' THEN COALESCE(monto_pagado, 0) ELSE 0 END), 0) AS monto_pagado,
		COALESCE(SUM(COALESCE(monto_anulado, 0)), 0) AS monto_anulado,
		COALESCE(SUM(COALESCE(devolucion_total, 0)), 0) AS devolucion_total,
		COALESCE(AVG(CASE WHEN evento_operacion = 'venta_pagada' AND COALESCE(duracion_segundos, 0) > 0 THEN duracion_segundos END), 0) AS tiempo_promedio_segundos,
		COALESCE(MIN(CASE WHEN evento_operacion = 'venta_pagada' AND COALESCE(duracion_segundos, 0) > 0 THEN duracion_segundos END), 0) AS tiempo_min_segundos,
		COALESCE(MAX(CASE WHEN evento_operacion = 'venta_pagada' AND COALESCE(duracion_segundos, 0) > 0 THEN duracion_segundos END), 0) AS tiempo_max_segundos,
		COALESCE(MAX(COALESCE(fecha_evento, fecha_creacion, '')), '') AS ultima_operacion
	FROM empresa_ventas_estacion_metricas
	WHERE empresa_id = ?
		AND COALESCE(estado, 'activo') = 'activo'
		AND datetime(COALESCE(fecha_evento, fecha_creacion, datetime('now','localtime'))) >= datetime('now','localtime', ?)`
	args := []interface{}{empresaID, fmt.Sprintf("-%d day", days)}

	if estacionID > 0 {
		query += ` AND estacion_id = ?`
		args = append(args, estacionID)
	}

	query += `
	GROUP BY COALESCE(estacion_id, 0)
	ORDER BY ventas_pagadas DESC, monto_pagado DESC, estacion_id ASC
	LIMIT ?`
	args = append(args, limit)

	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]CarritoStationMetricSummary, 0)
	for rows.Next() {
		var row CarritoStationMetricSummary
		if err := rows.Scan(
			&row.EstacionID,
			&row.EstacionCodigo,
			&row.EstacionNombre,
			&row.VentasPagadas,
			&row.Correcciones,
			&row.MontoVendido,
			&row.MontoPagado,
			&row.MontoAnulado,
			&row.DevolucionTotal,
			&row.TiempoPromedioSegundos,
			&row.TiempoMinSegundos,
			&row.TiempoMaxSegundos,
			&row.UltimaOperacion,
		); err != nil {
			return nil, err
		}
		if strings.TrimSpace(row.EstacionNombre) == "" {
			if row.EstacionID > 0 {
				row.EstacionNombre = fmt.Sprintf("Estacion %d", row.EstacionID)
			} else {
				row.EstacionNombre = "Estacion"
			}
		}
		row.MontoVendido = round2(row.MontoVendido)
		row.MontoPagado = round2(row.MontoPagado)
		row.MontoAnulado = round2(row.MontoAnulado)
		row.DevolucionTotal = round2(row.DevolucionTotal)
		row.TiempoPromedioSegundos = round2(row.TiempoPromedioSegundos)
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
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

	id := int64(0)
	err := withCarritoTxRetry(dbConn, func(tx *sql.Tx) error {
		if err := validateCarritoEnEmpresaTx(tx, payload.EmpresaID, payload.CarritoID); err != nil {
			return err
		}

		itemID, insertErr := insertTxSQLCompat(tx, `INSERT INTO carrito_compra_items (
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
		if insertErr != nil {
			return insertErr
		}
		id = itemID

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
				return err
			}
		}

		return recalculateCarritoTotalsTx(tx, payload.EmpresaID, payload.CarritoID)
	})
	if err != nil {
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

	return withCarritoTxRetry(dbConn, func(tx *sql.Tx) error {
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

		return recalculateCarritoTotalsTx(tx, payload.EmpresaID, payload.CarritoID)
	})
}

// DeleteCarritoCompraItem elimina item y recalcula totales.
func DeleteCarritoCompraItem(dbConn *sql.DB, empresaID, carritoID, itemID int64) error {
	return withCarritoTxRetry(dbConn, func(tx *sql.Tx) error {
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
		return recalculateCarritoTotalsTx(tx, empresaID, carritoID)
	})
}

// SetCarritoCompraItemEstado activa/desactiva item y recalcula totales.
func SetCarritoCompraItemEstado(dbConn *sql.DB, empresaID, carritoID, itemID int64, estado string) error {
	nuevoEstado := strings.TrimSpace(estado)
	if nuevoEstado == "" {
		nuevoEstado = "activo"
	}

	return withCarritoTxRetry(dbConn, func(tx *sql.Tx) error {
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
		return recalculateCarritoTotalsTx(tx, empresaID, carritoID)
	})
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
			res, err := tx.Exec(`UPDATE inventario_existencias
			SET cantidad = cantidad - ?,
				fecha_actualizacion = datetime('now','localtime')
			WHERE empresa_id = ?
				AND producto_id = ?
				AND bodega_id = ?
				AND cantidad >= ?`, ctx.Cantidad, empresaID, ctx.ProductoID, ctx.BodegaID, ctx.Cantidad)
			if err != nil {
				return err
			}
			affected, _ := res.RowsAffected()
			if affected == 0 {
				return ErrStockInsuficiente
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
