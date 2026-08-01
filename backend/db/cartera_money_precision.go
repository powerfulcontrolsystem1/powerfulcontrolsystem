package db

import (
	"context"
	"database/sql"
	"fmt"
)

const empresaCarteraMoneyPrecisionFingerprint = "empresa-cartera:v1:numeric-18-2:exact-balance-invariant"

type carteraMoneyTable struct {
	name       string
	constraint string
}

var empresaCarteraMoneyTables = []carteraMoneyTable{
	{name: "empresa_cuentas_por_cobrar", constraint: "ck_empresa_cxc_money_invariants"},
	{name: "empresa_cuentas_por_pagar", constraint: "ck_empresa_cxp_money_invariants"},
}

// applyEmpresaCarteraMoneyPrecisionTx replaces PostgreSQL REAL money columns
// with exact two-decimal NUMERIC values. It fails closed when existing business
// drift is larger than a rounding artifact instead of silently rewriting it.
func applyEmpresaCarteraMoneyPrecisionTx(ctx context.Context, tx *sql.Tx) error {
	if tx == nil {
		return fmt.Errorf("migration transaction is required")
	}
	for _, table := range empresaCarteraMoneyTables {
		var incompatible int64
		query := fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE
			ROUND(COALESCE(valor_original, 0)::numeric, 2) < 0 OR
			ROUND(COALESCE(valor_pagado, 0)::numeric, 2) < 0 OR
			ROUND(COALESCE(valor_pagado, 0)::numeric, 2) > ROUND(COALESCE(valor_original, 0)::numeric, 2) OR
			ABS(ROUND(COALESCE(saldo, 0)::numeric, 2) - GREATEST(
				ROUND(COALESCE(valor_original, 0)::numeric, 2) - ROUND(COALESCE(valor_pagado, 0)::numeric, 2), 0
			)) > 0.02`, table.name)
		if err := tx.QueryRowContext(ctx, query).Scan(&incompatible); err != nil {
			return fmt.Errorf("inspect %s money precision: %w", table.name, err)
		}
		if incompatible > 0 {
			return fmt.Errorf("%s contains %d monetary rows requiring manual reconciliation", table.name, incompatible)
		}
	}

	for _, table := range empresaCarteraMoneyTables {
		statements := []string{
			fmt.Sprintf(`ALTER TABLE %s
				ALTER COLUMN valor_original TYPE NUMERIC(18,2) USING ROUND(COALESCE(valor_original, 0)::numeric, 2),
				ALTER COLUMN valor_pagado TYPE NUMERIC(18,2) USING ROUND(COALESCE(valor_pagado, 0)::numeric, 2),
				ALTER COLUMN saldo TYPE NUMERIC(18,2) USING ROUND(COALESCE(saldo, 0)::numeric, 2)`, table.name),
			fmt.Sprintf(`UPDATE %s SET saldo = GREATEST(valor_original - valor_pagado, 0)`, table.name),
			fmt.Sprintf(`ALTER TABLE %s
				ALTER COLUMN valor_original SET DEFAULT 0,
				ALTER COLUMN valor_original SET NOT NULL,
				ALTER COLUMN valor_pagado SET DEFAULT 0,
				ALTER COLUMN valor_pagado SET NOT NULL,
				ALTER COLUMN saldo SET DEFAULT 0,
				ALTER COLUMN saldo SET NOT NULL`, table.name),
			fmt.Sprintf(`ALTER TABLE %s ADD CONSTRAINT %s CHECK (
				valor_original >= 0 AND valor_pagado >= 0 AND valor_pagado <= valor_original AND
				saldo = valor_original - valor_pagado
			)`, table.name, table.constraint),
		}
		for _, statement := range statements {
			if _, err := tx.ExecContext(ctx, statement); err != nil {
				return fmt.Errorf("migrate %s money precision: %w", table.name, err)
			}
		}
	}
	return nil
}
