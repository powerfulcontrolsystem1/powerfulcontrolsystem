package db

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/you/pos-backend/secure"
)

// EmpresaAIOpenAIProviderConfig keeps an optional customer-owned OpenAI
// credential. The API key is deliberately never returned by this package.
type EmpresaAIOpenAIProviderConfig struct {
	EmpresaID     int64  `json:"empresa_id"`
	Habilitado    bool   `json:"habilitado"`
	ClaveCargada  bool   `json:"clave_cargada"`
	ActualizadoEn string `json:"actualizado_en,omitempty"`
}

const empresaAIOpenAIEncryptionPurpose = "empresa-openai-provider"

func EnsureEmpresaAIOpenAIProviderSchema(dbConn *sql.DB) error {
	if SchemaBootstrapDisabled() {
		return nil
	}
	if dbConn == nil {
		return errors.New("db connection is nil")
	}
	_, err := execSQLCompat(dbConn, `CREATE TABLE IF NOT EXISTS empresa_ai_openai_proveedor_configuracion (
		empresa_id BIGINT PRIMARY KEY,
		habilitado INTEGER NOT NULL DEFAULT 0,
		api_key_cifrada TEXT NOT NULL DEFAULT '',
		usuario_actualizacion TEXT NOT NULL DEFAULT '',
		fecha_actualizacion TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`)
	return err
}

func GetEmpresaAIOpenAIProviderConfig(dbConn *sql.DB, empresaID int64) (EmpresaAIOpenAIProviderConfig, error) {
	if err := EnsureEmpresaAIOpenAIProviderSchema(dbConn); err != nil {
		return EmpresaAIOpenAIProviderConfig{}, err
	}
	if empresaID <= 0 {
		return EmpresaAIOpenAIProviderConfig{}, fmt.Errorf("empresa_id invalido")
	}
	cfg := EmpresaAIOpenAIProviderConfig{EmpresaID: empresaID}
	var enabled int
	var encrypted string
	err := queryRowSQLCompat(dbConn, `SELECT COALESCE(habilitado,0), COALESCE(api_key_cifrada,''), COALESCE(fecha_actualizacion::text,'') FROM empresa_ai_openai_proveedor_configuracion WHERE empresa_id=?`, empresaID).Scan(&enabled, &encrypted, &cfg.ActualizadoEn)
	if errors.Is(err, sql.ErrNoRows) {
		return cfg, nil
	}
	if err != nil {
		return EmpresaAIOpenAIProviderConfig{}, err
	}
	cfg.Habilitado = enabled != 0
	cfg.ClaveCargada = strings.TrimSpace(encrypted) != ""
	return cfg, nil
}

// GetEmpresaAIOpenAIProviderKey is intentionally server-only. It decrypts the
// credential only for an outbound provider request and never serializes it.
func GetEmpresaAIOpenAIProviderKey(dbConn *sql.DB, empresaID int64) (string, bool, error) {
	if err := EnsureEmpresaAIOpenAIProviderSchema(dbConn); err != nil {
		return "", false, err
	}
	if empresaID <= 0 {
		return "", false, fmt.Errorf("empresa_id invalido")
	}
	var enabled int
	var encrypted string
	err := queryRowSQLCompat(dbConn, `SELECT COALESCE(habilitado,0), COALESCE(api_key_cifrada,'') FROM empresa_ai_openai_proveedor_configuracion WHERE empresa_id=?`, empresaID).Scan(&enabled, &encrypted)
	if errors.Is(err, sql.ErrNoRows) || enabled == 0 || strings.TrimSpace(encrypted) == "" {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	plain, err := secure.DecryptStringForPurpose(empresaAIOpenAIEncryptionPurpose, encrypted)
	if err != nil {
		return "", false, fmt.Errorf("no se pudo descifrar la clave OpenAI empresarial")
	}
	plain = strings.TrimSpace(plain)
	if plain == "" {
		return "", false, nil
	}
	return plain, true, nil
}

// UpsertEmpresaAIOpenAIProviderConfig retains a stored key when apiKey is
// empty. Sending replaceKey=true replaces it; sending clearKey=true erases it.
func UpsertEmpresaAIOpenAIProviderConfig(dbConn *sql.DB, empresaID int64, enabled bool, apiKey, actor string, replaceKey, clearKey bool) (EmpresaAIOpenAIProviderConfig, error) {
	if err := EnsureEmpresaAIOpenAIProviderSchema(dbConn); err != nil {
		return EmpresaAIOpenAIProviderConfig{}, err
	}
	if empresaID <= 0 {
		return EmpresaAIOpenAIProviderConfig{}, fmt.Errorf("empresa_id invalido")
	}
	apiKey = strings.TrimSpace(apiKey)
	if len(apiKey) > 512 {
		return EmpresaAIOpenAIProviderConfig{}, fmt.Errorf("la clave OpenAI supera el tamano permitido")
	}
	if clearKey && apiKey != "" {
		return EmpresaAIOpenAIProviderConfig{}, fmt.Errorf("no combines eliminar clave con una clave nueva")
	}
	if enabled && clearKey {
		return EmpresaAIOpenAIProviderConfig{}, fmt.Errorf("carga una clave OpenAI antes de activar el proveedor propio")
	}

	current, err := GetEmpresaAIOpenAIProviderConfig(dbConn, empresaID)
	if err != nil {
		return EmpresaAIOpenAIProviderConfig{}, err
	}
	if enabled && !current.ClaveCargada && apiKey == "" {
		return EmpresaAIOpenAIProviderConfig{}, fmt.Errorf("carga una clave OpenAI antes de activar el proveedor propio")
	}

	encrypted := ""
	if apiKey != "" {
		encrypted, err = secure.EncryptStringForPurpose(empresaAIOpenAIEncryptionPurpose, apiKey)
		if err != nil {
			return EmpresaAIOpenAIProviderConfig{}, fmt.Errorf("no se pudo cifrar la clave OpenAI empresarial")
		}
	}
	if !replaceKey && apiKey == "" && !clearKey {
		encrypted = ""
	}
	_, err = execSQLCompat(dbConn, `INSERT INTO empresa_ai_openai_proveedor_configuracion (empresa_id,habilitado,api_key_cifrada,usuario_actualizacion,fecha_actualizacion)
		VALUES (?,?,?,?,CURRENT_TIMESTAMP)
		ON CONFLICT (empresa_id) DO UPDATE SET
			habilitado=EXCLUDED.habilitado,
			api_key_cifrada=CASE WHEN ? <> '' THEN ? WHEN ? <> 0 THEN '' ELSE empresa_ai_openai_proveedor_configuracion.api_key_cifrada END,
			usuario_actualizacion=EXCLUDED.usuario_actualizacion,
			fecha_actualizacion=CURRENT_TIMESTAMP`, empresaID, boolToInt(enabled), encrypted, strings.TrimSpace(actor), encrypted, encrypted, boolToInt(clearKey))
	if err != nil {
		return EmpresaAIOpenAIProviderConfig{}, err
	}
	return GetEmpresaAIOpenAIProviderConfig(dbConn, empresaID)
}
