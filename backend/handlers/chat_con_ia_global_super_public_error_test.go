package handlers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSuperAIProviderErrorsAreRedacted(t *testing.T) {
	path := filepath.Join("chat_con_ia_global_super.go")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	source := string(data)
	for _, forbidden := range []string{
		`http.Error(w, err.Error(), http.StatusBadGateway)`,
		`openAIStreamEvent{Error: err.Error()}`,
	} {
		if strings.Contains(source, forbidden) {
			t.Fatalf("provider error remains publicly exposed: %s", forbidden)
		}
	}
	for _, required := range []string{
		`No se pudo consultar el proveedor de IA`,
		`No se pudo completar la consulta de IA`,
	} {
		if !strings.Contains(source, required) {
			t.Fatalf("redacted public message missing: %s", required)
		}
	}
}
