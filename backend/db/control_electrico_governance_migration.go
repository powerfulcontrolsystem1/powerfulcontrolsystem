package db

import (
	"database/sql"
	"fmt"
)

const empresaControlElectricoGovernanceSchemaFingerprint = "empresa_control_electrico_governance:v1:disconnect-alerts-monthly-bandwidth-policy"

// applyEmpresaControlElectricoGovernanceSchemaTx agrega gobierno operativo al
// tunel: alertas de desconexion con gracia y una cuota mensual administrada
// exclusivamente desde el panel super.
func applyEmpresaControlElectricoGovernanceSchemaTx(tx *sql.Tx) error {
	if tx == nil {
		return fmt.Errorf("migration transaction is required")
	}
	statements := []string{
		`ALTER TABLE IF EXISTS empresa_control_electrico_config ADD COLUMN IF NOT EXISTS disconnect_alert_enabled INTEGER NOT NULL DEFAULT 0`,
		`ALTER TABLE IF EXISTS empresa_control_electrico_config ADD COLUMN IF NOT EXISTS disconnect_alert_email TEXT`,
		`ALTER TABLE IF EXISTS empresa_control_electrico_config ADD COLUMN IF NOT EXISTS disconnect_grace_minutes INTEGER NOT NULL DEFAULT 5`,
		`ALTER TABLE IF EXISTS empresa_control_electrico_raspberry_pis ADD COLUMN IF NOT EXISTS disconnect_alerted_for_last_seen TEXT`,
		`ALTER TABLE IF EXISTS empresa_control_electrico_raspberry_pis ADD COLUMN IF NOT EXISTS disconnect_alerted_at TEXT`,
		`CREATE TABLE IF NOT EXISTS empresa_control_electrico_limites_tunel (
			empresa_id BIGINT PRIMARY KEY,
			limite_mensual_mb BIGINT NOT NULL DEFAULT 2048,
			alerta_porcentaje INTEGER NOT NULL DEFAULT 80,
			bloquear_al_superar INTEGER NOT NULL DEFAULT 1,
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			observaciones TEXT
		)`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_control_electrico_trafico_empresa_fecha ON empresa_control_electrico_trafico_diario(empresa_id, fecha)`,
	}
	for _, statement := range statements {
		if _, err := execTxSQLCompat(tx, statement); err != nil {
			return err
		}
	}
	return nil
}
