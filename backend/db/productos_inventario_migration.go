package db

import (
	"context"
	"database/sql"
	"fmt"
)

const empresaProductosInventarioSchemaFingerprint = "empresa-productos-inventario:v1:recepcion-bodega-not-null:productos-tenant-state-index"

// applyEmpresaProductosInventarioSchemaTx owns the schema additions required
// by the atomic purchase-receipt flow. Runtime API and worker processes must
// never create these structures.
func applyEmpresaProductosInventarioSchemaTx(ctx context.Context, tx *sql.Tx) error {
	if tx == nil {
		return fmt.Errorf("migration transaction is required")
	}
	statements := []string{
		`ALTER TABLE empresa_compras_recepciones_avanzadas
			ADD COLUMN IF NOT EXISTS bodega_id INTEGER`,
		`UPDATE empresa_compras_recepciones_avanzadas
			SET bodega_id = 0
			WHERE bodega_id IS NULL`,
		`ALTER TABLE empresa_compras_recepciones_avanzadas
			ALTER COLUMN bodega_id SET DEFAULT 0,
			ALTER COLUMN bodega_id SET NOT NULL`,
		`CREATE INDEX IF NOT EXISTS ix_productos_empresa_estado_id
			ON productos(empresa_id, estado, id)`,
	}
	for _, statement := range statements {
		if _, err := tx.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("migrate products and inventory schema: %w", err)
		}
	}
	return nil
}
