package handlers

import (
	"encoding/base64"
	"strings"
	"testing"
	"time"

	dbpkg "github.com/you/pos-backend/db"
	"github.com/you/pos-backend/secure"
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

func TestAdminTOTPSecretIsEncryptedBeforeVerification(t *testing.T) {
	t.Setenv("CONFIG_ENC_KEY", base64.StdEncoding.EncodeToString(make([]byte, 32)))
	secret := strings.Repeat("A", 32)
	payload, err := secure.EncryptStringForPurpose(secure.TOTPEncryptionPurpose, secret)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(payload, secret) {
		t.Fatal("encrypted TOTP payload contains plaintext secret")
	}
	plain, err := adminTOTPSecretForVerification(&dbpkg.Admin{TOTPSecret: payload})
	if err != nil || plain != secret {
		t.Fatalf("encrypted TOTP secret cannot be recovered for verification: %q %v", plain, err)
	}
}

func TestAdminTOTPCodesValidateAndRejectIncorrectCode(t *testing.T) {
	secret := strings.Repeat("A", 32)
	now := time.Unix(1_700_000_000, 0)
	code, err := totpCodeAt(secret, now.Unix()/30)
	if err != nil {
		t.Fatal(err)
	}
	if !verifyAdminTOTPCode(secret, code, now) {
		t.Fatal("valid TOTP code rejected")
	}
	if verifyAdminTOTPCode(secret, "000000", now) && code != "000000" {
		t.Fatal("incorrect TOTP code accepted")
	}
}

func TestAdminTOTPRecoveryCodesAreHighEntropyAndDistinct(t *testing.T) {
	codes, batchID, err := generateAdminTOTPRecoveryCodes(3)
	if err != nil {
		t.Fatal(err)
	}
	if len(batchID) < 20 || len(codes) != 3 {
		t.Fatal("recovery code batch is incomplete")
	}
	seen := map[string]bool{}
	for _, code := range codes {
		if len(code) < 40 || seen[code] {
			t.Fatal("recovery code is weak or duplicated")
		}
		seen[code] = true
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
