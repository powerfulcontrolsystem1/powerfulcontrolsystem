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

func TestFacturacionDIANFrontendUsesSanitizedCredentialPresence(t *testing.T) {
	path := filepath.Join("..", "..", "web", "administrar_empresa", "facturacion_electronica_pruebas_dian.html")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read DIAN page: %v", err)
	}
	page := string(raw)
	for _, marker := range []string{
		`function dianFieldConfigured(cfg, field)`,
		`field + "_configurado"`,
		`dianFieldConfigured(cfg, "certificado_url")`,
		`dianFieldConfigured(cfg, "test_set_id")`,
		`dianConfiguredLabel(cfg, "software_id")`,
		`const historial = Array.isArray(state.dianTrackHistory)`,
		`updateDianProgress();`,
	} {
		if !strings.Contains(page, marker) {
			t.Fatalf("DIAN sanitized-state UI missing marker %q", marker)
		}
	}
	if strings.Contains(page, `Mostrar PIN y clave tecnica`) {
		t.Fatal("DIAN page must not offer to reveal write-only secrets")
	}
}

func TestCarritoSurfacesNestedElectronicInvoiceResult(t *testing.T) {
	path := filepath.Join("..", "..", "web", "administrar_empresa", "carrito_de_compras.html")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read carrito frontend: %v", err)
	}
	page := string(raw)
	for _, expected := range []string{
		"documentoVenta.factura_electronica",
		"Aviso factura electrónica:",
		"setStationPaymentPersistentMessage(successMsg, Boolean(facturaWarning))",
	} {
		if !strings.Contains(page, expected) {
			t.Fatalf("carrito must surface fiscal result and warning; missing %q", expected)
		}
	}
}

