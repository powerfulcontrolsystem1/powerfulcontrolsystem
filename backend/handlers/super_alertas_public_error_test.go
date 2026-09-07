package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWriteSuperAlertasPublicErrorRedactsInternalCause(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/super/alertas", nil)
	req.Header.Set("X-Request-ID", "req-alert-105")
	recorder := httptest.NewRecorder()
	writeSuperAlertasPublicError(recorder, req, http.StatusInternalServerError, "prueba", errors.New("postgres://internal:secret@db"), map[string]interface{}{"sent": false})

	body := recorder.Body.String()
	if strings.Contains(body, "postgres://") || strings.Contains(body, "secret") {
		t.Fatalf("public body exposes internal cause: %q", body)
	}
	if !strings.Contains(body, `"code":"super_alertas_error"`) || !strings.Contains(body, `"request_id":"req-alert-105"`) || !strings.Contains(body, `"sent":false`) {
		t.Fatalf("public body missing safe fields: %q", body)
	}
}

func TestRedactSuperAlertasEvaluationError(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/super/alertas?action=evaluar", nil)
	eval := superAlertEvaluation{OK: false, Error: "Authorization: Bearer secret"}
	redactSuperAlertasEvaluationError(&eval, req)
	if strings.Contains(eval.Error, "Bearer") || !strings.Contains(eval.Error, "No se pudo evaluar") {
		t.Fatalf("evaluation error was not redacted: %q", eval.Error)
	}
}
