package db

import (
	"strings"
	"testing"
)

func TestNormalizeDatafonoProviderAndEstado(t *testing.T) {
	if got := NormalizeDatafonoProvider("Crediban-Co"); got != DatafonoProviderCredibanco {
		t.Fatalf("provider = %q, want %q", got, DatafonoProviderCredibanco)
	}
	if got := NormalizeDatafonoEstadoPago("PAID"); got != DatafonoEstadoAprobado {
		t.Fatalf("estado = %q, want %q", got, DatafonoEstadoAprobado)
	}
	if got := NormalizeDatafonoEstadoPago("declined"); got != DatafonoEstadoRechazado {
		t.Fatalf("estado = %q, want %q", got, DatafonoEstadoRechazado)
	}
}

func TestValidateDatafonoAmountAndReference(t *testing.T) {
	req := EmpresaDatafonoPaymentRequest{Monto: 50000, Referencia: "VENTA-1"}
	resp := EmpresaDatafonoProviderResponse{ProviderTransactionID: "TX-1", EstadoPago: DatafonoEstadoAprobado, Monto: 50000, Moneda: "COP", Referencia: "VENTA-1"}
	if err := ValidateDatafonoAmountAndReference(req, resp); err != nil {
		t.Fatalf("expected valid response, got %v", err)
	}

	resp.Monto = 49900
	if err := ValidateDatafonoAmountAndReference(req, resp); err == nil {
		t.Fatalf("expected amount mismatch")
	}

	resp.Monto = 50000
	resp.Referencia = "VENTA-2"
	if err := ValidateDatafonoAmountAndReference(req, resp); err == nil {
		t.Fatalf("expected reference mismatch")
	}
}

func TestDatafonoRequestJSONDoesNotPersistCustomerPII(t *testing.T) {
	raw := DatafonoRequestJSON(EmpresaDatafonoPaymentRequest{EmpresaID: 1, Referencia: "VENTA-1", Cliente: EmpresaDatafonoCliente{Documento: "123456789", Email: "cliente@example.com", Telefono: "3001234567"}})
	for _, forbidden := range []string{"123456789", "cliente@example.com", "3001234567"} {
		if strings.Contains(raw, forbidden) {
			t.Fatalf("request audit contains customer PII %q: %s", forbidden, raw)
		}
	}
}
