package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRuntimeHealthHandlerAllowsOnlySafeProbeMethods(t *testing.T) {
	get := httptest.NewRequest(http.MethodGet, "/health", nil)
	getResponse := httptest.NewRecorder()
	RuntimeHealthHandler(getResponse, get)
	if getResponse.Code != http.StatusOK || getResponse.Body.String() != `{"status":"ok"}` {
		t.Fatalf("unexpected health response: status=%d body=%q", getResponse.Code, getResponse.Body.String())
	}
	if got := getResponse.Header().Get("Cache-Control"); got != "no-store" {
		t.Fatalf("health response cache control = %q, want no-store", got)
	}

	head := httptest.NewRequest(http.MethodHead, "/health", nil)
	headResponse := httptest.NewRecorder()
	RuntimeHealthHandler(headResponse, head)
	if headResponse.Code != http.StatusOK || headResponse.Body.Len() != 0 {
		t.Fatalf("unexpected HEAD health response: status=%d body=%q", headResponse.Code, headResponse.Body.String())
	}

	post := httptest.NewRequest(http.MethodPost, "/health", nil)
	postResponse := httptest.NewRecorder()
	RuntimeHealthHandler(postResponse, post)
	if postResponse.Code != http.StatusMethodNotAllowed {
		t.Fatalf("POST health status=%d, want %d", postResponse.Code, http.StatusMethodNotAllowed)
	}
}

func TestRuntimeReadyHandlerFailsClosedWithoutDatabases(t *testing.T) {
	h := RuntimeReadyHandler(nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	response := httptest.NewRecorder()
	h(response, req)
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("ready without databases status=%d, want %d", response.Code, http.StatusServiceUnavailable)
	}
	if response.Body.String() != `{"status":"not_ready"}` {
		t.Fatalf("ready without databases body=%q", response.Body.String())
	}
}
