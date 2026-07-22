package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWriteSuperMantenimientoPublicErrorRedactsInternalCause(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/super/mantenimiento", nil)
	req.Header.Set("X-Request-ID", "req-maint-105")
	recorder := httptest.NewRecorder()
	writeSuperMantenimientoPublicError(recorder, req, http.StatusInternalServerError, "prueba", errors.New("postgres://private-token@db/internal"))

	body := recorder.Body.String()
	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status=%d", recorder.Code)
	}
	if strings.Contains(body, "postgres://") || strings.Contains(body, "private-token") {
		t.Fatalf("public body exposes internal cause: %q", body)
	}
	if !strings.Contains(body, `"code":"super_mantenimiento_error"`) || !strings.Contains(body, `"request_id":"req-maint-105"`) {
		t.Fatalf("public body missing safe correlation fields: %q", body)
	}
}

func TestWriteSuperMantenimientoAgentPublicErrorRedactsInternalCause(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/super/mantenimiento?action=run_now", nil)
	recorder := httptest.NewRecorder()
	writeSuperMantenimientoAgentPublicError(recorder, req, http.StatusBadGateway, map[string]any{"ok": false, "error": "provider returned Authorization: Bearer secret"})

	body := recorder.Body.String()
	if strings.Contains(body, "Bearer secret") {
		t.Fatalf("agent public body exposes provider detail: %q", body)
	}
	if !strings.Contains(body, `"code":"super_mantenimiento_agent_failed"`) {
		t.Fatalf("agent public body missing stable code: %q", body)
	}
}
