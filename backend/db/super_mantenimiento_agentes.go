package db

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

const SuperAgenteDIANNoticiasCodigo = "dian_noticias"

type SuperMantenimientoAgente struct {
	ID                  int64  `json:"id"`
	Codigo              string `json:"codigo"`
	Nombre              string `json:"nombre"`
	Descripcion         string `json:"descripcion"`
	Habilitado          bool   `json:"habilitado"`
	HoraEjecucion       string `json:"hora_ejecucion"`
	EmailNotificacion   string `json:"email_notificacion"`
	UltimaEjecucion     string `json:"ultima_ejecucion"`
	UltimaEjecucionDia  string `json:"ultima_ejecucion_dia"`
	ProximaEjecucion    string `json:"proxima_ejecucion"`
	EstadoUltimaLectura string `json:"estado_ultima_lectura"`
	Observaciones       string `json:"observaciones"`
	FechaCreacion       string `json:"fecha_creacion"`
	FechaActualizacion  string `json:"fecha_actualizacion"`
	UsuarioCreador      string `json:"usuario_creador"`
	Estado              string `json:"estado"`
}

type SuperMantenimientoHallazgo struct {
	ID                 int64  `json:"id"`
	AgenteCodigo       string `json:"agente_codigo"`
	Titulo             string `json:"titulo"`
	URL                string `json:"url"`
	Fuente             string `json:"fuente"`
	Resumen            string `json:"resumen"`
	ImpactoSistema     string `json:"impacto_sistema"`
	Relevancia         string `json:"relevancia"`
	FechaPublicacion   string `json:"fecha_publicacion"`
	FechaDetectado     string `json:"fecha_detectado"`
	NotificadoEmail    string `json:"notificado_email"`
	NotificadoEn       string `json:"notificado_en"`
	HashContenido      string `json:"hash_contenido"`
	MetadataJSON       string `json:"metadata_json"`
	Estado             string `json:"estado"`
	FechaCreacion      string `json:"fecha_creacion"`
	FechaActualizacion string `json:"fecha_actualizacion"`
	UsuarioCreador     string `json:"usuario_creador"`
}

