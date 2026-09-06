package db

import (
	"context"
	"database/sql"
	"fmt"
)

const empresaOfflineCashReportSchemaFingerprint = "empresa-offline-cash-report:v1:immutable-sale-operation-cash-register-breakdown"

func applyEmpresaOfflineCashReportSchemaTx(_ context.Context, tx *sql.Tx) error {
	if tx == nil {
		return fmt.Errorf("migration transaction is required")
	}
	statements := []string{
		`ALTER TABLE empresa_ventas_offline_sync ADD COLUMN IF NOT EXISTS operacion_codigo TEXT`,
		`ALTER TABLE empresa_ventas_offline_sync ADD COLUMN IF NOT EXISTS cierre_caja_id BIGINT DEFAULT 0`,
		`ALTER TABLE empresa_ventas_offline_sync ADD COLUMN IF NOT EXISTS caja_codigo TEXT`,
		`ALTER TABLE empresa_ventas_offline_sync ADD COLUMN IF NOT EXISTS caja_turno TEXT`,
		`ALTER TABLE empresa_ventas_offline_sync ADD COLUMN IF NOT EXISTS caja_sucursal_id BIGINT DEFAULT 0`,
		`ALTER TABLE empresa_ventas_offline_sync ADD COLUMN IF NOT EXISTS metodo_pago TEXT`,
		`ALTER TABLE empresa_ventas_offline_sync ADD COLUMN IF NOT EXISTS moneda TEXT`,
		`ALTER TABLE empresa_ventas_offline_sync ADD COLUMN IF NOT EXISTS monto_total NUMERIC(18,2) DEFAULT 0`,
		`ALTER TABLE empresa_ventas_offline_sync ADD COLUMN IF NOT EXISTS fecha_venta TEXT`,
		`ALTER TABLE empresa_ventas_offline_sync ADD COLUMN IF NOT EXISTS estacion_nombre TEXT`,
		`ALTER TABLE empresa_ventas_offline_sync ADD COLUMN IF NOT EXISTS usuario_cajero TEXT`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_empresa_ventas_offline_operacion
			ON empresa_ventas_offline_sync(empresa_id, operacion_codigo)
			WHERE COALESCE(operacion_codigo, '') <> ''`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_ventas_offline_corte_caja
			ON empresa_ventas_offline_sync(empresa_id, cierre_caja_id, caja_codigo, fecha_venta DESC)
			WHERE estado_sync = 'sincronizado'`,
	}
	for _, statement := range statements {
		if _, err := tx.Exec(statement); err != nil {
			return err
		}
	}
	return nil
}
