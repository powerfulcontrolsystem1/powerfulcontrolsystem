package db

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type SuperErrorSistema struct {
	ID                 int64  `json:"id"`
	Nivel              string `json:"nivel"`
	TipoError          string `json:"tipo_error"`
	Mensaje            string `json:"mensaje"`
	MensajePublico     string `json:"mensaje_publico"`
	Detalle            string `json:"detalle,omitempty"`
	StackTrace         string `json:"stack_trace,omitempty"`
	EmpresaID          int64  `json:"empresa_id,omitempty"`
	UsuarioEmail       string `json:"usuario_email,omitempty"`
	Endpoint           string `json:"endpoint,omitempty"`
	Modulo             string `json:"modulo,omitempty"`
	MetodoHTTP         string `json:"metodo_http,omitempty"`
	CodigoHTTP         int    `json:"codigo_http,omitempty"`
	RequestID          string `json:"request_id,omitempty"`
	Origen             string `json:"origen,omitempty"`
	IP                 string `json:"ip,omitempty"`
	UserAgent          string `json:"user_agent,omitempty"`
	MetadataJSON       string `json:"metadata_json,omitempty"`
	FechaError         string `json:"fecha_error,omitempty"`
	FechaCreacion      string `json:"fecha_creacion,omitempty"`
	FechaActualizacion string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string `json:"usuario_creador,omitempty"`
	Estado             string `json:"estado,omitempty"`
	Observaciones      string `json:"observaciones,omitempty"`
}

type SuperErrorSistemaFiltro struct {
	EmpresaID int64
	Nivel     string
	TipoError string
	Desde     string
	Hasta     string
	Search    string
	Limit     int
	Offset    int
}

type SuperErrorSistemaResumen struct {
	Total          int64 `json:"total"`
	Info           int64 `json:"info"`
	Warning        int64 `json:"warning"`
	Error          int64 `json:"error"`
	Critical       int64 `json:"critical"`
	Ultimas24Horas int64 `json:"ultimas_24_horas"`
}

func normalizeSuperErrorLevel(raw string) string {
	v := strings.ToUpper(strings.TrimSpace(raw))
	switch v {
	case "INFO":
		return "INFO"
	case "WARN", "WARNING":
		return "WARNING"
	case "ERR", "ERROR":
		return "ERROR"
	case "CRITICAL", "FATAL", "PANIC":
		return "CRITICAL"
	default:
		return ""
	}
}

func normalizeSuperErrorType(raw string) string {
	v := strings.ToLower(strings.TrimSpace(raw))
	v = strings.ReplaceAll(v, "-", "_")
	v = strings.ReplaceAll(v, " ", "_")
	for strings.Contains(v, "__") {
		v = strings.ReplaceAll(v, "__", "_")
	}
	return strings.Trim(v, "_")
}

func normalizeSuperErrorDateFilter(raw string, endOfDay bool) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	layouts := []string{time.RFC3339, "2006-01-02 15:04:05", "2006-01-02T15:04", "2006-01-02"}
	for _, layout := range layouts {
		parsed, err := time.Parse(layout, raw)
		if err != nil {
			continue
		}
		if layout == "2006-01-02" {
			if endOfDay {
				parsed = parsed.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			}
		}
		return parsed.Format("2006-01-02 15:04:05")
	}
	return ""
}

func normalizeSuperErrorFiltro(filter SuperErrorSistemaFiltro) SuperErrorSistemaFiltro {
	filter.Nivel = normalizeSuperErrorLevel(filter.Nivel)
	filter.TipoError = normalizeSuperErrorType(filter.TipoError)
	filter.Search = strings.TrimSpace(filter.Search)
	filter.Desde = normalizeSuperErrorDateFilter(filter.Desde, false)
	filter.Hasta = normalizeSuperErrorDateFilter(filter.Hasta, true)
	if filter.Limit <= 0 {
		filter.Limit = 50
	}
	if filter.Limit > 200 {
		filter.Limit = 200
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}
	return filter
}

