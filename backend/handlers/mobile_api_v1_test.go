package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
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

func TestMobileAPIJSONPreservesExistingEnvelope(t *testing.T) {
	h := mobileAPIJSON(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusCreated, mobileAPIEnvelope{OK: true, Data: map[string]string{"estado": "creado"}})
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/v1/empresa/carritos", nil))
	if rec.Code != http.StatusCreated {
		t.Fatalf("status inesperado: %d", rec.Code)
	}
	var out mobileAPIEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if !out.OK || out.RequestID == "" {
		t.Fatalf("envoltorio v1 invalido: %#v", out)
	}
	data, ok := out.Data.(map[string]interface{})
	if !ok || data["estado"] != "creado" {
		t.Fatalf("respuesta se anido o perdio: %#v", out.Data)
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

func TestMobileNormalizeEmpresaJSONUsesQueryTenant(t *testing.T) {
	var got map[string]interface{}
	h := mobileNormalizeEmpresaJSON(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatal(err)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/empresa/carritos?empresa_id=19", bytes.NewBufferString(`{"empresa_id":999,"nombre":"Caja movil"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent || got["empresa_id"] != float64(19) {
		t.Fatalf("tenant no normalizado: status=%d payload=%#v", rec.Code, got)
	}
}

func TestValidMobileIdempotencyKey(t *testing.T) {
	for _, key := range []string{"mobile-20260713-0001", "aBc_1234567890-xyz"} {
		if !validMobileIdempotencyKey(key) {
			t.Fatalf("clave valida rechazada: %q", key)
		}
	}
	for _, key := range []string{"corta", "clave con espacios 123456", "clave/con/slash-123456"} {
		if validMobileIdempotencyKey(key) {
			t.Fatalf("clave insegura aceptada: %q", key)
		}
	}
}

func TestMobileCursorPagination(t *testing.T) {
	meta := mobilePageMeta(25, 0, 25, 25)
	cursor, _ := meta["next_cursor"].(string)
	if cursor == "" {
		t.Fatal("next cursor was not generated")
	}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/empresa/productos?limit=25&cursor="+url.QueryEscape(cursor), nil)
	limit, offset, err := mobileLimitOffset(req)
	if err != nil || limit != 25 || offset != 25 {
		t.Fatalf("cursor pagination failed: limit=%d offset=%d err=%v", limit, offset, err)
	}
	req = httptest.NewRequest(http.MethodGet, "/api/v1/empresa/productos?offset=1&cursor="+url.QueryEscape(cursor), nil)
	if _, _, err := mobileLimitOffset(req); err == nil {
		t.Fatal("offset and cursor were accepted together")
	}
}

func TestMobileIdempotentWhenMutatingLeavesReadsUntouched(t *testing.T) {
	calls := 0
	h := mobileIdempotentWhenMutating(nil, "prueba", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusNoContent)
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/empresa/productos", nil))
	if rec.Code != http.StatusNoContent || calls != 1 {
		t.Fatalf("lectura no debia exigir idempotencia: status=%d calls=%d", rec.Code, calls)
	}
}
