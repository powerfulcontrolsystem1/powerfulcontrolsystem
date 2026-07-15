package db

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
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

// EmpresaAsistenciaConfiguracion representa reglas de asistencia por empresa.
type EmpresaAsistenciaConfiguracion struct {
	ID                    int64  `json:"id"`
	EmpresaID             int64  `json:"empresa_id"`
	ToleranciaEntradaMin  int    `json:"tolerancia_entrada_minutos"`
	ToleranciaSalidaMin   int    `json:"tolerancia_salida_minutos"`
	HoraInicioTurnoManana string `json:"hora_inicio_turno_manana,omitempty"`
	HoraInicioTurnoTarde  string `json:"hora_inicio_turno_tarde,omitempty"`
	HoraInicioTurnoNoche  string `json:"hora_inicio_turno_noche,omitempty"`
	PermitirTurnoNocturno bool   `json:"permitir_turno_nocturno"`
	PermitirTurnoCruzado  bool   `json:"permitir_turno_cruzado"`
	FechaCreacion         string `json:"fecha_creacion,omitempty"`
	FechaActualizacion    string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador        string `json:"usuario_creador,omitempty"`
	Estado                string `json:"estado,omitempty"`
	Observaciones         string `json:"observaciones,omitempty"`
}

// EmpresaAsistenciaPeriodoCierre representa un periodo bloqueado para edicion de asistencia.
type EmpresaAsistenciaPeriodoCierre struct {
	ID                 int64  `json:"id"`
	EmpresaID          int64  `json:"empresa_id"`
	PeriodoDesde       string `json:"periodo_desde"`
	PeriodoHasta       string `json:"periodo_hasta"`
	FechaCierre        string `json:"fecha_cierre,omitempty"`
	CerradoPor         string `json:"cerrado_por,omitempty"`
	Motivo             string `json:"motivo,omitempty"`
	FechaCreacion      string `json:"fecha_creacion,omitempty"`
	FechaActualizacion string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string `json:"usuario_creador,omitempty"`
	Estado             string `json:"estado,omitempty"`
	Observaciones      string `json:"observaciones,omitempty"`
}

var (
	ErrAsistenciaPeriodoCerrado  = errors.New("el periodo de asistencia esta cerrado")
	ErrAsistenciaPeriodoSolapado = errors.New("el periodo de cierre se solapa con uno existente")
)

