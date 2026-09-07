package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"strings"
	"testing"

	"github.com/you/pos-backend/utils"
)

func TestNextcloudAdminCredentialIsEncryptedBeforeStorage(t *testing.T) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CONFIG_ENC_KEY", base64.StdEncoding.EncodeToString(key))
	t.Setenv("CONFIG_ENC_KEY_ID", "nextcloud-test")
	plain := "temporary-nextcloud-credential"
	encrypted, err := encryptNextcloudAdminCredential(plain)
	if err != nil {
		t.Fatal(err)
	}
	if encrypted == plain || !strings.HasPrefix(encrypted, "v1:nextcloud-test:") {
		t.Fatalf("nextcloud credential was not stored in a versioned encrypted envelope")
	}
	decrypted, err := utils.DecryptString(encrypted)
	if err != nil || decrypted != plain {
		t.Fatalf("nextcloud credential cannot be decrypted safely: %v", err)
	}
}

func TestNextcloudAccessURLsRequireActiveProvisionedAccount(t *testing.T) {
	account := nextcloudCompanyAccount{User: "pcs_empresa_7", Active: true, Provisioned: true}
	webURL, webDAVURL := nextcloudAccessURLs(account, "https://nextcloud.example.test")
	if webURL == "" || webDAVURL == "" {
		t.Fatal("active provisioned account must receive its scoped access URLs")
	}
	account.Active = false
	webURL, webDAVURL = nextcloudAccessURLs(account, "https://nextcloud.example.test")
	if webURL != "" || webDAVURL != "" {
		t.Fatal("deactivated company account must not receive Nextcloud access URLs")
	}
}

func TestValidateNextcloudAccountUser(t *testing.T) {
	user, err := validateNextcloudAccountUser(" Cuenta.Super_01 ")
	if err != nil || user != "cuenta.super_01" {
		t.Fatalf("expected normalized personal account user, got %q (%v)", user, err)
	}
	for _, invalid := range []string{"ab", "cuenta con espacios", "cuenta/otra", "cuenta@correo"} {
		if _, err := validateNextcloudAccountUser(invalid); err == nil {
			t.Fatalf("expected invalid Nextcloud user to be rejected: %q", invalid)
		}
	}
}
