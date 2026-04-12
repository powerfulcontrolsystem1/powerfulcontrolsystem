package secure

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
)

// getEncKeyFromEnv intenta obtener la clave de cifrado desde la variable `CONFIG_ENC_KEY`.
// Se admite una cadena Base64 (preferida) o una cadena de al menos 32 bytes.
func getEncKeyFromEnv() ([]byte, error) {
	s := os.Getenv("CONFIG_ENC_KEY")
	if s == "" {
		return nil, fmt.Errorf("CONFIG_ENC_KEY not set")
	}
	if decoded, err := base64.StdEncoding.DecodeString(s); err == nil {
		if len(decoded) >= 16 {
			if len(decoded) >= 32 {
				return decoded[:32], nil
			}
			key := make([]byte, 32)
			copy(key, decoded)
			return key, nil
		}
	}
	if len(s) >= 32 {
		return []byte(s)[:32], nil
	}
	return nil, fmt.Errorf("CONFIG_ENC_KEY invalid length; provide base64 or >=32 bytes")
}

// EncryptString cifra texto plano usando AES-GCM y devuelve Base64(nonce|ciphertext)
func EncryptString(plain string) (string, error) {
	key, err := getEncKeyFromEnv()
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
	ct := gcm.Seal(nil, nonce, []byte(plain), nil)
	out := append(nonce, ct...)
	return base64.StdEncoding.EncodeToString(out), nil
}

// DecryptString descifra un payload Base64(nonce|ciphertext) usando AES-GCM
func DecryptString(payload string) (string, error) {
	key, err := getEncKeyFromEnv()
	if err != nil {
		return "", err
	}
	raw, err := base64.StdEncoding.DecodeString(payload)
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
	pt, err := gcm.Open(nil, nonce, ct, nil)
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
