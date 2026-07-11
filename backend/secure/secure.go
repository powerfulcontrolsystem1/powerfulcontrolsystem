package secure

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
)

const encryptionFormatVersion = "v1"
const activeEncryptionKeyID = "active"

// getEncKeyFromEnv requires exactly 32 random bytes in canonical Base64. Older
// permissive behavior (padding/truncation) could silently weaken encryption.
func getEncKeyFromEnv() ([]byte, error) {
	s := strings.TrimSpace(os.Getenv("CONFIG_ENC_KEY"))
	if s == "" {
		return nil, fmt.Errorf("CONFIG_ENC_KEY not set")
	}
	decoded, err := base64.StdEncoding.DecodeString(s)
	if err != nil || len(decoded) != 32 || base64.StdEncoding.EncodeToString(decoded) != s {
		return nil, fmt.Errorf("CONFIG_ENC_KEY must be canonical base64 for exactly 32 bytes")
	}
	return decoded, nil
}

func encryptionKeyForID(keyID string) ([]byte, error) {
	if keyID == "" || keyID == activeEncryptionKeyID {
		return getEncKeyFromEnv()
	}
	for _, entry := range strings.Split(os.Getenv("CONFIG_ENC_KEY_PREVIOUS"), ",") {
		parts := strings.SplitN(strings.TrimSpace(entry), ":", 2)
		if len(parts) != 2 || strings.TrimSpace(parts[0]) != keyID {
			continue
		}
		decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(parts[1]))
		if err == nil && len(decoded) == 32 && base64.StdEncoding.EncodeToString(decoded) == strings.TrimSpace(parts[1]) {
			return decoded, nil
		}
		return nil, fmt.Errorf("invalid previous encryption key %q", keyID)
	}
	return nil, fmt.Errorf("encryption key id %q unavailable", keyID)
}

// EncryptString encrypts new data in a versioned envelope. Legacy payloads are
// still accepted by DecryptString during a controlled key-format transition.
func EncryptString(plain string) (string, error) {
	key, err := encryptionKeyForID(activeEncryptionKeyID)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	ct := gcm.Seal(nil, nonce, []byte(plain), []byte(encryptionFormatVersion+":"+activeEncryptionKeyID))
	out := append(nonce, ct...)
	return encryptionFormatVersion + ":" + activeEncryptionKeyID + ":" + base64.StdEncoding.EncodeToString(out), nil
}

// DecryptString descifra un payload Base64(nonce|ciphertext) usando AES-GCM
func DecryptString(payload string) (string, error) {
	keyID := activeEncryptionKeyID
	aad := []byte(nil)
	encoded := strings.TrimSpace(payload)
	parts := strings.SplitN(encoded, ":", 3)
	if len(parts) == 3 {
		if parts[0] != encryptionFormatVersion || strings.TrimSpace(parts[1]) == "" {
			return "", fmt.Errorf("unsupported encrypted payload format")
		}
		keyID = strings.TrimSpace(parts[1])
		encoded = parts[2]
		aad = []byte(encryptionFormatVersion + ":" + keyID)
	}
	key, err := encryptionKeyForID(keyID)
	if err != nil {
		return "", err
	}
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	ns := gcm.NonceSize()
	if len(raw) < ns {
		return "", fmt.Errorf("invalid payload")
	}
	nonce := raw[:ns]
	ct := raw[ns:]
	pt, err := gcm.Open(nil, nonce, ct, aad)
	if err != nil {
		return "", err
	}
	return string(pt), nil
}

// EncryptionAvailable devuelve true si la variable de entorno CONFIG_ENC_KEY está disponible y válida.
func EncryptionAvailable() bool {
	_, err := getEncKeyFromEnv()
	return err == nil
}
