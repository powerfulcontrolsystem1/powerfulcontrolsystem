package db

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type EmpresaGimnasioSocio struct {
	ID                         int64   `json:"id"`
	EmpresaID                  int64   `json:"empresa_id"`
	Codigo                     string  `json:"codigo"`
	NombreCompleto             string  `json:"nombre_completo"`
	Documento                  string  `json:"documento,omitempty"`
	Telefono                   string  `json:"telefono,omitempty"`
	Email                      string  `json:"email,omitempty"`
	FechaNacimiento            string  `json:"fecha_nacimiento,omitempty"`
	Genero                     string  `json:"genero,omitempty"`
	Objetivo                   string  `json:"objetivo,omitempty"`
	Estado                     string  `json:"estado,omitempty"`
	PlanID                     int64   `json:"plan_id,omitempty"`
	PlanNombre                 string  `json:"plan_nombre,omitempty"`
	FechaInicioPlan            string  `json:"fecha_inicio_plan,omitempty"`
	FechaFinPlan               string  `json:"fecha_fin_plan,omitempty"`
	Saldo                      float64 `json:"saldo"`
	ContactoEmergenciaNombre   string  `json:"contacto_emergencia_nombre,omitempty"`
	ContactoEmergenciaTelefono string  `json:"contacto_emergencia_telefono,omitempty"`
	Observaciones              string  `json:"observaciones,omitempty"`
	FechaCreacion              string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion         string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador             string  `json:"usuario_creador,omitempty"`
}

type EmpresaGimnasioPlan struct {
	ID                     int64   `json:"id"`
	EmpresaID              int64   `json:"empresa_id"`
	Nombre                 string  `json:"nombre"`
	Descripcion            string  `json:"descripcion,omitempty"`
	Precio                 float64 `json:"precio"`
	DuracionDias           int     `json:"duracion_dias"`
	ClasesIncluidas        int     `json:"clases_incluidas"`
	AccesoIlimitado        bool    `json:"acceso_ilimitado"`
	SesionesPersonalizadas int     `json:"sesiones_personalizadas"`
	Estado                 string  `json:"estado,omitempty"`
	FechaCreacion          string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion     string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador         string  `json:"usuario_creador,omitempty"`
}

type EmpresaGimnasioEntrenador struct {
	ID                 int64  `json:"id"`
	EmpresaID          int64  `json:"empresa_id"`
	NombreCompleto     string `json:"nombre_completo"`
	Especialidad       string `json:"especialidad,omitempty"`
	Telefono           string `json:"telefono,omitempty"`
	Email              string `json:"email,omitempty"`
	Certificaciones    string `json:"certificaciones,omitempty"`
	Estado             string `json:"estado,omitempty"`
	Disponibilidad     string `json:"disponibilidad,omitempty"`
	Observaciones      string `json:"observaciones,omitempty"`
	FechaCreacion      string `json:"fecha_creacion,omitempty"`
	FechaActualizacion string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string `json:"usuario_creador,omitempty"`
}

type EmpresaGimnasioClase struct {
	ID                 int64   `json:"id"`
	EmpresaID          int64   `json:"empresa_id"`
	Nombre             string  `json:"nombre"`
	Categoria          string  `json:"categoria,omitempty"`
	EntrenadorID       int64   `json:"entrenador_id,omitempty"`
	EntrenadorNombre   string  `json:"entrenador_nombre,omitempty"`
	Sede               string  `json:"sede,omitempty"`
	Canal              string  `json:"canal,omitempty"`
	Cupos              int     `json:"cupos"`
	DuracionMinutos    int     `json:"duracion_minutos"`
	FechaProgramada    string  `json:"fecha_programada,omitempty"`
	Estado             string  `json:"estado,omitempty"`
	Precio             float64 `json:"precio"`
	Descripcion        string  `json:"descripcion,omitempty"`
	FechaCreacion      string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string  `json:"usuario_creador,omitempty"`
}

type EmpresaGimnasioInscripcion struct {
	ID                int64  `json:"id"`
	EmpresaID         int64  `json:"empresa_id"`
	SocioID           int64  `json:"socio_id"`
	SocioNombre       string `json:"socio_nombre,omitempty"`
	ClaseID           int64  `json:"clase_id"`
	ClaseNombre       string `json:"clase_nombre,omitempty"`
	Estado            string `json:"estado,omitempty"`
	FechaInscripcion  string `json:"fecha_inscripcion,omitempty"`
	AsistenciaMarcada bool   `json:"asistencia_marcada"`
	Observaciones     string `json:"observaciones,omitempty"`
	FechaCreacion     string `json:"fecha_creacion,omitempty"`
	UsuarioCreador    string `json:"usuario_creador,omitempty"`
}

type EmpresaGimnasioAsistencia struct {
	ID             int64  `json:"id"`
	EmpresaID      int64  `json:"empresa_id"`
	SocioID        int64  `json:"socio_id"`
	SocioNombre    string `json:"socio_nombre,omitempty"`
	ClaseID        int64  `json:"clase_id,omitempty"`
	ClaseNombre    string `json:"clase_nombre,omitempty"`
	FechaHora      string `json:"fecha_hora,omitempty"`
	TipoAcceso     string `json:"tipo_acceso,omitempty"`
	Canal          string `json:"canal,omitempty"`
	Sede           string `json:"sede,omitempty"`
	Observaciones  string `json:"observaciones,omitempty"`
	FechaCreacion  string `json:"fecha_creacion,omitempty"`
	UsuarioCreador string `json:"usuario_creador,omitempty"`
}

type EmpresaGimnasioPago struct {
	ID             int64   `json:"id"`
	EmpresaID      int64   `json:"empresa_id"`
	SocioID        int64   `json:"socio_id"`
	SocioNombre    string  `json:"socio_nombre,omitempty"`
	PlanID         int64   `json:"plan_id,omitempty"`
	PlanNombre     string  `json:"plan_nombre,omitempty"`
	Concepto       string  `json:"concepto"`
	Monto          float64 `json:"monto"`
	Moneda         string  `json:"moneda,omitempty"`
	MetodoPago     string  `json:"metodo_pago,omitempty"`
	Canal          string  `json:"canal,omitempty"`
	Sede           string  `json:"sede,omitempty"`
	Estado         string  `json:"estado,omitempty"`
	Referencia     string  `json:"referencia,omitempty"`
	FechaPago      string  `json:"fecha_pago,omitempty"`
	Observaciones  string  `json:"observaciones,omitempty"`
	FechaCreacion  string  `json:"fecha_creacion,omitempty"`
	UsuarioCreador string  `json:"usuario_creador,omitempty"`
}

