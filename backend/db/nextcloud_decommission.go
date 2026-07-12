package db

import "database/sql"

// EnsureEmpresaNextcloudSchema keeps the per-company Nextcloud assignment.  The
// secret is deliberately not stored here: it belongs only to Nextcloud and the
// super-admin configuration/encrypted runtime environment.
func EnsureEmpresaNextcloudSchema(dbEmpresas *sql.DB) error {
	if dbEmpresas == nil {
		return nil
	}
	_, err := execSQLCompat(dbEmpresas, `CREATE TABLE IF NOT EXISTS empresa_nextcloud_accounts (
		id BIGSERIAL PRIMARY KEY, empresa_id BIGINT NOT NULL UNIQUE,
		nextcloud_user TEXT NOT NULL, quota_mb BIGINT NOT NULL DEFAULT 1024,
		provisioned BOOLEAN NOT NULL DEFAULT FALSE, created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		return err
	}
	_, err = execSQLCompat(dbEmpresas, `INSERT INTO empresa_nextcloud_accounts (empresa_id,nextcloud_user,quota_mb)
		SELECT e.id, 'empresa_' || e.id::text, 1024 FROM empresas e
		ON CONFLICT (empresa_id) DO NOTHING`)
	return err
}

// DecommissionNextcloudArtifacts is retained for source compatibility. Nextcloud
// is again a supported multi-company service, so it must never remove data.
func DecommissionNextcloudArtifacts(dbEmpresas, dbSuper *sql.DB) error {
	return EnsureEmpresaNextcloudSchema(dbEmpresas)
	/*
		if dbEmpresas != nil {
			if _, err := execSQLCompat(dbEmpresas, `DROP TABLE IF EXISTS empresa_nextcloud_accounts`); err != nil {
				return err
			}
		}
		if dbSuper != nil {
			if _, err := execSQLCompat(dbSuper, `
				DELETE FROM configuraciones
				WHERE config_key = 'nextcloud.enabled'
				   OR config_key = 'nextcloud.base_url'
				   OR config_key = 'nextcloud.admin_user'
				   OR config_key = 'nextcloud.admin_secret'
			`); err != nil {
				return err
			}
		}
		return nil
	*/
}
