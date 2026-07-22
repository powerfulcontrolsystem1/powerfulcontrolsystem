package db

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// EmpresaHojaVidaEntidad representa un activo, paciente, moto, equipo u objeto trazable.
type EmpresaHojaVidaEntidad struct {
	ID                 int64  `json:"id"`
	EmpresaID          int64  `json:"empresa_id"`
	TipoEntidad        string `json:"tipo_entidad"`
	Codigo             string `json:"codigo"`
	Nombre             string `json:"nombre"`
	ClienteID          int64  `json:"cliente_id,omitempty"`
	ClienteNombre      string `json:"cliente_nombre,omitempty"`
	Identificacion     string `json:"identificacion,omitempty"`
	Marca              string `json:"marca,omitempty"`
	Modelo             string `json:"modelo,omitempty"`
	Serie              string `json:"serie,omitempty"`
	Color              string `json:"color,omitempty"`
	FechaNacimiento    string `json:"fecha_nacimiento,omitempty"`
	FechaIngreso       string `json:"fecha_ingreso,omitempty"`
	EstadoOperativo    string `json:"estado_operativo"`
	UltimoEventoFecha  string `json:"ultimo_evento_fecha,omitempty"`
	ProximoEventoFecha string `json:"proximo_evento_fecha,omitempty"`
	MetadataJSON       string `json:"metadata_json,omitempty"`
	FechaCreacion      string `json:"fecha_creacion,omitempty"`
	FechaActualizacion string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string `json:"usuario_creador,omitempty"`
	Estado             string `json:"estado,omitempty"`
	Observaciones      string `json:"observaciones,omitempty"`
	EventosTotal       int64  `json:"eventos_total,omitempty"`
	AlertasPendientes  int64  `json:"alertas_pendientes,omitempty"`
}

// EmpresaHojaVidaEvento registra servicios, eventos clinicos, mantenimientos, alertas cumplidas o reportes.
type EmpresaHojaVidaEvento struct {
	ID                  int64   `json:"id"`
	EmpresaID           int64   `json:"empresa_id"`
	EntidadID           int64   `json:"entidad_id"`
	TipoEvento          string  `json:"tipo_evento"`
	Titulo              string  `json:"titulo"`
	Descripcion         string  `json:"descripcion,omitempty"`
	FechaEvento         string  `json:"fecha_evento,omitempty"`
	FechaProxima        string  `json:"fecha_proxima,omitempty"`
	Costo               float64 `json:"costo,omitempty"`
	Responsable         string  `json:"responsable,omitempty"`
	DocumentoReferencia string  `json:"documento_referencia,omitempty"`
	AdjuntoURL          string  `json:"adjunto_url,omitempty"`
	Recurrente          bool    `json:"recurrente"`
	RecurrenciaDias     int64   `json:"recurrencia_dias,omitempty"`
	FechaCreacion       string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion  string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador      string  `json:"usuario_creador,omitempty"`
	Estado              string  `json:"estado,omitempty"`
	Observaciones       string  `json:"observaciones,omitempty"`
}

// EmpresaHojaVidaAlerta representa una tarea/recordatorio operativo asociado a una hoja de vida.
type EmpresaHojaVidaAlerta struct {
	ID                 int64  `json:"id"`
	EmpresaID          int64  `json:"empresa_id"`
	EntidadID          int64  `json:"entidad_id"`
	Titulo             string `json:"titulo"`
	Descripcion        string `json:"descripcion,omitempty"`
	FechaProgramada    string `json:"fecha_programada,omitempty"`
	Prioridad          string `json:"prioridad,omitempty"`
	EstadoAlerta       string `json:"estado_alerta,omitempty"`
	FechaCompletada    string `json:"fecha_completada,omitempty"`
	Responsable        string `json:"responsable,omitempty"`
	FechaCreacion      string `json:"fecha_creacion,omitempty"`
	FechaActualizacion string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string `json:"usuario_creador,omitempty"`
	Estado             string `json:"estado,omitempty"`
	Observaciones      string `json:"observaciones,omitempty"`
}

