package db

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// EmpresaCalculadoraConfiguracion define opciones de integracion del modulo calculadora.
type EmpresaCalculadoraConfiguracion struct {
	ID                   int64  `json:"id"`
	EmpresaID            int64  `json:"empresa_id"`
	IntegrarCarritos     bool   `json:"integrar_carritos"`
	IntegrarCotizaciones bool   `json:"integrar_cotizaciones"`
	FechaCreacion        string `json:"fecha_creacion,omitempty"`
	FechaActualizacion   string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador       string `json:"usuario_creador,omitempty"`
	Estado               string `json:"estado,omitempty"`
	Observaciones        string `json:"observaciones,omitempty"`
}

// EmpresaCalculadoraOperacion representa una operacion registrada por empresa.
type EmpresaCalculadoraOperacion struct {
	ID                 int64    `json:"id"`
	EmpresaID          int64    `json:"empresa_id"`
	Expresion          string   `json:"expresion"`
	Resultado          string   `json:"resultado"`
	Etiquetas          []string `json:"etiquetas,omitempty"`
	ClienteID          int64    `json:"cliente_id,omitempty"`
	ClienteNombre      string   `json:"cliente_nombre,omitempty"`
	DocumentoTipo      string   `json:"documento_tipo,omitempty"`
	DocumentoCodigo    string   `json:"documento_codigo,omitempty"`
	CarritoID          int64    `json:"carrito_id,omitempty"`
	CotizacionID       int64    `json:"cotizacion_id,omitempty"`
	FechaOperacion     string   `json:"fecha_operacion,omitempty"`
	MetadataJSON       string   `json:"metadata_json,omitempty"`
	FechaCreacion      string   `json:"fecha_creacion,omitempty"`
	FechaActualizacion string   `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string   `json:"usuario_creador,omitempty"`
	Estado             string   `json:"estado,omitempty"`
	Observaciones      string   `json:"observaciones,omitempty"`
}

// EmpresaCalculadoraOperacionFilter aplica filtros de consulta sobre historial.
type EmpresaCalculadoraOperacionFilter struct {
	Desde           string
	Hasta           string
	UsuarioCreador  string
	ClienteID       int64
	Etiqueta        string
	IncludeInactive bool
	Limit           int
	Offset          int
}

func defaultEmpresaCalculadoraConfiguracion(empresaID int64) EmpresaCalculadoraConfiguracion {
	return EmpresaCalculadoraConfiguracion{
		EmpresaID:            empresaID,
		IntegrarCarritos:     true,
		IntegrarCotizaciones: true,
		Estado:               "activo",
	}
}

func normalizeEmpresaCalculadoraEstado(raw string) string {
	if strings.EqualFold(strings.TrimSpace(raw), "inactivo") {
		return "inactivo"
	}
	return "activo"
}

func calcNormalizeLimitOffset(limit, offset int) (int, int) {
	if limit <= 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = 1000
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}

func calcLikePattern(raw string) string {
	value := strings.TrimSpace(raw)
	value = strings.ReplaceAll(value, "!", "!!")
	value = strings.ReplaceAll(value, "%", "!%")
	value = strings.ReplaceAll(value, "_", "!_")
	return "%" + value + "%"
}

func calcNormalizeEtiquetas(values []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(values))
	for _, item := range values {
		clean := strings.TrimSpace(item)
		if clean == "" {
			continue
		}
		key := strings.ToLower(clean)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, clean)
	}
	return out
}

func calcMarshalEtiquetas(values []string) (string, error) {
	normalized := calcNormalizeEtiquetas(values)
	if len(normalized) == 0 {
		return "[]", nil
	}
	raw, err := json.Marshal(normalized)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func calcUnmarshalEtiquetas(raw string) []string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return []string{}
	}
	items := make([]string, 0)
	if err := json.Unmarshal([]byte(value), &items); err == nil {
		return calcNormalizeEtiquetas(items)
	}
	return calcNormalizeEtiquetas(strings.Split(value, ","))
}

// EnsureEmpresaCalculadoraSchema crea/migra tablas del modulo calculadora empresarial.
func EnsureEmpresaCalculadoraSchema(dbConn *sql.DB) error {
	if SchemaBootstrapDisabled() {
		return nil
	}
	if dbConn == nil {
		return errors.New("db connection is nil")
	}

	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_calculadora_configuracion (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL UNIQUE,
			integrar_carritos INTEGER DEFAULT 1,
			integrar_cotizaciones INTEGER DEFAULT 1,
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_empresa_calculadora_configuracion_empresa ON empresa_calculadora_configuracion(empresa_id);`,
		`CREATE TABLE IF NOT EXISTS empresa_calculadora_operaciones (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			expresion TEXT NOT NULL,
			resultado TEXT NOT NULL,
			etiquetas_json TEXT DEFAULT '[]',
			cliente_id INTEGER DEFAULT 0,
			cliente_nombre TEXT,
			documento_tipo TEXT,
			documento_codigo TEXT,
			carrito_id INTEGER DEFAULT 0,
			cotizacion_id INTEGER DEFAULT 0,
			fecha_operacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			metadata_json TEXT DEFAULT '{}',
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_calculadora_operaciones_empresa_fecha ON empresa_calculadora_operaciones(empresa_id, fecha_operacion DESC, id DESC);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_calculadora_operaciones_usuario ON empresa_calculadora_operaciones(empresa_id, usuario_creador, fecha_operacion DESC);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_calculadora_operaciones_refs ON empresa_calculadora_operaciones(empresa_id, carrito_id, cotizacion_id);`,
	}

	for _, stmt := range stmts {
		if _, err := dbConn.Exec(stmt); err != nil {
			return err
		}
	}

	if err := ensureColumnIfMissing(dbConn, "empresa_calculadora_configuracion", "integrar_carritos", "INTEGER DEFAULT 1"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_calculadora_configuracion", "integrar_cotizaciones", "INTEGER DEFAULT 1"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_calculadora_configuracion", "fecha_actualizacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_calculadora_configuracion", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_calculadora_configuracion", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_calculadora_configuracion", "observaciones", "TEXT"); err != nil {
		return err
	}

	if err := ensureColumnIfMissing(dbConn, "empresa_calculadora_operaciones", "etiquetas_json", "TEXT DEFAULT '[]'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_calculadora_operaciones", "cliente_id", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_calculadora_operaciones", "cliente_nombre", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_calculadora_operaciones", "documento_tipo", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_calculadora_operaciones", "documento_codigo", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_calculadora_operaciones", "carrito_id", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_calculadora_operaciones", "cotizacion_id", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_calculadora_operaciones", "fecha_operacion", "TEXT DEFAULT (CURRENT_TIMESTAMP)"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_calculadora_operaciones", "metadata_json", "TEXT DEFAULT '{}'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_calculadora_operaciones", "fecha_actualizacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_calculadora_operaciones", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_calculadora_operaciones", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_calculadora_operaciones", "observaciones", "TEXT"); err != nil {
		return err
	}

	return nil
}

