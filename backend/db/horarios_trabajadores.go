package db

import (
	"database/sql"
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	ErrHorarioTrabajadorConflict = errors.New("conflicto de horario")

	ensureHorariosTrabajadoresSchemaOnce sync.Once
	ensureHorariosTrabajadoresSchemaErr  error
)

type HorarioTrabajador struct {
	ID                   int64    `json:"id"`
	EmpresaID            int64    `json:"empresa_id"`
	UsuarioID            *int64   `json:"usuario_id,omitempty"`
	NombreEmpleado       string   `json:"nombre_empleado"`
	Cargo                string   `json:"cargo,omitempty"`
	Area                 string   `json:"area,omitempty"`
	Sede                 string   `json:"sede,omitempty"`
	Fecha                string   `json:"fecha"`
	HoraInicio           string   `json:"hora_inicio"`
	HoraFin              string   `json:"hora_fin"`
	DescansoMinutos      int      `json:"descanso_minutos"`
	TurnoNombre          string   `json:"turno_nombre,omitempty"`
	TipoTurno            string   `json:"tipo_turno,omitempty"`
	Canal                string   `json:"canal,omitempty"`
	Color                string   `json:"color,omitempty"`
	Estado               string   `json:"estado,omitempty"`
	Publicado            bool     `json:"publicado"`
	Conflicto            bool     `json:"conflicto"`
	RequiereCobertura    bool     `json:"requiere_cobertura"`
	HorasProgramadas     float64  `json:"horas_programadas"`
	Observaciones        string   `json:"observaciones,omitempty"`
	FechaCreacion        string   `json:"fecha_creacion,omitempty"`
	FechaActualizacion   string   `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador       string   `json:"usuario_creador,omitempty"`
	ConflictosDetectados []string `json:"conflictos_detectados,omitempty"`
}

type HorarioTrabajadorConfig struct {
	EmpresaID                    int64   `json:"empresa_id"`
	HorasObjetivoDia             float64 `json:"horas_objetivo_dia"`
	HorasObjetivoSemana          float64 `json:"horas_objetivo_semana"`
	DescansoMinimoMinutos        int     `json:"descanso_minimo_minutos"`
	PermitirSolapados            bool    `json:"permitir_solapados"`
	AnticipacionPublicacionHoras int     `json:"anticipacion_publicacion_horas"`
	ColorManana                  string  `json:"color_manana,omitempty"`
	ColorTarde                   string  `json:"color_tarde,omitempty"`
	ColorNoche                   string  `json:"color_noche,omitempty"`
	ColorLibre                   string  `json:"color_libre,omitempty"`
	FechaActualizacion           string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador               string  `json:"usuario_creador,omitempty"`
}

type HorarioTrabajadorResumen struct {
	Clave    string  `json:"clave"`
	Etiqueta string  `json:"etiqueta"`
	Cantidad int     `json:"cantidad"`
	Horas    float64 `json:"horas"`
}

type HorarioTrabajadorSemaforo struct {
	Clave   string `json:"clave"`
	Titulo  string `json:"titulo"`
	Estado  string `json:"estado"`
	Detalle string `json:"detalle"`
}

type HorarioTrabajadorDashboard struct {
	EmpresaID                int64                       `json:"empresa_id"`
	Desde                    string                      `json:"desde"`
	Hasta                    string                      `json:"hasta"`
	TotalTurnos              int                         `json:"total_turnos"`
	EmpleadosProgramados     int                         `json:"empleados_programados"`
	HorasProgramadas         float64                     `json:"horas_programadas"`
	HorasPublicadas          float64                     `json:"horas_publicadas"`
	PromedioHorasPorEmpleado float64                     `json:"promedio_horas_por_empleado"`
	TurnosPublicados         int                         `json:"turnos_publicados"`
	TurnosPendientes         int                         `json:"turnos_pendientes"`
	Conflictos               int                         `json:"conflictos"`
	CoberturasPendientes     int                         `json:"coberturas_pendientes"`
	Areas                    []HorarioTrabajadorResumen  `json:"areas"`
	Sedes                    []HorarioTrabajadorResumen  `json:"sedes"`
	Estados                  []HorarioTrabajadorResumen  `json:"estados"`
	Semaforos                []HorarioTrabajadorSemaforo `json:"semaforos"`
	Alertas                  []string                    `json:"alertas"`
	Oportunidades            []string                    `json:"oportunidades"`
}

type HorarioTrabajadorBulkInput struct {
	EmpresaID         int64  `json:"empresa_id"`
	UsuarioID         *int64 `json:"usuario_id,omitempty"`
	NombreEmpleado    string `json:"nombre_empleado"`
	Cargo             string `json:"cargo,omitempty"`
	Area              string `json:"area,omitempty"`
	Sede              string `json:"sede,omitempty"`
	FechaInicio       string `json:"fecha_inicio"`
	FechaFin          string `json:"fecha_fin"`
	HoraInicio        string `json:"hora_inicio"`
	HoraFin           string `json:"hora_fin"`
	DescansoMinutos   int    `json:"descanso_minutos"`
	TurnoNombre       string `json:"turno_nombre,omitempty"`
	TipoTurno         string `json:"tipo_turno,omitempty"`
	Canal             string `json:"canal,omitempty"`
	Color             string `json:"color,omitempty"`
	Estado            string `json:"estado,omitempty"`
	Publicado         bool   `json:"publicado"`
	RequiereCobertura bool   `json:"requiere_cobertura"`
	Observaciones     string `json:"observaciones,omitempty"`
	DiasSemana        []int  `json:"dias_semana,omitempty"`
	UsuarioCreador    string `json:"usuario_creador,omitempty"`
}

type HorarioTrabajadorPublishInput struct {
	EmpresaID int64   `json:"empresa_id"`
	IDs       []int64 `json:"ids,omitempty"`
	Desde     string  `json:"desde,omitempty"`
	Hasta     string  `json:"hasta,omitempty"`
}

func ensureHorariosTrabajadoresSchemaReady() error {
	ensureHorariosTrabajadoresSchemaOnce.Do(func() {
		ensureHorariosTrabajadoresSchemaErr = EnsureHorariosTrabajadoresSchema()
	})
	return ensureHorariosTrabajadoresSchemaErr
}

func EnsureHorariosTrabajadoresSchema() error {
	dbConn := GetDB()
	if dbConn == nil {
		return fmt.Errorf("base de datos de empresas no inicializada")
	}

	statements := []string{
		`CREATE TABLE IF NOT EXISTS empresa_horarios_trabajadores (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL,
			usuario_id BIGINT,
			nombre_empleado TEXT NOT NULL,
			cargo TEXT,
			area TEXT,
			sede TEXT,
			fecha TEXT NOT NULL,
			hora_inicio TEXT NOT NULL,
			hora_fin TEXT NOT NULL,
			descanso_minutos INTEGER DEFAULT 0,
			turno_nombre TEXT,
			tipo_turno TEXT DEFAULT 'operativo',
			canal TEXT DEFAULT 'presencial',
			color TEXT,
			estado TEXT DEFAULT 'programado',
			publicado INTEGER DEFAULT 0,
			conflicto INTEGER DEFAULT 0,
			requiere_cobertura INTEGER DEFAULT 0,
			horas_programadas NUMERIC(10,2) DEFAULT 0,
			observaciones TEXT,
			fecha_creacion TEXT DEFAULT CURRENT_TIMESTAMP,
			fecha_actualizacion TEXT,
			usuario_creador TEXT
		);`,
		`ALTER TABLE empresa_horarios_trabajadores ADD COLUMN IF NOT EXISTS cargo TEXT`,
		`ALTER TABLE empresa_horarios_trabajadores ADD COLUMN IF NOT EXISTS area TEXT`,
		`ALTER TABLE empresa_horarios_trabajadores ADD COLUMN IF NOT EXISTS sede TEXT`,
		`ALTER TABLE empresa_horarios_trabajadores ADD COLUMN IF NOT EXISTS descanso_minutos INTEGER DEFAULT 0`,
		`ALTER TABLE empresa_horarios_trabajadores ADD COLUMN IF NOT EXISTS turno_nombre TEXT`,
		`ALTER TABLE empresa_horarios_trabajadores ADD COLUMN IF NOT EXISTS tipo_turno TEXT DEFAULT 'operativo'`,
		`ALTER TABLE empresa_horarios_trabajadores ADD COLUMN IF NOT EXISTS canal TEXT DEFAULT 'presencial'`,
		`ALTER TABLE empresa_horarios_trabajadores ADD COLUMN IF NOT EXISTS color TEXT`,
		`ALTER TABLE empresa_horarios_trabajadores ADD COLUMN IF NOT EXISTS publicado INTEGER DEFAULT 0`,
		`ALTER TABLE empresa_horarios_trabajadores ADD COLUMN IF NOT EXISTS conflicto INTEGER DEFAULT 0`,
		`ALTER TABLE empresa_horarios_trabajadores ADD COLUMN IF NOT EXISTS requiere_cobertura INTEGER DEFAULT 0`,
		`ALTER TABLE empresa_horarios_trabajadores ADD COLUMN IF NOT EXISTS horas_programadas NUMERIC(10,2) DEFAULT 0`,
		`ALTER TABLE empresa_horarios_trabajadores ADD COLUMN IF NOT EXISTS fecha_actualizacion TEXT`,
		`CREATE TABLE IF NOT EXISTS empresa_horarios_trabajadores_config (
			empresa_id BIGINT PRIMARY KEY,
			horas_objetivo_dia NUMERIC(10,2) DEFAULT 8,
			horas_objetivo_semana NUMERIC(10,2) DEFAULT 48,
			descanso_minimo_minutos INTEGER DEFAULT 30,
			permitir_solapados INTEGER DEFAULT 0,
			anticipacion_publicacion_horas INTEGER DEFAULT 24,
			color_manana TEXT DEFAULT '#2563eb',
			color_tarde TEXT DEFAULT '#f97316',
			color_noche TEXT DEFAULT '#7c3aed',
			color_libre TEXT DEFAULT '#64748b',
			fecha_actualizacion TEXT DEFAULT CURRENT_TIMESTAMP,
			usuario_creador TEXT
		);`,
		`CREATE INDEX IF NOT EXISTS idx_horarios_trabajadores_empresa_fecha ON empresa_horarios_trabajadores(empresa_id, fecha)`,
		`CREATE INDEX IF NOT EXISTS idx_horarios_trabajadores_empresa_usuario_fecha ON empresa_horarios_trabajadores(empresa_id, usuario_id, fecha)`,
		`CREATE INDEX IF NOT EXISTS idx_horarios_trabajadores_empresa_estado_fecha ON empresa_horarios_trabajadores(empresa_id, estado, fecha)`,
		`CREATE INDEX IF NOT EXISTS idx_horarios_trabajadores_empresa_area_fecha ON empresa_horarios_trabajadores(empresa_id, area, fecha)`,
	}
	for _, stmt := range statements {
		if _, err := ExecCompat(dbConn, stmt); err != nil {
			return fmt.Errorf("no se pudo preparar el esquema de horarios: %w", err)
		}
	}
	return nil
}

func horarioTrabajadorDefaultConfig(empresaID int64) HorarioTrabajadorConfig {
	return HorarioTrabajadorConfig{
		EmpresaID:                    empresaID,
		HorasObjetivoDia:             8,
		HorasObjetivoSemana:          48,
		DescansoMinimoMinutos:        30,
		PermitirSolapados:            false,
		AnticipacionPublicacionHoras: 24,
		ColorManana:                  "#2563eb",
		ColorTarde:                   "#f97316",
		ColorNoche:                   "#7c3aed",
		ColorLibre:                   "#64748b",
	}
}

func GetHorarioTrabajadorConfig(dbConn *sql.DB, empresaID int64) (HorarioTrabajadorConfig, error) {
	if err := ensureHorariosTrabajadoresSchemaReady(); err != nil {
		return HorarioTrabajadorConfig{}, err
	}
	cfg := horarioTrabajadorDefaultConfig(empresaID)
	row := QueryRowCompat(dbConn, `SELECT empresa_id, horas_objetivo_dia, horas_objetivo_semana, descanso_minimo_minutos, permitir_solapados, anticipacion_publicacion_horas, color_manana, color_tarde, color_noche, color_libre, COALESCE(fecha_actualizacion, ''), COALESCE(usuario_creador, '') FROM empresa_horarios_trabajadores_config WHERE empresa_id = ?`, empresaID)
	var permitir int
	if err := row.Scan(&cfg.EmpresaID, &cfg.HorasObjetivoDia, &cfg.HorasObjetivoSemana, &cfg.DescansoMinimoMinutos, &permitir, &cfg.AnticipacionPublicacionHoras, &cfg.ColorManana, &cfg.ColorTarde, &cfg.ColorNoche, &cfg.ColorLibre, &cfg.FechaActualizacion, &cfg.UsuarioCreador); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return cfg, nil
		}
		return HorarioTrabajadorConfig{}, err
	}
	cfg.PermitirSolapados = permitir > 0
	return cfg, nil
}

func UpsertHorarioTrabajadorConfig(dbConn *sql.DB, cfg HorarioTrabajadorConfig) error {
	if err := ensureHorariosTrabajadoresSchemaReady(); err != nil {
		return err
	}
	cfg = normalizeHorarioTrabajadorConfig(cfg)
	now := time.Now().Format("2006-01-02 15:04:05")
	_, err := ExecCompat(dbConn, `INSERT INTO empresa_horarios_trabajadores_config (empresa_id, horas_objetivo_dia, horas_objetivo_semana, descanso_minimo_minutos, permitir_solapados, anticipacion_publicacion_horas, color_manana, color_tarde, color_noche, color_libre, fecha_actualizacion, usuario_creador)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (empresa_id) DO UPDATE SET
			horas_objetivo_dia = EXCLUDED.horas_objetivo_dia,
			horas_objetivo_semana = EXCLUDED.horas_objetivo_semana,
			descanso_minimo_minutos = EXCLUDED.descanso_minimo_minutos,
			permitir_solapados = EXCLUDED.permitir_solapados,
			anticipacion_publicacion_horas = EXCLUDED.anticipacion_publicacion_horas,
			color_manana = EXCLUDED.color_manana,
			color_tarde = EXCLUDED.color_tarde,
			color_noche = EXCLUDED.color_noche,
			color_libre = EXCLUDED.color_libre,
			fecha_actualizacion = EXCLUDED.fecha_actualizacion,
			usuario_creador = EXCLUDED.usuario_creador`,
		cfg.EmpresaID,
		cfg.HorasObjetivoDia,
		cfg.HorasObjetivoSemana,
		cfg.DescansoMinimoMinutos,
		horarioBoolToInt(cfg.PermitirSolapados),
		cfg.AnticipacionPublicacionHoras,
		cfg.ColorManana,
		cfg.ColorTarde,
		cfg.ColorNoche,
		cfg.ColorLibre,
		now,
		cfg.UsuarioCreador,
	)
	return err
}

func ListHorariosTrabajadores(dbConn *sql.DB, empresaID int64, desde, hasta, q, area, sede, estado string, publishedOnly bool, limit int) ([]HorarioTrabajador, error) {
	if err := ensureHorariosTrabajadoresSchemaReady(); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 1000 {
		limit = 400
	}
	query := `SELECT id, empresa_id, usuario_id, nombre_empleado, COALESCE(cargo, ''), COALESCE(area, ''), COALESCE(sede, ''), fecha, hora_inicio, hora_fin, COALESCE(descanso_minutos, 0), COALESCE(turno_nombre, ''), COALESCE(tipo_turno, ''), COALESCE(canal, ''), COALESCE(color, ''), COALESCE(estado, ''), COALESCE(publicado, 0), COALESCE(conflicto, 0), COALESCE(requiere_cobertura, 0), COALESCE(horas_programadas, 0), COALESCE(observaciones, ''), COALESCE(fecha_creacion, ''), COALESCE(fecha_actualizacion, ''), COALESCE(usuario_creador, '') FROM empresa_horarios_trabajadores WHERE empresa_id = ?`
	args := []interface{}{empresaID}
	if strings.TrimSpace(desde) != "" {
		query += ` AND fecha >= ?`
		args = append(args, strings.TrimSpace(desde))
	}
	if strings.TrimSpace(hasta) != "" {
		query += ` AND fecha <= ?`
		args = append(args, strings.TrimSpace(hasta))
	}
	if strings.TrimSpace(area) != "" {
		query += ` AND LOWER(COALESCE(area, '')) = LOWER(?)`
		args = append(args, strings.TrimSpace(area))
	}
	if strings.TrimSpace(sede) != "" {
		query += ` AND LOWER(COALESCE(sede, '')) = LOWER(?)`
		args = append(args, strings.TrimSpace(sede))
	}
	if strings.TrimSpace(estado) != "" {
		query += ` AND LOWER(COALESCE(estado, '')) = LOWER(?)`
		args = append(args, strings.TrimSpace(estado))
	}
	if publishedOnly {
		query += ` AND COALESCE(publicado, 0) = 1`
	}
	if strings.TrimSpace(q) != "" {
		like := "%" + strings.ToLower(strings.TrimSpace(q)) + "%"
		query += ` AND (LOWER(COALESCE(nombre_empleado, '')) LIKE ? OR LOWER(COALESCE(cargo, '')) LIKE ? OR LOWER(COALESCE(area, '')) LIKE ? OR LOWER(COALESCE(turno_nombre, '')) LIKE ?)`
		args = append(args, like, like, like, like)
	}
	query += ` ORDER BY fecha ASC, hora_inicio ASC, nombre_empleado ASC LIMIT ?`
	args = append(args, limit)
	rows, err := ExecQueryCompat(dbConn, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []HorarioTrabajador
	for rows.Next() {
		item, err := scanHorarioTrabajador(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func GetHorariosTrabajadorByUsuario(dbConn *sql.DB, empresaID int64, usuarioID int64, desde, hasta string) ([]HorarioTrabajador, error) {
	if err := ensureHorariosTrabajadoresSchemaReady(); err != nil {
		return nil, err
	}
	query := `SELECT id, empresa_id, usuario_id, nombre_empleado, COALESCE(cargo, ''), COALESCE(area, ''), COALESCE(sede, ''), fecha, hora_inicio, hora_fin, COALESCE(descanso_minutos, 0), COALESCE(turno_nombre, ''), COALESCE(tipo_turno, ''), COALESCE(canal, ''), COALESCE(color, ''), COALESCE(estado, ''), COALESCE(publicado, 0), COALESCE(conflicto, 0), COALESCE(requiere_cobertura, 0), COALESCE(horas_programadas, 0), COALESCE(observaciones, ''), COALESCE(fecha_creacion, ''), COALESCE(fecha_actualizacion, ''), COALESCE(usuario_creador, '') FROM empresa_horarios_trabajadores WHERE empresa_id = ? AND usuario_id = ?`
	args := []interface{}{empresaID, usuarioID}
	if strings.TrimSpace(desde) != "" {
		query += ` AND fecha >= ?`
		args = append(args, strings.TrimSpace(desde))
	}
	if strings.TrimSpace(hasta) != "" {
		query += ` AND fecha <= ?`
		args = append(args, strings.TrimSpace(hasta))
	}
	query += ` ORDER BY fecha ASC, hora_inicio ASC LIMIT 200`
	rows, err := ExecQueryCompat(dbConn, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []HorarioTrabajador
	for rows.Next() {
		item, err := scanHorarioTrabajador(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func GetHorariosTrabajadorByUsuarioPerfil(dbConn *sql.DB, empresaID int64, usuarioID int64, email, desde, hasta string, publishedOnly bool, limit int) ([]HorarioTrabajador, error) {
	if err := ensureHorariosTrabajadoresSchemaReady(); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	email = strings.TrimSpace(email)
	query := `SELECT id, empresa_id, usuario_id, nombre_empleado, COALESCE(cargo, ''), COALESCE(area, ''), COALESCE(sede, ''), fecha, hora_inicio, hora_fin, COALESCE(descanso_minutos, 0), COALESCE(turno_nombre, ''), COALESCE(tipo_turno, ''), COALESCE(canal, ''), COALESCE(color, ''), COALESCE(estado, ''), COALESCE(publicado, 0), COALESCE(conflicto, 0), COALESCE(requiere_cobertura, 0), COALESCE(horas_programadas, 0), COALESCE(observaciones, ''), COALESCE(fecha_creacion, ''), COALESCE(fecha_actualizacion, ''), COALESCE(usuario_creador, '') FROM empresa_horarios_trabajadores WHERE empresa_id = ? AND (usuario_id = ?`
	args := []interface{}{empresaID, usuarioID}
	if email != "" {
		query += ` OR LOWER(COALESCE(nombre_empleado, '')) = LOWER(?)`
		args = append(args, email)
	}
	query += `)`
	if strings.TrimSpace(desde) != "" {
		query += ` AND fecha >= ?`
		args = append(args, strings.TrimSpace(desde))
	}
	if strings.TrimSpace(hasta) != "" {
		query += ` AND fecha <= ?`
		args = append(args, strings.TrimSpace(hasta))
	}
	if publishedOnly {
		query += ` AND COALESCE(publicado, 0) = 1`
	}
	query += ` ORDER BY fecha ASC, hora_inicio ASC LIMIT ?`
	args = append(args, limit)
	rows, err := ExecQueryCompat(dbConn, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []HorarioTrabajador
	for rows.Next() {
		item, err := scanHorarioTrabajador(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func BuildHorarioTrabajadorDashboard(dbConn *sql.DB, empresaID int64, desde, hasta string) (HorarioTrabajadorDashboard, error) {
	if err := ensureHorariosTrabajadoresSchemaReady(); err != nil {
		return HorarioTrabajadorDashboard{}, err
	}
	cfg, err := GetHorarioTrabajadorConfig(dbConn, empresaID)
	if err != nil {
		return HorarioTrabajadorDashboard{}, err
	}
	items, err := ListHorariosTrabajadores(dbConn, empresaID, desde, hasta, "", "", "", "", false, 1000)
	if err != nil {
		return HorarioTrabajadorDashboard{}, err
	}
	dashboard := HorarioTrabajadorDashboard{EmpresaID: empresaID, Desde: desde, Hasta: hasta}
	areaMap := map[string]*HorarioTrabajadorResumen{}
	sedeMap := map[string]*HorarioTrabajadorResumen{}
	estadoMap := map[string]*HorarioTrabajadorResumen{}
	employees := map[string]struct{}{}
	var horasNoPublicadas float64
	var horasRiesgo float64
	for _, item := range items {
		dashboard.TotalTurnos++
		if item.NombreEmpleado != "" {
			employees[strings.ToLower(item.NombreEmpleado)] = struct{}{}
		}
		dashboard.HorasProgramadas += item.HorasProgramadas
		if item.Publicado {
			dashboard.TurnosPublicados++
			dashboard.HorasPublicadas += item.HorasProgramadas
		} else {
			dashboard.TurnosPendientes++
			horasNoPublicadas += item.HorasProgramadas
		}
		if item.Conflicto {
			dashboard.Conflictos++
		}
		if item.RequiereCobertura {
			dashboard.CoberturasPendientes++
		}
		if item.HorasProgramadas > cfg.HorasObjetivoDia {
			horasRiesgo += item.HorasProgramadas - cfg.HorasObjetivoDia
		}
		accumulateHorarioResumen(areaMap, normalizeHorarioKey(item.Area, "Sin área"), item.HorasProgramadas)
		accumulateHorarioResumen(sedeMap, normalizeHorarioKey(item.Sede, "Sin sede"), item.HorasProgramadas)
		accumulateHorarioResumen(estadoMap, normalizeHorarioKey(item.Estado, "programado"), item.HorasProgramadas)
	}
	dashboard.EmpleadosProgramados = len(employees)
	if dashboard.EmpleadosProgramados > 0 {
		dashboard.PromedioHorasPorEmpleado = roundHorarioFloat(dashboard.HorasProgramadas / float64(dashboard.EmpleadosProgramados))
	}
	dashboard.Areas = flattenHorarioResumen(areaMap)
	dashboard.Sedes = flattenHorarioResumen(sedeMap)
	dashboard.Estados = flattenHorarioResumen(estadoMap)
	dashboard.Semaforos = []HorarioTrabajadorSemaforo{
		buildHorarioSemaforo("publicacion", "Publicación", ratioHorarioEstado(dashboard.TurnosPublicados, dashboard.TotalTurnos), fmt.Sprintf("%d de %d turnos ya fueron publicados al equipo.", dashboard.TurnosPublicados, dashboard.TotalTurnos)),
		buildHorarioSemaforo("conflictos", "Solapes y conflictos", inverseRatioHorarioEstado(dashboard.Conflictos, maxHorarioInt(dashboard.TotalTurnos, 1)), fmt.Sprintf("%d turnos presentan conflicto o superposición.", dashboard.Conflictos)),
		buildHorarioSemaforo("cobertura", "Cobertura", inverseRatioHorarioEstado(dashboard.CoberturasPendientes, maxHorarioInt(dashboard.TotalTurnos, 1)), fmt.Sprintf("%d turnos siguen marcados como cobertura pendiente.", dashboard.CoberturasPendientes)),
		buildHorarioSemaforo("carga", "Carga horaria", inverseRatioHorarioEstado(int(math.Round(horasRiesgo)), maxHorarioInt(int(math.Round(cfg.HorasObjetivoDia*float64(maxHorarioInt(dashboard.EmpleadosProgramados, 1)))), 1)), fmt.Sprintf("%.1f horas están por encima del objetivo diario configurado.", horasRiesgo)),
	}
	if dashboard.TurnosPendientes > 0 {
		dashboard.Alertas = append(dashboard.Alertas, fmt.Sprintf("Hay %d turnos pendientes por publicar al equipo.", dashboard.TurnosPendientes))
	}
	if dashboard.Conflictos > 0 {
		dashboard.Alertas = append(dashboard.Alertas, fmt.Sprintf("Se detectaron %d turnos con solapes o conflictos que conviene corregir antes de publicar.", dashboard.Conflictos))
	}
	if dashboard.CoberturasPendientes > 0 {
		dashboard.Alertas = append(dashboard.Alertas, fmt.Sprintf("Existen %d turnos que necesitan reemplazo o cobertura adicional.", dashboard.CoberturasPendientes))
	}
	if len(dashboard.Alertas) == 0 {
		dashboard.Alertas = append(dashboard.Alertas, "La programación actual luce consistente y lista para operar.")
	}
	if dashboard.HorasProgramadas > 0 {
		dashboard.Oportunidades = append(dashboard.Oportunidades, fmt.Sprintf("Publicar las %.1f horas aún pendientes le dará al equipo un cierre operativo más claro.", horasNoPublicadas))
	}
	if len(dashboard.Areas) > 0 {
		dashboard.Oportunidades = append(dashboard.Oportunidades, fmt.Sprintf("El frente con mayor carga es %s con %.1f horas programadas.", dashboard.Areas[0].Etiqueta, dashboard.Areas[0].Horas))
	}
	if len(dashboard.Oportunidades) == 0 {
		dashboard.Oportunidades = append(dashboard.Oportunidades, "Todavía no hay suficientes turnos para generar recomendaciones de programación.")
	}
	dashboard.HorasProgramadas = roundHorarioFloat(dashboard.HorasProgramadas)
	dashboard.HorasPublicadas = roundHorarioFloat(dashboard.HorasPublicadas)
	dashboard.PromedioHorasPorEmpleado = roundHorarioFloat(dashboard.PromedioHorasPorEmpleado)
	return dashboard, nil
}

func CreateHorarioTrabajador(dbConn *sql.DB, item *HorarioTrabajador) (int64, error) {
	if err := ensureHorariosTrabajadoresSchemaReady(); err != nil {
		return 0, err
	}
	if item == nil {
		return 0, fmt.Errorf("horario vacío")
	}
	cfg, err := GetHorarioTrabajadorConfig(dbConn, item.EmpresaID)
	if err != nil {
		return 0, err
	}
	if err := validateHorarioTrabajadorInput(*item); err != nil {
		return 0, err
	}
	normalizeHorarioTrabajador(item, cfg)
	if err := validateHorarioTrabajadorNormalized(*item, cfg); err != nil {
		return 0, err
	}
	if conflicts, err := findHorarioTrabajadorConflicts(dbConn, *item); err != nil {
		return 0, err
	} else if len(conflicts) > 0 {
		item.Conflicto = true
		item.ConflictosDetectados = conflicts
		if !cfg.PermitirSolapados {
			return 0, fmt.Errorf("%w: %s", ErrHorarioTrabajadorConflict, strings.Join(conflicts, " | "))
		}
	}
	now := time.Now().Format("2006-01-02 15:04:05")
	item.FechaCreacion = now
	item.FechaActualizacion = now
	id, err := insertSQLCompat(dbConn, `INSERT INTO empresa_horarios_trabajadores (empresa_id, usuario_id, nombre_empleado, cargo, area, sede, fecha, hora_inicio, hora_fin, descanso_minutos, turno_nombre, tipo_turno, canal, color, estado, publicado, conflicto, requiere_cobertura, horas_programadas, observaciones, fecha_creacion, fecha_actualizacion, usuario_creador)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		item.EmpresaID,
		item.UsuarioID,
		item.NombreEmpleado,
		item.Cargo,
		item.Area,
		item.Sede,
		item.Fecha,
		item.HoraInicio,
		item.HoraFin,
		item.DescansoMinutos,
		item.TurnoNombre,
		item.TipoTurno,
		item.Canal,
		item.Color,
		item.Estado,
		horarioBoolToInt(item.Publicado),
		horarioBoolToInt(item.Conflicto),
		horarioBoolToInt(item.RequiereCobertura),
		item.HorasProgramadas,
		item.Observaciones,
		item.FechaCreacion,
		item.FechaActualizacion,
		item.UsuarioCreador,
	)
	return id, err
}

