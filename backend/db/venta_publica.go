package db

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// EnsureVentaPublicaSchema crea las tablas necesarias para el módulo de venta pública.
// Debe ser idempotente y segura para SQLite y Postgres (usa IF NOT EXISTS cuando es soportado).
func EnsureVentaPublicaSchema(db *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS paginas_publicas (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			slug TEXT NOT NULL,
			titulo TEXT NOT NULL,
			descripcion TEXT,
			video_url TEXT,
			activo INTEGER NOT NULL DEFAULT 1,
			creado_en DATETIME DEFAULT (datetime('now')),
			actualizado_en DATETIME DEFAULT (datetime('now'))
		);`,

		`CREATE UNIQUE INDEX IF NOT EXISTS idx_paginas_publicas_empresa_slug ON paginas_publicas(empresa_id, slug);`,

		`CREATE TABLE IF NOT EXISTS productos_publicos (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			pagina_id INTEGER NOT NULL,
			nombre TEXT NOT NULL,
			descripcion TEXT,
			precio_cents INTEGER NOT NULL DEFAULT 0,
			moneda TEXT NOT NULL DEFAULT 'COP',
			stock INTEGER,
			sku TEXT,
			youtube_url TEXT,
			activo INTEGER NOT NULL DEFAULT 1,
			creado_en DATETIME DEFAULT (datetime('now')),
			actualizado_en DATETIME DEFAULT (datetime('now')),
			FOREIGN KEY(pagina_id) REFERENCES paginas_publicas(id) ON DELETE CASCADE
		);`,

		`CREATE INDEX IF NOT EXISTS idx_productos_publicos_pagina_id ON productos_publicos(pagina_id);`,

		`CREATE TABLE IF NOT EXISTS imagenes_productos_publicos (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			producto_id INTEGER NOT NULL,
			url TEXT NOT NULL,
			orden INTEGER NOT NULL DEFAULT 0,
			FOREIGN KEY(producto_id) REFERENCES productos_publicos(id) ON DELETE CASCADE
		);`,

		`CREATE INDEX IF NOT EXISTS idx_imagenes_producto_id ON imagenes_productos_publicos(producto_id);`,

		`CREATE TABLE IF NOT EXISTS empresa_payment_settings (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			provider TEXT NOT NULL,
			config TEXT,
			activo INTEGER NOT NULL DEFAULT 1,
			creado_en DATETIME DEFAULT (datetime('now')),
			actualizado_en DATETIME DEFAULT (datetime('now'))
		);`,

		`CREATE UNIQUE INDEX IF NOT EXISTS idx_empresa_payment_provider ON empresa_payment_settings(empresa_id, provider);`,
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction venta_publica schema: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	for _, s := range stmts {
		if _, err := tx.Exec(s); err != nil {
			return fmt.Errorf("exec venta_publica schema stmt: %w; stmt=%s", err, s)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit venta_publica schema: %w", err)
	}

	return nil
}

// (Se definirá EnsureEmpresaVentaPublicaSchema más abajo con migraciones completas.)

type PaginaPublica struct {
	ID          int64  `json:"id"`
	EmpresaID   int64  `json:"empresa_id"`
	Slug        string `json:"slug"`
	Titulo      string `json:"titulo"`
	Descripcion string `json:"descripcion"`
	VideoURL    string `json:"video_url"`
	Activo      bool   `json:"activo"`
}

type ProductoPublico struct {
	ID          int64          `json:"id"`
	PaginaID    int64          `json:"pagina_id"`
	Nombre      string         `json:"nombre"`
	Descripcion string         `json:"descripcion"`
	PrecioCents int64          `json:"precio_cents"`
	Moneda      string         `json:"moneda"`
	Stock       sql.NullInt64  `json:"-"`
	SKU         string         `json:"sku"`
	YoutubeURL  string         `json:"youtube_url"`
	Activo      bool           `json:"activo"`
}

// Use ventaPublicaBoolToInt más abajo; evitar colisiones de nombre en paquete db.

func CreatePaginaPublica(db *sql.DB, empresaID int64, slug, titulo, descripcion, videoURL string, activo bool) (int64, error) {
	if empresaID <= 0 || strings.TrimSpace(slug) == "" || strings.TrimSpace(titulo) == "" {
		return 0, fmt.Errorf("empresa_id, slug y titulo son obligatorios")
	}
	res, err := db.Exec(`INSERT INTO paginas_publicas (
		empresa_id, slug, titulo, descripcion, video_url, activo, creado_en, actualizado_en
	) VALUES (?, ?, ?, ?, ?, ?, datetime('now','localtime'), datetime('now','localtime'))`,
		empresaID, strings.TrimSpace(slug), strings.TrimSpace(titulo), strings.TrimSpace(descripcion), strings.TrimSpace(videoURL), ventaPublicaBoolToInt(activo))
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func CreateProductoPublico(db *sql.DB, paginaID int64, nombre, descripcion string, precioCents int64, moneda string, stock sql.NullInt64, sku, youtubeURL string, activo bool) (int64, error) {
	if paginaID <= 0 || strings.TrimSpace(nombre) == "" {
		return 0, fmt.Errorf("pagina_id y nombre son obligatorios")
	}
	res, err := db.Exec(`INSERT INTO productos_publicos (
		pagina_id, nombre, descripcion, precio_cents, moneda, stock, sku, youtube_url, activo, creado_en, actualizado_en
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now','localtime'), datetime('now','localtime'))`,
		paginaID, strings.TrimSpace(nombre), strings.TrimSpace(descripcion), precioCents, strings.TrimSpace(moneda), nullableInt64Value(stock), strings.TrimSpace(sku), strings.TrimSpace(youtubeURL), ventaPublicaBoolToInt(activo))
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func nullableInt64Value(v sql.NullInt64) interface{} {
	if v.Valid {
		return v.Int64
	}
	return nil
}

func AddImagenProductoPublico(db *sql.DB, productoID int64, url string, orden int) error {
	if productoID <= 0 || strings.TrimSpace(url) == "" {
		return fmt.Errorf("producto_id y url son obligatorios")
	}
	_, err := db.Exec(`INSERT INTO imagenes_productos_publicos (producto_id, url, orden) VALUES (?, ?, ?)`, productoID, strings.TrimSpace(url), orden)
	return err
}

func GetPaginaPublicaBySlug(db *sql.DB, slug string) (PaginaPublica, error) {
	var p PaginaPublica
	row := db.QueryRow(`SELECT id, empresa_id, slug, titulo, COALESCE(descripcion,''), COALESCE(video_url,''), activo FROM paginas_publicas WHERE slug = ? AND activo = 1`, strings.TrimSpace(slug))
	var activoInt int
	if err := row.Scan(&p.ID, &p.EmpresaID, &p.Slug, &p.Titulo, &p.Descripcion, &p.VideoURL, &activoInt); err != nil {
		return p, err
	}
	p.Activo = activoInt == 1
	return p, nil
}

func ListPaginasPublicasByEmpresa(db *sql.DB, empresaID int64) ([]PaginaPublica, error) {
	rows, err := db.Query(`SELECT id, empresa_id, slug, titulo, COALESCE(descripcion,''), COALESCE(video_url,''), activo FROM paginas_publicas WHERE empresa_id = ? ORDER BY id DESC`, empresaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []PaginaPublica{}
	for rows.Next() {
		var p PaginaPublica
		var activoInt int
		if err := rows.Scan(&p.ID, &p.EmpresaID, &p.Slug, &p.Titulo, &p.Descripcion, &p.VideoURL, &activoInt); err != nil {
			return nil, err
		}
		p.Activo = activoInt == 1
		out = append(out, p)
	}
	return out, rows.Err()
}


const (
	ventaPublicaWompiModeSandbox = "sandbox"
	ventaPublicaWompiModeReal    = "real"
	ventaPublicaEpaycoModeSandbox   = "sandbox"
	ventaPublicaEpaycoModeProduction = "production"
)

// EmpresaVentaPublicaConfig define la configuracion de catalogo/pagos publicos por empresa.
type EmpresaVentaPublicaConfig struct {
	ID                 int64  `json:"id"`
	EmpresaID          int64  `json:"empresa_id"`
	EmpresaSlug        string `json:"empresa_slug"`
	NombreTienda       string `json:"nombre_tienda"`
	DescripcionTienda  string `json:"descripcion_tienda,omitempty"`
	LogoURL            string `json:"logo_url,omitempty"`
	BannerURL          string `json:"banner_url,omitempty"`
	ColorPrimario      string `json:"color_primario,omitempty"`
	TemaVisual         string `json:"tema_visual,omitempty"`
	Moneda             string `json:"moneda"`
	DominioPublico     string `json:"dominio_publico,omitempty"`
	MostrarStock       bool   `json:"mostrar_stock"`
	WompiActivo        bool   `json:"wompi_activo"`
	WompiMode          string `json:"wompi_mode"`
	WompiPublicKey     string `json:"wompi_public_key,omitempty"`
	WompiPrivateKeyRef string `json:"wompi_private_key_ref,omitempty"`
	WompiIntegrityRef  string `json:"wompi_integrity_key_ref,omitempty"`
	WompiEventKeyRef   string `json:"wompi_event_key_ref,omitempty"`
	EpaycoActivo       bool   `json:"epayco_activo"`
	EpaycoMode         string `json:"epayco_mode"`
	EpaycoPublicKey    string `json:"epayco_public_key,omitempty"`
	EpaycoPrivateKeyRef string `json:"epayco_private_key_ref,omitempty"`
	EpaycoCustomerID   string `json:"epayco_customer_id,omitempty"`
	FechaCreacion      string `json:"fecha_creacion,omitempty"`
	FechaActualizacion string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string `json:"usuario_creador,omitempty"`
	Estado             string `json:"estado,omitempty"`
	Observaciones      string `json:"observaciones,omitempty"`
}

// EmpresaVentaPublicaItem representa un producto publicado para venta por internet.
type EmpresaVentaPublicaItem struct {
	ID                 int64   `json:"id"`
	EmpresaID          int64   `json:"empresa_id"`
	ProductoID         int64   `json:"producto_id,omitempty"`
	CodigoPublico      string  `json:"codigo_publico"`
	Nombre             string  `json:"nombre"`
	Descripcion        string  `json:"descripcion,omitempty"`
	Precio             float64 `json:"precio"`
	Moneda             string  `json:"moneda"`
	ImagenURL          string  `json:"imagen_url,omitempty"`
	StockPublicado     float64 `json:"stock_publicado,omitempty"`
	OrdenVisual        int     `json:"orden_visual,omitempty"`
	Destacado          bool    `json:"destacado"`
	FechaCreacion      string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string  `json:"usuario_creador,omitempty"`
	Estado             string  `json:"estado,omitempty"`
	Observaciones      string  `json:"observaciones,omitempty"`
}

// EmpresaVentaPublicaItemsFilter aplica filtros de listado para catalogo publico empresarial.
type EmpresaVentaPublicaItemsFilter struct {
	IncludeInactive bool
	Q               string
	Limit           int
	Offset          int
}

// EmpresaVentaPublicaOrder representa una orden creada desde la pagina publica.
type EmpresaVentaPublicaOrder struct {
	ID                  int64   `json:"id"`
	EmpresaID           int64   `json:"empresa_id"`
	CodigoOrden         string  `json:"codigo_orden"`
	CompradorNombre     string  `json:"comprador_nombre,omitempty"`
	CompradorEmail      string  `json:"comprador_email,omitempty"`
	CompradorTelefono   string  `json:"comprador_telefono,omitempty"`
	Moneda              string  `json:"moneda"`
	Subtotal            float64 `json:"subtotal"`
	DescuentoTotal      float64 `json:"descuento_total"`
	ImpuestoTotal       float64 `json:"impuesto_total"`
	Total               float64 `json:"total"`
	MetodoPago          string  `json:"metodo_pago"`
	EstadoPago          string  `json:"estado_pago"`
	ReferenciaExterna   string  `json:"referencia_externa,omitempty"`
	TransactionID       string  `json:"transaction_id,omitempty"`
	ItemsJSON           string  `json:"items_json,omitempty"`
	PasarelaPayloadJSON string  `json:"pasarela_payload_json,omitempty"`
	PagadoEn            string  `json:"pagado_en,omitempty"`
	FechaCreacion       string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion  string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador      string  `json:"usuario_creador,omitempty"`
	Estado              string  `json:"estado,omitempty"`
	Observaciones       string  `json:"observaciones,omitempty"`
}

// EmpresaVentaPublicaOrdersFilter aplica filtros de listado para ordenes publicas.
type EmpresaVentaPublicaOrdersFilter struct {
	IncludeInactive bool
	EstadoPago      string
	Q               string
	Limit           int
	Offset          int
}

func ventaPublicaNormalizeLimitOffset(limit, offset int) (int, int) {
	if limit <= 0 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}

func ventaPublicaNormalizeEstado(raw string) string {
	if strings.EqualFold(strings.TrimSpace(raw), "inactivo") {
		return "inactivo"
	}
	return "activo"
}

func ventaPublicaNormalizeMoneda(raw string) string {
	moneda := strings.ToUpper(strings.TrimSpace(raw))
	if moneda == "" {
		return "COP"
	}
	if len(moneda) > 8 {
		moneda = moneda[:8]
	}
	return moneda
}

func ventaPublicaNormalizeTemaVisual(raw string) string {
	tema := strings.ToLower(strings.TrimSpace(raw))
	switch tema {
	case "", "default":
		return "default"
	case "light", "minimal", "moderno":
		return tema
	default:
		return "default"
	}
}

func ventaPublicaNormalizeWompiMode(raw string) string {
	mode := strings.ToLower(strings.TrimSpace(raw))
	switch mode {
	case ventaPublicaWompiModeSandbox, ventaPublicaWompiModeReal:
		return mode
	default:
		return ventaPublicaWompiModeSandbox
	}
}

func ventaPublicaNormalizeEpaycoMode(raw string) string {
	mode := strings.ToLower(strings.TrimSpace(raw))
	switch mode {
	case ventaPublicaEpaycoModeSandbox, ventaPublicaEpaycoModeProduction:
		return mode
	case "real", "prod":
		return ventaPublicaEpaycoModeProduction
	default:
		return ventaPublicaEpaycoModeSandbox
	}
}

func ventaPublicaLikePattern(raw string) string {
	value := strings.TrimSpace(raw)
	value = strings.ReplaceAll(value, "!", "!!")
	value = strings.ReplaceAll(value, "%", "!%")
	value = strings.ReplaceAll(value, "_", "!_")
	return "%" + value + "%"
}

func ventaPublicaBoolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

func ventaPublicaIntToBool(v int64) bool {
	return v > 0
}

// NormalizeEmpresaPublicSlug genera un slug URL-safe para URL publica por empresa.
func NormalizeEmpresaPublicSlug(raw string) string {
	raw = strings.TrimSpace(strings.ToLower(raw))
	if raw == "" {
		return "empresa"
	}

	replacements := map[rune]string{
		'á': "a",
		'à': "a",
		'ä': "a",
		'â': "a",
		'ã': "a",
		'é': "e",
		'è': "e",
		'ë': "e",
		'ê': "e",
		'í': "i",
		'ì': "i",
		'ï': "i",
		'î': "i",
		'ó': "o",
		'ò': "o",
		'ö': "o",
		'ô': "o",
		'õ': "o",
		'ú': "u",
		'ù': "u",
		'ü': "u",
		'û': "u",
		'ñ': "n",
		'ç': "c",
	}

	var b strings.Builder
	prevDash := false
	for _, r := range raw {
		if mapped, ok := replacements[r]; ok {
			for _, mr := range mapped {
				b.WriteRune(mr)
			}
			prevDash = false
			continue
		}
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			prevDash = false
			continue
		}
		switch r {
		case ' ', '-', '_', '.', '/', '\\':
			if b.Len() > 0 && !prevDash {
				b.WriteByte('-')
				prevDash = true
			}
		}
	}

	slug := strings.Trim(b.String(), "-")
	if slug == "" {
		return "empresa"
	}
	if len(slug) > 120 {
		slug = strings.Trim(slug[:120], "-")
	}
	if slug == "" {
		return "empresa"
	}
	return slug
}

func ventaPublicaGenerateItemCode(empresaID int64, nombre string) string {
	base := NormalizeEmpresaPublicSlug(nombre)
	if base == "" || base == "empresa" {
		base = "item"
	}
	if len(base) > 22 {
		base = base[:22]
	}
	stamp := time.Now().In(time.Local).Format("20060102150405")
	return strings.ToUpper(fmt.Sprintf("VP-%d-%s-%s", empresaID, base, stamp))
}

func ventaPublicaGenerateOrderCode(empresaID int64) string {
	stamp := time.Now().In(time.Local).Format("20060102150405")
	return fmt.Sprintf("VP-ORD-%d-%s", empresaID, stamp)
}

func getEmpresaNombreByID(dbConn *sql.DB, empresaID int64) (string, error) {
	if dbConn == nil {
		return "", errors.New("db connection is nil")
	}
	if empresaID <= 0 {
		return "", fmt.Errorf("empresa_id invalido")
	}
	var nombre string
	err := dbConn.QueryRow(`SELECT COALESCE(nombre, '') FROM empresas WHERE id = ? OR COALESCE(empresa_id, 0) = ? ORDER BY id ASC LIMIT 1`, empresaID, empresaID).Scan(&nombre)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(nombre), nil
}

// EnsureEmpresaVentaPublicaSchema crea y migra tablas para venta publica empresarial.
func EnsureEmpresaVentaPublicaSchema(dbConn *sql.DB) error {
	if dbConn == nil {
		return errors.New("db connection is nil")
	}

	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_venta_publica_configuracion (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL UNIQUE,
			empresa_slug TEXT NOT NULL,
			nombre_tienda TEXT,
			descripcion_tienda TEXT,
			logo_url TEXT,
			banner_url TEXT,
			color_primario TEXT DEFAULT '#0f4c81',
			tema_visual TEXT DEFAULT 'default',
			moneda TEXT DEFAULT 'COP',
			dominio_publico TEXT,
			mostrar_stock INTEGER DEFAULT 1,
			wompi_activo INTEGER DEFAULT 0,
			wompi_mode TEXT DEFAULT 'sandbox',
			wompi_public_key TEXT,
			wompi_private_key_ref TEXT,
			wompi_integrity_key_ref TEXT,
			wompi_event_key_ref TEXT,
			epayco_activo INTEGER DEFAULT 0,
			epayco_mode TEXT DEFAULT 'sandbox',
			epayco_public_key TEXT,
			epayco_private_key_ref TEXT,
			epayco_customer_id TEXT,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_venta_publica_cfg_slug ON empresa_venta_publica_configuracion(empresa_slug);`,
		`CREATE INDEX IF NOT EXISTS ix_venta_publica_cfg_empresa_estado ON empresa_venta_publica_configuracion(empresa_id, estado);`,
		`CREATE TABLE IF NOT EXISTS empresa_venta_publica_items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			producto_id INTEGER DEFAULT 0,
			codigo_publico TEXT NOT NULL,
			nombre TEXT NOT NULL,
			descripcion TEXT,
			precio REAL DEFAULT 0,
			moneda TEXT DEFAULT 'COP',
			imagen_url TEXT,
			stock_publicado REAL DEFAULT 0,
			orden_visual INTEGER DEFAULT 0,
			destacado INTEGER DEFAULT 0,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			UNIQUE(empresa_id, codigo_publico)
		);`,
		`CREATE INDEX IF NOT EXISTS ix_venta_publica_items_empresa_estado ON empresa_venta_publica_items(empresa_id, estado, orden_visual, id);`,
		`CREATE TABLE IF NOT EXISTS empresa_venta_publica_ordenes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			codigo_orden TEXT NOT NULL,
			comprador_nombre TEXT,
			comprador_email TEXT,
			comprador_telefono TEXT,
			moneda TEXT DEFAULT 'COP',
			subtotal REAL DEFAULT 0,
			descuento_total REAL DEFAULT 0,
			impuesto_total REAL DEFAULT 0,
			total REAL DEFAULT 0,
			metodo_pago TEXT DEFAULT 'wompi_nequi',
			estado_pago TEXT DEFAULT 'pendiente',
			referencia_externa TEXT,
			transaction_id TEXT,
			items_json TEXT,
			pasarela_payload_json TEXT,
			pagado_en TEXT,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			UNIQUE(empresa_id, codigo_orden)
		);`,
		`CREATE INDEX IF NOT EXISTS ix_venta_publica_ordenes_empresa_estado ON empresa_venta_publica_ordenes(empresa_id, estado_pago, fecha_creacion DESC);`,
		`CREATE INDEX IF NOT EXISTS ix_venta_publica_ordenes_tx ON empresa_venta_publica_ordenes(transaction_id);`,
	}

	for _, stmt := range stmts {
		if _, err := dbConn.Exec(stmt); err != nil {
			return err
		}
	}

	if err := ensureColumnIfMissing(dbConn, "empresa_venta_publica_configuracion", "empresa_slug", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_venta_publica_configuracion", "wompi_mode", "TEXT DEFAULT 'sandbox'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_venta_publica_configuracion", "wompi_public_key", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_venta_publica_configuracion", "wompi_private_key_ref", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_venta_publica_configuracion", "wompi_integrity_key_ref", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_venta_publica_configuracion", "wompi_event_key_ref", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_venta_publica_configuracion", "epayco_activo", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_venta_publica_configuracion", "epayco_mode", "TEXT DEFAULT 'sandbox'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_venta_publica_configuracion", "epayco_public_key", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_venta_publica_configuracion", "tema_visual", "TEXT DEFAULT 'default'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_venta_publica_configuracion", "epayco_private_key_ref", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_venta_publica_configuracion", "epayco_customer_id", "TEXT"); err != nil {
		return err
	}

	return nil
}

// GetEmpresaVentaPublicaConfig obtiene la configuracion de venta publica por empresa.
func GetEmpresaVentaPublicaConfig(dbConn *sql.DB, empresaID int64) (EmpresaVentaPublicaConfig, error) {
	if dbConn == nil {
		return EmpresaVentaPublicaConfig{}, errors.New("db connection is nil")
	}
	if empresaID <= 0 {
		return EmpresaVentaPublicaConfig{}, fmt.Errorf("empresa_id invalido")
	}

	var out EmpresaVentaPublicaConfig
	var mostrarStock sql.NullInt64
	var wompiActivo sql.NullInt64
	var epaycoActivo sql.NullInt64
	err := dbConn.QueryRow(`SELECT
		id,
		empresa_id,
		COALESCE(empresa_slug, ''),
		COALESCE(nombre_tienda, ''),
		COALESCE(descripcion_tienda, ''),
		COALESCE(logo_url, ''),
		COALESCE(banner_url, ''),
		COALESCE(color_primario, ''),
		COALESCE(tema_visual, ''),
		COALESCE(moneda, 'COP'),
		COALESCE(dominio_publico, ''),
		COALESCE(mostrar_stock, 1),
		COALESCE(wompi_activo, 0),
		COALESCE(wompi_mode, 'sandbox'),
		COALESCE(wompi_public_key, ''),
		COALESCE(wompi_private_key_ref, ''),
		COALESCE(wompi_integrity_key_ref, ''),
		COALESCE(wompi_event_key_ref, ''),
		COALESCE(epayco_activo, 0),
		COALESCE(epayco_mode, 'sandbox'),
		COALESCE(epayco_public_key, ''),
		COALESCE(epayco_private_key_ref, ''),
		COALESCE(epayco_customer_id, ''),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM empresa_venta_publica_configuracion
	WHERE empresa_id = ?
	LIMIT 1`, empresaID).Scan(
		&out.ID,
		&out.EmpresaID,
		&out.EmpresaSlug,
		&out.NombreTienda,
		&out.DescripcionTienda,
		&out.LogoURL,
		&out.BannerURL,
		&out.ColorPrimario,
		&out.TemaVisual,
		&out.Moneda,
		&out.DominioPublico,
		&mostrarStock,
		&wompiActivo,
		&out.WompiMode,
		&out.WompiPublicKey,
		&out.WompiPrivateKeyRef,
		&out.WompiIntegrityRef,
		&out.WompiEventKeyRef,
		&epaycoActivo,
		&out.EpaycoMode,
		&out.EpaycoPublicKey,
		&out.EpaycoPrivateKeyRef,
		&out.EpaycoCustomerID,
		&out.FechaCreacion,
		&out.FechaActualizacion,
		&out.UsuarioCreador,
		&out.Estado,
		&out.Observaciones,
	)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return EmpresaVentaPublicaConfig{}, err
		}
		nombre, _ := getEmpresaNombreByID(dbConn, empresaID)
		if strings.TrimSpace(nombre) == "" {
			nombre = fmt.Sprintf("Empresa %d", empresaID)
		}
		out = EmpresaVentaPublicaConfig{
			EmpresaID:     empresaID,
			EmpresaSlug:   NormalizeEmpresaPublicSlug(nombre),
			NombreTienda:  nombre,
			ColorPrimario: "#0f4c81",
			TemaVisual:    "default",
			Moneda:        "COP",
			MostrarStock:  true,
			WompiActivo:   false,
			WompiMode:     ventaPublicaWompiModeSandbox,
			EpaycoActivo:  false,
			EpaycoMode:    ventaPublicaEpaycoModeSandbox,
			Estado:        "activo",
		}
		return out, nil
	}

	if out.EmpresaID <= 0 {
		out.EmpresaID = empresaID
	}
	out.EmpresaSlug = NormalizeEmpresaPublicSlug(out.EmpresaSlug)
	if out.EmpresaSlug == "empresa" {
		nombre, _ := getEmpresaNombreByID(dbConn, empresaID)
		if nombre != "" {
			out.EmpresaSlug = NormalizeEmpresaPublicSlug(nombre)
		}
	}
	if strings.TrimSpace(out.NombreTienda) == "" {
		nombre, _ := getEmpresaNombreByID(dbConn, empresaID)
		if nombre == "" {
			nombre = fmt.Sprintf("Empresa %d", empresaID)
		}
		out.NombreTienda = nombre
	}
	out.Moneda = ventaPublicaNormalizeMoneda(out.Moneda)
	out.TemaVisual = ventaPublicaNormalizeTemaVisual(out.TemaVisual)
	out.WompiMode = ventaPublicaNormalizeWompiMode(out.WompiMode)
	out.EpaycoMode = ventaPublicaNormalizeEpaycoMode(out.EpaycoMode)
	out.MostrarStock = mostrarStock.Valid && mostrarStock.Int64 > 0
	if !mostrarStock.Valid {
		out.MostrarStock = true
	}
	out.WompiActivo = wompiActivo.Valid && wompiActivo.Int64 > 0
	out.EpaycoActivo = epaycoActivo.Valid && epaycoActivo.Int64 > 0
	out.Estado = ventaPublicaNormalizeEstado(out.Estado)
	return out, nil
}

// UpsertEmpresaVentaPublicaConfig crea/actualiza configuracion de tienda publica por empresa.
func UpsertEmpresaVentaPublicaConfig(dbConn *sql.DB, cfg EmpresaVentaPublicaConfig) (int64, error) {
	if dbConn == nil {
		return 0, errors.New("db connection is nil")
	}
	if cfg.EmpresaID <= 0 {
		return 0, fmt.Errorf("empresa_id invalido")
	}

	nombreEmpresa, _ := getEmpresaNombreByID(dbConn, cfg.EmpresaID)
	if strings.TrimSpace(cfg.NombreTienda) == "" {
		if strings.TrimSpace(nombreEmpresa) != "" {
			cfg.NombreTienda = nombreEmpresa
		} else {
			cfg.NombreTienda = fmt.Sprintf("Empresa %d", cfg.EmpresaID)
		}
	}
	cfg.NombreTienda = strings.TrimSpace(cfg.NombreTienda)
	if cfg.EmpresaSlug == "" {
		cfg.EmpresaSlug = NormalizeEmpresaPublicSlug(cfg.NombreTienda)
	}
	cfg.EmpresaSlug = NormalizeEmpresaPublicSlug(cfg.EmpresaSlug)
	if cfg.EmpresaSlug == "" {
		cfg.EmpresaSlug = fmt.Sprintf("empresa-%d", cfg.EmpresaID)
	}
	cfg.Moneda = ventaPublicaNormalizeMoneda(cfg.Moneda)
	cfg.TemaVisual = ventaPublicaNormalizeTemaVisual(cfg.TemaVisual)
	cfg.WompiMode = ventaPublicaNormalizeWompiMode(cfg.WompiMode)
	cfg.EpaycoMode = ventaPublicaNormalizeEpaycoMode(cfg.EpaycoMode)
	cfg.Estado = ventaPublicaNormalizeEstado(cfg.Estado)
	if strings.TrimSpace(cfg.ColorPrimario) == "" {
		cfg.ColorPrimario = "#0f4c81"
	}

	var existingID int64
	err := dbConn.QueryRow(`SELECT id FROM empresa_venta_publica_configuracion WHERE empresa_id = ? LIMIT 1`, cfg.EmpresaID).Scan(&existingID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}

	if existingID > 0 {
		_, err = dbConn.Exec(`UPDATE empresa_venta_publica_configuracion
			SET empresa_slug = ?,
				nombre_tienda = ?,
				descripcion_tienda = ?,
				logo_url = ?,
				banner_url = ?,
			color_primario = ?,
			tema_visual = ?,
				moneda = ?,
				dominio_publico = ?,
				mostrar_stock = ?,
				wompi_activo = ?,
				wompi_mode = ?,
				wompi_public_key = ?,
				wompi_private_key_ref = ?,
				wompi_integrity_key_ref = ?,
				wompi_event_key_ref = ?,
				epayco_activo = ?,
				epayco_mode = ?,
				epayco_public_key = ?,
				epayco_private_key_ref = ?,
				epayco_customer_id = ?,
				usuario_creador = ?,
				estado = ?,
				observaciones = ?,
				fecha_actualizacion = datetime('now','localtime')
			WHERE empresa_id = ?`,
			cfg.EmpresaSlug,
			cfg.NombreTienda,
			strings.TrimSpace(cfg.DescripcionTienda),
			strings.TrimSpace(cfg.LogoURL),
			strings.TrimSpace(cfg.BannerURL),
			strings.TrimSpace(cfg.ColorPrimario),
			strings.TrimSpace(cfg.TemaVisual),
			cfg.Moneda,
			strings.TrimSpace(cfg.DominioPublico),
			ventaPublicaBoolToInt(cfg.MostrarStock),
			ventaPublicaBoolToInt(cfg.WompiActivo),
			cfg.WompiMode,
			strings.TrimSpace(cfg.WompiPublicKey),
			strings.TrimSpace(cfg.WompiPrivateKeyRef),
			strings.TrimSpace(cfg.WompiIntegrityRef),
			strings.TrimSpace(cfg.WompiEventKeyRef),
			ventaPublicaBoolToInt(cfg.EpaycoActivo),
			cfg.EpaycoMode,
			strings.TrimSpace(cfg.EpaycoPublicKey),
			strings.TrimSpace(cfg.EpaycoPrivateKeyRef),
			strings.TrimSpace(cfg.EpaycoCustomerID),
			strings.TrimSpace(cfg.UsuarioCreador),
			cfg.Estado,
			strings.TrimSpace(cfg.Observaciones),
			cfg.EmpresaID,
		)
		if err != nil {
			return 0, err
		}
		return existingID, nil
	}

	res, err := dbConn.Exec(`INSERT INTO empresa_venta_publica_configuracion (
		empresa_id,
		empresa_slug,
		nombre_tienda,
		descripcion_tienda,
		logo_url,
		banner_url,
		color_primario,
		tema_visual,
		moneda,
		dominio_publico,
		mostrar_stock,
		wompi_activo,
		wompi_mode,
		wompi_public_key,
		wompi_private_key_ref,
		wompi_integrity_key_ref,
		wompi_event_key_ref,
		epayco_activo,
		epayco_mode,
		epayco_public_key,
		epayco_private_key_ref,
		epayco_customer_id,
		usuario_creador,
		estado,
		observaciones
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		cfg.EmpresaID,
		cfg.EmpresaSlug,
		cfg.NombreTienda,
		strings.TrimSpace(cfg.DescripcionTienda),
		strings.TrimSpace(cfg.LogoURL),
		strings.TrimSpace(cfg.BannerURL),
		strings.TrimSpace(cfg.ColorPrimario),
		strings.TrimSpace(cfg.TemaVisual),
		cfg.Moneda,
		strings.TrimSpace(cfg.DominioPublico),
		ventaPublicaBoolToInt(cfg.MostrarStock),
		ventaPublicaBoolToInt(cfg.WompiActivo),
		cfg.WompiMode,
		strings.TrimSpace(cfg.WompiPublicKey),
		strings.TrimSpace(cfg.WompiPrivateKeyRef),
		strings.TrimSpace(cfg.WompiIntegrityRef),
		strings.TrimSpace(cfg.WompiEventKeyRef),
		ventaPublicaBoolToInt(cfg.EpaycoActivo),
		cfg.EpaycoMode,
		strings.TrimSpace(cfg.EpaycoPublicKey),
		strings.TrimSpace(cfg.EpaycoPrivateKeyRef),
		strings.TrimSpace(cfg.EpaycoCustomerID),
		strings.TrimSpace(cfg.UsuarioCreador),
		cfg.Estado,
		strings.TrimSpace(cfg.Observaciones),
	)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

// ResolveEmpresaIDByVentaPublicaSlug resuelve empresa_id a partir del slug publico.
func ResolveEmpresaIDByVentaPublicaSlug(dbConn *sql.DB, slug string) (int64, error) {
	if dbConn == nil {
		return 0, errors.New("db connection is nil")
	}
	slug = NormalizeEmpresaPublicSlug(slug)
	if strings.TrimSpace(slug) == "" {
		return 0, fmt.Errorf("empresa_slug invalido")
	}

	var empresaID int64
	err := dbConn.QueryRow(`SELECT empresa_id FROM empresa_venta_publica_configuracion WHERE empresa_slug = ? AND COALESCE(estado, 'activo') <> 'inactivo' LIMIT 1`, slug).Scan(&empresaID)
	if err == nil && empresaID > 0 {
		return empresaID, nil
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}

	rows, err := dbConn.Query(`SELECT id, COALESCE(empresa_id, id), COALESCE(nombre, '') FROM empresas WHERE COALESCE(estado, 'activo') <> 'inactivo' ORDER BY id ASC`)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var empresaIDAlt int64
		var nombre string
		if err := rows.Scan(&id, &empresaIDAlt, &nombre); err != nil {
			return 0, err
		}
		if NormalizeEmpresaPublicSlug(nombre) == slug {
			if empresaIDAlt > 0 {
				return empresaIDAlt, nil
			}
			return id, nil
		}
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}
	return 0, sql.ErrNoRows
}

