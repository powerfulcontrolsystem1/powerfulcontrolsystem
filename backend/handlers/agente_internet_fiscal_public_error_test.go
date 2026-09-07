package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	dbpkg "github.com/you/pos-backend/db"
)

func TestWriteAgenteInternetFiscalPublicErrorRedactsInternalCause(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/empresa/agente_internet/nomina", nil)
	req.Header.Set("X-Request-ID", "req-agent-105")
	recorder := httptest.NewRecorder()
	writeAgenteInternetFiscalPublicError(recorder, req, http.StatusTooManyRequests, errors.New("postgres://internal:secret@db"), dbpkg.EmpresaAgenteUsoDiario{}, map[string]int64{"consultas_ligeras_diarias": 20})

	body := recorder.Body.String()
	if strings.Contains(body, "postgres://") || strings.Contains(body, "secret") {
		t.Fatalf("public body exposes internal cause: %q", body)
	}
	if !strings.Contains(body, `"code":"agente_internet_usage_unavailable"`) || !strings.Contains(body, `"request_id":"req-agent-105"`) {
		t.Fatalf("public body missing safe fields: %q", body)
	}
}