// GetEmpresaCalculadoraConfiguracion obtiene configuracion por empresa y crea default si no existe.
func GetEmpresaCalculadoraConfiguracion(dbConn *sql.DB, empresaID int64) (*EmpresaCalculadoraConfiguracion, error) {
	if dbConn == nil {
		return nil, errors.New("db connection is nil")
	}
	if empresaID <= 0 {
		return nil, errors.New("empresa_id invalido")
	}

	row := dbConn.QueryRow(`SELECT
		id,
		empresa_id,
		COALESCE(integrar_carritos, 1),
		COALESCE(integrar_cotizaciones, 1),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM empresa_calculadora_configuracion
	WHERE empresa_id = ?
	LIMIT 1`, empresaID)

	var out EmpresaCalculadoraConfiguracion
	var integrarCarritos int
	var integrarCotizaciones int
	err := row.Scan(
		&out.ID,
		&out.EmpresaID,
		&integrarCarritos,
		&integrarCotizaciones,
		&out.FechaCreacion,
		&out.FechaActualizacion,
		&out.UsuarioCreador,
		&out.Estado,
		&out.Observaciones,
	)
	if err == sql.ErrNoRows {
		defaultCfg := defaultEmpresaCalculadoraConfiguracion(empresaID)
		id, upsertErr := UpsertEmpresaCalculadoraConfiguracion(dbConn, defaultCfg)
		if upsertErr != nil {
			return nil, upsertErr
		}
		defaultCfg.ID = id
		return &defaultCfg, nil
	}
	if err != nil {
		return nil, err
	}

	out.IntegrarCarritos = integrarCarritos > 0
	out.IntegrarCotizaciones = integrarCotizaciones > 0
	out.Estado = normalizeEmpresaCalculadoraEstado(out.Estado)

	return &out, nil
}

