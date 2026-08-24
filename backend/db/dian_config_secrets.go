package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/you/pos-backend/secure"
)

// #nosec G101 -- immutable migration fingerprint; this value is not a credential or encryption key.
const empresaDIANConfigSecretsFingerprint = "empresa-dian-config-secrets:v1:purpose-bound-encryption"

var empresaDIANConfigSecretColumns = []string{
	"software_id",
	"software_pin",
	"software_id_compartido_ref",
	"software_pin_compartido_ref",
	"test_set_id",
	"certificado_url",
	"certificado_clave_ref",
	"llave_tecnica",
	"token_emisor_ref",
}

func EmpresaDIANConfigSecretColumns() []string {
	out := make([]string, len(empresaDIANConfigSecretColumns))
	copy(out, empresaDIANConfigSecretColumns)
	return out
}

func empresaDIANConfigSecretPurpose(empresaID int64, field string) (string, error) {
	field = strings.ToLower(strings.TrimSpace(field))
	if empresaID <= 0 {
		return "", fmt.Errorf("empresa_id invalido para secreto DIAN")
	}
	allowed := false
	for _, candidate := range empresaDIANConfigSecretColumns {
		if field == candidate {
			allowed = true
			break
		}
	}
	if !allowed {
		return "", fmt.Errorf("campo secreto DIAN no permitido")
	}
	// EncryptStringForPurpose serializa el proposito separado por dos puntos;
	// el propio proposito no puede contener ese delimitador.
	return fmt.Sprintf("dian-config-empresa-%d-campo-%s", empresaID, field), nil
}

func EncryptEmpresaDIANConfigSecret(empresaID int64, field, plain string) (string, error) {
	plain = strings.TrimSpace(plain)
	if plain == "" {
		return "", nil
	}
	purpose, err := empresaDIANConfigSecretPurpose(empresaID, field)
	if err != nil {
		return "", err
	}
	if strings.HasPrefix(plain, "v1:") {
		if _, err := secure.DecryptStringForPurpose(purpose, plain); err != nil {
			return "", fmt.Errorf("secreto DIAN cifrado con proposito invalido: %w", err)
		}
		return plain, nil
	}
	return secure.EncryptStringForPurpose(purpose, plain)
}

func DecryptEmpresaDIANConfigSecret(empresaID int64, field, encrypted string) (string, error) {
	encrypted = strings.TrimSpace(encrypted)
	if encrypted == "" {
		return "", nil
	}
	purpose, err := empresaDIANConfigSecretPurpose(empresaID, field)
	if err != nil {
		return "", err
	}
	if !strings.HasPrefix(encrypted, "v1:") {
		return "", fmt.Errorf("secreto DIAN heredado sin cifrar")
	}
	return secure.DecryptStringForPurpose(purpose, encrypted)
}

func applyEmpresaDIANConfigSecretsTx(ctx context.Context, tx *sql.Tx) error {
	if tx == nil {
		return fmt.Errorf("migration transaction is required")
	}
	var registered sql.NullString
	if err := tx.QueryRowContext(ctx, `SELECT to_regclass('empresa_dian_configuracion')`).Scan(&registered); err != nil {
		return fmt.Errorf("verify empresa_dian_configuracion before secret encryption: %w", err)
	}
	if !registered.Valid || strings.TrimSpace(registered.String) == "" {
		return nil
	}

	type row struct {
		id        int64
		empresaID int64
		values    []string
	}
	selectColumns := make([]string, 0, len(empresaDIANConfigSecretColumns))
	for _, field := range empresaDIANConfigSecretColumns {
		selectColumns = append(selectColumns, "COALESCE("+field+", '')")
	}
	// #nosec G202 -- column names come only from the fixed server-side catalog above.
	rows, err := tx.QueryContext(ctx, "SELECT id, empresa_id, "+strings.Join(selectColumns, ", ")+" FROM empresa_dian_configuracion")
	if err != nil {
		return fmt.Errorf("read DIAN secrets for encryption: %w", err)
	}
	items := make([]row, 0)
	for rows.Next() {
		item := row{values: make([]string, len(empresaDIANConfigSecretColumns))}
		dest := make([]interface{}, 0, len(item.values)+2)
		dest = append(dest, &item.id, &item.empresaID)
		for index := range item.values {
			dest = append(dest, &item.values[index])
		}
		if err := rows.Scan(dest...); err != nil {
			_ = rows.Close()
			return err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return err
	}
	if err := rows.Close(); err != nil {
		return err
	}

	for _, item := range items {
		for index, field := range empresaDIANConfigSecretColumns {
			plain := strings.TrimSpace(item.values[index])
			if plain == "" {
				continue
			}
			encrypted, err := EncryptEmpresaDIANConfigSecret(item.empresaID, field, plain)
			if err != nil {
				return fmt.Errorf("encrypt %s for empresa %d: %w", field, item.empresaID, err)
			}
			if encrypted == plain {
				continue
			}
			// #nosec G202 -- field comes only from empresaDIANConfigSecretColumns.
			query := "UPDATE empresa_dian_configuracion SET " + field + " = $1 WHERE id = $2 AND empresa_id = $3 AND COALESCE(" + field + ", '') = $4"
			result, err := tx.ExecContext(ctx, query, encrypted, item.id, item.empresaID, plain)
			if err != nil {
				return fmt.Errorf("persist encrypted %s for empresa %d: %w", field, item.empresaID, err)
			}
			affected, err := result.RowsAffected()
			if err != nil || affected != 1 {
				return fmt.Errorf("encrypted %s for empresa %d was not updated exactly once", field, item.empresaID)
			}
		}
	}
	return nil
}
