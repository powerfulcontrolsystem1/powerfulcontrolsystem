package db

import (
	"context"
	"database/sql"
)

const sessionPrincipalClaimsSchemaFingerprint = "sessions-principal-claims:v1:principal-type:principal-id:empresa-id:role"

func applySessionPrincipalClaimsSchemaTx(ctx context.Context, tx *sql.Tx) error {
	statements := []string{
		`ALTER TABLE sesiones ADD COLUMN IF NOT EXISTS principal_type TEXT NOT NULL DEFAULT 'admin'`,
		`ALTER TABLE sesiones ADD COLUMN IF NOT EXISTS principal_id BIGINT NOT NULL DEFAULT 0`,
		`ALTER TABLE sesiones ADD COLUMN IF NOT EXISTS empresa_id BIGINT NOT NULL DEFAULT 0`,
		`ALTER TABLE sesiones ADD COLUMN IF NOT EXISTS principal_role TEXT NOT NULL DEFAULT ''`,
		`UPDATE sesiones SET principal_type = 'admin' WHERE principal_type IS NULL OR btrim(principal_type) = ''`,
		`CREATE INDEX IF NOT EXISTS ix_sesiones_principal_scope ON sesiones (principal_type, empresa_id, principal_id, activo)`,
	}
	for _, statement := range statements {
		if _, err := tx.ExecContext(ctx, statement); err != nil {
			return err
		}
	}
	return nil
}
