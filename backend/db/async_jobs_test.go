package db

import (
	"fmt"
	"strings"
	"testing"
)

func TestValidateAsyncJobRejectsUnsafeInput(t *testing.T) {
	t.Parallel()
	if err := ValidateAsyncJob(AsyncJob{EmpresaID: -1, Kind: "mail", MaxAttempts: 1}); err == nil {
		t.Fatal("expected tenant validation error")
	}
	if err := ValidateAsyncJob(AsyncJob{EmpresaID: 1, Kind: "", MaxAttempts: 1}); err == nil {
		t.Fatal("expected kind validation error")
	}
	if err := ValidateAsyncJob(AsyncJob{EmpresaID: 1, Kind: "mail", MaxAttempts: 26}); err == nil {
		t.Fatal("expected retry validation error")
	}
}

func TestValidateAsyncJobAcceptsTenantScopedWork(t *testing.T) {
	t.Parallel()
	if err := ValidateAsyncJob(AsyncJob{EmpresaID: 42, Kind: "email.send", PayloadJSON: `{"message_id":"safe-reference"}`, MaxAttempts: 5}); err != nil {
		t.Fatalf("ValidateAsyncJob: %v", err)
	}
}

func TestAsyncJobRedactionAndIdempotencyHashDoNotExposeRawInput(t *testing.T) {
	t.Parallel()
	if got := hashAsyncJobKey("retry-key"); got == "" || got == "retry-key" {
		t.Fatalf("expected stable non-raw key hash, got %q", got)
	}
	message := redactAsyncJobError(fmt.Errorf("provider failed token=private-value"))
	if strings.Contains(message, "private-value") {
		t.Fatalf("job error leaked sensitive value: %q", message)
	}
}
