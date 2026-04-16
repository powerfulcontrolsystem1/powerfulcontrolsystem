package handlers

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	dbpkg "github.com/you/pos-backend/db"
	"github.com/you/pos-backend/utils"
)

func ensurePaymentsHandlerTestSchema(t *testing.T, dbSuper *sql.DB) {
	t.Helper()

	_, err := dbSuper.Exec(`CREATE TABLE IF NOT EXISTS configuraciones (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		config_key TEXT UNIQUE,
		value TEXT,
		encrypted INTEGER DEFAULT 0,
		fecha_creacion TEXT,
		fecha_actualizacion TEXT
	);`)
	if err != nil {
		t.Fatalf("create configuraciones schema: %v", err)
	}

	_, err = dbSuper.Exec(`CREATE TABLE IF NOT EXISTS licencias (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		empresa_id INTEGER,
		tipo_id INTEGER,
		nombre TEXT,
		descripcion TEXT,
		valor REAL,
		duracion_dias INTEGER,
		modulos_habilitados TEXT,
		super_rol_habilitado INTEGER DEFAULT 0,
		fecha_creacion TEXT,
		activo INTEGER DEFAULT 1
	);`)
	if err != nil {
		t.Fatalf("create licencias schema: %v", err)
	}

	_, err = dbSuper.Exec(`CREATE TABLE IF NOT EXISTS pagos_epayco (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		licencia_id INTEGER,
		empresa_id INTEGER,
		transaction_id TEXT,
		reference TEXT,
		status TEXT,
		raw_payload TEXT,
		discount_code TEXT,
		asesor_id TEXT,
		fecha_creacion TEXT,
		fecha_actualizacion TEXT
	);`)
	if err != nil {
		t.Fatalf("create pagos_epayco schema: %v", err)
	}

	_, err = dbSuper.Exec(`CREATE TABLE IF NOT EXISTS pagos_wompi (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		licencia_id INTEGER,
		empresa_id INTEGER,
		transaction_id TEXT,
		reference TEXT,
		status TEXT,
		raw_payload TEXT,
		discount_code TEXT,
		asesor_id TEXT,
		fecha_creacion TEXT,
		fecha_actualizacion TEXT
	);`)
	if err != nil {
		t.Fatalf("create pagos_wompi schema: %v", err)
	}
}

func TestResolvePaymentBaseURLFallsBackToCanonicalDomainOnLocalhost(t *testing.T) {
	dbSuper := openTestSQLite(t, "super_resolve_payment_base_url_localhost.db")
	ensurePaymentsHandlerTestSchema(t, dbSuper)

	req := httptest.NewRequest(http.MethodGet, "http://localhost:8080/pagar_licencia.html", nil)
	baseURL, err := resolvePaymentBaseURL(req, dbSuper)
	if err != nil {
		t.Fatalf("resolvePaymentBaseURL returned error: %v", err)
	}
	if baseURL != canonicalPaymentPublicBaseURL {
		t.Fatalf("expected canonical fallback base URL %q, got %q", canonicalPaymentPublicBaseURL, baseURL)
	}
}

