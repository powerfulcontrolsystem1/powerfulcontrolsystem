package db

import (
	"strings"
	"testing"
)

func TestApartTurCoreCodeNormalizaPrefijoYLongitud(t *testing.T) {
	code := apartTurCoreCode("apt-unidad", " Suite 101 ", "Vista ciudad con balcon y jacuzzi")
	if !strings.HasPrefix(code, "APT-UNIDAD-") {
		t.Fatalf("prefijo inesperado: %s", code)
	}
	if strings.ContainsAny(code, " #") {
		t.Fatalf("codigo sin normalizar: %s", code)
	}
	if len(strings.TrimPrefix(code, "APT-UNIDAD-")) > 42 {
		t.Fatalf("codigo demasiado largo: %s", code)
	}
}

func TestApartTurMetodoPagoUsaCatalogoCarrito(t *testing.T) {
	if got := NormalizeMetodoPagoCarrito("tarjeta debito"); got != "tarjeta_debito" {
		t.Fatalf("metodo normalizado = %q", got)
	}
	if got := NormalizeMetodoPagoCarrito("airbnb"); got != "" {
		t.Fatalf("metodo no soportado debe quedar vacio, got %q", got)
	}
}
