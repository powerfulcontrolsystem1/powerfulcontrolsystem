package db

import "testing"

func TestTaxiCoreCodeEsEstable(t *testing.T) {
	got := taxiCoreCode("TX-VIAJE", "12", "TX-260511-120000")
	if got != "TX-VIAJE-12-TX-260511-120000" {
		t.Fatalf("taxiCoreCode() = %q", got)
	}
}

func TestTaxiMetodoPagoUsaNucleo(t *testing.T) {
	got := NormalizeMetodoPagoCarrito("transferencia")
	if got != "transferencia_bancaria" {
		t.Fatalf("NormalizeMetodoPagoCarrito() = %q", got)
	}
}
