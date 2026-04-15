package db

import (
	"database/sql"
	"fmt"
	"strings"
)

// ChatConversacion representa un hilo de comunicacion interna por empresa.
type ChatConversacion struct {
	ID                 int64  `json:"id"`
	EmpresaID          int64  `json:"empresa_id"`
	Titulo             string `json:"titulo"`
	Descripcion        string `json:"descripcion,omitempty"`
	Prioridad          string `json:"prioridad,omitempty"`
	EstadoConversacion string `json:"estado_conversacion,omitempty"`
	UltimoMensajeEn    string `json:"ultimo_mensaje_en,omitempty"`
	MensajesCount      int64  `json:"mensajes_count,omitempty"`
	TareasPendientes   int64  `json:"tareas_pendientes,omitempty"`
	FechaCreacion      string `json:"fecha_creacion,omitempty"`
	FechaActualizacion string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string `json:"usuario_creador,omitempty"`
	Estado             string `json:"estado,omitempty"`
	Observaciones      string `json:"observaciones,omitempty"`
}

// ChatParticipante representa un usuario/admin participante de una conversacion.
type ChatParticipante struct {
	ID                 int64  `json:"id"`
	EmpresaID          int64  `json:"empresa_id"`
	ConversacionID     int64  `json:"conversacion_id"`
	ParticipanteTipo   string `json:"participante_tipo,omitempty"`
	ParticipanteRefID  int64  `json:"participante_ref_id,omitempty"`
	Nombre             string `json:"nombre,omitempty"`
	Email              string `json:"email,omitempty"`
	ActivoChat         int    `json:"activo_chat,omitempty"`
	FechaCreacion      string `json:"fecha_creacion,omitempty"`
	FechaActualizacion string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string `json:"usuario_creador,omitempty"`
	Estado             string `json:"estado,omitempty"`
	Observaciones      string `json:"observaciones,omitempty"`
}

// ChatAdjunto representa un archivo adjunto (imagen, audio, etc.) en un mensaje.
type ChatAdjunto struct {
	ID                 int64   `json:"id"`
	EmpresaID          int64   `json:"empresa_id"`
	MensajeID          int64   `json:"mensaje_id"`
	TipoArchivo        string  `json:"tipo_archivo,omitempty"`
	NombreArchivo      string  `json:"nombre_archivo,omitempty"`
	MimeType           string  `json:"mime_type,omitempty"`
	FileURL            string  `json:"file_url"`
	TamanoBytes        int64   `json:"tamano_bytes,omitempty"`
	DuracionSegundos   float64 `json:"duracion_segundos,omitempty"`
	FechaCreacion      string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string  `json:"usuario_creador,omitempty"`
	Estado             string  `json:"estado,omitempty"`
	Observaciones      string  `json:"observaciones,omitempty"`
}

