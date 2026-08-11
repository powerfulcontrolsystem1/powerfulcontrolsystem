package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	dbpkg "github.com/you/pos-backend/db"
)

func TestRegistrarPagoCxPRequiresIdempotencyKeyBeforeDatabaseAccess(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodPost, "/api/empresa/finanzas/cuentas_pagar?action=registrar_pago&empresa_id=12", strings.NewReader(`{"empresa_id":12,"id":7,"monto":100}`))
	recorder := httptest.NewRecorder()
	handleRegistrarPagoCarteraAction(nil, cfgCxP, "egreso", "proveedor_nombre", "cuentas_pagar", recorder, req)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body=%s", recorder.Code, http.StatusBadRequest, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "Idempotency-Key") {
		t.Fatalf("response must explain the idempotency requirement: %s", recorder.Body.String())
	}
}

func TestRegistrarPagoCxPRequiresEmpresaIDBeforeDatabaseAccess(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodPost, "/api/empresa/finanzas/cuentas_pagar?action=registrar_pago", strings.NewReader(`{"id":7,"monto":100}`))
	req.Header.Set("Idempotency-Key", "p106-tenant-check")
	recorder := httptest.NewRecorder()
	handleRegistrarPagoCarteraAction(nil, cfgCxP, "egreso", "proveedor_nombre", "cuentas_pagar", recorder, req)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body=%s", recorder.Code, http.StatusBadRequest, recorder.Body.String())
	}
	if !strings.Contains(strings.ToLower(recorder.Body.String()), "empresa") {
		t.Fatalf("response must explain the missing tenant context: %s", recorder.Body.String())
	}
}

func TestValidateEmpresaCxPProveedorPayloadRejectsMissingSupplierBeforeDatabase(t *testing.T) {
	t.Parallel()
	err := validateEmpresaCxPProveedorPayload(nil, 12, map[string]interface{}{}, true)
	if err == nil || !strings.Contains(err.Error(), "proveedor") {
		t.Fatalf("missing CxP supplier must be rejected before database access, err=%v", err)
	}
}

func TestCxPConfigurationRequiresCanonicalRegisteredSupplier(t *testing.T) {
	t.Parallel()
	if cfgCxP.ValidatePayload == nil {
		t.Fatal("CxP must validate its supplier payload")
	}
	for _, required := range []string{"proveedor_id", "documento_codigo"} {
		found := false
		for _, column := range cfgCxP.RequiredOnCreate {
			if column == required {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("CxP creation must require %q", required)
		}
	}
	for _, required := range cfgCxP.RequiredOnCreate {
		if required == "proveedor_nombre" {
			t.Fatal("supplier name must be derived from the tenant-scoped proveedor_id, not supplied by the client")
		}
	}
}

func TestCxPConcurrentNoBalanceIsReportedAsConflict(t *testing.T) {
	for _, err := range []error{
		dbpkg.ErrEmpresaCxPNoPendingBalance,
		dbpkg.ErrEmpresaCxPAmountExceedsBalance,
		dbpkg.ErrPeriodoFinancieroCerrado,
	} {
		if got := registrarPagoCxPErrorStatus(err); got != http.StatusConflict {
			t.Fatalf("status for %v = %d, want %d", err, got, http.StatusConflict)
		}
	}
	if got := registrarPagoCxPErrorStatus(http.ErrBodyNotAllowed); got != http.StatusBadRequest {
		t.Fatalf("generic error status = %d, want %d", got, http.StatusBadRequest)
	}
}