func TestDIANCatalogShowsCreditNoteTotalCancellationAsPartial(t *testing.T) {
	page := readDIANFrontendContractPage(t)
	for _, expected := range []string{
		`doc.disponible_anulacion_total === true`,
		`partial ? "Parcial"`,
	} {
		if !strings.Contains(page, expected) {
			t.Fatalf("DIAN catalog must distinguish total credit-note cancellation; missing %q", expected)
		}
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

func TestDIANConfigUpdateDoesNotReapplyCreateDefaults(t *testing.T) {
	payload := map[string]interface{}{
		"razon_social": "Edicion RUT revisada",
	}
	prepareDIANConfigSaveActivation(payload)
	prepareDIANConfigPersistenceDefaults(payload, false, "actor@example.test")

	for _, field := range []string{"estado_dian", "tipo_ambiente", "url_dian", "codigo", "usuario_creador", "estado", "resolucion_alerta_dias"} {
		if value, exists := payload[field]; exists {
			t.Fatalf("una edicion DIAN no debe inventar %s=%#v ni sobrescribir el valor persistido", field, value)
		}
	}

	createPayload := map[string]interface{}{
		"nit":          "900000000",
		"razon_social": "Empresa nueva",
	}
	prepareDIANConfigPersistenceDefaults(createPayload, true, "actor@example.test")
	if got := genericStringValue(createPayload["estado_dian"]); got != "pendiente" {
		t.Fatalf("estado inicial=%q, want pendiente", got)
	}
	if got := genericStringValue(createPayload["tipo_ambiente"]); got != "habilitacion" {
		t.Fatalf("ambiente inicial=%q, want habilitacion", got)
	}
	if got := genericStringValue(createPayload["codigo"]); !strings.HasPrefix(got, "DIAN-") {
		t.Fatalf("codigo inicial=%q, want prefijo DIAN-", got)
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

func TestDIANFrontendDisablesFreeFormAndFamiliesWithoutDedicatedAdapter(t *testing.T) {
	path := filepath.Join("..", "..", "web", "administrar_empresa", "facturacion_electronica_pruebas_dian.html")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read DIAN page: %v", err)
	}
	page := string(raw)
	for _, marker := range []string{
		`value="documento_soporte" disabled`,
		`Documento soporte (usar borrador contable estructurado)`,
		`La emisión se realiza desde el borrador contable estructurado, con preflight y confirmación`,
		`value="nomina_electronica" disabled`,
		`value="documento_equivalente_pos" disabled`,
		`value="eventos_radian_recepcion" disabled`,
		`codigo: "emision_fiscal_libre_bloqueada"`,
	} {
		if !strings.Contains(page, marker) {
			t.Fatalf("DIAN unsupported-family guard missing %q", marker)
		}
	}
}

func TestFacturasElectronicasOnlyCancelsWithOfficialCUFEAndShowsCreditNoteArtifacts(t *testing.T) {
	path := filepath.Join("..", "..", "web", "administrar_empresa", "facturas_electronicas.html")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read invoice tray: %v", err)
	}
	page := string(raw)
	for _, marker := range []string{
		`function isOfficialDIANDocumentKey(value)`,
		`!isOfficialDIANDocumentKey(item && item.codigo_validacion)`,
		`item.fuente_fiscal_disponible !== true`,
		`Documento histórico sin fuente fiscal inmutable`,
		`artifactType === "factura_electronica" || artifactType === "nota_credito"`,
	} {
		if !strings.Contains(page, marker) {
			t.Fatalf("safe invoice/credit-note UI missing %q", marker)
		}
	}
}

func TestNominaFrontendCannotCallGenericFiscalEndpoint(t *testing.T) {
	path := filepath.Join("..", "..", "web", "administrar_empresa", "nomina_sueldos.html")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read payroll page: %v", err)
	}
	page := string(raw)
	if strings.Contains(page, `/api/empresa/facturacion_electronica?action=nomina_electronica`) {
		t.Fatal("payroll UI still contains a dead generic DIAN submission path")
	}
	if !strings.Contains(page, `Adaptador DIAN de nomina electronica pendiente`) || !strings.Contains(page, `Preparar el lote no genera XML ni transmite información`) {
		t.Fatal("payroll UI must explain the fail-closed preflight-only scope")
	}
}

func TestPaidSaleFiscalSourceIsCheckedBeforeNumberReservationAndEmail(t *testing.T) {
	path := filepath.Join("carritos_compras.go")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read cart backend: %v", err)
	}
	source := string(raw)
	start := strings.Index(source, "func registrarFacturaElectronicaDesdeDocumentoVentaContext(")
	if start < 0 {
		t.Fatal("could not isolate paid-sale invoice function")
	}
	end := strings.Index(source[start:], "func registrarDocumentoVentaDesdeCarritoPagado(")
	if end < 0 {
		t.Fatal("could not isolate paid-sale invoice function")
	}
	body := source[start : start+end]
	sourceCheck := strings.Index(body, "loadFacturacionFuenteFiscalParaDocumento(")
	reserveNumber := strings.Index(body, "PrepareFacturacionDocumentoLegalContext(")
	if sourceCheck < 0 || reserveNumber < 0 || sourceCheck > reserveNumber {
		t.Fatal("immutable fiscal source must be checked before reserving a legal number")
	}
	for _, marker := range []string{
		`correoFiscalPermitido := strings.EqualFold`,
		`facturaElectronicaVentaIntegracionConfirmada(resultadoIntegracion)`,
		`correo fiscal pendiente hasta recibir aceptacion DIAN`,
	} {
		if !strings.Contains(body, marker) {
			t.Fatalf("accepted-only fiscal email gate missing %q", marker)
		}
	}
}

func TestCreditNoteCancellationIsSerializedResumableAndPendingUntilDIAN(t *testing.T) {
	path := filepath.Join("facturacion_electronica.go")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read billing backend: %v", err)
	}
	source := string(raw)
	for _, marker := range []string{
		`releaseFacturaLock`,
		`fuente_fiscal_factura_no_disponible`,
		`EstadoDocumento:      "pendiente_emision"`,
		`EventoUltimo:         "nota_credito_pendiente_dian"`,
		`nota_credito_ya_existia`,
		`facturacionNotaCreditoAceptadaParaAnulacion`,
	} {
		if !strings.Contains(source, marker) {
			t.Fatalf("credit-note cancellation hardening missing %q", marker)
		}
	}
}

