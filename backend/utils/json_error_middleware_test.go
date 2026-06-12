package utils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestJSONErrorMiddlewarePublicAPIErrorPassesSpecificMessage(t *testing.T) {
	t.Parallel()

	handler := JSONErrorMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		MarkPublicAPIError(w)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":      false,
			"error":   "No se pudo validar el contrato vigente.",
			"detalle": "Recarga la pagina e intenta de nuevo.",
		})
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/empresa/usuarios/establecer_password", strings.NewReader(`{"empresa_id":12}`))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)

	if res.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", res.Code)
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(res.Body.Bytes(), &payload); err != nil {
		t.Fatalf("invalid json: %v body=%s", err, res.Body.String())
	}
	if payload["error"] != "No se pudo validar el contrato vigente." {
		t.Fatalf("error = %#v", payload["error"])
	}
	if payload["request_id"] == "" {
		t.Fatalf("expected request_id in payload: %#v", payload)
	}
	if _, leaked := res.Header()[publicAPIErrorHeader]; leaked {
		t.Fatalf("public marker header leaked to client")
	}
}

func TestJSONErrorMiddlewareUnmarkedServerErrorStaysFriendly(t *testing.T) {
	t.Parallel()

	handler := JSONErrorMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "sql: private detail", http.StatusInternalServerError)
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/empresa/usuarios/establecer_password", strings.NewReader(`{"empresa_id":12}`))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)

	if res.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", res.Code)
	}
	if strings.Contains(res.Body.String(), "sql: private detail") {
		t.Fatalf("unmarked internal detail leaked: %s", res.Body.String())
	}
}
