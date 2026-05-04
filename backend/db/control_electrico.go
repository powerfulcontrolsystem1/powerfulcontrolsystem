package db

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"sync"
)

const (
	DefaultControlElectricoPort      = 8081
	DefaultControlElectricoAPIPath   = "/api/gpio/relay"
	DefaultControlElectricoTimeoutMS = 2500
)

var (
	empresaControlElectricoSchemaMu    sync.Mutex
	empresaControlElectricoSchemaReady bool
)

// EmpresaControlElectricoConfig guarda la conexion principal contra la Raspberry Pi.
type EmpresaControlElectricoConfig struct {
	ID                 int64  `json:"id"`
	EmpresaID          int64  `json:"empresa_id"`
	Habilitado         bool   `json:"habilitado"`
	RaspberryIP        string `json:"raspberry_ip"`
	RaspberryPort      int    `json:"raspberry_port"`
	APIPath            string `json:"api_path"`
	APIToken           string `json:"api_token,omitempty"`
	APITokenConfigured bool   `json:"api_token_configured"`
	TimeoutMS          int    `json:"timeout_ms"`
	AutoSyncEstaciones bool   `json:"auto_sync_estaciones"`
	FailSafeOnError    bool   `json:"fail_safe_on_error"`
	FechaCreacion      string `json:"fecha_creacion,omitempty"`
	FechaActualizacion string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string `json:"usuario_creador,omitempty"`
	Estado             string `json:"estado,omitempty"`
	Observaciones      string `json:"observaciones,omitempty"`
}

// EmpresaControlElectricoRaspberry representa un controlador fisico GPIO adicional.
type EmpresaControlElectricoRaspberry struct {
	ID                 int64  `json:"id"`
	EmpresaID          int64  `json:"empresa_id"`
	Codigo             string `json:"codigo"`
	Nombre             string `json:"nombre"`
	RaspberryIP        string `json:"raspberry_ip"`
	RaspberryPort      int    `json:"raspberry_port"`
	APIPath            string `json:"api_path"`
	APIToken           string `json:"api_token,omitempty"`
	APITokenConfigured bool   `json:"api_token_configured"`
	TimeoutMS          int    `json:"timeout_ms"`
	FechaCreacion      string `json:"fecha_creacion,omitempty"`
	FechaActualizacion string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string `json:"usuario_creador,omitempty"`
	Estado             string `json:"estado,omitempty"`
	Observaciones      string `json:"observaciones,omitempty"`
}

// EmpresaControlElectricoRele representa la salida fisica asociada a una estacion.
type EmpresaControlElectricoRele struct {
	ID                   int64  `json:"id"`
	EmpresaID            int64  `json:"empresa_id"`
	RaspberryID          int64  `json:"raspberry_id,omitempty"`
	RaspberryCodigo      string `json:"raspberry_codigo,omitempty"`
	RaspberryNombre      string `json:"raspberry_nombre,omitempty"`
	RaspberryIP          string `json:"raspberry_ip,omitempty"`
	EstacionID           int64  `json:"estacion_id"`
	EstacionCodigo       string `json:"estacion_codigo,omitempty"`
	EstacionNombre       string `json:"estacion_nombre,omitempty"`
	SalidaCodigo         string `json:"salida_codigo"`
	TipoCarga            string `json:"tipo_carga,omitempty"`
	GPIOPin              int    `json:"gpio_pin"`
	RelayName            string `json:"relay_name"`
	ActiveHigh           bool   `json:"active_high"`
	PulsoMS              int    `json:"pulso_ms"`
	Modo                 string `json:"modo"`
	UltimoEstado         string `json:"ultimo_estado,omitempty"`
	UltimoComando        string `json:"ultimo_comando,omitempty"`
	UltimoError          string `json:"ultimo_error,omitempty"`
	UltimaSincronizacion string `json:"ultima_sincronizacion,omitempty"`
	FechaCreacion        string `json:"fecha_creacion,omitempty"`
	FechaActualizacion   string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador       string `json:"usuario_creador,omitempty"`
	Estado               string `json:"estado,omitempty"`
	Observaciones        string `json:"observaciones,omitempty"`
}

// EmpresaControlElectricoEvento deja trazabilidad de cada comando enviado.
type EmpresaControlElectricoEvento struct {
	ID             int64  `json:"id"`
	EmpresaID      int64  `json:"empresa_id"`
	EstacionID     int64  `json:"estacion_id,omitempty"`
	ReleID         int64  `json:"rele_id,omitempty"`
	RaspberryID    int64  `json:"raspberry_id,omitempty"`
	GPIOPin        int    `json:"gpio_pin,omitempty"`
	Comando        string `json:"comando"`
	EstadoObjetivo string `json:"estado_objetivo"`
	Resultado      string `json:"resultado"`
	HTTPStatus     int    `json:"http_status,omitempty"`
	RaspberryIP    string `json:"raspberry_ip,omitempty"`
	ResponseBody   string `json:"response_body,omitempty"`
	Error          string `json:"error,omitempty"`
	FechaEvento    string `json:"fecha_evento,omitempty"`
	Actor          string `json:"actor,omitempty"`
	Origen         string `json:"origen,omitempty"`
	MetadataJSON   string `json:"metadata_json,omitempty"`
}

// EmpresaControlElectricoEstacion resume una estacion operativa y su mapeo electrico.
type EmpresaControlElectricoEstacion struct {
	EstacionID     int64                        `json:"estacion_id"`
	EstacionCodigo string                       `json:"estacion_codigo,omitempty"`
	EstacionNombre string                       `json:"estacion_nombre,omitempty"`
	CarritoID      int64                        `json:"carrito_id,omitempty"`
	Activa         bool                         `json:"activa"`
	Estado         string                       `json:"estado,omitempty"`
	EstadoCarrito  string                       `json:"estado_carrito,omitempty"`
	EstadoVenta    string                       `json:"estado_venta,omitempty"`
	ActivadoEn     string                       `json:"activado_en,omitempty"`
	PagadoEn       string                       `json:"pagado_en,omitempty"`
	Rele           *EmpresaControlElectricoRele `json:"rele,omitempty"`
}

