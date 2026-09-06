package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

const empresaVidaSchemaFingerprint = "empresa-vida:v1:personal-expenses-private-receipts-subscriptions-user-isolation"
const empresaVidaPriceHistorySchemaFingerprint = "empresa-vida:v2:personal-price-history-barcode-ai-invoices-user-isolation"
const empresaVidaReportsNotificationsSchemaFingerprint = "empresa-vida:v3:filtered-reports-user-opt-in-email-whatsapp-reminders-idempotent"

var ErrEmpresaVidaIdempotencyConflict = errors.New("vida idempotency key reused with different payload")

type EmpresaVidaGasto struct {
	ID                 int64   `json:"id"`
	EmpresaID          int64   `json:"empresa_id"`
	UsuarioID          string  `json:"-"`
	FechaGasto         string  `json:"fecha_gasto"`
	Categoria          string  `json:"categoria"`
	Comercio           string  `json:"comercio"`
	Descripcion        string  `json:"descripcion"`
	Monto              float64 `json:"monto"`
	Moneda             string  `json:"moneda"`
	MetodoPago         string  `json:"metodo_pago"`
	ReciboRef          string  `json:"-"`
	ReciboNombre       string  `json:"recibo_nombre,omitempty"`
	ReciboDisponible   bool    `json:"recibo_disponible"`
	ReciboURL          string  `json:"recibo_url,omitempty"`
	ClientRequestID    string  `json:"-"`
	RequestHash        string  `json:"-"`
	FechaCreacion      string  `json:"fecha_creacion"`
	FechaActualizacion string  `json:"fecha_actualizacion"`
}

type EmpresaVidaSuscripcion struct {
	ID                 int64   `json:"id"`
	EmpresaID          int64   `json:"empresa_id"`
	UsuarioID          string  `json:"-"`
	Nombre             string  `json:"nombre"`
	Proveedor          string  `json:"proveedor"`
	Costo              float64 `json:"costo"`
	Moneda             string  `json:"moneda"`
	Periodicidad       string  `json:"periodicidad"`
	Intervalo          int     `json:"intervalo"`
	FechaInicio        string  `json:"fecha_inicio"`
	ProximaRenovacion  string  `json:"proxima_renovacion"`
	RecordatorioDias   int     `json:"recordatorio_dias"`
	TipoRecordatorio   string  `json:"tipo_recordatorio"`
	AutoRenovacion     bool    `json:"auto_renovacion"`
	Estado             string  `json:"estado"`
	Notas              string  `json:"notas"`
	DiasRestantes      int     `json:"dias_restantes"`
	CostoMensual       float64 `json:"costo_mensual"`
	CostoAnual         float64 `json:"costo_anual"`
	ClientRequestID    string  `json:"-"`
	RequestHash        string  `json:"-"`
	FechaCreacion      string  `json:"fecha_creacion"`
	FechaActualizacion string  `json:"fecha_actualizacion"`
}

type EmpresaVidaCategoriaTotal struct {
	Categoria string  `json:"categoria"`
	Total     float64 `json:"total"`
}

type EmpresaVidaPrecio struct {
	ID             int64   `json:"id"`
	EmpresaID      int64   `json:"empresa_id"`
	UsuarioID      string  `json:"-"`
	GastoID        int64   `json:"gasto_id"`
	FechaCompra    string  `json:"fecha_compra"`
	CodigoBarras   string  `json:"codigo_barras"`
	ProductoNombre string  `json:"producto_nombre"`
	Comercio       string  `json:"comercio"`
	Cantidad       float64 `json:"cantidad"`
	PrecioUnitario float64 `json:"precio_unitario"`
	PrecioTotal    float64 `json:"precio_total"`
	Moneda         string  `json:"moneda"`
	Origen         string  `json:"origen"`
	FechaCreacion  string  `json:"fecha_creacion"`
}

type EmpresaVidaResumen struct {
	Mes                  string                      `json:"mes"`
	TotalMes             float64                     `json:"total_mes"`
	CantidadGastos       int64                       `json:"cantidad_gastos"`
	PromedioDiario       float64                     `json:"promedio_diario"`
	SuscripcionesActivas int64                       `json:"suscripciones_activas"`
	SuscripcionesMensual float64                     `json:"suscripciones_mensual"`
	SuscripcionesAnual   float64                     `json:"suscripciones_anual"`
	AlertasProximas      int64                       `json:"alertas_proximas"`
	PorCategoria         []EmpresaVidaCategoriaTotal `json:"por_categoria"`
}

// EmpresaVidaReporteFiltro is deliberately scoped by the caller; EmpresaID and
// UsuarioID never come from a browser-supplied filter.
type EmpresaVidaReporteFiltro struct {
	Desde      string
	Hasta      string
	Categoria  string
	Comercio   string
	MetodoPago string
}

type EmpresaVidaReporteFila struct {
	Clave    string  `json:"clave"`
	Total    float64 `json:"total"`
	Cantidad int64   `json:"cantidad"`
}

type EmpresaVidaReporte struct {
	Desde         string                   `json:"desde"`
	Hasta         string                   `json:"hasta"`
	Total         float64                  `json:"total"`
	Cantidad      int64                    `json:"cantidad"`
	Promedio      float64                  `json:"promedio"`
	PorCategoria  []EmpresaVidaReporteFila `json:"por_categoria"`
	PorComercio   []EmpresaVidaReporteFila `json:"por_comercio"`
	PorMetodoPago []EmpresaVidaReporteFila `json:"por_metodo_pago"`
	PorDia        []EmpresaVidaReporteFila `json:"por_dia"`
}

type EmpresaVidaNotificacionConfiguracion struct {
	EmpresaID        int64  `json:"empresa_id"`
	UsuarioID        string `json:"-"`
	EmailActiva      bool   `json:"email_activa"`
	WhatsAppActiva   bool   `json:"whatsapp_activa"`
	WhatsAppTelefono string `json:"whatsapp_telefono,omitempty"`
	HoraLocal        string `json:"hora_local"`
}

type EmpresaVidaRecordatorioPendiente struct {
	Suscripcion EmpresaVidaSuscripcion
	Config      EmpresaVidaNotificacionConfiguracion
}

func applyEmpresaVidaSchemaTx(_ context.Context, tx *sql.Tx) error {
	if tx == nil {
		return fmt.Errorf("migration transaction is required")
	}
	for _, statement := range empresaVidaSchemaStatements() {
		if _, err := tx.Exec(statement); err != nil {
			return err
		}
	}
	return nil
}

