package db

import (
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func setDIANConfigTestEncryptionKey(t *testing.T) {
	t.Helper()
	t.Setenv("CONFIG_ENC_KEY", base64.StdEncoding.EncodeToString([]byte("0123456789abcdef0123456789abcdef")))
	t.Setenv("CONFIG_ENC_KEY_ID", "dian-test")
	t.Setenv("CONFIG_ENC_KEY_PREVIOUS", "")
}

func TestEmpresaDIANConfigSecretEncryptionIsPurposeBound(t *testing.T) {
	setDIANConfigTestEncryptionKey(t)

	const (
		empresaID = int64(12)
		field     = "software_pin"
		plain     = "pin-confidencial-empresa-12"
	)
	encrypted, err := EncryptEmpresaDIANConfigSecret(empresaID, field, plain)
	if err != nil {
		t.Fatal(err)
	}
	if encrypted == plain || strings.Contains(encrypted, plain) || !strings.HasPrefix(encrypted, "v1:dian-config-empresa-12-campo-software_pin:dian-test:") {
		t.Fatalf("secreto no quedo protegido por envelope: %q", encrypted)
	}
	decrypted, err := DecryptEmpresaDIANConfigSecret(empresaID, field, encrypted)
	if err != nil || decrypted != plain {
		t.Fatalf("descifrado valido=%q err=%v", decrypted, err)
	}
	if _, err := DecryptEmpresaDIANConfigSecret(13, field, encrypted); err == nil {
		t.Fatal("un secreto cifrado para empresa 12 fue aceptado por empresa 13")
	}
	if _, err := DecryptEmpresaDIANConfigSecret(empresaID, "software_id", encrypted); err == nil {
		t.Fatal("un secreto cifrado para software_pin fue aceptado como software_id")
	}
	if _, err := EncryptEmpresaDIANConfigSecret(empresaID, "campo_no_permitido", plain); err == nil {
		t.Fatal("se acepto cifrar un campo fuera del catalogo DIAN")
	}
	if _, err := DecryptEmpresaDIANConfigSecret(empresaID, field, plain); err == nil {
		t.Fatal("se acepto como cifrado un secreto heredado en texto plano")
	}
}

func TestEmpresaDIANConfigSecretsMigrationCataloguedWithoutDestructiveDraftRewrite(t *testing.T) {
	migrations, err := PlatformMigrations(MigrationTargetEmpresas)
	if err != nil {
		t.Fatal(err)
	}
	const (
		encryptedVersion   = "20260823-002-dian-config-secrets-encrypted-v1"
		destructiveVersion = "20260823-002-dian-manual-document-drafts-v1"
	)
	foundEncrypted := false
	for _, migration := range migrations {
		if migration.Version == destructiveVersion {
			t.Fatalf("la migracion destructiva retirada sigue catalogada: %s", migration.Version)
		}
		if migration.Version != encryptedVersion {
			continue
		}
		foundEncrypted = true
		if migration.Body != empresaDIANConfigSecretsFingerprint || migration.Apply == nil {
			t.Fatalf("migracion de secretos DIAN mutable o no ejecutable: %#v", migration)
		}
	}
	if !foundEncrypted {
		t.Fatalf("no se encontro la migracion %s", encryptedVersion)
	}
}

func TestEmpresaDIANConfigSecretsMigrationEncryptsPlaintextPostgres(t *testing.T) {
	dsn := strings.TrimSpace(os.Getenv("PCS_TEST_POSTGRES_DSN"))
	if dsn == "" {
		t.Skip("PCS_TEST_POSTGRES_DSN is not configured")
	}
	setDIANConfigTestEncryptionKey(t)

	dbConn, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer dbConn.Close()
	tx, err := dbConn.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()

	schema := fmt.Sprintf("dian_config_secrets_test_%d", time.Now().UnixNano())
	if _, err := tx.Exec(`CREATE SCHEMA ` + schema); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(`SET LOCAL search_path TO ` + schema); err != nil {
		t.Fatal(err)
	}
	columns := []string{"id BIGSERIAL PRIMARY KEY", "empresa_id BIGINT NOT NULL"}
	for _, field := range EmpresaDIANConfigSecretColumns() {
		columns = append(columns, field+" TEXT")
	}
	if _, err := tx.Exec(`CREATE TABLE empresa_dian_configuracion (` + strings.Join(columns, ", ") + `)`); err != nil {
		t.Fatal(err)
	}
	const (
		empresaID = int64(12)
		plainID   = "software-id-plano"
		plainPIN  = "software-pin-plano"
	)
	if _, err := tx.Exec(`INSERT INTO empresa_dian_configuracion (empresa_id, software_id, software_pin) VALUES ($1, $2, $3)`, empresaID, plainID, plainPIN); err != nil {
		t.Fatal(err)
	}

	if err := applyEmpresaDIANConfigSecretsTx(context.Background(), tx); err != nil {
		t.Fatal(err)
	}
	var encryptedID, encryptedPIN string
	if err := tx.QueryRow(`SELECT software_id, software_pin FROM empresa_dian_configuracion WHERE empresa_id = $1`, empresaID).Scan(&encryptedID, &encryptedPIN); err != nil {
		t.Fatal(err)
	}
	for field, pair := range map[string][2]string{
		"software_id":  {encryptedID, plainID},
		"software_pin": {encryptedPIN, plainPIN},
	} {
		if pair[0] == pair[1] || strings.Contains(pair[0], pair[1]) {
			t.Fatalf("%s persistio texto plano: %q", field, pair[0])
		}
		plain, err := DecryptEmpresaDIANConfigSecret(empresaID, field, pair[0])
		if err != nil || plain != pair[1] {
			t.Fatalf("%s no descifra al valor original: plain=%q err=%v", field, plain, err)
		}
	}

	if err := applyEmpresaDIANConfigSecretsTx(context.Background(), tx); err != nil {
		t.Fatalf("la migracion no fue idempotente: %v", err)
	}
	var encryptedIDAgain, encryptedPINAgain string
	if err := tx.QueryRow(`SELECT software_id, software_pin FROM empresa_dian_configuracion WHERE empresa_id = $1`, empresaID).Scan(&encryptedIDAgain, &encryptedPINAgain); err != nil {
		t.Fatal(err)
	}
	if encryptedIDAgain != encryptedID || encryptedPINAgain != encryptedPIN {
		t.Fatal("la segunda ejecucion volvio a cifrar envelopes ya protegidos")
	}
}
