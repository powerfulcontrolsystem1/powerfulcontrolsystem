package db

import "testing"

func TestPropHCoreCodeSanitizesStablePrefix(t *testing.T) {
	got := propHCoreCode("ph cargo", "Cuota Administraci\u00f3n", "Torre A-101")
	want := "PH-CARGO-CUOTA-ADMINISTRACION-TORRE-A-101"
	if got != want {
		t.Fatalf("codigo normalizado = %q, want %q", got, want)
	}
}

func TestPropHMetodoPagoCarritoMapsOperationalMethods(t *testing.T) {
	cases := map[string]string{
		"efectivo":       "efectivo",
		"tarjeta debito": "tarjeta_debito",
		"transferencia":  "transferencia_bancaria",
		"pse":            "transferencia_bancaria",
		"consignacion":   "transferencia_bancaria",
		"otro":           "transferencia_bancaria",
		"cripto":         "efectivo",
	}
	for in, want := range cases {
		if got := propHMetodoPagoCarrito(in); got != want {
			t.Fatalf("metodo %q = %q, want %q", in, got, want)
		}
	}
}
