package db

import (
	"database/sql"
	"fmt"
)

const empresaControlElectricoRestartCategorySchemaFingerprint = "empresa_control_electrico_restart_category:v1:boot-id-category"

// applyEmpresaControlElectricoRestartCategorySchemaTx prepara la restauracion
// idempotente tras reinicio y la clasificacion visible de equipos.
func applyEmpresaControlElectricoRestartCategorySchemaTx(tx *sql.Tx) error {
	if tx == nil {
		return fmt.Errorf("migration transaction is required")
	}
	for _, statement := range []string{
		`ALTER TABLE IF EXISTS empresa_control_electrico_raspberry_pis ADD COLUMN IF NOT EXISTS last_boot_id TEXT`,
		`ALTER TABLE IF EXISTS empresa_control_electrico_reles ADD COLUMN IF NOT EXISTS categoria TEXT`,
	} {
		if _, err := execTxSQLCompat(tx, statement); err != nil {
			return err
		}
	}
	return nil
}
