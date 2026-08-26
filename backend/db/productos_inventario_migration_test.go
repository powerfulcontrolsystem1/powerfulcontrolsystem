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
		if err := applyEmpresaProductosInventarioLegacyWarehouseSchemaTx(context.Background(), tx); err != nil {
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

func TestEmpresaProductosInventarioMigrationOnRestoredSchemaPostgres(t *testing.T) {
	if os.Getenv("PCS_TEST_POSTGRES_LIVE_SCHEMA") != "isolated" {
		t.Skip("requires PCS_TEST_POSTGRES_LIVE_SCHEMA=isolated and a disposable restored database")
	}
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

	var receiptsBefore int
	if err := tx.QueryRow(`SELECT count(*) FROM empresa_compras_recepciones_avanzadas`).Scan(&receiptsBefore); err != nil {
		t.Fatal(err)
	}
	if err := applyEmpresaProductosInventarioSchemaTx(context.Background(), tx); err != nil {
		t.Fatalf("v1: %v", err)
	}
	for attempt := 0; attempt < 2; attempt++ {
		if err := applyEmpresaProductosInventarioLegacyWarehouseSchemaTx(context.Background(), tx); err != nil {
			t.Fatalf("attempt %d: %v", attempt+1, err)
		}
	}

	var receiptsAfter, invalidWarehouseRefs int
	if err := tx.QueryRow(`SELECT count(*) FROM empresa_compras_recepciones_avanzadas`).Scan(&receiptsAfter); err != nil {
		t.Fatal(err)
	}
	if receiptsAfter != receiptsBefore {
		t.Fatalf("receipt count changed from %d to %d", receiptsBefore, receiptsAfter)
	}
	if err := tx.QueryRow(`SELECT count(*)
		FROM empresa_compras_recepciones_avanzadas AS recepcion
		WHERE bodega_id IS NOT NULL
			AND (
				bodega_id <= 0
				OR NOT EXISTS (
					SELECT 1 FROM bodegas AS bodega
					WHERE bodega.empresa_id=recepcion.empresa_id
						AND bodega.id=recepcion.bodega_id
				)
			)`).Scan(&invalidWarehouseRefs); err != nil {
		t.Fatal(err)
	}
	if invalidWarehouseRefs != 0 {
		t.Fatalf("migration left %d invalid warehouse references", invalidWarehouseRefs)
	}

	var nullable string
	var defaultValue sql.NullString
	if err := tx.QueryRow(`SELECT is_nullable, column_default
		FROM information_schema.columns
		WHERE table_schema='public'
			AND table_name='empresa_compras_recepciones_avanzadas'
			AND column_name='bodega_id'`).Scan(&nullable, &defaultValue); err != nil {
		t.Fatal(err)
	}
	if nullable != "YES" || defaultValue.Valid {
		t.Fatalf("restored bodega_id nullable=%s default=%v", nullable, defaultValue)
	}

	var constraintCount, indexCount int
	if err := tx.QueryRow(`SELECT count(*) FROM pg_constraint
		WHERE conname='ck_empresa_compras_recepciones_bodega_positiva'
			AND conrelid='empresa_compras_recepciones_avanzadas'::regclass`).Scan(&constraintCount); err != nil {
		t.Fatal(err)
	}
	if err := tx.QueryRow(`SELECT count(*) FROM pg_indexes
		WHERE schemaname='public' AND tablename='productos'
			AND indexname='ix_productos_empresa_estado_id'`).Scan(&indexCount); err != nil {
		t.Fatal(err)
	}
	if constraintCount != 1 || indexCount != 1 {
		t.Fatalf("constraint=%d index=%d, want both present", constraintCount, indexCount)
	}
}

func TestEmpresaCatalogosNombreUniqueMigrationPostgres(t *testing.T) {
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
		`CREATE TEMP TABLE categorias_productos (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			nombre TEXT NOT NULL
		) ON COMMIT DROP`,
		`CREATE TEMP TABLE proveedores (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			nombre TEXT NOT NULL
		) ON COMMIT DROP`,
		`INSERT INTO categorias_productos (empresa_id,nombre) VALUES (1,'Retail')`,
		`INSERT INTO proveedores (empresa_id,nombre) VALUES (1,'Proveedor Norte')`,
	} {
		if _, err := tx.Exec(statement); err != nil {
			t.Fatal(err)
		}
	}

	for attempt := 0; attempt < 2; attempt++ {
		if err := applyEmpresaCatalogosNombreUniqueSchemaTx(context.Background(), tx); err != nil {
			t.Fatalf("attempt %d: %v", attempt+1, err)
		}
	}
	if _, err := tx.Exec(`INSERT INTO categorias_productos (empresa_id,nombre) VALUES (2,'  retail  ')`); err != nil {
		t.Fatalf("same normalized category name in another tenant must remain valid: %v", err)
	}
	if _, err := tx.Exec(`INSERT INTO proveedores (empresa_id,nombre) VALUES (2,'PROVEEDOR NORTE')`); err != nil {
		t.Fatalf("same normalized provider name in another tenant must remain valid: %v", err)
	}

	for _, duplicate := range []struct {
		name      string
		savepoint string
		statement string
	}{
		{name: "category", savepoint: "duplicate_category", statement: `INSERT INTO categorias_productos (empresa_id,nombre) VALUES (1,'  RETAIL ')`},
		{name: "provider", savepoint: "duplicate_provider", statement: `INSERT INTO proveedores (empresa_id,nombre) VALUES (1,' proveedor norte ')`},
	} {
		if _, err := tx.Exec("SAVEPOINT " + duplicate.savepoint); err != nil {
			t.Fatal(err)
		}
		if _, err := tx.Exec(duplicate.statement); err == nil {
			t.Fatalf("normalized duplicate %s name must be rejected", duplicate.name)
		}
		if _, err := tx.Exec("ROLLBACK TO SAVEPOINT " + duplicate.savepoint); err != nil {
			t.Fatal(err)
		}
	}
}