func TestGenericInvoiceEmissionBlocksBeforeNumberReservation(t *testing.T) {
	path := filepath.Join("facturacion_electronica.go")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read billing backend: %v", err)
	}
	source := string(raw)
	start := strings.Index(source, `if !facturacionDocumentoElectronicoDIANCreacionGenericaSoportada(documentoTipo) {`)
	if start < 0 {
		t.Fatal("generic fiscal guard missing")
	}
	reserve := strings.Index(source[start:], `PrepareFacturacionDocumentoLegalContext(`)
	guardReturn := strings.Index(source[start:], `return`)
	if reserve < 0 || guardReturn < 0 || guardReturn > reserve {
		t.Fatal("generic invoice guard must return before legal-number reservation")
	}
	if !strings.Contains(source[start:start+reserve], `emision_factura_libre_bloqueada`) {
		t.Fatal("generic invoice guard must expose an actionable fail-closed code")
	}
}

func TestReconciliationRepairsAcceptedDocumentsWithoutResending(t *testing.T) {
	path := filepath.Join("facturacion_electronica.go")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read billing backend: %v", err)
	}
	source := string(raw)
	helperStart := strings.Index(source, "func reconciliarFacturacionAceptacionLocal(")
	start := strings.Index(source, "func reconcileFacturacionEstados(")
	if helperStart < 0 || start < 0 || helperStart > start {
		t.Fatal("accepted-repair helper or reconciliation function missing")
	}
	body := source[start:]
	accepted := strings.Index(body, `reconciliarFacturacionAceptacionLocal(`)
	sentGuard := strings.Index(body, `if estadoRetry == "enviado"`)
	dispatch := strings.Index(body, `processFacturacionIntegracionForDocumento(`)
	if accepted < 0 || sentGuard < 0 || dispatch < 0 || accepted > dispatch || sentGuard > dispatch {
		t.Fatal("accepted local repair must run before any integration dispatch")
	}
	reconciliationSource := source[helperStart:]
	for _, marker := range []string{
		`facturacionDocumentoAceptadoDIAN`,
		`facturacionRetryAceptadoConCodigoValidacion`,
		`reparaciones_locales`,
		`documento_pendiente_sin_aceptacion`,
	} {
		if !strings.Contains(reconciliationSource, marker) {
			t.Fatalf("safe accepted reconciliation missing %q", marker)
		}
	}
}

func TestAcceptedOnlyReconciliationStopsBeforeAnyDispatch(t *testing.T) {
	path := filepath.Join("facturacion_electronica.go")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read billing backend: %v", err)
	}
	source := string(raw)
	actionStart := strings.Index(source, `if action == "reconciliar_estados" || action == "reconciliar_aceptados_local"`)
	if actionStart < 0 {
		t.Fatal("accepted-only reconciliation action missing")
	}
	actionEnd := strings.Index(source[actionStart:], `if action == "facturar_desde_venta"`)
	if actionEnd < 0 {
		t.Fatal("accepted-only reconciliation action boundary missing")
	}
	actionBody := source[actionStart : actionStart+actionEnd]
	for _, marker := range []string{
		`soloAcusesAceptados := action == "reconciliar_aceptados_local"`,
		`reconcileFacturacionEstados(dbEmp, empresaID, aplicar, soloAcusesAceptados`,
	} {
		if !strings.Contains(actionBody, marker) {
			t.Fatalf("accepted-only action missing %q", marker)
		}
	}

	start := strings.Index(source, "func reconcileFacturacionEstados(")
	if start < 0 {
		t.Fatal("reconciliation function missing")
	}
	body := source[start:]
	guard := strings.Index(body, `if soloAcusesAceptados {`)
	dispatch := strings.Index(body, `processFacturacionIntegracionForDocumento(`)
	if guard < 0 || dispatch < 0 || guard > dispatch {
		t.Fatal("accepted-only guard must run before any integration dispatch")
	}
	guardBody := body[guard:dispatch]
	if !strings.Contains(guardBody, "continue") {
		t.Fatal("accepted-only guard must stop processing non-accepted documents")
	}
	for _, marker := range []string{`"solo_acuses_aceptados"`, `"transmision_xml_habilitada"`, `"omitidos_no_aceptados"`} {
		if !strings.Contains(body, marker) {
			t.Fatalf("accepted-only evidence missing %q", marker)
		}
	}
}

