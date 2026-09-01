package db

import (
	"os"
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

func TestMobileAPIIdempotencyCleanupIsBoundedAndExpiryIndexed(t *testing.T) {
	raw, err := os.ReadFile("mobile_api_idempotency.go")
	if err != nil {
		t.Fatal(err)
	}
	source := string(raw)
	for _, required := range []string{
		"func CleanupExpiredMobileAPIIdempotency",
		"WHERE estado = 'completado'",
		"AND fecha_expiracion IS NOT NULL",
		"AND fecha_expiracion < CURRENT_TIMESTAMP",
		"ORDER BY fecha_expiracion ASC",
		"LIMIT ?",
		"mobileAPIIdempotencyCleanupCounter.Add(1)%256",
		"ix_empresa_mobile_api_idempotencia_expiracion",
		"func MarkMobileAPIIdempotencyUncertain",
		"ON CONFLICT (empresa_id, operacion, clave_hash) DO NOTHING",
	} {
		if !strings.Contains(source, required) {
			t.Errorf("bounded idempotency cleanup missing %q", required)
		}
	}
}
