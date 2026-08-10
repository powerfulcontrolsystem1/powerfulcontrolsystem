package db

import (
	"database/sql"
	"fmt"
)

const empresaControlElectricoTunnelSchemaFingerprint = "empresa_control_electrico_tunnel:v1:device-identity-command-queue-gpio-input-daily-traffic"

// applyEmpresaControlElectricoTunnelSchemaTx versiona el canal saliente HTTPS
// usado por las Raspberry Pi. Los secretos de dispositivo solo se persisten
// como huellas SHA-256 y nunca forman parte de las respuestas administrativas.
func applyEmpresaControlElectricoTunnelSchemaTx(tx *sql.Tx) error {
	if tx == nil {
		return fmt.Errorf("migration transaction is required")
	}
	statements := []string{
		`ALTER TABLE IF EXISTS empresa_control_electrico_raspberry_pis ADD COLUMN IF NOT EXISTS device_uid TEXT`,
		`ALTER TABLE IF EXISTS empresa_control_electrico_raspberry_pis ADD COLUMN IF NOT EXISTS enrollment_token_hash TEXT`,
		`ALTER TABLE IF EXISTS empresa_control_electrico_raspberry_pis ADD COLUMN IF NOT EXISTS enrollment_expires_at TEXT`,
		`ALTER TABLE IF EXISTS empresa_control_electrico_raspberry_pis ADD COLUMN IF NOT EXISTS device_token_hash TEXT`,
		`ALTER TABLE IF EXISTS empresa_control_electrico_raspberry_pis ADD COLUMN IF NOT EXISTS tunnel_enabled INTEGER NOT NULL DEFAULT 0`,
		`ALTER TABLE IF EXISTS empresa_control_electrico_raspberry_pis ADD COLUMN IF NOT EXISTS tunnel_status TEXT NOT NULL DEFAULT 'sin_configurar'`,
		`ALTER TABLE IF EXISTS empresa_control_electrico_raspberry_pis ADD COLUMN IF NOT EXISTS last_seen TEXT`,
		`ALTER TABLE IF EXISTS empresa_control_electrico_raspberry_pis ADD COLUMN IF NOT EXISTS last_ip TEXT`,
		`ALTER TABLE IF EXISTS empresa_control_electrico_raspberry_pis ADD COLUMN IF NOT EXISTS agent_version TEXT`,
		`ALTER TABLE IF EXISTS empresa_control_electrico_raspberry_pis ADD COLUMN IF NOT EXISTS bytes_rx BIGINT NOT NULL DEFAULT 0`,
		`ALTER TABLE IF EXISTS empresa_control_electrico_raspberry_pis ADD COLUMN IF NOT EXISTS bytes_tx BIGINT NOT NULL DEFAULT 0`,
		`ALTER TABLE IF EXISTS empresa_control_electrico_raspberry_pis ADD COLUMN IF NOT EXISTS last_tunnel_error TEXT`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_empresa_control_electrico_raspberry_device_uid ON empresa_control_electrico_raspberry_pis(device_uid) WHERE NULLIF(BTRIM(device_uid), '') IS NOT NULL`,
		`ALTER TABLE IF EXISTS empresa_control_electrico_reglas ADD COLUMN IF NOT EXISTS raspberry_id BIGINT`,
		`ALTER TABLE IF EXISTS empresa_control_electrico_reglas ADD COLUMN IF NOT EXISTS entrada_gpio_pin INTEGER NOT NULL DEFAULT -1`,
		`ALTER TABLE IF EXISTS empresa_control_electrico_reglas ADD COLUMN IF NOT EXISTS entrada_pull TEXT NOT NULL DEFAULT 'none'`,
		`ALTER TABLE IF EXISTS empresa_control_electrico_reglas ADD COLUMN IF NOT EXISTS debounce_ms INTEGER NOT NULL DEFAULT 250`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_control_electrico_reglas_gpio ON empresa_control_electrico_reglas(empresa_id, raspberry_id, entrada_gpio_pin, estado)`,
		`CREATE TABLE IF NOT EXISTS empresa_control_electrico_comandos (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL,
			raspberry_id BIGINT NOT NULL,
			command_uid TEXT NOT NULL,
			rele_id BIGINT,
			estacion_id BIGINT,
			gpio_pin INTEGER,
			estado_objetivo TEXT,
			payload_json TEXT NOT NULL,
			estado TEXT NOT NULL DEFAULT 'pendiente',
			intentos INTEGER NOT NULL DEFAULT 0,
			resultado TEXT,
			error TEXT,
			solicitado_en TEXT DEFAULT (CURRENT_TIMESTAMP),
			entregado_en TEXT,
			completado_en TEXT,
			expira_en TEXT,
			usuario_creador TEXT,
			origen TEXT
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_empresa_control_electrico_comando_uid ON empresa_control_electrico_comandos(command_uid)`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_control_electrico_comandos_cola ON empresa_control_electrico_comandos(empresa_id, raspberry_id, estado, id)`,
		`CREATE TABLE IF NOT EXISTS empresa_control_electrico_trafico_diario (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL,
			raspberry_id BIGINT NOT NULL,
			fecha TEXT NOT NULL,
			bytes_rx BIGINT NOT NULL DEFAULT 0,
			bytes_tx BIGINT NOT NULL DEFAULT 0,
			solicitudes BIGINT NOT NULL DEFAULT 0,
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			UNIQUE(empresa_id, raspberry_id, fecha)
		)`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_control_electrico_trafico_fecha ON empresa_control_electrico_trafico_diario(fecha, empresa_id, raspberry_id)`,
	}
	for _, statement := range statements {
		if _, err := execTxSQLCompat(tx, statement); err != nil {
			return err
		}
	}
	return nil
}
