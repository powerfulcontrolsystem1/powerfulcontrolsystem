package handlers

import (
	"os"
	"strings"
	"testing"
)

func TestFacturacionRetryProcessorClaimsBeforeSending(t *testing.T) {
	raw, err := os.ReadFile("facturacion_electronica.go")
	if err != nil {
		t.Fatal(err)
	}
	source := string(raw)
	if !strings.Contains(source, "ClaimFacturacionElectronicaRetriesByEmpresa") {
		t.Fatal("el procesador FE debe reclamar filas antes de enviarlas")
	}
	processorStart := strings.Index(source, "func processFacturacionRetryQueue")
	if processorStart < 0 {
		t.Fatal("no se encontro el procesador FE")
	}
	processorEnd := strings.Index(source[processorStart:], "func buildFacturacionReconciliacion")
	if processorEnd < 0 {
		t.Fatal("no se encontro el final del procesador FE")
	}
	processor := source[processorStart : processorStart+processorEnd]
	if strings.Contains(processor, "ListFacturacionElectronicaRetriesByEmpresa") {
		t.Fatal("el procesador FE aun lista sin claim atomico")
	}
	if count := strings.Count(source, "ReleaseFacturacionElectronicaRetryClaim"); count < 3 {
		t.Fatalf("no se liberan todos los caminos del procesador FE: %d", count)
	}
}
