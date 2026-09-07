package db

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/you/pos-backend/secure"
)

func TestEmpresaOpenAIKeyUsesDedicatedEncryptedPurpose(t *testing.T) {
	t.Setenv("CONFIG_ENC_KEY", base64.StdEncoding.EncodeToString(make([]byte, 32)))
	t.Setenv("CONFIG_ENC_KEY_ID", "test-key")
	plain := "key-that-must-never-reach-a-response"
	ciphertext, err := secure.EncryptStringForPurpose(empresaAIOpenAIEncryptionPurpose, plain)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(ciphertext, plain) || !strings.HasPrefix(ciphertext, "v1:"+empresaAIOpenAIEncryptionPurpose+":test-key:") {
		t.Fatalf("la clave empresarial no quedo en una envoltura cifrada dedicada: %q", ciphertext)
	}
	if _, err := secure.DecryptStringForPurpose("totp", ciphertext); err == nil {
		t.Fatal("un cifrado OpenAI empresarial no debe descifrarse como TOTP")
	}
}
