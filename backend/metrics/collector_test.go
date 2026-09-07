package metrics

import "testing"

func TestDefaultIntervalSecondsUsesPositiveConfiguredValue(t *testing.T) {
	t.Setenv("METRICS_INTERVAL_SECONDS", "45")
	if got := DefaultIntervalSeconds(); got != 45 {
		t.Fatalf("DefaultIntervalSeconds() = %d, want 45", got)
	}
}

func TestDefaultIntervalSecondsFallsBackForInvalidValue(t *testing.T) {
	t.Setenv("METRICS_INTERVAL_SECONDS", "invalid")
	if got := DefaultIntervalSeconds(); got != 10 {
		t.Fatalf("DefaultIntervalSeconds() = %d, want 10", got)
	}
}

func TestDefaultIntervalSecondsRejectsUnsafeShortInterval(t *testing.T) {
	t.Setenv("METRICS_INTERVAL_SECONDS", "1")
	if got := DefaultIntervalSeconds(); got != 10 {
		t.Fatalf("DefaultIntervalSeconds() = %d, want safe fallback 10", got)
	}
}
