package db

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

const (
	SuperCorreoNotificacionTipoConfirmacion   = "confirmacion_correo_usuario_empresa"
	SuperCorreoNotificacionTipoRecuperacion   = "recuperacion_password_usuario_empresa"
	SuperCorreoNotificacionTipoInicioServidor = "inicio_servidor"
)

// SuperCorreoNotificacionPrueba representa una notificacion de correo capturada en entorno de pruebas.
type SuperCorreoNotificacionPrueba struct {
	ID                 int64  `json:"id"`
	Tipo               string `json:"tipo"`
	EmpresaID          int64  `json:"empresa_id"`
	Destinatario       string `json:"destinatario"`
	Asunto             string `json:"asunto"`
	Cuerpo             string `json:"cuerpo"`
	TokenRef           string `json:"token_ref,omitempty"`
	MetadataJSON       string `json:"metadata_json,omitempty"`
	FechaEvento        string `json:"fecha_evento,omitempty"`
	FechaCreacion      string `json:"fecha_creacion,omitempty"`
	FechaActualizacion string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string `json:"usuario_creador,omitempty"`
	Estado             string `json:"estado,omitempty"`
	Observaciones      string `json:"observaciones,omitempty"`
}

// SuperCorreoNotificacionPruebaFilter define filtros de consulta para notificaciones capturadas.
type SuperCorreoNotificacionPruebaFilter struct {
	EmpresaID    int64
	Tipo         string
	Destinatario string
	Limit        int
	Offset       int
}

// EnsureSuperCorreoNotificacionesPruebaSchema crea/migra la tabla de notificaciones de correo en modo pruebas.
func EnsureSuperCorreoNotificacionesPruebaSchema(dbConn *sql.DB) error {
	if dbConn == nil {
		return fmt.Errorf("db connection is required")
	}

	stmts := []string{
		`CREATE TABLE IF NOT EXISTS super_correo_notificaciones_prueba (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tipo TEXT NOT NULL,
			empresa_id INTEGER,
			destinatario TEXT NOT NULL,
			asunto TEXT,
			cuerpo TEXT,
			token_ref TEXT,
			metadata_json TEXT,
			fecha_evento TEXT DEFAULT (datetime('now','localtime')),
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'capturado',
			observaciones TEXT
		);`,
		`CREATE INDEX IF NOT EXISTS ix_super_correo_notificaciones_prueba_empresa_tipo_fecha ON super_correo_notificaciones_prueba(empresa_id, tipo, fecha_evento DESC);`,
		`CREATE INDEX IF NOT EXISTS ix_super_correo_notificaciones_prueba_destinatario ON super_correo_notificaciones_prueba(destinatario);`,
	}
	for _, stmt := range stmts {
		if _, err := dbConn.Exec(stmt); err != nil {
			return err
		}
	}

	if err := ensureColumnIfMissing(dbConn, "super_correo_notificaciones_prueba", "tipo", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "super_correo_notificaciones_prueba", "empresa_id", "INTEGER"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "super_correo_notificaciones_prueba", "destinatario", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "super_correo_notificaciones_prueba", "asunto", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "super_correo_notificaciones_prueba", "cuerpo", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "super_correo_notificaciones_prueba", "token_ref", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "super_correo_notificaciones_prueba", "metadata_json", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "super_correo_notificaciones_prueba", "fecha_evento", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "super_correo_notificaciones_prueba", "fecha_creacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "super_correo_notificaciones_prueba", "fecha_actualizacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "super_correo_notificaciones_prueba", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "super_correo_notificaciones_prueba", "estado", "TEXT DEFAULT 'capturado'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "super_correo_notificaciones_prueba", "observaciones", "TEXT"); err != nil {
		return err
	}

	return nil
}

