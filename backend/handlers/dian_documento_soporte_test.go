package handlers

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateDIANUBLDocumentoSoporteCumpleContratoAnexo11(t *testing.T) {
	documento, configuracion, empresa, principal := documentoSoportePreflightFixture()
	documento.NumeroLegal = "DS1"
	fuente, err := buildDocumentoSoporteFuenteFiscal(documento, empresa)
	if err != nil {
		t.Fatal(err)
	}
	cfg := documentoSoporteMergeDIANConfig(principal, documentoSoporteConfigSnapshotFromRow(configuracion))
	generated, status, err := generateDIANUBLDocumentoSoporteDesdeFuenteFiscal(cfg, documento.EmpresaID, map[string]interface{}{
		"documento_codigo": documentoSoporteCodigoInterno(documento.ID),
		"numero_legal":     documento.NumeroLegal,
		"fecha_emision":    "2026-08-26T10:15:30-05:00",
	}, fuente)
	if err != nil || status != 200 {
		t.Fatalf("generacion DIAN fallo status=%d err=%v", status, err)
	}
	xml := genericStringValue(generated["xml_ubl_base"])
	values, root, err := parseDIANXMLTextValues(xml)
	if err != nil {
		t.Fatalf("XML no bien formado: %v", err)
	}
	if root != "Invoice" || dianXMLFirst(values, "InvoiceTypeCode") != "05" || dianXMLFirst(values, "CustomizationID") != "10" {
		t.Fatalf("cabecera de documento soporte inesperada: root=%s tipo=%s custom=%s", root, dianXMLFirst(values, "InvoiceTypeCode"), dianXMLFirst(values, "CustomizationID"))
	}
	if got := dianXMLFirst(values, "ProfileID"); got != "DIAN 2.1: documento soporte en adquisiciones efectuadas a no obligados a facturar." {
		t.Fatalf("ProfileID=%q", got)
	}
	cuds := genericStringValue(generated["cuds"])
	if !facturacionCodigoSHA384Valido(cuds) || facturacionCodigoValidacionDesdeXML(xml) != cuds {
		t.Fatalf("CUDS invalido o no coincide con UUID: %q", cuds)
	}
	for _, fragment := range []string{
		`schemeName="CUDS-SHA384"`,
		`<cbc:PayableAmount currencyID="COP">119.00</cbc:PayableAmount>`,
		`<cbc:TaxableAmount currencyID="COP">19.00</cbc:TaxableAmount><cbc:TaxAmount currencyID="COP">2.85</cbc:TaxAmount>`,
		`<cbc:TaxableAmount currencyID="COP">100.00</cbc:TaxableAmount><cbc:TaxAmount currencyID="COP">11.00</cbc:TaxAmount>`,
		`https://catalogo-vpfe-hab.dian.gov.co/document/searchqr?documentkey=`,
		`NumDS: DS1`, `DocAdq: 800197268`,
	} {
		if !strings.Contains(xml, fragment) {
			t.Fatalf("XML no contiene %q", fragment)
		}
	}
	if strings.Contains(xml, "Software-PIN") || strings.Contains(xml, ">12345<") {
		t.Fatal("el PIN del software no debe quedar expuesto en el XML ni en el QR")
	}
	if strings.Contains(xml, `<cbc:Telephone></cbc:Telephone>`) || strings.Contains(xml, `<cbc:ElectronicMail></cbc:ElectronicMail>`) {
		t.Fatal("los contactos UBL no deben contener telefono o correo vacios")
	}
}

