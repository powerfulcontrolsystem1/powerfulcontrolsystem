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

func VerifyEmpresaVidaSchema(dbConn *sql.DB) error {
	if dbConn == nil {
		return fmt.Errorf("database not available")
	}
	for _, table := range []string{"empresa_vida_gastos", "empresa_vida_suscripciones", "empresa_vida_precios"} {
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

func normalizeVidaUsuario(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}
