package db

import (
	"context"
	"database/sql"
	"os"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestEmpresaCatalogIncludesDIANLocalProductionFlagMigration(t *testing.T) {
	t.Parallel()
	migrations, err := PlatformMigrations(MigrationTargetEmpresas)
	if err != nil {
		t.Fatal(err)
	}
	for _, migration := range migrations {
		if migration.Version != "20260809-001-dian-local-production-flag-v1" {
			continue
		}
		if migration.Body != empresaDIANLocalProductionFlagFingerprint || migration.Apply == nil {
			t.Fatal("DIAN local production flag migration must be immutable and executable")
		}
		return
	}
	t.Fatal("DIAN local production flag migration is missing from empresas catalog")
}

func TestEmpresaDIANLocalProductionFlagPostgres(t *testing.T) {
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
	if _, err := tx.Exec(`CREATE TEMP TABLE empresa_dian_configuracion (
		id BIGSERIAL PRIMARY KEY,
		tipo_ambiente TEXT,
		estado_dian TEXT,
		observaciones TEXT
	) ON COMMIT DROP`); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(`INSERT INTO empresa_dian_configuracion (tipo_ambiente, estado_dian, observaciones) VALUES
		('produccion', 'produccion_local_activa', ''),
		('produccion', 'enviado', 'Activado por dian_activar_produccion_local'),
		('produccion', 'enviado', ''),
		('habilitacion', 'produccion_local_activa', 'por dian_activar_produccion_local'),
		(NULL, NULL, NULL)`); err != nil {
		t.Fatal(err)
	}
	for attempt := 0; attempt < 2; attempt++ {
		if err := applyEmpresaDIANLocalProductionFlagTx(context.Background(), tx); err != nil {
			t.Fatalf("attempt %d: %v", attempt+1, err)
		}
	}
	rows, err := tx.Query(`SELECT produccion_local_activa FROM empresa_dian_configuracion ORDER BY id`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	want := []int{1, 1, 0, 0, 0}
	for index := 0; rows.Next(); index++ {
		var got int
		if err := rows.Scan(&got); err != nil {
			t.Fatal(err)
		}
		if index >= len(want) || got != want[index] {
			t.Fatalf("row %d flag=%d, want %d", index+1, got, want[index])
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	var nullable, defaultValue string
	if err := tx.QueryRow(`SELECT is_nullable, column_default
		FROM information_schema.columns
		WHERE table_schema LIKE 'pg_temp_%'
		  AND table_name = 'empresa_dian_configuracion'
		  AND column_name = 'produccion_local_activa'`).Scan(&nullable, &defaultValue); err != nil {
		t.Fatal(err)
	}
	if nullable != "NO" || defaultValue != "0" {
		t.Fatalf("flag nullable=%s default=%s", nullable, defaultValue)
	}
}