func empresaVidaSchemaStatements() []string {
	return []string{
		`CREATE TABLE IF NOT EXISTS empresa_vida_gastos (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL,
			usuario_id TEXT NOT NULL,
			fecha_gasto DATE NOT NULL,
			categoria TEXT NOT NULL,
			comercio TEXT NOT NULL DEFAULT '',
			descripcion TEXT NOT NULL DEFAULT '',
			monto NUMERIC(18,2) NOT NULL,
			moneda VARCHAR(3) NOT NULL DEFAULT 'COP',
			metodo_pago TEXT NOT NULL DEFAULT 'otro',
			recibo_ref TEXT NOT NULL DEFAULT '',
			recibo_nombre TEXT NOT NULL DEFAULT '',
			client_request_id TEXT NOT NULL DEFAULT '',
			request_hash TEXT NOT NULL DEFAULT '',
			fecha_creacion TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			fecha_actualizacion TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			CHECK (monto > 0),
			CHECK (char_length(usuario_id) BETWEEN 3 AND 320),
			CHECK (char_length(moneda) = 3)
		)`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_vida_gastos_usuario_fecha ON empresa_vida_gastos(empresa_id, usuario_id, fecha_gasto DESC, id DESC)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_empresa_vida_gastos_idempotencia ON empresa_vida_gastos(empresa_id, usuario_id, client_request_id) WHERE client_request_id <> ''`,
		`CREATE TABLE IF NOT EXISTS empresa_vida_suscripciones (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL,
			usuario_id TEXT NOT NULL,
			nombre TEXT NOT NULL,
			proveedor TEXT NOT NULL DEFAULT '',
			costo NUMERIC(18,2) NOT NULL,
			moneda VARCHAR(3) NOT NULL DEFAULT 'COP',
			periodicidad TEXT NOT NULL DEFAULT 'mensual',
			intervalo INTEGER NOT NULL DEFAULT 1,
			fecha_inicio DATE NOT NULL,
			proxima_renovacion DATE NOT NULL,
			recordatorio_dias INTEGER NOT NULL DEFAULT 5,
			tipo_recordatorio TEXT NOT NULL DEFAULT 'renovar',
			auto_renovacion BOOLEAN NOT NULL DEFAULT TRUE,
			estado TEXT NOT NULL DEFAULT 'activa',
			notas TEXT NOT NULL DEFAULT '',
			client_request_id TEXT NOT NULL DEFAULT '',
			request_hash TEXT NOT NULL DEFAULT '',
			fecha_creacion TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			fecha_actualizacion TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			CHECK (costo >= 0),
			CHECK (intervalo BETWEEN 1 AND 120),
			CHECK (recordatorio_dias BETWEEN 0 AND 365),
			CHECK (periodicidad IN ('semanal','mensual','trimestral','semestral','anual','personalizada')),
			CHECK (tipo_recordatorio IN ('renovar','cancelar','ambos')),
			CHECK (estado IN ('activa','pausada','cancelada','vencida'))
		)`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_vida_suscripciones_usuario_fecha ON empresa_vida_suscripciones(empresa_id, usuario_id, estado, proxima_renovacion, id)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_empresa_vida_suscripciones_idempotencia ON empresa_vida_suscripciones(empresa_id, usuario_id, client_request_id) WHERE client_request_id <> ''`,
	}
}

func applyEmpresaVidaPriceHistorySchemaTx(_ context.Context, tx *sql.Tx) error {
	if tx == nil {
		return fmt.Errorf("migration transaction is required")
	}
	for _, statement := range empresaVidaPriceHistorySchemaStatements() {
		if _, err := tx.Exec(statement); err != nil {
			return err
		}
	}
	return nil
}