// UpsertEmpresaCalculadoraConfiguracion crea o actualiza configuracion por empresa.
func UpsertEmpresaCalculadoraConfiguracion(dbConn *sql.DB, cfg EmpresaCalculadoraConfiguracion) (int64, error) {
	if dbConn == nil {
		return 0, errors.New("db connection is nil")
	}
	if cfg.EmpresaID <= 0 {
		return 0, errors.New("empresa_id invalido")
	}

	cfg.Estado = normalizeEmpresaCalculadoraEstado(cfg.Estado)
	var existingID int64
	err := dbConn.QueryRow("SELECT id FROM empresa_calculadora_configuracion WHERE empresa_id = ? LIMIT 1", cfg.EmpresaID).Scan(&existingID)
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}

	if err == sql.ErrNoRows {
		res, insertErr := dbConn.Exec(`INSERT INTO empresa_calculadora_configuracion (
			empresa_id,
			integrar_carritos,
			integrar_cotizaciones,
			fecha_creacion,
			fecha_actualizacion,
			usuario_creador,
			estado,
			observaciones
		) VALUES (
			?, ?, ?,
			CURRENT_TIMESTAMP,
			CURRENT_TIMESTAMP,
			?, ?, ?
		)`,
			cfg.EmpresaID,
			calcBoolToInt(cfg.IntegrarCarritos),
			calcBoolToInt(cfg.IntegrarCotizaciones),
			strings.TrimSpace(cfg.UsuarioCreador),
			cfg.Estado,
			strings.TrimSpace(cfg.Observaciones),
		)
		if insertErr != nil {
			return 0, insertErr
		}
		return res.LastInsertId()
	}

	_, updateErr := dbConn.Exec(`UPDATE empresa_calculadora_configuracion SET
		integrar_carritos = ?,
		integrar_cotizaciones = ?,
		fecha_actualizacion = CURRENT_TIMESTAMP,
		usuario_creador = ?,
		estado = ?,
		observaciones = ?
	WHERE empresa_id = ?`,
		calcBoolToInt(cfg.IntegrarCarritos),
		calcBoolToInt(cfg.IntegrarCotizaciones),
		strings.TrimSpace(cfg.UsuarioCreador),
		cfg.Estado,
		strings.TrimSpace(cfg.Observaciones),
		cfg.EmpresaID,
	)
	if updateErr != nil {
		return 0, updateErr
	}

	return existingID, nil
}

// CreateEmpresaCalculadoraOperacion registra una operacion en historial empresarial.
func CreateEmpresaCalculadoraOperacion(dbConn *sql.DB, op EmpresaCalculadoraOperacion) (int64, error) {
	if dbConn == nil {
		return 0, errors.New("db connection is nil")
	}
	if op.EmpresaID <= 0 {
		return 0, errors.New("empresa_id invalido")
	}
	op.Expresion = strings.TrimSpace(op.Expresion)
	op.Resultado = strings.TrimSpace(op.Resultado)
	if op.Expresion == "" {
		return 0, errors.New("expresion es obligatoria")
	}
	if op.Resultado == "" {
		op.Resultado = "0"
	}

	etiquetasJSON, err := calcMarshalEtiquetas(op.Etiquetas)
	if err != nil {
		return 0, fmt.Errorf("no se pudo serializar etiquetas: %w", err)
	}

	metadataJSON := strings.TrimSpace(op.MetadataJSON)
	if metadataJSON == "" {
		metadataJSON = "{}"
	} else if !json.Valid([]byte(metadataJSON)) {
		return 0, errors.New("metadata_json invalido")
	}

	op.Estado = normalizeEmpresaCalculadoraEstado(op.Estado)
	if op.ClienteID < 0 {
		op.ClienteID = 0
	}
	if op.CarritoID < 0 {
		op.CarritoID = 0
	}
	if op.CotizacionID < 0 {
		op.CotizacionID = 0
	}

	res, err := dbConn.Exec(`INSERT INTO empresa_calculadora_operaciones (
		empresa_id,
		expresion,
		resultado,
		etiquetas_json,
		cliente_id,
		cliente_nombre,
		documento_tipo,
		documento_codigo,
		carrito_id,
		cotizacion_id,
		fecha_operacion,
		metadata_json,
		fecha_creacion,
		fecha_actualizacion,
		usuario_creador,
		estado,
		observaciones
	) VALUES (
		?, ?, ?, ?,
		?, ?, ?, ?,
		?, ?,
		COALESCE(NULLIF(?, ''), CURRENT_TIMESTAMP),
		?,
		CURRENT_TIMESTAMP,
		CURRENT_TIMESTAMP,
		?, ?, ?
	)`,
		op.EmpresaID,
		op.Expresion,
		op.Resultado,
		etiquetasJSON,
		op.ClienteID,
		strings.TrimSpace(op.ClienteNombre),
		strings.TrimSpace(op.DocumentoTipo),
		strings.TrimSpace(op.DocumentoCodigo),
		op.CarritoID,
		op.CotizacionID,
		strings.TrimSpace(op.FechaOperacion),
		metadataJSON,
		strings.TrimSpace(op.UsuarioCreador),
		op.Estado,
		strings.TrimSpace(op.Observaciones),
	)
	if err != nil {
		return 0, err
	}

	return res.LastInsertId()
}