func TestDIANFrontendDisablesFreeFormFiscalEmission(t *testing.T) {
	path := filepath.Join("..", "..", "web", "administrar_empresa", "facturacion_electronica_pruebas_dian.html")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read DIAN page: %v", err)
	}
	page := string(raw)
	for _, marker := range []string{
		`id="btnDianEnviarFactura" type="button" class="btn secondary" disabled`,
		`id="btnDianEnviarND" type="button" class="btn secondary" disabled`,
		`id="btnDianEnviarNC" type="button" class="btn secondary" disabled`,
		`id="op_tipo_documento" class="form-input" disabled`,
		`id="btnEmitirDocumento" class="btn" disabled`,
		`id="btnAnularDocumento" class="btn secondary" disabled`,
		`La emision comercial real nace exclusivamente de una venta pagada con fuente fiscal inmutable.`,
	} {
		if !strings.Contains(page, marker) {
			t.Fatalf("free-form fiscal emission must be disabled; missing %q", marker)
		}
	}
}

func TestDIANFrontendProgressRequiresIndependentEvidence(t *testing.T) {
	path := filepath.Join("..", "..", "web", "administrar_empresa", "facturacion_electronica_pruebas_dian.html")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read DIAN page: %v", err)
	}
	page := string(raw)
	for _, forbidden := range []string{
		`const acuseReady = hasDianAcuseFinal(cfg, result) || productionReady;`,
		`const validationReady = state.dianValidationOk || acuseReady || productionReady;`,
		`done: hasDianEnvioReal(result) || acuseReady || productionReady`,
		`estado === "habilitacion_aprobada" || estado === "produccion_local_activa"`,
		`const validationReady = state.dianValidationOk || hasDianEnvioReal(result)`,
		`if (estado === "habilitacion_aprobada" || estado === "aceptado") return true;`,
	} {
		if strings.Contains(page, forbidden) {
			t.Fatalf("local production activation must not forge DIAN evidence: found %q", forbidden)
		}
	}
	if !strings.Contains(page, "Servidor DIAN/proveedor alcanzable. Esta prueba no valida SOAP, credenciales ni aceptación fiscal.") {
		t.Fatal("connection probe must distinguish reachability from DIAN validation")
	}
	if !strings.Contains(page, "El 100 % solo resume las comprobaciones ejecutadas en esta sesion") {
		t.Fatal("progress must explicitly deny production-readiness semantics")
	}
}

