package db

import (
	"context"
	"database/sql"
	"errors"
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

func TestEmpresaCarteraMoneyPrecisionMigrationIsInEnterpriseCatalog(t *testing.T) {
	t.Parallel()
	migrations, err := PlatformMigrations(MigrationTargetEmpresas)
	if err != nil {
		t.Fatal(err)
	}
	for _, migration := range migrations {
		if migration.Version == "20260801-001-cartera-money-precision-v1" {
			if migration.Apply == nil || migration.Body != empresaCarteraMoneyPrecisionFingerprint {
				t.Fatal("cartera money precision migration must be executable and checksummed")
			}
			return
		}
	}
	t.Fatal("cartera money precision migration is missing")
}

func TestEmpresaCarteraMoneyPrecisionUsesExactColumnsAndFailsClosed(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile("cartera_money_precision.go")
	if err != nil {
		t.Fatal(err)
	}
	source := string(raw)
	for _, required := range []string{
		"empresa_cuentas_por_cobrar", "empresa_cuentas_por_pagar",
		"NUMERIC(18,2)", "manual reconciliation",
		"saldo = GREATEST(valor_original - valor_pagado, 0)",
		"saldo = valor_original - valor_pagado",
	} {
		if !strings.Contains(source, required) {
			t.Fatalf("money precision migration missing %q", required)
		}
	}

	legacy, err := os.ReadFile("modulos_faltantes.go")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(legacy), "valor_original NUMERIC(18,2) DEFAULT 0") < 2 ||
		strings.Count(string(legacy), "valor_pagado NUMERIC(18,2) DEFAULT 0") < 2 ||
		strings.Count(string(legacy), "saldo NUMERIC(18,2) DEFAULT 0") < 2 {
		t.Fatal("new CxC/CxP schemas must use exact NUMERIC money columns")
	}
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
		"func RegistrarEmpresaCxPAbonoContext(ctx context.Context",
		"BeginTx(ctx, nil)",
		"queryRowTxSQLCompatContext(ctx, tx",
		"execTxSQLCompatContext(ctx, tx",
		"insertTxSQLCompatContext(ctx, tx",
		"FROM empresa_cuentas_por_pagar WHERE empresa_id = ? AND id = ? FOR UPDATE",
		"FROM empresa_cxp_pagos WHERE empresa_id = ? AND idempotency_key_hash = ?",
		"INSERT INTO empresa_cxp_pagos",
		"InsertOutboxEvent(tx",
		"ErrEmpresaCxPAmountExceedsBalance",
		"EmpresaCxPPaymentOutboxTopic",
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

func TestProcessEmpresaCxPPaymentAccountingRejectsInvalidPayloadBeforeDatabase(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		empresaID int64
		payload   string
	}{
		{0, `{"cuenta_por_pagar_id":1,"pago_id":1,"movimiento_finanzas_id":1,"monto":1}`},
		{12, `{}`},
		{12, `{"cuenta_por_pagar_id":1,"pago_id":1,"movimiento_finanzas_id":1,"monto":-1}`},
		{12, `no-json`},
	} {
		if _, err := ProcessEmpresaCxPPaymentAccounting(context.Background(), &sql.DB{}, test.empresaID, test.payload); err == nil {
			t.Fatalf("empresa=%d payload=%q must fail before database access", test.empresaID, test.payload)
		}
	}
}

func TestProcessEmpresaCxPPaymentAccountingPreservesTenantAndRetryIdempotency(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile("cuentas_por_pagar.go")
	if err != nil {
		t.Fatal(err)
	}
	source := string(raw)
	for _, required := range []string{
		"c.empresa_id = p.empresa_id AND c.id = p.cuenta_por_pagar_id",
		"m.empresa_id = p.empresa_id AND m.id = p.movimiento_finanzas_id",
		"WHERE p.empresa_id = ? AND p.id = ? AND p.cuenta_por_pagar_id = ?",
		"FOR UPDATE OF p",
		"entidad = 'empresa_cxp_pagos' AND entidad_id = ?",
		"'abono_proveedor_registrado'",
	} {
		if !strings.Contains(source, required) {
			t.Fatalf("CxP accounting worker contract missing %q", required)
		}
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

func TestLegacyAccountingCXPRejectsNewWritesBeforeDatabase(t *testing.T) {
	t.Parallel()
	_, err := CreateEmpresaCarteraCXP(&sql.DB{}, EmpresaCarteraCXP{
		EmpresaID:     12,
		Tipo:          "cxp",
		TerceroNombre: "Proveedor de prueba",
		Documento:     "P110-CXP-LEGACY",
	})
	if !errors.Is(err, ErrEmpresaCarteraCXPHistoricaReadOnly) {
		t.Fatalf("legacy CxP write must be rejected before SQL, got %v", err)
	}
	if err := rejectEmpresaCarteraCXPHistorica(" CxP "); !errors.Is(err, ErrEmpresaCarteraCXPHistoricaReadOnly) {
		t.Fatalf("legacy CxP type must remain read-only, got %v", err)
	}
	if err := rejectEmpresaCarteraCXPHistorica("cxc"); err != nil {
		t.Fatalf("CXC history must remain available, got %v", err)
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
