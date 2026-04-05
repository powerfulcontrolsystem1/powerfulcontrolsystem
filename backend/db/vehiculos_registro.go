package db

import (
	"database/sql"
	"fmt"
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

// EnsureEmpresaVehiculosRegistroSchema crea/migra la tabla de registro de vehiculos por empresa.
func EnsureEmpresaVehiculosRegistroSchema(dbConn *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_vehiculos_registro (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
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
			fecha_ingreso TEXT DEFAULT (datetime('now','localtime')),
			fecha_salida TEXT,
			estado_registro TEXT DEFAULT 'en_empresa',
			usuario_salida TEXT,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_vehiculos_registro_empresa_fecha ON empresa_vehiculos_registro(empresa_id, fecha_ingreso DESC, id DESC);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_vehiculos_registro_empresa_patente ON empresa_vehiculos_registro(empresa_id, patente);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_vehiculos_registro_empresa_estado ON empresa_vehiculos_registro(empresa_id, estado, estado_registro);`,
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
	if err := ensureColumnIfMissing(dbConn, "empresa_vehiculos_registro", "fecha_ingreso", "TEXT DEFAULT (datetime('now','localtime'))"); err != nil {
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

	return nil
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

	fechaIngreso, err := normalizeVehiculoDateTime(item.FechaIngreso, true)
	if err != nil {
		return 0, err
	}
	fechaSalida, err := normalizeVehiculoDateTime(item.FechaSalida, false)
	if err != nil {
		return 0, err
	}
	estadoRegistro := normalizeEstadoRegistroVehiculo(item.EstadoRegistro)
	if estadoRegistro == "retirado" && fechaSalida == "" {
		fechaSalida = time.Now().Format("2006-01-02 15:04:05")
	}

	res, err := dbConn.Exec(`INSERT INTO empresa_vehiculos_registro (
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
		datetime('now','localtime'), datetime('now','localtime')
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
		strings.TrimSpace(item.Estado),
		strings.TrimSpace(item.Observaciones),
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
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

// UpdateEmpresaVehiculoRegistro actualiza un registro de vehiculo existente.
func UpdateEmpresaVehiculoRegistro(dbConn *sql.DB, item EmpresaVehiculoRegistro) error {
	if item.EmpresaID <= 0 || item.ID <= 0 {
		return fmt.Errorf("empresa_id e id son obligatorios")
	}
	item.Patente = normalizePatenteVehiculo(item.Patente)
	if item.Patente == "" {
		return fmt.Errorf("patente es obligatoria")
	}

	fechaIngreso, err := normalizeVehiculoDateTime(item.FechaIngreso, true)
	if err != nil {
		return err
	}
	fechaSalida, err := normalizeVehiculoDateTime(item.FechaSalida, false)
	if err != nil {
		return err
	}
	estadoRegistro := normalizeEstadoRegistroVehiculo(item.EstadoRegistro)
	if estadoRegistro == "retirado" && fechaSalida == "" {
		fechaSalida = time.Now().Format("2006-01-02 15:04:05")
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
		fecha_actualizacion = datetime('now','localtime')
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
	estado = strings.ToLower(strings.TrimSpace(estado))
	if estado != "activo" && estado != "inactivo" {
		estado = "activo"
	}

	res, err := dbConn.Exec(`UPDATE empresa_vehiculos_registro
	SET estado = ?, fecha_actualizacion = datetime('now','localtime')
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
	fechaSalidaNorm, err := normalizeVehiculoDateTime(fechaSalida, false)
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
		fecha_actualizacion = datetime('now','localtime')
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

func normalizeVehiculoDateTime(raw string, allowNow bool) (string, error) {
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
