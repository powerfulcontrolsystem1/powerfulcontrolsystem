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

func TestDirectSaleInitialLoadIncludesLegacyCartBeforeCreatingCanonicalCart(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "web", "administrar_empresa", "carrito_de_compras.html"))
	if err != nil {
		t.Fatalf("read carrito frontend: %v", err)
	}
	source := string(raw)
	start := strings.Index(source, "async function loadCarritos(")
	if start < 0 {
		t.Fatal("loadCarritos frontend contract is missing")
	}
	directEnd := strings.Index(source[start:], "} else if (state.stationMode)")
	if directEnd < 0 {
		t.Fatal("direct sale load branch is missing")
	}
	directBranch := source[start : start+directEnd]
	if strings.Contains(directBranch, "url.searchParams.set('carrito_codigo'") {
		t.Fatal("direct sale initial load must not hide the legacy cart behind the canonical code filter")
	}
	if !strings.Contains(source, "getLegacyDirectSaleCarritoCode()") {
		t.Fatal("direct sale lifecycle must preserve the legacy cart reconciliation path")
	}
}
