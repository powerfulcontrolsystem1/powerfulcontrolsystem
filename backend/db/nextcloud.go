package db

import (
	"database/sql"
	"strconv"
)

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
		activo BOOLEAN NOT NULL DEFAULT TRUE,
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
	if err := ensureColumnIfMissing(dbEmpresas, "empresa_nextcloud_accounts", "activo", "BOOLEAN NOT NULL DEFAULT TRUE"); err != nil {
		return err
	}
	_, err = execSQLCompat(dbEmpresas, `CREATE INDEX IF NOT EXISTS idx_empresa_nextcloud_accounts_empresa
		ON empresa_nextcloud_accounts(empresa_id)`)
	return err
}

// EnsureEmpresaNextcloudAssignment creates the technical account assignment
// immediately when a company is created. It does not contact Nextcloud.
func EnsureEmpresaNextcloudAssignment(dbEmpresas *sql.DB, empresaID, quotaMB int64) error {
	if dbEmpresas == nil || empresaID <= 0 {
		return nil
	}
	if quotaMB <= 0 {
		quotaMB = 1024
	}
	if err := EnsureEmpresaNextcloudSchema(dbEmpresas); err != nil {
		return err
	}
	_, err := dbEmpresas.Exec(`INSERT INTO empresa_nextcloud_accounts (empresa_id, nextcloud_user, quota_mb, activo)
		VALUES ($1, $2, $3, TRUE) ON CONFLICT (empresa_id) DO UPDATE SET quota_mb=EXCLUDED.quota_mb, updated_at=CURRENT_TIMESTAMP`, empresaID, "pcs_empresa_"+strconv.FormatInt(empresaID, 10), quotaMB)
	return err
}

// EnsureEmpresaNextcloudAssignmentsForAll applies the global default to every
// existing company, preserving the empresa_id boundary.
func EnsureEmpresaNextcloudAssignmentsForAll(dbEmpresas *sql.DB, quotaMB int64) (int64, error) {
	if dbEmpresas == nil {
		return 0, nil
	}
	if quotaMB <= 0 {
		quotaMB = 1024
	}
	if err := EnsureEmpresaNextcloudSchema(dbEmpresas); err != nil {
		return 0, err
	}
	res, err := dbEmpresas.Exec(`INSERT INTO empresa_nextcloud_accounts (empresa_id, nextcloud_user, quota_mb)
		SELECT id, 'pcs_empresa_' || id::text, $1 FROM empresas
		ON CONFLICT (empresa_id) DO UPDATE SET quota_mb=EXCLUDED.quota_mb, updated_at=CURRENT_TIMESTAMP`, quotaMB)
	if err != nil {
		return 0, err
	}
	count, _ := res.RowsAffected()
	return count, nil
}
