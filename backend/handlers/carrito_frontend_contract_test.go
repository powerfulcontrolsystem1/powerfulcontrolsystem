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

func TestCarritoStationSearchNavigationAndGlobalVIPRetirementContract(t *testing.T) {
	cartPath := filepath.Join("..", "..", "web", "administrar_empresa", "carrito_de_compras.html")
	raw, err := os.ReadFile(cartPath)
	if err != nil {
		t.Fatalf("read carrito frontend: %v", err)
	}
	source := string(raw)
	for _, required := range []string{
		`id="backToStationsHeaderBtn" class="btn secondary" type="button" hidden>Regresar</button>`,
		`aria-label="Abrir en pantalla completa"`,
		`const canPrepareFocusedSale = focusedSaleMode && state.empresaID > 0;`,
		`const canUse = !!selected || canPrepareFocusedSale;`,
		`const showVipCard = false;`,
		`mostrar_tarjeta_vip_cliente: false,`,
	} {
		if !strings.Contains(source, required) {
			t.Fatalf("carrito operational contract must keep %q", required)
		}
	}
	if strings.Contains(source, `<span class="carrito-fullscreen-label">`) {
		t.Fatal("fullscreen control must render only its icon")
	}

	for _, page := range []string{
		filepath.Join("..", "..", "web", "administrar_empresa", "configuracion_carrito_de_compra_empresa.html"),
		filepath.Join("..", "..", "web", "administrar_empresa", "configuracion_de_estaciones.html"),
	} {
		pageRaw, readErr := os.ReadFile(page)
		if readErr != nil {
			t.Fatalf("read %s: %v", page, readErr)
		}
		pageSource := string(pageRaw)
		if !strings.Contains(pageSource, `id="carritoCfgTarjetaVip" type="checkbox" disabled`) ||
			!strings.Contains(pageSource, `mostrar_tarjeta_vip_cliente: false,`) {
			t.Fatalf("%s must keep VIP cart access globally disabled", page)
		}
	}
}
