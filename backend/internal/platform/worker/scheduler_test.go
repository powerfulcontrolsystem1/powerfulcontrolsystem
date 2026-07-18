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
