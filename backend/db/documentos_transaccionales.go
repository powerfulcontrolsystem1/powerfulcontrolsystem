package db

import (
	"database/sql"
	"errors"
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
	RequiereAprobacion   bool    `json:"requiere_aprobacion"`
	NivelesAprobacion    int     `json:"niveles_aprobacion_requeridos"`
	NivelAprobacion      int     `json:"nivel_aprobacion_actual"`
	AprobadoresJSON      string  `json:"aprobadores_json,omitempty"`
	RecepcionDetalleJSON string  `json:"recepcion_detalle_json,omitempty"`
	RecepcionResumenJSON string  `json:"recepcion_resumen_json,omitempty"`
	ValidacionEstado     string  `json:"validacion_documental_estado,omitempty"`
	ProveedorDocRef      string  `json:"proveedor_documento_ref,omitempty"`
	FacturaDocRef        string  `json:"factura_documento_ref,omitempty"`
	EntradaDocRef        string  `json:"entrada_documento_ref,omitempty"`
	ComprobanteURL       string  `json:"comprobante_url,omitempty"`
	ComprobanteNombre    string  `json:"comprobante_nombre_archivo,omitempty"`
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
	CajeroQuery     string
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
			id BIGSERIAL PRIMARY KEY,
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
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			UNIQUE(empresa_id, tipo_documento, documento_codigo)
		);`,
		`CREATE TABLE IF NOT EXISTS empresa_compras_documentos (
			id BIGSERIAL PRIMARY KEY,
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
			requiere_aprobacion INTEGER DEFAULT 0,
			niveles_aprobacion_requeridos INTEGER DEFAULT 1,
			nivel_aprobacion_actual INTEGER DEFAULT 0,
			aprobadores_json TEXT,
			recepcion_detalle_json TEXT,
			recepcion_resumen_json TEXT,
			validacion_documental_estado TEXT DEFAULT 'no_aplica',
			proveedor_documento_ref TEXT,
			factura_documento_ref TEXT,
			entrada_documento_ref TEXT,
			comprobante_url TEXT,
			comprobante_nombre_archivo TEXT,
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
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
	if err := ensureColumnIfMissing(dbConn, "empresa_compras_documentos", "requiere_aprobacion", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_compras_documentos", "niveles_aprobacion_requeridos", "INTEGER DEFAULT 1"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_compras_documentos", "nivel_aprobacion_actual", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_compras_documentos", "aprobadores_json", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_compras_documentos", "recepcion_detalle_json", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_compras_documentos", "recepcion_resumen_json", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_compras_documentos", "validacion_documental_estado", "TEXT DEFAULT 'no_aplica'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_compras_documentos", "proveedor_documento_ref", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_compras_documentos", "factura_documento_ref", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_compras_documentos", "entrada_documento_ref", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_compras_documentos", "comprobante_url", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_compras_documentos", "comprobante_nombre_archivo", "TEXT"); err != nil {
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

	indexStmts := []string{
		`CREATE INDEX IF NOT EXISTS ix_empresa_compras_documentos_aprobacion ON empresa_compras_documentos(empresa_id, estado_documento, requiere_aprobacion, nivel_aprobacion_actual, niveles_aprobacion_requeridos);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_compras_documentos_validacion ON empresa_compras_documentos(empresa_id, validacion_documental_estado, estado);`,
	}
	for _, stmt := range indexStmts {
		if _, err := dbConn.Exec(stmt); err != nil {
			return err
		}
	}

	if err := ensurePostgresDocumentTableIDSequence(dbConn, "empresa_facturacion_documentos"); err != nil {
		return err
	}
	if err := ensurePostgresDocumentTableIDSequence(dbConn, "empresa_compras_documentos"); err != nil {
		return err
	}

	return nil
}

// VerifyEmpresaDocumentosTransaccionalesSchema is safe for request paths. DDL
// belongs to the migration role, so handlers only confirm the required table
// exists before executing a business transition.
func VerifyEmpresaDocumentosTransaccionalesSchema(dbConn *sql.DB) error {
	if dbConn == nil {
		return errors.New("db connection is nil")
	}
	var exists int
	err := queryRowSQLCompat(dbConn, `SELECT 1 FROM information_schema.tables WHERE table_schema = current_schema() AND table_name = ? LIMIT 1`, "empresa_facturacion_documentos").Scan(&exists)
	if err == sql.ErrNoRows {
		return errors.New("esquema de documentos no migrado; ejecute pcs-migrate")
	}
	return err
}

func ensurePostgresDocumentTableIDSequence(dbConn *sql.DB, tableName string) error {
	if dbConn == nil || !isPostgresDialect() {
		return nil
	}

	var columnDefault string
	err := dbConn.QueryRow(`SELECT COALESCE(column_default, '')
		FROM information_schema.columns
		WHERE table_schema = current_schema() AND table_name = $1 AND column_name = 'id'`, tableName).Scan(&columnDefault)
	if err != nil {
		errLower := strings.ToLower(err.Error())
		if strings.Contains(errLower, "no such table") || (strings.Contains(errLower, "relation") && strings.Contains(errLower, "does not exist")) {
			return nil
		}
	}
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return err
	}
	if strings.Contains(strings.ToLower(columnDefault), "nextval(") {
		return nil
	}

	seqName := tableName + "_id_seq"
	if _, err := dbConn.Exec(fmt.Sprintf(`CREATE SEQUENCE IF NOT EXISTS %s`, seqName)); err != nil {
		return err
	}
	if _, err := dbConn.Exec(fmt.Sprintf(`ALTER SEQUENCE %s OWNED BY %s.id`, seqName, tableName)); err != nil {
		return err
	}
	if _, err := dbConn.Exec(fmt.Sprintf(`ALTER TABLE %s ALTER COLUMN id SET DEFAULT nextval('%s')`, tableName, seqName)); err != nil {
		return err
	}

	var maxID int64
	if err := dbConn.QueryRow(fmt.Sprintf(`SELECT COALESCE(MAX(id), 0) FROM %s`, tableName)).Scan(&maxID); err != nil {
		return err
	}
	if maxID > 0 {
		if _, err := dbConn.Exec(`SELECT setval($1, $2, true)`, seqName, maxID); err != nil {
			return err
		}
		return nil
	}
	if _, err := dbConn.Exec(`SELECT setval($1, 1, false)`, seqName); err != nil {
		return err
	}
	return nil
}

