package utils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteAuthFailureUsesMobileEnvelope(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	response := httptest.NewRecorder()
	writeAuthFailure(response, request, http.StatusUnauthorized)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d want=%d", response.Code, http.StatusUnauthorized)
	}
	var payload struct {
		OK    bool `json:"ok"`
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
		RequestID string `json:"request_id"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("invalid mobile JSON: %v", err)
	}
	if payload.OK || payload.Error.Code != "unauthenticated" || payload.RequestID == "" {
		t.Fatalf("unexpected mobile auth payload: %#v", payload)
	}
}
