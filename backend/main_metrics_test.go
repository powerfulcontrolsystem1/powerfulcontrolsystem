package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestPrometheusMetricsHandlerExposesAggregateOperationalSignals(t *testing.T) {
	originalBusinessDB, originalSuperDB := dbEmpresas, dbSuper
	dbEmpresas, dbSuper = nil, nil
	operationalMetricsCache = prometheusOperationalCache{}
	t.Cleanup(func() {
		dbEmpresas, dbSuper = originalBusinessDB, originalSuperDB
		operationalMetricsCache = prometheusOperationalCache{}
	})

	rec := httptest.NewRecorder()
	prometheusMetricsHandler(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /metrics status = %d, want 200", rec.Code)
	}
	if got := rec.Header().Get("Content-Type"); !strings.Contains(got, "text/plain") {
		t.Fatalf("Content-Type = %q, want Prometheus text", got)
	}
	body := rec.Body.String()
	expected := []string{
		"pcs_backend_up 1",
		`pcs_postgres_ready{database="business"} 0`,
		`pcs_postgres_ready{database="super"} 0`,
		"pcs_worker_heartbeat_age_seconds -1.000",
		`pcs_outbox_ready_total{source="business_outbox"} 0`,
		`pcs_outbox_expired_leases_total{source="super_outbox"} 0`,
		`pcs_async_jobs_ready_total{source="super_jobs"} 0`,
		`pcs_observability_query_success{source="super_jobs"} 0`,
		"pcs_support_purge_pending_total 0",
		"pcs_support_purge_stale_total 0",
		`pcs_observability_query_success{source="support_purge"} 0`,
	}
	for _, want := range expected {
		if !strings.Contains(body, want) {
			t.Fatalf("metrics body does not contain %q: %q", want, body)
		}
	}
	for _, forbidden := range []string{"empresa_id", "payload", "last_error", "credential", "email"} {
		if strings.Contains(strings.ToLower(body), forbidden) {
			t.Fatalf("metrics body leaked forbidden field %q: %q", forbidden, body)
		}
	}
	if got := rec.Header().Get("Cache-Control"); got != "no-store" {
		t.Fatalf("Cache-Control = %q, want no-store", got)
	}
}

func TestPrometheusOperationalCacheReusesSnapshotWithinTTL(t *testing.T) {
	cache := prometheusOperationalCache{}
	first := cache.get(context.Background(), nil, nil)
	if first.workerAge != -1 {
		t.Fatalf("first worker age = %v, want -1", first.workerAge)
	}
	cache.values.workerAge = 42
	second := cache.get(context.Background(), nil, nil)
	if second.workerAge != 42 {
		t.Fatalf("cached worker age = %v, want 42", second.workerAge)
	}
}

func TestRenderPrometheusMetricsUsesOnlyBoundedAggregateLabels(t *testing.T) {
	values := defaultPrometheusOperationalMetrics()
	values.businessDBReady = 1
	values.superDBReady = 1
	values.workerAge = 12.5
	values.workerQueryOK = 1
	values.businessOutbox = prometheusQueueMetrics{ready: 2, processing: 1, dead: 3, expiredLeases: 4, queryOK: 1}
	values.superOutbox = prometheusQueueMetrics{ready: 5, queryOK: 1}
	values.asyncJobs = prometheusQueueMetrics{ready: 6, processing: 7, dead: 8, expiredLeases: 9, queryOK: 1}
	values.supportPurge = prometheusSupportPurgeMetrics{pending: 10, stale: 2, purged: 11, queryOK: 1}

	body := renderPrometheusMetrics(values)
	for _, want := range []string{
		"pcs_worker_heartbeat_age_seconds 12.500",
		`pcs_outbox_dead_total{source="business_outbox"} 3`,
		`pcs_async_jobs_expired_leases_total{source="super_jobs"} 9`,
		"pcs_support_purge_pending_total 10",
		"pcs_support_purge_stale_total 2",
		"pcs_support_purged_total 11",
		`pcs_observability_query_success{source="support_purge"} 1`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("metrics body does not contain %q: %q", want, body)
		}
	}
	if strings.Contains(strings.ToLower(body), "tenant") {
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

func TestSupportPurgeMonitoringConfigurationContract(t *testing.T) {
	rules, err := os.ReadFile("../deploy/monitoring/alert_rules.yml")
	if err != nil {
		t.Fatal(err)
	}
	rulesText := string(rules)
	for _, want := range []string{
		"alert: PCSSoporteIAPurgaVencida",
		`expr: pcs_support_purge_stale_total{job=~"pcs-backend|pcs-staging-backend"} > 0`,
		"runbook",
	} {
		if !strings.Contains(rulesText, want) {
			t.Fatalf("alert rules do not contain %q", want)
		}
	}
	if strings.Count(rulesText, "alert: PCSSoporteIAPurgaVencida") != 1 {
		t.Fatal("support purge alert must be unique")
	}

	dashboard, err := os.ReadFile("../deploy/monitoring/grafana/dashboards/pcs-operacion.json")
	if err != nil {
		t.Fatal(err)
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal(dashboard, &parsed); err != nil {
		t.Fatalf("dashboard JSON invalid: %v", err)
	}
	for _, want := range []string{"Depuraciones IA pendientes", "Depuraciones IA vencidas", "pcs_support_purge_stale_total"} {
		if !strings.Contains(string(dashboard), want) {
			t.Fatalf("dashboard does not contain %q", want)
		}
	}
}
