package main

import (
	"testing"

	platformworker "github.com/you/pos-backend/internal/platform/worker"
)

func TestProductionRegistryContainsOnlyBusinessHandlers(t *testing.T) {
	registry := businessRegistry(nil, nil)
	if len(registry) < 10 {
		t.Fatalf("production registry has only %d handlers", len(registry))
	}
	if _, exists := registry[platformworker.JobKindPlatformNoop]; exists {
		t.Fatal("production registry includes diagnostic no-op handler")
	}
	if _, exists := registry[jobSystemMetrics]; !exists {
		t.Fatal("production registry must collect metrics through the durable worker")
	}
	if _, exists := registry[jobCxPPayment]; !exists {
		t.Fatal("production registry must reconcile committed CxP payment events")
	}
	if _, exists := registry[jobFacturacionRetries]; !exists {
		t.Fatal("production registry must process durable electronic-invoicing retries")
	}
	for kind, spec := range registry {
		if !spec.Enabled || spec.Handle == nil || spec.Timeout <= 0 || spec.MaxAttempts < 1 {
			t.Errorf("business handler %s is incomplete: %+v", kind, spec)
		}
	}
}

func TestBusinessSchedulesAreAcceptedByWorkerScheduler(t *testing.T) {
	schedules := businessSchedules()
	fiscalShards := map[string]bool{}
	for _, schedule := range schedules {
		if err := platformworker.ValidateScheduleSpec(schedule); err != nil {
			t.Fatalf("business schedule %s is invalid: %v", schedule.Kind, err)
		}
		if schedule.Kind == jobFacturacionRetries {
			t.Fatal("el trabajo fiscal global no debe programarse porque puede causar inanicion entre empresas")
		}
		for i := 0; i < facturacionRetryShardCount(); i++ {
			if schedule.Kind == facturacionRetryShardKind(i) {
				fiscalShards[schedule.Kind] = true
			}
		}
	}
	if len(fiscalShards) != facturacionRetryShardCount() {
		t.Fatalf("shards fiscales programados: got %d want %d", len(fiscalShards), facturacionRetryShardCount())
	}
	registry := businessRegistry(nil, nil)
	for kind := range fiscalShards {
		if _, ok := registry[kind]; !ok {
			t.Fatalf("falta handler para shard fiscal %s", kind)
		}
	}
}
