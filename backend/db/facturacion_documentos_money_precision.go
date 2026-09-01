package db

import (
	"context"
	"database/sql"
	"fmt"
)

const empresaFacturacionDocumentosMoneyPrecisionFingerprint = "empresa-facturacion-documentos:v1:numeric-18-2:fail-closed"

// applyEmpresaFacturacionDocumentosMoneyPrecisionTx replaces the legacy REAL
// fiscal amount with NUMERIC(18,2). Amounts that would materially change at
// two decimals are deliberately left for reconciliation instead of being
// silently modified during a production migration.
func applyEmpresaFacturacionDocumentosMoneyPrecisionTx(ctx context.Context, tx *sql.Tx) error {
	if tx == nil {
		return fmt.Errorf("migration transaction is required")
	}

	var incompatible int64
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*)
		FROM empresa_facturacion_documentos
		WHERE COALESCE(monto_total, 0) < 0
		   OR ABS(COALESCE(monto_total, 0)::numeric - ROUND(COALESCE(monto_total, 0)::numeric, 2)) > 0.005`).Scan(&incompatible); err != nil {
		return fmt.Errorf("inspect empresa_facturacion_documentos money precision: %w", err)
	}
	if incompatible > 0 {
		return fmt.Errorf("empresa_facturacion_documentos contains %d monetary rows requiring manual reconciliation", incompatible)
	}

	for _, statement := range []string{
		`UPDATE empresa_facturacion_documentos SET monto_total = 0 WHERE monto_total IS NULL`,
		`ALTER TABLE empresa_facturacion_documentos
			ALTER COLUMN monto_total TYPE NUMERIC(18,2) USING ROUND(COALESCE(monto_total, 0)::numeric, 2)`,
		`ALTER TABLE empresa_facturacion_documentos
			ALTER COLUMN monto_total SET DEFAULT 0,
			ALTER COLUMN monto_total SET NOT NULL`,
		`ALTER TABLE empresa_facturacion_documentos
			ADD CONSTRAINT ck_empresa_facturacion_documentos_monto_total_nonnegative CHECK (monto_total >= 0)`,
	} {
		if _, err := tx.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("migrate empresa_facturacion_documentos money precision: %w", err)
		}
	}
	return nil
}