// EnsureEmpresaControlElectricoSchema crea/migra tablas del modulo control electrico.
func EnsureEmpresaControlElectricoSchema(dbConn *sql.DB) error {
	if dbConn == nil {
		return errors.New("db connection is nil")
	}

	empresaControlElectricoSchemaMu.Lock()
	defer empresaControlElectricoSchemaMu.Unlock()

	if empresaControlElectricoSchemaReady {
		return nil
	}

	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_control_electrico_config (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			habilitado INTEGER DEFAULT 0,
			raspberry_ip TEXT,
			raspberry_port INTEGER DEFAULT 8081,
			api_path TEXT DEFAULT '/api/gpio/relay',
			api_token TEXT,
			timeout_ms INTEGER DEFAULT 2500,
			auto_sync_estaciones INTEGER DEFAULT 1,
			fail_safe_on_error INTEGER DEFAULT 0,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_empresa_control_electrico_config_empresa ON empresa_control_electrico_config(empresa_id);`,
		`CREATE TABLE IF NOT EXISTS empresa_control_electrico_raspberry_pis (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			codigo TEXT,
			nombre TEXT,
			raspberry_ip TEXT NOT NULL,
			raspberry_port INTEGER DEFAULT 8081,
			api_path TEXT DEFAULT '/api/gpio/relay',
			api_token TEXT,
			timeout_ms INTEGER DEFAULT 2500,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_empresa_control_electrico_raspberry_codigo ON empresa_control_electrico_raspberry_pis(empresa_id, codigo);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_control_electrico_raspberry_estado ON empresa_control_electrico_raspberry_pis(empresa_id, estado);`,
		`CREATE TABLE IF NOT EXISTS empresa_control_electrico_reles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			raspberry_id INTEGER,
			estacion_id INTEGER NOT NULL,
			estacion_codigo TEXT,
			estacion_nombre TEXT,
			salida_codigo TEXT DEFAULT 'principal',
			tipo_carga TEXT DEFAULT 'luces',
			gpio_pin INTEGER NOT NULL,
			relay_name TEXT,
			active_high INTEGER DEFAULT 1,
			pulso_ms INTEGER DEFAULT 0,
			modo TEXT DEFAULT 'seguimiento_estacion',
			ultimo_estado TEXT DEFAULT 'desconocido',
			ultimo_comando TEXT,
			ultimo_error TEXT,
			ultima_sincronizacion TEXT,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`DROP INDEX IF EXISTS ux_empresa_control_electrico_rele_estacion;`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_control_electrico_reles_empresa_estado ON empresa_control_electrico_reles(empresa_id, estado);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_control_electrico_reles_gpio ON empresa_control_electrico_reles(empresa_id, gpio_pin);`,
		`CREATE TABLE IF NOT EXISTS empresa_control_electrico_eventos (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			estacion_id INTEGER,
			rele_id INTEGER,
			raspberry_id INTEGER,
			gpio_pin INTEGER,
			comando TEXT,
			estado_objetivo TEXT,
			resultado TEXT,
			http_status INTEGER DEFAULT 0,
			raspberry_ip TEXT,
			response_body TEXT,
			error TEXT,
			fecha_evento TEXT DEFAULT (datetime('now','localtime')),
			actor TEXT,
			origen TEXT,
			metadata_json TEXT
		);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_control_electrico_eventos_empresa_fecha ON empresa_control_electrico_eventos(empresa_id, fecha_evento);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_control_electrico_eventos_estacion ON empresa_control_electrico_eventos(empresa_id, estacion_id);`,
	}

	for _, stmt := range stmts {
		if _, err := execSQLCompat(dbConn, stmt); err != nil {
			return err
		}
	}

	configCols := map[string]string{
		"empresa_id": "INTEGER NOT NULL", "habilitado": "INTEGER DEFAULT 0", "raspberry_ip": "TEXT",
		"raspberry_port": "INTEGER DEFAULT 8081", "api_path": "TEXT DEFAULT '/api/gpio/relay'", "api_token": "TEXT",
		"timeout_ms": "INTEGER DEFAULT 2500", "auto_sync_estaciones": "INTEGER DEFAULT 1", "fail_safe_on_error": "INTEGER DEFAULT 0",
		"fecha_creacion": "TEXT DEFAULT (datetime('now','localtime'))", "fecha_actualizacion": "TEXT DEFAULT (datetime('now','localtime'))",
		"usuario_creador": "TEXT", "estado": "TEXT DEFAULT 'activo'", "observaciones": "TEXT",
	}
	for col, def := range configCols {
		if err := ensureColumnIfMissing(dbConn, "empresa_control_electrico_config", col, def); err != nil {
			return err
		}
	}

	raspberryCols := map[string]string{
		"empresa_id": "INTEGER NOT NULL", "codigo": "TEXT", "nombre": "TEXT", "raspberry_ip": "TEXT",
		"raspberry_port": "INTEGER DEFAULT 8081", "api_path": "TEXT DEFAULT '/api/gpio/relay'", "api_token": "TEXT",
		"timeout_ms": "INTEGER DEFAULT 2500", "fecha_creacion": "TEXT DEFAULT (datetime('now','localtime'))",
		"fecha_actualizacion": "TEXT DEFAULT (datetime('now','localtime'))", "usuario_creador": "TEXT",
		"estado": "TEXT DEFAULT 'activo'", "observaciones": "TEXT",
	}
	for col, def := range raspberryCols {
		if err := ensureColumnIfMissing(dbConn, "empresa_control_electrico_raspberry_pis", col, def); err != nil {
			return err
		}
	}

	releCols := map[string]string{
		"empresa_id": "INTEGER NOT NULL", "raspberry_id": "INTEGER", "estacion_id": "INTEGER NOT NULL", "estacion_codigo": "TEXT",
		"estacion_nombre": "TEXT", "salida_codigo": "TEXT DEFAULT 'principal'", "tipo_carga": "TEXT DEFAULT 'luces'",
		"gpio_pin": "INTEGER NOT NULL", "relay_name": "TEXT", "active_high": "INTEGER DEFAULT 1",
		"pulso_ms": "INTEGER DEFAULT 0", "modo": "TEXT DEFAULT 'seguimiento_estacion'", "ultimo_estado": "TEXT DEFAULT 'desconocido'",
		"ultimo_comando": "TEXT", "ultimo_error": "TEXT", "ultima_sincronizacion": "TEXT",
		"fecha_creacion": "TEXT DEFAULT (datetime('now','localtime'))", "fecha_actualizacion": "TEXT DEFAULT (datetime('now','localtime'))",
		"usuario_creador": "TEXT", "estado": "TEXT DEFAULT 'activo'", "observaciones": "TEXT",
	}
	for col, def := range releCols {
		if err := ensureColumnIfMissing(dbConn, "empresa_control_electrico_reles", col, def); err != nil {
			return err
		}
	}
	if _, err := execSQLCompat(dbConn, `UPDATE empresa_control_electrico_reles SET salida_codigo = 'principal' WHERE COALESCE(TRIM(salida_codigo), '') = '';`); err != nil {
		return err
	}
	if _, err := execSQLCompat(dbConn, `UPDATE empresa_control_electrico_reles SET tipo_carga = 'luces' WHERE COALESCE(TRIM(tipo_carga), '') = '';`); err != nil {
		return err
	}
	if _, err := execSQLCompat(dbConn, `CREATE UNIQUE INDEX IF NOT EXISTS ux_empresa_control_electrico_rele_salida ON empresa_control_electrico_reles(empresa_id, estacion_id, salida_codigo);`); err != nil {
		return err
	}

	eventCols := map[string]string{
		"empresa_id": "INTEGER NOT NULL", "estacion_id": "INTEGER", "rele_id": "INTEGER", "raspberry_id": "INTEGER", "gpio_pin": "INTEGER",
		"comando": "TEXT", "estado_objetivo": "TEXT", "resultado": "TEXT", "http_status": "INTEGER DEFAULT 0",
		"raspberry_ip": "TEXT", "response_body": "TEXT", "error": "TEXT", "fecha_evento": "TEXT DEFAULT (datetime('now','localtime'))",
		"actor": "TEXT", "origen": "TEXT", "metadata_json": "TEXT",
	}
	for col, def := range eventCols {
		if err := ensureColumnIfMissing(dbConn, "empresa_control_electrico_eventos", col, def); err != nil {
			return err
		}
	}

	empresaControlElectricoSchemaReady = true
	return nil
}

// GetEmpresaControlElectricoConfig obtiene configuracion o entrega defaults no persistidos.
func GetEmpresaControlElectricoConfig(dbConn *sql.DB, empresaID int64, includeToken bool) (*EmpresaControlElectricoConfig, error) {
	if empresaID <= 0 {
		return nil, errors.New("empresa_id invalido")
	}
	cfg := defaultEmpresaControlElectricoConfig(empresaID)
	row := queryRowSQLCompat(dbConn, `SELECT id, empresa_id, COALESCE(habilitado,0), COALESCE(raspberry_ip,''), COALESCE(raspberry_port,8081), COALESCE(api_path,''), COALESCE(api_token,''), COALESCE(timeout_ms,2500), COALESCE(auto_sync_estaciones,1), COALESCE(fail_safe_on_error,0), COALESCE(fecha_creacion,''), COALESCE(fecha_actualizacion,''), COALESCE(usuario_creador,''), COALESCE(estado,'activo'), COALESCE(observaciones,'') FROM empresa_control_electrico_config WHERE empresa_id = ? LIMIT 1`, empresaID)
	var habilitado, autoSync, failSafe int
	var token string
	if err := row.Scan(&cfg.ID, &cfg.EmpresaID, &habilitado, &cfg.RaspberryIP, &cfg.RaspberryPort, &cfg.APIPath, &token, &cfg.TimeoutMS, &autoSync, &failSafe, &cfg.FechaCreacion, &cfg.FechaActualizacion, &cfg.UsuarioCreador, &cfg.Estado, &cfg.Observaciones); err != nil {
		if err == sql.ErrNoRows {
			return cfg, nil
		}
		return nil, err
	}
	cfg.Habilitado = habilitado == 1
	cfg.AutoSyncEstaciones = autoSync == 1
	cfg.FailSafeOnError = failSafe == 1
	cfg.APITokenConfigured = strings.TrimSpace(token) != ""
	if includeToken {
		cfg.APIToken = token
	}
	normalizeEmpresaControlElectricoConfig(cfg)
	return cfg, nil
}

// EnsureEmpresaControlElectricoPrimaryRaspberry crea el nodo principal desde la configuracion legacy.
func EnsureEmpresaControlElectricoPrimaryRaspberry(dbConn *sql.DB, cfg *EmpresaControlElectricoConfig) (*EmpresaControlElectricoRaspberry, error) {
	if cfg == nil || cfg.EmpresaID <= 0 || strings.TrimSpace(cfg.RaspberryIP) == "" {
		return nil, nil
	}
	row := queryRowSQLCompat(dbConn, `SELECT id FROM empresa_control_electrico_raspberry_pis WHERE empresa_id = ? AND codigo = 'principal' LIMIT 1`, cfg.EmpresaID)
	var existingID int64
	if err := row.Scan(&existingID); err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	item := &EmpresaControlElectricoRaspberry{
		ID:             existingID,
		EmpresaID:      cfg.EmpresaID,
		Codigo:         "principal",
		Nombre:         "Raspberry principal",
		RaspberryIP:    cfg.RaspberryIP,
		RaspberryPort:  cfg.RaspberryPort,
		APIPath:        cfg.APIPath,
		APIToken:       cfg.APIToken,
		TimeoutMS:      cfg.TimeoutMS,
		UsuarioCreador: cfg.UsuarioCreador,
		Estado:         "activo",
		Observaciones:  "Nodo principal sincronizado desde la configuracion global",
	}
	id, err := UpsertEmpresaControlElectricoRaspberry(dbConn, item)
	if err != nil {
		return nil, err
	}
	return GetEmpresaControlElectricoRaspberryByID(dbConn, cfg.EmpresaID, id, false)
}

// ListEmpresaControlElectricoRaspberry lista controladores GPIO configurados.
func ListEmpresaControlElectricoRaspberry(dbConn *sql.DB, empresaID int64, includeInactive bool) ([]EmpresaControlElectricoRaspberry, error) {
	if empresaID <= 0 {
		return nil, errors.New("empresa_id invalido")
	}
	q := `SELECT id, empresa_id, COALESCE(codigo,''), COALESCE(nombre,''), COALESCE(raspberry_ip,''), COALESCE(raspberry_port,8081), COALESCE(api_path,''), COALESCE(api_token,''), COALESCE(timeout_ms,2500), COALESCE(fecha_creacion,''), COALESCE(fecha_actualizacion,''), COALESCE(usuario_creador,''), COALESCE(estado,'activo'), COALESCE(observaciones,'') FROM empresa_control_electrico_raspberry_pis WHERE empresa_id = ?`
	if !includeInactive {
		q += " AND LOWER(COALESCE(estado,'activo')) = 'activo'"
	}
	q += " ORDER BY CASE LOWER(COALESCE(codigo,'')) WHEN 'principal' THEN 0 ELSE 1 END, nombre, id"
	rows, err := querySQLCompat(dbConn, q, empresaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaControlElectricoRaspberry{}
	for rows.Next() {
		var item EmpresaControlElectricoRaspberry
		var token string
		if err := rows.Scan(&item.ID, &item.EmpresaID, &item.Codigo, &item.Nombre, &item.RaspberryIP, &item.RaspberryPort, &item.APIPath, &token, &item.TimeoutMS, &item.FechaCreacion, &item.FechaActualizacion, &item.UsuarioCreador, &item.Estado, &item.Observaciones); err != nil {
			return nil, err
		}
		item.APITokenConfigured = strings.TrimSpace(token) != ""
		normalizeEmpresaControlElectricoRaspberry(&item)
		out = append(out, item)
	}
	return out, rows.Err()
}

// GetEmpresaControlElectricoRaspberryByID obtiene un controlador GPIO por id.
func GetEmpresaControlElectricoRaspberryByID(dbConn *sql.DB, empresaID, raspberryID int64, includeToken bool) (*EmpresaControlElectricoRaspberry, error) {
	if empresaID <= 0 || raspberryID <= 0 {
		return nil, errors.New("empresa_id y raspberry_id son obligatorios")
	}
	row := queryRowSQLCompat(dbConn, `SELECT id, empresa_id, COALESCE(codigo,''), COALESCE(nombre,''), COALESCE(raspberry_ip,''), COALESCE(raspberry_port,8081), COALESCE(api_path,''), COALESCE(api_token,''), COALESCE(timeout_ms,2500), COALESCE(fecha_creacion,''), COALESCE(fecha_actualizacion,''), COALESCE(usuario_creador,''), COALESCE(estado,'activo'), COALESCE(observaciones,'') FROM empresa_control_electrico_raspberry_pis WHERE empresa_id = ? AND id = ? AND LOWER(COALESCE(estado,'activo')) = 'activo' LIMIT 1`, empresaID, raspberryID)
	var item EmpresaControlElectricoRaspberry
	var token string
	if err := row.Scan(&item.ID, &item.EmpresaID, &item.Codigo, &item.Nombre, &item.RaspberryIP, &item.RaspberryPort, &item.APIPath, &token, &item.TimeoutMS, &item.FechaCreacion, &item.FechaActualizacion, &item.UsuarioCreador, &item.Estado, &item.Observaciones); err != nil {
		return nil, err
	}
	item.APITokenConfigured = strings.TrimSpace(token) != ""
	if includeToken {
		item.APIToken = token
	}
	normalizeEmpresaControlElectricoRaspberry(&item)
	return &item, nil
}

// UpsertEmpresaControlElectricoRaspberry crea o actualiza un controlador GPIO.
func UpsertEmpresaControlElectricoRaspberry(dbConn *sql.DB, item *EmpresaControlElectricoRaspberry) (int64, error) {
	if item == nil || item.EmpresaID <= 0 {
		return 0, errors.New("empresa_id invalido")
	}
	normalizeEmpresaControlElectricoRaspberry(item)
	if strings.TrimSpace(item.RaspberryIP) == "" {
		return 0, errors.New("raspberry_ip es obligatorio")
	}
	var existingID int64
	var existingToken string
	var err error
	if item.ID > 0 {
		err = queryRowSQLCompat(dbConn, `SELECT id, COALESCE(api_token,'') FROM empresa_control_electrico_raspberry_pis WHERE empresa_id = ? AND id = ? LIMIT 1`, item.EmpresaID, item.ID).Scan(&existingID, &existingToken)
	} else {
		err = queryRowSQLCompat(dbConn, `SELECT id, COALESCE(api_token,'') FROM empresa_control_electrico_raspberry_pis WHERE empresa_id = ? AND codigo = ? LIMIT 1`, item.EmpresaID, item.Codigo).Scan(&existingID, &existingToken)
	}
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}
	token := existingToken
	if strings.TrimSpace(item.APIToken) != "" {
		token = strings.TrimSpace(item.APIToken)
	}
	if existingID > 0 {
		_, err := execSQLCompat(dbConn, `UPDATE empresa_control_electrico_raspberry_pis SET codigo=?, nombre=?, raspberry_ip=?, raspberry_port=?, api_path=?, api_token=?, timeout_ms=?, fecha_actualizacion=datetime('now','localtime'), usuario_creador=?, estado=?, observaciones=? WHERE empresa_id=? AND id=?`,
			item.Codigo, item.Nombre, item.RaspberryIP, item.RaspberryPort, item.APIPath, token, item.TimeoutMS, strings.TrimSpace(item.UsuarioCreador), item.Estado, strings.TrimSpace(item.Observaciones), item.EmpresaID, existingID)
		return existingID, err
	}
	return insertSQLCompat(dbConn, `INSERT INTO empresa_control_electrico_raspberry_pis (empresa_id, codigo, nombre, raspberry_ip, raspberry_port, api_path, api_token, timeout_ms, fecha_creacion, fecha_actualizacion, usuario_creador, estado, observaciones) VALUES (?, ?, ?, ?, ?, ?, ?, ?, datetime('now','localtime'), datetime('now','localtime'), ?, ?, ?)`,
		item.EmpresaID, item.Codigo, item.Nombre, item.RaspberryIP, item.RaspberryPort, item.APIPath, token, item.TimeoutMS, strings.TrimSpace(item.UsuarioCreador), item.Estado, strings.TrimSpace(item.Observaciones))
}

// SetEmpresaControlElectricoRaspberryEstado activa o desactiva un controlador GPIO.
func SetEmpresaControlElectricoRaspberryEstado(dbConn *sql.DB, empresaID, raspberryID int64, estado string) error {
	if empresaID <= 0 || raspberryID <= 0 {
		return errors.New("empresa_id y raspberry_id son obligatorios")
	}
	_, err := execSQLCompat(dbConn, `UPDATE empresa_control_electrico_raspberry_pis SET estado=?, fecha_actualizacion=datetime('now','localtime') WHERE empresa_id=? AND id=?`, normalizeControlElectricoEstado(estado), empresaID, raspberryID)
	return err
}

// UpsertEmpresaControlElectricoConfig crea o actualiza configuracion.
func UpsertEmpresaControlElectricoConfig(dbConn *sql.DB, cfg *EmpresaControlElectricoConfig) (int64, error) {
	if cfg == nil || cfg.EmpresaID <= 0 {
		return 0, errors.New("empresa_id invalido")
	}
	normalizeEmpresaControlElectricoConfig(cfg)
	var existingID int64
	var existingToken string
	err := queryRowSQLCompat(dbConn, `SELECT id, COALESCE(api_token,'') FROM empresa_control_electrico_config WHERE empresa_id = ? LIMIT 1`, cfg.EmpresaID).Scan(&existingID, &existingToken)
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}
	token := existingToken
	if strings.TrimSpace(cfg.APIToken) != "" {
		token = strings.TrimSpace(cfg.APIToken)
	}
	if existingID > 0 {
		_, err := execSQLCompat(dbConn, `UPDATE empresa_control_electrico_config SET habilitado=?, raspberry_ip=?, raspberry_port=?, api_path=?, api_token=?, timeout_ms=?, auto_sync_estaciones=?, fail_safe_on_error=?, fecha_actualizacion=datetime('now','localtime'), usuario_creador=?, estado=?, observaciones=? WHERE id=?`,
			boolInt(cfg.Habilitado), cfg.RaspberryIP, cfg.RaspberryPort, cfg.APIPath, token, cfg.TimeoutMS, boolInt(cfg.AutoSyncEstaciones), boolInt(cfg.FailSafeOnError), strings.TrimSpace(cfg.UsuarioCreador), cfg.Estado, strings.TrimSpace(cfg.Observaciones), existingID)
		return existingID, err
	}
	return insertSQLCompat(dbConn, `INSERT INTO empresa_control_electrico_config (empresa_id, habilitado, raspberry_ip, raspberry_port, api_path, api_token, timeout_ms, auto_sync_estaciones, fail_safe_on_error, fecha_creacion, fecha_actualizacion, usuario_creador, estado, observaciones) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now','localtime'), datetime('now','localtime'), ?, ?, ?)`,
		cfg.EmpresaID, boolInt(cfg.Habilitado), cfg.RaspberryIP, cfg.RaspberryPort, cfg.APIPath, token, cfg.TimeoutMS, boolInt(cfg.AutoSyncEstaciones), boolInt(cfg.FailSafeOnError), strings.TrimSpace(cfg.UsuarioCreador), cfg.Estado, strings.TrimSpace(cfg.Observaciones))
}

// ListEmpresaControlElectricoReles lista relays configurados.
func ListEmpresaControlElectricoReles(dbConn *sql.DB, empresaID int64, includeInactive bool) ([]EmpresaControlElectricoRele, error) {
	if empresaID <= 0 {
		return nil, errors.New("empresa_id invalido")
	}
	q := `SELECT r.id, r.empresa_id, COALESCE(r.raspberry_id,0), COALESCE(rp.codigo,''), COALESCE(rp.nombre,''), COALESCE(rp.raspberry_ip,''), r.estacion_id, COALESCE(r.estacion_codigo,''), COALESCE(r.estacion_nombre,''), COALESCE(r.salida_codigo,'principal'), COALESCE(r.tipo_carga,'luces'), COALESCE(r.gpio_pin,0), COALESCE(r.relay_name,''), COALESCE(r.active_high,1), COALESCE(r.pulso_ms,0), COALESCE(r.modo,'seguimiento_estacion'), COALESCE(r.ultimo_estado,'desconocido'), COALESCE(r.ultimo_comando,''), COALESCE(r.ultimo_error,''), COALESCE(r.ultima_sincronizacion,''), COALESCE(r.fecha_creacion,''), COALESCE(r.fecha_actualizacion,''), COALESCE(r.usuario_creador,''), COALESCE(r.estado,'activo'), COALESCE(r.observaciones,'') FROM empresa_control_electrico_reles r LEFT JOIN empresa_control_electrico_raspberry_pis rp ON rp.empresa_id = r.empresa_id AND rp.id = r.raspberry_id WHERE r.empresa_id = ?`
	if !includeInactive {
		q += " AND LOWER(COALESCE(r.estado,'activo')) = 'activo'"
	}
	q += " ORDER BY r.estacion_id, r.salida_codigo"
	rows, err := querySQLCompat(dbConn, q, empresaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaControlElectricoRele{}
	for rows.Next() {
		var item EmpresaControlElectricoRele
		var activeHigh int
		if err := rows.Scan(&item.ID, &item.EmpresaID, &item.RaspberryID, &item.RaspberryCodigo, &item.RaspberryNombre, &item.RaspberryIP, &item.EstacionID, &item.EstacionCodigo, &item.EstacionNombre, &item.SalidaCodigo, &item.TipoCarga, &item.GPIOPin, &item.RelayName, &activeHigh, &item.PulsoMS, &item.Modo, &item.UltimoEstado, &item.UltimoComando, &item.UltimoError, &item.UltimaSincronizacion, &item.FechaCreacion, &item.FechaActualizacion, &item.UsuarioCreador, &item.Estado, &item.Observaciones); err != nil {
			return nil, err
		}
		item.ActiveHigh = activeHigh == 1
		normalizeEmpresaControlElectricoRele(&item)
		out = append(out, item)
	}
	return out, rows.Err()
}

// GetEmpresaControlElectricoReleByEstacion obtiene el relay de una estacion.
func GetEmpresaControlElectricoReleByEstacion(dbConn *sql.DB, empresaID, estacionID int64) (*EmpresaControlElectricoRele, error) {
	reles, err := ListEmpresaControlElectricoRelesByEstacion(dbConn, empresaID, estacionID, false)
	if err != nil {
		return nil, err
	}
	if len(reles) == 0 {
		return nil, sql.ErrNoRows
	}
	return &reles[0], nil
}

// GetEmpresaControlElectricoReleByID obtiene una salida electrica puntual.
func GetEmpresaControlElectricoReleByID(dbConn *sql.DB, empresaID, releID int64) (*EmpresaControlElectricoRele, error) {
	if empresaID <= 0 || releID <= 0 {
		return nil, errors.New("empresa_id y rele_id son obligatorios")
	}
	row := queryRowSQLCompat(dbConn, `SELECT r.id, r.empresa_id, COALESCE(r.raspberry_id,0), COALESCE(rp.codigo,''), COALESCE(rp.nombre,''), COALESCE(rp.raspberry_ip,''), r.estacion_id, COALESCE(r.estacion_codigo,''), COALESCE(r.estacion_nombre,''), COALESCE(r.salida_codigo,'principal'), COALESCE(r.tipo_carga,'luces'), COALESCE(r.gpio_pin,0), COALESCE(r.relay_name,''), COALESCE(r.active_high,1), COALESCE(r.pulso_ms,0), COALESCE(r.modo,'seguimiento_estacion'), COALESCE(r.ultimo_estado,'desconocido'), COALESCE(r.ultimo_comando,''), COALESCE(r.ultimo_error,''), COALESCE(r.ultima_sincronizacion,''), COALESCE(r.fecha_creacion,''), COALESCE(r.fecha_actualizacion,''), COALESCE(r.usuario_creador,''), COALESCE(r.estado,'activo'), COALESCE(r.observaciones,'') FROM empresa_control_electrico_reles r LEFT JOIN empresa_control_electrico_raspberry_pis rp ON rp.empresa_id = r.empresa_id AND rp.id = r.raspberry_id WHERE r.empresa_id = ? AND r.id = ? AND LOWER(COALESCE(r.estado,'activo')) = 'activo' LIMIT 1`, empresaID, releID)
	var item EmpresaControlElectricoRele
	var activeHigh int
	if err := row.Scan(&item.ID, &item.EmpresaID, &item.RaspberryID, &item.RaspberryCodigo, &item.RaspberryNombre, &item.RaspberryIP, &item.EstacionID, &item.EstacionCodigo, &item.EstacionNombre, &item.SalidaCodigo, &item.TipoCarga, &item.GPIOPin, &item.RelayName, &activeHigh, &item.PulsoMS, &item.Modo, &item.UltimoEstado, &item.UltimoComando, &item.UltimoError, &item.UltimaSincronizacion, &item.FechaCreacion, &item.FechaActualizacion, &item.UsuarioCreador, &item.Estado, &item.Observaciones); err != nil {
		return nil, err
	}
	item.ActiveHigh = activeHigh == 1
	normalizeEmpresaControlElectricoRele(&item)
	return &item, nil
}

// ListEmpresaControlElectricoRelesByEstacion lista todas las salidas de una estacion.
func ListEmpresaControlElectricoRelesByEstacion(dbConn *sql.DB, empresaID, estacionID int64, includeInactive bool) ([]EmpresaControlElectricoRele, error) {
	if empresaID <= 0 || estacionID <= 0 {
		return nil, errors.New("empresa_id y estacion_id son obligatorios")
	}
	q := `SELECT r.id, r.empresa_id, COALESCE(r.raspberry_id,0), COALESCE(rp.codigo,''), COALESCE(rp.nombre,''), COALESCE(rp.raspberry_ip,''), r.estacion_id, COALESCE(r.estacion_codigo,''), COALESCE(r.estacion_nombre,''), COALESCE(r.salida_codigo,'principal'), COALESCE(r.tipo_carga,'luces'), COALESCE(r.gpio_pin,0), COALESCE(r.relay_name,''), COALESCE(r.active_high,1), COALESCE(r.pulso_ms,0), COALESCE(r.modo,'seguimiento_estacion'), COALESCE(r.ultimo_estado,'desconocido'), COALESCE(r.ultimo_comando,''), COALESCE(r.ultimo_error,''), COALESCE(r.ultima_sincronizacion,''), COALESCE(r.fecha_creacion,''), COALESCE(r.fecha_actualizacion,''), COALESCE(r.usuario_creador,''), COALESCE(r.estado,'activo'), COALESCE(r.observaciones,'') FROM empresa_control_electrico_reles r LEFT JOIN empresa_control_electrico_raspberry_pis rp ON rp.empresa_id = r.empresa_id AND rp.id = r.raspberry_id WHERE r.empresa_id = ? AND r.estacion_id = ?`
	if !includeInactive {
		q += " AND LOWER(COALESCE(r.estado,'activo')) = 'activo'"
	}
	q += " ORDER BY CASE COALESCE(r.salida_codigo,'principal') WHEN 'principal' THEN 0 WHEN 'luces' THEN 1 WHEN 'jacuzzi' THEN 2 ELSE 10 END, r.id"
	rows, err := querySQLCompat(dbConn, q, empresaID, estacionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaControlElectricoRele{}
	for rows.Next() {
		var item EmpresaControlElectricoRele
		var activeHigh int
		if err := rows.Scan(&item.ID, &item.EmpresaID, &item.RaspberryID, &item.RaspberryCodigo, &item.RaspberryNombre, &item.RaspberryIP, &item.EstacionID, &item.EstacionCodigo, &item.EstacionNombre, &item.SalidaCodigo, &item.TipoCarga, &item.GPIOPin, &item.RelayName, &activeHigh, &item.PulsoMS, &item.Modo, &item.UltimoEstado, &item.UltimoComando, &item.UltimoError, &item.UltimaSincronizacion, &item.FechaCreacion, &item.FechaActualizacion, &item.UsuarioCreador, &item.Estado, &item.Observaciones); err != nil {
			return nil, err
		}
		item.ActiveHigh = activeHigh == 1
		normalizeEmpresaControlElectricoRele(&item)
		out = append(out, item)
	}
	return out, rows.Err()
}

// UpsertEmpresaControlElectricoRele crea o actualiza un relay por estacion.
func UpsertEmpresaControlElectricoRele(dbConn *sql.DB, item *EmpresaControlElectricoRele) (int64, error) {
	if item == nil || item.EmpresaID <= 0 || item.EstacionID <= 0 {
		return 0, errors.New("empresa_id y estacion_id son obligatorios")
	}
	normalizeEmpresaControlElectricoRele(item)
	if item.GPIOPin < 0 || item.GPIOPin > 27 {
		return 0, fmt.Errorf("gpio_pin debe estar entre 0 y 27")
	}
	if item.RaspberryID > 0 {
		if _, err := GetEmpresaControlElectricoRaspberryByID(dbConn, item.EmpresaID, item.RaspberryID, false); err != nil {
			if err == sql.ErrNoRows {
				return 0, fmt.Errorf("raspberry_id no pertenece a la empresa o esta inactiva")
			}
			return 0, err
		}
	}
	var existingID int64
	var err error
	if item.ID > 0 {
		err = queryRowSQLCompat(dbConn, `SELECT id FROM empresa_control_electrico_reles WHERE empresa_id = ? AND id = ? LIMIT 1`, item.EmpresaID, item.ID).Scan(&existingID)
	} else {
		err = queryRowSQLCompat(dbConn, `SELECT id FROM empresa_control_electrico_reles WHERE empresa_id = ? AND estacion_id = ? AND salida_codigo = ? LIMIT 1`, item.EmpresaID, item.EstacionID, item.SalidaCodigo).Scan(&existingID)
	}
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}
	if existingID > 0 {
		_, err := execSQLCompat(dbConn, `UPDATE empresa_control_electrico_reles SET raspberry_id=NULLIF(?,0), estacion_codigo=?, estacion_nombre=?, salida_codigo=?, tipo_carga=?, gpio_pin=?, relay_name=?, active_high=?, pulso_ms=?, modo=?, fecha_actualizacion=datetime('now','localtime'), usuario_creador=?, estado=?, observaciones=? WHERE id=?`,
			item.RaspberryID, item.EstacionCodigo, item.EstacionNombre, item.SalidaCodigo, item.TipoCarga, item.GPIOPin, item.RelayName, boolInt(item.ActiveHigh), item.PulsoMS, item.Modo, strings.TrimSpace(item.UsuarioCreador), item.Estado, strings.TrimSpace(item.Observaciones), existingID)
		return existingID, err
	}
	return insertSQLCompat(dbConn, `INSERT INTO empresa_control_electrico_reles (empresa_id, raspberry_id, estacion_id, estacion_codigo, estacion_nombre, salida_codigo, tipo_carga, gpio_pin, relay_name, active_high, pulso_ms, modo, ultimo_estado, fecha_creacion, fecha_actualizacion, usuario_creador, estado, observaciones) VALUES (?, NULLIF(?,0), ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'desconocido', datetime('now','localtime'), datetime('now','localtime'), ?, ?, ?)`,
		item.EmpresaID, item.RaspberryID, item.EstacionID, item.EstacionCodigo, item.EstacionNombre, item.SalidaCodigo, item.TipoCarga, item.GPIOPin, item.RelayName, boolInt(item.ActiveHigh), item.PulsoMS, item.Modo, strings.TrimSpace(item.UsuarioCreador), item.Estado, strings.TrimSpace(item.Observaciones))
}

// SetEmpresaControlElectricoReleEstado cambia estado logico del mapeo.
func SetEmpresaControlElectricoReleEstado(dbConn *sql.DB, empresaID, releID int64, estado string) error {
	if empresaID <= 0 || releID <= 0 {
		return errors.New("empresa_id y rele_id son obligatorios")
	}
	_, err := execSQLCompat(dbConn, `UPDATE empresa_control_electrico_reles SET estado=?, fecha_actualizacion=datetime('now','localtime') WHERE empresa_id=? AND id=?`, normalizeControlElectricoEstado(estado), empresaID, releID)
	return err
}

// UpdateEmpresaControlElectricoReleRuntime actualiza el ultimo resultado del relay.
func UpdateEmpresaControlElectricoReleRuntime(dbConn *sql.DB, empresaID, releID int64, ultimoEstado, ultimoComando, ultimoError string) error {
	if empresaID <= 0 || releID <= 0 {
		return nil
	}
	_, err := execSQLCompat(dbConn, `UPDATE empresa_control_electrico_reles SET ultimo_estado=?, ultimo_comando=?, ultimo_error=?, ultima_sincronizacion=datetime('now','localtime'), fecha_actualizacion=datetime('now','localtime') WHERE empresa_id=? AND id=?`, strings.TrimSpace(ultimoEstado), strings.TrimSpace(ultimoComando), strings.TrimSpace(ultimoError), empresaID, releID)
	return err
}

// InsertEmpresaControlElectricoEvento registra una accion electrica.
func InsertEmpresaControlElectricoEvento(dbConn *sql.DB, ev EmpresaControlElectricoEvento) (int64, error) {
	if ev.EmpresaID <= 0 {
		return 0, errors.New("empresa_id invalido")
	}
	return insertSQLCompat(dbConn, `INSERT INTO empresa_control_electrico_eventos (empresa_id, estacion_id, rele_id, raspberry_id, gpio_pin, comando, estado_objetivo, resultado, http_status, raspberry_ip, response_body, error, fecha_evento, actor, origen, metadata_json) VALUES (?, NULLIF(?,0), NULLIF(?,0), NULLIF(?,0), ?, ?, ?, ?, ?, ?, ?, ?, datetime('now','localtime'), ?, ?, ?)`,
		ev.EmpresaID, ev.EstacionID, ev.ReleID, ev.RaspberryID, ev.GPIOPin, strings.TrimSpace(ev.Comando), strings.TrimSpace(ev.EstadoObjetivo), strings.TrimSpace(ev.Resultado), ev.HTTPStatus, strings.TrimSpace(ev.RaspberryIP), truncateControlElectricoText(ev.ResponseBody, 1200), truncateControlElectricoText(ev.Error, 800), strings.TrimSpace(ev.Actor), strings.TrimSpace(ev.Origen), strings.TrimSpace(ev.MetadataJSON))
}

// ListEmpresaControlElectricoEventos lista los eventos recientes.
func ListEmpresaControlElectricoEventos(dbConn *sql.DB, empresaID int64, limit int) ([]EmpresaControlElectricoEvento, error) {
	if empresaID <= 0 {
		return nil, errors.New("empresa_id invalido")
	}
	if limit <= 0 {
		limit = 30
	}
	if limit > 200 {
		limit = 200
	}
	rows, err := querySQLCompat(dbConn, `SELECT id, empresa_id, COALESCE(estacion_id,0), COALESCE(rele_id,0), COALESCE(raspberry_id,0), COALESCE(gpio_pin,0), COALESCE(comando,''), COALESCE(estado_objetivo,''), COALESCE(resultado,''), COALESCE(http_status,0), COALESCE(raspberry_ip,''), COALESCE(response_body,''), COALESCE(error,''), COALESCE(fecha_evento,''), COALESCE(actor,''), COALESCE(origen,''), COALESCE(metadata_json,'') FROM empresa_control_electrico_eventos WHERE empresa_id=? ORDER BY id DESC LIMIT ?`, empresaID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaControlElectricoEvento{}
	for rows.Next() {
		var ev EmpresaControlElectricoEvento
		if err := rows.Scan(&ev.ID, &ev.EmpresaID, &ev.EstacionID, &ev.ReleID, &ev.RaspberryID, &ev.GPIOPin, &ev.Comando, &ev.EstadoObjetivo, &ev.Resultado, &ev.HTTPStatus, &ev.RaspberryIP, &ev.ResponseBody, &ev.Error, &ev.FechaEvento, &ev.Actor, &ev.Origen, &ev.MetadataJSON); err != nil {
			return nil, err
		}
		out = append(out, ev)
	}
	return out, rows.Err()
}

// ListEmpresaControlElectricoEstaciones arma el tablero electrico desde las estaciones/carritos.
func ListEmpresaControlElectricoEstaciones(dbConn *sql.DB, empresaID int64) ([]EmpresaControlElectricoEstacion, error) {
	carritos, err := GetCarritosCompraByEmpresa(dbConn, empresaID, true, "")
	if err != nil {
		return nil, err
	}
	reles, err := ListEmpresaControlElectricoReles(dbConn, empresaID, true)
	if err != nil {
		return nil, err
	}
	relesByStation := map[int64]*EmpresaControlElectricoRele{}
	for i := range reles {
		rele := reles[i]
		if _, exists := relesByStation[rele.EstacionID]; !exists {
			relesByStation[rele.EstacionID] = &rele
		}
	}
	out := []EmpresaControlElectricoEstacion{}
	seen := map[int64]bool{}
	for _, carrito := range carritos {
		estacionID, codigo, nombre := ResolveCarritoStationIdentity(&carrito)
		if estacionID <= 0 || seen[estacionID] {
			continue
		}
		seen[estacionID] = true
		estado := strings.ToLower(strings.TrimSpace(carrito.Estado))
		estadoCarrito := strings.ToLower(strings.TrimSpace(carrito.EstadoCarrito))
		activa := estado == "activo" && estadoCarrito != "cerrado" && strings.TrimSpace(carrito.PagadoEn) == ""
		out = append(out, EmpresaControlElectricoEstacion{
			EstacionID:     estacionID,
			EstacionCodigo: codigo,
			EstacionNombre: nombre,
			CarritoID:      carrito.ID,
			Activa:         activa,
			Estado:         carrito.Estado,
			EstadoCarrito:  carrito.EstadoCarrito,
			EstadoVenta:    carrito.EstadoVenta,
			ActivadoEn:     carrito.ActivadoEn,
			PagadoEn:       carrito.PagadoEn,
			Rele:           relesByStation[estacionID],
		})
	}
	return out, nil
}

func defaultEmpresaControlElectricoConfig(empresaID int64) *EmpresaControlElectricoConfig {
	return &EmpresaControlElectricoConfig{
		EmpresaID:          empresaID,
		RaspberryPort:      DefaultControlElectricoPort,
		APIPath:            DefaultControlElectricoAPIPath,
		TimeoutMS:          DefaultControlElectricoTimeoutMS,
		AutoSyncEstaciones: true,
		Estado:             "activo",
	}
}

func normalizeEmpresaControlElectricoConfig(cfg *EmpresaControlElectricoConfig) {
	if cfg == nil {
		return
	}
	cfg.RaspberryIP = strings.TrimSpace(cfg.RaspberryIP)
	if cfg.RaspberryPort <= 0 {
		cfg.RaspberryPort = DefaultControlElectricoPort
	}
	cfg.APIPath = strings.TrimSpace(cfg.APIPath)
	if cfg.APIPath == "" {
		cfg.APIPath = DefaultControlElectricoAPIPath
	}
	if !strings.HasPrefix(cfg.APIPath, "/") {
		cfg.APIPath = "/" + cfg.APIPath
	}
	if cfg.TimeoutMS <= 0 {
		cfg.TimeoutMS = DefaultControlElectricoTimeoutMS
	}
	if cfg.TimeoutMS < 500 {
		cfg.TimeoutMS = 500
	}
	if cfg.TimeoutMS > 15000 {
		cfg.TimeoutMS = 15000
	}
	cfg.Estado = normalizeControlElectricoEstado(cfg.Estado)
}

func normalizeEmpresaControlElectricoRaspberry(item *EmpresaControlElectricoRaspberry) {
	if item == nil {
		return
	}
	item.Codigo = normalizeControlElectricoSalidaCodigo(item.Codigo, "", item.Nombre)
	if item.Codigo == "principal" && strings.TrimSpace(item.Nombre) == "" {
		item.Nombre = "Raspberry principal"
	}
	if item.Codigo == "" {
		item.Codigo = "raspberry"
	}
	item.Nombre = strings.TrimSpace(item.Nombre)
	if item.Nombre == "" {
		item.Nombre = strings.ReplaceAll(item.Codigo, "_", " ")
	}
	item.RaspberryIP = strings.TrimSpace(item.RaspberryIP)
	if item.RaspberryPort <= 0 {
		item.RaspberryPort = DefaultControlElectricoPort
	}
	if item.RaspberryPort > 65535 {
		item.RaspberryPort = DefaultControlElectricoPort
	}
	item.APIPath = strings.TrimSpace(item.APIPath)
	if item.APIPath == "" {
		item.APIPath = DefaultControlElectricoAPIPath
	}
	if !strings.HasPrefix(item.APIPath, "/") {
		item.APIPath = "/" + item.APIPath
	}
	if item.TimeoutMS <= 0 {
		item.TimeoutMS = DefaultControlElectricoTimeoutMS
	}
	if item.TimeoutMS < 500 {
		item.TimeoutMS = 500
	}
	if item.TimeoutMS > 15000 {
		item.TimeoutMS = 15000
	}
	item.Estado = normalizeControlElectricoEstado(item.Estado)
}

func normalizeEmpresaControlElectricoRele(item *EmpresaControlElectricoRele) {
	if item == nil {
		return
	}
	item.RaspberryCodigo = strings.TrimSpace(item.RaspberryCodigo)
	item.RaspberryNombre = strings.TrimSpace(item.RaspberryNombre)
	item.RaspberryIP = strings.TrimSpace(item.RaspberryIP)
	item.EstacionCodigo = strings.TrimSpace(item.EstacionCodigo)
	item.EstacionNombre = strings.TrimSpace(item.EstacionNombre)
	item.SalidaCodigo = normalizeControlElectricoSalidaCodigo(item.SalidaCodigo, item.TipoCarga, item.RelayName)
	item.TipoCarga = strings.ToLower(strings.TrimSpace(item.TipoCarga))
	if item.TipoCarga == "" {
		item.TipoCarga = item.SalidaCodigo
	}
	item.RelayName = strings.TrimSpace(item.RelayName)
	if item.RelayName == "" {
		item.RelayName = fmt.Sprintf("Rele estacion %d", item.EstacionID)
	}
	item.Modo = strings.ToLower(strings.TrimSpace(item.Modo))
	if item.Modo == "" {
		item.Modo = "seguimiento_estacion"
	}
	if item.PulsoMS < 0 {
		item.PulsoMS = 0
	}
	item.Estado = normalizeControlElectricoEstado(item.Estado)
}

func normalizeControlElectricoEstado(raw string) string {
	value := strings.ToLower(strings.TrimSpace(raw))
	if value == "" {
		return "activo"
	}
	return value
}

func normalizeControlElectricoSalidaCodigo(raw, tipoCarga, relayName string) string {
	value := strings.ToLower(strings.TrimSpace(raw))
	if value == "" {
		value = strings.ToLower(strings.TrimSpace(tipoCarga))
	}
	if value == "" {
		value = strings.ToLower(strings.TrimSpace(relayName))
	}
	if value == "" {
		value = "principal"
	}
	replacer := strings.NewReplacer("á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u", "ñ", "n", " ", "_", "-", "_")
	value = replacer.Replace(value)
	clean := make([]rune, 0, len(value))
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			clean = append(clean, r)
		}
	}
	value = strings.Trim(string(clean), "_")
	if value == "" {
		return "principal"
	}
	if len(value) > 40 {
		value = strings.Trim(value[:40], "_")
	}
	if value == "" {
		return "principal"
	}
	return value
}

func truncateControlElectricoText(raw string, limit int) string {
	value := strings.TrimSpace(raw)
	if limit <= 0 || len(value) <= limit {
		return value
	}
	return value[:limit]
}
