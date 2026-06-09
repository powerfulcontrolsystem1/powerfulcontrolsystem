package db

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// SuperVentaDigitalConfig define la configuracion global de la tienda digital.
type SuperVentaDigitalConfig struct {
	ID                 int64  `json:"id"`
	ScopeKey           string `json:"scope_key"`
	NombreTienda       string `json:"nombre_tienda"`
	DescripcionTienda  string `json:"descripcion_tienda,omitempty"`
	LogoURL            string `json:"logo_url,omitempty"`
	BannerURL          string `json:"banner_url,omitempty"`
	ColorPrimario      string `json:"color_primario,omitempty"`
	Moneda             string `json:"moneda"`
	WompiActivo        bool   `json:"wompi_activo"`
	FechaCreacion      string `json:"fecha_creacion,omitempty"`
	FechaActualizacion string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string `json:"usuario_creador,omitempty"`
	Estado             string `json:"estado,omitempty"`
	Observaciones      string `json:"observaciones,omitempty"`
}

// SuperVentaDigitalItem representa un producto digital del catalogo global.
type SuperVentaDigitalItem struct {
	ID                      int64   `json:"id"`
	CodigoPublico           string  `json:"codigo_publico"`
	Nombre                  string  `json:"nombre"`
	Descripcion             string  `json:"descripcion,omitempty"`
	Precio                  float64 `json:"precio"`
	Moneda                  string  `json:"moneda"`
	ImagenURL               string  `json:"imagen_url,omitempty"`
	LicenciaCodigo          string  `json:"licencia_codigo,omitempty"`
	InstruccionesArchivoURL string  `json:"instrucciones_archivo_url,omitempty"`
	OrdenVisual             int     `json:"orden_visual,omitempty"`
	Destacado               bool    `json:"destacado"`
	FechaCreacion           string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion      string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador          string  `json:"usuario_creador,omitempty"`
	Estado                  string  `json:"estado,omitempty"`
	Observaciones           string  `json:"observaciones,omitempty"`
}

// SuperVentaDigitalItemsFilter define filtros para listado de catalogo.
type SuperVentaDigitalItemsFilter struct {
	IncludeInactive bool
	Q               string
	Limit           int
	Offset          int
}

// SuperVentaDigitalOrder representa una compra publica de producto digital.
type SuperVentaDigitalOrder struct {
	ID                      int64   `json:"id"`
	CodigoOrden             string  `json:"codigo_orden"`
	ItemID                  int64   `json:"item_id"`
	ItemNombre              string  `json:"item_nombre"`
	ItemPrecio              float64 `json:"item_precio"`
	ItemMoneda              string  `json:"item_moneda"`
	CompradorNombre         string  `json:"comprador_nombre,omitempty"`
	CompradorEmail          string  `json:"comprador_email,omitempty"`
	CompradorTelefono       string  `json:"comprador_telefono,omitempty"`
	MetodoPago              string  `json:"metodo_pago"`
	EstadoPago              string  `json:"estado_pago"`
	TransactionID           string  `json:"transaction_id,omitempty"`
	ReferenciaExterna       string  `json:"referencia_externa,omitempty"`
	PasarelaPayloadJSON     string  `json:"pasarela_payload_json,omitempty"`
	PagadoEn                string  `json:"pagado_en,omitempty"`
	CorreoEntregado         bool    `json:"correo_entregado"`
	CorreoEntregadoEn       string  `json:"correo_entregado_en,omitempty"`
	CorreoEntregaError      string  `json:"correo_entrega_error,omitempty"`
	LicenciaCodigoEnviado   string  `json:"licencia_codigo_enviado,omitempty"`
	InstruccionesArchivoURL string  `json:"instrucciones_archivo_url,omitempty"`
	FechaCreacion           string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion      string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador          string  `json:"usuario_creador,omitempty"`
	Estado                  string  `json:"estado,omitempty"`
	Observaciones           string  `json:"observaciones,omitempty"`
}

// SuperVentaDigitalOrdersFilter define filtros para listado de ordenes.
type SuperVentaDigitalOrdersFilter struct {
	IncludeInactive bool
	EstadoPago      string
	Q               string
	Limit           int
	Offset          int
}

func ventaDigitalNormalizeEstado(raw string) string {
	if strings.EqualFold(strings.TrimSpace(raw), "inactivo") {
		return "inactivo"
	}
	return "activo"
}

func ventaDigitalNormalizeMoneda(raw string) string {
	moneda := strings.ToUpper(strings.TrimSpace(raw))
	if moneda == "" {
		return "COP"
	}
	if len(moneda) > 8 {
		moneda = moneda[:8]
	}
	return moneda
}

