package db

import (
	"database/sql"
	"fmt"
)

const empresaControlElectricoSSHCredentialsSchemaFingerprint = "empresa_control_electrico_ssh_credentials:v1:tenant-scoped-encrypted-passwords-host-key-pinning"

// applyEmpresaControlElectricoSSHCredentialsSchemaTx agrega el perfil visible y
// los secretos cifrados de instalacion SSH. El API nunca serializa las columnas
// *_enc y la migracion no admite credenciales en texto plano.
func applyEmpresaControlElectricoSSHCredentialsSchemaTx(tx *sql.Tx) error {
	if tx == nil {
		return fmt.Errorf("migration transaction is required")
	}
	statements := []string{
		`ALTER TABLE IF EXISTS empresa_control_electrico_raspberry_pis ADD COLUMN IF NOT EXISTS ssh_host TEXT`,
		`ALTER TABLE IF EXISTS empresa_control_electrico_raspberry_pis ADD COLUMN IF NOT EXISTS ssh_port INTEGER NOT NULL DEFAULT 22`,
		`ALTER TABLE IF EXISTS empresa_control_electrico_raspberry_pis ADD COLUMN IF NOT EXISTS ssh_username TEXT`,
		`ALTER TABLE IF EXISTS empresa_control_electrico_raspberry_pis ADD COLUMN IF NOT EXISTS ssh_password_enc TEXT`,
		`ALTER TABLE IF EXISTS empresa_control_electrico_raspberry_pis ADD COLUMN IF NOT EXISTS ssh_sudo_password_enc TEXT`,
		`ALTER TABLE IF EXISTS empresa_control_electrico_raspberry_pis ADD COLUMN IF NOT EXISTS ssh_host_key_fingerprint TEXT`,
		`ALTER TABLE IF EXISTS empresa_control_electrico_raspberry_pis ADD COLUMN IF NOT EXISTS ssh_credentials_updated_at TEXT`,
		`ALTER TABLE IF EXISTS empresa_control_electrico_raspberry_pis ADD COLUMN IF NOT EXISTS ssh_credentials_updated_by TEXT`,
	}
	for _, statement := range statements {
		if _, err := execTxSQLCompat(tx, statement); err != nil {
			return err
		}
	}
	return nil
}