// ChatMensaje representa un mensaje dentro de una conversacion.
type ChatMensaje struct {
	ID                 int64         `json:"id"`
	EmpresaID          int64         `json:"empresa_id"`
	ConversacionID     int64         `json:"conversacion_id"`
	AutorTipo          string        `json:"autor_tipo,omitempty"`
	AutorRefID         int64         `json:"autor_ref_id,omitempty"`
	AutorNombre        string        `json:"autor_nombre,omitempty"`
	AutorEmail         string        `json:"autor_email,omitempty"`
	Contenido          string        `json:"contenido,omitempty"`
	TipoMensaje        string        `json:"tipo_mensaje,omitempty"`
	FechaEnvio         string        `json:"fecha_envio,omitempty"`
	FechaCreacion      string        `json:"fecha_creacion,omitempty"`
	FechaActualizacion string        `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string        `json:"usuario_creador,omitempty"`
	Estado             string        `json:"estado,omitempty"`
	Observaciones      string        `json:"observaciones,omitempty"`
	Adjuntos           []ChatAdjunto `json:"adjuntos,omitempty"`
}

// ChatTarea representa una tarea asociada a una empresa y opcionalmente a una conversacion.
type ChatTarea struct {
	ID                 int64   `json:"id"`
	EmpresaID          int64   `json:"empresa_id"`
	ConversacionID     int64   `json:"conversacion_id,omitempty"`
	Titulo             string  `json:"titulo"`
	Descripcion        string  `json:"descripcion,omitempty"`
	Prioridad          string  `json:"prioridad,omitempty"`
	FechaLimite        string  `json:"fecha_limite,omitempty"`
	AsignadoTipo       string  `json:"asignado_tipo,omitempty"`
	AsignadoRefID      int64   `json:"asignado_ref_id,omitempty"`
	AsignadoNombre     string  `json:"asignado_nombre,omitempty"`
	AsignadoEmail      string  `json:"asignado_email,omitempty"`
	CreadoPorTipo      string  `json:"creado_por_tipo,omitempty"`
	CreadoPorEmail     string  `json:"creado_por_email,omitempty"`
	EstadoTarea        string  `json:"estado_tarea,omitempty"`
	PorcentajeAvance   int     `json:"porcentaje_avance,omitempty"`
	CompletadaEn       string  `json:"completada_en,omitempty"`
	NotaVozURL         string  `json:"nota_voz_url,omitempty"`
	NotaVozMimeType    string  `json:"nota_voz_mime_type,omitempty"`
	NotaVozTamanoBytes int64   `json:"nota_voz_tamano_bytes,omitempty"`
	NotaVozDuracionSeg float64 `json:"nota_voz_duracion_segundos,omitempty"`
	FechaCreacion      string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string  `json:"usuario_creador,omitempty"`
	Estado             string  `json:"estado,omitempty"`
	Observaciones      string  `json:"observaciones,omitempty"`
}

// ChatCita representa una cita/reunion de agenda compartida por empresa.
type ChatCita struct {
	ID                    int64  `json:"id"`
	EmpresaID             int64  `json:"empresa_id"`
	ConversacionID        int64  `json:"conversacion_id,omitempty"`
	Titulo                string `json:"titulo"`
	Descripcion           string `json:"descripcion,omitempty"`
	TipoCita              string `json:"tipo_cita,omitempty"`
	FechaInicio           string `json:"fecha_inicio"`
	FechaFin              string `json:"fecha_fin,omitempty"`
	Ubicacion             string `json:"ubicacion,omitempty"`
	NotificarMinutosAntes int    `json:"notificar_minutos_antes,omitempty"`
	CreadoPorTipo         string `json:"creado_por_tipo,omitempty"`
	CreadoPorRefID        int64  `json:"creado_por_ref_id,omitempty"`
	CreadoPorNombre       string `json:"creado_por_nombre,omitempty"`
	CreadoPorEmail        string `json:"creado_por_email,omitempty"`
	EstadoCita            string `json:"estado_cita,omitempty"`
	RecordatorioEnviado   int    `json:"recordatorio_enviado,omitempty"`
	RecordatorioEnviadoEn string `json:"recordatorio_enviado_en,omitempty"`
	Visibilidad           string `json:"visibilidad,omitempty"`
	FechaCreacion         string `json:"fecha_creacion,omitempty"`
	FechaActualizacion    string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador        string `json:"usuario_creador,omitempty"`
	Estado                string `json:"estado,omitempty"`
	Observaciones         string `json:"observaciones,omitempty"`
}

// EnsureEmpresaChatTareasSchema crea y migra las tablas del modulo chat/tareas en empresas.db.
func EnsureEmpresaChatTareasSchema(dbConn *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS chat_tareas_conversaciones (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			titulo TEXT NOT NULL,
			descripcion TEXT,
			prioridad TEXT DEFAULT 'media',
			estado_conversacion TEXT DEFAULT 'abierta',
			ultimo_mensaje_en TEXT,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS chat_tareas_participantes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			conversacion_id INTEGER NOT NULL,
			participante_tipo TEXT DEFAULT 'usuario',
			participante_ref_id INTEGER,
			nombre TEXT,
			email TEXT,
			activo_chat INTEGER DEFAULT 1,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			UNIQUE(empresa_id, conversacion_id, participante_tipo, participante_ref_id, email)
		);`,
		`CREATE TABLE IF NOT EXISTS chat_tareas_mensajes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			conversacion_id INTEGER NOT NULL,
			autor_tipo TEXT DEFAULT 'admin',
			autor_ref_id INTEGER,
			autor_nombre TEXT,
			autor_email TEXT,
			contenido TEXT,
			tipo_mensaje TEXT DEFAULT 'texto',
			fecha_envio TEXT DEFAULT (datetime('now','localtime')),
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS chat_tareas_adjuntos (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			mensaje_id INTEGER NOT NULL,
			tipo_archivo TEXT DEFAULT 'otro',
			nombre_archivo TEXT,
			mime_type TEXT,
			file_url TEXT NOT NULL,
			tamano_bytes INTEGER DEFAULT 0,
			duracion_segundos REAL DEFAULT 0,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS chat_tareas (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			conversacion_id INTEGER,
			titulo TEXT NOT NULL,
			descripcion TEXT,
			prioridad TEXT DEFAULT 'media',
			fecha_limite TEXT,
			asignado_tipo TEXT DEFAULT 'usuario',
			asignado_ref_id INTEGER,
			asignado_nombre TEXT,
			asignado_email TEXT,
			creado_por_tipo TEXT DEFAULT 'admin',
			creado_por_email TEXT,
			estado_tarea TEXT DEFAULT 'pendiente',
			porcentaje_avance INTEGER DEFAULT 0,
			completada_en TEXT,
			nota_voz_url TEXT,
			nota_voz_mime_type TEXT,
			nota_voz_tamano_bytes INTEGER DEFAULT 0,
			nota_voz_duracion_segundos REAL DEFAULT 0,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS chat_tareas_citas (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			conversacion_id INTEGER,
			titulo TEXT NOT NULL,
			descripcion TEXT,
			tipo_cita TEXT DEFAULT 'reunion',
			fecha_inicio TEXT NOT NULL,
			fecha_fin TEXT,
			ubicacion TEXT,
			notificar_minutos_antes INTEGER DEFAULT 30,
			creado_por_tipo TEXT DEFAULT 'admin',
			creado_por_ref_id INTEGER,
			creado_por_nombre TEXT,
			creado_por_email TEXT,
			estado_cita TEXT DEFAULT 'programada',
			recordatorio_enviado INTEGER DEFAULT 0,
			recordatorio_enviado_en TEXT,
			visibilidad TEXT DEFAULT 'empresa',
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE INDEX IF NOT EXISTS ix_chat_conv_empresa_estado ON chat_tareas_conversaciones(empresa_id, estado, estado_conversacion);`,
		`CREATE INDEX IF NOT EXISTS ix_chat_participantes_empresa_conv ON chat_tareas_participantes(empresa_id, conversacion_id, estado);`,
		`CREATE INDEX IF NOT EXISTS ix_chat_msg_empresa_conv ON chat_tareas_mensajes(empresa_id, conversacion_id, estado, fecha_envio);`,
		`CREATE INDEX IF NOT EXISTS ix_chat_adj_empresa_mensaje ON chat_tareas_adjuntos(empresa_id, mensaje_id, estado);`,
		`CREATE INDEX IF NOT EXISTS ix_chat_tareas_empresa_estado ON chat_tareas(empresa_id, estado, estado_tarea);`,
		`CREATE INDEX IF NOT EXISTS ix_chat_tareas_empresa_conv ON chat_tareas(empresa_id, conversacion_id);`,
		`CREATE INDEX IF NOT EXISTS ix_chat_citas_empresa_fecha ON chat_tareas_citas(empresa_id, fecha_inicio, estado, estado_cita);`,
		`CREATE INDEX IF NOT EXISTS ix_chat_citas_empresa_conv ON chat_tareas_citas(empresa_id, conversacion_id);`,
	}
	for _, stmt := range stmts {
		if _, err := dbConn.Exec(stmt); err != nil {
			return err
		}
	}

	if err := ensureColumnIfMissing(dbConn, "chat_tareas_conversaciones", "descripcion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "chat_tareas_conversaciones", "prioridad", "TEXT DEFAULT 'media'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "chat_tareas_conversaciones", "estado_conversacion", "TEXT DEFAULT 'abierta'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "chat_tareas_conversaciones", "ultimo_mensaje_en", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "chat_tareas_conversaciones", "fecha_actualizacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "chat_tareas_conversaciones", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "chat_tareas_conversaciones", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "chat_tareas_conversaciones", "observaciones", "TEXT"); err != nil {
		return err
	}

	if err := ensureColumnIfMissing(dbConn, "chat_tareas_participantes", "participante_tipo", "TEXT DEFAULT 'usuario'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "chat_tareas_participantes", "participante_ref_id", "INTEGER"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "chat_tareas_participantes", "nombre", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "chat_tareas_participantes", "email", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "chat_tareas_participantes", "activo_chat", "INTEGER DEFAULT 1"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "chat_tareas_participantes", "fecha_actualizacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "chat_tareas_participantes", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "chat_tareas_participantes", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "chat_tareas_participantes", "observaciones", "TEXT"); err != nil {
		return err
	}

	if err := ensureColumnIfMissing(dbConn, "chat_tareas_mensajes", "autor_tipo", "TEXT DEFAULT 'admin'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "chat_tareas_mensajes", "autor_ref_id", "INTEGER"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "chat_tareas_mensajes", "autor_nombre", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "chat_tareas_mensajes", "autor_email", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "chat_tareas_mensajes", "contenido", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "chat_tareas_mensajes", "tipo_mensaje", "TEXT DEFAULT 'texto'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "chat_tareas_mensajes", "fecha_envio", "TEXT DEFAULT (datetime('now','localtime'))"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "chat_tareas_mensajes", "fecha_actualizacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "chat_tareas_mensajes", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "chat_tareas_mensajes", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "chat_tareas_mensajes", "observaciones", "TEXT"); err != nil {
		return err
	}

	if err := ensureColumnIfMissing(dbConn, "chat_tareas_adjuntos", "tipo_archivo", "TEXT DEFAULT 'otro'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "chat_tareas_adjuntos", "nombre_archivo", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "chat_tareas_adjuntos", "mime_type", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "chat_tareas_adjuntos", "tamano_bytes", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "chat_tareas_adjuntos", "duracion_segundos", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "chat_tareas_adjuntos", "fecha_actualizacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "chat_tareas_adjuntos", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "chat_tareas_adjuntos", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "chat_tareas_adjuntos", "observaciones", "TEXT"); err != nil {
		return err
	}

	if err := ensureColumnIfMissing(dbConn, "chat_tareas", "conversacion_id", "INTEGER"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "chat_tareas", "descripcion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "chat_tareas", "prioridad", "TEXT DEFAULT 'media'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "chat_tareas", "fecha_limite", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "chat_tareas", "asignado_tipo", "TEXT DEFAULT 'usuario'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "chat_tareas", "asignado_ref_id", "INTEGER"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "chat_tareas", "asignado_nombre", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "chat_tareas", "asignado_email", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "chat_tareas", "creado_por_tipo", "TEXT DEFAULT 'admin'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "chat_tareas", "creado_por_email", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "chat_tareas", "estado_tarea", "TEXT DEFAULT 'pendiente'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "chat_tareas", "porcentaje_avance", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "chat_tareas", "completada_en", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "chat_tareas", "nota_voz_url", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "chat_tareas", "nota_voz_mime_type", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "chat_tareas", "nota_voz_tamano_bytes", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "chat_tareas", "nota_voz_duracion_segundos", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "chat_tareas", "fecha_actualizacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "chat_tareas", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "chat_tareas", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "chat_tareas", "observaciones", "TEXT"); err != nil {
		return err
	}

	if err := ensureColumnIfMissing(dbConn, "chat_tareas_citas", "conversacion_id", "INTEGER"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "chat_tareas_citas", "descripcion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "chat_tareas_citas", "tipo_cita", "TEXT DEFAULT 'reunion'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "chat_tareas_citas", "fecha_inicio", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "chat_tareas_citas", "fecha_fin", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "chat_tareas_citas", "ubicacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "chat_tareas_citas", "notificar_minutos_antes", "INTEGER DEFAULT 30"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "chat_tareas_citas", "creado_por_tipo", "TEXT DEFAULT 'admin'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "chat_tareas_citas", "creado_por_ref_id", "INTEGER"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "chat_tareas_citas", "creado_por_nombre", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "chat_tareas_citas", "creado_por_email", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "chat_tareas_citas", "estado_cita", "TEXT DEFAULT 'programada'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "chat_tareas_citas", "recordatorio_enviado", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "chat_tareas_citas", "recordatorio_enviado_en", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "chat_tareas_citas", "visibilidad", "TEXT DEFAULT 'empresa'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "chat_tareas_citas", "fecha_actualizacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "chat_tareas_citas", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "chat_tareas_citas", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "chat_tareas_citas", "observaciones", "TEXT"); err != nil {
		return err
	}

	return nil
}

