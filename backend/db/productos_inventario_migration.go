package db

import (
	"context"
	"database/sql"
	"fmt"
)

const empresaProductosInventarioSchemaFingerprint = "empresa-productos-inventario:v1:recepcion-bodega-legacy-nullable-positive:productos-tenant-state-index"

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
		`ALTER TABLE empresa_compras_recepciones_avanzadas
			ALTER COLUMN bodega_id DROP DEFAULT,
			ALTER COLUMN bodega_id DROP NOT NULL`,
		`UPDATE empresa_compras_recepciones_avanzadas AS recepcion
			SET bodega_id = NULL
			WHERE bodega_id IS NOT NULL
				AND (
					bodega_id <= 0
					OR NOT EXISTS (
						SELECT 1
						FROM bodegas AS bodega
						WHERE bodega.empresa_id = recepcion.empresa_id
							AND bodega.id = recepcion.bodega_id
					)
				)`,
		`WITH bodega_unica AS (
			SELECT empresa_id, MIN(id) AS bodega_id
			FROM bodegas
			WHERE LOWER(BTRIM(COALESCE(estado, ''))) = 'activo'
			GROUP BY empresa_id
			HAVING COUNT(*) = 1
		)
		UPDATE empresa_compras_recepciones_avanzadas AS recepcion
		SET bodega_id = bodega_unica.bodega_id
		FROM bodega_unica
		WHERE recepcion.empresa_id = bodega_unica.empresa_id
			AND recepcion.bodega_id IS NULL`,
		`ALTER TABLE empresa_compras_recepciones_avanzadas
			DROP CONSTRAINT IF EXISTS ck_empresa_compras_recepciones_bodega_positiva`,
		`ALTER TABLE empresa_compras_recepciones_avanzadas
			ADD CONSTRAINT ck_empresa_compras_recepciones_bodega_positiva
			CHECK (bodega_id IS NULL OR bodega_id > 0)`,
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
