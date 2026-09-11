package db

import (
	"context"
	"database/sql"
	"fmt"
)

const empresaComisionesProductosSchemaFingerprint = "empresa-comisiones-productos:v1:configurable-product-lines-in-commission-base"

func applyEmpresaComisionesProductosSchemaTx(_ context.Context, tx *sql.Tx) error {
	if tx == nil {
		return fmt.Errorf("migration transaction is required")
	}
	_, err := tx.Exec(`ALTER TABLE empresa_comisiones_servicio_configuracion
		ADD COLUMN IF NOT EXISTS incluir_productos INTEGER DEFAULT 0`)
	return err
}