func normalizeChatEstado(v string) string {
	if strings.EqualFold(strings.TrimSpace(v), "inactivo") {
		return "inactivo"
	}
	return "activo"
}

func normalizeChatPrioridad(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "baja", "media", "alta", "urgente":
		return strings.ToLower(strings.TrimSpace(v))
	default:
		return "media"
	}
}

func normalizeConversacionEstado(v string) string {
	if strings.EqualFold(strings.TrimSpace(v), "cerrada") {
		return "cerrada"
	}
	return "abierta"
}

func normalizeParticipanteTipo(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "admin", "usuario", "sistema":
		return strings.ToLower(strings.TrimSpace(v))
	default:
		return "usuario"
	}
}

func normalizeAutorTipo(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "admin", "usuario", "sistema":
		return strings.ToLower(strings.TrimSpace(v))
	default:
		return "admin"
	}
}

func normalizeTipoMensaje(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "texto", "tarea", "sistema":
		return strings.ToLower(strings.TrimSpace(v))
	default:
		return "texto"
	}
}

func normalizeTipoAdjunto(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "imagen", "audio", "archivo", "otro":
		return strings.ToLower(strings.TrimSpace(v))
	default:
		return "otro"
	}
}

func normalizeEstadoTarea(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "pendiente", "en_progreso", "bloqueada", "completada", "cancelada":
		return strings.ToLower(strings.TrimSpace(v))
	default:
		return "pendiente"
	}
}

func normalizeAsignadoTipo(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "admin", "usuario", "sistema":
		return strings.ToLower(strings.TrimSpace(v))
	default:
		return "usuario"
	}
}

func normalizeEstadoCita(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "programada", "completada", "cancelada":
		return strings.ToLower(strings.TrimSpace(v))
	default:
		return "programada"
	}
}

func normalizeTipoCita(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	if v == "" {
		return "reunion"
	}
	if len(v) > 40 {
		return v[:40]
	}
	return v
}

func normalizeVisibilidadCita(v string) string {
	if strings.EqualFold(strings.TrimSpace(v), "privada") {
		return "privada"
	}
	return "empresa"
}

func normalizeReminderMinutes(v int) int {
	if v <= 0 {
		return 30
	}
	if v > 10080 {
		return 10080
	}
	return v
}

func clampPercent(v int) int {
	if v < 0 {
		return 0
	}
	if v > 100 {
		return 100
	}
	return v
}

