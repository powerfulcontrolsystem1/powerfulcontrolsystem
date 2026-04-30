package handlers

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"
)

func TestPickEpaycoFieldReadsNestedAliases(t *testing.T) {
	payload := map[string]interface{}{
		"data": map[string]interface{}{
			"auth": map[string]interface{}{
				"bearerToken": " token-123 ",
			},
		},
	}

	got := pickEpaycoField(payload, "token", "access_token", "bearer_token")
	if got != "token-123" {
		t.Fatalf("expected nested bearer token, got %q", got)
	}
}

func TestPickEpaycoFieldIgnoresNonScalarMatches(t *testing.T) {
	payload := map[string]interface{}{
		"data": map[string]interface{}{
			"session": map[string]interface{}{
				"id": map[string]interface{}{
					"value": "not-a-direct-session-id",
				},
			},
			"session_id": "session-456",
		},
	}

	got := pickEpaycoField(payload, "sessionId", "session_id", "id")
	if got != "session-456" {
		t.Fatalf("expected direct nested session id, got %q", got)
	}
}

func TestBuildEpaycoClassicCheckoutFormUsesSecurePaycoPostAndSignature(t *testing.T) {
	checkoutKey := "a1c7200f0e2029d11b62bfd863422d5db10a8397"
	form := buildEpaycoClassicCheckoutForm(
		"https://powerfulcontrolsystem.com",
		"123456",
		checkoutKey,
		"EPAYCO-LIC-1",
		"Premium",
		11,
		22,
		12345.674,
		"cliente@example.com",
		"sandbox",
	)

	if form.Method != "POST" {
		t.Fatalf("expected POST method, got %q", form.Method)
	}
	if form.Action != "https://secure.payco.co/checkout.php" {
		t.Fatalf("expected secure.payco.co classic checkout action, got %q", form.Action)
	}
	if form.Action == "https://checkout.epayco.co/checkout.php" {
		t.Fatal("classic checkout must not use checkout.epayco.co/checkout.php")
	}

	fields := form.Fields
	if fields["p_cust_id_cliente"] != "123456" {
		t.Fatalf("expected customer id field, got %q", fields["p_cust_id_cliente"])
	}
	if fields["p_key"] != checkoutKey {
		t.Fatalf("expected private key field, got %q", fields["p_key"])
	}
	if fields["p_id_invoice"] != "EPAYCO-LIC-1" {
		t.Fatalf("expected invoice reference, got %q", fields["p_id_invoice"])
	}
	if fields["p_amount"] != "12345.67" {
		t.Fatalf("expected rounded amount, got %q", fields["p_amount"])
	}
	if fields["p_currency_code"] != "COP" {
		t.Fatalf("expected COP currency, got %q", fields["p_currency_code"])
	}
	if fields["p_test_request"] != "TRUE" {
		t.Fatalf("expected sandbox test request, got %q", fields["p_test_request"])
	}
	if fields["p_extra1"] != "11" || fields["p_extra2"] != "22" || fields["p_extra3"] != "EPAYCO-LIC-1" {
		t.Fatalf("expected extra fields with licencia/empresa/reference, got %#v", fields)
	}
	if fields["p_url_response"] == "" || fields["p_url_confirmation"] != "https://powerfulcontrolsystem.com/epayco/webhook" {
		t.Fatalf("expected response and confirmation URLs, got response=%q confirmation=%q", fields["p_url_response"], fields["p_url_confirmation"])
	}

	expectedSignature := fmt.Sprintf("%x", md5.Sum([]byte("123456^"+checkoutKey+"^EPAYCO-LIC-1^12345.67^COP")))
	if fields["p_signature"] != expectedSignature {
		t.Fatalf("expected signature %q, got %q", expectedSignature, fields["p_signature"])
	}
}

func TestBuildEpaycoClassicCheckoutFormUsesProductionFlag(t *testing.T) {
	form := buildEpaycoClassicCheckoutForm(
		"https://powerfulcontrolsystem.com",
		"123456",
		"a1c7200f0e2029d11b62bfd863422d5db10a8397",
		"EPAYCO-LIC-2",
		"Premium",
		11,
		22,
		90000,
		"",
		"production",
	)

	if form.Fields["p_test_request"] != "FALSE" {
		t.Fatalf("expected production checkout to disable test request, got %q", form.Fields["p_test_request"])
	}
}

