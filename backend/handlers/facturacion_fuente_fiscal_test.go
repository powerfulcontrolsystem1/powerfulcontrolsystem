package handlers

import (
	"bytes"
	"os"
	"strings"
	"testing"

	dbpkg "github.com/you/pos-backend/db"
)

func TestBuildFacturacionFuenteFiscalSnapshotUsesRealCartLinesAndParties(t *testing.T) {
	carrito, items, cfg, cliente, doc := facturacionFuenteFiscalFixture()
	snapshot, err := buildFacturacionFuenteFiscalSnapshot(carrito, items, cfg, cliente, doc)
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.EmpresaID != 12 || snapshot.Documento.CodigoOrigen != "CP-VENTA-CRT-77-PG-20260823123000" || snapshot.Documento.CodigoDestino != "FV-VENTA-CRT-77-PG-20260823123000" {
		t.Fatalf("documento fuente inesperado: %+v", snapshot.Documento)
	}
	if len(snapshot.Lineas) != 2 {
		t.Fatalf("lineas=%d, se esperaban 2", len(snapshot.Lineas))
	}
	if snapshot.Lineas[0].ItemID != 9 || snapshot.Lineas[0].CodigoItem != "SKU-CAFE" || snapshot.Lineas[0].Descripcion != "Cafe colombiano" {
		t.Fatalf("primera linea no conserva el item real: %+v", snapshot.Lineas[0])
	}
	if snapshot.Lineas[1].ItemID != 10 || snapshot.Lineas[1].CodigoItem != "SKU-PAN" || snapshot.Lineas[1].Descripcion != "Pan artesanal" {
		t.Fatalf("segunda linea no conserva el item real: %+v", snapshot.Lineas[1])
	}
	if snapshot.Emisor.NumeroDocumento != "900123456" || snapshot.Emisor.NombreRazonSocial != "Empresa Real SAS" {
		t.Fatalf("emisor inesperado: %+v", snapshot.Emisor)
	}
	if snapshot.Cliente.ID != 22 || snapshot.Cliente.NumeroDocumento != "84456779" || snapshot.Cliente.NombreRazonSocial != "Cliente Real" {
		t.Fatalf("cliente inesperado: %+v", snapshot.Cliente)
	}
	if snapshot.Totales.BaseGravableLineas != 200 || snapshot.Totales.ImpuestoLineas != 38 || snapshot.Totales.TotalLineas != 238 {
		t.Fatalf("totales inesperados: %+v", snapshot.Totales)
	}
	raw, err := marshalFacturacionFuenteFiscal(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	for _, fabricated := range []string{"Servicio de habilitacion DIAN", "PCS-DIAN-001", "7700000000019", "cliente@example.com", "11001"} {
		if bytes.Contains(raw, []byte(fabricated)) {
			t.Fatalf("el snapshot contiene default fabricado %q: %s", fabricated, raw)
		}
	}
	if snapshot.Emisor.DepartamentoCodigoDANE != "25" || snapshot.Emisor.MunicipioCodigoDANE != "25175" || snapshot.Cliente.DepartamentoCodigoDANE != "25" || snapshot.Cliente.MunicipioCodigoDANE != "25175" {
		t.Fatalf("los codigos DANE reales no se conservaron: emisor=%+v cliente=%+v", snapshot.Emisor, snapshot.Cliente)
	}
	if len(snapshot.Bloqueantes) != 0 {
		t.Fatalf("fixture fiscal completa no debe quedar bloqueada: %v", snapshot.Bloqueantes)
	}
}

func TestBuildFacturacionFuenteFiscalSnapshotBlocksMissingOrInconsistentDANECodes(t *testing.T) {
	carrito, items, cfg, cliente, doc := facturacionFuenteFiscalFixture()
	cfg.DepartamentoCodigoDANE = ""
	cfg.MunicipioCodigoDANE = "11001"
	cliente.DepartamentoCodigoDANE = "25"
	cliente.MunicipioCodigoDANE = "11001"
	snapshot, err := buildFacturacionFuenteFiscalSnapshot(carrito, items, cfg, cliente, doc)
	if err != nil {
		t.Fatal(err)
	}
	for _, blocker := range []string{
		"emisor.departamento_codigo_dane_faltante", "emisor.municipio_codigo_dane_faltante",
		"cliente.municipio_codigo_dane_faltante",
	} {
		if !facturacionFuenteFiscalContains(snapshot.Bloqueantes, blocker) {
			t.Fatalf("falta bloqueante %q en %v", blocker, snapshot.Bloqueantes)
		}
	}
}

func TestBuildFacturacionFuenteFiscalSnapshotBlocksMissingItemCodeAndNonColombianParty(t *testing.T) {
	carrito, items, cfg, cliente, doc := facturacionFuenteFiscalFixture()
	items[0].CodigoItem = ""
	cliente.Pais = "US"
	snapshot, err := buildFacturacionFuenteFiscalSnapshot(carrito, items, cfg, cliente, doc)
	if err != nil {
		t.Fatal(err)
	}
	for _, blocker := range []string{"lineas.2.codigo_item_faltante", "cliente.pais_codigo_colombia_requerido"} {
		if !facturacionFuenteFiscalContains(snapshot.Bloqueantes, blocker) {
			t.Fatalf("falta bloqueante %q en %v", blocker, snapshot.Bloqueantes)
		}
	}
}

func TestBuildFacturacionFuenteFiscalSnapshotBlocksWrongIssuerNITDV(t *testing.T) {
	carrito, items, cfg, cliente, doc := facturacionFuenteFiscalFixture()
	cfg.DigitoVerificacion = "7"
	snapshot, err := buildFacturacionFuenteFiscalSnapshot(carrito, items, cfg, cliente, doc)
	if err != nil {
		t.Fatal(err)
	}
	if !facturacionFuenteFiscalContains(snapshot.Bloqueantes, "emisor.digito_verificacion_invalido") {
		t.Fatalf("DV incorrecto debe bloquear la fuente fiscal: %v", snapshot.Bloqueantes)
	}
}

func TestBuildFacturacionFuenteFiscalSnapshotIsDeterministic(t *testing.T) {
	carrito, items, cfg, cliente, doc := facturacionFuenteFiscalFixture()
	first, err := buildFacturacionFuenteFiscalSnapshot(carrito, items, cfg, cliente, doc)
	if err != nil {
		t.Fatal(err)
	}
	items[0], items[1] = items[1], items[0]
	second, err := buildFacturacionFuenteFiscalSnapshot(carrito, items, cfg, cliente, doc)
	if err != nil {
		t.Fatal(err)
	}
	firstJSON, err := marshalFacturacionFuenteFiscal(first)
	if err != nil {
		t.Fatal(err)
	}
	secondJSON, err := marshalFacturacionFuenteFiscal(second)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(firstJSON, secondJSON) {
		t.Fatalf("el mismo origen debe producir el mismo hash\nfirst=%s\nsecond=%s", firstJSON, secondJSON)
	}
}

func TestGenerateDIANUBLDesdeFuenteFiscalUsesRealMultiLineData(t *testing.T) {
	carrito, items, cfgMaestro, cliente, doc := facturacionFuenteFiscalFixture()
	snapshot, err := buildFacturacionFuenteFiscalSnapshot(carrito, items, cfgMaestro, cliente, doc)
	if err != nil {
		t.Fatal(err)
	}
	cfg := testDIANValidConfig(t, "")
	cfg["nit"] = "900123456"
	cfg["digito_verificacion"] = "7"
	cfg["prefijo"] = "FV"
	cfg["resolucion_numero"] = "18760000001"
	cfg["resolucion_fecha_desde"] = "2026-01-01"
	cfg["resolucion_fecha_hasta"] = "2027-12-31"
	cfg["llave_tecnica"] = "technical-key"
	result, status, err := generateDIANUBLBase(cfg, 12, map[string]interface{}{
		"documento_tipo": "factura_electronica", "documento_codigo": snapshot.Documento.CodigoDestino,
		"numero_legal": "FV100", "fecha_emision": "2026-08-24T09:30:00-05:00", "total": "238.00",
	}, snapshot)
	if err != nil || status != 200 {
		t.Fatalf("generacion real fallo status=%d err=%v", status, err)
	}
	xmlBase := genericStringValue(result["xml_ubl_base"])
	signed, signStatus, signErr := signDIANXMLXAdESBase(cfg, 12, map[string]interface{}{"xml_ubl_base": xmlBase, "documento_codigo": "FV100"})
	if signErr != nil || signStatus != 200 {
		t.Fatalf("firma del UBL real fallo status=%d err=%v", signStatus, signErr)
	}
	xml := genericStringValue(signed["xml_firmado"])
	if outputPath := strings.TrimSpace(os.Getenv("PCS_TEST_DIAN_XML_OUTPUT")); outputPath != "" {
		if err := os.WriteFile(outputPath, []byte(xml), 0o600); err != nil {
			t.Fatalf("guardar XML de QA: %v", err)
		}
	}
	for _, expected := range []string{"<cbc:ProfileID>DIAN 2.1</cbc:ProfileID>", "<cbc:LineCountNumeric>2</cbc:LineCountNumeric>", "Cafe colombiano", "Pan artesanal", "SKU-CAFE", "SKU-PAN", "25175", "Carrera 5 6-7", "Calle 10 20-30", `<cbc:TaxLevelCode listName="05">O-13</cbc:TaxLevelCode>`, `<cbc:TaxLevelCode listName="04">R-99-PN</cbc:TaxLevelCode>`} {
		if !strings.Contains(xml, expected) {
			t.Fatalf("UBL real no contiene %q", expected)
		}
	}
	if strings.Count(xml, "<cac:InvoiceLine>") != 2 {
		t.Fatalf("se esperaban dos lineas reales: %s", xml)
	}
	for _, fabricated := range []string{"Servicio de habilitacion DIAN", "PCS-DIAN-001", "7700000000019", "cliente@example.com", "Direccion registrada en la empresa", "Direccion del adquiriente"} {
		if strings.Contains(xml, fabricated) {
			t.Fatalf("UBL comercial contiene default fabricado %q", fabricated)
		}
	}
	if !strings.Contains(xml, `<cbc:AllowanceTotalAmount currencyID="COP">0.00</cbc:AllowanceTotalAmount>`) {
		t.Fatalf("los descuentos de linea no deben repetirse como descuento global: %s", xml)
	}
}

func TestDIANFuenteFiscalMonetaryTotalDoesNotDoubleSubtractLineDiscounts(t *testing.T) {
	snapshot := &facturacionFuenteFiscalSnapshot{Totales: facturacionFuenteFiscalTotales{
		BaseGravableLineas: 90, DescuentoLineas: 10, ImpuestoLineas: 17.10, TotalDocumentoOrigen: 107.10,
	}}
	xml := dianFuenteFiscalMonetaryTotalXML(snapshot, "COP")
	if !strings.Contains(xml, `<cbc:AllowanceTotalAmount currencyID="COP">0.00</cbc:AllowanceTotalAmount>`) ||
		!strings.Contains(xml, `<cbc:PayableAmount currencyID="COP">107.10</cbc:PayableAmount>`) {
		t.Fatalf("totales UBL inconsistentes: %s", xml)
	}
}

func TestDIANFuenteFiscalLineDiscountUsesPercentageAndRealItemCode(t *testing.T) {
	lines := []facturacionFuenteFiscalLinea{{
		Numero: 1, CodigoItem: "SKU-DESC", Descripcion: "Producto con descuento", UnidadMedida: "EA",
		Cantidad: 1, PrecioUnitario: 100, DescuentoPorcentaje: 10, ValorDescuento: 10,
		BaseGravable: 90, ImpuestoCodigo: "01", ImpuestoPorcentaje: 19, ValorImpuesto: 17.10,
		TotalLinea: 107.10,
	}}
	xml, err := dianFuenteFiscalLinesXML(lines, "COP")
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{"<cbc:MultiplierFactorNumeric>10.000000</cbc:MultiplierFactorNumeric>", "<cbc:ID>SKU-DESC</cbc:ID>"} {
		if !strings.Contains(xml, expected) {
			t.Fatalf("linea UBL no contiene %q: %s", expected, xml)
		}
	}
}

func TestBuildFacturacionFuenteFiscalSnapshotRejectsCrossTenantData(t *testing.T) {
	carrito, items, cfg, cliente, doc := facturacionFuenteFiscalFixture()
	items[0].EmpresaID = 99
	if _, err := buildFacturacionFuenteFiscalSnapshot(carrito, items, cfg, cliente, doc); err == nil {
		t.Fatal("se esperaba rechazo de linea de otra empresa")
	}
	carrito, items, cfg, cliente, doc = facturacionFuenteFiscalFixture()
	cliente.EmpresaID = 99
	if _, err := buildFacturacionFuenteFiscalSnapshot(carrito, items, cfg, cliente, doc); err == nil {
		t.Fatal("se esperaba rechazo de cliente de otra empresa")
	}
	carrito, items, cfg, cliente, doc = facturacionFuenteFiscalFixture()
	cfg.EmpresaID = 99
	if _, err := buildFacturacionFuenteFiscalSnapshot(carrito, items, cfg, cliente, doc); err == nil {
		t.Fatal("se esperaba rechazo de emisor de otra empresa")
	}
}

func TestBuildFacturacionFuenteFiscalSnapshotMarksUnreconciledTotals(t *testing.T) {
	carrito, items, cfg, cliente, doc := facturacionFuenteFiscalFixture()
	carrito.DescuentoTotal = 15
	carrito.ImpuestoTotal = 30
	doc.MontoTotal = 220
	snapshot, err := buildFacturacionFuenteFiscalSnapshot(carrito, items, cfg, cliente, doc)
	if err != nil {
		t.Fatal(err)
	}
	for _, blocker := range []string{
		"totales.descuento_global_sin_asignar", "totales.impuestos_no_conciliados",
		"totales.lineas_no_concilian_documento", "totales.carrito_no_concilia_documento",
	} {
		if !facturacionFuenteFiscalContains(snapshot.Bloqueantes, blocker) {
			t.Fatalf("falta bloqueante %q en %v", blocker, snapshot.Bloqueantes)
		}
	}
}

func facturacionFuenteFiscalFixture() (*dbpkg.CarritoCompra, []dbpkg.CarritoCompraItem, *dbpkg.EmpresaConfiguracionAvanzada, *dbpkg.Cliente, dbpkg.EmpresaDocumentoFacturacion) {
	carrito := &dbpkg.CarritoCompra{
		ID: 77, EmpresaID: 12, Codigo: "VENTA", ClienteID: 22, Moneda: "COP",
		Subtotal: 200, DescuentoTotal: 0, ImpuestoTotal: 38, Total: 238,
		PagadoEn: "2026-08-23 12:30:00", MetodoPago: "efectivo", ReferenciaPago: "CAJA-1",
	}
	items := []dbpkg.CarritoCompraItem{
		{ID: 10, EmpresaID: 12, CarritoID: 77, TipoItem: "producto", ReferenciaID: 102, CodigoItem: "SKU-PAN", Descripcion: "Pan artesanal", UnidadMedida: "EA", Cantidad: 1, PrecioUnitario: 100, ImpuestoCodigo: "01", ImpuestoPorcentaje: 19, BaseGravable: 100, ValorImpuesto: 19, SubtotalLinea: 100, TotalLinea: 119},
		{ID: 9, EmpresaID: 12, CarritoID: 77, TipoItem: "producto", ReferenciaID: 101, CodigoItem: "SKU-CAFE", Descripcion: "Cafe colombiano", UnidadMedida: "EA", Cantidad: 1, PrecioUnitario: 100, ImpuestoCodigo: "01", ImpuestoPorcentaje: 19, BaseGravable: 100, ValorImpuesto: 19, SubtotalLinea: 100, TotalLinea: 119},
	}
	cfg := &dbpkg.EmpresaConfiguracionAvanzada{
		EmpresaID: 12, TipoDocumentoEmisor: "NIT", NIT: "900123456", DigitoVerificacion: "8", TipoPersonaFiscal: "juridica",
		RazonSocial: "Empresa Real SAS", NombreComercial: "Empresa Real", RegimenFiscal: "responsable_iva", ResponsabilidadTributaria: "O-13",
		EmailFacturacion: "facturacion@empresa-real.test", TelefonoFacturacion: "6010000000", DireccionFiscal: "Calle 10 20-30",
		PaisCodigo: "CO", Departamento: "Cundinamarca", DepartamentoCodigoDANE: "25", Municipio: "Chia", MunicipioCodigoDANE: "25175", CodigoPostal: "250001",
	}
	cliente := &dbpkg.Cliente{
		ID: 22, EmpresaID: 12, TipoDocumento: "CC", NumeroDocumento: "84456779", TipoPersona: "natural",
		NombreRazonSocial: "Cliente Real", ResponsabilidadTributaria: "R-99-PN", Email: "cliente@correo.test", Telefono: "3000000000", Direccion: "Carrera 5 6-7",
		Pais: "CO", Departamento: "Cundinamarca", DepartamentoCodigoDANE: "25", Municipio: "Chia", MunicipioCodigoDANE: "25175", CodigoPostal: "250001",
	}
	doc := dbpkg.EmpresaDocumentoFacturacion{
		EmpresaID: 12, TipoDocumento: "comprobante_pago", DocumentoCodigo: "CP-VENTA-CRT-77-PG-20260823123000",
		MontoTotal: 238, Moneda: "COP", FechaDocumento: "2026-08-23 12:30:00", EntidadRelacionadaID: 22,
	}
	return carrito, items, cfg, cliente, doc
}

func facturacionFuenteFiscalContains(values []string, want string) bool {
	for _, value := range values {
		if strings.TrimSpace(value) == want {
			return true
		}
	}
	return false
}
