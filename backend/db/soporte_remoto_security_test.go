package db

import (
	"strings"
	"testing"
)

func TestSoporteRemotoSecureSecretUsesThirtyTwoRandomBytes(t *testing.T) {
	first, err := soporteRemotoGenerateSecureSecret("test")
	if err != nil {
		t.Fatalf("generate first secret: %v", err)
	}
	second, err := soporteRemotoGenerateSecureSecret("test")
	if err != nil {
		t.Fatalf("generate second secret: %v", err)
	}
	if first == second {
		t.Fatal("two independently generated secrets matched")
	}
	encoded := strings.TrimPrefix(first, "TEST-")
	if len(encoded) != 64 {
		t.Fatalf("expected 32 random bytes encoded as 64 hex chars, got %d", len(encoded))
	}
}

func TestSoporteRemotoHashComparison(t *testing.T) {
	hash := soporteRemotoHash("credential")
	if !soporteRemotoHashEqual("credential", hash) {
		t.Fatal("valid credential was rejected")
	}
	if soporteRemotoHashEqual("different", hash) {
		t.Fatal("invalid credential was accepted")
	}
}

func TestSoporteRemotoSignalingRoleAllowlist(t *testing.T) {
	for _, role := range []string{"host", "HOST", "viewer", " Viewer "} {
		if soporteRemotoNormalizeSignalingRole(role) == "" {
			t.Fatalf("valid role rejected: %q", role)
		}
	}
	for _, role := range []string{"", "admin", "host;viewer"} {
		if soporteRemotoNormalizeSignalingRole(role) != "" {
			t.Fatalf("invalid role accepted: %q", role)
		}
	}
}
