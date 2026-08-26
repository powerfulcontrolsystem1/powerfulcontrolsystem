package handlers

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	dbpkg "github.com/you/pos-backend/db"
)

func TestBuildDIANNominaCUNEFormulaAnexoConCamposPublicados(t *testing.T) {
	got := buildDIANNominaCUNE(
		"N00001", "2020-01-16", "10:53:10-05:00",
		"3500000.00", "1000000.00", "2500000.00",
		"700085371", "800199436", "693", "1",
	)
	// La Resolucion 000013 imprime 16560d... para estos campos, pero ese
	// digest no corresponde a su propia cadena. Este valor es el SHA-384
	// reproducible de la formula normativa, incluida HoraGen sin alterarla.
	const want = "dae35b4dfdf10939502c96278feb97be1af6cd5653bd662ca29b5c9cdf2729464b51d272ba0008586cf153b46f87a2bd"
	if got != want {
		t.Fatalf("CUNE=%q, want %q", got, want)
	}
}

func TestGenerateDIANNominaIndividualCumpleContratoOficial(t *testing.T) {
	source := nominaDIANFuenteFixture()
	generated, status, err := generateDIANNominaIndividualDesdeFuente(source, "693")
	if err != nil || status != http.StatusOK {
		t.Fatalf("generacion de NominaIndividual fallo status=%d err=%v", status, err)
	}
	xmlPayload := genericStringValue(generated["xml_ubl_base"])
	var root struct {
		XMLName xml.Name
	}
	if err := xml.Unmarshal([]byte(xmlPayload), &root); err != nil {
		t.Fatalf("XML de nomina no esta bien formado: %v", err)
	}
	if root.XMLName.Local != "NominaIndividual" || root.XMLName.Space != nominaElectronicaNamespace {
		t.Fatalf("raiz de nomina inesperada: %+v", root.XMLName)
	}
	for _, fragment := range []string{
		`SchemaLocation=""`,
		`<NumeroSecuenciaXML CodigoTrabajador="EMP-1" Prefijo="N" Consecutivo="1" Numero="N00001"/>`,
		`Version="V1.0: Documento Soporte de Pago de Nómina Electrónica"`,
		`Ambiente="1" TipoXML="102"`,
		`EncripCUNE="CUNE-SHA384"`,
		`<Pago Forma="1" Metodo="42"`,
		`<DevengadosTotal>3500000.00</DevengadosTotal>`,
		`<DeduccionesTotal>1000000.00</DeduccionesTotal>`,
		`<ComprobanteTotal>2500000.00</ComprobanteTotal>`,
		`https://catalogo-vpfe.dian.gov.co/document/searchqr?documentkey=`,
	} {
		if !strings.Contains(xmlPayload, fragment) {
			t.Fatalf("NominaIndividual no contiene %q", fragment)
		}
	}
	if got := genericStringValue(generated["cune"]); got != "dae35b4dfdf10939502c96278feb97be1af6cd5653bd662ca29b5c9cdf2729464b51d272ba0008586cf153b46f87a2bd" {
		t.Fatalf("CUNE generado=%q", got)
	}
	if genericStringValue(generated["documento_codigo"]) != "NE-NOMINA-91" || genericStringValue(generated["soap_operacion"]) != "SendNominaSync" {
		t.Fatalf("metadatos de nomina inesperados: %#v", generated)
	}
	if strings.Contains(xmlPayload, "693") || strings.Contains(xmlPayload, "Software-PIN") {
		t.Fatal("el PIN del software no debe quedar expuesto en NominaIndividual")
	}
}

