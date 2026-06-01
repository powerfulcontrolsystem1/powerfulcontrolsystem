package handlers

import (
	"strings"
	"testing"
)

func TestValidateOfflineVentaSessionOwnerRejectsDifferentCashier(t *testing.T) {
	err := validateOfflineVentaSessionOwner(offlineVentaPayload{
		UsuarioEmail: "cajero.uno@example.com",
	}, "cajero.dos@example.com")
	if err == nil {
		t.Fatal("expected error for offline sale owned by another cashier")
	}
	if !strings.Contains(err.Error(), "cajero.uno@example.com") {
		t.Fatalf("expected owner in error, got %q", err.Error())
	}
}

func TestValidateOfflineVentaSessionOwnerAllowsSameCashier(t *testing.T) {
	err := validateOfflineVentaSessionOwner(offlineVentaPayload{
		UsuarioEmail: "Cajero.Uno@Example.com",
	}, "cajero.uno@example.com")
	if err != nil {
		t.Fatalf("expected same cashier to be accepted, got %v", err)
	}
}

func TestValidateOfflineVentaCajaRequiresExplicitCashRegister(t *testing.T) {
	if err := validateOfflineVentaCaja(offlinePagoPayload{}); err == nil {
		t.Fatal("expected error when offline sale has no cash register")
	}
	if err := validateOfflineVentaCaja(offlinePagoPayload{CajaCodigo: "CAJA-2"}); err != nil {
		t.Fatalf("expected caja_codigo to be accepted, got %v", err)
	}
	if err := validateOfflineVentaCaja(offlinePagoPayload{CierreCajaID: 12}); err == nil {
		t.Fatal("expected cierre_caja_id without caja_codigo to be rejected")
	}
}

func TestNormalizeOfflineSyncKeyKeepsMultiCashierReferenceSafe(t *testing.T) {
	got := normalizeOfflineSyncKey(" off-7-cajero@example.com-caja 2-est 1 ")
	if got != "OFF-7-CAJEROEXAMPLECOM-CAJA2-EST1" {
		t.Fatalf("unexpected sync key normalization: %q", got)
	}
}
