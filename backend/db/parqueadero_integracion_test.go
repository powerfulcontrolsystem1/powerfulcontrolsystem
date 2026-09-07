package db

import "testing"

func TestParqueaderoCoreCodeEsEstable(t *testing.T) {
	got := parqueaderoCoreCode("PK-SERV", "Carro ejecutivo")
	if got != "PK-SERV-CARRO-EJECUTIVO" {
		t.Fatalf("parqueaderoCoreCode() = %q", got)
	}
}

func TestParqueaderoMetodoPagoUsaNucleo(t *testing.T) {
	got := NormalizeMetodoPagoCarrito("transferencia")
	if got != "transferencia_bancaria" {
		t.Fatalf("NormalizeMetodoPagoCarrito() = %q", got)
	}
}
