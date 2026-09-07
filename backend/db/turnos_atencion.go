package db

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type EmpresaTurnoAtencionConfig struct {
	EmpresaID                 int64  `json:"empresa_id"`
	NombreSistema             string `json:"nombre_sistema"`
	NombrePantalla            string `json:"nombre_pantalla"`
	PrefijoGeneral            string `json:"prefijo_general"`
	TiempoLlamadoSegundos     int    `json:"tiempo_llamado_segundos"`
	PermitirEmisionPublica    bool   `json:"permitir_emision_publica"`
	MostrarTicketsCompletados bool   `json:"mostrar_tickets_completados"`
	FechaActualizacion        string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador            string `json:"usuario_creador,omitempty"`
}

type EmpresaTurnoAtencionServicio struct {
	ID                 int64  `json:"id"`
	EmpresaID          int64  `json:"empresa_id"`
	Codigo             string `json:"codigo"`
	Nombre             string `json:"nombre"`
	Descripcion        string `json:"descripcion,omitempty"`
	Prefijo            string `json:"prefijo"`
	Prioridad          int    `json:"prioridad"`
	Color              string `json:"color,omitempty"`
	Estado             string `json:"estado,omitempty"`
	FechaCreacion      string `json:"fecha_creacion,omitempty"`
	FechaActualizacion string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string `json:"usuario_creador,omitempty"`
}

type EmpresaTurnoAtencionPuesto struct {
	ID                  int64  `json:"id"`
	EmpresaID           int64  `json:"empresa_id"`
	Codigo              string `json:"codigo"`
	Nombre              string `json:"nombre"`
	Area                string `json:"area,omitempty"`
	Ubicacion           string `json:"ubicacion,omitempty"`
	ServiciosPermitidos string `json:"servicios_permitidos,omitempty"`
	Estado              string `json:"estado,omitempty"`
	FechaCreacion       string `json:"fecha_creacion,omitempty"`
	FechaActualizacion  string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador      string `json:"usuario_creador,omitempty"`
}

type EmpresaTurnoAtencionTicket struct {
	ID                  int64   `json:"id"`
	EmpresaID           int64   `json:"empresa_id"`
	ServicioID          int64   `json:"servicio_id"`
	ServicioNombre      string  `json:"servicio_nombre,omitempty"`
	ServicioColor       string  `json:"servicio_color,omitempty"`
	PuestoID            int64   `json:"puesto_id,omitempty"`
	PuestoNombre        string  `json:"puesto_nombre,omitempty"`
	NumeroDia           int     `json:"numero_dia"`
	CodigoTurno         string  `json:"codigo_turno"`
	DocumentoCliente    string  `json:"documento_cliente,omitempty"`
	NombreCliente       string  `json:"nombre_cliente,omitempty"`
	CanalEmision        string  `json:"canal_emision,omitempty"`
	Estado              string  `json:"estado,omitempty"`
	Observaciones       string  `json:"observaciones,omitempty"`
	FechaOperacion      string  `json:"fecha_operacion,omitempty"`
	FechaEmision        string  `json:"fecha_emision,omitempty"`
	FechaLlamado        string  `json:"fecha_llamado,omitempty"`
	FechaInicioAtencion string  `json:"fecha_inicio_atencion,omitempty"`
	FechaCierre         string  `json:"fecha_cierre,omitempty"`
	TiempoEsperaMin     float64 `json:"tiempo_espera_min,omitempty"`
	TiempoAtencionMin   float64 `json:"tiempo_atencion_min,omitempty"`
	UsuarioCreador      string  `json:"usuario_creador,omitempty"`
}

type EmpresaTurnoAtencionDashboard struct {
	EmpresaID           int64                         `json:"empresa_id"`
	Fecha               string                        `json:"fecha"`
	Esperando           int                           `json:"esperando"`
	Llamando            int                           `json:"llamando"`
	EnAtencion          int                           `json:"en_atencion"`
	Completados         int                           `json:"completados"`
	Cancelados          int                           `json:"cancelados"`
	TiempoEsperaPromMin float64                       `json:"tiempo_espera_prom_min"`
	TiempoAtencionProm  float64                       `json:"tiempo_atencion_prom_min"`
	Servicios           []EmpresaTurnoAtencionResumen `json:"servicios"`
	Puestos             []EmpresaTurnoAtencionResumen `json:"puestos"`
	LlamadosRecientes   []EmpresaTurnoAtencionTicket  `json:"llamados_recientes"`
}

type EmpresaTurnoAtencionResumen struct {
	Clave      string `json:"clave"`
	Etiqueta   string `json:"etiqueta"`
	Cantidad   int    `json:"cantidad"`
	EnEspera   int    `json:"en_espera"`
	EnAtencion int    `json:"en_atencion"`
}

type EmpresaTurnoAtencionDisplay struct {
	EmpresaID         int64                         `json:"empresa_id"`
	Titulo            string                        `json:"titulo"`
	Fecha             string                        `json:"fecha"`
	TicketsLlamando   []EmpresaTurnoAtencionTicket  `json:"tickets_llamando"`
	LlamadosRecientes []EmpresaTurnoAtencionTicket  `json:"llamados_recientes"`
	ResumenServicios  []EmpresaTurnoAtencionResumen `json:"resumen_servicios"`
}