func UpdateHorarioTrabajador(dbConn *sql.DB, item *HorarioTrabajador) error {
	if err := ensureHorariosTrabajadoresSchemaReady(); err != nil {
		return err
	}
	if item == nil || item.ID <= 0 {
		return fmt.Errorf("horario inválido")
	}
	cfg, err := GetHorarioTrabajadorConfig(dbConn, item.EmpresaID)
	if err != nil {
		return err
	}
	if err := validateHorarioTrabajadorInput(*item); err != nil {
		return err
	}
	normalizeHorarioTrabajador(item, cfg)
	if err := validateHorarioTrabajadorNormalized(*item, cfg); err != nil {
		return err
	}
	if conflicts, err := findHorarioTrabajadorConflicts(dbConn, *item); err != nil {
		return err
	} else if len(conflicts) > 0 {
		item.Conflicto = true
		item.ConflictosDetectados = conflicts
		if !cfg.PermitirSolapados {
			return fmt.Errorf("%w: %s", ErrHorarioTrabajadorConflict, strings.Join(conflicts, " | "))
		}
	} else {
		item.Conflicto = false
	}
	item.FechaActualizacion = time.Now().Format("2006-01-02 15:04:05")
	res, err := ExecCompat(dbConn, `UPDATE empresa_horarios_trabajadores SET usuario_id = ?, nombre_empleado = ?, cargo = ?, area = ?, sede = ?, fecha = ?, hora_inicio = ?, hora_fin = ?, descanso_minutos = ?, turno_nombre = ?, tipo_turno = ?, canal = ?, color = ?, estado = ?, publicado = ?, conflicto = ?, requiere_cobertura = ?, horas_programadas = ?, observaciones = ?, fecha_actualizacion = ? WHERE id = ? AND empresa_id = ?`,
		item.UsuarioID,
		item.NombreEmpleado,
		item.Cargo,
		item.Area,
		item.Sede,
		item.Fecha,
		item.HoraInicio,
		item.HoraFin,
		item.DescansoMinutos,
		item.TurnoNombre,
		item.TipoTurno,
		item.Canal,
		item.Color,
		item.Estado,
		horarioBoolToInt(item.Publicado),
		horarioBoolToInt(item.Conflicto),
		horarioBoolToInt(item.RequiereCobertura),
		item.HorasProgramadas,
		item.Observaciones,
		item.FechaActualizacion,
		item.ID,
		item.EmpresaID,
	)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("no se encontró el turno a actualizar")
	}
	return nil
}

