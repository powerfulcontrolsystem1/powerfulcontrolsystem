package db

import (
	"os"
	"strings"
	"testing"
)

func TestPrepareFacturacionDocumentoLegalLocksTenantConsecutive(t *testing.T) {
	raw, err := os.ReadFile("facturacion_electronica.go")
	if err != nil {
		t.Fatalf("read facturacion_electronica.go: %v", err)
	}
	src := strings.ReplaceAll(string(raw), "\r\n", "\n")
	start := strings.Index(src, "func PrepareFacturacionDocumentoLegalContext(")
	end := strings.Index(src[start:], "func normalizeFacturacionRetryEstado")
	if start < 0 || end < 0 {
		t.Fatal("no se encontro el flujo de reserva legal de facturacion")
	}
	body := src[start : start+end]
	lock := strings.Index(body, "WHERE empresa_id = ?\n\tFOR UPDATE")
	advance := strings.Index(body, "UPDATE empresa_configuracion_avanzada")
	if lock < 0 {
		t.Fatal("la reserva de consecutivo debe bloquear la configuracion de la empresa con FOR UPDATE")
	}
	if advance < 0 || lock > advance {
		t.Fatal("el bloqueo de consecutivo debe ocurrir antes de avanzar proximo_consecutivo")
	}
	if !strings.Contains(body, `strings.EqualFold(strings.TrimSpace(cfg.PaisCodigo), "CO")`) || !strings.Contains(body, `codigoValidacion = ""`) {
		t.Fatal("Colombia must not persist a local hash as if it were an official DIAN CUFE")
	}
}

func TestFacturacionDocumentosMoneyPrecisionMigrationIsRegistered(t *testing.T) {
	migrations, err := PlatformMigrations(MigrationTargetEmpresas)
	if err != nil {
		t.Fatal(err)
	}
	for _, migration := range migrations {
		if migration.Version != "20260821-001-facturacion-documentos-money-v1" {
			continue
		}
		if migration.Body != empresaFacturacionDocumentosMoneyPrecisionFingerprint || migration.Apply == nil {
			t.Fatal("fiscal money migration must be executable and checksummed")
		}
		return
	}
	t.Fatal("fiscal money migration is missing from enterprise catalog")
}

func TestFacturacionDocumentosUseExactMoneyAndFailClosedMigration(t *testing.T) {
	legacy, err := os.ReadFile("documentos_transaccionales.go")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(legacy), "monto_total NUMERIC(18,2) NOT NULL DEFAULT 0") < 2 {
		t.Fatal("new fiscal and purchase document schemas must use exact monetary columns")
	}
	migration, err := os.ReadFile("facturacion_documentos_money_precision.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, required := range []string{
		"NUMERIC(18,2)", "manual reconciliation", "monto_total >= 0", "empresa_facturacion_documentos",
	} {
		if !strings.Contains(string(migration), required) {
			t.Fatalf("fiscal money migration missing %q", required)
		}
	}
}

func TestFacturacionArtefactosMigrationIsRegistered(t *testing.T) {
	migrations, err := PlatformMigrations(MigrationTargetEmpresas)
	if err != nil {
		t.Fatal(err)
	}
	for _, migration := range migrations {
		if migration.Version == "20260821-002-facturacion-artefactos-v1" {
			if migration.Body != empresaFacturacionArtefactosFingerprint || migration.Apply == nil {
				t.Fatal("fiscal artifact migration must be executable and checksummed")
			}
			return
		}
	}
	t.Fatal("fiscal artifact migration is missing from enterprise catalog")
}

func TestFacturacionRetryDueQueryDoesNotRepeatExhaustedDocuments(t *testing.T) {
	raw, err := os.ReadFile("facturacion_electronica.go")
	if err != nil {
		t.Fatal(err)
	}
	source := string(raw)
	for _, required := range []string{
		"COALESCE(fr.intentos, 0) < COALESCE(fr.max_intentos, 5)",
		"CAST(fr.proximo_intento AS TIMESTAMPTZ)",
		"ListFacturacionElectronicaRetryEmpresaIDsDueShardContext",
		"MOD(fr.empresa_id, $2) = $3",
		"pcs_queue_tenant_state",
		"qs.last_served_at ASC NULLS FIRST",
	} {
		if !strings.Contains(source, required) {
			t.Fatalf("durable fiscal retry query missing %q", required)
		}
	}
}

func TestFacturacionListingsCanExcludeSensitiveDocumentFamiliesBeforePagination(t *testing.T) {
	t.Parallel()
	for _, file := range []string{"documentos_transaccionales.go", "facturacion_electronica.go"} {
		raw, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		source := string(raw)
		for _, required := range []string{"ExcluirTiposDocumento []string", "tipo_documento <> ?"} {
			if !strings.Contains(source, required) {
				t.Fatalf("%s does not enforce sensitive-family exclusion at SQL level: missing %q", file, required)
			}
		}
	}
}

func TestFacturacionPayrollEntityNeverJoinsCustomerIdentity(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile("documentos_transaccionales.go")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), "COALESCE(d.tipo_documento, '') NOT IN ('nomina_electronica', 'nota_ajuste_nomina_electronica')") {
		t.Fatal("generic fiscal listing can still interpret a payroll employee id as a customer id")
	}
}
