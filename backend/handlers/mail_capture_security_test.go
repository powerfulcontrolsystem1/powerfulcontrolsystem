package handlers

import (
	"strings"
	"testing"
)

func TestRedactCapturedMailTokenNeverPersistsOriginalSecret(t *testing.T) {
	secret := "token-temporal-que-no-debe-persistirse"
	ref, body, metadata := redactCapturedMailToken(secret, "Enlace: "+secret, `{"token":"`+secret+`"}`)
	if ref == secret || !strings.HasPrefix(ref, "sha256:") {
		t.Fatalf("expected hashed token reference, got %q", ref)
	}
	if strings.Contains(body, secret) || strings.Contains(metadata, secret) {
		t.Fatal("captured email content must not retain the raw token")
	}
}
