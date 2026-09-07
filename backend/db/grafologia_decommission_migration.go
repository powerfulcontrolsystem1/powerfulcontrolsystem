package db

import (
	"context"
	"database/sql"
)

const empresaGrafologiaDecommissionFingerprint = "DROP TABLE IF EXISTS empresa_grafologia_analisis"

// applyEmpresaGrafologiaDecommissionTx removes the retired module's only
// database artifact. PostgreSQL removes its dependent indexes atomically with
// the table, and the migration ledger makes the operation idempotent.
func applyEmpresaGrafologiaDecommissionTx(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, empresaGrafologiaDecommissionFingerprint)
	return err
}