func TestBuildEpaycoClassicCheckoutPayloadUsesOfficialCheckoutJSShape(t *testing.T) {
	payload := buildEpaycoClassicCheckoutPayload(
		"https://powerfulcontrolsystem.com",
		"public-key-123",
		"EPAYCO-LIC-9",
		"Plan Empresa",
		9,
		15,
		99000,
		"cliente@example.com",
		"production",
	)

	if payload.ScriptURL != "https://checkout.epayco.co/checkout.js" {
		t.Fatalf("expected official checkout.js URL, got %q", payload.ScriptURL)
	}
	if payload.Config["key"] != "public-key-123" {
		t.Fatalf("expected public key in checkout config, got %#v", payload.Config["key"])
	}
	if payload.Config["test"] != false {
		t.Fatalf("expected production test=false, got %#v", payload.Config["test"])
	}
	data := payload.Data
	if data["external"] != "true" || data["currency"] != "cop" || data["country"] != "co" {
		t.Fatalf("expected standard external checkout data, got %#v", data)
	}
	if data["invoice"] != "EPAYCO-LIC-9" || data["amount"] != "99000.00" {
		t.Fatalf("expected invoice and amount, got invoice=%#v amount=%#v", data["invoice"], data["amount"])
	}
	if data["confirmation"] != "https://powerfulcontrolsystem.com/epayco/webhook" {
		t.Fatalf("expected webhook confirmation URL, got %#v", data["confirmation"])
	}
	if _, ok := data["p_key"]; ok {
		t.Fatal("classic checkout.js payload must not expose P_KEY to the browser")
	}
	if data["extra1"] != "9" || data["extra2"] != "15" || data["extra3"] != "EPAYCO-LIC-9" {
		t.Fatalf("expected extras with licencia/empresa/reference, got %#v", data)
	}
	if data["email_billing"] != "cliente@example.com" {
		t.Fatalf("expected billing email, got %#v", data["email_billing"])
	}
}

func TestBuildEpaycoSmartCheckoutSessionPayloadMatchesOfficialV2Shape(t *testing.T) {
	payload := buildEpaycoSmartCheckoutSessionPayload(
		"https://powerfulcontrolsystem.com",
		"EPAYCO-LIC-7",
		"Plan Empresarial",
		7,
		8,
		50000.456,
		"cliente@example.com",
	)

	if payload["checkout_version"] != "2" {
		t.Fatalf("expected Smart Checkout v2, got %#v", payload["checkout_version"])
	}
	if payload["currency"] != "COP" || payload["country"] != "CO" || payload["method"] != "POST" {
		t.Fatalf("expected COP/CO/POST payload, got %#v", payload)
	}
	if payload["amount"] != 50000.46 {
		t.Fatalf("expected rounded numeric amount, got %#v", payload["amount"])
	}
	if payload["invoice"] != "EPAYCO-LIC-7" {
		t.Fatalf("expected invoice reference, got %#v", payload["invoice"])
	}
	if payload["response"] == "" || payload["confirmation"] != "https://powerfulcontrolsystem.com/epayco/webhook" {
		t.Fatalf("expected public response and confirmation URLs, got response=%#v confirmation=%#v", payload["response"], payload["confirmation"])
	}
	if payload["forceResponse"] != true || payload["uniqueTransactionPerBill"] != true {
		t.Fatalf("expected deterministic response and unique invoice protection, got %#v", payload)
	}
	extras, ok := payload["extras"].(map[string]interface{})
	if !ok || extras["extra1"] != "7" || extras["extra2"] != "8" || extras["extra3"] != "EPAYCO-LIC-7" {
		t.Fatalf("expected extras with licencia/empresa/reference, got %#v", payload["extras"])
	}
	billing, ok := payload["billing"].(map[string]interface{})
	if !ok || billing["email"] != "cliente@example.com" {
		t.Fatalf("expected billing email, got %#v", payload["billing"])
	}
}

func TestVerifyEpaycoConfirmationSignatureUsesOfficialSha256Formula(t *testing.T) {
	payload := map[string]interface{}{
		"x_ref_payco":      "123456789",
		"x_transaction_id": "TX-123",
		"x_amount":         "50000.00",
		"x_currency_code":  "COP",
		"x_cod_response":   "1",
		"x_response":       "Aceptada",
	}
	source := "9695^p-key-secret^123456789^TX-123^50000.00^COP"
	sum := sha256.Sum256([]byte(source))
	payload["x_signature"] = hex.EncodeToString(sum[:])

	valid, provided, _, expected := verifyEpaycoConfirmationSignature("9695", "p-key-secret", payload)
	if !provided {
		t.Fatal("expected signature to be detected")
	}
	if !valid {
		t.Fatalf("expected valid signature, expected hash %q", expected)
	}

	payload["x_amount"] = "50001.00"
	valid, provided, _, _ = verifyEpaycoConfirmationSignature("9695", "p-key-secret", payload)
	if !provided || valid {
		t.Fatal("expected tampered amount to invalidate signature")
	}
}

func TestParseEpaycoPaymentStatusReadsTransactionStateCode(t *testing.T) {
	payload := map[string]interface{}{
		"x_cod_transaction_state": "1",
	}
	if got := parseEpaycoPaymentStatus(payload); got != "APPROVED" {
		t.Fatalf("expected APPROVED from x_cod_transaction_state, got %q", got)
	}
}

func TestEpaycoApprovedStatusAliasesActivateLicenses(t *testing.T) {
	for _, status := range []string{"APPROVED", "accepted", "Aceptada", "ACEPTADO", "Aprobada", "acreditado", "1", "success"} {
		if !isApprovedPaymentStatus(status) {
			t.Fatalf("expected %q to be treated as approved", status)
		}
	}
	for _, status := range []string{"PENDING", "DECLINED", "Rechazada", "ERROR", ""} {
		if isApprovedPaymentStatus(status) {
			t.Fatalf("did not expect %q to be treated as approved", status)
		}
	}
}

