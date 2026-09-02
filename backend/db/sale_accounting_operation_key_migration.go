package db

import (
	"context"
	"database/sql"
)

const empresaSaleAccountingOperationKeyFingerprint = "empresa-sale-accounting-operation-key:v1:durable-paid-session-key"

func applyEmpresaSaleAccountingOperationKeyTx(_ context.Context, tx *sql.Tx) error {
	for _, statement := range []string{
		`ALTER TABLE empresa_eventos_contables ADD COLUMN IF NOT EXISTS clave_idempotencia TEXT`,
		`DROP INDEX IF EXISTS ux_empresa_eventos_contables_venta_pagada_carrito`,
		`CREATE UNIQUE INDEX ux_empresa_eventos_contables_venta_pagada_carrito
			ON empresa_eventos_contables(empresa_id, clave_idempotencia)
			WHERE modulo = 'ventas' AND evento = 'venta_pagada' AND entidad = 'carrito_compra'
				AND NULLIF(TRIM(clave_idempotencia), '') IS NOT NULL AND estado = 'activo'`,
	} {
		if _, err := tx.Exec(statement); err != nil {
			return err
		}
	}
	return nil
}
