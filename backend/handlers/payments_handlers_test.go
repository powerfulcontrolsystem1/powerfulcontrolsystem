package handlers

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
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
		fecha_inicio TEXT,
		fecha_fin TEXT,
		fecha_creacion TEXT,
		activo INTEGER DEFAULT 1
	);`)
	if err != nil {
		t.Fatalf("create licencias schema: %v", err)
	}

	_, err = dbSuper.Exec(`CREATE TABLE IF NOT EXISTS tipos_de_empresas (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		nombre TEXT
	);`)
	if err != nil {
		t.Fatalf("create tipos_de_empresas schema: %v", err)
	}

	_, err = dbSuper.Exec(`CREATE TABLE IF NOT EXISTS empresas (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		nombre TEXT,
		usuario_creador TEXT
	);`)
	if err != nil {
		t.Fatalf("create empresas schema: %v", err)
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

	if err := dbpkg.EnsureSuperCorreoNotificacionesPruebaSchema(dbSuper); err != nil {
		t.Fatalf("create super_correo_notificaciones_prueba schema: %v", err)
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

func TestLicenciasHandlerGetReturnsHistorialFieldsForCreatorScope(t *testing.T) {
	dbSuper := openTestSQLite(t, "super_licencias_historial_scope.db")
	ensurePaymentsHandlerTestSchema(t, dbSuper)

	if _, err := dbSuper.Exec(`INSERT INTO tipos_de_empresas (id, nombre) VALUES (1, 'Restaurante')`); err != nil {
		t.Fatalf("seed tipo empresa: %v", err)
	}
	if _, err := dbSuper.Exec(`INSERT INTO empresas (id, nombre, usuario_creador) VALUES (10, 'Cafe Central', 'owner@demo.com'), (20, 'Hotel Ajeno', 'other@demo.com')`); err != nil {
		t.Fatalf("seed empresas: %v", err)
	}
	if _, err := dbSuper.Exec(`
		INSERT INTO licencias (
			id, empresa_id, tipo_id, nombre, descripcion, valor, duracion_dias, modulos_habilitados, super_rol_habilitado, fecha_inicio, fecha_fin, fecha_creacion, activo
		) VALUES
			(1, 10, 1, 'Plan Oro', 'Licencia vigente', 199900, 30, 'ventas,finanzas', 0, '2026-04-01 00:00:00', '2026-05-01 00:00:00', '2026-04-01 00:00:00', 1),
			(2, 20, 1, 'Plan Externo', 'Licencia de otra empresa', 99900, 15, 'ventas', 0, '2026-04-02 00:00:00', '2026-04-17 00:00:00', '2026-04-02 00:00:00', 1)
	`); err != nil {
		t.Fatalf("seed licencias: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/super/api/licencias?con_empresa=1&usuario_creador=owner@demo.com", nil)
	rr := httptest.NewRecorder()
	LicenciasHandler(dbSuper).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var items []dbpkg.Licencia
	if err := json.Unmarshal(rr.Body.Bytes(), &items); err != nil {
		t.Fatalf("decode licencias response: %v body=%s", err, rr.Body.String())
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 licencia for owner scope, got %d: %+v", len(items), items)
	}
	if items[0].EmpresaID != 10 {
		t.Fatalf("expected empresa_id 10, got %d", items[0].EmpresaID)
	}
	if items[0].EmpresaNombre != "Cafe Central" {
		t.Fatalf("expected empresa nombre Cafe Central, got %q", items[0].EmpresaNombre)
	}
	if items[0].FechaInicio != "2026-04-01 00:00:00" {
		t.Fatalf("expected fecha_inicio in response, got %q", items[0].FechaInicio)
	}
	if items[0].FechaFin != "2026-05-01 00:00:00" {
		t.Fatalf("expected fecha_fin in response, got %q", items[0].FechaFin)
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

	var loginAuth string
	var sessionAuth string
	var sessionRequest map[string]interface{}
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/login":
			loginAuth = r.Header.Get("Authorization")
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"token":"apify-token-demo"}`))
		case "/payment/session/create":
			sessionAuth = r.Header.Get("Authorization")
			body, _ := io.ReadAll(r.Body)
			if err := json.Unmarshal(body, &sessionRequest); err != nil {
				t.Fatalf("decode session request: %v body=%s", err, string(body))
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"success":true,"data":{"sessionId":"sess_demo_123"}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer testServer.Close()
	originalApifyBaseURL := epaycoApifyBaseURL
	epaycoApifyBaseURL = testServer.URL
	defer func() { epaycoApifyBaseURL = originalApifyBaseURL }()

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
		Reference         string `json:"reference"`
		TransactionID     string `json:"transaction_id"`
		SessionID         string `json:"session_id"`
		CheckoutType      string `json:"checkout_type"`
		CheckoutScriptURL string `json:"checkout_script_url"`
		PaymentMethod     string `json:"payment_method"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v body=%s", err, rr.Body.String())
	}
	if resp.SessionID != "sess_demo_123" {
		t.Fatalf("expected session_id sess_demo_123, got %q", resp.SessionID)
	}
	if resp.CheckoutType != "standard" {
		t.Fatalf("expected checkout_type standard, got %q", resp.CheckoutType)
	}
	if resp.CheckoutScriptURL != "https://checkout.epayco.co/checkout-v2.js" {
		t.Fatalf("unexpected checkout_script_url: %q", resp.CheckoutScriptURL)
	}
	if resp.PaymentMethod != "SMART_CHECKOUT" {
		t.Fatalf("expected payment_method SMART_CHECKOUT, got %q", resp.PaymentMethod)
	}
	if loginAuth == "" || !strings.HasPrefix(loginAuth, "Basic ") {
		t.Fatalf("expected Basic auth on apify login, got %q", loginAuth)
	}
	decodedLoginAuth, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(loginAuth, "Basic "))
	if err != nil {
		t.Fatalf("decode login auth: %v", err)
	}
	if string(decodedLoginAuth) != "pub_test_checkout_123:prv_test_checkout_456" {
		t.Fatalf("unexpected login auth payload: %q", string(decodedLoginAuth))
	}
	if sessionAuth != "Bearer apify-token-demo" {
		t.Fatalf("expected Bearer token on session create, got %q", sessionAuth)
	}
	if sessionRequest["invoice"] != resp.Reference {
		t.Fatalf("expected invoice %q, got %#v", resp.Reference, sessionRequest["invoice"])
	}
	if sessionRequest["currency"] != "COP" {
		t.Fatalf("expected currency COP, got %#v", sessionRequest["currency"])
	}
	responseURL, err := url.Parse(strings.TrimSpace(sessionRequest["response"].(string)))
	if err != nil {
		t.Fatalf("parse response URL: %v", err)
	}
	responseQuery := responseURL.Query()
	if responseURL.Scheme != "https" || responseURL.Host != "powerfulcontrolsystem.com" || responseURL.Path != "/epayco/respuesta.html" {
		t.Fatalf("unexpected response URL target: %q", sessionRequest["response"])
	}
	if responseQuery.Get("provider") != "epayco" || responseQuery.Get("status") != "pending" {
		t.Fatalf("unexpected response query: %s", responseURL.RawQuery)
	}
	if responseQuery.Get("reference") != resp.Reference || responseQuery.Get("licencia_id") != "1" || responseQuery.Get("empresa_id") != "44" {
		t.Fatalf("unexpected response query values: %s", responseURL.RawQuery)
	}
	if sessionRequest["confirmation"] != "https://powerfulcontrolsystem.com/epayco/webhook" {
		t.Fatalf("unexpected confirmation URL: %#v", sessionRequest["confirmation"])
	}
	extras, _ := sessionRequest["extras"].(map[string]interface{})
	if fmt.Sprint(extras["extra1"]) != "1" || fmt.Sprint(extras["extra2"]) != "44" || fmt.Sprint(extras["extra3"]) != resp.Reference {
		t.Fatalf("unexpected extras payload: %#v", extras)
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
	if !rec.RawPayload.Valid || !strings.Contains(rec.RawPayload.String, "sess_demo_123") || !strings.Contains(rec.RawPayload.String, "smart_checkout_v2") {
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

	if rr.Code != http.StatusPreconditionFailed {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusPreconditionFailed, rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "Private Key") {
		t.Fatalf("expected response to mention missing Private Key, got %s", rr.Body.String())
	}
}

func TestEpaycoCreateTransactionHandlerAcceptsSamboxAlias(t *testing.T) {
	rawKey := make([]byte, 32)
	for i := range rawKey {
		rawKey[i] = byte(73 + i)
	}
	t.Setenv("CONFIG_ENC_KEY", base64.StdEncoding.EncodeToString(rawKey))

	dbSuper := openTestSQLite(t, "super_epayco_checkout_sambox_alias.db")
	ensurePaymentsHandlerTestSchema(t, dbSuper)

	if _, err := dbSuper.Exec(`
		INSERT INTO licencias (id, empresa_id, tipo_id, nombre, descripcion, valor, duracion_dias, modulos_habilitados, super_rol_habilitado, fecha_creacion, activo)
		VALUES (1, 0, 1, 'Plan Sandbox', 'Licencia con modo sambox', 149900, 30, '', 0, datetime('now','localtime'), 1)
	`); err != nil {
		t.Fatalf("seed licencia: %v", err)
	}

	if err := dbpkg.SetConfigValue(dbSuper, "epayco.enabled", "1", false); err != nil {
		t.Fatalf("seed epayco.enabled: %v", err)
	}
	if err := dbpkg.SetConfigValue(dbSuper, "epayco.mode", "sambox", false); err != nil {
		t.Fatalf("seed epayco.mode: %v", err)
	}
	if err := dbpkg.SetConfigValue(dbSuper, "epayco.public_key", "pub_prod_like_alias_test", false); err != nil {
		t.Fatalf("seed epayco.public_key: %v", err)
	}
	encPrivateKey, err := utils.EncryptString("prv_prod_like_alias_test")
	if err != nil {
		t.Fatalf("encrypt epayco.private_key: %v", err)
	}
	if err := dbpkg.SetConfigValue(dbSuper, "epayco.private_key", encPrivateKey, true); err != nil {
		t.Fatalf("seed epayco.private_key: %v", err)
	}
	if err := dbpkg.SetConfigValue(dbSuper, "gmail.confirm_base_url", "https://powerfulcontrolsystem.com", false); err != nil {
		t.Fatalf("seed gmail.confirm_base_url: %v", err)
	}

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/login":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"token":"apify-token-sandbox"}`))
		case "/payment/session/create":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"success":true,"data":{"sessionId":"sess_sandbox_456"}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer testServer.Close()
	originalApifyBaseURL := epaycoApifyBaseURL
	epaycoApifyBaseURL = testServer.URL
	defer func() { epaycoApifyBaseURL = originalApifyBaseURL }()

	h := EpaycoCreateTransactionHandler(dbSuper)
	body := strings.NewReader(`{"licencia_id":1,"empresa_id":55,"customer_email":"sandbox@demo.com"}`)
	req := httptest.NewRequest(http.MethodPost, "http://localhost:8080/epayco/create_transaction", body)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var resp struct {
		Mode      string `json:"mode"`
		SessionID string `json:"session_id"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v body=%s", err, rr.Body.String())
	}
	if resp.Mode != "sandbox" {
		t.Fatalf("expected mode sandbox when epayco.mode=sambox, got %q", resp.Mode)
	}
	if resp.SessionID != "sess_sandbox_456" {
		t.Fatalf("expected session_id sess_sandbox_456, got %q", resp.SessionID)
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

func TestEpaycoTransactionStatusHandlerActivatesOnceAndCapturesEmail(t *testing.T) {
	t.Setenv("PCS_MAIL_TEST_MODE", "1")

	dbSuper := openTestSQLite(t, "super_epayco_status_activation_email.db")
	ensurePaymentsHandlerTestSchema(t, dbSuper)

	if _, err := dbSuper.Exec(`
		INSERT INTO empresas (id, nombre, usuario_creador) VALUES (44, 'Hotel Demo', 'owner@demo.com')
	`); err != nil {
		t.Fatalf("seed empresa: %v", err)
	}
	if _, err := dbSuper.Exec(`
		INSERT INTO licencias (id, empresa_id, tipo_id, nombre, descripcion, valor, duracion_dias, modulos_habilitados, super_rol_habilitado, fecha_creacion, activo)
		VALUES (1, 0, 1, 'Plan Correo', 'Prueba activacion con correo', 159900, 30, '', 0, datetime('now','localtime'), 1)
	`); err != nil {
		t.Fatalf("seed licencia: %v", err)
	}

	internalRef := "EPAYCO-LIC-1-EMP-44-EMAIL"
	if _, err := dbpkg.CreateEpaycoPaymentRecord(dbSuper, 1, 44, internalRef, internalRef, "PENDING", `{"customer_email":"cliente@demo.com","provider":"epayco"}`, "", ""); err != nil {
		t.Fatalf("seed pagos_epayco: %v", err)
	}

	originalTransport := http.DefaultTransport
	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		body := `{"data":{"x_transaction_id":"ep_tx_email_1","x_ref_payco":"ep_ref_email_1","invoice":"EPAYCO-LIC-1-EMP-44-EMAIL","x_cod_response":"1","x_response":"Aceptada"},"status":true}`
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(body)),
		}, nil
	})
	defer func() { http.DefaultTransport = originalTransport }()

	h := EpaycoTransactionStatusHandler(dbSuper)
	req := httptest.NewRequest(http.MethodGet, "/epayco/transaction_status?id=ep_tx_email_1&reference=ep_ref_email_1", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var resp struct {
		Status    string `json:"status"`
		Activated bool   `json:"activated"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode first response: %v body=%s", err, rr.Body.String())
	}
	if resp.Status != "APPROVED" {
		t.Fatalf("expected status APPROVED, got %q", resp.Status)
	}
	if !resp.Activated {
		t.Fatal("expected activated=true on first approval")
	}

	var notifCount int
	if err := dbSuper.QueryRow(`SELECT COUNT(*) FROM super_correo_notificaciones_prueba WHERE tipo = 'licencia_activada_pago' AND destinatario = 'cliente@demo.com'`).Scan(&notifCount); err != nil {
		t.Fatalf("count notifications: %v", err)
	}
	if notifCount != 1 {
		t.Fatalf("expected 1 activation email capture, got %d", notifCount)
	}

	lic, err := dbpkg.GetLicenciaByID(dbSuper, 1)
	if err != nil {
		t.Fatalf("reload licencia: %v", err)
	}
	if lic == nil || lic.EmpresaID != 44 || lic.Activo != 1 {
		t.Fatalf("expected licencia active for empresa 44, got %+v", lic)
	}

	rec, err := dbpkg.GetEpaycoPaymentByReference(dbSuper, internalRef)
	if err != nil {
		t.Fatalf("reload epayco record: %v", err)
	}
	if rec == nil || !rec.RawPayload.Valid || !strings.Contains(rec.RawPayload.String, `"customer_email":"cliente@demo.com"`) {
		t.Fatalf("expected raw_payload to preserve customer_email, got %+v", rec)
	}

	rrSecond := httptest.NewRecorder()
	h.ServeHTTP(rrSecond, req)
	if rrSecond.Code != http.StatusOK {
		t.Fatalf("expected second status %d, got %d body=%s", http.StatusOK, rrSecond.Code, rrSecond.Body.String())
	}
	var secondResp struct {
		Activated bool `json:"activated"`
	}
	if err := json.Unmarshal(rrSecond.Body.Bytes(), &secondResp); err != nil {
		t.Fatalf("decode second response: %v body=%s", err, rrSecond.Body.String())
	}
	if secondResp.Activated {
		t.Fatal("expected activated=false on repeated approved poll")
	}
	if err := dbSuper.QueryRow(`SELECT COUNT(*) FROM super_correo_notificaciones_prueba WHERE tipo = 'licencia_activada_pago' AND destinatario = 'cliente@demo.com'`).Scan(&notifCount); err != nil {
		t.Fatalf("count notifications after second poll: %v", err)
	}
	if notifCount != 1 {
		t.Fatalf("expected notification count to remain 1 after repeated poll, got %d", notifCount)
	}
}

