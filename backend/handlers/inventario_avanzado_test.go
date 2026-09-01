package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWriteInventarioAvanzadoErrorNoExponeDetalleInterno(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/empresa/inventario_avanzado", nil)
	req.Header.Set("X-Request-ID", "qa-inv-001")
	rec := httptest.NewRecorder()

	writeInventarioAvanzadoError(rec, req, "crear_reserva", errors.New("pq: relation inventario_secret does not exist"), http.StatusBadRequest, "No se pudo crear la reserva.")

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status=%d, want %d", rec.Code, http.StatusBadRequest)
	}
	body := rec.Body.String()
	if strings.Contains(body, "inventario_secret") || strings.Contains(body, "pq:") {
		t.Fatalf("internal error leaked in response: %q", body)
	}
	if !strings.Contains(body, "qa-inv-001") {
		t.Fatalf("request reference missing: %q", body)
	}
}
