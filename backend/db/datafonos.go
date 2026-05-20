package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"strings"
)

const (
	DatafonoProviderRedeban    = "redeban"
	DatafonoProviderCredibanco = "credibanco"
	DatafonoProviderBold       = "bold"
	DatafonoProviderBBVA       = "bbva"

	DatafonoEstadoPendiente = "pendiente"
	DatafonoEstadoAprobado  = "aprobado"
	DatafonoEstadoRechazado = "rechazado"
	DatafonoEstadoError     = "error"
)

// EmpresaDatafonoConfig guarda la configuracion contractual del proveedor por empresa.
type EmpresaDatafonoConfig struct {
	ID                 int64  `json:"id"`
	EmpresaID          int64  `json:"empresa_id"`
	Proveedor          string `json:"proveedor"`
	Nombre             string `json:"nombre"`
	TerminalID         string `json:"terminal_id"`
	ComercioID         string `json:"comercio_id"`
	ApiBaseURL         string `json:"api_base_url"`
	CrearPagoPath      string `json:"crear_pago_path"`
	ConsultarPagoPath  string `json:"consultar_pago_path"`
	AuthMode           string `json:"auth_mode"`
	AuthHeader         string `json:"auth_header"`
	ApiKeyRef          string `json:"api_key_ref"`
	TimeoutMs          int    `json:"timeout_ms"`
	Moneda             string `json:"moneda"`
	MetodoPagoPOS      string `json:"metodo_pago_pos"`
	Activo             bool   `json:"activo"`
	FechaCreacion      string `json:"fecha_creacion,omitempty"`
	FechaActualizacion string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string `json:"usuario_creador,omitempty"`
	Estado             string `json:"estado,omitempty"`
	Observaciones      string `json:"observaciones,omitempty"`
}

type EmpresaDatafonoCliente struct {
	Nombre    string `json:"nombre"`
	Documento string `json:"documento"`
	Email     string `json:"email"`
	Telefono  string `json:"telefono"`
}

