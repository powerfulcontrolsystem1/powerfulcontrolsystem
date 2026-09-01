package db

import (
	"os"
	"strings"
	"testing"
)

func TestOfflineVentaSyncHasImmutablePayloadAndLease(t *testing.T) {
	raw, err := os.ReadFile("offline_ventas.go")
	if err != nil {
		t.Fatal(err)
	}
	source := string(raw)
	for _, required := range []string{
		"ErrEmpresaVentaOfflineIdempotencyConflict",
		"payload_hash",
		"func ClaimEmpresaVentaOfflineSync",
		"processing_until IS NULL OR processing_until < CURRENT_TIMESTAMP",
		"processing_token=NULL",
	} {
		if !strings.Contains(source, required) {
			t.Errorf("falta contrato idempotente offline %q", required)
		}
	}
}
