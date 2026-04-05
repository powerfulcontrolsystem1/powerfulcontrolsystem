package db

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// EmpresaAsistenciaEmpleado representa el registro de asistencia diario de un colaborador.
type EmpresaAsistenciaEmpleado struct {
	ID                 int64   `json:"id"`
	EmpresaID          int64   `json:"empresa_id"`
	EmpleadoID         int64   `json:"empleado_id"`
	EmpleadoCodigo     string  `json:"empleado_codigo,omitempty"`
	EmpleadoNombre     string  `json:"empleado_nombre"`
	EmpleadoDocumento  string  `json:"empleado_documento,omitempty"`
	Cargo              string  `json:"cargo,omitempty"`
	Turno              string  `json:"turno,omitempty"`
	FechaAsistencia    string  `json:"fecha_asistencia,omitempty"`
	HoraEntrada        string  `json:"hora_entrada,omitempty"`
	HoraSalida         string  `json:"hora_salida,omitempty"`
	MinutosTarde       int     `json:"minutos_tarde,omitempty"`
	HorasTrabajadas    float64 `json:"horas_trabajadas,omitempty"`
	EstadoAsistencia   string  `json:"estado_asistencia,omitempty"`
	Novedad            string  `json:"novedad,omitempty"`
	FechaCreacion      string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string  `json:"usuario_creador,omitempty"`
	Estado             string  `json:"estado,omitempty"`
	Observaciones      string  `json:"observaciones,omitempty"`
}

