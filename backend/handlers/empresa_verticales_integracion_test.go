package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEmpresaVerticalesIntegracionCatalogoContrato(t *testing.T) {
	items := buildEmpresaVerticalesIntegracionCatalogo()
	if len(items) < 12 {
		t.Fatalf("catalogo clasico incompleto: %d", len(items))
	}
	seen := map[string]bool{}
	required := map[string]bool{
		"gimnasio":                false,
		"odontologia":             false,
		"parqueadero":             false,
		"taxi_system":             false,
		"domicilios":              false,
		"apartamentos_turisticos": false,
		"propiedad_horizontal":    false,
		"alquileres":              false,
		"drogueria_farmacia":      false,
		"aiu_construccion":        false,
	}
	for _, item := range items {
		if item.ID == "" || item.Page == "" || item.Modulo == "" || item.Titulo == "" {
			t.Fatalf("item incompleto: %+v", item)
		}
		if seen[item.Modulo] {
			t.Fatalf("modulo duplicado: %s", item.Modulo)
		}
		seen[item.Modulo] = true
		if _, ok := required[item.Modulo]; ok {
			required[item.Modulo] = true
		}
		if item.OperationalVisible && len(item.DuplicatesCore) != 0 {
			t.Fatalf("vertical visible con duplicados del nucleo: %+v", item)
		}
		if item.OperationalVisible && item.IntegrationStatus != "plantilla_integrada_nucleo" && item.IntegrationStatus != "integrado_soporte" {
			t.Fatalf("estado visible invalido para %s: %s", item.Modulo, item.IntegrationStatus)
		}
		if item.AliasDe == "" && len(item.OwnFlowAllowed) == 0 {
			t.Fatalf("vertical sin flujo especializado declarado: %+v", item)
		}
		if item.OperationalVisible && len(item.TemplateActivates) == 0 {
			t.Fatalf("vertical visible sin plantilla declarada: %+v", item)
		}
		if item.OperationalVisible && len(item.TablesTouched) == 0 {
			t.Fatalf("vertical visible sin tablas declaradas: %+v", item)
		}
		if item.OperationalVisible && len(item.RequiredPermissions) == 0 {
			t.Fatalf("vertical visible sin permisos declarados: %+v", item)
		}
		if item.OperationalVisible && len(item.SaleFlow) == 0 {
			t.Fatalf("vertical visible sin flujo de venta declarado: %+v", item)
		}
		if item.OperationalVisible && len(item.ReportsProduced) == 0 {
			t.Fatalf("vertical visible sin reportes declarados: %+v", item)
		}
		if item.SyncAction != "" && (item.SyncPath == "" || item.SyncActionName == "") {
			t.Fatalf("sincronizacion sin contrato estructurado: %+v", item)
		}
	}
	for module, ok := range required {
		if !ok {
			t.Fatalf("vertical requerido faltante en catalogo: %s", module)
		}
	}
}

func TestPublicVerticalesIntegracionCatalogoHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/public/verticales_integracion/catalogo", nil)
	rr := httptest.NewRecorder()
	PublicVerticalesIntegracionCatalogoHandler().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	var payload struct {
		OK    bool                             `json:"ok"`
		Total int                              `json:"total"`
		Items []empresaVerticalIntegracionItem `json:"items"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&payload); err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	if !payload.OK || payload.Total != len(payload.Items) || payload.Total < 12 {
		t.Fatalf("payload inesperado: %+v", payload)
	}
}
