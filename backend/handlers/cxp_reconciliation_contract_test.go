package handlers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCxPSourceReconciliationIsReadOnlyAction(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("modulos_faltantes.go"))
	if err != nil {
		t.Fatal(err)
	}
	source := string(raw)
	for _, want := range []string{"case \"reconciliacion_fuentes\":", "r.Method != http.MethodGet", "BuildEmpresaCxPReconciliacion(dbEmp, empresaID)"} {
		if !strings.Contains(source, want) {
			t.Fatalf("CxP reconciliation handler contract missing %q", want)
		}
	}
}