func TestDocumentoSoporteNumeracionSinPrefijoOmiteElementoUBL(t *testing.T) {
	documento, configuracion, empresa, principal := documentoSoportePreflightFixture()
	documento.NumeroLegal = "1"
	configuracion.Prefijo = ""
	if result := buildDocumentoSoporteDIANPreflight(documento, configuracion, empresa, principal); !result.PuedeEmitir {
		t.Fatalf("preflight sin prefijo debe ser valido: %+v", result.Bloqueos)
	}
	fuente, err := buildDocumentoSoporteFuenteFiscal(documento, empresa)
	if err != nil {
		t.Fatal(err)
	}
	cfg := documentoSoporteMergeDIANConfig(principal, documentoSoporteConfigSnapshotFromRow(configuracion))
	generated, status, err := generateDIANUBLDocumentoSoporteDesdeFuenteFiscal(cfg, documento.EmpresaID, map[string]interface{}{
		"documento_codigo": documentoSoporteCodigoInterno(documento.ID), "numero_legal": documento.NumeroLegal,
		"fecha_emision": "2026-08-26T10:15:30-05:00",
	}, fuente)
	if err != nil || status != 200 {
		t.Fatalf("generacion sin prefijo fallo status=%d err=%v", status, err)
	}
	xml := genericStringValue(generated["xml_ubl_base"])
	if strings.Contains(xml, "<sts:Prefix>") || !strings.Contains(xml, "<cbc:ID>1</cbc:ID>") {
		t.Fatalf("el XML sin prefijo no debe emitir sts:Prefix: %s", xml)
	}
}

func TestDocumentoSoporteNoResidenteUsaCustomization11SinDANEVacio(t *testing.T) {
	documento, configuracion, empresa, principal := documentoSoportePreflightFixture()
	documento.NumeroLegal = "DS2"
	documento.TipoDocumento = "PAS"
	documento.Documento = "P-123456"
	documento.VendedorDigitoVerificacion = ""
	documento.VendedorResidencia = "no_residente"
	documento.VendedorPaisCodigo = "US"
	documento.VendedorDepartamento = "Florida"
	documento.VendedorDepartamentoCodigoDANE = ""
	documento.VendedorMunicipio = "Miami"
	documento.VendedorMunicipioCodigoDANE = ""
	fuente, err := buildDocumentoSoporteFuenteFiscal(documento, empresa)
	if err != nil {
		t.Fatal(err)
	}
	cfg := documentoSoporteMergeDIANConfig(principal, documentoSoporteConfigSnapshotFromRow(configuracion))
	generated, status, err := generateDIANUBLDocumentoSoporteDesdeFuenteFiscal(cfg, documento.EmpresaID, map[string]interface{}{
		"documento_codigo": documentoSoporteCodigoInterno(documento.ID), "numero_legal": documento.NumeroLegal,
		"fecha_emision": "2026-08-26T10:15:30-05:00",
	}, fuente)
	if err != nil || status != 200 {
		t.Fatalf("generacion no residente fallo status=%d err=%v", status, err)
	}
	xml := genericStringValue(generated["xml_ubl_base"])
	if !strings.Contains(xml, `<cbc:CustomizationID>11</cbc:CustomizationID>`) || !strings.Contains(xml, `schemeName="41">P123456</cbc:CompanyID>`) {
		t.Fatal("el vendedor extranjero no conserva CustomizationID 11 y pasaporte")
	}
	if strings.Contains(xml, `<cbc:ID></cbc:ID>`) || strings.Contains(xml, `<cbc:CountrySubentityCode></cbc:CountrySubentityCode>`) {
		t.Fatal("la direccion extranjera no debe emitir elementos DANE vacios")
	}
}

func TestDocumentoSoporteAgrupaTarifasPorTributo(t *testing.T) {
	xml := documentoSoporteTaxGroupXML(map[string]*documentoSoporteTaxGroup{
		"iva19": {Code: "01", Name: "IVA", Rate: 19, Base: 100, Amount: 19},
		"iva5":  {Code: "01", Name: "IVA", Rate: 5, Base: 200, Amount: 10},
	}, "COP", "TaxTotal")
	if strings.Count(xml, "<cac:TaxTotal>") != 1 || strings.Count(xml, "<cac:TaxSubtotal>") != 2 || !strings.Contains(xml, `>29.00</cbc:TaxAmount>`) {
		t.Fatalf("agrupacion tributaria inesperada: %s", xml)
	}
}

