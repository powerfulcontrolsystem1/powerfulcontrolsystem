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
	for kind, spec := range registry {
		if !spec.Enabled || spec.Handle == nil || spec.Timeout <= 0 || spec.MaxAttempts < 1 {
			t.Errorf("business handler %s is incomplete: %+v", kind, spec)
		}
	}
}

func TestBusinessSchedulesAreAcceptedByWorkerScheduler(t *testing.T) {
	for _, schedule := range businessSchedules() {
		if err := platformworker.ValidateScheduleSpec(schedule); err != nil {
			t.Fatalf("business schedule %s is invalid: %v", schedule.Kind, err)
		}
	}
}
