package db

import (
	"context"
	"database/sql"
)

const empresaSaleAccountingIdempotencyFingerprint = "empresa-sale-accounting-idempotency:v1:unique-active-paid-cart-event"

func applyEmpresaSaleAccountingIdempotencyTx(_ context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS ux_empresa_eventos_contables_venta_pagada_carrito
		ON empresa_eventos_contables(empresa_id, modulo, evento, entidad, entidad_id)
		WHERE modulo = 'ventas' AND evento = 'venta_pagada' AND entidad = 'carrito_compra'
			AND entidad_id IS NOT NULL AND estado = 'activo'`)
	return err
}
