package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

const documentoSoporteConfirmacion = "EMITIR DOCUMENTO SOPORTE DIAN"

type documentoSoporteEmisionRequest struct {
	EmpresaID               int64  `json:"empresa_id"`
	DocumentoSoporteID      int64  `json:"documento_soporte_id"`
	ConfirmarEmision        bool   `json:"confirmar_emision"`
	MensajeConfirmacionDIAN string `json:"mensaje_confirmacion_dian"`
}

func documentoSoporteCodigoInterno(id int64) string {
	return fmt.Sprintf("DS-SOPORTE-%d", id)
}

func documentoSoporteEstadoDIANMirror(raw string) string {
	raw = strings.ToLower(strings.TrimSpace(raw))
	switch raw {
	case "preparado", "pendiente", "enviado", "aceptado", "rechazado", "fallido", "contingencia":
		return raw
	case "reconciliado":
		return "aceptado"
	case "fallido_terminal":
		return "fallido"
	case "", "no_aplica", "procesando", "recibido", "validando":
		return "pendiente"
	default:
		return "fallido"
	}
}

func buildDocumentoSoporteFuenteFiscal(document *dbpkg.EmpresaDocumentoSoporteElectronico, company *dbpkg.EmpresaConfiguracionAvanzada) (*facturacionFuenteFiscalSnapshot, error) {
	if document == nil || company == nil || document.EmpresaID <= 0 || company.EmpresaID != document.EmpresaID || document.ID <= 0 {
		return nil, fmt.Errorf("documento soporte y configuracion empresarial deben pertenecer a la misma empresa")
	}
	var storedLines []dbpkg.EmpresaDocumentoSoporteLinea
	decoder := json.NewDecoder(strings.NewReader(strings.TrimSpace(document.LineasJSON)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&storedLines); err != nil {
		return nil, fmt.Errorf("decodificar lineas estructuradas: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("lineas estructuradas contienen datos adicionales")
	}
	if len(storedLines) == 0 {
		return nil, fmt.Errorf("documento soporte sin lineas estructuradas")
	}

	companyPart, _ := facturacionFuenteFiscalPartes(company, nil)
	companyPart.ID = company.EmpresaID
	companyPart.ResidenciaFiscal = "residente"
	seller := facturacionFuenteFiscalParte{
		ID: document.ProveedorID, ResidenciaFiscal: strings.ToLower(strings.TrimSpace(document.VendedorResidencia)),
		TipoDocumento: strings.TrimSpace(document.TipoDocumento), NumeroDocumento: strings.TrimSpace(document.Documento),
		DigitoVerificacion: strings.TrimSpace(document.VendedorDigitoVerificacion), TipoPersona: strings.TrimSpace(document.VendedorTipoPersona),
		NombreRazonSocial: strings.TrimSpace(document.NombreProveedor), ResponsabilidadTributaria: strings.TrimSpace(document.VendedorResponsabilidadTributaria),
		Email: strings.TrimSpace(document.VendedorEmail), Telefono: strings.TrimSpace(document.VendedorTelefono), Direccion: strings.TrimSpace(document.VendedorDireccion),
		PaisCodigo: strings.ToUpper(strings.TrimSpace(document.VendedorPaisCodigo)), Departamento: strings.TrimSpace(document.VendedorDepartamento),
		DepartamentoCodigoDANE: strings.TrimSpace(document.VendedorDepartamentoCodigoDANE), Municipio: strings.TrimSpace(document.VendedorMunicipio),
		MunicipioCodigoDANE: strings.TrimSpace(document.VendedorMunicipioCodigoDANE), CodigoPostal: strings.TrimSpace(document.VendedorCodigoPostal),
	}
	if seller.ResidenciaFiscal == "" {
		seller.ResidenciaFiscal = "residente"
	}

	snapshot := &facturacionFuenteFiscalSnapshot{
		Esquema:   facturacionFuenteFiscalEsquema,
		Version:   facturacionFuenteFiscalVersion,
		EmpresaID: document.EmpresaID,
		Documento: facturacionFuenteFiscalDocumento{
			TipoOrigen: "documento_soporte", CodigoOrigen: documentoSoporteCodigoInterno(document.ID),
			TipoDestino: "documento_soporte", CodigoDestino: documentoSoporteCodigoInterno(document.ID),
			NumeroLegal: strings.TrimSpace(document.NumeroLegal), Fecha: strings.TrimSpace(document.FechaDocumento),
			Moneda: strings.ToUpper(strings.TrimSpace(document.Moneda)), MontoTotal: document.Total,
		},
		Emisor:  seller,
		Cliente: companyPart,
		Lineas:  make([]facturacionFuenteFiscalLinea, 0, len(storedLines)),
		Pago: facturacionFuenteFiscalPago{
			FormaCodigo: strings.TrimSpace(document.FormaPagoCodigo), MedioCodigo: strings.TrimSpace(document.MedioPagoCodigo),
			FechaVencimiento: strings.TrimSpace(document.FechaVencimiento), Referencia: documentoSoporteCodigoInterno(document.ID),
		},
	}
	if snapshot.Documento.Moneda == "" {
		snapshot.Documento.Moneda = "COP"
	}
	for _, line := range storedLines {
		snapshot.Lineas = append(snapshot.Lineas, facturacionFuenteFiscalLinea{
			Numero: line.Numero, CodigoItem: strings.TrimSpace(line.Codigo), Descripcion: strings.TrimSpace(line.Descripcion),
			UnidadMedida: strings.TrimSpace(line.UnidadMedida), Cantidad: line.Cantidad, PrecioUnitario: line.PrecioUnitario,
			DescuentoPorcentaje: line.DescuentoPorcentaje, ValorDescuento: line.ValorDescuento, BaseGravable: line.BaseGravable,
			ImpuestoCodigo: "01", ImpuestoPorcentaje: line.IVAPorcentaje, ValorImpuesto: line.IVAValor,
			ReteIVAPorcentaje: line.ReteIVAPorcentaje, ReteIVAValor: line.ReteIVAValor,
			ReteRentaPorcentaje: line.ReteRentaPorcentaje, ReteRentaValor: line.ReteRentaValor,
			SubtotalLinea: line.SubtotalLinea, TotalLinea: line.TotalLinea, TotalNetoContable: line.TotalNetoContableLinea,
		})
		snapshot.Totales.BrutoLineas += line.Cantidad * line.PrecioUnitario
		snapshot.Totales.DescuentoLineas += line.ValorDescuento
		snapshot.Totales.BaseGravableLineas += line.BaseGravable
		snapshot.Totales.ImpuestoLineas += line.IVAValor
		snapshot.Totales.RetencionesLineas += line.ReteIVAValor + line.ReteRentaValor
		snapshot.Totales.TotalLineas += line.TotalLinea
		snapshot.Totales.TotalNetoContable += line.TotalNetoContableLinea
	}
	snapshot.Totales.BrutoLineas = facturacionFuenteFiscalRound(snapshot.Totales.BrutoLineas)
	snapshot.Totales.DescuentoLineas = facturacionFuenteFiscalRound(snapshot.Totales.DescuentoLineas)
	snapshot.Totales.BaseGravableLineas = facturacionFuenteFiscalRound(snapshot.Totales.BaseGravableLineas)
	snapshot.Totales.ImpuestoLineas = facturacionFuenteFiscalRound(snapshot.Totales.ImpuestoLineas)
	snapshot.Totales.RetencionesLineas = facturacionFuenteFiscalRound(snapshot.Totales.RetencionesLineas)
	snapshot.Totales.TotalLineas = facturacionFuenteFiscalRound(snapshot.Totales.TotalLineas)
	snapshot.Totales.TotalNetoContable = facturacionFuenteFiscalRound(snapshot.Totales.TotalNetoContable)
	snapshot.Totales.TotalDocumentoOrigen = document.Total

	documentoSoporteCompletarBloqueantes(snapshot, document)
	if len(snapshot.Bloqueantes) > 0 {
		return snapshot, fmt.Errorf("fuente fiscal de documento soporte incompleta: %s", strings.Join(snapshot.Bloqueantes, ", "))
	}
	return snapshot, nil
}

func documentoSoporteCompletarBloqueantes(snapshot *facturacionFuenteFiscalSnapshot, document *dbpkg.EmpresaDocumentoSoporteElectronico) {
	if snapshot == nil || document == nil {
		return
	}
	add := func(value string) {
		for _, existing := range snapshot.Bloqueantes {
			if existing == value {
				return
			}
		}
		snapshot.Bloqueantes = append(snapshot.Bloqueantes, value)
	}
	if snapshot.Documento.Moneda != "COP" {
		add("documento.moneda_distinta_cop_requiere_tasa_cambio")
	}
	if strings.TrimSpace(snapshot.Documento.Fecha) == "" {
		add("documento.fecha_faltante")
	}
	if snapshot.Emisor.NombreRazonSocial == "" || snapshot.Emisor.NumeroDocumento == "" || snapshot.Emisor.TipoPersona == "" || snapshot.Emisor.ResponsabilidadTributaria == "" {
		add("vendedor.identidad_fiscal_incompleta")
	}
	sellerScheme, sellerDocument, ok := documentoSoporteSellerScheme(snapshot.Emisor.TipoDocumento, snapshot.Emisor.NumeroDocumento, snapshot.Emisor.ResidenciaFiscal)
	if !ok || sellerDocument == "" {
		add("vendedor.tipo_identificacion_no_admitido_dian")
	}
	if snapshot.Emisor.ResidenciaFiscal == "residente" {
		if sellerScheme != "31" || !strings.EqualFold(snapshot.Emisor.PaisCodigo, "CO") {
			add("vendedor.residente_requiere_nit_y_pais_co")
		}
		if expected, valid := calculateColombianNITDV(sellerDocument); !valid || strconv.Itoa(expected) != dianOnlyDigits(snapshot.Emisor.DigitoVerificacion) {
			add("vendedor.digito_verificacion_invalido")
		}
		if !facturacionFuenteFiscalDANEDepartamentoValido(snapshot.Emisor.DepartamentoCodigoDANE) ||
			!facturacionFuenteFiscalDANEMunicipioValido(snapshot.Emisor.MunicipioCodigoDANE, snapshot.Emisor.DepartamentoCodigoDANE) {
			add("vendedor.codigos_dane_invalidos")
		}
	} else if snapshot.Emisor.ResidenciaFiscal == "no_residente" {
		if snapshot.Emisor.PaisCodigo == "" || strings.EqualFold(snapshot.Emisor.PaisCodigo, "CO") {
			add("vendedor.no_residente_requiere_pais_extranjero")
		}
	} else {
		add("vendedor.residencia_fiscal_invalida")
	}
	if snapshot.Emisor.Direccion == "" || snapshot.Emisor.Municipio == "" || snapshot.Emisor.Departamento == "" {
		add("vendedor.ubicacion_fiscal_incompleta")
	}
	if snapshot.Cliente.NumeroDocumento == "" || snapshot.Cliente.DigitoVerificacion == "" || snapshot.Cliente.NombreRazonSocial == "" || snapshot.Cliente.ResponsabilidadTributaria == "" {
		add("adquirente.configuracion_fiscal_incompleta")
	}
	buyerNIT := dianOnlyDigits(snapshot.Cliente.NumeroDocumento)
	if expected, valid := calculateColombianNITDV(buyerNIT); !valid || strconv.Itoa(expected) != dianOnlyDigits(snapshot.Cliente.DigitoVerificacion) {
		add("adquirente.digito_verificacion_invalido")
	}
	if !strings.EqualFold(snapshot.Cliente.PaisCodigo, "CO") || snapshot.Cliente.Direccion == "" ||
		!facturacionFuenteFiscalDANEDepartamentoValido(snapshot.Cliente.DepartamentoCodigoDANE) ||
		!facturacionFuenteFiscalDANEMunicipioValido(snapshot.Cliente.MunicipioCodigoDANE, snapshot.Cliente.DepartamentoCodigoDANE) {
		add("adquirente.ubicacion_fiscal_colombia_incompleta")
	}
	if document.FormaPagoCodigo != "1" && document.FormaPagoCodigo != "2" {
		add("pago.forma_codigo_invalido")
	}
	if strings.TrimSpace(document.MedioPagoCodigo) == "" {
		add("pago.medio_codigo_faltante")
	}
	if document.FormaPagoCodigo == "2" && strings.TrimSpace(document.FechaVencimiento) == "" {
		add("pago.fecha_vencimiento_faltante")
	}
	if !facturacionFuenteFiscalClose(snapshot.Totales.BaseGravableLineas, document.Subtotal) ||
		!facturacionFuenteFiscalClose(snapshot.Totales.ImpuestoLineas, document.IVA) ||
		!facturacionFuenteFiscalClose(snapshot.Totales.RetencionesLineas, document.Retenciones) ||
		!facturacionFuenteFiscalClose(snapshot.Totales.TotalLineas, document.Total) ||
		!facturacionFuenteFiscalClose(snapshot.Totales.TotalNetoContable, document.TotalNetoContable) ||
		!facturacionFuenteFiscalClose(document.Subtotal+document.IVA, document.Total) ||
		!facturacionFuenteFiscalClose(document.Total-document.Retenciones, document.TotalNetoContable) {
		add("documento.totales_no_conciliados")
	}
	for _, line := range snapshot.Lineas {
		if line.Numero <= 0 || line.CodigoItem == "" || line.Descripcion == "" || line.Cantidad <= 0 || line.PrecioUnitario < 0 {
			add(fmt.Sprintf("lineas.%d.incompleta", line.Numero))
		}
		if !dbpkg.EmpresaDocumentoSoporteUnitCodeValid(line.UnidadMedida) {
			add(fmt.Sprintf("lineas.%d.unidad_medida_invalida", line.Numero))
		}
		if !facturacionFuenteFiscalClose(line.BaseGravable+line.ValorImpuesto, line.TotalLinea) ||
			!facturacionFuenteFiscalClose(line.TotalLinea-line.ReteIVAValor-line.ReteRentaValor, line.TotalNetoContable) {
			add(fmt.Sprintf("lineas.%d.totales_no_conciliados", line.Numero))
		}
	}
	snapshot.Bloqueantes = facturacionFuenteFiscalUniqueSorted(snapshot.Bloqueantes)
}

func documentoSoporteSellerScheme(rawType, rawDocument, residence string) (string, string, bool) {
	typeCode := strings.ToUpper(strings.TrimSpace(rawType))
	typeCode = strings.ReplaceAll(typeCode, " ", "")
	aliases := map[string]string{
		"TE": "21", "TARJETADEEXTRANJERIA": "21", "TARJETAEXTRANJERIA": "21", "21": "21",
		"CE": "22", "CEDULADEEXTRANJERIA": "22", "CEDULAEXTRANJERIA": "22", "22": "22",
		"NIT": "31", "31": "31", "PAS": "41", "PASAPORTE": "41", "41": "41",
		"DIE": "42", "DOCUMENTODEIDENTIFICACIONEXTRANJERO": "42", "DOCUMENTOIDENTIFICACIONEXTRANJERO": "42", "DOCUMENTOEXTRANJERO": "42", "42": "42",
		"PEP": "47", "47": "47", "NITOTROPAIS": "50", "NITDEOTROPAIS": "50", "NITEXTRANJERO": "50", "50": "50",
	}
	scheme := aliases[typeCode]
	residence = strings.ToLower(strings.TrimSpace(residence))
	if residence == "residente" && scheme != "31" {
		return "", "", false
	}
	if residence == "no_residente" {
		switch scheme {
		case "21", "22", "31", "41", "42", "47", "50":
		default:
			return "", "", false
		}
	}
	document := strings.TrimSpace(rawDocument)
	if scheme == "31" {
		document = dianOnlyDigits(document)
	} else {
		document = strings.Map(func(r rune) rune {
			if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' {
				return r
			}
			return -1
		}, document)
	}
	return scheme, document, scheme != "" && document != ""
}

func documentoSoporteMergeDIANConfig(base map[string]interface{}, snapshot *dbpkg.EmpresaDocumentoSoporteConfiguracionSnapshot) map[string]interface{} {
	out := make(map[string]interface{}, len(base)+12)
	for key, value := range base {
		out[key] = value
	}
	if snapshot == nil {
		return out
	}
	out["tipo_ambiente"] = snapshot.TipoAmbiente
	out["prefijo"] = snapshot.Prefijo
	out["resolucion_numero"] = snapshot.ResolucionNumero
	out["resolucion_fecha_desde"] = snapshot.ResolucionFechaDesde
	out["resolucion_fecha_hasta"] = snapshot.ResolucionFechaHasta
	out["rango_desde"] = snapshot.RangoDesde
	out["rango_hasta"] = snapshot.RangoHasta
	out["consecutivo_actual"] = snapshot.ConsecutivoAsignado
	out["modo_operacion_codigo"] = snapshot.ModoOperacionCodigo
	if strings.TrimSpace(snapshot.URLDIANOverride) != "" {
		out["url_dian"] = snapshot.URLDIANOverride
	}
	return out
}

func generateDIANUBLDocumentoSoporteDesdeFuenteFiscal(cfg map[string]interface{}, empresaID int64, payload map[string]interface{}, snapshot *facturacionFuenteFiscalSnapshot) (map[string]interface{}, int, error) {
	if snapshot == nil || snapshot.EmpresaID != empresaID || normalizeFacturacionDocumentoElectronicoTipo(snapshot.Documento.TipoOrigen) != "documento_soporte" || len(snapshot.Bloqueantes) > 0 {
		return nil, http.StatusUnprocessableEntity, fmt.Errorf("fuente fiscal inmutable de documento soporte invalida")
	}
	number := strings.ReplaceAll(dianFirstNonBlank(genericStringValue(payload["numero_legal"]), snapshot.Documento.NumeroLegal), " ", "")
	if number == "" || len(snapshot.Lineas) == 0 || !strings.EqualFold(snapshot.Documento.CodigoDestino, genericStringValue(payload["documento_codigo"])) {
		return nil, http.StatusUnprocessableEntity, fmt.Errorf("numero, lineas o identidad de documento soporte invalidos")
	}
	issueDate, issueTime := dianIssuepcs_ts(dianFirstNonBlank(genericStringValue(payload["fecha_emision"]), snapshot.Documento.Fecha))
	profileExecutionID := "2"
	if chooseDIANAmbiente(cfg) == "produccion" {
		profileExecutionID = "1"
	}
	customizationID := "10"
	if snapshot.Emisor.ResidenciaFiscal == "no_residente" {
		customizationID = "11"
	}
	sellerScheme, sellerDocument, ok := documentoSoporteSellerScheme(snapshot.Emisor.TipoDocumento, snapshot.Emisor.NumeroDocumento, snapshot.Emisor.ResidenciaFiscal)
	if !ok {
		return nil, http.StatusUnprocessableEntity, fmt.Errorf("identificacion del vendedor no admitida por DIAN")
	}
	buyerNIT := dianOnlyDigits(snapshot.Cliente.NumeroDocumento)
	softwareID, softwarePIN, _, credentialErr := resolveDIANSoftwareCredentials(cfg, payload, empresaID)
	if credentialErr != nil || softwareID == "" || softwarePIN == "" {
		return nil, http.StatusUnprocessableEntity, fmt.Errorf("software DIAN no disponible para documento soporte")
	}
	lineExtension := facturacionFuenteFiscalRound(snapshot.Totales.BaseGravableLineas)
	ivaTotal := facturacionFuenteFiscalRound(snapshot.Totales.ImpuestoLineas)
	total := facturacionFuenteFiscalRound(snapshot.Totales.TotalDocumentoOrigen)
	if !facturacionFuenteFiscalClose(lineExtension+ivaTotal, total) {
		return nil, http.StatusUnprocessableEntity, fmt.Errorf("total DIAN no concilia con base mas IVA")
	}
	cuds := buildDIANCUDSDocumentoSoporte(number, issueDate, issueTime, fmt.Sprintf("%.2f", lineExtension), fmt.Sprintf("%.2f", ivaTotal), fmt.Sprintf("%.2f", total), sellerDocument, buyerNIT, softwarePIN, profileExecutionID)
	securityCode := buildDIANSHA384Hex(softwareID, softwarePIN, number)
	qrBase := "https://catalogo-vpfe-hab.dian.gov.co/document/searchqr?documentkey="
	if profileExecutionID == "1" {
		qrBase = "https://catalogo-vpfe.dian.gov.co/document/searchqr?documentkey="
	}
	qrURL := qrBase + strings.ToLower(cuds)
	prefix := strings.ToUpper(strings.TrimSpace(genericStringValue(cfg["prefijo"])))
	resolution := strings.TrimSpace(genericStringValue(cfg["resolucion_numero"]))
	resolutionFrom := strings.TrimSpace(genericStringValue(cfg["resolucion_fecha_desde"]))
	resolutionTo := strings.TrimSpace(genericStringValue(cfg["resolucion_fecha_hasta"]))
	rangeFrom := anyToInt64(cfg["rango_desde"])
	rangeTo := anyToInt64(cfg["rango_hasta"])
	prefixXML := ""
	if prefix != "" {
		prefixXML = `<sts:Prefix>` + escapeXML(prefix) + `</sts:Prefix>`
	}
	invoiceControl := fmt.Sprintf(`<sts:InvoiceControl><sts:InvoiceAuthorization>%s</sts:InvoiceAuthorization><sts:AuthorizationPeriod><cbc:StartDate>%s</cbc:StartDate><cbc:EndDate>%s</cbc:EndDate></sts:AuthorizationPeriod><sts:AuthorizedInvoices>%s<sts:From>%d</sts:From><sts:To>%d</sts:To></sts:AuthorizedInvoices></sts:InvoiceControl>`,
		escapeXML(resolution), escapeXML(resolutionFrom), escapeXML(resolutionTo), prefixXML, rangeFrom, rangeTo)
	dianExtensions := fmt.Sprintf(`<ext:UBLExtensions><ext:UBLExtension><ext:ExtensionContent><sts:DianExtensions>%s<sts:InvoiceSource><cbc:IdentificationCode listAgencyID="6" listAgencyName="United Nations Economic Commission for Europe" listSchemeURI="urn:oasis:names:specification:ubl:codelist:gc:CountryIdentificationCode-2.1">CO</cbc:IdentificationCode></sts:InvoiceSource><sts:SoftwareProvider><sts:ProviderID schemeAgencyID="195" schemeAgencyName="%s" schemeID="%s" schemeName="31">%s</sts:ProviderID><sts:SoftwareID schemeAgencyID="195" schemeAgencyName="%s">%s</sts:SoftwareID></sts:SoftwareProvider><sts:SoftwareSecurityCode schemeAgencyID="195" schemeAgencyName="%s">%s</sts:SoftwareSecurityCode><sts:AuthorizationProvider><sts:AuthorizationProviderID schemeAgencyID="195" schemeAgencyName="%s" schemeID="4" schemeName="31">800197268</sts:AuthorizationProviderID></sts:AuthorizationProvider><sts:QRCode>NumDS: %s&#10;FecDS: %s&#10;HorDS: %s&#10;NumSNO: %s&#10;DocAdq: %s&#10;ValDS: %.2f&#10;ValIva: %.2f&#10;ValTolDS: %.2f&#10;CUDS: %s&#10;QRCode: %s</sts:QRCode></sts:DianExtensions></ext:ExtensionContent></ext:UBLExtension><ext:UBLExtension><ext:ExtensionContent></ext:ExtensionContent></ext:UBLExtension></ext:UBLExtensions>`,
		invoiceControl, escapeXML(dianAgencyName), escapeXML(dianCompanyIDSchemeID(buyerNIT, snapshot.Cliente.DigitoVerificacion)), escapeXML(buyerNIT),
		escapeXML(dianAgencyName), escapeXML(softwareID), escapeXML(dianAgencyName), escapeXML(securityCode), escapeXML(dianAgencyName),
		escapeXML(number), escapeXML(issueDate), escapeXML(issueTime), escapeXML(sellerDocument), escapeXML(buyerNIT), lineExtension, ivaTotal, total, escapeXML(strings.ToLower(cuds)), escapeXML(qrURL))
	dueDate := strings.TrimSpace(snapshot.Pago.FechaVencimiento)
	if dueDate == "" {
		dueDate = issueDate
	}
	header := fmt.Sprintf(`<cbc:UBLVersionID>UBL 2.1</cbc:UBLVersionID><cbc:CustomizationID>%s</cbc:CustomizationID><cbc:ProfileID>%s</cbc:ProfileID><cbc:ProfileExecutionID>%s</cbc:ProfileExecutionID><cbc:ID>%s</cbc:ID><cbc:UUID schemeID="%s" schemeName="CUDS-SHA384">%s</cbc:UUID><cbc:IssueDate>%s</cbc:IssueDate><cbc:IssueTime>%s</cbc:IssueTime><cbc:DueDate>%s</cbc:DueDate><cbc:InvoiceTypeCode>05</cbc:InvoiceTypeCode><cbc:Note>%s</cbc:Note><cbc:DocumentCurrencyCode>%s</cbc:DocumentCurrencyCode><cbc:LineCountNumeric>%d</cbc:LineCountNumeric>`,
		escapeXML(customizationID), escapeXML(dianDocumentProfileID("DocumentoSoporte")), escapeXML(profileExecutionID), escapeXML(number), escapeXML(profileExecutionID), escapeXML(strings.ToLower(cuds)),
		escapeXML(issueDate), escapeXML(issueTime), escapeXML(dueDate), escapeXML(snapshot.Documento.CodigoOrigen), escapeXML(snapshot.Documento.Moneda), len(snapshot.Lineas))
	supplier, err := documentoSoporteSupplierPartyXML(snapshot.Emisor, sellerScheme, sellerDocument)
	if err != nil {
		return nil, http.StatusUnprocessableEntity, err
	}
	customer, err := documentoSoporteCustomerPartyXML(snapshot.Cliente)
	if err != nil {
		return nil, http.StatusUnprocessableEntity, err
	}
	payment := fmt.Sprintf(`<cac:PaymentMeans><cbc:ID>%s</cbc:ID><cbc:PaymentMeansCode>%s</cbc:PaymentMeansCode><cbc:PaymentDueDate>%s</cbc:PaymentDueDate><cbc:PaymentID>%s</cbc:PaymentID></cac:PaymentMeans>`,
		escapeXML(snapshot.Pago.FormaCodigo), escapeXML(snapshot.Pago.MedioCodigo), escapeXML(dueDate), escapeXML(snapshot.Pago.Referencia))
	taxes, withholdings, err := documentoSoporteTaxTotalsXML(snapshot.Lineas, snapshot.Documento.Moneda)
	if err != nil {
		return nil, http.StatusUnprocessableEntity, err
	}
	monetary := dianFuenteFiscalMonetaryTotalXML(snapshot, snapshot.Documento.Moneda)
	lines, err := documentoSoporteLinesXML(snapshot.Lineas, snapshot.Documento.Moneda, issueDate)
	if err != nil {
		return nil, http.StatusUnprocessableEntity, err
	}
	xmlPayload := `<?xml version="1.0" encoding="UTF-8" standalone="no"?>` +
		`<Invoice xmlns="urn:oasis:names:specification:ubl:schema:xsd:Invoice-2" xmlns:cac="urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2" xmlns:cbc="urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2" xmlns:ds="http://www.w3.org/2000/09/xmldsig#" xmlns:ext="urn:oasis:names:specification:ubl:schema:xsd:CommonExtensionComponents-2" xmlns:sts="dian:gov:co:facturaelectronica:Structures-2-1" xmlns:xades="http://uri.etsi.org/01903/v1.3.2#" xmlns:xades141="http://uri.etsi.org/01903/v1.4.1#" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="urn:oasis:names:specification:ubl:schema:xsd:Invoice-2 http://docs.oasis-open.org/ubl/os-UBL-2.1/xsd/maindoc/UBL-Invoice-2.1.xsd">` +
		dianExtensions + header + supplier + customer + payment + taxes + withholdings + monetary + lines + `</Invoice>`
	return map[string]interface{}{
		"ok": true, "empresa_id": empresaID, "documento_codigo": snapshot.Documento.CodigoOrigen, "documento_tipo": "documento_soporte",
		"ubl_version": "UBL 2.1", "profile_execution_id": profileExecutionID, "customization_id": customizationID,
		"uuid_scheme": "CUDS-SHA384", "uuid": strings.ToLower(cuds), "cuds": strings.ToLower(cuds), "software_security_code": "[calculado]",
		"xml_ubl_base": xmlPayload, "estado_preparacion": "pre_envio_validable",
		"fuente_fiscal": map[string]interface{}{"tipo": snapshot.Documento.TipoOrigen, "codigo": snapshot.Documento.CodigoOrigen, "lineas": len(snapshot.Lineas)},
	}, http.StatusOK, nil
}

func documentoSoporteAddressXML(part facturacionFuenteFiscalParte) string {
	var fields strings.Builder
	if code := strings.TrimSpace(part.MunicipioCodigoDANE); code != "" {
		fields.WriteString("<cbc:ID>" + escapeXML(code) + "</cbc:ID>")
	}
	if city := strings.TrimSpace(part.Municipio); city != "" {
		fields.WriteString("<cbc:CityName>" + escapeXML(city) + "</cbc:CityName>")
	}
	if postal := strings.TrimSpace(part.CodigoPostal); postal != "" {
		fields.WriteString("<cbc:PostalZone>" + escapeXML(postal) + "</cbc:PostalZone>")
	}
	if department := strings.TrimSpace(part.Departamento); department != "" {
		fields.WriteString("<cbc:CountrySubentity>" + escapeXML(department) + "</cbc:CountrySubentity>")
	}
	if departmentCode := strings.TrimSpace(part.DepartamentoCodigoDANE); departmentCode != "" {
		fields.WriteString("<cbc:CountrySubentityCode>" + escapeXML(departmentCode) + "</cbc:CountrySubentityCode>")
	}
	if address := strings.TrimSpace(part.Direccion); address != "" {
		fields.WriteString("<cac:AddressLine><cbc:Line>" + escapeXML(address) + "</cbc:Line></cac:AddressLine>")
	}
	fields.WriteString(fmt.Sprintf(`<cac:Country><cbc:IdentificationCode>%s</cbc:IdentificationCode><cbc:Name languageID="es">%s</cbc:Name></cac:Country>`, escapeXML(part.PaisCodigo), escapeXML(documentoSoporteCountryName(part.PaisCodigo))))
	return fields.String()
}

func documentoSoporteCountryName(code string) string {
	if strings.EqualFold(strings.TrimSpace(code), "CO") {
		return "Colombia"
	}
	return strings.ToUpper(strings.TrimSpace(code))
}

func documentoSoporteAdditionalAccountID(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "juridica", "juridico", "persona_juridica", "1":
		return "1"
	case "natural", "persona_natural", "2":
		return "2"
	default:
		return ""
	}
}