// GetEmpresaDocumentoFacturacionByCodigo obtiene un documento de facturacion por llave de negocio.
func GetEmpresaDocumentoFacturacionByCodigo(dbConn *sql.DB, empresaID int64, tipoDocumento, documentoCodigo string) (*EmpresaDocumentoFacturacion, error) {
	if err := EnsureEmpresaDocumentosTransaccionalesSchema(dbConn); err != nil {
		return nil, err
	}

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

// UpdateEmpresaDocumentoFacturacionCliente asocia un documento de facturacion a un cliente de la misma empresa.
func UpdateEmpresaDocumentoFacturacionCliente(dbConn *sql.DB, empresaID int64, tipoDocumento, documentoCodigo string, clienteID int64) (*EmpresaDocumentoFacturacion, error) {
	if empresaID <= 0 {
		return nil, fmt.Errorf("empresa_id es obligatorio")
	}
	if clienteID <= 0 {
		return nil, fmt.Errorf("cliente_id es obligatorio")
	}
	tipoDocumento = normalizeDocumentoTransaccionalTipo(tipoDocumento, "comprobante_pago")
	documentoCodigo = normalizeDocumentoTransaccionalCodigo(documentoCodigo)
	if documentoCodigo == "" {
		return nil, fmt.Errorf("documento_codigo es obligatorio")
	}
	res, err := dbConn.Exec(`UPDATE empresa_facturacion_documentos
		SET entidad_relacionada_id = ?, fecha_actualizacion = CURRENT_TIMESTAMP
		WHERE empresa_id = ? AND tipo_documento = ? AND documento_codigo = ?`,
		clienteID,
		empresaID,
		tipoDocumento,
		documentoCodigo,
	)
	if err != nil {
		return nil, err
	}
	if affected, affErr := res.RowsAffected(); affErr == nil && affected == 0 {
		return nil, sql.ErrNoRows
	}
	return GetEmpresaDocumentoFacturacionByCodigo(dbConn, empresaID, tipoDocumento, documentoCodigo)
}

// UpsertEmpresaDocumentoFacturacion registra o actualiza estado transaccional de facturacion.
func UpsertEmpresaDocumentoFacturacion(dbConn *sql.DB, payload EmpresaDocumentoFacturacion) (*EmpresaDocumentoFacturacion, error) {
	if err := EnsureEmpresaDocumentosTransaccionalesSchema(dbConn); err != nil {
		return nil, err
	}

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
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, ?, ?, ?)
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
		fecha_actualizacion = CURRENT_TIMESTAMP,
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
	cajeroQuery := strings.TrimSpace(filter.CajeroQuery)
	fechaDesde := strings.TrimSpace(filter.FechaDesde)
	fechaHasta := strings.TrimSpace(filter.FechaHasta)
	busqueda := strings.TrimSpace(filter.Query)
	fechaDocumentoExpr := `COALESCE(
		NULLIF(CASE WHEN COALESCE(d.fecha_documento, '') LIKE 'date(%' OR COALESCE(d.fecha_documento, '') LIKE 'pcs_ts(%' THEN '' ELSE COALESCE(d.fecha_documento, '') END, ''),
		NULLIF(CASE WHEN COALESCE(d.fecha_creacion, '') LIKE 'date(%' OR COALESCE(d.fecha_creacion, '') LIKE 'pcs_ts(%' THEN '' ELSE COALESCE(d.fecha_creacion, '') END, ''),
		''
	)`
	fechaDocumentoDiaExpr := `substr(` + fechaDocumentoExpr + `, 1, 10)`

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
	if cajeroQuery != "" {
		query += ` AND lower(COALESCE(d.usuario_creador, '')) LIKE lower(?)`
		args = append(args, "%"+cajeroQuery+"%")
	}
	if fechaDesde != "" {
		query += ` AND ` + fechaDocumentoDiaExpr + ` >= ?`
		args = append(args, fechaDesde)
	}
	if fechaHasta != "" {
		query += ` AND ` + fechaDocumentoDiaExpr + ` <= ?`
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
	ORDER BY ` + fechaDocumentoExpr + ` DESC, d.id DESC
	LIMIT ? OFFSET ?`
	args = append(args, filter.Limit, filter.Offset)

	rows, err := ExecQueryCompat(dbConn, query, args...)
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
	var requiereAprobacionRaw int64
	var nivelesAprobacionRaw int64
	var nivelAprobacionRaw int64
	err := queryRowSQLCompat(dbConn, `SELECT
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
		COALESCE(requiere_aprobacion, 0),
		COALESCE(niveles_aprobacion_requeridos, 1),
		COALESCE(nivel_aprobacion_actual, 0),
		COALESCE(aprobadores_json, ''),
		COALESCE(recepcion_detalle_json, ''),
		COALESCE(recepcion_resumen_json, ''),
		COALESCE(validacion_documental_estado, 'no_aplica'),
		COALESCE(proveedor_documento_ref, ''),
		COALESCE(factura_documento_ref, ''),
		COALESCE(entrada_documento_ref, ''),
		COALESCE(comprobante_url, ''),
		COALESCE(comprobante_nombre_archivo, ''),
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
		&requiereAprobacionRaw,
		&nivelesAprobacionRaw,
		&nivelAprobacionRaw,
		&item.AprobadoresJSON,
		&item.RecepcionDetalleJSON,
		&item.RecepcionResumenJSON,
		&item.ValidacionEstado,
		&item.ProveedorDocRef,
		&item.FacturaDocRef,
		&item.EntradaDocRef,
		&item.ComprobanteURL,
		&item.ComprobanteNombre,
		&item.FechaCreacion,
		&item.FechaActualizacion,
		&item.UsuarioCreador,
		&item.Estado,
		&item.Observaciones,
	)
	if err != nil {
		return nil, err
	}
	item.RequiereAprobacion = int64ToBool(requiereAprobacionRaw)
	item.NivelesAprobacion = normalizeComprasNivelesAprobacion(int(nivelesAprobacionRaw), 1)
	item.NivelAprobacion = normalizeComprasNivelActual(int(nivelAprobacionRaw), item.NivelesAprobacion)
	item.ValidacionEstado = normalizeComprasValidacionEstado(item.ValidacionEstado)
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
	payload.RequiereAprobacion = int64ToBool(boolToInt64(payload.RequiereAprobacion))
	payload.NivelesAprobacion = normalizeComprasNivelesAprobacion(payload.NivelesAprobacion, 1)
	payload.NivelAprobacion = normalizeComprasNivelActual(payload.NivelAprobacion, payload.NivelesAprobacion)
	payload.AprobadoresJSON = strings.TrimSpace(payload.AprobadoresJSON)
	payload.RecepcionDetalleJSON = strings.TrimSpace(payload.RecepcionDetalleJSON)
	payload.RecepcionResumenJSON = strings.TrimSpace(payload.RecepcionResumenJSON)
	payload.ValidacionEstado = normalizeComprasValidacionEstado(payload.ValidacionEstado)
	payload.ProveedorDocRef = normalizeDocumentoTransaccionalCodigo(payload.ProveedorDocRef)
	payload.FacturaDocRef = normalizeDocumentoTransaccionalCodigo(payload.FacturaDocRef)
	payload.EntradaDocRef = normalizeDocumentoTransaccionalCodigo(payload.EntradaDocRef)
	payload.ComprobanteURL = strings.TrimSpace(payload.ComprobanteURL)
	payload.ComprobanteNombre = strings.TrimSpace(payload.ComprobanteNombre)
	payload.Observaciones = strings.TrimSpace(payload.Observaciones)
	if payload.MontoTotal < 0 {
		payload.MontoTotal = 0
	}

	_, err := execSQLCompat(dbConn, `INSERT INTO empresa_compras_documentos (
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
		requiere_aprobacion,
		niveles_aprobacion_requeridos,
		nivel_aprobacion_actual,
		aprobadores_json,
		recepcion_detalle_json,
		recepcion_resumen_json,
		validacion_documental_estado,
		proveedor_documento_ref,
		factura_documento_ref,
		entrada_documento_ref,
		comprobante_url,
		comprobante_nombre_archivo,
		fecha_creacion,
		fecha_actualizacion,
		usuario_creador,
		estado,
		observaciones
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, ?, ?, ?)
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
		requiere_aprobacion = excluded.requiere_aprobacion,
		niveles_aprobacion_requeridos = CASE
			WHEN excluded.niveles_aprobacion_requeridos > 0 THEN excluded.niveles_aprobacion_requeridos
			ELSE empresa_compras_documentos.niveles_aprobacion_requeridos
		END,
		nivel_aprobacion_actual = CASE
			WHEN excluded.nivel_aprobacion_actual >= 0 THEN excluded.nivel_aprobacion_actual
			ELSE empresa_compras_documentos.nivel_aprobacion_actual
		END,
		aprobadores_json = CASE WHEN excluded.aprobadores_json <> '' THEN excluded.aprobadores_json ELSE empresa_compras_documentos.aprobadores_json END,
		recepcion_detalle_json = CASE WHEN excluded.recepcion_detalle_json <> '' THEN excluded.recepcion_detalle_json ELSE empresa_compras_documentos.recepcion_detalle_json END,
		recepcion_resumen_json = CASE WHEN excluded.recepcion_resumen_json <> '' THEN excluded.recepcion_resumen_json ELSE empresa_compras_documentos.recepcion_resumen_json END,
		validacion_documental_estado = CASE
			WHEN excluded.validacion_documental_estado <> '' THEN excluded.validacion_documental_estado
			ELSE empresa_compras_documentos.validacion_documental_estado
		END,
		proveedor_documento_ref = CASE WHEN excluded.proveedor_documento_ref <> '' THEN excluded.proveedor_documento_ref ELSE empresa_compras_documentos.proveedor_documento_ref END,
		factura_documento_ref = CASE WHEN excluded.factura_documento_ref <> '' THEN excluded.factura_documento_ref ELSE empresa_compras_documentos.factura_documento_ref END,
		entrada_documento_ref = CASE WHEN excluded.entrada_documento_ref <> '' THEN excluded.entrada_documento_ref ELSE empresa_compras_documentos.entrada_documento_ref END,
		comprobante_url = CASE WHEN excluded.comprobante_url <> '' THEN excluded.comprobante_url ELSE empresa_compras_documentos.comprobante_url END,
		comprobante_nombre_archivo = CASE WHEN excluded.comprobante_nombre_archivo <> '' THEN excluded.comprobante_nombre_archivo ELSE empresa_compras_documentos.comprobante_nombre_archivo END,
		fecha_actualizacion = CURRENT_TIMESTAMP,
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
		boolToInt64(payload.RequiereAprobacion),
		payload.NivelesAprobacion,
		payload.NivelAprobacion,
		payload.AprobadoresJSON,
		payload.RecepcionDetalleJSON,
		payload.RecepcionResumenJSON,
		payload.ValidacionEstado,
		payload.ProveedorDocRef,
		payload.FacturaDocRef,
		payload.EntradaDocRef,
		payload.ComprobanteURL,
		payload.ComprobanteNombre,
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
		COALESCE(requiere_aprobacion, 0),
		COALESCE(niveles_aprobacion_requeridos, 1),
		COALESCE(nivel_aprobacion_actual, 0),
		COALESCE(aprobadores_json, ''),
		COALESCE(recepcion_detalle_json, ''),
		COALESCE(recepcion_resumen_json, ''),
		COALESCE(validacion_documental_estado, 'no_aplica'),
		COALESCE(proveedor_documento_ref, ''),
		COALESCE(factura_documento_ref, ''),
		COALESCE(entrada_documento_ref, ''),
		COALESCE(comprobante_url, ''),
		COALESCE(comprobante_nombre_archivo, ''),
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
	ORDER BY pcs_ts(COALESCE(NULLIF(fecha_actualizacion, ''), fecha_creacion)) DESC, id DESC
	LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	rows, err := querySQLCompat(dbConn, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]EmpresaDocumentoCompra, 0)
	for rows.Next() {
		var item EmpresaDocumentoCompra
		var requiereAprobacionRaw int64
		var nivelesAprobacionRaw int64
		var nivelAprobacionRaw int64
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
			&requiereAprobacionRaw,
			&nivelesAprobacionRaw,
			&nivelAprobacionRaw,
			&item.AprobadoresJSON,
			&item.RecepcionDetalleJSON,
			&item.RecepcionResumenJSON,
			&item.ValidacionEstado,
			&item.ProveedorDocRef,
			&item.FacturaDocRef,
			&item.EntradaDocRef,
			&item.ComprobanteURL,
			&item.ComprobanteNombre,
			&item.FechaCreacion,
			&item.FechaActualizacion,
			&item.UsuarioCreador,
			&item.Estado,
			&item.Observaciones,
		); err != nil {
			return nil, err
		}
		item.RequiereAprobacion = int64ToBool(requiereAprobacionRaw)
		item.NivelesAprobacion = normalizeComprasNivelesAprobacion(int(nivelesAprobacionRaw), 1)
		item.NivelAprobacion = normalizeComprasNivelActual(int(nivelAprobacionRaw), item.NivelesAprobacion)
		item.ValidacionEstado = normalizeComprasValidacionEstado(item.ValidacionEstado)
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
		SET estado = ?, fecha_actualizacion = CURRENT_TIMESTAMP
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

// UpdateEmpresaDocumentoCompraComprobante actualiza la evidencia adjunta de un documento de compras.
func UpdateEmpresaDocumentoCompraComprobante(dbConn *sql.DB, empresaID int64, tipoDocumento, documentoCodigo, comprobanteURL, comprobanteNombre string) error {
	if err := EnsureEmpresaDocumentosTransaccionalesSchema(dbConn); err != nil {
		return err
	}
	if empresaID <= 0 {
		return fmt.Errorf("empresa_id es obligatorio")
	}
	tipo := normalizeDocumentoTransaccionalTipo(tipoDocumento, "orden_compra")
	codigo := normalizeDocumentoTransaccionalCodigo(documentoCodigo)
	if codigo == "" {
		return fmt.Errorf("documento_codigo es obligatorio")
	}
	comprobanteURL = strings.TrimSpace(comprobanteURL)
	comprobanteNombre = strings.TrimSpace(comprobanteNombre)
	if comprobanteURL == "" {
		return fmt.Errorf("comprobante_url es obligatorio")
	}

	res, err := dbConn.Exec(`UPDATE empresa_compras_documentos
		SET comprobante_url = ?,
			comprobante_nombre_archivo = ?,
			fecha_actualizacion = CURRENT_TIMESTAMP
		WHERE empresa_id = ? AND tipo_documento = ? AND documento_codigo = ?`,
		comprobanteURL,
		comprobanteNombre,
		empresaID,
		tipo,
		codigo,
	)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected <= 0 {
		return sql.ErrNoRows
	}
	return nil
}

func boolToInt64(v bool) int64 {
	if v {
		return 1
	}
	return 0
}

func int64ToBool(v int64) bool { return v > 0 }

func normalizeComprasNivelesAprobacion(v int, fallback int) int {
	if fallback <= 0 {
		fallback = 1
	}
	if v <= 0 {
		v = fallback
	}
	if v < 1 {
		v = 1
	}
	if v > 10 {
		v = 10
	}
	return v
}

func normalizeComprasNivelActual(v int, niveles int) int {
	niveles = normalizeComprasNivelesAprobacion(niveles, 1)
	if v < 0 {
		return 0
	}
	if v > niveles {
		return niveles
	}
	return v
}

func normalizeComprasValidacionEstado(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "pendiente":
		return "pendiente"
	case "validada":
		return "validada"
	case "inconsistente":
		return "inconsistente"
	default:
		return "no_aplica"
	}
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
