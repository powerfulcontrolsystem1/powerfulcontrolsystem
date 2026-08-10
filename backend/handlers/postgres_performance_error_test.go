package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWritePostgresPerformanceReadErrorIgnoresClientCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	req := httptest.NewRequest(http.MethodGet, "/super/api/postgres/performance", nil).WithContext(ctx)
	recorder := httptest.NewRecorder()

	writePostgresPerformanceReadError(recorder, req, "cluster", "mensaje publico", context.Canceled)

	if recorder.Body.Len() != 0 {
		t.Fatalf("client cancellation must not write a synthetic error response: %q", recorder.Body.String())
	}
}

func TestWritePostgresPerformanceReadErrorMapsTimeoutToServiceUnavailable(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/super/api/postgres/performance", nil)
	recorder := httptest.NewRecorder()

	writePostgresPerformanceReadError(recorder, req, "cluster", "lectura temporalmente no disponible", context.DeadlineExceeded)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("status=%d want=%d", recorder.Code, http.StatusServiceUnavailable)
	}
	if strings.Contains(recorder.Body.String(), context.DeadlineExceeded.Error()) {
		t.Fatalf("public response leaked internal error: %s", recorder.Body.String())
	}
}

func TestWritePostgresPerformanceReadErrorKeepsUnexpectedFailureGeneric(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/super/api/postgres/performance", nil)
	recorder := httptest.NewRecorder()
	internal := errors.New("driver detail with private schema")

	writePostgresPerformanceReadError(recorder, req, "empresas", "no se pudo leer metricas", internal)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status=%d want=%d", recorder.Code, http.StatusInternalServerError)
	}
	if strings.Contains(recorder.Body.String(), internal.Error()) {
		t.Fatalf("public response leaked internal error: %s", recorder.Body.String())
	}
}
