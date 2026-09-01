package db

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"

	"github.com/you/pos-backend/secure"
)

func TestControlElectricoSSHCredentialPurposeIsTenantAndDeviceScoped(t *testing.T) {
	t.Setenv("CONFIG_ENC_KEY", base64.StdEncoding.EncodeToString([]byte("0123456789abcdef0123456789abcdef")))
	t.Setenv("CONFIG_ENC_KEY_ID", "test")
	purpose := controlElectricoSSHPurpose(12, 8, "password")
	ciphertext, err := secure.EncryptStringForPurpose(purpose, " secret with spaces ")
	if err != nil {
		t.Fatal(err)
	}
	plain, err := secure.DecryptStringForPurpose(purpose, ciphertext)
	if err != nil || plain != " secret with spaces " {
		t.Fatalf("roundtrip cifrado invalido: plain=%q err=%v", plain, err)
	}
	if _, err := secure.DecryptStringForPurpose(controlElectricoSSHPurpose(13, 8, "password"), ciphertext); err == nil {
		t.Fatal("otra empresa no debe poder descifrar la credencial")
	}
	if _, err := secure.DecryptStringForPurpose(controlElectricoSSHPurpose(12, 9, "password"), ciphertext); err == nil {
		t.Fatal("otra Raspberry no debe poder descifrar la credencial")
	}
}

func TestControlElectricoSSHProfileNeverSerializesSecrets(t *testing.T) {
	raw, err := json.Marshal(EmpresaControlElectricoSSHProfile{RaspberryID: 1, Host: "10.0.0.2", Username: "pi", CredentialsConfigured: true})
	if err != nil {
		t.Fatal(err)
	}
	serialized := strings.ToLower(string(raw))
	for _, forbidden := range []string{"password", "sudo_password", "ciphertext", "_enc"} {
		if strings.Contains(serialized, forbidden) {
			t.Fatalf("el perfil visible expone marcador secreto %q: %s", forbidden, serialized)
		}
	}
}
