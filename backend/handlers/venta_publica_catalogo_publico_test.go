package handlers

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVentaPublicaSlugFromRequestCatalogoSoloLecturaPath(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/motel-calipso/visualizar_productos_y_precios_publico.html", nil)

	got := ventaPublicaSlugFromRequest(req)
	if got != "motel-calipso" {
		t.Fatalf("expected slug from read-only catalog path, got %q", got)
	}
}

func TestVentaPublicaSlugFromRequestCatalogoQueryOverridesPath(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/motel-calipso/visualizar_productos_y_precios_publico.html?empresa_slug=Hotel+Principal", nil)

	got := ventaPublicaSlugFromRequest(req)
	if got != "hotel-principal" {
		t.Fatalf("expected normalized query slug, got %q", got)
	}
}

func TestVentaPublicaCatalogoDoesNotTreatStagingHostAsCompany(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "web", "visualizar_productos_y_precios_publico.html"))
	if err != nil {
		t.Fatal(err)
	}
	source := string(raw)
	for _, want := range []string{
		"['www', 'powerfulcontrolsystem', 'staging'].includes(first)",
		"overflow-wrap: anywhere",
		"async function responseErrorMessage(res, fallback)",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("public catalog staging/error contract missing %q", want)
		}
	}
}
