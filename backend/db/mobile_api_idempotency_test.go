package db

import (
	"strings"
	"testing"
)

func TestMobileAPIIdempotencyHashDoesNotPersistRawKey(t *testing.T) {
	key := "mobile-20260713-unique-key-0001"
	hash := mobileAPIHash(key)
	if hash == "" || hash == key || strings.Contains(hash, key) {
		t.Fatalf("la clave de idempotencia no quedo protegida: %q", hash)
	}
	if hash != mobileAPIHash(key) || hash == mobileAPIHash(key+"-other") {
		t.Fatal("el hash de idempotencia debe ser determinista y diferenciar claves")
	}
}
