package db

import (
	"database/sql"
	"fmt"
)

const empresaAIUsoDiarioUniqueSchemaFingerprint = "empresa_ai_uso_diario:v1:unique-tenant-provider-model-day"

// applyEmpresaAIUsoDiarioUniqueSchemaTx repairs legacy installations where
// CREATE TABLE IF NOT EXISTS preserved the table but not its UNIQUE clause.
// Existing duplicates deliberately make the migration fail closed so an
// operator can reconcile audit totals instead of silently discarding usage.
func applyEmpresaAIUsoDiarioUniqueSchemaTx(tx *sql.Tx) error {
	if tx == nil {
		return fmt.Errorf("migration transaction is required")
	}
	_, err := execTxSQLCompat(tx, `CREATE UNIQUE INDEX IF NOT EXISTS ux_empresa_ai_uso_diario_tenant_model_day
		ON empresa_ai_uso_diario (empresa_id, provider, model_id, fecha_uso)`)
	return err
}