func EnsureSuperErroresSistemaSchema(dbConn *sql.DB) error {
	if SchemaBootstrapDisabled() {
		return nil
	}
	if dbConn == nil {
		return fmt.Errorf("db connection is required")
	}

	if isPostgresDialect() {
		if _, err := execSQLCompat(dbConn, `CREATE TABLE IF NOT EXISTS super_errores_sistema (
			id BIGSERIAL PRIMARY KEY,
			nivel TEXT NOT NULL,
			tipo_error TEXT,
			mensaje TEXT NOT NULL,
			mensaje_publico TEXT,
			detalle TEXT,
			stack_trace TEXT,
			empresa_id BIGINT,
			usuario_email TEXT,
			endpoint TEXT,
			modulo TEXT,
			metodo_http TEXT,
			codigo_http INTEGER,
			request_id TEXT,
			origen TEXT,
			ip TEXT,
			user_agent TEXT,
			metadata_json TEXT,
			fecha_error TEXT DEFAULT (CAST(CURRENT_TIMESTAMP AS TEXT)),
			fecha_creacion TEXT DEFAULT (CAST(CURRENT_TIMESTAMP AS TEXT)),
			fecha_actualizacion TEXT DEFAULT (CAST(CURRENT_TIMESTAMP AS TEXT)),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		)`); err != nil {
			return err
		}
	} else {
		if _, err := execSQLCompat(dbConn, `CREATE TABLE IF NOT EXISTS super_errores_sistema (
			id BIGSERIAL PRIMARY KEY,
			nivel TEXT NOT NULL,
			tipo_error TEXT,
			mensaje TEXT NOT NULL,
			mensaje_publico TEXT,
			detalle TEXT,
			stack_trace TEXT,
			empresa_id INTEGER,
			usuario_email TEXT,
			endpoint TEXT,
			modulo TEXT,
			metodo_http TEXT,
			codigo_http INTEGER,
			request_id TEXT,
			origen TEXT,
			ip TEXT,
			user_agent TEXT,
			metadata_json TEXT,
			fecha_error TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		)`); err != nil {
			return err
		}
	}

	columns := []struct {
		name string
		def  string
	}{
		{"nivel", "TEXT NOT NULL DEFAULT 'ERROR'"},
		{"tipo_error", "TEXT"},
		{"mensaje", "TEXT"},
		{"mensaje_publico", "TEXT"},
		{"detalle", "TEXT"},
		{"stack_trace", "TEXT"},
		{"empresa_id", "INTEGER"},
		{"usuario_email", "TEXT"},
		{"endpoint", "TEXT"},
		{"modulo", "TEXT"},
		{"metodo_http", "TEXT"},
		{"codigo_http", "INTEGER"},
		{"request_id", "TEXT"},
		{"origen", "TEXT"},
		{"ip", "TEXT"},
		{"user_agent", "TEXT"},
		{"metadata_json", "TEXT"},
		{"fecha_error", "TEXT DEFAULT (CURRENT_TIMESTAMP)"},
		{"fecha_creacion", "TEXT DEFAULT (CURRENT_TIMESTAMP)"},
		{"fecha_actualizacion", "TEXT DEFAULT (CURRENT_TIMESTAMP)"},
		{"usuario_creador", "TEXT"},
		{"estado", "TEXT DEFAULT 'activo'"},
		{"observaciones", "TEXT"},
	}
	for _, column := range columns {
		if err := ensureColumnIfMissing(dbConn, "super_errores_sistema", column.name, column.def); err != nil {
			return err
		}
	}

	indices := []string{
		"CREATE INDEX IF NOT EXISTS ix_super_errores_sistema_fecha ON super_errores_sistema(fecha_error DESC, id DESC)",
		"CREATE INDEX IF NOT EXISTS ix_super_errores_sistema_nivel_fecha ON super_errores_sistema(nivel, fecha_error DESC)",
		"CREATE INDEX IF NOT EXISTS ix_super_errores_sistema_empresa_fecha ON super_errores_sistema(empresa_id, fecha_error DESC)",
		"CREATE INDEX IF NOT EXISTS ix_super_errores_sistema_tipo_fecha ON super_errores_sistema(tipo_error, fecha_error DESC)",
		"CREATE INDEX IF NOT EXISTS ix_super_errores_sistema_request ON super_errores_sistema(request_id)",
	}
	for _, stmt := range indices {
		if _, err := execSQLCompat(dbConn, stmt); err != nil {
			return err
		}
	}

	return nil
}

