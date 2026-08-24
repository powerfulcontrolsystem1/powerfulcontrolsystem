package handlers

import (
	"bytes"
	"strings"
	"testing"

	dbpkg "github.com/you/pos-backend/db"
)

func TestFacturacionSafeDispatchJSONRemovesRawFiscalPayload(t *testing.T) {
	raw := facturacionSafeDispatchJSON(map[string]interface{}{
		"ok": true, "track_id": "abc", "raw_response": "<secret/>",
		"request_resumen": map[string]interface{}{"xml_firmado": "<Invoice/>"},
	})
	if strings.Contains(raw, "secret") || strings.Contains(raw, "xml_firmado") || !strings.Contains(raw, "track_id") {
		t.Fatalf("unsafe fiscal response summary: %s", raw)
	}
}

func TestFacturaElectronicaRepresentationPDFIsRealPDF(t *testing.T) {
	pdf := buildFacturaElectronicaRepresentationPDF(dbpkg.EmpresaDocumentoFacturacion{
		EmpresaID: 12, TipoDocumento: "factura_electronica", DocumentoCodigo: "FAC-1", NumeroLegal: "1PCS4",
		CodigoValidacion: "CUFE123", MontoTotal: 100, Moneda: "COP", FechaDocumento: "2026-08-21",
	}, facturacionOperacionPayload{ClienteNombre: "Cliente", ClienteNumeroDocumento: "123"})
	if !bytes.HasPrefix(pdf, []byte("%PDF-1.4")) || !bytes.Contains(pdf, []byte("1PCS4")) || !bytes.Contains(pdf, []byte("CUFE123")) {
		t.Fatal("fiscal representation must be a readable PDF containing legal identifiers")
	}
}

func TestNormalizeFacturacionDocumentoElectronicoTipoIncluyeDocumentosSiigoDian(t *testing.T) {
	cases := map[string]string{
		"factura":                         "factura_electronica",
		"nota debito ventas":              "nota_debito",
		"documento soporte adquisiciones": "documento_soporte",
		"documento soporte de pago nomina electronica": "nomina_electronica",
		"tiquete maquina registradora pos":             "documento_equivalente_pos",
		"documento equivalente electronico POS":        "documento_equivalente_pos",
		"nota credito":                                 "nota_credito",
		"nota ajuste documento soporte":                "nota_ajuste_documento_soporte",
		"factura papel contingencia":                   "factura_talonario_contingencia",
		"eventos RADIAN":                               "eventos_radian_recepcion",
	}
	for raw, want := range cases {
		if got := normalizeFacturacionDocumentoElectronicoTipo(raw); got != want {
			t.Fatalf("normalizeFacturacionDocumentoElectronicoTipo(%q)=%q, want %q", raw, got, want)
		}
	}
}

func TestDIANUBLVentaNoConvierteFamiliasDistintasEnFactura(t *testing.T) {
	for _, tipo := range []string{"documento_soporte", "nomina_electronica", "documento_equivalente_pos", "eventos_radian_recepcion"} {
		t.Run(tipo, func(t *testing.T) {
			if facturacionDocumentoElectronicoDIANUBLVentaSoportado(tipo) {
				t.Fatalf("%s no pertenece al anexo UBL de factura de venta", tipo)
			}
			root, _, _, _, _, _, _, _ := dianDocumentKind(tipo)
			if root != "" {
				t.Fatalf("%s cayo silenciosamente en raiz %s", tipo, root)
			}
			_, status, err := generateDIANUBLBase(map[string]interface{}{"pais_codigo": "CO"}, 12, map[string]interface{}{
				"documento_tipo":   tipo,
				"documento_codigo": "NO-ENVIAR-1",
			})
			if err == nil || status != 422 || !strings.Contains(err.Error(), "no se envio informacion fiscal") {
				t.Fatalf("se esperaba bloqueo 422 para %s, status=%d err=%v", tipo, status, err)
			}
		})
	}
}

