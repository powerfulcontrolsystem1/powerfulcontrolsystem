package db

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// EmpresaVehiculoRegistro representa el ingreso/salida de vehiculos por empresa.
type EmpresaVehiculoRegistro struct {
	ID                   int64  `json:"id"`
	EmpresaID            int64  `json:"empresa_id"`
	Patente              string `json:"patente"`
	TipoVehiculo         string `json:"tipo_vehiculo,omitempty"`
	Marca                string `json:"marca,omitempty"`
	Modelo               string `json:"modelo,omitempty"`
	Color                string `json:"color,omitempty"`
	PropietarioNombre    string `json:"propietario_nombre,omitempty"`
	PropietarioDocumento string `json:"propietario_documento,omitempty"`
	ConductorNombre      string `json:"conductor_nombre,omitempty"`
	ConductorDocumento   string `json:"conductor_documento,omitempty"`
	MotivoIngreso        string `json:"motivo_ingreso,omitempty"`
	ReferenciaExterna    string `json:"referencia_externa,omitempty"`
	FechaIngreso         string `json:"fecha_ingreso,omitempty"`
	FechaSalida          string `json:"fecha_salida,omitempty"`
	EstadoRegistro       string `json:"estado_registro,omitempty"`
	UsuarioSalida        string `json:"usuario_salida,omitempty"`
	FechaCreacion        string `json:"fecha_creacion,omitempty"`
	FechaActualizacion   string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador       string `json:"usuario_creador,omitempty"`
	Estado               string `json:"estado,omitempty"`
	Observaciones        string `json:"observaciones,omitempty"`
}

// EmpresaVehiculosRegistroConfiguracion define reglas de placa/patente por empresa.
type EmpresaVehiculosRegistroConfiguracion struct {
	ID                    int64  `json:"id"`
	EmpresaID             int64  `json:"empresa_id"`
	PaisCodigo            string `json:"pais_codigo"`
	PatenteRegex          string `json:"patente_regex,omitempty"`
	PatenteDescripcion    string `json:"patente_descripcion,omitempty"`
	EvitarDuplicadoActivo bool   `json:"evitar_duplicado_activo"`
	FechaCreacion         string `json:"fecha_creacion,omitempty"`
	FechaActualizacion    string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador        string `json:"usuario_creador,omitempty"`
	Estado                string `json:"estado,omitempty"`
	Observaciones         string `json:"observaciones,omitempty"`
}

// EmpresaVehiculoPermanenciaReporteItem representa la permanencia y tiempo de estancia por registro.
type EmpresaVehiculoPermanenciaReporteItem struct {
	ID                int64   `json:"id"`
	EmpresaID         int64   `json:"empresa_id"`
	Patente           string  `json:"patente"`
	TipoVehiculo      string  `json:"tipo_vehiculo,omitempty"`
	ConductorNombre   string  `json:"conductor_nombre,omitempty"`
	PropietarioNombre string  `json:"propietario_nombre,omitempty"`
	FechaIngreso      string  `json:"fecha_ingreso,omitempty"`
	FechaSalida       string  `json:"fecha_salida,omitempty"`
	EstadoRegistro    string  `json:"estado_registro,omitempty"`
	Estado            string  `json:"estado,omitempty"`
	MinutosEstadia    int64   `json:"minutos_estadia"`
	HorasEstadia      float64 `json:"horas_estadia"`
	DiasEstadia       float64 `json:"dias_estadia"`
}

var ErrEmpresaVehiculoDuplicadoActivo = errors.New("ya existe un vehiculo activo en la empresa con la misma patente/placa")

