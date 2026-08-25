package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

const empresaFacturacionArtefactosFingerprint = "empresa-facturacion-artefactos:v1:private-tenant-files:sha256"
const empresaFacturacionFuenteFiscalFingerprint = "empresa-facturacion-artefactos:v2:immutable-fuente-fiscal-json"

const EmpresaFacturacionArtefactoTipoFuenteFiscalJSON = "fuente_fiscal_json"

var ErrEmpresaFacturacionFuenteFiscalInmutable = errors.New("la fuente fiscal ya existe con contenido diferente")

type EmpresaFacturacionArtefacto struct {
	ID                 int64  `json:"id"`
	EmpresaID          int64  `json:"empresa_id"`
	TipoDocumento      string `json:"tipo_documento"`
	DocumentoCodigo    string `json:"documento_codigo"`
	TipoArtefacto      string `json:"tipo_artefacto"`
	StorageRef         string `json:"-"`
	SHA256             string `json:"sha256"`
	MimeType           string `json:"mime_type"`
	TamanoBytes        int64  `json:"tamano_bytes"`
	Estado             string `json:"estado"`
	FechaCreacion      string `json:"fecha_creacion"`
	FechaActualizacion string `json:"fecha_actualizacion"`
}

func applyEmpresaFacturacionArtefactosTx(ctx context.Context, tx *sql.Tx) error {
	if tx == nil {
		return fmt.Errorf("migration transaction is required")
	}
	for _, statement := range []string{
		`CREATE TABLE empresa_facturacion_artefactos (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			tipo_documento TEXT NOT NULL,
			documento_codigo TEXT NOT NULL,
			tipo_artefacto TEXT NOT NULL,
			storage_ref TEXT NOT NULL,
			sha256 CHAR(64) NOT NULL,
			mime_type TEXT NOT NULL,
			tamano_bytes BIGINT NOT NULL,
			estado TEXT NOT NULL DEFAULT 'activo',
			fecha_creacion TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			fecha_actualizacion TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			CONSTRAINT uq_empresa_facturacion_artefacto UNIQUE (empresa_id, tipo_documento, documento_codigo, tipo_artefacto),
			CONSTRAINT ck_empresa_facturacion_artefacto_tipo CHECK (tipo_artefacto IN ('xml_firmado','respuesta_proveedor','representacion_pdf')),
			CONSTRAINT ck_empresa_facturacion_artefacto_sha256 CHECK (sha256 ~ '^[0-9a-f]{64}$'),
			CONSTRAINT ck_empresa_facturacion_artefacto_tamano CHECK (tamano_bytes > 0)
		)`,
		`CREATE INDEX ix_empresa_facturacion_artefactos_documento ON empresa_facturacion_artefactos (empresa_id, tipo_documento, documento_codigo, estado)`,
	} {
		if _, err := tx.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("migrate empresa_facturacion_artefactos: %w", err)
		}
	}
	return nil
}

