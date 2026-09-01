package db

import (
	"os"
	"strings"
	"testing"
)

func TestVentaPublicaPaymentUpdateIsMonotonic(t *testing.T) {
	raw, err := os.ReadFile("venta_publica.go")
	if err != nil {
		t.Fatal(err)
	}
	source := string(raw)
	for _, required := range []string{
		"WHEN LOWER(COALESCE(estado_pago, 'pendiente')) = 'aprobado' THEN estado_pago",
		"WHEN LOWER(?) = 'aprobado' THEN ?",
		"= 'rechazado' AND LOWER(?) = 'pendiente' THEN estado_pago",
	} {
		if !strings.Contains(source, required) {
			t.Errorf("falta transicion monotona de pago publico %q", required)
		}
	}
}