func ventaDigitalNormalizeLimitOffset(limit, offset int) (int, int) {
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

func ventaDigitalLikePattern(raw string) string {
	value := strings.TrimSpace(raw)
	value = strings.ReplaceAll(value, "!", "!!")
	value = strings.ReplaceAll(value, "%", "!%")
	value = strings.ReplaceAll(value, "_", "!_")
	return "%" + value + "%"
}

func ventaDigitalBoolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

func ventaDigitalGenerateItemCode(nombre string) string {
	base := NormalizeEmpresaPublicSlug(nombre)
	if base == "" || base == "empresa" {
		base = "digital"
	}
	if len(base) > 22 {
		base = base[:22]
	}
	stamp := time.Now().In(time.Local).Format("20060102150405")
	return strings.ToUpper("VD-" + base + "-" + stamp)
}

func ventaDigitalGenerateOrderCode() string {
	stamp := time.Now().In(time.Local).Format("20060102150405")
	return "VD-ORD-" + stamp
}

// BuildSuperVentaDigitalOrderReference genera referencia externa para Wompi.
func BuildSuperVentaDigitalOrderReference(orderCode string) string {
	code := strings.TrimSpace(orderCode)
	if code == "" {
		code = ventaDigitalGenerateOrderCode()
	}
	return "VD|" + code
}

// TryParseSuperVentaDigitalOrderCodeFromReference intenta recuperar codigo de orden desde una referencia.
func TryParseSuperVentaDigitalOrderCodeFromReference(reference string) string {
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
	if strings.Contains(reference, "VD-ORD-") {
		idx := strings.Index(reference, "VD-ORD-")
		if idx >= 0 {
			candidate := strings.TrimSpace(reference[idx:])
			if candidate != "" {
				return candidate
			}
		}
	}
	if strings.HasPrefix(reference, "VD") {
		return reference
	}
	return ""
}

// EnsureSuperVentaDigitalSchema crea y migra tablas de venta digital global en PostgreSQL.
func EnsureSuperVentaDigitalSchema(dbConn *sql.DB) error {
	if dbConn == nil {
		return errors.New("db connection is nil")
	}

	stmts := []string{
		`CREATE TABLE IF NOT EXISTS super_venta_digital_configuracion (
			id BIGSERIAL PRIMARY KEY,
			scope_key TEXT NOT NULL UNIQUE DEFAULT 'global',
			nombre_tienda TEXT,
			descripcion_tienda TEXT,
			logo_url TEXT,
			banner_url TEXT,
			color_primario TEXT DEFAULT '#0f4c81',
			moneda TEXT DEFAULT 'COP',
			wompi_activo INTEGER DEFAULT 1,
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS super_venta_digital_items (
			id BIGSERIAL PRIMARY KEY,
			codigo_publico TEXT NOT NULL UNIQUE,
			nombre TEXT NOT NULL,
			descripcion TEXT,
			precio REAL DEFAULT 0,
			moneda TEXT DEFAULT 'COP',
			imagen_url TEXT,
			licencia_codigo TEXT NOT NULL,
			instrucciones_archivo_url TEXT NOT NULL,
			orden_visual INTEGER DEFAULT 0,
			destacado INTEGER DEFAULT 0,
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE INDEX IF NOT EXISTS ix_super_venta_digital_items_estado_orden ON super_venta_digital_items(estado, orden_visual, id DESC);`,
		`CREATE TABLE IF NOT EXISTS super_venta_digital_ordenes (
			id BIGSERIAL PRIMARY KEY,
			codigo_orden TEXT NOT NULL UNIQUE,
			item_id INTEGER NOT NULL,
			item_nombre TEXT,
			item_precio REAL DEFAULT 0,
			item_moneda TEXT DEFAULT 'COP',
			comprador_nombre TEXT,
			comprador_email TEXT,
			comprador_telefono TEXT,
			metodo_pago TEXT DEFAULT 'wompi_nequi',
			estado_pago TEXT DEFAULT 'pendiente',
			transaction_id TEXT,
			referencia_externa TEXT,
			pasarela_payload_json TEXT,
			pagado_en TEXT,
			correo_entregado INTEGER DEFAULT 0,
			correo_entregado_en TEXT,
			correo_entrega_error TEXT,
			licencia_codigo_enviado TEXT,
			instrucciones_archivo_url TEXT,
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE INDEX IF NOT EXISTS ix_super_venta_digital_ordenes_estado ON super_venta_digital_ordenes(estado_pago, fecha_creacion DESC);`,
		`CREATE INDEX IF NOT EXISTS ix_super_venta_digital_ordenes_tx ON super_venta_digital_ordenes(transaction_id);`,
		`CREATE INDEX IF NOT EXISTS ix_super_venta_digital_ordenes_ref ON super_venta_digital_ordenes(referencia_externa);`,
	}

	for _, stmt := range stmts {
		if _, err := dbConn.Exec(stmt); err != nil {
			return err
		}
	}

	if err := ensureColumnIfMissing(dbConn, "super_venta_digital_configuracion", "scope_key", "TEXT DEFAULT 'global'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "super_venta_digital_configuracion", "wompi_activo", "INTEGER DEFAULT 1"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "super_venta_digital_items", "licencia_codigo", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "super_venta_digital_items", "instrucciones_archivo_url", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "super_venta_digital_ordenes", "correo_entregado", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "super_venta_digital_ordenes", "correo_entregado_en", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "super_venta_digital_ordenes", "correo_entrega_error", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "super_venta_digital_ordenes", "licencia_codigo_enviado", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "super_venta_digital_ordenes", "instrucciones_archivo_url", "TEXT"); err != nil {
		return err
	}

	_, _ = dbConn.Exec(`INSERT INTO super_venta_digital_configuracion (scope_key, nombre_tienda, descripcion_tienda, moneda, wompi_activo, usuario_creador, estado)
		SELECT 'global', 'Venta Digital Powerful Control System', 'Catalogo digital global administrado por super administrador', 'COP', 1, 'sistema', 'activo'
		WHERE NOT EXISTS (SELECT 1 FROM super_venta_digital_configuracion WHERE scope_key='global')`)

	return nil
}

// GetSuperVentaDigitalConfig obtiene la configuracion global de venta digital.
func GetSuperVentaDigitalConfig(dbConn *sql.DB) (SuperVentaDigitalConfig, error) {
	if dbConn == nil {
		return SuperVentaDigitalConfig{}, errors.New("db connection is nil")
	}

	var out SuperVentaDigitalConfig
	var wompiActivo sql.NullInt64
	err := dbConn.QueryRow(`SELECT
		id,
		COALESCE(scope_key, 'global'),
		COALESCE(nombre_tienda, ''),
		COALESCE(descripcion_tienda, ''),
		COALESCE(logo_url, ''),
		COALESCE(banner_url, ''),
		COALESCE(color_primario, '#0f4c81'),
		COALESCE(moneda, 'COP'),
		COALESCE(wompi_activo, 1),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM super_venta_digital_configuracion
	WHERE scope_key = 'global'
	LIMIT 1`).Scan(
		&out.ID,
		&out.ScopeKey,
		&out.NombreTienda,
		&out.DescripcionTienda,
		&out.LogoURL,
		&out.BannerURL,
		&out.ColorPrimario,
		&out.Moneda,
		&wompiActivo,
		&out.FechaCreacion,
		&out.FechaActualizacion,
		&out.UsuarioCreador,
		&out.Estado,
		&out.Observaciones,
	)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return SuperVentaDigitalConfig{}, err
		}
		return SuperVentaDigitalConfig{
			ScopeKey:          "global",
			NombreTienda:      "Venta Digital Powerful Control System",
			DescripcionTienda: "Catalogo digital global administrado por super administrador",
			ColorPrimario:     "#0f4c81",
			Moneda:            "COP",
			WompiActivo:       true,
			Estado:            "activo",
		}, nil
	}

	out.Moneda = ventaDigitalNormalizeMoneda(out.Moneda)
	out.Estado = ventaDigitalNormalizeEstado(out.Estado)
	out.WompiActivo = wompiActivo.Valid && wompiActivo.Int64 > 0
	if !wompiActivo.Valid {
		out.WompiActivo = true
	}
	if strings.TrimSpace(out.ScopeKey) == "" {
		out.ScopeKey = "global"
	}
	if strings.TrimSpace(out.NombreTienda) == "" {
		out.NombreTienda = "Venta Digital Powerful Control System"
	}
	if strings.TrimSpace(out.ColorPrimario) == "" {
		out.ColorPrimario = "#0f4c81"
	}
	return out, nil
}

