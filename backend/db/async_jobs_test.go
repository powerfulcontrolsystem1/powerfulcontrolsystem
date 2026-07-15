package db

import "testing"

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

func TestValidateAsyncJobRejectsInvalidOperationalMetadata(t *testing.T) {
	t.Parallel()
	if err := ValidateAsyncJob(AsyncJob{EmpresaID: 1, Kind: "email.send", MaxAttempts: 1, IdempotencyKey: "short"}); err == nil {
		t.Fatal("expected invalid short idempotency key")
	}
	if err := ValidateAsyncJob(AsyncJob{EmpresaID: 1, OriginUserID: -1, Kind: "email.send", MaxAttempts: 1}); err == nil {
		t.Fatal("expected invalid origin user id")
	}
}

func TestAsyncJobIdempotencyHashDoesNotExposeSource(t *testing.T) {
	t.Parallel()
	value := "job-key-0123456789"
	first := asyncJobIdempotencyHash(value)
	if first == value || len(first) != 64 || first != asyncJobIdempotencyHash(value) {
		t.Fatalf("unexpected idempotency hash %q", first)
	}
}
