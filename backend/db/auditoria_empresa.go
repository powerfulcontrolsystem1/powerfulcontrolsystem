package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"
)

const (
	defaultEmpresaAuditoriaRetencionDias = int64(180)
	maxEmpresaAuditoriaRetencionDias     = int64(3650)
)

// EmpresaAuditoriaEvento representa un evento de auditoria por empresa.
type EmpresaAuditoriaEvento struct {
	ID                 int64  `json:"id"`
	EmpresaID          int64  `json:"empresa_id"`
	Modulo             string `json:"modulo"`
	Accion             string `json:"accion"`
	Recurso            string `json:"recurso"`
	RecursoID          int64  `json:"recurso_id"`
	MetodoHTTP         string `json:"metodo_http"`
	Endpoint           string `json:"endpoint"`
	Resultado          string `json:"resultado"`
	CodigoHTTP         int64  `json:"codigo_http"`
	RequestID          string `json:"request_id"`
	IPOrigen           string `json:"ip_origen"`
	UserAgent          string `json:"user_agent"`
	MetadataJSON       string `json:"metadata_json"`
	RetencionDias      int64  `json:"retencion_dias"`
	FechaEvento        string `json:"fecha_evento"`
	FechaExpiracion    string `json:"fecha_expiracion"`
	FechaCreacion      string `json:"fecha_creacion"`
	FechaActualizacion string `json:"fecha_actualizacion"`
	UsuarioCreador     string `json:"usuario_creador"`
	Estado             string `json:"estado"`
	Observaciones      string `json:"observaciones"`
}

// EmpresaAuditoriaEventoFilter permite consultar auditoria por filtros.
type EmpresaAuditoriaEventoFilter struct {
	Modulo          string
	Accion          string
	MetodoHTTP      string
	Recurso         string
	Endpoint        string
	Search          string
	RecursoID       int64
	CodigoHTTP      int64
	Resultado       string
	UsuarioCreador  string
	RequestID       string
	Desde           string
	Hasta           string
	IncludeInactive bool
	Limit           int
	Offset          int
}