func DeleteHorarioTrabajador(dbConn *sql.DB, id, empresaID int64) error {
	if err := ensureHorariosTrabajadoresSchemaReady(); err != nil {
		return err
	}
	_, err := ExecCompat(dbConn, `DELETE FROM empresa_horarios_trabajadores WHERE id = ? AND empresa_id = ?`, id, empresaID)
	return err
}

func PublishHorariosTrabajadores(dbConn *sql.DB, payload HorarioTrabajadorPublishInput) (int64, error) {
	if err := ensureHorariosTrabajadoresSchemaReady(); err != nil {
		return 0, err
	}
	payload.EmpresaID = maxHorarioInt64(payload.EmpresaID, 0)
	if payload.EmpresaID <= 0 {
		return 0, fmt.Errorf("empresa_id es obligatorio")
	}
	args := []interface{}{time.Now().Format("2006-01-02 15:04:05"), payload.EmpresaID}
	query := `UPDATE empresa_horarios_trabajadores SET publicado = 1, fecha_actualizacion = ? WHERE empresa_id = ?`
	if len(payload.IDs) > 0 {
		placeholders := make([]string, 0, len(payload.IDs))
		for _, id := range payload.IDs {
			placeholders = append(placeholders, "?")
			args = append(args, id)
		}
		query += ` AND id IN (` + strings.Join(placeholders, ",") + `)`
	} else {
		if strings.TrimSpace(payload.Desde) != "" {
			query += ` AND fecha >= ?`
			args = append(args, strings.TrimSpace(payload.Desde))
		}
		if strings.TrimSpace(payload.Hasta) != "" {
			query += ` AND fecha <= ?`
			args = append(args, strings.TrimSpace(payload.Hasta))
		}
	}
	res, err := ExecCompat(dbConn, query, args...)
	if err != nil {
		return 0, err
	}
	rows, _ := res.RowsAffected()
	return rows, nil
}

