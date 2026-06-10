package db

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	empresaVentaPublicaSchemaMu    sync.Mutex
	empresaVentaPublicaSchemaReady bool
)

// EnsureVentaPublicaSchema crea las tablas necesarias para el módulo de venta pública.
// Debe ser idempotente y segura (usa IF NOT EXISTS cuando es soportado).
func EnsureVentaPublicaSchema(db *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS paginas_publicas (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			slug TEXT NOT NULL,
			titulo TEXT NOT NULL,
			descripcion TEXT,
			video_url TEXT,
			activo INTEGER NOT NULL DEFAULT 1,
			creado_en DATETIME DEFAULT (CURRENT_TIMESTAMP),
			actualizado_en DATETIME DEFAULT (CURRENT_TIMESTAMP)
		);`,

		`CREATE UNIQUE INDEX IF NOT EXISTS idx_paginas_publicas_empresa_slug ON paginas_publicas(empresa_id, slug);`,

		`CREATE TABLE IF NOT EXISTS productos_publicos (
			id BIGSERIAL PRIMARY KEY,
			pagina_id INTEGER NOT NULL,
			nombre TEXT NOT NULL,
			descripcion TEXT,
			precio_cents INTEGER NOT NULL DEFAULT 0,
			moneda TEXT NOT NULL DEFAULT 'COP',
			stock INTEGER,
			sku TEXT,
			youtube_url TEXT,
			activo INTEGER NOT NULL DEFAULT 1,
			creado_en DATETIME DEFAULT (CURRENT_TIMESTAMP),
			actualizado_en DATETIME DEFAULT (CURRENT_TIMESTAMP),
			FOREIGN KEY(pagina_id) REFERENCES paginas_publicas(id) ON DELETE CASCADE
		);`,

		`CREATE INDEX IF NOT EXISTS idx_productos_publicos_pagina_id ON productos_publicos(pagina_id);`,

		`CREATE TABLE IF NOT EXISTS imagenes_productos_publicos (
			id BIGSERIAL PRIMARY KEY,
			producto_id INTEGER NOT NULL,
			url TEXT NOT NULL,
			orden INTEGER NOT NULL DEFAULT 0,
			FOREIGN KEY(producto_id) REFERENCES productos_publicos(id) ON DELETE CASCADE
		);`,

		`CREATE INDEX IF NOT EXISTS idx_imagenes_producto_id ON imagenes_productos_publicos(producto_id);`,

		`CREATE TABLE IF NOT EXISTS empresa_payment_settings (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			provider TEXT NOT NULL,
			config TEXT,
			activo INTEGER NOT NULL DEFAULT 1,
			creado_en DATETIME DEFAULT (CURRENT_TIMESTAMP),
			actualizado_en DATETIME DEFAULT (CURRENT_TIMESTAMP)
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
	ID          int64         `json:"id"`
	PaginaID    int64         `json:"pagina_id"`
	Nombre      string        `json:"nombre"`
	Descripcion string        `json:"descripcion"`
	PrecioCents int64         `json:"precio_cents"`
	Moneda      string        `json:"moneda"`
	Stock       sql.NullInt64 `json:"-"`
	SKU         string        `json:"sku"`
	YoutubeURL  string        `json:"youtube_url"`
	Activo      bool          `json:"activo"`
}

// Use ventaPublicaBoolToInt más abajo; evitar colisiones de nombre en paquete db.

func CreatePaginaPublica(db *sql.DB, empresaID int64, slug, titulo, descripcion, videoURL string, activo bool) (int64, error) {
	if empresaID <= 0 || strings.TrimSpace(slug) == "" || strings.TrimSpace(titulo) == "" {
		return 0, fmt.Errorf("empresa_id, slug y titulo son obligatorios")
	}
	id, err := insertSQLCompat(db, `INSERT INTO paginas_publicas (
		empresa_id, slug, titulo, descripcion, video_url, activo, creado_en, actualizado_en
	) VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		empresaID, strings.TrimSpace(slug), strings.TrimSpace(titulo), strings.TrimSpace(descripcion), strings.TrimSpace(videoURL), ventaPublicaBoolToInt(activo))
	if err != nil {
		return 0, err
	}
	return id, nil
}

func CreateProductoPublico(db *sql.DB, paginaID int64, nombre, descripcion string, precioCents int64, moneda string, stock sql.NullInt64, sku, youtubeURL string, activo bool) (int64, error) {
	if paginaID <= 0 || strings.TrimSpace(nombre) == "" {
		return 0, fmt.Errorf("pagina_id y nombre son obligatorios")
	}
	id, err := insertSQLCompat(db, `INSERT INTO productos_publicos (
		pagina_id, nombre, descripcion, precio_cents, moneda, stock, sku, youtube_url, activo, creado_en, actualizado_en
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		paginaID, strings.TrimSpace(nombre), strings.TrimSpace(descripcion), precioCents, strings.TrimSpace(moneda), nullableInt64Value(stock), strings.TrimSpace(sku), strings.TrimSpace(youtubeURL), ventaPublicaBoolToInt(activo))
	if err != nil {
		return 0, err
	}
	return id, nil
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
	_, err := execSQLCompat(db, `INSERT INTO imagenes_productos_publicos (producto_id, url, orden) VALUES (?, ?, ?)`, productoID, strings.TrimSpace(url), orden)
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
	ventaPublicaWompiModeSandbox     = "sandbox"
	ventaPublicaWompiModeReal        = "production"
	ventaPublicaEpaycoModeSandbox    = "sandbox"
	ventaPublicaEpaycoModeProduction = "production"
)

// EmpresaVentaPublicaConfig define la configuracion de catalogo/pagos publicos por empresa.
type EmpresaVentaPublicaConfig struct {
	ID                              int64  `json:"id"`
	EmpresaID                       int64  `json:"empresa_id"`
	EmpresaSlug                     string `json:"empresa_slug"`
	NombreTienda                    string `json:"nombre_tienda"`
	DescripcionTienda               string `json:"descripcion_tienda,omitempty"`
	LogoURL                         string `json:"logo_url,omitempty"`
	BannerURL                       string `json:"banner_url,omitempty"`
	ColorPrimario                   string `json:"color_primario,omitempty"`
	TemaVisual                      string `json:"tema_visual,omitempty"`
	Moneda                          string `json:"moneda"`
	DominioPublico                  string `json:"dominio_publico,omitempty"`
	MostrarStock                    bool   `json:"mostrar_stock"`
	ContactoFormularioActivo        bool   `json:"contacto_formulario_activo"`
	PedidosRestauranteActivo        bool   `json:"pedidos_restaurante_activo"`
	PedidosRegistroOpcionalCliente  bool   `json:"pedidos_registro_opcional_cliente"`
	PedidosPermitirRecogerEnTienda  bool   `json:"pedidos_permitir_recoger_en_tienda"`
	PedidosPermitirDomicilio        bool   `json:"pedidos_permitir_domicilio"`
	PedidosTrackingDomiciliario     bool   `json:"pedidos_tracking_domiciliario"`
	PedidosDespachoAutomatico       bool   `json:"pedidos_despacho_automatico"`
	PedidosNombreSistema            string `json:"pedidos_nombre_sistema,omitempty"`
	PedidosTiempoPreparacionMinutos int    `json:"pedidos_tiempo_preparacion_minutos,omitempty"`
	WompiActivo                     bool   `json:"wompi_activo"`
	WompiMode                       string `json:"wompi_mode"`
	WompiPublicKey                  string `json:"wompi_public_key,omitempty"`
	WompiPrivateKeyRef              string `json:"wompi_private_key_ref,omitempty"`
	WompiIntegrityRef               string `json:"wompi_integrity_key_ref,omitempty"`
	WompiEventKeyRef                string `json:"wompi_event_key_ref,omitempty"`
	EpaycoActivo                    bool   `json:"epayco_activo"`
	EpaycoMode                      string `json:"epayco_mode"`
	EpaycoPublicKey                 string `json:"epayco_public_key,omitempty"`
	EpaycoPrivateKeyRef             string `json:"epayco_private_key_ref,omitempty"`
	EpaycoCustomerID                string `json:"epayco_customer_id,omitempty"`
	FechaCreacion                   string `json:"fecha_creacion,omitempty"`
	FechaActualizacion              string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador                  string `json:"usuario_creador,omitempty"`
	Estado                          string `json:"estado,omitempty"`
	Observaciones                   string `json:"observaciones,omitempty"`
}

// EmpresaVentaPublicaPagina representa una pagina publica creada por una empresa bajo su dominio/slug.
type EmpresaVentaPublicaPagina struct {
	ID                 int64  `json:"id"`
	EmpresaID          int64  `json:"empresa_id"`
	Slug               string `json:"slug"`
	Nombre             string `json:"nombre"`
	Descripcion        string `json:"descripcion,omitempty"`
	BannerURL          string `json:"banner_url,omitempty"`
	OrdenVisual        int    `json:"orden_visual,omitempty"`
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
	PaginaID           int64   `json:"pagina_id,omitempty"`
	PaginaSlug         string  `json:"pagina_slug,omitempty"`
	PaginaNombre       string  `json:"pagina_nombre,omitempty"`
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
	PaginaID        int64
	PaginaSlug      string
	Q               string
	Sort            string
	Limit           int
	Offset          int
}

// EmpresaVentaPublicaOrder representa una orden creada desde la pagina publica.
type EmpresaVentaPublicaOrder struct {
	ID                       int64   `json:"id"`
	EmpresaID                int64   `json:"empresa_id"`
	CodigoOrden              string  `json:"codigo_orden"`
	TipoOrden                string  `json:"tipo_orden,omitempty"`
	CompradorNombre          string  `json:"comprador_nombre,omitempty"`
	CompradorEmail           string  `json:"comprador_email,omitempty"`
	CompradorTelefono        string  `json:"comprador_telefono,omitempty"`
	Moneda                   string  `json:"moneda"`
	Subtotal                 float64 `json:"subtotal"`
	DescuentoTotal           float64 `json:"descuento_total"`
	ImpuestoTotal            float64 `json:"impuesto_total"`
	Total                    float64 `json:"total"`
	MetodoPago               string  `json:"metodo_pago"`
	EstadoPago               string  `json:"estado_pago"`
	EstadoPedido             string  `json:"estado_pedido,omitempty"`
	CanalEntrega             string  `json:"canal_entrega,omitempty"`
	DireccionEntrega         string  `json:"direccion_entrega,omitempty"`
	NotasEntrega             string  `json:"notas_entrega,omitempty"`
	ClienteComparteUbicacion bool    `json:"cliente_comparte_ubicacion"`
	EntregaLatitud           float64 `json:"entrega_latitud,omitempty"`
	EntregaLongitud          float64 `json:"entrega_longitud,omitempty"`
	TaxiRequestID            int64   `json:"taxi_request_id,omitempty"`
	TrackingToken            string  `json:"tracking_token,omitempty"`
	ReferenciaExterna        string  `json:"referencia_externa,omitempty"`
	TransactionID            string  `json:"transaction_id,omitempty"`
	ItemsJSON                string  `json:"items_json,omitempty"`
	PasarelaPayloadJSON      string  `json:"pasarela_payload_json,omitempty"`
	PagadoEn                 string  `json:"pagado_en,omitempty"`
	FechaCreacion            string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion       string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador           string  `json:"usuario_creador,omitempty"`
	Estado                   string  `json:"estado,omitempty"`
	Observaciones            string  `json:"observaciones,omitempty"`
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

func ventaPublicaNormalizeOrderType(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "restaurante", "pedido_restaurante", "restaurant":
		return "pedido_restaurante"
	default:
		return "catalogo"
	}
}

func ventaPublicaNormalizeDeliveryChannel(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "recoger", "pickup", "retiro":
		return "recoger"
	case "domicilio", "delivery":
		return "domicilio"
	default:
		return "domicilio"
	}
}

func ventaPublicaNormalizeOrderOperationalState(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "nuevo", "recibido":
		return "recibido"
	case "preparando", "cocinando":
		return "preparando"
	case "listo", "listo_para_entrega":
		return "listo_para_entrega"
	case "entregado_al_mensajero", "despachado":
		return "entregado_al_mensajero"
	case "en_camino":
		return "en_camino"
	case "entregado", "completado":
		return "entregado"
	case "cancelado", "cancelada":
		return "cancelado"
	default:
		return "recibido"
	}
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
	case ventaPublicaWompiModeSandbox, "test", "testing", "pruebas":
		return ventaPublicaWompiModeSandbox
	case ventaPublicaWompiModeReal, "real", "prod", "live", "reales":
		return ventaPublicaWompiModeReal
	default:
		return ventaPublicaWompiModeSandbox
	}
}