// CreateSuperCorreoNotificacionPrueba registra una notificacion de correo capturada para entorno de pruebas.
func CreateSuperCorreoNotificacionPrueba(dbConn *sql.DB, payload SuperCorreoNotificacionPrueba) (int64, error) {
	if dbConn == nil {
		return 0, fmt.Errorf("db connection is required")
	}
	if err := EnsureSuperCorreoNotificacionesPruebaSchema(dbConn); err != nil {
		return 0, err
	}

	payload.Tipo = strings.TrimSpace(payload.Tipo)
	payload.Destinatario = strings.TrimSpace(payload.Destinatario)
	payload.Asunto = strings.TrimSpace(payload.Asunto)
	payload.Cuerpo = strings.TrimSpace(payload.Cuerpo)
	payload.TokenRef = strings.TrimSpace(payload.TokenRef)
	payload.MetadataJSON = strings.TrimSpace(payload.MetadataJSON)
	payload.UsuarioCreador = strings.TrimSpace(payload.UsuarioCreador)
	payload.Observaciones = strings.TrimSpace(payload.Observaciones)
	payload.Estado = strings.TrimSpace(payload.Estado)
	if payload.Estado == "" {
		payload.Estado = "capturado"
	}
	if payload.Tipo == "" {
		return 0, fmt.Errorf("tipo is required")
	}
	if payload.Destinatario == "" {
		return 0, fmt.Errorf("destinatario is required")
	}
	if payload.FechaEvento == "" {
		payload.FechaEvento = time.Now().Format("2006-01-02 15:04:05")
	}

	res, err := dbConn.Exec(`INSERT INTO super_correo_notificaciones_prueba (
		tipo,
		empresa_id,
		destinatario,
		asunto,
		cuerpo,
		token_ref,
		metadata_json,
		fecha_evento,
		usuario_creador,
		estado,
		observaciones,
		fecha_creacion,
		fecha_actualizacion
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now','localtime'), datetime('now','localtime'))`,
		payload.Tipo,
		payload.EmpresaID,
		payload.Destinatario,
		payload.Asunto,
		payload.Cuerpo,
		payload.TokenRef,
		payload.MetadataJSON,
		payload.FechaEvento,
		payload.UsuarioCreador,
		payload.Estado,
		payload.Observaciones,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// ListSuperCorreoNotificacionesPrueba lista notificaciones capturadas en modo pruebas.
func ListSuperCorreoNotificacionesPrueba(dbConn *sql.DB, filter SuperCorreoNotificacionPruebaFilter) ([]SuperCorreoNotificacionPrueba, error) {
	if dbConn == nil {
		return nil, fmt.Errorf("db connection is required")
	}
	if err := EnsureSuperCorreoNotificacionesPruebaSchema(dbConn); err != nil {
		return nil, err
	}

	query := `SELECT
		id,
		COALESCE(tipo, ''),
		COALESCE(empresa_id, 0),
		COALESCE(destinatario, ''),
		COALESCE(asunto, ''),
		COALESCE(cuerpo, ''),
		COALESCE(token_ref, ''),
		COALESCE(metadata_json, ''),
		COALESCE(fecha_evento, ''),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'capturado'),
		COALESCE(observaciones, '')
	FROM super_correo_notificaciones_prueba
	WHERE 1=1`
	args := make([]interface{}, 0)

	if filter.EmpresaID > 0 {
		query += ` AND empresa_id = ?`
		args = append(args, filter.EmpresaID)
	}
	if tipo := strings.TrimSpace(filter.Tipo); tipo != "" {
		query += ` AND lower(tipo) = lower(?)`
		args = append(args, tipo)
	}
	if destinatario := strings.TrimSpace(filter.Destinatario); destinatario != "" {
		query += ` AND lower(destinatario) = lower(?)`
		args = append(args, destinatario)
	}

	query += ` ORDER BY id DESC`
	if filter.Limit <= 0 {
		filter.Limit = 100
	}
	if filter.Limit > 500 {
		filter.Limit = 500
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}
	query += ` LIMIT ? OFFSET ?`
	args = append(args, filter.Limit, filter.Offset)

	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]SuperCorreoNotificacionPrueba, 0)
	for rows.Next() {
		var item SuperCorreoNotificacionPrueba
		if err := rows.Scan(
			&item.ID,
			&item.Tipo,
			&item.EmpresaID,
			&item.Destinatario,
			&item.Asunto,
			&item.Cuerpo,
			&item.TokenRef,
			&item.MetadataJSON,
			&item.FechaEvento,
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
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