// CreateChatConversacion crea una conversacion de chat por empresa.
func CreateChatConversacion(dbConn *sql.DB, payload ChatConversacion) (int64, error) {
	res, err := dbConn.Exec(`INSERT INTO chat_tareas_conversaciones (
		empresa_id,
		titulo,
		descripcion,
		prioridad,
		estado_conversacion,
		ultimo_mensaje_en,
		usuario_creador,
		estado,
		observaciones,
		fecha_creacion,
		fecha_actualizacion
	) VALUES (?, ?, ?, ?, ?, '', ?, ?, ?, datetime('now','localtime'), datetime('now','localtime'))`,
		payload.EmpresaID,
		strings.TrimSpace(payload.Titulo),
		strings.TrimSpace(payload.Descripcion),
		normalizeChatPrioridad(payload.Prioridad),
		normalizeConversacionEstado(payload.EstadoConversacion),
		strings.TrimSpace(payload.UsuarioCreador),
		normalizeChatEstado(payload.Estado),
		strings.TrimSpace(payload.Observaciones),
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// GetChatConversaciones lista conversaciones por empresa.
func GetChatConversaciones(dbConn *sql.DB, empresaID int64, includeInactive bool, q string) ([]ChatConversacion, error) {
	query := `SELECT
		c.id,
		c.empresa_id,
		COALESCE(c.titulo, ''),
		COALESCE(c.descripcion, ''),
		COALESCE(c.prioridad, 'media'),
		COALESCE(c.estado_conversacion, 'abierta'),
		COALESCE(c.ultimo_mensaje_en, ''),
		COALESCE((
			SELECT COUNT(1)
			FROM chat_tareas_mensajes m
			WHERE m.empresa_id = c.empresa_id
				AND m.conversacion_id = c.id
				AND COALESCE(m.estado, 'activo') = 'activo'
		), 0) AS mensajes_count,
		COALESCE((
			SELECT COUNT(1)
			FROM chat_tareas t
			WHERE t.empresa_id = c.empresa_id
				AND t.conversacion_id = c.id
				AND COALESCE(t.estado, 'activo') = 'activo'
				AND COALESCE(t.estado_tarea, 'pendiente') NOT IN ('completada', 'cancelada')
		), 0) AS tareas_pendientes,
		COALESCE(c.fecha_creacion, ''),
		COALESCE(c.fecha_actualizacion, ''),
		COALESCE(c.usuario_creador, ''),
		COALESCE(c.estado, 'activo'),
		COALESCE(c.observaciones, '')
	FROM chat_tareas_conversaciones c
	WHERE c.empresa_id = ?`
	args := []interface{}{empresaID}

	if !includeInactive {
		query += ` AND COALESCE(c.estado, 'activo') = 'activo'`
	}

	q = strings.TrimSpace(q)
	if q != "" {
		query += ` AND (
			lower(COALESCE(c.titulo, '')) LIKE lower(?)
			OR lower(COALESCE(c.descripcion, '')) LIKE lower(?)
		)`
		pat := "%" + q + "%"
		args = append(args, pat, pat)
	}

	query += ` ORDER BY
		CASE WHEN COALESCE(c.estado_conversacion, 'abierta') = 'abierta' THEN 0 ELSE 1 END,
		COALESCE(c.ultimo_mensaje_en, c.fecha_creacion) DESC,
		c.id DESC`

	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]ChatConversacion, 0)
	for rows.Next() {
		var item ChatConversacion
		if err := rows.Scan(
			&item.ID,
			&item.EmpresaID,
			&item.Titulo,
			&item.Descripcion,
			&item.Prioridad,
			&item.EstadoConversacion,
			&item.UltimoMensajeEn,
			&item.MensajesCount,
			&item.TareasPendientes,
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
	return out, nil
}

// UpdateChatConversacion actualiza metadata de una conversacion.
func UpdateChatConversacion(dbConn *sql.DB, payload ChatConversacion) error {
	_, err := dbConn.Exec(`UPDATE chat_tareas_conversaciones
	SET titulo = ?,
		descripcion = ?,
		prioridad = ?,
		estado_conversacion = ?,
		observaciones = ?,
		fecha_actualizacion = datetime('now','localtime')
	WHERE id = ? AND empresa_id = ?`,
		strings.TrimSpace(payload.Titulo),
		strings.TrimSpace(payload.Descripcion),
		normalizeChatPrioridad(payload.Prioridad),
		normalizeConversacionEstado(payload.EstadoConversacion),
		strings.TrimSpace(payload.Observaciones),
		payload.ID,
		payload.EmpresaID,
	)
	return err
}

// SetChatConversacionEstado activa/desactiva una conversacion.
func SetChatConversacionEstado(dbConn *sql.DB, empresaID, id int64, estado string) error {
	_, err := dbConn.Exec(`UPDATE chat_tareas_conversaciones
	SET estado = ?,
		fecha_actualizacion = datetime('now','localtime')
	WHERE empresa_id = ? AND id = ?`, normalizeChatEstado(estado), empresaID, id)
	return err
}

// SetChatConversacionOperacionEstado abre/cierra una conversacion.
func SetChatConversacionOperacionEstado(dbConn *sql.DB, empresaID, id int64, estadoConversacion string) error {
	_, err := dbConn.Exec(`UPDATE chat_tareas_conversaciones
	SET estado_conversacion = ?,
		fecha_actualizacion = datetime('now','localtime')
	WHERE empresa_id = ? AND id = ?`, normalizeConversacionEstado(estadoConversacion), empresaID, id)
	return err
}

// DeleteChatConversacion elimina una conversacion y sus dependencias.
func DeleteChatConversacion(dbConn *sql.DB, empresaID, id int64) error {
	tx, err := dbConn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM chat_tareas_adjuntos
		WHERE empresa_id = ?
			AND mensaje_id IN (
				SELECT id FROM chat_tareas_mensajes WHERE empresa_id = ? AND conversacion_id = ?
			)`, empresaID, empresaID, id); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM chat_tareas_mensajes WHERE empresa_id = ? AND conversacion_id = ?`, empresaID, id); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM chat_tareas_participantes WHERE empresa_id = ? AND conversacion_id = ?`, empresaID, id); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM chat_tareas WHERE empresa_id = ? AND conversacion_id = ?`, empresaID, id); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM chat_tareas_conversaciones WHERE empresa_id = ? AND id = ?`, empresaID, id); err != nil {
		return err
	}

	return tx.Commit()
}

// CreateChatParticipante agrega un participante a una conversacion.
func CreateChatParticipante(dbConn *sql.DB, payload ChatParticipante) (int64, error) {
	email := strings.ToLower(strings.TrimSpace(payload.Email))
	tipo := normalizeParticipanteTipo(payload.ParticipanteTipo)
	res, err := dbConn.Exec(`INSERT OR IGNORE INTO chat_tareas_participantes (
		empresa_id,
		conversacion_id,
		participante_tipo,
		participante_ref_id,
		nombre,
		email,
		activo_chat,
		usuario_creador,
		estado,
		observaciones,
		fecha_creacion,
		fecha_actualizacion
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now','localtime'), datetime('now','localtime'))`,
		payload.EmpresaID,
		payload.ConversacionID,
		tipo,
		nullableInt64(payload.ParticipanteRefID),
		strings.TrimSpace(payload.Nombre),
		email,
		1,
		strings.TrimSpace(payload.UsuarioCreador),
		normalizeChatEstado(payload.Estado),
		strings.TrimSpace(payload.Observaciones),
	)
	if err != nil {
		return 0, err
	}

	if id, err := res.LastInsertId(); err == nil && id > 0 {
		return id, nil
	}

	row := dbConn.QueryRow(`SELECT id FROM chat_tareas_participantes
	WHERE empresa_id = ?
		AND conversacion_id = ?
		AND participante_tipo = ?
		AND COALESCE(participante_ref_id, 0) = ?
		AND COALESCE(email, '') = COALESCE(?, '')
	LIMIT 1`, payload.EmpresaID, payload.ConversacionID, tipo, payload.ParticipanteRefID, email)
	var existingID int64
	if err := row.Scan(&existingID); err != nil {
		return 0, err
	}

	_, _ = dbConn.Exec(`UPDATE chat_tareas_participantes
	SET nombre = ?,
		estado = 'activo',
		activo_chat = 1,
		fecha_actualizacion = datetime('now','localtime')
	WHERE id = ? AND empresa_id = ? AND conversacion_id = ?`,
		strings.TrimSpace(payload.Nombre), existingID, payload.EmpresaID, payload.ConversacionID)

	return existingID, nil
}