func TestPaymentContextFromEpaycoPayloadReadsExtrasAndInvoice(t *testing.T) {
	payload := map[string]interface{}{
		"x_extra1": "12",
		"x_extra2": "34",
	}
	licenciaID, empresaID, ok := paymentContextFromEpaycoPayload(payload)
	if !ok || licenciaID != 12 || empresaID != 34 {
		t.Fatalf("expected context from extras, got ok=%v licencia=%d empresa=%d", ok, licenciaID, empresaID)
	}

	payload = map[string]interface{}{
		"x_id_invoice": "EPAYCO-LIC-56-EMP-78-123456",
	}
	licenciaID, empresaID, ok = paymentContextFromEpaycoPayload(payload)
	if !ok || licenciaID != 56 || empresaID != 78 {
		t.Fatalf("expected context from invoice, got ok=%v licencia=%d empresa=%d", ok, licenciaID, empresaID)
	}
}

func TestStrongEpaycoApprovedReturnEvidenceRequiresGatewayAndInvoice(t *testing.T) {
	payload := map[string]interface{}{
		"x_transaction_id": "TX-123",
		"x_ref_payco":      "987654321",
		"x_id_invoice":     "EPAYCO-LIC-56-EMP-78-123456",
		"x_cod_response":   "1",
		"x_response":       "Aceptada",
	}
	if !hasStrongEpaycoApprovedReturnEvidence(payload, false) {
		t.Fatal("expected approved Epayco return with transaction, x_ref_payco and invoice to be strong evidence")
	}

	delete(payload, "x_ref_payco")
	if hasStrongEpaycoApprovedReturnEvidence(payload, false) {
		t.Fatal("expected unsigned return without x_ref_payco to be rejected as weak evidence")
	}

	if !hasStrongEpaycoApprovedReturnEvidence(payload, true) {
		t.Fatal("expected valid signed approved return to be accepted")
	}
}

func TestResolveEpaycoClassicModeUsesClassicCredentials(t *testing.T) {
	mode, source := resolveEpaycoClassicMode(nil, "123456", "a1c7200f0e2029d11b62bfd863422d5db10a8397")
	if mode != "production" || source != "classic_credentials" {
		t.Fatalf("expected production mode from classic credentials, got mode=%q source=%q", mode, source)
	}
}

func TestEpaycoCheckoutCredentialReadinessSeparatesSmartAndClassicKeys(t *testing.T) {
	if !epaycoSmartCheckoutReady("pub_prod_123", "prv_prod_456") {
		t.Fatal("expected prefixed API credentials to enable Smart Checkout")
	}
	if epaycoSmartCheckoutReady("pub_prod_123", "checkout-secret-key") {
		t.Fatal("checkout P_KEY must not be treated as Smart Checkout private_key API")
	}
	if !epaycoClassicCheckoutReady("9695", "a1c7200f0e2029d11b62bfd863422d5db10a8397") {
		t.Fatal("expected customer id + P_KEY to enable classic checkout")
	}
	if !epaycoCustomCheckoutReady("491d6a0b6e992cf924edd8d3d088aff1", "9695", "a1c7200f0e2029d11b62bfd863422d5db10a8397") {
		t.Fatal("expected public key + customer id + P_KEY to enable checkout.js fallback")
	}
	if epaycoCustomCheckoutReady("", "9695", "a1c7200f0e2029d11b62bfd863422d5db10a8397") {
		t.Fatal("checkout.js fallback must require the public key")
	}
	if epaycoClassicCheckoutReady("9695", "clave-corta#") {
		t.Fatal("short password-like values must not enable classic checkout")
	}
	if epaycoClassicCheckoutReady("pub_prod_123", "prv_prod_456") {
		t.Fatal("API public/private keys must not be treated as classic checkout credentials")
	}
}

func TestSanitizeEpaycoClassicCheckoutFormMasksPrivateKey(t *testing.T) {
	form := epaycoClassicCheckoutForm{
		Method: "POST",
		Action: "https://secure.payco.co/checkout.php",
		Fields: map[string]string{
			"p_key":        "private-key",
			"p_id_invoice": "EPAYCO-LIC-1",
		},
	}

	sanitized := sanitizeEpaycoClassicCheckoutForm(form)
	if sanitized.Fields["p_key"] != "********" {
		t.Fatalf("expected masked private key, got %q", sanitized.Fields["p_key"])
	}
	if sanitized.Fields["p_id_invoice"] != "EPAYCO-LIC-1" {
		t.Fatalf("expected invoice to be preserved, got %q", sanitized.Fields["p_id_invoice"])
	}
	if form.Fields["p_key"] != "private-key" {
		t.Fatalf("sanitize should not mutate original form, got %q", form.Fields["p_key"])
	}
}