func documentoSoporteContactXML(part facturacionFuenteFiscalParte) string {
	var fields strings.Builder
	if telephone := strings.TrimSpace(part.Telefono); telephone != "" {
		fields.WriteString("<cbc:Telephone>" + escapeXML(telephone) + "</cbc:Telephone>")
	}
	if email := strings.TrimSpace(part.Email); email != "" {
		fields.WriteString("<cbc:ElectronicMail>" + escapeXML(email) + "</cbc:ElectronicMail>")
	}
	if fields.Len() == 0 {
		return ""
	}
	return "<cac:Contact>" + fields.String() + "</cac:Contact>"
}

func documentoSoporteSupplierPartyXML(part facturacionFuenteFiscalParte, scheme, document string) (string, error) {
	accountID := documentoSoporteAdditionalAccountID(part.TipoPersona)
	if accountID == "" || scheme == "" || document == "" || part.NombreRazonSocial == "" || part.ResponsabilidadTributaria == "" {
		return "", fmt.Errorf("vendedor no obligado incompleto")
	}
	attrs := fmt.Sprintf(`schemeAgencyID="195" schemeAgencyName="%s" schemeName="%s"`, escapeXML(dianAgencyName), escapeXML(scheme))
	if scheme == "31" {
		attrs += fmt.Sprintf(` schemeID="%s"`, escapeXML(dianCompanyIDSchemeID(document, part.DigitoVerificacion)))
	}
	contact := documentoSoporteContactXML(part)
	address := documentoSoporteAddressXML(part)
	return fmt.Sprintf(`<cac:AccountingSupplierParty><cbc:AdditionalAccountID>%s</cbc:AdditionalAccountID><cac:Party><cac:PhysicalLocation><cac:Address>%s</cac:Address></cac:PhysicalLocation><cac:PartyTaxScheme><cbc:RegistrationName>%s</cbc:RegistrationName><cbc:CompanyID %s>%s</cbc:CompanyID><cbc:TaxLevelCode>%s</cbc:TaxLevelCode><cac:TaxScheme><cbc:ID>ZZ</cbc:ID><cbc:Name>No aplica</cbc:Name></cac:TaxScheme></cac:PartyTaxScheme>%s</cac:Party></cac:AccountingSupplierParty>`,
		escapeXML(accountID), address, escapeXML(part.NombreRazonSocial), attrs, escapeXML(document), escapeXML(part.ResponsabilidadTributaria), contact), nil
}