type EmpresaDatafonoPaymentRequest struct {
	EmpresaID  int64                  `json:"empresa_id"`
	ConfigID   int64                  `json:"config_id"`
	Proveedor  string                 `json:"proveedor"`
	CarritoID  int64                  `json:"carrito_id"`
	Monto      float64                `json:"monto"`
	Moneda     string                 `json:"moneda"`
	Referencia string                 `json:"referencia"`
	Cliente    EmpresaDatafonoCliente `json:"cliente"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

type EmpresaDatafonoProviderResponse struct {
	ProviderTransactionID string                 `json:"provider_transaction_id"`
	EstadoPago            string                 `json:"estado_pago"`
	CodigoAutorizacion    string                 `json:"codigo_autorizacion"`
	MensajeRespuesta      string                 `json:"mensaje_respuesta"`
	Referencia            string                 `json:"referencia"`
	Monto                 float64                `json:"monto"`
	Moneda                string                 `json:"moneda"`
	Raw                   map[string]interface{} `json:"raw,omitempty"`
}

type EmpresaDatafonoTransaction struct {
	ID                    int64   `json:"id"`
	EmpresaID             int64   `json:"empresa_id"`
	ConfigID              int64   `json:"config_id"`
	Proveedor             string  `json:"proveedor"`
	CarritoID             int64   `json:"carrito_id"`
	Referencia            string  `json:"referencia"`
	Monto                 float64 `json:"monto"`
	Moneda                string  `json:"moneda"`
	ClienteNombre         string  `json:"cliente_nombre"`
	ClienteDocumento      string  `json:"cliente_documento"`
	ClienteEmail          string  `json:"cliente_email"`
	ClienteTelefono       string  `json:"cliente_telefono"`
	EstadoPago            string  `json:"estado_pago"`
	ProviderTransactionID string  `json:"provider_transaction_id"`
	CodigoAutorizacion    string  `json:"codigo_autorizacion"`
	MensajeRespuesta      string  `json:"mensaje_respuesta"`
	RequestJSON           string  `json:"request_json,omitempty"`
	ResponseJSON          string  `json:"response_json,omitempty"`
	UltimaConsultaJSON    string  `json:"ultima_consulta_json,omitempty"`
	Validado              bool    `json:"validado"`
	ErrorMensaje          string  `json:"error_mensaje,omitempty"`
	FechaSolicitud        string  `json:"fecha_solicitud,omitempty"`
	FechaConfirmacion     string  `json:"fecha_confirmacion,omitempty"`
	FechaCreacion         string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion    string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador        string  `json:"usuario_creador,omitempty"`
	Estado                string  `json:"estado,omitempty"`
	Observaciones         string  `json:"observaciones,omitempty"`
}

func EnsureEmpresaDatafonosSchema(dbConn *sql.DB) error {
	if dbConn == nil {
		return fmt.Errorf("conexion de base de datos no disponible")
	}
	if _, err := execSQLCompat(dbConn, `CREATE TABLE IF NOT EXISTS empresa_datafonos_config (
		id BIGSERIAL PRIMARY KEY,
		empresa_id INTEGER NOT NULL,
		proveedor TEXT NOT NULL,
		nombre TEXT,
		terminal_id TEXT,
		comercio_id TEXT,
		api_base_url TEXT,
		crear_pago_path TEXT,
		consultar_pago_path TEXT,
		auth_mode TEXT DEFAULT 'none',
		auth_header TEXT,
		api_key_ref TEXT,
		timeout_ms INTEGER DEFAULT 15000,
		moneda TEXT DEFAULT 'COP',
		metodo_pago_pos TEXT DEFAULT 'tarjeta_debito',
		activo INTEGER DEFAULT 1,
		fecha_creacion TEXT DEFAULT (CAST(CURRENT_TIMESTAMP AS TEXT)),
		fecha_actualizacion TEXT DEFAULT (CAST(CURRENT_TIMESTAMP AS TEXT)),
		usuario_creador TEXT,
		estado TEXT DEFAULT 'activo',
		observaciones TEXT
	)`); err != nil {
		return err
	}
	if _, err := execSQLCompat(dbConn, `CREATE TABLE IF NOT EXISTS empresa_datafonos_transacciones (
		id BIGSERIAL PRIMARY KEY,
		empresa_id INTEGER NOT NULL,
		config_id INTEGER DEFAULT 0,
		proveedor TEXT NOT NULL,
		carrito_id INTEGER DEFAULT 0,
		referencia TEXT NOT NULL,
		monto REAL DEFAULT 0,
		moneda TEXT DEFAULT 'COP',
		cliente_nombre TEXT,
		cliente_documento TEXT,
		cliente_email TEXT,
		cliente_telefono TEXT,
		estado_pago TEXT DEFAULT 'pendiente',
		provider_transaction_id TEXT,
		codigo_autorizacion TEXT,
		mensaje_respuesta TEXT,
		request_json TEXT,
		response_json TEXT,
		ultima_consulta_json TEXT,
		validado INTEGER DEFAULT 0,
		error_mensaje TEXT,
		fecha_solicitud TEXT DEFAULT (CAST(CURRENT_TIMESTAMP AS TEXT)),
		fecha_confirmacion TEXT,
		fecha_creacion TEXT DEFAULT (CAST(CURRENT_TIMESTAMP AS TEXT)),
		fecha_actualizacion TEXT DEFAULT (CAST(CURRENT_TIMESTAMP AS TEXT)),
		usuario_creador TEXT,
		estado TEXT DEFAULT 'activo',
		observaciones TEXT
	)`); err != nil {
		return err
	}

	for _, stmt := range []string{
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_empresa_datafonos_config_terminal ON empresa_datafonos_config(empresa_id, proveedor, terminal_id) WHERE COALESCE(terminal_id,'') <> ''`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_datafonos_config_empresa ON empresa_datafonos_config(empresa_id, estado, activo)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_empresa_datafonos_tx_ref ON empresa_datafonos_transacciones(empresa_id, proveedor, referencia)`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_datafonos_tx_carrito ON empresa_datafonos_transacciones(empresa_id, carrito_id)`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_datafonos_tx_estado ON empresa_datafonos_transacciones(empresa_id, estado_pago, fecha_creacion)`,
	} {
		if _, err := execSQLCompat(dbConn, stmt); err != nil {
			return err
		}
	}
	return nil
}

func NormalizeDatafonoProvider(provider string) string {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "redeban", "red-eban":
		return DatafonoProviderRedeban
	case "credibanco", "crediban-co", "crediban":
		return DatafonoProviderCredibanco
	case "bold":
		return DatafonoProviderBold
	case "bbva":
		return DatafonoProviderBBVA
	default:
		return ""
	}
}

func NormalizeDatafonoEstadoPago(estado string) string {
	switch strings.ToLower(strings.TrimSpace(estado)) {
	case "aprobado", "approved", "approve", "paid", "pagado", "success", "successful", "ok", "completed", "complete", "captured":
		return DatafonoEstadoAprobado
	case "rechazado", "declined", "rejected", "denied", "failed", "failure", "cancelled", "canceled", "voided":
		return DatafonoEstadoRechazado
	case "error":
		return DatafonoEstadoError
	default:
		return DatafonoEstadoPendiente
	}
}

func NormalizeDatafonoMoneda(moneda string) string {
	moneda = strings.ToUpper(strings.TrimSpace(moneda))
	if moneda == "" {
		return "COP"
	}
	if len(moneda) > 8 {
		return moneda[:8]
	}
	return moneda
}

func UpsertEmpresaDatafonoConfig(dbConn *sql.DB, cfg EmpresaDatafonoConfig) (int64, error) {
	if dbConn == nil {
		return 0, fmt.Errorf("conexion de base de datos no disponible")
	}
	if cfg.EmpresaID <= 0 {
		return 0, fmt.Errorf("empresa_id invalido")
	}
	cfg.Proveedor = NormalizeDatafonoProvider(cfg.Proveedor)
	if cfg.Proveedor == "" {
		return 0, fmt.Errorf("proveedor de datafono no soportado")
	}
	cfg.Nombre = strings.TrimSpace(cfg.Nombre)
	if cfg.Nombre == "" {
		cfg.Nombre = strings.Title(cfg.Proveedor)
	}
	cfg.AuthMode = normalizeDatafonoAuthMode(cfg.AuthMode)
	cfg.AuthHeader = strings.TrimSpace(cfg.AuthHeader)
	cfg.ApiKeyRef = strings.TrimSpace(cfg.ApiKeyRef)
	if cfg.ApiKeyRef != "" && !strings.HasPrefix(strings.ToLower(cfg.ApiKeyRef), "env:") {
		return 0, fmt.Errorf("api_key_ref debe usar referencia segura env:NOMBRE_VARIABLE")
	}
	if cfg.TimeoutMs <= 0 {
		cfg.TimeoutMs = 15000
	}
	if cfg.TimeoutMs < 3000 {
		cfg.TimeoutMs = 3000
	}
	if cfg.TimeoutMs > 60000 {
		cfg.TimeoutMs = 60000
	}
	cfg.Moneda = NormalizeDatafonoMoneda(cfg.Moneda)
	cfg.MetodoPagoPOS = NormalizeMetodoPagoCarrito(cfg.MetodoPagoPOS)
	if cfg.MetodoPagoPOS == "" || cfg.MetodoPagoPOS == "efectivo" || cfg.MetodoPagoPOS == "codigo_descuento" {
		cfg.MetodoPagoPOS = "tarjeta_debito"
	}
	estado := strings.TrimSpace(cfg.Estado)
	if estado == "" {
		estado = "activo"
	}
	activo := 0
	if cfg.Activo {
		activo = 1
	}

	if cfg.ID > 0 {
		_, err := execSQLCompat(dbConn, `UPDATE empresa_datafonos_config SET
			proveedor = ?, nombre = ?, terminal_id = ?, comercio_id = ?, api_base_url = ?,
			crear_pago_path = ?, consultar_pago_path = ?, auth_mode = ?, auth_header = ?,
			api_key_ref = ?, timeout_ms = ?, moneda = ?, metodo_pago_pos = ?, activo = ?,
			fecha_actualizacion = `+sqlNowExpr()+`, usuario_creador = ?, estado = ?, observaciones = ?
			WHERE empresa_id = ? AND id = ?`,
			cfg.Proveedor, cfg.Nombre, strings.TrimSpace(cfg.TerminalID), strings.TrimSpace(cfg.ComercioID), strings.TrimSpace(cfg.ApiBaseURL),
			strings.TrimSpace(cfg.CrearPagoPath), strings.TrimSpace(cfg.ConsultarPagoPath), cfg.AuthMode, cfg.AuthHeader,
			cfg.ApiKeyRef, cfg.TimeoutMs, cfg.Moneda, cfg.MetodoPagoPOS, activo,
			strings.TrimSpace(cfg.UsuarioCreador), estado, strings.TrimSpace(cfg.Observaciones), cfg.EmpresaID, cfg.ID)
		if err != nil {
			return 0, err
		}
		return cfg.ID, nil
	}

	return insertSQLCompat(dbConn, `INSERT INTO empresa_datafonos_config (
		empresa_id, proveedor, nombre, terminal_id, comercio_id, api_base_url,
		crear_pago_path, consultar_pago_path, auth_mode, auth_header, api_key_ref,
		timeout_ms, moneda, metodo_pago_pos, activo, fecha_creacion, fecha_actualizacion,
		usuario_creador, estado, observaciones
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, `+sqlNowExpr()+`, `+sqlNowExpr()+`, ?, ?, ?)`,
		cfg.EmpresaID, cfg.Proveedor, cfg.Nombre, strings.TrimSpace(cfg.TerminalID), strings.TrimSpace(cfg.ComercioID), strings.TrimSpace(cfg.ApiBaseURL),
		strings.TrimSpace(cfg.CrearPagoPath), strings.TrimSpace(cfg.ConsultarPagoPath), cfg.AuthMode, cfg.AuthHeader, cfg.ApiKeyRef,
		cfg.TimeoutMs, cfg.Moneda, cfg.MetodoPagoPOS, activo, strings.TrimSpace(cfg.UsuarioCreador), estado, strings.TrimSpace(cfg.Observaciones))
}

func ListEmpresaDatafonoConfigs(dbConn *sql.DB, empresaID int64, includeInactive bool) ([]EmpresaDatafonoConfig, error) {
	if empresaID <= 0 {
		return nil, fmt.Errorf("empresa_id invalido")
	}
	query := `SELECT id, empresa_id, proveedor, COALESCE(nombre,''), COALESCE(terminal_id,''), COALESCE(comercio_id,''),
		COALESCE(api_base_url,''), COALESCE(crear_pago_path,''), COALESCE(consultar_pago_path,''),
		COALESCE(auth_mode,'none'), COALESCE(auth_header,''), COALESCE(api_key_ref,''), COALESCE(timeout_ms,15000),
		COALESCE(moneda,'COP'), COALESCE(metodo_pago_pos,'tarjeta_debito'), COALESCE(activo,1),
		COALESCE(fecha_creacion,''), COALESCE(fecha_actualizacion,''), COALESCE(usuario_creador,''), COALESCE(estado,'activo'), COALESCE(observaciones,'')
		FROM empresa_datafonos_config WHERE empresa_id = ?`
	if !includeInactive {
		query += ` AND COALESCE(estado,'activo') = 'activo' AND COALESCE(activo,1) = 1`
	}
	query += ` ORDER BY proveedor, nombre, id`
	rows, err := querySQLCompat(dbConn, query, empresaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]EmpresaDatafonoConfig, 0)
	for rows.Next() {
		item, err := scanEmpresaDatafonoConfig(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func GetEmpresaDatafonoConfig(dbConn *sql.DB, empresaID, configID int64, provider string) (EmpresaDatafonoConfig, error) {
	provider = NormalizeDatafonoProvider(provider)
	var row *sql.Row
	if configID > 0 {
		row = queryRowSQLCompat(dbConn, `SELECT id, empresa_id, proveedor, COALESCE(nombre,''), COALESCE(terminal_id,''), COALESCE(comercio_id,''),
			COALESCE(api_base_url,''), COALESCE(crear_pago_path,''), COALESCE(consultar_pago_path,''),
			COALESCE(auth_mode,'none'), COALESCE(auth_header,''), COALESCE(api_key_ref,''), COALESCE(timeout_ms,15000),
			COALESCE(moneda,'COP'), COALESCE(metodo_pago_pos,'tarjeta_debito'), COALESCE(activo,1),
			COALESCE(fecha_creacion,''), COALESCE(fecha_actualizacion,''), COALESCE(usuario_creador,''), COALESCE(estado,'activo'), COALESCE(observaciones,'')
			FROM empresa_datafonos_config WHERE empresa_id = ? AND id = ? LIMIT 1`, empresaID, configID)
	} else {
		row = queryRowSQLCompat(dbConn, `SELECT id, empresa_id, proveedor, COALESCE(nombre,''), COALESCE(terminal_id,''), COALESCE(comercio_id,''),
			COALESCE(api_base_url,''), COALESCE(crear_pago_path,''), COALESCE(consultar_pago_path,''),
			COALESCE(auth_mode,'none'), COALESCE(auth_header,''), COALESCE(api_key_ref,''), COALESCE(timeout_ms,15000),
			COALESCE(moneda,'COP'), COALESCE(metodo_pago_pos,'tarjeta_debito'), COALESCE(activo,1),
			COALESCE(fecha_creacion,''), COALESCE(fecha_actualizacion,''), COALESCE(usuario_creador,''), COALESCE(estado,'activo'), COALESCE(observaciones,'')
			FROM empresa_datafonos_config WHERE empresa_id = ? AND proveedor = ? AND COALESCE(estado,'activo') = 'activo' AND COALESCE(activo,1) = 1 ORDER BY id DESC LIMIT 1`, empresaID, provider)
	}
	return scanEmpresaDatafonoConfigRow(row)
}

func CreateEmpresaDatafonoTransaction(dbConn *sql.DB, tx EmpresaDatafonoTransaction) (int64, error) {
	if dbConn == nil {
		return 0, fmt.Errorf("conexion de base de datos no disponible")
	}
	if tx.EmpresaID <= 0 {
		return 0, fmt.Errorf("empresa_id invalido")
	}
	tx.Proveedor = NormalizeDatafonoProvider(tx.Proveedor)
	if tx.Proveedor == "" {
		return 0, fmt.Errorf("proveedor de datafono no soportado")
	}
	tx.Referencia = strings.TrimSpace(tx.Referencia)
	if tx.Referencia == "" {
		return 0, fmt.Errorf("referencia de pago obligatoria")
	}
	if tx.Monto <= 0 || math.IsNaN(tx.Monto) || math.IsInf(tx.Monto, 0) {
		return 0, fmt.Errorf("monto invalido")
	}
	tx.Moneda = NormalizeDatafonoMoneda(tx.Moneda)
	tx.EstadoPago = NormalizeDatafonoEstadoPago(tx.EstadoPago)
	estado := strings.TrimSpace(tx.Estado)
	if estado == "" {
		estado = "activo"
	}
	validado := 0
	if tx.Validado {
		validado = 1
	}
	return insertSQLCompat(dbConn, `INSERT INTO empresa_datafonos_transacciones (
		empresa_id, config_id, proveedor, carrito_id, referencia, monto, moneda,
		cliente_nombre, cliente_documento, cliente_email, cliente_telefono,
		estado_pago, provider_transaction_id, codigo_autorizacion, mensaje_respuesta,
		request_json, response_json, ultima_consulta_json, validado, error_mensaje,
		fecha_solicitud, fecha_confirmacion, fecha_creacion, fecha_actualizacion,
		usuario_creador, estado, observaciones
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, COALESCE(NULLIF(?, ''), `+sqlNowExpr()+`), NULLIF(?, ''), `+sqlNowExpr()+`, `+sqlNowExpr()+`, ?, ?, ?)`,
		tx.EmpresaID, tx.ConfigID, tx.Proveedor, tx.CarritoID, tx.Referencia, round2(tx.Monto), tx.Moneda,
		strings.TrimSpace(tx.ClienteNombre), strings.TrimSpace(tx.ClienteDocumento), strings.TrimSpace(tx.ClienteEmail), strings.TrimSpace(tx.ClienteTelefono),
		tx.EstadoPago, strings.TrimSpace(tx.ProviderTransactionID), strings.TrimSpace(tx.CodigoAutorizacion), strings.TrimSpace(tx.MensajeRespuesta),
		strings.TrimSpace(tx.RequestJSON), strings.TrimSpace(tx.ResponseJSON), strings.TrimSpace(tx.UltimaConsultaJSON), validado, strings.TrimSpace(tx.ErrorMensaje),
		strings.TrimSpace(tx.FechaSolicitud), strings.TrimSpace(tx.FechaConfirmacion), strings.TrimSpace(tx.UsuarioCreador), estado, strings.TrimSpace(tx.Observaciones))
}

func GetEmpresaDatafonoTransaction(dbConn *sql.DB, empresaID, txID int64, provider, reference string) (EmpresaDatafonoTransaction, error) {
	provider = NormalizeDatafonoProvider(provider)
	reference = strings.TrimSpace(reference)
	var row *sql.Row
	if txID > 0 {
		row = queryRowSQLCompat(dbConn, datafonoTransactionSelectSQL()+` WHERE empresa_id = ? AND id = ? LIMIT 1`, empresaID, txID)
	} else {
		row = queryRowSQLCompat(dbConn, datafonoTransactionSelectSQL()+` WHERE empresa_id = ? AND proveedor = ? AND referencia = ? LIMIT 1`, empresaID, provider, reference)
	}
	return scanEmpresaDatafonoTransactionRow(row)
}

func ListEmpresaDatafonoTransactions(dbConn *sql.DB, empresaID int64, limit int) ([]EmpresaDatafonoTransaction, error) {
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	rows, err := querySQLCompat(dbConn, datafonoTransactionSelectSQL()+` WHERE empresa_id = ? ORDER BY id DESC LIMIT ?`, empresaID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]EmpresaDatafonoTransaction, 0)
	for rows.Next() {
		item, err := scanEmpresaDatafonoTransaction(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func UpdateEmpresaDatafonoTransactionFromProvider(dbConn *sql.DB, empresaID, txID int64, resp EmpresaDatafonoProviderResponse, responseJSON, consultaJSON, errorMensaje string) error {
	estadoPago := NormalizeDatafonoEstadoPago(resp.EstadoPago)
	validado := 0
	fechaConfirmacion := ""
	if estadoPago == DatafonoEstadoAprobado {
		validado = 1
		fechaConfirmacion = sqlNowExpr()
	}
	setFecha := "NULLIF(fecha_confirmacion, '')"
	if fechaConfirmacion != "" {
		setFecha = sqlNowExpr()
	}
	_, err := execSQLCompat(dbConn, `UPDATE empresa_datafonos_transacciones SET
		estado_pago = ?,
		provider_transaction_id = COALESCE(NULLIF(?, ''), provider_transaction_id),
		codigo_autorizacion = COALESCE(NULLIF(?, ''), codigo_autorizacion),
		mensaje_respuesta = COALESCE(NULLIF(?, ''), mensaje_respuesta),
		response_json = COALESCE(NULLIF(?, ''), response_json),
		ultima_consulta_json = COALESCE(NULLIF(?, ''), ultima_consulta_json),
		validado = ?,
		error_mensaje = ?,
		fecha_confirmacion = `+setFecha+`,
		fecha_actualizacion = `+sqlNowExpr()+`
		WHERE empresa_id = ? AND id = ?`,
		estadoPago, strings.TrimSpace(resp.ProviderTransactionID), strings.TrimSpace(resp.CodigoAutorizacion), strings.TrimSpace(resp.MensajeRespuesta),
		strings.TrimSpace(responseJSON), strings.TrimSpace(consultaJSON), validado, strings.TrimSpace(errorMensaje), empresaID, txID)
	return err
}

func MarkEmpresaDatafonoTransactionPOSApplied(dbConn *sql.DB, empresaID, txID int64, carritoID int64) error {
	_, err := execSQLCompat(dbConn, `UPDATE empresa_datafonos_transacciones SET
		observaciones = TRIM(COALESCE(observaciones,'') || CASE WHEN COALESCE(observaciones,'') = '' THEN '' ELSE ' | ' END || ?),
		fecha_actualizacion = `+sqlNowExpr()+`
		WHERE empresa_id = ? AND id = ?`,
		fmt.Sprintf("pago aplicado al POS en carrito %d", carritoID), empresaID, txID)
	return err
}

func ValidateDatafonoAmountAndReference(req EmpresaDatafonoPaymentRequest, resp EmpresaDatafonoProviderResponse) error {
	if NormalizeDatafonoEstadoPago(resp.EstadoPago) != DatafonoEstadoAprobado {
		return nil
	}
	if resp.Monto > 0 && math.Abs(round2(resp.Monto)-round2(req.Monto)) > 0.01 {
		return fmt.Errorf("monto confirmado por datafono no coincide con la venta")
	}
	if strings.TrimSpace(resp.Referencia) != "" && strings.TrimSpace(req.Referencia) != "" && strings.TrimSpace(resp.Referencia) != strings.TrimSpace(req.Referencia) {
		return fmt.Errorf("referencia confirmada por datafono no coincide con la venta")
	}
	return nil
}

func DatafonoRequestJSON(req EmpresaDatafonoPaymentRequest) string {
	raw, _ := json.Marshal(req)
	return string(raw)
}

func DatafonoProviderResponseJSON(resp EmpresaDatafonoProviderResponse) string {
	raw, _ := json.Marshal(resp)
	return string(raw)
}

func normalizeDatafonoAuthMode(mode string) string {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "bearer", "api_key", "basic":
		return strings.ToLower(strings.TrimSpace(mode))
	default:
		return "none"
	}
}

func scanEmpresaDatafonoConfig(rows *sql.Rows) (EmpresaDatafonoConfig, error) {
	var item EmpresaDatafonoConfig
	var activo int
	err := rows.Scan(&item.ID, &item.EmpresaID, &item.Proveedor, &item.Nombre, &item.TerminalID, &item.ComercioID,
		&item.ApiBaseURL, &item.CrearPagoPath, &item.ConsultarPagoPath, &item.AuthMode, &item.AuthHeader, &item.ApiKeyRef,
		&item.TimeoutMs, &item.Moneda, &item.MetodoPagoPOS, &activo, &item.FechaCreacion, &item.FechaActualizacion,
		&item.UsuarioCreador, &item.Estado, &item.Observaciones)
	item.Activo = activo > 0
	return item, err
}

func scanEmpresaDatafonoConfigRow(row *sql.Row) (EmpresaDatafonoConfig, error) {
	var item EmpresaDatafonoConfig
	var activo int
	err := row.Scan(&item.ID, &item.EmpresaID, &item.Proveedor, &item.Nombre, &item.TerminalID, &item.ComercioID,
		&item.ApiBaseURL, &item.CrearPagoPath, &item.ConsultarPagoPath, &item.AuthMode, &item.AuthHeader, &item.ApiKeyRef,
		&item.TimeoutMs, &item.Moneda, &item.MetodoPagoPOS, &activo, &item.FechaCreacion, &item.FechaActualizacion,
		&item.UsuarioCreador, &item.Estado, &item.Observaciones)
	item.Activo = activo > 0
	return item, err
}

func datafonoTransactionSelectSQL() string {
	return `SELECT id, empresa_id, COALESCE(config_id,0), proveedor, COALESCE(carrito_id,0), referencia,
		COALESCE(monto,0), COALESCE(moneda,'COP'), COALESCE(cliente_nombre,''), COALESCE(cliente_documento,''),
		COALESCE(cliente_email,''), COALESCE(cliente_telefono,''), COALESCE(estado_pago,'pendiente'),
		COALESCE(provider_transaction_id,''), COALESCE(codigo_autorizacion,''), COALESCE(mensaje_respuesta,''),
		COALESCE(request_json,''), COALESCE(response_json,''), COALESCE(ultima_consulta_json,''), COALESCE(validado,0),
		COALESCE(error_mensaje,''), COALESCE(fecha_solicitud,''), COALESCE(fecha_confirmacion,''),
		COALESCE(fecha_creacion,''), COALESCE(fecha_actualizacion,''), COALESCE(usuario_creador,''), COALESCE(estado,'activo'), COALESCE(observaciones,'')
		FROM empresa_datafonos_transacciones`
}

func scanEmpresaDatafonoTransaction(rows *sql.Rows) (EmpresaDatafonoTransaction, error) {
	var item EmpresaDatafonoTransaction
	var validado int
	err := rows.Scan(&item.ID, &item.EmpresaID, &item.ConfigID, &item.Proveedor, &item.CarritoID, &item.Referencia,
		&item.Monto, &item.Moneda, &item.ClienteNombre, &item.ClienteDocumento, &item.ClienteEmail, &item.ClienteTelefono,
		&item.EstadoPago, &item.ProviderTransactionID, &item.CodigoAutorizacion, &item.MensajeRespuesta,
		&item.RequestJSON, &item.ResponseJSON, &item.UltimaConsultaJSON, &validado, &item.ErrorMensaje,
		&item.FechaSolicitud, &item.FechaConfirmacion, &item.FechaCreacion, &item.FechaActualizacion,
		&item.UsuarioCreador, &item.Estado, &item.Observaciones)
	item.Validado = validado > 0
	return item, err
}

func scanEmpresaDatafonoTransactionRow(row *sql.Row) (EmpresaDatafonoTransaction, error) {
	var item EmpresaDatafonoTransaction
	var validado int
	err := row.Scan(&item.ID, &item.EmpresaID, &item.ConfigID, &item.Proveedor, &item.CarritoID, &item.Referencia,
		&item.Monto, &item.Moneda, &item.ClienteNombre, &item.ClienteDocumento, &item.ClienteEmail, &item.ClienteTelefono,
		&item.EstadoPago, &item.ProviderTransactionID, &item.CodigoAutorizacion, &item.MensajeRespuesta,
		&item.RequestJSON, &item.ResponseJSON, &item.UltimaConsultaJSON, &validado, &item.ErrorMensaje,
		&item.FechaSolicitud, &item.FechaConfirmacion, &item.FechaCreacion, &item.FechaActualizacion,
		&item.UsuarioCreador, &item.Estado, &item.Observaciones)
	item.Validado = validado > 0
	return item, err
}
