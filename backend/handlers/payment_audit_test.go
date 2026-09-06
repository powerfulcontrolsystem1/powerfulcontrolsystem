package handlers

import (
	"net/http/httptest"
	"testing"
)

func TestParseSuperPaymentAuditFilters(t *testing.T) {
	r := httptest.NewRequest("GET", "/super/api/pagos/auditoria?provider=epayco&status=approved&empresa_id=42&q=INV-1&limit=100&offset=20", nil)
	filters, validationError := parseSuperPaymentAuditFilters(r)
	if validationError != "" {
		t.Fatal(validationError)
	}
	if filters.Provider != "epayco" || filters.Status != "approved" || filters.EmpresaID != 42 || filters.Search != "INV-1" || filters.Limit != 100 || filters.Offset != 20 {
		t.Fatalf("unexpected filters: %#v", filters)
	}
}

func TestParseSuperPaymentAuditFiltersRejectsUnboundedInputs(t *testing.T) {
	for _, rawURL := range []string{
		"/super/api/pagos/auditoria?provider=otro",
		"/super/api/pagos/auditoria?empresa_id=-1",
		"/super/api/pagos/auditoria?limit=201",
		"/super/api/pagos/auditoria?offset=-1",
	} {
		r := httptest.NewRequest("GET", rawURL, nil)
		if _, validationError := parseSuperPaymentAuditFilters(r); validationError == "" {
			t.Fatalf("expected validation error for %s", rawURL)
		}
	}
}
