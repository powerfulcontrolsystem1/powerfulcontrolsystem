package handlers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenericPayloadDropsServerControlledIdentityAndLifecycle(t *testing.T) {
	payload := map[string]interface{}{
		"empresa_id":         99,
		"id":                 88,
		"usuario_creador":    "attacker@example.invalid",
		"estado":             "activo",
		"estado_orden":       "cerrada",
		"cantidad_producida": 500,
		"costo_real":         9000,
		"producto_id":        7,
	}
	sanitizeEmpresaGenericPayload(payload, cfgProduccionOrdenes, false)
	for _, field := range []string{"empresa_id", "id", "usuario_creador", "estado", "estado_orden", "cantidad_producida", "costo_real"} {
		if _, exists := payload[field]; exists {
			t.Fatalf("campo controlado por servidor sobrevivio al saneamiento: %s", field)
		}
	}
	if payload["producto_id"] != 7 {
		t.Fatal("el saneamiento elimino un campo operativo permitido")
	}
}

func TestGenericPayloadCreateForcesActiveBaseState(t *testing.T) {
	payload := map[string]interface{}{"estado": "inactivo", "estado_envio": "entregado"}
	sanitizeEmpresaGenericPayload(payload, cfgLogisticaEnvios, true)
	if payload["estado"] != "activo" {
		t.Fatalf("estado inicial=%v, want activo", payload["estado"])
	}
	if _, exists := payload["estado_envio"]; exists {
		t.Fatal("el cliente pudo elegir el estado inicial del flujo")
	}
}

func TestCRMFormsDelegateLifecycleChangesToTransitionActions(t *testing.T) {
	pageRaw, err := os.ReadFile(filepath.Join("..", "..", "web", "administrar_empresa", "crm_comercial.html"))
	if err != nil {
		t.Fatal(err)
	}
	page := string(pageRaw)
	for _, field := range []string{"leadEstado", "interaccionEstado", "cotizacionEstado", "campanaEstado"} {
		marker := `id="` + field + `" class="form-input" disabled`
		if !strings.Contains(page, marker) {
			t.Fatalf("el estado %s debe ser solo lectura en el formulario general", field)
		}
	}

	scriptRaw, err := os.ReadFile(filepath.Join("..", "..", "web", "js", "crm_comercial.js"))
	if err != nil {
		t.Fatal(err)
	}
	script := string(scriptRaw)
	for _, assignment := range []string{"estado_lead: normalize", "estado_interaccion: normalize", "estado_documento: normalize", "estado_campana: normalize"} {
		if strings.Contains(script, assignment) {
			t.Fatalf("el CRUD general de CRM aun intenta escribir el ciclo de vida: %s", assignment)
		}
	}
	if !strings.Contains(script, `interaccion: ["/api/empresa/crm/interacciones?action=transicionar"`) {
		t.Fatal("CRM debe conservar la accion dedicada para transiciones de seguimiento")
	}
}
