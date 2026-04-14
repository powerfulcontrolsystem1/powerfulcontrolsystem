package db

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

const (
	SuperServidorEventoTipoInicio = "inicio_servidor"
)

// SuperServidorEvento representa un evento operativo de arranque/reinicio del servidor.
type SuperServidorEvento struct {
	ID                 int64  `json:"id"`
	TipoEvento         string `json:"tipo_evento"`
	Motivo             string `json:"motivo,omitempty"`
	MotivoDetalle      string `json:"motivo_detalle,omitempty"`
	OrigenArranque     string `json:"origen_arranque,omitempty"`
	Hostname           string `json:"hostname,omitempty"`
	ProcessID          int64  `json:"process_id,omitempty"`
	ListenAddr         string `json:"listen_addr,omitempty"`
	ReinicioInesperado bool   `json:"reinicio_inesperado"`
	PrevioEstado       string `json:"previo_estado,omitempty"`
	PrevioProcessID    int64  `json:"previo_process_id,omitempty"`
	PrevioInicioEn     string `json:"previo_inicio_en,omitempty"`
	PrevioFinEn        string `json:"previo_fin_en,omitempty"`
	CorreoDestino      string `json:"correo_destino,omitempty"`
	CorreoEnviado      bool   `json:"correo_enviado"`
	CorreoError        string `json:"correo_error,omitempty"`
	MetadataJSON       string `json:"metadata_json,omitempty"`
	FechaEvento        string `json:"fecha_evento,omitempty"`
	FechaCreacion      string `json:"fecha_creacion,omitempty"`
	FechaActualizacion string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string `json:"usuario_creador,omitempty"`
	Estado             string `json:"estado,omitempty"`
	Observaciones      string `json:"observaciones,omitempty"`
}

// EnsureSuperServidorEventosSchema crea/migra la tabla de eventos de arranque/reinicio de servidor.
func EnsureSuperServidorEventosSchema(dbConn *sql.DB) error {
	if dbConn == nil {
		return fmt.Errorf("db connection is required")
	}

	stmts := []string{
		`CREATE TABLE IF NOT EXISTS super_servidor_eventos (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tipo_evento TEXT NOT NULL,
			motivo TEXT,
			motivo_detalle TEXT,
			origen_arranque TEXT,
			hostname TEXT,
			process_id INTEGER,
			listen_addr TEXT,
			reinicio_inesperado INTEGER DEFAULT 0,
			previo_estado TEXT,
			previo_process_id INTEGER,
			previo_inicio_en TEXT,
			previo_fin_en TEXT,
			correo_destino TEXT,
			correo_enviado INTEGER DEFAULT 0,
			correo_error TEXT,
			metadata_json TEXT,
			fecha_evento TEXT DEFAULT (datetime('now','localtime')),
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE INDEX IF NOT EXISTS ix_super_servidor_eventos_tipo_fecha ON super_servidor_eventos(tipo_evento, fecha_evento DESC);`,
		`CREATE INDEX IF NOT EXISTS ix_super_servidor_eventos_reinicio_fecha ON super_servidor_eventos(reinicio_inesperado, fecha_evento DESC);`,
	}
	for _, stmt := range stmts {
		if _, err := execSQLCompat(dbConn, stmt); err != nil {
			return err
		}
	}

	if err := ensureColumnIfMissing(dbConn, "super_servidor_eventos", "tipo_evento", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "super_servidor_eventos", "motivo", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "super_servidor_eventos", "motivo_detalle", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "super_servidor_eventos", "origen_arranque", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "super_servidor_eventos", "hostname", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "super_servidor_eventos", "process_id", "INTEGER"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "super_servidor_eventos", "listen_addr", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "super_servidor_eventos", "reinicio_inesperado", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "super_servidor_eventos", "previo_estado", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "super_servidor_eventos", "previo_process_id", "INTEGER"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "super_servidor_eventos", "previo_inicio_en", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "super_servidor_eventos", "previo_fin_en", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "super_servidor_eventos", "correo_destino", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "super_servidor_eventos", "correo_enviado", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "super_servidor_eventos", "correo_error", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "super_servidor_eventos", "metadata_json", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "super_servidor_eventos", "fecha_evento", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "super_servidor_eventos", "fecha_creacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "super_servidor_eventos", "fecha_actualizacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "super_servidor_eventos", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "super_servidor_eventos", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "super_servidor_eventos", "observaciones", "TEXT"); err != nil {
		return err
	}

	return nil
}