// EnsureEmpresaAsistenciaSchema crea y migra la tabla de asistencia de empleados por empresa.
func EnsureEmpresaAsistenciaSchema(dbConn *sql.DB) error {
	if SchemaBootstrapDisabled() {
		return nil
	}
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_asistencia_empleados (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			empleado_id INTEGER DEFAULT 0,
			empleado_codigo TEXT,
			empleado_nombre TEXT NOT NULL,
			empleado_documento TEXT,
			cargo TEXT,
			turno TEXT,
			fecha_asistencia TEXT DEFAULT (CURRENT_DATE),
			hora_entrada TEXT,
			hora_salida TEXT,
			minutos_tarde INTEGER DEFAULT 0,
			horas_trabajadas REAL DEFAULT 0,
			estado_asistencia TEXT DEFAULT 'pendiente',
			novedad TEXT,
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_asistencia_empleados_empresa_fecha ON empresa_asistencia_empleados(empresa_id, fecha_asistencia DESC, id DESC);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_asistencia_empleados_empresa_empleado ON empresa_asistencia_empleados(empresa_id, empleado_documento, empleado_nombre);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_asistencia_empleados_empresa_estado ON empresa_asistencia_empleados(empresa_id, estado, estado_asistencia);`,
		`CREATE TABLE IF NOT EXISTS empresa_asistencia_configuracion (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL UNIQUE,
			tolerancia_entrada_minutos INTEGER DEFAULT 10,
			tolerancia_salida_minutos INTEGER DEFAULT 0,
			hora_inicio_turno_manana TEXT DEFAULT '06:00:00',
			hora_inicio_turno_tarde TEXT DEFAULT '14:00:00',
			hora_inicio_turno_noche TEXT DEFAULT '22:00:00',
			permitir_turno_nocturno INTEGER DEFAULT 1,
			permitir_turno_cruzado INTEGER DEFAULT 1,
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_empresa_asistencia_configuracion_empresa ON empresa_asistencia_configuracion(empresa_id);`,
		`CREATE TABLE IF NOT EXISTS empresa_asistencia_periodos_cerrados (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			periodo_desde TEXT NOT NULL,
			periodo_hasta TEXT NOT NULL,
			fecha_cierre TEXT DEFAULT (CURRENT_TIMESTAMP),
			cerrado_por TEXT,
			motivo TEXT,
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			UNIQUE(empresa_id, periodo_desde, periodo_hasta)
		);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_asistencia_cierres_empresa_periodo ON empresa_asistencia_periodos_cerrados(empresa_id, periodo_desde DESC, periodo_hasta DESC);`,
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
	if err := ensureColumnIfMissing(dbConn, "empresa_asistencia_empleados", "fecha_asistencia", "TEXT DEFAULT (CURRENT_DATE)"); err != nil {
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

	if err := ensureColumnIfMissing(dbConn, "empresa_asistencia_configuracion", "tolerancia_entrada_minutos", "INTEGER DEFAULT 10"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_asistencia_configuracion", "tolerancia_salida_minutos", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_asistencia_configuracion", "hora_inicio_turno_manana", "TEXT DEFAULT '06:00:00'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_asistencia_configuracion", "hora_inicio_turno_tarde", "TEXT DEFAULT '14:00:00'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_asistencia_configuracion", "hora_inicio_turno_noche", "TEXT DEFAULT '22:00:00'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_asistencia_configuracion", "permitir_turno_nocturno", "INTEGER DEFAULT 1"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_asistencia_configuracion", "permitir_turno_cruzado", "INTEGER DEFAULT 1"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_asistencia_configuracion", "fecha_actualizacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_asistencia_configuracion", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_asistencia_configuracion", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_asistencia_configuracion", "observaciones", "TEXT"); err != nil {
		return err
	}

	if err := ensureColumnIfMissing(dbConn, "empresa_asistencia_periodos_cerrados", "fecha_cierre", "TEXT DEFAULT (CURRENT_TIMESTAMP)"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_asistencia_periodos_cerrados", "cerrado_por", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_asistencia_periodos_cerrados", "motivo", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_asistencia_periodos_cerrados", "fecha_actualizacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_asistencia_periodos_cerrados", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_asistencia_periodos_cerrados", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_asistencia_periodos_cerrados", "observaciones", "TEXT"); err != nil {
		return err
	}

	return nil
}

func defaultEmpresaAsistenciaConfiguracion(empresaID int64) *EmpresaAsistenciaConfiguracion {
	return &EmpresaAsistenciaConfiguracion{
		EmpresaID:             empresaID,
		ToleranciaEntradaMin:  10,
		ToleranciaSalidaMin:   0,
		HoraInicioTurnoManana: "06:00:00",
		HoraInicioTurnoTarde:  "14:00:00",
		HoraInicioTurnoNoche:  "22:00:00",
		PermitirTurnoNocturno: true,
		PermitirTurnoCruzado:  true,
		Estado:                "activo",
	}
}

func normalizeAsistenciaConfig(cfg EmpresaAsistenciaConfiguracion) (*EmpresaAsistenciaConfiguracion, error) {
	out := defaultEmpresaAsistenciaConfiguracion(cfg.EmpresaID)
	out.ID = cfg.ID
	out.UsuarioCreador = strings.TrimSpace(cfg.UsuarioCreador)
	out.Observaciones = strings.TrimSpace(cfg.Observaciones)
	if strings.TrimSpace(cfg.Estado) != "" {
		out.Estado = strings.ToLower(strings.TrimSpace(cfg.Estado))
	}

	out.ToleranciaEntradaMin = cfg.ToleranciaEntradaMin
	if out.ToleranciaEntradaMin < 0 {
		out.ToleranciaEntradaMin = 0
	}
	if out.ToleranciaEntradaMin > 240 {
		out.ToleranciaEntradaMin = 240
	}

	out.ToleranciaSalidaMin = cfg.ToleranciaSalidaMin
	if out.ToleranciaSalidaMin < 0 {
		out.ToleranciaSalidaMin = 0
	}
	if out.ToleranciaSalidaMin > 240 {
		out.ToleranciaSalidaMin = 240
	}

	manana, err := normalizeAsistenciaTime(cfg.HoraInicioTurnoManana)
	if err != nil {
		return nil, fmt.Errorf("hora_inicio_turno_manana invalida")
	}
	if manana != "" {
		out.HoraInicioTurnoManana = manana
	}

	tarde, err := normalizeAsistenciaTime(cfg.HoraInicioTurnoTarde)
	if err != nil {
		return nil, fmt.Errorf("hora_inicio_turno_tarde invalida")
	}
	if tarde != "" {
		out.HoraInicioTurnoTarde = tarde
	}

	noche, err := normalizeAsistenciaTime(cfg.HoraInicioTurnoNoche)
	if err != nil {
		return nil, fmt.Errorf("hora_inicio_turno_noche invalida")
	}
	if noche != "" {
		out.HoraInicioTurnoNoche = noche
	}

	out.PermitirTurnoNocturno = cfg.PermitirTurnoNocturno
	out.PermitirTurnoCruzado = cfg.PermitirTurnoCruzado
	return out, nil
}

// GetEmpresaAsistenciaConfiguracion obtiene reglas de asistencia por empresa.
func GetEmpresaAsistenciaConfiguracion(dbConn *sql.DB, empresaID int64) (*EmpresaAsistenciaConfiguracion, error) {
	if empresaID <= 0 {
		return nil, fmt.Errorf("empresa_id es obligatorio")
	}

	cfg := defaultEmpresaAsistenciaConfiguracion(empresaID)
	err := queryRowSQLCompat(dbConn, `SELECT
		id,
		empresa_id,
		COALESCE(tolerancia_entrada_minutos, 10),
		COALESCE(tolerancia_salida_minutos, 0),
		COALESCE(hora_inicio_turno_manana, '06:00:00'),
		COALESCE(hora_inicio_turno_tarde, '14:00:00'),
		COALESCE(hora_inicio_turno_noche, '22:00:00'),
		COALESCE(permitir_turno_nocturno, 1),
		COALESCE(permitir_turno_cruzado, 1),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM empresa_asistencia_configuracion
	WHERE empresa_id = ?
	LIMIT 1`, empresaID).Scan(
		&cfg.ID,
		&cfg.EmpresaID,
		&cfg.ToleranciaEntradaMin,
		&cfg.ToleranciaSalidaMin,
		&cfg.HoraInicioTurnoManana,
		&cfg.HoraInicioTurnoTarde,
		&cfg.HoraInicioTurnoNoche,
		&cfg.PermitirTurnoNocturno,
		&cfg.PermitirTurnoCruzado,
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

	normalized, err := normalizeAsistenciaConfig(*cfg)
	if err != nil {
		return nil, err
	}
	normalized.ID = cfg.ID
	normalized.FechaCreacion = cfg.FechaCreacion
	normalized.FechaActualizacion = cfg.FechaActualizacion
	return normalized, nil
}

// UpsertEmpresaAsistenciaConfiguracion crea o actualiza reglas de asistencia por empresa.
func UpsertEmpresaAsistenciaConfiguracion(dbConn *sql.DB, payload EmpresaAsistenciaConfiguracion) (int64, error) {
	if payload.EmpresaID <= 0 {
		return 0, fmt.Errorf("empresa_id es obligatorio")
	}

	cfg, err := normalizeAsistenciaConfig(payload)
	if err != nil {
		return 0, err
	}

	var existingID int64
	err = queryRowSQLCompat(dbConn, `SELECT id FROM empresa_asistencia_configuracion WHERE empresa_id = ? LIMIT 1`, cfg.EmpresaID).Scan(&existingID)
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}

	nowExpr := sqlNowExpr()
	if existingID > 0 {
		_, err = execSQLCompat(dbConn, `UPDATE empresa_asistencia_configuracion
		SET
			tolerancia_entrada_minutos = ?,
			tolerancia_salida_minutos = ?,
			hora_inicio_turno_manana = ?,
			hora_inicio_turno_tarde = ?,
			hora_inicio_turno_noche = ?,
			permitir_turno_nocturno = ?,
			permitir_turno_cruzado = ?,
			usuario_creador = ?,
			estado = ?,
			observaciones = ?,
			fecha_actualizacion = `+nowExpr+`
		WHERE empresa_id = ?`,
			cfg.ToleranciaEntradaMin,
			cfg.ToleranciaSalidaMin,
			cfg.HoraInicioTurnoManana,
			cfg.HoraInicioTurnoTarde,
			cfg.HoraInicioTurnoNoche,
			asistenciaBoolToInt(cfg.PermitirTurnoNocturno),
			asistenciaBoolToInt(cfg.PermitirTurnoCruzado),
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

	id, err := insertSQLCompat(dbConn, `INSERT INTO empresa_asistencia_configuracion (
		empresa_id,
		tolerancia_entrada_minutos,
		tolerancia_salida_minutos,
		hora_inicio_turno_manana,
		hora_inicio_turno_tarde,
		hora_inicio_turno_noche,
		permitir_turno_nocturno,
		permitir_turno_cruzado,
		usuario_creador,
		estado,
		observaciones,
		fecha_creacion,
		fecha_actualizacion
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, `+nowExpr+`, `+nowExpr+`)`,
		cfg.EmpresaID,
		cfg.ToleranciaEntradaMin,
		cfg.ToleranciaSalidaMin,
		cfg.HoraInicioTurnoManana,
		cfg.HoraInicioTurnoTarde,
		cfg.HoraInicioTurnoNoche,
		asistenciaBoolToInt(cfg.PermitirTurnoNocturno),
		asistenciaBoolToInt(cfg.PermitirTurnoCruzado),
		cfg.UsuarioCreador,
		cfg.Estado,
		cfg.Observaciones,
	)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// ListEmpresaAsistenciaPeriodosCerrados lista cierres de periodo registrados para una empresa.
func ListEmpresaAsistenciaPeriodosCerrados(dbConn *sql.DB, empresaID int64, includeInactive bool, desde, hasta string, limit int) ([]EmpresaAsistenciaPeriodoCierre, error) {
	if empresaID <= 0 {
		return nil, fmt.Errorf("empresa_id es obligatorio")
	}

	query := `SELECT
		id,
		empresa_id,
		COALESCE(periodo_desde, ''),
		COALESCE(periodo_hasta, ''),
		COALESCE(fecha_cierre, ''),
		COALESCE(cerrado_por, ''),
		COALESCE(motivo, ''),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM empresa_asistencia_periodos_cerrados
	WHERE empresa_id = ?`
	args := []interface{}{empresaID}

	if !includeInactive {
		query += ` AND LOWER(COALESCE(estado, 'activo')) = 'activo'`
	}

	if strings.TrimSpace(desde) != "" {
		desdeNorm, err := normalizeAsistenciaDate(desde)
		if err != nil {
			return nil, err
		}
		query += ` AND periodo_hasta >= ?`
		args = append(args, desdeNorm)
	}
	if strings.TrimSpace(hasta) != "" {
		hastaNorm, err := normalizeAsistenciaDate(hasta)
		if err != nil {
			return nil, err
		}
		query += ` AND periodo_desde <= ?`
		args = append(args, hastaNorm)
	}

	if limit <= 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = 1000
	}

	query += ` ORDER BY periodo_desde DESC, periodo_hasta DESC, id DESC LIMIT ?`
	args = append(args, limit)

	rows, err := querySQLCompat(dbConn, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]EmpresaAsistenciaPeriodoCierre, 0)
	for rows.Next() {
		var item EmpresaAsistenciaPeriodoCierre
		if err := rows.Scan(
			&item.ID,
			&item.EmpresaID,
			&item.PeriodoDesde,
			&item.PeriodoHasta,
			&item.FechaCierre,
			&item.CerradoPor,
			&item.Motivo,
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

// CreateEmpresaAsistenciaPeriodoCierre registra un cierre de periodo para bloquear ediciones posteriores.
func CreateEmpresaAsistenciaPeriodoCierre(dbConn *sql.DB, payload EmpresaAsistenciaPeriodoCierre) (int64, error) {
	if payload.EmpresaID <= 0 {
		return 0, fmt.Errorf("empresa_id es obligatorio")
	}
	desde, err := normalizeAsistenciaDate(payload.PeriodoDesde)
	if err != nil {
		return 0, fmt.Errorf("periodo_desde invalido")
	}
	hasta, err := normalizeAsistenciaDate(payload.PeriodoHasta)
	if err != nil {
		return 0, fmt.Errorf("periodo_hasta invalido")
	}
	if hasta < desde {
		return 0, fmt.Errorf("periodo_hasta no puede ser menor que periodo_desde")
	}

	var overlap EmpresaAsistenciaPeriodoCierre
	err = queryRowSQLCompat(dbConn, `SELECT
		id,
		COALESCE(periodo_desde, ''),
		COALESCE(periodo_hasta, ''),
		COALESCE(cerrado_por, ''),
		COALESCE(fecha_cierre, ''),
		COALESCE(motivo, '')
	FROM empresa_asistencia_periodos_cerrados
	WHERE empresa_id = ?
		AND LOWER(COALESCE(estado, 'activo')) = 'activo'
		AND NOT (periodo_hasta < ? OR periodo_desde > ?)
	LIMIT 1`, payload.EmpresaID, desde, hasta).Scan(
		&overlap.ID,
		&overlap.PeriodoDesde,
		&overlap.PeriodoHasta,
		&overlap.CerradoPor,
		&overlap.FechaCierre,
		&overlap.Motivo,
	)
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}
	if err == nil && overlap.ID > 0 {
		return 0, fmt.Errorf("%w: [%s a %s]", ErrAsistenciaPeriodoSolapado, overlap.PeriodoDesde, overlap.PeriodoHasta)
	}

	nowExpr := sqlNowExpr()
	id, err := insertSQLCompat(dbConn, `INSERT INTO empresa_asistencia_periodos_cerrados (
		empresa_id,
		periodo_desde,
		periodo_hasta,
		fecha_cierre,
		cerrado_por,
		motivo,
		usuario_creador,
		estado,
		observaciones,
		fecha_creacion,
		fecha_actualizacion
	) VALUES (
		?, ?, ?, `+nowExpr+`, ?, ?, ?, COALESCE(NULLIF(?, ''), 'activo'), ?, `+nowExpr+`, `+nowExpr+`
	)`,
		payload.EmpresaID,
		desde,
		hasta,
		strings.TrimSpace(payload.CerradoPor),
		strings.TrimSpace(payload.Motivo),
		strings.TrimSpace(payload.UsuarioCreador),
		strings.TrimSpace(payload.Estado),
		strings.TrimSpace(payload.Observaciones),
	)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func getAsistenciaCierreByFecha(dbConn *sql.DB, empresaID int64, fechaAsistencia string) (*EmpresaAsistenciaPeriodoCierre, error) {
	item := &EmpresaAsistenciaPeriodoCierre{}
	err := queryRowSQLCompat(dbConn, `SELECT
		id,
		empresa_id,
		COALESCE(periodo_desde, ''),
		COALESCE(periodo_hasta, ''),
		COALESCE(fecha_cierre, ''),
		COALESCE(cerrado_por, ''),
		COALESCE(motivo, '')
	FROM empresa_asistencia_periodos_cerrados
	WHERE empresa_id = ?
		AND LOWER(COALESCE(estado, 'activo')) = 'activo'
		AND periodo_desde <= ?
		AND periodo_hasta >= ?
	ORDER BY periodo_desde DESC, id DESC
	LIMIT 1`, empresaID, fechaAsistencia, fechaAsistencia).Scan(
		&item.ID,
		&item.EmpresaID,
		&item.PeriodoDesde,
		&item.PeriodoHasta,
		&item.FechaCierre,
		&item.CerradoPor,
		&item.Motivo,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return item, nil
}

func validateAsistenciaFechaNoCerrada(dbConn *sql.DB, empresaID int64, fechaAsistencia string) error {
	fecha, err := normalizeAsistenciaDate(fechaAsistencia)
	if err != nil {
		return err
	}
	cierre, err := getAsistenciaCierreByFecha(dbConn, empresaID, fecha)
	if err != nil {
		return err
	}
	if cierre == nil {
		return nil
	}
	return fmt.Errorf("%w: fecha %s bloqueada por cierre [%s a %s]", ErrAsistenciaPeriodoCerrado, fecha, cierre.PeriodoDesde, cierre.PeriodoHasta)
}

func getAsistenciaRegistroContext(dbConn *sql.DB, empresaID, id int64) (string, string, string, string, error) {
	var fechaAsistencia string
	var turno string
	var horaEntrada string
	var horaSalida string
	err := queryRowSQLCompat(dbConn, `SELECT
		COALESCE(fecha_asistencia, ''),
		COALESCE(turno, 'general'),
		COALESCE(hora_entrada, ''),
		COALESCE(hora_salida, '')
	FROM empresa_asistencia_empleados
	WHERE empresa_id = ? AND id = ?
	LIMIT 1`, empresaID, id).Scan(&fechaAsistencia, &turno, &horaEntrada, &horaSalida)
	if err != nil {
		return "", "", "", "", err
	}
	return fechaAsistencia, normalizeTurnoAsistencia(turno), horaEntrada, horaSalida, nil
}

func resolveTurnoInicio(cfg *EmpresaAsistenciaConfiguracion, turno string) string {
	switch normalizeTurnoAsistencia(turno) {
	case "manana":
		return cfg.HoraInicioTurnoManana
	case "tarde":
		return cfg.HoraInicioTurnoTarde
	case "noche":
		return cfg.HoraInicioTurnoNoche
	default:
		return ""
	}
}

func timeToMinutes(raw string) (int, error) {
	parsed, err := normalizeAsistenciaTime(raw)
	if err != nil {
		return 0, err
	}
	if parsed == "" {
		return 0, nil
	}
	parts := strings.Split(parsed, ":")
	if len(parts) < 2 {
		return 0, fmt.Errorf("hora invalida")
	}
	h, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, err
	}
	m, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, err
	}
	return h*60 + m, nil
}

func isTurnoCruzado(horaEntrada, horaSalida string) bool {
	if strings.TrimSpace(horaEntrada) == "" || strings.TrimSpace(horaSalida) == "" {
		return false
	}
	entradaMin, err := timeToMinutes(horaEntrada)
	if err != nil {
		return false
	}
	salidaMin, err := timeToMinutes(horaSalida)
	if err != nil {
		return false
	}
	return salidaMin < entradaMin
}

func tardanzaMinutosConfigurada(cfg *EmpresaAsistenciaConfiguracion, turno, horaEntrada string, manual int) int {
	if manual < 0 {
		manual = 0
	}
	expected := resolveTurnoInicio(cfg, turno)
	if strings.TrimSpace(expected) == "" || strings.TrimSpace(horaEntrada) == "" {
		return manual
	}

	entradaMin, err := timeToMinutes(horaEntrada)
	if err != nil {
		return manual
	}
	expectedMin, err := timeToMinutes(expected)
	if err != nil {
		return manual
	}

	diff := entradaMin - expectedMin
	if diff < -720 {
		diff += 24 * 60
	}
	if diff > 720 {
		diff -= 24 * 60
	}

	calc := 0
	if diff > cfg.ToleranciaEntradaMin {
		calc = diff - cfg.ToleranciaEntradaMin
	}
	if manual > calc {
		return manual
	}
	return calc
}

func validateAsistenciaTurnoConfig(cfg *EmpresaAsistenciaConfiguracion, turno, horaEntrada, horaSalida string) error {
	normalizedTurno := normalizeTurnoAsistencia(turno)
	if normalizedTurno == "noche" && !cfg.PermitirTurnoNocturno {
		return fmt.Errorf("el turno nocturno esta deshabilitado en configuracion de asistencia")
	}
	if isTurnoCruzado(horaEntrada, horaSalida) && !cfg.PermitirTurnoCruzado {
		return fmt.Errorf("el turno cruzado esta deshabilitado en configuracion de asistencia")
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
	if err := validateAsistenciaFechaNoCerrada(dbConn, item.EmpresaID, fechaAsistencia); err != nil {
		return 0, err
	}

	cfg, err := GetEmpresaAsistenciaConfiguracion(dbConn, item.EmpresaID)
	if err != nil {
		return 0, err
	}
	turno := normalizeTurnoAsistencia(item.Turno)
	if err := validateAsistenciaTurnoConfig(cfg, turno, horaEntrada, horaSalida); err != nil {
		return 0, err
	}

	if item.MinutosTarde < 0 {
		item.MinutosTarde = 0
	}
	item.MinutosTarde = tardanzaMinutosConfigurada(cfg, turno, horaEntrada, item.MinutosTarde)
	if item.HorasTrabajadas < 0 {
		item.HorasTrabajadas = 0
	}
	if item.HorasTrabajadas <= 0 && horaEntrada != "" && horaSalida != "" {
		if hours, calcErr := calculateWorkedHours(fechaAsistencia, horaEntrada, horaSalida); calcErr == nil {
			item.HorasTrabajadas = hours
		}
	}
	estadoAsistencia := normalizeEstadoAsistencia(item.EstadoAsistencia)
	if estadoAsistencia == "pendiente" && horaEntrada != "" {
		if item.MinutosTarde > 0 {
			estadoAsistencia = "tarde"
		} else {
			estadoAsistencia = "presente"
		}
	}

	nowExpr := sqlNowExpr()
	id, err := insertSQLCompat(dbConn, `INSERT INTO empresa_asistencia_empleados (
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
		`+nowExpr+`, `+nowExpr+`
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
	return id, nil
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

	rows, err := querySQLCompat(dbConn, query, args...)
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
	fechaActual, _, _, _, err := getAsistenciaRegistroContext(dbConn, item.EmpresaID, item.ID)
	if err != nil {
		return err
	}
	if err := validateAsistenciaFechaNoCerrada(dbConn, item.EmpresaID, fechaActual); err != nil {
		return err
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
	if err := validateAsistenciaFechaNoCerrada(dbConn, item.EmpresaID, fechaAsistencia); err != nil {
		return err
	}

	cfg, err := GetEmpresaAsistenciaConfiguracion(dbConn, item.EmpresaID)
	if err != nil {
		return err
	}
	turno := normalizeTurnoAsistencia(item.Turno)
	if err := validateAsistenciaTurnoConfig(cfg, turno, horaEntrada, horaSalida); err != nil {
		return err
	}

	if item.MinutosTarde < 0 {
		item.MinutosTarde = 0
	}
	item.MinutosTarde = tardanzaMinutosConfigurada(cfg, turno, horaEntrada, item.MinutosTarde)
	if item.HorasTrabajadas < 0 {
		item.HorasTrabajadas = 0
	}
	if item.HorasTrabajadas <= 0 && horaEntrada != "" && horaSalida != "" {
		if hours, calcErr := calculateWorkedHours(fechaAsistencia, horaEntrada, horaSalida); calcErr == nil {
			item.HorasTrabajadas = hours
		}
	}

	estadoAsistencia := normalizeEstadoAsistencia(item.EstadoAsistencia)
	if estadoAsistencia == "pendiente" && horaEntrada != "" {
		if item.MinutosTarde > 0 {
			estadoAsistencia = "tarde"
		} else {
			estadoAsistencia = "presente"
		}
	}

	nowExpr := sqlNowExpr()
	res, err := execSQLCompat(dbConn, `UPDATE empresa_asistencia_empleados
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
		fecha_actualizacion = `+nowExpr+`
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
		estadoAsistencia,
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

	fechaAsistencia, _, _, _, err := getAsistenciaRegistroContext(dbConn, empresaID, id)
	if err != nil {
		return err
	}
	if err := validateAsistenciaFechaNoCerrada(dbConn, empresaID, fechaAsistencia); err != nil {
		return err
	}

	nowExpr := sqlNowExpr()
	res, err := execSQLCompat(dbConn, `UPDATE empresa_asistencia_empleados
	SET estado = ?, fecha_actualizacion = `+nowExpr+`
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

	fechaAsistencia, turno, _, horaSalidaActual, err := getAsistenciaRegistroContext(dbConn, empresaID, id)
	if err != nil {
		return err
	}
	if err := validateAsistenciaFechaNoCerrada(dbConn, empresaID, fechaAsistencia); err != nil {
		return err
	}

	cfg, err := GetEmpresaAsistenciaConfiguracion(dbConn, empresaID)
	if err != nil {
		return err
	}
	if err := validateAsistenciaTurnoConfig(cfg, turno, parsedTime, horaSalidaActual); err != nil {
		return err
	}

	if minutosTarde < 0 {
		minutosTarde = 0
	}
	minutosTarde = tardanzaMinutosConfigurada(cfg, turno, parsedTime, minutosTarde)
	estadoRaw := strings.TrimSpace(estadoAsistencia)
	estadoAsistencia = normalizeEstadoAsistencia(estadoRaw)
	if estadoRaw == "" || estadoAsistencia == "pendiente" {
		if minutosTarde > 0 {
			estadoAsistencia = "tarde"
		} else {
			estadoAsistencia = "presente"
		}
	}

	nowExpr := sqlNowExpr()
	res, err := execSQLCompat(dbConn, `UPDATE empresa_asistencia_empleados
	SET
		hora_entrada = ?,
		minutos_tarde = ?,
		estado_asistencia = ?,
		novedad = CASE WHEN ? <> '' THEN ? ELSE novedad END,
		fecha_actualizacion = `+nowExpr+`
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
	var turno string
	var horaEntrada string
	var estadoActual string
	err = queryRowSQLCompat(dbConn, `SELECT
		COALESCE(fecha_asistencia, ''),
		COALESCE(turno, 'general'),
		COALESCE(hora_entrada, ''),
		COALESCE(estado_asistencia, 'pendiente')
	FROM empresa_asistencia_empleados
	WHERE empresa_id = ? AND id = ?
	LIMIT 1`, empresaID, id).Scan(&fechaAsistencia, &turno, &horaEntrada, &estadoActual)
	if err != nil {
		return err
	}
	if err := validateAsistenciaFechaNoCerrada(dbConn, empresaID, fechaAsistencia); err != nil {
		return err
	}

	cfg, err := GetEmpresaAsistenciaConfiguracion(dbConn, empresaID)
	if err != nil {
		return err
	}
	if err := validateAsistenciaTurnoConfig(cfg, turno, horaEntrada, parsedTime); err != nil {
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

	nowExpr := sqlNowExpr()
	res, err := execSQLCompat(dbConn, `UPDATE empresa_asistencia_empleados
	SET
		hora_salida = ?,
		horas_trabajadas = ?,
		estado_asistencia = ?,
		novedad = CASE WHEN ? <> '' THEN ? ELSE novedad END,
		fecha_actualizacion = `+nowExpr+`
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
	fechaAsistencia, _, _, _, err := getAsistenciaRegistroContext(dbConn, empresaID, id)
	if err != nil {
		return err
	}
	if err := validateAsistenciaFechaNoCerrada(dbConn, empresaID, fechaAsistencia); err != nil {
		return err
	}

	res, err := execSQLCompat(dbConn, `DELETE FROM empresa_asistencia_empleados WHERE empresa_id = ? AND id = ?`, empresaID, id)
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

func asistenciaBoolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
