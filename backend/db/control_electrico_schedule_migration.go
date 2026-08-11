package db

import (
	"database/sql"
	"fmt"
)

const empresaControlElectricoScheduleSchemaFingerprint = "empresa_control_electrico_schedule:v1:datetime-start-end"

// applyEmpresaControlElectricoScheduleSchemaTx agrega el rango fechado de
// programacion sin alterar los horarios heredados ni los datos de otra empresa.
func applyEmpresaControlElectricoScheduleSchemaTx(tx *sql.Tx) error {
	if tx == nil {
		return fmt.Errorf("migration transaction is required")
	}
	for _, statement := range []string{
		`ALTER TABLE IF EXISTS empresa_control_electrico_reles ADD COLUMN IF NOT EXISTS programacion_inicio TEXT`,
		`ALTER TABLE IF EXISTS empresa_control_electrico_reles ADD COLUMN IF NOT EXISTS programacion_fin TEXT`,
	} {
		if _, err := execTxSQLCompat(tx, statement); err != nil {
			return err
		}
	}
	return nil
}