func TestGenerateDIANNominaIndividualConservaTodasLasFechasPagoMensuales(t *testing.T) {
	source := nominaDIANFuenteFixture()
	source.LiquidacionIDs = []int64{91, 92}
	source.PagoIDs = []int64{101, 102}
	source.FechasPago = []string{"2020-01-16", "2020-01-31"}
	source.Pago.FechaPago = source.FechasPago[0]
	generated, status, err := generateDIANNominaIndividualDesdeFuente(source, "693")
	if err != nil || status != http.StatusOK {
		t.Fatalf("generación mensual con dos pagos falló status=%d err=%v", status, err)
	}
	xmlPayload := genericStringValue(generated["xml_ubl_base"])
	if got := strings.Count(xmlPayload, "<FechaPago>"); got != 2 {
		t.Fatalf("FechasPagos=%d, se esperaban 2: %s", got, xmlPayload)
	}
	if !strings.Contains(xmlPayload, "<FechaPago>2020-01-16</FechaPago><FechaPago>2020-01-31</FechaPago>") {
		t.Fatal("NominaIndividual no conservó las fechas de pago ordenadas")
	}
}

func TestDIANNominaMoneyTruncaSinRedondear(t *testing.T) {
	if got := dianNominaMoney("123.999"); got != "123.99" {
		t.Fatalf("monto truncado=%q", got)
	}
}

func TestGenerateDIANNominaIndividualBloqueaHorasSinIntervalos(t *testing.T) {
	source := nominaDIANFuenteFixture()
	source.Devengados.TieneHorasSinTrazabilidad = true
	if _, status, err := generateDIANNominaIndividualDesdeFuente(source, "693"); err == nil || status != http.StatusUnprocessableEntity || !strings.Contains(err.Error(), "intervalos") {
		t.Fatalf("las horas agregadas sin intervalos deben bloquear: status=%d err=%v", status, err)
	}
}

func TestSignDIANNominaIndividualInsertaXAdES(t *testing.T) {
	source := nominaDIANFuenteFixture()
	generated, status, err := generateDIANNominaIndividualDesdeFuente(source, "693")
	if err != nil || status != http.StatusOK {
		t.Fatalf("generacion de NominaIndividual fallo status=%d err=%v", status, err)
	}
	cfg := testDIANValidConfig(t, "")
	signed, signStatus, signErr := signDIANXMLXAdESBase(cfg, source.EmpresaID, map[string]interface{}{
		"documento_codigo": genericStringValue(generated["documento_codigo"]),
		"xml_ubl_base":     genericStringValue(generated["xml_ubl_base"]),
	})
	if signErr != nil || signStatus != http.StatusOK {
		t.Fatalf("firma XAdES de NominaIndividual fallo status=%d err=%v", signStatus, signErr)
	}
	xmlSigned := genericStringValue(signed["xml_firmado"])
	for _, fragment := range []string{
		`<NominaIndividual xmlns="` + nominaElectronicaNamespace + `"`,
		`<ext:ExtensionContent><ds:Signature`,
		`<xades:QualifyingProperties`,
		`<ds:SignatureValue `,
	} {
		if !strings.Contains(xmlSigned, fragment) {
			t.Fatalf("nomina firmada no contiene %q", fragment)
		}
	}
	if strings.Contains(xmlSigned, `xmlns:cac=`) || strings.Contains(xmlSigned, `xmlns:cbc=`) || strings.Contains(xmlSigned, `xmlns:sts=`) {
		t.Fatal("la firma de nomina no debe convertir la raiz al perfil UBL de factura")
	}
}

