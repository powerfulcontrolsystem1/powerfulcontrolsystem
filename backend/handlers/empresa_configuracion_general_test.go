package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	dbpkg "github.com/you/pos-backend/db"
)

func TestEmpresaConfiguracionGeneralHandlerGetAndSave(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_configuracion_general_handler.db")
	if err := dbpkg.EnsureEmpresaConfiguracionGeneralSchema(dbEmp); err != nil {
		t.Fatalf("ensure configuracion general schema: %v", err)
	}

	h := EmpresaConfiguracionGeneralHandler(dbEmp)

	getReq := httptest.NewRequest(http.MethodGet, "/api/empresa/configuracion_general?empresa_id=17", nil)
	getRR := httptest.NewRecorder()
	h.ServeHTTP(getRR, getReq)
	if getRR.Code != http.StatusOK {
		t.Fatalf("expected status %d for initial get, got %d body=%s", http.StatusOK, getRR.Code, getRR.Body.String())
	}

	var initial struct {
		EmpresaID int64 `json:"empresa_id"`
		Productos struct {
			CopiasOrdenServicio          int64 `json:"copias_orden_servicio"`
			DescuentosHabilitados        bool  `json:"descuentos_habilitados"`
			LectorCodigoBarrasHabilitado bool  `json:"lector_codigo_barras_habilitado"`
		} `json:"productos"`
	}
	if err := json.Unmarshal(getRR.Body.Bytes(), &initial); err != nil {
		t.Fatalf("decode initial get response: %v", err)
	}
	if initial.EmpresaID != 17 || initial.Productos.CopiasOrdenServicio != 1 || !initial.Productos.DescuentosHabilitados || !initial.Productos.LectorCodigoBarrasHabilitado {
		t.Fatalf("expected default config for empresa 17, got %+v", initial)
	}

	body := `{"empresa_id":17,"productos":{"imprimir_orden_servicio":true,"area_despacho":"Cocina fria","copias_orden_servicio":3,"nota_orden_servicio":"Entregar con prioridad","descuentos_habilitados":true,"permitir_descuento_porcentaje":true,"permitir_descuento_codigo":false,"permitir_descuento_valor":true,"codigos_descuento":"VIP10=10%","lector_codigo_barras_habilitado":true,"lector_codigo_barras_autofoco":false,"lector_codigo_barras_acumular":true}}`
	putReq := httptest.NewRequest(http.MethodPut, "/api/empresa/configuracion_general?empresa_id=17", strings.NewReader(body))
	putReq.Header.Set("Content-Type", "application/json")
	putReq.Header.Set("X-Admin-Email", "qa@example.com")
	putRR := httptest.NewRecorder()
	h.ServeHTTP(putRR, putReq)
	if putRR.Code != http.StatusOK {
		t.Fatalf("expected status %d for put, got %d body=%s", http.StatusOK, putRR.Code, putRR.Body.String())
	}

	var putResp struct {
		OK            bool `json:"ok"`
		Configuracion struct {
			EmpresaID int64 `json:"empresa_id"`
			Productos struct {
				ImprimirOrdenServicio      bool   `json:"imprimir_orden_servicio"`
				AreaDespacho               string `json:"area_despacho"`
				CopiasOrdenServicio        int64  `json:"copias_orden_servicio"`
				PermitirDescuentoCodigo    bool   `json:"permitir_descuento_codigo"`
				LectorCodigoBarrasAutofoco bool   `json:"lector_codigo_barras_autofoco"`
				CodigosDescuento           string `json:"codigos_descuento"`
			} `json:"productos"`
		} `json:"configuracion"`
	}
	if err := json.Unmarshal(putRR.Body.Bytes(), &putResp); err != nil {
		t.Fatalf("decode put response: %v", err)
	}
	if !putResp.OK || putResp.Configuracion.EmpresaID != 17 {
		t.Fatalf("expected ok response for empresa 17, got %+v", putResp)
	}
	if !putResp.Configuracion.Productos.ImprimirOrdenServicio || putResp.Configuracion.Productos.AreaDespacho != "Cocina fria" || putResp.Configuracion.Productos.CopiasOrdenServicio != 3 || putResp.Configuracion.Productos.PermitirDescuentoCodigo || putResp.Configuracion.Productos.LectorCodigoBarrasAutofoco || putResp.Configuracion.Productos.CodigosDescuento != "VIP10=10%" {
		t.Fatalf("expected persisted payload in response, got %+v", putResp.Configuracion.Productos)
	}

	finalReq := httptest.NewRequest(http.MethodGet, "/api/empresa/configuracion_general?empresa_id=17", nil)
	finalRR := httptest.NewRecorder()
	h.ServeHTTP(finalRR, finalReq)
	if finalRR.Code != http.StatusOK {
		t.Fatalf("expected status %d for final get, got %d body=%s", http.StatusOK, finalRR.Code, finalRR.Body.String())
	}

	var finalResp map[string]interface{}
	if err := json.Unmarshal(finalRR.Body.Bytes(), &finalResp); err != nil {
		t.Fatalf("decode final get response: %v", err)
	}
	productos, _ := finalResp["productos"].(map[string]interface{})
	if productos == nil {
		t.Fatalf("expected productos object in final response, got %+v", finalResp)
	}
	if productos["area_despacho"] != "Cocina fria" {
		t.Fatalf("expected persisted area_despacho, got %+v", productos)
	}
}