func EnsureSuperMantenimientoAgentesSchema(dbConn *sql.DB) error {
	if dbConn == nil {
		return nil
	}
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS super_mantenimiento_agentes (
			id BIGSERIAL PRIMARY KEY,
			codigo TEXT NOT NULL UNIQUE,
			nombre TEXT NOT NULL,
			descripcion TEXT,
			habilitado INTEGER DEFAULT 0,
			hora_ejecucion TEXT DEFAULT '07:00',
			email_notificacion TEXT,
			ultima_ejecucion TEXT,
			ultima_ejecucion_dia TEXT,
			proxima_ejecucion TEXT,
			estado_ultima_lectura TEXT,
			observaciones TEXT,
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo'
		);`,
		`CREATE TABLE IF NOT EXISTS super_mantenimiento_agente_hallazgos (
			id BIGSERIAL PRIMARY KEY,
			agente_codigo TEXT NOT NULL,
			titulo TEXT NOT NULL,
			url TEXT,
			fuente TEXT,
			resumen TEXT,
			impacto_sistema TEXT,
			relevancia TEXT DEFAULT 'media',
			fecha_publicacion TEXT,
			fecha_detectado TEXT DEFAULT (CURRENT_TIMESTAMP),
			notificado_email TEXT,
			notificado_en TEXT,
			hash_contenido TEXT NOT NULL UNIQUE,
			metadata_json TEXT,
			estado TEXT DEFAULT 'nuevo',
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT
		);`,
		`CREATE INDEX IF NOT EXISTS ix_super_mantenimiento_hallazgos_agente_fecha ON super_mantenimiento_agente_hallazgos(agente_codigo, fecha_detectado DESC, id DESC);`,
		`CREATE INDEX IF NOT EXISTS ix_super_mantenimiento_agentes_estado ON super_mantenimiento_agentes(estado, habilitado);`,
	}
	for _, stmt := range stmts {
		if _, err := execSQLCompat(dbConn, stmt); err != nil {
			return err
		}
	}
	_, err := execSQLCompat(dbConn, `INSERT INTO super_mantenimiento_agentes
		(codigo, nombre, descripcion, habilitado, hora_ejecucion, estado, fecha_actualizacion)
		VALUES (?, ?, ?, 0, ?, 'activo', CURRENT_TIMESTAMP)
		ON CONFLICT (codigo) DO NOTHING`,
		SuperAgenteDIANNoticiasCodigo,
		"Noticias DIAN",
		"Revisa noticias oficiales de la DIAN y alerta cambios que puedan afectar facturacion electronica, impuestos o integraciones.",
		"07:00")
	return err
}

func UpsertSuperMantenimientoAgente(dbConn *sql.DB, item SuperMantenimientoAgente) error {
	if dbConn == nil {
		return nil
	}
	item.Codigo = strings.TrimSpace(item.Codigo)
	if item.Codigo == "" {
		return fmt.Errorf("codigo de agente obligatorio")
	}
	if strings.TrimSpace(item.Nombre) == "" {
		item.Nombre = item.Codigo
	}
	if strings.TrimSpace(item.HoraEjecucion) == "" {
		item.HoraEjecucion = "07:00"
	}
	if strings.TrimSpace(item.Estado) == "" {
		item.Estado = "activo"
	}
	_, err := execSQLCompat(dbConn, `INSERT INTO super_mantenimiento_agentes
		(codigo, nombre, descripcion, habilitado, hora_ejecucion, email_notificacion, observaciones, usuario_creador, estado, fecha_actualizacion)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT (codigo) DO UPDATE SET
			nombre = EXCLUDED.nombre,
			descripcion = EXCLUDED.descripcion,
			habilitado = EXCLUDED.habilitado,
			hora_ejecucion = EXCLUDED.hora_ejecucion,
			email_notificacion = EXCLUDED.email_notificacion,
			observaciones = EXCLUDED.observaciones,
			usuario_creador = EXCLUDED.usuario_creador,
			estado = EXCLUDED.estado,
			fecha_actualizacion = CURRENT_TIMESTAMP`,
		item.Codigo, item.Nombre, item.Descripcion, boolToDBInt(item.Habilitado), item.HoraEjecucion, item.EmailNotificacion, item.Observaciones, item.UsuarioCreador, item.Estado)
	return err
}

func GetSuperMantenimientoAgente(dbConn *sql.DB, codigo string) (*SuperMantenimientoAgente, error) {
	if err := EnsureSuperMantenimientoAgentesSchema(dbConn); err != nil {
		return nil, err
	}
	row := queryRowSQLCompat(dbConn, `SELECT id, codigo, COALESCE(nombre,''), COALESCE(descripcion,''), COALESCE(habilitado,0),
		COALESCE(hora_ejecucion,''), COALESCE(email_notificacion,''), COALESCE(ultima_ejecucion,''), COALESCE(ultima_ejecucion_dia,''),
		COALESCE(proxima_ejecucion,''), COALESCE(estado_ultima_lectura,''), COALESCE(observaciones,''), COALESCE(fecha_creacion,''),
		COALESCE(fecha_actualizacion,''), COALESCE(usuario_creador,''), COALESCE(estado,'activo')
		FROM super_mantenimiento_agentes WHERE codigo = ? LIMIT 1`, strings.TrimSpace(codigo))
	var item SuperMantenimientoAgente
	var enabled int
	if err := row.Scan(&item.ID, &item.Codigo, &item.Nombre, &item.Descripcion, &enabled, &item.HoraEjecucion, &item.EmailNotificacion, &item.UltimaEjecucion, &item.UltimaEjecucionDia, &item.ProximaEjecucion, &item.EstadoUltimaLectura, &item.Observaciones, &item.FechaCreacion, &item.FechaActualizacion, &item.UsuarioCreador, &item.Estado); err != nil {
		return nil, err
	}
	item.Habilitado = enabled != 0
	return &item, nil
}

func ListSuperMantenimientoAgentes(dbConn *sql.DB) ([]SuperMantenimientoAgente, error) {
	if err := EnsureSuperMantenimientoAgentesSchema(dbConn); err != nil {
		return nil, err
	}
	rows, err := querySQLCompat(dbConn, `SELECT id, codigo, COALESCE(nombre,''), COALESCE(descripcion,''), COALESCE(habilitado,0),
		COALESCE(hora_ejecucion,''), COALESCE(email_notificacion,''), COALESCE(ultima_ejecucion,''), COALESCE(ultima_ejecucion_dia,''),
		COALESCE(proxima_ejecucion,''), COALESCE(estado_ultima_lectura,''), COALESCE(observaciones,''), COALESCE(fecha_creacion,''),
		COALESCE(fecha_actualizacion,''), COALESCE(usuario_creador,''), COALESCE(estado,'activo')
		FROM super_mantenimiento_agentes ORDER BY id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []SuperMantenimientoAgente{}
	for rows.Next() {
		var item SuperMantenimientoAgente
		var enabled int
		if err := rows.Scan(&item.ID, &item.Codigo, &item.Nombre, &item.Descripcion, &enabled, &item.HoraEjecucion, &item.EmailNotificacion, &item.UltimaEjecucion, &item.UltimaEjecucionDia, &item.ProximaEjecucion, &item.EstadoUltimaLectura, &item.Observaciones, &item.FechaCreacion, &item.FechaActualizacion, &item.UsuarioCreador, &item.Estado); err != nil {
			return nil, err
		}
		item.Habilitado = enabled != 0
		out = append(out, item)
	}
	return out, rows.Err()
}

