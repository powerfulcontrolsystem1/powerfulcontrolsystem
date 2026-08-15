package handlers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestControlElectricoLecturasHTTPPropaganContextoAlRepositorio(t *testing.T) {
	source, err := os.ReadFile(filepath.Join("control_electrico.go"))
	if err != nil {
		t.Fatalf("leer handler: %v", err)
	}
	body := string(source)
	for _, expected := range []string{
		"GetEmpresaControlElectricoConfigContext(r.Context(), dbEmp",
		"ListEmpresaControlElectricoRaspberryContext(r.Context(), dbEmp",
		"ListEmpresaControlElectricoRelesContext(r.Context(), dbEmp",
		"ListEmpresaControlElectricoRelesByEstacionContext(r.Context(), dbEmp",
		"ListEmpresaControlElectricoReglasContext(r.Context(), dbEmp",
		"ListEmpresaControlElectricoReglasByEstacionContext(r.Context(), dbEmp",
		"ListEmpresaControlElectricoEventosContext(r.Context(), dbEmp",
		"ListEmpresaControlElectricoEventosByEstacionContext(r.Context(), dbEmp",
		"ListEmpresaControlElectricoLecturasContext(r.Context(), dbEmp",
	} {
		if !strings.Contains(body, expected) {
			t.Fatalf("falta propagacion cancelable: %s", expected)
		}
	}
}
