package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
)

const empresaFacturacionNumeracionReservasFingerprint = "empresa-facturacion-reservas:v1:tenant-document:unique-fiscal-number:immutable-legal-snapshot"

// Only pcs-migrate applies this schema. Do not backfill or renumber fiscal history.
func applyEmpresaFacturacionNumeracionReservasTx(ctx context.Context, tx *sql.Tx) error {
	for _, statement := range []string{
		`CREATE TABLE empresa_facturacion_reservas_numeracion (
			empresa_id BIGINT NOT NULL CHECK (empresa_id > 0),
			documento_codigo TEXT NOT NULL CHECK (documento_codigo <> ''),
			pais_codigo TEXT NOT NULL CHECK (pais_codigo = 'CO'),
			ambiente TEXT NOT NULL CHECK (ambiente IN ('produccion', 'sandbox')),
			numero_legal TEXT NOT NULL CHECK (numero_legal <> ''),
			monto_total NUMERIC(18,2) NOT NULL CHECK (monto_total >= 0),
			moneda TEXT NOT NULL,
			legal_json TEXT NOT NULL,
			fecha_reserva TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (empresa_id, documento_codigo),
			UNIQUE (empresa_id, pais_codigo, ambiente, numero_legal)
		)`,
		`CREATE INDEX ix_factura_numero_legal_reserva ON empresa_facturacion_documentos
			(empresa_id, UPPER(REPLACE(TRIM(numero_legal), '-', '')))
			WHERE tipo_documento = 'factura_electronica'`,
		`CREATE INDEX ix_factura_retry_numero_legal_reserva ON facturacion_electronica_reintentos
			(empresa_id, UPPER(REPLACE(TRIM(numero_legal), '-', '')))
			WHERE tipo_documento = 'factura_electronica'`,
	} {
		if _, err := tx.ExecContext(ctx, statement); err != nil {
			return err
		}
	}
	return nil
}

func getFacturacionNumeroReservado(ctx context.Context, tx *sql.Tx, empresaID int64, codigo, pais, moneda string, monto float64) (*FacturacionDocumentoLegal, error) {
	var raw, storedAmount, storedCurrency, storedCountry string
	err := tx.QueryRowContext(ctx, `SELECT legal_json, monto_total::text, moneda, pais_codigo
		FROM empresa_facturacion_reservas_numeracion WHERE empresa_id = ? AND documento_codigo = ?`, empresaID, codigo).
		Scan(&raw, &storedAmount, &storedCurrency, &storedCountry)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if storedCountry != pais || storedCurrency != moneda || storedAmount != fmt.Sprintf("%.2f", monto) {
		return nil, fmt.Errorf("la reserva fiscal ya existe con otro pais, moneda o importe; no se puede renumerar la venta")
	}
	var legal FacturacionDocumentoLegal
	if err := json.Unmarshal([]byte(raw), &legal); err != nil {
		return nil, err
	}
	if legal.EmpresaID != empresaID || legal.PaisCodigo != pais || legal.NumeroLegal == "" {
		return nil, fmt.Errorf("reserva fiscal inconsistente; requiere revision sin nueva numeracion")
	}
	return &legal, nil
}

func persistFacturacionNumeroReservado(ctx context.Context, tx *sql.Tx, codigo, moneda string, monto float64, legal *FacturacionDocumentoLegal) error {
	// Legacy documents/queues are never rewritten. A historical collision must be
	// reconciled from official evidence, not skipped by silently consuming a folio.
	var occupied bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS (
		SELECT 1 FROM empresa_facturacion_documentos
		WHERE empresa_id = ? AND tipo_documento = 'factura_electronica'
		AND ((documento_codigo = ? AND COALESCE(TRIM(numero_legal), '') <> '') OR UPPER(REPLACE(TRIM(numero_legal), '-', '')) = ?)
		UNION ALL SELECT 1 FROM facturacion_electronica_reintentos
		WHERE empresa_id = ? AND tipo_documento = 'factura_electronica'
		AND ((documento_codigo = ? AND COALESCE(TRIM(numero_legal), '') <> '') OR UPPER(REPLACE(TRIM(numero_legal), '-', '')) = ?)
	)`, legal.EmpresaID, codigo, legal.NumeroLegal, legal.EmpresaID, codigo, legal.NumeroLegal).Scan(&occupied); err != nil {
		return err
	}
	if occupied {
		return fmt.Errorf("el documento o numero fiscal ya figura en documentos o cola; concilie la numeracion antes de emitir")
	}
	raw, err := json.Marshal(legal)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO empresa_facturacion_reservas_numeracion
		(empresa_id, documento_codigo, pais_codigo, ambiente, numero_legal, monto_total, moneda, legal_json)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, legal.EmpresaID, codigo, legal.PaisCodigo, legal.Ambiente, legal.NumeroLegal, fmt.Sprintf("%.2f", monto), moneda, string(raw))
	return err
}
