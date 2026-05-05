package handlers

import (
	"net/http"
	"net/http/httptest"
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
