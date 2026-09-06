package handlers

import (
	"testing"

	dbpkg "github.com/you/pos-backend/db"
)

func TestQueueCapacityCandidateTriggersOnlyItsLaneThreshold(t *testing.T) {
	cfg := dbpkg.QueueCapacityConfig{
		Lane: dbpkg.QueueLanePrinting, Label: "Impresiones", AlertsEnabled: true,
		RateLimitPerMinute: 120, PendingAlertThreshold: 20, OldestAlertSeconds: 60, MaxPendingPerTenant: 10,
	}
	snapshot := dbpkg.QueueCapacitySnapshot{Lane: dbpkg.QueueLanePrinting, Label: "Impresiones", Pending: 21, SaturationPercent: 105, QueryOK: true}
	candidate := queueCapacityCandidate(snapshot, cfg)
	if !candidate.Triggered || candidate.Tipo != "queue_capacity_printing" || candidate.Severidad != "warning" {
		t.Fatalf("candidato de impresion inesperado: %+v", candidate)
	}
}

func TestProductAddCapacityCandidateUsesPerTenantPressure(t *testing.T) {
	cfg := dbpkg.QueueCapacityConfig{Lane: dbpkg.QueueLaneProductAdd, Label: "Agregar productos", RateLimitPerMinute: 240}
	snapshot := dbpkg.QueueCapacitySnapshot{Lane: dbpkg.QueueLaneProductAdd, Label: cfg.Label, RequestsCurrentMinute: 2000, BusiestTenantPending: 239, SaturationPercent: 99.58, QueryOK: true}
	if candidate := queueCapacityCandidate(snapshot, cfg); candidate.Triggered {
		t.Fatalf("la actividad global no debe bloquear empresas independientes: %+v", candidate)
	}
	snapshot.BusiestTenantPending = 240
	snapshot.SaturationPercent = 100
	if candidate := queueCapacityCandidate(snapshot, cfg); !candidate.Triggered {
		t.Fatalf("debe alertar cuando una empresa alcanza su limite: %+v", candidate)
	}
}
