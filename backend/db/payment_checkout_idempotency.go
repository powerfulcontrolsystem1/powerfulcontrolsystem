package db

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

var ErrPaymentCheckoutIdempotencyConflict = errors.New("payment checkout idempotency key conflict")

const paymentIdempotencySchemaFingerprint = "payment-idempotency:v1:checkout-ledger:post-effect-ledger:activation-at-most-once"

type PaymentCheckoutIdempotency struct {
	Provider     string
	EmpresaID    int64
	KeyHash      string
	RequestHash  string
	Reference    string
	Status       string
	ResponseCode int
	ResponseJSON string
}

type PaymentPostEffectClaim struct {
	Provider        string
	PaymentRecordID int64
	Effect          string
}

func paymentCheckoutHash(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func normalizePaymentCheckoutProvider(provider string) (string, error) {
	provider = strings.ToLower(strings.TrimSpace(provider))
	if provider != "wompi" && provider != "epayco" {
		return "", fmt.Errorf("proveedor de checkout no soportado")
	}
	return provider, nil
}

// PaymentCheckoutReference derives a stable, opaque gateway reference from the
// client key. The raw idempotency key is never persisted or sent to a provider.
func PaymentCheckoutReference(provider string, licenciaID, empresaID int64, key string) (string, error) {
	provider, err := normalizePaymentCheckoutProvider(provider)
	if err != nil {
		return "", err
	}
	if licenciaID <= 0 || empresaID <= 0 || strings.TrimSpace(key) == "" {
		return "", fmt.Errorf("datos de referencia de checkout invalidos")
	}
	prefix := strings.ToUpper(provider)
	return fmt.Sprintf("%s-LIC-%d-EMP-%d-IDEM-%s", prefix, licenciaID, empresaID, paymentCheckoutHash(key)[:24]), nil
}

func EnsurePaymentCheckoutIdempotencySchema(dbConn *sql.DB) error {
	if dbConn == nil {
		return fmt.Errorf("db connection is nil")
	}
	statements := []string{
		`CREATE TABLE IF NOT EXISTS payment_checkout_idempotencia (
			proveedor TEXT NOT NULL,
			empresa_id BIGINT NOT NULL,
			clave_hash TEXT NOT NULL,
			solicitud_hash TEXT NOT NULL,
			referencia TEXT NOT NULL,
			estado TEXT NOT NULL DEFAULT 'procesando',
			codigo_respuesta INTEGER NOT NULL DEFAULT 0,
			respuesta_json TEXT NOT NULL DEFAULT '',
			fecha_creacion TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			fecha_actualizacion TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			fecha_completado TIMESTAMPTZ,
			PRIMARY KEY (proveedor, empresa_id, clave_hash),
			UNIQUE (proveedor, referencia)
		)`,
		`CREATE INDEX IF NOT EXISTS ix_payment_checkout_idempotencia_estado
			ON payment_checkout_idempotencia (estado, fecha_actualizacion)`,
		`CREATE TABLE IF NOT EXISTS payment_post_effect_idempotencia (
			proveedor TEXT NOT NULL,
			pago_registro_id BIGINT NOT NULL,
			efecto TEXT NOT NULL,
			estado TEXT NOT NULL DEFAULT 'procesando',
			intentos INTEGER NOT NULL DEFAULT 1,
			ultimo_error TEXT NOT NULL DEFAULT '',
			fecha_creacion TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			fecha_actualizacion TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			fecha_completado TIMESTAMPTZ,
			PRIMARY KEY (proveedor, pago_registro_id, efecto)
		)`,
		`CREATE INDEX IF NOT EXISTS ix_payment_post_effect_estado
			ON payment_post_effect_idempotencia (estado, fecha_actualizacion)`,
	}
	for _, statement := range statements {
		if _, err := execSQLCompat(dbConn, statement); err != nil {
			return err
		}
	}
	return nil
}

func applyPaymentIdempotencySchemaTx(_ context.Context, tx *sql.Tx) error {
	if tx == nil {
		return fmt.Errorf("payment idempotency migration transaction is nil")
	}
	statements := []string{
		`ALTER TABLE pagos_wompi ADD COLUMN IF NOT EXISTS licencia_activation_lease_until TIMESTAMPTZ`,
		`ALTER TABLE pagos_wompi ADD COLUMN IF NOT EXISTS licencia_activation_attempts INTEGER NOT NULL DEFAULT 0`,
		`ALTER TABLE pagos_epayco ADD COLUMN IF NOT EXISTS licencia_activation_lease_until TIMESTAMPTZ`,
		`ALTER TABLE pagos_epayco ADD COLUMN IF NOT EXISTS licencia_activation_attempts INTEGER NOT NULL DEFAULT 0`,
		`CREATE TABLE IF NOT EXISTS payment_checkout_idempotencia (
			proveedor TEXT NOT NULL,
			empresa_id BIGINT NOT NULL,
			clave_hash TEXT NOT NULL,
			solicitud_hash TEXT NOT NULL,
			referencia TEXT NOT NULL,
			estado TEXT NOT NULL DEFAULT 'procesando',
			codigo_respuesta INTEGER NOT NULL DEFAULT 0,
			respuesta_json TEXT NOT NULL DEFAULT '',
			fecha_creacion TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			fecha_actualizacion TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			fecha_completado TIMESTAMPTZ,
			PRIMARY KEY (proveedor, empresa_id, clave_hash),
			UNIQUE (proveedor, referencia)
		)`,
		`CREATE INDEX IF NOT EXISTS ix_payment_checkout_idempotencia_estado
			ON payment_checkout_idempotencia (estado, fecha_actualizacion)`,
		`CREATE TABLE IF NOT EXISTS payment_post_effect_idempotencia (
			proveedor TEXT NOT NULL,
			pago_registro_id BIGINT NOT NULL,
			efecto TEXT NOT NULL,
			estado TEXT NOT NULL DEFAULT 'procesando',
			intentos INTEGER NOT NULL DEFAULT 1,
			ultimo_error TEXT NOT NULL DEFAULT '',
			fecha_creacion TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			fecha_actualizacion TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			fecha_completado TIMESTAMPTZ,
			PRIMARY KEY (proveedor, pago_registro_id, efecto)
		)`,
		`CREATE INDEX IF NOT EXISTS ix_payment_post_effect_estado
			ON payment_post_effect_idempotencia (estado, fecha_actualizacion)`,
	}
	for _, statement := range statements {
		if _, err := tx.Exec(statement); err != nil {
			return err
		}
	}
	return nil
}

// ClaimPaymentPostEffect applies at-most-once semantics to effects outside the
// payment database transaction (for example SMTP). A processing/uncertain row
// is never reclaimed automatically because the remote outcome is unknown.
func ClaimPaymentPostEffect(dbConn *sql.DB, provider, transactionID, reference, effect string) (*PaymentPostEffectClaim, bool, error) {
	provider, err := normalizePaymentCheckoutProvider(provider)
	if err != nil {
		return nil, false, err
	}
	effect = strings.ToLower(strings.TrimSpace(effect))
	if effect == "" {
		return nil, false, fmt.Errorf("efecto de pago obligatorio")
	}
	if err := EnsurePaymentGatewaySchema(dbConn); err != nil {
		return nil, false, err
	}
	if err := EnsurePaymentCheckoutIdempotencySchema(dbConn); err != nil {
		return nil, false, err
	}
	table, _ := licenciaPaymentTable(provider)
	recordID, found, err := getLicenciaPaymentRowID(dbConn, table, transactionID, reference)
	if err != nil {
		return nil, false, err
	}
	if !found {
		return nil, false, fmt.Errorf("pago no registrado para reservar efecto")
	}
	claim := &PaymentPostEffectClaim{Provider: provider, PaymentRecordID: recordID, Effect: effect}
	result, err := execSQLCompat(dbConn, `INSERT INTO payment_post_effect_idempotencia (
		proveedor, pago_registro_id, efecto, estado, intentos, ultimo_error, fecha_creacion, fecha_actualizacion
	) VALUES (?, ?, ?, 'procesando', 1, '', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	ON CONFLICT (proveedor, pago_registro_id, efecto) DO UPDATE SET
		estado = 'procesando', intentos = payment_post_effect_idempotencia.intentos + 1,
		ultimo_error = '', fecha_actualizacion = CURRENT_TIMESTAMP
	WHERE payment_post_effect_idempotencia.estado = 'fallido'`, provider, recordID, effect)
	if err != nil {
		return nil, false, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return nil, false, err
	}
	return claim, affected == 1, nil
}

func FinishPaymentPostEffect(dbConn *sql.DB, claim *PaymentPostEffectClaim, status, errorMessage string) error {
	if claim == nil {
		return nil
	}
	status = strings.ToLower(strings.TrimSpace(status))
	if status != "completado" && status != "incierto" && status != "fallido" {
		return fmt.Errorf("estado final de efecto de pago invalido")
	}
	completedExpr := "NULL"
	if status == "completado" {
		completedExpr = "CURRENT_TIMESTAMP"
	}
	result, err := execSQLCompat(dbConn, `UPDATE payment_post_effect_idempotencia
		SET estado = ?, ultimo_error = ?, fecha_actualizacion = CURRENT_TIMESTAMP,
			fecha_completado = `+completedExpr+`
		WHERE proveedor = ? AND pago_registro_id = ? AND efecto = ? AND estado = 'procesando'`,
		status, strings.TrimSpace(errorMessage), claim.Provider, claim.PaymentRecordID, claim.Effect)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected != 1 {
		return fmt.Errorf("payment post-effect claim is no longer owned")
	}
	return nil
}

func ClaimPaymentCheckoutIdempotency(dbConn *sql.DB, provider string, empresaID int64, key, requestFingerprint, reference string) (*PaymentCheckoutIdempotency, bool, error) {
	provider, err := normalizePaymentCheckoutProvider(provider)
	if err != nil {
		return nil, false, err
	}
	if empresaID <= 0 || strings.TrimSpace(key) == "" || strings.TrimSpace(requestFingerprint) == "" || strings.TrimSpace(reference) == "" {
		return nil, false, fmt.Errorf("datos de idempotencia de checkout invalidos")
	}
	if err := EnsurePaymentCheckoutIdempotencySchema(dbConn); err != nil {
		return nil, false, err
	}
	claim := &PaymentCheckoutIdempotency{
		Provider:    provider,
		EmpresaID:   empresaID,
		KeyHash:     paymentCheckoutHash(key),
		RequestHash: paymentCheckoutHash(requestFingerprint),
		Reference:   strings.TrimSpace(reference),
		Status:      "procesando",
	}
	result, err := execSQLCompat(dbConn, `INSERT INTO payment_checkout_idempotencia (
		proveedor, empresa_id, clave_hash, solicitud_hash, referencia, estado,
		codigo_respuesta, respuesta_json, fecha_creacion, fecha_actualizacion
	) VALUES (?, ?, ?, ?, ?, 'procesando', 0, '', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	ON CONFLICT (proveedor, empresa_id, clave_hash) DO NOTHING`,
		claim.Provider, claim.EmpresaID, claim.KeyHash, claim.RequestHash, claim.Reference)
	if err != nil {
		return nil, false, err
	}
	if affected, _ := result.RowsAffected(); affected == 1 {
		return claim, true, nil
	}
	stored, err := GetPaymentCheckoutIdempotency(dbConn, claim.Provider, claim.EmpresaID, claim.KeyHash)
	if err != nil {
		return nil, false, err
	}
	if stored.RequestHash != claim.RequestHash || stored.Reference != claim.Reference {
		return nil, false, ErrPaymentCheckoutIdempotencyConflict
	}
	return stored, false, nil
}

func GetPaymentCheckoutIdempotency(dbConn *sql.DB, provider string, empresaID int64, keyHash string) (*PaymentCheckoutIdempotency, error) {
	row := queryRowSQLCompat(dbConn, `SELECT proveedor, empresa_id, clave_hash, solicitud_hash, referencia,
		COALESCE(estado, 'procesando'), COALESCE(codigo_respuesta, 0), COALESCE(respuesta_json, '')
		FROM payment_checkout_idempotencia
		WHERE proveedor = ? AND empresa_id = ? AND clave_hash = ?`, provider, empresaID, keyHash)
	var out PaymentCheckoutIdempotency
	if err := row.Scan(&out.Provider, &out.EmpresaID, &out.KeyHash, &out.RequestHash, &out.Reference, &out.Status, &out.ResponseCode, &out.ResponseJSON); err != nil {
		return nil, err
	}
	return &out, nil
}

func CompletePaymentCheckoutIdempotency(dbConn *sql.DB, claim *PaymentCheckoutIdempotency, responseCode int, responseJSON string) error {
	if claim == nil || responseCode < 200 || responseCode >= 300 || strings.TrimSpace(responseJSON) == "" {
		return fmt.Errorf("resultado de checkout idempotente invalido")
	}
	result, err := execSQLCompat(dbConn, `UPDATE payment_checkout_idempotencia
		SET estado = 'completado', codigo_respuesta = ?, respuesta_json = ?,
			fecha_actualizacion = CURRENT_TIMESTAMP, fecha_completado = CURRENT_TIMESTAMP
		WHERE proveedor = ? AND empresa_id = ? AND clave_hash = ? AND solicitud_hash = ? AND estado = 'procesando'`,
		responseCode, responseJSON, claim.Provider, claim.EmpresaID, claim.KeyHash, claim.RequestHash)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected != 1 {
		return fmt.Errorf("checkout idempotency claim is no longer owned")
	}
	return nil
}

func MarkPaymentCheckoutIdempotencyUncertain(dbConn *sql.DB, claim *PaymentCheckoutIdempotency) error {
	if claim == nil {
		return nil
	}
	_, err := execSQLCompat(dbConn, `UPDATE payment_checkout_idempotencia
		SET estado = 'incierto', fecha_actualizacion = CURRENT_TIMESTAMP
		WHERE proveedor = ? AND empresa_id = ? AND clave_hash = ? AND solicitud_hash = ? AND estado = 'procesando'`,
		claim.Provider, claim.EmpresaID, claim.KeyHash, claim.RequestHash)
	return err
}

func AbandonPaymentCheckoutIdempotency(dbConn *sql.DB, claim *PaymentCheckoutIdempotency) error {
	if claim == nil {
		return nil
	}
	_, err := execSQLCompat(dbConn, `DELETE FROM payment_checkout_idempotencia
		WHERE proveedor = ? AND empresa_id = ? AND clave_hash = ? AND solicitud_hash = ? AND estado = 'procesando'`,
		claim.Provider, claim.EmpresaID, claim.KeyHash, claim.RequestHash)
	return err
}
