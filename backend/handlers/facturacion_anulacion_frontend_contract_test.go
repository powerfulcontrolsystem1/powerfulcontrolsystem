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

func TestFacturacionDIANProgressDoesNotTreatTransportEnvironmentAsActivation(t *testing.T) {
	path := filepath.Join("..", "..", "web", "administrar_empresa", "facturacion_electronica_pruebas_dian.html")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read DIAN page: %v", err)
	}
	page := string(raw)
	if !strings.Contains(page, `return lowerValue(cfg.estado_dian || cfg.estado) === "produccion_local_activa";`) {
		t.Fatal("DIAN progress must require the explicit local-production activation state")
	}
	if strings.Contains(page, `return lowerValue(cfg.tipo_ambiente) === "produccion" || lowerValue(cfg.estado_dian || cfg.estado) === "produccion_local_activa";`) {
		t.Fatal("DIAN progress still treats a production transport environment as activation")
	}
}

func TestFacturasElectronicasFrontendOffersSafeDIANRetry(t *testing.T) {
	path := filepath.Join("..", "..", "web", "administrar_empresa", "facturas_electronicas.html")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read invoices page: %v", err)
	}
	page := string(raw)
	for _, marker := range []string{
		`data-action="reenviar_dian"`,
		`function retryInvoiceDIAN(item)`,
		`action=reenviar_dian&empresa_id=`,
		`return runSearch();`,
		`estadoDoc === "pendiente_emision" || estadoDoc === "fallida" || estadoDoc === "rechazada"`,
		`estadoDoc === "anulada"`,
	} {
		if !strings.Contains(page, marker) {
			t.Fatalf("DIAN retry UI missing marker %q", marker)
		}
	}
}

func TestFacturasElectronicasExportButtonsHaveAccessibleLabels(t *testing.T) {
	path := filepath.Join("..", "..", "web", "administrar_empresa", "facturas_electronicas.html")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read invoices page: %v", err)
	}
	page := string(raw)
	for _, marker := range []string{
		`id="btnExportCsv" type="button" class="btn secondary" aria-label="Exportar resultados en CSV"`,
		`id="btnExportExcel" type="button" class="btn secondary" aria-label="Exportar resultados en Excel"`,
	} {
		if !strings.Contains(page, marker) {
			t.Fatalf("invoice export control missing accessible label %q", marker)
		}
	}
}
