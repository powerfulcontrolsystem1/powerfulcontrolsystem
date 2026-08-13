package handlers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCarritoPaymentButtonHasExplicitAndDelegatedHandlerWithoutInlineEvent(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "web", "administrar_empresa", "carrito_de_compras.html"))
	if err != nil {
		t.Fatalf("read carrito frontend: %v", err)
	}
	source := string(raw)
	if strings.Contains(source, `onclick="handlePagarCarritoClick(event)"`) {
		t.Fatal("carrito payment button must not restore an inline click event")
	}
	for _, required := range []string{
		`id="btnPagarCarrito"`,
		"document.addEventListener('click'",
		"btnPagarCarrito.addEventListener('click', handlePagarCarritoClick)",
		"btn.dataset.paymentClickLock === '1'",
	} {
		if !strings.Contains(source, required) {
			t.Fatalf("carrito payment contract must keep %q", required)
		}
	}
}
