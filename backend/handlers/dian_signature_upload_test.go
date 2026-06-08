package handlers

import (
	"archive/zip"
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func testDIANKeyAndCertPEM(t *testing.T) (string, string) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	template := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject:      pkix.Name{CommonName: "Empresa DIAN Test"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().AddDate(1, 0, 0),
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}
	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("create cert: %v", err)
	}
	var keyPEM strings.Builder
	_ = pem.Encode(&keyPEM, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	var certPEM strings.Builder
	_ = pem.Encode(&certPEM, &pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	return keyPEM.String(), certPEM.String()
}

func testDIANValidConfig(t *testing.T, endpoint string) map[string]interface{} {
	t.Helper()
	keyPEM, certPEM := testDIANKeyAndCertPEM(t)
	return map[string]interface{}{
		"nit":                       "900373913",
		"digito_verificacion":       "4",
		"razon_social":              "Empresa Test SAS",
		"tipo_ambiente":             "habilitacion",
		"prefijo":                   "SETP",
		"resolucion_numero":         "18760000000001",
		"resolucion_fecha_desde":    time.Now().AddDate(0, 0, -1).Format("2006-01-02"),
		"resolucion_fecha_hasta":    time.Now().AddDate(0, 1, 0).Format("2006-01-02"),
		"rango_desde":               1,
		"rango_hasta":               999999,
		"consecutivo_actual":        1,
		"llave_tecnica":             "llave-tecnica-test",
		"url_dian":                  endpoint,
		"certificado_clave_ref":     keyPEM,
		"certificado_url":           certPEM,
		"test_set_id":               "test-set",
		"software_id":               "software-id",
		"software_pin":              "software-pin",
		"token_emisor_ref":          "inline-token",
		"usar_software_compartido":  0,
		"set_documentos_requeridos": 1,
	}
}

type testDIANSOAPServerStats struct {
	sendTestSet int32
	statusZip   int32
}

func newTestDIANSOAPServer(t *testing.T) (*httptest.Server, *testDIANSOAPServerStats) {
	t.Helper()
	stats := &testDIANSOAPServerStats{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, _ := io.ReadAll(r.Body)
		body := string(bodyBytes)
		w.Header().Set("Content-Type", "application/soap+xml; charset=utf-8")
		switch {
		case strings.Contains(body, "SendTestSetAsync"):
			count := atomic.AddInt32(&stats.sendTestSet, 1)
			if !strings.Contains(body, "<wcf:testSetId>test-set</wcf:testSetId>") {
				t.Errorf("SendTestSetAsync without expected testSetId: %s", body)
			}
			content := extractDIANSOAPTag(body, "contentFile")
			if content == "" {
				t.Errorf("SendTestSetAsync without contentFile")
			} else {
				zipBytes, err := base64.StdEncoding.DecodeString(content)
				if err != nil {
					t.Errorf("invalid contentFile base64: %v", err)
				} else {
					zr, err := zip.NewReader(bytes.NewReader(zipBytes), int64(len(zipBytes)))
					if err != nil {
						t.Errorf("invalid DIAN zip: %v", err)
					} else if len(zr.File) != 1 {
						t.Errorf("expected one XML in DIAN zip, got %d", len(zr.File))
					} else {
						rc, err := zr.File[0].Open()
						if err != nil {
							t.Errorf("open XML in DIAN zip: %v", err)
						} else {
							xmlBytes, _ := io.ReadAll(rc)
							_ = rc.Close()
							xmlPayload := string(xmlBytes)
							if !strings.Contains(xmlPayload, "<ds:Signature") {
								t.Errorf("XML sent to DIAN must be signed, got %s", xmlPayload)
							}
							if strings.Contains(strings.ToLower(xmlPayload), "simulado") {
								t.Errorf("XML sent to DIAN must not be simulated, got %s", xmlPayload)
							}
						}
					}
				}
			}
			_, _ = w.Write([]byte(`<s:Envelope><s:Body><SendTestSetAsyncResponse><SendTestSetAsyncResult><ZipKey>TRACK-` + genericStringValue(count) + `</ZipKey><StatusCode>00</StatusCode><StatusMessage>Recibido</StatusMessage></SendTestSetAsyncResult></SendTestSetAsyncResponse></s:Body></s:Envelope>`))
		case strings.Contains(body, "GetStatusZip"):
			atomic.AddInt32(&stats.statusZip, 1)
			trackID := extractDIANSOAPTag(body, "trackId")
			if trackID == "" {
				t.Errorf("GetStatusZip without trackId: %s", body)
			}
			_, _ = w.Write([]byte(`<s:Envelope><s:Body><GetStatusZipResponse><GetStatusZipResult><IsValid>true</IsValid><StatusCode>00</StatusCode><StatusMessage>Documento aceptado</StatusMessage><XmlDocumentKey>` + trackID + `</XmlDocumentKey></GetStatusZipResult></GetStatusZipResponse></s:Body></s:Envelope>`))
		default:
			http.Error(w, "unexpected SOAP operation", http.StatusBadRequest)
		}
	}))
	return server, stats
}