func TestDIANUBLVentaConservaTiposImplementados(t *testing.T) {
	wantRoots := map[string]string{
		"factura_electronica": "Invoice",
		"nota_credito":        "CreditNote",
		"nota_debito":         "DebitNote",
	}
	for tipo, wantRoot := range wantRoots {
		if !facturacionDocumentoElectronicoDIANUBLVentaSoportado(tipo) {
			t.Fatalf("%s debe permanecer operativo", tipo)
		}
		root, _, _, _, _, _, _, _ := dianDocumentKind(tipo)
		if root != wantRoot {
			t.Fatalf("dianDocumentKind(%s)=%s, want %s", tipo, root, wantRoot)
		}
	}
}

func TestConfiguracionDIANPorDocumentoNoHabilitaFamiliasSinAdaptador(t *testing.T) {
	if err := facturacionValidarConfiguracionDIANDocumento("documento_soporte", "configurando"); err != nil {
		t.Fatalf("documento soporte debe permitir preparacion separada: %v", err)
	}
	if err := facturacionValidarConfiguracionDIANDocumento("documento_soporte", "activo"); err == nil {
		t.Fatal("documento soporte no debe activarse sin adaptador propio")
	}
	if err := facturacionValidarConfiguracionDIANDocumento("factura_electronica", "configurando"); err == nil {
		t.Fatal("factura no debe duplicar su configuracion DIAN existente")
	}
}

func TestFacturacionFiltrarDocumentosDianOperativosSaneaConfiguracionLegacy(t *testing.T) {
	got := facturacionFiltrarDocumentosDianOperativos([]string{
		"factura_electronica", "documento_soporte", "nota_credito", "nomina_electronica", "nota_debito", "factura_electronica", "eventos_radian_recepcion",
	})
	want := []string{"factura_electronica", "nota_credito", "nota_debito"}
	if len(got) != len(want) {
		t.Fatalf("documentos operativos=%v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("documentos operativos=%v, want %v", got, want)
		}
	}
}

func TestBuildDIANCUDSDocumentoSoporteCoincideConEjemploOficial(t *testing.T) {
	// Vector tomado del ejemplo público DIAN de operación con residente. No se
	// usan credenciales ni datos de una empresa PCS.
	got := buildDIANCUDSDocumentoSoporte(
		"DS236000000", "2022-02-18", "13:34:59-05:00",
		"3899000.00", "322430.00", "4152176.00",
		"1020", "800197268", "12345", "2",
	)
	const want = "C96A728F4453822BFC69B94253880D21D29DD1A9424444DA07610799C203506D33FA4F16830DBD6EE0FEBB4711BFA23A"
	if got != want {
		t.Fatalf("CUDS=%s, want %s", got, want)
	}
}

func TestResolveFacturacionTransitionForDocumentosElectronicosNuevos(t *testing.T) {
	cases := []struct {
		name          string
		action        string
		tipoDocumento string
		wantAccion    string
		wantEstado    string
		wantEvento    string
	}{
		{name: "nota debito", action: "nota_debito", tipoDocumento: "nota_debito", wantAccion: "nota_debito", wantEstado: "emitida", wantEvento: "nota_debito_emitida"},
		{name: "documento soporte", action: "documento_soporte", tipoDocumento: "documento_soporte", wantAccion: "documento_soporte", wantEstado: "emitida", wantEvento: "documento_soporte_emitido"},
		{name: "nomina electronica", action: "nomina_electronica", tipoDocumento: "nomina_electronica", wantAccion: "nomina_electronica", wantEstado: "emitida", wantEvento: "nomina_electronica_emitida"},
		{name: "pos electronico", action: "documento_equivalente_pos", tipoDocumento: "documento_equivalente_pos", wantAccion: "documento_equivalente_pos", wantEstado: "emitida", wantEvento: "documento_equivalente_pos_emitido"},
		{name: "eventos radian", action: "eventos_radian_recepcion", tipoDocumento: "eventos_radian_recepcion", wantAccion: "eventos_radian_recepcion", wantEstado: "emitida", wantEvento: "eventos_radian_recepcion_emitido"},
		{name: "nota ajuste soporte", action: "emitir_nota_ajuste_documento_soporte", tipoDocumento: "nota_ajuste_documento_soporte", wantAccion: "nota_ajuste_documento_soporte", wantEstado: "emitida", wantEvento: "nota_ajuste_documento_soporte_emitido"},
		{name: "documento equivalente peajes", action: "documento_equivalente_peajes", tipoDocumento: "documento_equivalente_peajes", wantAccion: "documento_equivalente_peajes", wantEstado: "emitida", wantEvento: "documento_equivalente_peajes_emitido"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := resolveFacturacionTransitionForDocument(tc.action, "borrador", tc.tipoDocumento)
			if err != nil {
				t.Fatalf("resolve transition returned error: %v", err)
			}
			if got.Accion != tc.wantAccion || got.EstadoNuevo != tc.wantEstado || got.Evento != tc.wantEvento {
				t.Fatalf("unexpected transition: %#v", got)
			}
		})
	}
}

