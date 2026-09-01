package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestEmpresaFacturacionDANECodesMigrationIsCatalogued(t *testing.T) {
	migrations, err := PlatformMigrations(MigrationTargetEmpresas)
	if err != nil {
		t.Fatal(err)
	}
	for _, migration := range migrations {
		if migration.Version == "20260824-001-facturacion-dane-codes-v1" {
			if migration.Body != empresaFacturacionDANECodesFingerprint || migration.Apply == nil {
				t.Fatal("migracion DANE incompleta")
			}
			return
		}
	}
	t.Fatal("migracion DANE no catalogada")
}

func TestNormalizeFacturacionDANECodesRejectsInvalidOrCrossDepartmentValues(t *testing.T) {
	if departamento, municipio, err := normalizeFacturacionDANECodes("25", "25175"); err != nil || departamento != "25" || municipio != "25175" {
		t.Fatalf("codigos validos rechazados: departamento=%q municipio=%q err=%v", departamento, municipio, err)
	}
	for _, testCase := range [][2]string{{"2", "25175"}, {"25", "11001"}, {"AA", "AA001"}, {"", "25175"}} {
		if _, _, err := normalizeFacturacionDANECodes(testCase[0], testCase[1]); err == nil {
			t.Fatalf("se esperaban codigos DANE invalidos: %v", testCase)
		}
	}
}

func TestEmpresaFacturacionDANECodesPostgresDoesNotInferExistingRows(t *testing.T) {
	dsn := strings.TrimSpace(os.Getenv("PCS_TEST_POSTGRES_DSN"))
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
	schema := fmt.Sprintf("facturacion_dane_%d", time.Now().UnixNano())
	if _, err := tx.Exec(`CREATE SCHEMA ` + schema); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(`SET LOCAL search_path TO ` + schema); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(`CREATE TABLE empresa_configuracion_avanzada (id BIGSERIAL PRIMARY KEY, empresa_id INTEGER NOT NULL, departamento TEXT, municipio TEXT)`); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(`CREATE TABLE clientes (id BIGSERIAL PRIMARY KEY, empresa_id INTEGER NOT NULL, departamento TEXT, municipio TEXT)`); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(`INSERT INTO empresa_configuracion_avanzada (empresa_id,departamento,municipio) VALUES (12,'Cundinamarca','Chia')`); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(`INSERT INTO clientes (empresa_id,departamento,municipio) VALUES (12,'Cundinamarca','Chia')`); err != nil {
		t.Fatal(err)
	}
	if err := applyEmpresaFacturacionDANECodesTx(context.Background(), tx); err != nil {
		t.Fatal(err)
	}
	var emisorDepartamento, emisorMunicipio, clienteDepartamento, clienteMunicipio sql.NullString
	if err := tx.QueryRow(`SELECT departamento_codigo_dane, municipio_codigo_dane FROM empresa_configuracion_avanzada WHERE empresa_id=12`).Scan(&emisorDepartamento, &emisorMunicipio); err != nil {
		t.Fatal(err)
	}
	if err := tx.QueryRow(`SELECT departamento_codigo_dane, municipio_codigo_dane FROM clientes WHERE empresa_id=12`).Scan(&clienteDepartamento, &clienteMunicipio); err != nil {
		t.Fatal(err)
	}
	if emisorDepartamento.Valid || emisorMunicipio.Valid || clienteDepartamento.Valid || clienteMunicipio.Valid {
		t.Fatal("la migracion no debe inventar codigos DANE")
	}
}