func CreateHorariosTrabajadoresBulk(dbConn *sql.DB, payload HorarioTrabajadorBulkInput) (int, []string, error) {
	if err := ensureHorariosTrabajadoresSchemaReady(); err != nil {
		return 0, nil, err
	}
	cfg, err := GetHorarioTrabajadorConfig(dbConn, payload.EmpresaID)
	if err != nil {
		return 0, nil, err
	}
	if strings.TrimSpace(payload.NombreEmpleado) == "" {
		return 0, nil, fmt.Errorf("nombre_empleado es obligatorio")
	}
	if !isValidHorarioHour(payload.HoraInicio) || !isValidHorarioHour(payload.HoraFin) {
		return 0, nil, fmt.Errorf("hora_inicio y hora_fin deben tener formato HH:MM")
	}
	payload = normalizeHorarioTrabajadorBulkInput(payload, cfg)
	startDate, err := time.Parse("2006-01-02", payload.FechaInicio)
	if err != nil {
		return 0, nil, fmt.Errorf("fecha_inicio inválida")
	}
	endDate, err := time.Parse("2006-01-02", payload.FechaFin)
	if err != nil {
		return 0, nil, fmt.Errorf("fecha_fin inválida")
	}
	if endDate.Before(startDate) {
		return 0, nil, fmt.Errorf("fecha_fin no puede ser anterior a fecha_inicio")
	}
	weekdays := buildHorarioWeekdaySet(payload.DiasSemana)
	var warnings []string
	created := 0
	for day := startDate; !day.After(endDate); day = day.AddDate(0, 0, 1) {
		if len(weekdays) > 0 {
			if _, ok := weekdays[day.Weekday()]; !ok {
				continue
			}
		}
		item := HorarioTrabajador{
			EmpresaID:         payload.EmpresaID,
			UsuarioID:         payload.UsuarioID,
			NombreEmpleado:    payload.NombreEmpleado,
			Cargo:             payload.Cargo,
			Area:              payload.Area,
			Sede:              payload.Sede,
			Fecha:             day.Format("2006-01-02"),
			HoraInicio:        payload.HoraInicio,
			HoraFin:           payload.HoraFin,
			DescansoMinutos:   payload.DescansoMinutos,
			TurnoNombre:       payload.TurnoNombre,
			TipoTurno:         payload.TipoTurno,
			Canal:             payload.Canal,
			Color:             payload.Color,
			Estado:            payload.Estado,
			Publicado:         payload.Publicado,
			RequiereCobertura: payload.RequiereCobertura,
			Observaciones:     payload.Observaciones,
			UsuarioCreador:    payload.UsuarioCreador,
		}
		if _, err := CreateHorarioTrabajador(dbConn, &item); err != nil {
			if errors.Is(err, ErrHorarioTrabajadorConflict) {
				warnings = append(warnings, fmt.Sprintf("%s: %s", item.Fecha, err.Error()))
				continue
			}
			return created, warnings, err
		}
		created++
	}
	return created, warnings, nil
}

