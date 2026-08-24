package db

import (
	"context"
	"database/sql"
)

const administradorLoginThrottleSchemaFingerprint = "administrador-login-throttle:v1:email-pk:failed-attempts:window-started-at:locked-until"

func applyAdministradorLoginThrottleSchemaTx(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS administrador_login_intentos (
		email TEXT PRIMARY KEY,
		failed_attempts INTEGER NOT NULL DEFAULT 0,
		window_started_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
		locked_until TIMESTAMPTZ NOT NULL DEFAULT TIMESTAMPTZ 'epoch',
		updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`)
	return err
}
