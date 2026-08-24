package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	dbpkg "github.com/you/pos-backend/db"
)

func TestFacturacionBindAuthorizedEmpresaIDRejectsBodyTenantWithoutContentType(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/empresa/facturacion_electronica?empresa_id=999", strings.NewReader(`{"empresa_id":13}`))
	req = requestWithTenantContext(req, TenantContext{EmpresaID: 12})
	if got := req.Header.Get("Content-Type"); got != "" {
		t.Fatalf("la precondicion de la prueba requiere Content-Type ausente, se obtuvo %q", got)
	}

	payloadEmpresaID := int64(13)
	if err := facturacionBindAuthorizedEmpresaID(req, &payloadEmpresaID); err == nil {
		t.Fatal("se acepto empresa_id del body distinto al contexto autenticado")
	}

	payloadEmpresaID = 0
	if err := facturacionBindAuthorizedEmpresaID(req, &payloadEmpresaID); err != nil {
		t.Fatalf("no se pudo enlazar el contexto autenticado: %v", err)
	}
	if payloadEmpresaID != 12 {
		t.Fatalf("empresa_id enlazado=%d, want 12", payloadEmpresaID)
	}
}

func TestResolveEmpresaIDRejectsBodyTenantWithoutContentType(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/empresa/facturacion_electronica/dian?empresa_id=999", strings.NewReader(`{"empresa_id":13}`))
	req = requestWithTenantContext(req, TenantContext{EmpresaID: 12})
	if got := req.Header.Get("Content-Type"); got != "" {
		t.Fatalf("la precondicion de la prueba requiere Content-Type ausente, se obtuvo %q", got)
	}

	payload := map[string]interface{}{"empresa_id": float64(13)}
	if _, err := resolveEmpresaIDFromPayloadOrRequest(req, payload); err == nil {
		t.Fatal("se acepto empresa_id del body distinto al contexto autenticado")
	}

	payload = map[string]interface{}{}
	empresaID, err := resolveEmpresaIDFromPayloadOrRequest(req, payload)
	if err != nil {
		t.Fatalf("no se pudo resolver el contexto autenticado: %v", err)
	}
	if empresaID != 12 || anyToInt64(payload["empresa_id"]) != 12 {
		t.Fatalf("contexto no canonico: empresaID=%d payload=%v", empresaID, payload["empresa_id"])
	}
}

func TestDIANFreeFormXMLActionsFailClosedBeforeDBOrBodyParsing(t *testing.T) {
	tests := []struct {
		action string
		code   string
	}{
		{action: "firmar_xml_real", code: "firma_xml_directa_bloqueada"},
		{action: "firmar_xml_xades_base", code: "firma_xml_directa_bloqueada"},
		{action: "enviar_documento_real", code: "envio_dian_directo_bloqueado"},
	}

	for _, test := range tests {
		t.Run(test.action, func(t *testing.T) {
			// El handler recibe DB nil y JSON deliberadamente truncado. Si la ruta
			// intentara decodificar, firmar, enviar o persistir antes del bloqueo,
			// esta prueba no podria devolver el contrato 409 esperado.
			req := httptest.NewRequest(http.MethodPost, "/api/empresa/facturacion_electronica/dian?action="+test.action+"&empresa_id=999", strings.NewReader(`{"xml":"<Invoice>","certificado_clave_ref":"file:C:\\outside.pem"`))
			req = requestWithTenantContext(req, TenantContext{EmpresaID: 12})
			rec := httptest.NewRecorder()

			EmpresaDIANColombiaHandler(nil, nil).ServeHTTP(rec, req)

			if rec.Code != http.StatusConflict {
				t.Fatalf("status=%d body=%s, want %d", rec.Code, rec.Body.String(), http.StatusConflict)
			}
			var response map[string]interface{}
			if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
				t.Fatalf("respuesta no JSON: %v body=%q", err, rec.Body.String())
			}
			if response["codigo"] != test.code || response["bloqueado"] != true || response["ok"] != false {
				t.Fatalf("contrato fail-closed inesperado: %#v", response)
			}
			if strings.Contains(rec.Body.String(), "outside.pem") || strings.Contains(rec.Body.String(), "<Invoice>") {
				t.Fatalf("la respuesta reflejo material controlado por el cliente: %s", rec.Body.String())
			}
		})
	}
}

func TestSanitizeDIANConfigForResponseRemovesSecretsAndReportsPresence(t *testing.T) {
	cfg := map[string]interface{}{
		"id":         int64(7),
		"empresa_id": int64(12),
		"nit":        "900000000",
	}
	secretValues := map[string]string{}
	for _, field := range dbpkg.EmpresaDIANConfigSecretColumns() {
		value := "valor-confidencial-" + field
		if field == "software_pin" {
			value = ""
		}
		cfg[field] = value
		secretValues[field] = value
	}

	response := sanitizeDIANConfigForResponse(cfg)
	if response["nit"] != cfg["nit"] || anyToInt64(response["empresa_id"]) != 12 {
		t.Fatalf("se perdieron campos publicos: %#v", response)
	}
	for _, field := range dbpkg.EmpresaDIANConfigSecretColumns() {
		if _, exists := response[field]; exists {
			t.Fatalf("el secreto %s quedo expuesto en la respuesta", field)
		}
		wantConfigured := secretValues[field] != ""
		gotConfigured, ok := response[field+"_configurado"].(bool)
		if !ok || gotConfigured != wantConfigured {
			t.Fatalf("%s_configurado=%#v, want %v", field, response[field+"_configurado"], wantConfigured)
		}
		if _, stillPresent := cfg[field]; !stillPresent {
			t.Fatalf("el sanitizador modifico el mapa interno para %s", field)
		}
	}

	raw, err := json.Marshal(response)
	if err != nil {
		t.Fatal(err)
	}
	for field, value := range secretValues {
		if value != "" && strings.Contains(string(raw), value) {
			t.Fatalf("el valor de %s quedo serializado: %s", field, raw)
		}
	}
}