func documentoSoporteCustomerPartyXML(part facturacionFuenteFiscalParte) (string, error) {
	nit := dianOnlyDigits(part.NumeroDocumento)
	accountID := documentoSoporteAdditionalAccountID(part.TipoPersona)
	if accountID == "" || nit == "" || part.NombreRazonSocial == "" || part.ResponsabilidadTributaria == "" {
		return "", fmt.Errorf("adquirente empresarial incompleto")
	}
	attrs := fmt.Sprintf(`schemeAgencyID="195" schemeAgencyName="%s" schemeID="%s" schemeName="31"`, escapeXML(dianAgencyName), escapeXML(dianCompanyIDSchemeID(nit, part.DigitoVerificacion)))
	contact := documentoSoporteContactXML(part)
	return fmt.Sprintf(`<cac:AccountingCustomerParty><cbc:AdditionalAccountID>%s</cbc:AdditionalAccountID><cac:Party><cac:PartyTaxScheme><cbc:RegistrationName>%s</cbc:RegistrationName><cbc:CompanyID %s>%s</cbc:CompanyID><cbc:TaxLevelCode>%s</cbc:TaxLevelCode><cac:TaxScheme><cbc:ID>01</cbc:ID><cbc:Name>IVA</cbc:Name></cac:TaxScheme></cac:PartyTaxScheme>%s</cac:Party></cac:AccountingCustomerParty>`,
		escapeXML(accountID), escapeXML(part.NombreRazonSocial), attrs, escapeXML(nit), escapeXML(part.ResponsabilidadTributaria), contact), nil
}

