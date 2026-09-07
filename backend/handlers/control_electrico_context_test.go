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
		"SetEmpresaControlElectricoRaspberryEstadoContext(r.Context(), dbEmp",
		"SetEmpresaControlElectricoReglaEstadoContext(r.Context(), dbEmp",
		"SetEmpresaControlElectricoReleEstadoContext(r.Context(), dbEmp",
	} {
		if !strings.Contains(body, expected) {
			t.Fatalf("falta propagacion cancelable: %s", expected)
		}
	}
	for _, ignored := range []string{
		"eventos, _ := dbpkg.ListEmpresaControlElectricoEventos",
		"lecturas, _ := dbpkg.ListEmpresaControlElectricoLecturas",
	} {
		if strings.Contains(body, ignored) {
			t.Fatalf("el reporte no puede ocultar error de lectura: %s", ignored)
		}
	}
}