// applyEmpresaFacturacionFuenteFiscalTx extends the released artifact catalog
// without rewriting the original migration. The source JSON is internal,
// tenant-scoped and immutable once attached to a fiscal document.
func applyEmpresaFacturacionFuenteFiscalTx(ctx context.Context, tx *sql.Tx) error {
	if tx == nil {
		return fmt.Errorf("migration transaction is required")
	}
	if _, err := tx.ExecContext(ctx, `ALTER TABLE empresa_facturacion_artefactos
		DROP CONSTRAINT IF EXISTS ck_empresa_facturacion_artefacto_tipo`); err != nil {
		return fmt.Errorf("drop legacy fiscal artifact type constraint: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `ALTER TABLE empresa_facturacion_artefactos
		ADD CONSTRAINT ck_empresa_facturacion_artefacto_tipo
		CHECK (tipo_artefacto IN ('xml_firmado','respuesta_proveedor','representacion_pdf','fuente_fiscal_json'))`); err != nil {
		return fmt.Errorf("add fiscal source artifact type constraint: %w", err)
	}
	return nil
}

func normalizeEmpresaFacturacionArtefacto(item *EmpresaFacturacionArtefacto) error {
	if item == nil || item.EmpresaID <= 0 {
		return fmt.Errorf("empresa_id es obligatorio")
	}
	item.TipoDocumento = normalizeDocumentoTransaccionalTipo(item.TipoDocumento, "factura_electronica")
	item.DocumentoCodigo = normalizeDocumentoTransaccionalCodigo(item.DocumentoCodigo)
	item.TipoArtefacto = strings.ToLower(strings.TrimSpace(item.TipoArtefacto))
	item.StorageRef = strings.TrimSpace(item.StorageRef)
	item.SHA256 = strings.ToLower(strings.TrimSpace(item.SHA256))
	item.MimeType = strings.ToLower(strings.TrimSpace(item.MimeType))
	item.Estado = strings.ToLower(strings.TrimSpace(item.Estado))
	if item.Estado == "" {
		item.Estado = "activo"
	}
	if item.DocumentoCodigo == "" || item.StorageRef == "" || len(item.SHA256) != 64 || item.TamanoBytes <= 0 {
		return fmt.Errorf("artefacto fiscal incompleto")
	}
	switch item.TipoArtefacto {
	case "xml_firmado", "respuesta_proveedor", "representacion_pdf", EmpresaFacturacionArtefactoTipoFuenteFiscalJSON:
	default:
		return fmt.Errorf("tipo_artefacto invalido")
	}
	return nil
}

func UpsertEmpresaFacturacionArtefactoContext(ctx context.Context, dbConn *sql.DB, item EmpresaFacturacionArtefacto) (*EmpresaFacturacionArtefacto, error) {
	if dbConn == nil {
		return nil, fmt.Errorf("conexion empresarial no disponible")
	}
	if err := normalizeEmpresaFacturacionArtefacto(&item); err != nil {
		return nil, err
	}
	if item.TipoArtefacto == EmpresaFacturacionArtefactoTipoFuenteFiscalJSON {
		return nil, fmt.Errorf("fuente_fiscal_json solo admite insercion inmutable")
	}
	_, err := execSQLCompatContext(ctx, dbConn, `INSERT INTO empresa_facturacion_artefactos (
		empresa_id, tipo_documento, documento_codigo, tipo_artefacto, storage_ref, sha256, mime_type, tamano_bytes, estado
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT (empresa_id, tipo_documento, documento_codigo, tipo_artefacto) DO UPDATE SET
		storage_ref = EXCLUDED.storage_ref,
		sha256 = EXCLUDED.sha256,
		mime_type = EXCLUDED.mime_type,
		tamano_bytes = EXCLUDED.tamano_bytes,
		estado = EXCLUDED.estado,
		fecha_actualizacion = CURRENT_TIMESTAMP`, item.EmpresaID, item.TipoDocumento, item.DocumentoCodigo, item.TipoArtefacto, item.StorageRef, item.SHA256, item.MimeType, item.TamanoBytes, item.Estado)
	if err != nil {
		return nil, err
	}
	return GetEmpresaFacturacionArtefactoByTypeContext(ctx, dbConn, item.EmpresaID, item.TipoDocumento, item.DocumentoCodigo, item.TipoArtefacto)
}

type empresaFacturacionArtefactoSQLStore interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}

// InsertEmpresaFacturacionFuenteFiscalContext stores the canonical fiscal
// source exactly once. Repeating the same SHA-256 is idempotent; attempting to
// replace it with different content fails closed.
func InsertEmpresaFacturacionFuenteFiscalContext(ctx context.Context, dbConn *sql.DB, item EmpresaFacturacionArtefacto) (*EmpresaFacturacionArtefacto, error) {
	if dbConn == nil {
		return nil, fmt.Errorf("conexion empresarial no disponible")
	}
	return insertEmpresaFacturacionFuenteFiscalStore(ctx, dbConn, item)
}

func insertEmpresaFacturacionFuenteFiscalStore(ctx context.Context, store empresaFacturacionArtefactoSQLStore, item EmpresaFacturacionArtefacto) (*EmpresaFacturacionArtefacto, error) {
	if store == nil {
		return nil, fmt.Errorf("almacen empresarial no disponible")
	}
	item.TipoArtefacto = EmpresaFacturacionArtefactoTipoFuenteFiscalJSON
	if err := normalizeEmpresaFacturacionArtefacto(&item); err != nil {
		return nil, err
	}
	if item.MimeType != "application/json" {
		return nil, fmt.Errorf("fuente_fiscal_json requiere application/json")
	}
	if _, err := store.ExecContext(ctx, `INSERT INTO empresa_facturacion_artefactos (
		empresa_id, tipo_documento, documento_codigo, tipo_artefacto, storage_ref, sha256, mime_type, tamano_bytes, estado
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'activo')
	ON CONFLICT (empresa_id, tipo_documento, documento_codigo, tipo_artefacto) DO NOTHING`,
		item.EmpresaID, item.TipoDocumento, item.DocumentoCodigo, item.TipoArtefacto,
		item.StorageRef, item.SHA256, item.MimeType, item.TamanoBytes); err != nil {
		return nil, err
	}

	stored, err := scanEmpresaFacturacionArtefacto(store.QueryRowContext(ctx,
		empresaFacturacionArtefactoSelect+` WHERE empresa_id = $1 AND tipo_documento = $2 AND documento_codigo = $3 AND tipo_artefacto = $4 AND estado = 'activo' LIMIT 1`,
		item.EmpresaID, item.TipoDocumento, item.DocumentoCodigo, item.TipoArtefacto))
	if err != nil {
		return nil, err
	}
	if !strings.EqualFold(stored.SHA256, item.SHA256) || stored.MimeType != item.MimeType || stored.TamanoBytes != item.TamanoBytes {
		return nil, ErrEmpresaFacturacionFuenteFiscalInmutable
	}
	return stored, nil
}

func scanEmpresaFacturacionArtefacto(scanner interface{ Scan(...interface{}) error }) (*EmpresaFacturacionArtefacto, error) {
	var item EmpresaFacturacionArtefacto
	err := scanner.Scan(&item.ID, &item.EmpresaID, &item.TipoDocumento, &item.DocumentoCodigo, &item.TipoArtefacto, &item.StorageRef, &item.SHA256, &item.MimeType, &item.TamanoBytes, &item.Estado, &item.FechaCreacion, &item.FechaActualizacion)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

const empresaFacturacionArtefactoSelect = `SELECT id, empresa_id, tipo_documento, documento_codigo, tipo_artefacto, storage_ref, sha256, mime_type, tamano_bytes, estado, fecha_creacion, fecha_actualizacion FROM empresa_facturacion_artefactos`

func GetEmpresaFacturacionArtefactoByTypeContext(ctx context.Context, dbConn *sql.DB, empresaID int64, tipoDocumento, documentoCodigo, tipoArtefacto string) (*EmpresaFacturacionArtefacto, error) {
	return scanEmpresaFacturacionArtefacto(queryRowSQLCompatContext(ctx, dbConn, empresaFacturacionArtefactoSelect+` WHERE empresa_id = ? AND tipo_documento = ? AND documento_codigo = ? AND tipo_artefacto = ? AND estado = 'activo' LIMIT 1`, empresaID, normalizeDocumentoTransaccionalTipo(tipoDocumento, "factura_electronica"), normalizeDocumentoTransaccionalCodigo(documentoCodigo), strings.ToLower(strings.TrimSpace(tipoArtefacto))))
}

func GetEmpresaFacturacionArtefactoByIDContext(ctx context.Context, dbConn *sql.DB, empresaID, id int64) (*EmpresaFacturacionArtefacto, error) {
	if empresaID <= 0 || id <= 0 {
		return nil, fmt.Errorf("empresa_id e id son obligatorios")
	}
	return scanEmpresaFacturacionArtefacto(queryRowSQLCompatContext(ctx, dbConn, empresaFacturacionArtefactoSelect+` WHERE empresa_id = ? AND id = ? AND estado = 'activo' LIMIT 1`, empresaID, id))
}

func ListEmpresaFacturacionArtefactosContext(ctx context.Context, dbConn *sql.DB, empresaID int64, tipoDocumento, documentoCodigo string) ([]EmpresaFacturacionArtefacto, error) {
	if empresaID <= 0 || strings.TrimSpace(documentoCodigo) == "" {
		return nil, fmt.Errorf("empresa_id y documento_codigo son obligatorios")
	}
	rows, err := querySQLCompatContext(ctx, dbConn, empresaFacturacionArtefactoSelect+` WHERE empresa_id = ? AND tipo_documento = ? AND documento_codigo = ? AND estado = 'activo' ORDER BY tipo_artefacto`, empresaID, normalizeDocumentoTransaccionalTipo(tipoDocumento, "factura_electronica"), normalizeDocumentoTransaccionalCodigo(documentoCodigo))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]EmpresaFacturacionArtefacto, 0)
	for rows.Next() {
		item, err := scanEmpresaFacturacionArtefacto(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}
	return items, rows.Err()
}