func TestDocumentoSoporteEstadoDIANMirrorNoPropagaEstadosInternos(t *testing.T) {
	cases := map[string]string{
		"aceptado": "aceptado", "reconciliado": "aceptado", "fallido_terminal": "fallido",
		"rechazado": "rechazado", "no_aplica": "pendiente", "procesando": "pendiente", "desconocido": "fallido",
	}
	for raw, want := range cases {
		if got := documentoSoporteEstadoDIANMirror(raw); got != want {
			t.Fatalf("estado %q=%q, want %q", raw, got, want)
		}
	}
}

func TestDocumentoSoporteListaIdentificacionesVendedorAnexo11(t *testing.T) {
	allowed := []struct{ input, want, rawDocument, wantDocument string }{
		{"TE", "21", "AB-123", "AB123"}, {"CE", "22", "AB-123", "AB123"}, {"NIT", "31", "900-123", "900123"},
		{"PAS", "41", "AB-123", "AB123"}, {"DIE", "42", "AB-123", "AB123"}, {"PEP", "47", "AB-123", "AB123"},
		{"NIT otro pais", "50", "AB-123", "AB123"},
	}
	for _, tc := range allowed {
		got, document, ok := documentoSoporteSellerScheme(tc.input, tc.rawDocument, "no_residente")
		if !ok || got != tc.want || document != tc.wantDocument {
			t.Fatalf("identificacion %q: scheme=%q doc=%q ok=%v", tc.input, got, document, ok)
		}
	}
	if _, _, ok := documentoSoporteSellerScheme("CC", "123", "residente"); ok {
		t.Fatal("tabla 16.2.1 no permite CC para vendedor residente de documento soporte")
	}
}

