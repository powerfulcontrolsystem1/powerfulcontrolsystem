package handlers

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRedactEmailForLog(t *testing.T) {
	tests := map[string]string{
		"ana@example.com":   "a***@example.com",
		"a@example.com":     "a***@example.com",
		" invalid-address ": "[redacted]",
		"":                  "[redacted]",
	}
	for input, want := range tests {
		if got := redactEmailForLog(input); got != want {
			t.Fatalf("redactEmailForLog(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestLegacyEpaycoRuntimeToolDoesNotReadOrPrintSensitiveData(t *testing.T) {
	_, err := os.Stat(filepath.Join("..", "tools", "query_epayco_runtime.go"))
	if !os.IsNotExist(err) {
		t.Fatalf("legacy sensitive payment runtime tool must remain removed, stat err=%v", err)
	}
}
