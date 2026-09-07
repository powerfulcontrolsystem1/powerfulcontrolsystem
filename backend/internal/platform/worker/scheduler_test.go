package worker

import (
	"testing"
	"time"
)

func TestValidateScheduleSpec(t *testing.T) {
	t.Parallel()
	valid := ScheduleSpec{Kind: "maintenance.audit-retention", Version: 1, Interval: time.Hour, MaxAttempts: 5, Priority: 100}
	if err := ValidateScheduleSpec(valid); err != nil {
		t.Fatalf("valid schedule rejected: %v", err)
	}
	valid.Interval = 5 * time.Second
	if err := ValidateScheduleSpec(valid); err != nil {
		t.Fatalf("bounded sub-minute schedule rejected: %v", err)
	}
	valid.Interval = 4 * time.Second
	if err := ValidateScheduleSpec(valid); err == nil {
		t.Fatal("unbounded short schedule must be rejected")
	}
}

func TestScheduledPayloadForBucketIsStableWithinInterval(t *testing.T) {
	t.Parallel()
	interval := time.Minute
	first := time.Date(2026, time.September, 6, 6, 54, 1, 0, time.UTC)
	second := time.Date(2026, time.September, 6, 6, 54, 59, 0, time.UTC)
	next := time.Date(2026, time.September, 6, 6, 55, 0, 0, time.UTC)

	firstPayload := scheduledPayloadForBucket(first, interval)
	if secondPayload := scheduledPayloadForBucket(second, interval); secondPayload != firstPayload {
		t.Fatalf("same interval must preserve payload: first=%s second=%s", firstPayload, secondPayload)
	}
	if nextPayload := scheduledPayloadForBucket(next, interval); nextPayload == firstPayload {
		t.Fatalf("next interval must use a different payload: %s", nextPayload)
	}
}