func hydrateEmpresaVentaPublicaItemFromProducto(dbConn *sql.DB, item *EmpresaVentaPublicaItem) error {
	if dbConn == nil || item == nil {
		return nil
	}
	if item.EmpresaID <= 0 || item.ProductoID <= 0 {
		return nil
	}

	var nombre string
	var descripcion string
	var precio float64
	var imagenURL string
	var codigoBarras string
	var sku string
	err := dbConn.QueryRow(`SELECT
		COALESCE(nombre, ''),
		COALESCE(descripcion, ''),
		COALESCE(precio, 0),
		COALESCE(imagen_url, ''),
		COALESCE(codigo_barras, ''),
		COALESCE(sku, '')
	FROM productos
	WHERE empresa_id = ? AND id = ?
	LIMIT 1`, item.EmpresaID, item.ProductoID).Scan(&nombre, &descripcion, &precio, &imagenURL, &codigoBarras, &sku)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("producto_id no encontrado para la empresa")
		}
		return err
	}

	if strings.TrimSpace(item.Nombre) == "" {
		item.Nombre = strings.TrimSpace(nombre)
	}
	if strings.TrimSpace(item.Descripcion) == "" {
		item.Descripcion = strings.TrimSpace(descripcion)
	}
	if strings.TrimSpace(item.ImagenURL) == "" {
		item.ImagenURL = strings.TrimSpace(imagenURL)
	}
	if item.Precio <= 0 && precio > 0 {
		item.Precio = precio
	}
	if strings.TrimSpace(item.CodigoPublico) == "" {
		if strings.TrimSpace(codigoBarras) != "" {
			item.CodigoPublico = strings.TrimSpace(codigoBarras)
		} else if strings.TrimSpace(sku) != "" {
			item.CodigoPublico = strings.TrimSpace(sku)
		}
	}
	return nil
}

