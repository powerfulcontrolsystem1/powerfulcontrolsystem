package db

import "database/sql"

// EnsureEmpresaNextcloudSchema stores only the technical assignment. Passwords
// stay in Nextcloud and temporary credentials are returned once to the caller.
func EnsureEmpresaNextcloudSchema(dbEmpresas *sql.DB) error {
	if dbEmpresas == nil {
		return nil
	}
	_, err := execSQLCompat(dbEmpresas, `CREATE TABLE IF NOT EXISTS empresa_nextcloud_accounts (
		id BIGSERIAL PRIMARY KEY,
		empresa_id BIGINT NOT NULL UNIQUE REFERENCES empresas(id) ON DELETE CASCADE,
		nextcloud_user TEXT NOT NULL UNIQUE,
		quota_mb BIGINT NOT NULL DEFAULT 1024 CHECK (quota_mb > 0),
		provisioned BOOLEAN NOT NULL DEFAULT FALSE,
		provisioned_at TIMESTAMP,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbEmpresas, "empresa_nextcloud_accounts", "provisioned_at", "TIMESTAMP"); err != nil {
		return err
	}
	_, err = execSQLCompat(dbEmpresas, `CREATE INDEX IF NOT EXISTS idx_empresa_nextcloud_accounts_empresa
		ON empresa_nextcloud_accounts(empresa_id)`)
	return err
}
