package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPrometheusMetricsHandlerExposesOnlyAvailability(t *testing.T) {
	rec := httptest.NewRecorder()
	prometheusMetricsHandler(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /metrics status = %d, want 200", rec.Code)
	}
	if got := rec.Header().Get("Content-Type"); !strings.Contains(got, "text/plain") {
		t.Fatalf("Content-Type = %q, want Prometheus text", got)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "pcs_backend_up 1") || strings.Contains(strings.ToLower(body), "empresa") {
		t.Fatalf("unexpected metrics body: %q", body)
	}
}

func TestPrometheusMetricsHandlerRejectsMutations(t *testing.T) {
	rec := httptest.NewRecorder()
	prometheusMetricsHandler(rec, httptest.NewRequest(http.MethodPost, "/metrics", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("POST /metrics status = %d, want 405", rec.Code)
	}
}
