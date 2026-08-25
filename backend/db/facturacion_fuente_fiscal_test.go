package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestNormalizeEmpresaFacturacionArtefactoAcceptsFuenteFiscalJSON(t *testing.T) {
	item := &EmpresaFacturacionArtefacto{
		EmpresaID: 12, TipoDocumento: " Comprobante_Pago ", DocumentoCodigo: " cp-venta-1 ",
		TipoArtefacto: " FUENTE_FISCAL_JSON ", StorageRef: "snapshot.json", SHA256: strings.Repeat("a", 64),
		MimeType: "Application/JSON", TamanoBytes: 123,
	}
	if err := normalizeEmpresaFacturacionArtefacto(item); err != nil {
		t.Fatal(err)
	}
	if item.EmpresaID != 12 || item.TipoDocumento != "comprobante_pago" || item.DocumentoCodigo != "CP-VENTA-1" || item.TipoArtefacto != EmpresaFacturacionArtefactoTipoFuenteFiscalJSON || item.MimeType != "application/json" {
		t.Fatalf("normalizacion inesperada: %+v", item)
	}
}

func TestEmpresaFacturacionFuenteFiscalMigrationIsCatalogued(t *testing.T) {
	migrations, err := PlatformMigrations(MigrationTargetEmpresas)
	if err != nil {
		t.Fatal(err)
	}
	for _, migration := range migrations {
		if migration.Version != "20260823-003-facturacion-fuente-fiscal-v1" {
			continue
		}
		if migration.Body != empresaFacturacionFuenteFiscalFingerprint || migration.Apply == nil {
			t.Fatal("la migracion de fuente fiscal debe ser inmutable y ejecutable")
		}
		return
	}
	t.Fatal("no se encontro la migracion de fuente fiscal")
}

func TestEmpresaFacturacionFuenteFiscalPostgresImmutableAndTenantScoped(t *testing.T) {
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

	schema := fmt.Sprintf("facturacion_fuente_fiscal_%d", time.Now().UnixNano())
	if _, err := tx.Exec(`CREATE SCHEMA ` + schema); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(`SET LOCAL search_path TO ` + schema); err != nil {
		t.Fatal(err)
	}
	if err := applyEmpresaFacturacionArtefactosTx(context.Background(), tx); err != nil {
		t.Fatal(err)
	}
	if err := applyEmpresaFacturacionFuenteFiscalTx(context.Background(), tx); err != nil {
		t.Fatal(err)
	}

	base := EmpresaFacturacionArtefacto{
		EmpresaID: 12, TipoDocumento: "comprobante_pago", DocumentoCodigo: "CP-VENTA-1",
		StorageRef: "source-a.json", SHA256: strings.Repeat("a", 64), MimeType: "application/json", TamanoBytes: 321,
	}
	created, err := insertEmpresaFacturacionFuenteFiscalStore(context.Background(), tx, base)
	if err != nil {
		t.Fatal(err)
	}
	if created.EmpresaID != 12 || created.StorageRef != "source-a.json" {
		t.Fatalf("artefacto creado inesperado: %+v", created)
	}

	same := base
	same.StorageRef = "source-retry.json"
	idempotent, err := insertEmpresaFacturacionFuenteFiscalStore(context.Background(), tx, same)
	if err != nil {
		t.Fatalf("repetir el mismo hash debe ser idempotente: %v", err)
	}
	if idempotent.ID != created.ID || idempotent.StorageRef != created.StorageRef {
		t.Fatalf("la repeticion no debe reemplazar metadatos: created=%+v retry=%+v", created, idempotent)
	}

	changed := base
	changed.SHA256 = strings.Repeat("b", 64)
	if _, err := insertEmpresaFacturacionFuenteFiscalStore(context.Background(), tx, changed); !errors.Is(err, ErrEmpresaFacturacionFuenteFiscalInmutable) {
		t.Fatalf("se esperaba rechazo inmutable, obtuvo %v", err)
	}

	otherTenant := base
	otherTenant.EmpresaID = 99
	otherTenant.StorageRef = "tenant-b.json"
	otherTenant.SHA256 = strings.Repeat("c", 64)
	if _, err := insertEmpresaFacturacionFuenteFiscalStore(context.Background(), tx, otherTenant); err != nil {
		t.Fatalf("otra empresa debe tener su propio snapshot: %v", err)
	}
	var tenantACount, tenantBCount int
	if err := tx.QueryRow(`SELECT COUNT(*) FROM empresa_facturacion_artefactos WHERE empresa_id = 12 AND documento_codigo = 'CP-VENTA-1'`).Scan(&tenantACount); err != nil {
		t.Fatal(err)
	}
	if err := tx.QueryRow(`SELECT COUNT(*) FROM empresa_facturacion_artefactos WHERE empresa_id = 99 AND documento_codigo = 'CP-VENTA-1'`).Scan(&tenantBCount); err != nil {
		t.Fatal(err)
	}
	if tenantACount != 1 || tenantBCount != 1 {
		t.Fatalf("aislamiento inesperado: empresa12=%d empresa99=%d", tenantACount, tenantBCount)
	}

	if _, err := tx.Exec(`INSERT INTO empresa_facturacion_artefactos (empresa_id,tipo_documento,documento_codigo,tipo_artefacto,storage_ref,sha256,mime_type,tamano_bytes) VALUES (12,'comprobante_pago','CP-X','inventado','x.bin',$1,'application/octet-stream',1)`, strings.Repeat("d", 64)); err == nil {
		t.Fatal("la restriccion debe rechazar tipos de artefacto no catalogados")
	}
}
