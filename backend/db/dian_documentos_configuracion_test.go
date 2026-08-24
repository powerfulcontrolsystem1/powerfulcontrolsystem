package db

import (
	"context"
	"database/sql"
	"os"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestNormalizeEmpresaDIANDocumentoConfiguracionKeepsRangesSeparate(t *testing.T) {
	item := &EmpresaDIANDocumentoConfiguracion{
		EmpresaID:         12,
		TipoDocumento:     " Documento_Soporte ",
		Estado:            " CONFIGURANDO ",
		TipoAmbiente:      " PRODUCCION ",
		Prefijo:           " ds ",
		RangoDesde:        1,
		RangoHasta:        100,
		ConsecutivoActual: 1,
	}
	if err := normalizeEmpresaDIANDocumentoConfiguracion(item); err != nil {
		t.Fatal(err)
	}
	if item.TipoDocumento != "documento_soporte" || item.Prefijo != "DS" {
		t.Fatalf("normalizacion inesperada: %+v", item)
	}
}

func TestNormalizeEmpresaDIANDocumentoConfiguracionRejectsInvalidRange(t *testing.T) {
	item := &EmpresaDIANDocumentoConfiguracion{EmpresaID: 12, TipoDocumento: "documento_soporte", RangoDesde: 20, RangoHasta: 10}
	if err := normalizeEmpresaDIANDocumentoConfiguracion(item); err == nil {
		t.Fatal("se esperaba rechazo del rango inverso")
	}
}

func TestNormalizeEmpresaDIANDocumentoConfiguracionRejectsPrivateOverride(t *testing.T) {
	item := &EmpresaDIANDocumentoConfiguracion{
		EmpresaID: 12, TipoDocumento: "documento_soporte", URLDIANOverride: "https://127.0.0.1:8443",
	}
	if err := normalizeEmpresaDIANDocumentoConfiguracion(item); err == nil {
		t.Fatal("se esperaba rechazo de URL privada")
	}
}

func TestEmpresaDIANDocumentosConfiguracionMigrationIsCatalogued(t *testing.T) {
	migrations, err := PlatformMigrations(MigrationTargetEmpresas)
	if err != nil {
		t.Fatal(err)
	}
	for _, migration := range migrations {
		if migration.Version != "20260823-001-dian-documentos-configuracion-v1" {
			continue
		}
		if migration.Body != empresaDIANDocumentosConfiguracionFingerprint || migration.Apply == nil {
			t.Fatal("la migracion DIAN por documento debe ser inmutable y ejecutable")
		}
		return
	}
	t.Fatal("no se encontro la migracion DIAN por documento")
}

func TestEmpresaDIANDocumentosConfiguracionPostgres(t *testing.T) {
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
	if _, err := tx.Exec(`CREATE SCHEMA dian_documentos_configuracion_test`); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(`SET LOCAL search_path TO dian_documentos_configuracion_test`); err != nil {
		t.Fatal(err)
	}
	if err := applyEmpresaDIANDocumentosConfiguracionTx(context.Background(), tx); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(`INSERT INTO empresa_dian_documentos_configuracion (empresa_id, tipo_documento, rango_desde, rango_hasta) VALUES (12, 'documento_soporte', 1, 10)`); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(`SAVEPOINT dian_documentos_configuracion_unique`); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(`INSERT INTO empresa_dian_documentos_configuracion (empresa_id, tipo_documento, rango_desde, rango_hasta) VALUES (12, 'documento_soporte', 1, 10)`); err == nil {
		t.Fatal("se esperaba clave unica empresa/tipo")
	}
	if _, err := tx.Exec(`ROLLBACK TO SAVEPOINT dian_documentos_configuracion_unique`); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(`INSERT INTO empresa_dian_documentos_configuracion (empresa_id, tipo_documento, rango_desde, rango_hasta) VALUES (12, 'documento_soporte_2', 11, 10)`); err == nil {
		t.Fatal("se esperaba rechazo del rango inverso")
	}
}
