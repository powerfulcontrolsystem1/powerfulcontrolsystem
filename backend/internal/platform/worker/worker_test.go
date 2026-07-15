package worker

import (
	"testing"
	"time"
)

func TestRunnerRequiresDatabaseAndWorkerID(t *testing.T) {
	t.Parallel()
	if err := (&Runner{}).Run(t.Context()); err == nil {
		t.Fatal("expected database validation error")
	}
}

func TestRunnerUsesSafeOperationalDefaults(t *testing.T) {
	runner := &Runner{}
	runner.normalize()
	if runner.Poll != 2*time.Second {
		t.Fatalf("poll=%s", runner.Poll)
	}
	if runner.Batch != 20 {
		t.Fatalf("batch=%d", runner.Batch)
	}
	if runner.Lease != 5*time.Minute {
		t.Fatalf("lease=%s", runner.Lease)
	}
}