// GetChatParticipantes lista participantes de una conversacion.
func GetChatParticipantes(dbConn *sql.DB, empresaID, conversacionID int64, includeInactive bool) ([]ChatParticipante, error) {
	query := `SELECT
		id,
		empresa_id,
		conversacion_id,
		COALESCE(participante_tipo, 'usuario'),
		COALESCE(participante_ref_id, 0),
		COALESCE(nombre, ''),
		COALESCE(email, ''),
		COALESCE(activo_chat, 1),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM chat_tareas_participantes
	WHERE empresa_id = ? AND conversacion_id = ?`
	args := []interface{}{empresaID, conversacionID}
	if !includeInactive {
		query += ` AND COALESCE(estado, 'activo') = 'activo'`
	}
	query += ` ORDER BY id ASC`

	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]ChatParticipante, 0)
	for rows.Next() {
		var item ChatParticipante
		if err := rows.Scan(
			&item.ID,
			&item.EmpresaID,
			&item.ConversacionID,
			&item.ParticipanteTipo,
			&item.ParticipanteRefID,
			&item.Nombre,
			&item.Email,
			&item.ActivoChat,
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
	return out, nil
}

// UpdateChatParticipante actualiza datos de un participante.
func UpdateChatParticipante(dbConn *sql.DB, payload ChatParticipante) error {
	_, err := dbConn.Exec(`UPDATE chat_tareas_participantes
	SET participante_tipo = ?,
		participante_ref_id = ?,
		nombre = ?,
		email = ?,
		activo_chat = ?,
		observaciones = ?,
		fecha_actualizacion = datetime('now','localtime')
	WHERE id = ? AND empresa_id = ? AND conversacion_id = ?`,
		normalizeParticipanteTipo(payload.ParticipanteTipo),
		nullableInt64(payload.ParticipanteRefID),
		strings.TrimSpace(payload.Nombre),
		strings.ToLower(strings.TrimSpace(payload.Email)),
		payload.ActivoChat,
		strings.TrimSpace(payload.Observaciones),
		payload.ID,
		payload.EmpresaID,
		payload.ConversacionID,
	)
	return err
}

// SetChatParticipanteEstado activa/desactiva un participante.
func SetChatParticipanteEstado(dbConn *sql.DB, empresaID, conversacionID, id int64, estado string) error {
	activoChat := 0
	if normalizeChatEstado(estado) == "activo" {
		activoChat = 1
	}
	_, err := dbConn.Exec(`UPDATE chat_tareas_participantes
	SET estado = ?,
		activo_chat = ?,
		fecha_actualizacion = datetime('now','localtime')
	WHERE empresa_id = ? AND conversacion_id = ? AND id = ?`, normalizeChatEstado(estado), activoChat, empresaID, conversacionID, id)
	return err
}

// DeleteChatParticipante elimina un participante.
func DeleteChatParticipante(dbConn *sql.DB, empresaID, conversacionID, id int64) error {
	_, err := dbConn.Exec(`DELETE FROM chat_tareas_participantes WHERE empresa_id = ? AND conversacion_id = ? AND id = ?`, empresaID, conversacionID, id)
	return err
}

// CreateChatMensaje crea un mensaje en una conversacion.
func CreateChatMensaje(dbConn *sql.DB, payload ChatMensaje) (int64, error) {
	res, err := dbConn.Exec(`INSERT INTO chat_tareas_mensajes (
		empresa_id,
		conversacion_id,
		autor_tipo,
		autor_ref_id,
		autor_nombre,
		autor_email,
		contenido,
		tipo_mensaje,
		fecha_envio,
		usuario_creador,
		estado,
		observaciones,
		fecha_creacion,
		fecha_actualizacion
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, datetime('now','localtime'), ?, ?, ?, datetime('now','localtime'), datetime('now','localtime'))`,
		payload.EmpresaID,
		payload.ConversacionID,
		normalizeAutorTipo(payload.AutorTipo),
		nullableInt64(payload.AutorRefID),
		strings.TrimSpace(payload.AutorNombre),
		strings.ToLower(strings.TrimSpace(payload.AutorEmail)),
		strings.TrimSpace(payload.Contenido),
		normalizeTipoMensaje(payload.TipoMensaje),
		strings.TrimSpace(payload.UsuarioCreador),
		normalizeChatEstado(payload.Estado),
		strings.TrimSpace(payload.Observaciones),
	)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	if err := refreshConversacionUltimoMensaje(dbConn, payload.EmpresaID, payload.ConversacionID); err != nil {
		return 0, err
	}
	return id, nil
}

func refreshConversacionUltimoMensaje(dbConn *sql.DB, empresaID, conversacionID int64) error {
	_, err := dbConn.Exec(`UPDATE chat_tareas_conversaciones
	SET ultimo_mensaje_en = (
		SELECT COALESCE(MAX(fecha_envio), '')
		FROM chat_tareas_mensajes
		WHERE empresa_id = ?
			AND conversacion_id = ?
			AND COALESCE(estado, 'activo') = 'activo'
	),
	fecha_actualizacion = datetime('now','localtime')
	WHERE empresa_id = ? AND id = ?`, empresaID, conversacionID, empresaID, conversacionID)
	return err
}

// GetChatMensajes lista mensajes de una conversacion con adjuntos.
func GetChatMensajes(dbConn *sql.DB, empresaID, conversacionID int64, includeInactive bool, limit, offset int) ([]ChatMensaje, error) {
	if limit <= 0 {
		limit = 250
	}
	if limit > 1000 {
		limit = 1000
	}
	if offset < 0 {
		offset = 0
	}

	query := `SELECT
		id,
		empresa_id,
		conversacion_id,
		COALESCE(autor_tipo, 'admin'),
		COALESCE(autor_ref_id, 0),
		COALESCE(autor_nombre, ''),
		COALESCE(autor_email, ''),
		COALESCE(contenido, ''),
		COALESCE(tipo_mensaje, 'texto'),
		COALESCE(fecha_envio, ''),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM chat_tareas_mensajes
	WHERE empresa_id = ?
		AND conversacion_id = ?`
	args := []interface{}{empresaID, conversacionID}

	if !includeInactive {
		query += ` AND COALESCE(estado, 'activo') = 'activo'`
	}

	query += ` ORDER BY id ASC LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]ChatMensaje, 0)
	ids := make([]int64, 0)
	for rows.Next() {
		var item ChatMensaje
		if err := rows.Scan(
			&item.ID,
			&item.EmpresaID,
			&item.ConversacionID,
			&item.AutorTipo,
			&item.AutorRefID,
			&item.AutorNombre,
			&item.AutorEmail,
			&item.Contenido,
			&item.TipoMensaje,
			&item.FechaEnvio,
			&item.FechaCreacion,
			&item.FechaActualizacion,
			&item.UsuarioCreador,
			&item.Estado,
			&item.Observaciones,
		); err != nil {
			return nil, err
		}
		out = append(out, item)
		ids = append(ids, item.ID)
	}

	byMessage, err := getChatAdjuntosByMensajeIDs(dbConn, empresaID, ids, includeInactive)
	if err != nil {
		return nil, err
	}
	for i := range out {
		out[i].Adjuntos = byMessage[out[i].ID]
	}

	return out, nil
}

func getChatAdjuntosByMensajeIDs(dbConn *sql.DB, empresaID int64, mensajeIDs []int64, includeInactive bool) (map[int64][]ChatAdjunto, error) {
	out := make(map[int64][]ChatAdjunto)
	if len(mensajeIDs) == 0 {
		return out, nil
	}

	placeholders := make([]string, len(mensajeIDs))
	args := make([]interface{}, 0, len(mensajeIDs)+1)
	args = append(args, empresaID)
	for i, id := range mensajeIDs {
		placeholders[i] = "?"
		args = append(args, id)
	}

	query := `SELECT
		id,
		empresa_id,
		mensaje_id,
		COALESCE(tipo_archivo, 'otro'),
		COALESCE(nombre_archivo, ''),
		COALESCE(mime_type, ''),
		COALESCE(file_url, ''),
		COALESCE(tamano_bytes, 0),
		COALESCE(duracion_segundos, 0),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM chat_tareas_adjuntos
	WHERE empresa_id = ?
		AND mensaje_id IN (` + strings.Join(placeholders, ",") + `)`
	if !includeInactive {
		query += ` AND COALESCE(estado, 'activo') = 'activo'`
	}
	query += ` ORDER BY id ASC`

	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item ChatAdjunto
		if err := rows.Scan(
			&item.ID,
			&item.EmpresaID,
			&item.MensajeID,
			&item.TipoArchivo,
			&item.NombreArchivo,
			&item.MimeType,
			&item.FileURL,
			&item.TamanoBytes,
			&item.DuracionSegundos,
			&item.FechaCreacion,
			&item.FechaActualizacion,
			&item.UsuarioCreador,
			&item.Estado,
			&item.Observaciones,
		); err != nil {
			return nil, err
		}
		out[item.MensajeID] = append(out[item.MensajeID], item)
	}
	return out, nil
}

// UpdateChatMensaje actualiza un mensaje existente.
func UpdateChatMensaje(dbConn *sql.DB, payload ChatMensaje) error {
	_, err := dbConn.Exec(`UPDATE chat_tareas_mensajes
	SET contenido = ?,
		tipo_mensaje = ?,
		observaciones = ?,
		fecha_actualizacion = datetime('now','localtime')
	WHERE id = ?
		AND empresa_id = ?
		AND conversacion_id = ?`,
		strings.TrimSpace(payload.Contenido),
		normalizeTipoMensaje(payload.TipoMensaje),
		strings.TrimSpace(payload.Observaciones),
		payload.ID,
		payload.EmpresaID,
		payload.ConversacionID,
	)
	return err
}

// SetChatMensajeEstado activa/desactiva un mensaje.
func SetChatMensajeEstado(dbConn *sql.DB, empresaID, conversacionID, id int64, estado string) error {
	if _, err := dbConn.Exec(`UPDATE chat_tareas_mensajes
	SET estado = ?,
		fecha_actualizacion = datetime('now','localtime')
	WHERE empresa_id = ? AND conversacion_id = ? AND id = ?`, normalizeChatEstado(estado), empresaID, conversacionID, id); err != nil {
		return err
	}
	return refreshConversacionUltimoMensaje(dbConn, empresaID, conversacionID)
}

// DeleteChatMensaje elimina un mensaje y sus adjuntos.
func DeleteChatMensaje(dbConn *sql.DB, empresaID, conversacionID, id int64) error {
	tx, err := dbConn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM chat_tareas_adjuntos WHERE empresa_id = ? AND mensaje_id = ?`, empresaID, id); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM chat_tareas_mensajes WHERE empresa_id = ? AND conversacion_id = ? AND id = ?`, empresaID, conversacionID, id); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return refreshConversacionUltimoMensaje(dbConn, empresaID, conversacionID)
}

// CreateChatAdjunto inserta metadata de un archivo adjunto.
func CreateChatAdjunto(dbConn *sql.DB, payload ChatAdjunto) (int64, error) {
	if strings.TrimSpace(payload.FileURL) == "" {
		return 0, fmt.Errorf("file_url es obligatorio")
	}
	res, err := dbConn.Exec(`INSERT INTO chat_tareas_adjuntos (
		empresa_id,
		mensaje_id,
		tipo_archivo,
		nombre_archivo,
		mime_type,
		file_url,
		tamano_bytes,
		duracion_segundos,
		usuario_creador,
		estado,
		observaciones,
		fecha_creacion,
		fecha_actualizacion
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now','localtime'), datetime('now','localtime'))`,
		payload.EmpresaID,
		payload.MensajeID,
		normalizeTipoAdjunto(payload.TipoArchivo),
		strings.TrimSpace(payload.NombreArchivo),
		strings.TrimSpace(payload.MimeType),
		strings.TrimSpace(payload.FileURL),
		payload.TamanoBytes,
		payload.DuracionSegundos,
		strings.TrimSpace(payload.UsuarioCreador),
		normalizeChatEstado(payload.Estado),
		strings.TrimSpace(payload.Observaciones),
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// GetChatAdjuntosByMensaje lista adjuntos por mensaje.
func GetChatAdjuntosByMensaje(dbConn *sql.DB, empresaID, mensajeID int64, includeInactive bool) ([]ChatAdjunto, error) {
	query := `SELECT
		id,
		empresa_id,
		mensaje_id,
		COALESCE(tipo_archivo, 'otro'),
		COALESCE(nombre_archivo, ''),
		COALESCE(mime_type, ''),
		COALESCE(file_url, ''),
		COALESCE(tamano_bytes, 0),
		COALESCE(duracion_segundos, 0),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM chat_tareas_adjuntos
	WHERE empresa_id = ? AND mensaje_id = ?`
	args := []interface{}{empresaID, mensajeID}
	if !includeInactive {
		query += ` AND COALESCE(estado, 'activo') = 'activo'`
	}
	query += ` ORDER BY id ASC`

	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]ChatAdjunto, 0)
	for rows.Next() {
		var item ChatAdjunto
		if err := rows.Scan(
			&item.ID,
			&item.EmpresaID,
			&item.MensajeID,
			&item.TipoArchivo,
			&item.NombreArchivo,
			&item.MimeType,
			&item.FileURL,
			&item.TamanoBytes,
			&item.DuracionSegundos,
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
	return out, nil
}

// CreateChatTarea crea una tarea.
func CreateChatTarea(dbConn *sql.DB, payload ChatTarea) (int64, error) {
	estadoTarea := normalizeEstadoTarea(payload.EstadoTarea)
	porcentaje := clampPercent(payload.PorcentajeAvance)
	if estadoTarea == "completada" {
		porcentaje = 100
	}

	res, err := dbConn.Exec(`INSERT INTO chat_tareas (
		empresa_id,
		conversacion_id,
		titulo,
		descripcion,
		prioridad,
		fecha_limite,
		asignado_tipo,
		asignado_ref_id,
		asignado_nombre,
		asignado_email,
		creado_por_tipo,
		creado_por_email,
		estado_tarea,
		porcentaje_avance,
		completada_en,
		nota_voz_url,
		nota_voz_mime_type,
		nota_voz_tamano_bytes,
		nota_voz_duracion_segundos,
		usuario_creador,
		estado,
		observaciones,
		fecha_creacion,
		fecha_actualizacion
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CASE WHEN ? = 'completada' THEN datetime('now','localtime') ELSE NULL END, ?, ?, ?, ?, ?, ?, ?, datetime('now','localtime'), datetime('now','localtime'))`,
		payload.EmpresaID,
		nullableInt64(payload.ConversacionID),
		strings.TrimSpace(payload.Titulo),
		strings.TrimSpace(payload.Descripcion),
		normalizeChatPrioridad(payload.Prioridad),
		strings.TrimSpace(payload.FechaLimite),
		normalizeAsignadoTipo(payload.AsignadoTipo),
		nullableInt64(payload.AsignadoRefID),
		strings.TrimSpace(payload.AsignadoNombre),
		strings.ToLower(strings.TrimSpace(payload.AsignadoEmail)),
		normalizeAutorTipo(payload.CreadoPorTipo),
		strings.ToLower(strings.TrimSpace(payload.CreadoPorEmail)),
		estadoTarea,
		porcentaje,
		estadoTarea,
		strings.TrimSpace(payload.NotaVozURL),
		strings.TrimSpace(payload.NotaVozMimeType),
		payload.NotaVozTamanoBytes,
		payload.NotaVozDuracionSeg,
		strings.TrimSpace(payload.UsuarioCreador),
		normalizeChatEstado(payload.Estado),
		strings.TrimSpace(payload.Observaciones),
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// GetChatTareas lista tareas por empresa con filtros opcionales.
func GetChatTareas(dbConn *sql.DB, empresaID, conversacionID int64, includeInactive bool, estadoTarea, q string) ([]ChatTarea, error) {
	query := `SELECT
		id,
		empresa_id,
		COALESCE(conversacion_id, 0),
		COALESCE(titulo, ''),
		COALESCE(descripcion, ''),
		COALESCE(prioridad, 'media'),
		COALESCE(fecha_limite, ''),
		COALESCE(asignado_tipo, 'usuario'),
		COALESCE(asignado_ref_id, 0),
		COALESCE(asignado_nombre, ''),
		COALESCE(asignado_email, ''),
		COALESCE(creado_por_tipo, 'admin'),
		COALESCE(creado_por_email, ''),
		COALESCE(estado_tarea, 'pendiente'),
		COALESCE(porcentaje_avance, 0),
		COALESCE(completada_en, ''),
		COALESCE(nota_voz_url, ''),
		COALESCE(nota_voz_mime_type, ''),
		COALESCE(nota_voz_tamano_bytes, 0),
		COALESCE(nota_voz_duracion_segundos, 0),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM chat_tareas
	WHERE empresa_id = ?`
	args := []interface{}{empresaID}

	if conversacionID > 0 {
		query += ` AND conversacion_id = ?`
		args = append(args, conversacionID)
	}
	if !includeInactive {
		query += ` AND COALESCE(estado, 'activo') = 'activo'`
	}
	if e := normalizeEstadoTarea(estadoTarea); strings.TrimSpace(estadoTarea) != "" {
		query += ` AND COALESCE(estado_tarea, 'pendiente') = ?`
		args = append(args, e)
	}
	if q = strings.TrimSpace(q); q != "" {
		query += ` AND (
			lower(COALESCE(titulo, '')) LIKE lower(?)
			OR lower(COALESCE(descripcion, '')) LIKE lower(?)
			OR lower(COALESCE(asignado_nombre, '')) LIKE lower(?)
		)`
		pat := "%" + q + "%"
		args = append(args, pat, pat, pat)
	}

	query += ` ORDER BY
		CASE COALESCE(estado_tarea, 'pendiente')
			WHEN 'pendiente' THEN 0
			WHEN 'en_progreso' THEN 1
			WHEN 'bloqueada' THEN 2
			WHEN 'completada' THEN 3
			ELSE 4
		END,
		CASE COALESCE(prioridad, 'media')
			WHEN 'urgente' THEN 0
			WHEN 'alta' THEN 1
			WHEN 'media' THEN 2
			ELSE 3
		END,
		id DESC`

	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]ChatTarea, 0)
	for rows.Next() {
		var item ChatTarea
		if err := rows.Scan(
			&item.ID,
			&item.EmpresaID,
			&item.ConversacionID,
			&item.Titulo,
			&item.Descripcion,
			&item.Prioridad,
			&item.FechaLimite,
			&item.AsignadoTipo,
			&item.AsignadoRefID,
			&item.AsignadoNombre,
			&item.AsignadoEmail,
			&item.CreadoPorTipo,
			&item.CreadoPorEmail,
			&item.EstadoTarea,
			&item.PorcentajeAvance,
			&item.CompletadaEn,
			&item.NotaVozURL,
			&item.NotaVozMimeType,
			&item.NotaVozTamanoBytes,
			&item.NotaVozDuracionSeg,
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
	return out, nil
}

// UpdateChatTarea actualiza una tarea.
func UpdateChatTarea(dbConn *sql.DB, payload ChatTarea) error {
	estadoTarea := normalizeEstadoTarea(payload.EstadoTarea)
	porcentaje := clampPercent(payload.PorcentajeAvance)
	if estadoTarea == "completada" {
		porcentaje = 100
	}

	_, err := dbConn.Exec(`UPDATE chat_tareas
	SET conversacion_id = ?,
		titulo = ?,
		descripcion = ?,
		prioridad = ?,
		fecha_limite = ?,
		asignado_tipo = ?,
		asignado_ref_id = ?,
		asignado_nombre = ?,
		asignado_email = ?,
		estado_tarea = ?,
		porcentaje_avance = ?,
		completada_en = CASE
			WHEN ? = 'completada' THEN COALESCE(completada_en, datetime('now','localtime'))
			ELSE NULL
		END,
		nota_voz_url = ?,
		nota_voz_mime_type = ?,
		nota_voz_tamano_bytes = ?,
		nota_voz_duracion_segundos = ?,
		observaciones = ?,
		fecha_actualizacion = datetime('now','localtime')
	WHERE id = ? AND empresa_id = ?`,
		nullableInt64(payload.ConversacionID),
		strings.TrimSpace(payload.Titulo),
		strings.TrimSpace(payload.Descripcion),
		normalizeChatPrioridad(payload.Prioridad),
		strings.TrimSpace(payload.FechaLimite),
		normalizeAsignadoTipo(payload.AsignadoTipo),
		nullableInt64(payload.AsignadoRefID),
		strings.TrimSpace(payload.AsignadoNombre),
		strings.ToLower(strings.TrimSpace(payload.AsignadoEmail)),
		estadoTarea,
		porcentaje,
		estadoTarea,
		strings.TrimSpace(payload.NotaVozURL),
		strings.TrimSpace(payload.NotaVozMimeType),
		payload.NotaVozTamanoBytes,
		payload.NotaVozDuracionSeg,
		strings.TrimSpace(payload.Observaciones),
		payload.ID,
		payload.EmpresaID,
	)
	return err
}

// SetChatTareaNotaVoz registra o reemplaza la nota de voz de una tarea.
func SetChatTareaNotaVoz(dbConn *sql.DB, empresaID, id int64, fileURL, mimeType string, tamanoBytes int64, duracionSegundos float64) error {
	_, err := dbConn.Exec(`UPDATE chat_tareas
	SET nota_voz_url = ?,
		nota_voz_mime_type = ?,
		nota_voz_tamano_bytes = ?,
		nota_voz_duracion_segundos = ?,
		fecha_actualizacion = datetime('now','localtime')
	WHERE empresa_id = ? AND id = ?`,
		strings.TrimSpace(fileURL),
		strings.TrimSpace(mimeType),
		tamanoBytes,
		duracionSegundos,
		empresaID,
		id,
	)
	return err
}

// SetChatTareaEstado activa/desactiva una tarea.
func SetChatTareaEstado(dbConn *sql.DB, empresaID, id int64, estado string) error {
	_, err := dbConn.Exec(`UPDATE chat_tareas
	SET estado = ?,
		fecha_actualizacion = datetime('now','localtime')
	WHERE empresa_id = ? AND id = ?`, normalizeChatEstado(estado), empresaID, id)
	return err
}

// SetChatTareaWorkflowEstado cambia el estado funcional de una tarea.
func SetChatTareaWorkflowEstado(dbConn *sql.DB, empresaID, id int64, estadoTarea string, porcentaje int) error {
	est := normalizeEstadoTarea(estadoTarea)
	pct := clampPercent(porcentaje)
	if est == "completada" {
		pct = 100
	}
	_, err := dbConn.Exec(`UPDATE chat_tareas
	SET estado_tarea = ?,
		porcentaje_avance = ?,
		completada_en = CASE WHEN ? = 'completada' THEN datetime('now','localtime') ELSE NULL END,
		fecha_actualizacion = datetime('now','localtime')
	WHERE empresa_id = ? AND id = ?`, est, pct, est, empresaID, id)
	return err
}

// DeleteChatTarea elimina una tarea.
func DeleteChatTarea(dbConn *sql.DB, empresaID, id int64) error {
	_, err := dbConn.Exec(`DELETE FROM chat_tareas WHERE empresa_id = ? AND id = ?`, empresaID, id)
	return err
}

// CreateChatCita crea una cita de agenda compartida por empresa.
func CreateChatCita(dbConn *sql.DB, payload ChatCita) (int64, error) {
	fechaInicio := strings.TrimSpace(payload.FechaInicio)
	if fechaInicio == "" {
		return 0, fmt.Errorf("fecha_inicio es obligatoria")
	}
	fechaFin := strings.TrimSpace(payload.FechaFin)
	if fechaFin == "" {
		fechaFin = fechaInicio
	}

	recordatorioEnviado := 0
	if payload.RecordatorioEnviado > 0 {
		recordatorioEnviado = 1
	}

	res, err := dbConn.Exec(`INSERT INTO chat_tareas_citas (
		empresa_id,
		conversacion_id,
		titulo,
		descripcion,
		tipo_cita,
		fecha_inicio,
		fecha_fin,
		ubicacion,
		notificar_minutos_antes,
		creado_por_tipo,
		creado_por_ref_id,
		creado_por_nombre,
		creado_por_email,
		estado_cita,
		recordatorio_enviado,
		recordatorio_enviado_en,
		visibilidad,
		usuario_creador,
		estado,
		observaciones,
		fecha_creacion,
		fecha_actualizacion
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CASE WHEN ? = 1 THEN datetime('now','localtime') ELSE NULL END, ?, ?, ?, ?, datetime('now','localtime'), datetime('now','localtime'))`,
		payload.EmpresaID,
		nullableInt64(payload.ConversacionID),
		strings.TrimSpace(payload.Titulo),
		strings.TrimSpace(payload.Descripcion),
		normalizeTipoCita(payload.TipoCita),
		fechaInicio,
		fechaFin,
		strings.TrimSpace(payload.Ubicacion),
		normalizeReminderMinutes(payload.NotificarMinutosAntes),
		normalizeAutorTipo(payload.CreadoPorTipo),
		nullableInt64(payload.CreadoPorRefID),
		strings.TrimSpace(payload.CreadoPorNombre),
		strings.ToLower(strings.TrimSpace(payload.CreadoPorEmail)),
		normalizeEstadoCita(payload.EstadoCita),
		recordatorioEnviado,
		recordatorioEnviado,
		normalizeVisibilidadCita(payload.Visibilidad),
		strings.TrimSpace(payload.UsuarioCreador),
		normalizeChatEstado(payload.Estado),
		strings.TrimSpace(payload.Observaciones),
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// GetChatCitas lista citas de agenda por empresa con filtros opcionales.
func GetChatCitas(dbConn *sql.DB, empresaID int64, desde, hasta string, includeInactive bool, estadoCita, q string) ([]ChatCita, error) {
	query := `SELECT
		id,
		empresa_id,
		COALESCE(conversacion_id, 0),
		COALESCE(titulo, ''),
		COALESCE(descripcion, ''),
		COALESCE(tipo_cita, 'reunion'),
		COALESCE(fecha_inicio, ''),
		COALESCE(fecha_fin, ''),
		COALESCE(ubicacion, ''),
		COALESCE(notificar_minutos_antes, 30),
		COALESCE(creado_por_tipo, 'admin'),
		COALESCE(creado_por_ref_id, 0),
		COALESCE(creado_por_nombre, ''),
		COALESCE(creado_por_email, ''),
		COALESCE(estado_cita, 'programada'),
		COALESCE(recordatorio_enviado, 0),
		COALESCE(recordatorio_enviado_en, ''),
		COALESCE(visibilidad, 'empresa'),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM chat_tareas_citas
	WHERE empresa_id = ?`
	args := []interface{}{empresaID}

	if !includeInactive {
		query += ` AND COALESCE(estado, 'activo') = 'activo'`
	}

	desde = strings.TrimSpace(desde)
	if desde != "" {
		query += ` AND COALESCE(fecha_inicio, '') >= ?`
		args = append(args, desde)
	}
	hasta = strings.TrimSpace(hasta)
	if hasta != "" {
		query += ` AND COALESCE(fecha_inicio, '') <= ?`
		args = append(args, hasta)
	}

	if e := normalizeEstadoCita(estadoCita); strings.TrimSpace(estadoCita) != "" {
		query += ` AND COALESCE(estado_cita, 'programada') = ?`
		args = append(args, e)
	}

	if q = strings.TrimSpace(q); q != "" {
		query += ` AND (
			lower(COALESCE(titulo, '')) LIKE lower(?)
			OR lower(COALESCE(descripcion, '')) LIKE lower(?)
			OR lower(COALESCE(ubicacion, '')) LIKE lower(?)
			OR lower(COALESCE(creado_por_nombre, '')) LIKE lower(?)
		)`
		pat := "%" + q + "%"
		args = append(args, pat, pat, pat, pat)
	}

	query += ` ORDER BY
		COALESCE(fecha_inicio, '') ASC,
		id ASC`

	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]ChatCita, 0)
	for rows.Next() {
		var item ChatCita
		if err := rows.Scan(
			&item.ID,
			&item.EmpresaID,
			&item.ConversacionID,
			&item.Titulo,
			&item.Descripcion,
			&item.TipoCita,
			&item.FechaInicio,
			&item.FechaFin,
			&item.Ubicacion,
			&item.NotificarMinutosAntes,
			&item.CreadoPorTipo,
			&item.CreadoPorRefID,
			&item.CreadoPorNombre,
			&item.CreadoPorEmail,
			&item.EstadoCita,
			&item.RecordatorioEnviado,
			&item.RecordatorioEnviadoEn,
			&item.Visibilidad,
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
	return out, nil
}

// UpdateChatCita actualiza una cita existente.
func UpdateChatCita(dbConn *sql.DB, payload ChatCita) error {
	fechaInicio := strings.TrimSpace(payload.FechaInicio)
	if fechaInicio == "" {
		return fmt.Errorf("fecha_inicio es obligatoria")
	}
	fechaFin := strings.TrimSpace(payload.FechaFin)
	if fechaFin == "" {
		fechaFin = fechaInicio
	}

	_, err := dbConn.Exec(`UPDATE chat_tareas_citas
	SET conversacion_id = ?,
		titulo = ?,
		descripcion = ?,
		tipo_cita = ?,
		fecha_inicio = ?,
		fecha_fin = ?,
		ubicacion = ?,
		notificar_minutos_antes = ?,
		estado_cita = ?,
		recordatorio_enviado = ?,
		recordatorio_enviado_en = CASE WHEN ? = 1 THEN COALESCE(recordatorio_enviado_en, datetime('now','localtime')) ELSE NULL END,
		visibilidad = ?,
		observaciones = ?,
		fecha_actualizacion = datetime('now','localtime')
	WHERE id = ? AND empresa_id = ?`,
		nullableInt64(payload.ConversacionID),
		strings.TrimSpace(payload.Titulo),
		strings.TrimSpace(payload.Descripcion),
		normalizeTipoCita(payload.TipoCita),
		fechaInicio,
		fechaFin,
		strings.TrimSpace(payload.Ubicacion),
		normalizeReminderMinutes(payload.NotificarMinutosAntes),
		normalizeEstadoCita(payload.EstadoCita),
		func() int {
			if payload.RecordatorioEnviado > 0 {
				return 1
			}
			return 0
		}(),
		func() int {
			if payload.RecordatorioEnviado > 0 {
				return 1
			}
			return 0
		}(),
		normalizeVisibilidadCita(payload.Visibilidad),
		strings.TrimSpace(payload.Observaciones),
		payload.ID,
		payload.EmpresaID,
	)
	return err
}

// SetChatCitaEstado activa/desactiva una cita.
func SetChatCitaEstado(dbConn *sql.DB, empresaID, id int64, estado string) error {
	_, err := dbConn.Exec(`UPDATE chat_tareas_citas
	SET estado = ?,
		fecha_actualizacion = datetime('now','localtime')
	WHERE empresa_id = ? AND id = ?`, normalizeChatEstado(estado), empresaID, id)
	return err
}

// SetChatCitaWorkflowEstado cambia el estado operativo de la cita.
func SetChatCitaWorkflowEstado(dbConn *sql.DB, empresaID, id int64, estadoCita string) error {
	estado := normalizeEstadoCita(estadoCita)
	_, err := dbConn.Exec(`UPDATE chat_tareas_citas
	SET estado_cita = ?,
		recordatorio_enviado = CASE WHEN ? = 'programada' THEN 0 ELSE recordatorio_enviado END,
		recordatorio_enviado_en = CASE WHEN ? = 'programada' THEN NULL ELSE recordatorio_enviado_en END,
		fecha_actualizacion = datetime('now','localtime')
	WHERE empresa_id = ? AND id = ?`, estado, estado, estado, empresaID, id)
	return err
}

// SetChatCitaReminderSent marca el recordatorio de una cita como enviado/no enviado.
func SetChatCitaReminderSent(dbConn *sql.DB, empresaID, id int64, sent bool) error {
	recordatorio := 0
	if sent {
		recordatorio = 1
	}
	_, err := dbConn.Exec(`UPDATE chat_tareas_citas
	SET recordatorio_enviado = ?,
		recordatorio_enviado_en = CASE WHEN ? = 1 THEN datetime('now','localtime') ELSE NULL END,
		fecha_actualizacion = datetime('now','localtime')
	WHERE empresa_id = ? AND id = ?`, recordatorio, recordatorio, empresaID, id)
	return err
}

// DeleteChatCita elimina una cita.
func DeleteChatCita(dbConn *sql.DB, empresaID, id int64) error {
	_, err := dbConn.Exec(`DELETE FROM chat_tareas_citas WHERE empresa_id = ? AND id = ?`, empresaID, id)
	return err
}
