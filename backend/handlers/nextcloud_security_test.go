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
