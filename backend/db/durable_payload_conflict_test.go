package db

import (
	"os"
	"strings"
	"testing"
)

func TestDurableIdempotencyRejectsPayloadMismatch(t *testing.T) {
	checks := map[string][]string{
		"outbox.go": {
			"ErrOutboxIdempotencyConflict",
			"existing.Version != event.Version",
			"existing.PayloadJSON",
		},
		"async_jobs.go": {
			"ErrAsyncJobIdempotencyConflict",
			"stored.Version != job.Version",
			"stored.PayloadJSON",
		},
	}
	for file, required := range checks {
		raw, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		for _, contract := range required {
			if !strings.Contains(string(raw), contract) {
				t.Errorf("%s no rechaza mismatch durable %q", file, contract)
			}
		}
	}
}
