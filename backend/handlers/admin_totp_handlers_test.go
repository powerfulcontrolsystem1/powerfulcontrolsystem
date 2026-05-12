package handlers

import (
	"strings"
	"testing"

	dbpkg "github.com/you/pos-backend/db"
)

func TestAdminTOTPLoginRequiredForAdminDependsOnGlobalSwitch(t *testing.T) {
	admin := &dbpkg.Admin{TOTPEnabled: 1, TOTPSecret: strings.Repeat("A", 16)}
	if adminTOTPLoginRequiredForAdmin(admin, false) {
		t.Fatal("2FA no debe exigirse cuando el switch global esta apagado")
	}
	if !adminTOTPLoginRequiredForAdmin(admin, true) {
		t.Fatal("2FA debe exigirse cuando el switch global esta encendido y la cuenta tiene TOTP")
	}
}

func TestAdminTOTPLoginRequiredForAdminRequiresConfiguredAccount(t *testing.T) {
	if adminTOTPLoginRequiredForAdmin(nil, true) {
		t.Fatal("nil admin no debe exigir 2FA")
	}
	if adminTOTPLoginRequiredForAdmin(&dbpkg.Admin{TOTPEnabled: 0, TOTPSecret: strings.Repeat("A", 16)}, true) {
		t.Fatal("TOTP desactivado en la cuenta no debe exigir 2FA")
	}
	if adminTOTPLoginRequiredForAdmin(&dbpkg.Admin{TOTPEnabled: 1, TOTPSecret: ""}, true) {
		t.Fatal("TOTP sin secreto confirmado no debe exigir 2FA")
	}
}
