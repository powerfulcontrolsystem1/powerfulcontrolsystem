package db

import (
	"database/sql"
	"fmt"
)

const empresaControlElectricoScenesSchemaFingerprint = "empresa_control_electrico_scenes:v1:tenant-scenes-ordered-relay-targets"

func applyEmpresaControlElectricoScenesSchemaTx(tx *sql.Tx) error {
	if tx == nil {
		return fmt.Errorf("migration transaction is required")
	}
	statements := []string{
		`CREATE TABLE IF NOT EXISTS empresa_control_electrico_escenas (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL,
			nombre TEXT NOT NULL,
			descripcion TEXT,
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			estado TEXT NOT NULL DEFAULT 'activo'
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_empresa_control_electrico_escena_nombre ON empresa_control_electrico_escenas(empresa_id, LOWER(nombre))`,
		`CREATE TABLE IF NOT EXISTS empresa_control_electrico_escena_items (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL,
			escena_id BIGINT NOT NULL,
			rele_id BIGINT NOT NULL,
			estado_objetivo TEXT NOT NULL,
			orden INTEGER NOT NULL DEFAULT 0,
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP)
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_empresa_control_electrico_escena_item ON empresa_control_electrico_escena_items(empresa_id, escena_id, rele_id)`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_control_electrico_escena_items_orden ON empresa_control_electrico_escena_items(empresa_id, escena_id, orden, id)`,
	}
	for _, statement := range statements {
		if _, err := execTxSQLCompat(tx, statement); err != nil {
			return err
		}
	}
	return nil
}