func CreateSuperErrorSistema(dbConn *sql.DB, payload SuperErrorSistema) (int64, error) {
	if dbConn == nil {
		return 0, fmt.Errorf("db connection is required")
	}
	if err := EnsureSuperErroresSistemaSchema(dbConn); err != nil {
		return 0, err
	}

	payload.Nivel = normalizeSuperErrorLevel(payload.Nivel)
	if payload.Nivel == "" {
		payload.Nivel = "ERROR"
	}
	payload.TipoError = normalizeSuperErrorType(payload.TipoError)
	payload.Mensaje = strings.TrimSpace(payload.Mensaje)
	payload.MensajePublico = strings.TrimSpace(payload.MensajePublico)
	payload.Detalle = strings.TrimSpace(payload.Detalle)
	payload.StackTrace = strings.TrimSpace(payload.StackTrace)
	payload.UsuarioEmail = strings.TrimSpace(payload.UsuarioEmail)
	payload.Endpoint = strings.TrimSpace(payload.Endpoint)
	payload.Modulo = strings.TrimSpace(payload.Modulo)
	payload.MetodoHTTP = strings.TrimSpace(payload.MetodoHTTP)
	payload.RequestID = strings.TrimSpace(payload.RequestID)
	payload.Origen = strings.TrimSpace(payload.Origen)
	payload.IP = strings.TrimSpace(payload.IP)
	payload.UserAgent = strings.TrimSpace(payload.UserAgent)
	payload.MetadataJSON = strings.TrimSpace(payload.MetadataJSON)
	payload.FechaError = normalizeSuperErrorDateFilter(payload.FechaError, false)
	payload.UsuarioCreador = strings.TrimSpace(payload.UsuarioCreador)
	payload.Estado = strings.TrimSpace(payload.Estado)
	payload.Observaciones = strings.TrimSpace(payload.Observaciones)

	if payload.Mensaje == "" {
		payload.Mensaje = "Error del sistema sin detalle adicional"
	}
	if payload.FechaError == "" {
		payload.FechaError = time.Now().Format("2006-01-02 15:04:05")
	}
	if payload.UsuarioCreador == "" {
		payload.UsuarioCreador = "sistema"
	}
	if payload.Estado == "" {
		payload.Estado = "activo"
	}

	nowExpr := sqlNowExpr()
	query := `INSERT INTO super_errores_sistema (
		nivel,
		tipo_error,
		mensaje,
		mensaje_publico,
		detalle,
		stack_trace,
		empresa_id,
		usuario_email,
		endpoint,
		modulo,
		metodo_http,
		codigo_http,
		request_id,
		origen,
		ip,
		user_agent,
		metadata_json,
		fecha_error,
		fecha_creacion,
		fecha_actualizacion,
		usuario_creador,
		estado,
		observaciones
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ` + nowExpr + `, ` + nowExpr + `, ?, ?, ?)`

	return insertSQLCompat(dbConn, query,
		payload.Nivel,
		payload.TipoError,
		payload.Mensaje,
		payload.MensajePublico,
		payload.Detalle,
		payload.StackTrace,
		payload.EmpresaID,
		payload.UsuarioEmail,
		payload.Endpoint,
		payload.Modulo,
		payload.MetodoHTTP,
		payload.CodigoHTTP,
		payload.RequestID,
		payload.Origen,
		payload.IP,
		payload.UserAgent,
		payload.MetadataJSON,
		payload.FechaError,
		payload.UsuarioCreador,
		payload.Estado,
		payload.Observaciones,
	)
}

func buildSuperErroresSistemaWhere(filter SuperErrorSistemaFiltro) (string, []interface{}) {
	clauses := make([]string, 0, 6)
	args := make([]interface{}, 0, 12)

	if filter.EmpresaID > 0 {
		clauses = append(clauses, "empresa_id = ?")
		args = append(args, filter.EmpresaID)
	}
	if filter.Nivel != "" {
		clauses = append(clauses, "UPPER(COALESCE(nivel, '')) = ?")
		args = append(args, filter.Nivel)
	}
	if filter.TipoError != "" {
		clauses = append(clauses, "LOWER(COALESCE(tipo_error, '')) = ?")
		args = append(args, filter.TipoError)
	}
	if filter.Desde != "" {
		clauses = append(clauses, "COALESCE(fecha_error, fecha_creacion, '') >= ?")
		args = append(args, filter.Desde)
	}
	if filter.Hasta != "" {
		clauses = append(clauses, "COALESCE(fecha_error, fecha_creacion, '') <= ?")
		args = append(args, filter.Hasta)
	}
	if filter.Search != "" {
		like := "%" + strings.ToLower(filter.Search) + "%"
		clauses = append(clauses, `(LOWER(COALESCE(tipo_error, '')) LIKE ? OR LOWER(COALESCE(mensaje, '')) LIKE ? OR LOWER(COALESCE(mensaje_publico, '')) LIKE ? OR LOWER(COALESCE(detalle, '')) LIKE ? OR LOWER(COALESCE(endpoint, '')) LIKE ? OR LOWER(COALESCE(modulo, '')) LIKE ? OR LOWER(COALESCE(usuario_email, '')) LIKE ? OR LOWER(COALESCE(request_id, '')) LIKE ?)`)
		for i := 0; i < 8; i++ {
			args = append(args, like)
		}
	}

	if len(clauses) == 0 {
		return "", args
	}
	return " WHERE " + strings.Join(clauses, " AND "), args
}

