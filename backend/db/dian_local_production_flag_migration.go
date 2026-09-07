package db

import (
	"context"
	"database/sql"
	"fmt"
)

const empresaDIANLocalProductionFlagFingerprint = "empresa-dian:v1:local-production-flag:historical-evidence-only"

// applyEmpresaDIANLocalProductionFlagTx separates the local production gate
// from estado_dian, which is mutable because it records the latest provider
// response. Historical activation is preserved only when both the production
// environment and an explicit activation trace are present.
func applyEmpresaDIANLocalProductionFlagTx(ctx context.Context, tx *sql.Tx) error {
	if tx == nil {
		return fmt.Errorf("migration transaction is required")
	}
	statements := []string{
		`ALTER TABLE empresa_dian_configuracion
			ADD COLUMN IF NOT EXISTS produccion_local_activa INTEGER`,
		`UPDATE empresa_dian_configuracion
			SET produccion_local_activa = 0
			WHERE produccion_local_activa IS NULL`,
		`UPDATE empresa_dian_configuracion
			SET produccion_local_activa = 1
			WHERE produccion_local_activa = 0
			  AND LOWER(TRIM(COALESCE(tipo_ambiente, ''))) = 'produccion'
			  AND (
				LOWER(TRIM(COALESCE(estado_dian, ''))) = 'produccion_local_activa'
				OR LOWER(COALESCE(observaciones, '')) LIKE '%por dian_activar_produccion_local%'
			  )`,
		`ALTER TABLE empresa_dian_configuracion
			ALTER COLUMN produccion_local_activa SET DEFAULT 0,
			ALTER COLUMN produccion_local_activa SET NOT NULL`,
		`DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM pg_constraint
				WHERE conname = 'ck_empresa_dian_produccion_local_activa'
				  AND conrelid = 'empresa_dian_configuracion'::regclass
			) THEN
				ALTER TABLE empresa_dian_configuracion
					ADD CONSTRAINT ck_empresa_dian_produccion_local_activa
					CHECK (produccion_local_activa IN (0, 1));
			END IF;
		END $$`,
	}
	for _, statement := range statements {
		if _, err := tx.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("migrate DIAN local production flag: %w", err)
		}
	}
	return nil
}
