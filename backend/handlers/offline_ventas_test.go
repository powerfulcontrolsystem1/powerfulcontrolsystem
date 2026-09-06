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

func TestOfflineIdentityFailsClosedBeforeClaim(t *testing.T) {
	valid := offlineVentaPayload{EmpresaID: 12, SyncKey: "OFF-CAJA2-UNIQUE", UsuarioEmail: "qa@example.invalid", Pago: offlinePagoPayload{CajaCodigo: "CAJA-2"}}
	if err := validateOfflineVentaIdentity(12, valid, "qa@example.invalid"); err != nil {
		t.Fatal(err)
	}
	for _, change := range []func(*offlineVentaPayload){
		func(v *offlineVentaPayload) { v.EmpresaID = 13 },
		func(v *offlineVentaPayload) { v.SyncKey = "" },
		func(v *offlineVentaPayload) { v.UsuarioEmail = "" },
		func(v *offlineVentaPayload) { v.UsuarioEmail = "otro@example.invalid" },
		func(v *offlineVentaPayload) { v.Pago.CajaCodigo = "" },
	} {
		v := valid
		change(&v)
		if err := validateOfflineVentaIdentity(12, v, "qa@example.invalid"); err == nil {
			t.Fatal("invalid offline identity accepted")
		}
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
