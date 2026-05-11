package db

import "testing"

func TestOdontoCoreCodeEsEstableParaServicios(t *testing.T) {
	got := odontoCoreCode("OD-TRAT", "15", "Ortodoncia estetica")
	if got != "OD-TRAT-15-ORTODONCIA-ESTETICA" {
		t.Fatalf("odontoCoreCode() = %q", got)
	}
	if len(got) > 50 {
		t.Fatalf("codigo demasiado largo: %q", got)
	}
}

func TestOdontoPagoUsaMetodoPagoCentral(t *testing.T) {
	got := NormalizeMetodoPagoCarrito("transferencia")
	if got != "transferencia_bancaria" {
		t.Fatalf("NormalizeMetodoPagoCarrito() = %q", got)
	}
}
