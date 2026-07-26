package db

import (
	"database/sql"
	"math"
	"os"
	"strings"
	"testing"
)

func TestEmpresaCxPAtomicSchemaHasTenantIdempotencyAndLedger(t *testing.T) {
	t.Parallel()
	statements := strings.Join(empresaCxPAtomicSchemaStatements(), "\n")
	for _, required := range []string{
		"empresa_cxp_pagos",
		"empresa_id BIGINT NOT NULL",
		"idempotency_key_hash TEXT NOT NULL",
		"UNIQUE (empresa_id, idempotency_key_hash)",
		"movimiento_finanzas_id BIGINT NOT NULL",
		"monto NUMERIC(18,2) NOT NULL",
	} {
		if !strings.Contains(statements, required) {
			t.Fatalf("CxP atomic schema must contain %q", required)
		}
	}
}

func TestEmpresaCxPStatusAndDaysDoNotCreateNegativeBalances(t *testing.T) {
	t.Parallel()
	if got := empresaCxPEstado(0, "2020-01-01"); got != "pagada" {
		t.Fatalf("estado saldo cero = %q, want pagada", got)
	}
	if got := empresaCxPDiasMora("2099-01-01", 100); got != 0 {
		t.Fatalf("dias mora futuro = %d, want 0", got)
	}
	if got := empresaCxPDiasMora("fecha-invalida", 100); got != 0 {
		t.Fatalf("dias mora invalida = %d, want 0", got)
	}
}

func TestEmpresaCxPMigrationIsInEnterpriseCatalog(t *testing.T) {
	t.Parallel()
	migrations, err := PlatformMigrations(MigrationTargetEmpresas)
	if err != nil {
		t.Fatal(err)
	}
	for _, migration := range migrations {
		if migration.Version == "20260724-001-cxp-atomic-payments-v1" {
			if migration.Body != empresaCxPAtomicSchemaFingerprint || migration.Apply == nil {
				t.Fatal("CxP migration must be executable and checksummed")
			}
			return
		}
	}
	t.Fatal("CxP migration is missing from enterprise catalog")
}

func TestNombreProveedorCxPCanonicoUsesRegisteredSupplierData(t *testing.T) {
	t.Parallel()
	if got := nombreProveedorCxPCanonico("  Proveedor Legal S.A.S.  ", "Nombre IA no confiable"); got != "Proveedor Legal S.A.S." {
		t.Fatalf("razon social canonical = %q", got)
	}
	if got := nombreProveedorCxPCanonico("", "  Comercial PCS  "); got != "Comercial PCS" {
		t.Fatalf("nombre comercial fallback = %q", got)
	}
}

func TestEmpresaCxPIdempotencyHashDoesNotExposeRawKey(t *testing.T) {
	t.Parallel()
	key := "p106-cxp-payment-key-123"
	hash := empresaCxPIdempotencyHash(key)
	if hash == "" || hash == key || strings.Contains(hash, key) {
		t.Fatalf("idempotency key must be protected, got %q", hash)
	}
	if hash != empresaCxPIdempotencyHash(key) || hash == empresaCxPIdempotencyHash(key+"-other") {
		t.Fatal("idempotency hash must be deterministic and distinct")
	}
}

func TestRegistrarEmpresaCxPAbonoKeepsTenantScopedAtomicInvariants(t *testing.T) {
	raw, err := os.ReadFile("cuentas_por_pagar.go")
	if err != nil {
		t.Fatal(err)
	}
	body := string(raw)
	for _, required := range []string{
		"FROM empresa_cuentas_por_pagar WHERE empresa_id = ? AND id = ? FOR UPDATE",
		"FROM empresa_cxp_pagos WHERE empresa_id = ? AND idempotency_key_hash = ?",
		"INSERT INTO empresa_cxp_pagos",
		"InsertOutboxEvent(tx",
		"ErrEmpresaCxPAmountExceedsBalance",
	} {
		if !strings.Contains(body, required) {
			t.Fatalf("CxP atomic flow must preserve %q", required)
		}
	}
	accountLock := strings.Index(body, "FROM empresa_cuentas_por_pagar WHERE empresa_id = ? AND id = ? FOR UPDATE")
	if accountLock < 0 {
		t.Fatal("CxP flow must lock the canonical account")
	}
	if strings.Count(body[accountLock:], "FROM empresa_cxp_pagos WHERE empresa_id = ? AND idempotency_key_hash = ?") == 0 {
		t.Fatal("CxP flow must recheck idempotency after the account lock for concurrent retries")
	}
}

func TestEmpresaSoportesComprasIACriticalTablesCoverConversionDependencies(t *testing.T) {
	t.Parallel()
	tables := strings.Join(empresaSoportesComprasIATablasCriticas(), "\n")
	for _, required := range []string{
		"empresa_soportes_compras_ia",
		"empresa_soportes_compras_ia_eventos",
		"empresa_cuentas_por_pagar",
		"empresa_proveedores",
		"pcs_outbox_events",
	} {
		if !strings.Contains(tables, required) {
			t.Fatalf("critical support-to-CxP schema must include %q", required)
		}
	}
}

func TestCxPSupplierCatalogIsTenantFilteredAndActiveOnly(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile("compras_y_proveedores.go")
	if err != nil {
		t.Fatal(err)
	}
	body := string(raw)
	for _, required := range []string{
		"func ListEmpresaProveedoresCxP",
		"FROM proveedores WHERE empresa_id=? AND COALESCE(estado,'activo')='activo'",
		"func GetEmpresaProveedorCxP",
		"FROM proveedores WHERE empresa_id=? AND id=?",
	} {
		if !strings.Contains(body, required) {
			t.Fatalf("CxP supplier catalog must preserve %q", required)
		}
	}
}

func TestNormalizeEmpresaCxPMetodoPago(t *testing.T) {
	t.Parallel()
	if got, err := normalizeEmpresaCxPMetodoPago("  Transferencia "); err != nil || got != "transferencia" {
		t.Fatalf("normalized method = %q, err=%v", got, err)
	}
	if got, err := normalizeEmpresaCxPMetodoPago(""); err != nil || got != "efectivo" {
		t.Fatalf("default method = %q, err=%v", got, err)
	}
	if _, err := normalizeEmpresaCxPMetodoPago("efectivo\ninyeccion"); err == nil {
		t.Fatal("method with control characters must be rejected")
	}
}

func TestRegistrarEmpresaCxPAbonoRejectsNonFiniteAmountBeforeDatabase(t *testing.T) {
	t.Parallel()
	for _, amount := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		_, err := RegistrarEmpresaCxPAbono(&sql.DB{}, EmpresaCxPAbonoInput{EmpresaID: 12, CuentaPorPagarID: 7, Monto: amount, IdempotencyKey: "p106-nonfinite"})
		if err == nil || !strings.Contains(err.Error(), "monto del abono") {
			t.Fatalf("amount %v must fail before database, err=%v", amount, err)
		}
	}
}
