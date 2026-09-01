package db

import (
	"context"
	"database/sql"
)

const empresaOperationalIdempotencyFingerprint = "empresa-operational-idempotency:v1:cxp-request-hash:offline-payload-claim:dian-retry-claim"

func applyEmpresaOperationalIdempotencyTx(_ context.Context, tx *sql.Tx) error {
	statements := []string{
		`ALTER TABLE empresa_cxp_pagos ADD COLUMN IF NOT EXISTS request_hash TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE empresa_ventas_offline_sync ADD COLUMN IF NOT EXISTS payload_hash TEXT`,
		`ALTER TABLE empresa_ventas_offline_sync ADD COLUMN IF NOT EXISTS processing_token TEXT`,
		`ALTER TABLE empresa_ventas_offline_sync ADD COLUMN IF NOT EXISTS processing_until TIMESTAMPTZ`,
		`ALTER TABLE facturacion_electronica_reintentos ADD COLUMN IF NOT EXISTS lease_token TEXT`,
		`ALTER TABLE facturacion_electronica_reintentos ADD COLUMN IF NOT EXISTS lease_until TIMESTAMPTZ`,
		`CREATE INDEX IF NOT EXISTS ix_fe_reintentos_lease ON facturacion_electronica_reintentos(empresa_id, lease_until)`,
	}
	for _, statement := range statements {
		if _, err := tx.Exec(statement); err != nil {
			return err
		}
	}
	return nil
}
