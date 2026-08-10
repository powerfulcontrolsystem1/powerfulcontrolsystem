package db

import (
	"database/sql"
	"fmt"
)

const empresaControlElectricoStationAutomationSchemaFingerprint = "empresa_control_electrico_station_automation:v1:per-relay-on-off-policies"

// applyEmpresaControlElectricoStationAutomationSchemaTx stores independent
// station-transition policies per relay. Defaults preserve the prior
// seguimiento_estacion behavior for already configured equipment.
func applyEmpresaControlElectricoStationAutomationSchemaTx(tx *sql.Tx) error {
	if tx == nil {
		return fmt.Errorf("migration transaction is required")
	}
	for _, statement := range []string{
		`ALTER TABLE IF EXISTS empresa_control_electrico_reles ADD COLUMN IF NOT EXISTS encender_al_activar_estacion INTEGER NOT NULL DEFAULT 1`,
		`ALTER TABLE IF EXISTS empresa_control_electrico_reles ADD COLUMN IF NOT EXISTS apagar_al_desactivar_estacion INTEGER NOT NULL DEFAULT 1`,
	} {
		if _, err := execTxSQLCompat(tx, statement); err != nil {
			return err
		}
	}
	return nil
}
