package db

import "testing"

func TestMigrationAdvisoryLockKeyIsStableAndScoped(t *testing.T) {
	t.Parallel()
	first := migrationAdvisoryLockKey("platform:20260715")
	if first <= 0 || first != migrationAdvisoryLockKey("platform:20260715") {
		t.Fatalf("unstable migration lock key %d", first)
	}
	if first == migrationAdvisoryLockKey("platform:20260716") {
		t.Fatal("different migration names must not share a lock key")
	}
}

func TestWithMigrationAdvisoryLockRejectsInvalidInput(t *testing.T) {
	t.Parallel()
	if err := WithMigrationAdvisoryLock(nil, "migration", func() error { return nil }); err == nil {
		t.Fatal("expected nil database error")
	}
}