func TestFacturaElectronicaVentaRequiereAcuseFiscalSoloColombiaProduccion(t *testing.T) {
	doc := &dbpkg.EmpresaDocumentoFacturacion{
		TipoDocumento: "factura_electronica",
		PaisCodigo:    "CO",
		AmbienteFE:    "produccion",
	}
	resultado := facturacionIntegracionResultado{Aplica: true, EstadoEnvio: "fallido"}
	if !facturaElectronicaVentaRequiereAcuseFiscal(doc, resultado) {
		t.Fatalf("expected Colombia production invoice to require fiscal acknowledgment")
	}
	if facturaElectronicaVentaIntegracionConfirmada(resultado) {
		t.Fatalf("failed fiscal integration must not be treated as confirmed")
	}

	resultado.EstadoEnvio = "enviado"
	if facturaElectronicaVentaIntegracionConfirmada(resultado) {
		t.Fatalf("sent without final acceptance must not be treated as confirmed")
	}
	resultado.EstadoEnvio = "aceptado"
	if !facturaElectronicaVentaIntegracionConfirmada(resultado) {
		t.Fatalf("accepted fiscal integration must be treated as confirmed")
	}

	doc.AmbienteFE = "habilitacion"
	if facturaElectronicaVentaRequiereAcuseFiscal(doc, resultado) {
		t.Fatalf("sandbox/habilitacion invoice must not require production fiscal acknowledgment")
	}
}

func TestDIANUserVisibleErrorRedactsPrivateSignaturePath(t *testing.T) {
	t.Parallel()
	raw := "firmar XML DIAN: open ../web/uploads/empresas/empresa_12/firma_privada.pem: permission denied"
	got := dianUserVisibleError(raw)
	if got != "No se pudo acceder a la clave privada de firma del certificado DIAN." {
		t.Fatalf("visible error = %q", got)
	}
	if dianErrorUserHelp(raw) == "" {
		t.Fatal("permission failure must include a remediation path")
	}
}

func TestFacturacionColombiaProduccionBloqueaProveedorManual(t *testing.T) {
	cfg := &dbpkg.FacturacionElectronicaPaisConfig{
		EmpresaID:  1,
		PaisCodigo: "CO",
		Ambiente:   "produccion",
		Proveedor:  "manual",
		Estado:     "activo",
	}
	status := facturacionProveedorConnectionStatus(cfg)
	if online, _ := status["online"].(bool); online {
		t.Fatalf("manual provider must not be online for Colombia production: %#v", status)
	}
	if got := status["estado_conexion"]; got != "sin_proveedor_dian" {
		t.Fatalf("unexpected connection state: %#v", status)
	}

	result := dispatchFacturacionProveedor(nil, cfg, facturacionOperacionPayload{PaisCodigo: "CO"}, dbpkg.EmpresaDocumentoFacturacion{
		EmpresaID:       1,
		TipoDocumento:   "factura_electronica",
		DocumentoCodigo: "FV-TEST",
		PaisCodigo:      "CO",
		AmbienteFE:      "produccion",
	}, "emitir")
	if result.Success {
		t.Fatalf("manual provider must not dispatch as success for Colombia production")
	}
}

