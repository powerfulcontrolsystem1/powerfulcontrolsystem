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

func TestCarritoInitialLoadShowsUIAndParallelizesReadOnlyData(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "web", "administrar_empresa", "carrito_de_compras.html"))
	if err != nil {
		t.Fatalf("read carrito frontend: %v", err)
	}
	source := string(raw)
	for _, required := range []string{
		"async function loadInitialCartData()",
		"setInitialLoadContentVisible(true);",
		"await Promise.all([",
		"loadEmpresaGeneralConfigFromServer(),",
		"loadEmpresaPermissionsContextFromServer(),",
		"loadStationConfigFromServer(),",
		"PCSEstacionLabels.loadPrefs",
		"loadClientes(false),",
		"loadOperativaConfig(false),",
		"loadTipConfig(false),",
		"loadCommissionConfig(false),",
		"loadUsuariosEmpresaForComisiones(false),",
		"loadActiveCashRegisters(false),",
		"loadCarritos(false, false)",
		"renderUsuariosEmpresaForComisiones();",
		"fillClientesSelect('');",
		"await loadInitialCartData();",
		"state.initialDataLoading = false;",
		"initialDataLoading ? 'Terminando de cargar los datos del carrito.'",
		"const hasPayableContent = cartHasItemsOrTotalsForCancel(carrito);",
		"Agrega al menos un producto o servicio antes de pagar.",
	} {
		if !strings.Contains(source, required) {
			t.Fatalf("carrito fast initial-load contract must keep %q", required)
		}
	}

	for _, required := range []string{
		"async function loadActiveCashRegisters(renderAfterLoad)",
		"async function loadOperativaConfig(renderAfterLoad)",
		"async function loadTipConfig(renderAfterLoad)",
		"async function loadCommissionConfig(renderAfterLoad)",
		"async function loadUsuariosEmpresaForComisiones(renderAfterLoad)",
		"async function loadClientes(renderAfterLoad)",
		"async function loadCarritos(renderAfterLoad, filterBySelectedCode)",
		"if (filterBySelectedCode !== false && state.selectedCarritoCode)",
		"if (renderAfterLoad !== false)",
	} {
		if !strings.Contains(source, required) {
			t.Fatalf("carrito initial load must avoid duplicate partial rendering; missing %q", required)
		}
	}

	if strings.Count(source, "fetch('/me'") != 1 {
		t.Fatal("carrito initial load must share one /me request between title and operational-role loaders")
	}

	labelsRaw, err := os.ReadFile(filepath.Join("..", "..", "web", "js", "estaciones_labels.js"))
	if err != nil {
		t.Fatalf("read station labels frontend: %v", err)
	}
	for _, required := range []string{"var prefsLoads", "async function loadPrefs", "loadPrefs: loadPrefs"} {
		if !strings.Contains(string(labelsRaw), required) {
			t.Fatalf("station labels must share prefs request with the carrito; missing %q", required)
		}
	}
}
