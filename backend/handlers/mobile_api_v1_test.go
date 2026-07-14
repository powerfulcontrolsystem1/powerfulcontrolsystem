package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMobileAPIJSONNormalizesLegacyErrors(t *testing.T) {
	h := mobileAPIJSON(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "detalle interno", http.StatusForbidden)
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/empresa/productos", nil))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status inesperado: %d", rec.Code)
	}
	var out mobileAPIEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if out.OK || out.Error == nil || out.Error.Code != "forbidden" || out.Error.Message == "detalle interno" || out.RequestID == "" {
		t.Fatalf("error movil no normalizado: %#v", out)
	}
}

func TestMobileFieldSelectionIsClosedList(t *testing.T) {
	items := []map[string]interface{}{{"id": 1, "nombre": "Producto", "costo_interno": 900}}
	selected := mobileSelectFields(items, "id,costo_interno", map[string]bool{"id": true, "nombre": true})
	b, err := json.Marshal(selected)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != `[{"id":1}]` {
		t.Fatalf("seleccion de campos expuso un valor no permitido: %s", b)
	}
}