func ventaPublicaNormalizeSort(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "precio_asc":
		return "precio_asc"
	case "precio_desc":
		return "precio_desc"
	case "nombre_asc":
		return "nombre_asc"
	case "nuevos", "recientes":
		return "nuevos"
	default:
		return "relevancia"
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
	err := queryRowSQLCompat(dbConn, `SELECT COALESCE(nombre, '') FROM empresas WHERE id = ? OR COALESCE(empresa_id, 0) = ? ORDER BY id ASC LIMIT 1`, empresaID, empresaID).Scan(&nombre)
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
	empresaVentaPublicaSchemaMu.Lock()
	defer empresaVentaPublicaSchemaMu.Unlock()

	if empresaVentaPublicaSchemaReady {
		return nil
	}

	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_venta_publica_configuracion (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL UNIQUE,
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
			contacto_formulario_activo INTEGER DEFAULT 1,
			pedidos_restaurante_activo INTEGER DEFAULT 0,
			pedidos_registro_opcional_cliente INTEGER DEFAULT 1,
			pedidos_permitir_recoger_en_tienda INTEGER DEFAULT 1,
			pedidos_permitir_domicilio INTEGER DEFAULT 1,
			pedidos_tracking_domiciliario INTEGER DEFAULT 1,
			pedidos_despacho_automatico INTEGER DEFAULT 1,
			pedidos_nombre_sistema TEXT DEFAULT 'Pedidos restaurante',
			pedidos_tiempo_preparacion_minutos INTEGER DEFAULT 25,
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
			fecha_creacion TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			fecha_actualizacion TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS empresa_venta_publica_paginas (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL,
			slug TEXT NOT NULL,
			nombre TEXT NOT NULL,
			descripcion TEXT,
			banner_url TEXT,
			orden_visual INTEGER DEFAULT 0,
			fecha_creacion TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			fecha_actualizacion TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			UNIQUE(empresa_id, slug)
		);`,
		`CREATE TABLE IF NOT EXISTS empresa_venta_publica_items (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL,
			pagina_id BIGINT DEFAULT 0,
			producto_id BIGINT DEFAULT 0,
			codigo_publico TEXT NOT NULL,
			nombre TEXT NOT NULL,
			descripcion TEXT,
			precio REAL DEFAULT 0,
			moneda TEXT DEFAULT 'COP',
			imagen_url TEXT,
			stock_publicado REAL DEFAULT 0,
			orden_visual INTEGER DEFAULT 0,
			destacado INTEGER DEFAULT 0,
			fecha_creacion TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			fecha_actualizacion TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			UNIQUE(empresa_id, codigo_publico)
		);`,
		`CREATE TABLE IF NOT EXISTS empresa_venta_publica_ordenes (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL,
			codigo_orden TEXT NOT NULL,
			comprador_nombre TEXT,
			comprador_email TEXT,
			comprador_telefono TEXT,
			tipo_orden TEXT DEFAULT 'catalogo',
			moneda TEXT DEFAULT 'COP',
			subtotal REAL DEFAULT 0,
			descuento_total REAL DEFAULT 0,
			impuesto_total REAL DEFAULT 0,
			total REAL DEFAULT 0,
			metodo_pago TEXT DEFAULT 'wompi_nequi',
			estado_pago TEXT DEFAULT 'pendiente',
			estado_pedido TEXT DEFAULT 'recibido',
			canal_entrega TEXT DEFAULT 'domicilio',
			direccion_entrega TEXT,
			notas_entrega TEXT,
			cliente_comparte_ubicacion INTEGER DEFAULT 0,
			entrega_latitud REAL DEFAULT 0,
			entrega_longitud REAL DEFAULT 0,
			taxi_request_id BIGINT DEFAULT 0,
			tracking_token TEXT,
			referencia_externa TEXT,
			transaction_id TEXT,
			items_json TEXT,
			pasarela_payload_json TEXT,
			pagado_en TEXT,
			fecha_creacion TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			fecha_actualizacion TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			UNIQUE(empresa_id, codigo_orden)
		);`,
	}

	for _, stmt := range stmts {
		if _, err := execSQLCompat(dbConn, stmt); err != nil {
			return err
		}
	}

	if err := ensureColumnIfMissing(dbConn, "empresa_venta_publica_configuracion", "empresa_slug", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_venta_publica_configuracion", "nombre_tienda", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_venta_publica_configuracion", "descripcion_tienda", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_venta_publica_configuracion", "logo_url", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_venta_publica_configuracion", "banner_url", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_venta_publica_configuracion", "color_primario", "TEXT DEFAULT '#0f4c81'"); err != nil {
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
	if err := ensureColumnIfMissing(dbConn, "empresa_venta_publica_configuracion", "dominio_publico", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_venta_publica_configuracion", "pedidos_restaurante_activo", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_venta_publica_configuracion", "contacto_formulario_activo", "INTEGER DEFAULT 1"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_venta_publica_configuracion", "pedidos_registro_opcional_cliente", "INTEGER DEFAULT 1"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_venta_publica_configuracion", "pedidos_permitir_recoger_en_tienda", "INTEGER DEFAULT 1"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_venta_publica_configuracion", "pedidos_permitir_domicilio", "INTEGER DEFAULT 1"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_venta_publica_configuracion", "pedidos_tracking_domiciliario", "INTEGER DEFAULT 1"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_venta_publica_configuracion", "pedidos_despacho_automatico", "INTEGER DEFAULT 1"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_venta_publica_configuracion", "pedidos_nombre_sistema", "TEXT DEFAULT 'Pedidos restaurante'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_venta_publica_configuracion", "pedidos_tiempo_preparacion_minutos", "INTEGER DEFAULT 25"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_venta_publica_configuracion", "epayco_private_key_ref", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_venta_publica_configuracion", "epayco_customer_id", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_venta_publica_configuracion", "fecha_creacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_venta_publica_configuracion", "fecha_actualizacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_venta_publica_configuracion", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_venta_publica_configuracion", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_venta_publica_configuracion", "observaciones", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_venta_publica_items", "producto_id", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_venta_publica_items", "pagina_id", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_venta_publica_items", "imagen_url", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_venta_publica_items", "stock_publicado", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_venta_publica_items", "orden_visual", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_venta_publica_items", "destacado", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_venta_publica_items", "fecha_actualizacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_venta_publica_items", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_venta_publica_items", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_venta_publica_items", "observaciones", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_venta_publica_ordenes", "descuento_total", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_venta_publica_ordenes", "tipo_orden", "TEXT DEFAULT 'catalogo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_venta_publica_ordenes", "estado_pedido", "TEXT DEFAULT 'recibido'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_venta_publica_ordenes", "canal_entrega", "TEXT DEFAULT 'domicilio'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_venta_publica_ordenes", "direccion_entrega", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_venta_publica_ordenes", "notas_entrega", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_venta_publica_ordenes", "cliente_comparte_ubicacion", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_venta_publica_ordenes", "entrega_latitud", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_venta_publica_ordenes", "entrega_longitud", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_venta_publica_ordenes", "taxi_request_id", "BIGINT DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_venta_publica_ordenes", "tracking_token", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_venta_publica_ordenes", "impuesto_total", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_venta_publica_ordenes", "referencia_externa", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_venta_publica_ordenes", "transaction_id", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_venta_publica_ordenes", "items_json", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_venta_publica_ordenes", "pasarela_payload_json", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_venta_publica_ordenes", "pagado_en", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_venta_publica_ordenes", "fecha_actualizacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_venta_publica_ordenes", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_venta_publica_ordenes", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_venta_publica_ordenes", "observaciones", "TEXT"); err != nil {
		return err
	}
	if _, err := execSQLCompat(dbConn, `CREATE UNIQUE INDEX IF NOT EXISTS ux_venta_publica_cfg_slug ON empresa_venta_publica_configuracion(empresa_slug)`); err != nil {
		return err
	}
	if _, err := execSQLCompat(dbConn, `CREATE INDEX IF NOT EXISTS ix_venta_publica_cfg_empresa_estado ON empresa_venta_publica_configuracion(empresa_id, estado)`); err != nil {
		return err
	}
	if _, err := execSQLCompat(dbConn, `CREATE INDEX IF NOT EXISTS ix_venta_publica_items_empresa_estado ON empresa_venta_publica_items(empresa_id, estado, orden_visual, id)`); err != nil {
		return err
	}
	if _, err := execSQLCompat(dbConn, `CREATE INDEX IF NOT EXISTS ix_venta_publica_items_empresa_pagina ON empresa_venta_publica_items(empresa_id, pagina_id, estado)`); err != nil {
		return err
	}
	if _, err := execSQLCompat(dbConn, `CREATE INDEX IF NOT EXISTS ix_venta_publica_paginas_empresa_estado ON empresa_venta_publica_paginas(empresa_id, estado, orden_visual, id)`); err != nil {
		return err
	}
	if _, err := execSQLCompat(dbConn, `CREATE INDEX IF NOT EXISTS ix_venta_publica_ordenes_empresa_estado ON empresa_venta_publica_ordenes(empresa_id, estado_pago, fecha_creacion DESC)`); err != nil {
		return err
	}
	if _, err := execSQLCompat(dbConn, `CREATE INDEX IF NOT EXISTS ix_venta_publica_ordenes_tx ON empresa_venta_publica_ordenes(transaction_id)`); err != nil {
		return err
	}
	if _, err := execSQLCompat(dbConn, `CREATE INDEX IF NOT EXISTS ix_venta_publica_ordenes_tracking ON empresa_venta_publica_ordenes(empresa_id, tracking_token)`); err != nil {
		return err
	}
	empresaVentaPublicaSchemaReady = true
	return nil
}

func empresaVentaPublicaSchemaLooksReady(dbConn *sql.DB) (bool, error) {
	requiredTables := []string{
		"empresa_venta_publica_configuracion",
		"empresa_venta_publica_paginas",
		"empresa_venta_publica_items",
		"empresa_venta_publica_ordenes",
	}
	for _, tableName := range requiredTables {
		ok, err := tableExists(dbConn, tableName)
		if err != nil || !ok {
			return false, err
		}
	}

	requiredIndexes := []string{
		"ux_venta_publica_cfg_slug",
		"ix_venta_publica_cfg_empresa_estado",
		"ix_venta_publica_items_empresa_estado",
		"ix_venta_publica_items_empresa_pagina",
		"ix_venta_publica_paginas_empresa_estado",
		"ix_venta_publica_ordenes_empresa_estado",
		"ix_venta_publica_ordenes_tx",
	}
	for _, indexName := range requiredIndexes {
		ok, err := empresaVentaPublicaIndexExists(dbConn, indexName)
		if err != nil || !ok {
			return false, err
		}
	}
	return true, nil
}

func empresaVentaPublicaIndexExists(dbConn *sql.DB, indexName string) (bool, error) {
	var exists bool
	err := queryRowSQLCompat(dbConn, `
		SELECT EXISTS (
			SELECT 1
			FROM pg_indexes
			WHERE schemaname = ANY (current_schemas(false))
			  AND indexname = ?
		)
	`, indexName).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// ListEmpresaVentaPublicaPaginas lista las paginas publicas de una empresa.
func ListEmpresaVentaPublicaPaginas(dbConn *sql.DB, empresaID int64, includeInactive bool) ([]EmpresaVentaPublicaPagina, error) {
	if dbConn == nil {
		return nil, errors.New("db connection is nil")
	}
	if empresaID <= 0 {
		return nil, fmt.Errorf("empresa_id invalido")
	}
	if err := EnsureEmpresaVentaPublicaSchema(dbConn); err != nil {
		return nil, err
	}
	where := `WHERE empresa_id = ?`
	args := []interface{}{empresaID}
	if !includeInactive {
		where += ` AND COALESCE(estado, 'activo') <> 'inactivo'`
	}
	rows, err := querySQLCompat(dbConn, `SELECT
		id, empresa_id, COALESCE(slug, ''), COALESCE(nombre, ''), COALESCE(descripcion, ''),
		COALESCE(banner_url, ''), COALESCE(orden_visual, 0), COALESCE(CAST(fecha_creacion AS TEXT), ''),
		COALESCE(CAST(fecha_actualizacion AS TEXT), ''), COALESCE(usuario_creador, ''), COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM empresa_venta_publica_paginas `+where+`
	ORDER BY COALESCE(orden_visual, 0) ASC, id DESC`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]EmpresaVentaPublicaPagina, 0)
	for rows.Next() {
		var p EmpresaVentaPublicaPagina
		if err := rows.Scan(&p.ID, &p.EmpresaID, &p.Slug, &p.Nombre, &p.Descripcion, &p.BannerURL, &p.OrdenVisual, &p.FechaCreacion, &p.FechaActualizacion, &p.UsuarioCreador, &p.Estado, &p.Observaciones); err != nil {
			return nil, err
		}
		p.Slug = NormalizeEmpresaPublicSlug(p.Slug)
		p.Estado = ventaPublicaNormalizeEstado(p.Estado)
		out = append(out, p)
	}
	return out, rows.Err()
}

// UpsertEmpresaVentaPublicaPagina crea o actualiza una pagina publica por empresa.
func UpsertEmpresaVentaPublicaPagina(dbConn *sql.DB, page EmpresaVentaPublicaPagina) (int64, error) {
	if dbConn == nil {
		return 0, errors.New("db connection is nil")
	}
	if page.EmpresaID <= 0 {
		return 0, fmt.Errorf("empresa_id invalido")
	}
	if err := EnsureEmpresaVentaPublicaSchema(dbConn); err != nil {
		return 0, err
	}
	page.Nombre = strings.TrimSpace(page.Nombre)
	if page.Nombre == "" {
		return 0, fmt.Errorf("nombre es obligatorio")
	}
	page.Slug = NormalizeEmpresaPublicSlug(page.Slug)
	if page.Slug == "" || page.Slug == "empresa" {
		page.Slug = NormalizeEmpresaPublicSlug(page.Nombre)
	}
	if page.Slug == "" {
		return 0, fmt.Errorf("slug invalido")
	}
	page.Estado = ventaPublicaNormalizeEstado(page.Estado)
	var existingID int64
	if page.ID > 0 {
		_ = queryRowSQLCompat(dbConn, `SELECT id FROM empresa_venta_publica_paginas WHERE empresa_id = ? AND id = ? LIMIT 1`, page.EmpresaID, page.ID).Scan(&existingID)
	}
	if existingID <= 0 {
		_ = queryRowSQLCompat(dbConn, `SELECT id FROM empresa_venta_publica_paginas WHERE empresa_id = ? AND slug = ? LIMIT 1`, page.EmpresaID, page.Slug).Scan(&existingID)
	}
	if existingID > 0 {
		_, err := execSQLCompat(dbConn, `UPDATE empresa_venta_publica_paginas
			SET slug = ?, nombre = ?, descripcion = ?, banner_url = ?, orden_visual = ?,
				usuario_creador = ?, estado = ?, observaciones = ?, fecha_actualizacion = CURRENT_TIMESTAMP
			WHERE empresa_id = ? AND id = ?`,
			page.Slug, page.Nombre, strings.TrimSpace(page.Descripcion), strings.TrimSpace(page.BannerURL), page.OrdenVisual,
			strings.TrimSpace(page.UsuarioCreador), page.Estado, strings.TrimSpace(page.Observaciones), page.EmpresaID, existingID)
		return existingID, err
	}
	return insertSQLCompat(dbConn, `INSERT INTO empresa_venta_publica_paginas (
		empresa_id, slug, nombre, descripcion, banner_url, orden_visual, usuario_creador, estado, observaciones
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		page.EmpresaID, page.Slug, page.Nombre, strings.TrimSpace(page.Descripcion), strings.TrimSpace(page.BannerURL),
		page.OrdenVisual, strings.TrimSpace(page.UsuarioCreador), page.Estado, strings.TrimSpace(page.Observaciones))
}

// GetEmpresaVentaPublicaPaginaByID obtiene una pagina por id/empresa.
func GetEmpresaVentaPublicaPaginaByID(dbConn *sql.DB, empresaID, pageID int64) (EmpresaVentaPublicaPagina, error) {
	if dbConn == nil {
		return EmpresaVentaPublicaPagina{}, errors.New("db connection is nil")
	}
	var p EmpresaVentaPublicaPagina
	err := queryRowSQLCompat(dbConn, `SELECT
		id, empresa_id, COALESCE(slug, ''), COALESCE(nombre, ''), COALESCE(descripcion, ''),
		COALESCE(banner_url, ''), COALESCE(orden_visual, 0), COALESCE(CAST(fecha_creacion AS TEXT), ''),
		COALESCE(CAST(fecha_actualizacion AS TEXT), ''), COALESCE(usuario_creador, ''), COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM empresa_venta_publica_paginas WHERE empresa_id = ? AND id = ? LIMIT 1`, empresaID, pageID).Scan(
		&p.ID, &p.EmpresaID, &p.Slug, &p.Nombre, &p.Descripcion, &p.BannerURL, &p.OrdenVisual, &p.FechaCreacion, &p.FechaActualizacion, &p.UsuarioCreador, &p.Estado, &p.Observaciones,
	)
	p.Slug = NormalizeEmpresaPublicSlug(p.Slug)
	p.Estado = ventaPublicaNormalizeEstado(p.Estado)
	return p, err
}

// GetEmpresaVentaPublicaPaginaBySlug obtiene una pagina activa por slug/empresa.
func GetEmpresaVentaPublicaPaginaBySlug(dbConn *sql.DB, empresaID int64, slug string) (EmpresaVentaPublicaPagina, error) {
	if dbConn == nil {
		return EmpresaVentaPublicaPagina{}, errors.New("db connection is nil")
	}
	slug = NormalizeEmpresaPublicSlug(slug)
	var p EmpresaVentaPublicaPagina
	err := queryRowSQLCompat(dbConn, `SELECT
		id, empresa_id, COALESCE(slug, ''), COALESCE(nombre, ''), COALESCE(descripcion, ''),
		COALESCE(banner_url, ''), COALESCE(orden_visual, 0), COALESCE(CAST(fecha_creacion AS TEXT), ''),
		COALESCE(CAST(fecha_actualizacion AS TEXT), ''), COALESCE(usuario_creador, ''), COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM empresa_venta_publica_paginas
	WHERE empresa_id = ? AND slug = ? AND COALESCE(estado, 'activo') <> 'inactivo'
	LIMIT 1`, empresaID, slug).Scan(
		&p.ID, &p.EmpresaID, &p.Slug, &p.Nombre, &p.Descripcion, &p.BannerURL, &p.OrdenVisual, &p.FechaCreacion, &p.FechaActualizacion, &p.UsuarioCreador, &p.Estado, &p.Observaciones,
	)
	p.Slug = NormalizeEmpresaPublicSlug(p.Slug)
	p.Estado = ventaPublicaNormalizeEstado(p.Estado)
	return p, err
}

// SetEmpresaVentaPublicaPaginaEstadoByID activa/desactiva una pagina publica.
func SetEmpresaVentaPublicaPaginaEstadoByID(dbConn *sql.DB, empresaID, pageID int64, estado string) error {
	if dbConn == nil {
		return errors.New("db connection is nil")
	}
	res, err := execSQLCompat(dbConn, `UPDATE empresa_venta_publica_paginas SET estado = ?, fecha_actualizacion = CURRENT_TIMESTAMP WHERE empresa_id = ? AND id = ?`, ventaPublicaNormalizeEstado(estado), empresaID, pageID)
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

// GetEmpresaVentaPublicaConfig obtiene la configuracion de venta publica por empresa.
func GetEmpresaVentaPublicaConfig(dbConn *sql.DB, empresaID int64) (EmpresaVentaPublicaConfig, error) {
	if dbConn == nil {
		return EmpresaVentaPublicaConfig{}, errors.New("db connection is nil")
	}
	if empresaID <= 0 {
		return EmpresaVentaPublicaConfig{}, fmt.Errorf("empresa_id invalido")
	}
	if err := EnsureEmpresaVentaPublicaSchema(dbConn); err != nil {
		return EmpresaVentaPublicaConfig{}, err
	}

	var out EmpresaVentaPublicaConfig
	var mostrarStock sql.NullInt64
	var contactoFormularioActivo sql.NullInt64
	var pedidosRestauranteActivo sql.NullInt64
	var pedidosRegistroOpcional sql.NullInt64
	var pedidosPermitirRecoger sql.NullInt64
	var pedidosPermitirDomicilio sql.NullInt64
	var pedidosTracking sql.NullInt64
	var pedidosDespacho sql.NullInt64
	var wompiActivo sql.NullInt64
	var epaycoActivo sql.NullInt64
	err := queryRowSQLCompat(dbConn, `SELECT
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
		COALESCE(contacto_formulario_activo, 1),
		COALESCE(pedidos_restaurante_activo, 0),
		COALESCE(pedidos_registro_opcional_cliente, 1),
		COALESCE(pedidos_permitir_recoger_en_tienda, 1),
		COALESCE(pedidos_permitir_domicilio, 1),
		COALESCE(pedidos_tracking_domiciliario, 1),
		COALESCE(pedidos_despacho_automatico, 1),
		COALESCE(pedidos_nombre_sistema, 'Pedidos restaurante'),
		COALESCE(pedidos_tiempo_preparacion_minutos, 25),
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
		COALESCE(CAST(fecha_creacion AS TEXT), ''),
		COALESCE(CAST(fecha_actualizacion AS TEXT), ''),
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
		&contactoFormularioActivo,
		&pedidosRestauranteActivo,
		&pedidosRegistroOpcional,
		&pedidosPermitirRecoger,
		&pedidosPermitirDomicilio,
		&pedidosTracking,
		&pedidosDespacho,
		&out.PedidosNombreSistema,
		&out.PedidosTiempoPreparacionMinutos,
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
			EmpresaID:                       empresaID,
			EmpresaSlug:                     NormalizeEmpresaPublicSlug(nombre),
			NombreTienda:                    nombre,
			ColorPrimario:                   "#0f4c81",
			TemaVisual:                      "default",
			Moneda:                          "COP",
			MostrarStock:                    true,
			ContactoFormularioActivo:        true,
			PedidosRegistroOpcionalCliente:  true,
			PedidosPermitirRecogerEnTienda:  true,
			PedidosPermitirDomicilio:        true,
			PedidosTrackingDomiciliario:     true,
			PedidosDespachoAutomatico:       true,
			PedidosNombreSistema:            "Pedidos restaurante",
			PedidosTiempoPreparacionMinutos: 25,
			WompiActivo:                     false,
			WompiMode:                       ventaPublicaWompiModeSandbox,
			EpaycoActivo:                    false,
			EpaycoMode:                      ventaPublicaEpaycoModeSandbox,
			Estado:                          "activo",
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
	out.ContactoFormularioActivo = !contactoFormularioActivo.Valid || contactoFormularioActivo.Int64 > 0
	out.PedidosRestauranteActivo = pedidosRestauranteActivo.Valid && pedidosRestauranteActivo.Int64 > 0
	out.PedidosRegistroOpcionalCliente = !pedidosRegistroOpcional.Valid || pedidosRegistroOpcional.Int64 > 0
	out.PedidosPermitirRecogerEnTienda = !pedidosPermitirRecoger.Valid || pedidosPermitirRecoger.Int64 > 0
	out.PedidosPermitirDomicilio = !pedidosPermitirDomicilio.Valid || pedidosPermitirDomicilio.Int64 > 0
	out.PedidosTrackingDomiciliario = !pedidosTracking.Valid || pedidosTracking.Int64 > 0
	out.PedidosDespachoAutomatico = !pedidosDespacho.Valid || pedidosDespacho.Int64 > 0
	if out.PedidosTiempoPreparacionMinutos <= 0 {
		out.PedidosTiempoPreparacionMinutos = 25
	}
	if strings.TrimSpace(out.PedidosNombreSistema) == "" {
		out.PedidosNombreSistema = "Pedidos restaurante"
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
	if err := EnsureEmpresaVentaPublicaSchema(dbConn); err != nil {
		return 0, err
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
	if strings.TrimSpace(cfg.PedidosNombreSistema) == "" {
		cfg.PedidosNombreSistema = "Pedidos restaurante"
	}
	if cfg.PedidosTiempoPreparacionMinutos <= 0 {
		cfg.PedidosTiempoPreparacionMinutos = 25
	}

	var existingID int64
	err := queryRowSQLCompat(dbConn, `SELECT id FROM empresa_venta_publica_configuracion WHERE empresa_id = ? LIMIT 1`, cfg.EmpresaID).Scan(&existingID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}

	if existingID > 0 {
		_, err = execSQLCompat(dbConn, `UPDATE empresa_venta_publica_configuracion
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
				contacto_formulario_activo = ?,
				pedidos_restaurante_activo = ?,
				pedidos_registro_opcional_cliente = ?,
				pedidos_permitir_recoger_en_tienda = ?,
				pedidos_permitir_domicilio = ?,
				pedidos_tracking_domiciliario = ?,
				pedidos_despacho_automatico = ?,
				pedidos_nombre_sistema = ?,
				pedidos_tiempo_preparacion_minutos = ?,
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
				fecha_actualizacion = CURRENT_TIMESTAMP
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
			ventaPublicaBoolToInt(cfg.ContactoFormularioActivo),
			ventaPublicaBoolToInt(cfg.PedidosRestauranteActivo),
			ventaPublicaBoolToInt(cfg.PedidosRegistroOpcionalCliente),
			ventaPublicaBoolToInt(cfg.PedidosPermitirRecogerEnTienda),
			ventaPublicaBoolToInt(cfg.PedidosPermitirDomicilio),
			ventaPublicaBoolToInt(cfg.PedidosTrackingDomiciliario),
			ventaPublicaBoolToInt(cfg.PedidosDespachoAutomatico),
			strings.TrimSpace(cfg.PedidosNombreSistema),
			cfg.PedidosTiempoPreparacionMinutos,
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

	id, err := insertSQLCompat(dbConn, `INSERT INTO empresa_venta_publica_configuracion (
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
		contacto_formulario_activo,
		pedidos_restaurante_activo,
		pedidos_registro_opcional_cliente,
		pedidos_permitir_recoger_en_tienda,
		pedidos_permitir_domicilio,
		pedidos_tracking_domiciliario,
		pedidos_despacho_automatico,
		pedidos_nombre_sistema,
		pedidos_tiempo_preparacion_minutos,
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
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
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
		ventaPublicaBoolToInt(cfg.ContactoFormularioActivo),
		ventaPublicaBoolToInt(cfg.PedidosRestauranteActivo),
		ventaPublicaBoolToInt(cfg.PedidosRegistroOpcionalCliente),
		ventaPublicaBoolToInt(cfg.PedidosPermitirRecogerEnTienda),
		ventaPublicaBoolToInt(cfg.PedidosPermitirDomicilio),
		ventaPublicaBoolToInt(cfg.PedidosTrackingDomiciliario),
		ventaPublicaBoolToInt(cfg.PedidosDespachoAutomatico),
		strings.TrimSpace(cfg.PedidosNombreSistema),
		cfg.PedidosTiempoPreparacionMinutos,
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
	err := queryRowSQLCompat(dbConn, `SELECT empresa_id FROM empresa_venta_publica_configuracion WHERE empresa_slug = ? AND COALESCE(estado, 'activo') <> 'inactivo' LIMIT 1`, slug).Scan(&empresaID)
	if err == nil && empresaID > 0 {
		return empresaID, nil
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}

	rows, err := querySQLCompat(dbConn, `SELECT id, COALESCE(empresa_id, id), COALESCE(nombre, '') FROM empresas WHERE COALESCE(estado, 'activo') <> 'inactivo' ORDER BY id ASC`)
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
	err := queryRowSQLCompat(dbConn, `SELECT
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
	if err := EnsureEmpresaVentaPublicaSchema(dbConn); err != nil {
		return 0, err
	}
	if item.Precio < 0 {
		return 0, fmt.Errorf("precio invalido")
	}
	if item.StockPublicado < 0 {
		return 0, fmt.Errorf("stock_publicado invalido")
	}
	if item.PaginaID > 0 {
		if _, err := GetEmpresaVentaPublicaPaginaByID(dbConn, item.EmpresaID, item.PaginaID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return 0, fmt.Errorf("pagina_id no encontrado para la empresa")
			}
			return 0, err
		}
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
	if item.PaginaID > 0 {
		item.CodigoPublico = fmt.Sprintf("%s-p%d", strings.TrimSuffix(item.CodigoPublico, fmt.Sprintf("-p%d", item.PaginaID)), item.PaginaID)
	}
	item.Moneda = ventaPublicaNormalizeMoneda(item.Moneda)
	item.Estado = ventaPublicaNormalizeEstado(item.Estado)

	id, err := insertSQLCompat(dbConn, `INSERT INTO empresa_venta_publica_items (
		empresa_id,
		pagina_id,
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
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		item.EmpresaID,
		item.PaginaID,
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
	if item.PaginaID > 0 {
		if _, err := GetEmpresaVentaPublicaPaginaByID(dbConn, item.EmpresaID, item.PaginaID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("pagina_id no encontrado para la empresa")
			}
			return err
		}
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
	if item.PaginaID > 0 {
		item.CodigoPublico = fmt.Sprintf("%s-p%d", strings.TrimSuffix(item.CodigoPublico, fmt.Sprintf("-p%d", item.PaginaID)), item.PaginaID)
	}
	item.Moneda = ventaPublicaNormalizeMoneda(item.Moneda)
	item.Estado = ventaPublicaNormalizeEstado(item.Estado)

	res, err := execSQLCompat(dbConn, `UPDATE empresa_venta_publica_items
		SET pagina_id = ?,
			producto_id = ?,
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
			fecha_actualizacion = CURRENT_TIMESTAMP
		WHERE empresa_id = ? AND id = ?`,
		item.PaginaID,
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
	res, err := execSQLCompat(dbConn, `UPDATE empresa_venta_publica_items SET estado = ?, fecha_actualizacion = CURRENT_TIMESTAMP WHERE empresa_id = ? AND id = ?`, estado, empresaID, itemID)
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
	err := queryRowSQLCompat(dbConn, `SELECT
		i.id,
		i.empresa_id,
		COALESCE(i.pagina_id, 0),
		COALESCE(p.slug, ''),
		COALESCE(p.nombre, ''),
		COALESCE(i.producto_id, 0),
		COALESCE(i.codigo_publico, ''),
		COALESCE(i.nombre, ''),
		COALESCE(i.descripcion, ''),
		COALESCE(i.precio, 0),
		COALESCE(i.moneda, 'COP'),
		COALESCE(i.imagen_url, ''),
		COALESCE(i.stock_publicado, 0),
		COALESCE(i.orden_visual, 0),
		COALESCE(i.destacado, 0),
		COALESCE(CAST(i.fecha_creacion AS TEXT), ''),
		COALESCE(CAST(i.fecha_actualizacion AS TEXT), ''),
		COALESCE(i.usuario_creador, ''),
		COALESCE(i.estado, 'activo'),
		COALESCE(i.observaciones, '')
	FROM empresa_venta_publica_items i
	LEFT JOIN empresa_venta_publica_paginas p ON p.empresa_id = i.empresa_id AND p.id = COALESCE(i.pagina_id, 0)
	WHERE i.empresa_id = ? AND i.id = ?
	LIMIT 1`, empresaID, itemID).Scan(
		&out.ID,
		&out.EmpresaID,
		&out.PaginaID,
		&out.PaginaSlug,
		&out.PaginaNombre,
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
	where := `WHERE i.empresa_id = ?`
	args := []interface{}{empresaID}
	if !filter.IncludeInactive {
		where += ` AND COALESCE(i.estado, 'activo') <> 'inactivo'`
	}
	if filter.PaginaID > 0 {
		where += ` AND COALESCE(i.pagina_id, 0) = ?`
		args = append(args, filter.PaginaID)
	} else if pageSlug := NormalizeEmpresaPublicSlug(filter.PaginaSlug); pageSlug != "" {
		where += ` AND p.slug = ?`
		args = append(args, pageSlug)
	}
	if q := strings.TrimSpace(filter.Q); q != "" {
		pattern := ventaPublicaLikePattern(q)
		where += ` AND (
			LOWER(COALESCE(i.codigo_publico, '')) LIKE LOWER(?) ESCAPE '!' OR
			LOWER(COALESCE(i.nombre, '')) LIKE LOWER(?) ESCAPE '!' OR
			LOWER(COALESCE(i.descripcion, '')) LIKE LOWER(?) ESCAPE '!'
		)`
		args = append(args, pattern, pattern, pattern)
	}

	countQuery := `SELECT COUNT(1)
	FROM empresa_venta_publica_items i
	LEFT JOIN empresa_venta_publica_paginas p ON p.empresa_id = i.empresa_id AND p.id = COALESCE(i.pagina_id, 0) ` + where
	var total int64
	if err := dbConn.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	orderBy := `ORDER BY COALESCE(p.orden_visual, 0) ASC, COALESCE(i.orden_visual, 0) ASC, i.id DESC`
	switch ventaPublicaNormalizeSort(filter.Sort) {
	case "precio_asc":
		orderBy = `ORDER BY COALESCE(i.precio, 0) ASC, COALESCE(i.orden_visual, 0) ASC, i.id DESC`
	case "precio_desc":
		orderBy = `ORDER BY COALESCE(i.precio, 0) DESC, COALESCE(i.orden_visual, 0) ASC, i.id DESC`
	case "nombre_asc":
		orderBy = `ORDER BY LOWER(COALESCE(i.nombre, '')) ASC, COALESCE(i.orden_visual, 0) ASC, i.id DESC`
	case "nuevos":
		orderBy = `ORDER BY i.id DESC`
	}

	query := `SELECT
		i.id,
		i.empresa_id,
		COALESCE(i.pagina_id, 0),
		COALESCE(p.slug, ''),
		COALESCE(p.nombre, ''),
		COALESCE(i.producto_id, 0),
		COALESCE(i.codigo_publico, ''),
		COALESCE(i.nombre, ''),
		COALESCE(i.descripcion, ''),
		COALESCE(i.precio, 0),
		COALESCE(i.moneda, 'COP'),
		COALESCE(i.imagen_url, ''),
		COALESCE(i.stock_publicado, 0),
		COALESCE(i.orden_visual, 0),
		COALESCE(i.destacado, 0),
		COALESCE(CAST(i.fecha_creacion AS TEXT), ''),
		COALESCE(CAST(i.fecha_actualizacion AS TEXT), ''),
		COALESCE(i.usuario_creador, ''),
		COALESCE(i.estado, 'activo'),
		COALESCE(i.observaciones, '')
	FROM empresa_venta_publica_items i
	LEFT JOIN empresa_venta_publica_paginas p ON p.empresa_id = i.empresa_id AND p.id = COALESCE(i.pagina_id, 0) ` + where + `
	` + orderBy + `
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
			&item.PaginaID,
			&item.PaginaSlug,
			&item.PaginaNombre,
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
	if err := EnsureEmpresaVentaPublicaSchema(dbConn); err != nil {
		return 0, err
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
	order.TipoOrden = ventaPublicaNormalizeOrderType(order.TipoOrden)
	order.EstadoPedido = ventaPublicaNormalizeOrderOperationalState(order.EstadoPedido)
	order.CanalEntrega = ventaPublicaNormalizeDeliveryChannel(order.CanalEntrega)
	if strings.TrimSpace(order.MetodoPago) == "" {
		order.MetodoPago = "wompi_nequi"
	}
	order.Moneda = ventaPublicaNormalizeMoneda(order.Moneda)
	order.Estado = ventaPublicaNormalizeEstado(order.Estado)
	if strings.TrimSpace(order.Estado) == "" {
		order.Estado = "activo"
	}

	id, err := insertSQLCompat(dbConn, `INSERT INTO empresa_venta_publica_ordenes (
		empresa_id,
		codigo_orden,
		comprador_nombre,
		comprador_email,
		comprador_telefono,
		tipo_orden,
		moneda,
		subtotal,
		descuento_total,
		impuesto_total,
		total,
		metodo_pago,
		estado_pago,
		estado_pedido,
		canal_entrega,
		direccion_entrega,
		notas_entrega,
		cliente_comparte_ubicacion,
		entrega_latitud,
		entrega_longitud,
		taxi_request_id,
		tracking_token,
		referencia_externa,
		transaction_id,
		items_json,
		pasarela_payload_json,
		pagado_en,
		usuario_creador,
		estado,
		observaciones
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		order.EmpresaID,
		order.CodigoOrden,
		strings.TrimSpace(order.CompradorNombre),
		strings.TrimSpace(order.CompradorEmail),
		strings.TrimSpace(order.CompradorTelefono),
		order.TipoOrden,
		order.Moneda,
		order.Subtotal,
		order.DescuentoTotal,
		order.ImpuestoTotal,
		order.Total,
		strings.TrimSpace(order.MetodoPago),
		strings.TrimSpace(order.EstadoPago),
		order.EstadoPedido,
		order.CanalEntrega,
		strings.TrimSpace(order.DireccionEntrega),
		strings.TrimSpace(order.NotasEntrega),
		ventaPublicaBoolToInt(order.ClienteComparteUbicacion),
		order.EntregaLatitud,
		order.EntregaLongitud,
		order.TaxiRequestID,
		strings.TrimSpace(order.TrackingToken),
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
	res, err := execSQLCompat(dbConn, `UPDATE empresa_venta_publica_ordenes
		SET estado_pago = ?,
			transaction_id = ?,
			referencia_externa = ?,
			pasarela_payload_json = CASE WHEN ? = '' THEN pasarela_payload_json ELSE ? END,
			pagado_en = CASE WHEN ? = '' THEN pagado_en ELSE ? END,
			observaciones = CASE WHEN ? = '' THEN observaciones ELSE ? END,
			fecha_actualizacion = CURRENT_TIMESTAMP
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
	if err := EnsureEmpresaVentaPublicaSchema(dbConn); err != nil {
		return EmpresaVentaPublicaOrder{}, err
	}
	codigoOrden = strings.TrimSpace(codigoOrden)
	if codigoOrden == "" {
		return EmpresaVentaPublicaOrder{}, fmt.Errorf("codigo_orden invalido")
	}

	var out EmpresaVentaPublicaOrder
	var comparteUbicacion int64
	err := queryRowSQLCompat(dbConn, `SELECT
		id,
		empresa_id,
		COALESCE(codigo_orden, ''),
		COALESCE(comprador_nombre, ''),
		COALESCE(comprador_email, ''),
		COALESCE(comprador_telefono, ''),
		COALESCE(tipo_orden, 'catalogo'),
		COALESCE(moneda, 'COP'),
		COALESCE(subtotal, 0),
		COALESCE(descuento_total, 0),
		COALESCE(impuesto_total, 0),
		COALESCE(total, 0),
		COALESCE(metodo_pago, ''),
		COALESCE(estado_pago, 'pendiente'),
		COALESCE(estado_pedido, 'recibido'),
		COALESCE(canal_entrega, 'domicilio'),
		COALESCE(direccion_entrega, ''),
		COALESCE(notas_entrega, ''),
		COALESCE(cliente_comparte_ubicacion, 0),
		COALESCE(entrega_latitud, 0),
		COALESCE(entrega_longitud, 0),
		COALESCE(taxi_request_id, 0),
		COALESCE(tracking_token, ''),
		COALESCE(referencia_externa, ''),
		COALESCE(transaction_id, ''),
		COALESCE(items_json, ''),
		COALESCE(pasarela_payload_json, ''),
		COALESCE(pagado_en, ''),
		COALESCE(CAST(fecha_creacion AS TEXT), ''),
		COALESCE(CAST(fecha_actualizacion AS TEXT), ''),
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
		&out.TipoOrden,
		&out.Moneda,
		&out.Subtotal,
		&out.DescuentoTotal,
		&out.ImpuestoTotal,
		&out.Total,
		&out.MetodoPago,
		&out.EstadoPago,
		&out.EstadoPedido,
		&out.CanalEntrega,
		&out.DireccionEntrega,
		&out.NotasEntrega,
		&comparteUbicacion,
		&out.EntregaLatitud,
		&out.EntregaLongitud,
		&out.TaxiRequestID,
		&out.TrackingToken,
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
	out.TipoOrden = ventaPublicaNormalizeOrderType(out.TipoOrden)
	out.Estado = ventaPublicaNormalizeEstado(out.Estado)
	out.EstadoPedido = ventaPublicaNormalizeOrderOperationalState(out.EstadoPedido)
	out.CanalEntrega = ventaPublicaNormalizeDeliveryChannel(out.CanalEntrega)
	out.ClienteComparteUbicacion = comparteUbicacion > 0
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
	if err := EnsureEmpresaVentaPublicaSchema(dbConn); err != nil {
		return EmpresaVentaPublicaOrder{}, err
	}
	transactionID = strings.TrimSpace(transactionID)
	referencia = strings.TrimSpace(referencia)
	if transactionID == "" && referencia == "" {
		return EmpresaVentaPublicaOrder{}, fmt.Errorf("transaction_id o referencia invalida")
	}

	var out EmpresaVentaPublicaOrder
	var comparteUbicacion int64
	err := queryRowSQLCompat(dbConn, `SELECT
		id,
		empresa_id,
		COALESCE(codigo_orden, ''),
		COALESCE(comprador_nombre, ''),
		COALESCE(comprador_email, ''),
		COALESCE(comprador_telefono, ''),
		COALESCE(tipo_orden, 'catalogo'),
		COALESCE(moneda, 'COP'),
		COALESCE(subtotal, 0),
		COALESCE(descuento_total, 0),
		COALESCE(impuesto_total, 0),
		COALESCE(total, 0),
		COALESCE(metodo_pago, ''),
		COALESCE(estado_pago, 'pendiente'),
		COALESCE(estado_pedido, 'recibido'),
		COALESCE(canal_entrega, 'domicilio'),
		COALESCE(direccion_entrega, ''),
		COALESCE(notas_entrega, ''),
		COALESCE(cliente_comparte_ubicacion, 0),
		COALESCE(entrega_latitud, 0),
		COALESCE(entrega_longitud, 0),
		COALESCE(taxi_request_id, 0),
		COALESCE(tracking_token, ''),
		COALESCE(referencia_externa, ''),
		COALESCE(transaction_id, ''),
		COALESCE(items_json, ''),
		COALESCE(pasarela_payload_json, ''),
		COALESCE(pagado_en, ''),
		COALESCE(CAST(fecha_creacion AS TEXT), ''),
		COALESCE(CAST(fecha_actualizacion AS TEXT), ''),
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
		&out.TipoOrden,
		&out.Moneda,
		&out.Subtotal,
		&out.DescuentoTotal,
		&out.ImpuestoTotal,
		&out.Total,
		&out.MetodoPago,
		&out.EstadoPago,
		&out.EstadoPedido,
		&out.CanalEntrega,
		&out.DireccionEntrega,
		&out.NotasEntrega,
		&comparteUbicacion,
		&out.EntregaLatitud,
		&out.EntregaLongitud,
		&out.TaxiRequestID,
		&out.TrackingToken,
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
	out.TipoOrden = ventaPublicaNormalizeOrderType(out.TipoOrden)
	out.Estado = ventaPublicaNormalizeEstado(out.Estado)
	out.EstadoPedido = ventaPublicaNormalizeOrderOperationalState(out.EstadoPedido)
	out.CanalEntrega = ventaPublicaNormalizeDeliveryChannel(out.CanalEntrega)
	out.ClienteComparteUbicacion = comparteUbicacion > 0
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
	if err := EnsureEmpresaVentaPublicaSchema(dbConn); err != nil {
		return nil, 0, err
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
	if err := queryRowSQLCompat(dbConn, `SELECT COUNT(1) FROM empresa_venta_publica_ordenes `+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `SELECT
		id,
		empresa_id,
		COALESCE(codigo_orden, ''),
		COALESCE(comprador_nombre, ''),
		COALESCE(comprador_email, ''),
		COALESCE(comprador_telefono, ''),
		COALESCE(tipo_orden, 'catalogo'),
		COALESCE(moneda, 'COP'),
		COALESCE(subtotal, 0),
		COALESCE(descuento_total, 0),
		COALESCE(impuesto_total, 0),
		COALESCE(total, 0),
		COALESCE(metodo_pago, ''),
		COALESCE(estado_pago, 'pendiente'),
		COALESCE(estado_pedido, 'recibido'),
		COALESCE(canal_entrega, 'domicilio'),
		COALESCE(direccion_entrega, ''),
		COALESCE(notas_entrega, ''),
		COALESCE(cliente_comparte_ubicacion, 0),
		COALESCE(entrega_latitud, 0),
		COALESCE(entrega_longitud, 0),
		COALESCE(taxi_request_id, 0),
		COALESCE(tracking_token, ''),
		COALESCE(referencia_externa, ''),
		COALESCE(transaction_id, ''),
		COALESCE(items_json, ''),
		COALESCE(pasarela_payload_json, ''),
		COALESCE(pagado_en, ''),
		COALESCE(CAST(fecha_creacion AS TEXT), ''),
		COALESCE(CAST(fecha_actualizacion AS TEXT), ''),
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
		var comparteUbicacion int64
		if err := rows.Scan(
			&item.ID,
			&item.EmpresaID,
			&item.CodigoOrden,
			&item.CompradorNombre,
			&item.CompradorEmail,
			&item.CompradorTelefono,
			&item.TipoOrden,
			&item.Moneda,
			&item.Subtotal,
			&item.DescuentoTotal,
			&item.ImpuestoTotal,
			&item.Total,
			&item.MetodoPago,
			&item.EstadoPago,
			&item.EstadoPedido,
			&item.CanalEntrega,
			&item.DireccionEntrega,
			&item.NotasEntrega,
			&comparteUbicacion,
			&item.EntregaLatitud,
			&item.EntregaLongitud,
			&item.TaxiRequestID,
			&item.TrackingToken,
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
		item.TipoOrden = ventaPublicaNormalizeOrderType(item.TipoOrden)
		item.Estado = ventaPublicaNormalizeEstado(item.Estado)
		item.EstadoPedido = ventaPublicaNormalizeOrderOperationalState(item.EstadoPedido)
		item.CanalEntrega = ventaPublicaNormalizeDeliveryChannel(item.CanalEntrega)
		item.ClienteComparteUbicacion = comparteUbicacion > 0
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

func GetEmpresaVentaPublicaOrderByTrackingToken(dbConn *sql.DB, empresaID int64, trackingToken string) (EmpresaVentaPublicaOrder, error) {
	if dbConn == nil {
		return EmpresaVentaPublicaOrder{}, errors.New("db connection is nil")
	}
	trackingToken = strings.TrimSpace(trackingToken)
	if empresaID <= 0 || trackingToken == "" {
		return EmpresaVentaPublicaOrder{}, fmt.Errorf("tracking_token invalido")
	}
	var orderCode string
	if err := queryRowSQLCompat(dbConn, `SELECT COALESCE(codigo_orden,'') FROM empresa_venta_publica_ordenes WHERE empresa_id = ? AND tracking_token = ? LIMIT 1`, empresaID, trackingToken).Scan(&orderCode); err != nil {
		return EmpresaVentaPublicaOrder{}, err
	}
	return GetEmpresaVentaPublicaOrderByCodigo(dbConn, empresaID, orderCode)
}

func UpdateEmpresaVentaPublicaOrderOperationalState(dbConn *sql.DB, empresaID int64, codigoOrden, estadoPedido, observaciones string, taxiRequestID int64) error {
	if dbConn == nil {
		return errors.New("db connection is nil")
	}
	if empresaID <= 0 || strings.TrimSpace(codigoOrden) == "" {
		return fmt.Errorf("empresa_id o codigo_orden invalido")
	}
	estadoPedido = ventaPublicaNormalizeOrderOperationalState(estadoPedido)
	res, err := execSQLCompat(dbConn, `UPDATE empresa_venta_publica_ordenes
		SET estado_pedido = ?,
			taxi_request_id = CASE WHEN ? > 0 THEN ? ELSE taxi_request_id END,
			observaciones = CASE WHEN ? = '' THEN observaciones ELSE ? END,
			fecha_actualizacion = CURRENT_TIMESTAMP
		WHERE empresa_id = ? AND codigo_orden = ?`,
		estadoPedido,
		taxiRequestID,
		taxiRequestID,
		strings.TrimSpace(observaciones),
		strings.TrimSpace(observaciones),
		empresaID,
		strings.TrimSpace(codigoOrden),
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

func UpdateEmpresaVentaPublicaOrderTracking(dbConn *sql.DB, empresaID int64, codigoOrden string, taxiRequestID int64, trackingToken string) error {
	if dbConn == nil {
		return errors.New("db connection is nil")
	}
	if empresaID <= 0 || strings.TrimSpace(codigoOrden) == "" {
		return fmt.Errorf("empresa_id o codigo_orden invalido")
	}
	res, err := execSQLCompat(dbConn, `UPDATE empresa_venta_publica_ordenes
		SET taxi_request_id = CASE WHEN ? > 0 THEN ? ELSE taxi_request_id END,
			tracking_token = CASE WHEN ? = '' THEN tracking_token ELSE ? END,
			fecha_actualizacion = CURRENT_TIMESTAMP
		WHERE empresa_id = ? AND codigo_orden = ?`,
		taxiRequestID,
		taxiRequestID,
		strings.TrimSpace(trackingToken),
		strings.TrimSpace(trackingToken),
		empresaID,
		strings.TrimSpace(codigoOrden),
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
