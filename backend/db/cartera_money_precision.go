package db

import (
	"context"
	"database/sql"
	"fmt"
)

const empresaCarteraMoneyPrecisionFingerprint = "empresa-cartera:v1:numeric-18-2:exact-balance-invariant"

type carteraMoneyTable struct {
	name          string
	inspectQuery  string
	alterTypeSQL  string
	recomputeSQL  string
	notNullSQL    string
	constraintSQL string
}

var empresaCarteraMoneyTables = []carteraMoneyTable{
	{
		name: "empresa_cuentas_por_cobrar",
		inspectQuery: `SELECT COUNT(*) FROM empresa_cuentas_por_cobrar WHERE
			ROUND(COALESCE(valor_original, 0)::numeric, 2) < 0 OR
			ROUND(COALESCE(valor_pagado, 0)::numeric, 2) < 0 OR
			ROUND(COALESCE(valor_pagado, 0)::numeric, 2) > ROUND(COALESCE(valor_original, 0)::numeric, 2) OR
			ABS(ROUND(COALESCE(saldo, 0)::numeric, 2) - GREATEST(
				ROUND(COALESCE(valor_original, 0)::numeric, 2) - ROUND(COALESCE(valor_pagado, 0)::numeric, 2), 0
			)) > 0.02`,
		alterTypeSQL: `ALTER TABLE empresa_cuentas_por_cobrar
			ALTER COLUMN valor_original TYPE NUMERIC(18,2) USING ROUND(COALESCE(valor_original, 0)::numeric, 2),
			ALTER COLUMN valor_pagado TYPE NUMERIC(18,2) USING ROUND(COALESCE(valor_pagado, 0)::numeric, 2),
			ALTER COLUMN saldo TYPE NUMERIC(18,2) USING ROUND(COALESCE(saldo, 0)::numeric, 2)`,
		recomputeSQL: `UPDATE empresa_cuentas_por_cobrar SET saldo = GREATEST(valor_original - valor_pagado, 0)`,
		notNullSQL: `ALTER TABLE empresa_cuentas_por_cobrar
			ALTER COLUMN valor_original SET DEFAULT 0,
			ALTER COLUMN valor_original SET NOT NULL,
			ALTER COLUMN valor_pagado SET DEFAULT 0,
			ALTER COLUMN valor_pagado SET NOT NULL,
			ALTER COLUMN saldo SET DEFAULT 0,
			ALTER COLUMN saldo SET NOT NULL`,
		constraintSQL: `ALTER TABLE empresa_cuentas_por_cobrar ADD CONSTRAINT ck_empresa_cxc_money_invariants CHECK (
			valor_original >= 0 AND valor_pagado >= 0 AND valor_pagado <= valor_original AND
			saldo = valor_original - valor_pagado
		)`,
	},
	{
		name: "empresa_cuentas_por_pagar",
		inspectQuery: `SELECT COUNT(*) FROM empresa_cuentas_por_pagar WHERE
			ROUND(COALESCE(valor_original, 0)::numeric, 2) < 0 OR
			ROUND(COALESCE(valor_pagado, 0)::numeric, 2) < 0 OR
			ROUND(COALESCE(valor_pagado, 0)::numeric, 2) > ROUND(COALESCE(valor_original, 0)::numeric, 2) OR
			ABS(ROUND(COALESCE(saldo, 0)::numeric, 2) - GREATEST(
				ROUND(COALESCE(valor_original, 0)::numeric, 2) - ROUND(COALESCE(valor_pagado, 0)::numeric, 2), 0
			)) > 0.02`,
		alterTypeSQL: `ALTER TABLE empresa_cuentas_por_pagar
			ALTER COLUMN valor_original TYPE NUMERIC(18,2) USING ROUND(COALESCE(valor_original, 0)::numeric, 2),
			ALTER COLUMN valor_pagado TYPE NUMERIC(18,2) USING ROUND(COALESCE(valor_pagado, 0)::numeric, 2),
			ALTER COLUMN saldo TYPE NUMERIC(18,2) USING ROUND(COALESCE(saldo, 0)::numeric, 2)`,
		recomputeSQL: `UPDATE empresa_cuentas_por_pagar SET saldo = GREATEST(valor_original - valor_pagado, 0)`,
		notNullSQL: `ALTER TABLE empresa_cuentas_por_pagar
			ALTER COLUMN valor_original SET DEFAULT 0,
			ALTER COLUMN valor_original SET NOT NULL,
			ALTER COLUMN valor_pagado SET DEFAULT 0,
			ALTER COLUMN valor_pagado SET NOT NULL,
			ALTER COLUMN saldo SET DEFAULT 0,
			ALTER COLUMN saldo SET NOT NULL`,
		constraintSQL: `ALTER TABLE empresa_cuentas_por_pagar ADD CONSTRAINT ck_empresa_cxp_money_invariants CHECK (
			valor_original >= 0 AND valor_pagado >= 0 AND valor_pagado <= valor_original AND
			saldo = valor_original - valor_pagado
		)`,
	},
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
		if err := tx.QueryRowContext(ctx, table.inspectQuery).Scan(&incompatible); err != nil {
			return fmt.Errorf("inspect %s money precision: %w", table.name, err)
		}
		if incompatible > 0 {
			return fmt.Errorf("%s contains %d monetary rows requiring manual reconciliation", table.name, incompatible)
		}
	}

	for _, table := range empresaCarteraMoneyTables {
		statements := []string{
			table.alterTypeSQL,
			table.recomputeSQL,
			table.notNullSQL,
			table.constraintSQL,
		}
		for _, statement := range statements {
			if _, err := tx.ExecContext(ctx, statement); err != nil {
				return fmt.Errorf("migrate %s money precision: %w", table.name, err)
			}
		}
	}
	return nil
}