type documentoSoporteTaxGroup struct {
	Code, Name string
	Rate       float64
	Base       float64
	Amount     float64
}

func documentoSoporteTaxTotalsXML(lines []facturacionFuenteFiscalLinea, currency string) (string, string, error) {
	iva := map[string]*documentoSoporteTaxGroup{}
	withholdings := map[string]*documentoSoporteTaxGroup{}
	for _, line := range lines {
		ivaKey := fmt.Sprintf("01|%.2f", line.ImpuestoPorcentaje)
		if iva[ivaKey] == nil {
			iva[ivaKey] = &documentoSoporteTaxGroup{Code: "01", Name: "IVA", Rate: line.ImpuestoPorcentaje}
		}
		iva[ivaKey].Base += line.BaseGravable
		iva[ivaKey].Amount += line.ValorImpuesto
		if line.ReteIVAPorcentaje > 0 {
			key := fmt.Sprintf("05|%.2f", line.ReteIVAPorcentaje)
			if withholdings[key] == nil {
				withholdings[key] = &documentoSoporteTaxGroup{Code: "05", Name: "ReteIVA", Rate: line.ReteIVAPorcentaje}
			}
			withholdings[key].Base += line.ValorImpuesto
			withholdings[key].Amount += line.ReteIVAValor
		}
		if line.ReteRentaPorcentaje > 0 {
			key := fmt.Sprintf("06|%.2f", line.ReteRentaPorcentaje)
			if withholdings[key] == nil {
				withholdings[key] = &documentoSoporteTaxGroup{Code: "06", Name: "ReteRenta", Rate: line.ReteRentaPorcentaje}
			}
			withholdings[key].Base += line.BaseGravable
			withholdings[key].Amount += line.ReteRentaValor
		}
	}
	return documentoSoporteTaxGroupXML(iva, currency, "TaxTotal"), documentoSoporteTaxGroupXML(withholdings, currency, "WithholdingTaxTotal"), nil
}

