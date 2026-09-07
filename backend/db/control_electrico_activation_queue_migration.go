package db

import (
	"database/sql"
	"fmt"
)

const empresaControlElectricoActivationQueueSchemaFingerprint = "empresa_control_electrico_activation_queue:v1:company-serialized-on-delay"

// applyEmpresaControlElectricoActivationQueueSchemaTx adds a company-wide
// activation reservation clock. It serializes ON commands even when several
// Raspberry agents poll at the same time.
func applyEmpresaControlElectricoActivationQueueSchemaTx(tx *sql.Tx) error {
	if tx == nil {
		return fmt.Errorf("migration transaction is required")
	}
	for _, statement := range []string{
		`ALTER TABLE IF EXISTS empresa_control_electrico_config ADD COLUMN IF NOT EXISTS activation_delay_seconds INTEGER NOT NULL DEFAULT 1`,
		`ALTER TABLE IF EXISTS empresa_control_electrico_config ADD COLUMN IF NOT EXISTS next_activation_at TEXT`,
		`ALTER TABLE IF EXISTS empresa_control_electrico_comandos ADD COLUMN IF NOT EXISTS disponible_desde TEXT`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_control_electrico_comandos_disponible ON empresa_control_electrico_comandos(empresa_id, estado, disponible_desde, id)`,
	} {
		if _, err := execTxSQLCompat(tx, statement); err != nil {
			return err
		}
	}
	return nil
}
