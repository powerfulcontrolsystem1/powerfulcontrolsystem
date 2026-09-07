package db

import "database/sql"

// DecommissionNextcloudArtifacts removes obsolete Nextcloud runtime data now that
// companies no longer use the VPS Nextcloud module.
func DecommissionNextcloudArtifacts(dbEmpresas, dbSuper *sql.DB) error {
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
}
