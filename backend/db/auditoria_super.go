package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
)

const defaultSuperAuditoriaRetencionDias = int64(365)

// SuperAuditoriaEvento representa trazabilidad del selector global y modulos super.
type SuperAuditoriaEvento struct {
	ID                 int64  `json:"id"`
	EmpresaID          int64  `json:"empresa_id"`
	PrincipalEmail     string `json:"principal_email"`
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

// SuperAuditoriaEventoFilter permite consultar auditoria global por alcance.
type SuperAuditoriaEventoFilter struct {
	EmpresaID       int64
	PrincipalEmail  string
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

// EnsureSuperAuditoriaSchema crea la bitacora global del selector y modulos super.
func EnsureSuperAuditoriaSchema(dbConn *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS super_auditoria_eventos (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER DEFAULT 0,
			principal_email TEXT,
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
			retencion_dias INTEGER DEFAULT 365,
			fecha_evento TEXT DEFAULT (datetime('now','localtime')),
			fecha_expiracion TEXT,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE INDEX IF NOT EXISTS ix_super_auditoria_fecha ON super_auditoria_eventos(fecha_evento DESC, id DESC);`,
		`CREATE INDEX IF NOT EXISTS ix_super_auditoria_principal_fecha ON super_auditoria_eventos(principal_email, fecha_evento DESC, id DESC);`,
		`CREATE INDEX IF NOT EXISTS ix_super_auditoria_empresa_fecha ON super_auditoria_eventos(empresa_id, fecha_evento DESC, id DESC);`,
		`CREATE INDEX IF NOT EXISTS ix_super_auditoria_modulo_accion ON super_auditoria_eventos(modulo, accion);`,
		`CREATE INDEX IF NOT EXISTS ix_super_auditoria_usuario_fecha ON super_auditoria_eventos(usuario_creador, fecha_evento DESC);`,
		`CREATE INDEX IF NOT EXISTS ix_super_auditoria_request_id ON super_auditoria_eventos(request_id);`,
		`CREATE INDEX IF NOT EXISTS ix_super_auditoria_http_resultado ON super_auditoria_eventos(codigo_http, resultado);`,
	}
	for _, stmt := range stmts {
		if _, err := dbConn.Exec(stmt); err != nil {
			return err
		}
	}
	columns := []struct {
		name string
		def  string
	}{
		{"empresa_id", "INTEGER DEFAULT 0"},
		{"principal_email", "TEXT"},
		{"modulo", "TEXT"},
		{"accion", "TEXT"},
		{"recurso", "TEXT"},
		{"recurso_id", "INTEGER"},
		{"metodo_http", "TEXT"},
		{"endpoint", "TEXT"},
		{"resultado", "TEXT DEFAULT 'ok'"},
		{"codigo_http", "INTEGER DEFAULT 0"},
		{"request_id", "TEXT"},
		{"ip_origen", "TEXT"},
		{"user_agent", "TEXT"},
		{"metadata_json", "TEXT"},
		{"retencion_dias", "INTEGER DEFAULT 365"},
		{"fecha_evento", "TEXT DEFAULT (datetime('now','localtime'))"},
		{"fecha_expiracion", "TEXT"},
		{"fecha_actualizacion", "TEXT DEFAULT (datetime('now','localtime'))"},
		{"usuario_creador", "TEXT"},
		{"estado", "TEXT DEFAULT 'activo'"},
		{"observaciones", "TEXT"},
	}
	for _, col := range columns {
		if err := ensureColumnIfMissing(dbConn, "super_auditoria_eventos", col.name, col.def); err != nil {
			return err
		}
	}
	return nil
}

// CreateSuperAuditoriaEvento registra un evento de auditoria del selector/global.
func CreateSuperAuditoriaEvento(dbConn *sql.DB, in SuperAuditoriaEvento) (int64, error) {
	if err := EnsureSuperAuditoriaSchema(dbConn); err != nil {
		return 0, err
	}
	in.PrincipalEmail = strings.ToLower(sanitizeAuditoriaText(strings.TrimSpace(in.PrincipalEmail), 160))
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
	in.UsuarioCreador = strings.ToLower(sanitizeAuditoriaText(strings.TrimSpace(in.UsuarioCreador), 160))
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
	retencionDias := in.RetencionDias
	if retencionDias <= 0 {
		retencionDias = defaultSuperAuditoriaRetencionDias
	}
	retencionDias = normalizeAuditoriaRetencionDias(retencionDias)
	ttlExpr := fmt.Sprintf("+%d days", retencionDias)

	return insertSQLCompat(dbConn, `INSERT INTO super_auditoria_eventos (
		empresa_id, principal_email, modulo, accion, recurso, recurso_id,
		metodo_http, endpoint, resultado, codigo_http, request_id, ip_origen,
		user_agent, metadata_json, retencion_dias, fecha_evento, fecha_expiracion,
		fecha_creacion, fecha_actualizacion, usuario_creador, estado, observaciones
	) VALUES (
		?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?,
		COALESCE(NULLIF(?, ''), datetime('now','localtime')),
		datetime(COALESCE(NULLIF(?, ''), datetime('now','localtime')), ?),
		datetime('now','localtime'), datetime('now','localtime'), ?, ?, ?
	)`,
		in.EmpresaID, in.PrincipalEmail, in.Modulo, in.Accion, in.Recurso, in.RecursoID,
		in.MetodoHTTP, in.Endpoint, in.Resultado, in.CodigoHTTP, in.RequestID, in.IPOrigen,
		in.UserAgent, metadata, retencionDias, strings.TrimSpace(in.FechaEvento),
		strings.TrimSpace(in.FechaEvento), ttlExpr, in.UsuarioCreador, in.Estado, in.Observaciones,
	)
}

// ListSuperAuditoriaEventos lista eventos globales con filtros.
func ListSuperAuditoriaEventos(dbConn *sql.DB, f SuperAuditoriaEventoFilter) ([]SuperAuditoriaEvento, error) {
	if err := EnsureSuperAuditoriaSchema(dbConn); err != nil {
		return nil, err
	}
	where, args := buildSuperAuditoriaWhereClause(f)
	query := `SELECT
		id, COALESCE(empresa_id, 0), COALESCE(principal_email, ''),
		COALESCE(modulo, ''), COALESCE(accion, ''), COALESCE(recurso, ''),
		COALESCE(recurso_id, 0), COALESCE(metodo_http, ''),
		COALESCE(endpoint, ''), COALESCE(resultado, 'ok'), COALESCE(codigo_http, 0),
		COALESCE(request_id, ''), COALESCE(ip_origen, ''), COALESCE(user_agent, ''),
		COALESCE(metadata_json, '{}'), COALESCE(retencion_dias, 365),
		COALESCE(fecha_evento, ''), COALESCE(fecha_expiracion, ''),
		COALESCE(fecha_creacion, ''), COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''), COALESCE(estado, 'activo'), COALESCE(observaciones, '')
	FROM super_auditoria_eventos` + where + `
	ORDER BY COALESCE(fecha_evento, '') DESC, id DESC
	LIMIT ? OFFSET ?`
	args = append(args, normalizeAuditoriaLimit(f.Limit), normalizeAuditoriaOffset(f.Offset))
	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]SuperAuditoriaEvento, 0)
	for rows.Next() {
		var item SuperAuditoriaEvento
		if err := rows.Scan(
			&item.ID, &item.EmpresaID, &item.PrincipalEmail, &item.Modulo,
			&item.Accion, &item.Recurso, &item.RecursoID, &item.MetodoHTTP,
			&item.Endpoint, &item.Resultado, &item.CodigoHTTP, &item.RequestID,
			&item.IPOrigen, &item.UserAgent, &item.MetadataJSON, &item.RetencionDias,
			&item.FechaEvento, &item.FechaExpiracion, &item.FechaCreacion,
			&item.FechaActualizacion, &item.UsuarioCreador, &item.Estado,
			&item.Observaciones,
		); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

// CountSuperAuditoriaEventos cuenta eventos globales con filtros.
func CountSuperAuditoriaEventos(dbConn *sql.DB, f SuperAuditoriaEventoFilter) (int64, error) {
	if err := EnsureSuperAuditoriaSchema(dbConn); err != nil {
		return 0, err
	}
	where, args := buildSuperAuditoriaWhereClause(f)
	var total int64
	if err := dbConn.QueryRow(`SELECT COUNT(1) FROM super_auditoria_eventos`+where, args...).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func buildSuperAuditoriaWhereClause(f SuperAuditoriaEventoFilter) (string, []interface{}) {
	where := ` WHERE 1=1`
	args := []interface{}{}
	if !f.IncludeInactive {
		where += ` AND COALESCE(estado, 'activo') = 'activo'`
	}
	if f.EmpresaID > 0 {
		where += ` AND COALESCE(empresa_id, 0) = ?`
		args = append(args, f.EmpresaID)
	}
	if principal := strings.ToLower(strings.TrimSpace(f.PrincipalEmail)); principal != "" {
		where += ` AND LOWER(COALESCE(principal_email, '')) = LOWER(?)`
		args = append(args, principal)
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
	if f.RecursoID > 0 {
		where += ` AND COALESCE(recurso_id, 0) = ?`
		args = append(args, f.RecursoID)
	}
	if f.CodigoHTTP > 0 {
		where += ` AND COALESCE(codigo_http, 0) = ?`
		args = append(args, f.CodigoHTTP)
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
			LOWER(COALESCE(principal_email, '')) LIKE ? ESCAPE '!'
			OR LOWER(COALESCE(modulo, '')) LIKE ? ESCAPE '!'
			OR LOWER(COALESCE(accion, '')) LIKE ? ESCAPE '!'
			OR LOWER(COALESCE(recurso, '')) LIKE ? ESCAPE '!'
			OR LOWER(COALESCE(endpoint, '')) LIKE ? ESCAPE '!'
			OR LOWER(COALESCE(usuario_creador, '')) LIKE ? ESCAPE '!'
			OR LOWER(COALESCE(request_id, '')) LIKE ? ESCAPE '!'
			OR LOWER(COALESCE(ip_origen, '')) LIKE ? ESCAPE '!'
			OR LOWER(COALESCE(observaciones, '')) LIKE ? ESCAPE '!'
			OR LOWER(COALESCE(metadata_json, '')) LIKE ? ESCAPE '!'
		)`
		args = append(args, searchLike, searchLike, searchLike, searchLike, searchLike, searchLike, searchLike, searchLike, searchLike, searchLike)
	}
	return where, args
}