func normalizeHorarioTrabajadorConfig(cfg HorarioTrabajadorConfig) HorarioTrabajadorConfig {
	if cfg.HorasObjetivoDia <= 0 {
		cfg.HorasObjetivoDia = 8
	}
	if cfg.HorasObjetivoSemana <= 0 {
		cfg.HorasObjetivoSemana = 48
	}
	if cfg.DescansoMinimoMinutos < 0 {
		cfg.DescansoMinimoMinutos = 0
	}
	if cfg.AnticipacionPublicacionHoras < 0 {
		cfg.AnticipacionPublicacionHoras = 0
	}
	if strings.TrimSpace(cfg.ColorManana) == "" {
		cfg.ColorManana = "#2563eb"
	}
	if strings.TrimSpace(cfg.ColorTarde) == "" {
		cfg.ColorTarde = "#f97316"
	}
	if strings.TrimSpace(cfg.ColorNoche) == "" {
		cfg.ColorNoche = "#7c3aed"
	}
	if strings.TrimSpace(cfg.ColorLibre) == "" {
		cfg.ColorLibre = "#64748b"
	}
	return cfg
}

func normalizeHorarioTrabajador(item *HorarioTrabajador, cfg HorarioTrabajadorConfig) {
	item.NombreEmpleado = strings.TrimSpace(item.NombreEmpleado)
	item.Cargo = strings.TrimSpace(item.Cargo)
	item.Area = strings.TrimSpace(item.Area)
	item.Sede = strings.TrimSpace(item.Sede)
	item.Fecha = strings.TrimSpace(item.Fecha)
	item.HoraInicio = normalizeHorarioHour(item.HoraInicio)
	item.HoraFin = normalizeHorarioHour(item.HoraFin)
	item.DescansoMinutos = maxHorarioInt(item.DescansoMinutos, 0)
	item.TurnoNombre = strings.TrimSpace(item.TurnoNombre)
	item.TipoTurno = strings.TrimSpace(item.TipoTurno)
	if item.TipoTurno == "" {
		item.TipoTurno = "operativo"
	}
	item.TipoTurno = normalizeHorarioChoice(item.TipoTurno, []string{"operativo", "administrativo", "cobertura", "guardia", "capacitacion"}, "operativo")
	item.Canal = strings.TrimSpace(item.Canal)
	if item.Canal == "" {
		item.Canal = "presencial"
	}
	item.Canal = normalizeHorarioChoice(item.Canal, []string{"presencial", "mixto", "remoto"}, "presencial")
	item.Color = strings.TrimSpace(item.Color)
	if item.Color == "" {
		item.Color = resolveHorarioDefaultColor(item.TurnoNombre, cfg)
	}
	item.Estado = strings.TrimSpace(item.Estado)
	if item.Estado == "" {
		item.Estado = "programado"
	}
	item.Estado = normalizeHorarioChoice(item.Estado, []string{"programado", "publicado", "cubierto", "incidencia", "cancelado"}, "programado")
	if strings.EqualFold(item.Estado, "publicado") {
		item.Publicado = true
	}
	item.Observaciones = strings.TrimSpace(item.Observaciones)
	item.HorasProgramadas = roundHorarioFloat(calcHorarioHours(item.HoraInicio, item.HoraFin, item.DescansoMinutos))
}

