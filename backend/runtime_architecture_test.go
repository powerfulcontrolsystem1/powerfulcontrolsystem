package main

import (
	"os"
	"strings"
	"testing"
)

func TestAPIMainDoesNotStartBusinessSchedulers(t *testing.T) {
	source, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatalf("read main.go: %v", err)
	}
	text := string(source)
	banned := []string{
		"StartSuperAlertasWorker(",
		"StartSuperVPSSnapshotWorker(",
		"StartSuperMantenimientoAgentesWorker(",
		"StartEmpresaCobranzaRecordatoriosWorker(",
		"StartLicenciaVencimientoAlertasWorker(",
		"StartControlElectricoProgramacionWorker(",
	}
	for _, call := range banned {
		if strings.Contains(text, call) {
			t.Errorf("API main starts business scheduler %s; register it in pcs-worker", call)
		}
	}
}
