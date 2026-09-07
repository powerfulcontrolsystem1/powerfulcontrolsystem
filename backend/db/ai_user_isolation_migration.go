package db

import (
	"database/sql"
	"fmt"
)

const empresaAIUserIsolationSchemaFingerprint = "empresa_ai_user_isolation:v1:tenant-user-preferences-private-history"

func applyEmpresaAIUserIsolationSchemaTx(tx *sql.Tx) error {
	if tx == nil {
		return fmt.Errorf("migration transaction is required")
	}
	statements := []string{
		`CREATE TABLE IF NOT EXISTS empresa_ai_memoria (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL,
			usuario_id TEXT NOT NULL DEFAULT '',
			tipo TEXT NOT NULL,
			clave TEXT NOT NULL,
			valor_json TEXT NOT NULL,
			consentida BOOLEAN NOT NULL DEFAULT FALSE,
			fecha_expiracion TIMESTAMP,
			fecha_creacion TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			fecha_actualizacion TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(empresa_id, usuario_id, tipo, clave)
		)`,
		`CREATE TABLE IF NOT EXISTS empresa_ai_usuario_preferencias (
			empresa_id BIGINT NOT NULL,
			usuario_id TEXT NOT NULL,
			model_id TEXT NOT NULL,
			modo_asistente TEXT NOT NULL DEFAULT 'operativo',
			agent_id TEXT NOT NULL DEFAULT 'general',
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			PRIMARY KEY (empresa_id, usuario_id)
		)`,
		`ALTER TABLE IF EXISTS empresa_ai_consultas ADD COLUMN IF NOT EXISTS conversation_id TEXT NOT NULL DEFAULT ''`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_ai_usuario_preferencias_empresa_usuario ON empresa_ai_usuario_preferencias(empresa_id, usuario_id, estado)`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_ai_consultas_empresa_usuario_fecha ON empresa_ai_consultas(empresa_id, usuario_creador, fecha_consulta DESC, id DESC)`,
	}
	for _, statement := range statements {
		if _, err := execTxSQLCompat(tx, statement); err != nil {
			return err
		}
	}
	return nil
}