func normalizeHorarioTrabajadorBulkInput(payload HorarioTrabajadorBulkInput, cfg HorarioTrabajadorConfig) HorarioTrabajadorBulkInput {
	payload.NombreEmpleado = strings.TrimSpace(payload.NombreEmpleado)
	payload.Cargo = strings.TrimSpace(payload.Cargo)
	payload.Area = strings.TrimSpace(payload.Area)
	payload.Sede = strings.TrimSpace(payload.Sede)
	payload.FechaInicio = strings.TrimSpace(payload.FechaInicio)
	payload.FechaFin = strings.TrimSpace(payload.FechaFin)
	payload.HoraInicio = normalizeHorarioHour(payload.HoraInicio)
	payload.HoraFin = normalizeHorarioHour(payload.HoraFin)
	payload.DescansoMinutos = maxHorarioInt(payload.DescansoMinutos, 0)
	payload.TurnoNombre = strings.TrimSpace(payload.TurnoNombre)
	payload.TipoTurno = strings.TrimSpace(payload.TipoTurno)
	if payload.TipoTurno == "" {
		payload.TipoTurno = "operativo"
	}
	payload.TipoTurno = normalizeHorarioChoice(payload.TipoTurno, []string{"operativo", "administrativo", "cobertura", "guardia", "capacitacion"}, "operativo")
	payload.Canal = strings.TrimSpace(payload.Canal)
	if payload.Canal == "" {
		payload.Canal = "presencial"
	}
	payload.Canal = normalizeHorarioChoice(payload.Canal, []string{"presencial", "mixto", "remoto"}, "presencial")
	payload.Color = strings.TrimSpace(payload.Color)
	if payload.Color == "" {
		payload.Color = resolveHorarioDefaultColor(payload.TurnoNombre, cfg)
	}
	payload.Estado = strings.TrimSpace(payload.Estado)
	if payload.Estado == "" {
		payload.Estado = "programado"
	}
	payload.Estado = normalizeHorarioChoice(payload.Estado, []string{"programado", "publicado", "cubierto", "incidencia", "cancelado"}, "programado")
	if strings.EqualFold(payload.Estado, "publicado") {
		payload.Publicado = true
	}
	payload.Observaciones = strings.TrimSpace(payload.Observaciones)
	return payload
}

