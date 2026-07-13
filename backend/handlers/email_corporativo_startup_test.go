package handlers

import (
	"os"
	"strings"
	"testing"
)

func TestCorporateEmailStartupProvisioningIsIdempotentAndEncrypted(t *testing.T) {
	raw, err := os.ReadFile("email_corporativo_handlers.go")
	if err != nil {
		t.Fatalf("read email_corporativo_handlers.go: %v", err)
	}
	src := string(raw)
	for _, expected := range []string{
		"func EnsureCorporateEmailProvisioningForExistingCompanies",
		"corporateEmailInitialPasswordForProvision(dbSuper, account)",
		"strings.EqualFold(strings.TrimSpace(account.EstadoProvision), \"provisionado\")",
	} {
		if !strings.Contains(src, expected) {
			t.Fatalf("falta proteccion de aprovisionamiento corporativo: %q", expected)
		}
	}
}
