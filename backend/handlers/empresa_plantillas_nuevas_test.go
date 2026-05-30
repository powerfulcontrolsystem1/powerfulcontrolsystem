package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNuevasPlantillasPermisosDerivadosDelCatalogo(t *testing.T) {
	modules := NuevasPlantillasEmpresaModules()
	if len(modules) != 20 {
		t.Fatalf("NuevasPlantillasEmpresaModules() len = %d, want 20", len(modules))
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

func TestEmpresaPlantillasNuevosCatalogoContrato(t *testing.T) {
	items := buildEmpresaPlantillasNuevosCatalogo()
	if len(items) != 20 {
		t.Fatalf("catalogo plantillas len=%d, want 20", len(items))
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
		if item.IntegrationStatus != "plantilla_integrada_nucleo" || !item.OperationalVisible {
			t.Fatalf("estado de integracion invalido para %q: %+v", item.Modulo, item)
		}
		if len(item.CoreModules) < 7 || len(item.DuplicatesCore) != 0 {
			t.Fatalf("contrato de nucleo invalido para %q: core=%v duplicados=%v", item.Modulo, item.CoreModules, item.DuplicatesCore)
		}
		if item.IntegracionPreconfig == nil {
			t.Fatalf("item sin integracion_preconfig: %s", item.Modulo)
		}
		if item.IntegracionPreconfig.Modulo != item.Modulo {
			t.Fatalf("integracion modulo=%q want %q", item.IntegracionPreconfig.Modulo, item.Modulo)
		}
		if item.DecisionPreconfig != item.IntegracionPreconfig.Decision {
			t.Fatalf("decision catalogo=%q integracion=%q", item.DecisionPreconfig, item.IntegracionPreconfig.Decision)
		}
		if len(item.IntegracionPreconfig.TemplateActivates) == 0 ||
			len(item.IntegracionPreconfig.TablesTouched) == 0 ||
			len(item.IntegracionPreconfig.RequiredPermissions) == 0 ||
			len(item.IntegracionPreconfig.SaleFlow) == 0 ||
			len(item.IntegracionPreconfig.ReportsProduced) == 0 {
			t.Fatalf("integracion extendida incompleta para %s: %+v", item.Modulo, item.IntegracionPreconfig)
		}
	}
}

func TestEmpresaPlantillasNuevosCatalogoProduccionMasiva(t *testing.T) {
	items := buildEmpresaPlantillasNuevosCatalogo()
	masivos := 0
	for _, item := range items {
		if item.ProduccionMasiva {
			masivos++
			if item.PrioridadProduccion < 1 || item.PrioridadProduccion > 20 {
				t.Fatalf("prioridad masiva invalida para %s: %d", item.Modulo, item.PrioridadProduccion)
			}
			if item.DecisionPreconfig != "integrar_v1_produccion_masiva" {
				t.Fatalf("masivo %s decision=%q", item.Modulo, item.DecisionPreconfig)
			}
			continue
		}
		t.Fatalf("%s debe estar marcado como produccion masiva", item.Modulo)
	}
	if masivos != 20 {
		t.Fatalf("plantillas masivos=%d, want 20", masivos)
	}
}

func TestPublicPlantillasNuevosCatalogoHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/public/plantillas_nuevas/catalogo", nil)
	rr := httptest.NewRecorder()
	PublicPlantillasNuevosCatalogoHandler().ServeHTTP(rr, req)
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

func TestSuperPlantillasNuevosCatalogoPostRequiereDB(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/super/api/plantillas_nuevas/catalogo?action=asegurar_v1_licencias", nil)
	rr := httptest.NewRecorder()
	SuperPlantillasNuevosCatalogoHandler().ServeHTTP(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
}

func TestSuperPlantillasNuevosCatalogoPostAccionInvalida(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/super/api/plantillas_nuevas/catalogo?action=borrar", nil)
	rr := httptest.NewRecorder()
	SuperPlantillasNuevosCatalogoHandler().ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
}

func TestLinkNuevasPlantillasRequiereAlgunVerticalPermitido(t *testing.T) {
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
	if pages["linkNuevasPlantillas"] {
		t.Fatal("linkNuevasPlantillas permitido sin plantillas habilitados")
	}

	rows = append(rows, permissionModuleMatrixRow{Modulo: "operador_turistico", Acciones: actionsOn})
	pages = buildPermissionPagesMapFromModuleRows(rows, nil)
	if !pages["linkNuevasPlantillas"] {
		t.Fatal("linkNuevasPlantillas deberia permitirse con al menos una plantilla habilitado")
	}
}