// EnsureEmpresaAuditoriaSchema crea/ajusta el esquema de auditoria empresarial.
func EnsureEmpresaAuditoriaSchema(dbConn *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_auditoria_eventos (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			modulo TEXT NOT NULL,
			accion TEXT NOT NULL,
			recurso TEXT,
			recurso_id INTEGER,
			metodo_http TEXT,
			endpoint TEXT,
			resultado TEXT DEFAULT 'ok',
			codigo_http INTEGER DEFAULT 0,
			request_id TEXT,
			ip_origen TEXT,
			user_agent TEXT,
			metadata_json TEXT,
			retencion_dias INTEGER DEFAULT 180,
			fecha_evento TEXT DEFAULT (datetime('now','localtime')),
			fecha_expiracion TEXT,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_auditoria_eventos_empresa_fecha ON empresa_auditoria_eventos(empresa_id, fecha_evento DESC, id DESC);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_auditoria_eventos_empresa_modulo_accion ON empresa_auditoria_eventos(empresa_id, modulo, accion);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_auditoria_eventos_empresa_usuario ON empresa_auditoria_eventos(empresa_id, usuario_creador, fecha_evento DESC);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_auditoria_eventos_request_id ON empresa_auditoria_eventos(request_id);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_auditoria_eventos_empresa_http_resultado ON empresa_auditoria_eventos(empresa_id, codigo_http, resultado);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_auditoria_eventos_empresa_recurso_id ON empresa_auditoria_eventos(empresa_id, recurso_id, fecha_evento DESC);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_auditoria_eventos_fecha_expiracion ON empresa_auditoria_eventos(fecha_expiracion);`,
	}
	for _, stmt := range stmts {
		if _, err := dbConn.Exec(stmt); err != nil {
			return err
		}
	}

	if err := ensureColumnIfMissing(dbConn, "empresa_auditoria_eventos", "modulo", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_auditoria_eventos", "accion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_auditoria_eventos", "recurso", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_auditoria_eventos", "recurso_id", "INTEGER"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_auditoria_eventos", "metodo_http", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_auditoria_eventos", "endpoint", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_auditoria_eventos", "resultado", "TEXT DEFAULT 'ok'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_auditoria_eventos", "codigo_http", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_auditoria_eventos", "request_id", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_auditoria_eventos", "ip_origen", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_auditoria_eventos", "user_agent", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_auditoria_eventos", "metadata_json", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_auditoria_eventos", "retencion_dias", "INTEGER DEFAULT 180"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_auditoria_eventos", "fecha_evento", "TEXT DEFAULT (datetime('now','localtime'))"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_auditoria_eventos", "fecha_expiracion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_auditoria_eventos", "fecha_actualizacion", "TEXT DEFAULT (datetime('now','localtime'))"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_auditoria_eventos", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_auditoria_eventos", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_auditoria_eventos", "observaciones", "TEXT"); err != nil {
		return err
	}

	return nil
}

// CreateEmpresaAuditoriaEvento registra un evento de auditoria por empresa.
func CreateEmpresaAuditoriaEvento(dbConn *sql.DB, in EmpresaAuditoriaEvento) (int64, error) {
	if in.EmpresaID <= 0 {
		return 0, fmt.Errorf("empresa_id es obligatorio")
	}
	if err := EnsureEmpresaAuditoriaSchema(dbConn); err != nil {
		return 0, err
	}

	in.Modulo = normalizeAuditoriaValue(in.Modulo)
	if in.Modulo == "" {
		return 0, fmt.Errorf("modulo es obligatorio")
	}
	in.Accion = normalizeAuditoriaValue(in.Accion)
	if in.Accion == "" {
		in.Accion = "desconocida"
	}
	in.Recurso = sanitizeAuditoriaText(strings.TrimSpace(in.Recurso), 160)
	in.MetodoHTTP = sanitizeAuditoriaText(strings.ToUpper(strings.TrimSpace(in.MetodoHTTP)), 12)
	in.Endpoint = sanitizeAuditoriaEndpoint(in.Endpoint)
	in.Resultado = normalizeAuditoriaResultado(in.Resultado)
	if in.Resultado == "" {
		in.Resultado = "ok"
	}
	in.RequestID = sanitizeAuditoriaText(strings.TrimSpace(in.RequestID), 120)
	in.IPOrigen = sanitizeAuditoriaText(strings.TrimSpace(in.IPOrigen), 64)
	in.UserAgent = sanitizeAuditoriaText(strings.TrimSpace(in.UserAgent), 320)
	in.UsuarioCreador = sanitizeAuditoriaText(strings.TrimSpace(in.UsuarioCreador), 120)
	if in.UsuarioCreador == "" {
		in.UsuarioCreador = "sistema"
	}
	in.Observaciones = sanitizeAuditoriaText(strings.TrimSpace(in.Observaciones), 500)
	in.Estado = normalizeAuditoriaEstado(in.Estado)
	if in.Estado == "" {
		in.Estado = "activo"
	}

	metadata := strings.TrimSpace(in.MetadataJSON)
	if metadata == "" {
		metadata = `{}`
	}
	if !json.Valid([]byte(metadata)) {
		return 0, fmt.Errorf("metadata_json invalido")
	}

	retencionDias := normalizeAuditoriaRetencionDias(in.RetencionDias)
	ttlExpr := fmt.Sprintf("+%d days", retencionDias)

	res, err := dbConn.Exec(`INSERT INTO empresa_auditoria_eventos (
		empresa_id,
		modulo,
		accion,
		recurso,
		recurso_id,
		metodo_http,
		endpoint,
		resultado,
		codigo_http,
		request_id,
		ip_origen,
		user_agent,
		metadata_json,
		retencion_dias,
		fecha_evento,
		fecha_expiracion,
		fecha_creacion,
		fecha_actualizacion,
		usuario_creador,
		estado,
		observaciones
	) VALUES (
		?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?,
		COALESCE(NULLIF(?, ''), datetime('now','localtime')),
		datetime(COALESCE(NULLIF(?, ''), datetime('now','localtime')), ?),
		datetime('now','localtime'),
		datetime('now','localtime'),
		?, ?, ?
	)`,
		in.EmpresaID,
		in.Modulo,
		in.Accion,
		in.Recurso,
		in.RecursoID,
		in.MetodoHTTP,
		in.Endpoint,
		in.Resultado,
		in.CodigoHTTP,
		in.RequestID,
		in.IPOrigen,
		in.UserAgent,
		metadata,
		retencionDias,
		strings.TrimSpace(in.FechaEvento),
		strings.TrimSpace(in.FechaEvento),
		ttlExpr,
		in.UsuarioCreador,
		in.Estado,
		in.Observaciones,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// ListEmpresaAuditoriaEventos lista eventos de auditoria por empresa con filtros.
func ListEmpresaAuditoriaEventos(dbConn *sql.DB, empresaID int64, f EmpresaAuditoriaEventoFilter) ([]EmpresaAuditoriaEvento, error) {
	if empresaID <= 0 {
		return nil, fmt.Errorf("empresa_id es obligatorio")
	}
	if err := EnsureEmpresaAuditoriaSchema(dbConn); err != nil {
		return nil, err
	}

	where, args := buildEmpresaAuditoriaWhereClause(empresaID, f)

	query := `SELECT
		id,
		empresa_id,
		COALESCE(modulo, ''),
		COALESCE(accion, ''),
		COALESCE(recurso, ''),
		COALESCE(recurso_id, 0),
		COALESCE(metodo_http, ''),
		COALESCE(endpoint, ''),
		COALESCE(resultado, 'ok'),
		COALESCE(codigo_http, 0),
		COALESCE(request_id, ''),
		COALESCE(ip_origen, ''),
		COALESCE(user_agent, ''),
		COALESCE(metadata_json, '{}'),
		COALESCE(retencion_dias, 180),
		COALESCE(fecha_evento, ''),
		COALESCE(fecha_expiracion, ''),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM empresa_auditoria_eventos` + where + `
	ORDER BY COALESCE(fecha_evento, '') DESC, id DESC
	LIMIT ? OFFSET ?`
	args = append(args, normalizeAuditoriaLimit(f.Limit), normalizeAuditoriaOffset(f.Offset))

	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]EmpresaAuditoriaEvento, 0)
	for rows.Next() {
		var item EmpresaAuditoriaEvento
		if err := rows.Scan(
			&item.ID,
			&item.EmpresaID,
			&item.Modulo,
			&item.Accion,
			&item.Recurso,
			&item.RecursoID,
			&item.MetodoHTTP,
			&item.Endpoint,
			&item.Resultado,
			&item.CodigoHTTP,
			&item.RequestID,
			&item.IPOrigen,
			&item.UserAgent,
			&item.MetadataJSON,
			&item.RetencionDias,
			&item.FechaEvento,
			&item.FechaExpiracion,
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

// CountEmpresaAuditoriaEventos cuenta eventos de auditoria por empresa aplicando filtros.
func CountEmpresaAuditoriaEventos(dbConn *sql.DB, empresaID int64, f EmpresaAuditoriaEventoFilter) (int64, error) {
	if empresaID <= 0 {
		return 0, fmt.Errorf("empresa_id es obligatorio")
	}
	if err := EnsureEmpresaAuditoriaSchema(dbConn); err != nil {
		return 0, err
	}

	where, args := buildEmpresaAuditoriaWhereClause(empresaID, f)
	query := `SELECT COUNT(1) FROM empresa_auditoria_eventos` + where

	var total int64
	if err := dbConn.QueryRow(query, args...).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func buildEmpresaAuditoriaWhereClause(empresaID int64, f EmpresaAuditoriaEventoFilter) (string, []interface{}) {
	where := ` WHERE empresa_id = ?`
	args := []interface{}{empresaID}

	if !f.IncludeInactive {
		where += ` AND COALESCE(estado, 'activo') = 'activo'`
	}
	if modulo := normalizeAuditoriaValue(f.Modulo); modulo != "" {
		where += ` AND COALESCE(modulo, '') = ?`
		args = append(args, modulo)
	}
	if accion := normalizeAuditoriaValue(f.Accion); accion != "" {
		where += ` AND COALESCE(accion, '') = ?`
		args = append(args, accion)
	}
	if metodo := sanitizeAuditoriaText(strings.ToUpper(strings.TrimSpace(f.MetodoHTTP)), 12); metodo != "" {
		where += ` AND COALESCE(metodo_http, '') = ?`
		args = append(args, metodo)
	}
	if recursoLike := normalizeAuditoriaContains(f.Recurso, 160); recursoLike != "" {
		where += ` AND LOWER(COALESCE(recurso, '')) LIKE ? ESCAPE '!'`
		args = append(args, recursoLike)
	}
	if endpointLike := normalizeAuditoriaContains(f.Endpoint, 300); endpointLike != "" {
		where += ` AND LOWER(COALESCE(endpoint, '')) LIKE ? ESCAPE '!'`
		args = append(args, endpointLike)
	}
	if recursoID := f.RecursoID; recursoID > 0 {
		where += ` AND COALESCE(recurso_id, 0) = ?`
		args = append(args, recursoID)
	}
	if codigoHTTP := f.CodigoHTTP; codigoHTTP > 0 {
		where += ` AND COALESCE(codigo_http, 0) = ?`
		args = append(args, codigoHTTP)
	}
	if resultado := normalizeAuditoriaResultado(f.Resultado); resultado != "" {
		where += ` AND COALESCE(resultado, 'ok') = ?`
		args = append(args, resultado)
	}
	if usuario := strings.TrimSpace(f.UsuarioCreador); usuario != "" {
		where += ` AND LOWER(COALESCE(usuario_creador, '')) = LOWER(?)`
		args = append(args, usuario)
	}
	if reqID := strings.TrimSpace(f.RequestID); reqID != "" {
		where += ` AND COALESCE(request_id, '') = ?`
		args = append(args, reqID)
	}
	if desde := strings.TrimSpace(f.Desde); desde != "" {
		where += ` AND datetime(COALESCE(fecha_evento, fecha_creacion, '')) >= datetime(?)`
		args = append(args, desde)
	}
	if hasta := strings.TrimSpace(f.Hasta); hasta != "" {
		where += ` AND datetime(COALESCE(fecha_evento, fecha_creacion, '')) <= datetime(?)`
		args = append(args, hasta)
	}
	if searchLike := normalizeAuditoriaContains(f.Search, 180); searchLike != "" {
		where += ` AND (
			LOWER(COALESCE(modulo, '')) LIKE ? ESCAPE '!'
			OR LOWER(COALESCE(accion, '')) LIKE ? ESCAPE '!'
			OR LOWER(COALESCE(recurso, '')) LIKE ? ESCAPE '!'
			OR LOWER(COALESCE(endpoint, '')) LIKE ? ESCAPE '!'
			OR LOWER(COALESCE(usuario_creador, '')) LIKE ? ESCAPE '!'
			OR LOWER(COALESCE(request_id, '')) LIKE ? ESCAPE '!'
			OR LOWER(COALESCE(ip_origen, '')) LIKE ? ESCAPE '!'
		)`
		args = append(args, searchLike, searchLike, searchLike, searchLike, searchLike, searchLike, searchLike)
	}

	return where, args
}

// PurgeEmpresaAuditoriaEventos elimina eventos que superan la politica de retencion.
func PurgeEmpresaAuditoriaEventos(dbConn *sql.DB, empresaID int64, retencionDias int64) (int64, error) {
	if empresaID <= 0 {
		return 0, fmt.Errorf("empresa_id es obligatorio")
	}
	if err := EnsureEmpresaAuditoriaSchema(dbConn); err != nil {
		return 0, err
	}
	retencionDias = normalizeAuditoriaRetencionDias(retencionDias)
	expr := fmt.Sprintf("-%d days", retencionDias)

	res, err := dbConn.Exec(`DELETE FROM empresa_auditoria_eventos
	WHERE empresa_id = ?
		AND datetime(COALESCE(fecha_evento, fecha_creacion, '')) < datetime('now','localtime', ?)`,
		empresaID,
		expr,
	)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// PurgeExpiredEmpresaAuditoriaEventos elimina eventos expirados usando fecha_expiracion o regla fallback por retencion_dias.
func PurgeExpiredEmpresaAuditoriaEventos(dbConn *sql.DB) (int64, error) {
	if err := EnsureEmpresaAuditoriaSchema(dbConn); err != nil {
		return 0, err
	}

	res, err := dbConn.Exec(`DELETE FROM empresa_auditoria_eventos
	WHERE datetime(
		CASE
			WHEN COALESCE(fecha_expiracion, '') <> '' THEN fecha_expiracion
			WHEN COALESCE(fecha_evento, '') <> '' THEN datetime(fecha_evento, printf('+%d days', COALESCE(retencion_dias, 180)))
			WHEN COALESCE(fecha_creacion, '') <> '' THEN datetime(fecha_creacion, printf('+%d days', COALESCE(retencion_dias, 180)))
			ELSE datetime('now','localtime','+36500 days')
		END
	) <= datetime('now','localtime')`)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// StartEmpresaAuditoriaRetentionWorker ejecuta limpieza periódica de auditoría expirada.
func StartEmpresaAuditoriaRetentionWorker(dbConn *sql.DB, interval time.Duration, stop <-chan struct{}) {
	if dbConn == nil {
		return
	}
	if interval <= 0 {
		interval = 12 * time.Hour
	}

	runPurge := func(origin string) {
		deleted, err := PurgeExpiredEmpresaAuditoriaEventos(dbConn)
		if err != nil {
			log.Printf("[auditoria] purge expirados (%s) error: %v", origin, err)
			return
		}
		if deleted > 0 {
			log.Printf("[auditoria] purge expirados (%s): %d registros eliminados", origin, deleted)
		}
	}

	runPurge("startup")
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			runPurge("ticker")
		case <-stop:
			return
		}
	}
}

func normalizeAuditoriaValue(raw string) string {
	v := strings.TrimSpace(strings.ToLower(raw))
	if v == "" {
		return ""
	}
	v = strings.ReplaceAll(v, "-", "_")
	v = strings.ReplaceAll(v, " ", "_")
	for strings.Contains(v, "__") {
		v = strings.ReplaceAll(v, "__", "_")
	}
	return strings.Trim(v, "_")
}

func normalizeAuditoriaResultado(raw string) string {
	v := normalizeAuditoriaValue(raw)
	switch v {
	case "ok", "error":
		return v
	default:
		return ""
	}
}

func normalizeAuditoriaEstado(raw string) string {
	v := normalizeAuditoriaValue(raw)
	switch v {
	case "activo", "inactivo", "anulado":
		return v
	default:
		return ""
	}
}

func normalizeAuditoriaRetencionDias(days int64) int64 {
	if days <= 0 {
		return defaultEmpresaAuditoriaRetencionDias
	}
	if days > maxEmpresaAuditoriaRetencionDias {
		return maxEmpresaAuditoriaRetencionDias
	}
	return days
}

func normalizeAuditoriaLimit(limit int) int {
	if limit <= 0 {
		return 200
	}
	if limit > 1000 {
		return 1000
	}
	return limit
}

func normalizeAuditoriaOffset(offset int) int {
	if offset < 0 {
		return 0
	}
	if offset > 500000 {
		return 500000
	}
	return offset
}

func normalizeAuditoriaContains(raw string, maxLen int) string {
	v := sanitizeAuditoriaText(strings.ToLower(strings.TrimSpace(raw)), maxLen)
	if v == "" {
		return ""
	}
	v = strings.ReplaceAll(v, "!", "!!")
	v = strings.ReplaceAll(v, "%", "!%")
	v = strings.ReplaceAll(v, "_", "!_")
	return "%" + v + "%"
}

func sanitizeAuditoriaText(raw string, maxLen int) string {
	v := strings.TrimSpace(raw)
	if maxLen <= 0 || len(v) <= maxLen {
		return v
	}
	return v[:maxLen]
}

func sanitizeAuditoriaEndpoint(raw string) string {
	v := strings.TrimSpace(raw)
	if v == "" {
		return "/"
	}
	if !strings.HasPrefix(v, "/") {
		v = "/" + v
	}
	return sanitizeAuditoriaText(v, 300)
}