func TestEpaycoTransactionStatusHandlerUsesEmpresaScopeForActivationMailBody(t *testing.T) {
	t.Setenv("PCS_MAIL_TEST_MODE", "1")

	dbSuper := openTestSQLite(t, "super_epayco_scope_activation_mail.db")
	ensurePaymentsHandlerTestSchema(t, dbSuper)

	if _, err := dbSuper.Exec(`ALTER TABLE empresas ADD COLUMN empresa_id INTEGER`); err != nil {
		t.Fatalf("extend empresas with empresa_id: %v", err)
	}
	if _, err := dbSuper.Exec(`ALTER TABLE empresas ADD COLUMN nit TEXT`); err != nil {
		t.Fatalf("extend empresas with nit: %v", err)
	}
	if _, err := dbSuper.Exec(`ALTER TABLE empresas ADD COLUMN tipo_id INTEGER`); err != nil {
		t.Fatalf("extend empresas with tipo_id: %v", err)
	}
	if _, err := dbSuper.Exec(`ALTER TABLE empresas ADD COLUMN tipo_nombre TEXT`); err != nil {
		t.Fatalf("extend empresas with tipo_nombre: %v", err)
	}
	if _, err := dbSuper.Exec(`ALTER TABLE empresas ADD COLUMN fecha_creacion TEXT`); err != nil {
		t.Fatalf("extend empresas with fecha_creacion: %v", err)
	}
	if _, err := dbSuper.Exec(`ALTER TABLE empresas ADD COLUMN fecha_actualizacion TEXT`); err != nil {
		t.Fatalf("extend empresas with fecha_actualizacion: %v", err)
	}
	if _, err := dbSuper.Exec(`ALTER TABLE empresas ADD COLUMN estado TEXT`); err != nil {
		t.Fatalf("extend empresas with estado: %v", err)
	}
	if _, err := dbSuper.Exec(`ALTER TABLE empresas ADD COLUMN observaciones TEXT`); err != nil {
		t.Fatalf("extend empresas with observaciones: %v", err)
	}

	if _, err := dbSuper.Exec(`INSERT INTO empresas (id, empresa_id, nombre, usuario_creador) VALUES (91, 77, 'Hotel Scope Demo', 'scope-owner@demo.com')`); err != nil {
		t.Fatalf("seed scoped empresa: %v", err)
	}
	if _, err := dbSuper.Exec(`
		INSERT INTO licencias (id, empresa_id, tipo_id, nombre, descripcion, valor, duracion_dias, modulos_habilitados, super_rol_habilitado, fecha_creacion, activo)
		VALUES (1, 0, 1, 'Plan Scope Mail', 'Prueba correo por alcance logico', 159900, 30, '', 0, datetime('now','localtime'), 1)
	`); err != nil {
		t.Fatalf("seed licencia: %v", err)
	}

	internalRef := "EPAYCO-LIC-1-EMP-77-SCOPE"
	if _, err := dbpkg.CreateEpaycoPaymentRecord(dbSuper, 1, 77, internalRef, internalRef, "PENDING", `{"customer_email":"scope-client@demo.com","provider":"epayco"}`, "", ""); err != nil {
		t.Fatalf("seed pagos_epayco: %v", err)
	}

	originalTransport := http.DefaultTransport
	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		body := `{"data":{"x_transaction_id":"ep_tx_scope_1","x_ref_payco":"ep_ref_scope_1","invoice":"EPAYCO-LIC-1-EMP-77-SCOPE","x_cod_response":"1","x_response":"Aceptada"},"status":true}`
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(body)),
		}, nil
	})
	defer func() { http.DefaultTransport = originalTransport }()

	h := EpaycoTransactionStatusHandler(dbSuper)
	req := httptest.NewRequest(http.MethodGet, "/epayco/transaction_status?id=ep_tx_scope_1&reference=ep_ref_scope_1&licencia_id=1&empresa_id=77", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var cuerpo string
	if err := dbSuper.QueryRow(`SELECT cuerpo FROM super_correo_notificaciones_prueba WHERE tipo = 'licencia_activada_pago' AND destinatario = 'scope-client@demo.com' ORDER BY id DESC LIMIT 1`).Scan(&cuerpo); err != nil {
		t.Fatalf("load captured mail body: %v", err)
	}
	if !strings.Contains(cuerpo, "Empresa: Hotel Scope Demo") {
		t.Fatalf("expected mail body to include scoped company name, got %q", cuerpo)
	}
}

