package handlers

import (
	"net/http"
	"net/http/httptest"
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
	for _, marker := range []string{
		`cfg.produccion_local_activa === true`,
		`Number(cfg.produccion_local_activa || 0) === 1`,
		`lowerValue(cfg.estado_dian || cfg.estado) === "produccion_local_activa"`,
	} {
		if !strings.Contains(page, marker) {
			t.Fatalf("DIAN progress must require persistent local-production activation; missing %q", marker)
		}
	}
	if strings.Contains(page, `return lowerValue(cfg.tipo_ambiente) === "produccion" || lowerValue(cfg.estado_dian || cfg.estado) === "produccion_local_activa";`) {
		t.Fatal("DIAN progress still treats a production transport environment as activation")
	}
}

func TestDIANConfigSaveCannotSelfActivateProduction(t *testing.T) {
	payload := map[string]interface{}{
		"tipo_ambiente":           "produccion",
		"produccion_local_activa": true,
		"estado_dian":             "produccion_local_activa",
	}
	prepareDIANConfigSaveActivation(payload)
	if _, exists := payload["produccion_local_activa"]; exists {
		t.Fatal("ordinary DIAN config save must not self-activate local production")
	}
	if _, exists := payload["estado_dian"]; exists {
		t.Fatal("ordinary DIAN config save must not forge the operational DIAN state")
	}

	payload = map[string]interface{}{
		"tipo_ambiente":           "habilitacion",
		"produccion_local_activa": true,
	}
	prepareDIANConfigSaveActivation(payload)
	if got := anyToInt64(payload["produccion_local_activa"]); got != 0 {
		t.Fatalf("returning to habilitation must clear local production, got %d", got)
	}

	payload = map[string]interface{}{"razon_social": "Edicion parcial"}
	prepareDIANConfigSaveActivation(payload)
	if _, exists := payload["produccion_local_activa"]; exists {
		t.Fatal("partial DIAN config save must preserve the existing production activation")
	}
}

func TestDIANUnknownMutationActionCannotBypassConfigGuard(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/empresa/facturacion_electronica/dian?action=crud_directo&empresa_id=12", strings.NewReader(`{"produccion_local_activa":true}`))
	rec := httptest.NewRecorder()
	EmpresaDIANColombiaHandler(nil, nil).ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("unknown DIAN mutation action status=%d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestDIANProductionActivationUsesDedicatedPersistentColumn(t *testing.T) {
	schemaPath := filepath.Join("..", "db", "modulos_faltantes.go")
	schemaRaw, err := os.ReadFile(schemaPath)
	if err != nil {
		t.Fatalf("read DIAN schema: %v", err)
	}
	schema := string(schemaRaw)
	for _, marker := range []string{
		`produccion_local_activa INTEGER DEFAULT 0`,
		`OR LOWER(COALESCE(observaciones, '')) LIKE '%por dian_activar_produccion_local%'`,
	} {
		if !strings.Contains(schema, marker) {
			t.Fatalf("DIAN schema is missing persistent activation marker %q", marker)
		}
	}

	handlerRaw, err := os.ReadFile("modulos_faltantes.go")
	if err != nil {
		t.Fatalf("read DIAN handler: %v", err)
	}
	handler := string(handlerRaw)
	for _, marker := range []string{
		`"produccion_local_activa": 1`,
		`if parseTruthy(genericStringValue(cfg["produccion_local_activa"]))`,
	} {
		if !strings.Contains(handler, marker) {
			t.Fatalf("DIAN activation must persist independently from send state; missing %q", marker)
		}
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
		`anulacion_confirmada_dian`,
		`DIAN no confirmó la anulación`,
		`data-action="artefactos"`,
		`function downloadFiscalArtifacts(item)`,
		`action", "artefactos"`,
	} {
		if !strings.Contains(page, marker) {
			t.Fatalf("DIAN retry UI missing marker %q", marker)
		}
	}
}

func TestFacturasElectronicasNoCuentaComprobanteComoFacturaFiscal(t *testing.T) {
	path := filepath.Join("..", "..", "web", "administrar_empresa", "facturas_electronicas.html")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read invoices page: %v", err)
	}
	page := string(raw)
	if !strings.Contains(page, `if (facturaAsociadaParaVenta(item)) return "Venta (factura asociada)";`) {
		t.Fatal("linked sale must remain visibly distinct from its fiscal invoice")
	}
	if strings.Contains(page, `if (facturaAsociadaParaVenta(item)) tipo = "factura_electronica";`) {
		t.Fatal("internal sale is still double-counted as an electronic invoice")
	}
}