// CreateSuperServidorEvento registra un evento de arranque/reinicio del servidor en superadministrador.db.
func CreateSuperServidorEvento(dbConn *sql.DB, payload SuperServidorEvento) (int64, error) {
	if dbConn == nil {
		return 0, fmt.Errorf("db connection is required")
	}
	if err := EnsureSuperServidorEventosSchema(dbConn); err != nil {
		return 0, err
	}

	payload.TipoEvento = strings.TrimSpace(payload.TipoEvento)
	payload.Motivo = strings.TrimSpace(payload.Motivo)
	payload.MotivoDetalle = strings.TrimSpace(payload.MotivoDetalle)
	payload.OrigenArranque = strings.TrimSpace(payload.OrigenArranque)
	payload.Hostname = strings.TrimSpace(payload.Hostname)
	payload.ListenAddr = strings.TrimSpace(payload.ListenAddr)
	payload.PrevioEstado = strings.TrimSpace(payload.PrevioEstado)
	payload.PrevioInicioEn = strings.TrimSpace(payload.PrevioInicioEn)
	payload.PrevioFinEn = strings.TrimSpace(payload.PrevioFinEn)
	payload.CorreoDestino = strings.TrimSpace(payload.CorreoDestino)
	payload.CorreoError = strings.TrimSpace(payload.CorreoError)
	payload.MetadataJSON = strings.TrimSpace(payload.MetadataJSON)
	payload.FechaEvento = strings.TrimSpace(payload.FechaEvento)
	payload.UsuarioCreador = strings.TrimSpace(payload.UsuarioCreador)
	payload.Estado = strings.TrimSpace(payload.Estado)
	payload.Observaciones = strings.TrimSpace(payload.Observaciones)

	if payload.TipoEvento == "" {
		payload.TipoEvento = SuperServidorEventoTipoInicio
	}
	if payload.FechaEvento == "" {
		payload.FechaEvento = time.Now().Format("2006-01-02 15:04:05")
	}
	if payload.UsuarioCreador == "" {
		payload.UsuarioCreador = "sistema"
	}
	if payload.Estado == "" {
		payload.Estado = "activo"
	}

	reinicioInesperado := 0
	if payload.ReinicioInesperado {
		reinicioInesperado = 1
	}
	correoEnviado := 0
	if payload.CorreoEnviado {
		correoEnviado = 1
	}

	nowExpr := sqlNowExpr()
	query := `INSERT INTO super_servidor_eventos (
		tipo_evento,
		motivo,
		motivo_detalle,
		origen_arranque,
		hostname,
		process_id,
		listen_addr,
		reinicio_inesperado,
		previo_estado,
		previo_process_id,
		previo_inicio_en,
		previo_fin_en,
		correo_destino,
		correo_enviado,
		correo_error,
		metadata_json,
		fecha_evento,
		fecha_creacion,
		fecha_actualizacion,
		usuario_creador,
		estado,
		observaciones
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ` + nowExpr + `, ` + nowExpr + `, ?, ?, ?)`

	return insertSQLCompat(dbConn, query,
		payload.TipoEvento,
		payload.Motivo,
		payload.MotivoDetalle,
		payload.OrigenArranque,
		payload.Hostname,
		payload.ProcessID,
		payload.ListenAddr,
		reinicioInesperado,
		payload.PrevioEstado,
		payload.PrevioProcessID,
		payload.PrevioInicioEn,
		payload.PrevioFinEn,
		payload.CorreoDestino,
		correoEnviado,
		payload.CorreoError,
		payload.MetadataJSON,
		payload.FechaEvento,
		payload.UsuarioCreador,
		payload.Estado,
		payload.Observaciones,
	)
}
