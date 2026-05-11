package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNuevosVerticalesPermisosDerivadosDelCatalogo(t *testing.T) {
	modules := NuevosVerticalesEmpresaModules()
	if len(modules) != 20 {
		t.Fatalf("NuevosVerticalesEmpresaModules() len = %d, want 20", len(modules))
	}

	seen := map[string]bool{}
	for _, module := range modules {
		if module == "" {
			t.Fatal("module key vacio")
		}
		if seen[module] {
			t.Fatalf("module key duplicado: %s", module)
		}
		seen[module] = true
		if !isPermModuleNuevoVertical(module) {
			t.Fatalf("isPermModuleNuevoVertical(%q) = false, want true", module)
		}
		if page := nuevoVerticalPageKey(module); page == "" || page[:4] != "link" {
			t.Fatalf("page key invalida para %q: %q", module, page)
		}
	}

	if got := nuevoVerticalPageKey("transporte_carga_tms"); got != "linkTransporteCargaTMS" {
		t.Fatalf("nuevoVerticalPageKey(transporte_carga_tms) = %q", got)
	}
	if page, ok := permissionPageForNuevoVerticalAPIPath("/api/empresa/agencia_viajes"); !ok || page != "linkAgenciaViajes" {
		t.Fatalf("permissionPageForNuevoVerticalAPIPath agencia_viajes = %q, %v", page, ok)
	}
}

func TestEmpresaVerticalesNuevosCatalogoContrato(t *testing.T) {
	items := buildEmpresaVerticalesNuevosCatalogo()
	if len(items) != 20 {
		t.Fatalf("catalogo verticales len=%d, want 20", len(items))
	}
	seen := map[string]bool{}
	for _, item := range items {
		if item.ID == "" || item.Page == "" || item.Modulo == "" || item.Titulo == "" {
			t.Fatalf("item incompleto: %+v", item)
		}
		if item.ID != item.Page {
			t.Fatalf("id/page deben coincidir para permisos frontend: %+v", item)
		}
		if seen[item.Modulo] {
			t.Fatalf("modulo duplicado: %s", item.Modulo)
		}
		seen[item.Modulo] = true
		if len(item.Secciones) < 4 || len(item.Plantilla.SeccionesFlujo) < 4 {
			t.Fatalf("secciones incompletas: %+v", item)
		}
		if item.Plantilla.Modulo != item.Modulo {
			t.Fatalf("plantilla modulo=%q want %q", item.Plantilla.Modulo, item.Modulo)
		}
	}
}

func TestPublicVerticalesNuevosCatalogoHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/public/verticales_nuevos/catalogo", nil)
	rr := httptest.NewRecorder()
	PublicVerticalesNuevosCatalogoHandler().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	var payload struct {
		OK    bool                               `json:"ok"`
		Total int                                `json:"total"`
		Items []empresaVerticalNuevoCatalogoItem `json:"items"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&payload); err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	if !payload.OK || payload.Total != 20 || len(payload.Items) != 20 {
		t.Fatalf("payload inesperado: %+v", payload)
	}
}

func TestLinkNuevosVerticalesRequiereAlgunVerticalPermitido(t *testing.T) {
	actionsOff := map[string]bool{
		permActionRead:    false,
		permActionCreate:  false,
		permActionUpdate:  false,
		permActionDelete:  false,
		permActionApprove: false,
	}
	actionsOn := map[string]bool{
		permActionRead:    true,
		permActionCreate:  true,
		permActionUpdate:  false,
		permActionDelete:  false,
		permActionApprove: false,
	}
	rows := []permissionModuleMatrixRow{
		{Modulo: permModuleVentas, Acciones: actionsOn},
		{Modulo: "agencia_viajes", Acciones: actionsOff},
	}
	pages := buildPermissionPagesMapFromModuleRows(rows, nil)
	if pages["linkNuevosVerticales"] {
		t.Fatal("linkNuevosVerticales permitido sin verticales habilitados")
	}

	rows = append(rows, permissionModuleMatrixRow{Modulo: "operador_turistico", Acciones: actionsOn})
	pages = buildPermissionPagesMapFromModuleRows(rows, nil)
	if !pages["linkNuevosVerticales"] {
		t.Fatal("linkNuevosVerticales deberia permitirse con al menos un vertical habilitado")
	}
}
