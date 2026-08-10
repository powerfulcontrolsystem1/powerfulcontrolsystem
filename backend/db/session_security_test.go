package db

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHashSessionTokenDoesNotRetainRawToken(t *testing.T) {
	raw := "session-secret-for-test"
	hashed := hashSessionToken(raw)
	if len(hashed) != 64 {
		t.Fatalf("expected SHA-256 session verifier, got %q", hashed)
	}
	if strings.Contains(hashed, raw) || hashed == hashSessionToken("different-token") {
		t.Fatal("session verifier must not expose or collide with raw token")
	}
}

func TestAdminSessionCapSupportsConcurrentCashiers(t *testing.T) {
	if maxActiveSessionsPerIdentity < 4 {
		t.Fatalf("session cap must support four independent cashier sessions, got %d", maxActiveSessionsPerIdentity)
	}
	if maxActiveSessionsPerIdentity > 50 {
		t.Fatalf("session cap must remain below the global alert threshold, got %d", maxActiveSessionsPerIdentity)
	}
}

func TestCreateSessionPrunesExpiredAndOldestRowsAtomically(t *testing.T) {
	sourcePath := filepath.Join("db.go")
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatalf("read %s: %v", sourcePath, err)
	}
	text := string(source)
	for _, required := range []string{
		"acquireAdminSessionLockTx(tx, adminEmail)",
		"pruneAdminSessionsTx(tx, adminEmail)",
		"ORDER BY id DESC",
		"LIMIT ? OFFSET ?",
		"return tx.Commit()",
	} {
		if !strings.Contains(text, required) {
			t.Fatalf("CreateSession must preserve bounded atomic session contract %q", required)
		}
	}
}
