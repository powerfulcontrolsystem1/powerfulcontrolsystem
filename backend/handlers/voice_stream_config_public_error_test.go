package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWriteVoiceStreamPublicErrorRedactsInternalCause(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/super/voice", nil)
	req.Header.Set("X-Request-ID", "req-voice-105")
	rec := httptest.NewRecorder()
	writeVoiceStreamPublicError(rec, req, http.StatusBadGateway, "tts_upstream", "voice_stream_unavailable", errors.New("postgres://secret@10.0.0.5:5432/pcs token=abc"))

	body := rec.Body.String()
	for _, blocked := range []string{"postgres://", "10.0.0.5", "token=abc"} {
		if strings.Contains(body, blocked) {
			t.Fatalf("respuesta publica expone causa interna %q: %s", blocked, body)
		}
	}
	if !strings.Contains(body, `"error":"voice_stream_unavailable"`) || !strings.Contains(body, `"request_id":"req-voice-105"`) {
		t.Fatalf("respuesta publica sin codigo o correlacion: %s", body)
	}
}