func TestResolvePaymentBaseURLUsesConfiguredCanonicalDomain(t *testing.T) {
	dbSuper := openTestSQLite(t, "super_resolve_payment_base_url_configured.db")
	ensurePaymentsHandlerTestSchema(t, dbSuper)

	if err := dbpkg.SetConfigValue(dbSuper, "gmail.confirm_base_url", "http://www.powerfulcontrolsystem.com", false); err != nil {
		t.Fatalf("seed gmail.confirm_base_url: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "http://localhost:8080/pagar_licencia.html", nil)
	baseURL, err := resolvePaymentBaseURL(req, dbSuper)
	if err != nil {
		t.Fatalf("resolvePaymentBaseURL returned error: %v", err)
	}
	if baseURL != "https://powerfulcontrolsystem.com" {
		t.Fatalf("expected canonical configured base URL, got %q", baseURL)
	}
}

func TestResolvePaymentBaseURLIgnoresConfiguredLocalhostAndFallsBackToCanonicalDomain(t *testing.T) {
	dbSuper := openTestSQLite(t, "super_resolve_payment_base_url_configured_localhost.db")
	ensurePaymentsHandlerTestSchema(t, dbSuper)

	if err := dbpkg.SetConfigValue(dbSuper, "gmail.confirm_base_url", "http://localhost:8080", false); err != nil {
		t.Fatalf("seed gmail.confirm_base_url: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "http://localhost:8080/pagar_licencia.html", nil)
	baseURL, err := resolvePaymentBaseURL(req, dbSuper)
	if err != nil {
		t.Fatalf("resolvePaymentBaseURL returned error: %v", err)
	}
	if baseURL != canonicalPaymentPublicBaseURL {
		t.Fatalf("expected canonical fallback base URL %q, got %q", canonicalPaymentPublicBaseURL, baseURL)
	}
}

func TestEpaycoCreateTransactionHandlerUsesConfiguredPublicBaseURLAndKeys(t *testing.T) {
	rawKey := make([]byte, 32)
	for i := range rawKey {
		rawKey[i] = byte(41 + i)
	}
	t.Setenv("CONFIG_ENC_KEY", base64.StdEncoding.EncodeToString(rawKey))

	dbSuper := openTestSQLite(t, "super_epayco_checkout_url.db")
	ensurePaymentsHandlerTestSchema(t, dbSuper)

	if _, err := dbSuper.Exec(`
		INSERT INTO licencias (id, empresa_id, tipo_id, nombre, descripcion, valor, duracion_dias, modulos_habilitados, super_rol_habilitado, fecha_creacion, activo)
		VALUES (1, 0, 1, 'Plan Pro', 'Licencia de prueba', 249900, 30, '', 0, datetime('now','localtime'), 1)
	`); err != nil {
		t.Fatalf("seed licencia: %v", err)
	}

	if err := dbpkg.SetConfigValue(dbSuper, "epayco.enabled", "1", false); err != nil {
		t.Fatalf("seed epayco.enabled: %v", err)
	}
	if err := dbpkg.SetConfigValue(dbSuper, "epayco.public_key", "pub_test_checkout_123", false); err != nil {
		t.Fatalf("seed epayco.public_key: %v", err)
	}
	if err := dbpkg.SetConfigValue(dbSuper, "epayco.customer_id", "1579238", false); err != nil {
		t.Fatalf("seed epayco.customer_id: %v", err)
	}
	encPrivateKey, err := utils.EncryptString("prv_test_checkout_456")
	if err != nil {
		t.Fatalf("encrypt epayco.private_key: %v", err)
	}
	if err := dbpkg.SetConfigValue(dbSuper, "epayco.private_key", encPrivateKey, true); err != nil {
		t.Fatalf("seed epayco.private_key: %v", err)
	}
	if err := dbpkg.SetConfigValue(dbSuper, "gmail.confirm_base_url", "https://powerfulcontrolsystem.com", false); err != nil {
		t.Fatalf("seed gmail.confirm_base_url: %v", err)
	}

	h := EpaycoCreateTransactionHandler(dbSuper)
	body := strings.NewReader(`{"licencia_id":1,"empresa_id":44,"customer_email":"cliente@demo.com"}`)
	req := httptest.NewRequest(http.MethodPost, "http://localhost:8080/api/public/licencias/payment/epayco", body)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var resp struct {
		CheckoutURL   string `json:"checkout_url"`
		Reference     string `json:"reference"`
		TransactionID string `json:"transaction_id"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v body=%s", err, rr.Body.String())
	}
	if resp.CheckoutURL == "" {
		t.Fatalf("expected checkout_url in response: %s", rr.Body.String())
	}
	if strings.Contains(resp.CheckoutURL, "localhost") {
		t.Fatalf("checkout URL must not include localhost: %s", resp.CheckoutURL)
	}

	parsed, err := url.Parse(resp.CheckoutURL)
	if err != nil {
		t.Fatalf("parse checkout URL: %v", err)
	}
	query := parsed.Query()
	if query.Get("public_key") != "pub_test_checkout_123" {
		t.Fatalf("expected public_key to use epayco.public_key, got %q", query.Get("public_key"))
	}
	if query.Get("p_cust_id_cliente") != "1579238" {
		t.Fatalf("expected p_cust_id_cliente to use customer_id, got %q", query.Get("p_cust_id_cliente"))
	}
	responseURL, err := url.Parse(query.Get("response"))
	if err != nil {
		t.Fatalf("parse response URL: %v", err)
	}
	responseQuery := responseURL.Query()
	if responseURL.Scheme != "https" || responseURL.Host != "powerfulcontrolsystem.com" || responseURL.Path != "/epayco/respuesta.html" {
		t.Fatalf("unexpected response URL target: %q", query.Get("response"))
	}
	if responseQuery.Get("provider") != "epayco" {
		t.Fatalf("expected response provider epayco, got %q", responseQuery.Get("provider"))
	}
	if responseQuery.Get("status") != "pending" {
		t.Fatalf("expected response status pending, got %q", responseQuery.Get("status"))
	}
	if responseQuery.Get("reference") != resp.Reference {
		t.Fatalf("expected response reference %q, got %q", resp.Reference, responseQuery.Get("reference"))
	}
	if responseQuery.Get("licencia_id") != "1" {
		t.Fatalf("expected response licencia_id 1, got %q", responseQuery.Get("licencia_id"))
	}
	if responseQuery.Get("extra1") != "1" {
		t.Fatalf("expected response extra1 1, got %q", responseQuery.Get("extra1"))
	}
	if responseQuery.Get("empresa_id") != "44" {
		t.Fatalf("expected response empresa_id 44, got %q", responseQuery.Get("empresa_id"))
	}
	if responseQuery.Get("extra2") != "44" {
		t.Fatalf("expected response extra2 44, got %q", responseQuery.Get("extra2"))
	}
	if query.Get("confirmation") != "https://powerfulcontrolsystem.com/epayco/webhook" {
		t.Fatalf("unexpected confirmation URL: %q", query.Get("confirmation"))
	}

	rec, err := dbpkg.GetEpaycoPaymentByReference(dbSuper, resp.Reference)
	if err != nil {
		t.Fatalf("read epayco record: %v", err)
	}
	if rec == nil {
		t.Fatalf("expected pagos_epayco record for reference %q", resp.Reference)
	}
	if !rec.TransactionID.Valid || strings.TrimSpace(rec.TransactionID.String) != resp.TransactionID {
		t.Fatalf("expected stored transaction_id %q, got %+v", resp.TransactionID, rec.TransactionID)
	}
	if !rec.RawPayload.Valid || !strings.Contains(rec.RawPayload.String, "https://powerfulcontrolsystem.com") {
		t.Fatalf("expected raw payload with public base URL, got %+v", rec.RawPayload)
	}
}

func TestEpaycoCreateTransactionHandlerAllowsCheckoutWithoutPrivateKey(t *testing.T) {
	dbSuper := openTestSQLite(t, "super_epayco_checkout_without_private_key.db")
	ensurePaymentsHandlerTestSchema(t, dbSuper)

	if _, err := dbSuper.Exec(`
		INSERT INTO licencias (id, empresa_id, tipo_id, nombre, descripcion, valor, duracion_dias, modulos_habilitados, super_rol_habilitado, fecha_creacion, activo)
		VALUES (1, 0, 1, 'Plan Base', 'Licencia sin llave secreta', 99900, 30, '', 0, datetime('now','localtime'), 1)
	`); err != nil {
		t.Fatalf("seed licencia: %v", err)
	}

	if err := dbpkg.SetConfigValue(dbSuper, "epayco.enabled", "1", false); err != nil {
		t.Fatalf("seed epayco.enabled: %v", err)
	}
	if err := dbpkg.SetConfigValue(dbSuper, "epayco.public_key", "pub_test_checkout_only_public", false); err != nil {
		t.Fatalf("seed epayco.public_key: %v", err)
	}
	if err := dbpkg.SetConfigValue(dbSuper, "epayco.customer_id", "1579238", false); err != nil {
		t.Fatalf("seed epayco.customer_id: %v", err)
	}
	if err := dbpkg.SetConfigValue(dbSuper, "gmail.confirm_base_url", "https://powerfulcontrolsystem.com", false); err != nil {
		t.Fatalf("seed gmail.confirm_base_url: %v", err)
	}

	h := EpaycoCreateTransactionHandler(dbSuper)
	body := strings.NewReader(`{"licencia_id":1,"empresa_id":77,"customer_email":"cliente@demo.com"}`)
	req := httptest.NewRequest(http.MethodPost, "http://localhost:8080/api/public/licencias/payment/epayco", body)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var resp struct {
		CheckoutURL string `json:"checkout_url"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v body=%s", err, rr.Body.String())
	}
	if resp.CheckoutURL == "" {
		t.Fatalf("expected checkout_url in response: %s", rr.Body.String())
	}
	parsed, err := url.Parse(resp.CheckoutURL)
	if err != nil {
		t.Fatalf("parse checkout URL: %v", err)
	}
	if got := parsed.Query().Get("public_key"); got != "pub_test_checkout_only_public" {
		t.Fatalf("expected public_key to use epayco.public_key, got %q", got)
	}
}

func TestEpaycoTransactionStatusHandlerPreservesPendingOnGenericValidationError(t *testing.T) {
	dbSuper := openTestSQLite(t, "super_epayco_status_pending_on_gateway_error.db")
	ensurePaymentsHandlerTestSchema(t, dbSuper)

	if _, err := dbpkg.CreateEpaycoPaymentRecord(dbSuper, 1, 44, "EPAYCO-TX-1", "EPAYCO-REF-1", "PENDING", `{}`, "", ""); err != nil {
		t.Fatalf("seed pagos_epayco: %v", err)
	}

	originalTransport := http.DefaultTransport
	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		body := `{"data":{"description":"Error de datos o conexión verifique de nuevo.","status":"error"},"message":"Error de datos o conexión.","status":false}`
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(body)),
		}, nil
	})
	defer func() { http.DefaultTransport = originalTransport }()

	h := EpaycoTransactionStatusHandler(dbSuper)
	req := httptest.NewRequest(http.MethodGet, "/epayco/transaction_status?reference=EPAYCO-REF-1", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var resp struct {
		Reference    string `json:"reference"`
		Status       string `json:"status"`
		ContextFound bool   `json:"context_found"`
		LicenciaID   int64  `json:"licencia_id"`
		EmpresaID    int64  `json:"empresa_id"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode epayco status response: %v body=%s", err, rr.Body.String())
	}
	if resp.Reference != "EPAYCO-REF-1" {
		t.Fatalf("expected reference EPAYCO-REF-1, got %q", resp.Reference)
	}
	if resp.Status != "PENDING" {
		t.Fatalf("expected status PENDING, got %q", resp.Status)
	}
	if !resp.ContextFound {
		t.Fatal("expected context_found=true for stored payment record")
	}
	if resp.LicenciaID != 1 || resp.EmpresaID != 44 {
		t.Fatalf("expected licencia_id=1 empresa_id=44, got licencia_id=%d empresa_id=%d", resp.LicenciaID, resp.EmpresaID)
	}

	rec, err := dbpkg.GetEpaycoPaymentByReference(dbSuper, "EPAYCO-REF-1")
	if err != nil {
		t.Fatalf("read epayco record: %v", err)
	}
	if rec == nil || !rec.Status.Valid || strings.ToUpper(strings.TrimSpace(rec.Status.String)) != "PENDING" {
		t.Fatalf("expected stored status to remain PENDING, got %+v", rec)
	}
}

func TestEpaycoTransactionStatusHandlerFindsContextUsingInvoiceWhenGatewayIDsDiffer(t *testing.T) {
	dbSuper := openTestSQLite(t, "super_epayco_status_context_invoice_fallback.db")
	ensurePaymentsHandlerTestSchema(t, dbSuper)

	if _, err := dbSuper.Exec(`
		INSERT INTO licencias (id, empresa_id, tipo_id, nombre, descripcion, valor, duracion_dias, modulos_habilitados, super_rol_habilitado, fecha_creacion, activo)
		VALUES (1, 0, 1, 'Plan Contexto', 'Prueba de contexto epayco', 129900, 30, '', 0, datetime('now','localtime'), 1)
	`); err != nil {
		t.Fatalf("seed licencia: %v", err)
	}

	internalRef := "EPAYCO-LIC-1-EMP-44-INT"
	if _, err := dbpkg.CreateEpaycoPaymentRecord(dbSuper, 1, 44, internalRef, internalRef, "PENDING", `{}`, "", ""); err != nil {
		t.Fatalf("seed pagos_epayco: %v", err)
	}

	originalTransport := http.DefaultTransport
	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		body := `{"data":{"x_transaction_id":"ep_tx_789","x_ref_payco":"ep_ref_999","invoice":"EPAYCO-LIC-1-EMP-44-INT","x_cod_response":"1","x_response":"Aceptada"},"status":true}`
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(body)),
		}, nil
	})
	defer func() { http.DefaultTransport = originalTransport }()

	h := EpaycoTransactionStatusHandler(dbSuper)
	req := httptest.NewRequest(http.MethodGet, "/epayco/transaction_status?id=ep_tx_789&reference=ep_ref_999", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var resp struct {
		Status       string `json:"status"`
		ContextFound bool   `json:"context_found"`
		LicenciaID   int64  `json:"licencia_id"`
		EmpresaID    int64  `json:"empresa_id"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode epayco status response: %v body=%s", err, rr.Body.String())
	}
	if resp.Status != "APPROVED" {
		t.Fatalf("expected status APPROVED, got %q", resp.Status)
	}
	if !resp.ContextFound {
		t.Fatal("expected context_found=true when invoice matches internal reference")
	}
	if resp.LicenciaID != 1 || resp.EmpresaID != 44 {
		t.Fatalf("expected licencia_id=1 empresa_id=44, got licencia_id=%d empresa_id=%d", resp.LicenciaID, resp.EmpresaID)
	}

	rec, err := dbpkg.GetEpaycoPaymentByReference(dbSuper, internalRef)
	if err != nil {
		t.Fatalf("read epayco record: %v", err)
	}
	if rec == nil || !rec.Status.Valid || strings.ToUpper(strings.TrimSpace(rec.Status.String)) != "APPROVED" {
		t.Fatalf("expected stored status APPROVED for internal reference, got %+v", rec)
	}
}

func TestWompiTransactionStatusHandlerAllowsReferenceLookup(t *testing.T) {
	dbSuper := openTestSQLite(t, "super_wompi_reference_lookup.db")
	ensurePaymentsHandlerTestSchema(t, dbSuper)

	if err := dbpkg.SetConfigValue(dbSuper, "wompi.public_key", "pub_test_reference_lookup", false); err != nil {
		t.Fatalf("seed wompi.public_key: %v", err)
	}
	if _, err := dbpkg.CreateWompiPaymentRecord(dbSuper, 1, 44, "wompi_tx_123", "WOMPI-LIC-1-EMP-44-REF", "PENDING", `{}`, "", ""); err != nil {
		t.Fatalf("seed pagos_wompi: %v", err)
	}

	originalTransport := http.DefaultTransport
	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		body := `{"data":{"id":"wompi_tx_123","status":"APPROVED","reference":"WOMPI-LIC-1-EMP-44-REF"}}`
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(body)),
		}, nil
	})
	defer func() { http.DefaultTransport = originalTransport }()

	h := WompiTransactionStatusHandler(dbSuper)
	req := httptest.NewRequest(http.MethodGet, "/wompi/transaction_status?reference=WOMPI-LIC-1-EMP-44-REF", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var resp struct {
		TransactionID string `json:"transaction_id"`
		Reference     string `json:"reference"`
		Status        string `json:"status"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode wompi status response: %v body=%s", err, rr.Body.String())
	}
	if resp.TransactionID != "wompi_tx_123" {
		t.Fatalf("expected transaction_id wompi_tx_123, got %q", resp.TransactionID)
	}
	if resp.Reference != "WOMPI-LIC-1-EMP-44-REF" {
		t.Fatalf("expected reference WOMPI-LIC-1-EMP-44-REF, got %q", resp.Reference)
	}
	if resp.Status != "APPROVED" {
		t.Fatalf("expected status APPROVED, got %q", resp.Status)
	}
}
