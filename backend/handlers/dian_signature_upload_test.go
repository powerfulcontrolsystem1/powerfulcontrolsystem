package handlers

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestDecodeDIANSignatureUploadPEMWithPrivateKeyAndCertificate(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	template := &x509.Certificate{
		SerialNumber: big.NewInt(12345),
		Subject:      pkix.Name{CommonName: "Empresa DIAN Test"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}
	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("create cert: %v", err)
	}

	var pemBody strings.Builder
	_ = pem.Encode(&pemBody, &pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	_ = pem.Encode(&pemBody, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})

	material, err := decodeDIANSignatureUpload([]byte(pemBody.String()), "firma.pem", "")
	if err != nil {
		t.Fatalf("decode upload: %v", err)
	}
	if !strings.Contains(material.PrivateKeyPEM, "RSA PRIVATE KEY") {
		t.Fatalf("expected private key PEM, got %q", material.PrivateKeyPEM)
	}
	if !strings.Contains(material.CertificatePEM, "CERTIFICATE") {
		t.Fatalf("expected certificate PEM, got %q", material.CertificatePEM)
	}
	if material.Subject == "" || material.Serial == "" {
		t.Fatalf("expected certificate metadata, got subject=%q serial=%q", material.Subject, material.Serial)
	}
	if material.NotAfter.IsZero() {
		t.Fatalf("expected certificate expiration metadata")
	}
}

func TestDIANCertificateExpiryStatusWarnsBeforeExpiration(t *testing.T) {
	now := time.Date(2026, 6, 1, 9, 0, 0, 0, time.UTC)
	notAfter := now.AddDate(0, 0, 10)
	status := dianCertificateExpiryStatus(now, now.AddDate(0, -1, 0), notAfter, 30)
	if !parseTruthy(genericStringValue(status["proximo_a_vencer"])) {
		t.Fatalf("expected certificate to be close to expiration, got %#v", status)
	}
	if parseTruthy(genericStringValue(status["vencido"])) {
		t.Fatalf("certificate must not be expired, got %#v", status)
	}
	if genericStringValue(status["fecha_vencimiento"]) != notAfter.Format("2006-01-02") {
		t.Fatalf("unexpected expiration date: %#v", status)
	}
}

func TestDIANCertificateExpiryStatusDetectsExpired(t *testing.T) {
	now := time.Date(2026, 6, 1, 9, 0, 0, 0, time.UTC)
	status := dianCertificateExpiryStatus(now, now.AddDate(0, -2, 0), now.AddDate(0, 0, -1), 30)
	if !parseTruthy(genericStringValue(status["vencido"])) {
		t.Fatalf("expected expired certificate, got %#v", status)
	}
	if parseTruthy(genericStringValue(status["ok"])) {
		t.Fatalf("expired certificate must not be ok, got %#v", status)
	}
}

