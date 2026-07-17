package handlers

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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

func TestRuntimePrivateStorageReadyWritesAndCleansProbe(t *testing.T) {
	root := t.TempDir()
	t.Setenv("PCS_PRIVATE_STORAGE_DIR", root)
	if err := runtimePrivateStorageReady(); err != nil {
		t.Fatalf("storage readiness failed: %v", err)
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatalf("read storage dir: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("storage readiness left %d temporary files", len(entries))
	}
}

func TestRuntimePrivateStorageReadyRejectsFileAsRoot(t *testing.T) {
	root := filepath.Join(t.TempDir(), "not-a-directory")
	if err := os.WriteFile(root, []byte("blocked"), 0o600); err != nil {
		t.Fatalf("prepare storage file: %v", err)
	}
	t.Setenv("PCS_PRIVATE_STORAGE_DIR", root)
	if err := runtimePrivateStorageReady(); err == nil {
		t.Fatal("storage readiness accepted a regular file as storage root")
	}
}