func dianTestContainsString(items []string, expected string) bool {
	for _, item := range items {
		if item == expected {
			return true
		}
	}
	return false
}

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

func TestBuildDIANSOAPEnvelopeWithWSSecuritySendTestSetAsync(t *testing.T) {
	keyPEM, certPEM := testDIANKeyAndCertPEM(t)
	privateKey, err := parseDIANRSAPrivateKey(keyPEM)
	if err != nil {
		t.Fatalf("parse key: %v", err)
	}
	certificate, err := parseDIANCertificate(certPEM)
	if err != nil {
		t.Fatalf("parse cert: %v", err)
	}
	zipBytes, err := buildDIANZipContent("FV1.xml", "<Invoice><ID>FV1</ID></Invoice>")
	if err != nil {
		t.Fatalf("zip content: %v", err)
	}
	envelope, meta, err := buildDIANSOAPEnvelopeWithWSSecurity("SendTestSetAsync", "https://vpfe-hab.dian.gov.co/WcfDianCustomerServices.svc", "FV1.zip", zipBytes, "abc-test-set", privateKey, certificate, time.Date(2026, 6, 5, 10, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("secure envelope: %v", err)
	}
	for _, expected := range []string{
		`<wsse:Security`,
		`<wsu:Timestamp`,
		`<wsse:BinarySecurityToken`,
		`<ds:Signature`,
		`<ds:SignatureValue>`,
		`<ec:InclusiveNamespaces xmlns:ec="http://www.w3.org/2001/10/xml-exc-c14n#" PrefixList="wsa soap wcf"></ec:InclusiveNamespaces>`,
		`<ec:InclusiveNamespaces xmlns:ec="http://www.w3.org/2001/10/xml-exc-c14n#" PrefixList="soap wcf"></ec:InclusiveNamespaces>`,
		`<wsse:Reference URI="#X509-`,
		`<ds:Reference URI="#ID-`,
		`<wsa:Action xmlns:wsa="http://www.w3.org/2005/08/addressing">`,
		`<wsa:MessageID`,
		`<wsa:ReplyTo`,
		`<wcf:testSetId>abc-test-set</wcf:testSetId>`,
		`http://wcf.dian.colombia/IWcfDianCustomerServices/SendTestSetAsync`,
	} {
		if !strings.Contains(envelope, expected) {
			t.Fatalf("expected WS-Security envelope to contain %q, got %s", expected, envelope)
		}
	}
	if !parseTruthy(genericStringValue(meta["ws_security"])) {
		t.Fatalf("expected ws_security meta true, got %#v", meta)
	}
	if genericStringValue(meta["key_reference"]) != "BinarySecurityTokenReference" {
		t.Fatalf("expected BinarySecurityTokenReference key reference, got %#v", meta)
	}
	if genericStringValue(meta["security_layout"]) != "strict_timestamp_token_signature" {
		t.Fatalf("expected strict layout metadata, got %#v", meta)
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

func TestValidateDIANCredentialRefsDoesNotRequireTokenForOfficialSOAP(t *testing.T) {
	cfg := testDIANValidConfig(t, "https://vpfe-hab.dian.gov.co/WcfDianCustomerServices.svc?wsdl")
	response, status, err := validateDIANCredentialRefs(cfg, 12, map[string]interface{}{})
	if err != nil || status != 200 {
		t.Fatalf("validate credentials returned status=%d err=%v response=%#v", status, err, response)
	}
	if !parseTruthy(genericStringValue(response["ok"])) {
		t.Fatalf("expected official SOAP credentials to pass without token, got %#v", response)
	}
	checks, _ := response["checks"].(map[string]interface{})
	tokenCheck, _ := checks["token_emisor"].(map[string]interface{})
	if !parseTruthy(genericStringValue(tokenCheck["ok"])) {
		t.Fatalf("expected token check ok for official SOAP endpoint, got %#v", tokenCheck)
	}
	if parseTruthy(genericStringValue(tokenCheck["required"])) {
		t.Fatalf("token must not be required for official SOAP endpoint, got %#v", tokenCheck)
	}
	if dianTestContainsString(missingDIANFields(cfg), "token_emisor_ref") {
		t.Fatalf("token_emisor_ref must not be listed as missing for official SOAP endpoint")
	}
}

func TestValidateDIANCredentialRefsRequiresTokenForProviderAPI(t *testing.T) {
	cfg := testDIANValidConfig(t, "https://proveedor.example.com/api/dian")
	delete(cfg, "token_emisor_ref")
	response, status, err := validateDIANCredentialRefs(cfg, 12, map[string]interface{}{})
	if err != nil || status != 200 {
		t.Fatalf("validate credentials returned status=%d err=%v response=%#v", status, err, response)
	}
	if parseTruthy(genericStringValue(response["ok"])) {
		t.Fatalf("expected provider API credentials to fail without token, got %#v", response)
	}
	checks, _ := response["checks"].(map[string]interface{})
	tokenCheck, _ := checks["token_emisor"].(map[string]interface{})
	if !parseTruthy(genericStringValue(tokenCheck["required"])) {
		t.Fatalf("token must be required for provider API endpoint, got %#v", tokenCheck)
	}
	if !dianTestContainsString(missingDIANFields(cfg), "token_emisor_ref") {
		t.Fatalf("token_emisor_ref must be listed as missing for provider API endpoint")
	}
}

func TestValidateDIANCredentialRefsRequiresTestSetForRealHabilitacion(t *testing.T) {
	cfg := testDIANValidConfig(t, "https://vpfe-hab.dian.gov.co/WcfDianCustomerServices.svc?wsdl")
	delete(cfg, "test_set_id")
	response, status, err := validateDIANCredentialRefs(cfg, 12, map[string]interface{}{})
	if err != nil || status != 200 {
		t.Fatalf("validate credentials returned status=%d err=%v response=%#v", status, err, response)
	}
	if parseTruthy(genericStringValue(response["ok"])) {
		t.Fatalf("expected real habilitacion validation to fail without test_set_id, got %#v", response)
	}
	checks, _ := response["checks"].(map[string]interface{})
	testSetCheck, _ := checks["test_set_id"].(map[string]interface{})
	if !parseTruthy(genericStringValue(testSetCheck["required"])) || parseTruthy(genericStringValue(testSetCheck["ok"])) {
		t.Fatalf("expected required missing test_set_id check, got %#v", testSetCheck)
	}
	if !dianTestContainsString(missingDIANFields(cfg), "test_set_id") {
		t.Fatalf("test_set_id must be listed as missing for real habilitacion")
	}
}

func TestValidateDIANCredentialRefsAllowsMissingTestSetInSimulation(t *testing.T) {
	cfg := testDIANValidConfig(t, "https://vpfe-hab.dian.gov.co/WcfDianCustomerServices.svc?wsdl")
	delete(cfg, "test_set_id")
	response, status, err := validateDIANCredentialRefs(cfg, 12, map[string]interface{}{"simular": true})
	if err != nil || status != 200 {
		t.Fatalf("validate credentials returned status=%d err=%v response=%#v", status, err, response)
	}
	if !parseTruthy(genericStringValue(response["ok"])) {
		t.Fatalf("expected simulation validation to pass without test_set_id, got %#v", response)
	}
	checks, _ := response["checks"].(map[string]interface{})
	testSetCheck, _ := checks["test_set_id"].(map[string]interface{})
	if !parseTruthy(genericStringValue(testSetCheck["required"])) || parseTruthy(genericStringValue(testSetCheck["ok"])) {
		t.Fatalf("expected required but non-blocking missing test_set_id check in simulation, got %#v", testSetCheck)
	}
}

func TestRunDIANPruebasHabilitacionReportsMissingTestSetForRealRun(t *testing.T) {
	cfg := testDIANValidConfig(t, "https://vpfe-hab.dian.gov.co/WcfDianCustomerServices.svc?wsdl")
	delete(cfg, "test_set_id")
	result, status, err := runDIANPruebasHabilitacion(nil, cfg, 12, map[string]interface{}{
		"simular": false,
	})
	if err != nil {
		t.Fatalf("run pruebas returned err=%v", err)
	}
	if status != http.StatusConflict {
		t.Fatalf("expected conflict for missing test_set_id, got status=%d result=%#v", status, result)
	}
	if parseTruthy(genericStringValue(result["ok"])) {
		t.Fatalf("expected blocked response, got %#v", result)
	}
	if !strings.Contains(genericStringValue(result["motivo"]), "test_set_id") {
		t.Fatalf("expected motivo to mention test_set_id, got %#v", result)
	}
	faltantes, _ := result["faltantes"].([]string)
	if !dianTestContainsString(faltantes, "test_set_id") {
		t.Fatalf("expected top-level faltantes to include test_set_id, got %#v", result["faltantes"])
	}
}

func TestRunDIANPruebasHabilitacionRejectsSimulation(t *testing.T) {
	cfg := testDIANValidConfig(t, "https://vpfe-hab.dian.gov.co/WcfDianCustomerServices.svc?wsdl")
	result, status, err := runDIANPruebasHabilitacion(nil, cfg, 12, map[string]interface{}{
		"simular": true,
	})
	if err != nil {
		t.Fatalf("run pruebas returned err=%v", err)
	}
	if status != http.StatusBadRequest {
		t.Fatalf("expected bad request for simulated DIAN set, got status=%d result=%#v", status, result)
	}
	if !parseTruthy(genericStringValue(result["bloqueado"])) {
		t.Fatalf("expected simulated DIAN set to be blocked, got %#v", result)
	}
}

func TestRunDIANPruebasHabilitacionTwoEachRealSOAPWithStatusZip(t *testing.T) {
	server, stats := newTestDIANSOAPServer(t)
	defer server.Close()

	cfg := testDIANValidConfig(t, server.URL)
	result, status, err := runDIANPruebasHabilitacion(nil, cfg, 12, map[string]interface{}{
		"simular":                false,
		"facturas_electronicas":  2,
		"notas_debito":           2,
		"notas_credito":          2,
		"total_documentos":       6,
		"max_envios":             6,
		"detener_en_error":       true,
		"total_por_documento":    "1000.00",
		"set_habilitacion":       true,
		"cliente_nombre":         "Cliente habilitacion DIAN",
		"cliente_nit":            "222222222222",
		"set_documentos_minimos": 1,
		"acuse_intentos":         1,
		"acuse_espera_ms":        0,
		"usar_soap_dian":         true,
	})
	if err != nil || status != 200 {
		t.Fatalf("run pruebas returned status=%d err=%v result=%#v", status, err, result)
	}
	if !parseTruthy(genericStringValue(result["ok"])) {
		t.Fatalf("expected real 2+2+2 set to be ok, got %#v", result)
	}
	if anyToInt64(result["procesados"]) != 6 {
		t.Fatalf("expected 6 processed docs, got %#v", result)
	}
	resumen, _ := result["resumen"].(map[string]int)
	if resumen["aceptado"] != 6 || resumen["simulado"] != 0 {
		t.Fatalf("expected 6 accepted real docs and no simulation, got %#v", resumen)
	}
	if atomic.LoadInt32(&stats.sendTestSet) != 6 || atomic.LoadInt32(&stats.statusZip) != 6 {
		t.Fatalf("expected 6 SendTestSetAsync and 6 GetStatusZip calls, got send=%d status=%d", stats.sendTestSet, stats.statusZip)
	}
	objetivo, _ := result["objetivo"].(map[string]interface{})
	if anyToInt64(objetivo["facturas_electronicas"]) != 2 || anyToInt64(objetivo["notas_debito"]) != 2 || anyToInt64(objetivo["notas_credito"]) != 2 {
		t.Fatalf("unexpected 2+2+2 objective: %#v", objetivo)
	}
	if !parseTruthy(genericStringValue(result["habilitacion_aprobada"])) {
		t.Fatalf("expected accepted real set to approve habilitation, got %#v", result)
	}
}

func TestDIANDefaultSetRequirementUsesSoftwarePropioProveedorTarget(t *testing.T) {
	got := dianDefaultSetRequirement()
	if got["facturas_electronicas"] != 30 || got["notas_debito"] != 10 || got["notas_credito"] != 10 || got["total_documentos"] != 50 {
		t.Fatalf("unexpected default DIAN set requirement: %#v", got)
	}
	if !strings.Contains(genericStringValue(got["nota"]), "software propio") {
		t.Fatalf("expected default note to explain software mode, got %#v", got)
	}
}

func TestRunDIANPruebasHabilitacionUsesConfiguredPortalSet(t *testing.T) {
	server, stats := newTestDIANSOAPServer(t)
	defer server.Close()

	cfg := testDIANValidConfig(t, server.URL)
	cfg["set_documentos_requeridos"] = 4
	cfg["set_facturas_requeridas"] = 2
	cfg["set_notas_debito_requeridas"] = 1
	cfg["set_notas_credito_requeridas"] = 1
	cfg["set_documentos_aceptados_requeridos"] = 1
	cfg["set_facturas_aceptadas_requeridas"] = 1
	cfg["set_notas_debito_aceptadas_requeridas"] = 0
	cfg["set_notas_credito_aceptadas_requeridas"] = 0

	result, status, err := runDIANPruebasHabilitacion(nil, cfg, 12, map[string]interface{}{
		"simular":         false,
		"acuse_intentos":  1,
		"acuse_espera_ms": 0,
		"usar_soap_dian":  true,
	})
	if err != nil || status != 200 {
		t.Fatalf("run pruebas returned status=%d err=%v result=%#v", status, err, result)
	}
	if anyToInt64(result["procesados"]) != 4 {
		t.Fatalf("expected configured portal set to process 4 docs, got %#v", result["procesados"])
	}
	objetivo, _ := result["objetivo"].(map[string]interface{})
	if anyToInt64(objetivo["facturas_electronicas"]) != 2 || anyToInt64(objetivo["notas_debito"]) != 1 || anyToInt64(objetivo["notas_credito"]) != 1 {
		t.Fatalf("unexpected configured portal objective: %#v", objetivo)
	}
	req, _ := result["requisito_set_dian"].(map[string]interface{})
	if anyToInt64(req["facturas_electronicas_aceptadas_minimo"]) != 1 || anyToInt64(req["notas_debito_aceptadas_minimo"]) != 0 || anyToInt64(req["notas_credito_aceptadas_minimo"]) != 0 {
		t.Fatalf("unexpected accepted minimums: %#v", req)
	}
	if !parseTruthy(genericStringValue(result["habilitacion_aprobada"])) {
		t.Fatalf("accepted real set must mark habilitation approved: %#v", result)
	}
	if atomic.LoadInt32(&stats.sendTestSet) != 4 || atomic.LoadInt32(&stats.statusZip) != 4 {
		t.Fatalf("expected configured set to call DIAN 4+4 times, got send=%d status=%d", stats.sendTestSet, stats.statusZip)
	}
}

func TestDIANEffectiveSetRequirementAllowsManualSingleDocument(t *testing.T) {
	got := dianEffectiveSetRequirement(map[string]interface{}{
		"facturas_electronicas": 1,
		"notas_debito":          0,
		"notas_credito":         0,
	})
	if got["facturas_electronicas"] != 1 || got["notas_debito"] != 0 || got["notas_credito"] != 0 || got["total_documentos"] != 1 {
		t.Fatalf("unexpected manual DIAN set requirement: %#v", got)
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

func TestValidateDIANDocumentPreflightAcceptsGeneratedDIANUBL(t *testing.T) {
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
		"llave_tecnica":          "llave-tecnica-test",
		"url_dian":               "https://vpfe-hab.dian.gov.co/WcfDianCustomerServices.svc?wsdl",
		"token_emisor_ref":       "env:DIAN_TOKEN_TEST",
		"certificado_clave_ref":  "file:/tmp/key.pem",
		"certificado_url":        "file:/tmp/cert.pem",
		"test_set_id":            "test-set",
		"software_id":            "software-id",
		"software_pin":           "software-pin",
	}
	payload := map[string]interface{}{
		"empresa_id":       1,
		"documento_codigo": "SETP1",
		"fecha_emision":    time.Now().Format(time.RFC3339),
		"cliente_nombre":   "Cliente Test",
		"cliente_nit":      "2222222222",
		"total":            "1190.00",
		"impuesto_total":   "190.00",
		"moneda":           "COP",
	}
	generated, status, err := generateDIANUBLBase(cfg, 1, payload)
	if err != nil || status != http.StatusOK {
		t.Fatalf("generateDIANUBLBase returned status=%d err=%v result=%#v", status, err, generated)
	}
	xmlPayload := genericStringValue(generated["xml_ubl_base"])
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
	if !strings.Contains(xmlPayload, `schemeName="CUFE-SHA384"`) {
		t.Fatalf("expected CUFE scheme marker, got %s", xmlPayload)
	}
	if !strings.Contains(xmlPayload, "<sts:DianExtensions>") || !strings.Contains(xmlPayload, "<cac:InvoiceLine>") {
		t.Fatalf("expected DIAN extensions and invoice line, got %s", xmlPayload)
	}
	for _, expected := range []string{
		"<cbc:ProfileID>DIAN 2.1: Factura Electrónica de Venta</cbc:ProfileID>",
		"<cbc:PrepaidAmount",
		"<cac:PaymentMeans>",
		"<cbc:AdditionalAccountID>2</cbc:AdditionalAccountID>",
		"<cbc:Name>consumidor o usuario final</cbc:Name>",
		"<cac:TaxScheme><cbc:ID>ZZ</cbc:ID><cbc:Name>No aplica</cbc:Name></cac:TaxScheme>",
		"CO, DIAN (Dirección de Impuestos y Aduanas Nacionales)",
	} {
		if !strings.Contains(xmlPayload, expected) {
			t.Fatalf("expected %q in invoice XML: %s", expected, xmlPayload)
		}
	}
	if strings.Contains(xmlPayload, "PrePaidAmount") || strings.Contains(xmlPayload, "Direccion de Impuestos") {
		t.Fatalf("invoice XML contains DIAN-rejected literal/casing: %s", xmlPayload)
	}
}

func TestGenerateDIANUBLBaseUsesCorrectNoteLines(t *testing.T) {
	cfg := map[string]interface{}{
		"nit":                    "900373913",
		"digito_verificacion":    "4",
		"razon_social":           "Empresa Test SAS",
		"tipo_ambiente":          "habilitacion",
		"prefijo":                "SETP",
		"resolucion_numero":      "18760000000001",
		"resolucion_fecha_desde": time.Now().AddDate(0, 0, -1).Format("2006-01-02"),
		"resolucion_fecha_hasta": time.Now().AddDate(0, 1, 0).Format("2006-01-02"),
		"llave_tecnica":          "llave-tecnica-test",
		"software_id":            "software-id",
		"software_pin":           "software-pin",
	}
	for _, tc := range []struct {
		docType      string
		expectedRoot string
		expectedLine string
		expectedCUDE string
		expectedID   string
	}{
		{docType: "nota_credito", expectedRoot: "<CreditNote ", expectedLine: "<cac:CreditNoteLine>", expectedCUDE: `schemeName="CUDE-SHA384"`, expectedID: "<cbc:ProfileID>DIAN 2.1: Nota Crédito de Factura Electrónica de Venta</cbc:ProfileID>"},
		{docType: "nota_debito", expectedRoot: "<DebitNote ", expectedLine: "<cac:DebitNoteLine>", expectedCUDE: `schemeName="CUDE-SHA384"`, expectedID: "<cbc:ProfileID>DIAN 2.1: Nota Débito de Factura Electrónica de Venta</cbc:ProfileID>"},
	} {
		result, status, err := generateDIANUBLBase(cfg, 1, map[string]interface{}{
			"documento_codigo": "SETP99",
			"documento_tipo":   tc.docType,
			"cliente_nombre":   "Cliente Test",
			"cliente_nit":      "2222222222",
			"total":            "1190.00",
			"impuesto_total":   "190.00",
			"moneda":           "COP",
		})
		if err != nil || status != http.StatusOK {
			t.Fatalf("%s generate status=%d err=%v result=%#v", tc.docType, status, err, result)
		}
		xmlPayload := genericStringValue(result["xml_ubl_base"])
		for _, expected := range []string{tc.expectedRoot, tc.expectedLine, tc.expectedCUDE, tc.expectedID, "<cac:DiscrepancyResponse>", "<cac:BillingReference>", "<cac:PaymentMeans>", "<cbc:PrepaidAmount"} {
			if !strings.Contains(xmlPayload, expected) {
				t.Fatalf("%s expected %q in XML: %s", tc.docType, expected, xmlPayload)
			}
		}
		if strings.Contains(xmlPayload, "<cac:InvoiceLine>") || strings.Contains(xmlPayload, "PrePaidAmount") {
			t.Fatalf("%s must not use InvoiceLine: %s", tc.docType, xmlPayload)
		}
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

func TestExtractDIANSOAPResponseMapIncludesStatusDescription(t *testing.T) {
	raw := `<s:Envelope><s:Body><GetStatusZipResponse><GetStatusZipResult><b:IsValid>false</b:IsValid><b:StatusDescription>Batch en proceso de validacion.</b:StatusDescription><b:StatusCode>00</b:StatusCode></GetStatusZipResult></GetStatusZipResponse></s:Body></s:Envelope>`
	result := extractDIANSOAPResponseMap(raw)
	if got := genericStringValue(result["status_description"]); got != "Batch en proceso de validacion." {
		t.Fatalf("expected StatusDescription, got %#v", result)
	}
	if got := genericStringValue(result["is_valid"]); got != "false" {
		t.Fatalf("expected IsValid=false, got %#v", result)
	}
}

func TestExtractDIANSOAPResponseMapIncludesErrorMessageList(t *testing.T) {
	raw := `<s:Envelope><s:Body><GetStatusZipResponse><GetStatusZipResult><b:IsValid>false</b:IsValid><b:StatusCode>99</b:StatusCode><b:ErrorMessageList><b:string>Regla DIAN A fallida</b:string><b:string>Regla DIAN B fallida</b:string></b:ErrorMessageList></GetStatusZipResult></GetStatusZipResponse></s:Body></s:Envelope>`
	result := extractDIANSOAPResponseMap(raw)
	messages, _ := result["error_messages"].([]string)
	if len(messages) != 2 {
		t.Fatalf("expected both DIAN error messages, got %#v", result)
	}
	estado, mensaje := resolveDIANAcuseFromResponse(http.StatusOK, result)
	if estado != "rechazado" {
		t.Fatalf("expected rejected DIAN acuse, got estado=%s mensaje=%s", estado, mensaje)
	}
	if !strings.Contains(mensaje, "Regla DIAN A fallida") || !strings.Contains(mensaje, "Regla DIAN B fallida") {
		t.Fatalf("expected joined error list as message, got %q", mensaje)
	}
}

func TestResolveDIANAcuseFromStatusDescriptionPending(t *testing.T) {
	response := map[string]interface{}{
		"is_valid":           "false",
		"status_description": "Batch en proceso de validacion.",
	}
	estado, mensaje := resolveDIANAcuseFromResponse(http.StatusOK, response)
	if estado != "pendiente" {
		t.Fatalf("expected pendiente, got estado=%s mensaje=%s", estado, mensaje)
	}
	if mensaje != "Batch en proceso de validacion." {
		t.Fatalf("expected DIAN status description as message, got %q", mensaje)
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