func TestEpaycoTransactionStatusHandlerRejectsUnexpectedEmpresaContext(t *testing.T) {
	dbSuper := openTestSQLite(t, "super_epayco_context_mismatch.db")
	ensurePaymentsHandlerTestSchema(t, dbSuper)

	if _, err := dbSuper.Exec(`INSERT INTO empresas (id, nombre, usuario_creador) VALUES (7, 'Hotel Siete', 'owner7@demo.com')`); err != nil {
		t.Fatalf("seed empresa: %v", err)
	}
	if _, err := dbSuper.Exec(`
		INSERT INTO licencias (id, empresa_id, tipo_id, nombre, descripcion, valor, duracion_dias, modulos_habilitados, super_rol_habilitado, fecha_creacion, activo)
		VALUES (1, 0, 1, 'Plan Mismatch', 'Prueba mismatch de contexto', 159900, 30, '', 0, datetime('now','localtime'), 1)
	`); err != nil {
		t.Fatalf("seed licencia: %v", err)
	}

	internalRef := "EPAYCO-LIC-1-EMP-7-MISMATCH"
	if _, err := dbpkg.CreateEpaycoPaymentRecord(dbSuper, 1, 7, internalRef, internalRef, "PENDING", `{"customer_email":"mismatch@demo.com","provider":"epayco"}`, "", ""); err != nil {
		t.Fatalf("seed pagos_epayco: %v", err)
	}

	originalTransport := http.DefaultTransport
	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		body := `{"data":{"x_transaction_id":"ep_tx_mismatch_1","x_ref_payco":"ep_ref_mismatch_1","invoice":"EPAYCO-LIC-1-EMP-7-MISMATCH","x_cod_response":"1","x_response":"Aceptada"},"status":true}`
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(body)),
		}, nil
	})
	defer func() { http.DefaultTransport = originalTransport }()

	h := EpaycoTransactionStatusHandler(dbSuper)
	req := httptest.NewRequest(http.MethodGet, "/epayco/transaction_status?id=ep_tx_mismatch_1&reference=ep_ref_mismatch_1&licencia_id=1&empresa_id=6", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusConflict, rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "no corresponde a la empresa o licencia abierta") {
		t.Fatalf("expected mismatch message, got %s", rr.Body.String())
	}
}