func validateHorarioTrabajadorInput(item HorarioTrabajador) error {
	if item.EmpresaID <= 0 {
		return fmt.Errorf("empresa_id es obligatorio")
	}
	if strings.TrimSpace(item.NombreEmpleado) == "" {
		return fmt.Errorf("nombre_empleado es obligatorio")
	}
	if _, err := time.Parse("2006-01-02", strings.TrimSpace(item.Fecha)); err != nil {
		return fmt.Errorf("fecha invalida")
	}
	if !isValidHorarioHour(item.HoraInicio) || !isValidHorarioHour(item.HoraFin) {
		return fmt.Errorf("hora_inicio y hora_fin deben tener formato HH:MM")
	}
	if item.DescansoMinutos < 0 || item.DescansoMinutos > 720 {
		return fmt.Errorf("descanso_minutos debe estar entre 0 y 720")
	}
	return nil
}

func validateHorarioTrabajadorNormalized(item HorarioTrabajador, cfg HorarioTrabajadorConfig) error {
	if item.HorasProgramadas <= 0 {
		return fmt.Errorf("el turno debe tener duracion positiva despues de descontar descansos")
	}
	if cfg.HorasObjetivoDia > 0 && item.HorasProgramadas > 24 {
		return fmt.Errorf("la duracion del turno no puede superar 24 horas")
	}
	return nil
}

func normalizeHorarioChoice(value string, allowed []string, fallback string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	for _, item := range allowed {
		if normalized == item {
			return item
		}
	}
	return fallback
}

