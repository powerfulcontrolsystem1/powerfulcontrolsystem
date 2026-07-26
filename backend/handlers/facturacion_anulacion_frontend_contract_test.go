package handlers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFacturacionDIANFrontendUsesCreditNoteCancellation(t *testing.T) {
	path := filepath.Join("..", "..", "web", "administrar_empresa", "facturacion_electronica_pruebas_dian.html")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read DIAN page: %v", err)
	}
	page := string(raw)
	required := []string{
		`id="op_observaciones"`,
		`anular_factura_nota_credito`,
		`La factura solo quedara anulada cuando DIAN acepte la nota.`,
		`addEventListener("click", ejecutarAnulacionFacturacion)`,
	}
	for _, marker := range required {
		if !strings.Contains(page, marker) {
			t.Fatalf("DIAN cancellation UI missing marker %q", marker)
		}
	}
	if strings.Contains(page, `btnAnularDocumento").addEventListener("click", function() { ejecutarOperacionFacturacion("anular")`) {
		t.Fatal("invoice cancellation still calls the generic local transition")
	}
}