type EmpresaGimnasioResumenGrupo struct {
	Clave     string  `json:"clave"`
	Etiqueta  string  `json:"etiqueta"`
	Cantidad  int     `json:"cantidad"`
	Monto     float64 `json:"monto"`
	Margen    float64 `json:"margen"`
	Ocupacion float64 `json:"ocupacion"`
}

type EmpresaGimnasioDashboard struct {
	EmpresaID            int64                         `json:"empresa_id"`
	SociosActivos        int                           `json:"socios_activos"`
	PlanesActivos        int                           `json:"planes_activos"`
	ClasesHoy            int                           `json:"clases_hoy"`
	AccesosHoy           int                           `json:"accesos_hoy"`
	IngresosMes          float64                       `json:"ingresos_mes"`
	RenovacionesProximas int                           `json:"renovaciones_proximas"`
	InscripcionesActivas int                           `json:"inscripciones_activas"`
	VencimientosProximos []EmpresaGimnasioSocio        `json:"vencimientos_proximos"`
	ClasesProgramadasHoy []EmpresaGimnasioClase        `json:"clases_programadas_hoy"`
	IngresosPorCanal     []EmpresaGimnasioResumenGrupo `json:"ingresos_por_canal"`
	RentabilidadPorLinea []EmpresaGimnasioResumenGrupo `json:"rentabilidad_por_linea"`
	RentabilidadPorSede  []EmpresaGimnasioResumenGrupo `json:"rentabilidad_por_sede"`
}