func findHorarioTrabajadorConflicts(dbConn *sql.DB, item HorarioTrabajador) ([]string, error) {
	query := `SELECT id, nombre_empleado, fecha, hora_inicio, hora_fin, COALESCE(area, ''), COALESCE(sede, '') FROM empresa_horarios_trabajadores WHERE empresa_id = ? AND fecha = ?`
	args := []interface{}{item.EmpresaID, item.Fecha}
	if item.ID > 0 {
		query += ` AND id <> ?`
		args = append(args, item.ID)
	}
	if item.UsuarioID != nil && *item.UsuarioID > 0 {
		query += ` AND usuario_id = ?`
		args = append(args, *item.UsuarioID)
	} else {
		query += ` AND LOWER(COALESCE(nombre_empleado, '')) = LOWER(?)`
		args = append(args, item.NombreEmpleado)
	}
	rows, err := ExecQueryCompat(dbConn, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	startCandidate, endCandidate, err := horarioInterval(item.HoraInicio, item.HoraFin)
	if err != nil {
		return []string{"horario inválido"}, nil
	}
	var conflicts []string
	for rows.Next() {
		var id int64
		var nombre, fecha, horaInicio, horaFin, area, sede string
		if err := rows.Scan(&id, &nombre, &fecha, &horaInicio, &horaFin, &area, &sede); err != nil {
			return nil, err
		}
		startExisting, endExisting, err := horarioInterval(horaInicio, horaFin)
		if err != nil {
			continue
		}
		if startCandidate.Before(endExisting) && endCandidate.After(startExisting) {
			conflicts = append(conflicts, fmt.Sprintf("Se cruza con %s (%s-%s) en %s / %s", strings.TrimSpace(nombre), horaInicio, horaFin, normalizeHorarioKey(area, "sin área"), normalizeHorarioKey(sede, "sin sede")))
		}
	}
	return conflicts, rows.Err()
}

func scanHorarioTrabajador(scanner interface {
	Scan(dest ...interface{}) error
}) (HorarioTrabajador, error) {
	var item HorarioTrabajador
	var usuarioID sql.NullInt64
	var publicado, conflicto, cobertura int
	if err := scanner.Scan(&item.ID, &item.EmpresaID, &usuarioID, &item.NombreEmpleado, &item.Cargo, &item.Area, &item.Sede, &item.Fecha, &item.HoraInicio, &item.HoraFin, &item.DescansoMinutos, &item.TurnoNombre, &item.TipoTurno, &item.Canal, &item.Color, &item.Estado, &publicado, &conflicto, &cobertura, &item.HorasProgramadas, &item.Observaciones, &item.FechaCreacion, &item.FechaActualizacion, &item.UsuarioCreador); err != nil {
		return HorarioTrabajador{}, err
	}
	if usuarioID.Valid {
		item.UsuarioID = &usuarioID.Int64
	}
	item.Publicado = publicado > 0
	item.Conflicto = conflicto > 0
	item.RequiereCobertura = cobertura > 0
	item.HorasProgramadas = roundHorarioFloat(item.HorasProgramadas)
	return item, nil
}

func horarioInterval(horaInicio, horaFin string) (time.Time, time.Time, error) {
	start, err := time.Parse("15:04", normalizeHorarioHour(horaInicio))
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	end, err := time.Parse("15:04", normalizeHorarioHour(horaFin))
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	if !end.After(start) {
		end = end.Add(24 * time.Hour)
	}
	return start, end, nil
}

func calcHorarioHours(horaInicio, horaFin string, descansoMinutos int) float64 {
	start, end, err := horarioInterval(horaInicio, horaFin)
	if err != nil {
		return 0
	}
	duration := end.Sub(start).Hours() - (float64(maxHorarioInt(descansoMinutos, 0)) / 60)
	if duration < 0 {
		return 0
	}
	return duration
}

func normalizeHorarioHour(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "00:00"
	}
	if len(value) >= 5 {
		value = value[:5]
	}
	parts := strings.Split(value, ":")
	if len(parts) < 2 {
		return "00:00"
	}
	h, _ := strconv.Atoi(parts[0])
	m, _ := strconv.Atoi(parts[1])
	if h < 0 || h > 23 {
		h = 0
	}
	if m < 0 || m > 59 {
		m = 0
	}
	return fmt.Sprintf("%02d:%02d", h, m)
}

func isValidHorarioHour(raw string) bool {
	value := strings.TrimSpace(raw)
	if len(value) < 5 {
		return false
	}
	if len(value) > 5 {
		value = value[:5]
	}
	parts := strings.Split(value, ":")
	if len(parts) != 2 {
		return false
	}
	h, errH := strconv.Atoi(parts[0])
	m, errM := strconv.Atoi(parts[1])
	return errH == nil && errM == nil && h >= 0 && h <= 23 && m >= 0 && m <= 59
}

func resolveHorarioDefaultColor(turnoNombre string, cfg HorarioTrabajadorConfig) string {
	name := strings.ToLower(strings.TrimSpace(turnoNombre))
	switch {
	case strings.Contains(name, "man"):
		return cfg.ColorManana
	case strings.Contains(name, "tar"):
		return cfg.ColorTarde
	case strings.Contains(name, "noc"):
		return cfg.ColorNoche
	case strings.Contains(name, "lib"):
		return cfg.ColorLibre
	default:
		return cfg.ColorManana
	}
}

func accumulateHorarioResumen(target map[string]*HorarioTrabajadorResumen, key string, horas float64) {
	if target[key] == nil {
		target[key] = &HorarioTrabajadorResumen{Clave: strings.ToLower(key), Etiqueta: key}
	}
	target[key].Cantidad++
	target[key].Horas += horas
}

func flattenHorarioResumen(source map[string]*HorarioTrabajadorResumen) []HorarioTrabajadorResumen {
	items := make([]HorarioTrabajadorResumen, 0, len(source))
	for _, item := range source {
		item.Horas = roundHorarioFloat(item.Horas)
		items = append(items, *item)
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].Horas == items[j].Horas {
			return items[i].Etiqueta < items[j].Etiqueta
		}
		return items[i].Horas > items[j].Horas
	})
	return items
}

func normalizeHorarioKey(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}

func buildHorarioSemaforo(clave, titulo string, ratio float64, detalle string) HorarioTrabajadorSemaforo {
	estado := "solido"
	switch {
	case ratio < 0.45:
		estado = "critico"
	case ratio < 0.7:
		estado = "atencion"
	case ratio < 0.88:
		estado = "estable"
	}
	return HorarioTrabajadorSemaforo{Clave: clave, Titulo: titulo, Estado: estado, Detalle: detalle}
}

func ratioHorarioEstado(good, total int) float64 {
	if total <= 0 {
		return 1
	}
	return float64(good) / float64(total)
}

func inverseRatioHorarioEstado(problem, total int) float64 {
	if total <= 0 {
		return 1
	}
	value := 1 - (float64(problem) / float64(total))
	if value < 0 {
		return 0
	}
	return value
}

func buildHorarioWeekdaySet(days []int) map[time.Weekday]struct{} {
	result := map[time.Weekday]struct{}{}
	for _, day := range days {
		switch day {
		case 0:
			result[time.Sunday] = struct{}{}
		case 1:
			result[time.Monday] = struct{}{}
		case 2:
			result[time.Tuesday] = struct{}{}
		case 3:
			result[time.Wednesday] = struct{}{}
		case 4:
			result[time.Thursday] = struct{}{}
		case 5:
			result[time.Friday] = struct{}{}
		case 6:
			result[time.Saturday] = struct{}{}
		}
	}
	return result
}

func roundHorarioFloat(value float64) float64 {
	return math.Round(value*100) / 100
}

func maxHorarioInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func maxHorarioInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func horarioBoolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