// EmpresaHojaVidaReporte resume la operacion del modulo por empresa.
type EmpresaHojaVidaReporte struct {
	EmpresaID            int64 `json:"empresa_id"`
	EntidadesActivas     int64 `json:"entidades_activas"`
	EventosActivos       int64 `json:"eventos_activos"`
	AlertasPendientes    int64 `json:"alertas_pendientes"`
	AlertasVencidas      int64 `json:"alertas_vencidas"`
	ServiciosRecurrentes int64 `json:"servicios_recurrentes"`
}

// EnsureEmpresaHojaVidaOperativaSchema crea/migra tablas del modulo universal de hoja de vida.
func EnsureEmpresaHojaVidaOperativaSchema(dbConn *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_hoja_vida_entidades (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			tipo_entidad TEXT DEFAULT 'activo',
			codigo TEXT,
			nombre TEXT NOT NULL,
			cliente_id INTEGER,
			cliente_nombre TEXT,
			identificacion TEXT,
			marca TEXT,
			modelo TEXT,
			serie TEXT,
			color TEXT,
			fecha_nacimiento TEXT,
			fecha_ingreso TEXT,
			estado_operativo TEXT DEFAULT 'activo',
			ultimo_evento_fecha TEXT,
			proximo_evento_fecha TEXT,
			metadata_json TEXT,
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_hoja_vida_entidades_empresa ON empresa_hoja_vida_entidades(empresa_id, estado, tipo_entidad);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_hoja_vida_entidades_codigo ON empresa_hoja_vida_entidades(empresa_id, codigo);`,
		`CREATE TABLE IF NOT EXISTS empresa_hoja_vida_eventos (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			entidad_id INTEGER NOT NULL,
			tipo_evento TEXT DEFAULT 'evento',
			titulo TEXT NOT NULL,
			descripcion TEXT,
			fecha_evento TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_proxima TEXT,
			costo REAL DEFAULT 0,
			responsable TEXT,
			documento_referencia TEXT,
			adjunto_url TEXT,
			recurrente INTEGER DEFAULT 0,
			recurrencia_dias INTEGER DEFAULT 0,
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_hoja_vida_eventos_entidad ON empresa_hoja_vida_eventos(empresa_id, entidad_id, fecha_evento DESC);`,
		`CREATE TABLE IF NOT EXISTS empresa_hoja_vida_alertas (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			entidad_id INTEGER NOT NULL,
			titulo TEXT NOT NULL,
			descripcion TEXT,
			fecha_programada TEXT,
			prioridad TEXT DEFAULT 'media',
			estado_alerta TEXT DEFAULT 'pendiente',
			fecha_completada TEXT,
			responsable TEXT,
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_hoja_vida_alertas_entidad ON empresa_hoja_vida_alertas(empresa_id, entidad_id, estado_alerta, fecha_programada);`,
	}
	for _, stmt := range stmts {
		if _, err := dbConn.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}

// EmpresaHojaVidaOperativaSchemaReady verifies the migration-owned tables
// without altering schema while a company operates its records.
func EmpresaHojaVidaOperativaSchemaReady(dbConn *sql.DB) error {
	if dbConn == nil {
		return fmt.Errorf("conexion de base de datos no disponible")
	}
	for _, table := range []string{
		"empresa_hoja_vida_entidades",
		"empresa_hoja_vida_eventos",
		"empresa_hoja_vida_alertas",
	} {
		var marker int
		err := dbConn.QueryRow("SELECT 1 FROM " + table + " WHERE 1=0").Scan(&marker)
		if errors.Is(err, sql.ErrNoRows) {
			continue
		}
		if err != nil {
			return fmt.Errorf("esquema de hoja de vida no disponible (%s): %w", table, err)
		}
	}
	return nil
}

func normalizeHojaVidaTipo(raw string) string {
	v := strings.ToLower(strings.TrimSpace(raw))
	if v == "" {
		return "activo"
	}
	return v
}

func normalizeHojaVidaEstadoOperativo(raw string) string {
	v := strings.ToLower(strings.TrimSpace(raw))
	if v == "" {
		return "activo"
	}
	return v
}

func hojaVidaBoolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

func hojaVidaIntToBool(v int) bool {
	return v != 0
}

// CreateEmpresaHojaVidaEntidad crea una hoja de vida.
func CreateEmpresaHojaVidaEntidad(dbConn *sql.DB, item EmpresaHojaVidaEntidad) (int64, error) {
	item.TipoEntidad = normalizeHojaVidaTipo(item.TipoEntidad)
	item.EstadoOperativo = normalizeHojaVidaEstadoOperativo(item.EstadoOperativo)
	res, err := dbConn.Exec(`INSERT INTO empresa_hoja_vida_entidades (
		empresa_id, tipo_entidad, codigo, nombre, cliente_id, cliente_nombre, identificacion,
		marca, modelo, serie, color, fecha_nacimiento, fecha_ingreso, estado_operativo,
		metadata_json, usuario_creador, estado, observaciones
	) VALUES (?, ?, ?, ?, NULLIF(?,0), ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'activo', ?)`,
		item.EmpresaID, item.TipoEntidad, strings.TrimSpace(item.Codigo), strings.TrimSpace(item.Nombre), item.ClienteID,
		strings.TrimSpace(item.ClienteNombre), strings.TrimSpace(item.Identificacion), strings.TrimSpace(item.Marca),
		strings.TrimSpace(item.Modelo), strings.TrimSpace(item.Serie), strings.TrimSpace(item.Color),
		strings.TrimSpace(item.FechaNacimiento), strings.TrimSpace(item.FechaIngreso), item.EstadoOperativo,
		strings.TrimSpace(item.MetadataJSON), strings.TrimSpace(item.UsuarioCreador), strings.TrimSpace(item.Observaciones))
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// UpdateEmpresaHojaVidaEntidad actualiza una hoja de vida.
func UpdateEmpresaHojaVidaEntidad(dbConn *sql.DB, item EmpresaHojaVidaEntidad) error {
	item.TipoEntidad = normalizeHojaVidaTipo(item.TipoEntidad)
	item.EstadoOperativo = normalizeHojaVidaEstadoOperativo(item.EstadoOperativo)
	res, err := dbConn.Exec(`UPDATE empresa_hoja_vida_entidades SET
		tipo_entidad = ?, codigo = ?, nombre = ?, cliente_id = NULLIF(?,0), cliente_nombre = ?,
		identificacion = ?, marca = ?, modelo = ?, serie = ?, color = ?, fecha_nacimiento = ?,
		fecha_ingreso = ?, estado_operativo = ?, metadata_json = ?, fecha_actualizacion = CURRENT_TIMESTAMP,
		observaciones = ?
		WHERE empresa_id = ? AND id = ? AND estado = 'activo'`,
		item.TipoEntidad, strings.TrimSpace(item.Codigo), strings.TrimSpace(item.Nombre), item.ClienteID,
		strings.TrimSpace(item.ClienteNombre), strings.TrimSpace(item.Identificacion), strings.TrimSpace(item.Marca),
		strings.TrimSpace(item.Modelo), strings.TrimSpace(item.Serie), strings.TrimSpace(item.Color),
		strings.TrimSpace(item.FechaNacimiento), strings.TrimSpace(item.FechaIngreso), item.EstadoOperativo,
		strings.TrimSpace(item.MetadataJSON), strings.TrimSpace(item.Observaciones), item.EmpresaID, item.ID)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err == nil && n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// DeleteEmpresaHojaVidaEntidad inactiva una hoja de vida.
func DeleteEmpresaHojaVidaEntidad(dbConn *sql.DB, empresaID, id int64) error {
	res, err := dbConn.Exec(`UPDATE empresa_hoja_vida_entidades SET estado = 'inactivo', fecha_actualizacion = CURRENT_TIMESTAMP WHERE empresa_id = ? AND id = ?`, empresaID, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err == nil && n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// ListEmpresaHojaVidaEntidades lista hojas de vida por empresa.
func ListEmpresaHojaVidaEntidades(dbConn *sql.DB, empresaID int64, tipo, q string, includeInactive bool, limit int) ([]EmpresaHojaVidaEntidad, error) {
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	args := []interface{}{empresaID}
	where := "WHERE e.empresa_id = ?"
	if !includeInactive {
		where += " AND e.estado = 'activo'"
	}
	if t := strings.TrimSpace(tipo); t != "" {
		where += " AND lower(e.tipo_entidad) = ?"
		args = append(args, strings.ToLower(t))
	}
	if term := strings.TrimSpace(q); term != "" {
		like := "%" + strings.ToLower(term) + "%"
		where += ` AND (lower(COALESCE(e.codigo,'')) LIKE ? OR lower(COALESCE(e.nombre,'')) LIKE ? OR lower(COALESCE(e.cliente_nombre,'')) LIKE ? OR lower(COALESCE(e.identificacion,'')) LIKE ? OR lower(COALESCE(e.marca,'')) LIKE ? OR lower(COALESCE(e.modelo,'')) LIKE ? OR lower(COALESCE(e.serie,'')) LIKE ?)`
		args = append(args, like, like, like, like, like, like, like)
	}
	args = append(args, limit)
	// #nosec G202 -- SQL structure is assembled only from server-side allowlists; all external values remain bound parameters.
	rows, err := dbConn.Query(`SELECT
		e.id, e.empresa_id, COALESCE(e.tipo_entidad,''), COALESCE(e.codigo,''), COALESCE(e.nombre,''),
		COALESCE(e.cliente_id,0), COALESCE(e.cliente_nombre,''), COALESCE(e.identificacion,''),
		COALESCE(e.marca,''), COALESCE(e.modelo,''), COALESCE(e.serie,''), COALESCE(e.color,''),
		COALESCE(e.fecha_nacimiento,''), COALESCE(e.fecha_ingreso,''), COALESCE(e.estado_operativo,''),
		COALESCE(e.ultimo_evento_fecha,''), COALESCE(e.proximo_evento_fecha,''), COALESCE(e.metadata_json,''),
		COALESCE(e.fecha_creacion,''), COALESCE(e.fecha_actualizacion,''), COALESCE(e.usuario_creador,''),
		COALESCE(e.estado,''), COALESCE(e.observaciones,''),
		(SELECT COUNT(1) FROM empresa_hoja_vida_eventos ev WHERE ev.empresa_id = e.empresa_id AND ev.entidad_id = e.id AND ev.estado = 'activo') AS eventos_total,
		(SELECT COUNT(1) FROM empresa_hoja_vida_alertas al WHERE al.empresa_id = e.empresa_id AND al.entidad_id = e.id AND al.estado = 'activo' AND al.estado_alerta = 'pendiente') AS alertas_pendientes
		FROM empresa_hoja_vida_entidades e `+where+` ORDER BY e.fecha_actualizacion DESC, e.id DESC LIMIT ?`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaHojaVidaEntidad{}
	for rows.Next() {
		var item EmpresaHojaVidaEntidad
		if err := rows.Scan(&item.ID, &item.EmpresaID, &item.TipoEntidad, &item.Codigo, &item.Nombre, &item.ClienteID, &item.ClienteNombre, &item.Identificacion, &item.Marca, &item.Modelo, &item.Serie, &item.Color, &item.FechaNacimiento, &item.FechaIngreso, &item.EstadoOperativo, &item.UltimoEventoFecha, &item.ProximoEventoFecha, &item.MetadataJSON, &item.FechaCreacion, &item.FechaActualizacion, &item.UsuarioCreador, &item.Estado, &item.Observaciones, &item.EventosTotal, &item.AlertasPendientes); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

// CreateEmpresaHojaVidaEvento crea un evento/servicio.
func CreateEmpresaHojaVidaEvento(dbConn *sql.DB, item EmpresaHojaVidaEvento) (int64, error) {
	if strings.TrimSpace(item.TipoEvento) == "" {
		item.TipoEvento = "evento"
	}
	res, err := dbConn.Exec(`INSERT INTO empresa_hoja_vida_eventos (
		empresa_id, entidad_id, tipo_evento, titulo, descripcion, fecha_evento, fecha_proxima,
		costo, responsable, documento_referencia, adjunto_url, recurrente, recurrencia_dias,
		usuario_creador, estado, observaciones
	) VALUES (?, ?, ?, ?, ?, COALESCE(NULLIF(?,''), CURRENT_TIMESTAMP), ?, ?, ?, ?, ?, ?, ?, ?, 'activo', ?)`,
		item.EmpresaID, item.EntidadID, strings.TrimSpace(item.TipoEvento), strings.TrimSpace(item.Titulo),
		strings.TrimSpace(item.Descripcion), strings.TrimSpace(item.FechaEvento), strings.TrimSpace(item.FechaProxima),
		item.Costo, strings.TrimSpace(item.Responsable), strings.TrimSpace(item.DocumentoReferencia),
		strings.TrimSpace(item.AdjuntoURL), hojaVidaBoolToInt(item.Recurrente), item.RecurrenciaDias,
		strings.TrimSpace(item.UsuarioCreador), strings.TrimSpace(item.Observaciones))
	if err != nil {
		return 0, err
	}
	_ = refreshEmpresaHojaVidaEntidadDates(dbConn, item.EmpresaID, item.EntidadID)
	return res.LastInsertId()
}

// ListEmpresaHojaVidaEventos lista eventos por entidad.
func ListEmpresaHojaVidaEventos(dbConn *sql.DB, empresaID, entidadID int64, limit int) ([]EmpresaHojaVidaEvento, error) {
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	rows, err := dbConn.Query(`SELECT id, empresa_id, entidad_id, COALESCE(tipo_evento,''), COALESCE(titulo,''), COALESCE(descripcion,''), COALESCE(fecha_evento,''), COALESCE(fecha_proxima,''), COALESCE(costo,0), COALESCE(responsable,''), COALESCE(documento_referencia,''), COALESCE(adjunto_url,''), COALESCE(recurrente,0), COALESCE(recurrencia_dias,0), COALESCE(fecha_creacion,''), COALESCE(fecha_actualizacion,''), COALESCE(usuario_creador,''), COALESCE(estado,''), COALESCE(observaciones,'') FROM empresa_hoja_vida_eventos WHERE empresa_id = ? AND entidad_id = ? AND estado = 'activo' ORDER BY fecha_evento DESC, id DESC LIMIT ?`, empresaID, entidadID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaHojaVidaEvento{}
	for rows.Next() {
		var item EmpresaHojaVidaEvento
		var recurrente int
		if err := rows.Scan(&item.ID, &item.EmpresaID, &item.EntidadID, &item.TipoEvento, &item.Titulo, &item.Descripcion, &item.FechaEvento, &item.FechaProxima, &item.Costo, &item.Responsable, &item.DocumentoReferencia, &item.AdjuntoURL, &recurrente, &item.RecurrenciaDias, &item.FechaCreacion, &item.FechaActualizacion, &item.UsuarioCreador, &item.Estado, &item.Observaciones); err != nil {
			return nil, err
		}
		item.Recurrente = hojaVidaIntToBool(recurrente)
		out = append(out, item)
	}
	return out, rows.Err()
}

// DeleteEmpresaHojaVidaEvento inactiva un evento.
func DeleteEmpresaHojaVidaEvento(dbConn *sql.DB, empresaID, id int64) error {
	var entidadID int64
	_ = dbConn.QueryRow(`SELECT entidad_id FROM empresa_hoja_vida_eventos WHERE empresa_id = ? AND id = ?`, empresaID, id).Scan(&entidadID)
	res, err := dbConn.Exec(`UPDATE empresa_hoja_vida_eventos SET estado = 'inactivo', fecha_actualizacion = CURRENT_TIMESTAMP WHERE empresa_id = ? AND id = ?`, empresaID, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err == nil && n == 0 {
		return sql.ErrNoRows
	}
	if entidadID > 0 {
		_ = refreshEmpresaHojaVidaEntidadDates(dbConn, empresaID, entidadID)
	}
	return nil
}

// CreateEmpresaHojaVidaAlerta crea una alerta.
func CreateEmpresaHojaVidaAlerta(dbConn *sql.DB, item EmpresaHojaVidaAlerta) (int64, error) {
	if strings.TrimSpace(item.Prioridad) == "" {
		item.Prioridad = "media"
	}
	if strings.TrimSpace(item.EstadoAlerta) == "" {
		item.EstadoAlerta = "pendiente"
	}
	res, err := dbConn.Exec(`INSERT INTO empresa_hoja_vida_alertas (
		empresa_id, entidad_id, titulo, descripcion, fecha_programada, prioridad,
		estado_alerta, responsable, usuario_creador, estado, observaciones
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 'activo', ?)`,
		item.EmpresaID, item.EntidadID, strings.TrimSpace(item.Titulo), strings.TrimSpace(item.Descripcion),
		strings.TrimSpace(item.FechaProgramada), strings.TrimSpace(item.Prioridad), strings.TrimSpace(item.EstadoAlerta),
		strings.TrimSpace(item.Responsable), strings.TrimSpace(item.UsuarioCreador), strings.TrimSpace(item.Observaciones))
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// ListEmpresaHojaVidaAlertas lista alertas por entidad o empresa.
func ListEmpresaHojaVidaAlertas(dbConn *sql.DB, empresaID, entidadID int64, estadoAlerta string, limit int) ([]EmpresaHojaVidaAlerta, error) {
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	args := []interface{}{empresaID}
	where := "WHERE empresa_id = ? AND estado = 'activo'"
	if entidadID > 0 {
		where += " AND entidad_id = ?"
		args = append(args, entidadID)
	}
	if estadoAlerta = strings.TrimSpace(estadoAlerta); estadoAlerta != "" {
		where += " AND estado_alerta = ?"
		args = append(args, estadoAlerta)
	}
	args = append(args, limit)
	// #nosec G202 -- SQL structure is assembled only from server-side allowlists; all external values remain bound parameters.
	rows, err := dbConn.Query(`SELECT id, empresa_id, entidad_id, COALESCE(titulo,''), COALESCE(descripcion,''), COALESCE(fecha_programada,''), COALESCE(prioridad,''), COALESCE(estado_alerta,''), COALESCE(fecha_completada,''), COALESCE(responsable,''), COALESCE(fecha_creacion,''), COALESCE(fecha_actualizacion,''), COALESCE(usuario_creador,''), COALESCE(estado,''), COALESCE(observaciones,'') FROM empresa_hoja_vida_alertas `+where+` ORDER BY CASE estado_alerta WHEN 'pendiente' THEN 0 ELSE 1 END, fecha_programada ASC, id DESC LIMIT ?`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaHojaVidaAlerta{}
	for rows.Next() {
		var item EmpresaHojaVidaAlerta
		if err := rows.Scan(&item.ID, &item.EmpresaID, &item.EntidadID, &item.Titulo, &item.Descripcion, &item.FechaProgramada, &item.Prioridad, &item.EstadoAlerta, &item.FechaCompletada, &item.Responsable, &item.FechaCreacion, &item.FechaActualizacion, &item.UsuarioCreador, &item.Estado, &item.Observaciones); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

// SetEmpresaHojaVidaAlertaEstado cambia el estado de una alerta.
func SetEmpresaHojaVidaAlertaEstado(dbConn *sql.DB, empresaID, id int64, estadoAlerta string) error {
	estadoAlerta = strings.ToLower(strings.TrimSpace(estadoAlerta))
	if estadoAlerta == "" {
		estadoAlerta = "completada"
	}
	fechaCompletadaExpr := "NULL"
	if estadoAlerta == "completada" {
		fechaCompletadaExpr = "CURRENT_TIMESTAMP"
	}
	// #nosec G202 -- SQL structure is assembled only from server-side allowlists; all external values remain bound parameters.
	res, err := dbConn.Exec(`UPDATE empresa_hoja_vida_alertas SET estado_alerta = ?, fecha_completada = `+fechaCompletadaExpr+`, fecha_actualizacion = CURRENT_TIMESTAMP WHERE empresa_id = ? AND id = ? AND estado = 'activo'`, estadoAlerta, empresaID, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err == nil && n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// DeleteEmpresaHojaVidaAlerta inactiva una alerta.
func DeleteEmpresaHojaVidaAlerta(dbConn *sql.DB, empresaID, id int64) error {
	res, err := dbConn.Exec(`UPDATE empresa_hoja_vida_alertas SET estado = 'inactivo', fecha_actualizacion = CURRENT_TIMESTAMP WHERE empresa_id = ? AND id = ?`, empresaID, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err == nil && n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// GetEmpresaHojaVidaReporte devuelve conteos de control.
func GetEmpresaHojaVidaReporte(dbConn *sql.DB, empresaID int64) (EmpresaHojaVidaReporte, error) {
	out := EmpresaHojaVidaReporte{EmpresaID: empresaID}
	err := dbConn.QueryRow(`SELECT
		(SELECT COUNT(1) FROM empresa_hoja_vida_entidades WHERE empresa_id = ? AND estado = 'activo'),
		(SELECT COUNT(1) FROM empresa_hoja_vida_eventos WHERE empresa_id = ? AND estado = 'activo'),
		(SELECT COUNT(1) FROM empresa_hoja_vida_alertas WHERE empresa_id = ? AND estado = 'activo' AND estado_alerta = 'pendiente'),
		(SELECT COUNT(1) FROM empresa_hoja_vida_alertas WHERE empresa_id = ? AND estado = 'activo' AND estado_alerta = 'pendiente' AND fecha_programada IS NOT NULL AND fecha_programada < CURRENT_TIMESTAMP),
		(SELECT COUNT(1) FROM empresa_hoja_vida_eventos WHERE empresa_id = ? AND estado = 'activo' AND recurrente = 1)`,
		empresaID, empresaID, empresaID, empresaID, empresaID,
	).Scan(&out.EntidadesActivas, &out.EventosActivos, &out.AlertasPendientes, &out.AlertasVencidas, &out.ServiciosRecurrentes)
	return out, err
}

func refreshEmpresaHojaVidaEntidadDates(dbConn *sql.DB, empresaID, entidadID int64) error {
	_, err := dbConn.Exec(`UPDATE empresa_hoja_vida_entidades SET
		ultimo_evento_fecha = (SELECT MAX(fecha_evento) FROM empresa_hoja_vida_eventos WHERE empresa_id = ? AND entidad_id = ? AND estado = 'activo'),
		proximo_evento_fecha = (SELECT MIN(fecha_proxima) FROM empresa_hoja_vida_eventos WHERE empresa_id = ? AND entidad_id = ? AND estado = 'activo' AND fecha_proxima IS NOT NULL AND fecha_proxima <> ''),
		fecha_actualizacion = CURRENT_TIMESTAMP
		WHERE empresa_id = ? AND id = ?`, empresaID, entidadID, empresaID, entidadID, empresaID, entidadID)
	return err
}
