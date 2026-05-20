package db

import "testing"

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
	resp := EmpresaDatafonoProviderResponse{EstadoPago: DatafonoEstadoAprobado, Monto: 50000, Referencia: "VENTA-1"}
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