func EnsureEmpresaGimnasioSchema(dbConn *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_gimnasio_planes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			nombre TEXT NOT NULL,
			descripcion TEXT,
			precio REAL DEFAULT 0,
			duracion_dias INTEGER DEFAULT 30,
			clases_incluidas INTEGER DEFAULT 0,
			acceso_ilimitado INTEGER DEFAULT 0,
			sesiones_personalizadas INTEGER DEFAULT 0,
			estado TEXT DEFAULT 'activo',
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT
		);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_gimnasio_planes_empresa ON empresa_gimnasio_planes(empresa_id, estado, id DESC);`,
		`CREATE TABLE IF NOT EXISTS empresa_gimnasio_socios (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			codigo TEXT,
			nombre_completo TEXT NOT NULL,
			documento TEXT,
			telefono TEXT,
			email TEXT,
			fecha_nacimiento TEXT,
			genero TEXT,
			objetivo TEXT,
			estado TEXT DEFAULT 'activo',
			plan_id INTEGER,
			fecha_inicio_plan TEXT,
			fecha_fin_plan TEXT,
			saldo REAL DEFAULT 0,
			contacto_emergencia_nombre TEXT,
			contacto_emergencia_telefono TEXT,
			observaciones TEXT,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT
		);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_gimnasio_socios_empresa ON empresa_gimnasio_socios(empresa_id, estado, id DESC);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_gimnasio_socios_plan ON empresa_gimnasio_socios(empresa_id, plan_id);`,
		`CREATE TABLE IF NOT EXISTS empresa_gimnasio_entrenadores (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			nombre_completo TEXT NOT NULL,
			especialidad TEXT,
			telefono TEXT,
			email TEXT,
			certificaciones TEXT,
			estado TEXT DEFAULT 'activo',
			disponibilidad TEXT,
			observaciones TEXT,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT
		);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_gimnasio_entrenadores_empresa ON empresa_gimnasio_entrenadores(empresa_id, estado, id DESC);`,
		`CREATE TABLE IF NOT EXISTS empresa_gimnasio_clases (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			nombre TEXT NOT NULL,
			categoria TEXT,
			entrenador_id INTEGER,
			sede TEXT,
			canal TEXT,
			cupos INTEGER DEFAULT 0,
			duracion_minutos INTEGER DEFAULT 60,
			fecha_programada TEXT,
			estado TEXT DEFAULT 'programada',
			precio REAL DEFAULT 0,
			descripcion TEXT,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT
		);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_gimnasio_clases_empresa_fecha ON empresa_gimnasio_clases(empresa_id, fecha_programada DESC, id DESC);`,
		`CREATE TABLE IF NOT EXISTS empresa_gimnasio_inscripciones (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			socio_id INTEGER NOT NULL,
			clase_id INTEGER NOT NULL,
			estado TEXT DEFAULT 'activa',
			fecha_inscripcion TEXT DEFAULT (datetime('now','localtime')),
			asistencia_marcada INTEGER DEFAULT 0,
			observaciones TEXT,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT
		);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_gimnasio_inscripciones_empresa ON empresa_gimnasio_inscripciones(empresa_id, estado, clase_id, socio_id);`,
		`CREATE TABLE IF NOT EXISTS empresa_gimnasio_asistencias (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			socio_id INTEGER NOT NULL,
			clase_id INTEGER,
			fecha_hora TEXT DEFAULT (datetime('now','localtime')),
			tipo_acceso TEXT DEFAULT 'checkin',
			canal TEXT,
			sede TEXT,
			observaciones TEXT,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT
		);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_gimnasio_asistencias_empresa_fecha ON empresa_gimnasio_asistencias(empresa_id, fecha_hora DESC, id DESC);`,
		`CREATE TABLE IF NOT EXISTS empresa_gimnasio_pagos (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			socio_id INTEGER NOT NULL,
			plan_id INTEGER,
			concepto TEXT NOT NULL,
			monto REAL DEFAULT 0,
			moneda TEXT DEFAULT 'COP',
			metodo_pago TEXT DEFAULT 'efectivo',
			canal TEXT,
			sede TEXT,
			estado TEXT DEFAULT 'pagado',
			referencia TEXT,
			fecha_pago TEXT DEFAULT (datetime('now','localtime')),
			observaciones TEXT,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT
		);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_gimnasio_pagos_empresa_fecha ON empresa_gimnasio_pagos(empresa_id, fecha_pago DESC, id DESC);`,
	}
	for _, stmt := range stmts {
		if _, err := dbConn.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}

func normalizeGymState(raw, fallback string) string {
	v := strings.ToLower(strings.TrimSpace(raw))
	if v == "" {
		return fallback
	}
	return v
}

func normalizeGymSocio(payload EmpresaGimnasioSocio) (*EmpresaGimnasioSocio, error) {
	if payload.EmpresaID <= 0 {
		return nil, fmt.Errorf("empresa_id es obligatorio")
	}
	out := payload
	out.Codigo = strings.TrimSpace(out.Codigo)
	out.NombreCompleto = strings.TrimSpace(out.NombreCompleto)
	out.Documento = strings.TrimSpace(out.Documento)
	out.Telefono = strings.TrimSpace(out.Telefono)
	out.Email = strings.TrimSpace(out.Email)
	out.Genero = strings.TrimSpace(out.Genero)
	out.Objetivo = strings.TrimSpace(out.Objetivo)
	out.Estado = normalizeGymState(out.Estado, "activo")
	out.ContactoEmergenciaNombre = strings.TrimSpace(out.ContactoEmergenciaNombre)
	out.ContactoEmergenciaTelefono = strings.TrimSpace(out.ContactoEmergenciaTelefono)
	out.Observaciones = strings.TrimSpace(out.Observaciones)
	out.UsuarioCreador = strings.TrimSpace(out.UsuarioCreador)
	if out.NombreCompleto == "" {
		return nil, fmt.Errorf("nombre_completo es obligatorio")
	}
	if out.Codigo == "" {
		out.Codigo = fmt.Sprintf("GYM-%d", time.Now().Unix())
	}
	return &out, nil
}

func normalizeGymPlan(payload EmpresaGimnasioPlan) (*EmpresaGimnasioPlan, error) {
	if payload.EmpresaID <= 0 {
		return nil, fmt.Errorf("empresa_id es obligatorio")
	}
	out := payload
	out.Nombre = strings.TrimSpace(out.Nombre)
	out.Descripcion = strings.TrimSpace(out.Descripcion)
	out.Estado = normalizeGymState(out.Estado, "activo")
	out.UsuarioCreador = strings.TrimSpace(out.UsuarioCreador)
	if out.Nombre == "" {
		return nil, fmt.Errorf("nombre es obligatorio")
	}
	if out.DuracionDias <= 0 {
		out.DuracionDias = 30
	}
	if out.ClasesIncluidas < 0 {
		out.ClasesIncluidas = 0
	}
	if out.SesionesPersonalizadas < 0 {
		out.SesionesPersonalizadas = 0
	}
	return &out, nil
}

func normalizeGymEntrenador(payload EmpresaGimnasioEntrenador) (*EmpresaGimnasioEntrenador, error) {
	if payload.EmpresaID <= 0 {
		return nil, fmt.Errorf("empresa_id es obligatorio")
	}
	out := payload
	out.NombreCompleto = strings.TrimSpace(out.NombreCompleto)
	out.Especialidad = strings.TrimSpace(out.Especialidad)
	out.Telefono = strings.TrimSpace(out.Telefono)
	out.Email = strings.TrimSpace(out.Email)
	out.Certificaciones = strings.TrimSpace(out.Certificaciones)
	out.Estado = normalizeGymState(out.Estado, "activo")
	out.Disponibilidad = strings.TrimSpace(out.Disponibilidad)
	out.Observaciones = strings.TrimSpace(out.Observaciones)
	out.UsuarioCreador = strings.TrimSpace(out.UsuarioCreador)
	if out.NombreCompleto == "" {
		return nil, fmt.Errorf("nombre_completo es obligatorio")
	}
	return &out, nil
}

func normalizeGymClase(payload EmpresaGimnasioClase) (*EmpresaGimnasioClase, error) {
	if payload.EmpresaID <= 0 {
		return nil, fmt.Errorf("empresa_id es obligatorio")
	}
	out := payload
	out.Nombre = strings.TrimSpace(out.Nombre)
	out.Categoria = strings.TrimSpace(out.Categoria)
	out.Sede = strings.TrimSpace(out.Sede)
	out.Canal = strings.TrimSpace(out.Canal)
	out.FechaProgramada = strings.TrimSpace(out.FechaProgramada)
	out.Estado = normalizeGymState(out.Estado, "programada")
	out.Descripcion = strings.TrimSpace(out.Descripcion)
	out.UsuarioCreador = strings.TrimSpace(out.UsuarioCreador)
	if out.Nombre == "" {
		return nil, fmt.Errorf("nombre es obligatorio")
	}
	if out.Cupos <= 0 {
		out.Cupos = 20
	}
	if out.DuracionMinutos <= 0 {
		out.DuracionMinutos = 60
	}
	if out.Canal == "" {
		out.Canal = "presencial"
	}
	if out.Sede == "" {
		out.Sede = "principal"
	}
	return &out, nil
}

func normalizeGymInscripcion(payload EmpresaGimnasioInscripcion) (*EmpresaGimnasioInscripcion, error) {
	if payload.EmpresaID <= 0 || payload.SocioID <= 0 || payload.ClaseID <= 0 {
		return nil, fmt.Errorf("empresa_id, socio_id y clase_id son obligatorios")
	}
	out := payload
	out.Estado = normalizeGymState(out.Estado, "activa")
	out.Observaciones = strings.TrimSpace(out.Observaciones)
	out.UsuarioCreador = strings.TrimSpace(out.UsuarioCreador)
	return &out, nil
}

func normalizeGymAsistencia(payload EmpresaGimnasioAsistencia) (*EmpresaGimnasioAsistencia, error) {
	if payload.EmpresaID <= 0 || payload.SocioID <= 0 {
		return nil, fmt.Errorf("empresa_id y socio_id son obligatorios")
	}
	out := payload
	out.FechaHora = strings.TrimSpace(out.FechaHora)
	out.TipoAcceso = normalizeGymState(out.TipoAcceso, "checkin")
	out.Canal = strings.TrimSpace(out.Canal)
	out.Sede = strings.TrimSpace(out.Sede)
	out.Observaciones = strings.TrimSpace(out.Observaciones)
	out.UsuarioCreador = strings.TrimSpace(out.UsuarioCreador)
	if out.FechaHora == "" {
		out.FechaHora = time.Now().Format("2006-01-02 15:04:05")
	}
	if out.Canal == "" {
		out.Canal = "recepcion"
	}
	if out.Sede == "" {
		out.Sede = "principal"
	}
	return &out, nil
}

func normalizeGymPago(payload EmpresaGimnasioPago) (*EmpresaGimnasioPago, error) {
	if payload.EmpresaID <= 0 || payload.SocioID <= 0 {
		return nil, fmt.Errorf("empresa_id y socio_id son obligatorios")
	}
	out := payload
	out.Concepto = strings.TrimSpace(out.Concepto)
	out.Moneda = strings.ToUpper(strings.TrimSpace(out.Moneda))
	out.MetodoPago = strings.TrimSpace(out.MetodoPago)
	out.Canal = strings.TrimSpace(out.Canal)
	out.Sede = strings.TrimSpace(out.Sede)
	out.Estado = normalizeGymState(out.Estado, "pagado")
	out.Referencia = strings.TrimSpace(out.Referencia)
	out.FechaPago = strings.TrimSpace(out.FechaPago)
	out.Observaciones = strings.TrimSpace(out.Observaciones)
	out.UsuarioCreador = strings.TrimSpace(out.UsuarioCreador)
	if out.Concepto == "" {
		return nil, fmt.Errorf("concepto es obligatorio")
	}
	if out.Monto <= 0 {
		return nil, fmt.Errorf("monto debe ser mayor que cero")
	}
	if out.Moneda == "" {
		out.Moneda = "COP"
	}
	if out.MetodoPago == "" {
		out.MetodoPago = "efectivo"
	}
	if out.Canal == "" {
		out.Canal = "mostrador"
	}
	if out.Sede == "" {
		out.Sede = "principal"
	}
	if out.FechaPago == "" {
		out.FechaPago = time.Now().Format("2006-01-02 15:04:05")
	}
	return &out, nil
}

func ListEmpresaGimnasioSocios(dbConn *sql.DB, empresaID int64) ([]EmpresaGimnasioSocio, error) {
	if empresaID <= 0 {
		return nil, fmt.Errorf("empresa_id es obligatorio")
	}
	if err := EnsureEmpresaGimnasioSchema(dbConn); err != nil {
		return nil, err
	}
	rows, err := dbConn.Query(`SELECT
		s.id, s.empresa_id, COALESCE(s.codigo,''), COALESCE(s.nombre_completo,''), COALESCE(s.documento,''), COALESCE(s.telefono,''),
		COALESCE(s.email,''), COALESCE(s.fecha_nacimiento,''), COALESCE(s.genero,''), COALESCE(s.objetivo,''), COALESCE(s.estado,'activo'),
		COALESCE(s.plan_id,0), COALESCE(p.nombre,''), COALESCE(s.fecha_inicio_plan,''), COALESCE(s.fecha_fin_plan,''), COALESCE(s.saldo,0),
		COALESCE(s.contacto_emergencia_nombre,''), COALESCE(s.contacto_emergencia_telefono,''), COALESCE(s.observaciones,''), COALESCE(s.fecha_creacion,''),
		COALESCE(s.fecha_actualizacion,''), COALESCE(s.usuario_creador,'')
	FROM empresa_gimnasio_socios s
	LEFT JOIN empresa_gimnasio_planes p ON p.id = s.plan_id AND p.empresa_id = s.empresa_id
	WHERE s.empresa_id = ?
	ORDER BY s.id DESC`, empresaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []EmpresaGimnasioSocio
	for rows.Next() {
		var item EmpresaGimnasioSocio
		if err := rows.Scan(
			&item.ID, &item.EmpresaID, &item.Codigo, &item.NombreCompleto, &item.Documento, &item.Telefono,
			&item.Email, &item.FechaNacimiento, &item.Genero, &item.Objetivo, &item.Estado,
			&item.PlanID, &item.PlanNombre, &item.FechaInicioPlan, &item.FechaFinPlan, &item.Saldo,
			&item.ContactoEmergenciaNombre, &item.ContactoEmergenciaTelefono, &item.Observaciones, &item.FechaCreacion,
			&item.FechaActualizacion, &item.UsuarioCreador,
		); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func CreateEmpresaGimnasioSocio(dbConn *sql.DB, payload EmpresaGimnasioSocio) (int64, error) {
	if err := EnsureEmpresaGimnasioSchema(dbConn); err != nil {
		return 0, err
	}
	item, err := normalizeGymSocio(payload)
	if err != nil {
		return 0, err
	}
	res, err := dbConn.Exec(`INSERT INTO empresa_gimnasio_socios (
		empresa_id, codigo, nombre_completo, documento, telefono, email, fecha_nacimiento, genero, objetivo, estado,
		plan_id, fecha_inicio_plan, fecha_fin_plan, saldo, contacto_emergencia_nombre, contacto_emergencia_telefono,
		observaciones, usuario_creador, fecha_actualizacion
	) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,datetime('now','localtime'))`,
		item.EmpresaID, item.Codigo, item.NombreCompleto, item.Documento, item.Telefono, item.Email, item.FechaNacimiento, item.Genero, item.Objetivo, item.Estado,
		item.PlanID, item.FechaInicioPlan, item.FechaFinPlan, item.Saldo, item.ContactoEmergenciaNombre, item.ContactoEmergenciaTelefono,
		item.Observaciones, item.UsuarioCreador,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func UpdateEmpresaGimnasioSocio(dbConn *sql.DB, payload EmpresaGimnasioSocio) error {
	if err := EnsureEmpresaGimnasioSchema(dbConn); err != nil {
		return err
	}
	item, err := normalizeGymSocio(payload)
	if err != nil {
		return err
	}
	if item.ID <= 0 {
		return fmt.Errorf("id es obligatorio")
	}
	res, err := dbConn.Exec(`UPDATE empresa_gimnasio_socios SET
		codigo=?, nombre_completo=?, documento=?, telefono=?, email=?, fecha_nacimiento=?, genero=?, objetivo=?, estado=?,
		plan_id=?, fecha_inicio_plan=?, fecha_fin_plan=?, saldo=?, contacto_emergencia_nombre=?, contacto_emergencia_telefono=?,
		observaciones=?, fecha_actualizacion=datetime('now','localtime')
	WHERE id=? AND empresa_id=?`,
		item.Codigo, item.NombreCompleto, item.Documento, item.Telefono, item.Email, item.FechaNacimiento, item.Genero, item.Objetivo, item.Estado,
		item.PlanID, item.FechaInicioPlan, item.FechaFinPlan, item.Saldo, item.ContactoEmergenciaNombre, item.ContactoEmergenciaTelefono,
		item.Observaciones, item.ID, item.EmpresaID,
	)
	if err != nil {
		return err
	}
	return ensureRowsAffected(res)
}

func DeleteEmpresaGimnasioSocio(dbConn *sql.DB, empresaID, id int64) error {
	return simpleDeleteByEmpresa(dbConn, empresaID, id, "empresa_gimnasio_socios")
}

func ListEmpresaGimnasioPlanes(dbConn *sql.DB, empresaID int64) ([]EmpresaGimnasioPlan, error) {
	if empresaID <= 0 {
		return nil, fmt.Errorf("empresa_id es obligatorio")
	}
	if err := EnsureEmpresaGimnasioSchema(dbConn); err != nil {
		return nil, err
	}
	rows, err := dbConn.Query(`SELECT
		id, empresa_id, COALESCE(nombre,''), COALESCE(descripcion,''), COALESCE(precio,0), COALESCE(duracion_dias,30),
		COALESCE(clases_incluidas,0), COALESCE(acceso_ilimitado,0), COALESCE(sesiones_personalizadas,0), COALESCE(estado,'activo'),
		COALESCE(fecha_creacion,''), COALESCE(fecha_actualizacion,''), COALESCE(usuario_creador,'')
	FROM empresa_gimnasio_planes
	WHERE empresa_id = ?
	ORDER BY id DESC`, empresaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []EmpresaGimnasioPlan
	for rows.Next() {
		var item EmpresaGimnasioPlan
		var accesoIlimitado int
		if err := rows.Scan(&item.ID, &item.EmpresaID, &item.Nombre, &item.Descripcion, &item.Precio, &item.DuracionDias,
			&item.ClasesIncluidas, &accesoIlimitado, &item.SesionesPersonalizadas, &item.Estado,
			&item.FechaCreacion, &item.FechaActualizacion, &item.UsuarioCreador); err != nil {
			return nil, err
		}
		item.AccesoIlimitado = accesoIlimitado > 0
		out = append(out, item)
	}
	return out, rows.Err()
}

func CreateEmpresaGimnasioPlan(dbConn *sql.DB, payload EmpresaGimnasioPlan) (int64, error) {
	if err := EnsureEmpresaGimnasioSchema(dbConn); err != nil {
		return 0, err
	}
	item, err := normalizeGymPlan(payload)
	if err != nil {
		return 0, err
	}
	accesoIlimitado := 0
	if item.AccesoIlimitado {
		accesoIlimitado = 1
	}
	res, err := dbConn.Exec(`INSERT INTO empresa_gimnasio_planes (
		empresa_id, nombre, descripcion, precio, duracion_dias, clases_incluidas, acceso_ilimitado, sesiones_personalizadas,
		estado, usuario_creador, fecha_actualizacion
	) VALUES (?,?,?,?,?,?,?,?,?,?,datetime('now','localtime'))`,
		item.EmpresaID, item.Nombre, item.Descripcion, item.Precio, item.DuracionDias, item.ClasesIncluidas, accesoIlimitado, item.SesionesPersonalizadas,
		item.Estado, item.UsuarioCreador,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func UpdateEmpresaGimnasioPlan(dbConn *sql.DB, payload EmpresaGimnasioPlan) error {
	if err := EnsureEmpresaGimnasioSchema(dbConn); err != nil {
		return err
	}
	item, err := normalizeGymPlan(payload)
	if err != nil {
		return err
	}
	if item.ID <= 0 {
		return fmt.Errorf("id es obligatorio")
	}
	accesoIlimitado := 0
	if item.AccesoIlimitado {
		accesoIlimitado = 1
	}
	res, err := dbConn.Exec(`UPDATE empresa_gimnasio_planes SET
		nombre=?, descripcion=?, precio=?, duracion_dias=?, clases_incluidas=?, acceso_ilimitado=?, sesiones_personalizadas=?, estado=?,
		fecha_actualizacion=datetime('now','localtime')
	WHERE id=? AND empresa_id=?`,
		item.Nombre, item.Descripcion, item.Precio, item.DuracionDias, item.ClasesIncluidas, accesoIlimitado, item.SesionesPersonalizadas, item.Estado,
		item.ID, item.EmpresaID,
	)
	if err != nil {
		return err
	}
	return ensureRowsAffected(res)
}

func DeleteEmpresaGimnasioPlan(dbConn *sql.DB, empresaID, id int64) error {
	return simpleDeleteByEmpresa(dbConn, empresaID, id, "empresa_gimnasio_planes")
}

func ListEmpresaGimnasioEntrenadores(dbConn *sql.DB, empresaID int64) ([]EmpresaGimnasioEntrenador, error) {
	if empresaID <= 0 {
		return nil, fmt.Errorf("empresa_id es obligatorio")
	}
	if err := EnsureEmpresaGimnasioSchema(dbConn); err != nil {
		return nil, err
	}
	rows, err := dbConn.Query(`SELECT
		id, empresa_id, COALESCE(nombre_completo,''), COALESCE(especialidad,''), COALESCE(telefono,''), COALESCE(email,''),
		COALESCE(certificaciones,''), COALESCE(estado,'activo'), COALESCE(disponibilidad,''), COALESCE(observaciones,''),
		COALESCE(fecha_creacion,''), COALESCE(fecha_actualizacion,''), COALESCE(usuario_creador,'')
	FROM empresa_gimnasio_entrenadores
	WHERE empresa_id = ?
	ORDER BY id DESC`, empresaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []EmpresaGimnasioEntrenador
	for rows.Next() {
		var item EmpresaGimnasioEntrenador
		if err := rows.Scan(&item.ID, &item.EmpresaID, &item.NombreCompleto, &item.Especialidad, &item.Telefono, &item.Email,
			&item.Certificaciones, &item.Estado, &item.Disponibilidad, &item.Observaciones,
			&item.FechaCreacion, &item.FechaActualizacion, &item.UsuarioCreador); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func CreateEmpresaGimnasioEntrenador(dbConn *sql.DB, payload EmpresaGimnasioEntrenador) (int64, error) {
	if err := EnsureEmpresaGimnasioSchema(dbConn); err != nil {
		return 0, err
	}
	item, err := normalizeGymEntrenador(payload)
	if err != nil {
		return 0, err
	}
	res, err := dbConn.Exec(`INSERT INTO empresa_gimnasio_entrenadores (
		empresa_id, nombre_completo, especialidad, telefono, email, certificaciones, estado, disponibilidad, observaciones,
		usuario_creador, fecha_actualizacion
	) VALUES (?,?,?,?,?,?,?,?,?,?,datetime('now','localtime'))`,
		item.EmpresaID, item.NombreCompleto, item.Especialidad, item.Telefono, item.Email, item.Certificaciones, item.Estado, item.Disponibilidad, item.Observaciones,
		item.UsuarioCreador,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func UpdateEmpresaGimnasioEntrenador(dbConn *sql.DB, payload EmpresaGimnasioEntrenador) error {
	if err := EnsureEmpresaGimnasioSchema(dbConn); err != nil {
		return err
	}
	item, err := normalizeGymEntrenador(payload)
	if err != nil {
		return err
	}
	if item.ID <= 0 {
		return fmt.Errorf("id es obligatorio")
	}
	res, err := dbConn.Exec(`UPDATE empresa_gimnasio_entrenadores SET
		nombre_completo=?, especialidad=?, telefono=?, email=?, certificaciones=?, estado=?, disponibilidad=?, observaciones=?,
		fecha_actualizacion=datetime('now','localtime')
	WHERE id=? AND empresa_id=?`,
		item.NombreCompleto, item.Especialidad, item.Telefono, item.Email, item.Certificaciones, item.Estado, item.Disponibilidad, item.Observaciones,
		item.ID, item.EmpresaID,
	)
	if err != nil {
		return err
	}
	return ensureRowsAffected(res)
}

func DeleteEmpresaGimnasioEntrenador(dbConn *sql.DB, empresaID, id int64) error {
	return simpleDeleteByEmpresa(dbConn, empresaID, id, "empresa_gimnasio_entrenadores")
}

func ListEmpresaGimnasioClases(dbConn *sql.DB, empresaID int64) ([]EmpresaGimnasioClase, error) {
	if empresaID <= 0 {
		return nil, fmt.Errorf("empresa_id es obligatorio")
	}
	if err := EnsureEmpresaGimnasioSchema(dbConn); err != nil {
		return nil, err
	}
	rows, err := dbConn.Query(`SELECT
		c.id, c.empresa_id, COALESCE(c.nombre,''), COALESCE(c.categoria,''), COALESCE(c.entrenador_id,0), COALESCE(e.nombre_completo,''),
		COALESCE(c.sede,''), COALESCE(c.canal,''), COALESCE(c.cupos,0), COALESCE(c.duracion_minutos,60), COALESCE(c.fecha_programada,''),
		COALESCE(c.estado,'programada'), COALESCE(c.precio,0), COALESCE(c.descripcion,''), COALESCE(c.fecha_creacion,''),
		COALESCE(c.fecha_actualizacion,''), COALESCE(c.usuario_creador,'')
	FROM empresa_gimnasio_clases c
	LEFT JOIN empresa_gimnasio_entrenadores e ON e.id = c.entrenador_id AND e.empresa_id = c.empresa_id
	WHERE c.empresa_id = ?
	ORDER BY c.fecha_programada DESC, c.id DESC`, empresaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []EmpresaGimnasioClase
	for rows.Next() {
		var item EmpresaGimnasioClase
		if err := rows.Scan(&item.ID, &item.EmpresaID, &item.Nombre, &item.Categoria, &item.EntrenadorID, &item.EntrenadorNombre,
			&item.Sede, &item.Canal, &item.Cupos, &item.DuracionMinutos, &item.FechaProgramada,
			&item.Estado, &item.Precio, &item.Descripcion, &item.FechaCreacion,
			&item.FechaActualizacion, &item.UsuarioCreador); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func CreateEmpresaGimnasioClase(dbConn *sql.DB, payload EmpresaGimnasioClase) (int64, error) {
	if err := EnsureEmpresaGimnasioSchema(dbConn); err != nil {
		return 0, err
	}
	item, err := normalizeGymClase(payload)
	if err != nil {
		return 0, err
	}
	res, err := dbConn.Exec(`INSERT INTO empresa_gimnasio_clases (
		empresa_id, nombre, categoria, entrenador_id, sede, canal, cupos, duracion_minutos, fecha_programada, estado, precio, descripcion,
		usuario_creador, fecha_actualizacion
	) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,datetime('now','localtime'))`,
		item.EmpresaID, item.Nombre, item.Categoria, item.EntrenadorID, item.Sede, item.Canal, item.Cupos, item.DuracionMinutos, item.FechaProgramada, item.Estado, item.Precio, item.Descripcion,
		item.UsuarioCreador,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func UpdateEmpresaGimnasioClase(dbConn *sql.DB, payload EmpresaGimnasioClase) error {
	if err := EnsureEmpresaGimnasioSchema(dbConn); err != nil {
		return err
	}
	item, err := normalizeGymClase(payload)
	if err != nil {
		return err
	}
	if item.ID <= 0 {
		return fmt.Errorf("id es obligatorio")
	}
	res, err := dbConn.Exec(`UPDATE empresa_gimnasio_clases SET
		nombre=?, categoria=?, entrenador_id=?, sede=?, canal=?, cupos=?, duracion_minutos=?, fecha_programada=?, estado=?, precio=?, descripcion=?,
		fecha_actualizacion=datetime('now','localtime')
	WHERE id=? AND empresa_id=?`,
		item.Nombre, item.Categoria, item.EntrenadorID, item.Sede, item.Canal, item.Cupos, item.DuracionMinutos, item.FechaProgramada, item.Estado, item.Precio, item.Descripcion,
		item.ID, item.EmpresaID,
	)
	if err != nil {
		return err
	}
	return ensureRowsAffected(res)
}

func DeleteEmpresaGimnasioClase(dbConn *sql.DB, empresaID, id int64) error {
	return simpleDeleteByEmpresa(dbConn, empresaID, id, "empresa_gimnasio_clases")
}

func ListEmpresaGimnasioInscripciones(dbConn *sql.DB, empresaID int64) ([]EmpresaGimnasioInscripcion, error) {
	if empresaID <= 0 {
		return nil, fmt.Errorf("empresa_id es obligatorio")
	}
	if err := EnsureEmpresaGimnasioSchema(dbConn); err != nil {
		return nil, err
	}
	rows, err := dbConn.Query(`SELECT
		i.id, i.empresa_id, i.socio_id, COALESCE(s.nombre_completo,''), i.clase_id, COALESCE(c.nombre,''),
		COALESCE(i.estado,'activa'), COALESCE(i.fecha_inscripcion,''), COALESCE(i.asistencia_marcada,0), COALESCE(i.observaciones,''),
		COALESCE(i.fecha_creacion,''), COALESCE(i.usuario_creador,'')
	FROM empresa_gimnasio_inscripciones i
	INNER JOIN empresa_gimnasio_socios s ON s.id = i.socio_id AND s.empresa_id = i.empresa_id
	INNER JOIN empresa_gimnasio_clases c ON c.id = i.clase_id AND c.empresa_id = i.empresa_id
	WHERE i.empresa_id = ?
	ORDER BY i.id DESC`, empresaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []EmpresaGimnasioInscripcion
	for rows.Next() {
		var item EmpresaGimnasioInscripcion
		var asistenciaMarcada int
		if err := rows.Scan(&item.ID, &item.EmpresaID, &item.SocioID, &item.SocioNombre, &item.ClaseID, &item.ClaseNombre,
			&item.Estado, &item.FechaInscripcion, &asistenciaMarcada, &item.Observaciones, &item.FechaCreacion, &item.UsuarioCreador); err != nil {
			return nil, err
		}
		item.AsistenciaMarcada = asistenciaMarcada > 0
		out = append(out, item)
	}
	return out, rows.Err()
}

func CreateEmpresaGimnasioInscripcion(dbConn *sql.DB, payload EmpresaGimnasioInscripcion) (int64, error) {
	if err := EnsureEmpresaGimnasioSchema(dbConn); err != nil {
		return 0, err
	}
	item, err := normalizeGymInscripcion(payload)
	if err != nil {
		return 0, err
	}
	res, err := dbConn.Exec(`INSERT INTO empresa_gimnasio_inscripciones (
		empresa_id, socio_id, clase_id, estado, asistencia_marcada, observaciones, usuario_creador
	) VALUES (?,?,?,?,?,?,?)`,
		item.EmpresaID, item.SocioID, item.ClaseID, item.Estado, 0, item.Observaciones, item.UsuarioCreador,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func UpdateEmpresaGimnasioInscripcionEstado(dbConn *sql.DB, empresaID, id int64, estado string) error {
	if err := EnsureEmpresaGimnasioSchema(dbConn); err != nil {
		return err
	}
	res, err := dbConn.Exec(`UPDATE empresa_gimnasio_inscripciones SET estado=? WHERE id=? AND empresa_id=?`, normalizeGymState(estado, "cancelada"), id, empresaID)
	if err != nil {
		return err
	}
	return ensureRowsAffected(res)
}

func ListEmpresaGimnasioAsistencias(dbConn *sql.DB, empresaID int64) ([]EmpresaGimnasioAsistencia, error) {
	if empresaID <= 0 {
		return nil, fmt.Errorf("empresa_id es obligatorio")
	}
	if err := EnsureEmpresaGimnasioSchema(dbConn); err != nil {
		return nil, err
	}
	rows, err := dbConn.Query(`SELECT
		a.id, a.empresa_id, a.socio_id, COALESCE(s.nombre_completo,''), COALESCE(a.clase_id,0), COALESCE(c.nombre,''),
		COALESCE(a.fecha_hora,''), COALESCE(a.tipo_acceso,'checkin'), COALESCE(a.canal,''), COALESCE(a.sede,''),
		COALESCE(a.observaciones,''), COALESCE(a.fecha_creacion,''), COALESCE(a.usuario_creador,'')
	FROM empresa_gimnasio_asistencias a
	INNER JOIN empresa_gimnasio_socios s ON s.id = a.socio_id AND s.empresa_id = a.empresa_id
	LEFT JOIN empresa_gimnasio_clases c ON c.id = a.clase_id AND c.empresa_id = a.empresa_id
	WHERE a.empresa_id = ?
	ORDER BY a.fecha_hora DESC, a.id DESC`, empresaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []EmpresaGimnasioAsistencia
	for rows.Next() {
		var item EmpresaGimnasioAsistencia
		if err := rows.Scan(&item.ID, &item.EmpresaID, &item.SocioID, &item.SocioNombre, &item.ClaseID, &item.ClaseNombre,
			&item.FechaHora, &item.TipoAcceso, &item.Canal, &item.Sede, &item.Observaciones, &item.FechaCreacion, &item.UsuarioCreador); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func CreateEmpresaGimnasioAsistencia(dbConn *sql.DB, payload EmpresaGimnasioAsistencia) (int64, error) {
	if err := EnsureEmpresaGimnasioSchema(dbConn); err != nil {
		return 0, err
	}
	item, err := normalizeGymAsistencia(payload)
	if err != nil {
		return 0, err
	}
	res, err := dbConn.Exec(`INSERT INTO empresa_gimnasio_asistencias (
		empresa_id, socio_id, clase_id, fecha_hora, tipo_acceso, canal, sede, observaciones, usuario_creador
	) VALUES (?,?,?,?,?,?,?,?,?)`,
		item.EmpresaID, item.SocioID, item.ClaseID, item.FechaHora, item.TipoAcceso, item.Canal, item.Sede, item.Observaciones, item.UsuarioCreador,
	)
	if err != nil {
		return 0, err
	}
	if item.ClaseID > 0 {
		_, _ = dbConn.Exec(`UPDATE empresa_gimnasio_inscripciones SET asistencia_marcada=1 WHERE empresa_id=? AND socio_id=? AND clase_id=?`, item.EmpresaID, item.SocioID, item.ClaseID)
	}
	return res.LastInsertId()
}

func ListEmpresaGimnasioPagos(dbConn *sql.DB, empresaID int64) ([]EmpresaGimnasioPago, error) {
	if empresaID <= 0 {
		return nil, fmt.Errorf("empresa_id es obligatorio")
	}
	if err := EnsureEmpresaGimnasioSchema(dbConn); err != nil {
		return nil, err
	}
	rows, err := dbConn.Query(`SELECT
		p.id, p.empresa_id, p.socio_id, COALESCE(s.nombre_completo,''), COALESCE(p.plan_id,0), COALESCE(pl.nombre,''),
		COALESCE(p.concepto,''), COALESCE(p.monto,0), COALESCE(p.moneda,'COP'), COALESCE(p.metodo_pago,'efectivo'), COALESCE(p.canal,''),
		COALESCE(p.sede,''), COALESCE(p.estado,'pagado'), COALESCE(p.referencia,''), COALESCE(p.fecha_pago,''), COALESCE(p.observaciones,''),
		COALESCE(p.fecha_creacion,''), COALESCE(p.usuario_creador,'')
	FROM empresa_gimnasio_pagos p
	INNER JOIN empresa_gimnasio_socios s ON s.id = p.socio_id AND s.empresa_id = p.empresa_id
	LEFT JOIN empresa_gimnasio_planes pl ON pl.id = p.plan_id AND pl.empresa_id = p.empresa_id
	WHERE p.empresa_id = ?
	ORDER BY p.fecha_pago DESC, p.id DESC`, empresaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []EmpresaGimnasioPago
	for rows.Next() {
		var item EmpresaGimnasioPago
		if err := rows.Scan(&item.ID, &item.EmpresaID, &item.SocioID, &item.SocioNombre, &item.PlanID, &item.PlanNombre,
			&item.Concepto, &item.Monto, &item.Moneda, &item.MetodoPago, &item.Canal, &item.Sede, &item.Estado,
			&item.Referencia, &item.FechaPago, &item.Observaciones, &item.FechaCreacion, &item.UsuarioCreador); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func CreateEmpresaGimnasioPago(dbConn *sql.DB, payload EmpresaGimnasioPago) (int64, error) {
	if err := EnsureEmpresaGimnasioSchema(dbConn); err != nil {
		return 0, err
	}
	item, err := normalizeGymPago(payload)
	if err != nil {
		return 0, err
	}
	res, err := dbConn.Exec(`INSERT INTO empresa_gimnasio_pagos (
		empresa_id, socio_id, plan_id, concepto, monto, moneda, metodo_pago, canal, sede, estado, referencia, fecha_pago, observaciones, usuario_creador
	) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		item.EmpresaID, item.SocioID, item.PlanID, item.Concepto, item.Monto, item.Moneda, item.MetodoPago, item.Canal, item.Sede, item.Estado, item.Referencia, item.FechaPago, item.Observaciones, item.UsuarioCreador,
	)
	if err != nil {
		return 0, err
	}
	_, _ = dbConn.Exec(`UPDATE empresa_gimnasio_socios SET saldo=COALESCE(saldo,0)-?, fecha_actualizacion=datetime('now','localtime') WHERE empresa_id=? AND id=?`, item.Monto, item.EmpresaID, item.SocioID)
	return res.LastInsertId()
}