func TestEpaycoWebhookHandlerFindsContextUsingInvoiceFallback(t *testing.T) {
	t.Setenv("PCS_MAIL_TEST_MODE", "1")

	dbSuper := openTestSQLite(t, "super_epayco_webhook_invoice_fallback.db")
	ensurePaymentsHandlerTestSchema(t, dbSuper)

	if _, err := dbSuper.Exec(`INSERT INTO empresas (id, nombre, usuario_creador) VALUES (55, 'Motel Demo', 'owner55@demo.com')`); err != nil {
		t.Fatalf("seed empresa: %v", err)
	}
	if _, err := dbSuper.Exec(`
		INSERT INTO licencias (id, empresa_id, tipo_id, nombre, descripcion, valor, duracion_dias, modulos_habilitados, super_rol_habilitado, fecha_creacion, activo)
		VALUES (1, 0, 1, 'Plan Webhook', 'Prueba webhook epayco', 189900, 30, '', 0, datetime('now','localtime'), 1)
	`); err != nil {
		t.Fatalf("seed licencia: %v", err)
	}

	internalRef := "EPAYCO-LIC-1-EMP-55-INVOICE"
	if _, err := dbpkg.CreateEpaycoPaymentRecord(dbSuper, 1, 55, internalRef, internalRef, "PENDING", `{"customer_email":"webhook@demo.com","provider":"epayco"}`, "", ""); err != nil {
		t.Fatalf("seed pagos_epayco: %v", err)
	}

	h := EpaycoWebhookHandler(dbSuper)
	body := strings.NewReader(`{"x_transaction_id":"ep_tx_webhook_1","x_ref_payco":"ep_ref_webhook_1","invoice":"EPAYCO-LIC-1-EMP-55-INVOICE","x_cod_response":"1","x_response":"Aceptada"}`)
	req := httptest.NewRequest(http.MethodPost, "/epayco/webhook", body)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var resp struct {
		Status       string `json:"status"`
		ContextFound bool   `json:"context_found"`
		Activated    bool   `json:"activated"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode webhook response: %v body=%s", err, rr.Body.String())
	}
	if resp.Status != "APPROVED" {
		t.Fatalf("expected status APPROVED, got %q", resp.Status)
	}
	if !resp.ContextFound {
		t.Fatal("expected context_found=true using invoice fallback")
	}
	if !resp.Activated {
		t.Fatal("expected activated=true on webhook approval")
	}

	lic, err := dbpkg.GetLicenciaByID(dbSuper, 1)
	if err != nil {
		t.Fatalf("reload licencia after webhook: %v", err)
	}
	if lic == nil || lic.EmpresaID != 55 || lic.Activo != 1 {
		t.Fatalf("expected licencia active for empresa 55, got %+v", lic)
	}

	rec, err := dbpkg.GetEpaycoPaymentByReference(dbSuper, internalRef)
	if err != nil {
		t.Fatalf("reload record by invoice ref: %v", err)
	}
	if rec == nil || !rec.Status.Valid || strings.ToUpper(strings.TrimSpace(rec.Status.String)) != "APPROVED" {
		t.Fatalf("expected internal reference record approved after webhook, got %+v", rec)
	}
	}


func TestEpaycoTransactionStatusHandlerRetriesActivationEmailAfterWebhookActivatedFirst(t *testing.T) {
	t.Setenv("PCS_MAIL_TEST_MODE", "1")

	dbSuper := openTestSQLite(t, "super_epayco_retry_activation_mail.db")
	ensurePaymentsHandlerTestSchema(t, dbSuper)

	if _, err := dbSuper.Exec(`INSERT INTO empresas (id, nombre, usuario_creador) VALUES (66, 'Hotel Reintento', 'owner66@demo.com')`); err != nil {
		t.Fatalf("seed empresa: %v", err)
	}
	if _, err := dbSuper.Exec(`
		INSERT INTO licencias (id, empresa_id, tipo_id, nombre, descripcion, valor, duracion_dias, modulos_habilitados, super_rol_habilitado, fecha_creacion, activo)
		VALUES (1, 0, 1, 'Plan Retry Mail', 'Prueba reintento email epayco', 199900, 30, '', 0, datetime('now','localtime'), 1)
	`); err != nil {
		t.Fatalf("seed licencia: %v", err)
	}

	internalRef := "EPAYCO-LIC-1-EMP-66-RETRY"
	if _, err := dbpkg.CreateEpaycoPaymentRecord(dbSuper, 1, 66, internalRef, internalRef, "PENDING", `{"provider":"epayco"}`, "", ""); err != nil {
		t.Fatalf("seed pagos_epayco: %v", err)
	}

	webhookHandler := EpaycoWebhookHandler(dbSuper)
	webhookBody := strings.NewReader(`{"x_transaction_id":"ep_tx_retry_1","x_ref_payco":"ep_ref_retry_1","invoice":"EPAYCO-LIC-1-EMP-66-RETRY","x_cod_response":"1","x_response":"Aceptada"}`)
	webhookReq := httptest.NewRequest(http.MethodPost, "/epayco/webhook", webhookBody)
	webhookReq.Header.Set("Content-Type", "application/json")
	webhookRR := httptest.NewRecorder()
	webhookHandler.ServeHTTP(webhookRR, webhookReq)

	if webhookRR.Code != http.StatusOK {
		t.Fatalf("expected webhook status %d, got %d body=%s", http.StatusOK, webhookRR.Code, webhookRR.Body.String())
	}

	var notifCount int
	if err := dbSuper.QueryRow(`SELECT COUNT(*) FROM super_correo_notificaciones_prueba WHERE tipo = 'licencia_activada_pago'`).Scan(&notifCount); err != nil {
		t.Fatalf("count notifications after webhook: %v", err)
	}
	if notifCount != 0 {
		t.Fatalf("expected 0 activation email captures after webhook without recipient, got %d", notifCount)
	}

	originalTransport := http.DefaultTransport
	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		body := `{"data":{"x_transaction_id":"ep_tx_retry_1","x_ref_payco":"ep_ref_retry_1","invoice":"EPAYCO-LIC-1-EMP-66-RETRY","customer_email":"cliente.retry@demo.com","x_cod_response":"1","x_response":"Aceptada"},"status":true}`
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(body)),
		}, nil
	})
	defer func() { http.DefaultTransport = originalTransport }()

	statusHandler := EpaycoTransactionStatusHandler(dbSuper)
	statusReq := httptest.NewRequest(http.MethodGet, "/epayco/transaction_status?id=ep_tx_retry_1&reference=ep_ref_retry_1", nil)
	statusRR := httptest.NewRecorder()
	statusHandler.ServeHTTP(statusRR, statusReq)

	if statusRR.Code != http.StatusOK {
		t.Fatalf("expected status poll %d, got %d body=%s", http.StatusOK, statusRR.Code, statusRR.Body.String())
	}

	if err := dbSuper.QueryRow(`SELECT COUNT(*) FROM super_correo_notificaciones_prueba WHERE tipo = 'licencia_activada_pago' AND destinatario = 'cliente.retry@demo.com'`).Scan(&notifCount); err != nil {
		t.Fatalf("count notifications after status poll: %v", err)
	}
	if notifCount != 1 {
		t.Fatalf("expected 1 activation email capture after retry poll, got %d", notifCount)
	}

	rec, err := dbpkg.GetEpaycoPaymentByReference(dbSuper, internalRef)
	if err != nil {
		t.Fatalf("reload epayco record: %v", err)
	}
	if rec == nil || !rec.RawPayload.Valid || !strings.Contains(rec.RawPayload.String, `"licencia_activation_email_sent":true`) {
		t.Fatalf("expected raw_payload to mark activation email as sent, got %+v", rec)
	}

	statusRRSecond := httptest.NewRecorder()
	statusHandler.ServeHTTP(statusRRSecond, statusReq)
	if statusRRSecond.Code != http.StatusOK {
		t.Fatalf("expected second status poll %d, got %d body=%s", http.StatusOK, statusRRSecond.Code, statusRRSecond.Body.String())
	}
	if err := dbSuper.QueryRow(`SELECT COUNT(*) FROM super_correo_notificaciones_prueba WHERE tipo = 'licencia_activada_pago' AND destinatario = 'cliente.retry@demo.com'`).Scan(&notifCount); err != nil {
		t.Fatalf("count notifications after second retry poll: %v", err)
	}
	if notifCount != 1 {
		t.Fatalf("expected activation email capture to remain 1 after second poll, got %d", notifCount)
	}
}

func TestLicenciaCheckoutSummaryHandlerAllowsZeroTotalByConfiguredDiscount(t *testing.T) {
	dbSuper := openTestSQLite(t, "super_licencia_checkout_summary_discount.db")
	ensurePaymentsHandlerTestSchema(t, dbSuper)

	if _, err := dbSuper.Exec(`
		INSERT INTO licencias (id, empresa_id, tipo_id, nombre, descripcion, valor, duracion_dias, modulos_habilitados, super_rol_habilitado, fecha_creacion, activo)
		VALUES (1, 0, 1, 'Plan Promo', 'Licencia con descuento total', 120000, 30, '', 0, datetime('now','localtime'), 1)
	`); err != nil {
		t.Fatalf("seed licencia: %v", err)
	}
	if err := dbpkg.SetConfigValue(dbSuper, "licencias.discount_codes", "PROMO100=100%", false); err != nil {
		t.Fatalf("seed licencias.discount_codes: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/public/licencias/checkout_summary?licencia_id=1&empresa_id=77&discount_code=PROMO100", nil)
	rr := httptest.NewRecorder()
	LicenciaCheckoutSummaryHandler(dbSuper).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}
	var resp struct {
		Summary struct {
			OriginalValue             float64 `json:"original_value"`
			DiscountValue             float64 `json:"discount_value"`
			TotalValue                float64 `json:"total_value"`
			DiscountApplied           bool    `json:"discount_applied"`
			IsZeroTotal               bool    `json:"is_zero_total"`
			CanActivateWithoutPayment bool    `json:"can_activate_without_payment"`
		} `json:"summary"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode summary response: %v body=%s", err, rr.Body.String())
	}
	if resp.Summary.OriginalValue != 120000 || resp.Summary.DiscountValue != 120000 || resp.Summary.TotalValue != 0 {
		t.Fatalf("unexpected summary values: %+v", resp.Summary)
	}
	if !resp.Summary.DiscountApplied || !resp.Summary.IsZeroTotal || !resp.Summary.CanActivateWithoutPayment {
		t.Fatalf("expected zero-total summary ready for activation, got %+v", resp.Summary)
	}
}

