package worker

import "testing"

func TestDefaultRegistryIsValidAndNonEmpty(t *testing.T) {
	t.Parallel()
	registry := DefaultRegistry()
	if err := validateHandlerRegistry(registry); err != nil {
		t.Fatalf("validateHandlerRegistry: %v", err)
	}
	if _, ok := registry[JobKindPlatformNoop]; !ok {
		t.Fatal("expected safe platform diagnostic handler")
	}
}

func TestRetryBackoffIsBounded(t *testing.T) {
	t.Parallel()
	if got := retryBackoff(1); got.String() != "1m0s" {
		t.Fatalf("unexpected first retry: %s", got)
	}
	if got := retryBackoff(100); got.String() != "2h8m0s" {
		t.Fatalf("unexpected bounded retry: %s", got)
	}
}