func TestValidateDIANNominaDocumentPreflightFirmadoYDetectaTotalAlterado(t *testing.T) {
	source := nominaDIANFuenteFixture()
	generated, status, err := generateDIANNominaIndividualDesdeFuente(source, "693")
	if err != nil || status != http.StatusOK {
		t.Fatalf("generación de NominaIndividual falló status=%d err=%v", status, err)
	}
	cfg := testDIANValidConfig(t, "https://vpfe.dian.gov.co/WcfDianCustomerServices.svc")
	cfg["tipo_ambiente"] = "produccion"
	cfg["software_id"] = source.SoftwareID
	cfg["software_pin"] = "693"
	signed, signStatus, signErr := signDIANXMLXAdESBase(cfg, source.EmpresaID, map[string]interface{}{
		"documento_codigo": genericStringValue(generated["documento_codigo"]),
		"xml_ubl_base":     genericStringValue(generated["xml_ubl_base"]),
	})
	if signErr != nil || signStatus != http.StatusOK {
		t.Fatalf("firma XAdES de NominaIndividual falló status=%d err=%v", signStatus, signErr)
	}
	xmlSigned := genericStringValue(signed["xml_firmado"])
	preflight := validateDIANNominaDocumentPreflight(cfg, source.EmpresaID, source, xmlSigned, "pre_envio")
	if blocked, _ := preflight["bloqueado"].(bool); blocked {
		t.Fatalf("preflight firmado válido fue bloqueado: %#v", preflight["issues"])
	}
	tampered := strings.Replace(xmlSigned, "<ComprobanteTotal>2500000.00</ComprobanteTotal>", "<ComprobanteTotal>2500001.00</ComprobanteTotal>", 1)
	preflight = validateDIANNominaDocumentPreflight(cfg, source.EmpresaID, source, tampered, "pre_envio")
	if blocked, _ := preflight["bloqueado"].(bool); !blocked {
		t.Fatal("preflight permitió un ComprobanteTotal distinto de la fuente fiscal")
	}
	if !strings.Contains(strings.ToLower(fmt.Sprint(preflight["issues"])), "comprobantetotal") {
		t.Fatalf("preflight no identificó el total alterado: %#v", preflight["issues"])
	}
}

func TestValidateDIANNominaPreflightBloqueaEndpointDeAmbienteCruzado(t *testing.T) {
	source := nominaDIANFuenteFixture()
	source.TipoAmbiente = "habilitacion"
	cfg := testDIANValidConfig(t, "https://vpfe.dian.gov.co/WcfDianCustomerServices.svc")
	cfg["tipo_ambiente"] = "habilitacion"
	cfg["software_id"] = source.SoftwareID
	cfg["software_pin"] = "693"
	preflight := validateDIANNominaDocumentPreflight(cfg, source.EmpresaID, source, "", "reserva")
	raw := fmt.Sprint(preflight["issues"])
	if !strings.Contains(raw, "DIAN-NOM-ENDPOINT-AMBIENTE") {
		t.Fatalf("habilitación con endpoint productivo no se bloqueó: %#v", preflight)
	}

	cfg["url_dian"] = "https://vpfe-hab.dian.gov.co/WcfDianCustomerServices.svc"
	preflight = validateDIANNominaDocumentPreflight(cfg, source.EmpresaID, source, "", "reserva")
	if strings.Contains(fmt.Sprint(preflight["issues"]), "DIAN-NOM-ENDPOINT-AMBIENTE") {
		t.Fatalf("endpoint oficial de habilitación fue rechazado: %#v", preflight)
	}
}

func TestValidateDIANSoftwareProviderPayload(t *testing.T) {
	valid := map[string]interface{}{
		"software_proveedor_nit": " 900123456 ", "software_proveedor_dv": "8",
		"software_proveedor_razon_social": " Proveedor SAS ",
	}
	if err := validateDIANSoftwareProviderPayload(valid); err != nil {
		t.Fatalf("proveedor jurídico válido fue rechazado: %v", err)
	}
	if valid["software_proveedor_nit"] != "900123456" || valid["software_proveedor_razon_social"] != "Proveedor SAS" {
		t.Fatalf("proveedor no fue normalizado: %#v", valid)
	}
	for name, payload := range map[string]map[string]interface{}{
		"nit_no_numerico": {"software_proveedor_nit": "900-1", "software_proveedor_dv": "8", "software_proveedor_razon_social": "Proveedor"},
		"dv_no_numerico":  {"software_proveedor_nit": "9001", "software_proveedor_dv": "X", "software_proveedor_razon_social": "Proveedor"},
		"dv_multiple":     {"software_proveedor_nit": "9001", "software_proveedor_dv": "12", "software_proveedor_razon_social": "Proveedor"},
		"sin_identidad":   {"software_proveedor_nit": "9001", "software_proveedor_dv": "8", "software_proveedor_razon_social": ""},
	} {
		t.Run(name, func(t *testing.T) {
			if err := validateDIANSoftwareProviderPayload(payload); err == nil {
				t.Fatal("configuración incompleta del proveedor fue aceptada")
			}
		})
	}
	if err := validateDIANSoftwareProviderPayload(map[string]interface{}{"observaciones": "sin cambios de proveedor"}); err != nil {
		t.Fatalf("actualización ajena al proveedor no debe romper configuración legada: %v", err)
	}
}