func TestDecodeDIANSignatureUploadRejectsCertificateWithoutPrivateKey(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	template := &x509.Certificate{
		SerialNumber: big.NewInt(67890),
		Subject:      pkix.Name{CommonName: "Solo Certificado"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
	}
	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("create cert: %v", err)
	}
	var pemBody strings.Builder
	_ = pem.Encode(&pemBody, &pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	_, err = decodeDIANSignatureUpload([]byte(pemBody.String()), "certificado.crt", "")
	if err == nil || !strings.Contains(err.Error(), "no contiene llave privada RSA") {
		t.Fatalf("expected private-key validation error, got %v", err)
	}
}

func TestEmpresaFacturacionFirmaElectronicaUsesCompanyFolder(t *testing.T) {
	dir, publicPrefix, folder := empresaUploadsSubdir(nil, 15, empresaFacturacionElectronicaDirName, empresaFirmaElectronicaDirName)
	expectedSuffix := filepath.Join("uploads", "empresas", "empresa_15_empresa", "facturacion_electronica", "firma_electronica")
	if !strings.HasSuffix(filepath.Clean(dir), expectedSuffix) {
		t.Fatalf("expected firma dir suffix %q, got %q", expectedSuffix, dir)
	}
	if folder != "empresa_15_empresa" {
		t.Fatalf("expected company folder empresa_15_empresa, got %q", folder)
	}
	if publicPrefix != "/uploads/empresas/empresa_15_empresa/facturacion_electronica/firma_electronica" {
		t.Fatalf("unexpected public prefix: %q", publicPrefix)
	}
}

func TestBuildDIANSOAPEnvelopeSendTestSetAsync(t *testing.T) {
	zipBytes, err := buildDIANZipContent("FV1.xml", "<Invoice><ID>FV1</ID></Invoice>")
	if err != nil {
		t.Fatalf("zip content: %v", err)
	}
	envelope := buildDIANSOAPEnvelope("SendTestSetAsync", "https://vpfe-hab.dian.gov.co/WcfDianCustomerServices.svc", "FV1.zip", zipBytes, "abc-test-set")
	for _, expected := range []string{
		"http://wcf.dian.colombia/IWcfDianCustomerServices/SendTestSetAsync",
		"<wcf:fileName>FV1.zip</wcf:fileName>",
		"<wcf:contentFile>",
		"<wcf:testSetId>abc-test-set</wcf:testSetId>",
	} {
		if !strings.Contains(envelope, expected) {
			t.Fatalf("expected SOAP envelope to contain %q, got %s", expected, envelope)
		}
	}
}

func TestExtractDIANSOAPResponseMapTrackID(t *testing.T) {
	raw := `<s:Envelope><s:Body><SendTestSetAsyncResponse><SendTestSetAsyncResult><ZipKey>TRACK-123</ZipKey><StatusCode>00</StatusCode><StatusMessage>Procesado</StatusMessage></SendTestSetAsyncResult></SendTestSetAsyncResponse></s:Body></s:Envelope>`
	out := extractDIANSOAPResponseMap(raw)
	if out["track_id"] != "TRACK-123" {
		t.Fatalf("expected track id, got %#v", out)
	}
	if out["status_code"] != "00" {
		t.Fatalf("expected status code, got %#v", out)
	}
}

func TestBuildDIANGetStatusZipEnvelope(t *testing.T) {
	envelope := buildDIANGetStatusZipEnvelope("https://vpfe.dian.gov.co/WcfDianCustomerServices.svc", "TRACK-123")
	for _, expected := range []string{
		"http://wcf.dian.colombia/IWcfDianCustomerServices/GetStatusZip",
		"<wcf:GetStatusZip>",
		"<wcf:trackId>TRACK-123</wcf:trackId>",
	} {
		if !strings.Contains(envelope, expected) {
			t.Fatalf("expected SOAP envelope to contain %q, got %s", expected, envelope)
		}
	}
}

func TestDIANDefaultSetRequirementUsesSoftwarePropioProveedorTarget(t *testing.T) {
	got := dianDefaultSetRequirement()
	if got["facturas_electronicas"] != 60 || got["notas_debito"] != 20 || got["notas_credito"] != 20 || got["total_documentos"] != 100 {
		t.Fatalf("unexpected default DIAN set requirement: %#v", got)
	}
	if !strings.Contains(genericStringValue(got["nota"]), "software propio") {
		t.Fatalf("expected default note to explain software mode, got %#v", got)
	}
}

func TestCalculateColombianNITDV(t *testing.T) {
	dv, ok := calculateColombianNITDV("900373913")
	if !ok {
		t.Fatalf("expected NIT DV calculation to be valid")
	}
	if dv != 4 {
		t.Fatalf("expected DV 4, got %d", dv)
	}
}

func TestValidateDIANDocumentPreflightAcceptsBasicUBL(t *testing.T) {
	cfg := map[string]interface{}{
		"nit":                    "900373913",
		"digito_verificacion":    "4",
		"razon_social":           "Empresa Test SAS",
		"tipo_ambiente":          "habilitacion",
		"prefijo":                "SETP",
		"resolucion_numero":      "18760000000001",
		"resolucion_fecha_desde": time.Now().AddDate(0, 0, -1).Format("2006-01-02"),
		"resolucion_fecha_hasta": time.Now().AddDate(0, 1, 0).Format("2006-01-02"),
		"rango_desde":            1,
		"rango_hasta":            999999,
		"consecutivo_actual":     1,
		"url_dian":               "https://vpfe-hab.dian.gov.co/WcfDianCustomerServices.svc?wsdl",
		"token_emisor_ref":       "env:DIAN_TOKEN_TEST",
		"certificado_clave_ref":  "file:/tmp/key.pem",
		"certificado_url":        "file:/tmp/cert.pem",
		"test_set_id":            "test-set",
		"software_id":            "software-id",
		"software_pin":           "software-pin",
	}
	xmlPayload := `<?xml version="1.0" encoding="UTF-8"?><Invoice><UBLVersionID>2.1</UBLVersionID><ProfileExecutionID>2</ProfileExecutionID><ID>SETP1</ID><UUID>ABC123</UUID><TaxAmount>0.00</TaxAmount><PayableAmount>1000.00</PayableAmount></Invoice>`
	payload := map[string]interface{}{
		"empresa_id":       1,
		"documento_codigo": "SETP1",
		"fecha_emision":    time.Now().Format(time.RFC3339),
		"cliente_nombre":   "Cliente Test",
		"cliente_nit":      "222222222222",
		"total":            "1000.00",
		"impuesto_total":   "0.00",
		"moneda":           "COP",
		"xml":              xmlPayload,
	}
	result := validateDIANDocumentPreflight(cfg, 1, payload, xmlPayload, "validacion_manual")
	if blocked, _ := result["bloqueado"].(bool); blocked {
		t.Fatalf("expected preflight to pass, got %#v", result)
	}
}

func TestValidateDIANDocumentPreflightRejectsBadDV(t *testing.T) {
	cfg := map[string]interface{}{
		"nit":                    "900373913",
		"digito_verificacion":    "9",
		"razon_social":           "Empresa Test SAS",
		"tipo_ambiente":          "produccion",
		"prefijo":                "FV",
		"resolucion_numero":      "18760000000001",
		"resolucion_fecha_desde": time.Now().AddDate(0, 0, -1).Format("2006-01-02"),
		"resolucion_fecha_hasta": time.Now().AddDate(0, 1, 0).Format("2006-01-02"),
		"rango_desde":            1,
		"rango_hasta":            999999,
		"consecutivo_actual":     1,
		"url_dian":               "https://vpfe.dian.gov.co/WcfDianCustomerServices.svc?wsdl",
		"token_emisor_ref":       "env:DIAN_TOKEN_TEST",
		"certificado_clave_ref":  "file:/tmp/key.pem",
		"certificado_url":        "file:/tmp/cert.pem",
		"software_id":            "software-id",
		"software_pin":           "software-pin",
	}
	result := validateDIANDocumentPreflight(cfg, 1, map[string]interface{}{"empresa_id": 1, "documento_codigo": "FV1", "total": "1000.00", "cliente_nombre": "Cliente", "cliente_nit": "222222222222"}, "", "validacion_manual")
	if blocked, _ := result["bloqueado"].(bool); !blocked {
		t.Fatalf("expected preflight to reject invalid DV, got %#v", result)
	}
}

func TestGenerateDIANUBLBaseDoesNotEmitDemoOrPendingMarkers(t *testing.T) {
	cfg := map[string]interface{}{
		"nit":                 "900373913",
		"digito_verificacion": "4",
		"razon_social":        "Empresa Test SAS",
		"tipo_ambiente":       "habilitacion",
		"prefijo":             "SETP",
		"resolucion_numero":   "18760000000001",
	}
	payload := map[string]interface{}{
		"documento_codigo": "SETP1",
		"cliente_nombre":   "Cliente Test",
		"cliente_nit":      "222222222222",
		"total":            "1000.00",
		"impuesto_total":   "0.00",
		"moneda":           "COP",
	}
	result, status, err := generateDIANUBLBase(cfg, 1, payload)
	if err != nil || status != 200 {
		t.Fatalf("generateDIANUBLBase returned status=%d err=%v result=%#v", status, err, result)
	}
	xmlPayload := genericStringValue(result["xml_ubl_base"])
	upper := strings.ToUpper(xmlPayload)
	if strings.Contains(upper, "DEMO") || strings.Contains(upper, "PENDIENTE") {
		t.Fatalf("xml must not contain demo/pending markers: %s", xmlPayload)
	}
	if !strings.Contains(xmlPayload, `schemeName="CUFE-SHA256"`) {
		t.Fatalf("expected CUFE scheme marker, got %s", xmlPayload)
	}
}

func TestValidateDIANDocumentPreflightBlocksDemoMarkersForRealSend(t *testing.T) {
	cfg := map[string]interface{}{
		"nit":                    "900373913",
		"digito_verificacion":    "4",
		"razon_social":           "Empresa Test SAS",
		"tipo_ambiente":          "habilitacion",
		"prefijo":                "SETP",
		"resolucion_numero":      "18760000000001",
		"resolucion_fecha_desde": time.Now().AddDate(0, 0, -1).Format("2006-01-02"),
		"resolucion_fecha_hasta": time.Now().AddDate(0, 1, 0).Format("2006-01-02"),
		"rango_desde":            1,
		"rango_hasta":            999999,
		"consecutivo_actual":     1,
		"url_dian":               "https://vpfe-hab.dian.gov.co/WcfDianCustomerServices.svc?wsdl",
		"token_emisor_ref":       "inline-token",
		"certificado_clave_ref":  "inline-key",
		"certificado_url":        "inline-cert",
		"test_set_id":            "test-set",
		"software_id":            "software-id",
		"software_pin":           "software-pin",
	}
	xmlPayload := `<?xml version="1.0" encoding="UTF-8"?><Invoice><UBLVersionID>2.1</UBLVersionID><ProfileExecutionID>2</ProfileExecutionID><ID>SETP1</ID><UUID schemeName="CUFE-SHA384-PENDIENTE">ABC123</UUID><TaxAmount>0.00</TaxAmount><PayableAmount>1000.00</PayableAmount><ds:Signature><ds:X509Certificate>ABC</ds:X509Certificate></ds:Signature></Invoice>`
	payload := map[string]interface{}{
		"empresa_id":       1,
		"documento_codigo": "SETP1",
		"fecha_emision":    time.Now().Format(time.RFC3339),
		"cliente_nombre":   "Cliente Test",
		"cliente_nit":      "222222222222",
		"total":            "1000.00",
		"impuesto_total":   "0.00",
		"moneda":           "COP",
		"simular":          true,
	}
	result := validateDIANDocumentPreflight(cfg, 1, payload, xmlPayload, "envio_real")
	if blocked, _ := result["bloqueado"].(bool); !blocked {
		t.Fatalf("expected real preflight to block demo/pending markers, got %#v", result)
	}
	if !validationResultHasCode(result, "DIAN-UBL-DEMO") {
		t.Fatalf("expected DIAN-UBL-DEMO issue, got %#v", result)
	}
}

func validationResultHasCode(result map[string]interface{}, code string) bool {
	for _, key := range []string{"issues", "warnings"} {
		items, _ := result[key].([]map[string]interface{})
		for _, item := range items {
			if genericStringValue(item["code"]) == code {
				return true
			}
		}
	}
	return false
}
