package db

import (
	"context"
	"database/sql"
)

const empresaNextcloudSchemaFingerprint = "empresa_nextcloud_accounts:v1"

// applyEmpresaNextcloudSchemaTx is the checksummed migration counterpart of
// the legacy bootstrap. It is intentionally DDL-only: provisioning accounts
// remains an explicit business operation after the schema exists.
func applyEmpresaNextcloudSchemaTx(_ context.Context, tx *sql.Tx) error {
	for _, statement := range []string{
		`CREATE TABLE IF NOT EXISTS empresa_nextcloud_accounts (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL UNIQUE REFERENCES empresas(id) ON DELETE CASCADE,
			nextcloud_user TEXT NOT NULL UNIQUE,
			quota_mb BIGINT NOT NULL DEFAULT 1024 CHECK (quota_mb > 0),
			activo BOOLEAN NOT NULL DEFAULT TRUE,
			provisioned BOOLEAN NOT NULL DEFAULT FALSE,
			provisioned_at TIMESTAMP,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`ALTER TABLE empresa_nextcloud_accounts ADD COLUMN IF NOT EXISTS provisioned_at TIMESTAMP`,
		`ALTER TABLE empresa_nextcloud_accounts ADD COLUMN IF NOT EXISTS activo BOOLEAN NOT NULL DEFAULT TRUE`,
		`CREATE INDEX IF NOT EXISTS idx_empresa_nextcloud_accounts_empresa ON empresa_nextcloud_accounts(empresa_id)`,
	} {
		if _, err := execTxSQLCompat(tx, statement); err != nil {
			return err
		}
	}
	return nil
}