func TestDocumentoSoporteXSDOficialOpcional(t *testing.T) {
	xsdPath := strings.TrimSpace(os.Getenv("PCS_DIAN_DOCUMENTO_SOPORTE_XSD"))
	pythonPath := strings.TrimSpace(os.Getenv("PCS_PYTHON"))
	if xsdPath == "" || pythonPath == "" {
		t.Skip("configure PCS_DIAN_DOCUMENTO_SOPORTE_XSD y PCS_PYTHON para validar la caja de herramientas oficial")
	}
	documento, configuracion, empresa, _ := documentoSoportePreflightFixture()
	documento.NumeroLegal = "DS1"
	// El ejemplo oficial usa el NIT corto 1020, pero el XSL compilado de la misma
	// caja intenta aplicar modulo 11 solo a longitudes de 5 a 15 y aborta. Para
	// poder ejercer el resto de reglas usamos aqui un NIT colombiano convencional.
	documento.Documento = "900123456"
	documento.VendedorDigitoVerificacion = "8"
	fuente, err := buildDocumentoSoporteFuenteFiscal(documento, empresa)
	if err != nil {
		t.Fatal(err)
	}
	principal := testDIANValidConfig(t, "https://vpfe-hab.dian.gov.co/WcfDianCustomerServices.svc", documento.EmpresaID)
	principal["nit"] = empresa.NIT
	principal["digito_verificacion"] = empresa.DigitoVerificacion
	principal["razon_social"] = empresa.RazonSocial
	cfg := documentoSoporteMergeDIANConfig(principal, documentoSoporteConfigSnapshotFromRow(configuracion))
	generated, status, err := generateDIANUBLDocumentoSoporteDesdeFuenteFiscal(cfg, documento.EmpresaID, map[string]interface{}{
		"documento_codigo": documentoSoporteCodigoInterno(documento.ID), "numero_legal": documento.NumeroLegal,
		"fecha_emision": "2026-08-26T10:15:30-05:00",
	}, fuente)
	if err != nil || status != 200 {
		t.Fatalf("generacion DIAN fallo status=%d err=%v", status, err)
	}
	signed, signStatus, err := signDIANXMLXAdESBase(cfg, documento.EmpresaID, map[string]interface{}{
		"documento_codigo": documentoSoporteCodigoInterno(documento.ID), "xml_ubl_base": genericStringValue(generated["xml_ubl_base"]),
	})
	if err != nil || signStatus != 200 {
		t.Fatalf("firma XAdES fallo status=%d err=%v", signStatus, err)
	}
	xmlPath := filepath.Join(t.TempDir(), "documento-soporte-firmado.xml")
	if err := os.WriteFile(xmlPath, []byte(genericStringValue(signed["xml_firmado"])), 0o600); err != nil {
		t.Fatal(err)
	}
	script := `import sys
from lxml import etree
schema = etree.XMLSchema(etree.parse(sys.argv[1]))
document = etree.parse(sys.argv[2])
schema.assertValid(document)`
	command := exec.Command(pythonPath, "-c", script, xsdPath, xmlPath)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("XSD oficial rechazo el XML: %v\n%s", err, strings.TrimSpace(string(output)))
	}
	javaPath := strings.TrimSpace(os.Getenv("PCS_JAVA"))
	saxonPath := strings.TrimSpace(os.Getenv("PCS_DIAN_SAXON_JAR"))
	schematronPath := strings.TrimSpace(os.Getenv("PCS_DIAN_DOCUMENTO_SOPORTE_XSL"))
	if javaPath == "" || saxonPath == "" || schematronPath == "" {
		return
	}
	reportPath := filepath.Join(t.TempDir(), "documento-soporte-schematron.xml")
	saxonClass := strings.TrimSpace(os.Getenv("PCS_DIAN_SAXON_CLASS"))
	if saxonClass == "" {
		saxonClass = "net.sf.saxon.Transform"
	}
	if saxonClass == "com.icl.saxon.StyleSheet" {
		command = exec.Command(javaPath, "-cp", saxonPath, saxonClass, "-o", reportPath, xmlPath, schematronPath)
	} else {
		command = exec.Command(javaPath, "-cp", saxonPath, saxonClass,
			"-s:"+xmlPath, "-xsl:"+schematronPath, "-o:"+reportPath)
	}
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("no se pudo ejecutar Schematron oficial: %v\n%s", err, strings.TrimSpace(string(output)))
	}
	// La caja 1.1 es internamente contradictoria: sus XML de ejemplo usan el
	// ProfileID largo, CUDS-SHA384 e InvoiceTypeCode 05, mientras el unico XSL
	// incluido es el modelo de factura de venta y los rechaza como FAD03, FAD08 y
	// FAD12. Se registran solo esas incompatibilidades conocidas; cualquier otra
	// regla fatal continua cerrando la prueba.
	knownToolboxMismatch := false
	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSpace(line)
		if !strings.Contains(line, "Fatal:") {
			continue
		}
		if strings.Contains(line, "[FAD03]") || strings.Contains(line, "[FAD08]") || strings.Contains(line, "[FAD12]") {
			knownToolboxMismatch = true
			continue
		}
		if strings.Contains(line, "Fatal:") {
			t.Fatalf("Schematron oficial reporto una regla fatal inesperada:\n%s", strings.TrimSpace(string(output)))
		}
	}
	if knownToolboxMismatch {
		t.Log("caja DIAN 1.1 conserva un XSL de factura de venta incompatible con sus ejemplos de documento soporte")
	}
	report, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatal(err)
	}
	reportText := string(report)
	if strings.Contains(reportText, `failed-assert flag="fatal"`) || strings.Contains(reportText, `flag="fatal" role="fatal"`) {
		t.Fatalf("Schematron oficial reporto reglas fatales:\n%s", reportText)
	}
}