// EnsureEmpresaAsistenciaSchema crea y migra la tabla de asistencia de empleados por empresa.
func EnsureEmpresaAsistenciaSchema(dbConn *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_asistencia_empleados (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			empleado_id INTEGER DEFAULT 0,
			empleado_codigo TEXT,
			empleado_nombre TEXT NOT NULL,
			empleado_documento TEXT,
			cargo TEXT,
			turno TEXT,
			fecha_asistencia TEXT DEFAULT (date('now','localtime')),
			hora_entrada TEXT,
			hora_salida TEXT,
			minutos_tarde INTEGER DEFAULT 0,
			horas_trabajadas REAL DEFAULT 0,
			estado_asistencia TEXT DEFAULT 'pendiente',
			novedad TEXT,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_asistencia_empleados_empresa_fecha ON empresa_asistencia_empleados(empresa_id, fecha_asistencia DESC, id DESC);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_asistencia_empleados_empresa_empleado ON empresa_asistencia_empleados(empresa_id, empleado_documento, empleado_nombre);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_asistencia_empleados_empresa_estado ON empresa_asistencia_empleados(empresa_id, estado, estado_asistencia);`,
	}
	for _, stmt := range stmts {
		if _, err := dbConn.Exec(stmt); err != nil {
			return err
		}
	}

	if err := ensureColumnIfMissing(dbConn, "empresa_asistencia_empleados", "empleado_id", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_asistencia_empleados", "empleado_codigo", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_asistencia_empleados", "empleado_nombre", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_asistencia_empleados", "empleado_documento", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_asistencia_empleados", "cargo", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_asistencia_empleados", "turno", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_asistencia_empleados", "fecha_asistencia", "TEXT DEFAULT (date('now','localtime'))"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_asistencia_empleados", "hora_entrada", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_asistencia_empleados", "hora_salida", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_asistencia_empleados", "minutos_tarde", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_asistencia_empleados", "horas_trabajadas", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_asistencia_empleados", "estado_asistencia", "TEXT DEFAULT 'pendiente'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_asistencia_empleados", "novedad", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_asistencia_empleados", "fecha_actualizacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_asistencia_empleados", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_asistencia_empleados", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_asistencia_empleados", "observaciones", "TEXT"); err != nil {
		return err
	}

	return nil
}

// CreateEmpresaAsistenciaEmpleado crea un registro de asistencia para una empresa.
func CreateEmpresaAsistenciaEmpleado(dbConn *sql.DB, item EmpresaAsistenciaEmpleado) (int64, error) {
	if item.EmpresaID <= 0 {
		return 0, fmt.Errorf("empresa_id es obligatorio")
	}
	item.EmpleadoNombre = strings.TrimSpace(item.EmpleadoNombre)
	if item.EmpleadoNombre == "" {
		return 0, fmt.Errorf("empleado_nombre es obligatorio")
	}

	fechaAsistencia, err := normalizeAsistenciaDate(item.FechaAsistencia)
	if err != nil {
		return 0, err
	}
	horaEntrada, err := normalizeAsistenciaTime(item.HoraEntrada)
	if err != nil {
		return 0, err
	}
	horaSalida, err := normalizeAsistenciaTime(item.HoraSalida)
	if err != nil {
		return 0, err
	}
	if item.MinutosTarde < 0 {
		item.MinutosTarde = 0
	}
	if item.HorasTrabajadas < 0 {
		item.HorasTrabajadas = 0
	}
	estadoAsistencia := normalizeEstadoAsistencia(item.EstadoAsistencia)
	turno := normalizeTurnoAsistencia(item.Turno)

	res, err := dbConn.Exec(`INSERT INTO empresa_asistencia_empleados (
		empresa_id, empleado_id, empleado_codigo, empleado_nombre, empleado_documento,
		cargo, turno, fecha_asistencia, hora_entrada, hora_salida,
		minutos_tarde, horas_trabajadas, estado_asistencia, novedad,
		usuario_creador, estado, observaciones,
		fecha_creacion, fecha_actualizacion
	) VALUES (
		?, ?, ?, ?, ?,
		?, ?, ?, ?, ?,
		?, ?, ?, ?,
		?, COALESCE(NULLIF(?, ''), 'activo'), ?,
		datetime('now','localtime'), datetime('now','localtime')
	)`,
		item.EmpresaID,
		item.EmpleadoID,
		strings.TrimSpace(item.EmpleadoCodigo),
		item.EmpleadoNombre,
		strings.TrimSpace(item.EmpleadoDocumento),
		strings.TrimSpace(item.Cargo),
		turno,
		fechaAsistencia,
		horaEntrada,
		horaSalida,
		item.MinutosTarde,
		item.HorasTrabajadas,
		estadoAsistencia,
		strings.TrimSpace(item.Novedad),
		strings.TrimSpace(item.UsuarioCreador),
		strings.TrimSpace(item.Estado),
		strings.TrimSpace(item.Observaciones),
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// ListEmpresaAsistenciaEmpleados lista asistencias por empresa con filtros operativos.
func ListEmpresaAsistenciaEmpleados(dbConn *sql.DB, empresaID int64, includeInactive bool, desde, hasta, estadoAsistencia, q string, limit int) ([]EmpresaAsistenciaEmpleado, error) {
	query := `SELECT
		id, empresa_id, COALESCE(empleado_id, 0), COALESCE(empleado_codigo, ''), COALESCE(empleado_nombre, ''),
		COALESCE(empleado_documento, ''), COALESCE(cargo, ''), COALESCE(turno, ''),
		COALESCE(fecha_asistencia, ''), COALESCE(hora_entrada, ''), COALESCE(hora_salida, ''),
		COALESCE(minutos_tarde, 0), COALESCE(horas_trabajadas, 0), COALESCE(estado_asistencia, 'pendiente'),
		COALESCE(novedad, ''), COALESCE(fecha_creacion, ''), COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''), COALESCE(estado, 'activo'), COALESCE(observaciones, '')
	FROM empresa_asistencia_empleados
	WHERE empresa_id = ?`
	args := []interface{}{empresaID}

	if !includeInactive {
		query += ` AND estado = 'activo'`
	}

	desde = strings.TrimSpace(desde)
	if desde != "" {
		query += ` AND fecha_asistencia >= ?`
		args = append(args, desde)
	}
	hasta = strings.TrimSpace(hasta)
	if hasta != "" {
		query += ` AND fecha_asistencia <= ?`
		args = append(args, hasta)
	}

	estadoAsistencia = strings.TrimSpace(strings.ToLower(estadoAsistencia))
	if estadoAsistencia != "" {
		query += ` AND LOWER(COALESCE(estado_asistencia,'')) = ?`
		args = append(args, estadoAsistencia)
	}

	q = strings.TrimSpace(strings.ToLower(q))
	if q != "" {
		query += ` AND (
			LOWER(COALESCE(empleado_codigo, '')) LIKE ?
			OR LOWER(COALESCE(empleado_nombre, '')) LIKE ?
			OR LOWER(COALESCE(empleado_documento, '')) LIKE ?
			OR LOWER(COALESCE(cargo, '')) LIKE ?
			OR LOWER(COALESCE(turno, '')) LIKE ?
		)`
		like := "%" + q + "%"
		args = append(args, like, like, like, like, like)
	}

	if limit <= 0 {
		limit = 300
	}
	if limit > 2000 {
		limit = 2000
	}
	query += ` ORDER BY fecha_asistencia DESC, hora_entrada DESC, id DESC LIMIT ?`
	args = append(args, limit)

	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]EmpresaAsistenciaEmpleado, 0)
	for rows.Next() {
		var item EmpresaAsistenciaEmpleado
		if err := rows.Scan(
			&item.ID,
			&item.EmpresaID,
			&item.EmpleadoID,
			&item.EmpleadoCodigo,
			&item.EmpleadoNombre,
			&item.EmpleadoDocumento,
			&item.Cargo,
			&item.Turno,
			&item.FechaAsistencia,
			&item.HoraEntrada,
			&item.HoraSalida,
			&item.MinutosTarde,
			&item.HorasTrabajadas,
			&item.EstadoAsistencia,
			&item.Novedad,
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

// UpdateEmpresaAsistenciaEmpleado actualiza un registro de asistencia existente.
func UpdateEmpresaAsistenciaEmpleado(dbConn *sql.DB, item EmpresaAsistenciaEmpleado) error {
	if item.EmpresaID <= 0 || item.ID <= 0 {
		return fmt.Errorf("empresa_id e id son obligatorios")
	}
	item.EmpleadoNombre = strings.TrimSpace(item.EmpleadoNombre)
	if item.EmpleadoNombre == "" {
		return fmt.Errorf("empleado_nombre es obligatorio")
	}

	fechaAsistencia, err := normalizeAsistenciaDate(item.FechaAsistencia)
	if err != nil {
		return err
	}
	horaEntrada, err := normalizeAsistenciaTime(item.HoraEntrada)
	if err != nil {
		return err
	}
	horaSalida, err := normalizeAsistenciaTime(item.HoraSalida)
	if err != nil {
		return err
	}
	if item.MinutosTarde < 0 {
		item.MinutosTarde = 0
	}
	if item.HorasTrabajadas < 0 {
		item.HorasTrabajadas = 0
	}

	res, err := dbConn.Exec(`UPDATE empresa_asistencia_empleados
	SET
		empleado_id = ?,
		empleado_codigo = ?,
		empleado_nombre = ?,
		empleado_documento = ?,
		cargo = ?,
		turno = ?,
		fecha_asistencia = ?,
		hora_entrada = ?,
		hora_salida = ?,
		minutos_tarde = ?,
		horas_trabajadas = ?,
		estado_asistencia = ?,
		novedad = ?,
		observaciones = ?,
		fecha_actualizacion = datetime('now','localtime')
	WHERE empresa_id = ? AND id = ?`,
		item.EmpleadoID,
		strings.TrimSpace(item.EmpleadoCodigo),
		item.EmpleadoNombre,
		strings.TrimSpace(item.EmpleadoDocumento),
		strings.TrimSpace(item.Cargo),
		normalizeTurnoAsistencia(item.Turno),
		fechaAsistencia,
		horaEntrada,
		horaSalida,
		item.MinutosTarde,
		item.HorasTrabajadas,
		normalizeEstadoAsistencia(item.EstadoAsistencia),
		strings.TrimSpace(item.Novedad),
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

// SetEmpresaAsistenciaEmpleadoEstado activa o desactiva un registro de asistencia.
func SetEmpresaAsistenciaEmpleadoEstado(dbConn *sql.DB, empresaID, id int64, estado string) error {
	estado = strings.ToLower(strings.TrimSpace(estado))
	if estado != "activo" && estado != "inactivo" {
		estado = "activo"
	}

	res, err := dbConn.Exec(`UPDATE empresa_asistencia_empleados
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

// MarkEmpresaAsistenciaEntrada marca la hora de entrada para un registro existente.
func MarkEmpresaAsistenciaEntrada(dbConn *sql.DB, empresaID, id int64, horaEntrada string, minutosTarde int, estadoAsistencia, novedad string) error {
	parsedTime, err := normalizeAsistenciaTime(horaEntrada)
	if err != nil {
		return err
	}
	if parsedTime == "" {
		parsedTime = time.Now().Format("15:04:05")
	}
	if minutosTarde < 0 {
		minutosTarde = 0
	}
	estadoAsistencia = normalizeEstadoAsistencia(strings.TrimSpace(estadoAsistencia))
	if strings.TrimSpace(estadoAsistencia) == "" {
		if minutosTarde > 0 {
			estadoAsistencia = "tarde"
		} else {
			estadoAsistencia = "presente"
		}
	}

	res, err := dbConn.Exec(`UPDATE empresa_asistencia_empleados
	SET
		hora_entrada = ?,
		minutos_tarde = ?,
		estado_asistencia = ?,
		novedad = CASE WHEN ? <> '' THEN ? ELSE novedad END,
		fecha_actualizacion = datetime('now','localtime')
	WHERE empresa_id = ? AND id = ?`,
		parsedTime,
		minutosTarde,
		estadoAsistencia,
		strings.TrimSpace(novedad),
		strings.TrimSpace(novedad),
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

// MarkEmpresaAsistenciaSalida marca salida y calcula horas trabajadas si hay hora de entrada.
func MarkEmpresaAsistenciaSalida(dbConn *sql.DB, empresaID, id int64, horaSalida, novedad string) error {
	parsedTime, err := normalizeAsistenciaTime(horaSalida)
	if err != nil {
		return err
	}
	if parsedTime == "" {
		parsedTime = time.Now().Format("15:04:05")
	}

	var fechaAsistencia string
	var horaEntrada string
	var estadoActual string
	err = dbConn.QueryRow(`SELECT
		COALESCE(fecha_asistencia, ''),
		COALESCE(hora_entrada, ''),
		COALESCE(estado_asistencia, 'pendiente')
	FROM empresa_asistencia_empleados
	WHERE empresa_id = ? AND id = ?
	LIMIT 1`, empresaID, id).Scan(&fechaAsistencia, &horaEntrada, &estadoActual)
	if err != nil {
		return err
	}

	horasTrabajadas := 0.0
	if strings.TrimSpace(horaEntrada) != "" {
		if hours, calcErr := calculateWorkedHours(fechaAsistencia, horaEntrada, parsedTime); calcErr == nil {
			horasTrabajadas = hours
		}
	}

	estadoAsistencia := normalizeEstadoAsistencia(estadoActual)
	if estadoAsistencia == "pendiente" {
		estadoAsistencia = "presente"
	}

	res, err := dbConn.Exec(`UPDATE empresa_asistencia_empleados
	SET
		hora_salida = ?,
		horas_trabajadas = ?,
		estado_asistencia = ?,
		novedad = CASE WHEN ? <> '' THEN ? ELSE novedad END,
		fecha_actualizacion = datetime('now','localtime')
	WHERE empresa_id = ? AND id = ?`,
		parsedTime,
		horasTrabajadas,
		estadoAsistencia,
		strings.TrimSpace(novedad),
		strings.TrimSpace(novedad),
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

// DeleteEmpresaAsistenciaEmpleado elimina un registro de asistencia por empresa.
func DeleteEmpresaAsistenciaEmpleado(dbConn *sql.DB, empresaID, id int64) error {
	res, err := dbConn.Exec(`DELETE FROM empresa_asistencia_empleados WHERE empresa_id = ? AND id = ?`, empresaID, id)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func normalizeAsistenciaDate(raw string) (string, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return time.Now().Format("2006-01-02"), nil
	}
	if len(value) >= 10 {
		candidate := value[:10]
		if _, err := time.Parse("2006-01-02", candidate); err == nil {
			return candidate, nil
		}
	}
	layouts := []string{time.RFC3339, "2006-01-02 15:04:05", "2006-01-02T15:04:05", "2006-01-02"}
	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed.Format("2006-01-02"), nil
		}
	}
	return "", fmt.Errorf("fecha_asistencia invalida (use YYYY-MM-DD)")
}

func normalizeAsistenciaTime(raw string) (string, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "", nil
	}
	if len(value) >= 8 {
		candidate := value[:8]
		if parsed, err := time.Parse("15:04:05", candidate); err == nil {
			return parsed.Format("15:04:05"), nil
		}
	}
	if parsed, err := time.Parse("15:04", value); err == nil {
		return parsed.Format("15:04:05"), nil
	}
	layouts := []string{time.RFC3339, "2006-01-02 15:04:05", "2006-01-02T15:04:05"}
	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed.Format("15:04:05"), nil
		}
	}
	return "", fmt.Errorf("hora invalida (use HH:MM o HH:MM:SS)")
}