func documentoSoporteTaxGroupXML(groups map[string]*documentoSoporteTaxGroup, currency, tag string) string {
	byCode := make(map[string][]*documentoSoporteTaxGroup)
	for _, group := range groups {
		if group == nil {
			continue
		}
		copyGroup := *group
		copyGroup.Base = facturacionFuenteFiscalRound(copyGroup.Base)
		copyGroup.Amount = facturacionFuenteFiscalRound(copyGroup.Amount)
		byCode[copyGroup.Code] = append(byCode[copyGroup.Code], &copyGroup)
	}
	codes := make([]string, 0, len(byCode))
	for code := range byCode {
		codes = append(codes, code)
	}
	sort.Strings(codes)
	var out strings.Builder
	for _, code := range codes {
		codeGroups := byCode[code]
		sort.Slice(codeGroups, func(i, j int) bool { return codeGroups[i].Rate < codeGroups[j].Rate })
		var total float64
		for _, group := range codeGroups {
			total += group.Amount
		}
		out.WriteString(fmt.Sprintf(`<cac:%s><cbc:TaxAmount currencyID="%s">%.2f</cbc:TaxAmount>`, tag, escapeXML(currency), facturacionFuenteFiscalRound(total)))
		for _, group := range codeGroups {
			out.WriteString(fmt.Sprintf(`<cac:TaxSubtotal><cbc:TaxableAmount currencyID="%s">%.2f</cbc:TaxableAmount><cbc:TaxAmount currencyID="%s">%.2f</cbc:TaxAmount><cac:TaxCategory><cbc:Percent>%.2f</cbc:Percent><cac:TaxScheme><cbc:ID>%s</cbc:ID><cbc:Name>%s</cbc:Name></cac:TaxScheme></cac:TaxCategory></cac:TaxSubtotal>`,
				escapeXML(currency), group.Base, escapeXML(currency), group.Amount, group.Rate, escapeXML(group.Code), escapeXML(group.Name)))
		}
		out.WriteString("</cac:" + tag + ">")
	}
	return out.String()
}