func GetEmpresaGimnasioDashboard(dbConn *sql.DB, empresaID int64) (*EmpresaGimnasioDashboard, error) {
	if empresaID <= 0 {
		return nil, fmt.Errorf("empresa_id es obligatorio")
	}
	if err := EnsureEmpresaGimnasioSchema(dbConn); err != nil {
		return nil, err
	}
	out := &EmpresaGimnasioDashboard{EmpresaID: empresaID}
	err := dbConn.QueryRow(`SELECT
		(SELECT COUNT(1) FROM empresa_gimnasio_socios WHERE empresa_id=? AND COALESCE(estado,'activo')='activo'),
		(SELECT COUNT(1) FROM empresa_gimnasio_planes WHERE empresa_id=? AND COALESCE(estado,'activo')='activo'),
		(SELECT COUNT(1) FROM empresa_gimnasio_clases WHERE empresa_id=? AND substr(COALESCE(fecha_programada,''),1,10)=date('now','localtime')),
		(SELECT COUNT(1) FROM empresa_gimnasio_asistencias WHERE empresa_id=? AND substr(COALESCE(fecha_hora,''),1,10)=date('now','localtime')),
		(SELECT COALESCE(SUM(monto),0) FROM empresa_gimnasio_pagos WHERE empresa_id=? AND substr(COALESCE(fecha_pago,''),1,7)=substr(date('now','localtime'),1,7) AND COALESCE(estado,'pagado')='pagado'),
		(SELECT COUNT(1) FROM empresa_gimnasio_socios WHERE empresa_id=? AND COALESCE(fecha_fin_plan,'')<>'' AND date(fecha_fin_plan) BETWEEN date('now','localtime') AND date('now','localtime','+10 day')),
		(SELECT COUNT(1) FROM empresa_gimnasio_inscripciones WHERE empresa_id=? AND COALESCE(estado,'activa')='activa')`,
		empresaID, empresaID, empresaID, empresaID, empresaID, empresaID, empresaID,
	).Scan(&out.SociosActivos, &out.PlanesActivos, &out.ClasesHoy, &out.AccesosHoy, &out.IngresosMes, &out.RenovacionesProximas, &out.InscripcionesActivas)
	if err != nil {
		return nil, err
	}

	vencimientos, err := ListEmpresaGimnasioSocios(dbConn, empresaID)
	if err == nil {
		for _, item := range vencimientos {
			if strings.TrimSpace(item.FechaFinPlan) != "" && len(out.VencimientosProximos) < 8 {
				out.VencimientosProximos = append(out.VencimientosProximos, item)
			}
		}
	}

	clases, err := ListEmpresaGimnasioClases(dbConn, empresaID)
	if err == nil {
		today := time.Now().Format("2006-01-02")
		for _, item := range clases {
			if strings.HasPrefix(strings.TrimSpace(item.FechaProgramada), today) && len(out.ClasesProgramadasHoy) < 8 {
				out.ClasesProgramadasHoy = append(out.ClasesProgramadasHoy, item)
			}
		}
	}

	out.IngresosPorCanal, _ = listGymResumenGrupo(dbConn, `SELECT COALESCE(canal,'Sin canal'), COUNT(1), COALESCE(SUM(monto),0), COALESCE(SUM(monto),0), 0 FROM empresa_gimnasio_pagos WHERE empresa_id=? AND COALESCE(estado,'pagado')='pagado' GROUP BY COALESCE(canal,'Sin canal') ORDER BY COALESCE(SUM(monto),0) DESC`, empresaID)
	out.RentabilidadPorLinea, _ = listGymResumenGrupo(dbConn, `SELECT COALESCE(concepto,'Sin concepto'), COUNT(1), COALESCE(SUM(monto),0), COALESCE(SUM(monto),0) - (COUNT(1) * 12000), 0 FROM empresa_gimnasio_pagos WHERE empresa_id=? AND COALESCE(estado,'pagado')='pagado' GROUP BY COALESCE(concepto,'Sin concepto') ORDER BY COALESCE(SUM(monto),0) DESC`, empresaID)
	out.RentabilidadPorSede, _ = listGymResumenGrupo(dbConn, `SELECT COALESCE(sede,'Principal'), COUNT(1), COALESCE(SUM(monto),0), COALESCE(SUM(monto),0) - (COUNT(1) * 9000), 0 FROM empresa_gimnasio_pagos WHERE empresa_id=? AND COALESCE(estado,'pagado')='pagado' GROUP BY COALESCE(sede,'Principal') ORDER BY COALESCE(SUM(monto),0) DESC`, empresaID)

	return out, nil
}

func listGymResumenGrupo(dbConn *sql.DB, query string, empresaID int64) ([]EmpresaGimnasioResumenGrupo, error) {
	rows, err := dbConn.Query(query, empresaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []EmpresaGimnasioResumenGrupo
	for rows.Next() {
		var item EmpresaGimnasioResumenGrupo
		if err := rows.Scan(&item.Clave, &item.Cantidad, &item.Monto, &item.Margen, &item.Ocupacion); err != nil {
			return nil, err
		}
		item.Etiqueta = item.Clave
		out = append(out, item)
	}
	return out, rows.Err()
}

func simpleDeleteByEmpresa(dbConn *sql.DB, empresaID, id int64, table string) error {
	if empresaID <= 0 || id <= 0 {
		return fmt.Errorf("empresa_id e id son obligatorios")
	}
	if err := EnsureEmpresaGimnasioSchema(dbConn); err != nil {
		return err
	}
	res, err := dbConn.Exec(`DELETE FROM `+table+` WHERE id=? AND empresa_id=?`, id, empresaID)
	if err != nil {
		return err
	}
	return ensureRowsAffected(res)
}

func ensureRowsAffected(res sql.Result) error {
	rows, err := res.RowsAffected()
	if err != nil {
		return nil
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}