func ListSuperErroresSistema(dbConn *sql.DB, filter SuperErrorSistemaFiltro) ([]SuperErrorSistema, int64, SuperErrorSistemaResumen, error) {
	var summary SuperErrorSistemaResumen
	if dbConn == nil {
		return nil, 0, summary, fmt.Errorf("db connection is required")
	}
	if err := EnsureSuperErroresSistemaSchema(dbConn); err != nil {
		return nil, 0, summary, err
	}

	filter = normalizeSuperErrorFiltro(filter)
	whereSQL, args := buildSuperErroresSistemaWhere(filter)

	countQuery := "SELECT COUNT(*) FROM super_errores_sistema" + whereSQL
	if err := queryRowSQLCompat(dbConn, countQuery, args...).Scan(&summary.Total); err != nil {
		return nil, 0, summary, err
	}

	statsQuery := `SELECT
		COALESCE(SUM(CASE WHEN UPPER(COALESCE(nivel, '')) = 'INFO' THEN 1 ELSE 0 END), 0),
		COALESCE(SUM(CASE WHEN UPPER(COALESCE(nivel, '')) = 'WARNING' THEN 1 ELSE 0 END), 0),
		COALESCE(SUM(CASE WHEN UPPER(COALESCE(nivel, '')) = 'ERROR' THEN 1 ELSE 0 END), 0),
		COALESCE(SUM(CASE WHEN UPPER(COALESCE(nivel, '')) = 'CRITICAL' THEN 1 ELSE 0 END), 0),
		COALESCE(SUM(CASE WHEN COALESCE(fecha_error, fecha_creacion, '') >= ? THEN 1 ELSE 0 END), 0)
	FROM super_errores_sistema` + whereSQL
	statsArgs := append([]interface{}{time.Now().Add(-24 * time.Hour).Format("2006-01-02 15:04:05")}, args...)
	if err := queryRowSQLCompat(dbConn, statsQuery, statsArgs...).Scan(&summary.Info, &summary.Warning, &summary.Error, &summary.Critical, &summary.Ultimas24Horas); err != nil {
		return nil, 0, summary, err
	}

	query := `SELECT
		id,
		COALESCE(nivel, ''),
		COALESCE(tipo_error, ''),
		COALESCE(mensaje, ''),
		COALESCE(mensaje_publico, ''),
		COALESCE(detalle, ''),
		COALESCE(stack_trace, ''),
		COALESCE(empresa_id, 0),
		COALESCE(usuario_email, ''),
		COALESCE(endpoint, ''),
		COALESCE(modulo, ''),
		COALESCE(metodo_http, ''),
		COALESCE(codigo_http, 0),
		COALESCE(request_id, ''),
		COALESCE(origen, ''),
		COALESCE(ip, ''),
		COALESCE(user_agent, ''),
		COALESCE(metadata_json, ''),
		COALESCE(fecha_error, ''),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, ''),
		COALESCE(observaciones, '')
	FROM super_errores_sistema` + whereSQL + ` ORDER BY COALESCE(fecha_error, fecha_creacion, '') DESC, id DESC LIMIT ? OFFSET ?`
	listArgs := append(append([]interface{}{}, args...), filter.Limit, filter.Offset)
	rows, err := querySQLCompat(dbConn, query, listArgs...)
	if err != nil {
		return nil, 0, summary, err
	}
	defer rows.Close()

	items := make([]SuperErrorSistema, 0, filter.Limit)
	for rows.Next() {
		var item SuperErrorSistema
		var empresaID int64
		if err := rows.Scan(
			&item.ID,
			&item.Nivel,
			&item.TipoError,
			&item.Mensaje,
			&item.MensajePublico,
			&item.Detalle,
			&item.StackTrace,
			&empresaID,
			&item.UsuarioEmail,
			&item.Endpoint,
			&item.Modulo,
			&item.MetodoHTTP,
			&item.CodigoHTTP,
			&item.RequestID,
			&item.Origen,
			&item.IP,
			&item.UserAgent,
			&item.MetadataJSON,
			&item.FechaError,
			&item.FechaCreacion,
			&item.FechaActualizacion,
			&item.UsuarioCreador,
			&item.Estado,
			&item.Observaciones,
		); err != nil {
			return nil, 0, summary, err
		}
		if empresaID > 0 {
			item.EmpresaID = empresaID
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, summary, err
	}

	return items, summary.Total, summary, nil
}

// ResetSuperErroresSistema limpia los indicadores del monitor centralizado de errores.
func ResetSuperErroresSistema(dbConn *sql.DB) (int64, error) {
	if dbConn == nil {
		return 0, fmt.Errorf("db connection is required")
	}
	if err := EnsureSuperErroresSistemaSchema(dbConn); err != nil {
		return 0, err
	}
	res, err := execSQLCompat(dbConn, "DELETE FROM super_errores_sistema")
	if err != nil {
		return 0, err
	}
	affected, _ := res.RowsAffected()
	return affected, nil
}