func TestActivateLicenciaSinPagoHandlerBlocksRepeatedFreeLicensePerEmpresa(t *testing.T) {
	dbSuper := openTestSQLite(t, "super_activate_licencia_sin_pago_once.db")
	ensurePaymentsHandlerTestSchema(t, dbSuper)

	if _, err := dbSuper.Exec(`
		INSERT INTO licencias (id, empresa_id, tipo_id, nombre, descripcion, valor, duracion_dias, modulos_habilitados, super_rol_habilitado, fecha_creacion, activo)
		VALUES (1, 0, 1, 'Plan Gratis', 'Licencia sin costo', 0, 30, '', 0, datetime('now','localtime'), 1)
	`); err != nil {
		t.Fatalf("seed licencia: %v", err)
	}

	h := ActivateLicenciaSinPagoHandler(dbSuper)
	firstReq := httptest.NewRequest(http.MethodPost, "/licencias/activar_sin_pago", strings.NewReader(`{"licencia_id":1,"empresa_id":15,"motivo":"licencia_valor_cero"}`))
	firstReq.Header.Set("Content-Type", "application/json")
	firstRR := httptest.NewRecorder()
	h.ServeHTTP(firstRR, firstReq)
	if firstRR.Code != http.StatusOK {
		t.Fatalf("expected first status %d, got %d body=%s", http.StatusOK, firstRR.Code, firstRR.Body.String())
	}

	secondReq := httptest.NewRequest(http.MethodPost, "/licencias/activar_sin_pago", strings.NewReader(`{"licencia_id":1,"empresa_id":15,"motivo":"licencia_valor_cero"}`))
	secondReq.Header.Set("Content-Type", "application/json")
	secondRR := httptest.NewRecorder()
	h.ServeHTTP(secondRR, secondReq)
	if secondRR.Code != http.StatusConflict {
		t.Fatalf("expected second status %d, got %d body=%s", http.StatusConflict, secondRR.Code, secondRR.Body.String())
	}

	var count int
	if err := dbSuper.QueryRow(`SELECT COUNT(*) FROM licencias_activaciones_gratis WHERE licencia_id = 1 AND empresa_id = 15`).Scan(&count); err != nil {
		t.Fatalf("count licencias_activaciones_gratis: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected single free activation record, got %d", count)
	}
}

