package db

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
)

// EmpresaSoporteRemotoConfig define configuracion operativa de soporte remoto por empresa.
type EmpresaSoporteRemotoConfig struct {
	ID                         int64  `json:"id"`
	EmpresaID                  int64  `json:"empresa_id"`
	Habilitado                 bool   `json:"habilitado"`
	ProveedorPreferido         string `json:"proveedor_preferido"`
	ModoOperacion              string `json:"modo_operacion"`
	RequiereAprobacionOperador bool   `json:"requiere_aprobacion_operador"`
	AutoCerrarMinutos          int    `json:"auto_cerrar_minutos"`
	FechaCreacion              string `json:"fecha_creacion,omitempty"`
	FechaActualizacion         string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador             string `json:"usuario_creador,omitempty"`
	Estado                     string `json:"estado,omitempty"`
	Observaciones              string `json:"observaciones,omitempty"`
}

// EmpresaSoporteRemotoDispositivo representa un equipo remoto registrado por empresa.
type EmpresaSoporteRemotoDispositivo struct {
	ID                 int64  `json:"id"`
	EmpresaID          int64  `json:"empresa_id"`
	CodigoDispositivo  string `json:"codigo_dispositivo"`
	NombreEquipo       string `json:"nombre_equipo"`
	AliasOperativo     string `json:"alias_operativo,omitempty"`
	Ubicacion          string `json:"ubicacion,omitempty"`
	SistemaOperativo   string `json:"sistema_operativo,omitempty"`
	AgenteVersion      string `json:"agente_version,omitempty"`
	StreamURL          string `json:"stream_url"`
	EstadoConexion     string `json:"estado_conexion"`
	UltimoHeartbeat    string `json:"ultimo_heartbeat,omitempty"`
	AccesoPINHash      string `json:"-"`
	FechaCreacion      string `json:"fecha_creacion,omitempty"`
	FechaActualizacion string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string `json:"usuario_creador,omitempty"`
	Estado             string `json:"estado,omitempty"`
	Observaciones      string `json:"observaciones,omitempty"`
}

// EmpresaSoporteRemotoSession representa una sesion de visualizacion remota por empresa.
type EmpresaSoporteRemotoSession struct {
	ID                    int64  `json:"id"`
	EmpresaID             int64  `json:"empresa_id"`
	DispositivoID         int64  `json:"dispositivo_id"`
	DispositivoCodigo     string `json:"dispositivo_codigo,omitempty"`
	DispositivoNombre     string `json:"dispositivo_nombre,omitempty"`
	CodigoSesion          string `json:"codigo_sesion"`
	SolicitadaPor         string `json:"solicitada_por,omitempty"`
	OperadorNombre        string `json:"operador_nombre,omitempty"`
	OperadorEmail         string `json:"operador_email,omitempty"`
	Motivo                string `json:"motivo,omitempty"`
	EstadoSesion          string `json:"estado_sesion"`
	URLVisualizacion      string `json:"url_visualizacion,omitempty"`
	TokenVisualizacionRaw string `json:"token_visualizacion,omitempty"`
	IniciadaEn            string `json:"iniciada_en,omitempty"`
	ExpiraEn              string `json:"expira_en,omitempty"`
	FinalizadaEn          string `json:"finalizada_en,omitempty"`
	FechaCreacion         string `json:"fecha_creacion,omitempty"`
	FechaActualizacion    string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador        string `json:"usuario_creador,omitempty"`
	Estado                string `json:"estado,omitempty"`
	Observaciones         string `json:"observaciones,omitempty"`

	tokenVisualizacionHash string
}

// EmpresaSoporteRemotoDispositivoFilter define filtros para listar dispositivos.
type EmpresaSoporteRemotoDispositivoFilter struct {
	IncludeInactive bool
	Q               string
	Limit           int
	Offset          int
}

// EmpresaSoporteRemotoSessionFilter define filtros para listar sesiones.
type EmpresaSoporteRemotoSessionFilter struct {
	IncludeInactive bool
	EstadoSesion    string
	Q               string
	Limit           int
	Offset          int
}

func soporteRemotoNormalizeEstado(raw string) string {
	if strings.EqualFold(strings.TrimSpace(raw), "inactivo") {
		return "inactivo"
	}
	return "activo"
}

func soporteRemotoNormalizeConexion(raw string) string {
	value := strings.ToLower(strings.TrimSpace(raw))
	switch value {
	case "online", "offline", "intermitente":
		return value
	default:
		return "offline"
	}
}

func soporteRemotoNormalizeSesionEstado(raw string) string {
	value := strings.ToLower(strings.TrimSpace(raw))
	switch value {
	case "pendiente", "aprobada", "activa", "finalizada", "rechazada", "expirada":
		return value
	default:
		return "pendiente"
	}
}

func soporteRemotoNormalizeProveedor(raw string) string {
	value := strings.ToLower(strings.TrimSpace(raw))
	switch value {
	case "novnc", "rustdesk_web", "guacamole", "custom_url":
		return value
	default:
		return "novnc"
	}
}

func soporteRemotoNormalizeModo(raw string) string {
	value := strings.ToLower(strings.TrimSpace(raw))
	switch value {
	case "agente_web", "agente_local", "hibrido":
		return value
	default:
		return "agente_web"
	}
}

func soporteRemotoNormalizeAutoCerrar(raw int) int {
	if raw <= 0 {
		return 30
	}
	if raw > 480 {
		return 480
	}
	return raw
}