func calcBuildOperacionesFilterClause(filter EmpresaCalculadoraOperacionFilter, includeInactive bool) (string, []interface{}) {
	clauses := make([]string, 0, 6)
	args := make([]interface{}, 0, 6)

	if !includeInactive {
		clauses = append(clauses, "LOWER(COALESCE(estado, 'activo')) = 'activo'")
	}

	if strings.TrimSpace(filter.Desde) != "" {
		clauses = append(clauses, "date(COALESCE(fecha_operacion, fecha_creacion)) >= date(?)")
		args = append(args, strings.TrimSpace(filter.Desde))
	}
	if strings.TrimSpace(filter.Hasta) != "" {
		clauses = append(clauses, "date(COALESCE(fecha_operacion, fecha_creacion)) <= date(?)")
		args = append(args, strings.TrimSpace(filter.Hasta))
	}
	if strings.TrimSpace(filter.UsuarioCreador) != "" {
		clauses = append(clauses, "LOWER(COALESCE(usuario_creador, '')) = LOWER(?)")
		args = append(args, strings.TrimSpace(filter.UsuarioCreador))
	}
	if filter.ClienteID > 0 {
		clauses = append(clauses, "COALESCE(cliente_id, 0) = ?")
		args = append(args, filter.ClienteID)
	}
	if strings.TrimSpace(filter.Etiqueta) != "" {
		clauses = append(clauses, "LOWER(COALESCE(etiquetas_json, '')) LIKE LOWER(?) ESCAPE '!'")
		args = append(args, calcLikePattern(filter.Etiqueta))
	}

	if len(clauses) == 0 {
		return "", args
	}
	return " AND " + strings.Join(clauses, " AND "), args
}

// ListEmpresaCalculadoraOperaciones lista historial filtrable por empresa.
func ListEmpresaCalculadoraOperaciones(dbConn *sql.DB, empresaID int64, filter EmpresaCalculadoraOperacionFilter) ([]EmpresaCalculadoraOperacion, int64, error) {
	if dbConn == nil {
		return nil, 0, errors.New("db connection is nil")
	}
	if empresaID <= 0 {
		return nil, 0, errors.New("empresa_id invalido")
	}

	whereSQL, whereArgs := calcBuildOperacionesFilterClause(filter, filter.IncludeInactive)

	countQuery := "SELECT COUNT(1) FROM empresa_calculadora_operaciones WHERE empresa_id = ?" + whereSQL
	countArgs := make([]interface{}, 0, 1+len(whereArgs))
	countArgs = append(countArgs, empresaID)
	countArgs = append(countArgs, whereArgs...)

	var total int64
	if err := dbConn.QueryRow(countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}

	limit, offset := calcNormalizeLimitOffset(filter.Limit, filter.Offset)
	// #nosec G202 -- SQL structure is assembled only from server-side allowlists; all external values remain bound parameters.
	query := `SELECT
		id,
		empresa_id,
		COALESCE(expresion, ''),
		COALESCE(resultado, ''),
		COALESCE(etiquetas_json, '[]'),
		COALESCE(cliente_id, 0),
		COALESCE(cliente_nombre, ''),
		COALESCE(documento_tipo, ''),
		COALESCE(documento_codigo, ''),
		COALESCE(carrito_id, 0),
		COALESCE(cotizacion_id, 0),
		COALESCE(fecha_operacion, ''),
		COALESCE(metadata_json, '{}'),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM empresa_calculadora_operaciones
	WHERE empresa_id = ?` + whereSQL + `
	ORDER BY pcs_ts(COALESCE(fecha_operacion, fecha_creacion)) DESC, id DESC
	LIMIT ? OFFSET ?`

	args := make([]interface{}, 0, 3+len(whereArgs))
	args = append(args, empresaID)
	args = append(args, whereArgs...)
	args = append(args, limit, offset)

	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := make([]EmpresaCalculadoraOperacion, 0)
	for rows.Next() {
		var item EmpresaCalculadoraOperacion
		var etiquetasJSON string
		if err := rows.Scan(
			&item.ID,
			&item.EmpresaID,
			&item.Expresion,
			&item.Resultado,
			&etiquetasJSON,
			&item.ClienteID,
			&item.ClienteNombre,
			&item.DocumentoTipo,
			&item.DocumentoCodigo,
			&item.CarritoID,
			&item.CotizacionID,
			&item.FechaOperacion,
			&item.MetadataJSON,
			&item.FechaCreacion,
			&item.FechaActualizacion,
			&item.UsuarioCreador,
			&item.Estado,
			&item.Observaciones,
		); err != nil {
			return nil, 0, err
		}
		item.Etiquetas = calcUnmarshalEtiquetas(etiquetasJSON)
		item.Estado = normalizeEmpresaCalculadoraEstado(item.Estado)
		out = append(out, item)
	}

	return out, total, nil
}