func UpdateSuperMantenimientoAgenteRun(dbConn *sql.DB, codigo, estadoLectura, observaciones string, ranAt time.Time) error {
	if dbConn == nil {
		return nil
	}
	fecha := ranAt.Format("2006-01-02 15:04:05")
	dia := ranAt.Format("2006-01-02")
	_, err := execSQLCompat(dbConn, `UPDATE super_mantenimiento_agentes
		SET ultima_ejecucion = ?, ultima_ejecucion_dia = ?, estado_ultima_lectura = ?, observaciones = ?, fecha_actualizacion = CURRENT_TIMESTAMP
		WHERE codigo = ?`, fecha, dia, strings.TrimSpace(estadoLectura), strings.TrimSpace(observaciones), strings.TrimSpace(codigo))
	return err
}

func CreateSuperMantenimientoHallazgoIfNew(dbConn *sql.DB, item SuperMantenimientoHallazgo) (bool, int64, error) {
	if err := EnsureSuperMantenimientoAgentesSchema(dbConn); err != nil {
		return false, 0, err
	}
	item.AgenteCodigo = strings.TrimSpace(item.AgenteCodigo)
	item.HashContenido = strings.TrimSpace(item.HashContenido)
	item.Titulo = strings.TrimSpace(item.Titulo)
	if item.AgenteCodigo == "" || item.HashContenido == "" || item.Titulo == "" {
		return false, 0, fmt.Errorf("hallazgo incompleto")
	}
	if strings.TrimSpace(item.Estado) == "" {
		item.Estado = "nuevo"
	}
	res, err := execSQLCompat(dbConn, `INSERT INTO super_mantenimiento_agente_hallazgos
		(agente_codigo, titulo, url, fuente, resumen, impacto_sistema, relevancia, fecha_publicacion, fecha_detectado, notificado_email, notificado_en, hash_contenido, metadata_json, estado, usuario_creador, fecha_actualizacion)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT (hash_contenido) DO NOTHING`,
		item.AgenteCodigo, item.Titulo, item.URL, item.Fuente, item.Resumen, item.ImpactoSistema, item.Relevancia, item.FechaPublicacion, item.NotificadoEmail, item.NotificadoEn, item.HashContenido, item.MetadataJSON, item.Estado, item.UsuarioCreador)
	if err != nil {
		return false, 0, err
	}
	affected, _ := res.RowsAffected()
	var id int64
	err = queryRowSQLCompat(dbConn, `SELECT id FROM super_mantenimiento_agente_hallazgos WHERE hash_contenido = ? LIMIT 1`, item.HashContenido).Scan(&id)
	if err != nil {
		return false, 0, err
	}
	return affected > 0, id, nil
}

func ListSuperMantenimientoHallazgos(dbConn *sql.DB, agenteCodigo string, limit int) ([]SuperMantenimientoHallazgo, error) {
	if err := EnsureSuperMantenimientoAgentesSchema(dbConn); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows, err := querySQLCompat(dbConn, `SELECT id, agente_codigo, COALESCE(titulo,''), COALESCE(url,''), COALESCE(fuente,''),
		COALESCE(resumen,''), COALESCE(impacto_sistema,''), COALESCE(relevancia,''), COALESCE(fecha_publicacion,''),
		COALESCE(fecha_detectado,''), COALESCE(notificado_email,''), COALESCE(notificado_en,''), COALESCE(hash_contenido,''),
		COALESCE(metadata_json,''), COALESCE(estado,''), COALESCE(fecha_creacion,''), COALESCE(fecha_actualizacion,''), COALESCE(usuario_creador,'')
		FROM super_mantenimiento_agente_hallazgos
		WHERE agente_codigo = ?
		ORDER BY fecha_detectado DESC, id DESC
		LIMIT ?`, strings.TrimSpace(agenteCodigo), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []SuperMantenimientoHallazgo{}
	for rows.Next() {
		var item SuperMantenimientoHallazgo
		if err := rows.Scan(&item.ID, &item.AgenteCodigo, &item.Titulo, &item.URL, &item.Fuente, &item.Resumen, &item.ImpactoSistema, &item.Relevancia, &item.FechaPublicacion, &item.FechaDetectado, &item.NotificadoEmail, &item.NotificadoEn, &item.HashContenido, &item.MetadataJSON, &item.Estado, &item.FechaCreacion, &item.FechaActualizacion, &item.UsuarioCreador); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}