func normalizeEstadoAsistencia(raw string) string {
	value := strings.ToLower(strings.TrimSpace(raw))
	switch value {
	case "pendiente", "presente", "tarde", "ausente", "permiso", "incapacidad", "vacaciones":
		return value
	default:
		return "pendiente"
	}
}

func normalizeTurnoAsistencia(raw string) string {
	value := strings.ToLower(strings.TrimSpace(raw))
	switch value {
	case "manana", "tarde", "noche", "mixto", "rotativo", "general":
		return value
	default:
		if value == "" {
			return "general"
		}
		return value
	}
}

func calculateWorkedHours(fechaAsistencia, horaEntrada, horaSalida string) (float64, error) {
	fecha, err := normalizeAsistenciaDate(fechaAsistencia)
	if err != nil {
		return 0, err
	}
	entrada, err := normalizeAsistenciaTime(horaEntrada)
	if err != nil {
		return 0, err
	}
	salida, err := normalizeAsistenciaTime(horaSalida)
	if err != nil {
		return 0, err
	}
	if entrada == "" || salida == "" {
		return 0, nil
	}

	start, err := time.Parse("2006-01-02 15:04:05", fecha+" "+entrada)
	if err != nil {
		return 0, err
	}
	end, err := time.Parse("2006-01-02 15:04:05", fecha+" "+salida)
	if err != nil {
		return 0, err
	}
	if end.Before(start) {
		end = end.Add(24 * time.Hour)
	}
	hours := end.Sub(start).Hours()
	if hours < 0 {
		hours = 0
	}
	if hours > 24 {
		hours = 24
	}
	return float64(int64(hours*100+0.5)) / 100, nil
}