func empresaVidaPriceHistorySchemaStatements() []string {
	return []string{
		`CREATE TABLE IF NOT EXISTS empresa_vida_precios (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL,
			usuario_id TEXT NOT NULL,
			gasto_id BIGINT NOT NULL REFERENCES empresa_vida_gastos(id) ON DELETE CASCADE,
			fecha_compra DATE NOT NULL,
			codigo_barras TEXT NOT NULL DEFAULT '',
			producto_nombre TEXT NOT NULL,
			comercio TEXT NOT NULL DEFAULT '',
			cantidad NUMERIC(14,3) NOT NULL DEFAULT 1,
			precio_unitario NUMERIC(18,2) NOT NULL,
			precio_total NUMERIC(18,2) NOT NULL,
			moneda VARCHAR(3) NOT NULL DEFAULT 'COP',
			origen TEXT NOT NULL DEFAULT 'manual',
			fecha_creacion TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			CHECK (cantidad > 0),
			CHECK (precio_unitario >= 0),
			CHECK (precio_total >= 0),
			CHECK (char_length(usuario_id) BETWEEN 3 AND 320),
			CHECK (char_length(moneda) = 3),
			CHECK (origen IN ('manual','codigo_barras','ia_factura'))
		)`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_vida_precios_usuario_fecha ON empresa_vida_precios(empresa_id, usuario_id, fecha_compra DESC, id DESC)`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_vida_precios_usuario_codigo ON empresa_vida_precios(empresa_id, usuario_id, codigo_barras, fecha_compra DESC, id DESC) WHERE codigo_barras <> ''`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_vida_precios_usuario_producto ON empresa_vida_precios(empresa_id, usuario_id, LOWER(producto_nombre), fecha_compra DESC, id DESC)`,
	}
}

func applyEmpresaVidaReportsNotificationsSchemaTx(_ context.Context, tx *sql.Tx) error {
	if tx == nil {
		return fmt.Errorf("migration transaction is required")
	}
	for _, statement := range empresaVidaReportsNotificationsSchemaStatements() {
		if _, err := tx.Exec(statement); err != nil {
			return err
		}
	}
	return nil
}

func empresaVidaReportsNotificationsSchemaStatements() []string {
	return []string{
		`CREATE TABLE IF NOT EXISTS empresa_vida_notificacion_configuracion (
			empresa_id BIGINT NOT NULL,
			usuario_id TEXT NOT NULL,
			email_activa BOOLEAN NOT NULL DEFAULT FALSE,
			whatsapp_activa BOOLEAN NOT NULL DEFAULT FALSE,
			whatsapp_telefono TEXT NOT NULL DEFAULT '',
			hora_local VARCHAR(5) NOT NULL DEFAULT '09:00',
			fecha_actualizacion TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (empresa_id, usuario_id),
			CHECK (char_length(usuario_id) BETWEEN 3 AND 320),
			CHECK (hora_local ~ '^[0-2][0-9]:[0-5][0-9]$')
		)`,
		`CREATE TABLE IF NOT EXISTS empresa_vida_notificaciones (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL,
			usuario_id TEXT NOT NULL,
			suscripcion_id BIGINT NOT NULL REFERENCES empresa_vida_suscripciones(id) ON DELETE CASCADE,
			proxima_renovacion DATE NOT NULL,
			canal TEXT NOT NULL,
			estado TEXT NOT NULL DEFAULT 'procesando',
			intentos INTEGER NOT NULL DEFAULT 1,
			fecha_ultimo_intento TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			fecha_envio TIMESTAMPTZ,
			error_publico TEXT NOT NULL DEFAULT '',
			fecha_creacion TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			CHECK (canal IN ('email','whatsapp')),
			CHECK (estado IN ('procesando','enviado','error','omitido')),
			CHECK (intentos BETWEEN 1 AND 5),
			UNIQUE (empresa_id, usuario_id, suscripcion_id, proxima_renovacion, canal)
		)`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_vida_notificaciones_estado ON empresa_vida_notificaciones(estado, fecha_ultimo_intento)`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_vida_notificaciones_usuario ON empresa_vida_notificaciones(empresa_id, usuario_id, fecha_creacion DESC)`,
	}
}

func VerifyEmpresaVidaSchema(dbConn *sql.DB) error {
	if dbConn == nil {
		return fmt.Errorf("database not available")
	}
	for _, table := range []string{"empresa_vida_gastos", "empresa_vida_suscripciones", "empresa_vida_precios", "empresa_vida_notificacion_configuracion", "empresa_vida_notificaciones"} {
		var name sql.NullString
		if err := QueryRowCompat(dbConn, `SELECT to_regclass(?)`, table).Scan(&name); err != nil {
			return err
		}
		if !name.Valid || strings.TrimSpace(name.String) == "" {
			return fmt.Errorf("%s schema missing; run pcs-migrate", table)
		}
	}
	return nil
}

func CreateEmpresaVidaGastoConPrecios(dbConn *sql.DB, item EmpresaVidaGasto, precios []EmpresaVidaPrecio) (*EmpresaVidaGasto, []EmpresaVidaPrecio, bool, error) {
	if dbConn == nil {
		return nil, nil, false, fmt.Errorf("database not available")
	}
	tx, err := dbConn.BeginTx(context.Background(), nil)
	if err != nil {
		return nil, nil, false, err
	}
	defer func() { _ = tx.Rollback() }()
	result, err := execTxSQLCompat(tx, `INSERT INTO empresa_vida_gastos
		(empresa_id, usuario_id, fecha_gasto, categoria, comercio, descripcion, monto, moneda, metodo_pago, recibo_ref, recibo_nombre, client_request_id, request_hash)
		VALUES (?, ?, CAST(? AS DATE), ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ON CONFLICT DO NOTHING`,
		item.EmpresaID, normalizeVidaUsuario(item.UsuarioID), item.FechaGasto, item.Categoria, item.Comercio, item.Descripcion,
		item.Monto, strings.ToUpper(item.Moneda), item.MetodoPago, item.ReciboRef, item.ReciboNombre, item.ClientRequestID, item.RequestHash)
	if err != nil {
		return nil, nil, false, err
	}
	createdRows, _ := result.RowsAffected()
	stored, err := scanEmpresaVidaGasto(queryRowTxSQLCompat(tx, vidaGastoSelect+` WHERE empresa_id=? AND usuario_id=? AND client_request_id=? LIMIT 1`, item.EmpresaID, normalizeVidaUsuario(item.UsuarioID), strings.TrimSpace(item.ClientRequestID)))
	if err != nil {
		return nil, nil, false, err
	}
	if createdRows == 0 {
		if stored.RequestHash != item.RequestHash {
			return nil, nil, false, ErrEmpresaVidaIdempotencyConflict
		}
		existing, err := listEmpresaVidaPreciosTx(tx, item.EmpresaID, item.UsuarioID, stored.ID)
		if err != nil {
			return nil, nil, false, err
		}
		if err := tx.Commit(); err != nil {
			return nil, nil, false, err
		}
		return stored, existing, false, nil
	}
	inserted := make([]EmpresaVidaPrecio, 0, len(precios))
	for _, precio := range precios {
		precio.EmpresaID, precio.UsuarioID, precio.GastoID = item.EmpresaID, normalizeVidaUsuario(item.UsuarioID), stored.ID
		id, err := insertTxSQLCompat(tx, `INSERT INTO empresa_vida_precios
			(empresa_id, usuario_id, gasto_id, fecha_compra, codigo_barras, producto_nombre, comercio, cantidad, precio_unitario, precio_total, moneda, origen)
			VALUES (?, ?, ?, CAST(? AS DATE), ?, ?, ?, ?, ?, ?, ?, ?)`, precio.EmpresaID, precio.UsuarioID, precio.GastoID,
			precio.FechaCompra, precio.CodigoBarras, precio.ProductoNombre, precio.Comercio, precio.Cantidad, precio.PrecioUnitario,
			precio.PrecioTotal, strings.ToUpper(precio.Moneda), precio.Origen)
		if err != nil {
			return nil, nil, false, err
		}
		precio.ID = id
		inserted = append(inserted, precio)
	}
	if err := tx.Commit(); err != nil {
		return nil, nil, false, err
	}
	return stored, inserted, true, nil
}

func ListEmpresaVidaPrecios(dbConn *sql.DB, empresaID int64, usuarioID, codigo, producto string, limit int) ([]EmpresaVidaPrecio, error) {
	if limit < 1 || limit > 500 {
		limit = 200
	}
	query := vidaPrecioSelect + ` WHERE empresa_id=? AND usuario_id=?`
	args := []interface{}{empresaID, normalizeVidaUsuario(usuarioID)}
	if strings.TrimSpace(codigo) != "" {
		query += ` AND codigo_barras=?`
		args = append(args, strings.TrimSpace(codigo))
	}
	if strings.TrimSpace(producto) != "" {
		query += ` AND LOWER(producto_nombre) LIKE LOWER(?)`
		args = append(args, "%"+strings.TrimSpace(producto)+"%")
	}
	query += ` ORDER BY fecha_compra DESC, id DESC LIMIT ?`
	args = append(args, limit)
	rows, err := ExecQueryCompat(dbConn, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanEmpresaVidaPrecios(rows)
}

func ListEmpresaVidaPreciosPorGasto(dbConn *sql.DB, empresaID int64, usuarioID string, gastoID int64) ([]EmpresaVidaPrecio, error) {
	rows, err := ExecQueryCompat(dbConn, vidaPrecioSelect+` WHERE empresa_id=? AND usuario_id=? AND gasto_id=? ORDER BY id`, empresaID, normalizeVidaUsuario(usuarioID), gastoID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanEmpresaVidaPrecios(rows)
}

const vidaPrecioSelect = `SELECT id, empresa_id, usuario_id, gasto_id, TO_CHAR(fecha_compra,'YYYY-MM-DD'), COALESCE(codigo_barras,''), producto_nombre, COALESCE(comercio,''), COALESCE(cantidad,1), COALESCE(precio_unitario,0), COALESCE(precio_total,0), COALESCE(moneda,'COP'), COALESCE(origen,'manual'), CAST(fecha_creacion AS TEXT) FROM empresa_vida_precios`

func listEmpresaVidaPreciosTx(tx *sql.Tx, empresaID int64, usuarioID string, gastoID int64) ([]EmpresaVidaPrecio, error) {
	rows, err := queryTxSQLCompat(tx, vidaPrecioSelect+` WHERE empresa_id=? AND usuario_id=? AND gasto_id=? ORDER BY id`, empresaID, normalizeVidaUsuario(usuarioID), gastoID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanEmpresaVidaPrecios(rows)
}

func scanEmpresaVidaPrecios(rows *sql.Rows) ([]EmpresaVidaPrecio, error) {
	out := make([]EmpresaVidaPrecio, 0)
	for rows.Next() {
		var item EmpresaVidaPrecio
		if err := rows.Scan(&item.ID, &item.EmpresaID, &item.UsuarioID, &item.GastoID, &item.FechaCompra, &item.CodigoBarras, &item.ProductoNombre, &item.Comercio, &item.Cantidad, &item.PrecioUnitario, &item.PrecioTotal, &item.Moneda, &item.Origen, &item.FechaCreacion); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func CreateEmpresaVidaGasto(dbConn *sql.DB, item EmpresaVidaGasto) (*EmpresaVidaGasto, bool, error) {
	result, err := ExecCompat(dbConn, `INSERT INTO empresa_vida_gastos
		(empresa_id, usuario_id, fecha_gasto, categoria, comercio, descripcion, monto, moneda, metodo_pago, recibo_ref, recibo_nombre, client_request_id, request_hash)
		VALUES (?, ?, CAST(? AS DATE), ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ON CONFLICT DO NOTHING`,
		item.EmpresaID, normalizeVidaUsuario(item.UsuarioID), item.FechaGasto, item.Categoria, item.Comercio, item.Descripcion,
		item.Monto, strings.ToUpper(item.Moneda), item.MetodoPago, item.ReciboRef, item.ReciboNombre, item.ClientRequestID, item.RequestHash)
	if err != nil {
		return nil, false, err
	}
	created, _ := result.RowsAffected()
	stored, err := getEmpresaVidaGastoByRequest(dbConn, item.EmpresaID, item.UsuarioID, item.ClientRequestID)
	if err != nil {
		return nil, false, err
	}
	if created == 0 && stored.RequestHash != item.RequestHash {
		return nil, false, ErrEmpresaVidaIdempotencyConflict
	}
	return stored, created == 1, nil
}

func getEmpresaVidaGastoByRequest(dbConn *sql.DB, empresaID int64, usuarioID, requestID string) (*EmpresaVidaGasto, error) {
	return scanEmpresaVidaGasto(QueryRowCompat(dbConn, vidaGastoSelect+` WHERE empresa_id=? AND usuario_id=? AND client_request_id=? LIMIT 1`, empresaID, normalizeVidaUsuario(usuarioID), strings.TrimSpace(requestID)))
}

func GetEmpresaVidaGastoByRequest(dbConn *sql.DB, empresaID int64, usuarioID, requestID string) (*EmpresaVidaGasto, error) {
	return getEmpresaVidaGastoByRequest(dbConn, empresaID, usuarioID, requestID)
}

func GetEmpresaVidaGasto(dbConn *sql.DB, empresaID, id int64, usuarioID string) (*EmpresaVidaGasto, error) {
	return scanEmpresaVidaGasto(QueryRowCompat(dbConn, vidaGastoSelect+` WHERE empresa_id=? AND usuario_id=? AND id=? LIMIT 1`, empresaID, normalizeVidaUsuario(usuarioID), id))
}

func ListEmpresaVidaGastos(dbConn *sql.DB, empresaID int64, usuarioID, desde, hasta, categoria string, limit int) ([]EmpresaVidaGasto, error) {
	if limit < 1 || limit > 500 {
		limit = 100
	}
	query := vidaGastoSelect + ` WHERE empresa_id=? AND usuario_id=?`
	args := []interface{}{empresaID, normalizeVidaUsuario(usuarioID)}
	if strings.TrimSpace(desde) != "" {
		query += ` AND fecha_gasto >= CAST(? AS DATE)`
		args = append(args, desde)
	}
	if strings.TrimSpace(hasta) != "" {
		query += ` AND fecha_gasto <= CAST(? AS DATE)`
		args = append(args, hasta)
	}
	if strings.TrimSpace(categoria) != "" {
		query += ` AND categoria=?`
		args = append(args, strings.TrimSpace(categoria))
	}
	query += ` ORDER BY fecha_gasto DESC, id DESC LIMIT ?`
	args = append(args, limit)
	rows, err := ExecQueryCompat(dbConn, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]EmpresaVidaGasto, 0)
	for rows.Next() {
		item, err := scanEmpresaVidaGasto(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *item)
	}
	return out, rows.Err()
}

func UpdateEmpresaVidaGasto(dbConn *sql.DB, item EmpresaVidaGasto) error {
	result, err := ExecCompat(dbConn, `UPDATE empresa_vida_gastos SET fecha_gasto=CAST(? AS DATE), categoria=?, comercio=?, descripcion=?, monto=?, moneda=?, metodo_pago=?, fecha_actualizacion=CURRENT_TIMESTAMP WHERE empresa_id=? AND usuario_id=? AND id=?`,
		item.FechaGasto, item.Categoria, item.Comercio, item.Descripcion, item.Monto, strings.ToUpper(item.Moneda), item.MetodoPago, item.EmpresaID, normalizeVidaUsuario(item.UsuarioID), item.ID)
	if err != nil {
		return err
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return sql.ErrNoRows
	}
	return nil
}

func DeleteEmpresaVidaGasto(dbConn *sql.DB, empresaID, id int64, usuarioID string) (*EmpresaVidaGasto, error) {
	item, err := GetEmpresaVidaGasto(dbConn, empresaID, id, usuarioID)
	if err != nil {
		return nil, err
	}
	result, err := ExecCompat(dbConn, `DELETE FROM empresa_vida_gastos WHERE empresa_id=? AND usuario_id=? AND id=?`, empresaID, normalizeVidaUsuario(usuarioID), id)
	if err != nil {
		return nil, err
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return nil, sql.ErrNoRows
	}
	return item, nil
}

const vidaGastoSelect = `SELECT id, empresa_id, usuario_id, TO_CHAR(fecha_gasto,'YYYY-MM-DD'), categoria, COALESCE(comercio,''), COALESCE(descripcion,''), COALESCE(monto,0), COALESCE(moneda,'COP'), COALESCE(metodo_pago,'otro'), COALESCE(recibo_ref,''), COALESCE(recibo_nombre,''), COALESCE(client_request_id,''), COALESCE(request_hash,''), CAST(fecha_creacion AS TEXT), CAST(fecha_actualizacion AS TEXT) FROM empresa_vida_gastos`

type vidaScanner interface{ Scan(...interface{}) error }

func scanEmpresaVidaGasto(scanner vidaScanner) (*EmpresaVidaGasto, error) {
	var item EmpresaVidaGasto
	if err := scanner.Scan(&item.ID, &item.EmpresaID, &item.UsuarioID, &item.FechaGasto, &item.Categoria, &item.Comercio, &item.Descripcion, &item.Monto, &item.Moneda, &item.MetodoPago, &item.ReciboRef, &item.ReciboNombre, &item.ClientRequestID, &item.RequestHash, &item.FechaCreacion, &item.FechaActualizacion); err != nil {
		return nil, err
	}
	item.ReciboDisponible = strings.TrimSpace(item.ReciboRef) != ""
	return &item, nil
}

func CreateEmpresaVidaSuscripcion(dbConn *sql.DB, item EmpresaVidaSuscripcion) (*EmpresaVidaSuscripcion, bool, error) {
	result, err := ExecCompat(dbConn, `INSERT INTO empresa_vida_suscripciones
		(empresa_id, usuario_id, nombre, proveedor, costo, moneda, periodicidad, intervalo, fecha_inicio, proxima_renovacion, recordatorio_dias, tipo_recordatorio, auto_renovacion, estado, notas, client_request_id, request_hash)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, CAST(? AS DATE), CAST(? AS DATE), ?, ?, ?, ?, ?, ?, ?) ON CONFLICT DO NOTHING`,
		item.EmpresaID, normalizeVidaUsuario(item.UsuarioID), item.Nombre, item.Proveedor, item.Costo, strings.ToUpper(item.Moneda), item.Periodicidad,
		item.Intervalo, item.FechaInicio, item.ProximaRenovacion, item.RecordatorioDias, item.TipoRecordatorio, item.AutoRenovacion,
		item.Estado, item.Notas, item.ClientRequestID, item.RequestHash)
	if err != nil {
		return nil, false, err
	}
	created, _ := result.RowsAffected()
	stored, err := getEmpresaVidaSuscripcionByRequest(dbConn, item.EmpresaID, item.UsuarioID, item.ClientRequestID)
	if err != nil {
		return nil, false, err
	}
	if created == 0 && stored.RequestHash != item.RequestHash {
		return nil, false, ErrEmpresaVidaIdempotencyConflict
	}
	return stored, created == 1, nil
}

func getEmpresaVidaSuscripcionByRequest(dbConn *sql.DB, empresaID int64, usuarioID, requestID string) (*EmpresaVidaSuscripcion, error) {
	return scanEmpresaVidaSuscripcion(QueryRowCompat(dbConn, vidaSuscripcionSelect+` WHERE empresa_id=? AND usuario_id=? AND client_request_id=? LIMIT 1`, empresaID, normalizeVidaUsuario(usuarioID), strings.TrimSpace(requestID)))
}

func GetEmpresaVidaSuscripcion(dbConn *sql.DB, empresaID, id int64, usuarioID string) (*EmpresaVidaSuscripcion, error) {
	return scanEmpresaVidaSuscripcion(QueryRowCompat(dbConn, vidaSuscripcionSelect+` WHERE empresa_id=? AND usuario_id=? AND id=? LIMIT 1`, empresaID, normalizeVidaUsuario(usuarioID), id))
}

func ListEmpresaVidaSuscripciones(dbConn *sql.DB, empresaID int64, usuarioID, estado string) ([]EmpresaVidaSuscripcion, error) {
	query := vidaSuscripcionSelect + ` WHERE empresa_id=? AND usuario_id=?`
	args := []interface{}{empresaID, normalizeVidaUsuario(usuarioID)}
	if strings.TrimSpace(estado) != "" {
		query += ` AND estado=?`
		args = append(args, strings.TrimSpace(estado))
	}
	query += ` ORDER BY CASE WHEN estado='activa' THEN 0 ELSE 1 END, proxima_renovacion, id DESC`
	rows, err := ExecQueryCompat(dbConn, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]EmpresaVidaSuscripcion, 0)
	for rows.Next() {
		item, err := scanEmpresaVidaSuscripcion(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *item)
	}
	return out, rows.Err()
}

func UpdateEmpresaVidaSuscripcion(dbConn *sql.DB, item EmpresaVidaSuscripcion) error {
	result, err := ExecCompat(dbConn, `UPDATE empresa_vida_suscripciones SET nombre=?, proveedor=?, costo=?, moneda=?, periodicidad=?, intervalo=?, fecha_inicio=CAST(? AS DATE), proxima_renovacion=CAST(? AS DATE), recordatorio_dias=?, tipo_recordatorio=?, auto_renovacion=?, estado=?, notas=?, fecha_actualizacion=CURRENT_TIMESTAMP WHERE empresa_id=? AND usuario_id=? AND id=?`,
		item.Nombre, item.Proveedor, item.Costo, strings.ToUpper(item.Moneda), item.Periodicidad, item.Intervalo, item.FechaInicio, item.ProximaRenovacion,
		item.RecordatorioDias, item.TipoRecordatorio, item.AutoRenovacion, item.Estado, item.Notas, item.EmpresaID, normalizeVidaUsuario(item.UsuarioID), item.ID)
	if err != nil {
		return err
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return sql.ErrNoRows
	}
	return nil
}

func DeleteEmpresaVidaSuscripcion(dbConn *sql.DB, empresaID, id int64, usuarioID string) error {
	result, err := ExecCompat(dbConn, `DELETE FROM empresa_vida_suscripciones WHERE empresa_id=? AND usuario_id=? AND id=?`, empresaID, normalizeVidaUsuario(usuarioID), id)
	if err != nil {
		return err
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return sql.ErrNoRows
	}
	return nil
}

func RenovarEmpresaVidaSuscripcion(dbConn *sql.DB, empresaID, id int64, usuarioID, proximaFecha string) error {
	result, err := ExecCompat(dbConn, `UPDATE empresa_vida_suscripciones SET proxima_renovacion=CAST(? AS DATE), estado='activa', fecha_actualizacion=CURRENT_TIMESTAMP WHERE empresa_id=? AND usuario_id=? AND id=?`, proximaFecha, empresaID, normalizeVidaUsuario(usuarioID), id)
	if err != nil {
		return err
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return sql.ErrNoRows
	}
	return nil
}

const vidaSuscripcionSelect = `SELECT id, empresa_id, usuario_id, nombre, COALESCE(proveedor,''), COALESCE(costo,0), COALESCE(moneda,'COP'), periodicidad, intervalo, TO_CHAR(fecha_inicio,'YYYY-MM-DD'), TO_CHAR(proxima_renovacion,'YYYY-MM-DD'), recordatorio_dias, tipo_recordatorio, auto_renovacion, estado, COALESCE(notas,''), (proxima_renovacion-CURRENT_DATE), COALESCE(client_request_id,''), COALESCE(request_hash,''), CAST(fecha_creacion AS TEXT), CAST(fecha_actualizacion AS TEXT) FROM empresa_vida_suscripciones`

func scanEmpresaVidaSuscripcion(scanner vidaScanner) (*EmpresaVidaSuscripcion, error) {
	var item EmpresaVidaSuscripcion
	if err := scanner.Scan(&item.ID, &item.EmpresaID, &item.UsuarioID, &item.Nombre, &item.Proveedor, &item.Costo, &item.Moneda, &item.Periodicidad, &item.Intervalo, &item.FechaInicio, &item.ProximaRenovacion, &item.RecordatorioDias, &item.TipoRecordatorio, &item.AutoRenovacion, &item.Estado, &item.Notas, &item.DiasRestantes, &item.ClientRequestID, &item.RequestHash, &item.FechaCreacion, &item.FechaActualizacion); err != nil {
		return nil, err
	}
	item.CostoMensual, item.CostoAnual = EmpresaVidaSubscriptionProjection(item.Costo, item.Periodicidad, item.Intervalo)
	return &item, nil
}

func GetEmpresaVidaResumen(dbConn *sql.DB, empresaID int64, usuarioID, mes string) (*EmpresaVidaResumen, error) {
	if strings.TrimSpace(mes) == "" {
		mes = time.Now().Format("2006-01")
	}
	start, err := time.Parse("2006-01", mes)
	if err != nil {
		return nil, fmt.Errorf("mes invalido")
	}
	end := start.AddDate(0, 1, 0)
	out := &EmpresaVidaResumen{Mes: mes, PorCategoria: []EmpresaVidaCategoriaTotal{}}
	err = QueryRowCompat(dbConn, `SELECT COALESCE(SUM(monto),0), COUNT(*), COALESCE(SUM(monto)/NULLIF(EXTRACT(DAY FROM LEAST(CURRENT_DATE, CAST(? AS DATE)))::INTEGER,0),0) FROM empresa_vida_gastos WHERE empresa_id=? AND usuario_id=? AND fecha_gasto>=CAST(? AS DATE) AND fecha_gasto<CAST(? AS DATE)`, end.AddDate(0, 0, -1).Format("2006-01-02"), empresaID, normalizeVidaUsuario(usuarioID), start.Format("2006-01-02"), end.Format("2006-01-02")).Scan(&out.TotalMes, &out.CantidadGastos, &out.PromedioDiario)
	if err != nil {
		return nil, err
	}
	rows, err := ExecQueryCompat(dbConn, `SELECT categoria, COALESCE(SUM(monto),0) FROM empresa_vida_gastos WHERE empresa_id=? AND usuario_id=? AND fecha_gasto>=CAST(? AS DATE) AND fecha_gasto<CAST(? AS DATE) GROUP BY categoria ORDER BY SUM(monto) DESC, categoria`, empresaID, normalizeVidaUsuario(usuarioID), start.Format("2006-01-02"), end.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var row EmpresaVidaCategoriaTotal
		if err := rows.Scan(&row.Categoria, &row.Total); err != nil {
			_ = rows.Close()
			return nil, err
		}
		out.PorCategoria = append(out.PorCategoria, row)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	subs, err := ListEmpresaVidaSuscripciones(dbConn, empresaID, usuarioID, "activa")
	if err != nil {
		return nil, err
	}
	for _, sub := range subs {
		out.SuscripcionesActivas++
		out.SuscripcionesMensual += sub.CostoMensual
		out.SuscripcionesAnual += sub.CostoAnual
		if sub.DiasRestantes <= sub.RecordatorioDias {
			out.AlertasProximas++
		}
	}
	return out, nil
}

func EmpresaVidaSubscriptionProjection(costo float64, periodicidad string, intervalo int) (float64, float64) {
	if intervalo < 1 {
		intervalo = 1
	}
	var cyclesPerYear float64
	switch strings.ToLower(strings.TrimSpace(periodicidad)) {
	case "semanal":
		cyclesPerYear = 52 / float64(intervalo)
	case "trimestral":
		cyclesPerYear = 4 / float64(intervalo)
	case "semestral":
		cyclesPerYear = 2 / float64(intervalo)
	case "anual":
		cyclesPerYear = 1 / float64(intervalo)
	default:
		cyclesPerYear = 12 / float64(intervalo)
	}
	anual := costo * cyclesPerYear
	return anual / 12, anual
}

func GetEmpresaVidaReporte(dbConn *sql.DB, empresaID int64, usuarioID string, filter EmpresaVidaReporteFiltro) (*EmpresaVidaReporte, error) {
	if empresaID <= 0 || normalizeVidaUsuario(usuarioID) == "" {
		return nil, fmt.Errorf("contexto de Vida invalido")
	}
	desde, hasta, err := normalizeVidaReportDates(filter.Desde, filter.Hasta)
	if err != nil {
		return nil, err
	}
	where, args := vidaReportWhere(empresaID, usuarioID, desde, hasta, filter)
	out := &EmpresaVidaReporte{Desde: desde, Hasta: hasta, PorCategoria: []EmpresaVidaReporteFila{}, PorComercio: []EmpresaVidaReporteFila{}, PorMetodoPago: []EmpresaVidaReporteFila{}, PorDia: []EmpresaVidaReporteFila{}}
	if err := QueryRowCompat(dbConn, `SELECT COALESCE(SUM(monto),0), COUNT(*), COALESCE(AVG(monto),0) FROM empresa_vida_gastos WHERE `+where, args...).Scan(&out.Total, &out.Cantidad, &out.Promedio); err != nil {
		return nil, err
	}
	for _, group := range []struct {
		column string
		target *[]EmpresaVidaReporteFila
	}{
		{"categoria", &out.PorCategoria}, {"COALESCE(NULLIF(comercio,''),'Sin comercio')", &out.PorComercio}, {"metodo_pago", &out.PorMetodoPago}, {"TO_CHAR(fecha_gasto,'YYYY-MM-DD')", &out.PorDia},
	} {
		rows, queryErr := ExecQueryCompat(dbConn, `SELECT `+group.column+`, COALESCE(SUM(monto),0), COUNT(*) FROM empresa_vida_gastos WHERE `+where+` GROUP BY `+group.column+` ORDER BY `+group.column, args...)
		if queryErr != nil {
			return nil, queryErr
		}
		for rows.Next() {
			var row EmpresaVidaReporteFila
			if err := rows.Scan(&row.Clave, &row.Total, &row.Cantidad); err != nil {
				_ = rows.Close()
				return nil, err
			}
			*group.target = append(*group.target, row)
		}
		if err := rows.Close(); err != nil {
			return nil, err
		}
	}
	return out, nil
}

func normalizeVidaReportDates(rawDesde, rawHasta string) (string, string, error) {
	today := time.Now().Format("2006-01-02")
	desde, hasta := strings.TrimSpace(rawDesde), strings.TrimSpace(rawHasta)
	if hasta == "" {
		hasta = today
	}
	if desde == "" {
		desde = time.Now().AddDate(0, -1, 0).Format("2006-01-02")
	}
	start, err := time.Parse("2006-01-02", desde)
	if err != nil {
		return "", "", fmt.Errorf("fecha desde invalida")
	}
	end, err := time.Parse("2006-01-02", hasta)
	if err != nil {
		return "", "", fmt.Errorf("fecha hasta invalida")
	}
	if start.After(end) || end.Sub(start).Hours() > 366*24 {
		return "", "", fmt.Errorf("el periodo debe estar entre 1 y 366 dias")
	}
	return desde, hasta, nil
}

func vidaReportWhere(empresaID int64, usuarioID, desde, hasta string, filter EmpresaVidaReporteFiltro) (string, []interface{}) {
	clauses := []string{"empresa_id=?", "usuario_id=?", "fecha_gasto>=CAST(? AS DATE)", "fecha_gasto<=CAST(? AS DATE)"}
	args := []interface{}{empresaID, normalizeVidaUsuario(usuarioID), desde, hasta}
	if value := strings.TrimSpace(filter.Categoria); value != "" {
		clauses, args = append(clauses, "categoria=?"), append(args, value)
	}
	if value := strings.TrimSpace(filter.Comercio); value != "" {
		clauses, args = append(clauses, "LOWER(comercio) LIKE LOWER(?)"), append(args, "%"+value+"%")
	}
	if value := strings.TrimSpace(filter.MetodoPago); value != "" {
		clauses, args = append(clauses, "metodo_pago=?"), append(args, value)
	}
	return strings.Join(clauses, " AND "), args
}

func GetEmpresaVidaNotificacionConfiguracion(dbConn *sql.DB, empresaID int64, usuarioID string) (*EmpresaVidaNotificacionConfiguracion, error) {
	item := &EmpresaVidaNotificacionConfiguracion{EmpresaID: empresaID, UsuarioID: normalizeVidaUsuario(usuarioID), HoraLocal: "09:00"}
	err := QueryRowCompat(dbConn, `SELECT email_activa, whatsapp_activa, COALESCE(whatsapp_telefono,''), hora_local FROM empresa_vida_notificacion_configuracion WHERE empresa_id=? AND usuario_id=?`, empresaID, item.UsuarioID).Scan(&item.EmailActiva, &item.WhatsAppActiva, &item.WhatsAppTelefono, &item.HoraLocal)
	if errors.Is(err, sql.ErrNoRows) {
		return item, nil
	}
	if err != nil {
		return nil, err
	}
	return item, nil
}

func SaveEmpresaVidaNotificacionConfiguracion(dbConn *sql.DB, item EmpresaVidaNotificacionConfiguracion) error {
	if item.EmpresaID <= 0 || normalizeVidaUsuario(item.UsuarioID) == "" {
		return fmt.Errorf("contexto de Vida invalido")
	}
	if _, err := time.Parse("15:04", strings.TrimSpace(item.HoraLocal)); err != nil {
		return fmt.Errorf("hora local invalida")
	}
	_, err := ExecCompat(dbConn, `INSERT INTO empresa_vida_notificacion_configuracion (empresa_id, usuario_id, email_activa, whatsapp_activa, whatsapp_telefono, hora_local) VALUES (?, ?, ?, ?, ?, ?) ON CONFLICT (empresa_id, usuario_id) DO UPDATE SET email_activa=EXCLUDED.email_activa, whatsapp_activa=EXCLUDED.whatsapp_activa, whatsapp_telefono=EXCLUDED.whatsapp_telefono, hora_local=EXCLUDED.hora_local, fecha_actualizacion=CURRENT_TIMESTAMP`, item.EmpresaID, normalizeVidaUsuario(item.UsuarioID), item.EmailActiva, item.WhatsAppActiva, strings.TrimSpace(item.WhatsAppTelefono), strings.TrimSpace(item.HoraLocal))
	return err
}

func ListEmpresaVidaRecordatoriosPendientes(dbConn *sql.DB, now time.Time, limit int) ([]EmpresaVidaRecordatorioPendiente, error) {
	if limit < 1 || limit > 500 {
		limit = 200
	}
	_ = now // PostgreSQL evaluates CURRENT_DATE consistently for every candidate in this cycle.
	rows, err := ExecQueryCompat(dbConn, `SELECT s.id, s.empresa_id, s.usuario_id, s.nombre, COALESCE(s.proveedor,''), COALESCE(s.costo,0), COALESCE(s.moneda,'COP'), s.periodicidad, s.intervalo, TO_CHAR(s.fecha_inicio,'YYYY-MM-DD'), TO_CHAR(s.proxima_renovacion,'YYYY-MM-DD'), s.recordatorio_dias, s.tipo_recordatorio, s.auto_renovacion, s.estado, COALESCE(s.notas,''), (s.proxima_renovacion-CURRENT_DATE), COALESCE(s.client_request_id,''), COALESCE(s.request_hash,''), CAST(s.fecha_creacion AS TEXT), CAST(s.fecha_actualizacion AS TEXT), c.email_activa, c.whatsapp_activa, COALESCE(c.whatsapp_telefono,''), c.hora_local FROM empresa_vida_suscripciones s JOIN empresa_vida_notificacion_configuracion c ON c.empresa_id=s.empresa_id AND c.usuario_id=s.usuario_id WHERE s.estado='activa' AND (c.email_activa=TRUE OR c.whatsapp_activa=TRUE) AND CURRENT_DATE >= (s.proxima_renovacion - s.recordatorio_dias) ORDER BY s.proxima_renovacion, s.id LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]EmpresaVidaRecordatorioPendiente, 0)
	for rows.Next() {
		var sub EmpresaVidaSuscripcion
		var cfg EmpresaVidaNotificacionConfiguracion
		if err := rows.Scan(&sub.ID, &sub.EmpresaID, &sub.UsuarioID, &sub.Nombre, &sub.Proveedor, &sub.Costo, &sub.Moneda, &sub.Periodicidad, &sub.Intervalo, &sub.FechaInicio, &sub.ProximaRenovacion, &sub.RecordatorioDias, &sub.TipoRecordatorio, &sub.AutoRenovacion, &sub.Estado, &sub.Notas, &sub.DiasRestantes, &sub.ClientRequestID, &sub.RequestHash, &sub.FechaCreacion, &sub.FechaActualizacion, &cfg.EmailActiva, &cfg.WhatsAppActiva, &cfg.WhatsAppTelefono, &cfg.HoraLocal); err != nil {
			return nil, err
		}
		cfg.EmpresaID, cfg.UsuarioID = sub.EmpresaID, sub.UsuarioID
		out = append(out, EmpresaVidaRecordatorioPendiente{Suscripcion: sub, Config: cfg})
	}
	return out, rows.Err()
}