func TestValidateNominaElectronicaEmployerDIANConfigExigeMismaIdentidad(t *testing.T) {
	source := nominaDIANFuenteFixture()
	cfg := map[string]interface{}{"nit": source.Empleador.NIT, "digito_verificacion": source.Empleador.DV}
	if err := validateNominaElectronicaEmployerDIANConfig(cfg, source); err != nil {
		t.Fatalf("identidad coincidente fue rechazada: %v", err)
	}
	cfg["nit"] = "900999888"
	if err := validateNominaElectronicaEmployerDIANConfig(cfg, source); err == nil || !strings.Contains(err.Error(), "no coincide") {
		t.Fatalf("identidad DIAN divergente no fue bloqueada: %v", err)
	}
	cfg["nit"] = source.Empleador.NIT
	cfg["digito_verificacion"] = source.Empleador.DV + "X"
	if err := validateNominaElectronicaEmployerDIANConfig(cfg, source); err == nil || !strings.Contains(err.Error(), "completos") {
		t.Fatalf("DV principal malformado no fue bloqueado: %v", err)
	}
}

func TestDIANNominaSOAPUsaContratoSendNominaSync(t *testing.T) {
	payload := map[string]interface{}{
		"documento_tipo": "nomina_electronica",
		"soap_operacion": "SendBillSync",
	}
	if got := dianSOAPOperationFromPayload(payload, "habilitacion", "test-set-no-aplica"); got != "SendNominaSync" {
		t.Fatalf("operacion SOAP de nomina=%q", got)
	}
	body := dianBuildSOAPBody("SendNominaSync", "nomina.zip", []byte("zip"), "test-set-no-aplica")
	if !strings.Contains(body, `<wcf:SendNominaSync><wcf:contentFile>emlw</wcf:contentFile></wcf:SendNominaSync>`) {
		t.Fatalf("body SendNominaSync inesperado: %s", body)
	}
	if strings.Contains(body, "fileName") || strings.Contains(body, "testSetId") {
		t.Fatalf("SendNominaSync solo admite contentFile segun WSDL: %s", body)
	}
	envelope := buildDIANSOAPEnvelope("SendNominaSync", "https://vpfe-hab.dian.gov.co/WcfDianCustomerServices.svc", "nomina.zip", []byte("zip"), "")
	if !strings.Contains(envelope, `IWcfDianCustomerServices/SendNominaSync`) || !strings.Contains(envelope, `<wcf:SendNominaSync>`) {
		t.Fatalf("envelope SOAP de nomina inesperado: %s", envelope)
	}
}