// UpsertSuperVentaDigitalConfig crea/actualiza la configuracion global de venta digital.
func UpsertSuperVentaDigitalConfig(dbConn *sql.DB, cfg SuperVentaDigitalConfig) (int64, error) {
	if dbConn == nil {
		return 0, errors.New("db connection is nil")
	}

	cfg.ScopeKey = "global"
	cfg.NombreTienda = strings.TrimSpace(cfg.NombreTienda)
	if cfg.NombreTienda == "" {
		cfg.NombreTienda = "Venta Digital Powerful Control System"
	}
	cfg.DescripcionTienda = strings.TrimSpace(cfg.DescripcionTienda)
	if cfg.DescripcionTienda == "" {
		cfg.DescripcionTienda = "Catalogo digital global administrado por super administrador"
	}
	if strings.TrimSpace(cfg.ColorPrimario) == "" {
		cfg.ColorPrimario = "#0f4c81"
	}
	cfg.Moneda = ventaDigitalNormalizeMoneda(cfg.Moneda)
	cfg.Estado = ventaDigitalNormalizeEstado(cfg.Estado)

	var existingID int64
	err := dbConn.QueryRow(`SELECT id FROM super_venta_digital_configuracion WHERE scope_key = 'global' LIMIT 1`).Scan(&existingID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}

	if existingID > 0 {
		_, err = dbConn.Exec(`UPDATE super_venta_digital_configuracion
			SET nombre_tienda = ?,
				descripcion_tienda = ?,
				logo_url = ?,
				banner_url = ?,
				color_primario = ?,
				moneda = ?,
				wompi_activo = ?,
				usuario_creador = ?,
				estado = ?,
				observaciones = ?,
				fecha_actualizacion = CURRENT_TIMESTAMP
			WHERE id = ?`,
			cfg.NombreTienda,
			cfg.DescripcionTienda,
			strings.TrimSpace(cfg.LogoURL),
			strings.TrimSpace(cfg.BannerURL),
			strings.TrimSpace(cfg.ColorPrimario),
			cfg.Moneda,
			ventaDigitalBoolToInt(cfg.WompiActivo),
			strings.TrimSpace(cfg.UsuarioCreador),
			cfg.Estado,
			strings.TrimSpace(cfg.Observaciones),
			existingID,
		)
		if err != nil {
			return 0, err
		}
		return existingID, nil
	}

	res, err := dbConn.Exec(`INSERT INTO super_venta_digital_configuracion (
		scope_key,
		nombre_tienda,
		descripcion_tienda,
		logo_url,
		banner_url,
		color_primario,
		moneda,
		wompi_activo,
		usuario_creador,
		estado,
		observaciones
	) VALUES ('global', ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		cfg.NombreTienda,
		cfg.DescripcionTienda,
		strings.TrimSpace(cfg.LogoURL),
		strings.TrimSpace(cfg.BannerURL),
		strings.TrimSpace(cfg.ColorPrimario),
		cfg.Moneda,
		ventaDigitalBoolToInt(cfg.WompiActivo),
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

// CreateSuperVentaDigitalItem crea un item de catalogo digital.
func CreateSuperVentaDigitalItem(dbConn *sql.DB, item SuperVentaDigitalItem) (int64, error) {
	if dbConn == nil {
		return 0, errors.New("db connection is nil")
	}
	item.Nombre = strings.TrimSpace(item.Nombre)
	if item.Nombre == "" {
		return 0, fmt.Errorf("nombre es obligatorio")
	}
	if item.Precio < 0 {
		return 0, fmt.Errorf("precio invalido")
	}
	item.LicenciaCodigo = strings.TrimSpace(item.LicenciaCodigo)
	if item.LicenciaCodigo == "" {
		return 0, fmt.Errorf("licencia_codigo es obligatorio")
	}
	item.InstruccionesArchivoURL = strings.TrimSpace(item.InstruccionesArchivoURL)
	if item.InstruccionesArchivoURL == "" {
		return 0, fmt.Errorf("instrucciones_archivo_url es obligatorio")
	}
	item.CodigoPublico = strings.TrimSpace(item.CodigoPublico)
	if item.CodigoPublico == "" {
		item.CodigoPublico = ventaDigitalGenerateItemCode(item.Nombre)
	}
	item.Moneda = ventaDigitalNormalizeMoneda(item.Moneda)
	item.Estado = ventaDigitalNormalizeEstado(item.Estado)

	res, err := dbConn.Exec(`INSERT INTO super_venta_digital_items (
		codigo_publico,
		nombre,
		descripcion,
		precio,
		moneda,
		imagen_url,
		licencia_codigo,
		instrucciones_archivo_url,
		orden_visual,
		destacado,
		usuario_creador,
		estado,
		observaciones
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		item.CodigoPublico,
		item.Nombre,
		strings.TrimSpace(item.Descripcion),
		item.Precio,
		item.Moneda,
		strings.TrimSpace(item.ImagenURL),
		item.LicenciaCodigo,
		item.InstruccionesArchivoURL,
		item.OrdenVisual,
		ventaDigitalBoolToInt(item.Destacado),
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

// UpdateSuperVentaDigitalItem actualiza un item del catalogo digital.
func UpdateSuperVentaDigitalItem(dbConn *sql.DB, item SuperVentaDigitalItem) error {
	if dbConn == nil {
		return errors.New("db connection is nil")
	}
	if item.ID <= 0 {
		return fmt.Errorf("id invalido")
	}
	item.Nombre = strings.TrimSpace(item.Nombre)
	if item.Nombre == "" {
		return fmt.Errorf("nombre es obligatorio")
	}
	if item.Precio < 0 {
		return fmt.Errorf("precio invalido")
	}
	item.LicenciaCodigo = strings.TrimSpace(item.LicenciaCodigo)
	if item.LicenciaCodigo == "" {
		return fmt.Errorf("licencia_codigo es obligatorio")
	}
	item.InstruccionesArchivoURL = strings.TrimSpace(item.InstruccionesArchivoURL)
	if item.InstruccionesArchivoURL == "" {
		return fmt.Errorf("instrucciones_archivo_url es obligatorio")
	}
	item.CodigoPublico = strings.TrimSpace(item.CodigoPublico)
	if item.CodigoPublico == "" {
		item.CodigoPublico = ventaDigitalGenerateItemCode(item.Nombre)
	}
	item.Moneda = ventaDigitalNormalizeMoneda(item.Moneda)
	item.Estado = ventaDigitalNormalizeEstado(item.Estado)

	res, err := dbConn.Exec(`UPDATE super_venta_digital_items
		SET codigo_publico = ?,
			nombre = ?,
			descripcion = ?,
			precio = ?,
			moneda = ?,
			imagen_url = ?,
			licencia_codigo = ?,
			instrucciones_archivo_url = ?,
			orden_visual = ?,
			destacado = ?,
			usuario_creador = ?,
			estado = ?,
			observaciones = ?,
			fecha_actualizacion = CURRENT_TIMESTAMP
		WHERE id = ?`,
		item.CodigoPublico,
		item.Nombre,
		strings.TrimSpace(item.Descripcion),
		item.Precio,
		item.Moneda,
		strings.TrimSpace(item.ImagenURL),
		item.LicenciaCodigo,
		item.InstruccionesArchivoURL,
		item.OrdenVisual,
		ventaDigitalBoolToInt(item.Destacado),
		strings.TrimSpace(item.UsuarioCreador),
		item.Estado,
		strings.TrimSpace(item.Observaciones),
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

// SetSuperVentaDigitalItemEstadoByID activa/desactiva un item de venta digital.
func SetSuperVentaDigitalItemEstadoByID(dbConn *sql.DB, itemID int64, estado string) error {
	if dbConn == nil {
		return errors.New("db connection is nil")
	}
	if itemID <= 0 {
		return fmt.Errorf("id invalido")
	}
	res, err := dbConn.Exec(`UPDATE super_venta_digital_items
		SET estado = ?, fecha_actualizacion = CURRENT_TIMESTAMP
		WHERE id = ?`, ventaDigitalNormalizeEstado(estado), itemID)
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

// GetSuperVentaDigitalItemByID obtiene un item por su id.
func GetSuperVentaDigitalItemByID(dbConn *sql.DB, itemID int64) (SuperVentaDigitalItem, error) {
	if dbConn == nil {
		return SuperVentaDigitalItem{}, errors.New("db connection is nil")
	}
	if itemID <= 0 {
		return SuperVentaDigitalItem{}, fmt.Errorf("id invalido")
	}

	var out SuperVentaDigitalItem
	var destacado sql.NullInt64
	err := dbConn.QueryRow(`SELECT
		id,
		COALESCE(codigo_publico, ''),
		COALESCE(nombre, ''),
		COALESCE(descripcion, ''),
		COALESCE(precio, 0),
		COALESCE(moneda, 'COP'),
		COALESCE(imagen_url, ''),
		COALESCE(licencia_codigo, ''),
		COALESCE(instrucciones_archivo_url, ''),
		COALESCE(orden_visual, 0),
		COALESCE(destacado, 0),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM super_venta_digital_items
	WHERE id = ?
	LIMIT 1`, itemID).Scan(
		&out.ID,
		&out.CodigoPublico,
		&out.Nombre,
		&out.Descripcion,
		&out.Precio,
		&out.Moneda,
		&out.ImagenURL,
		&out.LicenciaCodigo,
		&out.InstruccionesArchivoURL,
		&out.OrdenVisual,
		&destacado,
		&out.FechaCreacion,
		&out.FechaActualizacion,
		&out.UsuarioCreador,
		&out.Estado,
		&out.Observaciones,
	)
	if err != nil {
		return SuperVentaDigitalItem{}, err
	}
	out.Destacado = destacado.Valid && destacado.Int64 > 0
	out.Moneda = ventaDigitalNormalizeMoneda(out.Moneda)
	out.Estado = ventaDigitalNormalizeEstado(out.Estado)
	return out, nil
}

// ListSuperVentaDigitalItems lista catalogo de venta digital con filtros.
func ListSuperVentaDigitalItems(dbConn *sql.DB, filter SuperVentaDigitalItemsFilter) ([]SuperVentaDigitalItem, int64, error) {
	if dbConn == nil {
		return nil, 0, errors.New("db connection is nil")
	}

	limit, offset := ventaDigitalNormalizeLimitOffset(filter.Limit, filter.Offset)
	where := "WHERE 1=1"
	args := make([]interface{}, 0)

	if !filter.IncludeInactive {
		where += " AND COALESCE(estado, 'activo') <> 'inactivo'"
	}
	if q := strings.TrimSpace(filter.Q); q != "" {
		pattern := ventaDigitalLikePattern(q)
		where += ` AND (
			LOWER(COALESCE(codigo_publico, '')) LIKE LOWER(?) ESCAPE '!' OR
			LOWER(COALESCE(nombre, '')) LIKE LOWER(?) ESCAPE '!' OR
			LOWER(COALESCE(descripcion, '')) LIKE LOWER(?) ESCAPE '!'
		)`
		args = append(args, pattern, pattern, pattern)
	}

	var total int64
	if err := dbConn.QueryRow("SELECT COUNT(1) FROM super_venta_digital_items "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `SELECT
		id,
		COALESCE(codigo_publico, ''),
		COALESCE(nombre, ''),
		COALESCE(descripcion, ''),
		COALESCE(precio, 0),
		COALESCE(moneda, 'COP'),
		COALESCE(imagen_url, ''),
		COALESCE(licencia_codigo, ''),
		COALESCE(instrucciones_archivo_url, ''),
		COALESCE(orden_visual, 0),
		COALESCE(destacado, 0),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM super_venta_digital_items ` + where + `
	ORDER BY COALESCE(orden_visual, 0) ASC, id DESC
	LIMIT ? OFFSET ?`

	rows, err := dbConn.Query(query, append(args, limit, offset)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := make([]SuperVentaDigitalItem, 0)
	for rows.Next() {
		var item SuperVentaDigitalItem
		var destacado sql.NullInt64
		if err := rows.Scan(
			&item.ID,
			&item.CodigoPublico,
			&item.Nombre,
			&item.Descripcion,
			&item.Precio,
			&item.Moneda,
			&item.ImagenURL,
			&item.LicenciaCodigo,
			&item.InstruccionesArchivoURL,
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
		item.Moneda = ventaDigitalNormalizeMoneda(item.Moneda)
		item.Estado = ventaDigitalNormalizeEstado(item.Estado)
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

// ListSuperVentaDigitalItemsPublic lista solo items activos para vista publica.
func ListSuperVentaDigitalItemsPublic(dbConn *sql.DB) ([]SuperVentaDigitalItem, error) {
	rows, _, err := ListSuperVentaDigitalItems(dbConn, SuperVentaDigitalItemsFilter{
		IncludeInactive: false,
		Limit:           500,
		Offset:          0,
	})
	if err != nil {
		return nil, err
	}
	return rows, nil
}

// CreateSuperVentaDigitalOrder crea una orden publica para compra de producto digital.
func CreateSuperVentaDigitalOrder(dbConn *sql.DB, order SuperVentaDigitalOrder) (int64, error) {
	if dbConn == nil {
		return 0, errors.New("db connection is nil")
	}
	if order.ItemID <= 0 {
		return 0, fmt.Errorf("item_id invalido")
	}
	if order.ItemPrecio < 0 {
		return 0, fmt.Errorf("item_precio invalido")
	}
	order.CodigoOrden = strings.TrimSpace(order.CodigoOrden)
	if order.CodigoOrden == "" {
		order.CodigoOrden = ventaDigitalGenerateOrderCode()
	}
	if strings.TrimSpace(order.MetodoPago) == "" {
		order.MetodoPago = "wompi_nequi"
	}
	if strings.TrimSpace(order.EstadoPago) == "" {
		order.EstadoPago = "pendiente"
	}
	order.ItemMoneda = ventaDigitalNormalizeMoneda(order.ItemMoneda)
	order.Estado = ventaDigitalNormalizeEstado(order.Estado)

	res, err := dbConn.Exec(`INSERT INTO super_venta_digital_ordenes (
		codigo_orden,
		item_id,
		item_nombre,
		item_precio,
		item_moneda,
		comprador_nombre,
		comprador_email,
		comprador_telefono,
		metodo_pago,
		estado_pago,
		transaction_id,
		referencia_externa,
		pasarela_payload_json,
		pagado_en,
		correo_entregado,
		correo_entregado_en,
		correo_entrega_error,
		licencia_codigo_enviado,
		instrucciones_archivo_url,
		usuario_creador,
		estado,
		observaciones
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		order.CodigoOrden,
		order.ItemID,
		strings.TrimSpace(order.ItemNombre),
		order.ItemPrecio,
		order.ItemMoneda,
		strings.TrimSpace(order.CompradorNombre),
		strings.TrimSpace(order.CompradorEmail),
		strings.TrimSpace(order.CompradorTelefono),
		strings.TrimSpace(order.MetodoPago),
		strings.TrimSpace(order.EstadoPago),
		strings.TrimSpace(order.TransactionID),
		strings.TrimSpace(order.ReferenciaExterna),
		strings.TrimSpace(order.PasarelaPayloadJSON),
		strings.TrimSpace(order.PagadoEn),
		ventaDigitalBoolToInt(order.CorreoEntregado),
		strings.TrimSpace(order.CorreoEntregadoEn),
		strings.TrimSpace(order.CorreoEntregaError),
		strings.TrimSpace(order.LicenciaCodigoEnviado),
		strings.TrimSpace(order.InstruccionesArchivoURL),
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

// UpdateSuperVentaDigitalOrderPayment actualiza estado de pago y datos de transaccion.
func UpdateSuperVentaDigitalOrderPayment(dbConn *sql.DB, codigoOrden, estadoPago, transactionID, referenciaExterna, payloadJSON, pagadoEn, observaciones string) error {
	if dbConn == nil {
		return errors.New("db connection is nil")
	}
	codigoOrden = strings.TrimSpace(codigoOrden)
	if codigoOrden == "" {
		return fmt.Errorf("codigo_orden invalido")
	}
	estado := strings.TrimSpace(estadoPago)
	if estado == "" {
		estado = "pendiente"
	}

	res, err := dbConn.Exec(`UPDATE super_venta_digital_ordenes
		SET estado_pago = ?,
			transaction_id = CASE WHEN ? = '' THEN transaction_id ELSE ? END,
			referencia_externa = CASE WHEN ? = '' THEN referencia_externa ELSE ? END,
			pasarela_payload_json = CASE WHEN ? = '' THEN pasarela_payload_json ELSE ? END,
			pagado_en = CASE WHEN ? = '' THEN pagado_en ELSE ? END,
			observaciones = CASE WHEN ? = '' THEN observaciones ELSE ? END,
			fecha_actualizacion = CURRENT_TIMESTAMP
		WHERE codigo_orden = ?`,
		estado,
		strings.TrimSpace(transactionID), strings.TrimSpace(transactionID),
		strings.TrimSpace(referenciaExterna), strings.TrimSpace(referenciaExterna),
		strings.TrimSpace(payloadJSON), strings.TrimSpace(payloadJSON),
		strings.TrimSpace(pagadoEn), strings.TrimSpace(pagadoEn),
		strings.TrimSpace(observaciones), strings.TrimSpace(observaciones),
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

// SetSuperVentaDigitalOrderDelivery marca estado de entrega por correo de una orden.
func SetSuperVentaDigitalOrderDelivery(dbConn *sql.DB, codigoOrden string, delivered bool, deliveredAt, deliveryError, licenciaCodigoEnviado, instruccionesArchivoURL string) error {
	if dbConn == nil {
		return errors.New("db connection is nil")
	}
	codigoOrden = strings.TrimSpace(codigoOrden)
	if codigoOrden == "" {
		return fmt.Errorf("codigo_orden invalido")
	}

	res, err := dbConn.Exec(`UPDATE super_venta_digital_ordenes
		SET correo_entregado = ?,
			correo_entregado_en = CASE WHEN ? = '' THEN correo_entregado_en ELSE ? END,
			correo_entrega_error = CASE WHEN ? = '' THEN correo_entrega_error ELSE ? END,
			licencia_codigo_enviado = CASE WHEN ? = '' THEN licencia_codigo_enviado ELSE ? END,
			instrucciones_archivo_url = CASE WHEN ? = '' THEN instrucciones_archivo_url ELSE ? END,
			fecha_actualizacion = CURRENT_TIMESTAMP
		WHERE codigo_orden = ?`,
		ventaDigitalBoolToInt(delivered),
		strings.TrimSpace(deliveredAt), strings.TrimSpace(deliveredAt),
		strings.TrimSpace(deliveryError), strings.TrimSpace(deliveryError),
		strings.TrimSpace(licenciaCodigoEnviado), strings.TrimSpace(licenciaCodigoEnviado),
		strings.TrimSpace(instruccionesArchivoURL), strings.TrimSpace(instruccionesArchivoURL),
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

// GetSuperVentaDigitalOrderByCodigo obtiene una orden por codigo.
func GetSuperVentaDigitalOrderByCodigo(dbConn *sql.DB, codigoOrden string) (SuperVentaDigitalOrder, error) {
	if dbConn == nil {
		return SuperVentaDigitalOrder{}, errors.New("db connection is nil")
	}
	codigoOrden = strings.TrimSpace(codigoOrden)
	if codigoOrden == "" {
		return SuperVentaDigitalOrder{}, fmt.Errorf("codigo_orden invalido")
	}

	var out SuperVentaDigitalOrder
	var correoEntregado sql.NullInt64
	err := dbConn.QueryRow(`SELECT
		id,
		COALESCE(codigo_orden, ''),
		COALESCE(item_id, 0),
		COALESCE(item_nombre, ''),
		COALESCE(item_precio, 0),
		COALESCE(item_moneda, 'COP'),
		COALESCE(comprador_nombre, ''),
		COALESCE(comprador_email, ''),
		COALESCE(comprador_telefono, ''),
		COALESCE(metodo_pago, 'wompi_nequi'),
		COALESCE(estado_pago, 'pendiente'),
		COALESCE(transaction_id, ''),
		COALESCE(referencia_externa, ''),
		COALESCE(pasarela_payload_json, ''),
		COALESCE(pagado_en, ''),
		COALESCE(correo_entregado, 0),
		COALESCE(correo_entregado_en, ''),
		COALESCE(correo_entrega_error, ''),
		COALESCE(licencia_codigo_enviado, ''),
		COALESCE(instrucciones_archivo_url, ''),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM super_venta_digital_ordenes
	WHERE codigo_orden = ?
	LIMIT 1`, codigoOrden).Scan(
		&out.ID,
		&out.CodigoOrden,
		&out.ItemID,
		&out.ItemNombre,
		&out.ItemPrecio,
		&out.ItemMoneda,
		&out.CompradorNombre,
		&out.CompradorEmail,
		&out.CompradorTelefono,
		&out.MetodoPago,
		&out.EstadoPago,
		&out.TransactionID,
		&out.ReferenciaExterna,
		&out.PasarelaPayloadJSON,
		&out.PagadoEn,
		&correoEntregado,
		&out.CorreoEntregadoEn,
		&out.CorreoEntregaError,
		&out.LicenciaCodigoEnviado,
		&out.InstruccionesArchivoURL,
		&out.FechaCreacion,
		&out.FechaActualizacion,
		&out.UsuarioCreador,
		&out.Estado,
		&out.Observaciones,
	)
	if err != nil {
		return SuperVentaDigitalOrder{}, err
	}
	out.CorreoEntregado = correoEntregado.Valid && correoEntregado.Int64 > 0
	out.ItemMoneda = ventaDigitalNormalizeMoneda(out.ItemMoneda)
	out.Estado = ventaDigitalNormalizeEstado(out.Estado)
	if strings.TrimSpace(out.EstadoPago) == "" {
		out.EstadoPago = "pendiente"
	}
	return out, nil
}

// FindSuperVentaDigitalOrderByPaymentContext busca una orden por transaction_id o referencia externa.
func FindSuperVentaDigitalOrderByPaymentContext(dbConn *sql.DB, transactionID, reference string) (SuperVentaDigitalOrder, bool, error) {
	if dbConn == nil {
		return SuperVentaDigitalOrder{}, false, errors.New("db connection is nil")
	}

	transactionID = strings.TrimSpace(transactionID)
	reference = strings.TrimSpace(reference)

	if transactionID != "" {
		var codigo string
		err := dbConn.QueryRow(`SELECT COALESCE(codigo_orden, '')
			FROM super_venta_digital_ordenes
			WHERE transaction_id = ?
			ORDER BY id DESC
			LIMIT 1`, transactionID).Scan(&codigo)
		if err == nil && strings.TrimSpace(codigo) != "" {
			order, getErr := GetSuperVentaDigitalOrderByCodigo(dbConn, codigo)
			if getErr != nil {
				return SuperVentaDigitalOrder{}, false, getErr
			}
			return order, true, nil
		}
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return SuperVentaDigitalOrder{}, false, err
		}
	}

	if reference != "" {
		var codigo string
		err := dbConn.QueryRow(`SELECT COALESCE(codigo_orden, '')
			FROM super_venta_digital_ordenes
			WHERE referencia_externa = ?
			ORDER BY id DESC
			LIMIT 1`, reference).Scan(&codigo)
		if err == nil && strings.TrimSpace(codigo) != "" {
			order, getErr := GetSuperVentaDigitalOrderByCodigo(dbConn, codigo)
			if getErr != nil {
				return SuperVentaDigitalOrder{}, false, getErr
			}
			return order, true, nil
		}
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return SuperVentaDigitalOrder{}, false, err
		}

		if codeFromRef := TryParseSuperVentaDigitalOrderCodeFromReference(reference); codeFromRef != "" {
			order, getErr := GetSuperVentaDigitalOrderByCodigo(dbConn, codeFromRef)
			if getErr == nil {
				return order, true, nil
			}
			if !errors.Is(getErr, sql.ErrNoRows) {
				return SuperVentaDigitalOrder{}, false, getErr
			}
		}
	}

	return SuperVentaDigitalOrder{}, false, nil
}

// ListSuperVentaDigitalOrders lista ordenes de venta digital con filtros.
func ListSuperVentaDigitalOrders(dbConn *sql.DB, filter SuperVentaDigitalOrdersFilter) ([]SuperVentaDigitalOrder, int64, error) {
	if dbConn == nil {
		return nil, 0, errors.New("db connection is nil")
	}

	limit, offset := ventaDigitalNormalizeLimitOffset(filter.Limit, filter.Offset)
	where := "WHERE 1=1"
	args := make([]interface{}, 0)

	if !filter.IncludeInactive {
		where += " AND COALESCE(estado, 'activo') <> 'inactivo'"
	}
	if status := strings.TrimSpace(strings.ToLower(filter.EstadoPago)); status != "" {
		where += " AND LOWER(COALESCE(estado_pago, '')) = ?"
		args = append(args, status)
	}
	if q := strings.TrimSpace(filter.Q); q != "" {
		pattern := ventaDigitalLikePattern(q)
		where += ` AND (
			LOWER(COALESCE(codigo_orden, '')) LIKE LOWER(?) ESCAPE '!' OR
			LOWER(COALESCE(comprador_nombre, '')) LIKE LOWER(?) ESCAPE '!' OR
			LOWER(COALESCE(comprador_email, '')) LIKE LOWER(?) ESCAPE '!' OR
			LOWER(COALESCE(referencia_externa, '')) LIKE LOWER(?) ESCAPE '!'
		)`
		args = append(args, pattern, pattern, pattern, pattern)
	}

	var total int64
	if err := dbConn.QueryRow("SELECT COUNT(1) FROM super_venta_digital_ordenes "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `SELECT
		id,
		COALESCE(codigo_orden, ''),
		COALESCE(item_id, 0),
		COALESCE(item_nombre, ''),
		COALESCE(item_precio, 0),
		COALESCE(item_moneda, 'COP'),
		COALESCE(comprador_nombre, ''),
		COALESCE(comprador_email, ''),
		COALESCE(comprador_telefono, ''),
		COALESCE(metodo_pago, 'wompi_nequi'),
		COALESCE(estado_pago, 'pendiente'),
		COALESCE(transaction_id, ''),
		COALESCE(referencia_externa, ''),
		COALESCE(pasarela_payload_json, ''),
		COALESCE(pagado_en, ''),
		COALESCE(correo_entregado, 0),
		COALESCE(correo_entregado_en, ''),
		COALESCE(correo_entrega_error, ''),
		COALESCE(licencia_codigo_enviado, ''),
		COALESCE(instrucciones_archivo_url, ''),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM super_venta_digital_ordenes ` + where + `
	ORDER BY id DESC
	LIMIT ? OFFSET ?`

	rows, err := dbConn.Query(query, append(args, limit, offset)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := make([]SuperVentaDigitalOrder, 0)
	for rows.Next() {
		var item SuperVentaDigitalOrder
		var correoEntregado sql.NullInt64
		if err := rows.Scan(
			&item.ID,
			&item.CodigoOrden,
			&item.ItemID,
			&item.ItemNombre,
			&item.ItemPrecio,
			&item.ItemMoneda,
			&item.CompradorNombre,
			&item.CompradorEmail,
			&item.CompradorTelefono,
			&item.MetodoPago,
			&item.EstadoPago,
			&item.TransactionID,
			&item.ReferenciaExterna,
			&item.PasarelaPayloadJSON,
			&item.PagadoEn,
			&correoEntregado,
			&item.CorreoEntregadoEn,
			&item.CorreoEntregaError,
			&item.LicenciaCodigoEnviado,
			&item.InstruccionesArchivoURL,
			&item.FechaCreacion,
			&item.FechaActualizacion,
			&item.UsuarioCreador,
			&item.Estado,
			&item.Observaciones,
		); err != nil {
			return nil, 0, err
		}
		item.CorreoEntregado = correoEntregado.Valid && correoEntregado.Int64 > 0
		item.ItemMoneda = ventaDigitalNormalizeMoneda(item.ItemMoneda)
		item.Estado = ventaDigitalNormalizeEstado(item.Estado)
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

// BuildSuperVentaDigitalInstructionAbsoluteURL construye URL absoluta para archivo de instrucciones.
func BuildSuperVentaDigitalInstructionAbsoluteURL(baseURL, fileURL string) string {
	fileURL = strings.TrimSpace(fileURL)
	if fileURL == "" {
		return ""
	}
	if strings.HasPrefix(strings.ToLower(fileURL), "http://") || strings.HasPrefix(strings.ToLower(fileURL), "https://") {
		return fileURL
	}
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		return fileURL
	}
	if strings.HasPrefix(fileURL, "/") {
		return baseURL + fileURL
	}
	return baseURL + "/" + fileURL
}

// ParseSuperVentaDigitalOrderIDFromCode intenta extraer un ID numerico final desde un codigo de orden.
func ParseSuperVentaDigitalOrderIDFromCode(code string) int64 {
	code = strings.TrimSpace(code)
	if code == "" {
		return 0
	}
	parts := strings.Split(code, "-")
	if len(parts) == 0 {
		return 0
	}
	last := strings.TrimSpace(parts[len(parts)-1])
	if last == "" {
		return 0
	}
	v, err := strconv.ParseInt(last, 10, 64)
	if err != nil || v <= 0 {
		return 0
	}
	return v
}
