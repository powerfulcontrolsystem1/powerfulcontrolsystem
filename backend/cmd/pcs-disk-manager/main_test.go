package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"
)

func TestNormalizeCategoriesRejectsAnythingOutsideAllowlist(t *testing.T) {
	got, err := normalizeCategories([]string{"images", "containers"})
	if err != nil || len(got) != 2 || got[0] != "containers" || got[1] != "images" {
		t.Fatalf("unexpected normalized categories: %#v, %v", got, err)
	}
	if _, err := normalizeCategories([]string{"images; rm -rf /"}); err == nil {
		t.Fatal("arbitrary input must not become a Docker argument")
	}
}

func TestDiskManagerSignatureRejectsExpiredAndTamperedRequests(t *testing.T) {
	manager := &diskManager{secret: []byte("test-secret")}
	body := []byte(`{"action":"status"}`)
	stamp := time.Now().UTC().Format(time.RFC3339)
	mac := hmac.New(sha256.New, manager.secret)
	_, _ = mac.Write([]byte(stamp + "\n"))
	_, _ = mac.Write(body)
	signature := hex.EncodeToString(mac.Sum(nil))
	if !manager.validSignature(stamp, signature, body) {
		t.Fatal("expected current signed request to be accepted")
	}
	if manager.validSignature(stamp, signature, []byte(`{"action":"cleanup"}`)) {
		t.Fatal("signature must bind the request body")
	}
	if manager.validSignature(time.Now().Add(-2*time.Minute).UTC().Format(time.RFC3339), signature, body) {
		t.Fatal("expired request must be rejected")
	}
}

func TestParseDockerBytes(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want int64
	}{
		{name: "docker compact GB", raw: "10.5GB (40%)", want: int64(10.5 * 1024 * 1024 * 1024)},
		{name: "spaced MiB", raw: "2 MiB (10%)", want: 2 * 1024 * 1024},
		{name: "decimal comma KB", raw: "1,5kB", want: 1536},
		{name: "bytes", raw: "512B", want: 512},
		{name: "terabytes", raw: "1TB (5%)", want: 1024 * 1024 * 1024 * 1024},
		{name: "empty", raw: "", want: 0},
		{name: "invalid number", raw: "unknown", want: 0},
		{name: "unknown unit", raw: "12PB", want: 0},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := parseDockerBytes(test.raw); got != test.want {
				t.Fatalf("parseDockerBytes(%q) = %d, want %d", test.raw, got, test.want)
			}
		})
	}
}
