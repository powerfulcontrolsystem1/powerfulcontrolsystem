package handlers

import (
	"crypto/hmac"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"

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

func TestNextcloudSSOTokenIsShortLivedAndBoundToEmpresa(t *testing.T) {
	secret := strings.Repeat("s", 48)
	t.Setenv("NEXTCLOUD_SSO_SECRET", secret)
	now := time.Date(2026, time.September, 13, 12, 0, 0, 0, time.UTC)
	account := nextcloudCompanyAccount{User: "pcs_empresa_12", Active: true, Provisioned: true}
	token, err := createNextcloudSSOToken(account, 12, now)
	if err != nil {
		t.Fatal(err)
	}
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		t.Fatalf("unexpected signed token format: %q", token)
	}
	expectedSignature := signNextcloudSSOPayload(parts[0], secret)
	if !hmac.Equal([]byte(expectedSignature), []byte(parts[1])) {
		t.Fatal("Nextcloud SSO token signature is invalid")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		t.Fatal(err)
	}
	var claims nextcloudSSOToken
	if err := json.Unmarshal(payload, &claims); err != nil {
		t.Fatal(err)
	}
	if claims.User != "pcs_empresa_12" || claims.EmpresaID != 12 || claims.Audience != nextcloudSSOTokenAudience {
		t.Fatalf("token lost its company boundary: %+v", claims)
	}
	if claims.ExpiresAt-claims.IssuedAt != int64(nextcloudSSOTokenTTL/time.Second) {
		t.Fatalf("unexpected token lifetime: %d", claims.ExpiresAt-claims.IssuedAt)
	}

	account.User = "pcs_empresa_52"
	if _, err := createNextcloudSSOToken(account, 12, now); err == nil {
		t.Fatal("cross-company Nextcloud user must be rejected")
	}
}

func TestNextcloudAutologinURLKeepsConfiguredBasePath(t *testing.T) {
	got := nextcloudAutologinURL("https://nextcloud.example.test/cloud", "signed-token")
	want := "https://nextcloud.example.test/cloud/index.php/apps/pcs_sso/login?token=signed-token"
	if got != want {
		t.Fatalf("autologin URL = %q, want %q", got, want)
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
