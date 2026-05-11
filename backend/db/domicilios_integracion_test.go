package db

import (
	"strings"
	"testing"
)

func TestDomicilioCoreCodeNormalizaPrefijoYLongitud(t *testing.T) {
	code := domicilioCoreCode("dom-menu", " Restaurante Norte ", "Hamburguesa especial #1 con extra queso y salsa")
	if !strings.HasPrefix(code, "DOM-MENU-") {
		t.Fatalf("prefijo inesperado: %s", code)
	}
	if strings.ContainsAny(code, " #") {
		t.Fatalf("codigo sin normalizar: %s", code)
	}
	if len(strings.TrimPrefix(code, "DOM-MENU-")) > 42 {
		t.Fatalf("codigo demasiado largo: %s", code)
	}
}

func TestDomicilioMetodoPagoUsaCatalogoCarrito(t *testing.T) {
	if got := NormalizeMetodoPagoCarrito("transferencia bancaria"); got != "transferencia_bancaria" {
		t.Fatalf("metodo normalizado = %q", got)
	}
	if got := NormalizeMetodoPagoCarrito("contra entrega"); got != "" {
		t.Fatalf("metodo no soportado debe quedar vacio, got %q", got)
	}
}