// CreateEmpresaVentaPublicaItem crea un item para catalogo de venta publica.
func CreateEmpresaVentaPublicaItem(dbConn *sql.DB, item EmpresaVentaPublicaItem) (int64, error) {
	if dbConn == nil {
		return 0, errors.New("db connection is nil")
	}
	if item.EmpresaID <= 0 {
		return 0, fmt.Errorf("empresa_id invalido")
	}
	if item.Precio < 0 {
		return 0, fmt.Errorf("precio invalido")
	}
	if item.StockPublicado < 0 {
		return 0, fmt.Errorf("stock_publicado invalido")
	}
	if err := hydrateEmpresaVentaPublicaItemFromProducto(dbConn, &item); err != nil {
		return 0, err
	}

	item.Nombre = strings.TrimSpace(item.Nombre)
	if item.Nombre == "" {
		return 0, fmt.Errorf("nombre es obligatorio")
	}
	if item.CodigoPublico = strings.TrimSpace(item.CodigoPublico); item.CodigoPublico == "" {
		item.CodigoPublico = ventaPublicaGenerateItemCode(item.EmpresaID, item.Nombre)
	}
	item.Moneda = ventaPublicaNormalizeMoneda(item.Moneda)
	item.Estado = ventaPublicaNormalizeEstado(item.Estado)

	res, err := dbConn.Exec(`INSERT INTO empresa_venta_publica_items (
		empresa_id,
		producto_id,
		codigo_publico,
		nombre,
		descripcion,
		precio,
		moneda,
		imagen_url,
		stock_publicado,
		orden_visual,
		destacado,
		usuario_creador,
		estado,
		observaciones
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		item.EmpresaID,
		item.ProductoID,
		item.CodigoPublico,
		item.Nombre,
		strings.TrimSpace(item.Descripcion),
		item.Precio,
		item.Moneda,
		strings.TrimSpace(item.ImagenURL),
		item.StockPublicado,
		item.OrdenVisual,
		ventaPublicaBoolToInt(item.Destacado),
		strings.TrimSpace(item.UsuarioCreador),
		item.Estado,
		strings.TrimSpace(item.Observaciones),
	)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

// UpdateEmpresaVentaPublicaItem actualiza un item del catalogo publico.
func UpdateEmpresaVentaPublicaItem(dbConn *sql.DB, item EmpresaVentaPublicaItem) error {
	if dbConn == nil {
		return errors.New("db connection is nil")
	}
	if item.EmpresaID <= 0 || item.ID <= 0 {
		return fmt.Errorf("empresa_id/id invalidos")
	}
	if item.Precio < 0 {
		return fmt.Errorf("precio invalido")
	}
	if item.StockPublicado < 0 {
		return fmt.Errorf("stock_publicado invalido")
	}
	if err := hydrateEmpresaVentaPublicaItemFromProducto(dbConn, &item); err != nil {
		return err
	}
	item.Nombre = strings.TrimSpace(item.Nombre)
	if item.Nombre == "" {
		return fmt.Errorf("nombre es obligatorio")
	}
	item.CodigoPublico = strings.TrimSpace(item.CodigoPublico)
	if item.CodigoPublico == "" {
		item.CodigoPublico = ventaPublicaGenerateItemCode(item.EmpresaID, item.Nombre)
	}
	item.Moneda = ventaPublicaNormalizeMoneda(item.Moneda)
	item.Estado = ventaPublicaNormalizeEstado(item.Estado)

	res, err := dbConn.Exec(`UPDATE empresa_venta_publica_items
		SET producto_id = ?,
			codigo_publico = ?,
			nombre = ?,
			descripcion = ?,
			precio = ?,
			moneda = ?,
			imagen_url = ?,
			stock_publicado = ?,
			orden_visual = ?,
			destacado = ?,
			usuario_creador = ?,
			estado = ?,
			observaciones = ?,
			fecha_actualizacion = datetime('now','localtime')
		WHERE empresa_id = ? AND id = ?`,
		item.ProductoID,
		item.CodigoPublico,
		item.Nombre,
		strings.TrimSpace(item.Descripcion),
		item.Precio,
		item.Moneda,
		strings.TrimSpace(item.ImagenURL),
		item.StockPublicado,
		item.OrdenVisual,
		ventaPublicaBoolToInt(item.Destacado),
		strings.TrimSpace(item.UsuarioCreador),
		item.Estado,
		strings.TrimSpace(item.Observaciones),
		item.EmpresaID,
		item.ID,
	)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected <= 0 {
		return sql.ErrNoRows
	}
	return nil
}

// SetEmpresaVentaPublicaItemEstadoByID activa/desactiva un item de catalogo publico.
func SetEmpresaVentaPublicaItemEstadoByID(dbConn *sql.DB, empresaID, itemID int64, estado string) error {
	if dbConn == nil {
		return errors.New("db connection is nil")
	}
	if empresaID <= 0 || itemID <= 0 {
		return fmt.Errorf("empresa_id/id invalidos")
	}
	estado = ventaPublicaNormalizeEstado(estado)
	res, err := dbConn.Exec(`UPDATE empresa_venta_publica_items SET estado = ?, fecha_actualizacion = datetime('now','localtime') WHERE empresa_id = ? AND id = ?`, estado, empresaID, itemID)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected <= 0 {
		return sql.ErrNoRows
	}
	return nil
}

// GetEmpresaVentaPublicaItemByID obtiene un item por id/empresa.
func GetEmpresaVentaPublicaItemByID(dbConn *sql.DB, empresaID, itemID int64) (EmpresaVentaPublicaItem, error) {
	if dbConn == nil {
		return EmpresaVentaPublicaItem{}, errors.New("db connection is nil")
	}
	if empresaID <= 0 || itemID <= 0 {
		return EmpresaVentaPublicaItem{}, fmt.Errorf("empresa_id/id invalidos")
	}

	var out EmpresaVentaPublicaItem
	var destacado sql.NullInt64
	err := dbConn.QueryRow(`SELECT
		id,
		empresa_id,
		COALESCE(producto_id, 0),
		COALESCE(codigo_publico, ''),
		COALESCE(nombre, ''),
		COALESCE(descripcion, ''),
		COALESCE(precio, 0),
		COALESCE(moneda, 'COP'),
		COALESCE(imagen_url, ''),
		COALESCE(stock_publicado, 0),
		COALESCE(orden_visual, 0),
		COALESCE(destacado, 0),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM empresa_venta_publica_items
	WHERE empresa_id = ? AND id = ?
	LIMIT 1`, empresaID, itemID).Scan(
		&out.ID,
		&out.EmpresaID,
		&out.ProductoID,
		&out.CodigoPublico,
		&out.Nombre,
		&out.Descripcion,
		&out.Precio,
		&out.Moneda,
		&out.ImagenURL,
		&out.StockPublicado,
		&out.OrdenVisual,
		&destacado,
		&out.FechaCreacion,
		&out.FechaActualizacion,
		&out.UsuarioCreador,
		&out.Estado,
		&out.Observaciones,
	)
	if err != nil {
		return EmpresaVentaPublicaItem{}, err
	}
	out.Destacado = destacado.Valid && destacado.Int64 > 0
	out.Moneda = ventaPublicaNormalizeMoneda(out.Moneda)
	out.Estado = ventaPublicaNormalizeEstado(out.Estado)
	return out, nil
}

// ListEmpresaVentaPublicaItems lista items publicados por empresa con filtros.
func ListEmpresaVentaPublicaItems(dbConn *sql.DB, empresaID int64, filter EmpresaVentaPublicaItemsFilter) ([]EmpresaVentaPublicaItem, int64, error) {
	if dbConn == nil {
		return nil, 0, errors.New("db connection is nil")
	}
	if empresaID <= 0 {
		return nil, 0, fmt.Errorf("empresa_id invalido")
	}

	limit, offset := ventaPublicaNormalizeLimitOffset(filter.Limit, filter.Offset)
	where := `WHERE empresa_id = ?`
	args := []interface{}{empresaID}
	if !filter.IncludeInactive {
		where += ` AND COALESCE(estado, 'activo') <> 'inactivo'`
	}
	if q := strings.TrimSpace(filter.Q); q != "" {
		pattern := ventaPublicaLikePattern(q)
		where += ` AND (
			LOWER(COALESCE(codigo_publico, '')) LIKE LOWER(?) ESCAPE '!' OR
			LOWER(COALESCE(nombre, '')) LIKE LOWER(?) ESCAPE '!' OR
			LOWER(COALESCE(descripcion, '')) LIKE LOWER(?) ESCAPE '!'
		)`
		args = append(args, pattern, pattern, pattern)
	}

	countQuery := `SELECT COUNT(1) FROM empresa_venta_publica_items ` + where
	var total int64
	if err := dbConn.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `SELECT
		id,
		empresa_id,
		COALESCE(producto_id, 0),
		COALESCE(codigo_publico, ''),
		COALESCE(nombre, ''),
		COALESCE(descripcion, ''),
		COALESCE(precio, 0),
		COALESCE(moneda, 'COP'),
		COALESCE(imagen_url, ''),
		COALESCE(stock_publicado, 0),
		COALESCE(orden_visual, 0),
		COALESCE(destacado, 0),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM empresa_venta_publica_items ` + where + `
	ORDER BY COALESCE(orden_visual, 0) ASC, id DESC
	LIMIT ? OFFSET ?`

	rows, err := dbConn.Query(query, append(args, limit, offset)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := make([]EmpresaVentaPublicaItem, 0)
	for rows.Next() {
		var item EmpresaVentaPublicaItem
		var destacado sql.NullInt64
		if err := rows.Scan(
			&item.ID,
			&item.EmpresaID,
			&item.ProductoID,
			&item.CodigoPublico,
			&item.Nombre,
			&item.Descripcion,
			&item.Precio,
			&item.Moneda,
			&item.ImagenURL,
			&item.StockPublicado,
			&item.OrdenVisual,
			&destacado,
			&item.FechaCreacion,
			&item.FechaActualizacion,
			&item.UsuarioCreador,
			&item.Estado,
			&item.Observaciones,
		); err != nil {
			return nil, 0, err
		}
		item.Destacado = destacado.Valid && destacado.Int64 > 0
		item.Moneda = ventaPublicaNormalizeMoneda(item.Moneda)
		item.Estado = ventaPublicaNormalizeEstado(item.Estado)
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

// ListEmpresaVentaPublicaItemsPublic lista solo items activos para consumo publico.
func ListEmpresaVentaPublicaItemsPublic(dbConn *sql.DB, empresaID int64) ([]EmpresaVentaPublicaItem, error) {
	rows, _, err := ListEmpresaVentaPublicaItems(dbConn, empresaID, EmpresaVentaPublicaItemsFilter{
		IncludeInactive: false,
		Limit:           500,
		Offset:          0,
	})
	if err != nil {
		return nil, err
	}
	return rows, nil
}

// CreateEmpresaVentaPublicaOrder crea una orden publica.
func CreateEmpresaVentaPublicaOrder(dbConn *sql.DB, order EmpresaVentaPublicaOrder) (int64, error) {
	if dbConn == nil {
		return 0, errors.New("db connection is nil")
	}
	if order.EmpresaID <= 0 {
		return 0, fmt.Errorf("empresa_id invalido")
	}
	if order.Total < 0 || order.Subtotal < 0 || order.ImpuestoTotal < 0 || order.DescuentoTotal < 0 {
		return 0, fmt.Errorf("totales invalidos")
	}
	if strings.TrimSpace(order.CodigoOrden) == "" {
		order.CodigoOrden = ventaPublicaGenerateOrderCode(order.EmpresaID)
	}
	if strings.TrimSpace(order.EstadoPago) == "" {
		order.EstadoPago = "pendiente"
	}
	if strings.TrimSpace(order.MetodoPago) == "" {
		order.MetodoPago = "wompi_nequi"
	}
	order.Moneda = ventaPublicaNormalizeMoneda(order.Moneda)
	order.Estado = ventaPublicaNormalizeEstado(order.Estado)
	if strings.TrimSpace(order.Estado) == "" {
		order.Estado = "activo"
	}

	res, err := dbConn.Exec(`INSERT INTO empresa_venta_publica_ordenes (
		empresa_id,
		codigo_orden,
		comprador_nombre,
		comprador_email,
		comprador_telefono,
		moneda,
		subtotal,
		descuento_total,
		impuesto_total,
		total,
		metodo_pago,
		estado_pago,
		referencia_externa,
		transaction_id,
		items_json,
		pasarela_payload_json,
		pagado_en,
		usuario_creador,
		estado,
		observaciones
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		order.EmpresaID,
		order.CodigoOrden,
		strings.TrimSpace(order.CompradorNombre),
		strings.TrimSpace(order.CompradorEmail),
		strings.TrimSpace(order.CompradorTelefono),
		order.Moneda,
		order.Subtotal,
		order.DescuentoTotal,
		order.ImpuestoTotal,
		order.Total,
		strings.TrimSpace(order.MetodoPago),
		strings.TrimSpace(order.EstadoPago),
		strings.TrimSpace(order.ReferenciaExterna),
		strings.TrimSpace(order.TransactionID),
		strings.TrimSpace(order.ItemsJSON),
		strings.TrimSpace(order.PasarelaPayloadJSON),
		strings.TrimSpace(order.PagadoEn),
		strings.TrimSpace(order.UsuarioCreador),
		order.Estado,
		strings.TrimSpace(order.Observaciones),
	)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

// UpdateEmpresaVentaPublicaOrderPayment actualiza estado de pago/transaccion de una orden publica.
func UpdateEmpresaVentaPublicaOrderPayment(dbConn *sql.DB, empresaID int64, codigoOrden, estadoPago, transactionID, referenciaExterna, payloadJSON, pagadoEn, observaciones string) error {
	if dbConn == nil {
		return errors.New("db connection is nil")
	}
	if empresaID <= 0 {
		return fmt.Errorf("empresa_id invalido")
	}
	codigoOrden = strings.TrimSpace(codigoOrden)
	if codigoOrden == "" {
		return fmt.Errorf("codigo_orden invalido")
	}

	if strings.TrimSpace(estadoPago) == "" {
		estadoPago = "pendiente"
	}
	res, err := dbConn.Exec(`UPDATE empresa_venta_publica_ordenes
		SET estado_pago = ?,
			transaction_id = ?,
			referencia_externa = ?,
			pasarela_payload_json = CASE WHEN ? = '' THEN pasarela_payload_json ELSE ? END,
			pagado_en = CASE WHEN ? = '' THEN pagado_en ELSE ? END,
			observaciones = CASE WHEN ? = '' THEN observaciones ELSE ? END,
			fecha_actualizacion = datetime('now','localtime')
		WHERE empresa_id = ? AND codigo_orden = ?`,
		strings.TrimSpace(estadoPago),
		strings.TrimSpace(transactionID),
		strings.TrimSpace(referenciaExterna),
		strings.TrimSpace(payloadJSON),
		strings.TrimSpace(payloadJSON),
		strings.TrimSpace(pagadoEn),
		strings.TrimSpace(pagadoEn),
		strings.TrimSpace(observaciones),
		strings.TrimSpace(observaciones),
		empresaID,
		codigoOrden,
	)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected <= 0 {
		return sql.ErrNoRows
	}
	return nil
}

// GetEmpresaVentaPublicaOrderByCodigo obtiene una orden publica por codigo.
func GetEmpresaVentaPublicaOrderByCodigo(dbConn *sql.DB, empresaID int64, codigoOrden string) (EmpresaVentaPublicaOrder, error) {
	if dbConn == nil {
		return EmpresaVentaPublicaOrder{}, errors.New("db connection is nil")
	}
	if empresaID <= 0 {
		return EmpresaVentaPublicaOrder{}, fmt.Errorf("empresa_id invalido")
	}
	codigoOrden = strings.TrimSpace(codigoOrden)
	if codigoOrden == "" {
		return EmpresaVentaPublicaOrder{}, fmt.Errorf("codigo_orden invalido")
	}

	var out EmpresaVentaPublicaOrder
	err := dbConn.QueryRow(`SELECT
		id,
		empresa_id,
		COALESCE(codigo_orden, ''),
		COALESCE(comprador_nombre, ''),
		COALESCE(comprador_email, ''),
		COALESCE(comprador_telefono, ''),
		COALESCE(moneda, 'COP'),
		COALESCE(subtotal, 0),
		COALESCE(descuento_total, 0),
		COALESCE(impuesto_total, 0),
		COALESCE(total, 0),
		COALESCE(metodo_pago, ''),
		COALESCE(estado_pago, 'pendiente'),
		COALESCE(referencia_externa, ''),
		COALESCE(transaction_id, ''),
		COALESCE(items_json, ''),
		COALESCE(pasarela_payload_json, ''),
		COALESCE(pagado_en, ''),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM empresa_venta_publica_ordenes
	WHERE empresa_id = ? AND codigo_orden = ?
	LIMIT 1`, empresaID, codigoOrden).Scan(
		&out.ID,
		&out.EmpresaID,
		&out.CodigoOrden,
		&out.CompradorNombre,
		&out.CompradorEmail,
		&out.CompradorTelefono,
		&out.Moneda,
		&out.Subtotal,
		&out.DescuentoTotal,
		&out.ImpuestoTotal,
		&out.Total,
		&out.MetodoPago,
		&out.EstadoPago,
		&out.ReferenciaExterna,
		&out.TransactionID,
		&out.ItemsJSON,
		&out.PasarelaPayloadJSON,
		&out.PagadoEn,
		&out.FechaCreacion,
		&out.FechaActualizacion,
		&out.UsuarioCreador,
		&out.Estado,
		&out.Observaciones,
	)
	if err != nil {
		return EmpresaVentaPublicaOrder{}, err
	}
	out.Moneda = ventaPublicaNormalizeMoneda(out.Moneda)
	out.Estado = ventaPublicaNormalizeEstado(out.Estado)
	if strings.TrimSpace(out.EstadoPago) == "" {
		out.EstadoPago = "pendiente"
	}
	return out, nil
}

// FindEmpresaVentaPublicaOrderByTransactionOrReference busca una orden por transaction_id o referencia externa.
func FindEmpresaVentaPublicaOrderByTransactionOrReference(dbConn *sql.DB, transactionID, referencia string) (EmpresaVentaPublicaOrder, error) {
	if dbConn == nil {
		return EmpresaVentaPublicaOrder{}, errors.New("db connection is nil")
	}
	transactionID = strings.TrimSpace(transactionID)
	referencia = strings.TrimSpace(referencia)
	if transactionID == "" && referencia == "" {
		return EmpresaVentaPublicaOrder{}, fmt.Errorf("transaction_id o referencia invalida")
	}

	var out EmpresaVentaPublicaOrder
	err := dbConn.QueryRow(`SELECT
		id,
		empresa_id,
		COALESCE(codigo_orden, ''),
		COALESCE(comprador_nombre, ''),
		COALESCE(comprador_email, ''),
		COALESCE(comprador_telefono, ''),
		COALESCE(moneda, 'COP'),
		COALESCE(subtotal, 0),
		COALESCE(descuento_total, 0),
		COALESCE(impuesto_total, 0),
		COALESCE(total, 0),
		COALESCE(metodo_pago, ''),
		COALESCE(estado_pago, 'pendiente'),
		COALESCE(referencia_externa, ''),
		COALESCE(transaction_id, ''),
		COALESCE(items_json, ''),
		COALESCE(pasarela_payload_json, ''),
		COALESCE(pagado_en, ''),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM empresa_venta_publica_ordenes
	WHERE ((? <> '' AND transaction_id = ?) OR (? <> '' AND referencia_externa = ?))
	ORDER BY id DESC
	LIMIT 1`, transactionID, transactionID, referencia, referencia).Scan(
		&out.ID,
		&out.EmpresaID,
		&out.CodigoOrden,
		&out.CompradorNombre,
		&out.CompradorEmail,
		&out.CompradorTelefono,
		&out.Moneda,
		&out.Subtotal,
		&out.DescuentoTotal,
		&out.ImpuestoTotal,
		&out.Total,
		&out.MetodoPago,
		&out.EstadoPago,
		&out.ReferenciaExterna,
		&out.TransactionID,
		&out.ItemsJSON,
		&out.PasarelaPayloadJSON,
		&out.PagadoEn,
		&out.FechaCreacion,
		&out.FechaActualizacion,
		&out.UsuarioCreador,
		&out.Estado,
		&out.Observaciones,
	)
	if err != nil {
		return EmpresaVentaPublicaOrder{}, err
	}
	out.Moneda = ventaPublicaNormalizeMoneda(out.Moneda)
	out.Estado = ventaPublicaNormalizeEstado(out.Estado)
	if strings.TrimSpace(out.EstadoPago) == "" {
		out.EstadoPago = "pendiente"
	}
	return out, nil
}

// ListEmpresaVentaPublicaOrders lista ordenes publicas por empresa.
func ListEmpresaVentaPublicaOrders(dbConn *sql.DB, empresaID int64, filter EmpresaVentaPublicaOrdersFilter) ([]EmpresaVentaPublicaOrder, int64, error) {
	if dbConn == nil {
		return nil, 0, errors.New("db connection is nil")
	}
	if empresaID <= 0 {
		return nil, 0, fmt.Errorf("empresa_id invalido")
	}

	limit, offset := ventaPublicaNormalizeLimitOffset(filter.Limit, filter.Offset)
	where := `WHERE empresa_id = ?`
	args := []interface{}{empresaID}
	if !filter.IncludeInactive {
		where += ` AND COALESCE(estado, 'activo') <> 'inactivo'`
	}
	if status := strings.TrimSpace(strings.ToLower(filter.EstadoPago)); status != "" {
		where += ` AND LOWER(COALESCE(estado_pago, '')) = ?`
		args = append(args, status)
	}
	if q := strings.TrimSpace(filter.Q); q != "" {
		pattern := ventaPublicaLikePattern(q)
		where += ` AND (
			LOWER(COALESCE(codigo_orden, '')) LIKE LOWER(?) ESCAPE '!' OR
			LOWER(COALESCE(comprador_nombre, '')) LIKE LOWER(?) ESCAPE '!' OR
			LOWER(COALESCE(comprador_email, '')) LIKE LOWER(?) ESCAPE '!' OR
			LOWER(COALESCE(referencia_externa, '')) LIKE LOWER(?) ESCAPE '!'
		)`
		args = append(args, pattern, pattern, pattern, pattern)
	}

	var total int64
	if err := dbConn.QueryRow(`SELECT COUNT(1) FROM empresa_venta_publica_ordenes `+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `SELECT
		id,
		empresa_id,
		COALESCE(codigo_orden, ''),
		COALESCE(comprador_nombre, ''),
		COALESCE(comprador_email, ''),
		COALESCE(comprador_telefono, ''),
		COALESCE(moneda, 'COP'),
		COALESCE(subtotal, 0),
		COALESCE(descuento_total, 0),
		COALESCE(impuesto_total, 0),
		COALESCE(total, 0),
		COALESCE(metodo_pago, ''),
		COALESCE(estado_pago, 'pendiente'),
		COALESCE(referencia_externa, ''),
		COALESCE(transaction_id, ''),
		COALESCE(items_json, ''),
		COALESCE(pasarela_payload_json, ''),
		COALESCE(pagado_en, ''),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM empresa_venta_publica_ordenes ` + where + `
	ORDER BY id DESC
	LIMIT ? OFFSET ?`
	rows, err := dbConn.Query(query, append(args, limit, offset)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := make([]EmpresaVentaPublicaOrder, 0)
	for rows.Next() {
		var item EmpresaVentaPublicaOrder
		if err := rows.Scan(
			&item.ID,
			&item.EmpresaID,
			&item.CodigoOrden,
			&item.CompradorNombre,
			&item.CompradorEmail,
			&item.CompradorTelefono,
			&item.Moneda,
			&item.Subtotal,
			&item.DescuentoTotal,
			&item.ImpuestoTotal,
			&item.Total,
			&item.MetodoPago,
			&item.EstadoPago,
			&item.ReferenciaExterna,
			&item.TransactionID,
			&item.ItemsJSON,
			&item.PasarelaPayloadJSON,
			&item.PagadoEn,
			&item.FechaCreacion,
			&item.FechaActualizacion,
			&item.UsuarioCreador,
			&item.Estado,
			&item.Observaciones,
		); err != nil {
			return nil, 0, err
		}
		item.Moneda = ventaPublicaNormalizeMoneda(item.Moneda)
		item.Estado = ventaPublicaNormalizeEstado(item.Estado)
		if strings.TrimSpace(item.EstadoPago) == "" {
			item.EstadoPago = "pendiente"
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

// ParseEmpresaVentaPublicaOrderItems interpreta payload de items en formato JSON.
func ParseEmpresaVentaPublicaOrderItems(raw string) ([]map[string]interface{}, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return []map[string]interface{}{}, nil
	}
	var out []map[string]interface{}
	if err := json.Unmarshal([]byte(value), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// EncodeEmpresaVentaPublicaOrderItems serializa items para persistencia.
func EncodeEmpresaVentaPublicaOrderItems(items []map[string]interface{}) string {
	if len(items) == 0 {
		return "[]"
	}
	raw, err := json.Marshal(items)
	if err != nil {
		return "[]"
	}
	return string(raw)
}

// ResolveVentaPublicaEmpresaIDFromAny resuelve empresa por empresa_id o slug.
func ResolveVentaPublicaEmpresaIDFromAny(dbConn *sql.DB, empresaID int64, slug string) (int64, error) {
	if empresaID > 0 {
		return empresaID, nil
	}
	slug = strings.TrimSpace(slug)
	if slug == "" {
		return 0, fmt.Errorf("empresa_id o empresa_slug son obligatorios")
	}
	resolved, err := ResolveEmpresaIDByVentaPublicaSlug(dbConn, slug)
	if err != nil {
		return 0, err
	}
	return resolved, nil
}

// TryParseOrderCodeFromReference intenta extraer codigo de orden desde referencia externa.
func TryParseOrderCodeFromReference(reference string) string {
	reference = strings.TrimSpace(reference)
	if reference == "" {
		return ""
	}
	parts := strings.Split(reference, "|")
	if len(parts) >= 2 {
		candidate := strings.TrimSpace(parts[len(parts)-1])
		if candidate != "" {
			return candidate
		}
	}
	if strings.Contains(reference, "VP-ORD-") {
		idx := strings.Index(reference, "VP-ORD-")
		if idx >= 0 {
			candidate := strings.TrimSpace(reference[idx:])
			if candidate != "" {
				return candidate
			}
		}
	}
	if strings.HasPrefix(reference, "VP") {
		return reference
	}
	return ""
}

// BuildVentaPublicaOrderReference genera referencia externa para transaccion Wompi.
func BuildVentaPublicaOrderReference(empresaID int64, orderCode string) string {
	orderCode = strings.TrimSpace(orderCode)
	if orderCode == "" {
		orderCode = ventaPublicaGenerateOrderCode(empresaID)
	}
	return "VP|" + strconv.FormatInt(empresaID, 10) + "|" + orderCode
}
