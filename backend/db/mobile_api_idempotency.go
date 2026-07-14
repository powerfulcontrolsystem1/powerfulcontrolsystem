package db

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

// MobileAPIIdempotencyRecord stores the completed result of a mutation issued
// by a mobile device. The client supplied key is never persisted in plaintext.
type MobileAPIIdempotencyRecord struct {
	EmpresaID    int64
	Operation    string
	KeyHash      string
	RequestHash  string
	Status       string
	ResponseCode int
	ResponseJSON string
}

var ErrMobileAPIIdempotencyConflict = errors.New("mobile api idempotency key conflict")

func mobileAPIHash(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

// EnsureMobileAPIIdempotencySchema is additive and tenant-scoped. It allows a
// retry to return the original successful response instead of charging or
// emitting a fiscal document a second time.
func EnsureMobileAPIIdempotencySchema(dbConn *sql.DB) error {
	if dbConn == nil {
		return fmt.Errorf("db connection is nil")
	}
	_, err := execSQLCompat(dbConn, `CREATE TABLE IF NOT EXISTS empresa_mobile_api_idempotencia (
		empresa_id BIGINT NOT NULL,
		operacion TEXT NOT NULL,
		clave_hash TEXT NOT NULL,
		solicitud_hash TEXT NOT NULL,
		estado TEXT NOT NULL DEFAULT 'procesando',
		codigo_respuesta INTEGER NOT NULL DEFAULT 0,
		respuesta_json TEXT NOT NULL DEFAULT '',
		fecha_creacion TEXT DEFAULT CAST(CURRENT_TIMESTAMP AS TEXT),
		fecha_actualizacion TEXT DEFAULT CAST(CURRENT_TIMESTAMP AS TEXT),
		PRIMARY KEY (empresa_id, operacion, clave_hash)
	)`)
	if err != nil {
		return err
	}
	_, err = execSQLCompat(dbConn, `CREATE INDEX IF NOT EXISTS ix_empresa_mobile_api_idempotencia_actualizacion
		ON empresa_mobile_api_idempotencia (empresa_id, fecha_actualizacion DESC)`)
	return err
}

// ClaimMobileAPIIdempotency atomically reserves an operation. claimed is true
// only for the request that may execute the underlying financial mutation.
func ClaimMobileAPIIdempotency(dbConn *sql.DB, empresaID int64, operation, key, requestBody string) (*MobileAPIIdempotencyRecord, bool, error) {
	if empresaID <= 0 || strings.TrimSpace(operation) == "" || strings.TrimSpace(key) == "" {
		return nil, false, fmt.Errorf("idempotency input invalido")
	}
	if err := EnsureMobileAPIIdempotencySchema(dbConn); err != nil {
		return nil, false, err
	}
	record := &MobileAPIIdempotencyRecord{
		EmpresaID:   empresaID,
		Operation:   strings.TrimSpace(operation),
		KeyHash:     mobileAPIHash(key),
		RequestHash: mobileAPIHash(requestBody),
		Status:      "procesando",
	}
	result, err := execSQLCompat(dbConn, `INSERT INTO empresa_mobile_api_idempotencia (
		empresa_id, operacion, clave_hash, solicitud_hash, estado, codigo_respuesta, respuesta_json, fecha_creacion, fecha_actualizacion
	) VALUES (?, ?, ?, ?, 'procesando', 0, '', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	ON CONFLICT (empresa_id, operacion, clave_hash) DO NOTHING`, record.EmpresaID, record.Operation, record.KeyHash, record.RequestHash)
	if err != nil {
		return nil, false, err
	}
	if affected, _ := result.RowsAffected(); affected == 1 {
		return record, true, nil
	}

	stored, err := GetMobileAPIIdempotency(dbConn, record.EmpresaID, record.Operation, record.KeyHash)
	if err != nil {
		return nil, false, err
	}
	if stored.RequestHash != record.RequestHash {
		return nil, false, ErrMobileAPIIdempotencyConflict
	}
	return stored, false, nil
}

func GetMobileAPIIdempotency(dbConn *sql.DB, empresaID int64, operation, keyHash string) (*MobileAPIIdempotencyRecord, error) {
	row := queryRowSQLCompat(dbConn, `SELECT empresa_id, operacion, clave_hash, solicitud_hash, estado,
		COALESCE(codigo_respuesta, 0), COALESCE(respuesta_json, '')
		FROM empresa_mobile_api_idempotencia
		WHERE empresa_id = ? AND operacion = ? AND clave_hash = ?`, empresaID, operation, keyHash)
	var out MobileAPIIdempotencyRecord
	if err := row.Scan(&out.EmpresaID, &out.Operation, &out.KeyHash, &out.RequestHash, &out.Status, &out.ResponseCode, &out.ResponseJSON); err != nil {
		return nil, err
	}
	return &out, nil
}

func CompleteMobileAPIIdempotency(dbConn *sql.DB, record *MobileAPIIdempotencyRecord, responseCode int, responseJSON string) error {
	if record == nil || responseCode < 200 || responseCode >= 300 {
		return fmt.Errorf("resultado de idempotencia invalido")
	}
	_, err := execSQLCompat(dbConn, `UPDATE empresa_mobile_api_idempotencia
		SET estado = 'completado', codigo_respuesta = ?, respuesta_json = ?, fecha_actualizacion = CURRENT_TIMESTAMP
		WHERE empresa_id = ? AND operacion = ? AND clave_hash = ? AND solicitud_hash = ?`,
		responseCode, responseJSON, record.EmpresaID, record.Operation, record.KeyHash, record.RequestHash)
	return err
}

func AbandonMobileAPIIdempotency(dbConn *sql.DB, record *MobileAPIIdempotencyRecord) error {
	if record == nil {
		return nil
	}
	_, err := execSQLCompat(dbConn, `DELETE FROM empresa_mobile_api_idempotencia
		WHERE empresa_id = ? AND operacion = ? AND clave_hash = ? AND solicitud_hash = ? AND estado = 'procesando'`,
		record.EmpresaID, record.Operation, record.KeyHash, record.RequestHash)
	return err
}
