package worker

import (
	"context"
	"testing"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

func TestRunnerRequiresDatabaseAndWorkerID(t *testing.T) {
	t.Parallel()
	if err := (&Runner{}).Run(t.Context()); err == nil {
		t.Fatal("expected database validation error")
	}
}

func TestRunnerRejectsInvalidRegistry(t *testing.T) {
	t.Parallel()
	runner := &Runner{Registry: []HandlerSpec{{Kind: "", Version: 1, Enabled: true, Handler: func(context.Context, dbpkg.AsyncJob) error { return nil }}}}
	if err := runner.validateRegistry(); err == nil {
		t.Fatal("expected invalid registry")
	}
}

func TestRetryBackoffUsesDeclaredPolicy(t *testing.T) {
	t.Parallel()
	got := retryBackoff(HandlerSpec{Backoff: func(int) time.Duration { return 7 * time.Second }}, 3)
	if got != 7*time.Second {
		t.Fatalf("backoff=%s", got)
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

func TestRegisteredHandlerControlsRetryBudget(t *testing.T) {
	job := jobWithRegisteredMaxAttempts(dbpkg.AsyncJob{MaxAttempts: 5}, HandlerSpec{MaxAttempts: 3})
	if job.MaxAttempts != 3 {
		t.Fatalf("max attempts=%d", job.MaxAttempts)
	}
	unchanged := jobWithRegisteredMaxAttempts(job, HandlerSpec{MaxAttempts: 0})
	if unchanged.MaxAttempts != 3 {
		t.Fatalf("unexpected max attempts=%d", unchanged.MaxAttempts)
	}
}
