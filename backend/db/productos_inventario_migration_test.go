package db

import (
	"context"
	"database/sql"
	"os"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestEmpresaProductosInventarioMigrationPostgres(t *testing.T) {
	dsn := os.Getenv("PCS_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("PCS_TEST_POSTGRES_DSN is not configured")
	}
	dbConn, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer dbConn.Close()

	tx, err := dbConn.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()

	for _, statement := range []string{
		`CREATE TEMP TABLE bodegas (
			id BIGINT PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			estado TEXT
		) ON COMMIT DROP`,
		`CREATE TEMP TABLE empresa_compras_recepciones_avanzadas (
			id BIGINT PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			bodega_id INTEGER NOT NULL DEFAULT 0
		) ON COMMIT DROP`,
		`CREATE TEMP TABLE productos (
			id BIGINT PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			estado TEXT
		) ON COMMIT DROP`,
		`INSERT INTO bodegas (id, empresa_id, estado) VALUES
			(11, 1, 'activo'),
			(21, 2, 'activo'),
			(22, 2, 'activo')`,
		`INSERT INTO empresa_compras_recepciones_avanzadas (id, empresa_id, bodega_id) VALUES
			(101, 1, 0),
			(201, 2, 0),
			(202, 2, 999)`,
	} {
		if _, err := tx.Exec(statement); err != nil {
			t.Fatal(err)
		}
	}

	for attempt := 0; attempt < 2; attempt++ {
		if err := applyEmpresaProductosInventarioSchemaTx(context.Background(), tx); err != nil {
			t.Fatalf("attempt %d: %v", attempt+1, err)
		}
	}

	var uniqueWarehouse sql.NullInt64
	if err := tx.QueryRow(`SELECT bodega_id FROM empresa_compras_recepciones_avanzadas WHERE id=101`).Scan(&uniqueWarehouse); err != nil {
		t.Fatal(err)
	}
	if !uniqueWarehouse.Valid || uniqueWarehouse.Int64 != 11 {
		t.Fatalf("single active warehouse backfill=%v, want 11", uniqueWarehouse)
	}
	for _, receiptID := range []int64{201, 202} {
		var ambiguousWarehouse sql.NullInt64
		if err := tx.QueryRow(`SELECT bodega_id FROM empresa_compras_recepciones_avanzadas WHERE id=$1`, receiptID).Scan(&ambiguousWarehouse); err != nil {
			t.Fatal(err)
		}
		if ambiguousWarehouse.Valid {
			t.Fatalf("ambiguous receipt %d warehouse=%d, want NULL", receiptID, ambiguousWarehouse.Int64)
		}
	}

	var nullable string
	var defaultValue sql.NullString
	if err := tx.QueryRow(`SELECT is_nullable, column_default
		FROM information_schema.columns
		WHERE table_schema LIKE 'pg_temp_%'
			AND table_name='empresa_compras_recepciones_avanzadas'
			AND column_name='bodega_id'`).Scan(&nullable, &defaultValue); err != nil {
		t.Fatal(err)
	}
	if nullable != "YES" || defaultValue.Valid {
		t.Fatalf("bodega_id nullable=%s default=%v, want nullable without sentinel default", nullable, defaultValue)
	}
	if _, err := tx.Exec(`INSERT INTO empresa_compras_recepciones_avanzadas (id,empresa_id,bodega_id) VALUES (301,3,0)`); err == nil {
		t.Fatal("migration constraint must reject zero warehouse ids")
	}
}
