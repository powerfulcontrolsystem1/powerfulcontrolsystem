package db

import (
	"database/sql"
	"fmt"
	"strings"
)

// EmpresaDocumentoFacturacion representa un documento transaccional de facturacion.
type EmpresaDocumentoFacturacion struct {
	ID                   int64   `json:"id"`
	EmpresaID            int64   `json:"empresa_id"`
	TipoDocumento        string  `json:"tipo_documento"`
	DocumentoCodigo      string  `json:"documento_codigo"`
	NumeroLegal          string  `json:"numero_legal"`
	CodigoValidacion     string  `json:"codigo_validacion"`
	PaisCodigo           string  `json:"pais_codigo"`
	AmbienteFE           string  `json:"ambiente_fe"`
	EstadoDocumento      string  `json:"estado_documento"`
	EstadoAnterior       string  `json:"estado_anterior"`
	EventoUltimo         string  `json:"evento_ultimo"`
	PeriodoContable      string  `json:"periodo_contable"`
	MontoTotal           float64 `json:"monto_total"`
	Moneda               string  `json:"moneda"`
	FechaDocumento       string  `json:"fecha_documento"`
	EntidadRelacionadaID int64   `json:"entidad_relacionada_id"`
	FechaCreacion        string  `json:"fecha_creacion"`
	FechaActualizacion   string  `json:"fecha_actualizacion"`
	UsuarioCreador       string  `json:"usuario_creador"`
	Estado               string  `json:"estado"`
	Observaciones        string  `json:"observaciones"`
}

// EmpresaDocumentoCompra representa un documento transaccional de compras.
type EmpresaDocumentoCompra struct {
	ID                   int64   `json:"id"`
	EmpresaID            int64   `json:"empresa_id"`
	ProveedorID          int64   `json:"proveedor_id"`
	TipoDocumento        string  `json:"tipo_documento"`
	DocumentoCodigo      string  `json:"documento_codigo"`
	EstadoDocumento      string  `json:"estado_documento"`
	EstadoAnterior       string  `json:"estado_anterior"`
	EventoUltimo         string  `json:"evento_ultimo"`
	PeriodoContable      string  `json:"periodo_contable"`
	MontoTotal           float64 `json:"monto_total"`
	Moneda               string  `json:"moneda"`
	FechaDocumento       string  `json:"fecha_documento"`
	EntidadRelacionadaID int64   `json:"entidad_relacionada_id"`
	FechaCreacion        string  `json:"fecha_creacion"`
	FechaActualizacion   string  `json:"fecha_actualizacion"`
	UsuarioCreador       string  `json:"usuario_creador"`
	Estado               string  `json:"estado"`
	Observaciones        string  `json:"observaciones"`
}

// EmpresaDocumentoFacturacionListado representa un documento de facturación enriquecido con datos de cliente.
type EmpresaDocumentoFacturacionListado struct {
	EmpresaDocumentoFacturacion
	ClienteNombre    string `json:"cliente_nombre,omitempty"`
	ClienteEmail     string `json:"cliente_email,omitempty"`
	ClienteDocumento string `json:"cliente_documento,omitempty"`
}

// EmpresaDocumentoFacturacionListFilter define los filtros para consultar documentos de facturación.
type EmpresaDocumentoFacturacionListFilter struct {
	EmpresaID       int64
	TipoDocumento   string
	EstadoDocumento string
	IncludeInactive bool
	ClienteQuery    string
	DocumentoQuery  string
	FechaDesde      string
	FechaHasta      string
	Query           string
	Limit           int
	Offset          int
}