func TestDIANHistoricalTrackRequeryCannotMutateGlobalConfiguration(t *testing.T) {
	path := filepath.Join("modulos_faltantes.go")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read DIAN backend: %v", err)
	}
	source := string(raw)
	start := strings.Index(source, "func consultarDIANStatusZipSOAP(")
	end := strings.Index(source, "func consultarDIANNumberingRange(")
	if start < 0 || end <= start {
		t.Fatal("could not isolate consultarDIANStatusZipSOAP source")
	}
	body := source[start:end]
	if !strings.Contains(body, "upsertDIANTrackHistory(") {
		t.Fatal("GetStatusZip must persist the individual TrackId history")
	}
	for _, forbidden := range []string{
		"updateDIANConfigFields(",
		"empresa_dian_configuracion",
	} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("historical GetStatusZip must not mutate global DIAN configuration; found %q", forbidden)
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
		`El servidor recalcula los importes y elimina cualquier CUDS, respuesta o estado DIAN escrito desde el navegador.`,
		`Guardar no consume consecutivo, no genera XML y no transmite a DIAN.`,
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
	htmlPath := filepath.Join("..", "..", "web", "administrar_empresa", "contabilidad_colombia_avanzada.html")
	jsPath := filepath.Join("..", "..", "web", "js", "documento_soporte_dian.js")
	html, err := os.ReadFile(htmlPath)
	if err != nil {
		t.Fatalf("read Colombia accounting page: %v", err)
	}
	js, err := os.ReadFile(jsPath)
	if err != nil {
		t.Fatalf("read document-support frontend: %v", err)
	}
	page := string(html) + "\n" + string(js)
	for _, marker := range []string{
		`<script src="/js/documento_soporte_dian.js"></script>`,
		`id="soportePreflight"`,
		`data-ds-preflight`,
		`documento_soporte_preflight`,
		`No se generó XML, no se consumió consecutivo y no se transmitió información.`,
		`Preflight completado con bloqueos; no hubo emisión.`,
	} {
		if !strings.Contains(page, marker) {
			t.Fatalf("document-support preflight UI missing safety marker %q", marker)
		}
	}
}

func TestContabilidadDocumentoSoporteEmissionRequiresTypedConfirmation(t *testing.T) {
	htmlPath := filepath.Join("..", "..", "web", "administrar_empresa", "contabilidad_colombia_avanzada.html")
	jsPath := filepath.Join("..", "..", "web", "js", "documento_soporte_dian.js")
	html, err := os.ReadFile(htmlPath)
	if err != nil {
		t.Fatalf("read Colombia accounting page: %v", err)
	}
	js, err := os.ReadFile(jsPath)
	if err != nil {
		t.Fatalf("read document-support frontend: %v", err)
	}
	page := string(html) + "\n" + string(js)
	for _, marker := range []string{
		`var CONFIRMACION_DIAN = "EMITIR DOCUMENTO SOPORTE DIAN"`,
		`id="btnConfirmarSoporteEmision" class="btn" type="button" disabled`,
		`els.dsEmitConfirmacion.value !== CONFIRMACION_DIAN`,
		`facturacionAPI("emitir_documento_soporte")`,
		`confirmar_emision: true, mensaje_confirmacion_dian: CONFIRMACION_DIAN`,
		`out.puede_emitir ?`,
		`tipo_documento: "documento_soporte"`,
	} {
		if !strings.Contains(page, marker) {
			t.Fatalf("document-support emission guard missing %q", marker)
		}
	}
	start := strings.Index(string(js), "function draftPayload()")
	end := strings.Index(string(js), "function validateDraft(")
	if start < 0 || end <= start {
		t.Fatal("could not isolate document-support draft payload")
	}
	draft := string(js)[start:end]
	for _, forbidden := range []string{"estado_dian", "cuds:", "respuesta_dian", "numero_legal", "subtotal:", "retenciones:", "total_neto_contable:"} {
		if strings.Contains(draft, forbidden) {
			t.Fatalf("browser draft must not forge server or DIAN field %q", forbidden)
		}
	}
}

