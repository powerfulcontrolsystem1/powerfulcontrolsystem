package db

import (
	"database/sql"
	"fmt"
)

const empresaControlElectricoTimerSchemaFingerprint = "empresa_control_electrico_timer:v1:rule-duration-seconds"

func applyEmpresaControlElectricoTimerSchemaTx(tx *sql.Tx) error {
	if tx == nil {
		return fmt.Errorf("migration transaction is required")
	}
	_, err := tx.Exec(`ALTER TABLE IF EXISTS empresa_control_electrico_reglas ADD COLUMN IF NOT EXISTS temporizador_segundos INTEGER NOT NULL DEFAULT 0`)
	return err
}
