package handlers

import (
	"strings"
	"testing"
)

func TestFacturacionSafeDispatchJSONRemovesNestedRawXMLAndSecrets(t *testing.T) {
	t.Parallel()
	raw := facturacionSafeDispatchJSON(map[string]interface{}{
		"ok": true,
		"respuesta_dian": map[string]interface{}{
			"estado":       "pendiente",
			"raw_xml":      "<secret>PII SOAP</secret>",
			"software_pin": "pin-super-secreto",
			"nested": []interface{}{map[string]interface{}{
				"token_emisor": "bearer-secreto",
				"mensaje":      "visible",
			}},
		},
	})
	for _, forbidden := range []string{"PII SOAP", "pin-super-secreto", "bearer-secreto", "raw_xml", "software_pin", "token_emisor"} {
		if strings.Contains(raw, forbidden) {
			t.Fatalf("sanitized JSON leaked %q: %s", forbidden, raw)
		}
	}
	if !strings.Contains(raw, "visible") || !strings.Contains(raw, "pendiente") {
		t.Fatalf("sanitized JSON removed safe fields: %s", raw)
	}
}

func TestDIANSafeTrackHistoryJSONRemovesNestedRawResponse(t *testing.T) {
	t.Parallel()
	raw := dianSafeTrackHistoryJSON(map[string]interface{}{
		"status_code": "00",
		"detalle": map[string]interface{}{
			"raw_response": "SOAP PRIVADO",
			"message":      "Aceptado",
		},
	})
	if strings.Contains(raw, "SOAP PRIVADO") || strings.Contains(raw, "raw_response") {
		t.Fatalf("track history leaked raw response: %s", raw)
	}
	if !strings.Contains(raw, "Aceptado") {
		t.Fatalf("track history removed safe message: %s", raw)
	}
}
