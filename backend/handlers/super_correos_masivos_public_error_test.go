package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWriteSuperCorreosMasivosPublicErrorRedactsInternalCause(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/super/correos_masivos", nil)
	req.Header.Set("X-Request-ID", "req-mail-105")
	recorder := httptest.NewRecorder()
	writeSuperCorreosMasivosPublicError(recorder, req, http.StatusInternalServerError, "prueba", errors.New("smtp://user:secret@mail.internal"), map[string]interface{}{"result": map[string]interface{}{"campaign_id": int64(3)}})

	body := recorder.Body.String()
	if strings.Contains(body, "smtp://") || strings.Contains(body, "secret") {
		t.Fatalf("public body exposes internal cause: %q", body)
	}
	if !strings.Contains(body, `"code":"super_correos_masivos_error"`) || !strings.Contains(body, `"request_id":"req-mail-105"`) || !strings.Contains(body, `"campaign_id":3`) {
		t.Fatalf("public body missing safe fields: %q", body)
	}
}

func TestSuperCorreoMasivoValidationMessageDoesNotExposeUnexpectedCause(t *testing.T) {
	message := superCorreoMasivoValidationMessage(errors.New("smtp://user:secret@mail.internal"))
	if strings.Contains(message, "smtp://") || strings.Contains(message, "secret") {
		t.Fatalf("validation message exposes internal cause: %q", message)
	}
	if message != "los datos del correo masivo no son validos" {
		t.Fatalf("unexpected safe validation message: %q", message)
	}
}
