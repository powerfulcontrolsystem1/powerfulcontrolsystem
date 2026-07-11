package secure

import (
	"crypto/rand"
	"encoding/base64"
	"strings"
	"testing"
)

func testKey(t *testing.T) string {
	t.Helper()
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		t.Fatal(err)
	}
	return base64.StdEncoding.EncodeToString(raw)
}

func TestEncryptionKeyRequiresExactBase64ThirtyTwoBytes(t *testing.T) {
	for _, key := range []string{"", base64.StdEncoding.EncodeToString(make([]byte, 16)), strings.Repeat("x", 32), testKey(t) + "="} {
		t.Setenv("CONFIG_ENC_KEY", key)
		if _, err := getEncKeyFromEnv(); err == nil {
			t.Fatalf("invalid key accepted: %q", key)
		}
	}
	t.Setenv("CONFIG_ENC_KEY", testKey(t))
	if _, err := getEncKeyFromEnv(); err != nil {
		t.Fatalf("valid key rejected: %v", err)
	}
}

func TestEncryptionEnvelopeDetectsTampering(t *testing.T) {
	t.Setenv("CONFIG_ENC_KEY", testKey(t))
	payload, err := EncryptString("confidencial")
	if err != nil || !strings.HasPrefix(payload, "v1:active:") {
		t.Fatalf("versioned encryption failed: %q %v", payload, err)
	}
	plain, err := DecryptString(payload)
	if err != nil || plain != "confidencial" {
		t.Fatalf("decrypt failed: %q %v", plain, err)
	}
	if _, err := DecryptString(payload[:len(payload)-1] + "A"); err == nil {
		t.Fatal("tampered payload accepted")
	}
}
