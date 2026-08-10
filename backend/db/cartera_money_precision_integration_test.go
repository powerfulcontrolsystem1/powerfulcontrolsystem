package db

import (
	"context"
	"database/sql"
	"os"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestEmpresaCarteraMoneyPrecisionPostgres(t *testing.T) {
	dsn := os.Getenv("PCS_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("PCS_TEST_POSTGRES_DSN is not configured")
	}
	dbConn, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer dbConn.Close()

	t.Run("converts real rounding artifacts to exact numeric balances", func(t *testing.T) {
		tx, err := dbConn.BeginTx(context.Background(), nil)
		if err != nil {
			t.Fatal(err)
		}
		defer tx.Rollback()
		createTempCarteraMoneyTables(t, tx)
		for _, table := range empresaCarteraMoneyTables {
			if _, err := tx.Exec(`INSERT INTO ` + table.name + ` (valor_original,valor_pagado,saldo) VALUES (214200,0.01,214200)`); err != nil {
				t.Fatal(err)
			}
		}
		if err := applyEmpresaCarteraMoneyPrecisionTx(context.Background(), tx); err != nil {
			t.Fatal(err)
		}
		for _, table := range empresaCarteraMoneyTables {
			var columnType, original, paid, balance string
			if err := tx.QueryRow(`SELECT pg_typeof(saldo)::text,valor_original::text,valor_pagado::text,saldo::text FROM `+table.name).
				Scan(&columnType, &original, &paid, &balance); err != nil {
				t.Fatal(err)
			}
			if columnType != "numeric" || original != "214200.00" || paid != "0.01" || balance != "214199.99" {
				t.Fatalf("%s precision result type=%s original=%s paid=%s balance=%s", table.name, columnType, original, paid, balance)
			}
		}
	})

	t.Run("fails closed on business drift", func(t *testing.T) {
		tx, err := dbConn.BeginTx(context.Background(), nil)
		if err != nil {
			t.Fatal(err)
		}
		defer tx.Rollback()
		createTempCarteraMoneyTables(t, tx)
		if _, err := tx.Exec(`INSERT INTO empresa_cuentas_por_pagar (valor_original,valor_pagado,saldo) VALUES (100,10,100)`); err != nil {
			t.Fatal(err)
		}
		if err := applyEmpresaCarteraMoneyPrecisionTx(context.Background(), tx); err == nil {
			t.Fatal("migration must reject material balance drift")
		}
	})
}

func createTempCarteraMoneyTables(t *testing.T, tx *sql.Tx) {
	t.Helper()
	for _, table := range empresaCarteraMoneyTables {
		if _, err := tx.Exec(`CREATE TEMP TABLE ` + table.name + ` (
			id BIGSERIAL PRIMARY KEY,
			valor_original REAL DEFAULT 0,
			valor_pagado REAL DEFAULT 0,
			saldo REAL DEFAULT 0
		) ON COMMIT DROP`); err != nil {
			t.Fatal(err)
		}
	}
}