func TestAnulacionElectronicaSoloConfirmaConNotaCreditoAceptada(t *testing.T) {
	if facturacionIntegracionAceptada(facturacionIntegracionResultado{EstadoEnvio: "enviado"}) {
		t.Fatal("una nota credito solamente enviada no debe anular la factura original")
	}
	if facturacionIntegracionAceptada(facturacionIntegracionResultado{EstadoEnvio: "fallido"}) {
		t.Fatal("una nota credito fallida no debe anular la factura original")
	}
	if !facturacionIntegracionAceptada(facturacionIntegracionResultado{EstadoEnvio: "aceptado"}) {
		t.Fatal("la aceptacion DIAN debe permitir finalizar la anulacion")
	}

	observaciones := facturacionNotaCreditoFacturaOrigenMarker + "FV-PCS-1\nAnulacion total autorizada"
	if got := facturacionNotaCreditoFacturaOrigen(observaciones); got != "FV-PCS-1" {
		t.Fatalf("factura origen = %q", got)
	}
	if got := facturacionNotaCreditoFacturaOrigen("texto sin marcador"); got != "" {
		t.Fatalf("no debe inferirse una factura origen sin marcador, obtuvo %q", got)
	}
	if _, err := resolveFacturacionTransitionForDocument("anular", "emitida", "factura_electronica"); err == nil {
		t.Fatal("la transicion local generica no debe anular una factura electronica")
	}
}

func TestFacturacionDIANFechaEmisionDesdeXMLConservaInstanteFirmado(t *testing.T) {
	xml := `<Invoice xmlns:cbc="urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2"><cbc:IssueDate>2026-08-21</cbc:IssueDate><cbc:IssueTime>09:31:45-05:00</cbc:IssueTime></Invoice>`
	if got := facturacionDIANFechaEmisionDesdeXML(xml); got != "2026-08-21T09:31:45-05:00" {
		t.Fatalf("fecha fiscal firmada = %q", got)
	}
	if got := facturacionDIANFechaEmisionDesdeXML(`<Invoice><IssueDate>2026-08-21</IssueDate></Invoice>`); got != "2026-08-21" {
		t.Fatalf("fecha fiscal sin hora = %q", got)
	}
	if got := facturacionDIANFechaEmisionDesdeXML(`<Invoice>`); got != "" {
		t.Fatalf("XML invalido no debe producir fecha fiscal: %q", got)
	}
}

func TestFacturacionCUFEOficialAceptaAcuseDIANAnidado(t *testing.T) {
	cufe := strings.Repeat("aB", 48)
	got := facturacionCUFEOficialDesdeMap(map[string]interface{}{
		"codigo_validacion": strings.Repeat("f", 64),
		"respuesta_dian": map[string]interface{}{
			"xml_document_key": cufe,
		},
	})
	if got != strings.ToLower(cufe) {
		t.Fatalf("CUFE anidado = %q", got)
	}
	if got := facturacionCUFEOficialDesdeMap(map[string]interface{}{"cufe": strings.Repeat("f", 64)}); got != "" {
		t.Fatalf("un hash local de 64 caracteres no puede aceptarse como CUFE: %q", got)
	}
	if qr := facturaElectronicaDIANQRURL(dbpkg.EmpresaDocumentoFacturacion{PaisCodigo: "CO", CodigoValidacion: strings.Repeat("f", 64)}); qr != "" {
		t.Fatalf("no debe generarse QR DIAN para un codigo no oficial: %q", qr)
	}
}

func TestFacturacionDocumentoAdvisoryLockKeyAislaEmpresaYDocumento(t *testing.T) {
	base := facturacionDocumentoAdvisoryLockKey(12, "factura_electronica", "1PCS8")
	if base == 0 || base != facturacionDocumentoAdvisoryLockKey(12, "FACTURA_ELECTRONICA", "1pcs8") {
		t.Fatal("la clave documental debe ser estable y normalizada")
	}
	if base == facturacionDocumentoAdvisoryLockKey(13, "factura_electronica", "1PCS8") {
		t.Fatal("la clave documental debe aislar empresa_id")
	}
	if base == facturacionDocumentoAdvisoryLockKey(12, "factura_electronica", "1PCS9") {
		t.Fatal("la clave documental debe aislar el folio")
	}
}
