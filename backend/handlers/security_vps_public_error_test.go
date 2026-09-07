package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWriteSecurityVPSPublicErrorRedactsInternalCause(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/super/security/vps/run", nil)
	req.Header.Set("X-Request-ID", "req-vps-105")
	recorder := httptest.NewRecorder()
	writeSecurityVPSPublicError(recorder, req, http.StatusInternalServerError, "prueba", errors.New("/srv/private/.env Authorization: Bearer secret"), map[string]interface{}{"status": map[string]string{"state": "failed"}})

	body := recorder.Body.String()
	if strings.Contains(body, "/srv/private") || strings.Contains(body, "Bearer secret") {
		t.Fatalf("public body exposes internal cause: %q", body)
	}
	if !strings.Contains(body, `"code":"security_vps_error"`) || !strings.Contains(body, `"request_id":"req-vps-105"`) || !strings.Contains(body, `"state":"failed"`) {
		t.Fatalf("public body missing safe fields: %q", body)
	}
}