func TestContabilidadNominaPreflightIsVisibleAndNonEmitting(t *testing.T) {
	path := filepath.Join("..", "..", "web", "administrar_empresa", "contabilidad_colombia_avanzada.html")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read Colombia accounting page: %v", err)
	}
	page := string(raw)
	for _, marker := range []string{
		`id="nominaPreflight"`,
		`data-nom-preflight`,
		`nomina_electronica_preflight`,
		`La revisión no genera XML, no consume consecutivo y no envía información.`,
		`No se emitió ni transmitió ningún documento.`,
	} {
		if !strings.Contains(page, marker) {
			t.Fatalf("payroll preflight UI missing safety marker %q", marker)
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

func dianFrontendContractSection(t *testing.T, page, start, end string) string {
	t.Helper()
	startAt := strings.Index(page, start)
	if startAt < 0 {
		t.Fatalf("DIAN frontend contract is missing section %q", start)
	}
	endAt := strings.Index(page[startAt+len(start):], end)
	if endAt < 0 {
		t.Fatalf("DIAN frontend contract section %q has no end marker %q", start, end)
	}
	return page[startAt : startAt+len(start)+endAt]
}

func readDIANFrontendContractPage(t *testing.T) string {
	t.Helper()
	path := filepath.Join("..", "..", "web", "administrar_empresa", "facturacion_electronica.html")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read DIAN configuration page: %v", err)
	}
	return string(raw)
}

func TestDIANFrontendSecretsAreWriteOnlyAndUseConfiguredFlags(t *testing.T) {
	page := readDIANFrontendContractPage(t)
	for _, inputID := range []string{
		"dian_test_set_id",
		"dian_software_id",
		"dian_software_pin",
		"dian_llave_tecnica",
		"dian_token_emisor_ref",
		"dian_certificado_url",
		"dian_certificado_clave_ref",
	} {
		inputMarker := `id="` + inputID + `" class="form-input" type="password"`
		if !strings.Contains(page, inputMarker) {
			t.Errorf("sensitive DIAN input %q is not a password field", inputID)
		}
		if !strings.Contains(page, `id="`+inputID+`_estado"`) {
			t.Errorf("sensitive DIAN input %q has no visible configured-state indicator", inputID)
		}
	}

	setSection := dianFrontendContractSection(t, page, "function setDianConfigFormData(cfg)", "function collectDianConfigPayload()")
	for _, forbidden := range []string{
		"cfg.test_set_id",
		"cfg.software_id",
		"cfg.software_pin",
		"cfg.llave_tecnica",
		"cfg.token_emisor_ref",
		"cfg.certificado_url",
		"cfg.certificado_clave_ref",
	} {
		if strings.Contains(setSection, forbidden) {
			t.Errorf("DIAN frontend must not rehydrate a secret from API config: found %q", forbidden)
		}
	}
	if !strings.Contains(setSection, "resetDianSecretInputs(cfg);") {
		t.Error("DIAN frontend does not reset write-only secret fields after loading config")
	}
	if !strings.Contains(setSection, "cfg = dianConfigForUI(cfg);") {
		t.Error("DIAN frontend does not discard raw secrets before retaining API config in UI state")
	}
	if !strings.Contains(page, `field + "_configurado"`) {
		t.Error("DIAN frontend does not consume the server *_configurado indicators")
	}
	for _, hiddenResultKey := range []string{"software_id", "software_pin", "test_set_id", "llave_tecnica", "certificado_url", "certificado_clave"} {
		if !strings.Contains(page, "|"+hiddenResultKey) {
			t.Errorf("DIAN visible-result sanitizer does not hide %q", hiddenResultKey)
		}
	}
}

func TestDIANFrontendOmitsUnchangedSecretsAndCertificateReferences(t *testing.T) {
	page := readDIANFrontendContractPage(t)
	collectSection := dianFrontendContractSection(t, page, "function collectDianConfigPayload()", "async function loadDianConfig()")
	for _, certificateField := range []string{"certificado_url", "certificado_clave_ref"} {
		if strings.Contains(collectSection, certificateField) {
			t.Errorf("DIAN configuration save must never submit upload-managed field %q", certificateField)
		}
	}
	for _, secretField := range []string{"software_id", "software_pin", "test_set_id", "llave_tecnica", "token_emisor_ref"} {
		if !strings.Contains(collectSection, `{ field: "`+secretField+`"`) {
			t.Errorf("DIAN configuration save is missing write-only field %q", secretField)
		}
	}
	if !strings.Contains(collectSection, "if (value) payload[secret.field] = value;") {
		t.Error("DIAN configuration save does not omit empty write-only secret values")
	}
	if !strings.Contains(page, "state.dianConfig.certificado_clave_ref_configurado") {
		t.Error("signature replacement confirmation is not based on the server configured flag")
	}
}
