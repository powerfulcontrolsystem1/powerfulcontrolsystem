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
		fecha_documento,
		entidad_relacionada_id,
		fecha_creacion,
		fecha_actualizacion,
		usuario_creador,
		estado,
		observaciones
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now','localtime'), datetime('now','localtime'), ?, ?, ?)
	ON CONFLICT(empresa_id, tipo_documento, documento_codigo) DO UPDATE SET
		estado_documento = excluded.estado_documento,
		estado_anterior = excluded.estado_anterior,
		evento_ultimo = excluded.evento_ultimo,
		periodo_contable = CASE WHEN excluded.periodo_contable <> '' THEN excluded.periodo_contable ELSE empresa_facturacion_documentos.periodo_contable END,
		monto_total = CASE WHEN excluded.monto_total > 0 THEN excluded.monto_total ELSE empresa_facturacion_documentos.monto_total END,
		moneda = CASE WHEN excluded.moneda <> '' THEN excluded.moneda ELSE empresa_facturacion_documentos.moneda END,
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
