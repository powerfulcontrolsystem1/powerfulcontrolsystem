package db

import (
	"context"
	"database/sql"
	"fmt"
)

const (
	empresaProductosInventarioSchemaFingerprint                = "empresa-productos-inventario:v1:recepcion-bodega-not-null:productos-tenant-state-index"
	empresaProductosInventarioLegacyWarehouseSchemaFingerprint = "empresa-productos-inventario:v2:recepcion-bodega-legacy-nullable-positive"
	empresaServiciosNombreUniqueSchemaFingerprint              = "empresa-servicios:v1:tenant-normalized-name-unique"
)

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

// applyEmpresaProductosInventarioLegacyWarehouseSchemaTx repairs the sentinel
// introduced by v1 without changing its published checksum. Unknown legacy
// warehouses remain NULL; new receipts are still required and tenant-validated
// by the write path.
func applyEmpresaProductosInventarioLegacyWarehouseSchemaTx(ctx context.Context, tx *sql.Tx) error {
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
	}
	for _, statement := range statements {
		if _, err := tx.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("migrate products and inventory schema: %w", err)
		}
	}
	return nil
}

// applyEmpresaServiciosNombreUniqueSchemaTx prevents two concurrent writes from
// creating the same logical service name inside one tenant. Optional blank
// service codes remain NULL and therefore independent from this name invariant.
// Existing duplicate groups fail closed so an operator can reconcile them
// explicitly instead of silently changing business catalog data.
func applyEmpresaServiciosNombreUniqueSchemaTx(ctx context.Context, tx *sql.Tx) error {
	if tx == nil {
		return fmt.Errorf("migration transaction is required")
	}

	var duplicateGroups int64
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*)
		FROM (
			SELECT empresa_id, LOWER(BTRIM(nombre)) AS nombre_normalizado
			FROM servicios
			WHERE BTRIM(nombre) <> ''
			GROUP BY empresa_id, LOWER(BTRIM(nombre))
			HAVING COUNT(*) > 1
		) AS duplicados`).Scan(&duplicateGroups); err != nil {
		return fmt.Errorf("inspect duplicate enterprise service names: %w", err)
	}
	if duplicateGroups > 0 {
		return fmt.Errorf("cannot enforce unique enterprise service names: %d duplicate tenant/name groups require reconciliation", duplicateGroups)
	}

	if _, err := tx.ExecContext(ctx, `CREATE UNIQUE INDEX IF NOT EXISTS ux_servicios_empresa_nombre_normalizado
		ON servicios(empresa_id, LOWER(BTRIM(nombre)))
		WHERE BTRIM(nombre) <> ''`); err != nil {
		return fmt.Errorf("enforce unique enterprise service names: %w", err)
	}
	return nil
}