// EnsureEmpresaDocumentosTransaccionalesSchema crea/migra tablas de documentos de negocio para facturacion y compras.
func EnsureEmpresaDocumentosTransaccionalesSchema(dbConn *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_facturacion_documentos (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			tipo_documento TEXT NOT NULL,
			documento_codigo TEXT NOT NULL,
			estado_documento TEXT NOT NULL,
			estado_anterior TEXT,
			evento_ultimo TEXT,
			periodo_contable TEXT,
			monto_total REAL DEFAULT 0,
			moneda TEXT DEFAULT 'COP',
			numero_legal TEXT,
			codigo_validacion TEXT,
			pais_codigo TEXT,
			ambiente_fe TEXT,
			fecha_documento TEXT,
			entidad_relacionada_id INTEGER,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			UNIQUE(empresa_id, tipo_documento, documento_codigo)
		);`,
		`CREATE TABLE IF NOT EXISTS empresa_compras_documentos (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			proveedor_id INTEGER,
			tipo_documento TEXT NOT NULL,
			documento_codigo TEXT NOT NULL,
			estado_documento TEXT NOT NULL,
			estado_anterior TEXT,
			evento_ultimo TEXT,
			periodo_contable TEXT,
			monto_total REAL DEFAULT 0,
			moneda TEXT DEFAULT 'COP',
			fecha_documento TEXT,
			entidad_relacionada_id INTEGER,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			UNIQUE(empresa_id, tipo_documento, documento_codigo)
		);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_facturacion_documentos_lookup ON empresa_facturacion_documentos(empresa_id, tipo_documento, documento_codigo);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_facturacion_documentos_estado ON empresa_facturacion_documentos(empresa_id, estado_documento, estado);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_compras_documentos_lookup ON empresa_compras_documentos(empresa_id, tipo_documento, documento_codigo);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_compras_documentos_estado ON empresa_compras_documentos(empresa_id, estado_documento, estado);`,
	}
	for _, stmt := range stmts {
		if _, err := dbConn.Exec(stmt); err != nil {
			return err
		}
	}

	if err := ensureColumnIfMissing(dbConn, "empresa_facturacion_documentos", "empresa_id", "INTEGER NOT NULL"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_facturacion_documentos", "tipo_documento", "TEXT NOT NULL"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_facturacion_documentos", "documento_codigo", "TEXT NOT NULL"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_facturacion_documentos", "estado_documento", "TEXT NOT NULL"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_facturacion_documentos", "estado_anterior", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_facturacion_documentos", "evento_ultimo", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_facturacion_documentos", "periodo_contable", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_facturacion_documentos", "monto_total", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_facturacion_documentos", "moneda", "TEXT DEFAULT 'COP'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_facturacion_documentos", "numero_legal", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_facturacion_documentos", "codigo_validacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_facturacion_documentos", "pais_codigo", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_facturacion_documentos", "ambiente_fe", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_facturacion_documentos", "fecha_documento", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_facturacion_documentos", "entidad_relacionada_id", "INTEGER"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_facturacion_documentos", "fecha_actualizacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_facturacion_documentos", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_facturacion_documentos", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_facturacion_documentos", "observaciones", "TEXT"); err != nil {
		return err
	}

	if err := ensureColumnIfMissing(dbConn, "empresa_compras_documentos", "empresa_id", "INTEGER NOT NULL"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_compras_documentos", "proveedor_id", "INTEGER"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_compras_documentos", "tipo_documento", "TEXT NOT NULL"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_compras_documentos", "documento_codigo", "TEXT NOT NULL"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_compras_documentos", "estado_documento", "TEXT NOT NULL"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_compras_documentos", "estado_anterior", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_compras_documentos", "evento_ultimo", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_compras_documentos", "periodo_contable", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_compras_documentos", "monto_total", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_compras_documentos", "moneda", "TEXT DEFAULT 'COP'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_compras_documentos", "fecha_documento", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_compras_documentos", "entidad_relacionada_id", "INTEGER"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_compras_documentos", "fecha_actualizacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_compras_documentos", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_compras_documentos", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_compras_documentos", "observaciones", "TEXT"); err != nil {
		return err
	}

	return nil
}

// GetEmpresaDocumentoFacturacionByCodigo obtiene un documento de facturacion por llave de negocio.
func GetEmpresaDocumentoFacturacionByCodigo(dbConn *sql.DB, empresaID int64, tipoDocumento, documentoCodigo string) (*EmpresaDocumentoFacturacion, error) {
	if empresaID <= 0 {
		return nil, fmt.Errorf("empresa_id es obligatorio")
	}
	tipo := normalizeDocumentoTransaccionalTipo(tipoDocumento, "factura_electronica")
	codigo := normalizeDocumentoTransaccionalCodigo(documentoCodigo)
	if codigo == "" {
		return nil, fmt.Errorf("documento_codigo es obligatorio")
	}

	var item EmpresaDocumentoFacturacion
	err := dbConn.QueryRow(`SELECT
		id,
		empresa_id,
		COALESCE(tipo_documento, 'factura_electronica'),
		COALESCE(documento_codigo, ''),
		COALESCE(numero_legal, ''),
		COALESCE(codigo_validacion, ''),
		COALESCE(pais_codigo, ''),
		COALESCE(ambiente_fe, ''),
		COALESCE(estado_documento, 'borrador'),
		COALESCE(estado_anterior, ''),
		COALESCE(evento_ultimo, ''),
		COALESCE(periodo_contable, ''),
		COALESCE(monto_total, 0),
		COALESCE(moneda, 'COP'),
		COALESCE(fecha_documento, ''),
		COALESCE(entidad_relacionada_id, 0),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM empresa_facturacion_documentos
	WHERE empresa_id = ? AND tipo_documento = ? AND documento_codigo = ?
	LIMIT 1`, empresaID, tipo, codigo).Scan(
		&item.ID,
		&item.EmpresaID,
		&item.TipoDocumento,
		&item.DocumentoCodigo,
		&item.NumeroLegal,
		&item.CodigoValidacion,
		&item.PaisCodigo,
		&item.AmbienteFE,
		&item.EstadoDocumento,
		&item.EstadoAnterior,
		&item.EventoUltimo,
		&item.PeriodoContable,
		&item.MontoTotal,
		&item.Moneda,
		&item.FechaDocumento,
		&item.EntidadRelacionadaID,
		&item.FechaCreacion,
		&item.FechaActualizacion,
		&item.UsuarioCreador,
		&item.Estado,
		&item.Observaciones,
	)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// UpsertEmpresaDocumentoFacturacion registra o actualiza estado transaccional de facturacion.
func UpsertEmpresaDocumentoFacturacion(dbConn *sql.DB, payload EmpresaDocumentoFacturacion) (*EmpresaDocumentoFacturacion, error) {
	if payload.EmpresaID <= 0 {
		return nil, fmt.Errorf("empresa_id es obligatorio")
	}
	payload.TipoDocumento = normalizeDocumentoTransaccionalTipo(payload.TipoDocumento, "factura_electronica")
	payload.DocumentoCodigo = normalizeDocumentoTransaccionalCodigo(payload.DocumentoCodigo)
	if payload.DocumentoCodigo == "" {
		return nil, fmt.Errorf("documento_codigo es obligatorio")
	}
	payload.EstadoDocumento = normalizeDocumentoTransaccionalEstado(payload.EstadoDocumento, "borrador")
	payload.EstadoAnterior = normalizeDocumentoTransaccionalEstado(payload.EstadoAnterior, "")
	payload.EventoUltimo = strings.TrimSpace(strings.ToLower(payload.EventoUltimo))
	payload.PeriodoContable = normalizePeriodoEventoContable(payload.PeriodoContable)
	if payload.PeriodoContable == "" {
		payload.PeriodoContable = normalizePeriodoEventoContable(payload.FechaDocumento)
	}
	payload.Moneda = normalizeDocumentoTransaccionalMoneda(payload.Moneda)
	payload.NumeroLegal = strings.ToUpper(strings.TrimSpace(payload.NumeroLegal))
	payload.CodigoValidacion = strings.ToUpper(strings.TrimSpace(payload.CodigoValidacion))
	payload.PaisCodigo = strings.ToUpper(strings.TrimSpace(payload.PaisCodigo))
	payload.AmbienteFE = strings.ToLower(strings.TrimSpace(payload.AmbienteFE))
	payload.FechaDocumento = strings.TrimSpace(payload.FechaDocumento)
	payload.UsuarioCreador = strings.TrimSpace(payload.UsuarioCreador)
	payload.Estado = strings.TrimSpace(strings.ToLower(payload.Estado))
	if payload.Estado == "" {
		payload.Estado = "activo"
	}
	payload.Observaciones = strings.TrimSpace(payload.Observaciones)
	if payload.MontoTotal < 0 {
		payload.MontoTotal = 0
	}

	_, err := dbConn.Exec(`INSERT INTO empresa_facturacion_documentos (
		empresa_id,
		tipo_documento,
		documento_codigo,
		estado_documento,
		estado_anterior,
		evento_ultimo,
		periodo_contable,
		monto_total,
		moneda,
		numero_legal,
		codigo_validacion,
		pais_codigo,
		ambiente_fe,
		fecha_documento,
		entidad_relacionada_id,
		fecha_creacion,
		fecha_actualizacion,
		usuario_creador,
		estado,
		observaciones
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now','localtime'), datetime('now','localtime'), ?, ?, ?)
	ON CONFLICT(empresa_id, tipo_documento, documento_codigo) DO UPDATE SET
		estado_documento = excluded.estado_documento,
		estado_anterior = excluded.estado_anterior,
		evento_ultimo = excluded.evento_ultimo,
		periodo_contable = CASE WHEN excluded.periodo_contable <> '' THEN excluded.periodo_contable ELSE empresa_facturacion_documentos.periodo_contable END,
		monto_total = CASE WHEN excluded.monto_total > 0 THEN excluded.monto_total ELSE empresa_facturacion_documentos.monto_total END,
		moneda = CASE WHEN excluded.moneda <> '' THEN excluded.moneda ELSE empresa_facturacion_documentos.moneda END,
		numero_legal = CASE WHEN excluded.numero_legal <> '' THEN excluded.numero_legal ELSE empresa_facturacion_documentos.numero_legal END,
		codigo_validacion = CASE WHEN excluded.codigo_validacion <> '' THEN excluded.codigo_validacion ELSE empresa_facturacion_documentos.codigo_validacion END,
		pais_codigo = CASE WHEN excluded.pais_codigo <> '' THEN excluded.pais_codigo ELSE empresa_facturacion_documentos.pais_codigo END,
		ambiente_fe = CASE WHEN excluded.ambiente_fe <> '' THEN excluded.ambiente_fe ELSE empresa_facturacion_documentos.ambiente_fe END,
		fecha_documento = CASE WHEN excluded.fecha_documento <> '' THEN excluded.fecha_documento ELSE empresa_facturacion_documentos.fecha_documento END,
		entidad_relacionada_id = CASE WHEN excluded.entidad_relacionada_id > 0 THEN excluded.entidad_relacionada_id ELSE empresa_facturacion_documentos.entidad_relacionada_id END,
		fecha_actualizacion = datetime('now','localtime'),
		usuario_creador = CASE WHEN excluded.usuario_creador <> '' THEN excluded.usuario_creador ELSE empresa_facturacion_documentos.usuario_creador END,
		estado = excluded.estado,
		observaciones = CASE WHEN excluded.observaciones <> '' THEN excluded.observaciones ELSE empresa_facturacion_documentos.observaciones END`,
		payload.EmpresaID,
		payload.TipoDocumento,
		payload.DocumentoCodigo,
		payload.EstadoDocumento,
		payload.EstadoAnterior,
		payload.EventoUltimo,
		payload.PeriodoContable,
		payload.MontoTotal,
		payload.Moneda,
		payload.NumeroLegal,
		payload.CodigoValidacion,
		payload.PaisCodigo,
		payload.AmbienteFE,
		payload.FechaDocumento,
		payload.EntidadRelacionadaID,
		payload.UsuarioCreador,
		payload.Estado,
		payload.Observaciones,
	)
	if err != nil {
		return nil, err
	}

	return GetEmpresaDocumentoFacturacionByCodigo(dbConn, payload.EmpresaID, payload.TipoDocumento, payload.DocumentoCodigo)
}

// ListEmpresaDocumentosFacturacionByEmpresa lista documentos de facturación por empresa con filtros operativos.
func ListEmpresaDocumentosFacturacionByEmpresa(dbConn *sql.DB, filter EmpresaDocumentoFacturacionListFilter) ([]EmpresaDocumentoFacturacionListado, error) {
	if filter.EmpresaID <= 0 {
		return nil, fmt.Errorf("empresa_id es obligatorio")
	}

	if filter.Limit <= 0 || filter.Limit > 500 {
		filter.Limit = 100
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	tipo := normalizeDocumentoTransaccionalTipo(filter.TipoDocumento, "")
	estadoDoc := normalizeDocumentoTransaccionalEstado(filter.EstadoDocumento, "")
	clienteQuery := strings.TrimSpace(filter.ClienteQuery)
	documentoQuery := strings.TrimSpace(filter.DocumentoQuery)
	fechaDesde := strings.TrimSpace(filter.FechaDesde)
	fechaHasta := strings.TrimSpace(filter.FechaHasta)
	busqueda := strings.TrimSpace(filter.Query)

	query := `SELECT
		d.id,
		d.empresa_id,
		COALESCE(d.tipo_documento, 'factura_electronica'),
		COALESCE(d.documento_codigo, ''),
		COALESCE(d.numero_legal, ''),
		COALESCE(d.codigo_validacion, ''),
		COALESCE(d.pais_codigo, ''),
		COALESCE(d.ambiente_fe, ''),
		COALESCE(d.estado_documento, 'borrador'),
		COALESCE(d.estado_anterior, ''),
		COALESCE(d.evento_ultimo, ''),
		COALESCE(d.periodo_contable, ''),
		COALESCE(d.monto_total, 0),
		COALESCE(d.moneda, 'COP'),
		COALESCE(d.fecha_documento, ''),
		COALESCE(d.entidad_relacionada_id, 0),
		COALESCE(d.fecha_creacion, ''),
		COALESCE(d.fecha_actualizacion, ''),
		COALESCE(d.usuario_creador, ''),
		COALESCE(d.estado, 'activo'),
		COALESCE(d.observaciones, ''),
		COALESCE(c.nombre_razon_social, ''),
		COALESCE(c.email, ''),
		COALESCE(c.numero_documento, '')
	FROM empresa_facturacion_documentos d
	LEFT JOIN clientes c ON c.empresa_id = d.empresa_id AND c.id = d.entidad_relacionada_id
	WHERE d.empresa_id = ?`
	args := []interface{}{filter.EmpresaID}

	if tipo != "" {
		query += ` AND d.tipo_documento = ?`
		args = append(args, tipo)
	}
	if estadoDoc != "" {
		query += ` AND d.estado_documento = ?`
		args = append(args, estadoDoc)
	}
	if !filter.IncludeInactive {
		query += ` AND COALESCE(d.estado, 'activo') = 'activo'`
	}
	if clienteQuery != "" {
		query += ` AND (
			lower(COALESCE(c.nombre_razon_social, '')) LIKE lower(?)
			OR lower(COALESCE(c.email, '')) LIKE lower(?)
			OR lower(COALESCE(c.numero_documento, '')) LIKE lower(?)
		)`
		clienteLike := "%" + clienteQuery + "%"
		args = append(args, clienteLike, clienteLike, clienteLike)
	}
	if documentoQuery != "" {
		query += ` AND (
			upper(COALESCE(d.documento_codigo, '')) LIKE ?
			OR upper(COALESCE(d.numero_legal, '')) LIKE ?
			OR upper(COALESCE(d.codigo_validacion, '')) LIKE ?
		)`
		documentoLike := "%" + strings.ToUpper(documentoQuery) + "%"
		args = append(args, documentoLike, documentoLike, documentoLike)
	}
	if fechaDesde != "" {
		query += ` AND date(COALESCE(NULLIF(d.fecha_documento, ''), substr(d.fecha_creacion, 1, 10))) >= date(?)`
		args = append(args, fechaDesde)
	}
	if fechaHasta != "" {
		query += ` AND date(COALESCE(NULLIF(d.fecha_documento, ''), substr(d.fecha_creacion, 1, 10))) <= date(?)`
		args = append(args, fechaHasta)
	}
	if busqueda != "" {
		query += ` AND (
			upper(COALESCE(d.documento_codigo, '')) LIKE ?
			OR upper(COALESCE(d.numero_legal, '')) LIKE ?
			OR upper(COALESCE(d.codigo_validacion, '')) LIKE ?
			OR lower(COALESCE(c.nombre_razon_social, '')) LIKE lower(?)
			OR lower(COALESCE(c.email, '')) LIKE lower(?)
			OR lower(COALESCE(c.numero_documento, '')) LIKE lower(?)
			OR lower(COALESCE(d.observaciones, '')) LIKE lower(?)
		)`
		qUpper := "%" + strings.ToUpper(busqueda) + "%"
		qLower := "%" + busqueda + "%"
		args = append(args, qUpper, qUpper, qUpper, qLower, qLower, qLower, qLower)
	}

	query += `
	ORDER BY datetime(COALESCE(NULLIF(d.fecha_documento, ''), d.fecha_creacion)) DESC, d.id DESC
	LIMIT ? OFFSET ?`
	args = append(args, filter.Limit, filter.Offset)

	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]EmpresaDocumentoFacturacionListado, 0)
	for rows.Next() {
		var item EmpresaDocumentoFacturacionListado
		if err := rows.Scan(
			&item.ID,
			&item.EmpresaID,
			&item.TipoDocumento,
			&item.DocumentoCodigo,
			&item.NumeroLegal,
			&item.CodigoValidacion,
			&item.PaisCodigo,
			&item.AmbienteFE,
			&item.EstadoDocumento,
			&item.EstadoAnterior,
			&item.EventoUltimo,
			&item.PeriodoContable,
			&item.MontoTotal,
			&item.Moneda,
			&item.FechaDocumento,
			&item.EntidadRelacionadaID,
			&item.FechaCreacion,
			&item.FechaActualizacion,
			&item.UsuarioCreador,
			&item.Estado,
			&item.Observaciones,
			&item.ClienteNombre,
			&item.ClienteEmail,
			&item.ClienteDocumento,
		); err != nil {
			return nil, err
		}
		out = append(out, item)
	}

	return out, rows.Err()
}

// GetEmpresaDocumentoCompraByCodigo obtiene un documento de compras por llave de negocio.
func GetEmpresaDocumentoCompraByCodigo(dbConn *sql.DB, empresaID int64, tipoDocumento, documentoCodigo string) (*EmpresaDocumentoCompra, error) {
	if empresaID <= 0 {
		return nil, fmt.Errorf("empresa_id es obligatorio")
	}
	tipo := normalizeDocumentoTransaccionalTipo(tipoDocumento, "orden_compra")
	codigo := normalizeDocumentoTransaccionalCodigo(documentoCodigo)
	if codigo == "" {
		return nil, fmt.Errorf("documento_codigo es obligatorio")
	}

	var item EmpresaDocumentoCompra
	err := dbConn.QueryRow(`SELECT
		id,
		empresa_id,
		COALESCE(proveedor_id, 0),
		COALESCE(tipo_documento, 'orden_compra'),
		COALESCE(documento_codigo, ''),
		COALESCE(estado_documento, 'borrador'),
		COALESCE(estado_anterior, ''),
		COALESCE(evento_ultimo, ''),
		COALESCE(periodo_contable, ''),
		COALESCE(monto_total, 0),
		COALESCE(moneda, 'COP'),
		COALESCE(fecha_documento, ''),
		COALESCE(entidad_relacionada_id, 0),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM empresa_compras_documentos
	WHERE empresa_id = ? AND tipo_documento = ? AND documento_codigo = ?
	LIMIT 1`, empresaID, tipo, codigo).Scan(
		&item.ID,
		&item.EmpresaID,
		&item.ProveedorID,
		&item.TipoDocumento,
		&item.DocumentoCodigo,
		&item.EstadoDocumento,
		&item.EstadoAnterior,
		&item.EventoUltimo,
		&item.PeriodoContable,
		&item.MontoTotal,
		&item.Moneda,
		&item.FechaDocumento,
		&item.EntidadRelacionadaID,
		&item.FechaCreacion,
		&item.FechaActualizacion,
		&item.UsuarioCreador,
		&item.Estado,
		&item.Observaciones,
	)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// UpsertEmpresaDocumentoCompra registra o actualiza estado transaccional de compras.
func UpsertEmpresaDocumentoCompra(dbConn *sql.DB, payload EmpresaDocumentoCompra) (*EmpresaDocumentoCompra, error) {
	if payload.EmpresaID <= 0 {
		return nil, fmt.Errorf("empresa_id es obligatorio")
	}
	payload.TipoDocumento = normalizeDocumentoTransaccionalTipo(payload.TipoDocumento, "orden_compra")
	payload.DocumentoCodigo = normalizeDocumentoTransaccionalCodigo(payload.DocumentoCodigo)
	if payload.DocumentoCodigo == "" {
		return nil, fmt.Errorf("documento_codigo es obligatorio")
	}
	payload.EstadoDocumento = normalizeDocumentoTransaccionalEstado(payload.EstadoDocumento, "borrador")
	payload.EstadoAnterior = normalizeDocumentoTransaccionalEstado(payload.EstadoAnterior, "")
	payload.EventoUltimo = strings.TrimSpace(strings.ToLower(payload.EventoUltimo))
	payload.PeriodoContable = normalizePeriodoEventoContable(payload.PeriodoContable)
	if payload.PeriodoContable == "" {
		payload.PeriodoContable = normalizePeriodoEventoContable(payload.FechaDocumento)
	}
	payload.Moneda = normalizeDocumentoTransaccionalMoneda(payload.Moneda)
	payload.FechaDocumento = strings.TrimSpace(payload.FechaDocumento)
	payload.UsuarioCreador = strings.TrimSpace(payload.UsuarioCreador)
	payload.Estado = strings.TrimSpace(strings.ToLower(payload.Estado))
	if payload.Estado == "" {
		payload.Estado = "activo"
	}
	payload.Observaciones = strings.TrimSpace(payload.Observaciones)
	if payload.MontoTotal < 0 {
		payload.MontoTotal = 0
	}

	_, err := dbConn.Exec(`INSERT INTO empresa_compras_documentos (
		empresa_id,
		proveedor_id,
		tipo_documento,
		documento_codigo,
		estado_documento,
		estado_anterior,
		evento_ultimo,
		periodo_contable,
		monto_total,
		moneda,
		fecha_documento,
		entidad_relacionada_id,
		fecha_creacion,
		fecha_actualizacion,
		usuario_creador,
		estado,
		observaciones
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now','localtime'), datetime('now','localtime'), ?, ?, ?)
	ON CONFLICT(empresa_id, tipo_documento, documento_codigo) DO UPDATE SET
		proveedor_id = CASE WHEN excluded.proveedor_id > 0 THEN excluded.proveedor_id ELSE empresa_compras_documentos.proveedor_id END,
		estado_documento = excluded.estado_documento,
		estado_anterior = excluded.estado_anterior,
		evento_ultimo = excluded.evento_ultimo,
		periodo_contable = CASE WHEN excluded.periodo_contable <> '' THEN excluded.periodo_contable ELSE empresa_compras_documentos.periodo_contable END,
		monto_total = CASE WHEN excluded.monto_total > 0 THEN excluded.monto_total ELSE empresa_compras_documentos.monto_total END,
		moneda = CASE WHEN excluded.moneda <> '' THEN excluded.moneda ELSE empresa_compras_documentos.moneda END,
		fecha_documento = CASE WHEN excluded.fecha_documento <> '' THEN excluded.fecha_documento ELSE empresa_compras_documentos.fecha_documento END,
		entidad_relacionada_id = CASE WHEN excluded.entidad_relacionada_id > 0 THEN excluded.entidad_relacionada_id ELSE empresa_compras_documentos.entidad_relacionada_id END,
		fecha_actualizacion = datetime('now','localtime'),
		usuario_creador = CASE WHEN excluded.usuario_creador <> '' THEN excluded.usuario_creador ELSE empresa_compras_documentos.usuario_creador END,
		estado = excluded.estado,
		observaciones = CASE WHEN excluded.observaciones <> '' THEN excluded.observaciones ELSE empresa_compras_documentos.observaciones END`,
		payload.EmpresaID,
		payload.ProveedorID,
		payload.TipoDocumento,
		payload.DocumentoCodigo,
		payload.EstadoDocumento,
		payload.EstadoAnterior,
		payload.EventoUltimo,
		payload.PeriodoContable,
		payload.MontoTotal,
		payload.Moneda,
		payload.FechaDocumento,
		payload.EntidadRelacionadaID,
		payload.UsuarioCreador,
		payload.Estado,
		payload.Observaciones,
	)
	if err != nil {
		return nil, err
	}

	return GetEmpresaDocumentoCompraByCodigo(dbConn, payload.EmpresaID, payload.TipoDocumento, payload.DocumentoCodigo)
}

// ListEmpresaDocumentosCompraByEmpresa lista documentos de compras por filtros operativos.
func ListEmpresaDocumentosCompraByEmpresa(dbConn *sql.DB, empresaID int64, tipoDocumento string, proveedorID int64, estadoDocumento string, includeInactive bool, q string, limit int, offset int) ([]EmpresaDocumentoCompra, error) {
	if empresaID <= 0 {
		return nil, fmt.Errorf("empresa_id es obligatorio")
	}

	if limit <= 0 || limit > 500 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	tipo := normalizeDocumentoTransaccionalTipo(tipoDocumento, "")
	estadoDoc := normalizeDocumentoTransaccionalEstado(estadoDocumento, "")
	busqueda := strings.ToUpper(strings.TrimSpace(q))

	query := `SELECT
		id,
		empresa_id,
		COALESCE(proveedor_id, 0),
		COALESCE(tipo_documento, 'orden_compra'),
		COALESCE(documento_codigo, ''),
		COALESCE(estado_documento, 'borrador'),
		COALESCE(estado_anterior, ''),
		COALESCE(evento_ultimo, ''),
		COALESCE(periodo_contable, ''),
		COALESCE(monto_total, 0),
		COALESCE(moneda, 'COP'),
		COALESCE(fecha_documento, ''),
		COALESCE(entidad_relacionada_id, 0),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM empresa_compras_documentos
	WHERE empresa_id = ?`
	args := []interface{}{empresaID}

	if tipo != "" {
		query += ` AND tipo_documento = ?`
		args = append(args, tipo)
	}
	if proveedorID > 0 {
		query += ` AND proveedor_id = ?`
		args = append(args, proveedorID)
	}
	if estadoDoc != "" {
		query += ` AND estado_documento = ?`
		args = append(args, estadoDoc)
	}
	if !includeInactive {
		query += ` AND COALESCE(estado, 'activo') = 'activo'`
	}
	if busqueda != "" {
		query += ` AND (documento_codigo LIKE ? OR observaciones LIKE ?)`
		like := "%" + busqueda + "%"
		args = append(args, like, like)
	}

	query += `
	ORDER BY datetime(COALESCE(NULLIF(fecha_actualizacion, ''), fecha_creacion)) DESC, id DESC
	LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]EmpresaDocumentoCompra, 0)
	for rows.Next() {
		var item EmpresaDocumentoCompra
		if err := rows.Scan(
			&item.ID,
			&item.EmpresaID,
			&item.ProveedorID,
			&item.TipoDocumento,
			&item.DocumentoCodigo,
			&item.EstadoDocumento,
			&item.EstadoAnterior,
			&item.EventoUltimo,
			&item.PeriodoContable,
			&item.MontoTotal,
			&item.Moneda,
			&item.FechaDocumento,
			&item.EntidadRelacionadaID,
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

// SetEmpresaDocumentoCompraEstadoByCodigo actualiza estado activo/inactivo del documento de compras.
func SetEmpresaDocumentoCompraEstadoByCodigo(dbConn *sql.DB, empresaID int64, tipoDocumento, documentoCodigo, estado string) error {
	if empresaID <= 0 {
		return fmt.Errorf("empresa_id es obligatorio")
	}
	tipo := normalizeDocumentoTransaccionalTipo(tipoDocumento, "orden_compra")
	codigo := normalizeDocumentoTransaccionalCodigo(documentoCodigo)
	if codigo == "" {
		return fmt.Errorf("documento_codigo es obligatorio")
	}
	estadoNorm := strings.ToLower(strings.TrimSpace(estado))
	if estadoNorm == "" {
		estadoNorm = "activo"
	}

	res, err := dbConn.Exec(`UPDATE empresa_compras_documentos
		SET estado = ?, fecha_actualizacion = datetime('now','localtime')
		WHERE empresa_id = ? AND tipo_documento = ? AND documento_codigo = ?`, estadoNorm, empresaID, tipo, codigo)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected <= 0 {
		return sql.ErrNoRows
	}
	return nil
}

func normalizeDocumentoTransaccionalTipo(v, fallback string) string {
	v = strings.TrimSpace(strings.ToLower(v))
	if v == "" {
		return strings.TrimSpace(strings.ToLower(fallback))
	}
	return v
}

func normalizeDocumentoTransaccionalCodigo(v string) string {
	return strings.ToUpper(strings.TrimSpace(v))
}

func normalizeDocumentoTransaccionalEstado(v, fallback string) string {
	v = strings.TrimSpace(strings.ToLower(v))
	v = strings.ReplaceAll(v, "-", "_")
	v = strings.ReplaceAll(v, " ", "_")
	if v == "" {
		v = strings.TrimSpace(strings.ToLower(fallback))
		v = strings.ReplaceAll(v, "-", "_")
		v = strings.ReplaceAll(v, " ", "_")
	}
	return v
}

func normalizeDocumentoTransaccionalMoneda(v string) string {
	v = strings.TrimSpace(strings.ToUpper(v))
	if v == "" {
		return "COP"
	}
	return v
}
