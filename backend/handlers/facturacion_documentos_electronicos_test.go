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

func TestDecodeFacturacionOperacionPayloadIsStrict(t *testing.T) {
	var payload facturacionOperacionPayload
	if err := decodeFacturacionOperacionPayload(strings.NewReader(`{"empresa_id":12,"tipo_documento":"factura_electronica"}`), &payload); err != nil {
		t.Fatalf("payload válido rechazado: %v", err)
	}
	if payload.EmpresaID != 12 || payload.TipoDocumento != "factura_electronica" {
		t.Fatalf("payload válido no se decodificó: %+v", payload)
	}
	if err := decodeFacturacionOperacionPayload(strings.NewReader(`{"empresa_id":12,"campo_desconocido":true}`), &payload); err == nil {
		t.Fatal("un campo desconocido debe ser rechazado")
	}
	if err := decodeFacturacionOperacionPayload(strings.NewReader(`{"empresa_id":12}{"empresa_id":13}`), &payload); err == nil {
		t.Fatal("dos objetos JSON deben ser rechazados")
	}
	if err := decodeFacturacionOperacionPayload(strings.NewReader(""), &payload); err != nil {
		t.Fatalf("el cuerpo vacío con parámetros de consulta debe seguir permitido: %v", err)
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
	for _, tipo := range []string{"nomina_electronica", "documento_equivalente_pos", "eventos_radian_recepcion"} {
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

func TestDocumentoSoporteUsaAdaptadorPropioYNoElUBLDeVenta(t *testing.T) {
	if facturacionDocumentoElectronicoDIANUBLVentaSoportado("documento_soporte") {
		t.Fatal("documento soporte no puede usar el generador UBL comercial de factura")
	}
	if !facturacionDocumentoElectronicoDIANTransporteSoportado("documento_soporte") || !facturacionDocumentoElectronicoDIANComercialSoportado("documento_soporte") {
		t.Fatal("documento soporte debe llegar a su adaptador comercial dedicado")
	}
	root, line, customization, scheme, typeCode, total, _, _ := dianDocumentKind("documento_soporte")
	if root != "DocumentoSoporte" || line != "InvoiceLine" || customization != "10" || scheme != "CUDS-SHA384" || typeCode != "05" || total != "LegalMonetaryTotal" {
		t.Fatalf("contrato DIAN de documento soporte inesperado: %q %q %q %q %q %q", root, line, customization, scheme, typeCode, total)
	}
	_, status, err := generateDIANUBLBase(map[string]interface{}{"pais_codigo": "CO"}, 12, map[string]interface{}{
		"documento_tipo": "documento_soporte", "documento_codigo": "NO-ENVIAR-1",
	}, nil)
	if err == nil || status != 422 || !strings.Contains(err.Error(), "fuente fiscal") {
		t.Fatalf("el adaptador debe exigir fuente fiscal inmutable, status=%d err=%v", status, err)
	}
}

func TestDocumentoSoporteDetectaXMLFirmadoLegacyConRaizInvoice(t *testing.T) {
	legacy := `<Invoice xmlns:cac="urn:test"><ProfileID>DIAN 2.1: Factura Electrónica de Venta</ProfileID><cac:InvoiceLine/></Invoice>`
	if !facturacionDIANLegacySignedXMLNeedsManualRegeneration(legacy, "documento_soporte") {
		t.Fatal("un XML Invoice con perfil de factura no debe tratarse como documento soporte vigente")
	}
	current := `<Invoice xmlns:cac="urn:test"><ProfileID>` + dianDocumentProfileID("DocumentoSoporte") + `</ProfileID><cac:StandardItemIdentification></cac:StandardItemIdentification></Invoice>`
	if facturacionDIANLegacySignedXMLNeedsManualRegeneration(current, "documento_soporte") {
		t.Fatal("un documento soporte con perfil y detalle fiscal vigentes no debe regenerarse")
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

func TestNotaCreditoComercialOnlyUsesTotalCancellationWorkflow(t *testing.T) {
	if !facturacionDocumentoElectronicoDIANComercialSoportado("nota_credito") {
		t.Fatal("nota credito total con fuente de ajuste debe llegar al dispatcher DIAN")
	}
	if facturacionDocumentoElectronicoDIANCreacionGenericaSoportada("nota_credito") {
		t.Fatal("el endpoint generico no debe fabricar notas credito sin factura aceptada")
	}
	if facturacionDocumentoElectronicoDIANCreacionGenericaSoportada("factura_electronica") {
		t.Fatal("la factura libre no puede reservar numeracion; debe nacer de una venta pagada")
	}
	if facturacionDocumentoElectronicoDIANComercialSoportado("nota_debito") {
		t.Fatal("nota debito debe seguir bloqueada hasta tener ajuste estructurado real")
	}
}

func TestConfiguracionDIANPorDocumentoHabilitaSoloFamiliasConAdaptador(t *testing.T) {
	if err := facturacionValidarConfiguracionDIANDocumento("documento_soporte", "configurando"); err != nil {
		t.Fatalf("documento soporte debe permitir preparacion separada: %v", err)
	}
	if err := facturacionValidarConfiguracionDIANDocumento("documento_soporte", "activo"); err != nil {
		t.Fatalf("documento soporte debe permitir activacion con su adaptador propio: %v", err)
	}
	if err := facturacionValidarConfiguracionDIANDocumento("nomina_electronica", "activo"); err == nil {
		t.Fatal("nomina debe seguir bloqueada sin adaptador propio")
	}
	if err := facturacionValidarConfiguracionDIANDocumento("factura_electronica", "configurando"); err == nil {
		t.Fatal("factura no debe duplicar su configuracion DIAN existente")
	}
}

func TestFacturacionFiltrarDocumentosDianOperativosSaneaConfiguracionLegacy(t *testing.T) {
	got := facturacionFiltrarDocumentosDianOperativos([]string{
		"factura_electronica", "documento_soporte", "nota_credito", "nomina_electronica", "nota_debito", "factura_electronica", "eventos_radian_recepcion",
	})
	want := []string{"factura_electronica", "documento_soporte"}
	if len(got) != len(want) {
		t.Fatalf("documentos operativos=%v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("documentos operativos=%v, want %v", got, want)
		}
	}
}

func TestFacturacionMarcarDisponibilidadFuenteFiscalNoHabilitaLegados(t *testing.T) {
	items := []dbpkg.EmpresaDocumentoFacturacionListado{
		{EmpresaDocumentoFacturacion: dbpkg.EmpresaDocumentoFacturacion{TipoDocumento: "factura_electronica", DocumentoCodigo: "FV-VENTA-1"}},
		{EmpresaDocumentoFacturacion: dbpkg.EmpresaDocumentoFacturacion{TipoDocumento: "factura_electronica", DocumentoCodigo: "FV-LEGADA"}},
		{EmpresaDocumentoFacturacion: dbpkg.EmpresaDocumentoFacturacion{TipoDocumento: "nota_credito", DocumentoCodigo: "NC-1"}},
	}
	fuentes := []dbpkg.EmpresaFacturacionFuenteFiscalRef{
		{TipoDocumento: "comprobante_pago", DocumentoCodigo: "CP-VENTA-1"},
		{TipoDocumento: "nota_credito", DocumentoCodigo: "NC-1"},
	}
	facturacionMarcarDisponibilidadFuenteFiscal(items, fuentes)
	if !items[0].FuenteFiscalDisponible || items[1].FuenteFiscalDisponible || !items[2].FuenteFiscalDisponible {
		t.Fatalf("disponibilidad inesperada: %#v", items)
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

func TestBuildDIANCUDSDocumentoSoporteConservaIdentificacionExtranjeraAlfanumerica(t *testing.T) {
	got := buildDIANCUDSDocumentoSoporte(
		"DS1", "2026-08-26", "10:15:30-05:00", "100.00", "19.00", "119.00",
		"P123456", "800197268", "12345", "2",
	)
	want := buildDIANSHA384Hex(
		"DS1", "2026-08-26", "10:15:30-05:00", "100.00", "01", "19.00", "119.00",
		"P123456", "800197268", "12345", "2",
	)
	withoutLetter := buildDIANCUDSDocumentoSoporte(
		"DS1", "2026-08-26", "10:15:30-05:00", "100.00", "19.00", "119.00",
		"123456", "800197268", "12345", "2",
	)
	if got != want || got == withoutLetter {
		t.Fatalf("el CUDS debe conservar exactamente P123456: got=%s want=%s sin_letra=%s", got, want, withoutLetter)
	}
}

func TestDocumentoSoportePreflightListoSinTestSetYFallaConTotalesAlterados(t *testing.T) {
	documento, configuracion, empresa, principal := documentoSoportePreflightFixture()
	resultado := buildDocumentoSoporteDIANPreflight(documento, configuracion, empresa, principal)
	if !resultado.PuedeEmitir || resultado.Estado != "listo_para_emision" || len(resultado.Bloqueos) != 0 {
		t.Fatalf("preflight completo debe quedar listo sin TestSet para anexo 1.1: %+v", resultado)
	}

	documento.Total = 118
	resultado = buildDocumentoSoporteDIANPreflight(documento, configuracion, empresa, principal)
	if resultado.PuedeEmitir || !strings.Contains(strings.Join(resultado.Bloqueos, " "), "no cuadran") {
		t.Fatalf("la prevalidacion debe detectar totales inconsistentes: %+v", resultado.Bloqueos)
	}
}

func documentoSoportePreflightFixture() (*dbpkg.EmpresaDocumentoSoporteElectronico, *dbpkg.EmpresaDIANDocumentoConfiguracion, *dbpkg.EmpresaConfiguracionAvanzada, map[string]interface{}) {
	documento := &dbpkg.EmpresaDocumentoSoporteElectronico{
		ID: 7, EmpresaID: 12, ProveedorID: 1, TipoDocumento: "NIT", Documento: "1020", VendedorDigitoVerificacion: "2",
		VendedorTipoPersona: "juridica", VendedorResidencia: "residente", NombreProveedor: "Proveedor real de prueba",
		VendedorDireccion: "Calle 1 2-3", VendedorPaisCodigo: "CO", VendedorDepartamento: "Cundinamarca",
		VendedorDepartamentoCodigoDANE: "25", VendedorMunicipio: "Chia", VendedorMunicipioCodigoDANE: "25175",
		VendedorCodigoPostal: "250001", VendedorResponsabilidadTributaria: "O-23", VendedorEmail: "proveedor@example.test",
		FechaDocumento: "2026-08-26", Periodo: "2026-08", Concepto: "Servicio de prueba", Moneda: "COP",
		FormaPagoCodigo: "1", MedioPagoCodigo: "10",
		LineasJSON: `[{
			"numero":1,"codigo":"SERV-1","descripcion":"Servicio de prueba","unidad_medida":"94",
			"cantidad":1,"precio_unitario":100,"descuento_porcentaje":0,"valor_descuento":0,
			"base_gravable":100,"iva_porcentaje":19,"iva_valor":19,
			"reteiva_porcentaje":15,"reteiva_valor":2.85,"reterenta_porcentaje":11,"reterenta_valor":11,
			"subtotal_linea":100,"total_linea":119,"total_neto_contable_linea":105.15
		}]`,
		Subtotal: 100, IVA: 19, Retenciones: 13.85, Total: 119, TotalNetoContable: 105.15, EstadoDIAN: "borrador",
	}
	configuracion := &dbpkg.EmpresaDIANDocumentoConfiguracion{
		EmpresaID: 12, TipoDocumento: "documento_soporte", Estado: "habilitacion", TipoAmbiente: "habilitacion",
		Prefijo: "DS", ResolucionNumero: "18764000000001", ResolucionFechaDesde: "2026-01-01", ResolucionFechaHasta: "2027-12-31",
		RangoDesde: 1, RangoHasta: 1000, ConsecutivoActual: 1,
	}
	empresa := &dbpkg.EmpresaConfiguracionAvanzada{
		EmpresaID: 12, TipoDocumentoEmisor: "NIT", NIT: "800197268", DigitoVerificacion: "4", TipoPersonaFiscal: "juridica",
		RazonSocial: "Adquirente real SAS", ResponsabilidadTributaria: "O-15", EmailFacturacion: "facturacion@example.test",
		DireccionFiscal: "Carrera 1 2-3", PaisCodigo: "CO", Departamento: "Cundinamarca", DepartamentoCodigoDANE: "25",
		Municipio: "Chia", MunicipioCodigoDANE: "25175", CodigoPostal: "250001",
	}
	principal := map[string]interface{}{
		"nit": "800197268", "digito_verificacion": "4", "razon_social": "Adquirente real SAS", "tipo_ambiente": "habilitacion",
		"url_dian": "https://vpfe-hab.dian.gov.co/WcfDianCustomerServices.svc", "certificado_url": "secret://certificado",
		"certificado_clave_ref": "secret://clave", "software_id": "software-id-prueba", "software_pin": "12345",
	}
	return documento, configuracion, empresa, principal
}

func TestNominaElectronicaPreflightFailsClosed(t *testing.T) {
	nomina := &dbpkg.EmpresaNominaElectronica{
		ID:          8,
		EmpresaID:   12,
		Documento:   "123456",
		Nombre:      "Empleado de prueba",
		Periodo:     "2026-08",
		FechaPago:   "2026-08-23",
		EmpleadoID:  1,
		Devengados:  1000,
		Deducciones: 100,
		Total:       900,
		SalarioBase: 1000,
	}
	configuracion := &dbpkg.EmpresaDIANDocumentoConfiguracion{
		Estado:       "configurando",
		TipoAmbiente: "habilitacion",
		Prefijo:      "NE",
		RangoDesde:   1,
		RangoHasta:   10,
		TestSetID:    "test-set",
	}
	resultado := buildNominaElectronicaDIANPreflight(nomina, configuracion)
	if resultado.PuedeEmitir || resultado.Estado != "bloqueado_adaptador_dian" {
		t.Fatalf("la prevalidacion debe cerrar la emisión hasta tener adaptador propio: %+v", resultado)
	}
	if !strings.Contains(strings.Join(resultado.Bloqueos, " "), "NominaIndividual") {
		t.Fatalf("faltó el bloqueo explícito del adaptador de nómina: %+v", resultado.Bloqueos)
	}

	nomina.Total = 899
	resultado = buildNominaElectronicaDIANPreflight(nomina, configuracion)
	if !strings.Contains(strings.Join(resultado.Bloqueos, " "), "no cuadran") {
		t.Fatalf("la prevalidacion debe detectar totales inconsistentes: %+v", resultado.Bloqueos)
	}
}

func TestResolveFacturacionTransitionFailsClosedForDocumentosSinAdaptador(t *testing.T) {
	cases := []struct {
		name          string
		action        string
		tipoDocumento string
		bloqueado     bool
		wantAction    string
		wantEvent     string
	}{
		{name: "nota debito", action: "nota_debito", tipoDocumento: "nota_debito", wantAction: "nota_debito", wantEvent: "nota_debito_emitida"},
		{name: "documento soporte", action: "documento_soporte", tipoDocumento: "documento_soporte", wantAction: "documento_soporte", wantEvent: "documento_soporte_emitido"},
		{name: "nomina electronica", action: "nomina_electronica", tipoDocumento: "nomina_electronica", bloqueado: true},
		{name: "pos electronico", action: "documento_equivalente_pos", tipoDocumento: "documento_equivalente_pos", bloqueado: true},
		{name: "eventos radian", action: "eventos_radian_recepcion", tipoDocumento: "eventos_radian_recepcion", bloqueado: true},
		{name: "nota ajuste soporte", action: "emitir_nota_ajuste_documento_soporte", tipoDocumento: "nota_ajuste_documento_soporte", bloqueado: true},
		{name: "documento equivalente peajes", action: "documento_equivalente_peajes", tipoDocumento: "documento_equivalente_peajes", bloqueado: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := resolveFacturacionTransitionForDocument(tc.action, "borrador", tc.tipoDocumento)
			if tc.bloqueado {
				if err == nil || !strings.Contains(err.Error(), "no dispone aun de un adaptador DIAN") {
					t.Fatalf("se esperaba bloqueo DIAN para %s, transition=%#v err=%v", tc.tipoDocumento, got, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("resolve transition returned error: %v", err)
			}
			if got.Accion != tc.wantAction || got.EstadoNuevo != "emitida" || got.Evento != tc.wantEvent {
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
	}, "emitir", false)
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

func TestNotaCreditoAceptadaParaAnulacionExigeCUDEYColaCoincidentes(t *testing.T) {
	cude := strings.Repeat("ab", 48)
	nota := dbpkg.EmpresaDocumentoFacturacion{
		TipoDocumento: "nota_credito", EstadoDocumento: "emitida", CodigoValidacion: cude,
	}
	retry := &dbpkg.FacturacionElectronicaRetryItem{EstadoEnvio: "aceptado", CodigoValidacion: cude}
	if !facturacionNotaCreditoAceptadaParaAnulacion(nota, retry) {
		t.Fatal("nota emitida con CUDE y acuse aceptado coincidentes debe poder finalizar la anulacion")
	}

	for name, mutate := range map[string]func(*dbpkg.EmpresaDocumentoFacturacion, *dbpkg.FacturacionElectronicaRetryItem){
		"nota pendiente": func(n *dbpkg.EmpresaDocumentoFacturacion, _ *dbpkg.FacturacionElectronicaRetryItem) {
			n.EstadoDocumento = "pendiente_emision"
		},
		"sin CUDE": func(n *dbpkg.EmpresaDocumentoFacturacion, _ *dbpkg.FacturacionElectronicaRetryItem) {
			n.CodigoValidacion = ""
		},
		"cola fallida": func(_ *dbpkg.EmpresaDocumentoFacturacion, r *dbpkg.FacturacionElectronicaRetryItem) {
			r.EstadoEnvio = "fallido"
		},
		"CUDE distinto": func(_ *dbpkg.EmpresaDocumentoFacturacion, r *dbpkg.FacturacionElectronicaRetryItem) {
			r.CodigoValidacion = strings.Repeat("cd", 48)
		},
	} {
		t.Run(name, func(t *testing.T) {
			candidateNote := nota
			candidateRetry := *retry
			mutate(&candidateNote, &candidateRetry)
			if facturacionNotaCreditoAceptadaParaAnulacion(candidateNote, &candidateRetry) {
				t.Fatal("la anulacion debe permanecer cerrada")
			}
		})
	}
}

func TestNotaCreditoFuenteParaAnulacionDebeCoincidirConFactura(t *testing.T) {
	cufe := strings.Repeat("ef", 48)
	nota := dbpkg.EmpresaDocumentoFacturacion{DocumentoCodigo: "NC-1"}
	factura := dbpkg.EmpresaDocumentoFacturacion{DocumentoCodigo: "FV-1", NumeroLegal: "1PCS8", CodigoValidacion: cufe}
	fuente := &facturacionFuenteFiscalSnapshot{
		Documento: facturacionFuenteFiscalDocumento{TipoOrigen: "nota_credito", CodigoOrigen: nota.DocumentoCodigo},
		Referencia: &facturacionFuenteFiscalReferencia{
			TipoDocumento: "factura_electronica", DocumentoCodigo: factura.DocumentoCodigo,
			NumeroLegal: factura.NumeroLegal, CodigoValidacion: factura.CodigoValidacion,
		},
	}
	if !facturacionNotaCreditoFuenteValidaParaAnulacion(fuente, nota, factura) {
		t.Fatal("la fuente derivada y coincidente debe permitir finalizar")
	}

	for name, mutate := range map[string]func(*facturacionFuenteFiscalSnapshot){
		"sin referencia": func(s *facturacionFuenteFiscalSnapshot) { s.Referencia = nil },
		"otra factura":   func(s *facturacionFuenteFiscalSnapshot) { s.Referencia.DocumentoCodigo = "FV-2" },
		"otro CUFE":      func(s *facturacionFuenteFiscalSnapshot) { s.Referencia.CodigoValidacion = strings.Repeat("ab", 48) },
		"otra nota":      func(s *facturacionFuenteFiscalSnapshot) { s.Documento.CodigoOrigen = "NC-2" },
	} {
		t.Run(name, func(t *testing.T) {
			candidate := *fuente
			reference := *fuente.Referencia
			candidate.Referencia = &reference
			mutate(&candidate)
			if facturacionNotaCreditoFuenteValidaParaAnulacion(&candidate, nota, factura) {
				t.Fatal("la fuente discordante debe bloquear la anulacion")
			}
		})
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

func TestFacturacionDocumentoAceptadoDIANFinalizaPendienteConCUFE(t *testing.T) {
	cufe := strings.Repeat("ab", 48)
	doc, changed := facturacionDocumentoAceptadoDIAN(dbpkg.EmpresaDocumentoFacturacion{
		EmpresaID:       12,
		TipoDocumento:   "factura_electronica",
		DocumentoCodigo: "FV-PCS-ACEPTADA",
		EstadoDocumento: "pendiente_emision",
		EventoUltimo:    "integracion_fiscal_pendiente",
	}, `{"estado_dian":"aceptado","cufe":"`+cufe+`"}`)
	if !changed {
		t.Fatal("la aceptacion debe completar la transicion local pendiente")
	}
	if doc.EstadoDocumento != "emitida" || doc.EstadoAnterior != "pendiente_emision" {
		t.Fatalf("transicion aceptada inesperada: anterior=%q nuevo=%q", doc.EstadoAnterior, doc.EstadoDocumento)
	}
	if doc.EventoUltimo != "integracion_fiscal_aceptada" {
		t.Fatalf("evento final = %q", doc.EventoUltimo)
	}
	if doc.CodigoValidacion != cufe {
		t.Fatalf("CUFE final = %q", doc.CodigoValidacion)
	}
}

func TestFacturacionDocumentoAceptadoDIANEsIdempotenteYFailClosed(t *testing.T) {
	cufe := strings.Repeat("cd", 48)
	emitida := dbpkg.EmpresaDocumentoFacturacion{
		TipoDocumento:    "factura_electronica",
		EstadoDocumento:  "emitida",
		EventoUltimo:     "factura_emitida",
		CodigoValidacion: cufe,
	}
	got, changed := facturacionDocumentoAceptadoDIAN(emitida, `{"cufe":"`+cufe+`"}`)
	if changed || got.EventoUltimo != emitida.EventoUltimo {
		t.Fatal("una factura ya finalizada no debe reescribirse")
	}

	noSoportado := dbpkg.EmpresaDocumentoFacturacion{TipoDocumento: "nota_debito", EstadoDocumento: "pendiente_emision"}
	got, changed = facturacionDocumentoAceptadoDIAN(noSoportado, `{"cufe":"`+cufe+`"}`)
	if changed || got.EstadoDocumento != "pendiente_emision" {
		t.Fatal("un flujo comercial no implementado debe permanecer cerrado")
	}

	sinCUFE := dbpkg.EmpresaDocumentoFacturacion{TipoDocumento: "factura_electronica", EstadoDocumento: "pendiente_emision"}
	got, changed = facturacionDocumentoAceptadoDIAN(sinCUFE, `{"estado_dian":"aceptado"}`)
	if changed || got.EstadoDocumento != "pendiente_emision" {
		t.Fatal("un acuse sin CUFE/CUDE oficial no debe finalizar el documento")
	}
}

func TestFacturacionRetryAceptadoBackfillCUFEPreservaTrazabilidad(t *testing.T) {
	cufe := strings.Repeat("ef", 48)
	original := dbpkg.FacturacionElectronicaRetryItem{
		ID:                 91,
		EmpresaID:          12,
		EstadoEnvio:        "aceptado",
		Intentos:           1,
		FechaUltimoIntento: "2026-08-25 11:51:50",
		RespuestaProveedor: `{"estado_dian":"aceptado"}`,
	}
	got, changed := facturacionRetryAceptadoConCodigoValidacion(&original, strings.ToUpper(cufe))
	if !changed || got.CodigoValidacion != cufe {
		t.Fatalf("backfill CUFE inesperado: changed=%v codigo=%q", changed, got.CodigoValidacion)
	}
	if got.Intentos != original.Intentos || got.FechaUltimoIntento != original.FechaUltimoIntento || got.RespuestaProveedor != original.RespuestaProveedor {
		t.Fatal("el backfill no debe alterar intentos, fecha ni acuse")
	}
	if _, changed = facturacionRetryAceptadoConCodigoValidacion(&got, cufe); changed {
		t.Fatal("el backfill debe ser idempotente")
	}
	original.EstadoEnvio = "fallido"
	if _, changed = facturacionRetryAceptadoConCodigoValidacion(&original, cufe); changed {
		t.Fatal("una cola no aceptada no debe recibir codigo por esta ruta")
	}
}