// ClaimEmpresaVidaNotificacion reserves one channel. A failed notification can
// be retried at most five times, while a successful one is immutable for that
// subscription renewal date.
func ClaimEmpresaVidaNotificacion(dbConn *sql.DB, sub EmpresaVidaSuscripcion, canal string) (bool, error) {
	if canal != "email" && canal != "whatsapp" {
		return false, fmt.Errorf("canal de Vida invalido")
	}
	result, err := ExecCompat(dbConn, `INSERT INTO empresa_vida_notificaciones (empresa_id, usuario_id, suscripcion_id, proxima_renovacion, canal, estado, intentos) VALUES (?, ?, ?, CAST(? AS DATE), ?, 'procesando', 1) ON CONFLICT DO NOTHING`, sub.EmpresaID, normalizeVidaUsuario(sub.UsuarioID), sub.ID, sub.ProximaRenovacion, canal)
	if err != nil {
		return false, err
	}
	if affected, _ := result.RowsAffected(); affected == 1 {
		return true, nil
	}
	result, err = ExecCompat(dbConn, `UPDATE empresa_vida_notificaciones SET estado='procesando', intentos=intentos+1, fecha_ultimo_intento=CURRENT_TIMESTAMP, error_publico='' WHERE empresa_id=? AND usuario_id=? AND suscripcion_id=? AND proxima_renovacion=CAST(? AS DATE) AND canal=? AND estado='error' AND intentos<5 AND fecha_ultimo_intento < CURRENT_TIMESTAMP - INTERVAL '30 minutes'`, sub.EmpresaID, normalizeVidaUsuario(sub.UsuarioID), sub.ID, sub.ProximaRenovacion, canal)
	if err != nil {
		return false, err
	}
	affected, _ := result.RowsAffected()
	return affected == 1, nil
}

func CompleteEmpresaVidaNotificacion(dbConn *sql.DB, sub EmpresaVidaSuscripcion, canal, estado, publicError string) error {
	if estado != "enviado" && estado != "error" && estado != "omitido" {
		return fmt.Errorf("estado de Vida invalido")
	}
	if len(publicError) > 240 {
		publicError = publicError[:240]
	}
	_, err := ExecCompat(dbConn, `UPDATE empresa_vida_notificaciones SET estado=?, fecha_ultimo_intento=CURRENT_TIMESTAMP, fecha_envio=CASE WHEN ?='enviado' THEN CURRENT_TIMESTAMP ELSE fecha_envio END, error_publico=? WHERE empresa_id=? AND usuario_id=? AND suscripcion_id=? AND proxima_renovacion=CAST(? AS DATE) AND canal=?`, estado, estado, strings.TrimSpace(publicError), sub.EmpresaID, normalizeVidaUsuario(sub.UsuarioID), sub.ID, sub.ProximaRenovacion, canal)
	return err
}

func normalizeVidaUsuario(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}