func TestWompiCreateNequiTransactionHandlerRejectsZeroTotalAndSuggestsActivation(t *testing.T) {
	dbSuper := openTestSQLite(t, "super_wompi_zero_total_requires_activation.db")
	ensurePaymentsHandlerTestSchema(t, dbSuper)

	if _, err := dbSuper.Exec(`
		INSERT INTO licencias (id, empresa_id, tipo_id, nombre, descripcion, valor, duracion_dias, modulos_habilitados, super_rol_habilitado, fecha_creacion, activo)
		VALUES (1, 0, 1, 'Plan Descuento Total', 'Licencia con total cero', 95000, 30, '', 0, datetime('now','localtime'), 1)
	`); err != nil {
		t.Fatalf("seed licencia: %v", err)
	}
	if err := dbpkg.SetConfigValue(dbSuper, "wompi.public_key", "pub_test", false); err != nil {
		t.Fatalf("seed wompi.public_key: %v", err)
	}
	if err := dbpkg.SetConfigValue(dbSuper, "wompi.private_key", "priv_test", false); err != nil {
		t.Fatalf("seed wompi.private_key: %v", err)
	}
	if err := dbpkg.SetConfigValue(dbSuper, "wompi.integrity_key", "int_test", false); err != nil {
		t.Fatalf("seed wompi.integrity_key: %v", err)
	}
	if err := dbpkg.SetConfigValue(dbSuper, "wompi.enabled", "1", false); err != nil {
		t.Fatalf("seed wompi.enabled: %v", err)
	}
	if err := dbpkg.SetConfigValue(dbSuper, "licencias.discount_codes", "PROMO100=100%", false); err != nil {
		t.Fatalf("seed licencias.discount_codes: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/wompi/create_transaction_nequi", strings.NewReader(`{"licencia_id":1,"empresa_id":33,"phone_number":"3991111111","accept_terms":true,"discount_code":"PROMO100"}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	WompiCreateNequiTransactionHandler(dbSuper).ServeHTTP(rr, req)

	if rr.Code != http.StatusPreconditionFailed {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusPreconditionFailed, rr.Code, rr.Body.String())
	}
	if !strings.Contains(strings.ToLower(rr.Body.String()), "manual_activation") && !strings.Contains(strings.ToLower(rr.Body.String()), "activar") {
		t.Fatalf("expected zero-total response to suggest activation, got %s", rr.Body.String())
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