func soporteRemotoNormalizeLimitOffset(limit, offset int) (int, int) {
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

func soporteRemotoLikePattern(raw string) string {
	value := strings.TrimSpace(raw)
	value = strings.ReplaceAll(value, "!", "!!")
	value = strings.ReplaceAll(value, "%", "!%")
	value = strings.ReplaceAll(value, "_", "!_")
	return "%" + value + "%"
}

func soporteRemotoBoolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

func soporteRemotoHash(v string) string {
	raw := strings.TrimSpace(v)
	if raw == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func soporteRemotoGenerateToken(prefix string) string {
	buf := make([]byte, 18)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("%s-%d", strings.ToUpper(strings.TrimSpace(prefix)), time.Now().UnixNano())
	}
	base := strings.ToUpper(strings.TrimSpace(prefix))
	if base == "" {
		base = "SR"
	}
	return base + "-" + hex.EncodeToString(buf)
}

func soporteRemotoGenerateDeviceCode(nombre string) string {
	base := NormalizeEmpresaPublicSlug(nombre)
	if base == "" || base == "empresa" {
		base = "dispositivo"
	}
	if len(base) > 20 {
		base = base[:20]
	}
	return strings.ToUpper("SRD-" + base + "-" + time.Now().In(time.Local).Format("20060102150405"))
}

func soporteRemotoGenerateSessionCode() string {
	return strings.ToUpper("SRS-" + time.Now().In(time.Local).Format("20060102150405") + "-" + soporteRemotoGenerateToken(""))
}

// EnsureEmpresaSoporteRemotoSchema crea/migra tablas de soporte remoto por empresa.
func EnsureEmpresaSoporteRemotoSchema(dbConn *sql.DB) error {
	if dbConn == nil {
		return errors.New("db connection is nil")
	}

	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_soporte_remoto_configuracion (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL UNIQUE,
			habilitado INTEGER DEFAULT 1,
			proveedor_preferido TEXT DEFAULT 'novnc',
			modo_operacion TEXT DEFAULT 'agente_web',
			requiere_aprobacion_operador INTEGER DEFAULT 1,
			auto_cerrar_minutos INTEGER DEFAULT 30,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_soporte_remoto_configuracion_estado
		ON empresa_soporte_remoto_configuracion(empresa_id, estado);`,
		`CREATE TABLE IF NOT EXISTS empresa_soporte_remoto_dispositivos (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			codigo_dispositivo TEXT NOT NULL,
			nombre_equipo TEXT NOT NULL,
			alias_operativo TEXT,
			ubicacion TEXT,
			sistema_operativo TEXT,
			agente_version TEXT,
			stream_url TEXT NOT NULL,
			estado_conexion TEXT DEFAULT 'offline',
			ultimo_heartbeat TEXT,
			acceso_pin_hash TEXT,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_empresa_soporte_remoto_dispositivos_codigo
		ON empresa_soporte_remoto_dispositivos(empresa_id, codigo_dispositivo);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_soporte_remoto_dispositivos_lookup
		ON empresa_soporte_remoto_dispositivos(empresa_id, estado, estado_conexion, id DESC);`,
		`CREATE TABLE IF NOT EXISTS empresa_soporte_remoto_sesiones (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			dispositivo_id INTEGER NOT NULL,
			codigo_sesion TEXT NOT NULL,
			solicitada_por TEXT,
			operador_nombre TEXT,
			operador_email TEXT,
			motivo TEXT,
			estado_sesion TEXT DEFAULT 'pendiente',
			token_visualizacion_hash TEXT,
			url_visualizacion TEXT,
			iniciada_en TEXT,
			expira_en TEXT,
			finalizada_en TEXT,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_empresa_soporte_remoto_sesiones_codigo
		ON empresa_soporte_remoto_sesiones(empresa_id, codigo_sesion);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_soporte_remoto_sesiones_lookup
		ON empresa_soporte_remoto_sesiones(empresa_id, estado_sesion, fecha_creacion DESC);`,
	}

	for _, stmt := range stmts {
		if _, err := dbConn.Exec(stmt); err != nil {
			return err
		}
	}

	if err := ensureColumnIfMissing(dbConn, "empresa_soporte_remoto_configuracion", "proveedor_preferido", "TEXT DEFAULT 'novnc'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_soporte_remoto_configuracion", "modo_operacion", "TEXT DEFAULT 'agente_web'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_soporte_remoto_dispositivos", "acceso_pin_hash", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_soporte_remoto_sesiones", "token_visualizacion_hash", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_soporte_remoto_sesiones", "url_visualizacion", "TEXT"); err != nil {
		return err
	}

	return nil
}

// GetEmpresaSoporteRemotoConfig consulta configuracion de soporte remoto por empresa.
func GetEmpresaSoporteRemotoConfig(dbConn *sql.DB, empresaID int64) (EmpresaSoporteRemotoConfig, error) {
	if dbConn == nil {
		return EmpresaSoporteRemotoConfig{}, errors.New("db connection is nil")
	}
	if empresaID <= 0 {
		return EmpresaSoporteRemotoConfig{}, errors.New("empresa_id invalido")
	}

	var out EmpresaSoporteRemotoConfig
	var habilitado sql.NullInt64
	var requiereAprobacion sql.NullInt64
	err := dbConn.QueryRow(`SELECT
		id,
		empresa_id,
		COALESCE(habilitado, 1),
		COALESCE(proveedor_preferido, 'novnc'),
		COALESCE(modo_operacion, 'agente_web'),
		COALESCE(requiere_aprobacion_operador, 1),
		COALESCE(auto_cerrar_minutos, 30),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM empresa_soporte_remoto_configuracion
	WHERE empresa_id = ?
	LIMIT 1`, empresaID).Scan(
		&out.ID,
		&out.EmpresaID,
		&habilitado,
		&out.ProveedorPreferido,
		&out.ModoOperacion,
		&requiereAprobacion,
		&out.AutoCerrarMinutos,
		&out.FechaCreacion,
		&out.FechaActualizacion,
		&out.UsuarioCreador,
		&out.Estado,
		&out.Observaciones,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return EmpresaSoporteRemotoConfig{
				EmpresaID:                  empresaID,
				Habilitado:                 true,
				ProveedorPreferido:         "novnc",
				ModoOperacion:              "agente_web",
				RequiereAprobacionOperador: true,
				AutoCerrarMinutos:          30,
				Estado:                     "activo",
			}, nil
		}
		return EmpresaSoporteRemotoConfig{}, err
	}

	out.Habilitado = habilitado.Valid && habilitado.Int64 > 0
	out.RequiereAprobacionOperador = requiereAprobacion.Valid && requiereAprobacion.Int64 > 0
	out.ProveedorPreferido = soporteRemotoNormalizeProveedor(out.ProveedorPreferido)
	out.ModoOperacion = soporteRemotoNormalizeModo(out.ModoOperacion)
	out.AutoCerrarMinutos = soporteRemotoNormalizeAutoCerrar(out.AutoCerrarMinutos)
	out.Estado = soporteRemotoNormalizeEstado(out.Estado)
	return out, nil
}

// UpsertEmpresaSoporteRemotoConfig crea o actualiza la configuracion de soporte remoto por empresa.
func UpsertEmpresaSoporteRemotoConfig(dbConn *sql.DB, cfg EmpresaSoporteRemotoConfig) (int64, error) {
	if dbConn == nil {
		return 0, errors.New("db connection is nil")
	}
	if cfg.EmpresaID <= 0 {
		return 0, errors.New("empresa_id invalido")
	}

	cfg.ProveedorPreferido = soporteRemotoNormalizeProveedor(cfg.ProveedorPreferido)
	cfg.ModoOperacion = soporteRemotoNormalizeModo(cfg.ModoOperacion)
	cfg.AutoCerrarMinutos = soporteRemotoNormalizeAutoCerrar(cfg.AutoCerrarMinutos)
	cfg.Estado = soporteRemotoNormalizeEstado(cfg.Estado)
	if cfg.Estado == "" {
		cfg.Estado = "activo"
	}

	var existingID int64
	err := dbConn.QueryRow(`SELECT id FROM empresa_soporte_remoto_configuracion WHERE empresa_id = ? LIMIT 1`, cfg.EmpresaID).Scan(&existingID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}

	if existingID > 0 {
		_, err = dbConn.Exec(`UPDATE empresa_soporte_remoto_configuracion
			SET habilitado = ?,
				proveedor_preferido = ?,
				modo_operacion = ?,
				requiere_aprobacion_operador = ?,
				auto_cerrar_minutos = ?,
				usuario_creador = ?,
				estado = ?,
				observaciones = ?,
				fecha_actualizacion = datetime('now','localtime')
			WHERE id = ?`,
			soporteRemotoBoolToInt(cfg.Habilitado),
			cfg.ProveedorPreferido,
			cfg.ModoOperacion,
			soporteRemotoBoolToInt(cfg.RequiereAprobacionOperador),
			cfg.AutoCerrarMinutos,
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

	res, err := dbConn.Exec(`INSERT INTO empresa_soporte_remoto_configuracion (
		empresa_id,
		habilitado,
		proveedor_preferido,
		modo_operacion,
		requiere_aprobacion_operador,
		auto_cerrar_minutos,
		usuario_creador,
		estado,
		observaciones
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		cfg.EmpresaID,
		soporteRemotoBoolToInt(cfg.Habilitado),
		cfg.ProveedorPreferido,
		cfg.ModoOperacion,
		soporteRemotoBoolToInt(cfg.RequiereAprobacionOperador),
		cfg.AutoCerrarMinutos,
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

// CreateEmpresaSoporteRemotoDispositivo registra un dispositivo remoto por empresa.
func CreateEmpresaSoporteRemotoDispositivo(dbConn *sql.DB, item EmpresaSoporteRemotoDispositivo, accesoPINPlano string) (int64, error) {
	if dbConn == nil {
		return 0, errors.New("db connection is nil")
	}
	if item.EmpresaID <= 0 {
		return 0, errors.New("empresa_id invalido")
	}
	item.NombreEquipo = strings.TrimSpace(item.NombreEquipo)
	if item.NombreEquipo == "" {
		return 0, errors.New("nombre_equipo es obligatorio")
	}
	item.StreamURL = strings.TrimSpace(item.StreamURL)
	if item.StreamURL == "" {
		return 0, errors.New("stream_url es obligatorio")
	}
	item.CodigoDispositivo = strings.TrimSpace(item.CodigoDispositivo)
	if item.CodigoDispositivo == "" {
		item.CodigoDispositivo = soporteRemotoGenerateDeviceCode(item.NombreEquipo)
	}
	item.EstadoConexion = soporteRemotoNormalizeConexion(item.EstadoConexion)
	item.Estado = soporteRemotoNormalizeEstado(item.Estado)
	if item.Estado == "" {
		item.Estado = "activo"
	}

	accesoPINHash := strings.TrimSpace(item.AccesoPINHash)
	if accesoPINHash == "" {
		accesoPINHash = soporteRemotoHash(accesoPINPlano)
	}

	res, err := dbConn.Exec(`INSERT INTO empresa_soporte_remoto_dispositivos (
		empresa_id,
		codigo_dispositivo,
		nombre_equipo,
		alias_operativo,
		ubicacion,
		sistema_operativo,
		agente_version,
		stream_url,
		estado_conexion,
		ultimo_heartbeat,
		acceso_pin_hash,
		usuario_creador,
		estado,
		observaciones
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		item.EmpresaID,
		item.CodigoDispositivo,
		item.NombreEquipo,
		strings.TrimSpace(item.AliasOperativo),
		strings.TrimSpace(item.Ubicacion),
		strings.TrimSpace(item.SistemaOperativo),
		strings.TrimSpace(item.AgenteVersion),
		item.StreamURL,
		item.EstadoConexion,
		strings.TrimSpace(item.UltimoHeartbeat),
		accesoPINHash,
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

// UpdateEmpresaSoporteRemotoDispositivo actualiza metadata de un dispositivo.
func UpdateEmpresaSoporteRemotoDispositivo(dbConn *sql.DB, item EmpresaSoporteRemotoDispositivo, accesoPINPlano string) error {
	if dbConn == nil {
		return errors.New("db connection is nil")
	}
	if item.ID <= 0 || item.EmpresaID <= 0 {
		return errors.New("id/empresa_id invalidos")
	}
	item.NombreEquipo = strings.TrimSpace(item.NombreEquipo)
	if item.NombreEquipo == "" {
		return errors.New("nombre_equipo es obligatorio")
	}
	item.StreamURL = strings.TrimSpace(item.StreamURL)
	if item.StreamURL == "" {
		return errors.New("stream_url es obligatorio")
	}
	item.CodigoDispositivo = strings.TrimSpace(item.CodigoDispositivo)
	if item.CodigoDispositivo == "" {
		return errors.New("codigo_dispositivo es obligatorio")
	}
	item.EstadoConexion = soporteRemotoNormalizeConexion(item.EstadoConexion)
	item.Estado = soporteRemotoNormalizeEstado(item.Estado)
	if item.Estado == "" {
		item.Estado = "activo"
	}

	accesoPINHash := strings.TrimSpace(item.AccesoPINHash)
	if accesoPINHash == "" {
		accesoPINHash = soporteRemotoHash(accesoPINPlano)
	}

	res, err := dbConn.Exec(`UPDATE empresa_soporte_remoto_dispositivos
		SET codigo_dispositivo = ?,
			nombre_equipo = ?,
			alias_operativo = ?,
			ubicacion = ?,
			sistema_operativo = ?,
			agente_version = ?,
			stream_url = ?,
			estado_conexion = ?,
			ultimo_heartbeat = ?,
			acceso_pin_hash = CASE WHEN ? = '' THEN acceso_pin_hash ELSE ? END,
			usuario_creador = ?,
			estado = ?,
			observaciones = ?,
			fecha_actualizacion = datetime('now','localtime')
		WHERE id = ? AND empresa_id = ?`,
		item.CodigoDispositivo,
		item.NombreEquipo,
		strings.TrimSpace(item.AliasOperativo),
		strings.TrimSpace(item.Ubicacion),
		strings.TrimSpace(item.SistemaOperativo),
		strings.TrimSpace(item.AgenteVersion),
		item.StreamURL,
		item.EstadoConexion,
		strings.TrimSpace(item.UltimoHeartbeat),
		accesoPINHash,
		accesoPINHash,
		strings.TrimSpace(item.UsuarioCreador),
		item.Estado,
		strings.TrimSpace(item.Observaciones),
		item.ID,
		item.EmpresaID,
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

// SetEmpresaSoporteRemotoDispositivoEstadoByID activa/desactiva un dispositivo por empresa.
func SetEmpresaSoporteRemotoDispositivoEstadoByID(dbConn *sql.DB, empresaID, deviceID int64, estado string) error {
	if dbConn == nil {
		return errors.New("db connection is nil")
	}
	if empresaID <= 0 || deviceID <= 0 {
		return errors.New("empresa_id/id invalidos")
	}
	res, err := dbConn.Exec(`UPDATE empresa_soporte_remoto_dispositivos
		SET estado = ?, fecha_actualizacion = datetime('now','localtime')
		WHERE id = ? AND empresa_id = ?`, soporteRemotoNormalizeEstado(estado), deviceID, empresaID)
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

// GetEmpresaSoporteRemotoDispositivoByID obtiene un dispositivo por empresa e ID.
func GetEmpresaSoporteRemotoDispositivoByID(dbConn *sql.DB, empresaID, deviceID int64) (EmpresaSoporteRemotoDispositivo, error) {
	if dbConn == nil {
		return EmpresaSoporteRemotoDispositivo{}, errors.New("db connection is nil")
	}
	if empresaID <= 0 || deviceID <= 0 {
		return EmpresaSoporteRemotoDispositivo{}, errors.New("empresa_id/id invalidos")
	}

	var out EmpresaSoporteRemotoDispositivo
	err := dbConn.QueryRow(`SELECT
		id,
		empresa_id,
		COALESCE(codigo_dispositivo, ''),
		COALESCE(nombre_equipo, ''),
		COALESCE(alias_operativo, ''),
		COALESCE(ubicacion, ''),
		COALESCE(sistema_operativo, ''),
		COALESCE(agente_version, ''),
		COALESCE(stream_url, ''),
		COALESCE(estado_conexion, 'offline'),
		COALESCE(ultimo_heartbeat, ''),
		COALESCE(acceso_pin_hash, ''),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM empresa_soporte_remoto_dispositivos
	WHERE empresa_id = ? AND id = ?
	LIMIT 1`, empresaID, deviceID).Scan(
		&out.ID,
		&out.EmpresaID,
		&out.CodigoDispositivo,
		&out.NombreEquipo,
		&out.AliasOperativo,
		&out.Ubicacion,
		&out.SistemaOperativo,
		&out.AgenteVersion,
		&out.StreamURL,
		&out.EstadoConexion,
		&out.UltimoHeartbeat,
		&out.AccesoPINHash,
		&out.FechaCreacion,
		&out.FechaActualizacion,
		&out.UsuarioCreador,
		&out.Estado,
		&out.Observaciones,
	)
	if err != nil {
		return EmpresaSoporteRemotoDispositivo{}, err
	}
	out.EstadoConexion = soporteRemotoNormalizeConexion(out.EstadoConexion)
	out.Estado = soporteRemotoNormalizeEstado(out.Estado)
	return out, nil
}

// GetEmpresaSoporteRemotoDispositivoByCodigo obtiene un dispositivo por codigo y empresa.
func GetEmpresaSoporteRemotoDispositivoByCodigo(dbConn *sql.DB, empresaID int64, codigoDispositivo string) (EmpresaSoporteRemotoDispositivo, error) {
	if dbConn == nil {
		return EmpresaSoporteRemotoDispositivo{}, errors.New("db connection is nil")
	}
	if empresaID <= 0 {
		return EmpresaSoporteRemotoDispositivo{}, errors.New("empresa_id invalido")
	}
	codigoDispositivo = strings.TrimSpace(codigoDispositivo)
	if codigoDispositivo == "" {
		return EmpresaSoporteRemotoDispositivo{}, errors.New("codigo_dispositivo invalido")
	}

	var out EmpresaSoporteRemotoDispositivo
	err := dbConn.QueryRow(`SELECT
		id,
		empresa_id,
		COALESCE(codigo_dispositivo, ''),
		COALESCE(nombre_equipo, ''),
		COALESCE(alias_operativo, ''),
		COALESCE(ubicacion, ''),
		COALESCE(sistema_operativo, ''),
		COALESCE(agente_version, ''),
		COALESCE(stream_url, ''),
		COALESCE(estado_conexion, 'offline'),
		COALESCE(ultimo_heartbeat, ''),
		COALESCE(acceso_pin_hash, ''),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM empresa_soporte_remoto_dispositivos
	WHERE empresa_id = ? AND trim(codigo_dispositivo) = trim(?)
	LIMIT 1`, empresaID, codigoDispositivo).Scan(
		&out.ID,
		&out.EmpresaID,
		&out.CodigoDispositivo,
		&out.NombreEquipo,
		&out.AliasOperativo,
		&out.Ubicacion,
		&out.SistemaOperativo,
		&out.AgenteVersion,
		&out.StreamURL,
		&out.EstadoConexion,
		&out.UltimoHeartbeat,
		&out.AccesoPINHash,
		&out.FechaCreacion,
		&out.FechaActualizacion,
		&out.UsuarioCreador,
		&out.Estado,
		&out.Observaciones,
	)
	if err != nil {
		return EmpresaSoporteRemotoDispositivo{}, err
	}
	out.EstadoConexion = soporteRemotoNormalizeConexion(out.EstadoConexion)
	out.Estado = soporteRemotoNormalizeEstado(out.Estado)
	return out, nil
}

// ValidateEmpresaSoporteRemotoDispositivoAccess valida PIN de acceso del dispositivo.
func ValidateEmpresaSoporteRemotoDispositivoAccess(dbConn *sql.DB, empresaID int64, codigoDispositivo, accesoPIN string) (EmpresaSoporteRemotoDispositivo, error) {
	item, err := GetEmpresaSoporteRemotoDispositivoByCodigo(dbConn, empresaID, codigoDispositivo)
	if err != nil {
		return EmpresaSoporteRemotoDispositivo{}, err
	}
	if strings.EqualFold(item.Estado, "inactivo") {
		return EmpresaSoporteRemotoDispositivo{}, sql.ErrNoRows
	}
	hash := strings.TrimSpace(item.AccesoPINHash)
	if hash != "" {
		if soporteRemotoHash(accesoPIN) != hash {
			return EmpresaSoporteRemotoDispositivo{}, sql.ErrNoRows
		}
	}
	return item, nil
}

// RegisterEmpresaSoporteRemotoDispositivoHeartbeat actualiza latido/estado de conexion para un dispositivo.
func RegisterEmpresaSoporteRemotoDispositivoHeartbeat(dbConn *sql.DB, empresaID int64, codigoDispositivo, accesoPIN, streamURL, sistemaOperativo, agenteVersion string) (EmpresaSoporteRemotoDispositivo, error) {
	device, err := ValidateEmpresaSoporteRemotoDispositivoAccess(dbConn, empresaID, codigoDispositivo, accesoPIN)
	if err != nil {
		return EmpresaSoporteRemotoDispositivo{}, err
	}

	stream := strings.TrimSpace(streamURL)
	if stream == "" {
		stream = device.StreamURL
	}

	_, err = dbConn.Exec(`UPDATE empresa_soporte_remoto_dispositivos
		SET stream_url = ?,
			sistema_operativo = CASE WHEN ? = '' THEN sistema_operativo ELSE ? END,
			agente_version = CASE WHEN ? = '' THEN agente_version ELSE ? END,
			estado_conexion = 'online',
			ultimo_heartbeat = datetime('now','localtime'),
			fecha_actualizacion = datetime('now','localtime')
		WHERE id = ? AND empresa_id = ?`,
		stream,
		strings.TrimSpace(sistemaOperativo), strings.TrimSpace(sistemaOperativo),
		strings.TrimSpace(agenteVersion), strings.TrimSpace(agenteVersion),
		device.ID,
		empresaID,
	)
	if err != nil {
		return EmpresaSoporteRemotoDispositivo{}, err
	}

	return GetEmpresaSoporteRemotoDispositivoByID(dbConn, empresaID, device.ID)
}

// ListEmpresaSoporteRemotoDispositivos lista dispositivos remotos por empresa.
func ListEmpresaSoporteRemotoDispositivos(dbConn *sql.DB, empresaID int64, filter EmpresaSoporteRemotoDispositivoFilter) ([]EmpresaSoporteRemotoDispositivo, int64, error) {
	if dbConn == nil {
		return nil, 0, errors.New("db connection is nil")
	}
	if empresaID <= 0 {
		return nil, 0, errors.New("empresa_id invalido")
	}

	limit, offset := soporteRemotoNormalizeLimitOffset(filter.Limit, filter.Offset)
	where := "WHERE empresa_id = ?"
	args := make([]interface{}, 0)
	args = append(args, empresaID)

	if !filter.IncludeInactive {
		where += " AND COALESCE(estado, 'activo') <> 'inactivo'"
	}
	if q := strings.TrimSpace(filter.Q); q != "" {
		pattern := soporteRemotoLikePattern(q)
		where += ` AND (
			LOWER(COALESCE(codigo_dispositivo, '')) LIKE LOWER(?) ESCAPE '!' OR
			LOWER(COALESCE(nombre_equipo, '')) LIKE LOWER(?) ESCAPE '!' OR
			LOWER(COALESCE(alias_operativo, '')) LIKE LOWER(?) ESCAPE '!' OR
			LOWER(COALESCE(ubicacion, '')) LIKE LOWER(?) ESCAPE '!'
		)`
		args = append(args, pattern, pattern, pattern, pattern)
	}

	var total int64
	if err := dbConn.QueryRow("SELECT COUNT(1) FROM empresa_soporte_remoto_dispositivos "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := dbConn.Query(`SELECT
		id,
		empresa_id,
		COALESCE(codigo_dispositivo, ''),
		COALESCE(nombre_equipo, ''),
		COALESCE(alias_operativo, ''),
		COALESCE(ubicacion, ''),
		COALESCE(sistema_operativo, ''),
		COALESCE(agente_version, ''),
		COALESCE(stream_url, ''),
		COALESCE(estado_conexion, 'offline'),
		COALESCE(ultimo_heartbeat, ''),
		COALESCE(acceso_pin_hash, ''),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM empresa_soporte_remoto_dispositivos `+where+`
	ORDER BY id DESC
	LIMIT ? OFFSET ?`, append(args, limit, offset)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := make([]EmpresaSoporteRemotoDispositivo, 0)
	for rows.Next() {
		var item EmpresaSoporteRemotoDispositivo
		if err := rows.Scan(
			&item.ID,
			&item.EmpresaID,
			&item.CodigoDispositivo,
			&item.NombreEquipo,
			&item.AliasOperativo,
			&item.Ubicacion,
			&item.SistemaOperativo,
			&item.AgenteVersion,
			&item.StreamURL,
			&item.EstadoConexion,
			&item.UltimoHeartbeat,
			&item.AccesoPINHash,
			&item.FechaCreacion,
			&item.FechaActualizacion,
			&item.UsuarioCreador,
			&item.Estado,
			&item.Observaciones,
		); err != nil {
			return nil, 0, err
		}
		item.EstadoConexion = soporteRemotoNormalizeConexion(item.EstadoConexion)
		item.Estado = soporteRemotoNormalizeEstado(item.Estado)
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

// CreateEmpresaSoporteRemotoSession crea una sesion de visualizacion remota y retorna token de vista.
func CreateEmpresaSoporteRemotoSession(dbConn *sql.DB, empresaID, dispositivoID int64, solicitadaPor, operadorNombre, operadorEmail, motivo string, duracionMinutos int, requiereAprobacion bool) (EmpresaSoporteRemotoSession, error) {
	if dbConn == nil {
		return EmpresaSoporteRemotoSession{}, errors.New("db connection is nil")
	}
	if empresaID <= 0 || dispositivoID <= 0 {
		return EmpresaSoporteRemotoSession{}, errors.New("empresa_id/dispositivo_id invalidos")
	}

	device, err := GetEmpresaSoporteRemotoDispositivoByID(dbConn, empresaID, dispositivoID)
	if err != nil {
		return EmpresaSoporteRemotoSession{}, err
	}
	if strings.EqualFold(device.Estado, "inactivo") {
		return EmpresaSoporteRemotoSession{}, errors.New("dispositivo inactivo")
	}
	if strings.TrimSpace(device.StreamURL) == "" {
		return EmpresaSoporteRemotoSession{}, errors.New("dispositivo sin stream_url")
	}

	duracion := soporteRemotoNormalizeAutoCerrar(duracionMinutos)
	estadoSesion := "activa"
	iniciadaEn := time.Now().In(time.Local).Format("2006-01-02 15:04:05")
	if requiereAprobacion {
		estadoSesion = "pendiente"
		iniciadaEn = ""
	}
	expiraEn := time.Now().In(time.Local).Add(time.Duration(duracion) * time.Minute).Format("2006-01-02 15:04:05")
	codigoSesion := soporteRemotoGenerateSessionCode()
	tokenRaw := soporteRemotoGenerateToken("SRV")
	tokenHash := soporteRemotoHash(tokenRaw)

	res, err := dbConn.Exec(`INSERT INTO empresa_soporte_remoto_sesiones (
		empresa_id,
		dispositivo_id,
		codigo_sesion,
		solicitada_por,
		operador_nombre,
		operador_email,
		motivo,
		estado_sesion,
		token_visualizacion_hash,
		url_visualizacion,
		iniciada_en,
		expira_en,
		usuario_creador,
		estado,
		observaciones
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'activo', ?)`,
		empresaID,
		dispositivoID,
		codigoSesion,
		strings.TrimSpace(solicitadaPor),
		strings.TrimSpace(operadorNombre),
		strings.TrimSpace(operadorEmail),
		strings.TrimSpace(motivo),
		estadoSesion,
		tokenHash,
		strings.TrimSpace(device.StreamURL),
		iniciadaEn,
		expiraEn,
		strings.TrimSpace(solicitadaPor),
		"sesion creada por panel",
	)
	if err != nil {
		return EmpresaSoporteRemotoSession{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return EmpresaSoporteRemotoSession{}, err
	}

	session, err := GetEmpresaSoporteRemotoSessionByID(dbConn, empresaID, id)
	if err != nil {
		return EmpresaSoporteRemotoSession{}, err
	}
	session.TokenVisualizacionRaw = tokenRaw
	return session, nil
}

// GetEmpresaSoporteRemotoSessionByID consulta una sesion por ID.
func GetEmpresaSoporteRemotoSessionByID(dbConn *sql.DB, empresaID, sessionID int64) (EmpresaSoporteRemotoSession, error) {
	if dbConn == nil {
		return EmpresaSoporteRemotoSession{}, errors.New("db connection is nil")
	}
	if empresaID <= 0 || sessionID <= 0 {
		return EmpresaSoporteRemotoSession{}, errors.New("empresa_id/id invalidos")
	}

	var out EmpresaSoporteRemotoSession
	err := dbConn.QueryRow(`SELECT
		s.id,
		s.empresa_id,
		s.dispositivo_id,
		COALESCE(d.codigo_dispositivo, ''),
		COALESCE(d.nombre_equipo, ''),
		COALESCE(s.codigo_sesion, ''),
		COALESCE(s.solicitada_por, ''),
		COALESCE(s.operador_nombre, ''),
		COALESCE(s.operador_email, ''),
		COALESCE(s.motivo, ''),
		COALESCE(s.estado_sesion, 'pendiente'),
		COALESCE(s.token_visualizacion_hash, ''),
		COALESCE(s.url_visualizacion, ''),
		COALESCE(s.iniciada_en, ''),
		COALESCE(s.expira_en, ''),
		COALESCE(s.finalizada_en, ''),
		COALESCE(s.fecha_creacion, ''),
		COALESCE(s.fecha_actualizacion, ''),
		COALESCE(s.usuario_creador, ''),
		COALESCE(s.estado, 'activo'),
		COALESCE(s.observaciones, '')
	FROM empresa_soporte_remoto_sesiones s
	LEFT JOIN empresa_soporte_remoto_dispositivos d ON d.id = s.dispositivo_id
	WHERE s.empresa_id = ? AND s.id = ?
	LIMIT 1`, empresaID, sessionID).Scan(
		&out.ID,
		&out.EmpresaID,
		&out.DispositivoID,
		&out.DispositivoCodigo,
		&out.DispositivoNombre,
		&out.CodigoSesion,
		&out.SolicitadaPor,
		&out.OperadorNombre,
		&out.OperadorEmail,
		&out.Motivo,
		&out.EstadoSesion,
		&out.tokenVisualizacionHash,
		&out.URLVisualizacion,
		&out.IniciadaEn,
		&out.ExpiraEn,
		&out.FinalizadaEn,
		&out.FechaCreacion,
		&out.FechaActualizacion,
		&out.UsuarioCreador,
		&out.Estado,
		&out.Observaciones,
	)
	if err != nil {
		return EmpresaSoporteRemotoSession{}, err
	}
	out.EstadoSesion = soporteRemotoNormalizeSesionEstado(out.EstadoSesion)
	out.Estado = soporteRemotoNormalizeEstado(out.Estado)
	return out, nil
}

// GetEmpresaSoporteRemotoSessionByCodigo consulta una sesion por codigo.
func GetEmpresaSoporteRemotoSessionByCodigo(dbConn *sql.DB, empresaID int64, codigoSesion string) (EmpresaSoporteRemotoSession, error) {
	if dbConn == nil {
		return EmpresaSoporteRemotoSession{}, errors.New("db connection is nil")
	}
	if empresaID <= 0 {
		return EmpresaSoporteRemotoSession{}, errors.New("empresa_id invalido")
	}
	codigoSesion = strings.TrimSpace(codigoSesion)
	if codigoSesion == "" {
		return EmpresaSoporteRemotoSession{}, errors.New("codigo_sesion invalido")
	}

	var sessionID int64
	if err := dbConn.QueryRow(`SELECT id FROM empresa_soporte_remoto_sesiones WHERE empresa_id = ? AND trim(codigo_sesion) = trim(?) LIMIT 1`, empresaID, codigoSesion).Scan(&sessionID); err != nil {
		return EmpresaSoporteRemotoSession{}, err
	}
	return GetEmpresaSoporteRemotoSessionByID(dbConn, empresaID, sessionID)
}

// ResolveEmpresaSoporteRemotoViewerSession valida token de vista para una sesion.
func ResolveEmpresaSoporteRemotoViewerSession(dbConn *sql.DB, empresaID int64, codigoSesion, tokenVisualizacion string) (EmpresaSoporteRemotoSession, error) {
	session, err := GetEmpresaSoporteRemotoSessionByCodigo(dbConn, empresaID, codigoSesion)
	if err != nil {
		return EmpresaSoporteRemotoSession{}, err
	}
	if strings.EqualFold(session.Estado, "inactivo") {
		return EmpresaSoporteRemotoSession{}, sql.ErrNoRows
	}
	if strings.TrimSpace(session.tokenVisualizacionHash) == "" {
		return EmpresaSoporteRemotoSession{}, sql.ErrNoRows
	}
	if soporteHash := soporteRemotoHash(strings.TrimSpace(tokenVisualizacion)); soporteHash != session.tokenVisualizacionHash {
		return EmpresaSoporteRemotoSession{}, sql.ErrNoRows
	}

	if strings.TrimSpace(session.ExpiraEn) != "" {
		exp, err := time.ParseInLocation("2006-01-02 15:04:05", session.ExpiraEn, time.Local)
		if err == nil && time.Now().After(exp) && session.EstadoSesion != "finalizada" && session.EstadoSesion != "rechazada" {
			_ = SetEmpresaSoporteRemotoSessionEstadoByCodigo(dbConn, empresaID, session.CodigoSesion, "expirada", "sesion expirada automaticamente")
			session.EstadoSesion = "expirada"
		}
	}

	return session, nil
}

// SetEmpresaSoporteRemotoSessionEstadoByCodigo actualiza estado de una sesion por codigo.
func SetEmpresaSoporteRemotoSessionEstadoByCodigo(dbConn *sql.DB, empresaID int64, codigoSesion, estadoSesion, observaciones string) error {
	if dbConn == nil {
		return errors.New("db connection is nil")
	}
	if empresaID <= 0 {
		return errors.New("empresa_id invalido")
	}
	codigoSesion = strings.TrimSpace(codigoSesion)
	if codigoSesion == "" {
		return errors.New("codigo_sesion invalido")
	}

	estado := soporteRemotoNormalizeSesionEstado(estadoSesion)
	iniciadaEn := ""
	finalizadaEn := ""
	if estado == "activa" {
		iniciadaEn = time.Now().In(time.Local).Format("2006-01-02 15:04:05")
	}
	if estado == "finalizada" || estado == "rechazada" || estado == "expirada" {
		finalizadaEn = time.Now().In(time.Local).Format("2006-01-02 15:04:05")
	}

	res, err := dbConn.Exec(`UPDATE empresa_soporte_remoto_sesiones
		SET estado_sesion = ?,
			iniciada_en = CASE WHEN ? = '' THEN iniciada_en ELSE ? END,
			finalizada_en = CASE WHEN ? = '' THEN finalizada_en ELSE ? END,
			observaciones = CASE WHEN ? = '' THEN observaciones ELSE ? END,
			fecha_actualizacion = datetime('now','localtime')
		WHERE empresa_id = ? AND trim(codigo_sesion) = trim(?)`,
		estado,
		iniciadaEn, iniciadaEn,
		finalizadaEn, finalizadaEn,
		strings.TrimSpace(observaciones), strings.TrimSpace(observaciones),
		empresaID,
		codigoSesion,
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

// ListEmpresaSoporteRemotoSesiones lista sesiones de soporte remoto por empresa.
func ListEmpresaSoporteRemotoSesiones(dbConn *sql.DB, empresaID int64, filter EmpresaSoporteRemotoSessionFilter) ([]EmpresaSoporteRemotoSession, int64, error) {
	if dbConn == nil {
		return nil, 0, errors.New("db connection is nil")
	}
	if empresaID <= 0 {
		return nil, 0, errors.New("empresa_id invalido")
	}

	limit, offset := soporteRemotoNormalizeLimitOffset(filter.Limit, filter.Offset)
	where := "WHERE s.empresa_id = ?"
	args := make([]interface{}, 0)
	args = append(args, empresaID)

	if !filter.IncludeInactive {
		where += " AND COALESCE(s.estado, 'activo') <> 'inactivo'"
	}
	if estado := strings.TrimSpace(strings.ToLower(filter.EstadoSesion)); estado != "" {
		where += " AND LOWER(COALESCE(s.estado_sesion, '')) = ?"
		args = append(args, estado)
	}
	if q := strings.TrimSpace(filter.Q); q != "" {
		pattern := soporteRemotoLikePattern(q)
		where += ` AND (
			LOWER(COALESCE(s.codigo_sesion, '')) LIKE LOWER(?) ESCAPE '!' OR
			LOWER(COALESCE(s.solicitada_por, '')) LIKE LOWER(?) ESCAPE '!' OR
			LOWER(COALESCE(s.operador_nombre, '')) LIKE LOWER(?) ESCAPE '!' OR
			LOWER(COALESCE(d.nombre_equipo, '')) LIKE LOWER(?) ESCAPE '!'
		)`
		args = append(args, pattern, pattern, pattern, pattern)
	}

	var total int64
	if err := dbConn.QueryRow(`SELECT COUNT(1)
		FROM empresa_soporte_remoto_sesiones s
		LEFT JOIN empresa_soporte_remoto_dispositivos d ON d.id = s.dispositivo_id `+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := dbConn.Query(`SELECT
		s.id,
		s.empresa_id,
		s.dispositivo_id,
		COALESCE(d.codigo_dispositivo, ''),
		COALESCE(d.nombre_equipo, ''),
		COALESCE(s.codigo_sesion, ''),
		COALESCE(s.solicitada_por, ''),
		COALESCE(s.operador_nombre, ''),
		COALESCE(s.operador_email, ''),
		COALESCE(s.motivo, ''),
		COALESCE(s.estado_sesion, 'pendiente'),
		COALESCE(s.token_visualizacion_hash, ''),
		COALESCE(s.url_visualizacion, ''),
		COALESCE(s.iniciada_en, ''),
		COALESCE(s.expira_en, ''),
		COALESCE(s.finalizada_en, ''),
		COALESCE(s.fecha_creacion, ''),
		COALESCE(s.fecha_actualizacion, ''),
		COALESCE(s.usuario_creador, ''),
		COALESCE(s.estado, 'activo'),
		COALESCE(s.observaciones, '')
	FROM empresa_soporte_remoto_sesiones s
	LEFT JOIN empresa_soporte_remoto_dispositivos d ON d.id = s.dispositivo_id `+where+`
	ORDER BY s.id DESC
	LIMIT ? OFFSET ?`, append(args, limit, offset)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := make([]EmpresaSoporteRemotoSession, 0)
	for rows.Next() {
		var item EmpresaSoporteRemotoSession
		if err := rows.Scan(
			&item.ID,
			&item.EmpresaID,
			&item.DispositivoID,
			&item.DispositivoCodigo,
			&item.DispositivoNombre,
			&item.CodigoSesion,
			&item.SolicitadaPor,
			&item.OperadorNombre,
			&item.OperadorEmail,
			&item.Motivo,
			&item.EstadoSesion,
			&item.tokenVisualizacionHash,
			&item.URLVisualizacion,
			&item.IniciadaEn,
			&item.ExpiraEn,
			&item.FinalizadaEn,
			&item.FechaCreacion,
			&item.FechaActualizacion,
			&item.UsuarioCreador,
			&item.Estado,
			&item.Observaciones,
		); err != nil {
			return nil, 0, err
		}
		item.EstadoSesion = soporteRemotoNormalizeSesionEstado(item.EstadoSesion)
		item.Estado = soporteRemotoNormalizeEstado(item.Estado)
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return out, total, nil
}
