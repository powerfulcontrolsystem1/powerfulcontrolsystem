package worker

import "testing"

func TestRunnerRequiresDatabaseAndWorkerID(t *testing.T) {
	t.Parallel()
	if err := (&Runner{}).Run(t.Context()); err == nil {
		t.Fatal("expected database validation error")
	}
}