func TestDIANNominaIndividualXSDOficialOpcional(t *testing.T) {
	xsdPath := strings.TrimSpace(os.Getenv("PCS_DIAN_NOMINA_XSD"))
	schemesPath := strings.TrimSpace(os.Getenv("PCS_DIAN_NOMINA_SCHEMES"))
	pythonPath := strings.TrimSpace(os.Getenv("PCS_PYTHON"))
	if xsdPath == "" || schemesPath == "" || pythonPath == "" {
		t.Skip("configure PCS_DIAN_NOMINA_XSD, PCS_DIAN_NOMINA_SCHEMES y PCS_PYTHON para validar la caja oficial")
	}
	source := nominaDIANFuenteFixture()
	generated, status, err := generateDIANNominaIndividualDesdeFuente(source, "693")
	if err != nil || status != http.StatusOK {
		t.Fatalf("generacion de NominaIndividual fallo status=%d err=%v", status, err)
	}
	cfg := testDIANValidConfig(t, "")
	signed, signStatus, signErr := signDIANXMLXAdESBase(cfg, source.EmpresaID, map[string]interface{}{
		"documento_codigo": genericStringValue(generated["documento_codigo"]),
		"xml_ubl_base":     genericStringValue(generated["xml_ubl_base"]),
	})
	if signErr != nil || signStatus != http.StatusOK {
		t.Fatalf("firma XAdES de NominaIndividual fallo status=%d err=%v", signStatus, signErr)
	}
	xmlPath := filepath.Join(t.TempDir(), "nomina-individual-firmada.xml")
	if err := os.WriteFile(xmlPath, []byte(genericStringValue(signed["xml_firmado"])), 0o600); err != nil {
		t.Fatal(err)
	}
	script := `import pathlib, sys
from lxml import etree
xsd_path = pathlib.Path(sys.argv[1]).resolve()
schemes = pathlib.Path(sys.argv[2]).resolve()
tree = etree.parse(str(xsd_path))
ns = {'xsd': 'http://www.w3.org/2001/XMLSchema'}
for node in tree.xpath('//xsd:import[@schemaLocation]', namespaces=ns):
    if node.get('schemaLocation', '').startswith('../common/'):
        node.set('schemaLocation', (schemes / pathlib.Path(node.get('schemaLocation')).name).as_uri())
schema = etree.XMLSchema(tree)
schema.assertValid(etree.parse(sys.argv[3]))`
	command := exec.Command(pythonPath, "-c", script, xsdPath, schemesPath, xmlPath)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("XSD oficial rechazo NominaIndividual: %v\n%s", err, strings.TrimSpace(string(output)))
	}
}

func nominaDIANFuenteFixture() *dbpkg.EmpresaNominaDIANFuente {
	return &dbpkg.EmpresaNominaDIANFuente{
		Esquema: "pcs.nomina.fuente_fiscal", Version: 2, EmpresaID: 12,
		LiquidacionID: 91, LiquidacionIDs: []int64{91}, EmpleadoID: 71, EmpleadoNominaID: 81,
		PagoID: 101, PagoIDs: []int64{101}, FechasPago: []string{"2020-01-16"}, PeriodoReporte: "2020-01",
		PeriodoDesde: "2020-01-01", PeriodoHasta: "2020-01-15", FechaIngreso: "2014-10-28",
		TiempoLaborado: 1877, PeriodoNomina: 5, Prefijo: "N", Consecutivo: 1, NumeroLegal: "N00001",
		FechaEmisionLegal: "2020-01-16T10:53:10-05:00", TipoAmbiente: "produccion", SoftwareID: "software-id-nomina",
		Empleador: dbpkg.EmpresaNominaDIANParte{
			RazonSocial: "EMPLEADOR SAS", NIT: "700085371", DV: "2", Pais: "CO",
			Departamento: "11", Municipio: "11001", Direccion: "Calle 1 2 3",
		},
		ProveedorXML: dbpkg.EmpresaNominaDIANParte{RazonSocial: "PROVEEDOR SOFTWARE SAS", NIT: "900123456", DV: "8"},
		Trabajador: dbpkg.EmpresaNominaDIANTrabajador{
			TipoTrabajador: "01", SubTipoTrabajador: "00", TipoDocumento: "13", NumeroDocumento: "800199436",
			PrimerApellido: "PEREZ", SegundoApellido: "GOMEZ", PrimerNombre: "ANA", LugarTrabajoPais: "CO",
			LugarTrabajoDepartamento: "11", LugarTrabajoMunicipio: "11001", LugarTrabajoDireccion: "Carrera 4 5 6",
			TipoContrato: "1", Sueldo: 3500000, CodigoTrabajador: "EMP-1",
		},
		Pago:       dbpkg.EmpresaNominaDIANPagoFuente{FechaPago: "2020-01-16", Forma: "1", Metodo: "42", Banco: "BANCO", NetoPagado: 2500000},
		Devengados: dbpkg.EmpresaNominaDIANDevengados{DiasTrabajados: 15, SueldoTrabajado: 3500000, Total: 3500000},
		Deducciones: dbpkg.EmpresaNominaDIANDeducciones{
			PorcentajeSalud: 4, Salud: 140000, PorcentajePension: 4, Pension: 140000, OtrasDeducciones: 720000, Total: 1000000,
		},
		ComprobanteTotal: 2500000,
	}
}