// EnsureEmpresaVehiculosRegistroSchema crea/migra la tabla de registro de vehiculos por empresa.
func EnsureEmpresaVehiculosRegistroSchema(dbConn *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_vehiculos_registro (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			patente TEXT NOT NULL,
			tipo_vehiculo TEXT DEFAULT 'automovil',
			marca TEXT,
			modelo TEXT,
			color TEXT,
			propietario_nombre TEXT,
			propietario_documento TEXT,
			conductor_nombre TEXT,
			conductor_documento TEXT,
			motivo_ingreso TEXT,
			referencia_externa TEXT,
			fecha_ingreso TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_salida TEXT,
			estado_registro TEXT DEFAULT 'en_empresa',
			usuario_salida TEXT,
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_vehiculos_registro_empresa_fecha ON empresa_vehiculos_registro(empresa_id, fecha_ingreso DESC, id DESC);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_vehiculos_registro_empresa_patente ON empresa_vehiculos_registro(empresa_id, patente);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_vehiculos_registro_empresa_estado ON empresa_vehiculos_registro(empresa_id, estado, estado_registro);`,
		`CREATE TABLE IF NOT EXISTS empresa_vehiculos_configuracion (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL UNIQUE,
			pais_codigo TEXT DEFAULT 'CO',
			patente_regex TEXT,
			patente_descripcion TEXT,
			evitar_duplicado_activo INTEGER DEFAULT 1,
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_empresa_vehiculos_configuracion_empresa ON empresa_vehiculos_configuracion(empresa_id);`,
	}
	for _, stmt := range stmts {
		if _, err := dbConn.Exec(stmt); err != nil {
			return err
		}
	}

	if err := ensureColumnIfMissing(dbConn, "empresa_vehiculos_registro", "tipo_vehiculo", "TEXT DEFAULT 'automovil'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_vehiculos_registro", "marca", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_vehiculos_registro", "modelo", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_vehiculos_registro", "color", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_vehiculos_registro", "propietario_nombre", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_vehiculos_registro", "propietario_documento", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_vehiculos_registro", "conductor_nombre", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_vehiculos_registro", "conductor_documento", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_vehiculos_registro", "motivo_ingreso", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_vehiculos_registro", "referencia_externa", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_vehiculos_registro", "fecha_ingreso", "TEXT DEFAULT (CURRENT_TIMESTAMP)"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_vehiculos_registro", "fecha_salida", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_vehiculos_registro", "estado_registro", "TEXT DEFAULT 'en_empresa'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_vehiculos_registro", "usuario_salida", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_vehiculos_registro", "fecha_actualizacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_vehiculos_registro", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_vehiculos_registro", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_vehiculos_registro", "observaciones", "TEXT"); err != nil {
		return err
	}

	if err := ensureColumnIfMissing(dbConn, "empresa_vehiculos_configuracion", "pais_codigo", "TEXT DEFAULT 'CO'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_vehiculos_configuracion", "patente_regex", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_vehiculos_configuracion", "patente_descripcion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_vehiculos_configuracion", "evitar_duplicado_activo", "INTEGER DEFAULT 1"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_vehiculos_configuracion", "fecha_actualizacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_vehiculos_configuracion", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_vehiculos_configuracion", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_vehiculos_configuracion", "observaciones", "TEXT"); err != nil {
		return err
	}

	return nil
}

func defaultEmpresaVehiculosRegistroConfiguracion(empresaID int64) *EmpresaVehiculosRegistroConfiguracion {
	pattern, description := vehiculoPlatePatternByCountry("CO")
	return &EmpresaVehiculosRegistroConfiguracion{
		EmpresaID:             empresaID,
		PaisCodigo:            "CO",
		PatenteRegex:          pattern,
		PatenteDescripcion:    description,
		EvitarDuplicadoActivo: true,
		Estado:                "activo",
	}
}

func normalizeVehiculoPaisCodigo(raw string) string {
	v := strings.ToUpper(strings.TrimSpace(raw))
	if v == "" {
		return "CO"
	}
	if len(v) > 3 {
		v = v[:3]
	}
	return v
}

func vehiculoPlatePatternByCountry(pais string) (string, string) {
	switch normalizeVehiculoPaisCodigo(pais) {
	case "CO":
		return "^(?:[A-Z]{3}[0-9]{3}|[A-Z]{3}[0-9]{2}[A-Z])$", "CO: 3 letras + 3 digitos o 3 letras + 2 digitos + 1 letra"
	case "MX":
		return "^[A-Z0-9]{5,7}$", "MX: entre 5 y 7 caracteres alfanumericos"
	case "AR":
		return "^(?:[A-Z]{3}[0-9]{3}|[A-Z]{2}[0-9]{3}[A-Z]{2})$", "AR: formato antiguo ABC123 o Mercosur AA123AA"
	case "CL":
		return "^(?:[A-Z]{4}[0-9]{2}|[A-Z]{2}[0-9]{4})$", "CL: 4 letras + 2 digitos o 2 letras + 4 digitos"
	default:
		return "^[A-Z0-9]{5,8}$", "General: entre 5 y 8 caracteres alfanumericos"
	}
}

func normalizeVehiculoConfig(payload EmpresaVehiculosRegistroConfiguracion) (*EmpresaVehiculosRegistroConfiguracion, error) {
	if payload.EmpresaID <= 0 {
		return nil, fmt.Errorf("empresa_id es obligatorio")
	}

	out := defaultEmpresaVehiculosRegistroConfiguracion(payload.EmpresaID)
	out.ID = payload.ID
	out.PaisCodigo = normalizeVehiculoPaisCodigo(payload.PaisCodigo)
	out.UsuarioCreador = strings.TrimSpace(payload.UsuarioCreador)
	out.Observaciones = strings.TrimSpace(payload.Observaciones)
	out.Estado = normalizeVehiculoEstado(payload.Estado)
	out.EvitarDuplicadoActivo = payload.EvitarDuplicadoActivo

	defaultPattern, defaultDescription := vehiculoPlatePatternByCountry(out.PaisCodigo)
	regex := strings.TrimSpace(payload.PatenteRegex)
	if regex == "" {
		regex = defaultPattern
	}
	if _, err := regexp.Compile(regex); err != nil {
		return nil, fmt.Errorf("patente_regex invalido")
	}
	out.PatenteRegex = regex

	description := strings.TrimSpace(payload.PatenteDescripcion)
	if description == "" {
		description = defaultDescription
	}
	out.PatenteDescripcion = description

	return out, nil
}

// GetEmpresaVehiculosRegistroConfiguracion obtiene reglas de placa/patente por empresa.
func GetEmpresaVehiculosRegistroConfiguracion(dbConn *sql.DB, empresaID int64) (*EmpresaVehiculosRegistroConfiguracion, error) {
	if empresaID <= 0 {
		return nil, fmt.Errorf("empresa_id es obligatorio")
	}

	if err := EnsureEmpresaVehiculosRegistroSchema(dbConn); err != nil {
		return nil, err
	}

	cfg := defaultEmpresaVehiculosRegistroConfiguracion(empresaID)
	var evitarDuplicado int
	err := dbConn.QueryRow(`SELECT
		id,
		empresa_id,
		COALESCE(pais_codigo, 'CO'),
		COALESCE(patente_regex, ''),
		COALESCE(patente_descripcion, ''),
		COALESCE(evitar_duplicado_activo, 1),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM empresa_vehiculos_configuracion
	WHERE empresa_id = ?
	LIMIT 1`, empresaID).Scan(
		&cfg.ID,
		&cfg.EmpresaID,
		&cfg.PaisCodigo,
		&cfg.PatenteRegex,
		&cfg.PatenteDescripcion,
		&evitarDuplicado,
		&cfg.FechaCreacion,
		&cfg.FechaActualizacion,
		&cfg.UsuarioCreador,
		&cfg.Estado,
		&cfg.Observaciones,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return cfg, nil
		}
		return nil, err
	}

	cfg.EvitarDuplicadoActivo = evitarDuplicado == 1
	normalized, err := normalizeVehiculoConfig(*cfg)
	if err != nil {
		return nil, err
	}
	normalized.ID = cfg.ID
	normalized.FechaCreacion = cfg.FechaCreacion
	normalized.FechaActualizacion = cfg.FechaActualizacion
	return normalized, nil
}

// UpsertEmpresaVehiculosRegistroConfiguracion crea o actualiza reglas de placa/patente por empresa.
func UpsertEmpresaVehiculosRegistroConfiguracion(dbConn *sql.DB, payload EmpresaVehiculosRegistroConfiguracion) (int64, error) {
	if err := EnsureEmpresaVehiculosRegistroSchema(dbConn); err != nil {
		return 0, err
	}

	cfg, err := normalizeVehiculoConfig(payload)
	if err != nil {
		return 0, err
	}

	var existingID int64
	err = dbConn.QueryRow(`SELECT id FROM empresa_vehiculos_configuracion WHERE empresa_id = ? LIMIT 1`, cfg.EmpresaID).Scan(&existingID)
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}

	if existingID > 0 {
		_, err = dbConn.Exec(`UPDATE empresa_vehiculos_configuracion
		SET
			pais_codigo = ?,
			patente_regex = ?,
			patente_descripcion = ?,
			evitar_duplicado_activo = ?,
			usuario_creador = ?,
			estado = ?,
			observaciones = ?,
			fecha_actualizacion = CURRENT_TIMESTAMP
		WHERE empresa_id = ?`,
			cfg.PaisCodigo,
			cfg.PatenteRegex,
			cfg.PatenteDescripcion,
			vehiculoBoolToInt(cfg.EvitarDuplicadoActivo),
			cfg.UsuarioCreador,
			cfg.Estado,
			cfg.Observaciones,
			cfg.EmpresaID,
		)
		if err != nil {
			return 0, err
		}
		return existingID, nil
	}

	id, err := insertSQLCompat(dbConn, `INSERT INTO empresa_vehiculos_configuracion (
		empresa_id,
		pais_codigo,
		patente_regex,
		patente_descripcion,
		evitar_duplicado_activo,
		usuario_creador,
		estado,
		observaciones,
		fecha_creacion,
		fecha_actualizacion
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		cfg.EmpresaID,
		cfg.PaisCodigo,
		cfg.PatenteRegex,
		cfg.PatenteDescripcion,
		vehiculoBoolToInt(cfg.EvitarDuplicadoActivo),
		cfg.UsuarioCreador,
		cfg.Estado,
		cfg.Observaciones,
	)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// CreateEmpresaVehiculoRegistro crea un registro de ingreso de vehiculo.
func CreateEmpresaVehiculoRegistro(dbConn *sql.DB, item EmpresaVehiculoRegistro) (int64, error) {
	if item.EmpresaID <= 0 {
		return 0, fmt.Errorf("empresa_id es obligatorio")
	}
	item.Patente = normalizePatenteVehiculo(item.Patente)
	if item.Patente == "" {
		return 0, fmt.Errorf("patente es obligatoria")
	}

	cfg, err := GetEmpresaVehiculosRegistroConfiguracion(dbConn, item.EmpresaID)
	if err != nil {
		return 0, err
	}
	if err := validateVehiculoPatenteByConfig(item.Patente, cfg); err != nil {
		return 0, err
	}

	fechaIngreso, err := normalizeVehiculopcs_ts(item.FechaIngreso, true)
	if err != nil {
		return 0, err
	}
	fechaSalida, err := normalizeVehiculopcs_ts(item.FechaSalida, false)
	if err != nil {
		return 0, err
	}
	estadoRegistro := normalizeEstadoRegistroVehiculo(item.EstadoRegistro)
	estado := normalizeVehiculoEstado(item.Estado)
	if estadoRegistro == "retirado" && fechaSalida == "" {
		fechaSalida = time.Now().Format("2006-01-02 15:04:05")
	}

	if cfg.EvitarDuplicadoActivo && isVehiculoActivoEnPatio(estado, estadoRegistro) {
		duplicado, err := existsVehiculoActivoPatente(dbConn, item.EmpresaID, item.Patente, 0)
		if err != nil {
			return 0, err
		}
		if duplicado {
			return 0, fmt.Errorf("%w: %s", ErrEmpresaVehiculoDuplicadoActivo, item.Patente)
		}
	}

	id, err := insertSQLCompat(dbConn, `INSERT INTO empresa_vehiculos_registro (
		empresa_id, patente, tipo_vehiculo, marca, modelo, color,
		propietario_nombre, propietario_documento,
		conductor_nombre, conductor_documento,
		motivo_ingreso, referencia_externa,
		fecha_ingreso, fecha_salida, estado_registro, usuario_salida,
		usuario_creador, estado, observaciones,
		fecha_creacion, fecha_actualizacion
	) VALUES (
		?, ?, ?, ?, ?, ?,
		?, ?,
		?, ?,
		?, ?,
		?, ?, ?, ?,
		?, COALESCE(NULLIF(?, ''), 'activo'), ?,
		CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
	)`,
		item.EmpresaID,
		item.Patente,
		normalizeTipoVehiculo(item.TipoVehiculo),
		strings.TrimSpace(item.Marca),
		strings.TrimSpace(item.Modelo),
		strings.TrimSpace(item.Color),
		strings.TrimSpace(item.PropietarioNombre),
		strings.TrimSpace(item.PropietarioDocumento),
		strings.TrimSpace(item.ConductorNombre),
		strings.TrimSpace(item.ConductorDocumento),
		strings.TrimSpace(item.MotivoIngreso),
		strings.TrimSpace(item.ReferenciaExterna),
		fechaIngreso,
		fechaSalida,
		estadoRegistro,
		strings.TrimSpace(item.UsuarioSalida),
		strings.TrimSpace(item.UsuarioCreador),
		estado,
		strings.TrimSpace(item.Observaciones),
	)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// ListEmpresaVehiculosRegistros lista registros de vehiculos por empresa con filtros.
func ListEmpresaVehiculosRegistros(dbConn *sql.DB, empresaID int64, includeInactive bool, desde, hasta, estadoRegistro, patente, q string, limit int) ([]EmpresaVehiculoRegistro, error) {
	query := `SELECT
		id, empresa_id, COALESCE(patente, ''), COALESCE(tipo_vehiculo, 'automovil'),
		COALESCE(marca, ''), COALESCE(modelo, ''), COALESCE(color, ''),
		COALESCE(propietario_nombre, ''), COALESCE(propietario_documento, ''),
		COALESCE(conductor_nombre, ''), COALESCE(conductor_documento, ''),
		COALESCE(motivo_ingreso, ''), COALESCE(referencia_externa, ''),
		COALESCE(fecha_ingreso, ''), COALESCE(fecha_salida, ''), COALESCE(estado_registro, 'en_empresa'),
		COALESCE(usuario_salida, ''), COALESCE(fecha_creacion, ''), COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''), COALESCE(estado, 'activo'), COALESCE(observaciones, '')
	FROM empresa_vehiculos_registro
	WHERE empresa_id = ?`
	args := []interface{}{empresaID}

	if !includeInactive {
		query += ` AND COALESCE(estado, 'activo') = 'activo'`
	}

	if desdeTrim := strings.TrimSpace(desde); desdeTrim != "" {
		query += ` AND date(COALESCE(fecha_ingreso, '')) >= date(?)`
		args = append(args, desdeTrim)
	}
	if hastaTrim := strings.TrimSpace(hasta); hastaTrim != "" {
		query += ` AND date(COALESCE(fecha_ingreso, '')) <= date(?)`
		args = append(args, hastaTrim)
	}

	estadoRegistro = strings.TrimSpace(strings.ToLower(estadoRegistro))
	if estadoRegistro != "" {
		query += ` AND LOWER(COALESCE(estado_registro, 'en_empresa')) = ?`
		args = append(args, normalizeEstadoRegistroVehiculo(estadoRegistro))
	}

	patente = normalizePatenteVehiculo(patente)
	if patente != "" {
		query += ` AND UPPER(COALESCE(patente, '')) LIKE ?`
		args = append(args, "%"+patente+"%")
	}

	if qTrim := strings.TrimSpace(strings.ToLower(q)); qTrim != "" {
		like := "%" + qTrim + "%"
		query += ` AND (
			LOWER(COALESCE(patente, '')) LIKE ?
			OR LOWER(COALESCE(conductor_nombre, '')) LIKE ?
			OR LOWER(COALESCE(propietario_nombre, '')) LIKE ?
			OR LOWER(COALESCE(motivo_ingreso, '')) LIKE ?
			OR LOWER(COALESCE(referencia_externa, '')) LIKE ?
		)`
		args = append(args, like, like, like, like, like)
	}

	if limit <= 0 {
		limit = 300
	}
	if limit > 2000 {
		limit = 2000
	}
	query += ` ORDER BY COALESCE(fecha_ingreso, fecha_creacion) DESC, id DESC LIMIT ?`
	args = append(args, limit)

	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]EmpresaVehiculoRegistro, 0)
	for rows.Next() {
		var item EmpresaVehiculoRegistro
		if err := rows.Scan(
			&item.ID,
			&item.EmpresaID,
			&item.Patente,
			&item.TipoVehiculo,
			&item.Marca,
			&item.Modelo,
			&item.Color,
			&item.PropietarioNombre,
			&item.PropietarioDocumento,
			&item.ConductorNombre,
			&item.ConductorDocumento,
			&item.MotivoIngreso,
			&item.ReferenciaExterna,
			&item.FechaIngreso,
			&item.FechaSalida,
			&item.EstadoRegistro,
			&item.UsuarioSalida,
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

	return out, rows.Err()
}

// ListEmpresaVehiculosPermanenciaReporte lista permanencia y tiempo de estancia por registro.
func ListEmpresaVehiculosPermanenciaReporte(dbConn *sql.DB, empresaID int64, includeInactive bool, desde, hasta, patente, q string, limit int) ([]EmpresaVehiculoPermanenciaReporteItem, error) {
	if empresaID <= 0 {
		return nil, fmt.Errorf("empresa_id es obligatorio")
	}

	query := `SELECT
		id,
		empresa_id,
		COALESCE(patente, ''),
		COALESCE(tipo_vehiculo, 'automovil'),
		COALESCE(conductor_nombre, ''),
		COALESCE(propietario_nombre, ''),
		COALESCE(fecha_ingreso, ''),
		COALESCE(fecha_salida, ''),
		COALESCE(estado_registro, 'en_empresa'),
		COALESCE(estado, 'activo'),
		CAST(ROUND((pcs_julian_day(COALESCE(NULLIF(fecha_salida, ''), CURRENT_TIMESTAMP)) - pcs_julian_day(COALESCE(NULLIF(fecha_ingreso, ''), CURRENT_TIMESTAMP))) * 24.0 * 60.0, 0) AS INTEGER) AS minutos_estadia
	FROM empresa_vehiculos_registro
	WHERE empresa_id = ?`
	args := []interface{}{empresaID}

	if !includeInactive {
		query += ` AND COALESCE(estado, 'activo') = 'activo'`
	}

	if desdeTrim := strings.TrimSpace(desde); desdeTrim != "" {
		query += ` AND date(COALESCE(fecha_ingreso, '')) >= date(?)`
		args = append(args, desdeTrim)
	}
	if hastaTrim := strings.TrimSpace(hasta); hastaTrim != "" {
		query += ` AND date(COALESCE(fecha_ingreso, '')) <= date(?)`
		args = append(args, hastaTrim)
	}

	patente = normalizePatenteVehiculo(patente)
	if patente != "" {
		query += ` AND UPPER(COALESCE(patente, '')) LIKE ?`
		args = append(args, "%"+patente+"%")
	}

	if qTrim := strings.TrimSpace(strings.ToLower(q)); qTrim != "" {
		like := "%" + qTrim + "%"
		query += ` AND (
			LOWER(COALESCE(patente, '')) LIKE ?
			OR LOWER(COALESCE(conductor_nombre, '')) LIKE ?
			OR LOWER(COALESCE(propietario_nombre, '')) LIKE ?
			OR LOWER(COALESCE(motivo_ingreso, '')) LIKE ?
			OR LOWER(COALESCE(referencia_externa, '')) LIKE ?
		)`
		args = append(args, like, like, like, like, like)
	}

	if limit <= 0 {
		limit = 500
	}
	if limit > 5000 {
		limit = 5000
	}
	query += ` ORDER BY minutos_estadia DESC, COALESCE(fecha_ingreso, fecha_creacion) DESC, id DESC LIMIT ?`
	args = append(args, limit)

	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]EmpresaVehiculoPermanenciaReporteItem, 0)
	for rows.Next() {
		var item EmpresaVehiculoPermanenciaReporteItem
		var minutos sql.NullInt64
		if err := rows.Scan(
			&item.ID,
			&item.EmpresaID,
			&item.Patente,
			&item.TipoVehiculo,
			&item.ConductorNombre,
			&item.PropietarioNombre,
			&item.FechaIngreso,
			&item.FechaSalida,
			&item.EstadoRegistro,
			&item.Estado,
			&minutos,
		); err != nil {
			return nil, err
		}
		if minutos.Valid {
			item.MinutosEstadia = minutos.Int64
		}
		if item.MinutosEstadia < 0 {
			item.MinutosEstadia = 0
		}
		item.HorasEstadia = roundVehiculo2(float64(item.MinutosEstadia) / 60.0)
		item.DiasEstadia = roundVehiculo2(float64(item.MinutosEstadia) / 1440.0)
		out = append(out, item)
	}

	return out, rows.Err()
}

// UpdateEmpresaVehiculoRegistro actualiza un registro de vehiculo existente.
func UpdateEmpresaVehiculoRegistro(dbConn *sql.DB, item EmpresaVehiculoRegistro) error {
	if item.EmpresaID <= 0 || item.ID <= 0 {
		return fmt.Errorf("empresa_id e id son obligatorios")
	}
	item.Patente = normalizePatenteVehiculo(item.Patente)
	if item.Patente == "" {
		return fmt.Errorf("patente es obligatoria")
	}

	cfg, err := GetEmpresaVehiculosRegistroConfiguracion(dbConn, item.EmpresaID)
	if err != nil {
		return err
	}
	if err := validateVehiculoPatenteByConfig(item.Patente, cfg); err != nil {
		return err
	}

	fechaIngreso, err := normalizeVehiculopcs_ts(item.FechaIngreso, true)
	if err != nil {
		return err
	}
	fechaSalida, err := normalizeVehiculopcs_ts(item.FechaSalida, false)
	if err != nil {
		return err
	}
	estadoRegistro := normalizeEstadoRegistroVehiculo(item.EstadoRegistro)
	estado := normalizeVehiculoEstado(item.Estado)
	if estadoRegistro == "retirado" && fechaSalida == "" {
		fechaSalida = time.Now().Format("2006-01-02 15:04:05")
	}

	if cfg.EvitarDuplicadoActivo && isVehiculoActivoEnPatio(estado, estadoRegistro) {
		duplicado, err := existsVehiculoActivoPatente(dbConn, item.EmpresaID, item.Patente, item.ID)
		if err != nil {
			return err
		}
		if duplicado {
			return fmt.Errorf("%w: %s", ErrEmpresaVehiculoDuplicadoActivo, item.Patente)
		}
	}

	res, err := dbConn.Exec(`UPDATE empresa_vehiculos_registro
	SET
		patente = ?,
		tipo_vehiculo = ?,
		marca = ?,
		modelo = ?,
		color = ?,
		propietario_nombre = ?,
		propietario_documento = ?,
		conductor_nombre = ?,
		conductor_documento = ?,
		motivo_ingreso = ?,
		referencia_externa = ?,
		fecha_ingreso = ?,
		fecha_salida = ?,
		estado_registro = ?,
		usuario_salida = ?,
		observaciones = ?,
		fecha_actualizacion = CURRENT_TIMESTAMP
	WHERE empresa_id = ? AND id = ?`,
		item.Patente,
		normalizeTipoVehiculo(item.TipoVehiculo),
		strings.TrimSpace(item.Marca),
		strings.TrimSpace(item.Modelo),
		strings.TrimSpace(item.Color),
		strings.TrimSpace(item.PropietarioNombre),
		strings.TrimSpace(item.PropietarioDocumento),
		strings.TrimSpace(item.ConductorNombre),
		strings.TrimSpace(item.ConductorDocumento),
		strings.TrimSpace(item.MotivoIngreso),
		strings.TrimSpace(item.ReferenciaExterna),
		fechaIngreso,
		fechaSalida,
		estadoRegistro,
		strings.TrimSpace(item.UsuarioSalida),
		strings.TrimSpace(item.Observaciones),
		item.EmpresaID,
		item.ID,
	)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// SetEmpresaVehiculoRegistroEstado activa o desactiva un registro.
func SetEmpresaVehiculoRegistroEstado(dbConn *sql.DB, empresaID, id int64, estado string) error {
	estado = normalizeVehiculoEstado(estado)

	if estado == "activo" {
		cfg, err := GetEmpresaVehiculosRegistroConfiguracion(dbConn, empresaID)
		if err != nil {
			return err
		}
		if cfg.EvitarDuplicadoActivo {
			var patente string
			var estadoRegistro string
			err = dbConn.QueryRow(`SELECT COALESCE(patente, ''), COALESCE(estado_registro, 'en_empresa')
			FROM empresa_vehiculos_registro
			WHERE empresa_id = ? AND id = ?
			LIMIT 1`, empresaID, id).Scan(&patente, &estadoRegistro)
			if err != nil {
				return err
			}
			if isVehiculoActivoEnPatio(estado, estadoRegistro) {
				duplicado, err := existsVehiculoActivoPatente(dbConn, empresaID, patente, id)
				if err != nil {
					return err
				}
				if duplicado {
					return fmt.Errorf("%w: %s", ErrEmpresaVehiculoDuplicadoActivo, normalizePatenteVehiculo(patente))
				}
			}
		}
	}

	res, err := dbConn.Exec(`UPDATE empresa_vehiculos_registro
	SET estado = ?, fecha_actualizacion = CURRENT_TIMESTAMP
	WHERE empresa_id = ? AND id = ?`, estado, empresaID, id)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// MarkEmpresaVehiculoSalida marca la salida de un vehiculo.
func MarkEmpresaVehiculoSalida(dbConn *sql.DB, empresaID, id int64, fechaSalida, usuarioSalida, observaciones string) error {
	fechaSalidaNorm, err := normalizeVehiculopcs_ts(fechaSalida, false)
	if err != nil {
		return err
	}
	if fechaSalidaNorm == "" {
		fechaSalidaNorm = time.Now().Format("2006-01-02 15:04:05")
	}

	res, err := dbConn.Exec(`UPDATE empresa_vehiculos_registro
	SET
		fecha_salida = ?,
		estado_registro = 'retirado',
		usuario_salida = CASE WHEN ? <> '' THEN ? ELSE usuario_salida END,
		observaciones = CASE WHEN ? <> '' THEN ? ELSE observaciones END,
		fecha_actualizacion = CURRENT_TIMESTAMP
	WHERE empresa_id = ? AND id = ?`,
		fechaSalidaNorm,
		strings.TrimSpace(usuarioSalida), strings.TrimSpace(usuarioSalida),
		strings.TrimSpace(observaciones), strings.TrimSpace(observaciones),
		empresaID,
		id,
	)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// DeleteEmpresaVehiculoRegistro elimina un registro de vehiculo por empresa.
func DeleteEmpresaVehiculoRegistro(dbConn *sql.DB, empresaID, id int64) error {
	res, err := dbConn.Exec(`DELETE FROM empresa_vehiculos_registro WHERE empresa_id = ? AND id = ?`, empresaID, id)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func normalizePatenteVehiculo(raw string) string {
	value := strings.ToUpper(strings.TrimSpace(raw))
	value = strings.ReplaceAll(value, " ", "")
	if len(value) > 20 {
		value = value[:20]
	}
	return value
}

func canonicalPatenteVehiculo(raw string) string {
	v := normalizePatenteVehiculo(raw)
	v = strings.ReplaceAll(v, "-", "")
	v = strings.ReplaceAll(v, ".", "")
	v = strings.ReplaceAll(v, "_", "")
	return v
}

func validateVehiculoPatenteByConfig(patente string, cfg *EmpresaVehiculosRegistroConfiguracion) error {
	if cfg == nil {
		return nil
	}
	pattern := strings.TrimSpace(cfg.PatenteRegex)
	if pattern == "" {
		pattern, _ = vehiculoPlatePatternByCountry(cfg.PaisCodigo)
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("patente_regex invalido en configuracion")
	}
	value := canonicalPatenteVehiculo(patente)
	if value == "" {
		return fmt.Errorf("patente es obligatoria")
	}
	if !re.MatchString(value) {
		desc := strings.TrimSpace(cfg.PatenteDescripcion)
		if desc == "" {
			_, desc = vehiculoPlatePatternByCountry(cfg.PaisCodigo)
		}
		if desc != "" {
			return fmt.Errorf("la patente no cumple el formato configurado para %s (%s)", cfg.PaisCodigo, desc)
		}
		return fmt.Errorf("la patente no cumple el formato configurado para %s", cfg.PaisCodigo)
	}
	return nil
}

func existsVehiculoActivoPatente(dbConn *sql.DB, empresaID int64, patente string, excludeID int64) (bool, error) {
	canonical := canonicalPatenteVehiculo(patente)
	if canonical == "" {
		return false, nil
	}

	query := `SELECT COUNT(1)
	FROM empresa_vehiculos_registro
	WHERE empresa_id = ?
		AND COALESCE(estado, 'activo') = 'activo'
		AND COALESCE(estado_registro, 'en_empresa') = 'en_empresa'
		AND UPPER(REPLACE(REPLACE(REPLACE(REPLACE(COALESCE(patente, ''), '-', ''), ' ', ''), '.', ''), '_', '')) = ?`
	args := []interface{}{empresaID, canonical}
	if excludeID > 0 {
		query += ` AND id <> ?`
		args = append(args, excludeID)
	}

	var total int64
	if err := dbConn.QueryRow(query, args...).Scan(&total); err != nil {
		return false, err
	}
	return total > 0, nil
}

func normalizeVehiculoEstado(raw string) string {
	v := strings.ToLower(strings.TrimSpace(raw))
	if v == "inactivo" {
		return "inactivo"
	}
	return "activo"
}

func isVehiculoActivoEnPatio(estado, estadoRegistro string) bool {
	return normalizeVehiculoEstado(estado) == "activo" && normalizeEstadoRegistroVehiculo(estadoRegistro) == "en_empresa"
}

func vehiculoBoolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

func roundVehiculo2(v float64) float64 {
	if v < 0 {
		v = 0
	}
	return float64(int64(v*100+0.5)) / 100
}

func normalizeTipoVehiculo(raw string) string {
	value := strings.ToLower(strings.TrimSpace(raw))
	switch value {
	case "automovil", "moto", "camion", "camioneta", "bus", "van", "bicicleta", "otro":
		return value
	default:
		if value == "" {
			return "automovil"
		}
		return value
	}
}

func normalizeEstadoRegistroVehiculo(raw string) string {
	value := strings.ToLower(strings.TrimSpace(raw))
	switch value {
	case "en_empresa", "retirado":
		return value
	default:
		return "en_empresa"
	}
}

func normalizeVehiculopcs_ts(raw string, allowNow bool) (string, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		if allowNow {
			return time.Now().Format("2006-01-02 15:04:05"), nil
		}
		return "", nil
	}
	layouts := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04",
		"2006-01-02T15:04",
		"2006-01-02",
		time.RFC3339,
	}
	for _, layout := range layouts {
		if parsed, err := time.ParseInLocation(layout, value, time.Local); err == nil {
			return parsed.Format("2006-01-02 15:04:05"), nil
		}
	}
	return "", fmt.Errorf("fecha/hora invalida (use YYYY-MM-DD o YYYY-MM-DD HH:MM[:SS])")
}
