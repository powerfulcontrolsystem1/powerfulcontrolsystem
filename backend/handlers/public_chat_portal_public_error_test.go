package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWritePortalPublicChatStreamErrorRedactsProviderCause(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/public/chat", nil)
	req.Header.Set("X-Request-ID", "req-public-chat-105")
	recorder := httptest.NewRecorder()

	writePortalPublicChatStreamError(recorder, req, errors.New("https://provider.internal/v1 key=secret-token"))

	body := recorder.Body.String()
	if strings.Contains(body, "provider.internal") || strings.Contains(body, "secret-token") {
		t.Fatalf("stream exposes provider cause: %q", body)
	}
	if !strings.Contains(body, `"code":"public_chat_unavailable"`) || !strings.Contains(body, `"request_id":"req-public-chat-105"`) {
		t.Fatalf("stream missing safe correlation fields: %q", body)
	}
}
