package handlers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCarritoPaymentButtonHasDirectAndDelegatedHandler(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "web", "administrar_empresa", "carrito_de_compras.html"))
	if err != nil {
		t.Fatalf("read carrito frontend: %v", err)
	}
	source := string(raw)
	for _, required := range []string{
		"onclick=\"handlePagarCarritoClick(event)\"",
		"document.addEventListener('click'",
		"btnPagarCarrito.addEventListener('click', handlePagarCarritoClick)",
		"btn.dataset.paymentClickLock === '1'",
	} {
		if !strings.Contains(source, required) {
			t.Fatalf("carrito payment contract must keep %q", required)
		}
	}
}