func EnsureEmpresaTurnosAtencionSchema(dbConn *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_turnos_atencion_config (
			empresa_id BIGINT PRIMARY KEY,
			nombre_sistema TEXT DEFAULT 'Turnos de atencion',
			nombre_pantalla TEXT DEFAULT 'Pantalla de llamados',
			prefijo_general TEXT DEFAULT 'T',
			tiempo_llamado_segundos INTEGER DEFAULT 20,
			permitir_emision_publica INTEGER DEFAULT 1,
			mostrar_tickets_completados INTEGER DEFAULT 1,
			fecha_actualizacion TEXT DEFAULT CURRENT_TIMESTAMP,
			usuario_creador TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS empresa_turnos_atencion_servicios (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL,
			codigo TEXT NOT NULL,
			nombre TEXT NOT NULL,
			descripcion TEXT,
			prefijo TEXT NOT NULL,
			prioridad INTEGER DEFAULT 100,
			color TEXT DEFAULT '#2563eb',
			estado TEXT DEFAULT 'activo',
			fecha_creacion TEXT DEFAULT CURRENT_TIMESTAMP,
			fecha_actualizacion TEXT DEFAULT CURRENT_TIMESTAMP,
			usuario_creador TEXT
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_turnos_servicio_codigo_empresa ON empresa_turnos_atencion_servicios(empresa_id, codigo)`,
		`CREATE TABLE IF NOT EXISTS empresa_turnos_atencion_puestos (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL,
			codigo TEXT NOT NULL,
			nombre TEXT NOT NULL,
			area TEXT,
			ubicacion TEXT,
			servicios_permitidos TEXT,
			estado TEXT DEFAULT 'activo',
			fecha_creacion TEXT DEFAULT CURRENT_TIMESTAMP,
			fecha_actualizacion TEXT DEFAULT CURRENT_TIMESTAMP,
			usuario_creador TEXT
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_turnos_puesto_codigo_empresa ON empresa_turnos_atencion_puestos(empresa_id, codigo)`,
		`CREATE TABLE IF NOT EXISTS empresa_turnos_atencion_tickets (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL,
			servicio_id BIGINT NOT NULL,
			puesto_id BIGINT,
			numero_dia INTEGER NOT NULL,
			codigo_turno TEXT NOT NULL,
			documento_cliente TEXT,
			nombre_cliente TEXT,
			canal_emision TEXT DEFAULT 'modulo',
			estado TEXT DEFAULT 'espera',
			observaciones TEXT,
			fecha_operacion TEXT DEFAULT CURRENT_DATE,
			fecha_emision TEXT DEFAULT CURRENT_TIMESTAMP,
			fecha_llamado TEXT,
			fecha_inicio_atencion TEXT,
			fecha_cierre TEXT,
			tiempo_espera_min NUMERIC(10,2) DEFAULT 0,
			tiempo_atencion_min NUMERIC(10,2) DEFAULT 0,
			usuario_creador TEXT
		)`,
		`CREATE INDEX IF NOT EXISTS ix_turnos_tickets_empresa_fecha ON empresa_turnos_atencion_tickets(empresa_id, fecha_operacion, id)`,
		`CREATE INDEX IF NOT EXISTS ix_turnos_tickets_empresa_estado_fecha ON empresa_turnos_atencion_tickets(empresa_id, estado, fecha_operacion, id)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_turnos_tickets_empresa_codigo_fecha ON empresa_turnos_atencion_tickets(empresa_id, fecha_operacion, codigo_turno)`,
	}
	for _, stmt := range stmts {
		if _, err := ExecCompat(dbConn, stmt); err != nil {
			return err
		}
	}
	return nil
}

func defaultTurnoAtencionConfig(empresaID int64) EmpresaTurnoAtencionConfig {
	return EmpresaTurnoAtencionConfig{
		EmpresaID:                 empresaID,
		NombreSistema:             "Turnos de atención",
		NombrePantalla:            "Pantalla de llamados",
		PrefijoGeneral:            "T",
		TiempoLlamadoSegundos:     20,
		PermitirEmisionPublica:    true,
		MostrarTicketsCompletados: true,
	}
}

func GetEmpresaTurnoAtencionConfig(dbConn *sql.DB, empresaID int64) (EmpresaTurnoAtencionConfig, error) {
	if err := EnsureEmpresaTurnosAtencionSchema(dbConn); err != nil {
		return EmpresaTurnoAtencionConfig{}, err
	}
	cfg := defaultTurnoAtencionConfig(empresaID)
	var emision, completados int
	err := QueryRowCompat(dbConn, `SELECT empresa_id, COALESCE(nombre_sistema,''), COALESCE(nombre_pantalla,''), COALESCE(prefijo_general,'T'), COALESCE(tiempo_llamado_segundos,20), COALESCE(permitir_emision_publica,1), COALESCE(mostrar_tickets_completados,1), COALESCE(fecha_actualizacion,''), COALESCE(usuario_creador,'') FROM empresa_turnos_atencion_config WHERE empresa_id = ?`, empresaID).Scan(
		&cfg.EmpresaID, &cfg.NombreSistema, &cfg.NombrePantalla, &cfg.PrefijoGeneral, &cfg.TiempoLlamadoSegundos, &emision, &completados, &cfg.FechaActualizacion, &cfg.UsuarioCreador,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return cfg, nil
		}
		return EmpresaTurnoAtencionConfig{}, err
	}
	cfg.PermitirEmisionPublica = emision > 0
	cfg.MostrarTicketsCompletados = completados > 0
	return cfg, nil
}

func UpsertEmpresaTurnoAtencionConfig(dbConn *sql.DB, cfg EmpresaTurnoAtencionConfig) error {
	if err := EnsureEmpresaTurnosAtencionSchema(dbConn); err != nil {
		return err
	}
	if strings.TrimSpace(cfg.NombreSistema) == "" {
		cfg.NombreSistema = "Turnos de atención"
	}
	if strings.TrimSpace(cfg.NombrePantalla) == "" {
		cfg.NombrePantalla = "Pantalla de llamados"
	}
	if strings.TrimSpace(cfg.PrefijoGeneral) == "" {
		cfg.PrefijoGeneral = "T"
	}
	if cfg.TiempoLlamadoSegundos <= 0 {
		cfg.TiempoLlamadoSegundos = 20
	}
	_, err := ExecCompat(dbConn, `INSERT INTO empresa_turnos_atencion_config (empresa_id, nombre_sistema, nombre_pantalla, prefijo_general, tiempo_llamado_segundos, permitir_emision_publica, mostrar_tickets_completados, fecha_actualizacion, usuario_creador)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (empresa_id) DO UPDATE SET
			nombre_sistema = EXCLUDED.nombre_sistema,
			nombre_pantalla = EXCLUDED.nombre_pantalla,
			prefijo_general = EXCLUDED.prefijo_general,
			tiempo_llamado_segundos = EXCLUDED.tiempo_llamado_segundos,
			permitir_emision_publica = EXCLUDED.permitir_emision_publica,
			mostrar_tickets_completados = EXCLUDED.mostrar_tickets_completados,
			fecha_actualizacion = EXCLUDED.fecha_actualizacion,
			usuario_creador = EXCLUDED.usuario_creador`,
		cfg.EmpresaID, cfg.NombreSistema, cfg.NombrePantalla, strings.ToUpper(strings.TrimSpace(cfg.PrefijoGeneral)), cfg.TiempoLlamadoSegundos, turnosBoolInt(cfg.PermitirEmisionPublica), turnosBoolInt(cfg.MostrarTicketsCompletados), time.Now().Format("2006-01-02 15:04:05"), cfg.UsuarioCreador,
	)
	return err
}

func ListEmpresaTurnosAtencionServicios(dbConn *sql.DB, empresaID int64) ([]EmpresaTurnoAtencionServicio, error) {
	if err := EnsureEmpresaTurnosAtencionSchema(dbConn); err != nil {
		return nil, err
	}
	rows, err := ExecQueryCompat(dbConn, `SELECT id, empresa_id, codigo, nombre, COALESCE(descripcion,''), COALESCE(prefijo,''), COALESCE(prioridad,100), COALESCE(color,'#2563eb'), COALESCE(estado,'activo'), COALESCE(fecha_creacion,''), COALESCE(fecha_actualizacion,''), COALESCE(usuario_creador,'') FROM empresa_turnos_atencion_servicios WHERE empresa_id = ? ORDER BY prioridad ASC, nombre ASC`, empresaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]EmpresaTurnoAtencionServicio, 0)
	for rows.Next() {
		var item EmpresaTurnoAtencionServicio
		if err := rows.Scan(&item.ID, &item.EmpresaID, &item.Codigo, &item.Nombre, &item.Descripcion, &item.Prefijo, &item.Prioridad, &item.Color, &item.Estado, &item.FechaCreacion, &item.FechaActualizacion, &item.UsuarioCreador); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func CreateEmpresaTurnoAtencionServicio(dbConn *sql.DB, item EmpresaTurnoAtencionServicio) (int64, error) {
	if err := EnsureEmpresaTurnosAtencionSchema(dbConn); err != nil {
		return 0, err
	}
	if strings.TrimSpace(item.Codigo) == "" || strings.TrimSpace(item.Nombre) == "" {
		return 0, fmt.Errorf("codigo y nombre son obligatorios")
	}
	if strings.TrimSpace(item.Prefijo) == "" {
		item.Prefijo = strings.ToUpper(strings.TrimSpace(item.Codigo))
	}
	if item.Prioridad <= 0 {
		item.Prioridad = 100
	}
	if strings.TrimSpace(item.Color) == "" {
		item.Color = "#2563eb"
	}
	return insertSQLCompat(dbConn, `INSERT INTO empresa_turnos_atencion_servicios (empresa_id, codigo, nombre, descripcion, prefijo, prioridad, color, estado, fecha_creacion, fecha_actualizacion, usuario_creador)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		item.EmpresaID, strings.ToUpper(strings.TrimSpace(item.Codigo)), strings.TrimSpace(item.Nombre), strings.TrimSpace(item.Descripcion), strings.ToUpper(strings.TrimSpace(item.Prefijo)), item.Prioridad, strings.TrimSpace(item.Color), defaultText(item.Estado, "activo"), time.Now().Format("2006-01-02 15:04:05"), time.Now().Format("2006-01-02 15:04:05"), strings.TrimSpace(item.UsuarioCreador),
	)
}

func ListEmpresaTurnosAtencionPuestos(dbConn *sql.DB, empresaID int64) ([]EmpresaTurnoAtencionPuesto, error) {
	if err := EnsureEmpresaTurnosAtencionSchema(dbConn); err != nil {
		return nil, err
	}
	rows, err := ExecQueryCompat(dbConn, `SELECT id, empresa_id, codigo, nombre, COALESCE(area,''), COALESCE(ubicacion,''), COALESCE(servicios_permitidos,''), COALESCE(estado,'activo'), COALESCE(fecha_creacion,''), COALESCE(fecha_actualizacion,''), COALESCE(usuario_creador,'') FROM empresa_turnos_atencion_puestos WHERE empresa_id = ? ORDER BY nombre ASC`, empresaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]EmpresaTurnoAtencionPuesto, 0)
	for rows.Next() {
		var item EmpresaTurnoAtencionPuesto
		if err := rows.Scan(&item.ID, &item.EmpresaID, &item.Codigo, &item.Nombre, &item.Area, &item.Ubicacion, &item.ServiciosPermitidos, &item.Estado, &item.FechaCreacion, &item.FechaActualizacion, &item.UsuarioCreador); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func CreateEmpresaTurnoAtencionPuesto(dbConn *sql.DB, item EmpresaTurnoAtencionPuesto) (int64, error) {
	if err := EnsureEmpresaTurnosAtencionSchema(dbConn); err != nil {
		return 0, err
	}
	if strings.TrimSpace(item.Codigo) == "" || strings.TrimSpace(item.Nombre) == "" {
		return 0, fmt.Errorf("codigo y nombre son obligatorios")
	}
	return insertSQLCompat(dbConn, `INSERT INTO empresa_turnos_atencion_puestos (empresa_id, codigo, nombre, area, ubicacion, servicios_permitidos, estado, fecha_creacion, fecha_actualizacion, usuario_creador)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		item.EmpresaID, strings.ToUpper(strings.TrimSpace(item.Codigo)), strings.TrimSpace(item.Nombre), strings.TrimSpace(item.Area), strings.TrimSpace(item.Ubicacion), strings.TrimSpace(item.ServiciosPermitidos), defaultText(item.Estado, "activo"), time.Now().Format("2006-01-02 15:04:05"), time.Now().Format("2006-01-02 15:04:05"), strings.TrimSpace(item.UsuarioCreador),
	)
}

func nextNumeroTurnoDia(dbConn *sql.DB, empresaID, servicioID int64, fecha string) (int, error) {
	var numero int
	if err := QueryRowCompat(dbConn, `SELECT COALESCE(MAX(numero_dia), 0) + 1 FROM empresa_turnos_atencion_tickets WHERE empresa_id = ? AND fecha_operacion = ?`, empresaID, fecha).Scan(&numero); err != nil {
		return 0, err
	}
	if numero <= 0 {
		numero = 1
	}
	return numero, nil
}

func getEmpresaTurnoServicio(dbConn *sql.DB, empresaID, servicioID int64) (EmpresaTurnoAtencionServicio, error) {
	var item EmpresaTurnoAtencionServicio
	err := QueryRowCompat(dbConn, `SELECT id, empresa_id, codigo, nombre, COALESCE(descripcion,''), COALESCE(prefijo,''), COALESCE(prioridad,100), COALESCE(color,'#2563eb'), COALESCE(estado,'activo'), COALESCE(fecha_creacion,''), COALESCE(fecha_actualizacion,''), COALESCE(usuario_creador,'') FROM empresa_turnos_atencion_servicios WHERE empresa_id = ? AND id = ?`, empresaID, servicioID).Scan(
		&item.ID, &item.EmpresaID, &item.Codigo, &item.Nombre, &item.Descripcion, &item.Prefijo, &item.Prioridad, &item.Color, &item.Estado, &item.FechaCreacion, &item.FechaActualizacion, &item.UsuarioCreador,
	)
	return item, err
}

func CreateEmpresaTurnoAtencionTicket(dbConn *sql.DB, item EmpresaTurnoAtencionTicket) (EmpresaTurnoAtencionTicket, error) {
	if err := EnsureEmpresaTurnosAtencionSchema(dbConn); err != nil {
		return EmpresaTurnoAtencionTicket{}, err
	}
	if item.EmpresaID <= 0 || item.ServicioID <= 0 {
		return EmpresaTurnoAtencionTicket{}, fmt.Errorf("empresa_id y servicio_id son obligatorios")
	}
	servicio, err := getEmpresaTurnoServicio(dbConn, item.EmpresaID, item.ServicioID)
	if err != nil {
		return EmpresaTurnoAtencionTicket{}, fmt.Errorf("servicio no encontrado")
	}
	item.FechaOperacion = time.Now().Format("2006-01-02")
	item.FechaEmision = time.Now().Format("2006-01-02 15:04:05")
	item.Estado = "espera"
	item.ServicioNombre = servicio.Nombre
	item.ServicioColor = servicio.Color
	item.NumeroDia, err = nextNumeroTurnoDia(dbConn, item.EmpresaID, item.ServicioID, item.FechaOperacion)
	if err != nil {
		return EmpresaTurnoAtencionTicket{}, err
	}
	item.CodigoTurno = fmt.Sprintf("%s-%03d", strings.ToUpper(strings.TrimSpace(servicio.Prefijo)), item.NumeroDia)
	id, err := insertSQLCompat(dbConn, `INSERT INTO empresa_turnos_atencion_tickets (empresa_id, servicio_id, numero_dia, codigo_turno, documento_cliente, nombre_cliente, canal_emision, estado, observaciones, fecha_operacion, fecha_emision, usuario_creador)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		item.EmpresaID, item.ServicioID, item.NumeroDia, item.CodigoTurno, strings.TrimSpace(item.DocumentoCliente), strings.TrimSpace(item.NombreCliente), defaultText(item.CanalEmision, "modulo"), item.Estado, strings.TrimSpace(item.Observaciones), item.FechaOperacion, item.FechaEmision, strings.TrimSpace(item.UsuarioCreador),
	)
	if err != nil {
		return EmpresaTurnoAtencionTicket{}, err
	}
	item.ID = id
	return item, nil
}

func ListEmpresaTurnosAtencionTickets(dbConn *sql.DB, empresaID int64, fecha, estado string, limit int) ([]EmpresaTurnoAtencionTicket, error) {
	if err := EnsureEmpresaTurnosAtencionSchema(dbConn); err != nil {
		return nil, err
	}
	if strings.TrimSpace(fecha) == "" {
		fecha = time.Now().Format("2006-01-02")
	}
	if limit <= 0 || limit > 300 {
		limit = 120
	}
	query := `SELECT t.id, t.empresa_id, t.servicio_id, COALESCE(s.nombre,''), COALESCE(s.color,'#2563eb'), COALESCE(t.puesto_id,0), COALESCE(p.nombre,''), t.numero_dia, t.codigo_turno, COALESCE(t.documento_cliente,''), COALESCE(t.nombre_cliente,''), COALESCE(t.canal_emision,''), COALESCE(t.estado,'espera'), COALESCE(t.observaciones,''), COALESCE(t.fecha_operacion,''), COALESCE(t.fecha_emision,''), COALESCE(t.fecha_llamado,''), COALESCE(t.fecha_inicio_atencion,''), COALESCE(t.fecha_cierre,''), COALESCE(t.tiempo_espera_min,0), COALESCE(t.tiempo_atencion_min,0), COALESCE(t.usuario_creador,'')
		FROM empresa_turnos_atencion_tickets t
		LEFT JOIN empresa_turnos_atencion_servicios s ON s.id = t.servicio_id
		LEFT JOIN empresa_turnos_atencion_puestos p ON p.id = t.puesto_id
		WHERE t.empresa_id = ? AND t.fecha_operacion = ?`
	args := []interface{}{empresaID, fecha}
	if strings.TrimSpace(estado) != "" {
		query += ` AND t.estado = ?`
		args = append(args, strings.TrimSpace(estado))
	}
	query += ` ORDER BY CASE t.estado WHEN 'llamando' THEN 1 WHEN 'atendiendo' THEN 2 WHEN 'espera' THEN 3 ELSE 4 END, t.id ASC LIMIT ?`
	args = append(args, limit)
	rows, err := ExecQueryCompat(dbConn, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]EmpresaTurnoAtencionTicket, 0)
	for rows.Next() {
		var item EmpresaTurnoAtencionTicket
		if err := rows.Scan(&item.ID, &item.EmpresaID, &item.ServicioID, &item.ServicioNombre, &item.ServicioColor, &item.PuestoID, &item.PuestoNombre, &item.NumeroDia, &item.CodigoTurno, &item.DocumentoCliente, &item.NombreCliente, &item.CanalEmision, &item.Estado, &item.Observaciones, &item.FechaOperacion, &item.FechaEmision, &item.FechaLlamado, &item.FechaInicioAtencion, &item.FechaCierre, &item.TiempoEsperaMin, &item.TiempoAtencionMin, &item.UsuarioCreador); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func BuildEmpresaTurnosAtencionDashboard(dbConn *sql.DB, empresaID int64, fecha string) (EmpresaTurnoAtencionDashboard, error) {
	items, err := ListEmpresaTurnosAtencionTickets(dbConn, empresaID, fecha, "", 300)
	if err != nil {
		return EmpresaTurnoAtencionDashboard{}, err
	}
	dashboard := EmpresaTurnoAtencionDashboard{
		EmpresaID:         empresaID,
		Fecha:             defaultText(fecha, time.Now().Format("2006-01-02")),
		Servicios:         make([]EmpresaTurnoAtencionResumen, 0),
		Puestos:           make([]EmpresaTurnoAtencionResumen, 0),
		LlamadosRecientes: make([]EmpresaTurnoAtencionTicket, 0),
	}
	serviceMap := map[string]*EmpresaTurnoAtencionResumen{}
	puestoMap := map[string]*EmpresaTurnoAtencionResumen{}
	var waitSum, waitCount, attendSum, attendCount float64
	for _, item := range items {
		switch item.Estado {
		case "espera":
			dashboard.Esperando++
		case "llamando":
			dashboard.Llamando++
		case "atendiendo":
			dashboard.EnAtencion++
		case "completado":
			dashboard.Completados++
		case "cancelado":
			dashboard.Cancelados++
		}
		accTurnoResumen(serviceMap, item.ServicioNombre, item.Estado)
		accTurnoResumen(puestoMap, defaultText(item.PuestoNombre, "Sin puesto"), item.Estado)
		if item.TiempoEsperaMin > 0 {
			waitSum += item.TiempoEsperaMin
			waitCount++
		}
		if item.TiempoAtencionMin > 0 {
			attendSum += item.TiempoAtencionMin
			attendCount++
		}
		if item.Estado == "llamando" || item.Estado == "atendiendo" || item.Estado == "completado" {
			dashboard.LlamadosRecientes = append(dashboard.LlamadosRecientes, item)
		}
	}
	if waitCount > 0 {
		dashboard.TiempoEsperaPromMin = roundTurnoFloat(waitSum / waitCount)
	}
	if attendCount > 0 {
		dashboard.TiempoAtencionProm = roundTurnoFloat(attendSum / attendCount)
	}
	dashboard.Servicios = flattenTurnoResumen(serviceMap)
	dashboard.Puestos = flattenTurnoResumen(puestoMap)
	if len(dashboard.LlamadosRecientes) > 6 {
		dashboard.LlamadosRecientes = dashboard.LlamadosRecientes[:6]
	}
	return dashboard, nil
}

func BuildEmpresaTurnosAtencionDisplay(dbConn *sql.DB, empresaID int64, fecha string) (EmpresaTurnoAtencionDisplay, error) {
	cfg, err := GetEmpresaTurnoAtencionConfig(dbConn, empresaID)
	if err != nil {
		return EmpresaTurnoAtencionDisplay{}, err
	}
	dashboard, err := BuildEmpresaTurnosAtencionDashboard(dbConn, empresaID, fecha)
	if err != nil {
		return EmpresaTurnoAtencionDisplay{}, err
	}
	out := EmpresaTurnoAtencionDisplay{
		EmpresaID:         empresaID,
		Titulo:            cfg.NombrePantalla,
		Fecha:             dashboard.Fecha,
		ResumenServicios:  dashboard.Servicios,
		TicketsLlamando:   make([]EmpresaTurnoAtencionTicket, 0),
		LlamadosRecientes: make([]EmpresaTurnoAtencionTicket, 0),
	}
	for _, item := range dashboard.LlamadosRecientes {
		if item.Estado == "llamando" || item.Estado == "atendiendo" {
			out.TicketsLlamando = append(out.TicketsLlamando, item)
		} else {
			out.LlamadosRecientes = append(out.LlamadosRecientes, item)
		}
	}
	if len(out.LlamadosRecientes) > 8 {
		out.LlamadosRecientes = out.LlamadosRecientes[:8]
	}
	return out, nil
}

func nextWaitingTicketForPuesto(dbConn *sql.DB, empresaID, puestoID int64) (EmpresaTurnoAtencionTicket, error) {
	var puesto EmpresaTurnoAtencionPuesto
	err := QueryRowCompat(dbConn, `SELECT id, empresa_id, codigo, nombre, COALESCE(area,''), COALESCE(ubicacion,''), COALESCE(servicios_permitidos,''), COALESCE(estado,'activo'), COALESCE(fecha_creacion,''), COALESCE(fecha_actualizacion,''), COALESCE(usuario_creador,'') FROM empresa_turnos_atencion_puestos WHERE empresa_id = ? AND id = ?`, empresaID, puestoID).Scan(&puesto.ID, &puesto.EmpresaID, &puesto.Codigo, &puesto.Nombre, &puesto.Area, &puesto.Ubicacion, &puesto.ServiciosPermitidos, &puesto.Estado, &puesto.FechaCreacion, &puesto.FechaActualizacion, &puesto.UsuarioCreador)
	if err != nil {
		return EmpresaTurnoAtencionTicket{}, err
	}
	query := `SELECT t.id, t.empresa_id, t.servicio_id, COALESCE(s.nombre,''), COALESCE(s.color,'#2563eb'), COALESCE(t.puesto_id,0), COALESCE(p.nombre,''), t.numero_dia, t.codigo_turno, COALESCE(t.documento_cliente,''), COALESCE(t.nombre_cliente,''), COALESCE(t.canal_emision,''), COALESCE(t.estado,'espera'), COALESCE(t.observaciones,''), COALESCE(t.fecha_operacion,''), COALESCE(t.fecha_emision,''), COALESCE(t.fecha_llamado,''), COALESCE(t.fecha_inicio_atencion,''), COALESCE(t.fecha_cierre,''), COALESCE(t.tiempo_espera_min,0), COALESCE(t.tiempo_atencion_min,0), COALESCE(t.usuario_creador,'')
		FROM empresa_turnos_atencion_tickets t
		LEFT JOIN empresa_turnos_atencion_servicios s ON s.id = t.servicio_id
		LEFT JOIN empresa_turnos_atencion_puestos p ON p.id = t.puesto_id
		WHERE t.empresa_id = ? AND t.fecha_operacion = ? AND t.estado = 'espera'`
	args := []interface{}{empresaID, time.Now().Format("2006-01-02")}
	if strings.TrimSpace(puesto.ServiciosPermitidos) != "" {
		tokens := splitCSV(puesto.ServiciosPermitidos)
		if len(tokens) > 0 {
			codePlaceholders := make([]string, 0, len(tokens))
			idPlaceholders := make([]string, 0, len(tokens))
			idArgs := make([]interface{}, 0, len(tokens))
			codeArgs := make([]interface{}, 0, len(tokens))
			for _, token := range tokens {
				token = strings.TrimSpace(token)
				if token == "" {
					continue
				}
				if id, err := strconv.ParseInt(token, 10, 64); err == nil && id > 0 {
					idPlaceholders = append(idPlaceholders, "?")
					idArgs = append(idArgs, id)
					continue
				}
				codePlaceholders = append(codePlaceholders, "?")
				codeArgs = append(codeArgs, strings.ToUpper(token))
			}
			clauses := make([]string, 0, 2)
			if len(idPlaceholders) > 0 {
				clauses = append(clauses, `t.servicio_id IN (`+strings.Join(idPlaceholders, ",")+`)`)
			}
			if len(codePlaceholders) > 0 {
				clauses = append(clauses, `UPPER(COALESCE(s.codigo,'')) IN (`+strings.Join(codePlaceholders, ",")+`)`)
			}
			if len(clauses) > 0 {
				query += ` AND (` + strings.Join(clauses, " OR ") + `)`
				args = append(args, idArgs...)
				args = append(args, codeArgs...)
			}
		}
	}
	query += ` ORDER BY COALESCE(s.prioridad,100) ASC, t.id ASC LIMIT 1`
	row := QueryRowCompat(dbConn, query, args...)
	var item EmpresaTurnoAtencionTicket
	if err := row.Scan(&item.ID, &item.EmpresaID, &item.ServicioID, &item.ServicioNombre, &item.ServicioColor, &item.PuestoID, &item.PuestoNombre, &item.NumeroDia, &item.CodigoTurno, &item.DocumentoCliente, &item.NombreCliente, &item.CanalEmision, &item.Estado, &item.Observaciones, &item.FechaOperacion, &item.FechaEmision, &item.FechaLlamado, &item.FechaInicioAtencion, &item.FechaCierre, &item.TiempoEsperaMin, &item.TiempoAtencionMin, &item.UsuarioCreador); err != nil {
		return EmpresaTurnoAtencionTicket{}, err
	}
	return item, nil
}

func LlamarSiguienteTurnoAtencion(dbConn *sql.DB, empresaID, puestoID int64, usuario string) (EmpresaTurnoAtencionTicket, error) {
	if err := EnsureEmpresaTurnosAtencionSchema(dbConn); err != nil {
		return EmpresaTurnoAtencionTicket{}, err
	}
	item, err := nextWaitingTicketForPuesto(dbConn, empresaID, puestoID)
	if err != nil {
		return EmpresaTurnoAtencionTicket{}, err
	}
	item.FechaLlamado = time.Now().Format("2006-01-02 15:04:05")
	item.Estado = "llamando"
	item.PuestoID = puestoID
	var puestoNombre string
	_ = QueryRowCompat(dbConn, `SELECT COALESCE(nombre,'') FROM empresa_turnos_atencion_puestos WHERE id = ?`, puestoID).Scan(&puestoNombre)
	item.PuestoNombre = puestoNombre
	waitMin := minutesBetween(item.FechaEmision, item.FechaLlamado)
	item.TiempoEsperaMin = waitMin
	_, err = ExecCompat(dbConn, `UPDATE empresa_turnos_atencion_tickets SET puesto_id = ?, estado = 'llamando', fecha_llamado = ?, tiempo_espera_min = ?, usuario_creador = ? WHERE id = ? AND empresa_id = ?`, puestoID, item.FechaLlamado, item.TiempoEsperaMin, strings.TrimSpace(usuario), item.ID, empresaID)
	if err != nil {
		return EmpresaTurnoAtencionTicket{}, err
	}
	return item, nil
}

func CambiarEstadoTurnoAtencion(dbConn *sql.DB, empresaID, ticketID, puestoID int64, nuevoEstado, usuario, observaciones string) error {
	if err := EnsureEmpresaTurnosAtencionSchema(dbConn); err != nil {
		return err
	}
	var item EmpresaTurnoAtencionTicket
	err := QueryRowCompat(dbConn, `SELECT id, empresa_id, servicio_id, COALESCE(numero_dia,0), COALESCE(codigo_turno,''), COALESCE(fecha_emision,''), COALESCE(fecha_llamado,''), COALESCE(fecha_inicio_atencion,''), COALESCE(estado,'espera') FROM empresa_turnos_atencion_tickets WHERE empresa_id = ? AND id = ?`, empresaID, ticketID).Scan(&item.ID, &item.EmpresaID, &item.ServicioID, &item.NumeroDia, &item.CodigoTurno, &item.FechaEmision, &item.FechaLlamado, &item.FechaInicioAtencion, &item.Estado)
	if err != nil {
		return err
	}
	nuevoEstado = strings.TrimSpace(strings.ToLower(nuevoEstado))
	now := time.Now().Format("2006-01-02 15:04:05")
	switch nuevoEstado {
	case "atendiendo":
		_, err = ExecCompat(dbConn, `UPDATE empresa_turnos_atencion_tickets SET puesto_id = ?, estado = 'atendiendo', fecha_inicio_atencion = ?, observaciones = ? WHERE id = ? AND empresa_id = ?`, puestoID, now, strings.TrimSpace(observaciones), ticketID, empresaID)
	case "completado":
		atencionMin := minutesBetween(defaultText(item.FechaInicioAtencion, item.FechaLlamado), now)
		_, err = ExecCompat(dbConn, `UPDATE empresa_turnos_atencion_tickets SET estado = 'completado', fecha_cierre = ?, tiempo_atencion_min = ?, observaciones = ?, usuario_creador = ? WHERE id = ? AND empresa_id = ?`, now, atencionMin, strings.TrimSpace(observaciones), strings.TrimSpace(usuario), ticketID, empresaID)
	case "cancelado":
		_, err = ExecCompat(dbConn, `UPDATE empresa_turnos_atencion_tickets SET estado = 'cancelado', fecha_cierre = ?, observaciones = ?, usuario_creador = ? WHERE id = ? AND empresa_id = ?`, now, strings.TrimSpace(observaciones), strings.TrimSpace(usuario), ticketID, empresaID)
	case "llamando":
		waitMin := minutesBetween(item.FechaEmision, now)
		_, err = ExecCompat(dbConn, `UPDATE empresa_turnos_atencion_tickets SET puesto_id = ?, estado = 'llamando', fecha_llamado = ?, tiempo_espera_min = ?, observaciones = ?, usuario_creador = ? WHERE id = ? AND empresa_id = ?`, puestoID, now, waitMin, strings.TrimSpace(observaciones), strings.TrimSpace(usuario), ticketID, empresaID)
	default:
		return fmt.Errorf("estado no soportado")
	}
	return err
}

func defaultText(v, fallback string) string {
	if strings.TrimSpace(v) == "" {
		return fallback
	}
	return strings.TrimSpace(v)
}

func turnosBoolInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

func splitCSV(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		p := strings.TrimSpace(part)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func accTurnoResumen(target map[string]*EmpresaTurnoAtencionResumen, key, estado string) {
	key = defaultText(key, "General")
	if target[key] == nil {
		target[key] = &EmpresaTurnoAtencionResumen{Clave: strings.ToLower(key), Etiqueta: key}
	}
	target[key].Cantidad++
	if estado == "espera" {
		target[key].EnEspera++
	}
	if estado == "atendiendo" || estado == "llamando" {
		target[key].EnAtencion++
	}
}

func flattenTurnoResumen(source map[string]*EmpresaTurnoAtencionResumen) []EmpresaTurnoAtencionResumen {
	out := make([]EmpresaTurnoAtencionResumen, 0, len(source))
	for _, item := range source {
		out = append(out, *item)
	}
	// simple deterministic sort
	for i := 0; i < len(out); i++ {
		for j := i + 1; j < len(out); j++ {
			if out[j].Cantidad > out[i].Cantidad || (out[j].Cantidad == out[i].Cantidad && out[j].Etiqueta < out[i].Etiqueta) {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
	return out
}

func minutesBetween(from, to string) float64 {
	if strings.TrimSpace(from) == "" || strings.TrimSpace(to) == "" {
		return 0
	}
	start, err := time.Parse("2006-01-02 15:04:05", from)
	if err != nil {
		return 0
	}
	end, err := time.Parse("2006-01-02 15:04:05", to)
	if err != nil {
		return 0
	}
	return roundTurnoFloat(end.Sub(start).Minutes())
}

func roundTurnoFloat(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}