func documentoSoporteLinesXML(lines []facturacionFuenteFiscalLinea, currency, issueDate string) (string, error) {
	var out strings.Builder
	for _, line := range lines {
		if !dbpkg.EmpresaDocumentoSoporteUnitCodeValid(line.UnidadMedida) {
			return "", fmt.Errorf("linea %d con unidad DIAN no admitida", line.Numero)
		}
		allowance := ""
		if line.ValorDescuento > 0 {
			allowance = fmt.Sprintf(`<cac:AllowanceCharge><cbc:ID>1</cbc:ID><cbc:ChargeIndicator>false</cbc:ChargeIndicator><cbc:AllowanceChargeReason>Descuento por parte del vendedor</cbc:AllowanceChargeReason><cbc:MultiplierFactorNumeric>%.2f</cbc:MultiplierFactorNumeric><cbc:Amount currencyID="%s">%.2f</cbc:Amount><cbc:BaseAmount currencyID="%s">%.2f</cbc:BaseAmount></cac:AllowanceCharge>`,
				line.DescuentoPorcentaje, escapeXML(currency), line.ValorDescuento, escapeXML(currency), facturacionFuenteFiscalRound(line.Cantidad*line.PrecioUnitario))
		}
		iva := documentoSoporteTaxGroupXML(map[string]*documentoSoporteTaxGroup{"iva": {Code: "01", Name: "IVA", Rate: line.ImpuestoPorcentaje, Base: line.BaseGravable, Amount: line.ValorImpuesto}}, currency, "TaxTotal")
		retentions := map[string]*documentoSoporteTaxGroup{}
		if line.ReteIVAPorcentaje > 0 {
			retentions["reteiva"] = &documentoSoporteTaxGroup{Code: "05", Name: "ReteIVA", Rate: line.ReteIVAPorcentaje, Base: line.ValorImpuesto, Amount: line.ReteIVAValor}
		}
		if line.ReteRentaPorcentaje > 0 {
			retentions["reterenta"] = &documentoSoporteTaxGroup{Code: "06", Name: "ReteRenta", Rate: line.ReteRentaPorcentaje, Base: line.BaseGravable, Amount: line.ReteRentaValor}
		}
		withholding := documentoSoporteTaxGroupXML(retentions, currency, "WithholdingTaxTotal")
		out.WriteString(fmt.Sprintf(`<cac:InvoiceLine><cbc:ID>%d</cbc:ID><cbc:InvoicedQuantity unitCode="%s">%.6f</cbc:InvoicedQuantity><cbc:LineExtensionAmount currencyID="%s">%.2f</cbc:LineExtensionAmount><cac:InvoicePeriod><cbc:StartDate>%s</cbc:StartDate><cbc:DescriptionCode>1</cbc:DescriptionCode><cbc:Description>Por operación</cbc:Description></cac:InvoicePeriod>%s%s%s<cac:Item><cbc:Description>%s</cbc:Description><cac:SellersItemIdentification><cbc:ID>%s</cbc:ID></cac:SellersItemIdentification><cac:StandardItemIdentification><cbc:ID schemeID="999" schemeName="Estándar de adopción del contribuyente">%s</cbc:ID></cac:StandardItemIdentification></cac:Item><cac:Price><cbc:PriceAmount currencyID="%s">%.2f</cbc:PriceAmount><cbc:BaseQuantity unitCode="%s">1.000000</cbc:BaseQuantity></cac:Price></cac:InvoiceLine>`,
			line.Numero, escapeXML(strings.ToUpper(strings.TrimSpace(line.UnidadMedida))), line.Cantidad, escapeXML(currency), line.BaseGravable, escapeXML(issueDate), allowance, iva, withholding,
			escapeXML(line.Descripcion), escapeXML(line.CodigoItem), escapeXML(line.CodigoItem), escapeXML(currency), line.PrecioUnitario, escapeXML(strings.ToUpper(strings.TrimSpace(line.UnidadMedida)))))
	}
	return out.String(), nil
}

