package db

import (
	"crypto/sha256"
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
	auditoriaSeveridadCritica            = "critica"
	auditoriaSeveridadAlta               = "alta"
	auditoriaSeveridadMedia              = "media"
	auditoriaSeveridadBaja               = "baja"
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
			id BIGSERIAL PRIMARY KEY,
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
			fecha_evento TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_expiracion TEXT,
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
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
	if err := ensureColumnIfMissing(dbConn, "empresa_auditoria_eventos", "fecha_evento", "TEXT DEFAULT (CURRENT_TIMESTAMP)"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_auditoria_eventos", "fecha_expiracion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_auditoria_eventos", "fecha_actualizacion", "TEXT DEFAULT (CURRENT_TIMESTAMP)"); err != nil {
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
	if err := ensureEmpresaAuditoriaFTSSchema(dbConn); err != nil {
		return err
	}
	if err := ensureEmpresaAuditoriaIAConsultaSchema(dbConn); err != nil {
		return err
	}

	return nil
}

func ensureEmpresaAuditoriaIAConsultaSchema(dbConn *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_auditoria_ia_consultas (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			alcance TEXT NOT NULL,
			modelo TEXT,
			usuario_consulta TEXT,
			pregunta_hash TEXT,
			pregunta_resumen TEXT,
			filtros_json TEXT,
			resultados_json TEXT,
			eventos_consultados INTEGER DEFAULT 0,
			contexto_caracteres INTEGER DEFAULT 0,
			resultado TEXT DEFAULT 'ok',
			fecha_consulta TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_auditoria_ia_consultas_empresa_fecha ON empresa_auditoria_ia_consultas(empresa_id, fecha_consulta DESC, id DESC);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_auditoria_ia_consultas_alcance_fecha ON empresa_auditoria_ia_consultas(alcance, fecha_consulta DESC);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_auditoria_ia_consultas_hash ON empresa_auditoria_ia_consultas(pregunta_hash);`,
	}
	for _, stmt := range stmts {
		if _, err := dbConn.Exec(stmt); err != nil {
			return err
		}
	}
	for _, col := range []struct {
		name string
		def  string
	}{
		{"alcance", "TEXT"},
		{"modelo", "TEXT"},
		{"usuario_consulta", "TEXT"},
		{"pregunta_hash", "TEXT"},
		{"pregunta_resumen", "TEXT"},
		{"filtros_json", "TEXT"},
		{"resultados_json", "TEXT"},
		{"eventos_consultados", "INTEGER DEFAULT 0"},
		{"contexto_caracteres", "INTEGER DEFAULT 0"},
		{"resultado", "TEXT DEFAULT 'ok'"},
		{"fecha_consulta", "TEXT DEFAULT (CURRENT_TIMESTAMP)"},
		{"fecha_actualizacion", "TEXT DEFAULT (CURRENT_TIMESTAMP)"},
		{"usuario_creador", "TEXT"},
		{"estado", "TEXT DEFAULT 'activo'"},
		{"observaciones", "TEXT"},
	} {
		if err := ensureColumnIfMissing(dbConn, "empresa_auditoria_ia_consultas", col.name, col.def); err != nil {
			return err
		}
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

	severidad := resolveAuditoriaSeveridad(in.Modulo, in.Accion, in.Resultado, in.CodigoHTTP, metadata)
	retencionDias := in.RetencionDias
	if retencionDias <= 0 {
		retencionDias = resolveAuditoriaPoliticaRetencionDias(in.Modulo, severidad)
	}
	retencionDias = normalizeAuditoriaRetencionDias(retencionDias)
	metadata = enrichAuditoriaRetentionMetadata(metadata, in.Modulo, severidad, retencionDias)
	ttlExpr := fmt.Sprintf("+%d days", retencionDias)

	id, err := insertSQLCompat(dbConn, `INSERT INTO empresa_auditoria_eventos (
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
		COALESCE(NULLIF(?, ''), CURRENT_TIMESTAMP),
		pcs_ts(COALESCE(NULLIF(?, ''), 'now'), ?),
		CURRENT_TIMESTAMP,
		CURRENT_TIMESTAMP,
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
	return id, nil
}

// ListEmpresaAuditoriaEventos lista eventos de auditoria por empresa con filtros.
func ListEmpresaAuditoriaEventos(dbConn *sql.DB, empresaID int64, f EmpresaAuditoriaEventoFilter) ([]EmpresaAuditoriaEvento, error) {
	if empresaID <= 0 {
		return nil, fmt.Errorf("empresa_id es obligatorio")
	}
	if err := EnsureEmpresaAuditoriaSchema(dbConn); err != nil {
		return nil, err
	}

	where, args := buildEmpresaAuditoriaWhereClause(dbConn, empresaID, f)

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

	where, args := buildEmpresaAuditoriaWhereClause(dbConn, empresaID, f)
	query := `SELECT COUNT(1) FROM empresa_auditoria_eventos` + where

	var total int64
	if err := dbConn.QueryRow(query, args...).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func buildEmpresaAuditoriaWhereClause(dbConn *sql.DB, empresaID int64, f EmpresaAuditoriaEventoFilter) (string, []interface{}) {
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
		where += ` AND pcs_ts(COALESCE(fecha_evento, fecha_creacion, '')) >= pcs_ts(?)`
		args = append(args, desde)
	}
	if hasta := strings.TrimSpace(f.Hasta); hasta != "" {
		where += ` AND pcs_ts(COALESCE(fecha_evento, fecha_creacion, '')) <= pcs_ts(?)`
		args = append(args, hasta)
	}
	if searchClause, searchArgs := buildAuditoriaSearchClause(dbConn, empresaID, f.Search); searchClause != "" {
		where += searchClause
		args = append(args, searchArgs...)
	}

	return where, args
}

func buildAuditoriaSearchClause(dbConn *sql.DB, empresaID int64, rawSearch string) (string, []interface{}) {
	search := strings.TrimSpace(rawSearch)
	if search == "" {
		return "", nil
	}

	if searchLike := normalizeAuditoriaContains(search, 180); searchLike != "" {
		return ` AND (
			LOWER(COALESCE(modulo, '')) LIKE ? ESCAPE '!'
			OR LOWER(COALESCE(accion, '')) LIKE ? ESCAPE '!'
			OR LOWER(COALESCE(recurso, '')) LIKE ? ESCAPE '!'
			OR LOWER(COALESCE(endpoint, '')) LIKE ? ESCAPE '!'
			OR LOWER(COALESCE(usuario_creador, '')) LIKE ? ESCAPE '!'
			OR LOWER(COALESCE(request_id, '')) LIKE ? ESCAPE '!'
			OR LOWER(COALESCE(ip_origen, '')) LIKE ? ESCAPE '!'
			OR LOWER(COALESCE(observaciones, '')) LIKE ? ESCAPE '!'
			OR LOWER(COALESCE(metadata_json, '')) LIKE ? ESCAPE '!'
		)`, []interface{}{searchLike, searchLike, searchLike, searchLike, searchLike, searchLike, searchLike, searchLike, searchLike}
	}

	return "", nil
}

func auditoriaFTSEnabled(dbConn *sql.DB) bool {
	if dbConn == nil {
		return false
	}
	return false
}

func ensureEmpresaAuditoriaFTSSchema(dbConn *sql.DB) error {
	// PostgreSQL-only: no usamos tablas virtuales FTS.
	return nil
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
		AND pcs_ts(COALESCE(fecha_evento, fecha_creacion, '')) < pcs_ts('now','localtime', ?)`,
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
	WHERE CASE
		WHEN COALESCE(fecha_expiracion, '') <> '' THEN pcs_ts(fecha_expiracion)
		WHEN COALESCE(fecha_evento, '') <> '' THEN pcs_ts(fecha_evento, '+' || COALESCE(retencion_dias, 180)::text || ' days')
		WHEN COALESCE(fecha_creacion, '') <> '' THEN pcs_ts(fecha_creacion, '+' || COALESCE(retencion_dias, 180)::text || ' days')
		ELSE pcs_ts('now','localtime','+36500 days')
	END <= CURRENT_TIMESTAMP`)
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

// BuildEmpresaAuditoriaAIContext construye un resumen reciente de auditoria para la IA.
// Es best-effort: si la auditoria no existe o falla, devuelve una nota acotada y no propaga error.
func BuildEmpresaAuditoriaAIContext(dbConn *sql.DB, empresaID int64, limit int, window time.Duration) string {
	if dbConn == nil || empresaID <= 0 {
		return ""
	}
	if window <= 0 {
		window = 30 * time.Minute
	}
	if limit <= 0 {
		limit = 12
	}
	if limit > 40 {
		limit = 40
	}
	if ok, err := tableExists(dbConn, "empresa_auditoria_eventos"); err != nil || !ok {
		return "AUDITORIA_TIEMPO_REAL\n- estado=no_disponible\n- nota=la IA continua operando con el contexto transaccional disponible; la auditoria no bloqueo la respuesta.\n"
	}

	cutoff := time.Now().Add(-window).Format("2006-01-02 15:04:05")
	var total, errores, usuarios int64
	var ultima string
	if err := queryRowSQLCompat(dbConn, `SELECT
		COUNT(1),
		COALESCE(SUM(CASE WHEN LOWER(COALESCE(resultado,'')) = 'error' OR COALESCE(codigo_http,0) >= 400 THEN 1 ELSE 0 END), 0),
		COUNT(DISTINCT COALESCE(NULLIF(usuario_creador,''), 'sistema')),
		COALESCE(MAX(COALESCE(fecha_evento, fecha_creacion, '')), '')
		FROM empresa_auditoria_eventos
		WHERE empresa_id = ?
		  AND COALESCE(estado, 'activo') = 'activo'
		  AND COALESCE(fecha_evento, fecha_creacion, '') >= ?`, empresaID, cutoff).Scan(&total, &errores, &usuarios, &ultima); err != nil {
		return "AUDITORIA_TIEMPO_REAL\n- estado=error_lectura\n- nota=la auditoria reciente no se pudo consultar; la IA no debe inventar actividad de usuarios.\n"
	}

	var b strings.Builder
	b.WriteString("AUDITORIA_TIEMPO_REAL\n")
	b.WriteString(fmt.Sprintf("- ventana_minutos=%d\n", int(window.Minutes())))
	b.WriteString(fmt.Sprintf("- eventos_recientes=%d\n", total))
	b.WriteString(fmt.Sprintf("- eventos_error=%d\n", errores))
	b.WriteString(fmt.Sprintf("- usuarios_distintos=%d\n", usuarios))
	b.WriteString(fmt.Sprintf("- ultima_actividad=%s\n", safeAuditoriaAIValue(ultima)))
	b.WriteString("- regla=usar esta auditoria como senal de actividad reciente; no inventar acciones no auditadas.\n")

	writeAuditoriaAIContextSection(&b, "AUDITORIA_MODULOS_RECIENTES", empresaAuditoriaAIModuleSummary(dbConn, empresaID, cutoff, 8))
	writeAuditoriaAIContextSection(&b, "AUDITORIA_EVENTOS_RECIENTES", empresaAuditoriaAIRecentEvents(dbConn, empresaID, cutoff, limit))
	return b.String()
}

// BuildSuperAuditoriaAIContext resume actividad reciente de auditoria para el chat global.
func BuildSuperAuditoriaAIContext(dbConn *sql.DB, limit int, window time.Duration) string {
	if dbConn == nil {
		return ""
	}
	if window <= 0 {
		window = 30 * time.Minute
	}
	if limit <= 0 {
		limit = 15
	}
	if limit > 50 {
		limit = 50
	}
	if ok, err := tableExists(dbConn, "empresa_auditoria_eventos"); err != nil || !ok {
		return "AUDITORIA_GLOBAL_TIEMPO_REAL\n- estado=no_disponible\n- nota=la IA global continua sin depender de auditoria.\n"
	}

	cutoff := time.Now().Add(-window).Format("2006-01-02 15:04:05")
	var total, errores, empresas, usuarios int64
	var ultima string
	if err := queryRowSQLCompat(dbConn, `SELECT
		COUNT(1),
		COALESCE(SUM(CASE WHEN LOWER(COALESCE(resultado,'')) = 'error' OR COALESCE(codigo_http,0) >= 400 THEN 1 ELSE 0 END), 0),
		COUNT(DISTINCT empresa_id),
		COUNT(DISTINCT COALESCE(NULLIF(usuario_creador,''), 'sistema')),
		COALESCE(MAX(COALESCE(fecha_evento, fecha_creacion, '')), '')
		FROM empresa_auditoria_eventos
		WHERE COALESCE(estado, 'activo') = 'activo'
		  AND COALESCE(fecha_evento, fecha_creacion, '') >= ?`, cutoff).Scan(&total, &errores, &empresas, &usuarios, &ultima); err != nil {
		return "AUDITORIA_GLOBAL_TIEMPO_REAL\n- estado=error_lectura\n- nota=la auditoria global reciente no se pudo consultar; la IA no debe inventar actividad.\n"
	}

	var b strings.Builder
	b.WriteString("AUDITORIA_GLOBAL_TIEMPO_REAL\n")
	b.WriteString(fmt.Sprintf("- ventana_minutos=%d\n", int(window.Minutes())))
	b.WriteString(fmt.Sprintf("- eventos_recientes=%d\n", total))
	b.WriteString(fmt.Sprintf("- eventos_error=%d\n", errores))
	b.WriteString(fmt.Sprintf("- empresas_con_actividad=%d\n", empresas))
	b.WriteString(fmt.Sprintf("- usuarios_distintos=%d\n", usuarios))
	b.WriteString(fmt.Sprintf("- ultima_actividad=%s\n", safeAuditoriaAIValue(ultima)))
	b.WriteString("- regla=usar solo como senal operacional reciente; no revelar secretos ni asumir intencion del usuario.\n")

	writeAuditoriaAIContextSection(&b, "AUDITORIA_EMPRESAS_ACTIVAS", superAuditoriaAICompanySummary(dbConn, cutoff, 8))
	writeAuditoriaAIContextSection(&b, "AUDITORIA_EVENTOS_GLOBALES_RECIENTES", superAuditoriaAIRecentEvents(dbConn, cutoff, limit))
	return b.String()
}

// BuildEmpresaAuditoriaAIContextForQuestion agrega busqueda profunda de auditoria y consultas DB seguras por intencion.
func BuildEmpresaAuditoriaAIContextForQuestion(dbConn *sql.DB, empresaID int64, pregunta, usuario, modelo string, limit int, window time.Duration) string {
	context := strings.TrimSpace(BuildEmpresaAuditoriaAIContext(dbConn, empresaID, limit, window))
	if dbConn == nil || empresaID <= 0 {
		return context
	}

	var sections []string
	var eventosConsultados int64
	if searchContext, count := buildEmpresaAuditoriaAISearchContext(dbConn, empresaID, pregunta, limit); strings.TrimSpace(searchContext) != "" {
		sections = append(sections, searchContext)
		eventosConsultados += count
	}
	if dbContext := buildEmpresaAuditoriaAIDBFollowupContext(dbConn, empresaID, pregunta, 6); strings.TrimSpace(dbContext) != "" {
		sections = append(sections, dbContext)
	}
	if len(sections) > 0 {
		if context != "" {
			context += "\n"
		}
		context += strings.Join(sections, "\n")
	}
	registerAuditoriaIAConsultaNoBloqueante(dbConn, empresaID, "empresa", pregunta, usuario, modelo, context, eventosConsultados, map[string]interface{}{
		"modo":             "contexto_auditoria_profundo",
		"ventana_minutos":  int(normalizeAuditAIWindow(window).Minutes()),
		"limite_eventos":   normalizeAuditAILimit(limit, 12, 40),
		"consulta_segura":  true,
		"sql_libre_modelo": false,
	})
	return context
}

// BuildSuperAuditoriaAIContextForQuestion agrega busqueda profunda de auditoria global y consultas DB seguras.
func BuildSuperAuditoriaAIContextForQuestion(dbConn *sql.DB, pregunta, usuario, modelo string, limit int, window time.Duration) string {
	context := strings.TrimSpace(BuildSuperAuditoriaAIContext(dbConn, limit, window))
	if dbConn == nil {
		return context
	}

	var sections []string
	var eventosConsultados int64
	if searchContext, count := buildSuperAuditoriaAISearchContext(dbConn, pregunta, limit); strings.TrimSpace(searchContext) != "" {
		sections = append(sections, searchContext)
		eventosConsultados += count
	}
	if dbContext := buildSuperAuditoriaAIDBFollowupContext(dbConn, pregunta, 8); strings.TrimSpace(dbContext) != "" {
		sections = append(sections, dbContext)
	}
	if len(sections) > 0 {
		if context != "" {
			context += "\n"
		}
		context += strings.Join(sections, "\n")
	}
	registerAuditoriaIAConsultaNoBloqueante(dbConn, 0, "global_super", pregunta, usuario, modelo, context, eventosConsultados, map[string]interface{}{
		"modo":             "contexto_auditoria_global_profundo",
		"ventana_minutos":  int(normalizeAuditAIWindow(window).Minutes()),
		"limite_eventos":   normalizeAuditAILimit(limit, 15, 50),
		"consulta_segura":  true,
		"sql_libre_modelo": false,
	})
	return context
}

func empresaAuditoriaAIModuleSummary(dbConn *sql.DB, empresaID int64, cutoff string, limit int) []string {
	rows, err := querySQLCompat(dbConn, `SELECT
		COALESCE(modulo, 'sin_modulo'),
		COALESCE(accion, 'sin_accion'),
		COUNT(1),
		COALESCE(SUM(CASE WHEN LOWER(COALESCE(resultado,'')) = 'error' OR COALESCE(codigo_http,0) >= 400 THEN 1 ELSE 0 END), 0),
		COALESCE(MAX(COALESCE(fecha_evento, fecha_creacion, '')), '')
		FROM empresa_auditoria_eventos
		WHERE empresa_id = ?
		  AND COALESCE(estado, 'activo') = 'activo'
		  AND COALESCE(fecha_evento, fecha_creacion, '') >= ?
		GROUP BY COALESCE(modulo, 'sin_modulo'), COALESCE(accion, 'sin_accion')
		ORDER BY COUNT(1) DESC, COALESCE(MAX(COALESCE(fecha_evento, fecha_creacion, '')), '') DESC
		LIMIT ?`, empresaID, cutoff, limit)
	if err != nil {
		return nil
	}
	defer rows.Close()

	out := make([]string, 0, limit)
	for rows.Next() {
		var modulo, accion, ultima string
		var total, errores int64
		if err := rows.Scan(&modulo, &accion, &total, &errores, &ultima); err != nil {
			return nil
		}
		out = append(out, fmt.Sprintf("modulo=%s accion=%s eventos=%d errores=%d ultima=%s",
			safeAuditoriaAIValue(modulo), safeAuditoriaAIValue(accion), total, errores, safeAuditoriaAIValue(ultima)))
	}
	return out
}

func empresaAuditoriaAIRecentEvents(dbConn *sql.DB, empresaID int64, cutoff string, limit int) []string {
	rows, err := querySQLCompat(dbConn, `SELECT
		COALESCE(fecha_evento, fecha_creacion, ''),
		COALESCE(modulo, ''),
		COALESCE(accion, ''),
		COALESCE(recurso, ''),
		COALESCE(recurso_id, 0),
		COALESCE(metodo_http, ''),
		COALESCE(endpoint, ''),
		COALESCE(resultado, 'ok'),
		COALESCE(codigo_http, 0),
		COALESCE(request_id, ''),
		COALESCE(usuario_creador, 'sistema')
		FROM empresa_auditoria_eventos
		WHERE empresa_id = ?
		  AND COALESCE(estado, 'activo') = 'activo'
		  AND COALESCE(fecha_evento, fecha_creacion, '') >= ?
		ORDER BY COALESCE(fecha_evento, fecha_creacion, '') DESC, id DESC
		LIMIT ?`, empresaID, cutoff, limit)
	if err != nil {
		return nil
	}
	defer rows.Close()

	out := make([]string, 0, limit)
	for rows.Next() {
		var fecha, modulo, accion, recurso, metodo, endpoint, resultado, requestID, usuario string
		var recursoID, codigoHTTP int64
		if err := rows.Scan(&fecha, &modulo, &accion, &recurso, &recursoID, &metodo, &endpoint, &resultado, &codigoHTTP, &requestID, &usuario); err != nil {
			return nil
		}
		out = append(out, fmt.Sprintf("%s | usuario=%s | %s %s | modulo=%s accion=%s recurso=%s recurso_id=%d resultado=%s http=%d request_id=%s",
			safeAuditoriaAIValue(fecha), safeAuditoriaAIValue(usuario), safeAuditoriaAIValue(metodo), safeAuditoriaAIEndpoint(endpoint),
			safeAuditoriaAIValue(modulo), safeAuditoriaAIValue(accion), safeAuditoriaAIValue(recurso), recursoID, safeAuditoriaAIValue(resultado), codigoHTTP, safeAuditoriaAIValue(requestID)))
	}
	return out
}

func superAuditoriaAICompanySummary(dbConn *sql.DB, cutoff string, limit int) []string {
	rows, err := querySQLCompat(dbConn, `SELECT
		empresa_id,
		COUNT(1),
		COALESCE(SUM(CASE WHEN LOWER(COALESCE(resultado,'')) = 'error' OR COALESCE(codigo_http,0) >= 400 THEN 1 ELSE 0 END), 0),
		COUNT(DISTINCT COALESCE(NULLIF(usuario_creador,''), 'sistema')),
		COALESCE(MAX(COALESCE(fecha_evento, fecha_creacion, '')), '')
		FROM empresa_auditoria_eventos
		WHERE COALESCE(estado, 'activo') = 'activo'
		  AND COALESCE(fecha_evento, fecha_creacion, '') >= ?
		GROUP BY empresa_id
		ORDER BY COALESCE(MAX(COALESCE(fecha_evento, fecha_creacion, '')), '') DESC, COUNT(1) DESC
		LIMIT ?`, cutoff, limit)
	if err != nil {
		return nil
	}
	defer rows.Close()

	out := make([]string, 0, limit)
	for rows.Next() {
		var empresaID, total, errores, usuarios int64
		var ultima string
		if err := rows.Scan(&empresaID, &total, &errores, &usuarios, &ultima); err != nil {
			return nil
		}
		out = append(out, fmt.Sprintf("empresa_id=%d eventos=%d errores=%d usuarios=%d ultima=%s", empresaID, total, errores, usuarios, safeAuditoriaAIValue(ultima)))
	}
	return out
}

func superAuditoriaAIRecentEvents(dbConn *sql.DB, cutoff string, limit int) []string {
	rows, err := querySQLCompat(dbConn, `SELECT
		empresa_id,
		COALESCE(fecha_evento, fecha_creacion, ''),
		COALESCE(modulo, ''),
		COALESCE(accion, ''),
		COALESCE(recurso, ''),
		COALESCE(metodo_http, ''),
		COALESCE(endpoint, ''),
		COALESCE(resultado, 'ok'),
		COALESCE(codigo_http, 0),
		COALESCE(usuario_creador, 'sistema')
		FROM empresa_auditoria_eventos
		WHERE COALESCE(estado, 'activo') = 'activo'
		  AND COALESCE(fecha_evento, fecha_creacion, '') >= ?
		ORDER BY COALESCE(fecha_evento, fecha_creacion, '') DESC, id DESC
		LIMIT ?`, cutoff, limit)
	if err != nil {
		return nil
	}
	defer rows.Close()

	out := make([]string, 0, limit)
	for rows.Next() {
		var empresaID, codigoHTTP int64
		var fecha, modulo, accion, recurso, metodo, endpoint, resultado, usuario string
		if err := rows.Scan(&empresaID, &fecha, &modulo, &accion, &recurso, &metodo, &endpoint, &resultado, &codigoHTTP, &usuario); err != nil {
			return nil
		}
		out = append(out, fmt.Sprintf("empresa_id=%d | %s | usuario=%s | %s %s | modulo=%s accion=%s recurso=%s resultado=%s http=%d",
			empresaID, safeAuditoriaAIValue(fecha), safeAuditoriaAIValue(usuario), safeAuditoriaAIValue(metodo), safeAuditoriaAIEndpoint(endpoint),
			safeAuditoriaAIValue(modulo), safeAuditoriaAIValue(accion), safeAuditoriaAIValue(recurso), safeAuditoriaAIValue(resultado), codigoHTTP))
	}
	return out
}

func writeAuditoriaAIContextSection(b *strings.Builder, sectionName string, lines []string) {
	if len(lines) == 0 {
		return
	}
	b.WriteString(sectionName)
	b.WriteString("\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		b.WriteString("- ")
		b.WriteString(line)
		b.WriteString("\n")
	}
}

func safeAuditoriaAIValue(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return "sin_dato"
	}
	return sanitizeAuditoriaText(strings.ReplaceAll(v, "\n", " "), 220)
}

func safeAuditoriaAIEndpoint(v string) string {
	v = strings.TrimSpace(v)
	if idx := strings.Index(v, "?"); idx >= 0 {
		v = v[:idx]
	}
	return sanitizeAuditoriaEndpoint(v)
}

func buildEmpresaAuditoriaAISearchContext(dbConn *sql.DB, empresaID int64, pregunta string, limit int) (string, int64) {
	if !auditAIQuestionNeedsDeepSearch(pregunta) {
		return "", 0
	}
	filter := auditAIQuestionFilter(pregunta)
	filter.Limit = normalizeAuditAILimit(limit, 12, 40)
	filter.Offset = 0

	total, err := CountEmpresaAuditoriaEventos(dbConn, empresaID, filter)
	if err != nil {
		return "AUDITORIA_BUSQUEDA_PROFUNDA\n- estado=error_consulta\n- nota=la busqueda segura en auditoria no pudo ejecutarse.\n", 0
	}
	rows, err := ListEmpresaAuditoriaEventos(dbConn, empresaID, filter)
	if err != nil {
		return "AUDITORIA_BUSQUEDA_PROFUNDA\n- estado=error_lectura\n- nota=la auditoria profunda no pudo listar resultados.\n", 0
	}

	var b strings.Builder
	b.WriteString("AUDITORIA_BUSQUEDA_PROFUNDA\n")
	b.WriteString("- fuente=empresa_auditoria_eventos\n")
	b.WriteString("- alcance=empresa\n")
	b.WriteString(fmt.Sprintf("- total_coincidencias=%d\n", total))
	b.WriteString(fmt.Sprintf("- resultados_entregados=%d\n", len(rows)))
	b.WriteString(fmt.Sprintf("- filtro_modulo=%s\n", safeAuditoriaAIValue(filter.Modulo)))
	b.WriteString(fmt.Sprintf("- filtro_accion=%s\n", safeAuditoriaAIValue(filter.Accion)))
	b.WriteString(fmt.Sprintf("- filtro_resultado=%s\n", safeAuditoriaAIValue(filter.Resultado)))
	b.WriteString(fmt.Sprintf("- filtro_usuario=%s\n", safeAuditoriaAIValue(filter.UsuarioCreador)))
	b.WriteString(fmt.Sprintf("- filtro_busqueda=%s\n", safeAuditoriaAIValue(filter.Search)))
	for _, row := range rows {
		b.WriteString(fmt.Sprintf("- %s | usuario=%s | %s %s | modulo=%s accion=%s recurso=%s recurso_id=%d resultado=%s http=%d request_id=%s\n",
			safeAuditoriaAIValue(row.FechaEvento), safeAuditoriaAIValue(row.UsuarioCreador), safeAuditoriaAIValue(row.MetodoHTTP), safeAuditoriaAIEndpoint(row.Endpoint),
			safeAuditoriaAIValue(row.Modulo), safeAuditoriaAIValue(row.Accion), safeAuditoriaAIValue(row.Recurso), row.RecursoID, safeAuditoriaAIValue(row.Resultado), row.CodigoHTTP, safeAuditoriaAIValue(row.RequestID)))
	}
	return b.String(), int64(len(rows))
}

func buildSuperAuditoriaAISearchContext(dbConn *sql.DB, pregunta string, limit int) (string, int64) {
	if !auditAIQuestionNeedsDeepSearch(pregunta) {
		return "", 0
	}
	if ok, err := tableExists(dbConn, "empresa_auditoria_eventos"); err != nil || !ok {
		return "", 0
	}
	filter := auditAIQuestionFilter(pregunta)
	limit = normalizeAuditAILimit(limit, 15, 50)
	where := ` WHERE COALESCE(estado, 'activo') = 'activo'`
	args := make([]interface{}, 0, 8)
	if filter.Modulo != "" {
		where += ` AND COALESCE(modulo, '') = ?`
		args = append(args, filter.Modulo)
	}
	if filter.Accion != "" {
		where += ` AND COALESCE(accion, '') = ?`
		args = append(args, filter.Accion)
	}
	if filter.Resultado != "" {
		where += ` AND COALESCE(resultado, 'ok') = ?`
		args = append(args, filter.Resultado)
	}
	if filter.UsuarioCreador != "" {
		where += ` AND LOWER(COALESCE(usuario_creador, '')) = LOWER(?)`
		args = append(args, filter.UsuarioCreador)
	}
	if filter.Search != "" {
		if like := normalizeAuditoriaContains(filter.Search, 180); like != "" {
			where += ` AND (
				LOWER(COALESCE(modulo, '')) LIKE ? ESCAPE '!'
				OR LOWER(COALESCE(accion, '')) LIKE ? ESCAPE '!'
				OR LOWER(COALESCE(recurso, '')) LIKE ? ESCAPE '!'
				OR LOWER(COALESCE(endpoint, '')) LIKE ? ESCAPE '!'
				OR LOWER(COALESCE(usuario_creador, '')) LIKE ? ESCAPE '!'
				OR LOWER(COALESCE(request_id, '')) LIKE ? ESCAPE '!'
				OR LOWER(COALESCE(observaciones, '')) LIKE ? ESCAPE '!'
			)`
			args = append(args, like, like, like, like, like, like, like)
		}
	}

	var total int64
	if err := queryRowSQLCompat(dbConn, `SELECT COUNT(1) FROM empresa_auditoria_eventos`+where, args...).Scan(&total); err != nil {
		return "AUDITORIA_GLOBAL_BUSQUEDA_PROFUNDA\n- estado=error_consulta\n", 0
	}
	rowsArgs := append([]interface{}{}, args...)
	rowsArgs = append(rowsArgs, limit)
	rows, err := querySQLCompat(dbConn, `SELECT
		empresa_id,
		COALESCE(fecha_evento, fecha_creacion, ''),
		COALESCE(modulo, ''),
		COALESCE(accion, ''),
		COALESCE(recurso, ''),
		COALESCE(metodo_http, ''),
		COALESCE(endpoint, ''),
		COALESCE(resultado, 'ok'),
		COALESCE(codigo_http, 0),
		COALESCE(request_id, ''),
		COALESCE(usuario_creador, 'sistema')
		FROM empresa_auditoria_eventos`+where+`
		ORDER BY COALESCE(fecha_evento, fecha_creacion, '') DESC, id DESC
		LIMIT ?`, rowsArgs...)
	if err != nil {
		return "AUDITORIA_GLOBAL_BUSQUEDA_PROFUNDA\n- estado=error_lectura\n", 0
	}
	defer rows.Close()

	var b strings.Builder
	b.WriteString("AUDITORIA_GLOBAL_BUSQUEDA_PROFUNDA\n")
	b.WriteString("- fuente=empresa_auditoria_eventos\n")
	b.WriteString("- alcance=global_super\n")
	b.WriteString(fmt.Sprintf("- total_coincidencias=%d\n", total))
	count := int64(0)
	for rows.Next() {
		var empresaID, codigoHTTP int64
		var fecha, modulo, accion, recurso, metodo, endpoint, resultado, requestID, usuario string
		if err := rows.Scan(&empresaID, &fecha, &modulo, &accion, &recurso, &metodo, &endpoint, &resultado, &codigoHTTP, &requestID, &usuario); err != nil {
			return "", count
		}
		count++
		b.WriteString(fmt.Sprintf("- empresa_id=%d | %s | usuario=%s | %s %s | modulo=%s accion=%s recurso=%s resultado=%s http=%d request_id=%s\n",
			empresaID, safeAuditoriaAIValue(fecha), safeAuditoriaAIValue(usuario), safeAuditoriaAIValue(metodo), safeAuditoriaAIEndpoint(endpoint),
			safeAuditoriaAIValue(modulo), safeAuditoriaAIValue(accion), safeAuditoriaAIValue(recurso), safeAuditoriaAIValue(resultado), codigoHTTP, safeAuditoriaAIValue(requestID)))
	}
	return b.String(), count
}

func buildEmpresaAuditoriaAIDBFollowupContext(dbConn *sql.DB, empresaID int64, pregunta string, limit int) string {
	folded := aiFoldText(pregunta)
	if strings.TrimSpace(folded) == "" {
		return ""
	}
	availableTables, err := aiAvailableTables(dbConn, []string{
		"clientes",
		"productos",
		"inventario_existencias",
		"carritos_compras",
		"carrito_compra_items",
		"empresa_finanzas_movimientos",
		"empresa_auditoria_eventos",
	})
	if err != nil {
		return ""
	}
	var b strings.Builder
	b.WriteString("AUDITORIA_CONSULTAS_DB_SEGURAS\n")
	b.WriteString("- regla=consultas parametrizadas por whitelist y empresa_id; el modelo no genera SQL libre.\n")
	wrote := false
	if auditAIQuestionMentionsUsers(folded) {
		writeAuditoriaAIContextSection(&b, "DB_USUARIOS_CON_ACTIVIDAD_RECIENTE", empresaAuditoriaAIActiveUsers(dbConn, empresaID, 8))
		wrote = true
	}
	if aiLooksLikeSalesQuestion(folded) {
		writeAuditoriaAIContextSection(&b, "DB_VENTAS_RELACIONADAS", empresaAIVentasRecientes(dbConn, empresaID, availableTables, limit))
		wrote = true
	}
	if aiLooksLikeFinanceQuestion(folded) {
		writeAuditoriaAIContextSection(&b, "DB_FINANZAS_RELACIONADAS", empresaAIFinanzasRecientes(dbConn, empresaID, availableTables, limit))
		wrote = true
	}
	if aiLooksLikeInventoryQuestion(folded) || aiLooksLikeProductQuestion(folded) {
		writeAuditoriaAIContextSection(&b, "DB_INVENTARIO_RELACIONADO", empresaAIAlertasInventario(dbConn, empresaID, availableTables, limit))
		wrote = true
	}
	if aiLooksLikeClientQuestion(folded) {
		writeAuditoriaAIContextSection(&b, "DB_CLIENTES_RELACIONADOS", empresaAITopClientes(dbConn, empresaID, availableTables, limit))
		wrote = true
	}
	if !wrote {
		return ""
	}
	return b.String()
}

func buildSuperAuditoriaAIDBFollowupContext(dbConn *sql.DB, pregunta string, limit int) string {
	folded := aiFoldText(pregunta)
	if strings.TrimSpace(folded) == "" {
		return ""
	}
	availableTables, err := aiAvailableTables(dbConn, []string{
		"empresas",
		"clientes",
		"productos",
		"inventario_existencias",
		"carritos_compras",
		"empresa_finanzas_movimientos",
		"empresa_auditoria_eventos",
	})
	if err != nil {
		return ""
	}
	var b strings.Builder
	b.WriteString("AUDITORIA_GLOBAL_CONSULTAS_DB_SEGURAS\n")
	b.WriteString("- regla=consultas globales agregadas por whitelist; no se ejecuta SQL libre del modelo.\n")
	wrote := false
	if auditAIQuestionMentionsUsers(folded) {
		writeAuditoriaAIContextSection(&b, "DB_GLOBAL_USUARIOS_CON_ACTIVIDAD", superAuditoriaAIActiveUsers(dbConn, 10))
		wrote = true
	}
	if aiLooksLikeSalesQuestion(folded) {
		writeAuditoriaAIContextSection(&b, "DB_GLOBAL_VENTAS_RECIENTES", superAIVentasRecientes(dbConn, availableTables, limit))
		writeAuditoriaAIContextSection(&b, "DB_GLOBAL_EMPRESAS_TOP_VENTAS", superAITopEmpresasVentas(dbConn, availableTables, limit))
		wrote = true
	}
	if aiLooksLikeFinanceQuestion(folded) {
		writeAuditoriaAIContextSection(&b, "DB_GLOBAL_FINANZAS_MUESTRA", superAIFinanzasMuestraGlobal(dbConn, availableTables, limit))
		wrote = true
	}
	if aiLooksLikeInventoryQuestion(folded) || aiLooksLikeProductQuestion(folded) {
		writeAuditoriaAIContextSection(&b, "DB_GLOBAL_ALERTAS_INVENTARIO", superAIAlertasInventario(dbConn, availableTables, limit))
		wrote = true
	}
	if !wrote {
		return ""
	}
	return b.String()
}

func empresaAuditoriaAIActiveUsers(dbConn *sql.DB, empresaID int64, limit int) []string {
	rows, err := querySQLCompat(dbConn, `SELECT
		COALESCE(NULLIF(usuario_creador,''), 'sistema') AS usuario,
		COUNT(1) AS eventos,
		COALESCE(SUM(CASE WHEN LOWER(COALESCE(resultado,'')) = 'error' OR COALESCE(codigo_http,0) >= 400 THEN 1 ELSE 0 END), 0) AS errores,
		COALESCE(MAX(COALESCE(fecha_evento, fecha_creacion, '')), '') AS ultima
		FROM empresa_auditoria_eventos
		WHERE empresa_id = ?
		  AND COALESCE(estado, 'activo') = 'activo'
		  AND COALESCE(fecha_evento, fecha_creacion, '') >= ?
		GROUP BY COALESCE(NULLIF(usuario_creador,''), 'sistema')
		ORDER BY ultima DESC, eventos DESC
		LIMIT ?`, empresaID, time.Now().Add(-2*time.Hour).Format("2006-01-02 15:04:05"), limit)
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := make([]string, 0, limit)
	for rows.Next() {
		var usuario, ultima string
		var eventos, errores int64
		if err := rows.Scan(&usuario, &eventos, &errores, &ultima); err != nil {
			return nil
		}
		out = append(out, fmt.Sprintf("usuario=%s eventos=%d errores=%d ultima=%s", safeAuditoriaAIValue(usuario), eventos, errores, safeAuditoriaAIValue(ultima)))
	}
	return out
}

func superAuditoriaAIActiveUsers(dbConn *sql.DB, limit int) []string {
	rows, err := querySQLCompat(dbConn, `SELECT
		COALESCE(NULLIF(usuario_creador,''), 'sistema') AS usuario,
		COUNT(1) AS eventos,
		COUNT(DISTINCT empresa_id) AS empresas,
		COALESCE(SUM(CASE WHEN LOWER(COALESCE(resultado,'')) = 'error' OR COALESCE(codigo_http,0) >= 400 THEN 1 ELSE 0 END), 0) AS errores,
		COALESCE(MAX(COALESCE(fecha_evento, fecha_creacion, '')), '') AS ultima
		FROM empresa_auditoria_eventos
		WHERE COALESCE(estado, 'activo') = 'activo'
		  AND COALESCE(fecha_evento, fecha_creacion, '') >= ?
		GROUP BY COALESCE(NULLIF(usuario_creador,''), 'sistema')
		ORDER BY ultima DESC, eventos DESC
		LIMIT ?`, time.Now().Add(-2*time.Hour).Format("2006-01-02 15:04:05"), limit)
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := make([]string, 0, limit)
	for rows.Next() {
		var usuario, ultima string
		var eventos, empresas, errores int64
		if err := rows.Scan(&usuario, &eventos, &empresas, &errores, &ultima); err != nil {
			return nil
		}
		out = append(out, fmt.Sprintf("usuario=%s eventos=%d empresas=%d errores=%d ultima=%s", safeAuditoriaAIValue(usuario), eventos, empresas, errores, safeAuditoriaAIValue(ultima)))
	}
	return out
}

func auditAIQuestionFilter(pregunta string) EmpresaAuditoriaEventoFilter {
	folded := aiFoldText(pregunta)
	filter := EmpresaAuditoriaEventoFilter{IncludeInactive: false}
	for _, candidate := range []struct {
		modulo string
		tokens []string
	}{
		{"ventas", []string{"ventas", "venta", "vender"}},
		{"carritos", []string{"carritos", "carrito", "cierre de venta", "cerrar venta"}},
		{"venta_publica", []string{"venta_publica", "venta publica", "ecommerce", "tienda publica"}},
		{"inventario", []string{"inventario", "stock", "existencias"}},
		{"productos_import_export", []string{"productos_import_export", "importar productos", "importacion productos", "exportar productos", "plantilla productos"}},
		{"bodegas_traslados", []string{"bodegas", "bodega", "traslado", "trasladar producto", "transferir bodega"}},
		{"finanzas", []string{"finanzas", "ingresos", "egresos", "caja", "corte de caja"}},
		{"breb_qr", []string{"breb", "bre-b", "qr breb", "pago qr", "pagos qr", "transferencia bre"}},
		{"contabilidad_colombia", []string{"contabilidad_colombia", "contabilidad colombia", "niif", "dian"}},
		{"clientes", []string{"clientes", "cliente"}},
		{"crm_unificado", []string{"crm_unificado", "crm", "leads", "seguimientos"}},
		{"usuarios", []string{"usuarios", "usuario"}},
		{"compras", []string{"compras", "proveedores", "ordenes de compra"}},
		{"facturacion", []string{"facturacion", "factura", "facturas", "comprobante", "nota credito"}},
		{"reportes", []string{"reportes", "reporte", "informe", "exportacion"}},
		{"auditoria", []string{"auditoria", "auditar", "actividad"}},
		{"backups", []string{"backups", "backup", "copia", "copias", "respaldo"}},
		{"documentos_onlyoffice", []string{"documentos_onlyoffice", "onlyoffice", "documentos", "ofimatica"}},
		{"buzon", []string{"buzon", "buzón", "mensaje interno", "mensajeria interna", "notificacion", "campana"}},
		{"tareas_buzon", []string{"tareas_buzon", "tarea", "tareas internas", "asignar tarea", "finalizar tarea"}},
		{"chat_empresarial", []string{"chat empresarial", "chat interno", "chat usuarios"}},
		{"impresoras", []string{"impresoras", "impresora", "agente local", "cola impresion", "reglas de impresion"}},
		{"menu_visible", []string{"menu visible", "ocultar modulos", "ocultar modulo", "visibilidad menu"}},
		{"atajos_pos", []string{"atajos pos", "teclas funcion", "teclado pos", "f1", "f10", "f12"}},
		{"tickets_ayuda", []string{"tickets_ayuda", "ticket ayuda", "tickets de ayuda", "soporte"}},
		{"mantenimiento_programado", []string{"mantenimiento_programado", "mantenimiento programado"}},
		{"licencias", []string{"licencias", "licencia"}},
		{"propinas", []string{"propinas", "propina"}},
		{"comisiones", []string{"comisiones", "comision"}},
		{"gimnasio", []string{"gimnasio", "socios", "membresias"}},
		{"parqueadero", []string{"parqueadero", "parking"}},
		{"domicilios", []string{"domicilios", "delivery"}},
		{"turnos_atencion", []string{"turnos_atencion", "turnos", "fila"}},
		{"control_electrico", []string{"control_electrico", "control electrico", "rele", "raspberry"}},
		{"seguridad", []string{"seguridad", "permisos", "roles"}},
	} {
		for _, token := range candidate.tokens {
			if strings.Contains(folded, token) {
				filter.Modulo = candidate.modulo
				break
			}
		}
		if filter.Modulo != "" {
			break
		}
	}
	switch {
	case strings.Contains(folded, "error") || strings.Contains(folded, "fallo") || strings.Contains(folded, "falla") || strings.Contains(folded, "403") || strings.Contains(folded, "500"):
		filter.Resultado = "error"
	case strings.Contains(folded, "correct") || strings.Contains(folded, "exitos") || strings.Contains(folded, "ok"):
		filter.Resultado = "ok"
	}
	for _, candidate := range []struct {
		token  string
		action string
	}{
		{"crear", "crear"},
		{"creo", "crear"},
		{"actualizar", "actualizar"},
		{"actualizo", "actualizar"},
		{"editar", "actualizar"},
		{"edito", "actualizar"},
		{"eliminar", "eliminar"},
		{"elimino", "eliminar"},
		{"borrar", "eliminar"},
		{"aprobar", "aprobar"},
		{"aprobo", "aprobar"},
		{"leer", "leer"},
		{"consultar", "leer"},
		{"consulto", "leer"},
		{"listar", "leer"},
		{"ver", "leer"},
	} {
		if strings.Contains(folded, candidate.token) {
			filter.Accion = candidate.action
			break
		}
	}
	if terms := aiExtractSearchTerms(pregunta); len(terms) > 0 {
		filter.Search = terms[0]
	} else if auditAIQuestionNeedsDeepSearch(pregunta) {
		filter.Search = auditAICompactSearch(pregunta)
	}
	return filter
}

func auditAIQuestionNeedsDeepSearch(pregunta string) bool {
	folded := aiFoldText(pregunta)
	for _, token := range []string{"auditoria", "auditar", "audit", "quien", "usuario", "usuarios", "hizo", "hicieron", "actividad", "movimiento", "movimientos", "error", "fallo", "falla", "endpoint", "request", "ip"} {
		if strings.Contains(folded, token) {
			return true
		}
	}
	return len(aiExtractSearchTerms(pregunta)) > 0
}

func auditAIQuestionMentionsUsers(folded string) bool {
	return strings.Contains(folded, "usuario") || strings.Contains(folded, "usuarios") || strings.Contains(folded, "quien") || strings.Contains(folded, "hizo") || strings.Contains(folded, "hicieron")
}

func auditAICompactSearch(pregunta string) string {
	folded := aiFoldText(pregunta)
	stop := map[string]bool{
		"que": true, "quien": true, "cuando": true, "como": true, "para": true, "con": true, "por": true, "los": true, "las": true, "una": true, "uno": true, "del": true, "desde": true,
		"auditoria": true, "auditar": true, "actividad": true, "usuario": true, "usuarios": true, "sistema": true, "empresa": true, "base": true, "datos": true,
	}
	parts := strings.Fields(folded)
	for _, p := range parts {
		p = strings.Trim(p, ".,;:()[]{}¿?¡!\"'")
		if len([]rune(p)) < 4 || stop[p] {
			continue
		}
		return p
	}
	return ""
}

func normalizeAuditAIWindow(window time.Duration) time.Duration {
	if window <= 0 {
		return 30 * time.Minute
	}
	if window > 24*time.Hour {
		return 24 * time.Hour
	}
	return window
}

func normalizeAuditAILimit(limit, fallback, max int) int {
	if limit <= 0 {
		limit = fallback
	}
	if max > 0 && limit > max {
		limit = max
	}
	return limit
}

func registerAuditoriaIAConsultaNoBloqueante(dbConn *sql.DB, empresaID int64, alcance, pregunta, usuario, modelo, contexto string, eventosConsultados int64, filtros map[string]interface{}) {
	if dbConn == nil {
		return
	}
	if empresaID < 0 {
		empresaID = 0
	}
	if err := ensureEmpresaAuditoriaIAConsultaSchema(dbConn); err != nil {
		log.Printf("[auditoria_ia] no se pudo asegurar esquema empresa_id=%d alcance=%s error=%v", empresaID, alcance, err)
		return
	}
	pregunta = strings.TrimSpace(pregunta)
	sum := sha256.Sum256([]byte(pregunta))
	preguntaHash := fmt.Sprintf("%x", sum[:])
	if len([]rune(pregunta)) > 300 {
		pregunta = string([]rune(pregunta)[:300])
	}
	if usuario = strings.TrimSpace(usuario); usuario == "" {
		usuario = "sistema"
	}
	if modelo = strings.TrimSpace(modelo); modelo == "" {
		modelo = "modelo_no_especificado"
	}
	filtrosJSON, _ := json.Marshal(filtros)
	resultadosJSON, _ := json.Marshal(map[string]interface{}{
		"eventos_consultados": eventosConsultados,
		"contexto_caracteres": len([]rune(contexto)),
		"incluye_auditoria":   strings.Contains(contexto, "AUDITORIA_"),
	})
	if _, err := insertSQLCompat(dbConn, `INSERT INTO empresa_auditoria_ia_consultas (
		empresa_id, alcance, modelo, usuario_consulta, pregunta_hash, pregunta_resumen,
		filtros_json, resultados_json, eventos_consultados, contexto_caracteres,
		resultado, fecha_consulta, fecha_creacion, fecha_actualizacion, usuario_creador, estado, observaciones
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'ok', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, ?, 'activo', ?)`,
		empresaID,
		sanitizeAuditoriaText(strings.TrimSpace(alcance), 80),
		sanitizeAuditoriaText(modelo, 120),
		sanitizeAuditoriaText(usuario, 120),
		preguntaHash,
		sanitizeAuditoriaText(pregunta, 300),
		string(filtrosJSON),
		string(resultadosJSON),
		eventosConsultados,
		len([]rune(contexto)),
		sanitizeAuditoriaText(usuario, 120),
		"consulta de auditoria generada para contexto IA",
	); err != nil {
		log.Printf("[auditoria_ia] no se pudo registrar consulta empresa_id=%d alcance=%s error=%v", empresaID, alcance, err)
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

func normalizeAuditoriaSeveridad(raw string) string {
	v := normalizeAuditoriaValue(raw)
	switch v {
	case "critica", "critical", "critico":
		return auditoriaSeveridadCritica
	case "alta", "high":
		return auditoriaSeveridadAlta
	case "media", "medium":
		return auditoriaSeveridadMedia
	case "baja", "low":
		return auditoriaSeveridadBaja
	default:
		return ""
	}
}

func resolveAuditoriaSeveridad(modulo, accion, resultado string, codigoHTTP int64, metadataJSON string) string {
	if metadataSeverity := extractAuditoriaSeverityFromMetadata(metadataJSON); metadataSeverity != "" {
		return metadataSeverity
	}

	modulo = normalizeAuditoriaValue(modulo)
	accion = normalizeAuditoriaValue(accion)
	resultado = normalizeAuditoriaResultado(resultado)

	switch {
	case codigoHTTP >= 500:
		return auditoriaSeveridadCritica
	case modulo == "seguridad" && (accion == "eliminar" || accion == "rotar_credencial" || accion == "desactivar"):
		return auditoriaSeveridadCritica
	case resultado == "error" && codigoHTTP >= 400:
		return auditoriaSeveridadAlta
	case modulo == "seguridad" || modulo == "auditoria" || modulo == "backups" || modulo == "licencias" || modulo == "menu_visible":
		return auditoriaSeveridadAlta
	case modulo == "finanzas" || modulo == "breb_qr" || modulo == "facturacion" || modulo == "compras" || modulo == "nomina" || modulo == "documentos_onlyoffice" || modulo == "buzon" || modulo == "tareas_buzon" || modulo == "chat_empresarial" || modulo == "impresoras" || modulo == "productos_import_export" || modulo == "bodegas_traslados" || modulo == "atajos_pos":
		return auditoriaSeveridadMedia
	default:
		return auditoriaSeveridadBaja
	}
}

func extractAuditoriaSeverityFromMetadata(metadataJSON string) string {
	metadataJSON = strings.TrimSpace(metadataJSON)
	if metadataJSON == "" || !json.Valid([]byte(metadataJSON)) {
		return ""
	}

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(metadataJSON), &payload); err != nil {
		return ""
	}
	if len(payload) == 0 {
		return ""
	}

	for _, key := range []string{"severidad", "severity", "nivel_severidad"} {
		if raw, ok := payload[key]; ok {
			if sev := normalizeAuditoriaSeveridad(fmt.Sprint(raw)); sev != "" {
				return sev
			}
		}
	}

	return ""
}

func resolveAuditoriaPoliticaRetencionDias(modulo, severidad string) int64 {
	modulo = normalizeAuditoriaValue(modulo)
	severidad = normalizeAuditoriaSeveridad(severidad)
	if severidad == "" {
		severidad = auditoriaSeveridadBaja
	}

	switch modulo {
	case "seguridad", "auditoria", "backups", "licencias", "menu_visible":
		switch severidad {
		case auditoriaSeveridadCritica:
			return 3650
		case auditoriaSeveridadAlta:
			return 1825
		case auditoriaSeveridadMedia:
			return 1095
		default:
			return 365
		}
	case "finanzas", "breb_qr":
		switch severidad {
		case auditoriaSeveridadCritica:
			return 3650
		case auditoriaSeveridadAlta:
			return 1825
		case auditoriaSeveridadMedia:
			return 730
		default:
			return 365
		}
	case "facturacion", "compras", "nomina", "documentos_onlyoffice", "buzon", "tareas_buzon", "chat_empresarial", "impresoras", "productos_import_export", "bodegas_traslados", "atajos_pos":
		switch severidad {
		case auditoriaSeveridadCritica:
			return 1825
		case auditoriaSeveridadAlta:
			return 1095
		case auditoriaSeveridadMedia:
			return 730
		default:
			return 365
		}
	default:
		switch severidad {
		case auditoriaSeveridadCritica:
			return 1825
		case auditoriaSeveridadAlta:
			return 1095
		case auditoriaSeveridadMedia:
			return 365
		default:
			return defaultEmpresaAuditoriaRetencionDias
		}
	}
}

func enrichAuditoriaRetentionMetadata(metadataJSON, modulo, severidad string, retencionDias int64) string {
	metadataJSON = strings.TrimSpace(metadataJSON)
	if metadataJSON == "" || !json.Valid([]byte(metadataJSON)) {
		return metadataJSON
	}

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(metadataJSON), &payload); err != nil {
		return metadataJSON
	}
	if payload == nil {
		payload = map[string]interface{}{}
	}

	if _, exists := payload["severidad"]; !exists {
		payload["severidad"] = severidad
	}
	payload["retencion_politica_modulo"] = normalizeAuditoriaValue(modulo)
	payload["retencion_politica_severidad"] = severidad
	payload["retencion_dias_resuelto"] = retencionDias

	encoded, err := json.Marshal(payload)
	if err != nil {
		return metadataJSON
	}
	return string(encoded)
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