func TestDIANFrontendDisablesFamiliesWithoutDedicatedAdapter(t *testing.T) {
	path := filepath.Join("..", "..", "web", "administrar_empresa", "facturacion_electronica_pruebas_dian.html")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read DIAN page: %v", err)
	}
	page := string(raw)
	for _, marker := range []string{
		`value="documento_soporte" disabled`,
		`value="nomina_electronica" disabled`,
		`value="documento_equivalente_pos" disabled`,
		`value="eventos_radian_recepcion" disabled`,
		`codigo: "tipo_documento_dian_no_implementado"`,
	} {
		if !strings.Contains(page, marker) {
			t.Fatalf("DIAN unsupported-family guard missing %q", marker)
		}
	}
}

func TestContabilidadManualFormsDoNotOfferForgedDIANStates(t *testing.T) {
	path := filepath.Join("..", "..", "web", "administrar_empresa", "contabilidad_colombia_avanzada.html")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read Colombia accounting page: %v", err)
	}
	page := string(raw)
	for _, marker := range []string{
		`value="Borrador local - sin envío DIAN" readonly`,
		`CUNE y aceptación DIAN solo se registrarán desde el adaptador técnico`,
		`Un CUDS, firma o aceptación DIAN no se puede declarar manualmente`,
	} {
		if !strings.Contains(page, marker) {
			t.Fatalf("manual accounting UI missing fiscal-safety marker %q", marker)
		}
	}
	if strings.Contains(page, `estado_dian:els.nomEstado.value`) || strings.Contains(page, `estado_dian:els.dsEstado.value`) {
		t.Fatal("manual accounting form can still send a DIAN state")
	}
}

func TestContabilidadDocumentoSoportePreflightIsVisibleAndNonEmitting(t *testing.T) {
	path := filepath.Join("..", "..", "web", "administrar_empresa", "contabilidad_colombia_avanzada.html")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read Colombia accounting page: %v", err)
	}
	page := string(raw)
	for _, marker := range []string{
		`id="soportePreflight"`,
		`data-ds-preflight`,
		`documento_soporte_preflight`,
		`La revisión no genera XML, no consume consecutivo y no envía información.`,
		`No se emitió ni transmitió ningún documento.`,
	} {
		if !strings.Contains(page, marker) {
			t.Fatalf("document-support preflight UI missing safety marker %q", marker)
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

func TestFacturasElectronicasPrintUsesAbsoluteSameOriginAssets(t *testing.T) {
	path := filepath.Join("..", "..", "web", "administrar_empresa", "facturas_electronicas.html")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read invoices page: %v", err)
	}
	page := string(raw)
	for _, marker := range []string{
		`return window.location.origin.replace(/\/$/, "") + v;`,
		`escapeHtml(window.location.origin.replace(/\/$/, "") + "/estilos.css")`,
	} {
		if !strings.Contains(page, marker) {
			t.Fatalf("blob print assets must use an absolute same-origin URL; missing %q", marker)
		}
	}

	sharedPath := filepath.Join("..", "..", "web", "js", "print_documents.js")
	sharedRaw, err := os.ReadFile(sharedPath)
	if err != nil {
		t.Fatalf("read shared print renderer: %v", err)
	}
	shared := string(sharedRaw)
	for _, marker := range []string{`function absoluteSameOriginAssetURL(raw)`, `return absoluteSameOriginAssetURL(raw);`} {
		if !strings.Contains(shared, marker) {
			t.Fatalf("shared print assets must resolve root URLs before opening a blob; missing %q", marker)
		}
	}
}

func TestFacturacionElectronicaMobileContainsWideContent(t *testing.T) {
	path := filepath.Join("..", "..", "web", "administrar_empresa", "facturacion_electronica.html")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read electronic invoicing configuration page: %v", err)
	}
	page := string(raw)
	for _, marker := range []string{
		`#list{max-width:100%;overflow-x:auto;-webkit-overflow-scrolling:touch}`,
		`#list .table{min-width:556px}`,
		`.fe-firma-last-upload{max-width:100%;overflow-wrap:anywhere;word-break:break-word}`,
	} {
		if !strings.Contains(page, marker) {
			t.Fatalf("mobile DIAN layout is missing containment marker %q", marker)
		}
	}
}
