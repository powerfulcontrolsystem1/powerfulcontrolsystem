package db

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"
)

type EmpresaGimnasioSocio struct {
	ID                         int64   `json:"id"`
	EmpresaID                  int64   `json:"empresa_id"`
	ClienteID                  int64   `json:"cliente_id,omitempty"`
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
	ServicioID             int64   `json:"servicio_id,omitempty"`
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
	ClienteID      int64   `json:"cliente_id,omitempty"`
	PlanID         int64   `json:"plan_id,omitempty"`
	PlanNombre     string  `json:"plan_nombre,omitempty"`
	ServicioID     int64   `json:"servicio_id,omitempty"`
	CarritoID      int64   `json:"carrito_id,omitempty"`
	CarritoItemID  int64   `json:"carrito_item_id,omitempty"`
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

type EmpresaGimnasioAccesoConfig struct {
	ID                      int64  `json:"id"`
	EmpresaID               int64  `json:"empresa_id"`
	ModoValidacionPrincipal string `json:"modo_validacion_principal"`
	PermitirRFID            bool   `json:"permitir_rfid"`
	PermitirNFC             bool   `json:"permitir_nfc"`
	PermitirQR              bool   `json:"permitir_qr"`
	PermitirPIN             bool   `json:"permitir_pin"`
	PermitirBiometria       bool   `json:"permitir_biometria"`
	PermitirFacial          bool   `json:"permitir_facial"`
	AntiPassbackMinutos     int    `json:"anti_passback_minutos"`
	MinutosToleranciaMora   int    `json:"minutos_tolerancia_mora"`
	Estado                  string `json:"estado,omitempty"`
	FechaActualizacion      string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador          string `json:"usuario_creador,omitempty"`
}

type EmpresaGimnasioCredencial struct {
	ID                 int64  `json:"id"`
	EmpresaID          int64  `json:"empresa_id"`
	SocioID            int64  `json:"socio_id"`
	SocioNombre        string `json:"socio_nombre,omitempty"`
	TipoCredencial     string `json:"tipo_credencial"`
	CodigoCredencial   string `json:"codigo_credencial"`
	AliasCredencial    string `json:"alias_credencial,omitempty"`
	Estado             string `json:"estado,omitempty"`
	FechaExpiracion    string `json:"fecha_expiracion,omitempty"`
	UltimoUso          string `json:"ultimo_uso,omitempty"`
	FechaCreacion      string `json:"fecha_creacion,omitempty"`
	FechaActualizacion string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string `json:"usuario_creador,omitempty"`
}

type EmpresaGimnasioDispositivoAcceso struct {
	ID                 int64  `json:"id"`
	EmpresaID          int64  `json:"empresa_id"`
	Nombre             string `json:"nombre"`
	TipoDispositivo    string `json:"tipo_dispositivo"`
	Ubicacion          string `json:"ubicacion,omitempty"`
	Sede               string `json:"sede,omitempty"`
	Canal              string `json:"canal,omitempty"`
	Estado             string `json:"estado,omitempty"`
	Identificador      string `json:"identificador,omitempty"`
	Observaciones      string `json:"observaciones,omitempty"`
	FechaCreacion      string `json:"fecha_creacion,omitempty"`
	FechaActualizacion string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string `json:"usuario_creador,omitempty"`
}

type EmpresaGimnasioEventoAcceso struct {
	ID                int64  `json:"id"`
	EmpresaID         int64  `json:"empresa_id"`
	SocioID           int64  `json:"socio_id,omitempty"`
	SocioNombre       string `json:"socio_nombre,omitempty"`
	CredencialID      int64  `json:"credencial_id,omitempty"`
	CodigoCredencial  string `json:"codigo_credencial,omitempty"`
	DispositivoID     int64  `json:"dispositivo_id,omitempty"`
	DispositivoNombre string `json:"dispositivo_nombre,omitempty"`
	MetodoAcceso      string `json:"metodo_acceso,omitempty"`
	Resultado         string `json:"resultado,omitempty"`
	Motivo            string `json:"motivo,omitempty"`
	FechaEvento       string `json:"fecha_evento,omitempty"`
	Canal             string `json:"canal,omitempty"`
	Sede              string `json:"sede,omitempty"`
	Observaciones     string `json:"observaciones,omitempty"`
	UsuarioCreador    string `json:"usuario_creador,omitempty"`
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

type EmpresaGimnasioPreconfiguracion struct {
	EmpresaID             int64  `json:"empresa_id"`
	NombreSedePrincipal   string `json:"nombre_sede_principal"`
	PermitirRFID          bool   `json:"permitir_rfid"`
	PermitirNFC           bool   `json:"permitir_nfc"`
	PermitirQR            bool   `json:"permitir_qr"`
	CrearPlanesBase       bool   `json:"crear_planes_base"`
	CrearClasesBase       bool   `json:"crear_clases_base"`
	CrearDispositivosBase bool   `json:"crear_dispositivos_base"`
	CuposBaseClase        int    `json:"cupos_base_clase"`
	DuracionBaseClase     int    `json:"duracion_base_clase"`
	UsuarioCreador        string `json:"usuario_creador,omitempty"`
}

type EmpresaGimnasioPreconfiguracionResumen struct {
	EmpresaID         int64  `json:"empresa_id"`
	SedePrincipal     string `json:"sede_principal"`
	Socios            int    `json:"socios"`
	Planes            int    `json:"planes"`
	Entrenadores      int    `json:"entrenadores"`
	Clases            int    `json:"clases"`
	Credenciales      int    `json:"credenciales"`
	Dispositivos      int    `json:"dispositivos"`
	AccesoConfigurado bool   `json:"acceso_configurado"`
	TieneDatos        bool   `json:"tiene_datos"`
}

type EmpresaGimnasioIntegracionNucleoResumen struct {
	EmpresaID             int64    `json:"empresa_id"`
	SociosSincronizados   int      `json:"socios_sincronizados"`
	PlanesSincronizados   int      `json:"planes_sincronizados"`
	PagosSincronizados    int      `json:"pagos_sincronizados"`
	PagosPendientes       int      `json:"pagos_pendientes"`
	Errores               []string `json:"errores,omitempty"`
	EstadoIntegracion     string   `json:"estado_integracion"`
	VisibleOperativo      bool     `json:"visible_operativo"`
	RequiereRevisionDatos bool     `json:"requiere_revision_datos"`
}

var (
	empresaGimnasioSchemaEnsured sync.Map
	empresaGimnasioSchemaMu      sync.Mutex
)

func EnsureEmpresaGimnasioSchema(dbConn *sql.DB) error {
	if dbConn == nil {
		return fmt.Errorf("dbConn es obligatorio")
	}
	cacheKey := fmt.Sprintf("%p", dbConn)
	if _, ok := empresaGimnasioSchemaEnsured.Load(cacheKey); ok {
		return nil
	}
	empresaGimnasioSchemaMu.Lock()
	defer empresaGimnasioSchemaMu.Unlock()
	if _, ok := empresaGimnasioSchemaEnsured.Load(cacheKey); ok {
		return nil
	}

	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_gimnasio_planes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			servicio_id INTEGER,
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
		`CREATE INDEX IF NOT EXISTS ix_empresa_gimnasio_planes_servicio ON empresa_gimnasio_planes(empresa_id, servicio_id);`,
		`CREATE TABLE IF NOT EXISTS empresa_gimnasio_socios (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			cliente_id INTEGER,
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
		`CREATE INDEX IF NOT EXISTS ix_empresa_gimnasio_socios_cliente ON empresa_gimnasio_socios(empresa_id, cliente_id);`,
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
			cliente_id INTEGER,
			plan_id INTEGER,
			servicio_id INTEGER,
			carrito_id INTEGER,
			carrito_item_id INTEGER,
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
		`CREATE INDEX IF NOT EXISTS ix_empresa_gimnasio_pagos_carrito ON empresa_gimnasio_pagos(empresa_id, carrito_id);`,
		`CREATE TABLE IF NOT EXISTS empresa_gimnasio_acceso_config (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL UNIQUE,
			modo_validacion_principal TEXT DEFAULT 'rfid',
			permitir_rfid INTEGER DEFAULT 1,
			permitir_nfc INTEGER DEFAULT 1,
			permitir_qr INTEGER DEFAULT 1,
			permitir_pin INTEGER DEFAULT 0,
			permitir_biometria INTEGER DEFAULT 0,
			permitir_facial INTEGER DEFAULT 0,
			anti_passback_minutos INTEGER DEFAULT 10,
			minutos_tolerancia_mora INTEGER DEFAULT 0,
			estado TEXT DEFAULT 'activo',
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS empresa_gimnasio_credenciales (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			socio_id INTEGER NOT NULL,
			tipo_credencial TEXT NOT NULL,
			codigo_credencial TEXT NOT NULL,
			alias_credencial TEXT,
			estado TEXT DEFAULT 'activa',
			fecha_expiracion TEXT,
			ultimo_uso TEXT,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT
		);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_gimnasio_credenciales_empresa ON empresa_gimnasio_credenciales(empresa_id, socio_id, estado, id DESC);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_empresa_gimnasio_credenciales_codigo ON empresa_gimnasio_credenciales(empresa_id, codigo_credencial);`,
		`CREATE TABLE IF NOT EXISTS empresa_gimnasio_dispositivos_acceso (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			nombre TEXT NOT NULL,
			tipo_dispositivo TEXT NOT NULL,
			ubicacion TEXT,
			sede TEXT,
			canal TEXT,
			estado TEXT DEFAULT 'activo',
			identificador TEXT,
			observaciones TEXT,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT
		);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_gimnasio_dispositivos_empresa ON empresa_gimnasio_dispositivos_acceso(empresa_id, estado, id DESC);`,
		`CREATE TABLE IF NOT EXISTS empresa_gimnasio_eventos_acceso (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			socio_id INTEGER,
			credencial_id INTEGER,
			dispositivo_id INTEGER,
			codigo_credencial TEXT,
			metodo_acceso TEXT,
			resultado TEXT DEFAULT 'aprobado',
			motivo TEXT,
			fecha_evento TEXT DEFAULT (datetime('now','localtime')),
			canal TEXT,
			sede TEXT,
			observaciones TEXT,
			usuario_creador TEXT
		);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_gimnasio_eventos_empresa_fecha ON empresa_gimnasio_eventos_acceso(empresa_id, fecha_evento DESC, id DESC);`,
	}
	for _, stmt := range stmts {
		if _, err := dbConn.Exec(stmt); err != nil {
			return err
		}
	}
	columnGroups := []struct {
		table   string
		columns []struct {
			name string
			def  string
		}
	}{
		{
			table: "empresa_gimnasio_planes",
			columns: []struct {
				name string
				def  string
			}{
				{"empresa_id", "INTEGER NOT NULL"},
				{"servicio_id", "INTEGER"},
				{"nombre", "TEXT NOT NULL"},
				{"descripcion", "TEXT"},
				{"precio", "REAL DEFAULT 0"},
				{"duracion_dias", "INTEGER DEFAULT 30"},
				{"clases_incluidas", "INTEGER DEFAULT 0"},
				{"acceso_ilimitado", "INTEGER DEFAULT 0"},
				{"sesiones_personalizadas", "INTEGER DEFAULT 0"},
				{"estado", "TEXT DEFAULT 'activo'"},
				{"fecha_creacion", "TEXT DEFAULT (datetime('now','localtime'))"},
				{"fecha_actualizacion", "TEXT DEFAULT (datetime('now','localtime'))"},
				{"usuario_creador", "TEXT"},
			},
		},
		{
			table: "empresa_gimnasio_socios",
			columns: []struct {
				name string
				def  string
			}{
				{"empresa_id", "INTEGER NOT NULL"},
				{"cliente_id", "INTEGER"},
				{"codigo", "TEXT"},
				{"nombre_completo", "TEXT NOT NULL"},
				{"documento", "TEXT"},
				{"telefono", "TEXT"},
				{"email", "TEXT"},
				{"fecha_nacimiento", "TEXT"},
				{"genero", "TEXT"},
				{"objetivo", "TEXT"},
				{"estado", "TEXT DEFAULT 'activo'"},
				{"plan_id", "INTEGER"},
				{"fecha_inicio_plan", "TEXT"},
				{"fecha_fin_plan", "TEXT"},
				{"saldo", "REAL DEFAULT 0"},
				{"contacto_emergencia_nombre", "TEXT"},
				{"contacto_emergencia_telefono", "TEXT"},
				{"observaciones", "TEXT"},
				{"fecha_creacion", "TEXT DEFAULT (datetime('now','localtime'))"},
				{"fecha_actualizacion", "TEXT DEFAULT (datetime('now','localtime'))"},
				{"usuario_creador", "TEXT"},
			},
		},
		{
			table: "empresa_gimnasio_entrenadores",
			columns: []struct {
				name string
				def  string
			}{
				{"empresa_id", "INTEGER NOT NULL"},
				{"nombre_completo", "TEXT NOT NULL"},
				{"especialidad", "TEXT"},
				{"telefono", "TEXT"},
				{"email", "TEXT"},
				{"certificaciones", "TEXT"},
				{"estado", "TEXT DEFAULT 'activo'"},
				{"disponibilidad", "TEXT"},
				{"observaciones", "TEXT"},
				{"fecha_creacion", "TEXT DEFAULT (datetime('now','localtime'))"},
				{"fecha_actualizacion", "TEXT DEFAULT (datetime('now','localtime'))"},
				{"usuario_creador", "TEXT"},
			},
		},
		{
			table: "empresa_gimnasio_clases",
			columns: []struct {
				name string
				def  string
			}{
				{"empresa_id", "INTEGER NOT NULL"},
				{"nombre", "TEXT NOT NULL"},
				{"categoria", "TEXT"},
				{"entrenador_id", "INTEGER"},
				{"sede", "TEXT"},
				{"canal", "TEXT"},
				{"cupos", "INTEGER DEFAULT 0"},
				{"duracion_minutos", "INTEGER DEFAULT 60"},
				{"fecha_programada", "TEXT"},
				{"estado", "TEXT DEFAULT 'programada'"},
				{"precio", "REAL DEFAULT 0"},
				{"descripcion", "TEXT"},
				{"fecha_creacion", "TEXT DEFAULT (datetime('now','localtime'))"},
				{"fecha_actualizacion", "TEXT DEFAULT (datetime('now','localtime'))"},
				{"usuario_creador", "TEXT"},
			},
		},
		{
			table: "empresa_gimnasio_inscripciones",
			columns: []struct {
				name string
				def  string
			}{
				{"empresa_id", "INTEGER NOT NULL"},
				{"socio_id", "INTEGER NOT NULL"},
				{"clase_id", "INTEGER NOT NULL"},
				{"estado", "TEXT DEFAULT 'activa'"},
				{"fecha_inscripcion", "TEXT DEFAULT (datetime('now','localtime'))"},
				{"asistencia_marcada", "INTEGER DEFAULT 0"},
				{"observaciones", "TEXT"},
				{"fecha_creacion", "TEXT DEFAULT (datetime('now','localtime'))"},
				{"usuario_creador", "TEXT"},
			},
		},
		{
			table: "empresa_gimnasio_asistencias",
			columns: []struct {
				name string
				def  string
			}{
				{"empresa_id", "INTEGER NOT NULL"},
				{"socio_id", "INTEGER NOT NULL"},
				{"clase_id", "INTEGER"},
				{"fecha_hora", "TEXT DEFAULT (datetime('now','localtime'))"},
				{"tipo_acceso", "TEXT DEFAULT 'checkin'"},
				{"canal", "TEXT"},
				{"sede", "TEXT"},
				{"observaciones", "TEXT"},
				{"fecha_creacion", "TEXT DEFAULT (datetime('now','localtime'))"},
				{"usuario_creador", "TEXT"},
			},
		},
		{
			table: "empresa_gimnasio_pagos",
			columns: []struct {
				name string
				def  string
			}{
				{"empresa_id", "INTEGER NOT NULL"},
				{"socio_id", "INTEGER NOT NULL"},
				{"cliente_id", "INTEGER"},
				{"plan_id", "INTEGER"},
				{"servicio_id", "INTEGER"},
				{"carrito_id", "INTEGER"},
				{"carrito_item_id", "INTEGER"},
				{"concepto", "TEXT NOT NULL"},
				{"monto", "REAL DEFAULT 0"},
				{"moneda", "TEXT DEFAULT 'COP'"},
				{"metodo_pago", "TEXT DEFAULT 'efectivo'"},
				{"canal", "TEXT"},
				{"sede", "TEXT"},
				{"estado", "TEXT DEFAULT 'pagado'"},
				{"referencia", "TEXT"},
				{"fecha_pago", "TEXT DEFAULT (datetime('now','localtime'))"},
				{"observaciones", "TEXT"},
				{"fecha_creacion", "TEXT DEFAULT (datetime('now','localtime'))"},
				{"usuario_creador", "TEXT"},
			},
		},
		{
			table: "empresa_gimnasio_acceso_config",
			columns: []struct {
				name string
				def  string
			}{
				{"empresa_id", "INTEGER NOT NULL UNIQUE"},
				{"modo_validacion_principal", "TEXT DEFAULT 'rfid'"},
				{"permitir_rfid", "INTEGER DEFAULT 1"},
				{"permitir_nfc", "INTEGER DEFAULT 1"},
				{"permitir_qr", "INTEGER DEFAULT 1"},
				{"permitir_pin", "INTEGER DEFAULT 0"},
				{"permitir_biometria", "INTEGER DEFAULT 0"},
				{"permitir_facial", "INTEGER DEFAULT 0"},
				{"anti_passback_minutos", "INTEGER DEFAULT 10"},
				{"minutos_tolerancia_mora", "INTEGER DEFAULT 0"},
				{"estado", "TEXT DEFAULT 'activo'"},
				{"fecha_actualizacion", "TEXT DEFAULT (datetime('now','localtime'))"},
				{"usuario_creador", "TEXT"},
			},
		},
		{
			table: "empresa_gimnasio_credenciales",
			columns: []struct {
				name string
				def  string
			}{
				{"empresa_id", "INTEGER NOT NULL"},
				{"socio_id", "INTEGER NOT NULL"},
				{"tipo_credencial", "TEXT NOT NULL"},
				{"codigo_credencial", "TEXT NOT NULL"},
				{"alias_credencial", "TEXT"},
				{"estado", "TEXT DEFAULT 'activa'"},
				{"fecha_expiracion", "TEXT"},
				{"ultimo_uso", "TEXT"},
				{"fecha_creacion", "TEXT DEFAULT (datetime('now','localtime'))"},
				{"fecha_actualizacion", "TEXT DEFAULT (datetime('now','localtime'))"},
				{"usuario_creador", "TEXT"},
			},
		},
		{
			table: "empresa_gimnasio_dispositivos_acceso",
			columns: []struct {
				name string
				def  string
			}{
				{"empresa_id", "INTEGER NOT NULL"},
				{"nombre", "TEXT NOT NULL"},
				{"tipo_dispositivo", "TEXT NOT NULL"},
				{"ubicacion", "TEXT"},
				{"sede", "TEXT"},
				{"canal", "TEXT"},
				{"estado", "TEXT DEFAULT 'activo'"},
				{"identificador", "TEXT"},
				{"observaciones", "TEXT"},
				{"fecha_creacion", "TEXT DEFAULT (datetime('now','localtime'))"},
				{"fecha_actualizacion", "TEXT DEFAULT (datetime('now','localtime'))"},
				{"usuario_creador", "TEXT"},
			},
		},
		{
			table: "empresa_gimnasio_eventos_acceso",
			columns: []struct {
				name string
				def  string
			}{
				{"empresa_id", "INTEGER NOT NULL"},
				{"socio_id", "INTEGER"},
				{"credencial_id", "INTEGER"},
				{"dispositivo_id", "INTEGER"},
				{"codigo_credencial", "TEXT"},
				{"metodo_acceso", "TEXT"},
				{"resultado", "TEXT DEFAULT 'aprobado'"},
				{"motivo", "TEXT"},
				{"fecha_evento", "TEXT DEFAULT (datetime('now','localtime'))"},
				{"canal", "TEXT"},
				{"sede", "TEXT"},
				{"observaciones", "TEXT"},
				{"usuario_creador", "TEXT"},
			},
		},
	}
	for _, group := range columnGroups {
		for _, column := range group.columns {
			if err := ensureColumnIfMissing(dbConn, group.table, column.name, column.def); err != nil {
				return err
			}
		}
	}
	empresaGimnasioSchemaEnsured.Store(cacheKey, true)
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
	if coreMetodo := NormalizeMetodoPagoCarrito(out.MetodoPago); coreMetodo != "" {
		out.MetodoPago = coreMetodo
	} else {
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

func gymCoreCode(prefix string, parts ...string) string {
	var b strings.Builder
	for _, part := range parts {
		clean := strings.ToUpper(strings.TrimSpace(part))
		for _, r := range clean {
			if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
				b.WriteRune(r)
			} else if b.Len() > 0 {
				last := b.String()[b.Len()-1]
				if last != '-' {
					b.WriteRune('-')
				}
			}
		}
		if b.Len() > 0 {
			last := b.String()[b.Len()-1]
			if last != '-' {
				b.WriteRune('-')
			}
		}
	}
	code := strings.Trim(b.String(), "-")
	if code == "" {
		code = fmt.Sprintf("%d", time.Now().UnixNano())
	}
	if len(code) > 42 {
		code = code[:42]
	}
	return strings.Trim(strings.ToUpper(strings.TrimSpace(prefix)), "-") + "-" + code
}

func findEmpresaGimnasioClienteID(dbConn *sql.DB, socio EmpresaGimnasioSocio) (int64, error) {
	if socio.ClienteID > 0 {
		return socio.ClienteID, nil
	}
	documento := normalizeClienteDocumentoValue(socio.Documento)
	if documento != "" {
		query := fmt.Sprintf(`SELECT id FROM clientes WHERE empresa_id = ? AND %s = ? LIMIT 1`, clienteDocumentoSQLExpr("numero_documento"))
		return findClienteDuplicateID(dbConn, query, socio.EmpresaID, documento)
	}
	if email := normalizeClienteEmailValue(socio.Email); email != "" {
		return findClienteDuplicateID(dbConn, `SELECT id FROM clientes WHERE empresa_id = ? AND lower(trim(COALESCE(email, ''))) = ? LIMIT 1`, socio.EmpresaID, email)
	}
	if telefono := normalizeClienteTelefonoValue(socio.Telefono); telefono != "" {
		query := fmt.Sprintf(`SELECT id FROM clientes WHERE empresa_id = ? AND %s = ? LIMIT 1`, clienteTelefonoSQLExpr("telefono"))
		return findClienteDuplicateID(dbConn, query, socio.EmpresaID, telefono)
	}
	if codigo := strings.TrimSpace(socio.Codigo); codigo != "" {
		return findClienteDuplicateID(dbConn, `SELECT id FROM clientes WHERE empresa_id = ? AND tipo_documento = 'OTRO' AND numero_documento = ? LIMIT 1`, socio.EmpresaID, "GYM-"+codigo)
	}
	return 0, nil
}

func ensureEmpresaGimnasioSocioCliente(dbConn *sql.DB, socio EmpresaGimnasioSocio) (int64, error) {
	if err := EnsureEmpresaClientesSchema(dbConn); err != nil {
		return 0, err
	}
	if id, err := findEmpresaGimnasioClienteID(dbConn, socio); err != nil {
		return 0, err
	} else if id > 0 {
		return id, nil
	}
	tipoDocumento := "CC"
	numeroDocumento := strings.TrimSpace(socio.Documento)
	if numeroDocumento == "" {
		tipoDocumento = "OTRO"
		numeroDocumento = "GYM-" + strings.TrimSpace(socio.Codigo)
		if strings.TrimSpace(socio.Codigo) == "" {
			numeroDocumento = gymCoreCode("GYM-SOCIO", socio.NombreCompleto)
		}
	}
	id, err := CreateCliente(dbConn, Cliente{
		EmpresaID:         socio.EmpresaID,
		TipoDocumento:     tipoDocumento,
		NumeroDocumento:   numeroDocumento,
		TipoPersona:       "natural",
		NombreRazonSocial: socio.NombreCompleto,
		NombreComercial:   socio.NombreCompleto,
		Email:             socio.Email,
		Telefono:          socio.Telefono,
		Pais:              "CO",
		UsuarioCreador:    socio.UsuarioCreador,
		Estado:            "activo",
		Observaciones:     "Cliente creado/sincronizado desde gimnasio.",
	})
	if err != nil {
		var dup *ClienteDuplicadoError
		if strings.Contains(strings.ToLower(err.Error()), "cliente") && findErrClienteDuplicado(err, &dup) && dup.ClienteID > 0 {
			return dup.ClienteID, nil
		}
		return 0, err
	}
	return id, nil
}

func findErrClienteDuplicado(err error, out **ClienteDuplicadoError) bool {
	for err != nil {
		if dup, ok := err.(*ClienteDuplicadoError); ok {
			*out = dup
			return true
		}
		type unwrapper interface{ Unwrap() error }
		u, ok := err.(unwrapper)
		if !ok {
			break
		}
		err = u.Unwrap()
	}
	return false
}

func getEmpresaGimnasioSocioByID(dbConn *sql.DB, empresaID, socioID int64) (*EmpresaGimnasioSocio, error) {
	var socio EmpresaGimnasioSocio
	err := dbConn.QueryRow(`SELECT id, empresa_id, COALESCE(cliente_id,0), COALESCE(codigo,''), COALESCE(nombre_completo,''), COALESCE(documento,''), COALESCE(telefono,''), COALESCE(email,''), COALESCE(usuario_creador,'')
		FROM empresa_gimnasio_socios WHERE empresa_id=? AND id=? LIMIT 1`, empresaID, socioID).
		Scan(&socio.ID, &socio.EmpresaID, &socio.ClienteID, &socio.Codigo, &socio.NombreCompleto, &socio.Documento, &socio.Telefono, &socio.Email, &socio.UsuarioCreador)
	if err != nil {
		return nil, err
	}
	return &socio, nil
}

func syncEmpresaGimnasioSocioCliente(dbConn *sql.DB, socio EmpresaGimnasioSocio) (int64, error) {
	clienteID, err := ensureEmpresaGimnasioSocioCliente(dbConn, socio)
	if err != nil || clienteID <= 0 || socio.ID <= 0 {
		return clienteID, err
	}
	_, err = dbConn.Exec(`UPDATE empresa_gimnasio_socios SET cliente_id=?, fecha_actualizacion=datetime('now','localtime') WHERE empresa_id=? AND id=?`, clienteID, socio.EmpresaID, socio.ID)
	return clienteID, err
}

func syncEmpresaGimnasioPlanServicio(dbConn *sql.DB, plan EmpresaGimnasioPlan) (int64, error) {
	if err := EnsureEmpresaProductosSchema(dbConn); err != nil {
		return 0, err
	}
	code := gymCoreCode("GYM-PLAN", fmt.Sprintf("%d", plan.ID), plan.Nombre)
	if plan.ServicioID > 0 {
		if err := UpdateServicio(dbConn, Servicio{ID: plan.ServicioID, EmpresaID: plan.EmpresaID, Codigo: code, Nombre: plan.Nombre, Descripcion: plan.Descripcion, Categoria: "gimnasio", DuracionMinutos: plan.DuracionDias * 24 * 60, Precio: plan.Precio, Estado: plan.Estado, UsuarioCreador: plan.UsuarioCreador, Observaciones: "Servicio vendible sincronizado desde plan de gimnasio."}); err != nil {
			return 0, err
		}
		return plan.ServicioID, nil
	}
	var servicioID int64
	err := dbConn.QueryRow(`SELECT id FROM servicios WHERE empresa_id=? AND codigo=? LIMIT 1`, plan.EmpresaID, code).Scan(&servicioID)
	if err == sql.ErrNoRows {
		servicioID, err = CreateServicio(dbConn, Servicio{EmpresaID: plan.EmpresaID, Codigo: code, Nombre: plan.Nombre, Descripcion: plan.Descripcion, Categoria: "gimnasio", DuracionMinutos: plan.DuracionDias * 24 * 60, Precio: plan.Precio, Estado: plan.Estado, UsuarioCreador: plan.UsuarioCreador, Observaciones: "Servicio vendible sincronizado desde plan de gimnasio."})
	}
	if err != nil {
		return 0, err
	}
	if plan.ID > 0 {
		_, err = dbConn.Exec(`UPDATE empresa_gimnasio_planes SET servicio_id=?, fecha_actualizacion=datetime('now','localtime') WHERE empresa_id=? AND id=?`, servicioID, plan.EmpresaID, plan.ID)
	}
	return servicioID, err
}

func ensureEmpresaGimnasioConceptoServicio(dbConn *sql.DB, pago EmpresaGimnasioPago) (int64, error) {
	if pago.ServicioID > 0 {
		return pago.ServicioID, nil
	}
	if pago.PlanID > 0 {
		var plan EmpresaGimnasioPlan
		err := dbConn.QueryRow(`SELECT id, empresa_id, COALESCE(servicio_id,0), COALESCE(nombre,''), COALESCE(descripcion,''), COALESCE(precio,0), COALESCE(duracion_dias,30), COALESCE(estado,'activo'), COALESCE(usuario_creador,'')
			FROM empresa_gimnasio_planes WHERE empresa_id=? AND id=? LIMIT 1`, pago.EmpresaID, pago.PlanID).
			Scan(&plan.ID, &plan.EmpresaID, &plan.ServicioID, &plan.Nombre, &plan.Descripcion, &plan.Precio, &plan.DuracionDias, &plan.Estado, &plan.UsuarioCreador)
		if err != nil {
			return 0, err
		}
		if plan.UsuarioCreador == "" {
			plan.UsuarioCreador = pago.UsuarioCreador
		}
		return syncEmpresaGimnasioPlanServicio(dbConn, plan)
	}
	if err := EnsureEmpresaProductosSchema(dbConn); err != nil {
		return 0, err
	}
	code := gymCoreCode("GYM-SERV", pago.Concepto)
	var servicioID int64
	err := dbConn.QueryRow(`SELECT id FROM servicios WHERE empresa_id=? AND codigo=? LIMIT 1`, pago.EmpresaID, code).Scan(&servicioID)
	if err == sql.ErrNoRows {
		servicioID, err = CreateServicio(dbConn, Servicio{EmpresaID: pago.EmpresaID, Codigo: code, Nombre: pago.Concepto, Descripcion: "Servicio de gimnasio creado desde recaudo.", Categoria: "gimnasio", Precio: pago.Monto, Estado: "activo", UsuarioCreador: pago.UsuarioCreador, Observaciones: "Servicio vendible sincronizado desde pagos de gimnasio."})
	}
	return servicioID, err
}

func createEmpresaGimnasioPagoCarrito(dbConn *sql.DB, pago EmpresaGimnasioPago) (int64, int64, error) {
	if pago.Estado != "pagado" {
		return 0, 0, nil
	}
	if err := EnsureEmpresaCarritosSchema(dbConn); err != nil {
		return 0, 0, err
	}
	carritoID, err := CreateCarritoCompra(dbConn, CarritoCompra{
		EmpresaID:         pago.EmpresaID,
		Codigo:            gymCoreCode("GYM-PAGO", fmt.Sprintf("%d", pago.SocioID), fmt.Sprintf("%d", time.Now().UnixNano())),
		Nombre:            "Gimnasio - " + pago.Concepto,
		CanalVenta:        "gimnasio",
		ClienteID:         pago.ClienteID,
		EstadoCarrito:     "abierto",
		Moneda:            pago.Moneda,
		ReferenciaExterna: fmt.Sprintf("gimnasio:socio:%d:%s", pago.SocioID, pago.FechaPago),
		MetodoPago:        pago.MetodoPago,
		ReferenciaPago:    pago.Referencia,
		UsuarioCreador:    pago.UsuarioCreador,
		Observaciones:     "Venta central generada desde pago de gimnasio.",
	})
	if err != nil {
		return 0, 0, err
	}
	itemID, err := CreateCarritoCompraItem(dbConn, CarritoCompraItem{
		EmpresaID:          pago.EmpresaID,
		CarritoID:          carritoID,
		TipoItem:           "servicio",
		ReferenciaID:       pago.ServicioID,
		CodigoItem:         gymCoreCode("GYM-ITEM", pago.Concepto),
		Descripcion:        pago.Concepto,
		UnidadMedida:       "servicio",
		Cantidad:           1,
		PrecioUnitario:     pago.Monto,
		ImpuestoPorcentaje: 0,
		UsuarioCreador:     pago.UsuarioCreador,
		Estado:             "activo",
		Observaciones:      "Item central generado desde pago de gimnasio.",
	})
	if err != nil {
		return 0, 0, err
	}
	if err := PayCarritoStationSession(dbConn, pago.EmpresaID, carritoID, pago.MetodoPago, pago.Referencia, "", "", 0, 0, pago.Monto, 0, pago.UsuarioCreador); err != nil {
		return 0, 0, err
	}
	return carritoID, itemID, nil
}

func normalizeGymAccessConfig(payload EmpresaGimnasioAccesoConfig) *EmpresaGimnasioAccesoConfig {
	out := payload
	out.ModoValidacionPrincipal = strings.ToLower(strings.TrimSpace(out.ModoValidacionPrincipal))
	if out.ModoValidacionPrincipal == "" {
		out.ModoValidacionPrincipal = "rfid"
	}
	out.Estado = normalizeGymState(out.Estado, "activo")
	if out.AntiPassbackMinutos < 0 {
		out.AntiPassbackMinutos = 0
	}
	if out.MinutosToleranciaMora < 0 {
		out.MinutosToleranciaMora = 0
	}
	out.UsuarioCreador = strings.TrimSpace(out.UsuarioCreador)
	return &out
}

func normalizeGymCredencial(payload EmpresaGimnasioCredencial) (*EmpresaGimnasioCredencial, error) {
	if payload.EmpresaID <= 0 || payload.SocioID <= 0 {
		return nil, fmt.Errorf("empresa_id y socio_id son obligatorios")
	}
	out := payload
	out.TipoCredencial = strings.ToLower(strings.TrimSpace(out.TipoCredencial))
	out.CodigoCredencial = strings.TrimSpace(out.CodigoCredencial)
	out.AliasCredencial = strings.TrimSpace(out.AliasCredencial)
	out.Estado = normalizeGymState(out.Estado, "activa")
	out.FechaExpiracion = strings.TrimSpace(out.FechaExpiracion)
	out.UsuarioCreador = strings.TrimSpace(out.UsuarioCreador)
	if out.TipoCredencial == "" {
		out.TipoCredencial = "rfid"
	}
	if out.CodigoCredencial == "" {
		return nil, fmt.Errorf("codigo_credencial es obligatorio")
	}
	return &out, nil
}

func normalizeGymDispositivo(payload EmpresaGimnasioDispositivoAcceso) (*EmpresaGimnasioDispositivoAcceso, error) {
	if payload.EmpresaID <= 0 {
		return nil, fmt.Errorf("empresa_id es obligatorio")
	}
	out := payload
	out.Nombre = strings.TrimSpace(out.Nombre)
	out.TipoDispositivo = strings.ToLower(strings.TrimSpace(out.TipoDispositivo))
	out.Ubicacion = strings.TrimSpace(out.Ubicacion)
	out.Sede = strings.TrimSpace(out.Sede)
	out.Canal = strings.TrimSpace(out.Canal)
	out.Estado = normalizeGymState(out.Estado, "activo")
	out.Identificador = strings.TrimSpace(out.Identificador)
	out.Observaciones = strings.TrimSpace(out.Observaciones)
	out.UsuarioCreador = strings.TrimSpace(out.UsuarioCreador)
	if out.Nombre == "" {
		return nil, fmt.Errorf("nombre es obligatorio")
	}
	if out.TipoDispositivo == "" {
		out.TipoDispositivo = "lector_rfid"
	}
	if out.Sede == "" {
		out.Sede = "principal"
	}
	if out.Canal == "" {
		out.Canal = "ingreso"
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
		s.id, s.empresa_id, COALESCE(s.cliente_id,0), COALESCE(s.codigo,''), COALESCE(s.nombre_completo,''), COALESCE(s.documento,''), COALESCE(s.telefono,''),
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
			&item.ID, &item.EmpresaID, &item.ClienteID, &item.Codigo, &item.NombreCompleto, &item.Documento, &item.Telefono,
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
	clienteID, err := ensureEmpresaGimnasioSocioCliente(dbConn, *item)
	if err != nil {
		return 0, err
	}
	item.ClienteID = clienteID
	return insertSQLCompat(dbConn, `INSERT INTO empresa_gimnasio_socios (
		empresa_id, cliente_id, codigo, nombre_completo, documento, telefono, email, fecha_nacimiento, genero, objetivo, estado,
		plan_id, fecha_inicio_plan, fecha_fin_plan, saldo, contacto_emergencia_nombre, contacto_emergencia_telefono,
		observaciones, usuario_creador, fecha_actualizacion
	) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,datetime('now','localtime'))`,
		item.EmpresaID, nullableInt64(item.ClienteID), item.Codigo, item.NombreCompleto, item.Documento, item.Telefono, item.Email, item.FechaNacimiento, item.Genero, item.Objetivo, item.Estado,
		item.PlanID, item.FechaInicioPlan, item.FechaFinPlan, item.Saldo, item.ContactoEmergenciaNombre, item.ContactoEmergenciaTelefono,
		item.Observaciones, item.UsuarioCreador,
	)
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
	clienteID, err := syncEmpresaGimnasioSocioCliente(dbConn, *item)
	if err != nil {
		return err
	}
	item.ClienteID = clienteID
	res, err := dbConn.Exec(`UPDATE empresa_gimnasio_socios SET
		cliente_id=?, codigo=?, nombre_completo=?, documento=?, telefono=?, email=?, fecha_nacimiento=?, genero=?, objetivo=?, estado=?,
		plan_id=?, fecha_inicio_plan=?, fecha_fin_plan=?, saldo=?, contacto_emergencia_nombre=?, contacto_emergencia_telefono=?,
		observaciones=?, fecha_actualizacion=datetime('now','localtime')
	WHERE id=? AND empresa_id=?`,
		nullableInt64(item.ClienteID), item.Codigo, item.NombreCompleto, item.Documento, item.Telefono, item.Email, item.FechaNacimiento, item.Genero, item.Objetivo, item.Estado,
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
		id, empresa_id, COALESCE(servicio_id,0), COALESCE(nombre,''), COALESCE(descripcion,''), COALESCE(precio,0), COALESCE(duracion_dias,30),
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
		if err := rows.Scan(&item.ID, &item.EmpresaID, &item.ServicioID, &item.Nombre, &item.Descripcion, &item.Precio, &item.DuracionDias,
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
	id, err := insertSQLCompat(dbConn, `INSERT INTO empresa_gimnasio_planes (
		empresa_id, nombre, descripcion, precio, duracion_dias, clases_incluidas, acceso_ilimitado, sesiones_personalizadas,
		estado, usuario_creador, fecha_actualizacion
	) VALUES (?,?,?,?,?,?,?,?,?,?,datetime('now','localtime'))`,
		item.EmpresaID, item.Nombre, item.Descripcion, item.Precio, item.DuracionDias, item.ClasesIncluidas, accesoIlimitado, item.SesionesPersonalizadas,
		item.Estado, item.UsuarioCreador,
	)
	if err != nil {
		return 0, err
	}
	item.ID = id
	if servicioID, syncErr := syncEmpresaGimnasioPlanServicio(dbConn, *item); syncErr != nil {
		return 0, syncErr
	} else {
		item.ServicioID = servicioID
	}
	return id, nil
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
	servicioID, err := syncEmpresaGimnasioPlanServicio(dbConn, *item)
	if err != nil {
		return err
	}
	item.ServicioID = servicioID
	res, err := dbConn.Exec(`UPDATE empresa_gimnasio_planes SET
		servicio_id=?, nombre=?, descripcion=?, precio=?, duracion_dias=?, clases_incluidas=?, acceso_ilimitado=?, sesiones_personalizadas=?, estado=?,
		fecha_actualizacion=datetime('now','localtime')
	WHERE id=? AND empresa_id=?`,
		nullableInt64(item.ServicioID), item.Nombre, item.Descripcion, item.Precio, item.DuracionDias, item.ClasesIncluidas, accesoIlimitado, item.SesionesPersonalizadas, item.Estado,
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
	return insertSQLCompat(dbConn, `INSERT INTO empresa_gimnasio_entrenadores (
		empresa_id, nombre_completo, especialidad, telefono, email, certificaciones, estado, disponibilidad, observaciones,
		usuario_creador, fecha_actualizacion
	) VALUES (?,?,?,?,?,?,?,?,?,?,datetime('now','localtime'))`,
		item.EmpresaID, item.NombreCompleto, item.Especialidad, item.Telefono, item.Email, item.Certificaciones, item.Estado, item.Disponibilidad, item.Observaciones,
		item.UsuarioCreador,
	)
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
	return insertSQLCompat(dbConn, `INSERT INTO empresa_gimnasio_clases (
		empresa_id, nombre, categoria, entrenador_id, sede, canal, cupos, duracion_minutos, fecha_programada, estado, precio, descripcion,
		usuario_creador, fecha_actualizacion
	) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,datetime('now','localtime'))`,
		item.EmpresaID, item.Nombre, item.Categoria, item.EntrenadorID, item.Sede, item.Canal, item.Cupos, item.DuracionMinutos, item.FechaProgramada, item.Estado, item.Precio, item.Descripcion,
		item.UsuarioCreador,
	)
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
	return insertSQLCompat(dbConn, `INSERT INTO empresa_gimnasio_inscripciones (
		empresa_id, socio_id, clase_id, estado, asistencia_marcada, observaciones, usuario_creador
	) VALUES (?,?,?,?,?,?,?)`,
		item.EmpresaID, item.SocioID, item.ClaseID, item.Estado, 0, item.Observaciones, item.UsuarioCreador,
	)
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
	id, err := insertSQLCompat(dbConn, `INSERT INTO empresa_gimnasio_asistencias (
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
	return id, nil
}

func ListEmpresaGimnasioPagos(dbConn *sql.DB, empresaID int64) ([]EmpresaGimnasioPago, error) {
	if empresaID <= 0 {
		return nil, fmt.Errorf("empresa_id es obligatorio")
	}
	if err := EnsureEmpresaGimnasioSchema(dbConn); err != nil {
		return nil, err
	}
	rows, err := dbConn.Query(`SELECT
		p.id, p.empresa_id, p.socio_id, COALESCE(s.nombre_completo,''), COALESCE(p.cliente_id,0), COALESCE(p.plan_id,0), COALESCE(pl.nombre,''),
		COALESCE(p.servicio_id,0), COALESCE(p.carrito_id,0), COALESCE(p.carrito_item_id,0),
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
		if err := rows.Scan(&item.ID, &item.EmpresaID, &item.SocioID, &item.SocioNombre, &item.ClienteID, &item.PlanID, &item.PlanNombre,
			&item.ServicioID, &item.CarritoID, &item.CarritoItemID,
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
	socio, err := getEmpresaGimnasioSocioByID(dbConn, item.EmpresaID, item.SocioID)
	if err != nil {
		return 0, err
	}
	socio.UsuarioCreador = item.UsuarioCreador
	clienteID, err := syncEmpresaGimnasioSocioCliente(dbConn, *socio)
	if err != nil {
		return 0, err
	}
	item.ClienteID = clienteID
	servicioID, err := ensureEmpresaGimnasioConceptoServicio(dbConn, *item)
	if err != nil {
		return 0, err
	}
	item.ServicioID = servicioID
	carritoID, carritoItemID, err := createEmpresaGimnasioPagoCarrito(dbConn, *item)
	if err != nil {
		return 0, err
	}
	item.CarritoID = carritoID
	item.CarritoItemID = carritoItemID
	id, err := insertSQLCompat(dbConn, `INSERT INTO empresa_gimnasio_pagos (
		empresa_id, socio_id, cliente_id, plan_id, servicio_id, carrito_id, carrito_item_id, concepto, monto, moneda, metodo_pago, canal, sede, estado, referencia, fecha_pago, observaciones, usuario_creador
	) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		item.EmpresaID, item.SocioID, nullableInt64(item.ClienteID), item.PlanID, nullableInt64(item.ServicioID), nullableInt64(item.CarritoID), nullableInt64(item.CarritoItemID), item.Concepto, item.Monto, item.Moneda, item.MetodoPago, item.Canal, item.Sede, item.Estado, item.Referencia, item.FechaPago, item.Observaciones, item.UsuarioCreador,
	)
	if err != nil {
		return 0, err
	}
	_, _ = dbConn.Exec(`UPDATE empresa_gimnasio_socios SET saldo=COALESCE(saldo,0)-?, fecha_actualizacion=datetime('now','localtime') WHERE empresa_id=? AND id=?`, item.Monto, item.EmpresaID, item.SocioID)
	return id, nil
}

func SyncEmpresaGimnasioNucleo(dbConn *sql.DB, empresaID int64, usuario string) (*EmpresaGimnasioIntegracionNucleoResumen, error) {
	if empresaID <= 0 {
		return nil, fmt.Errorf("empresa_id es obligatorio")
	}
	if err := EnsureEmpresaGimnasioSchema(dbConn); err != nil {
		return nil, err
	}
	out := &EmpresaGimnasioIntegracionNucleoResumen{EmpresaID: empresaID, EstadoIntegracion: "plantilla_integrada_nucleo", VisibleOperativo: true}
	usuario = strings.TrimSpace(usuario)
	if usuario == "" {
		usuario = "sistema"
	}

	planes, err := ListEmpresaGimnasioPlanes(dbConn, empresaID)
	if err != nil {
		return nil, err
	}
	for _, plan := range planes {
		plan.UsuarioCreador = usuario
		if servicioID, syncErr := syncEmpresaGimnasioPlanServicio(dbConn, plan); syncErr != nil {
			out.Errores = append(out.Errores, fmt.Sprintf("plan %d: %v", plan.ID, syncErr))
		} else if servicioID > 0 {
			out.PlanesSincronizados++
		}
	}

	socios, err := ListEmpresaGimnasioSocios(dbConn, empresaID)
	if err != nil {
		return nil, err
	}
	for _, socio := range socios {
		socio.UsuarioCreador = usuario
		if clienteID, syncErr := syncEmpresaGimnasioSocioCliente(dbConn, socio); syncErr != nil {
			out.Errores = append(out.Errores, fmt.Sprintf("socio %d: %v", socio.ID, syncErr))
		} else if clienteID > 0 {
			out.SociosSincronizados++
		}
	}

	pagos, err := ListEmpresaGimnasioPagos(dbConn, empresaID)
	if err != nil {
		return nil, err
	}
	for _, pago := range pagos {
		if pago.CarritoID > 0 || pago.Estado != "pagado" {
			if pago.Estado != "pagado" {
				out.PagosPendientes++
			}
			continue
		}
		pago.UsuarioCreador = usuario
		socio, syncErr := getEmpresaGimnasioSocioByID(dbConn, pago.EmpresaID, pago.SocioID)
		if syncErr != nil {
			out.Errores = append(out.Errores, fmt.Sprintf("pago %d socio: %v", pago.ID, syncErr))
			continue
		}
		socio.UsuarioCreador = usuario
		pago.ClienteID, syncErr = syncEmpresaGimnasioSocioCliente(dbConn, *socio)
		if syncErr != nil {
			out.Errores = append(out.Errores, fmt.Sprintf("pago %d cliente: %v", pago.ID, syncErr))
			continue
		}
		pago.ServicioID, syncErr = ensureEmpresaGimnasioConceptoServicio(dbConn, pago)
		if syncErr != nil {
			out.Errores = append(out.Errores, fmt.Sprintf("pago %d servicio: %v", pago.ID, syncErr))
			continue
		}
		carritoID, carritoItemID, syncErr := createEmpresaGimnasioPagoCarrito(dbConn, pago)
		if syncErr != nil {
			out.Errores = append(out.Errores, fmt.Sprintf("pago %d venta: %v", pago.ID, syncErr))
			continue
		}
		_, syncErr = dbConn.Exec(`UPDATE empresa_gimnasio_pagos SET cliente_id=?, servicio_id=?, carrito_id=?, carrito_item_id=? WHERE empresa_id=? AND id=?`, nullableInt64(pago.ClienteID), nullableInt64(pago.ServicioID), nullableInt64(carritoID), nullableInt64(carritoItemID), pago.EmpresaID, pago.ID)
		if syncErr != nil {
			out.Errores = append(out.Errores, fmt.Sprintf("pago %d referencia: %v", pago.ID, syncErr))
			continue
		}
		out.PagosSincronizados++
	}
	out.RequiereRevisionDatos = len(out.Errores) > 0
	if out.RequiereRevisionDatos {
		out.EstadoIntegracion = "integrado_con_observaciones"
	}
	return out, nil
}

func GetEmpresaGimnasioDashboard(dbConn *sql.DB, empresaID int64) (*EmpresaGimnasioDashboard, error) {
	if empresaID <= 0 {
		return nil, fmt.Errorf("empresa_id es obligatorio")
	}
	if err := EnsureEmpresaGimnasioSchema(dbConn); err != nil {
		return nil, err
	}
	out := &EmpresaGimnasioDashboard{EmpresaID: empresaID}
	now := time.Now()
	today := now.Format("2006-01-02")
	currentMonth := now.Format("2006-01")
	renewalUntil := now.AddDate(0, 0, 10).Format("2006-01-02")
	err := dbConn.QueryRow(`SELECT
		(SELECT COUNT(1) FROM empresa_gimnasio_socios WHERE empresa_id=? AND COALESCE(estado,'activo')='activo'),
		(SELECT COUNT(1) FROM empresa_gimnasio_planes WHERE empresa_id=? AND COALESCE(estado,'activo')='activo'),
		(SELECT COUNT(1) FROM empresa_gimnasio_clases WHERE empresa_id=? AND substr(COALESCE(fecha_programada,''),1,10)=?),
		(SELECT COUNT(1) FROM empresa_gimnasio_asistencias WHERE empresa_id=? AND substr(COALESCE(fecha_hora,''),1,10)=?),
		(SELECT COALESCE(SUM(monto),0) FROM empresa_gimnasio_pagos WHERE empresa_id=? AND substr(COALESCE(fecha_pago,''),1,7)=? AND COALESCE(estado,'pagado')='pagado'),
		(SELECT COUNT(1) FROM empresa_gimnasio_socios WHERE empresa_id=? AND COALESCE(fecha_fin_plan,'')<>'' AND substr(COALESCE(fecha_fin_plan,''),1,10) BETWEEN ? AND ?),
		(SELECT COUNT(1) FROM empresa_gimnasio_inscripciones WHERE empresa_id=? AND COALESCE(estado,'activa')='activa')`,
		empresaID, empresaID, empresaID, today, empresaID, today, empresaID, currentMonth, empresaID, today, renewalUntil, empresaID,
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

func normalizeGymPreconfig(payload EmpresaGimnasioPreconfiguracion) *EmpresaGimnasioPreconfiguracion {
	out := payload
	out.NombreSedePrincipal = strings.TrimSpace(out.NombreSedePrincipal)
	if out.NombreSedePrincipal == "" {
		out.NombreSedePrincipal = "principal"
	}
	if out.CuposBaseClase <= 0 {
		out.CuposBaseClase = 18
	}
	if out.DuracionBaseClase <= 0 {
		out.DuracionBaseClase = 60
	}
	if !out.PermitirRFID && !out.PermitirNFC && !out.PermitirQR {
		out.PermitirRFID = true
		out.PermitirQR = true
	}
	if !out.CrearPlanesBase && !out.CrearClasesBase && !out.CrearDispositivosBase {
		out.CrearPlanesBase = true
		out.CrearClasesBase = true
		out.CrearDispositivosBase = true
	}
	out.UsuarioCreador = strings.TrimSpace(out.UsuarioCreador)
	return &out
}

func GetEmpresaGimnasioPreconfiguracionResumen(dbConn *sql.DB, empresaID int64) (*EmpresaGimnasioPreconfiguracionResumen, error) {
	if empresaID <= 0 {
		return nil, fmt.Errorf("empresa_id es obligatorio")
	}
	if err := EnsureEmpresaGimnasioSchema(dbConn); err != nil {
		return nil, err
	}
	out := &EmpresaGimnasioPreconfiguracionResumen{EmpresaID: empresaID, SedePrincipal: "principal"}
	var accesoConfigCount int
	if err := dbConn.QueryRow(`SELECT
		(SELECT COUNT(1) FROM empresa_gimnasio_socios WHERE empresa_id=?),
		(SELECT COUNT(1) FROM empresa_gimnasio_planes WHERE empresa_id=?),
		(SELECT COUNT(1) FROM empresa_gimnasio_entrenadores WHERE empresa_id=?),
		(SELECT COUNT(1) FROM empresa_gimnasio_clases WHERE empresa_id=?),
		(SELECT COUNT(1) FROM empresa_gimnasio_credenciales WHERE empresa_id=?),
		(SELECT COUNT(1) FROM empresa_gimnasio_dispositivos_acceso WHERE empresa_id=?),
		(SELECT COUNT(1) FROM empresa_gimnasio_acceso_config WHERE empresa_id=?)`,
		empresaID, empresaID, empresaID, empresaID, empresaID, empresaID, empresaID,
	).Scan(&out.Socios, &out.Planes, &out.Entrenadores, &out.Clases, &out.Credenciales, &out.Dispositivos, &accesoConfigCount); err != nil {
		return nil, err
	}
	out.AccesoConfigurado = accesoConfigCount > 0
	out.TieneDatos = out.Socios > 0 || out.Planes > 0 || out.Entrenadores > 0 || out.Clases > 0 || out.Credenciales > 0 || out.Dispositivos > 0 || out.AccesoConfigurado
	if row := dbConn.QueryRow(`SELECT COALESCE(sede,'') FROM empresa_gimnasio_clases WHERE empresa_id=? AND TRIM(COALESCE(sede,''))<>'' ORDER BY id ASC LIMIT 1`, empresaID); row != nil {
		var sede string
		if err := row.Scan(&sede); err == nil && strings.TrimSpace(sede) != "" {
			out.SedePrincipal = strings.TrimSpace(sede)
		}
	}
	return out, nil
}

func ApplyEmpresaGimnasioPreconfiguracion(dbConn *sql.DB, payload EmpresaGimnasioPreconfiguracion) (*EmpresaGimnasioPreconfiguracionResumen, error) {
	if payload.EmpresaID <= 0 {
		return nil, fmt.Errorf("empresa_id es obligatorio")
	}
	if err := EnsureEmpresaGimnasioSchema(dbConn); err != nil {
		return nil, err
	}
	cfg := normalizeGymPreconfig(payload)
	accessCfg := EmpresaGimnasioAccesoConfig{
		EmpresaID:               cfg.EmpresaID,
		ModoValidacionPrincipal: "rfid",
		PermitirRFID:            cfg.PermitirRFID,
		PermitirNFC:             cfg.PermitirNFC,
		PermitirQR:              cfg.PermitirQR,
		PermitirPIN:             false,
		PermitirBiometria:       false,
		PermitirFacial:          false,
		AntiPassbackMinutos:     10,
		MinutosToleranciaMora:   0,
		Estado:                  "activo",
		UsuarioCreador:          cfg.UsuarioCreador,
	}
	if cfg.PermitirRFID {
		accessCfg.ModoValidacionPrincipal = "rfid"
	} else if cfg.PermitirNFC {
		accessCfg.ModoValidacionPrincipal = "nfc"
	} else if cfg.PermitirQR {
		accessCfg.ModoValidacionPrincipal = "qr"
	}
	if _, err := UpsertEmpresaGimnasioAccesoConfig(dbConn, accessCfg); err != nil {
		return nil, err
	}
	if cfg.CrearPlanesBase {
		plans := []EmpresaGimnasioPlan{
			{EmpresaID: cfg.EmpresaID, Nombre: "Pase diario", Precio: 25000, DuracionDias: 1, ClasesIncluidas: 0, AccesoIlimitado: true, Estado: "activo", UsuarioCreador: cfg.UsuarioCreador},
			{EmpresaID: cfg.EmpresaID, Nombre: "Mensual estándar", Precio: 119000, DuracionDias: 30, ClasesIncluidas: 8, AccesoIlimitado: true, Estado: "activo", UsuarioCreador: cfg.UsuarioCreador},
			{EmpresaID: cfg.EmpresaID, Nombre: "Trimestral rendimiento", Precio: 309000, DuracionDias: 90, ClasesIncluidas: 24, AccesoIlimitado: true, SesionesPersonalizadas: 2, Estado: "activo", UsuarioCreador: cfg.UsuarioCreador},
		}
		for _, plan := range plans {
			var existingID int64
			err := dbConn.QueryRow(`SELECT id FROM empresa_gimnasio_planes WHERE empresa_id=? AND lower(nombre)=lower(?) LIMIT 1`, cfg.EmpresaID, plan.Nombre).Scan(&existingID)
			if err == sql.ErrNoRows {
				if _, createErr := CreateEmpresaGimnasioPlan(dbConn, plan); createErr != nil {
					return nil, createErr
				}
			} else if err != nil {
				return nil, err
			}
		}
	}
	if cfg.CrearClasesBase {
		classes := []EmpresaGimnasioClase{
			{EmpresaID: cfg.EmpresaID, Nombre: "Funcional AM", Categoria: "funcional", Sede: cfg.NombreSedePrincipal, Canal: "presencial", Cupos: cfg.CuposBaseClase, DuracionMinutos: cfg.DuracionBaseClase, Precio: 0, Estado: "programada", UsuarioCreador: cfg.UsuarioCreador},
			{EmpresaID: cfg.EmpresaID, Nombre: "Spinning prime", Categoria: "spinning", Sede: cfg.NombreSedePrincipal, Canal: "presencial", Cupos: cfg.CuposBaseClase, DuracionMinutos: cfg.DuracionBaseClase, Precio: 0, Estado: "programada", UsuarioCreador: cfg.UsuarioCreador},
			{EmpresaID: cfg.EmpresaID, Nombre: "Movilidad y core", Categoria: "movilidad", Sede: cfg.NombreSedePrincipal, Canal: "presencial", Cupos: cfg.CuposBaseClase, DuracionMinutos: cfg.DuracionBaseClase, Precio: 0, Estado: "programada", UsuarioCreador: cfg.UsuarioCreador},
		}
		for _, classItem := range classes {
			var existingID int64
			err := dbConn.QueryRow(`SELECT id FROM empresa_gimnasio_clases WHERE empresa_id=? AND lower(nombre)=lower(?) AND lower(COALESCE(sede,''))=lower(?) LIMIT 1`, cfg.EmpresaID, classItem.Nombre, classItem.Sede).Scan(&existingID)
			if err == sql.ErrNoRows {
				if _, createErr := CreateEmpresaGimnasioClase(dbConn, classItem); createErr != nil {
					return nil, createErr
				}
			} else if err != nil {
				return nil, err
			}
		}
	}
	if cfg.CrearDispositivosBase {
		devices := []EmpresaGimnasioDispositivoAcceso{}
		if cfg.PermitirRFID {
			devices = append(devices, EmpresaGimnasioDispositivoAcceso{
				EmpresaID:       cfg.EmpresaID,
				Nombre:          "Ingreso principal RFID",
				TipoDispositivo: "lector_rfid",
				Ubicacion:       "Recepción principal",
				Sede:            cfg.NombreSedePrincipal,
				Canal:           "ingreso",
				Estado:          "activo",
				Identificador:   "RFID-PRINCIPAL",
				UsuarioCreador:  cfg.UsuarioCreador,
			})
		}
		if cfg.PermitirNFC {
			devices = append(devices, EmpresaGimnasioDispositivoAcceso{
				EmpresaID:       cfg.EmpresaID,
				Nombre:          "Acceso NFC lobby",
				TipoDispositivo: "lector_nfc",
				Ubicacion:       "Lobby de ingreso",
				Sede:            cfg.NombreSedePrincipal,
				Canal:           "ingreso",
				Estado:          "activo",
				Identificador:   "NFC-LOBBY",
				UsuarioCreador:  cfg.UsuarioCreador,
			})
		}
		if cfg.PermitirQR {
			devices = append(devices, EmpresaGimnasioDispositivoAcceso{
				EmpresaID:       cfg.EmpresaID,
				Nombre:          "Torniquete QR recepción",
				TipoDispositivo: "torniquete_qr",
				Ubicacion:       "Recepción principal",
				Sede:            cfg.NombreSedePrincipal,
				Canal:           "ingreso",
				Estado:          "activo",
				Identificador:   "QR-RECEPCION",
				UsuarioCreador:  cfg.UsuarioCreador,
			})
		}
		for _, device := range devices {
			var existingID int64
			err := dbConn.QueryRow(`SELECT id FROM empresa_gimnasio_dispositivos_acceso WHERE empresa_id=? AND lower(nombre)=lower(?) LIMIT 1`, cfg.EmpresaID, device.Nombre).Scan(&existingID)
			if err == sql.ErrNoRows {
				if _, createErr := CreateEmpresaGimnasioDispositivoAcceso(dbConn, device); createErr != nil {
					return nil, createErr
				}
			} else if err != nil {
				return nil, err
			}
		}
	}
	return GetEmpresaGimnasioPreconfiguracionResumen(dbConn, cfg.EmpresaID)
}

func GetEmpresaGimnasioAccesoConfig(dbConn *sql.DB, empresaID int64) (*EmpresaGimnasioAccesoConfig, error) {
	if empresaID <= 0 {
		return nil, fmt.Errorf("empresa_id es obligatorio")
	}
	if err := EnsureEmpresaGimnasioSchema(dbConn); err != nil {
		return nil, err
	}
	cfg := normalizeGymAccessConfig(EmpresaGimnasioAccesoConfig{EmpresaID: empresaID, PermitirRFID: true, PermitirNFC: true, PermitirQR: true, Estado: "activo", AntiPassbackMinutos: 10})
	var permitirRFID, permitirNFC, permitirQR, permitirPIN, permitirBiometria, permitirFacial int
	err := dbConn.QueryRow(`SELECT id, empresa_id, COALESCE(modo_validacion_principal,'rfid'),
		COALESCE(permitir_rfid,1), COALESCE(permitir_nfc,1), COALESCE(permitir_qr,1), COALESCE(permitir_pin,0),
		COALESCE(permitir_biometria,0), COALESCE(permitir_facial,0), COALESCE(anti_passback_minutos,10), COALESCE(minutos_tolerancia_mora,0),
		COALESCE(estado,'activo'), COALESCE(fecha_actualizacion,''), COALESCE(usuario_creador,'')
	FROM empresa_gimnasio_acceso_config WHERE empresa_id=? LIMIT 1`, empresaID).Scan(
		&cfg.ID, &cfg.EmpresaID, &cfg.ModoValidacionPrincipal,
		&permitirRFID, &permitirNFC, &permitirQR, &permitirPIN, &permitirBiometria, &permitirFacial, &cfg.AntiPassbackMinutos, &cfg.MinutosToleranciaMora,
		&cfg.Estado, &cfg.FechaActualizacion, &cfg.UsuarioCreador,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return cfg, nil
		}
		return nil, err
	}
	cfg.PermitirRFID = permitirRFID > 0
	cfg.PermitirNFC = permitirNFC > 0
	cfg.PermitirQR = permitirQR > 0
	cfg.PermitirPIN = permitirPIN > 0
	cfg.PermitirBiometria = permitirBiometria > 0
	cfg.PermitirFacial = permitirFacial > 0
	return cfg, nil
}

func UpsertEmpresaGimnasioAccesoConfig(dbConn *sql.DB, payload EmpresaGimnasioAccesoConfig) (int64, error) {
	if payload.EmpresaID <= 0 {
		return 0, fmt.Errorf("empresa_id es obligatorio")
	}
	if err := EnsureEmpresaGimnasioSchema(dbConn); err != nil {
		return 0, err
	}
	cfg := normalizeGymAccessConfig(payload)
	toInt := func(v bool) int {
		if v {
			return 1
		}
		return 0
	}
	var existingID int64
	err := dbConn.QueryRow(`SELECT id FROM empresa_gimnasio_acceso_config WHERE empresa_id=? LIMIT 1`, cfg.EmpresaID).Scan(&existingID)
	if err == sql.ErrNoRows {
		return insertSQLCompat(dbConn, `INSERT INTO empresa_gimnasio_acceso_config (
			empresa_id, modo_validacion_principal, permitir_rfid, permitir_nfc, permitir_qr, permitir_pin, permitir_biometria, permitir_facial,
			anti_passback_minutos, minutos_tolerancia_mora, estado, fecha_actualizacion, usuario_creador
		) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)`,
			cfg.EmpresaID, cfg.ModoValidacionPrincipal, toInt(cfg.PermitirRFID), toInt(cfg.PermitirNFC), toInt(cfg.PermitirQR), toInt(cfg.PermitirPIN), toInt(cfg.PermitirBiometria), toInt(cfg.PermitirFacial),
			cfg.AntiPassbackMinutos, cfg.MinutosToleranciaMora, cfg.Estado, time.Now().Format("2006-01-02 15:04:05"), cfg.UsuarioCreador,
		)
	}
	if err != nil {
		return 0, err
	}
	_, err = dbConn.Exec(`UPDATE empresa_gimnasio_acceso_config SET
		modo_validacion_principal=?, permitir_rfid=?, permitir_nfc=?, permitir_qr=?, permitir_pin=?, permitir_biometria=?, permitir_facial=?,
		anti_passback_minutos=?, minutos_tolerancia_mora=?, estado=?, fecha_actualizacion=?, usuario_creador=?
	WHERE empresa_id=?`,
		cfg.ModoValidacionPrincipal, toInt(cfg.PermitirRFID), toInt(cfg.PermitirNFC), toInt(cfg.PermitirQR), toInt(cfg.PermitirPIN), toInt(cfg.PermitirBiometria), toInt(cfg.PermitirFacial),
		cfg.AntiPassbackMinutos, cfg.MinutosToleranciaMora, cfg.Estado, time.Now().Format("2006-01-02 15:04:05"), cfg.UsuarioCreador, cfg.EmpresaID,
	)
	return existingID, err
}

func ListEmpresaGimnasioCredenciales(dbConn *sql.DB, empresaID int64) ([]EmpresaGimnasioCredencial, error) {
	if empresaID <= 0 {
		return nil, fmt.Errorf("empresa_id es obligatorio")
	}
	if err := EnsureEmpresaGimnasioSchema(dbConn); err != nil {
		return nil, err
	}
	rows, err := dbConn.Query(`SELECT c.id, c.empresa_id, c.socio_id, COALESCE(s.nombre_completo,''), COALESCE(c.tipo_credencial,''), COALESCE(c.codigo_credencial,''),
		COALESCE(c.alias_credencial,''), COALESCE(c.estado,'activa'), COALESCE(c.fecha_expiracion,''), COALESCE(c.ultimo_uso,''), COALESCE(c.fecha_creacion,''),
		COALESCE(c.fecha_actualizacion,''), COALESCE(c.usuario_creador,'')
	FROM empresa_gimnasio_credenciales c
	INNER JOIN empresa_gimnasio_socios s ON s.id=c.socio_id AND s.empresa_id=c.empresa_id
	WHERE c.empresa_id=?
	ORDER BY c.id DESC`, empresaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []EmpresaGimnasioCredencial
	for rows.Next() {
		var item EmpresaGimnasioCredencial
		if err := rows.Scan(&item.ID, &item.EmpresaID, &item.SocioID, &item.SocioNombre, &item.TipoCredencial, &item.CodigoCredencial,
			&item.AliasCredencial, &item.Estado, &item.FechaExpiracion, &item.UltimoUso, &item.FechaCreacion, &item.FechaActualizacion, &item.UsuarioCreador); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func CreateEmpresaGimnasioCredencial(dbConn *sql.DB, payload EmpresaGimnasioCredencial) (int64, error) {
	if err := EnsureEmpresaGimnasioSchema(dbConn); err != nil {
		return 0, err
	}
	item, err := normalizeGymCredencial(payload)
	if err != nil {
		return 0, err
	}
	return insertSQLCompat(dbConn, `INSERT INTO empresa_gimnasio_credenciales (
		empresa_id, socio_id, tipo_credencial, codigo_credencial, alias_credencial, estado, fecha_expiracion, fecha_actualizacion, usuario_creador
	) VALUES (?,?,?,?,?,?,?,?,?)`,
		item.EmpresaID, item.SocioID, item.TipoCredencial, item.CodigoCredencial, item.AliasCredencial, item.Estado, item.FechaExpiracion, time.Now().Format("2006-01-02 15:04:05"), item.UsuarioCreador)
}

func UpdateEmpresaGimnasioCredencial(dbConn *sql.DB, payload EmpresaGimnasioCredencial) error {
	if err := EnsureEmpresaGimnasioSchema(dbConn); err != nil {
		return err
	}
	item, err := normalizeGymCredencial(payload)
	if err != nil {
		return err
	}
	if item.ID <= 0 {
		return fmt.Errorf("id es obligatorio")
	}
	res, err := dbConn.Exec(`UPDATE empresa_gimnasio_credenciales SET socio_id=?, tipo_credencial=?, codigo_credencial=?, alias_credencial=?, estado=?, fecha_expiracion=?, fecha_actualizacion=? WHERE id=? AND empresa_id=?`,
		item.SocioID, item.TipoCredencial, item.CodigoCredencial, item.AliasCredencial, item.Estado, item.FechaExpiracion, time.Now().Format("2006-01-02 15:04:05"), item.ID, item.EmpresaID)
	if err != nil {
		return err
	}
	return ensureRowsAffected(res)
}

func DeleteEmpresaGimnasioCredencial(dbConn *sql.DB, empresaID, id int64) error {
	return simpleDeleteByEmpresa(dbConn, empresaID, id, "empresa_gimnasio_credenciales")
}

func ListEmpresaGimnasioDispositivosAcceso(dbConn *sql.DB, empresaID int64) ([]EmpresaGimnasioDispositivoAcceso, error) {
	if empresaID <= 0 {
		return nil, fmt.Errorf("empresa_id es obligatorio")
	}
	if err := EnsureEmpresaGimnasioSchema(dbConn); err != nil {
		return nil, err
	}
	rows, err := dbConn.Query(`SELECT id, empresa_id, COALESCE(nombre,''), COALESCE(tipo_dispositivo,''), COALESCE(ubicacion,''), COALESCE(sede,''), COALESCE(canal,''), COALESCE(estado,'activo'), COALESCE(identificador,''), COALESCE(observaciones,''), COALESCE(fecha_creacion,''), COALESCE(fecha_actualizacion,''), COALESCE(usuario_creador,'')
	FROM empresa_gimnasio_dispositivos_acceso WHERE empresa_id=? ORDER BY id DESC`, empresaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []EmpresaGimnasioDispositivoAcceso
	for rows.Next() {
		var item EmpresaGimnasioDispositivoAcceso
		if err := rows.Scan(&item.ID, &item.EmpresaID, &item.Nombre, &item.TipoDispositivo, &item.Ubicacion, &item.Sede, &item.Canal, &item.Estado, &item.Identificador, &item.Observaciones, &item.FechaCreacion, &item.FechaActualizacion, &item.UsuarioCreador); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func CreateEmpresaGimnasioDispositivoAcceso(dbConn *sql.DB, payload EmpresaGimnasioDispositivoAcceso) (int64, error) {
	if err := EnsureEmpresaGimnasioSchema(dbConn); err != nil {
		return 0, err
	}
	item, err := normalizeGymDispositivo(payload)
	if err != nil {
		return 0, err
	}
	return insertSQLCompat(dbConn, `INSERT INTO empresa_gimnasio_dispositivos_acceso (empresa_id, nombre, tipo_dispositivo, ubicacion, sede, canal, estado, identificador, observaciones, fecha_actualizacion, usuario_creador) VALUES (?,?,?,?,?,?,?,?,?,?,?)`,
		item.EmpresaID, item.Nombre, item.TipoDispositivo, item.Ubicacion, item.Sede, item.Canal, item.Estado, item.Identificador, item.Observaciones, time.Now().Format("2006-01-02 15:04:05"), item.UsuarioCreador)
}

func UpdateEmpresaGimnasioDispositivoAcceso(dbConn *sql.DB, payload EmpresaGimnasioDispositivoAcceso) error {
	if err := EnsureEmpresaGimnasioSchema(dbConn); err != nil {
		return err
	}
	item, err := normalizeGymDispositivo(payload)
	if err != nil {
		return err
	}
	if item.ID <= 0 {
		return fmt.Errorf("id es obligatorio")
	}
	res, err := dbConn.Exec(`UPDATE empresa_gimnasio_dispositivos_acceso SET nombre=?, tipo_dispositivo=?, ubicacion=?, sede=?, canal=?, estado=?, identificador=?, observaciones=?, fecha_actualizacion=? WHERE id=? AND empresa_id=?`,
		item.Nombre, item.TipoDispositivo, item.Ubicacion, item.Sede, item.Canal, item.Estado, item.Identificador, item.Observaciones, time.Now().Format("2006-01-02 15:04:05"), item.ID, item.EmpresaID)
	if err != nil {
		return err
	}
	return ensureRowsAffected(res)
}

func DeleteEmpresaGimnasioDispositivoAcceso(dbConn *sql.DB, empresaID, id int64) error {
	return simpleDeleteByEmpresa(dbConn, empresaID, id, "empresa_gimnasio_dispositivos_acceso")
}

func ListEmpresaGimnasioEventosAcceso(dbConn *sql.DB, empresaID int64) ([]EmpresaGimnasioEventoAcceso, error) {
	if empresaID <= 0 {
		return nil, fmt.Errorf("empresa_id es obligatorio")
	}
	if err := EnsureEmpresaGimnasioSchema(dbConn); err != nil {
		return nil, err
	}
	rows, err := dbConn.Query(`SELECT e.id, e.empresa_id, COALESCE(e.socio_id,0), COALESCE(s.nombre_completo,''), COALESCE(e.credencial_id,0), COALESCE(e.codigo_credencial,''), COALESCE(e.dispositivo_id,0), COALESCE(d.nombre,''), COALESCE(e.metodo_acceso,''), COALESCE(e.resultado,''), COALESCE(e.motivo,''), COALESCE(e.fecha_evento,''), COALESCE(e.canal,''), COALESCE(e.sede,''), COALESCE(e.observaciones,''), COALESCE(e.usuario_creador,'')
	FROM empresa_gimnasio_eventos_acceso e
	LEFT JOIN empresa_gimnasio_socios s ON s.id=e.socio_id AND s.empresa_id=e.empresa_id
	LEFT JOIN empresa_gimnasio_dispositivos_acceso d ON d.id=e.dispositivo_id AND d.empresa_id=e.empresa_id
	WHERE e.empresa_id=?
	ORDER BY e.fecha_evento DESC, e.id DESC LIMIT 120`, empresaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []EmpresaGimnasioEventoAcceso
	for rows.Next() {
		var item EmpresaGimnasioEventoAcceso
		if err := rows.Scan(&item.ID, &item.EmpresaID, &item.SocioID, &item.SocioNombre, &item.CredencialID, &item.CodigoCredencial, &item.DispositivoID, &item.DispositivoNombre, &item.MetodoAcceso, &item.Resultado, &item.Motivo, &item.FechaEvento, &item.Canal, &item.Sede, &item.Observaciones, &item.UsuarioCreador); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func RegistrarEmpresaGimnasioEventoAcceso(dbConn *sql.DB, payload EmpresaGimnasioEventoAcceso) (int64, error) {
	if err := EnsureEmpresaGimnasioSchema(dbConn); err != nil {
		return 0, err
	}
	if payload.EmpresaID <= 0 {
		return 0, fmt.Errorf("empresa_id es obligatorio")
	}
	if strings.TrimSpace(payload.FechaEvento) == "" {
		payload.FechaEvento = time.Now().Format("2006-01-02 15:04:05")
	}
	return insertSQLCompat(dbConn, `INSERT INTO empresa_gimnasio_eventos_acceso (empresa_id, socio_id, credencial_id, dispositivo_id, codigo_credencial, metodo_acceso, resultado, motivo, fecha_evento, canal, sede, observaciones, usuario_creador) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		payload.EmpresaID, payload.SocioID, payload.CredencialID, payload.DispositivoID, strings.TrimSpace(payload.CodigoCredencial), strings.TrimSpace(payload.MetodoAcceso), normalizeGymState(payload.Resultado, "aprobado"), strings.TrimSpace(payload.Motivo), payload.FechaEvento, strings.TrimSpace(payload.Canal), strings.TrimSpace(payload.Sede), strings.TrimSpace(payload.Observaciones), strings.TrimSpace(payload.UsuarioCreador))
}

func ValidarEmpresaGimnasioAcceso(dbConn *sql.DB, empresaID int64, codigoCredencial, metodo string, dispositivoID int64, usuario string) (*EmpresaGimnasioEventoAcceso, error) {
	if empresaID <= 0 {
		return nil, fmt.Errorf("empresa_id es obligatorio")
	}
	if err := EnsureEmpresaGimnasioSchema(dbConn); err != nil {
		return nil, err
	}
	codigoCredencial = strings.TrimSpace(codigoCredencial)
	metodo = strings.ToLower(strings.TrimSpace(metodo))
	if codigoCredencial == "" {
		return nil, fmt.Errorf("codigo_credencial es obligatorio")
	}
	cfg, err := GetEmpresaGimnasioAccesoConfig(dbConn, empresaID)
	if err != nil {
		return nil, err
	}
	evento := &EmpresaGimnasioEventoAcceso{EmpresaID: empresaID, CodigoCredencial: codigoCredencial, MetodoAcceso: metodo, DispositivoID: dispositivoID, UsuarioCreador: strings.TrimSpace(usuario), Resultado: "denegado"}
	var cred EmpresaGimnasioCredencial
	err = dbConn.QueryRow(`SELECT c.id, c.empresa_id, c.socio_id, COALESCE(s.nombre_completo,''), COALESCE(c.tipo_credencial,''), COALESCE(c.codigo_credencial,''), COALESCE(c.alias_credencial,''), COALESCE(c.estado,'activa'), COALESCE(c.fecha_expiracion,''), COALESCE(c.ultimo_uso,''), COALESCE(c.fecha_creacion,''), COALESCE(c.fecha_actualizacion,''), COALESCE(c.usuario_creador,'')
	FROM empresa_gimnasio_credenciales c
	INNER JOIN empresa_gimnasio_socios s ON s.id=c.socio_id AND s.empresa_id=c.empresa_id
	WHERE c.empresa_id=? AND c.codigo_credencial=? LIMIT 1`, empresaID, codigoCredencial).Scan(&cred.ID, &cred.EmpresaID, &cred.SocioID, &cred.SocioNombre, &cred.TipoCredencial, &cred.CodigoCredencial, &cred.AliasCredencial, &cred.Estado, &cred.FechaExpiracion, &cred.UltimoUso, &cred.FechaCreacion, &cred.FechaActualizacion, &cred.UsuarioCreador)
	if err != nil {
		evento.Motivo = "credencial_no_encontrada"
		_, _ = RegistrarEmpresaGimnasioEventoAcceso(dbConn, *evento)
		return evento, nil
	}
	evento.SocioID = cred.SocioID
	evento.SocioNombre = cred.SocioNombre
	evento.CredencialID = cred.ID
	if !isGymMethodAllowed(cfg, cred.TipoCredencial) {
		evento.Motivo = "metodo_no_habilitado"
		_, _ = RegistrarEmpresaGimnasioEventoAcceso(dbConn, *evento)
		return evento, nil
	}
	if strings.ToLower(cred.Estado) != "activa" {
		evento.Motivo = "credencial_inactiva"
		_, _ = RegistrarEmpresaGimnasioEventoAcceso(dbConn, *evento)
		return evento, nil
	}
	var socioEstado, fechaFinPlan string
	var saldo float64
	err = dbConn.QueryRow(`SELECT COALESCE(estado,'activo'), COALESCE(fecha_fin_plan,''), COALESCE(saldo,0) FROM empresa_gimnasio_socios WHERE empresa_id=? AND id=? LIMIT 1`, empresaID, cred.SocioID).Scan(&socioEstado, &fechaFinPlan, &saldo)
	if err != nil {
		evento.Motivo = "socio_no_encontrado"
		_, _ = RegistrarEmpresaGimnasioEventoAcceso(dbConn, *evento)
		return evento, nil
	}
	if strings.ToLower(socioEstado) != "activo" {
		evento.Motivo = "socio_inactivo"
		_, _ = RegistrarEmpresaGimnasioEventoAcceso(dbConn, *evento)
		return evento, nil
	}
	if strings.TrimSpace(fechaFinPlan) != "" {
		today := time.Now().Format("2006-01-02")
		if strings.TrimSpace(fechaFinPlan) < today {
			evento.Motivo = "membresia_vencida"
			_, _ = RegistrarEmpresaGimnasioEventoAcceso(dbConn, *evento)
			return evento, nil
		}
	}
	if cfg.MinutosToleranciaMora <= 0 && saldo > 0 {
		evento.Motivo = "saldo_pendiente"
		_, _ = RegistrarEmpresaGimnasioEventoAcceso(dbConn, *evento)
		return evento, nil
	}
	if cfg.AntiPassbackMinutos > 0 {
		var ultimo string
		_ = dbConn.QueryRow(`SELECT COALESCE(fecha_evento,'') FROM empresa_gimnasio_eventos_acceso WHERE empresa_id=? AND socio_id=? AND resultado='aprobado' ORDER BY id DESC LIMIT 1`, empresaID, cred.SocioID).Scan(&ultimo)
		if strings.TrimSpace(ultimo) != "" {
			if t, parseErr := time.Parse("2006-01-02 15:04:05", ultimo); parseErr == nil && time.Since(t) < time.Duration(cfg.AntiPassbackMinutos)*time.Minute {
				evento.Motivo = "anti_passback"
				_, _ = RegistrarEmpresaGimnasioEventoAcceso(dbConn, *evento)
				return evento, nil
			}
		}
	}
	evento.Resultado = "aprobado"
	evento.Motivo = "ok"
	evento.FechaEvento = time.Now().Format("2006-01-02 15:04:05")
	_, _ = dbConn.Exec(`UPDATE empresa_gimnasio_credenciales SET ultimo_uso=?, fecha_actualizacion=? WHERE id=? AND empresa_id=?`, evento.FechaEvento, evento.FechaEvento, cred.ID, empresaID)
	_, _ = RegistrarEmpresaGimnasioEventoAcceso(dbConn, *evento)
	return evento, nil
}

func isGymMethodAllowed(cfg *EmpresaGimnasioAccesoConfig, method string) bool {
	method = strings.ToLower(strings.TrimSpace(method))
	switch method {
	case "rfid", "tarjeta_rfid":
		return cfg.PermitirRFID
	case "nfc", "tarjeta_nfc":
		return cfg.PermitirNFC
	case "qr", "qr_dinamico", "qr_estatico", "codigo_barras":
		return cfg.PermitirQR
	case "pin":
		return cfg.PermitirPIN
	case "biometria":
		return cfg.PermitirBiometria
	case "facial", "reconocimiento_facial":
		return cfg.PermitirFacial
	default:
		return true
	}
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
