package db

import (
	"os"
	"strings"
	"testing"
)

func TestFacturacionRetryQueueUsesDurableClaim(t *testing.T) {
	raw, err := os.ReadFile("facturacion_electronica.go")
	if err != nil {
		t.Fatal(err)
	}
	source := string(raw)
	for _, required := range []string{
		"func ClaimFacturacionElectronicaRetriesByEmpresa",
		"FOR UPDATE SKIP LOCKED",
		"lease_until < CURRENT_TIMESTAMP",
		"CURRENT_TIMESTAMP + interval '5 minutes'",
		"func ReleaseFacturacionElectronicaRetryClaim",
		"AND lease_token = ?",
	} {
		if !strings.Contains(source, required) {
			t.Errorf("falta contrato durable de claim FE %q", required)
		}
	}
}
