package handlers

import (
	"os"
	"strings"
	"testing"
)

func TestPaymentCheckoutRequiresDurableIdempotency(t *testing.T) {
	raw, err := os.ReadFile("payments_handlers.go")
	if err != nil {
		t.Fatal(err)
	}
	source := string(raw)
	for _, required := range []string{
		"beginPaymentCheckoutIdempotency",
		"Idempotency-Key valido es obligatorio",
		"CompletePaymentCheckoutIdempotency",
		"MarkPaymentCheckoutIdempotencyUncertain",
		"AbandonPaymentCheckoutIdempotency",
	} {
		if !strings.Contains(source, required) {
			t.Errorf("falta contrato de checkout idempotente %q", required)
		}
	}
	if strings.Contains(source, `fmt.Sprintf("WOMPI-LIC-%d-EMP-%d-%d", payload.LicenciaID, payload.EmpresaID, time.Now().UnixNano())`) ||
		strings.Contains(source, `fmt.Sprintf("EPAYCO-LIC-%d-EMP-%d-%d", payload.LicenciaID, payload.EmpresaID, time.Now().UnixNano())`) {
		t.Fatal("el checkout aun genera referencias nuevas en cada retry")
	}
}