func documentoSoporteConfigSnapshotFromRow(config *dbpkg.EmpresaDIANDocumentoConfiguracion) *dbpkg.EmpresaDocumentoSoporteConfiguracionSnapshot {
	if config == nil {
		return nil
	}
	return &dbpkg.EmpresaDocumentoSoporteConfiguracionSnapshot{
		TipoDocumento: "documento_soporte", TipoAmbiente: config.TipoAmbiente, ModoOperacionCodigo: config.ModoOperacionCodigo,
		Prefijo: config.Prefijo, ResolucionNumero: config.ResolucionNumero, ResolucionFechaDesde: config.ResolucionFechaDesde,
		ResolucionFechaHasta: config.ResolucionFechaHasta, RangoDesde: config.RangoDesde, RangoHasta: config.RangoHasta,
		ConsecutivoAsignado: config.ConsecutivoActual, URLDIANOverride: config.URLDIANOverride,
	}
}

func loadDocumentoSoporteParaDispatch(ctx context.Context, dbEmp *sql.DB, doc dbpkg.EmpresaDocumentoFacturacion) (*dbpkg.EmpresaDocumentoSoporteElectronico, *dbpkg.EmpresaDocumentoSoporteConfiguracionSnapshot, error) {
	if dbEmp == nil || doc.EmpresaID <= 0 || normalizeFacturacionDocumentoElectronicoTipo(doc.TipoDocumento) != "documento_soporte" {
		return nil, nil, fmt.Errorf("documento soporte fiscal invalido")
	}
	code := strings.ToUpper(strings.TrimSpace(doc.DocumentoCodigo))
	if !strings.HasPrefix(code, "DS-SOPORTE-") {
		return nil, nil, fmt.Errorf("documento soporte sin codigo interno trazable")
	}
	id, err := strconv.ParseInt(strings.TrimPrefix(code, "DS-SOPORTE-"), 10, 64)
	if err != nil || id <= 0 {
		return nil, nil, fmt.Errorf("codigo interno de documento soporte invalido")
	}
	support, err := dbpkg.GetEmpresaDocumentoSoporteByIDContext(ctx, dbEmp, doc.EmpresaID, id)
	if err != nil {
		return nil, nil, err
	}
	if strings.TrimSpace(support.NumeroLegal) == "" || !strings.EqualFold(strings.TrimSpace(support.NumeroLegal), strings.TrimSpace(doc.NumeroLegal)) {
		return nil, nil, fmt.Errorf("numeracion del documento soporte no coincide con su registro contable")
	}
	var config dbpkg.EmpresaDocumentoSoporteConfiguracionSnapshot
	decoder := json.NewDecoder(strings.NewReader(strings.TrimSpace(support.ConfiguracionDIANJSON)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&config); err != nil {
		return nil, nil, fmt.Errorf("instantanea DIAN de documento soporte invalida: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return nil, nil, fmt.Errorf("instantanea DIAN contiene datos adicionales")
	}
	if config.TipoDocumento != "documento_soporte" || config.ConsecutivoAsignado <= 0 {
		return nil, nil, fmt.Errorf("instantanea DIAN de documento soporte incompleta")
	}
	return support, &config, nil
}

func decodeDocumentoSoporteEmisionRequest(w http.ResponseWriter, r *http.Request) (documentoSoporteEmisionRequest, bool) {
	var request documentoSoporteEmisionRequest
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16<<10))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		http.Error(w, "Solicitud de emision de documento soporte invalida", http.StatusBadRequest)
		return request, false
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		http.Error(w, "Solicitud de emision de documento soporte invalida", http.StatusBadRequest)
		return request, false
	}
	if err := facturacionBindAuthorizedEmpresaID(r, &request.EmpresaID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return request, false
	}
	if request.DocumentoSoporteID <= 0 {
		http.Error(w, "documento_soporte_id es obligatorio", http.StatusBadRequest)
		return request, false
	}
	if !request.ConfirmarEmision || strings.TrimSpace(request.MensajeConfirmacionDIAN) != documentoSoporteConfirmacion {
		writeJSON(w, http.StatusConflict, map[string]interface{}{
			"ok": false, "bloqueado": true, "codigo": "confirmacion_dian_requerida",
			"mensaje_confirmacion_requerido": documentoSoporteConfirmacion,
		})
		return request, false
	}
	return request, true
}

type documentoSoporteEmisionContext struct {
	document *dbpkg.EmpresaDocumentoSoporteElectronico
	company  *dbpkg.EmpresaConfiguracionAvanzada
}

func loadDocumentoSoporteEmisionContext(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB, request documentoSoporteEmisionRequest) (*documentoSoporteEmisionContext, bool) {
	document, err := dbpkg.GetEmpresaDocumentoSoporteByIDContext(r.Context(), dbEmp, request.EmpresaID, request.DocumentoSoporteID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "documento soporte no encontrado", http.StatusNotFound)
			return nil, false
		}
		http.Error(w, "No se pudo consultar el documento soporte", http.StatusInternalServerError)
		return nil, false
	}
	company, err := dbpkg.GetEmpresaConfiguracionAvanzada(dbEmp, request.EmpresaID)
	if err != nil || company == nil {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]interface{}{"ok": false, "bloqueado": true, "codigo": "configuracion_empresa_incompleta", "error": "No existe configuracion fiscal empresarial completa."})
		return nil, false
	}
	config, err := dbpkg.GetEmpresaDIANDocumentoConfiguracionContext(r.Context(), dbEmp, request.EmpresaID, "documento_soporte")
	if err != nil || config == nil {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]interface{}{"ok": false, "bloqueado": true, "codigo": "configuracion_documento_soporte_incompleta", "error": "No existe configuracion DIAN separada para documento soporte."})
		return nil, false
	}
	mainConfig, err := getEmpresaDIANConfig(dbEmp, request.EmpresaID)
	if err != nil || len(mainConfig) == 0 {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]interface{}{"ok": false, "bloqueado": true, "codigo": "configuracion_dian_incompleta", "error": "No existe configuracion DIAN principal para firma y transporte."})
		return nil, false
	}
	preflight := buildDocumentoSoporteDIANPreflight(document, config, company, mainConfig)
	if !preflight.PuedeEmitir {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]interface{}{"ok": false, "bloqueado": true, "codigo": "preflight_documento_soporte", "preflight": preflight})
		return nil, false
	}
	_, err = buildDocumentoSoporteFuenteFiscal(document, company)
	if err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]interface{}{"ok": false, "bloqueado": true, "codigo": "fuente_fiscal_documento_soporte_invalida", "error": err.Error()})
		return nil, false
	}
	return &documentoSoporteEmisionContext{document: document, company: company}, true
}

