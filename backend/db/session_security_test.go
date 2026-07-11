package db

import (
	"strings"
	"testing"
)

func TestHashSessionTokenDoesNotRetainRawToken(t *testing.T) {
	raw := "session-secret-for-test"
	hashed := hashSessionToken(raw)
	if !strings.HasPrefix(hashed, sessionTokenHashPrefix) {
		t.Fatalf("expected versioned session verifier, got %q", hashed)
	}
	if strings.Contains(hashed, raw) || hashed == hashSessionToken("different-token") {
		t.Fatal("session verifier must not expose or collide with raw token")
	}
}