// GetEmpresaCalculadoraOperacionByID obtiene un registro puntual por empresa.
func GetEmpresaCalculadoraOperacionByID(dbConn *sql.DB, empresaID, operacionID int64) (*EmpresaCalculadoraOperacion, error) {
	if dbConn == nil {
		return nil, errors.New("db connection is nil")
	}
	if empresaID <= 0 || operacionID <= 0 {
		return nil, errors.New("empresa_id u operacion_id invalido")
	}

	row := dbConn.QueryRow(`SELECT
		id,
		empresa_id,
		COALESCE(expresion, ''),
		COALESCE(resultado, ''),
		COALESCE(etiquetas_json, '[]'),
		COALESCE(cliente_id, 0),
		COALESCE(cliente_nombre, ''),
		COALESCE(documento_tipo, ''),
		COALESCE(documento_codigo, ''),
		COALESCE(carrito_id, 0),
		COALESCE(cotizacion_id, 0),
		COALESCE(fecha_operacion, ''),
		COALESCE(metadata_json, '{}'),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM empresa_calculadora_operaciones
	WHERE empresa_id = ? AND id = ?
	LIMIT 1`, empresaID, operacionID)

	var item EmpresaCalculadoraOperacion
	var etiquetasJSON string
	if err := row.Scan(
		&item.ID,
		&item.EmpresaID,
		&item.Expresion,
		&item.Resultado,
		&etiquetasJSON,
		&item.ClienteID,
		&item.ClienteNombre,
		&item.DocumentoTipo,
		&item.DocumentoCodigo,
		&item.CarritoID,
		&item.CotizacionID,
		&item.FechaOperacion,
		&item.MetadataJSON,
		&item.FechaCreacion,
		&item.FechaActualizacion,
		&item.UsuarioCreador,
		&item.Estado,
		&item.Observaciones,
	); err != nil {
		return nil, err
	}
	item.Etiquetas = calcUnmarshalEtiquetas(etiquetasJSON)
	item.Estado = normalizeEmpresaCalculadoraEstado(item.Estado)

	return &item, nil
}

// SetEmpresaCalculadoraOperacionEstadoByID ajusta estado de un registro puntual.
func SetEmpresaCalculadoraOperacionEstadoByID(dbConn *sql.DB, empresaID, operacionID int64, estado string) error {
	if dbConn == nil {
		return errors.New("db connection is nil")
	}
	if empresaID <= 0 || operacionID <= 0 {
		return errors.New("empresa_id u operacion_id invalido")
	}

	_, err := dbConn.Exec(`UPDATE empresa_calculadora_operaciones SET
		estado = ?,
		fecha_actualizacion = CURRENT_TIMESTAMP
	WHERE empresa_id = ? AND id = ?`, normalizeEmpresaCalculadoraEstado(estado), empresaID, operacionID)
	return err
}

// SetEmpresaCalculadoraOperacionesEstado aplica estado por filtro (limpieza logica del historial).
func SetEmpresaCalculadoraOperacionesEstado(dbConn *sql.DB, empresaID int64, filter EmpresaCalculadoraOperacionFilter, estado string) (int64, error) {
	if dbConn == nil {
		return 0, errors.New("db connection is nil")
	}
	if empresaID <= 0 {
		return 0, errors.New("empresa_id invalido")
	}

	whereSQL, whereArgs := calcBuildOperacionesFilterClause(filter, false)
	// #nosec G202 -- SQL structure is assembled only from server-side allowlists; all external values remain bound parameters.
	query := "UPDATE empresa_calculadora_operaciones SET estado = ?, fecha_actualizacion = CURRENT_TIMESTAMP WHERE empresa_id = ?" + whereSQL
	args := make([]interface{}, 0, 2+len(whereArgs))
	args = append(args, normalizeEmpresaCalculadoraEstado(estado), empresaID)
	args = append(args, whereArgs...)

	res, err := dbConn.Exec(query, args...)
	if err != nil {
		return 0, err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	return rows, nil
}

func calcBoolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
