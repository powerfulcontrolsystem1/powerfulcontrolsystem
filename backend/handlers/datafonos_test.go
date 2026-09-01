package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	dbpkg "github.com/you/pos-backend/db"
)

func TestNormalizeDatafonoProviderHTTPResponse(t *testing.T) {
	resp := normalizeDatafonoProviderHTTPResponse("bold", map[string]interface{}{
		"id":                 "pay_123",
		"status":             "APPROVED",
		"authorization_code": "AUT-9",
		"reference":          "VENTA-9",
		"amount":             float64(12500),
		"card_number":        "4111111111111111",
		"payment_method":     map[string]interface{}{"card": map[string]interface{}{"pan": "4111111111111111", "cvv": "123"}},
	})
	if resp.EstadoPago != dbpkg.DatafonoEstadoAprobado {
		t.Fatalf("estado = %q", resp.EstadoPago)
	}
	if resp.ProviderTransactionID != "pay_123" || resp.CodigoAutorizacion != "AUT-9" {
		t.Fatalf("response identifiers not normalized: %+v", resp)
	}
	if got := resp.Raw["card_number"]; got != "[redacted]" {
		t.Fatalf("card_number not redacted: %v", got)
	}
	nested := resp.Raw["payment_method"].(map[string]interface{})
	if nested["card"] != "[redacted]" {
		t.Fatalf("nested card block not redacted: %#v", nested)
	}
}

func TestHTTPDatafonoProviderClientInitiatePayment(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/payments" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer secret-test" {
			t.Fatalf("authorization header = %q", got)
		}
		var payload map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		if payload["reference"] != "VENTA-1" {
			t.Fatalf("reference = %v", payload["reference"])
		}
		encodeJSONResponse(w, map[string]interface{}{
			"transaction_id":     "tx-1",
			"status":             "approved",
			"authorization_code": "AUT-1",
			"reference":          "VENTA-1",
			"amount":             10000,
			"currency":           "COP",
		})
	}))
	defer server.Close()

	t.Setenv("DATAFONO_TEST_KEY", "secret-test")
	t.Setenv("PCS_DATAFONO_TEST_ALLOW_HTTP", "1")
	client := &httpDatafonoProviderClient{httpClient: server.Client()}
	resp, err := client.InitiatePayment(context.Background(), dbpkg.EmpresaDatafonoConfig{
		Proveedor:     dbpkg.DatafonoProviderRedeban,
		ApiBaseURL:    server.URL,
		CrearPagoPath: "/payments",
		AuthMode:      "bearer",
		ApiKeyRef:     "env:DATAFONO_TEST_KEY",
	}, dbpkg.EmpresaDatafonoPaymentRequest{
		Monto:      10000,
		Moneda:     "COP",
		Referencia: "VENTA-1",
	})
	if err != nil {
		t.Fatalf("InitiatePayment returned error: %v", err)
	}
	if resp.EstadoPago != dbpkg.DatafonoEstadoAprobado || resp.ProviderTransactionID != "tx-1" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestJoinProviderURLRejectsPrivateAndCrossOriginTargets(t *testing.T) {
	for _, raw := range []string{"http://127.0.0.1:8080", "https://localhost", "https://169.254.169.254"} {
		if _, err := joinProviderURL(raw, "/payments"); err == nil {
			t.Fatalf("expected private endpoint %q to be rejected", raw)
		}
	}
	if _, err := joinProviderURL("https://payments.example.com", "https://attacker.example/pay"); err == nil {
		t.Fatal("expected absolute provider path to be rejected")
	}
}
