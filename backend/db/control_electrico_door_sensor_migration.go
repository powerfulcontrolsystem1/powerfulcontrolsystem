package db

import (
	"database/sql"
	"fmt"
)

const empresaControlElectricoDoorSensorSchemaFingerprint = "empresa_control_electrico_door_sensor:v1:raspberry-usage-four-input-multiplexed-door-channels"

// applyEmpresaControlElectricoDoorSensorSchemaTx versiona la modalidad de
// barrido de sensores de puertas. La Raspberry conserva la identidad del tunel
// y cada canal fisico queda relacionado con un dispositivo del modulo vigente
// sensor_puertas, siempre dentro de la misma empresa.
func applyEmpresaControlElectricoDoorSensorSchemaTx(tx *sql.Tx) error {
	if tx == nil {
		return fmt.Errorf("migration transaction is required")
	}
	statements := []string{
		`ALTER TABLE IF EXISTS empresa_control_electrico_raspberry_pis ADD COLUMN IF NOT EXISTS uso_tipo TEXT NOT NULL DEFAULT 'domotica'`,
		`ALTER TABLE IF EXISTS empresa_control_electrico_raspberry_pis ADD COLUMN IF NOT EXISTS puerta_reles_salida INTEGER NOT NULL DEFAULT 16`,
		`ALTER TABLE IF EXISTS empresa_control_electrico_raspberry_pis ADD COLUMN IF NOT EXISTS puerta_delay_ms INTEGER NOT NULL DEFAULT 100`,
		`ALTER TABLE IF EXISTS empresa_sensor_puertas_devices ADD COLUMN IF NOT EXISTS source_raspberry_id BIGINT`,
		`ALTER TABLE IF EXISTS empresa_sensor_puertas_devices ADD COLUMN IF NOT EXISTS selector_output INTEGER`,
		`ALTER TABLE IF EXISTS empresa_sensor_puertas_devices ADD COLUMN IF NOT EXISTS selector_input INTEGER`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_empresa_sensor_puertas_raspberry_channel ON empresa_sensor_puertas_devices(empresa_id, source_raspberry_id, selector_output, selector_input) WHERE source_raspberry_id IS NOT NULL AND selector_output IS NOT NULL AND selector_input IS NOT NULL`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_sensor_puertas_raspberry_active ON empresa_sensor_puertas_devices(empresa_id, source_raspberry_id, estado)`,
	}
	for _, statement := range statements {
		if _, err := execTxSQLCompat(tx, statement); err != nil {
			return err
		}
	}
	return nil
}