func handleEmitirDocumentoSoporte(w http.ResponseWriter, r *http.Request, dbEmp, dbSuper *sql.DB) {
	request, ok := decodeDocumentoSoporteEmisionRequest(w, r)
	if !ok {
		return
	}
	emissionContext, ok := loadDocumentoSoporteEmisionContext(w, r, dbEmp, request)
	if !ok {
		return
	}
	usuario := strings.TrimSpace(adminEmailFromRequest(r))

	code := documentoSoporteCodigoInterno(emissionContext.document.ID)
	lockContext, release, locked, lockErr := acquireFacturacionDocumentAdvisoryLock(r.Context(), dbEmp, request.EmpresaID, "documento_soporte", code)
	if lockErr != nil {
		http.Error(w, "No se pudo reservar el documento soporte", http.StatusInternalServerError)
		return
	}
	if !locked {
		http.Error(w, "el documento soporte ya tiene una emision en proceso", http.StatusConflict)
		return
	}
	defer release()

	document, configSnapshot, err := dbpkg.ReserveEmpresaDocumentoSoporteNumeroContext(lockContext, dbEmp, request.EmpresaID, request.DocumentoSoporteID, time.Now())
	if err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]interface{}{"ok": false, "bloqueado": true, "codigo": "reserva_numeracion_documento_soporte", "error": err.Error()})
		return
	}
	source, err := buildDocumentoSoporteFuenteFiscal(document, emissionContext.company)
	if err != nil {
		http.Error(w, "La numeracion quedo reservada, pero no se pudo sellar la fuente fiscal; reintente sin cambiar el borrador", http.StatusInternalServerError)
		return
	}
	doc := dbpkg.EmpresaDocumentoFacturacion{
		EmpresaID: request.EmpresaID, TipoDocumento: "documento_soporte", DocumentoCodigo: code,
		NumeroLegal: document.NumeroLegal, PaisCodigo: "CO", AmbienteFE: configSnapshot.TipoAmbiente,
		EstadoDocumento: "emitida", EstadoAnterior: "borrador", EventoUltimo: "documento_soporte_emitido",
		PeriodoContable: document.Periodo, MontoTotal: document.Total, Moneda: document.Moneda,
		FechaDocumento: document.FechaEmisionLegal, EntidadRelacionadaID: document.ProveedorID,
		UsuarioCreador: usuario, Estado: "activo", Observaciones: "documento soporte generado desde borrador contable estructurado id=" + strconv.FormatInt(document.ID, 10),
	}
	persisted, err := dbpkg.UpsertEmpresaDocumentoFacturacionContext(lockContext, dbEmp, doc)
	if err != nil {
		http.Error(w, "No se pudo persistir el documento fiscal reservado", http.StatusInternalServerError)
		return
	}
	if _, err := saveFacturacionFuenteFiscalSnapshot(lockContext, dbEmp, *persisted, source); err != nil {
		http.Error(w, "No se pudo sellar la fuente fiscal inmutable del documento soporte", http.StatusInternalServerError)
		return
	}
	if err := dbpkg.UpdateEmpresaDocumentoSoporteDIANResultContext(lockContext, dbEmp, request.EmpresaID, document.ID, "preparado", "", "", true); err != nil {
		http.Error(w, "El documento fiscal quedo preparado, pero no se pudo actualizar su estado contable; no se transmitio a DIAN", http.StatusInternalServerError)
		return
	}

	payload := facturacionBuildOperacionPayloadFromDocumento(*persisted)
	payload.ClienteNombre = source.Cliente.NombreRazonSocial
	payload.ClienteNumeroDocumento = source.Cliente.NumeroDocumento
	payload.ClienteTipoDocumento = source.Cliente.TipoDocumento
	payload.ClienteEmail = source.Cliente.Email
	payload.ClienteTelefono = source.Cliente.Telefono
	payload.ClienteDireccion = source.Cliente.Direccion
	integration, retry, integrationErr := processFacturacionIntegracionForDocumentoContext(lockContext, dbEmp, payload, *persisted, "emitir_documento_soporte", usuario, dbSuper)
	if integrationErr != nil {
		http.Error(w, "No se pudo completar la integracion DIAN del documento soporte", http.StatusInternalServerError)
		return
	}
	persistenceWarnings := make([]string, 0, 2)
	refreshed, refreshErr := dbpkg.GetEmpresaDocumentoFacturacionByCodigoContext(lockContext, dbEmp, request.EmpresaID, "documento_soporte", code)
	if refreshErr != nil {
		persistenceWarnings = append(persistenceWarnings, "La integracion termino, pero no se pudo refrescar el documento fiscal en la respuesta.")
	} else if refreshed != nil {
		persisted = refreshed
	}
	cuds := strings.ToLower(strings.TrimSpace(persisted.CodigoValidacion))
	state := documentoSoporteEstadoDIANMirror(integration.EstadoEnvio)
	response := ""
	if retry != nil {
		response = retry.RespuestaProveedor
		if cuds == "" {
			cuds = facturacionCUFEOficialDesdeRespuesta(response)
		}
	}
	if err := dbpkg.UpdateEmpresaDocumentoSoporteDIANResultContext(lockContext, dbEmp, request.EmpresaID, document.ID, state, cuds, response, true); err != nil {
		persistenceWarnings = append(persistenceWarnings, "La integracion termino, pero no se pudo actualizar el espejo contable final.")
	}
	status := http.StatusAccepted
	if integration.EstadoEnvio == "aceptado" {
		status = http.StatusOK
	}
	writeJSON(w, status, map[string]interface{}{
		"ok":     integration.EstadoEnvio == "aceptado" || integration.EstadoEnvio == "enviado",
		"accion": "emitir_documento_soporte", "empresa_id": request.EmpresaID,
		"documento_soporte": document, "documento_fiscal": persisted, "integracion_fiscal": integration,
		"cola_reintentos": retry, "fuente_fiscal_sellada": true, "advertencias_persistencia": persistenceWarnings,
	})
}
