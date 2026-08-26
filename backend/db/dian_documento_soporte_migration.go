package db

import (
	"context"
	"database/sql"
	"fmt"
)

const empresaDIANDocumentoSoporteFingerprint = "empresa-dian-documento-soporte:v1:numeric-lines-seller-numbering-source-snapshot"

// applyEmpresaDIANDocumentoSoporteTx upgrades the legacy accounting draft into
// a tenant-scoped fiscal source. Existing drafts are preserved as drafts: the
// migration never invents a legal number, CUDS, seller address, or DIAN result.
func applyEmpresaDIANDocumentoSoporteTx(ctx context.Context, tx *sql.Tx) error {
	if tx == nil {
		return fmt.Errorf("migration transaction is required")
	}

	var incompatible int64
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*)
		FROM empresa_contabilidad_documentos_soporte
		WHERE COALESCE(subtotal, 0) < 0
		   OR COALESCE(iva, 0) < 0
		   OR COALESCE(retenciones, 0) < 0
		   OR COALESCE(total, 0) < 0
		   OR COALESCE(retenciones, 0) > COALESCE(total, 0)
		   OR ABS(COALESCE(subtotal, 0)::numeric - ROUND(COALESCE(subtotal, 0)::numeric, 2)) > 0.005
		   OR ABS(COALESCE(iva, 0)::numeric - ROUND(COALESCE(iva, 0)::numeric, 2)) > 0.005
		   OR ABS(COALESCE(retenciones, 0)::numeric - ROUND(COALESCE(retenciones, 0)::numeric, 2)) > 0.005
		   OR ABS(COALESCE(total, 0)::numeric - ROUND(COALESCE(total, 0)::numeric, 2)) > 0.005`).Scan(&incompatible); err != nil {
		return fmt.Errorf("inspect documento soporte monetary precision: %w", err)
	}
	if incompatible > 0 {
		return fmt.Errorf("empresa_contabilidad_documentos_soporte contains %d monetary rows requiring manual reconciliation", incompatible)
	}

	for _, statement := range empresaDIANDocumentoSoporteMigrationStatements() {
		if _, err := tx.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("migrate DIAN documento soporte schema: %w", err)
		}
	}
	return nil
}

func empresaDIANDocumentoSoporteMigrationStatements() []string {
	return []string{
		`UPDATE empresa_contabilidad_documentos_soporte
			SET subtotal = COALESCE(subtotal, 0), iva = COALESCE(iva, 0),
				retenciones = COALESCE(retenciones, 0), total = COALESCE(total, 0)`,
		`ALTER TABLE empresa_contabilidad_documentos_soporte
			ALTER COLUMN subtotal TYPE NUMERIC(18,2) USING ROUND(COALESCE(subtotal, 0)::numeric, 2),
			ALTER COLUMN iva TYPE NUMERIC(18,2) USING ROUND(COALESCE(iva, 0)::numeric, 2),
			ALTER COLUMN retenciones TYPE NUMERIC(18,2) USING ROUND(COALESCE(retenciones, 0)::numeric, 2),
			ALTER COLUMN total TYPE NUMERIC(18,2) USING ROUND(COALESCE(total, 0)::numeric, 2)`,
		`ALTER TABLE empresa_contabilidad_documentos_soporte
			ALTER COLUMN subtotal SET DEFAULT 0,
			ALTER COLUMN subtotal SET NOT NULL,
			ALTER COLUMN iva SET DEFAULT 0,
			ALTER COLUMN iva SET NOT NULL,
			ALTER COLUMN retenciones SET DEFAULT 0,
			ALTER COLUMN retenciones SET NOT NULL,
			ALTER COLUMN total SET DEFAULT 0,
			ALTER COLUMN total SET NOT NULL`,
		`ALTER TABLE empresa_contabilidad_documentos_soporte
			ADD COLUMN IF NOT EXISTS numero_legal TEXT NOT NULL DEFAULT '',
			ADD COLUMN IF NOT EXISTS fecha_emision_legal TIMESTAMPTZ,
			ADD COLUMN IF NOT EXISTS vendedor_residencia TEXT NOT NULL DEFAULT 'residente',
			ADD COLUMN IF NOT EXISTS vendedor_digito_verificacion TEXT NOT NULL DEFAULT '',
			ADD COLUMN IF NOT EXISTS vendedor_tipo_persona TEXT NOT NULL DEFAULT '',
			ADD COLUMN IF NOT EXISTS vendedor_direccion TEXT NOT NULL DEFAULT '',
			ADD COLUMN IF NOT EXISTS vendedor_pais_codigo TEXT NOT NULL DEFAULT 'CO',
			ADD COLUMN IF NOT EXISTS vendedor_departamento TEXT NOT NULL DEFAULT '',
			ADD COLUMN IF NOT EXISTS vendedor_departamento_codigo_dane TEXT NOT NULL DEFAULT '',
			ADD COLUMN IF NOT EXISTS vendedor_municipio TEXT NOT NULL DEFAULT '',
			ADD COLUMN IF NOT EXISTS vendedor_municipio_codigo_dane TEXT NOT NULL DEFAULT '',
			ADD COLUMN IF NOT EXISTS vendedor_codigo_postal TEXT NOT NULL DEFAULT '',
			ADD COLUMN IF NOT EXISTS vendedor_responsabilidad_tributaria TEXT NOT NULL DEFAULT '',
			ADD COLUMN IF NOT EXISTS vendedor_email TEXT NOT NULL DEFAULT '',
			ADD COLUMN IF NOT EXISTS vendedor_telefono TEXT NOT NULL DEFAULT '',
			ADD COLUMN IF NOT EXISTS moneda TEXT NOT NULL DEFAULT 'COP',
			ADD COLUMN IF NOT EXISTS forma_pago_codigo TEXT NOT NULL DEFAULT '1',
			ADD COLUMN IF NOT EXISTS medio_pago_codigo TEXT NOT NULL DEFAULT '10',
			ADD COLUMN IF NOT EXISTS fecha_vencimiento DATE,
			ADD COLUMN IF NOT EXISTS lineas_json JSONB NOT NULL DEFAULT '[]'::jsonb,
			ADD COLUMN IF NOT EXISTS total_neto_contable NUMERIC(18,2) NOT NULL DEFAULT 0,
			ADD COLUMN IF NOT EXISTS configuracion_dian_json JSONB NOT NULL DEFAULT '{}'::jsonb,
			ADD COLUMN IF NOT EXISTS fuente_fiscal_sellada BOOLEAN NOT NULL DEFAULT FALSE`,
		`UPDATE empresa_contabilidad_documentos_soporte
			SET total_neto_contable = ROUND(GREATEST(total - retenciones, 0)::numeric, 2)
			WHERE total_neto_contable = 0 AND (total > 0 OR retenciones > 0)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS uq_empresa_documento_soporte_numero_legal
			ON empresa_contabilidad_documentos_soporte (empresa_id, numero_legal)
			WHERE numero_legal <> ''`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_documento_soporte_estado
			ON empresa_contabilidad_documentos_soporte (empresa_id, estado_dian, id)`,
		`DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM pg_constraint
				WHERE conname = 'ck_empresa_documento_soporte_importes'
				  AND conrelid = 'empresa_contabilidad_documentos_soporte'::regclass
			) THEN
				ALTER TABLE empresa_contabilidad_documentos_soporte
					ADD CONSTRAINT ck_empresa_documento_soporte_importes
					CHECK (subtotal >= 0 AND iva >= 0 AND retenciones >= 0 AND total >= 0 AND total_neto_contable >= 0);
			END IF;
		END $$`,
		`DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM pg_constraint
				WHERE conname = 'ck_empresa_documento_soporte_fuente_totales'
				  AND conrelid = 'empresa_contabilidad_documentos_soporte'::regclass
			) THEN
				ALTER TABLE empresa_contabilidad_documentos_soporte
					ADD CONSTRAINT ck_empresa_documento_soporte_fuente_totales
					CHECK (
						jsonb_typeof(lineas_json) = 'array'
						AND (lineas_json = '[]'::jsonb OR (
							total = subtotal + iva
							AND total_neto_contable = total - retenciones
						))
					);
			END IF;
		END $$`,
	}
}
