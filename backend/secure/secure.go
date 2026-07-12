package secure

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
)

const encryptionFormatVersion = "v1"
const defaultEncryptionKeyID = "active"

// TOTPEncryptionPurpose separates authenticator secrets from the generic
// configuration encryption domain. A derived key plus authenticated purpose
// prevents a ciphertext from one domain being accepted by another.
const TOTPEncryptionPurpose = "totp"

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

func activeEncryptionKeyID() (string, error) {
	keyID := strings.TrimSpace(os.Getenv("CONFIG_ENC_KEY_ID"))
	if keyID == "" {
		return defaultEncryptionKeyID, nil
	}
	for _, ch := range keyID {
		if !(ch >= 'a' && ch <= 'z') && !(ch >= 'A' && ch <= 'Z') && !(ch >= '0' && ch <= '9') && ch != '-' && ch != '_' {
			return "", fmt.Errorf("CONFIG_ENC_KEY_ID contains unsupported characters")
		}
	}
	return keyID, nil
}

func encryptionKeyForID(keyID string) ([]byte, error) {
	activeID, err := activeEncryptionKeyID()
	if err != nil {
		return nil, err
	}
	if keyID == "" || keyID == activeID {
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

func encryptionKeyForPurpose(keyID, purpose string) ([]byte, error) {
	master, err := encryptionKeyForID(keyID)
	if err != nil {
		return nil, err
	}
	purpose = strings.TrimSpace(purpose)
	if purpose == "" {
		return nil, fmt.Errorf("encryption purpose required")
	}
	mac := hmac.New(sha256.New, master)
	_, _ = mac.Write([]byte("pcs:encryption-purpose:" + purpose))
	return mac.Sum(nil), nil
}

// EncryptStringForPurpose encrypts a value in an isolated cryptographic
// domain. New values include format version, purpose and key identifier so a
// previous master key remains usable only for decryption during rotation.
func EncryptStringForPurpose(purpose, plain string) (string, error) {
	purpose = strings.TrimSpace(purpose)
	keyID, err := activeEncryptionKeyID()
	if err != nil {
		return "", err
	}
	key, err := encryptionKeyForPurpose(keyID, purpose)
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
	aad := []byte(encryptionFormatVersion + ":" + purpose + ":" + keyID)
	ciphertext := gcm.Seal(nil, nonce, []byte(plain), aad)
	return encryptionFormatVersion + ":" + purpose + ":" + keyID + ":" + base64.StdEncoding.EncodeToString(append(nonce, ciphertext...)), nil
}

// DecryptStringForPurpose only accepts the envelope issued for purpose. This
// intentionally does not fall back to legacy plaintext; callers must migrate
// legacy values explicitly so accidental plaintext use is never silent.
func DecryptStringForPurpose(purpose, payload string) (string, error) {
	purpose = strings.TrimSpace(purpose)
	parts := strings.SplitN(strings.TrimSpace(payload), ":", 4)
	if len(parts) != 4 || parts[0] != encryptionFormatVersion || parts[1] != purpose || strings.TrimSpace(parts[2]) == "" {
		return "", fmt.Errorf("unsupported encrypted payload format for purpose")
	}
	keyID := strings.TrimSpace(parts[2])
	key, err := encryptionKeyForPurpose(keyID, purpose)
	if err != nil {
		return "", err
	}
	raw, err := base64.StdEncoding.DecodeString(parts[3])
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
	if len(raw) < gcm.NonceSize() {
		return "", fmt.Errorf("invalid payload")
	}
	plaintext, err := gcm.Open(nil, raw[:gcm.NonceSize()], raw[gcm.NonceSize():], []byte(encryptionFormatVersion+":"+purpose+":"+keyID))
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

// EncryptString encrypts new data in a versioned envelope. Legacy payloads are
// still accepted by DecryptString during a controlled key-format transition.
func EncryptString(plain string) (string, error) {
	keyID, err := activeEncryptionKeyID()
	if err != nil {
		return "", err
	}
	key, err := encryptionKeyForID(keyID)
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
	ct := gcm.Seal(nil, nonce, []byte(plain), []byte(encryptionFormatVersion+":"+keyID))
	out := append(nonce, ct...)
	return encryptionFormatVersion + ":" + keyID + ":" + base64.StdEncoding.EncodeToString(out), nil
}

// DecryptString descifra un payload Base64(nonce|ciphertext) usando AES-GCM
func DecryptString(payload string) (string, error) {
	keyID, err := activeEncryptionKeyID()
	if err != nil {
		return "", err
	}
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
