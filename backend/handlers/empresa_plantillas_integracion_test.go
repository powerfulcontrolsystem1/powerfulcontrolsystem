package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	dbpkg "github.com/you/pos-backend/db"
)

func TestEmpresaPlantillasIntegracionCatalogoContrato(t *testing.T) {
	items := buildEmpresaPlantillasIntegracionCatalogo()
	if len(items) != 29 {
		t.Fatalf("catalogo universal debe publicar 29 plantillas canonicos exactos, obtuvo %d", len(items))
	}
	seen := map[string]bool{}
	forbidden := map[string]string{
		"consultorio_odontologico": "alias fusionado en odontologia",
		"taxi":                     "alias fusionado en taxi_system",
		"turnos_atencion":          "capacidad de soporte transversal",
		"turnos":                   "alias de soporte transversal",
	}
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
		if reason, ok := forbidden[item.Modulo]; ok {
			t.Fatalf("modulo no canonico publicado como plantilla: %s (%s)", item.Modulo, reason)
		}
		seen[item.Modulo] = true
		if _, ok := required[item.Modulo]; ok {
			required[item.Modulo] = true
		}
		if item.AliasDe != "" {
			t.Fatalf("alias publicado como plantilla canonica: %+v", item)
		}
		if item.OperationalVisible && len(item.DuplicatesCore) != 0 {
			t.Fatalf("plantilla visible con duplicados del nucleo: %+v", item)
		}
		if item.OperationalVisible && item.IntegrationStatus != "plantilla_integrada_nucleo" {
			t.Fatalf("estado visible invalido para %s: %s", item.Modulo, item.IntegrationStatus)
		}
		if len(item.OwnFlowAllowed) == 0 {
			t.Fatalf("plantilla sin flujo especializado declarado: %+v", item)
		}
		if item.OperationalVisible && len(item.TemplateActivates) == 0 {
			t.Fatalf("plantilla visible sin plantilla declarada: %+v", item)
		}
		if item.OperationalVisible && len(item.TablesTouched) == 0 {
			t.Fatalf("plantilla visible sin tablas declaradas: %+v", item)
		}
		if item.OperationalVisible && len(item.RequiredPermissions) == 0 {
			t.Fatalf("plantilla visible sin permisos declarados: %+v", item)
		}
		if item.OperationalVisible && len(item.SaleFlow) == 0 {
			t.Fatalf("plantilla visible sin flujo de venta declarado: %+v", item)
		}
		if item.OperationalVisible && len(item.ReportsProduced) == 0 {
			t.Fatalf("plantilla visible sin reportes declarados: %+v", item)
		}
		if item.OperationalVisible && !hasAllStringValues(item.FinancialCoreModules, []string{"ventas", "pagos", "finanzas", "reportes"}) {
			t.Fatalf("plantilla visible sin nucleo financiero declarado: %+v", item)
		}
		if item.OperationalVisible && len(item.IncomeFlow) == 0 {
			t.Fatalf("plantilla visible sin flujo de ingresos del nucleo: %+v", item)
		}
		if item.OperationalVisible && len(item.ExpenseFlow) == 0 {
			t.Fatalf("plantilla visible sin flujo de egresos del nucleo: %+v", item)
		}
		if item.OperationalVisible && !hasAllStringValues(item.FinancialTables, []string{"empresa_finanzas_movimientos"}) {
			t.Fatalf("plantilla visible sin tablas financieras del nucleo: %+v", item)
		}
		if item.OperationalVisible && len(item.FinancialReports) == 0 {
			t.Fatalf("plantilla visible sin reportes financieros del nucleo: %+v", item)
		}
		if item.OperationalVisible && !item.ProfessionalReady {
			t.Fatalf("plantilla visible sin preparacion profesional completa: %+v", item)
		}
		if item.OperationalVisible && item.ReadinessScore != 100 {
			t.Fatalf("plantilla visible con readiness score invalido: %+v", item)
		}
		if item.OperationalVisible && len(item.ConfigurationScope) == 0 {
			t.Fatalf("plantilla visible sin alcance de configuracion: %+v", item)
		}
		payload, err := json.Marshal(item)
		if err != nil {
			t.Fatalf("marshal item: %v", err)
		}
		if item.OperationalVisible && (jsonContainsKey(payload, "sync_action") || jsonContainsKey(payload, "sync_path") || jsonContainsKey(payload, "sync_action_name")) {
			t.Fatalf("plantilla visible publica contrato viejo de sincronizacion: %s", string(payload))
		}
	}
	for module, ok := range required {
		if !ok {
			t.Fatalf("plantilla requerida faltante en catalogo: %s", module)
		}
	}
	for _, nuevo := range dbpkg.NuevasPlantillasTipoEmpresaCatalog() {
		module := dbpkg.NormalizeEmpresaModuloColombia(nuevo.Modulo)
		if module == "" {
			continue
		}
		if !seen[module] {
			t.Fatalf("nueva plantilla faltante en matriz universal: %s", module)
		}
	}
	if len(items)-len(dbpkg.NuevasPlantillasTipoEmpresaCatalog()) != 10 {
		t.Fatalf("la matriz debe conservar 10 plantillas clasicos canonicos y 19 nuevos: total=%d nuevos=%d", len(items), len(dbpkg.NuevasPlantillasTipoEmpresaCatalog()))
	}
	if !hasAllStringValues(findVerticalForTest(items, "odontologia").FusedModules, []string{"consultorio_odontologico"}) {
		t.Fatalf("odontologia debe declarar fusion de consultorio_odontologico")
	}
	if !hasAllStringValues(findVerticalForTest(items, "taxi_system").FusedModules, []string{"taxi"}) {
		t.Fatalf("taxi_system debe declarar fusion de taxi")
	}
}

func findVerticalForTest(items []empresaVerticalIntegracionItem, module string) empresaVerticalIntegracionItem {
	for _, item := range items {
		if item.Modulo == module {
			return item
		}
	}
	return empresaVerticalIntegracionItem{}
}

func jsonContainsKey(payload []byte, key string) bool {
	var raw map[string]any
	if err := json.Unmarshal(payload, &raw); err != nil {
		return false
	}
	_, ok := raw[key]
	return ok
}

func TestPublicPlantillasIntegracionCatalogoHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/public/plantillas_integracion/catalogo", nil)
	rr := httptest.NewRecorder()
	PublicPlantillasIntegracionCatalogoHandler().ServeHTTP(rr, req)
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
	if !payload.OK || payload.Total != len(payload.Items) || payload.Total != 29 {
		t.Fatalf("payload inesperado: %+v", payload)
	}
}
