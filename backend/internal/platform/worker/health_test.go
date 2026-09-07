package worker

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestWorkerHealthStateTracksBatchOutcome(t *testing.T) {
	t.Parallel()
	state := &HealthState{}
	if state.Ready() {
		t.Fatal("new worker health state must not be ready")
	}
	state.MarkBatchSuccess(time.Now())
	if !state.Ready() {
		t.Fatal("successful worker batch must make health ready")
	}
	state.MarkBatchFailure(time.Now())
	if state.Ready() {
		t.Fatal("failed worker batch must make readiness fail closed")
	}
}

func TestWorkerHealthAddressStaysLoopback(t *testing.T) {
	t.Parallel()
	for _, addr := range []string{"127.0.0.1:8082", "[::1]:8082", "localhost:8082"} {
		if err := validateLoopbackHealthAddr(addr); err != nil {
			t.Fatalf("expected loopback address %q to be accepted: %v", addr, err)
		}
	}
	for _, addr := range []string{":8082", "0.0.0.0:8082", "10.0.0.1:8082", "127.0.0.1"} {
		if err := validateLoopbackHealthAddr(addr); err == nil {
			t.Fatalf("expected unsafe address %q to be rejected", addr)
		}
	}
}

func TestWorkerHealthMethodsAndNoStore(t *testing.T) {
	t.Parallel()
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/ready", nil)
	if workerHealthMethodAllowed(recorder, request) {
		t.Fatal("POST must not be accepted by worker health endpoint")
	}
	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("unexpected method status: %d", recorder.Code)
	}

	recorder = httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodGet, "/ready", nil)
	workerHealthResponse(recorder, request, http.StatusServiceUnavailable, "not_ready")
	if got := recorder.Header().Get("Cache-Control"); got != "no-store" {
		t.Fatalf("expected no-store cache header, got %q", got)
	}
}
